package open_interface

import "github.com/redresseur/message_bus/proto/message"

type MessageHandler interface {
	SendMessage(message *message.UnitMessage) error
}
