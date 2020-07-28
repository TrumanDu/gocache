package server

import (
	"container/list"
	"sync"
)

var lock sync.Mutex

type SyncList struct {
	list *list.List
}

func NewSyncList() *SyncList {
	list := list.New()
	return &SyncList{list}
}

func (syncList *SyncList) Add(data interface{}) {
	defer lock.Unlock()
	lock.Lock()
	(*syncList).list.PushFront(data)
}
