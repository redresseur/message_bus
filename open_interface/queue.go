package open_interface

type QueueSignal uint8

type QueueReader interface {
	// 查询从beginIndex 位置长度为 offset 的数据
	Seek(beginIndex uint64, offset int8)([]interface{}, error)

	// 消息队列更新的信号
	Single()(<- chan QueueSignal)

	// 队列的起始指针向后偏移1 , false 不从队列中移除取到的数据，true则从队列中移除取到的数据
	Next(remove bool)(interface{}, error)

	// 从起始到结束的位置数据清除掉
	// 如果beginIndex < 0 || endIndex < 0，则抛出错误
	// 如果beginIndex - endIndex > 0，则抛出错误
	Remove(beginIndex, endIndex int32)(error)

	// 清除所有数据
	Reset()(error)
}

type QueueWriter interface {
	// 成功则返回消息的队列地址
	Push(interface{})(uint64, error)
}