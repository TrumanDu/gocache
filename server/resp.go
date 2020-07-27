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
	StrFmt  string
	Err     string
	Integer int64
	Boolean bool
	Double  float64
	BigInt  *big.Int
	Elems   []*Value // for array & set
	// KV           *linkedhashmap.Map
	// Attrs        *linkedhashmap.Map
	StreamMarker string
}

var ErrInvalidSyntax = errors.New("resp:invalid syntax")
var defaultSize = 32 * 1024
var CRLF = []byte("\r\n")

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
	case TypeNumber, TypeBoolean, TypeDouble, TypeBigNumber:
		// TODO 待实现
		v.Str, err = r.readSimpleString(line)
	case TypeBlobString, TypeBlobError:
		v.Str, err = r.readBlobString(line)
	case TypeArray:
		v.Elems, err = r.readArray(line)
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

func (r *RedisReader) getCount(line []byte) (int, error) {
	end := bytes.IndexByte(line, '\r')
	return strconv.Atoi(string(line[1:end]))
}

func (r *RedisReader) readSimpleString(line []byte) (string, error) {
	return string(line[1 : len(line)-2]), nil
}

func (r *RedisReader) readBlobString(line []byte) (string, error) {
	count, err := r.getCount(line)
	if err != nil {
		return "", err
	}
	buf := make([]byte, count+2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return "", err
	}
	return string(buf[:count]), nil
}

func (r *RedisReader) readArray(line []byte) ([]*Value, error) {
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	var values []*Value
	for i := 0; i < count; i++ {
		v, err := r.ReadValue()
		if err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	return values, nil
}

type RedisWriter struct {
	*bufio.Writer
}

func NewRedisWriter(writer io.Writer) *RedisWriter {
	return &RedisWriter{bufio.NewWriterSize(writer, defaultSize)}
}

func InvalidSyntax() []byte {
	return []byte("-resp:invalid syntax\r\n")
}
func CommandNotSupport(command string) []byte {
	str := "not support redis command:" + command
	log.Info(str)
	return []byte("-resp:" + str + " \r\n")
}

func ReplyString(message string) []byte {
	bs := []byte{TypeSimpleString}
	my := []byte(message)
	bs = append(bs, my...)
	bs = append(bs, CRLF...)
	return bs
}

func ReplyArray(messages []string) []byte {
	bs := []byte{TypeArray}
	my := []byte(strconv.Itoa(len(messages)))
	bs = append(bs, my...)
	bs = append(bs, CRLF...)

	for _, arg := range messages {
		bs = append(bs, TypeBlobString)
		str := []byte(strconv.Itoa(len(arg)))
		bs = append(bs, str...)
		bs = append(bs, CRLF...)
		data := []byte(arg)
		bs = append(bs, data...)
		bs = append(bs, CRLF...)
	}
	return bs
}

func ReplyNull() []byte {
	bs := []byte{TypeBlobString, '-', '1'}
	bs = append(bs, CRLF...)
	return bs
}
func ReplyNumber(num int) []byte {
	bs := []byte{TypeNumber}
	my := []byte(strconv.Itoa(num))
	bs = append(bs, my...)
	bs = append(bs, CRLF...)
	return bs
}
