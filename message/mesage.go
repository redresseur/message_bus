package message

import (
	"errors"
	"github.com/redresseur/message_bus/open_interface"
)

type messageHandlerImpl struct {
	channelHandler open_interface.ChannelHandler
}

func NewMessageHandler( ch open_interface.ChannelHandler ) open_interface.MessageHandler  {
	return &messageHandlerImpl{
		channelHandler: ch,
	}
}

func (mh *messageHandlerImpl) SendMessageToEndpoint(payload interface{}, metadata []byte, channelId, dstEndPointId string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return open_interface.ErrChannelNotExisted
	}

	msg := &open_interface.Message{}
	msg.ChannelID = channelId
	msg.DstEndPointId = dstEndPointId
	msg.Payload = payload
	msg.Metadata = metadata

	return cc.SendMessage(msg)
}

func (mh *messageHandlerImpl) SendMessageToAll(payload interface{}, metadata []byte, channelId string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return open_interface.ErrChannelNotExisted
	}

	return cc.SendMessage(&open_interface.Message{
		Payload: payload,
		Metadata: metadata,
		ChannelID: channelId,
	})
}

func (mh *messageHandlerImpl) SendMessageToEndpoints(payload interface{}, metadata []byte, channelId string, points ... string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return open_interface.ErrChannelNotExisted
	}

	for _, point := range points{
		err := cc.SendMessage(&open_interface.Message{
			DstEndPointId: point,
			ChannelID: channelId,
			Payload: payload,
			Metadata: metadata,
		})

		if err != nil{
			return err
		}
	}

	return nil
}

func (mh *messageHandlerImpl) SendMessageToGroup(payload interface{}, metadata []byte, group string) error {
	// TODO: 待實現
	return errors.New("implement me")
}



