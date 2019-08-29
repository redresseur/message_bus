package context

import (
	"context"
	"errors"
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
	"github.com/redresseur/message_bus/queue"
	"github.com/redresseur/message_bus/storage"
	"testing"
)

var (
	cctx = WithChannelContext(context.Background(), &open_interface.ChannelInfo{
		ChannelId: "test",
	}, &storage.StorageImpl{})
)

type chanEndpointIO struct {
	sendCh chan *message.UnitMessage
	recvCh chan *message.UnitMessage
}

func newChanEndpointIO() (in, out *chanEndpointIO) {
	in = &chanEndpointIO{
		sendCh: make(chan *message.UnitMessage, 1024),
		recvCh: make(chan *message.UnitMessage, 1024),
	}

	out = &chanEndpointIO{
		sendCh: in.recvCh,
		recvCh: in.sendCh,
	}

	return
}

func (ci *chanEndpointIO) Write(msg *message.UnitMessage) (wErr error) {
	defer func() {
		r := recover()
		if err, ok := r.(error); ok {
			wErr = err
		}
	}()

	ci.sendCh <- msg
	return wErr
}

func (ci *chanEndpointIO) Read() (*message.UnitMessage, error) {
	msg, ok := <-ci.recvCh
	if !ok {
		return nil, errors.New("the recv chan closed")
	}
	return msg, nil
}

func TestMessageSync(t *testing.T) {
	cctx.AddEndPoint(&open_interface.EndPoint{
		Id:          "test_point",
		CacheEnable: true,
		Cache:       queue.NewQueueRW(false),
	})

	client, srv := newChanEndpointIO()
	cctx.BindRW("test_point", srv)

	for i := 0; i < 100; i++ {
		cctx.SendMessage(&message.UnitMessage{
			ChannelId:     "test",
			SrcEndPointId: "sender",
			DstEndPointId: []string{"test_point"},
			Flag:          message.UnitMessage_COMMON,
			Type:          message.UnitMessage_PointToPoint,
		})
	}

	client.Write(&message.UnitMessage{
		Ack:           0,
		Seq:           0,
		ChannelId:     "test",
		SrcEndPointId: "test_point",
		Flag:          message.UnitMessage_SYNC,
	})

	for i := 0; i < 100; i++ {
		msg, err := client.Read()
		if err != nil {
			t.Fatalf("read msg: %v", err)
		}

		t.Logf("msg: %v", *msg)
	}
}
