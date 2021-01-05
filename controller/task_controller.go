package controller

import (
	"dispatch/dto"
	"dispatch/variable"
	"net/http"
)

func TaskInit() map[string]func(http.ResponseWriter, *http.Request) {
	fun := make(map[string]func(http.ResponseWriter, *http.Request))
	fun["/task-get"] = Get
	fun["/task-confirm"] = Confirm
	fun["/task-complete"] = Complete
	fun["/task-discover"] = Discover
	fun["/task-file-info"] = FileInfo
	fun["/task-gen-file"] = GenFile
	fun["/task-rar-file-md5"] = RARFileMD5
	fun["/task-download-rar-file"] = DownloadRARFile

	return fun
}

func Get(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Get()).Bytes())
}

func Confirm(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Confirm(request.URL.Query().Get("group"))).Bytes())
}

func Complete(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Complete(request.URL.Query().Get("group"))).Bytes())
}

func Discover(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Discover(request.URL.Query().Get("group"), request.URL.Query().Get("key"))).Bytes())
}

func FileInfo(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.FileInfo()).Bytes())
}

func GenFile(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.SuccessBytes())
}

func RARFileMD5(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.RARFileMD5()).Bytes())
}

func DownloadRARFile(response http.ResponseWriter, request *http.Request) {
	response.Write(variable.FileWork.DownloadRARFile())
}
