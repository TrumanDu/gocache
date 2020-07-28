package server

import (
	"container/list"
	"net"
	"runtime"
	"strconv"
	"strings"

	"github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
	pool "github.com/TrumanDu/gocache/tools/pool"
)

var clientsMap = make(map[string]*clientConn)
var ioPool = pool.NewPool(runtime.NumCPU())

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
	log.Infof("cpu num:%d", runtime.NumCPU())
	if err != nil {
		log.Error("listen fail:", err)
		return
	}

	defer listen.Close()

	epoller, err := NewEpoller()

	if err != nil {
		panic(err)
	}

	go listenClientConnect(epoller, listen)

	for {
		connections, err := epoller.Wait()
		if err != nil {
			log.Error("epoll wait error:", err)
			continue
		}

		coreProcess(epoller, connections)
	}

}

func listenClientConnect(epoller *Epoller, listen net.Listener) {
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
	}
}

func coreProcess(epoller *Epoller, connections []net.Conn) {
	// 1.
	clientsRead := handleClientsWithPendingReadsUsingThreads(epoller, connections)
	// 2.
	responses := handleCommand((*clientsRead).l)
	// 3.
	handleClientsWithPendingWritesUsingThreads(responses)
}

func handleClientsWithPendingReadsUsingThreads(epoller *Epoller, connections []net.Conn) *SyncList {
	fs := make([]func(), len(connections))
	clientsRead := NewSyncList()
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
		id := conn.RemoteAddr().String()
		fs[i] = func() {
			clientConn := clientsMap[id]

			value, err1 := clientConn.rd.ReadValue()

			if err1 != nil {
				if err := epoller.Remove(conn); err != nil {
					log.Error("failed to remove :", err)
				}
				conn.Close()
				return
			}

			data := &readData{id, value}
			clientsRead.Add(data)
		}
	}
	ioPool.SyncRun(fs)
	return clientsRead
}

func handleCommand(clientsRead *list.List) *list.List {
	responseList := list.New()
	for e := clientsRead.Front(); e != nil; e = e.Next() {

		data := e.Value.(*readData)
		value := data.value
		var responseBytes []byte
		command := ""
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
				responseBytes = ReplyString("PONG")
			case "quit":
				responseBytes = ReplyString("OK")
			case "set":
				if len(array) < 3 {
					responseBytes = InvalidSyntax()
				} else {
					cache.Set(array[1].Str, array[2].Str)
					responseBytes = ReplyString("OK")
				}

			case "exists":
				if len(array) < 2 {
					responseBytes = InvalidSyntax()
				} else {
					responseBytes = ReplyNumber(cache.Exists(array[1].Str))
				}
			case "get":
				if len(array) < 2 {
					responseBytes = InvalidSyntax()
				} else {
					data := cache.Get(array[1].Str)
					if data != "" {
						responseBytes = ReplyString(data)
					} else {
						responseBytes = ReplyNull()
					}

				}
			case "del":
				if len(array) < 2 {
					responseBytes = InvalidSyntax()
				} else {
					responseBytes = ReplyNumber(cache.Del(array[1].Str))
				}
			case "command":
				empty := make([]string, 0)
				responseBytes = ReplyArray(empty)
			default:
				responseBytes = CommandNotSupport(array[0].Str)
			}

		default:
			responseBytes = InvalidSyntax()
		}
		obj := &responseData{id: data.id, command: command, data: responseBytes}
		responseList.PushFront(obj)
	}

	return responseList
}

func handleClientsWithPendingWritesUsingThreads(responses *list.List) {
	fs := make([]func(), (*responses).Len())
	i := 0
	for e := responses.Front(); e != nil; e = e.Next() {
		obj := e.Value.(*responseData)
		fs[i] = func() {
			clientConn := clientsMap[obj.id]
			_, err := clientConn.wt.Write(obj.data)
			if strings.ToLower(obj.command) == "quit" {
				clientConn.conn.Close()
			}
			if err != nil {
				log.Error("response message error:", err)
			}
			clientConn.wt.Flush()
		}
		i = i + 1
	}
	ioPool.SyncRun(fs)
}

type readData struct {
	id    string
	value *Value
}

type responseData struct {
	id      string
	command string
	data    []byte
}

type clientConn struct {
	conn net.Conn
	addr string
	rd   *RedisReader
	wt   *RedisWriter
}
