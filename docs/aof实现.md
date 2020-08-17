# AOF持久化实现
## 实现原理方案
redis aof持久化支持三种模式:always,everysec,no

appendfsync配置|解释
---|---
always|aof_buf缓冲区所有内容写入并同步到aof文件
everysec|aof_buf缓冲区所有内容写入aof文件，如果距离上一次同步文件超过1s,则将同步aof文件，该操作由另外一个线程负责
no|aof_buf缓冲区所有内容写入aof文件，但不对aof文件同步由系统决定何时同步

aof保存的为操作redis的所有写操作，例如：set,sadd,incr等。

aof格式(redis aof文件还包含版本信息，这里忽略)为：
```
*3
$3
set
$1
a
$1
a
*3
$3
set
$1
b
$1
b
```

aof主要包含三个部分：
1. 持久化aof文件
2. 载入aof文件
3. 重写aof文件

## 实现方案
### 持久化aof文件
实现思路：对客户端写事件，将该数据保存到aof_buf中，

定义一个全局切片`var aofBuf = make([]byte, 0)`

对于客户端发来的数据会读取为Value,但是这个数据已经被处理，没有相应的符号和`\r\n`

在主处理逻辑中针对解析的Value，循环遍历处理写事件
```
func appendAOFBuf(command string, value *Value) {
	if appendonly {
		switch strings.ToLower(command) {
		case "set", "del":
			raw := ValueToRow(value)
			aofBuf = append(aofBuf, raw...)
		}
	}
}
```
其中ValueToRow只是为了将结构化数据再转换为原始请求数据

```
func ValueToRow(value *Value) []byte {
	bs := []byte{value.Type}
	switch value.Type {

	case TypeSimpleError:
		str := []byte(strconv.Itoa(len(value.Str)) + CRLF + value.Str + CRLF)
		bs = append(bs, str...)
	case TypeSimpleString:
		log.Error("wait todo...")
	case TypeArray:
		array := value.Elems
		l := []byte(strconv.Itoa(len(array)) + CRLF)
		bs = append(bs, l...)
		for _, arg := range array {
			bs = append(bs, TypeBlobString)
			str := []byte(strconv.Itoa(len(arg.Str)) + CRLF + arg.Str + CRLF)
			bs = append(bs, str...)
		}
	default:
		log.Error("wait todo...")
	}
	return bs
}
```
接下来处理写aof文件，根据相应的appendfsync决定刷盘机制
```
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
```
### 载入aof文件
**redis实现思路**：程序启动先检查是否存在aof文件，如果存在，则创建一个fake client，读取aof文件，
然后利用伪客户端发送读取的命令。一直重复直至读取文件结束。

**本项目实现思路**：程序启动先检查是否存在aof文件，如果存在，读取aof文件，调用函数执行set操作。一直重复直至读取文件结束。

在解析文件的时候，需要计算相应的字节数，来确保下一次数据从哪里开始读取

```
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
```
```
func (handle *AOFHandle) ReadValue(offset int64) (*Value, error) {
	handle.file.Seek(offset, 1)
	value, error := handle.redisReader.ReadValue()
	return value, error
}
```
最后是根据解析的Value恢复数据

```
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
```

### 重写aof文件
1. 遍历db,根据不同数据类型，转换成resp协议，生成的数据写成temp文件
2. 重写过程命令写入aof缓存区,aof重写缓存区，待文件重写完成，原子替换原有aof文件
3. 将aof重写缓存区中的数据写入到新生成的aof文件中

为了避免堵塞服务器处理命令，重写过程会在子进程中执行。