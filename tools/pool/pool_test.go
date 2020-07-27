package pool

import (
	"testing"
	"time"
)

func TestPoolRun(test *testing.T) {
	pool := NewPool(3)
	count := 4
	fs := make([]func(), count)
	for i := 0; i < count; i++ {
		f := func() {
			time.Sleep(time.Duration(1) * time.Second)
		}
		fs[i] = f
	}
	pool.SyncRun(fs)
}

func TestPoolAsyncRun(test *testing.T) {
	pool := NewPool(3)
	count := 4
	fs := make([]func(), count)
	for i := 0; i < count; i++ {
		f := func() {
			time.Sleep(time.Duration(1) * time.Second)
		}
		fs[i] = f
	}
	pool.AsyncRun(fs)
}
