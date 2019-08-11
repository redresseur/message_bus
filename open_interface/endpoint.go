package open_interface

import (
	"context"
	"github.com/redresseur/message_bus/proto/message"
	"sync"
)

type EndPointStatus string

const (
	Online  EndPointStatus = `online`
	Offline EndPointStatus = `offline`
	Ready   EndPointStatus = `ready`
)

type EndPointIO interface {
	Write(message *message.UnitMessage) error
	Read() (*message.UnitMessage, error)
}

type EmptyEndPointIO struct {
}

func (e *EmptyEndPointIO) Write(message *message.UnitMessage) error {
	return nil
}

func (e *EmptyEndPointIO) Read() (*message.UnitMessage, error) {
	return nil, nil
}

// 节点信息和设置
type EndPoint struct {
	// 节点的ID, 唯一识别符
	Id string `json:"id"`

	// 当前消息序列
	Sequence uint32 `json:"sequence"`

	// 当前应答序列
	Ack uint32 `json:"ack"`

	// 是否为channel的创建者
	IsCreator bool `json:"isCreator"`

	// 當前節點的狀態
	Status EndPointStatus `json:"status"`

	Ctx context.Context `json:"_"`

	RW EndPointIO `json:"_"`

	// 是否开启消息缓存
	CacheEnable bool `json:"cacheEnable"`

	Cache QueueRW `json:"_"`

	l sync.Mutex `json:"_"`
}

func (e *EndPoint)Locker() sync.Locker {
	return &e.l
}

type EndPointGroup struct {
	// 節點組的id
	Id string

	// 節點隊列
	points []*EndPoint
}
