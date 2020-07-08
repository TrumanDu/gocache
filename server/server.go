package gocache

import (
	"fmt"
	cache "github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
	"net"
	"strconv"
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

	for {
		//阻塞等待用户链接
		conn, err := listen.Accept()
		if err != nil {
			log.Error(err)
			return
		}

		go handleConnect(conn)

	}

	cache.Set("truman", "trumandu")

	fmt.Println(cache.Get("truman"))
}

func handleConnect(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err1 := conn.Read(buf)
	if err1 != nil {
		log.Error(err1)
		return
	}
	fmt.Println("buf = ", string(buf[:n])) //指定打印的切片，即读了多少就打印多少
	conn.Write([]byte("OK"))
	defer conn.Close() //关闭当前用户链接
}
