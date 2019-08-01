package event

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/open_interface"
	"sync"
)

var logger = flogging.MustGetLogger("event.impl")

type hooker struct {
	req *open_interface.EventRequest
	c open_interface.CatchFunc
}

// 事件过滤器
type ChannelEventFilter open_interface.CatchFunc

type ChannelEventPayload struct {
	ChannelId string `json:"channelId"`
	Point *open_interface.EndPoint `json:"point"`
}

type channelEventHandlerImpl struct {
	hookers map[open_interface.EventType]sync.Map
	reader open_interface.QueueReader
	writer open_interface.QueueWriter
	ctx context.Context
}

func (eh *channelEventHandlerImpl) DeliverEvent(event *open_interface.Event) error {
	eh.writer.Push(event)
	return nil
}

func NewChannelEventHandlerImpl(ctx context.Context, reader open_interface.QueueReader,
	writer open_interface.QueueWriter) open_interface.EventHandler {
	eh := &channelEventHandlerImpl{
		hookers: map[open_interface.EventType]sync.Map{},
		reader: reader,
		writer: writer,
	}

	newCtx, cancel := context.WithCancel(ctx)
	eh.ctx = context.WithValue(newCtx, eh, cancel)

	go eh.worker()

	return eh
}

func (eh *channelEventHandlerImpl) RegistryEvent(req *open_interface.EventRequest,
	c open_interface.CatchFunc)(string, error) {
	eventID := uuid.New().String()
	if sets, ok := eh.hookers[req.Type]; ok {
		sets.Store(eventID, &hooker{
			c: c,
			req: req,
		})
	}else {
		sets = sync.Map{}
		sets.Store(eventID, &hooker{
			c: c,
			req: req,
		})
		eh.hookers[req.Type] = sets
	}


	return eventID, nil
}

func (eh *channelEventHandlerImpl)UnRegistryEvent(eventID string) error {
	for _, v := range eh.hookers{
		v.Delete(eventID)	
	}
	
	return nil
}

func (eh *channelEventHandlerImpl)CatchEvent(e *open_interface.Event) error {
	// 檢查類型組是否存在
	hookers, ok := eh.hookers[e.Type]
	if !ok{
		return nil
	}
	
	catchFuncs := []open_interface.CatchFunc{}

	payload := &ChannelEventPayload{}
	if err := json.Unmarshal(e.Payload, payload); err != nil{
		return err
	}

	q := func(key, value interface{}) bool {
		h := value.(*hooker)
		filter, ok := h.req.Filter.(ChannelEventFilter)
		if !ok{
			logger.Warnf("the event filter is invalid")
			return true
		}

		// 檢查當前的事件是否為其hooker 所期待的事件
		if filter(e) == nil{
			catchFuncs = append(catchFuncs, h.c)
		}
		return true
	}

	hookers.Range(q)

	cc := func(c open_interface.CatchFunc, e *open_interface.Event)(err error ){
		defer func() {
			if r := recover(); r != nil{
				if defErr, ok := r.(error); ok{
					err = defErr
				}
			}
		}()
		err = c(e)
		return
	}

	for _, c := range catchFuncs{
		if err := cc(c, e); err != nil{
			logger.Warningf("send service event: [%v]", err)
			continue
		}
	}

	return nil
}

func (eh *channelEventHandlerImpl) worker() {
	for {
		select {
		case _, ok := <-eh.reader.Single():
			if !ok {
				return
			}

			if e, err := eh.reader.Next(true); err != nil{
				logger.Errorf("Get Event %v", err)
				return
			}else {
			 	eh.CatchEvent(e.(*open_interface.Event))
			}
		case <-eh.ctx.Done():
			logger.Debug("the event worker over")
			return
		}
	}
}