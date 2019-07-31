package definitions

import (
	"context"
	"errors"
	"github.com/redresseur/flogging"
	"sync"
	"sync/atomic"
	"time"
)

var logger = flogging.MustGetLogger("def.channel")

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
	broadCastMsgStream chan *Message

	// 點對點消息stream
	p2pMsgStream chan *Message

	// 當前節點隊列
	endpoints sync.Map
}

func WithChannelContext(ctx context.Context, info *ChannelInfo) *ChannelContext{
	cm := &ChannelContext{
		info: info,
		broadCastMsgStream: make(chan *Message, 1024),
		p2pMsgStream: make(chan *Message, 1024),
	}

	newCtx, cancel := context.WithCancel(ctx)
	cm.ctx = context.WithValue(newCtx, cm, cancel)

	go cm.messageProcess()
	return cm
}

func (cc *ChannelContext)Cancel(){
	cc.ctx.Value(cc).(context.CancelFunc)()
}

func (cc *ChannelContext)EndPoint(endPointId string)*EndPoint{
	cu, ok := cc.endpoints.Load(endPointId)
	if !ok {
		return nil
	}

	return cu.(*EndPoint)
}

func (cc *ChannelContext)AddEndPoint(point *EndPoint) error {
	if cc.EndPoint(point.Id) != nil{
		return ErrEndPointHasBeenExisted
	}
	// 分配ctx给customer
	ctx, cancel := context.WithCancel(cc.ctx)
	ctx = context.WithValue(ctx, point, cancel)

	point.Ctx = ctx

	cc.endpoints.Store(point.Id, point)

	return nil
}

func (cc *ChannelContext)RemoveEndPoint(pointId string) {
	if ep ,ok := cc.endpoints.Load(pointId); ok {
		ep.(*EndPoint).Ctx.Value(ep).(context.CancelFunc)()
		cc.endpoints.Delete(pointId)
	}
}

func (cc *ChannelContext)BindRW(pointId string, rw EndPointIO) (context.Context, error) {
	ep := cc.EndPoint(pointId)
	if ep == nil{
		return nil, ErrEndPointNotExisted
	}

	return ep.Ctx, nil
}

func (cc *ChannelContext)endPointList() <-chan *EndPoint {
	// 要用同步指针
	output := make(chan *EndPoint, 1024)
	go func() {
		defer close(output)
		cc.endpoints.Range(func(key, v interface{})bool{
			select {
			case output <- v.(*EndPoint):
			case <-time.After(100*time.Millisecond): // 100 毫秒的等待延时
				return false
			}

			return true
		})
	}()

	return output
}

func (cc *ChannelContext)messageProcess(){
	for {
		select {
		case m, ok := <-cc.p2pMsgStream:
			{
				if !ok{
					return
				}

				endPoint := cc.EndPoint(m.DstEndPointId)
				if endPoint == nil{
					// 如果为空说明还没有建立监听
					logger.Errorf("%s is not existed", m.DstEndPointId)
					continue
				}

				if err := endPoint.RW.Write(m.Payload); err != nil{
					logger.Errorf("Write Message Failure: %v", err)
				}

				// 增加计数
				atomic.AddUint64(&endPoint.Sequence, 1)
			}
		case m, ok := <- cc.broadCastMsgStream:
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

					if err := endPoint.RW.Write(m.Payload); err != nil{
						logger.Errorf("Write Message Failure: %v", err)
					}

					// 增加计数
					atomic.AddUint64(&endPoint.Sequence, 1)
				}

			}
		case <-cc.ctx.Done():
			return
		}
	}

	return
}


func (cc *ChannelContext)SendMessage(message *Message) error {
	var send func(*Message)error
	if message.DstEndPointId != ""{
		send = func (p2pMessage *Message)(p2pErr error){
			defer func() {
				res := recover()
				if err, ok := res.(error); ok {
					p2pErr = err
				}else if _, ok = res.(string); ok{
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
	}else {
		send = func (broadcastMessage *Message)(broadCastErr error){
			defer func() {
				res := recover()
				if err, ok := res.(error); ok {
					broadCastErr = err
				}else if _, ok = res.(string); ok{
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

			return
		}
	}

	return send(message)
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
	ChildrenChannel(parentContext *ChannelContext, childInfo *ChannelInfo, point *EndPoint)(*ChannelContext, error)

	// 退出一个channel
	ExitChannel(channelId string, point *EndPoint)error

	// 获取channel context
	Channel(channelId string) *ChannelContext
}