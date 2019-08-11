package open_interface

import (
	"context"
	"github.com/redresseur/flogging"
	"github.com/redresseur/message_bus/proto/message"
)

var logger = flogging.MustGetLogger("open.interface")

// channel 的信息
type ChannelInfo struct {
	// channel 的唯一识别符
	ChannelId string

	// channel 的加入者上限容量
	Max uint64

	// 當前channel内的節點個數
	EndPointNum uint64
}

type ChannelContext interface {
	AddEndPoint(point *EndPoint) error

	EndPoint(endPointId string) *EndPoint

	RemoveEndPoint(pointId string)

	BindRW(pointId string, rw EndPointIO) (context.Context, error)

	SendMessage(msg *message.UnitMessage) error

	Cancel()
}

type ChannelHandler interface {
	// 创建一个channel
	CreateChannel(info *ChannelInfo, point *EndPoint) (ChannelContext, error)

	// 加入一个channel
	JoinChannel(channelId string, point *EndPoint) error

	// 关闭一个channel
	CloseChannel(channelId string, point *EndPoint) error

	// 监听一个channel
	ListenChannel(channelId string, point *EndPoint) (context.Context, error)

	// 开启一个子channel
	ChildrenChannel(parentContext *ChannelContext, childInfo *ChannelInfo, point *EndPoint) (ChannelContext, error)

	// 退出一个channel
	ExitChannel(channelId string, point *EndPoint) error

	// 获取channel context
	Channel(channelId string) ChannelContext
}
