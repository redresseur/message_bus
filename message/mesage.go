package message

import (
	"github.com/redresseur/message_bus/open_interface"
	"github.com/redresseur/message_bus/proto/message"
)

type messageHandlerImpl struct {
	channelHandler open_interface.ChannelHandler
}

func (mh *messageHandlerImpl) SendMessage(msg *message.UnitMessage) error {
	cc := mh.channelHandler.Channel(msg.ChannelId)
	if cc == nil {
		return open_interface.ErrChannelNotExisted
	}

	return cc.SendMessage(msg)
}

func NewMessageHandler(ch open_interface.ChannelHandler) open_interface.MessageHandler {
	return &messageHandlerImpl{
		channelHandler: ch,
	}
}
