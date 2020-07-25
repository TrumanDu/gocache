package main

import (
	"bytes"
	"fmt"
	"net"
	"strconv"

	"github.com/TrumanDu/gocache/tools/log"
)

func main() {

	fmt.Println(len([]byte("\r\n")))
	line := []byte("*3\r\n+set\r\n+truman\r\n+truman\r\n")
	end := bytes.IndexByte(line, '\r')
	count, _ := strconv.Atoi(string(line[1:end]))
	fmt.Println(strconv.Itoa(count))

	m := make(map[string]string)
	m["h"] = "h"
	//delete(m,"h")
	if _, ok := m["h"]; ok {
		fmt.Println(1)
	} else {
		fmt.Println(0)
	}

	conn, err := net.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	fmt.Println(send("set truman hello-world", conn))
	fmt.Println(send("get truman", conn))

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
