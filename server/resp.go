package server

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"math/big"
	"strconv"

	log "github.com/TrumanDu/gocache/tools/log"
)

const (
	// TypeBlobString simple types
	TypeBlobString     = '$' // $<length>\r\n<bytes>\r\n
	TypeSimpleString   = '+' // +<string>\r\n
	TypeSimpleError    = '-' // -<string>\r\n
	TypeNumber         = ':' // :<number>\r\n
	TypeNull           = '_' // _\r\n
	TypeDouble         = ',' // ,<floating-point-number>\r\n
	TypeBoolean        = '#' // #t\r\n or #f\r\n
	TypeBlobError      = '!' // !<length>\r\n<bytes>\r\n
	TypeVerbatimString = '=' // =<length>\r\n<format(3 bytes):><bytes>\r\n
	TypeBigNumber      = '(' // (<big number>\n
	// TypeArray Aggregate data types
	TypeArray     = '*' // *<elements number>\r\n... numelements other types ...
	TypeMap       = '%' // %<elements number>\r\n... numelements key/value pair of other types ...
	TypeSet       = '~' // ~<elements number>\r\n... numelements other types ...
	TypeAttribute = '|' // |~<elements number>\r\n... numelements map type ...
	TypePush      = '>' // ><elements number>\r\n<first item is String>\r\n... numelements-1 other types ...
	// TypeStream special type
	TypeStream = "$EOF:" //
)

type Value struct {
	Type    byte
	Str     string
	Err     string
	Integer int64
	Boolean bool
	Double  float64
	BigInt  *big.Int
	Elems   []*Value // for array & set
	// KV           *linkedhashmap.Map
	// Attrs        *linkedhashmap.Map
	StreamMarker string

	Size int64
}

var ErrInvalidSyntax = errors.New("resp:invalid syntax")
var defaultSize = 32 * 1024
var CRLF = "\r\n"

// Reader struct
type RedisReader struct {
	*bufio.Reader
}

// NewRedisReader method
func NewRedisReader(reader io.Reader) *RedisReader {

	return &RedisReader{Reader: bufio.NewReaderSize(reader, defaultSize)}
}

// ReadValue method
func (r *RedisReader) ReadValue() (*Value, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	if len(line) < 3 {
		return nil, ErrInvalidSyntax
	}
	v := &Value{
		Type: line[0],
	}
	switch v.Type {
	case TypeSimpleString, TypeSimpleError:
		v.Str, err = r.readSimpleString(line)
		v.Size = int64(3) + int64(len(v.Str))
	case TypeNumber, TypeBoolean, TypeDouble, TypeBigNumber:
		// TODO 待实现
		v.Str, err = r.readSimpleString(line)
	case TypeBlobString, TypeBlobError:
		v.Str, v.Size, err = r.readBlobString(line)
	case TypeArray:
		v.Elems, v.Size, err = r.readArray(line)
	}
	return v, err
}

// readLine \r\n
func (r *RedisReader) readLine() ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) > 1 && line[len(line)-2] == '\r' {
		return line, nil
	}

	return nil, ErrInvalidSyntax
}

func (r *RedisReader) getCount(line []byte) (int, int, error) {
	end := bytes.IndexByte(line, '\r')
	count, err := strconv.Atoi(string(line[1:end]))
	return count, end + 2, err
}

func (r *RedisReader) readSimpleString(line []byte) (string, error) {
	return string(line[1 : len(line)-2]), nil
}

func (r *RedisReader) readBlobString(line []byte) (string, int64, error) {
	count, index, err := r.getCount(line)
	if err != nil {
		return "", 0, err
	}
	buf := make([]byte, count+2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", 0, err
	}
	return string(buf[:count]), int64(count) + int64(index+2), nil
}

func (r *RedisReader) readArray(line []byte) ([]*Value, int64, error) {
	count, index, err := r.getCount(line)
	byteSize := int64(index)
	if err != nil {
		return nil, 0, err
	}
	var values []*Value
	for i := 0; i < count; i++ {
		v, err := r.ReadValue()
		if err != nil {
			return nil, 0, err
		}
		byteSize = byteSize + v.Size
		values = append(values, v)
	}

	return values, byteSize, nil
}

type RedisWriter struct {
	*bufio.Writer
}

func NewRedisWriter(writer io.Writer) *RedisWriter {
	return &RedisWriter{bufio.NewWriterSize(writer, defaultSize)}
}

func (w *RedisWriter) replyInvalidSyntax() []byte {
	return []byte("-resp:invalid syntax" + CRLF)
}
func (w *RedisWriter) replyCommandNotSupport(command string) []byte {
	str := "not support redis command:" + command
	log.Info(str)
	return []byte("-resp:" + str + CRLF)
}

func (w *RedisWriter) replyString(message string) []byte {
	bs := []byte{TypeSimpleString}
	my := []byte(message + CRLF)
	bs = append(bs, my...)
	return bs
}

func (w *RedisWriter) replyArray(messages []string) []byte {
	bs := []byte{TypeArray}
	my := []byte(strconv.Itoa(len(messages)) + CRLF)
	bs = append(bs, my...)

	for _, arg := range messages {
		bs = append(bs, TypeBlobString)
		str := []byte(strconv.Itoa(len(arg)) + CRLF + arg + CRLF)
		bs = append(bs, str...)
	}
	return bs
}

func (w *RedisWriter) replyNull() []byte {
	bs := []byte{TypeBlobString, '-', '1'}
	bs = append(bs, CRLF...)
	return bs
}
func (w *RedisWriter) replyNumber(num int) []byte {
	bs := []byte{TypeNumber}
	my := []byte(strconv.Itoa(num) + CRLF)
	bs = append(bs, my...)
	return bs
}

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
