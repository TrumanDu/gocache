package pool

import (
	"sync"
)

var wg = sync.WaitGroup{}

type Pool struct {
	threadNum int
	ch        chan struct{}
}

func NewPool(threadNum int) *Pool {
	ch := make(chan struct{}, threadNum)
	return &Pool{threadNum, ch}
}

func (pool *Pool) AsyncRun(fs []func()) {
	for i := 0; i < len(fs); i++ {
		pool.ch <- struct{}{}
		f := fs[i]
		go func() {
			f()
			<-pool.ch
		}()
	}
}

func (pool *Pool) SyncRun(fs []func()) {
	for i := 0; i < len(fs); i++ {
		pool.ch <- struct{}{}
		wg.Add(1)
		f := fs[i]
		go func() {
			f()
			<-pool.ch
			wg.Done()
		}()
	}

	wg.Wait()
}
