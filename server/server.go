package gocache

import (
	"net"
	"strconv"
	"strings"

	"github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
)

var clientsMap = make(map[string]*clientConn)

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
		key := conn.RemoteAddr().String()
		clientsMap[key] = &clientConn{conn, conn.RemoteAddr().String(), NewRedisReader(conn), NewRedisWriter(conn)}
		epoller.Add(conn)
		// go handleConnect(conn)
	}

}

func run(epoller *Epoller) {
	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Error("epoll wait error:", err)
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
	key := conn.RemoteAddr().String()
	clientConn := clientsMap[key]
	value, err1 := clientConn.rd.ReadValue()

	if err1 != nil {
		if err := epoller.Remove(conn); err != nil {
			log.Error("failed to remove :", err)
		}
		conn.Close()
		return
	}

	switch value.Type {
	case TypeSimpleError:
		log.Error(value.Err)
	case TypeSimpleString:
		log.Error("wait todo...")
	case TypeArray:
		array := value.Elems
		command := strings.ToLower(array[0].Str)
		switch command {
		case "ping":
			replyString(clientConn, "PONG")
		case "quit":
			replyString(clientConn, "OK")
			clientConn.conn.Close()
		case "set":
			if len(array) < 3 {
				invalidSyntax(clientConn)
			} else {
				cache.Set(array[1].Str, array[2].Str)
			}
			replyString(clientConn, "OK")
		case "exists":
			if len(array) < 2 {
				invalidSyntax(clientConn)
			} else {
				replyNumber(clientConn, cache.Exists(array[1].Str))
			}
		case "get":
			if len(array) < 2 {
				invalidSyntax(clientConn)
			} else {
				data := cache.Get(array[1].Str)
				if data != "" {
					replyString(clientConn, data)
				} else {
					replyNull(clientConn)
				}

			}
		case "del":
			if len(array) < 2 {
				invalidSyntax(clientConn)
			} else {
				replyNumber(clientConn, cache.Del(array[1].Str))
			}
		case "command":
			empty := make([]string, 0)
			replyArray(clientConn, empty)
		default:
			commandNotSupport(clientConn, array[0].Str)
		}

	default:
		invalidSyntax(clientConn)
	}
}

func invalidSyntax(conn *clientConn) {
	_, err := conn.wt.Write([]byte("-resp:invalid syntax \r\n"))
	if err != nil {
		log.Error("response message error:", err)
	}
}
func commandNotSupport(conn *clientConn, command string) {
	str := "not support redis command:" + command
	log.Info(str)
	_, err := conn.wt.Write([]byte("-resp:" + str + " \r\n"))
	if err != nil {
		log.Error("response message error:", err)
	}
}

func replyString(client *clientConn, message string) {
	err := client.wt.WriteSimpleString(message)
	if err != nil {
		log.Error("response message error:", err)
	}
}

func replyArray(client *clientConn, messages []string) {
	err := client.wt.WriteArray(messages)
	if err != nil {
		log.Error("response message error:", err)
	}
}

func replyNull(client *clientConn) {
	err := client.wt.WriteNull()
	if err != nil {
		log.Error("response null message error:", err)
	}
}
func replyNumber(client *clientConn, num int) {
	err := client.wt.WriteNumber(num)
	if err != nil {
		log.Error("response message error:", err)
	}
}

type clientConn struct {
	conn net.Conn
	addr string
	rd   *RedisReader
	wt   *RedisWriter
}
