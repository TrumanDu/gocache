package main

import (
	_ "github.com/TrumanDu/gocache/configs"
	server "github.com/TrumanDu/gocache/server"
	log "github.com/TrumanDu/gocache/tools/log"
)

func main() {

	log.Infof("start to run!")
	server.Run()
}
