package storage

import (
	"sync"
)

type LocalStorage struct {
	sync.Map
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{sync.Map{}}
}

func (ls *LocalStorage) Get(key string) (string, bool) {
	val, ok := ls.Load(key)
	if !ok {
		return "", false
	}
	return val.(string), true
}

func (ls *LocalStorage) Del(key string) bool {
	_, exist := ls.LoadAndDelete(key)
	return exist
}

func (ls *LocalStorage) Set(key, value string) error {
	ls.Store(key, value)
	return nil
}

func (ls *LocalStorage) Update(key, value string) bool {
	if _, exist := ls.Load(key); !exist {
		return false
	}
	ls.Store(key, value)
	return true
}

func (ls *LocalStorage) CAS(key, oldVal, newVal string) bool {
	return ls.CompareAndSwap(key, oldVal, newVal)
}
