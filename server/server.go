package server

import (
	"container/list"
	"net"
	"runtime"
	"strings"
	"time"

	"github.com/TrumanDu/gocache/store/cache"
	log "github.com/TrumanDu/gocache/tools/log"
	pool "github.com/TrumanDu/gocache/tools/pool"
	"github.com/spf13/viper"
)

var clientsMap = make(map[string]*clientConn)
var ioPool = pool.NewPool(runtime.NumCPU())
var aofHandle = NewAOFHandle()
var appendonly = false
var appendfsync = viper.GetString("gocache.appendfsync")
var aofBuf = make([]byte, 0)
var aofSyncTime = time.Now().UnixNano()

// 初始化socket 端口监听
// epoll 建立client连接
// 处理请求：解析命令
// 主线程执行命令
// 将结果返回给client
func Run() {

	loadAofFile()

	if strings.EqualFold("yes", viper.GetString("gocache.appendonly")) {
		appendonly = true
	}
	port := viper.GetString("gocache.port")
	listen, err := net.Listen("tcp", ":"+port)
	log.Infof("listen port:%s", port)
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

		if len(connections) > 0 {
			coreProcess(epoller, connections)
		}
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
	// 1.多线程读取解析client发来的数据（命令和数据类型等）
	clientsReadSyncList := handleClientsWithPendingReadsUsingThreads(epoller, connections)
	// 2.单线程执行command
	responses := handleCommand((*clientsReadSyncList).list)

	if appendonly {
		if len(aofBuf) > 0 {
			appendAOF(aofBuf)
		}
	}
	// 3.多线程向client发送响应数据
	handleClientsWithPendingWritesUsingThreads(responses)
}

func handleClientsWithPendingReadsUsingThreads(epoller *Epoller, connections []net.Conn) *SyncList {
	funcs := make([]func(), len(connections))
	clientsReadSyncList := NewSyncList()
	for i := 0; i < len(connections); i++ {
		conn := connections[i]
		id := conn.RemoteAddr().String()
		funcs[i] = func() {
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
			clientsReadSyncList.Add(data)
		}
	}
	ioPool.SyncRun(funcs)
	return clientsReadSyncList
}

func handleCommand(clientsRead *list.List) (responseList *list.List) {
	responseList = list.New()
	aofBuf = make([]byte, 0)
	for e := clientsRead.Front(); e != nil; e = e.Next() {

		data := e.Value.(*readData)
		value := data.value
		id := data.id
		wt := clientsMap[id].wt
		var responseBytes []byte
		command := ""
		switch value.Type {

		case TypeSimpleError:
			log.Error(value.Err)
		case TypeSimpleString:
			log.Error("wait todo...")
		case TypeArray:
			array := value.Elems
			command = strings.ToLower(array[0].Str)
			switch command {
			case "ping":
				responseBytes = wt.replyString("PONG")
			case "quit":
				responseBytes = wt.replyString("OK")
			case "set":
				if len(array) < 3 {
					responseBytes = wt.replyInvalidSyntax()
				} else {
					cache.Set(array[1].Str, array[2].Str)
					responseBytes = wt.replyString("OK")
				}

			case "exists":
				if len(array) < 2 {
					responseBytes = wt.replyInvalidSyntax()
				} else {
					responseBytes = wt.replyNumber(cache.Exists(array[1].Str))
				}
			case "get":
				if len(array) < 2 {
					responseBytes = wt.replyInvalidSyntax()
				} else {
					data := cache.Get(array[1].Str)
					if data != "" {
						responseBytes = wt.replyString(data)
					} else {
						responseBytes = wt.replyNull()
					}

				}
			case "del":
				if len(array) < 2 {
					responseBytes = wt.replyInvalidSyntax()
				} else {
					responseBytes = wt.replyNumber(cache.Del(array[1].Str))
				}
			case "command":
				empty := make([]string, 0)
				responseBytes = wt.replyArray(empty)
			default:
				responseBytes = wt.replyCommandNotSupport(array[0].Str)
			}

		default:
			responseBytes = wt.replyInvalidSyntax()
		}
		appendAOFBuf(command, value)
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

func appendAOF(aofBuf []byte) {
	if n := len(aofBuf); n > 0 {
		aofHandle.Write(aofBuf)
		switch appendfsync {
		case "always":
			fsync()
		case "everysec":
			if nowTime := time.Now().UnixNano(); (nowTime - aofSyncTime) > 1e9 {
				go fsync()
			}
		}
	}

}

func fsync() {
	aofHandle.Flush()
	aofSyncTime = time.Now().UnixNano()
}

func appendAOFBuf(command string, value *Value) {
	if appendonly {
		switch strings.ToLower(command) {
		case "set", "del":
			raw := ValueToRow(value)
			aofBuf = append(aofBuf, raw...)
		}
	}
}

func loadAofFile() {
	if aofHandle != nil {
		stat, err := aofHandle.file.Stat()
		if err != nil {
			log.Error("loadAofFile error:", err)
		}
		fileSize := stat.Size()

		if fileSize > 0 {
			i := int64(0)
			for i < fileSize {
				value, err := aofHandle.ReadValue(i)
				if err != nil {
					log.Error("loadAofFile ReadValue error:", err)
				}
				handleAOFReadCommand(value)
				i = i + value.Size
			}
		}

		log.Infof("loadAofFile success,aof size:%d", fileSize)
	}
}

func handleAOFReadCommand(value *Value) {
	command := ""
	switch value.Type {
	case TypeSimpleError:
		log.Error(value.Err)
	case TypeSimpleString:
		log.Error("wait todo...")
	case TypeArray:
		array := value.Elems
		command = strings.ToLower(array[0].Str)
		switch command {
		case "set":
			cache.Set(array[1].Str, array[2].Str)
		case "del":
			cache.Del(array[1].Str)
		default:
			log.Error("wait todo...")
		}
	default:
		log.Error("wait todo...")
	}
}

type clientConn struct {
	conn net.Conn
	addr string
	rd   *RedisReader
	wt   *RedisWriter
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
