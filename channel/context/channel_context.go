package context

import (
	"context"
	"errors"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"sync/atomic"
	"time"
)

var logger = flogging.MustGetLogger("channel.context")

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
}

func WithChannelContext(ctx context.Context, info *open_interface.ChannelInfo,
	storage open_interface.Storage) open_interface.ChannelContext {
	cm := &channelContextImpl{
		info:               info,
		broadCastMsgStream: make(chan *message.UnitMessage, 1024),
		p2pMsgStream:       make(chan *message.UnitMessage, 1024),
		endpoints:          storage,
	}

	newCtx, cancel := context.WithCancel(ctx)
	cm.ctx = context.WithValue(newCtx, cm, cancel)

	go cm.messageProcess()
	return cm
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
		return nil
	}

	// 分配ctx给customer
	ctx, cancel := context.WithCancel(cc.ctx)
	ctx = context.WithValue(ctx, point, cancel)

	point.Ctx = ctx

	// 判断是否开启缓存
	if point.CacheEnable && point.Cache == nil {
		return errors.New("the cache is nil")
	}

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
		if err != nil {
			logger.Errorf("Recv message from %s: %v", point.Id, err)
			break
		}

		// 检查ack 和 seq
		// 如果msg.ack 大于 point.seq， 说明对方串包了
		if msg.Ack > atomic.LoadUint32(&point.Sequence) {
			logger.Warnf("Ack %d greater than Seq %d", msg.Ack, point.Sequence)
			continue
		}

		// msg.seq 小于 point.ack, 说明对方重复发包了
		if msg.Seq < atomic.LoadUint32(&point.Ack) {
			continue
		}

		// ack 累计增长
		atomic.AddUint32(&point.Ack, 1)

		switch msg.Flag {
		// 心跳包
		case message.UnitMessage_HEART_BEAT:
			// TODO: 更新心跳状态
		case message.UnitMessage_REAL_TIME:
			// TODO: 发送消息
		case message.UnitMessage_SYNC:
			{
				if msg.Ack == point.Sequence {
					break
				}

				offset := uint32(point.Sequence - msg.Ack - 1)
				beginIndex := uint64(msg.Ack + 1)
				res, err := point.Cache.Seek(beginIndex, offset)
				if err != nil {
					logger.Warningf("Seek from %d offset %d: %v", msg.Ack, offset, err)
				}

				for _, v := range res {
					unitMsg := v.(*message.UnitMessage)
					cc.sendMsgToEndPoint(point, &message.UnitMessage{
						ChannelId:     unitMsg.ChannelId,
						SrcEndPointId: unitMsg.SrcEndPointId,
						DstEndPointId: unitMsg.DstEndPointId,
						Flag:          unitMsg.Flag,
						Type:          unitMsg.Type,
						Metadata:      unitMsg.Metadata,
						Payload:       unitMsg.Payload,
						// TODO: 更新时间戳
					})
				}

				// 移除消息，避免重复
				point.Cache.Remove(int32(beginIndex), int32(beginIndex+uint64(offset)))
			}
		case message.UnitMessage_COMMON:
			// TODO: 发送消息
		}
	}
}

func (cc *channelContextImpl) sendMsgToEndPoint(point *open_interface.EndPoint, msg *message.UnitMessage) {
	// 增加计数
	point.Locker().Lock()
	defer point.Locker().Unlock()

	msg.Seq = atomic.LoadUint32(&point.Sequence)
	msg.Ack = point.Ack
	if err := point.RW.Write(msg); err != nil {
		logger.Errorf("Write Message Failure: %v", err)
	}

	atomic.AddUint32(&point.Sequence, 1)

	// 添加到缓存中
	// 目前除了普通消息外都不缓存
	if point.CacheEnable {
		if msg.Flag == message.UnitMessage_COMMON {
			point.Cache.Push(msg)
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
