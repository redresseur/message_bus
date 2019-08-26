package context

import (
	"context"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"github.com/redresseur/message_bus/queue"
	"github.com/redresseur/message_bus/storage"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"testing"
	"time"
)

var (
	ccctx *ChannelContextClient
	once  sync.Once
)

func clientTestInitial() {
	ccctx = &ChannelContextClient{
		ctx:          context.Background(),
		catchSet:     map[message.UnitMessage_MessageFlag]CatchMsgFunc{},
		reopenMax:    ReopenMax,
		autoSyncOpen: true,
		channelId:    "test",
	}

	ccctx.endPoint.Sequence = 1
}

func TestChannelContextClient_Bind(t *testing.T) {
	once.Do(clientTestInitial)

	client, srv := newChanEndpointIO()

	// 绑定服务端
	cctx.AddEndPoint(&open_interface.EndPoint{
		Id:          "test_point",
		CacheEnable: true,
		Cache:       queue.NewQueueRW(false),
		Sequence:    1,
	})

	cctx.BindRW("test_point", srv)

	// 绑定客户端
	assert.NoError(t, ccctx.Bind(client))
}

func TestChannelContextImpl_SendMessage(t *testing.T) {
	once.Do(clientTestInitial)

	cctx = WithChannelContext(context.Background(), &open_interface.ChannelInfo{
		ChannelId: "test",
	}, &storage.StorageImpl{},
		WithCatchMsgFunc(message.UnitMessage_SYNC, func(unitMessage *message.UnitMessage) error {
			log.Printf("recv sync message: %v", unitMessage)
			return nil
		}),
		WithCatchMsgFunc(message.UnitMessage_COMMON, func(unitMessage *message.UnitMessage) error {
			cctx.SendMessage(unitMessage)
			return nil
		}))

	TestChannelContextClient_Bind(t)

	ccctx.catchSet[message.UnitMessage_COMMON] = func(unitMessage *message.UnitMessage) error {
		log.Printf("recv comman message: %v", unitMessage)
		return nil
	}

	for i := 1; i < 2048; i++ {
		assert.NoError(t, ccctx.SendMsg(&message.UnitMessage{
			ChannelId: "test",
			Flag:      message.UnitMessage_COMMON,
			Type:      message.UnitMessage_BroadCast,
			Payload:   []byte("hello world!"),
		}))

		time.Sleep(5 * time.Millisecond)
	}

	// 查看缓存
	ccImpl := cctx.(*channelContextImpl)
	res, _ := ccImpl.endpoints.Get("test_point")
	endPoint := res.(*open_interface.EndPoint)

	time.Sleep(6 * time.Second)
	cache, err := endPoint.Cache.Seek(0, 10000)
	assert.NoError(t, err)

	assert.EqualValues(t, 0, len(cache))
}

func TestChannelContextClient_Reopen(t *testing.T) {
	once.Do(clientTestInitial)

	cctx = WithChannelContext(context.Background(), &open_interface.ChannelInfo{
		ChannelId: "test",
	}, &storage.StorageImpl{},
		WithCatchMsgFunc(message.UnitMessage_SYNC, func(unitMessage *message.UnitMessage) error {
			log.Printf("recv sync message: %v", unitMessage)
			return nil
		}),
		WithCatchMsgFunc(message.UnitMessage_COMMON, func(unitMessage *message.UnitMessage) error {
			cctx.SendMessage(unitMessage)
			return nil
		}))

	TestChannelContextClient_Bind(t)

	var reopen Reopen = func() chan open_interface.EndPointIO {
		res := make(chan open_interface.EndPointIO)
		go func() {
			time.Sleep(1 * time.Second)
			client, srv := newChanEndpointIO()
			cctx.BindRW("test_point", srv)
			res <- client
		}()
		return res
	}

	ccctx.reConnFunc = reopen
	ccctx.catchSet[message.UnitMessage_COMMON] = func(unitMessage *message.UnitMessage) error {
		log.Printf("recv comman message: %v", unitMessage)
		return nil
	}

	go func() {
		time.Sleep(1 * time.Second)
		close(ccctx.endPoint.RW.(*chanEndpointIO).recvCh)
		close(ccctx.endPoint.RW.(*chanEndpointIO).sendCh)
	}()

	for i := 1; i < 2048; i++ {
		assert.NoError(t, ccctx.SendMsg(&message.UnitMessage{
			ChannelId: "test",
			Flag:      message.UnitMessage_COMMON,
			Type:      message.UnitMessage_BroadCast,
			Payload:   []byte("hello world!"),
		}))

		time.Sleep(5 * time.Millisecond)
	}

	time.Sleep(6 * time.Second)
	ccImpl := cctx.(*channelContextImpl)
	res, _ := ccImpl.endpoints.Get("test_point")
	endPoint := res.(*open_interface.EndPoint)
	cache, err := endPoint.Cache.Seek(0, 10000)
	assert.NoError(t, err)

	assert.EqualValues(t, 0, len(cache))
}
