package definitions

type Message struct {
	// Channel ID
	ChannelID string `json:"channelId"`

	// 目標節點
	DstEndPointId string `json:"dstEndPointId"`

	// 消息負載
	Payload interface{} `json:"payload"`

	// 元數據 暫時沒有用到
	Metadata []byte `json:"metadata"`
}

type MessageHandler interface {
	// 点对点发送
	SendMessageToEndpoint(payload interface{}, metadata []byte, channelId, dstEndPointId string)error

	// 广播
	SendMessageToAll(payload interface{}, metadata []byte, channelId string)error

	// 多播
	SendMessageToEndpoints(payload interface{}, metadata []byte, channelId string, points... string)error

	// 组播
	SendMessageToGroup(data interface{}, metadata []byte, group string)error
}
