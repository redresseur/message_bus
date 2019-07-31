package event

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/definitions"
	"sync"
)

var logger = flogging.MustGetLogger("event.impl")

type hooker struct {
	req *definitions.EventRequest
	c definitions.CatchFunc
}

type ChannelEvent struct {
	ChannelId string `json:"channelId"`
	Point definitions.EndPoint `json:"point"`
}

type channelEventHandlerImpl struct {
	hookers sync.Map
}

func NewChannelEventHandlerImpl() definitions.EventHandler {
	return &channelEventHandlerImpl{
		hookers: sync.Map{},
	}
}

func (eh *channelEventHandlerImpl) RegistryEvent(req *definitions.EventRequest,
	c definitions.CatchFunc)(string, error) {
	eventID := uuid.New().String()
	eh.hookers.Store(eventID, &hooker{
		c: c,
		req: req,
	})

	return eventID, nil
}

func (eh *channelEventHandlerImpl)UnRegistryEvent(eventID string) error {
	eh.hookers.Delete(eventID)
	return nil
}

func (eh *channelEventHandlerImpl)CatchEvent(e *definitions.Event) error {
	catchFuncs := []definitions.CatchFunc{}

	ce := &ChannelEvent{}
	if err := json.Unmarshal(e.Payload, ce); err != nil{
		return err
	}

	q := func(key, value interface{}) bool {
		h := value.(*hooker)

		cond :=  &ChannelEvent{}
		if err := json.Unmarshal(h.req.Payload, ce); err != nil{
			logger.Warnf("Unmarshal the event condition %v", err)
			return true
		}

		// 檢查當前的事件是否為其hooker 所期待的事件
		if cond.ChannelId == ce.ChannelId &&
			cond.Point.Id == ce.Point.Id{
			catchFuncs = append(catchFuncs, h.c)
		}
		return true
	}

	eh.hookers.Range(q)

	cc := func(c definitions.CatchFunc, e *definitions.Event)(err error ){
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
