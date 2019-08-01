package queue

import (
	"errors"
	"github.com/redresseur/message_bus/open_interface"
	"sync"
)

type QueueImpl struct {
	cacheLock sync.Mutex
	cache map[uint64]interface{}

	// 当前索引地址
	index  uint64

	// 最新消息的地址
	// 初始值要为 -1
	latest uint64

	signal chan open_interface.QueueSignal
}

func (qi *QueueImpl) Push(data interface{}) (uint64, error) {
	var(
		index uint64
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	qi.latest++
	index = qi.latest
	qi.cache[qi.latest] = data

	go func() {
		qi.signal <- 0
	}()

	return index, nil
}

func (qi *QueueImpl) Seek(beginIndex uint64, offset int8) ([]interface{}, error) {
	var (
		res = []interface{}{}
		err error
		endIndex uint64
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if qi.latest < beginIndex {
		err = errors.New("queue seek over range")
	}else {

		if offset < 0 || (beginIndex + uint64(offset)) > qi.latest{
			endIndex = qi.latest
		}

		var i = beginIndex
		for ; i <= endIndex; i++{
			if v, ok := qi.cache[i]; ok{
				res = append(res, v)
			}
		}

		// 移动索引地址
		qi.index = i
	}

	return res, err
}

func (qi *QueueImpl) Single() (<-chan open_interface.QueueSignal) {
	return qi.signal
}

func (qi *QueueImpl) Next(remove bool) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if v, ok := qi.cache[qi.index]; ok {
		res = v
		if remove{
			qi.cache[qi.index] = nil
		}

		qi.index++
	}else {
		err = errors.New("unexcepted EOF")
	}

	return res, err
}

func (qi *QueueImpl) Remove(beginIndex, endIndex int32) (error) {
	var(
		start uint64 = 0
		end uint64 = 0
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if beginIndex < 0 && endIndex < 0{
		return errors.New("the range is illegal")
	}

	if uint64(beginIndex) > qi.latest || uint64(endIndex) > qi.latest {
		return errors.New("the range is illegal")
	}

	if beginIndex > 0 {
		start = uint64(beginIndex)
		if endIndex < 0{
			end = qi.latest
		}else if endIndex - beginIndex < 0{
			return errors.New("the range is illegal")
		} else {
			end = uint64(endIndex)
		}
	}else {
		end = uint64(endIndex)
	}

	for i := start; i <= end; i++{
		delete(qi.cache, i)
	}

	return nil
}

func (qi *QueueImpl) Reset() (error) {
	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	qi.index = 0
	qi.latest = 0

	close(qi.signal)

	qi.cache = map[uint64]interface{}{}
	qi.signal = make(chan open_interface.QueueSignal, 1024)

	return nil
}
