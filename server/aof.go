package server

import (
	"bufio"
	"os"

	"github.com/TrumanDu/gocache/tools/log"
)

type AOFHandle struct {
	file           *os.File
	bufferedWriter *bufio.Writer
}

func NewAOFHandle() *AOFHandle {
	fileName := "appendonly.aof"
	_, err := os.Stat(fileName)
	if err != nil {
		_, err = os.Create(fileName)
		if err != nil {
			log.Error("create ppendonly.aof has error", err)
		}
	}

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Error("open appendonly.aof has error", err)
	}

	bufferedWriter := bufio.NewWriter(file)

	return &AOFHandle{file, bufferedWriter}
}

func (handle *AOFHandle) Write(bytes []byte) {
	_, err := handle.bufferedWriter.Write(bytes)
	if err != nil {
		log.Error("Write appendonly.aof has error ", err)
	}
}

func (handle *AOFHandle) Flush() {
	err := handle.bufferedWriter.Flush()
	if err != nil {
		log.Error("flush appendonly.aof has error ", err)
	}
}
