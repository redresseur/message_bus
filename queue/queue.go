package queue

import (
	"errors"
	"github.com/redresseur/message_bus/open_interface"
	"sync"
)

type queueImpl struct {
	cacheLock sync.Mutex
	cache     map[uint64]interface{}

	// 当前索引地址
	index uint64

	// 最新消息的地址
	// 此地址对应的数据永远为空
	latest uint64

	signalEnable bool

	signal chan open_interface.QueueSignal
}

func NewQueueRW(signalEnable bool) open_interface.QueueRW {
	return &queueImpl{
		signalEnable: signalEnable,
		index:        0,
		latest:       0,
		signal:       make(chan open_interface.QueueSignal, 1024),
		cache:        map[uint64]interface{}{},
	}
}

func (qi *queueImpl) Push(data interface{}) (uint64, error) {
	var (
		index uint64
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	index = qi.latest
	qi.cache[qi.latest] = data
	qi.latest++

	if qi.signalEnable {
		go func() {
			qi.signal <- 0
		}()
	}

	return index, nil
}

func (qi *queueImpl) Seek(beginIndex uint64, offset uint32) ([]interface{}, error) {
	var (
		res      = []interface{}{}
		err      error
		endIndex uint64
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if qi.latest < beginIndex {
		err = errors.New("queue seek over range")
	} else {

		if offset < 0 || (beginIndex+uint64(offset)) > qi.latest {
			endIndex = qi.latest
		} else {
			endIndex = beginIndex + uint64(offset)
		}

		var i = beginIndex
		for ; i <= endIndex; i++ {
			if v, ok := qi.cache[i]; ok {
				res = append(res, v)
			}
			// 移动索引地址
			// 放在外面会多移动一个地址
			qi.index = i
		}
	}

	return res, err
}

func (qi *queueImpl) Single() <-chan open_interface.QueueSignal {
	return qi.signal
}

func (qi *queueImpl) Next(remove bool) (interface{}, error) {
	var (
		res interface{}
		err error
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if v, ok := qi.cache[qi.index]; ok {
		res = v
		if remove {
			delete(qi.cache, qi.index)
		}

		qi.index++
	} else {
		err = errors.New("unexcepted EOF")
	}

	return res, err
}

func (qi *queueImpl) Remove(beginIndex, endIndex int32) error {
	var (
		start uint64 = 0
		end   uint64 = 0
	)

	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	if beginIndex < 0 && endIndex < 0 {
		return errors.New("the range is illegal")
	}

	if uint64(beginIndex) > qi.latest || uint64(endIndex) > qi.latest {
		return errors.New("the range is illegal")
	}

	if beginIndex > 0 {
		start = uint64(beginIndex)
		if endIndex < 0 {
			end = qi.latest
		} else if endIndex-beginIndex < 0 {
			return errors.New("the range is illegal")
		} else {
			end = uint64(endIndex)
		}
	} else {
		start = qi.index
		end = uint64(endIndex)
	}

	for i := start; i <= end; i++ {
		delete(qi.cache, i)
	}

	return nil
}

func (qi *queueImpl) Reset() error {
	qi.cacheLock.Lock()
	defer qi.cacheLock.Unlock()

	qi.index = 0
	qi.latest = 0

	close(qi.signal)

	qi.cache = map[uint64]interface{}{}

	if qi.signalEnable {
		qi.signal = make(chan open_interface.QueueSignal, 1024)
	}

	return nil
}
