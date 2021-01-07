package controller

import (
	"dispatch/dto"
	"dispatch/time"
	"dispatch/variable"
	"net/http"
	"sync"
	time2 "time"
)

var (
	threadMutex sync.Mutex
	thread      = make(map[string]map[string]*RunInfo)
	taskInfo    *TaskInfo
	interval    = int64(1000 * 60)
)

type RunInfo struct {
	Key string `json:"key"`
	Ts  int64  `json:"ts"`
}

type TaskInfo struct {
	CoreThreadCount int    `json:"core-thread-count"`
	RARMD5          string `json:"rar-md5"`
}

func TaskInit() map[string]func(http.ResponseWriter, *http.Request) {
	fun := make(map[string]func(http.ResponseWriter, *http.Request))
	fun["/task-get"] = Get
	fun["/task-confirm"] = Confirm
	fun["/task-complete"] = Complete
	fun["/task-discover"] = Discover
	fun["/task-file-info"] = FileInfo
	fun["/task-file-info-over-view"] = FileInfoOverView
	fun["/task-gen-file"] = GenFile
	fun["/task-download-rar-file"] = DownloadRARFile
	fun["/mining-info"] = MiningInfo
	fun["/mining-run-report"] = MiningRunReport
	fun["/mining-run-state"] = MiningRunState

	go checkThread()

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

func FileInfoOverView(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.FileInfoOverView()).Bytes())
}

func GenFile(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.SuccessBytes())
}

func DownloadRARFile(response http.ResponseWriter, request *http.Request) {
	response.Write(variable.FileWork.DownloadRARFile())
}

func MiningInfo(response http.ResponseWriter, request *http.Request) {
	if nil == taskInfo {
		taskInfo = &TaskInfo{
			CoreThreadCount: variable.Conf.CoreThreadCount,
			RARMD5:          variable.FileWork.RARFileMD5(),
		}
	}
	response.Write(dto.Success().SetData(taskInfo).Bytes())
}

func MiningRunReport(response http.ResponseWriter, request *http.Request) {
	addThread(request.URL.Query().Get("ip"), request.URL.Query().Get("group"), request.URL.Query().Get("key"))
	response.Write(dto.Success().SetData(true).Bytes())
}

func MiningRunState(response http.ResponseWriter, request *http.Request) {
	data := make(map[string]interface{})
	threadMutex.Lock()
	ipCount := 0
	groupCount := 0
	for _, groups := range thread {
		ipCount++
		groupCount += len(groups)
	}
	data["server-count"] = ipCount
	data["thread-count"] = groupCount
	data["thread"] = thread
	bytes := dto.Success().SetData(data).Bytes()
	threadMutex.Unlock()

	response.Write(bytes)
}

func checkThread() {
	for {
		time2.Sleep(time2.Minute)
		ts := time.TimestampNowMs()
		threadMutex.Lock()
		for _, groups := range thread {
			for group, info := range groups {
				if (ts - info.Ts) > interval {
					delete(groups, group)
				}
			}
		}
		threadMutex.Unlock()
	}
}

func addThread(ip string, group string, key string) {
	if "" != ip && "" != group {
		threadMutex.Lock()
		defer threadMutex.Unlock()
		groups, ok := thread[ip]
		if !ok {
			thread[ip] = make(map[string]*RunInfo)
			groups, _ = thread[ip]
		}
		info, ok := groups[group]
		if !ok {
			groups[group] = &RunInfo{
				Key: "",
				Ts:  0,
			}
			info, _ = groups[group]
		}
		info.Key = key
		info.Ts = time.TimestampNowMs()
	}
}

func removeThread(name string) {
	if "" != name {
		threadMutex.Lock()
		defer threadMutex.Unlock()
		delete(thread, name)
	}
}
