package main

import (
	"dispatch/common"
	"dispatch/conf"
	"dispatch/log"
	"dispatch/server"
	"dispatch/variable"
)

func main() {
	log.Init("./log", 4)

	conf.Conf = common.ConfigInit()

	variable.FileWork.Run()

	server.Run()
}
