package main

import (
	"fmt"
	"github.com/TrumanDu/gocache/tools/log"
	"net"
	"strconv"
)

func main() {

	event := 0x1

	fmt.Println(strconv.Itoa(event))
	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	//发送数据
	conn.Write([]byte("set a a"))
	buf := make([]byte, 1024)
	n, err1 := conn.Read(buf)
	if err1 != nil {
		log.Error(err1)
		return
	}
	fmt.Println(string(buf[:n]))
}
