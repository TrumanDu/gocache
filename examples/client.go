package main

import (
	"fmt"
	"net"

	"github.com/TrumanDu/gocache/tools/log"
)

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	fmt.Println(send("set a a", conn))
	fmt.Println(send("get a", conn))

}

func send(command string, conn net.Conn) string {
	conn.Write([]byte(command))
	buf := make([]byte, 1024)
	n, err1 := conn.Read(buf)
	if err1 != nil {
		log.Error(err1)
		return ""
	}

	return string(buf[:n])
}
