package definitions

import (
	"context"
)

type EndPointStatus string

const (
	Online EndPointStatus = `online`
	Offline EndPointStatus = `offline`
	Ready EndPointStatus = `ready`
)

type EndPointIO interface {
	Write(p interface{}) (error)
	Read() (interface{}, error)
}

// 节点信息和设置
type EndPoint struct {
	// 节点的ID, 唯一识别符
	Id string `json:"id"`

	// 是否开启消息缓存
	CacheEnable bool `json:"cacheEnable"`

	// 当前消息序列
	Sequence uint64 `json:"sequence"`

	// 当前应答序列
	Ack uint64 `json:"ack"`

	// 是否为channel的创建者
	IsCreator bool `json:"isCreator"`

	// 當前節點的狀態
	Status EndPointStatus `json:"status"`

	Ctx context.Context `json:"ctx"`

	RW EndPointIO `json:"rw"`
}

type EndPointGroup struct {
	// 節點組的id
	Id string

	// 節點隊列
	points []*EndPoint

}