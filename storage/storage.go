package storage

import (
	"github.com/redresseur/message_bus/open_interface"
	"sync"
)

type StorageImpl struct {
	cache sync.Map
}

func (si *StorageImpl) Get(key interface{}) (interface{}, error) {
	v, ok := si.cache.Load(key)
	if !ok {
		return nil, open_interface.ErrKeyIsInvalid
	}

	return v, nil
}

func (si *StorageImpl) Put(key interface{}, value interface{}) (error) {
	si.cache.Store(key, value)
	return nil
}

func (si *StorageImpl) Delete(key interface{}) (error) {
	si.cache.Delete(key)
	return nil
}

func (si *StorageImpl) Range(list func(key, v interface{}) bool) (error) {
	si.cache.Range(list)
	return nil
}

type StorageHandlerImpl struct {

}

func (*StorageHandlerImpl) New(id string) (open_interface.Storage, error) {
	return &StorageImpl{}, nil
}

func (*StorageHandlerImpl) Release(storage open_interface.Storage) (error) {
	storage = nil
	return nil
}
