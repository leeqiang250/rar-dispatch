package main

import (
	"dispatch/common"
	"dispatch/log"
	"dispatch/server"
	"dispatch/variable"
)

func main() {
	log.Init("./log", 4)

	variable.Conf = common.ConfigInit()

	go variable.FileWork.Run()

	server.Run()
}
