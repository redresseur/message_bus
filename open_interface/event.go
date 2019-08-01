package open_interface

import "github.com/golang/protobuf/ptypes/timestamp"

type EventType string

const (
	EventCreateChannel EventType = "create_channel"
	EventCloseChannel EventType = "close_channel"
	EventJoinChannel EventType = "join_channel"
	EventListenChannel EventType = "listen_channel"
	EventExitChannel EventType = "exit_channel"
)

type Event struct {
	Type  	 EventType `json:"type"`
	Payload  []byte `json:"payload"`
	ErrMsg   string `json:"errMsg"`
	Timestamp  *timestamp.Timestamp `json:"timestamp"`
}

// 监听事件请求
type EventRequest struct {
	// 事件类型
	Type EventType `json:"type"`

	// 事件过滤器
	Filter interface{} `json:"payload"`
}

// 用来接收事件
type CatchFunc func (*Event) error

type EventHandler interface {
	// 注册一个事件处理的hooker
	RegistryEvent(req *EventRequest, c CatchFunc) (string, error)

	// 注销监听
	UnRegistryEvent(eventID string) error

	// 处理事件
	CatchEvent(event *Event) error

	// 提交事件
	DeliverEvent(event *Event)error
}

type EmptyEventHandler struct {

}

func (*EmptyEventHandler) DeliverEvent(event *Event) error {
	return nil
}

func (*EmptyEventHandler) RegistryEvent(req *EventRequest, c CatchFunc) (string, error) {
	return "", nil
}

func (*EmptyEventHandler) UnRegistryEvent(eventID string) error {
	return nil
}

func (*EmptyEventHandler) CatchEvent(event *Event) error {
	return nil
}

type Manager struct {
	ss map[EventType]EventHandler
}

func (m *Manager)RegistryService(eventType EventType, s EventHandler){
	m.ss[eventType] = s
}

func (m *Manager)Service( eventType EventType) EventHandler {
	s, ok := m.ss[eventType]
	if ok{
		return s
	}

	return &EmptyEventHandler{}
}

func NewManager()*Manager {
	return &Manager{
		ss: map[EventType]EventHandler{},
	}
}