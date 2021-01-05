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
	http.Handle("/resource-file/", http.StripPrefix("/resource-file/", http.FileServer(http.Dir("/resource"))))

	for k, v := range controller.TaskInit() {
		http.HandleFunc(k, v)
	}

	fmt.Println("starting dispatch at port", variable.Conf.App.Port)

	err := http.ListenAndServe(":"+variable.Conf.App.Port, nil)
	if err != nil {
		log.Error.Println("ListenAndServe: ", err)
		os.Exit(0)
	}
}
