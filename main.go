package main

import (
	server "github.com/TrumanDu/gocache/server"
	log "github.com/TrumanDu/gocache/tools/log"
)

func main() {

	log.Infof("start to run!")
	server.Run()
}
