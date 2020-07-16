package gocache

import (
	"net"
	"strconv"
	"strings"

	"github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
)

func Run() {
	// 初始化socket 端口监听
	// epoll 建立client连接
	// 处理请求：解析命令
	// 主线程执行命令
	// 将结果返回给client
	port := 6379
	address := ":" + strconv.Itoa(port)
	listen, err := net.Listen("tcp", address)
	log.Infof("listen port:%d", port)
	if err != nil {
		log.Error("listen fail:", err)
		return
	}

	defer listen.Close()

	epoller, err := NewEpoller()

	if err != nil {
		panic(err)
	}

	go run(epoller)

	for {
		//阻塞等待用户链接
		conn, err := listen.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		epoller.Add(conn)

		// go handleConnect(conn)
	}

}

func run(epoller *Epoller) {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Error(err)
			continue
		}

		for _, conn := range connections {
			handleConnect(epoller, conn)
		}
	}

}

func handleConnect(epoller *Epoller, conn net.Conn) {
	if conn == nil {
		return
	}

	buf := make([]byte, 1024)
	n, err1 := conn.Read(buf)
	if err1 != nil {
		if err := epoller.Remove(conn); err != nil {
			log.Error("failed to remove %v", err)
		}
		conn.Close()
		return
	}

	msg := string(buf[:n])
	array := strings.Split(msg, " ")

	switch strings.ToLower(array[0]) {
	case "set":
		cache.Set(array[1], array[2])
		conn.Write([]byte("OK"))
	case "get":
		conn.Write([]byte(cache.Get(array[1])))
	}
}
