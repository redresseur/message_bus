package context

import (
	"context"
	"errors"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"math/bits"
	"strconv"
	"sync/atomic"
	"time"
)

var logger = flogging.MustGetLogger("channel.context")

const (
	HeartBeat = `HEART_BEAT`
)

//	channel的上下文
type channelContextImpl struct {
	info *open_interface.ChannelInfo

	ctx context.Context

	// 廣播消息stream
	broadCastMsgStream chan *message.UnitMessage

	// 點對點消息stream
	p2pMsgStream chan *message.UnitMessage

	// 當前節點隊列
	endpoints open_interface.Storage

	catchMsgFuncs map[message.UnitMessage_MessageFlag]CatchMsgFunc
}

func WithCatchMsgFunc(messageFlag message.UnitMessage_MessageFlag, msgFunc CatchMsgFunc) func(impl *channelContextImpl) {
	return func(impl *channelContextImpl) {
		if impl.catchMsgFuncs == nil {
			impl.catchMsgFuncs = map[message.UnitMessage_MessageFlag]CatchMsgFunc{}
		}

		impl.catchMsgFuncs[messageFlag] = msgFunc
	}
}

func WithChannelContext(ctx context.Context, info *open_interface.ChannelInfo,
	storage open_interface.Storage, ops ...func(impl *channelContextImpl)) open_interface.ChannelContext {
	cm := &channelContextImpl{
		info:               info,
		broadCastMsgStream: make(chan *message.UnitMessage, 1024),
		p2pMsgStream:       make(chan *message.UnitMessage, 1024),
		endpoints:          storage,
	}

	for _, op := range ops {
		op(cm)
	}

	newCtx, cancel := context.WithCancel(ctx)
	cm.ctx = context.WithValue(newCtx, cm, cancel)

	go cm.messageProcess()
	return cm
}

func (cc *channelContextImpl) catch(messageFlag message.UnitMessage_MessageFlag) CatchMsgFunc {
	var res CatchMsgFunc = func(unitMessage *message.UnitMessage) error {
		return nil
	}

	if c, ok := cc.catchMsgFuncs[messageFlag]; ok {
		res = c
	}

	return res
}

func (cc *channelContextImpl) Cancel() {
	cc.ctx.Value(cc).(context.CancelFunc)()
}

func (cc *channelContextImpl) EndPoint(endPointId string) *open_interface.EndPoint {
	cu, err := cc.endpoints.Get(endPointId)
	if err != nil {
		return nil
	}

	return cu.(*open_interface.EndPoint)
}

func (cc *channelContextImpl) AddEndPoint(point *open_interface.EndPoint) error {
	if ep := cc.EndPoint(point.Id); ep != nil {
		// 关闭原有的rw handler
		ep.Ctx.Value(ep).(context.CancelFunc)()
		cc.endpoints.Delete(point.Id)

	}

	// 分配ctx给customer
	ctx, cancel := context.WithCancel(cc.ctx)
	ctx, _ = WithMultiValueContext(ctx, point, cancel)

	point.Ctx = ctx

	// 判断是否开启缓存
	if point.CacheEnable && point.Cache == nil {
		return errors.New("the cache is nil")
	}

	// 初始化的序列号为1
	point.Sequence = 1
	return cc.endpoints.Put(point.Id, point)
}

func (cc *channelContextImpl) RemoveEndPoint(pointId string) {
	if ep, err := cc.endpoints.Get(pointId); err == nil {
		ep.(*open_interface.EndPoint).Ctx.Value(ep).(context.CancelFunc)()
		cc.endpoints.Delete(pointId)
	}
}

func (cc *channelContextImpl) BindRW(pointId string, rw open_interface.EndPointIO) (context.Context, error) {
	ep := cc.EndPoint(pointId)
	if ep == nil {
		return nil, open_interface.ErrEndPointNotExisted
	}

	ep.RW = rw

	// 判断是否需要启动心跳
	if ep.KeepAliveEnable {
		if ep.Keeper != nil {
			ep.Keeper.Finish()
		}

		deathFunc := func() {
			// 心跳超时则移除该节点
			UpdateMultiValueContext(ep.Ctx, HeartBeat, open_interface.ErrHeartBeatTimeOut)
			cc.RemoveEndPoint(ep.Id)
		}
		ep.Keeper = open_interface.NewKeeper(ep.Ctx, ep.HeartBeatDuration, deathFunc)
		go ep.Keeper.Run()
	}

	// 启动一个协程，用来接收消息
	go cc.recvMsgFromEndPoint(ep)
	return ep.Ctx, nil
}

func (cc *channelContextImpl) endPointList() <-chan *open_interface.EndPoint {
	// 要用同步指针
	output := make(chan *open_interface.EndPoint, 1024)
	go func() {
		defer close(output)
		cc.endpoints.Range(func(key, v interface{}) bool {
			select {
			case output <- v.(*open_interface.EndPoint):
			case <-time.After(100 * time.Millisecond): // 100 毫秒的等待延时
				return false
			}

			return true
		})
	}()

	return output
}

func (cc *channelContextImpl) recvMsgFromEndPoint(point *open_interface.EndPoint) {
	for {
		msg, err := point.RW.Read()
		if err != nil || point.Ctx.Err() != nil {
			logger.Errorf("Recv message from %s: %v", point.Id, err)
			break
		}

		if point.KeepAliveEnable && point.Keeper != nil {
			point.Keeper.Update()
		}

		// 检查ack 和 seq
		// 如果msg.ack 大于 point.seq， 说明对方串包了
		if msg.Ack > atomic.LoadUint32(&point.Sequence) {
			logger.Warnf("Msg %+v: Ack %d greater than Seq %d", msg, msg.Ack, point.Sequence)
			continue
		}

		// msg.seq 小于 point.ack, 说明对方重复发包了
		if msg.Seq <= atomic.LoadUint32(&point.Ack) {
			continue
		}

		// ack 累计增长
		switch msg.Flag {
		// 心跳包
		case message.UnitMessage_HEART_BEAT:
			// TODO: 更新心跳状态
			atomic.AddUint32(&point.Ack, 1)
		case message.UnitMessage_REAL_TIME:
			// TODO: 发送消息
			atomic.AddUint32(&point.Ack, 1)
		case message.UnitMessage_SYNC:
			{
				// ack 累计
				atomic.AddUint32(&point.Ack, 1)
				// 移除已经确认的消息
				stableUint64, err := strconv.ParseUint(string(msg.Payload), 10, bits.UintSize)
				if err != nil {
					logger.Errorf("the payload of sync message: %v is not uint type", string(msg.Payload))
					break
				}

				stable := uint32(stableUint64)
				point.Cache.Remove(-1, int32(stable-1)) // cache的起始值为0,所以此處减1

				offset := uint32(point.Sequence - stable)
				beginIndex := stableUint64 // cache的起始值为0,所以此處不加1
				res, err := point.Cache.Seek(beginIndex, offset)
				if err != nil {
					logger.Warningf("Seek from %d offset %d: %v", msg.Ack, offset, err)
					break
				}

				//如果最后的消息是实时消息，要替换为一个同步消息
				lenSeek := len(res)
				if lenSeek > 0 {
					for _, v := range res[:lenSeek-1] {
						if unitMsg, ok := v.(*message.UnitMessage); ok {
							// TODO: 更新时间戳
							cc.reSendMsgToEndPoint(point, unitMsg)
						}
					}

					latest, ok := res[lenSeek-1].(*message.UnitMessage)
					if !ok {
						latest = &message.UnitMessage{
							ChannelId:     msg.ChannelId,
							DstEndPointId: []string{msg.SrcEndPointId},
							Flag:          message.UnitMessage_SYNC,
							Seq:           stable + uint32(lenSeek),
						}
					}

					cc.reSendMsgToEndPoint(point, latest)
				}

				point.Cache.Seek(beginIndex, 0)
			}
		case message.UnitMessage_COMMON:
			atomic.AddUint32(&point.Ack, 1)
		}

		// 在末尾添加自定义处理
		if err := cc.catch(msg.Flag)(msg); err != nil {
			logger.Warningf("catch the common message: %v", err)
		}
	}
}

// 重发
func (cc *channelContextImpl) reSendMsgToEndPoint(point *open_interface.EndPoint, msg *message.UnitMessage) {
	// 增加计数
	point.Locker().Lock()
	defer point.Locker().Unlock()

	//msg.Seq = seq
	logger.Debugf("ReSend Message: %+v", msg)
	msg.Ack = atomic.LoadUint32(&point.Ack)
	if err := point.RW.Write(msg); err != nil {
		logger.Errorf("Write Message Failure: %v", err)
	}
}

func (cc *channelContextImpl) sendMsgToEndPoint(point *open_interface.EndPoint, msg *message.UnitMessage) {
	// 增加计数
	point.Locker().Lock()
	defer point.Locker().Unlock()

	// 复制消息副本
	dMsg := new(message.UnitMessage)
	dMsg.Seq = atomic.LoadUint32(&point.Sequence)
	dMsg.Ack = atomic.LoadUint32(&point.Ack)
	dMsg.DstEndPointId = []string{point.Id}
	dMsg.ChannelId = msg.ChannelId
	dMsg.SrcEndPointId = msg.SrcEndPointId
	dMsg.Type = msg.Type
	dMsg.Flag = msg.Flag
	dMsg.Timestamp = msg.Timestamp
	dMsg.Metadata = msg.Metadata
	dMsg.Payload = msg.Payload

	logger.Debugf("Send Message: %+v", dMsg)
	if err := point.RW.Write(dMsg); err != nil {
		logger.Errorf("Write Message Failure: %v", err)
	}

	atomic.AddUint32(&point.Sequence, 1)

	// 添加到缓存中
	// 目前除了普通消息外都不缓存
	if point.CacheEnable {
		if dMsg.Flag == message.UnitMessage_COMMON {
			point.Cache.Push(dMsg)
		} else {
			point.Cache.Push(nil)
		}
	}
}

func (cc *channelContextImpl) messageProcess() {
	for {
		select {
		case m, ok := <-cc.p2pMsgStream:
			{
				if !ok {
					return
				}

				endPoint := cc.EndPoint(m.DstEndPointId[0])
				if endPoint == nil {
					// 如果为空说明还没有建立监听
					logger.Errorf("%s is not existed", m.DstEndPointId)
					continue
				}

				cc.sendMsgToEndPoint(endPoint, m)
			}
		case m, ok := <-cc.broadCastMsgStream:
			{
				if !ok {
					return
				}

				endPoints := cc.endPointList()
				for {
					endPoint, ok := <-endPoints
					if !ok {
						break
					}

					cc.sendMsgToEndPoint(endPoint, m)
				}

			}
		case <-cc.ctx.Done():
			return
		}
	}

	return
}

func (cc *channelContextImpl) SendMessage(msg *message.UnitMessage) error {
	var send func(*message.UnitMessage) error

	switch msg.Type {
	case message.UnitMessage_BroadCast:
		{
			send = func(broadcastMessage *message.UnitMessage) (broadCastErr error) {
				defer func() {
					res := recover()
					if err, ok := res.(error); ok {
						broadCastErr = err
					} else if _, ok = res.(string); ok {
						broadCastErr = errors.New(res.(string))
					}
				}()

				// 设置5秒的超时时间
				ctx, _ := context.WithTimeout(cc.ctx, 5*time.Second)
				select {
				case cc.broadCastMsgStream <- broadcastMessage:
				case <-ctx.Done():
					return errors.New("the send time out")
				}

				return nil
			}
		}
	case message.UnitMessage_PointToPoint:
		{
			send = func(p2pMessage *message.UnitMessage) (p2pErr error) {
				defer func() {
					res := recover()
					if err, ok := res.(error); ok {
						p2pErr = err
					} else if _, ok = res.(string); ok {
						p2pErr = errors.New(res.(string))
					}
				}()

				// 设置5秒的超时时间
				ctx, _ := context.WithTimeout(cc.ctx, 5*time.Second)
				select {
				case cc.p2pMsgStream <- p2pMessage:
				case <-ctx.Done():
					return errors.New("the send time out")
				}

				return
			}
		}
	default:
		return errors.New("send type is not supported")
	}

	return send(msg)
}
