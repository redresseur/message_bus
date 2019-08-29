package context

import (
	"context"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"strconv"
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

	syncDuration time.Duration

	channelId string

	// 确认已经收到的消息序列号
	stable uint32
}

func WithSyncConfig(autoSync bool, syncDuration time.Duration) func(client *ChannelContextClient) {
	return func(client *ChannelContextClient) {
		client.autoSyncOpen = autoSync
		client.syncDuration = syncDuration
	}
}

// 输入自定义的重启函数和最多尝试次数
// 注意：最大重启次数是只一次断连的过程中
// 最多能尝试次数
func WithReopenFunc(reopen Reopen, reopenMax uint8) func(client *ChannelContextClient) {
	return func(client *ChannelContextClient) {
		client.reConnFunc = reopen
		client.reopenMax = reopenMax
	}
}

func NewChannelContextClient(ctx context.Context, channelId string, ops ...func(client *ChannelContextClient)) *ChannelContextClient {
	res := &ChannelContextClient{
		catchSet:     map[message.UnitMessage_MessageFlag]CatchMsgFunc{},
		reopenMax:    ReopenMax,
		autoSyncOpen: false,
		channelId:    channelId,
	}

	// 初始化的序列号为1
	res.endPoint.Sequence = 1

	ctx, cancel := context.WithCancel(ctx)
	ctx = context.WithValue(ctx, res, cancel)
	res.ctx = ctx

	for _, op := range ops {
		op(res)
	}

	return res
}

func (cc *ChannelContextClient) SetCatchMsgFunc(flag message.UnitMessage_MessageFlag, msgFunc CatchMsgFunc) {
	cc.catchSet[flag] = msgFunc
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

func (cc *ChannelContextClient) sendSync() {
	cc.SendMsg(&message.UnitMessage{
		ChannelId:     cc.channelId,
		SrcEndPointId: cc.endPoint.Id,
		Flag:          message.UnitMessage_SYNC,
		Payload:       []byte(strconv.FormatUint(uint64(cc.stable), 10)),
	})
}

func (cc *ChannelContextClient) sync() {
	// 启动时先同步一次
	cc.sendSync()

	// 每隔2秒同步一次
	ticker := time.NewTicker(cc.syncDuration)
	for {
		// 判断上下文是否有效
		select {
		case <-cc.ctx.Done():
			return
		case <-ticker.C:
			{
				cc.sendSync()
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

		if msg.Ack > atomic.LoadUint32(&cc.endPoint.Sequence) {
			// 异常消息，忽略
			continue
		}

		// 检查序列号
		if msg.Seq <= atomic.LoadUint32(&cc.stable) {
			// 消息已经处理过，忽略
			continue
		} else if msg.Seq <= atomic.LoadUint32(&cc.endPoint.Ack) {
			// 重傳的消息，此時直接把確實收到的消息序列號變成收到的序列號
			atomic.StoreUint32(&cc.stable, msg.Seq)
			// 重傳的實時數據直接丟棄掉
			if catch, ok := cc.catchSet[msg.Flag]; ok && msg.Flag != message.UnitMessage_REAL_TIME {
				if err = catch(msg); err != nil {
					logger.Warningf("Catch Message %v", err)
				}
			}
		} else if msg.Seq > atomic.LoadUint32(&cc.endPoint.Ack) {
			atomic.StoreUint32(&cc.endPoint.Ack, msg.Seq)

			// 实时数据直接执行
			ok := atomic.CompareAndSwapUint32(&cc.stable, msg.Seq-1, msg.Seq)
			if msg.Flag != message.UnitMessage_REAL_TIME && !ok {
				cc.sendSync()
			} else {
				if catch, ok := cc.catchSet[msg.Flag]; ok {
					if err = catch(msg); err != nil {
						logger.Warningf("Catch Message %v", err)
					}
				}
			}
		}
	}
}
