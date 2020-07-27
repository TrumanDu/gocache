package server

import (
	"container/list"
	"sync"
)

var lock sync.Mutex

type SyncList struct {
	l *list.List
}

func NewSyncList() *SyncList {
	l := list.New()
	return &SyncList{l}
}

func (syncList *SyncList) Add(data interface{}) {
	defer lock.Unlock()
	lock.Lock()
	(*syncList).l.PushFront(data)
}
