package controller

import (
	"dispatch/dto"
	"dispatch/time"
	"dispatch/variable"
	"net/http"
	"runtime"
	"strconv"
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
	OS              string `json:"os"`
	CoreThreadCount int    `json:"core-thread-count"`
	ReportInterval  int    `json:"report-interval"`
	RarFilePath     string `json:"rar-file-path"`
	RarFilePathMD5  string `json:"rar-file-path-md5"`
	ProgramPath     string `json:"program-path"`
	ProgramPathMD5  string `json:"program-path-md5"`
}

func TaskInit() map[string]func(http.ResponseWriter, *http.Request) {
	fun := make(map[string]func(http.ResponseWriter, *http.Request))
	fun["/task-test"] = Test
	fun["/task-get"] = Get
	fun["/task-confirm"] = Confirm
	fun["/task-complete"] = Complete
	fun["/task-discover"] = Discover
	fun["/task-file-info"] = FileInfo
	fun["/task-file-info-over-view"] = FileInfoOverView
	fun["/task-gen-file"] = GenFile
	fun["/task-download-file"] = DownloadRARFile
	fun["/mining-info"] = MiningInfo
	fun["/mining-result"] = MiningResult
	fun["/mining-run-report"] = MiningRunReport
	fun["/mining-run-state"] = MiningRunState

	go checkThread()

	return fun
}

func Test(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Test()).Bytes())
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
	response.Write(dto.Success().SetData(variable.FileWork.Discover(request.URL.Query().Get("group"))).Bytes())
}

func FileInfo(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.FileInfo()).Bytes())
}

func FileInfoOverView(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.FileInfoOverView()).Bytes())
}

func GenFile(response http.ResponseWriter, request *http.Request) {
	count, err := strconv.Atoi(request.URL.Query().Get("count"))
	if nil != err || count < 1 {
		count = 1
	}
	for count > 0 {
		variable.FileWork.GenFile()
		count--
	}
	response.Write(dto.SuccessBytes())
}

func DownloadRARFile(response http.ResponseWriter, request *http.Request) {
	response.Write(variable.FileWork.DownloadFile(request.URL.Query().Get("md5")))
}

func MiningInfo(response http.ResponseWriter, request *http.Request) {
	if nil == taskInfo {
		taskInfo = &TaskInfo{
			OS:              runtime.GOOS,
			CoreThreadCount: variable.Conf.CoreThreadCount,
			ReportInterval:  variable.Conf.ReportInterval,
			RarFilePath:     "./f",
			RarFilePathMD5:  variable.FileWork.RARFileMD5(),
			ProgramPath:     "./u",
			ProgramPathMD5:  variable.FileWork.ProgramFileMD5(),
		}
	}
	response.Write(dto.Success().SetData(taskInfo).Bytes())
}

func MiningResult(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Result()).Bytes())
}

func MiningRunReport(response http.ResponseWriter, request *http.Request) {
	addThread(request.URL.Query().Get("ip"), request.URL.Query().Get("group"), request.URL.Query().Get("index"))
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

func addThread(ip string, group string, index string) {
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
		info.Key = index
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
