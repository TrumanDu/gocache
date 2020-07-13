package gocache

import (
	"golang.org/x/sys/unix"
)

type Epoller struct {
	fd int
}

func NewEpoller() (*Epoller, error) {
	epoller := &Epoller{fd: 1}
	unix.EpollCreate1(0)
	return epoller, nil
}
