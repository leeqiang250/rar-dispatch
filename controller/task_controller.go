package controller

import (
	"dispatch/conf"
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
	thread      = make(map[string]*ServerRunInfo2)
	taskInfo    *TaskInfo
	interval    = int64(1000 * 60)
)

type GroupRunInfo struct {
	Group string `json:"group"`
	Ts    int64  `json:"ts"`
}

type ServerRunInfo2 struct {
	ServerRunInfo1
	GroupRunInfo map[string]*GroupRunInfo `json:"group-run-info"`
}

type ServerRunInfo1 struct {
	TotalPWD int64 `json:"total-pwd"`
	TotalTs  int64 `json:"total-ts"`
	Speed    int64 `json:"speed"`
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
	fun["/task-cancel"] = Cancel
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
	size, errSize := strconv.ParseInt(request.URL.Query().Get("size"), 10, 64)
	ts, errTs := strconv.ParseInt(request.URL.Query().Get("ts"), 10, 64)
	if nil != errSize || nil != errTs {
		size = 0
		ts = 0
	}
	removeThread(request.URL.Query().Get("ip"), request.URL.Query().Get("group"), size, ts)
	response.Write(dto.Success().SetData(variable.FileWork.Complete(request.URL.Query().Get("group"))).Bytes())
}

func Discover(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Discover(request.URL.Query().Get("group"))).Bytes())
}

func Cancel(response http.ResponseWriter, request *http.Request) {
	response.Write(dto.Success().SetData(variable.FileWork.Cancel(request.URL.Query().Get("group"))).Bytes())
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
			CoreThreadCount: conf.Conf.CoreThreadCount,
			ReportInterval:  conf.Conf.ReportInterval,
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
	detail := request.URL.Query().Get("detail")
	data := make(map[string]interface{})
	threadMutex.Lock()
	ipCount := 0
	groupCount := 0
	serverRunInfo1 := make(map[string]*ServerRunInfo1)
	for ip, serverRunInfo := range thread {
		ipCount++
		groupCount += len(serverRunInfo.GroupRunInfo)
		serverRunInfo1[ip] = &ServerRunInfo1{
			TotalPWD: serverRunInfo.TotalPWD,
			TotalTs:  serverRunInfo.TotalTs,
			Speed:    serverRunInfo.Speed,
		}
	}
	data["total-server-count"] = ipCount
	data["total-thread-count"] = groupCount
	if detail == "1" {
		data["thread"] = thread
	} else {
		data["thread"] = serverRunInfo1
	}

	bytes := dto.Success().SetData(data).Bytes()
	threadMutex.Unlock()

	response.Write(bytes)
}

func checkThread() {
	for {
		time2.Sleep(time2.Minute)
		ts := time.TimestampNowMs()
		threadMutex.Lock()
		for _, serverRunInfo := range thread {
			for group, info := range serverRunInfo.GroupRunInfo {
				if (ts - info.Ts) > interval {
					delete(serverRunInfo.GroupRunInfo, group)
				}
			}
		}
		threadMutex.Unlock()
	}
}

func addThread(ip string, group string, index string) {
	if "" != ip && "" != group {
		threadMutex.Lock()
		serverRunInfo, ok := thread[ip]
		if !ok {
			serverRunInfo = &ServerRunInfo2{}
			serverRunInfo.TotalPWD = 0
			serverRunInfo.TotalTs = 0
			serverRunInfo.Speed = 0
			serverRunInfo.GroupRunInfo = make(map[string]*GroupRunInfo)
			thread[ip] = serverRunInfo
		}
		groupRunInfo, ok := serverRunInfo.GroupRunInfo[group]
		if !ok {
			serverRunInfo.GroupRunInfo[group] = &GroupRunInfo{}
			groupRunInfo, _ = serverRunInfo.GroupRunInfo[group]
		}
		groupRunInfo.Group = index
		groupRunInfo.Ts = time.TimestampNowMs()
		threadMutex.Unlock()
	}
}

func removeThread(ip string, group string, size int64, ts int64) {
	if "" != ip && "" != group {
		threadMutex.Lock()
		serverRunInfo, ok := thread[ip]
		if !ok {
			serverRunInfo = &ServerRunInfo2{}
			serverRunInfo.TotalPWD = 0
			serverRunInfo.TotalTs = 0
			serverRunInfo.Speed = 0
			serverRunInfo.GroupRunInfo = make(map[string]*GroupRunInfo)
			thread[ip] = serverRunInfo
		}
		serverRunInfo.TotalPWD += size
		serverRunInfo.TotalTs += ts
		if serverRunInfo.TotalTs > 0 {
			serverRunInfo.Speed = serverRunInfo.TotalPWD / serverRunInfo.TotalTs
		} else {
			serverRunInfo.Speed = 0
		}
		delete(serverRunInfo.GroupRunInfo, group)
		threadMutex.Unlock()
	}
}
