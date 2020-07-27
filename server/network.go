package server

import (
	"net"
	"reflect"

	"golang.org/x/sys/unix"
)

//Op op
type Op uint32

//Epoller Epoller
type Epoller struct {
	epfd        int
	connections map[int]net.Conn
}

//NewEpoller new epoller
func NewEpoller() (*Epoller, error) {

	epfd, err := unix.EpollCreate(1)
	if err != nil {
		return nil, err
	}
	epoller := &Epoller{epfd: epfd, connections: make(map[int]net.Conn)}
	return epoller, nil
}

func (epoller *Epoller) Close() error {
	return unix.Close(epoller.epfd)
}

func (epoller *Epoller) Add(conn net.Conn) error {
	fd := socketFD(conn)
	event := unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLHUP,
		Fd:     int32(fd),
	}
	err := unix.EpollCtl(epoller.epfd, unix.EPOLL_CTL_ADD, fd, &event)

	if err != nil {
		return err
	}
	epoller.connections[fd] = conn
	return nil
}

func (epoller *Epoller) Remove(conn net.Conn) error {
	fd := socketFD(conn)
	err := unix.EpollCtl(epoller.epfd, unix.EPOLL_CTL_DEL, fd, nil)

	if err != nil {
		return err
	}
	delete(epoller.connections, fd)
	return nil
}

func (epoller *Epoller) Wait() ([]net.Conn, error) {
	events := make([]unix.EpollEvent, 10)
	n, err := unix.EpollWait(epoller.epfd, events, 10)
	if err != nil && err != unix.EINTR {
		return nil, err
	}
	var connections []net.Conn
	for i := 0; i < n; i++ {
		conn := epoller.connections[int(events[i].Fd)]
		connections = append(connections, conn)
	}
	return connections, nil
}

func socketFD(conn net.Conn) int {

	tcpConn := reflect.Indirect(reflect.ValueOf(conn)).FieldByName("conn")
	fdVal := tcpConn.FieldByName("fd")
	pfdVal := reflect.Indirect(fdVal).FieldByName("pfd")
	return int(pfdVal.FieldByName("Sysfd").Int())
}
