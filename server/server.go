package gocache

import (
	"net"
	"strconv"
	"strings"

	"github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
)

var clientsMap = make(map[*net.Conn]*clientConn)

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
		clientsMap[&conn] = &clientConn{conn, conn.RemoteAddr().String(), NewReader(conn), NewWriter(conn)}
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

	clientConn := clientsMap[&conn]
	value, err1 := clientConn.rd.ReadValue()

	if err1 != nil {
		if err := epoller.Remove(conn); err != nil {
			log.Error("failed to remove %v", err)
		}
		conn.Close()
		return
	}

	switch value.Type {
	case TypeSimpleError:
		log.Error(value.Err)
	case TypeSimpleString:
		if strings.EqualFold(strings.ToLower(value.Str), "ping") {
			clientConn.wt.WriteCommand("PONG")
		}
	case TypeArray:
		array := value.Elems
		command := strings.ToLower(array[0].Str)
		switch command {
		case "set":
			if len(array) < 3 {
				invalidSyntax(clientConn)
			} else {
				cache.Set(array[1].Str, array[2].Str)
			}
		case "get":
			if len(array) < 2 {
				invalidSyntax(clientConn)
			} else {
				data := cache.Get(array[1].Str)
				clientConn.wt.WriteCommand(data)
			}
		case "del":
			if len(array) < 2 {
				invalidSyntax(clientConn)
			} else {
				cache.Del(array[1].Str)
				clientConn.wt.WriteCommand("OK")
			}
		default:
			invalidSyntax(clientConn)
		}

	default:
		invalidSyntax(clientConn)
	}
}

func invalidSyntax(conn *clientConn) {
	conn.wt.Write([]byte("-resp:invalid syntax \r\n"))
}

type clientConn struct {
	conn net.Conn
	addr string
	rd   *Reader
	wt   *Writer
}
