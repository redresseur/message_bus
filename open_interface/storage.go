package open_interface

// kv结构的存储接口
type Storage interface {
	Get(key interface{})(interface{}, error)
	Put(key interface{}, value interface{})(error)
	Delete(key interface{})(error)
	Range(func(key, v interface{})bool)(error)
}

type StorageHandler interface {
	New(id string) (Storage, error)
	Release(storage Storage)(error)
}
