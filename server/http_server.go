package server

import (
	"dispatch/controller"
	"dispatch/log"
	"dispatch/variable"
	"fmt"
	"net/http"
	"os"
)

func Run() {
	http.Handle("/file/", http.StripPrefix("/file/", http.FileServer(http.Dir("./"))))

	for k, v := range controller.TaskInit() {
		http.HandleFunc(k, v)
	}

	fmt.Println("starting dispatch at port", variable.Conf.App.Port)
	log.Info.Println("starting dispatch at port", variable.Conf.App.Port)

	err := http.ListenAndServe(":"+variable.Conf.App.Port, nil)
	if err != nil {
		log.Error.Println("ListenAndServe: ", err)
		os.Exit(0)
	}
}
