package gocache

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strconv"
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

var ErrInvalidSyntax = errors.New("resp:invalid syntax")

// Reader struct
type Reader struct {
	*bufio.Reader
}

// NewReader method
func NewReader(reader io.Reader) *Reader {
	defaultSize := 32 * 1024
	return &Reader{Reader: bufio.NewReaderSize(reader, defaultSize)}
}

// ReadValue method
func (r *Reader) ReadValue() ([]byte, error) {
	line, err := r.readLine()
	if err != nil {
		return nil, err
	}
	if len(line) < 3 {
		return nil, ErrInvalidSyntax
	}

	switch line[0] {
	case TypeSimpleString, TypeNumber, TypeSimpleError, TypeBoolean, TypeDouble, TypeBigNumber:
		return line, nil
	case TypeBlobString, TypeBlobError:
		return r.readBlobString(line)
	case TypeArray:

	default:
		return nil, ErrInvalidSyntax
	}
}

// readLine \r\n
func (r *Reader) readLine() ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	if len(line) > 1 && line[len(line)-2] == '\r' {
		return line, nil
	}

	return nil, ErrInvalidSyntax
}

func (r *Reader) getCount(line []byte) (int, error) {
	end := bytes.IndexByte(line, '\r')
	return strconv.Atoi(string(line[1:end]))
}

func (r *Reader) readBlobString(line []byte) ([]byte, error) {
	count, err := r.getCount(line)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, count+2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
