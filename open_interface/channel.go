package open_interface

import (
	"context"
	"errors"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/proto/message"
	"sync/atomic"
	"time"
)

var logger = flogging.MustGetLogger("open.interface")

// channel 的信息
type ChannelInfo struct {
	// channel 的唯一识别符
	ChannelId string

	// channel 的加入者上限容量
	Max uint64

	// 當前channel内的節點個數
	EndPointNum uint64
}

//	channel的上下文
type ChannelContext struct {
	info *ChannelInfo

	ctx context.Context

	// 廣播消息stream
	broadCastMsgStream chan *message.UnitMessage

	// 點對點消息stream
	p2pMsgStream chan *message.UnitMessage

	// 當前節點隊列
	endpoints Storage
}

func WithChannelContext(ctx context.Context, info *ChannelInfo, storage Storage) *ChannelContext {
	cm := &ChannelContext{
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

func (cc *ChannelContext) Cancel() {
	cc.ctx.Value(cc).(context.CancelFunc)()
}

func (cc *ChannelContext) EndPoint(endPointId string) *EndPoint {
	cu, err := cc.endpoints.Get(endPointId)
	if err != nil {
		return nil
	}

	return cu.(*EndPoint)
}

func (cc *ChannelContext) AddEndPoint(point *EndPoint) error {
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

func (cc *ChannelContext) RemoveEndPoint(pointId string) {
	if ep, err := cc.endpoints.Get(pointId); err == nil {
		ep.(*EndPoint).Ctx.Value(ep).(context.CancelFunc)()
		cc.endpoints.Delete(pointId)
	}
}

func (cc *ChannelContext) BindRW(pointId string, rw EndPointIO) (context.Context, error) {
	ep := cc.EndPoint(pointId)
	if ep == nil {
		return nil, ErrEndPointNotExisted
	}

	ep.RW = rw

	// 启动一个协程，用来接收消息
	go func() {

	}()
	return ep.Ctx, nil
}

func (cc *ChannelContext) endPointList() <-chan *EndPoint {
	// 要用同步指针
	output := make(chan *EndPoint, 1024)
	go func() {
		defer close(output)
		cc.endpoints.Range(func(key, v interface{}) bool {
			select {
			case output <- v.(*EndPoint):
			case <-time.After(100 * time.Millisecond): // 100 毫秒的等待延时
				return false
			}

			return true
		})
	}()

	return output
}

func (cc *ChannelContext) recvMsgFromEndPoint(point *EndPoint) {
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

		}
	}
}

func (cc *ChannelContext) sendMsgToEndPoint(point *EndPoint, msg *message.UnitMessage) {
	// 增加计数
	point.l.Lock()
	defer point.l.Unlock()

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

func (cc *ChannelContext) messageProcess() {
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

func (cc *ChannelContext) SendMessage(msg *message.UnitMessage) error {
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

type ChannelHandler interface {
	// 创建一个channel
	CreateChannel(info *ChannelInfo, point *EndPoint) (*ChannelContext, error)

	// 加入一个channel
	JoinChannel(channelId string, point *EndPoint) error

	// 关闭一个channel
	CloseChannel(channelId string, point *EndPoint) error

	// 监听一个channel
	ListenChannel(channelId string, point *EndPoint) (context.Context, error)

	// 开启一个子channel
	ChildrenChannel(parentContext *ChannelContext, childInfo *ChannelInfo, point *EndPoint) (*ChannelContext, error)

	// 退出一个channel
	ExitChannel(channelId string, point *EndPoint) error

	// 获取channel context
	Channel(channelId string) *ChannelContext
}
