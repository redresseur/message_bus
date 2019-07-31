package message

import (
	"errors"
	"github.com/redresseur/message_bus/definitions"
)

type messageHandlerImpl struct {
	channelHandler definitions.ChannelHandler
}

func NewMessageHandler( ch definitions.ChannelHandler ) definitions.MessageHandler  {
	return &messageHandlerImpl{
		channelHandler: ch,
	}
}

func (mh *messageHandlerImpl) SendMessageToEndpoint(data, metadata []byte, channelId, dstEndPointId string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return definitions.ErrChannelNotExisted
	}

	msg := &definitions.Message{}
	msg.ChannelID = channelId
	msg.DstEndPointId = dstEndPointId
	msg.Payload = data
	msg.Metadata = data

	return cc.SendMessage(msg)
}

func (mh *messageHandlerImpl) SendMessageToAll(data, metadata []byte, channelId string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return definitions.ErrChannelNotExisted
	}

	return cc.SendMessage(&definitions.Message{
		Payload: data,
		Metadata: metadata,
		ChannelID: channelId,
	})
}

func (mh *messageHandlerImpl) SendMessageToEndpoints(data, metadata []byte, channelId string, points ... string) error {
	cc := mh.channelHandler.Channel(channelId)
	if cc != nil{
		return definitions.ErrChannelNotExisted
	}

	for _, point := range points{
		err := cc.SendMessage(&definitions.Message{
			DstEndPointId: point,
			ChannelID: channelId,
			Payload: data,
			Metadata: metadata,
		})

		if err != nil{
			return err
		}
	}

	return nil
}

func (mh *messageHandlerImpl) SendMessageToGroup(data, metadata []byte, group string) error {
	// TODO: 待實現
	return errors.New("implement me")
}



