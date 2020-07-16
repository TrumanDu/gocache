package main

import (
	cmd "github.com/TrumanDu/gocache/cmd/gocache"
	log "github.com/TrumanDu/gocache/tools/log"
)

func main() {

	log.Infof("start to run!")
	cmd.Run()

}
