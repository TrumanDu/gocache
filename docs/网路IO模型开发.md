# 网路IO模型开发
redis选用的单Reactor模型，虽然go 编程模型对于goroutine创建属于轻量级的，比
线程耗的资源更低，但是一个goroutine stack也会占用2k-8k。对于百万级连接，内存
占用也会很高，达到几十G,为了更好的性能，和更贴近redis设计。我这边也采用Reactor。
## Reactor模型
![](https://img2018.cnblogs.com/blog/1485398/201810/1485398-20181022232220631-1867817712.jpg)

Reactor模型其实就是IO多路复用+池化技术。

多说一句：复用指的是复用了1个线程，一个线程可以同时处理多个fd(文件描述符)

Reactor架构模式允许事件驱动的应用通过多路分发的机制去处理来自不同客户端的多个请求。

## 实践

```
package epoll

import (
	"golang.org/x/sys/unix"
	"os"
)

type Op uint32

const (
	// Just a subset, check the complete x/sys/unix documentation
	EpollIn  Op = unix.EPOLLIN
	EpollOut    = unix.EPOLLOUT
	EpollPri    = unix.EPOLLPRI
	EpollErr    = unix.EPOLLERR
	// It will probably not do what you have in mind without this flag.
	// Without you will get multiple events while the fd is ready.
	EpollEt = unix.EPOLLET
)

type Event struct {
	File *os.File
	Ops  Op
}

type EpollWatcher struct {
	Events chan Event
	Errors chan error

	epfd  int
	files map[int]*os.File
}

/* internal waiting function */
func (w *EpollWatcher) epollWait() {

	events := make([]unix.EpollEvent, 10)
	timeout_msec := -1

	for {
		n, err := unix.EpollWait(w.epfd, events, timeout_msec)
		if err != nil {
			w.Errors <- err
			return
		}
		for _, e := range events[:n] {
			w.Events <- Event{File: w.files[int(e.Fd)], Ops: Op(e.Events)}
		}
	}
}

func NewEpollWatcher() (*EpollWatcher, error) {
	epfd, err := unix.EpollCreate(1) // argument ignored since Linux 2.6.8
	if err != nil {
		return nil, err
	}

	w := &EpollWatcher{Events: make(chan Event), Errors: make(chan error), epfd: epfd}
	w.files = make(map[int]*os.File)

	go w.epollWait() // enter waiting loop

	return w, nil
}

func (w *EpollWatcher) Close() error {
	return unix.Close(w.epfd)
}

func (w *EpollWatcher) Add(file *os.File, ops Op) error {
	events := unix.EpollEvent{
		Events: uint32(ops),
		Fd:     int32(file.Fd()),
	}

	err := unix.EpollCtl(w.epfd, unix.EPOLL_CTL_ADD, int(file.Fd()), &events)
	if err != nil {
		return err
	}
	// keep a map of fd to file so we can return the file object
	w.files[int(file.Fd())] = file
	return nil
}

func (w *EpollWatcher) Modify(file *os.File, ops Op) error {
	events := unix.EpollEvent{
		Events: uint32(ops),
		Fd:     int32(file.Fd()),
	}
	return unix.EpollCtl(w.epfd, unix.EPOLL_CTL_MOD, int(file.Fd()), &events)
}

func (w *EpollWatcher) Remove(file os.File) error {
	err := unix.EpollCtl(w.epfd, unix.EPOLL_CTL_DEL, int(file.Fd()), nil)
	if err != nil {
		return err
	}
	delete(w.files, int(file.Fd()))
	return nil
}
```

## 参考
1. [百万 Go TCP 连接的思考: epoll方式减少资源占用](https://colobu.com/2019/02/23/1m-go-tcp-connection/)
2. [smallnest/1m-go-tcp-server](https://github.com/smallnest/1m-go-tcp-server)
3. [epoll多路复用-----epoll_create1()、epoll_ctl()、epoll_wait()](https://blog.csdn.net/displayMessage/article/details/81151646)
4. [Minimal viable epoll package for go](https://gist.github.com/ast/a41816345e94e065890440e87e41a219)