package server

import (
	"bufio"
	"os"

	"github.com/spf13/viper"

	"github.com/TrumanDu/gocache/tools/log"
)

type AOFHandle struct {
	file           *os.File
	bufferedWriter *bufio.Writer
	redisReader    *RedisReader
}

func NewAOFHandle() *AOFHandle {
	fileName := viper.GetString("gocache.appendfilename")
	_, err := os.Stat(fileName)
	if err != nil {
		_, err = os.Create(fileName)
		if err != nil {
			log.Error("create ppendonly.aof has error", err)
		}
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR, os.ModeAppend)
	if err != nil {
		log.Error("open appendonly.aof has error", err)
	}

	bufferedWriter := bufio.NewWriter(file)
	redisReader := NewRedisReader(file)
	return &AOFHandle{file, bufferedWriter, redisReader}
}

func (handle *AOFHandle) Write(bytes []byte) {
	_, err := handle.bufferedWriter.Write(bytes)
	if err != nil {
		log.Error("Write appendonly.aof has error ", err)
	}
}

func (handle *AOFHandle) ReadValue(offset int64) (*Value, error) {
	handle.file.Seek(offset, 1)
	value, error := handle.redisReader.ReadValue()
	return value, error
}

func (handle *AOFHandle) Flush() {
	err := handle.bufferedWriter.Flush()
	if err != nil {
		log.Error("flush appendonly.aof has error ", err)
	}
}
