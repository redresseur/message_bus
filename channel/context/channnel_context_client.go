package context

import (
	"context"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"sync/atomic"
	"time"
)

const ReopenMax = 5

type CatchMsgFunc func(unitMessage *message.UnitMessage) error

// 重连函数
// 如果重连失败则直接关闭 chan
// 如果重连成功则返回结果后，再关闭chan
type Reopen func() chan open_interface.EndPointIO

type ChannelContextClient struct {
	ctx context.Context

	// 客户端信息
	endPoint open_interface.EndPoint

	catchSet map[message.UnitMessage_MessageFlag]CatchMsgFunc

	reConnFunc Reopen

	// 默认重连次数为5次
	reopenMax uint8

	// 自动同步消息
	autoSyncOpen bool

	channelId string
}

func NewChannelContextClient(ctx context.Context, channelId string) *ChannelContextClient {
	res := &ChannelContextClient{
		catchSet:     map[message.UnitMessage_MessageFlag]CatchMsgFunc{},
		reopenMax:    ReopenMax,
		autoSyncOpen: true,
		channelId:    channelId,
	}

	// 初始化的序列号为1
	res.endPoint.Sequence = 1

	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, res, cancel)

	res.ctx = ctx
	return res
}

// 绑定读写io
// 只能调用一次
func (cc *ChannelContextClient) Bind(rw open_interface.EndPointIO) error {
	cc.endPoint.RW = rw
	go cc.worker()

	if cc.autoSyncOpen {
		go cc.sync()
	}

	return nil
}

func (cc *ChannelContextClient) SendMsg(unitMessage *message.UnitMessage) error {
	unitMessage.Seq = atomic.LoadUint32(&cc.endPoint.Sequence)
	unitMessage.Ack = atomic.LoadUint32(&cc.endPoint.Ack)
	unitMessage.SrcEndPointId = cc.endPoint.Id
	if err := cc.endPoint.RW.Write(unitMessage); err != nil {
		return err
	}

	atomic.AddUint32(&cc.endPoint.Sequence, 1)
	return nil
}

func (cc *ChannelContextClient) Close() {
	cc.ctx.Value(cc).(context.CancelFunc)()
}

func (cc *ChannelContextClient) sync() {
	// 每隔5秒同步一次
	ticker := time.NewTicker(5 * time.Second)
	for {
		// 判断上下文是否有效
		select {
		case <-cc.ctx.Done():
			return
		case <-ticker.C:
			{
				cc.SendMsg(&message.UnitMessage{
					ChannelId:     cc.channelId,
					SrcEndPointId: cc.endPoint.Id,
					Flag:          message.UnitMessage_SYNC,
				})
			}
		}
	}
}

// 当断开时，触发重连机制
func (cc *ChannelContextClient) reopen() {
	if cc.reConnFunc == nil {
		return
	}

	for i := uint8(0); i < cc.reopenMax; i++ {
		select {
		case <-cc.ctx.Done():
			return
		case rw, ok := <-cc.reConnFunc():
			{
				if !ok {
					continue
				}

				logger.Infof("Reopen Successfully")
				// 重连成功后，恢复工作协程
				cc.endPoint.RW = rw
				go cc.worker()
				return
			}
		}
	}
}

// 接受消息，以协程的方式启动
func (cc *ChannelContextClient) worker() {
	for {
		msg, err := cc.endPoint.RW.Read()
		if err != nil {
			logger.Errorf("Client read message: %v", err)
			// 判断当前上下文是否有效
			if err := cc.ctx.Err(); err == nil {
				cc.reopen()
			}
			break
		}
		// 检查序列号
		if msg.Seq <= atomic.LoadUint32(&cc.endPoint.Ack) {
			// 消息已经处理过，忽略
			continue
		}

		if msg.Ack > atomic.LoadUint32(&cc.endPoint.Sequence) {
			// 异常消息，忽略
			continue
		}

		atomic.StoreUint32(&cc.endPoint.Ack, msg.Seq)

		if catch, ok := cc.catchSet[msg.Flag]; ok {
			if err = catch(msg); err != nil {
				logger.Warningf("Catch Message %v", err)
			}
		}
	}
}
