package work

import (
	"bufio"
	"crypto/md5"
	"dispatch/conf"
	"dispatch/log"
	"dispatch/time"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	time2 "time"
)

type FileWork struct {
	mutex          sync.Mutex
	fileSplitMutex sync.Mutex
	data           map[string]int64
	rar2MD5        string
	programMD5     string
	file           map[string][]byte
	result         map[string]bool
}

type File struct {
	Name string `json:"group"`
	Text string `json:"text"`
}

const (
	PasswordPath = "./data/"
	CompletePath = "./complete/"
	LogPath      = "./log/"
	Test         = "./test.txt"
	Pwd          = ".pwd"
	Waiting      = ".waiting"
	Confirming   = ".confirming"
	Processing   = ".processing"
	Complete     = ".complete"
	Right        = ".right"
	Tmp          = ".tmp"

	Key              = ".key"
	RARFile          = "./data.rar"
	ProgramFileLinux = "./unrarlinux"
	ProgramFileMacOS = "./unrarmacos"

	FileStateWaitingInterval = int64(1000 * 60)
)

func NewFileWork() *FileWork {
	work := FileWork{}
	work.data = make(map[string]int64)
	work.file = make(map[string][]byte)
	work.result = make(map[string]bool)
	work.LoadFile()
	return &work
}

func (this *FileWork) Run() {
	{
		err := os.Mkdir(PasswordPath, 0777)
		if nil != err && !os.IsExist(err) {
			log.Error.Println("os.Mkdir", err)
			os.Exit(0)
		}
		err = os.Mkdir(CompletePath, 0777)
		if nil != err && !os.IsExist(err) {
			log.Error.Println("os.Mkdir", err)
			os.Exit(0)
		}
		err = os.Mkdir(LogPath, 0777)
		if nil != err && !os.IsExist(err) {
			log.Error.Println("os.Mkdir", err)
			os.Exit(0)
		}
	}

	this.CancelAll()

	go func() {
		for {
			this.CheckAndCancel()
			time2.Sleep(time2.Minute)
		}
	}()

	//go func() {
	//	for {
	//		if !this.ProductFile() {
	//			time2.Sleep(time2.Minute)
	//		}
	//	}
	//}()
}

func (this *FileWork) Test() *File {
	file := File{
		Name: uuid.NewV4().String(),
		Text: "",
	}

	text, err := ioutil.ReadFile(Test)
	if nil == err {
		file.Text = string(text)
	}

	return &file
}

func (this *FileWork) Get() *File {
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		group := ""
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), Waiting) && this.isSmallSize(file) {
				group = file.Name()
				break
			}
		}

		if "" == group {
			go this.SplitFile()
		} else {
			text, err := ioutil.ReadFile(PasswordPath + group)
			if nil == err {
				err = os.Rename(PasswordPath+group, PasswordPath+group+Confirming)
				if nil == err {
					this.mutex.Lock()
					_, ok := this.data[group]
					if !ok {
						this.data[group] = time.TimestampNowMs()
					}
					this.mutex.Unlock()
					if ok {
						return nil
					}
					log.Info.Println("FileWork Get", group)
					return &File{
						Name: group,
						Text: string(text),
					}
				} else {
					log.Error.Println("FileWork Get", group, err)
				}
			} else {
				log.Error.Println("FileWork Get", group, err)
			}
		}
	} else {
		log.Error.Println("FileWork Get", err)
	}

	return nil
}

func (this *FileWork) Confirm(group string) bool {
	err := os.Rename(PasswordPath+group+Confirming, PasswordPath+group+Processing)
	if nil == err {
		this.mutex.Lock()
		delete(this.data, group)
		this.mutex.Unlock()
		log.Info.Println("FileWork Confirm", group)
	} else {
		log.Error.Println("FileWork Confirm", group, err)
	}

	return nil == err
}

func (this *FileWork) Complete(group string) bool {
	if !this.IsExist(PasswordPath + group + Processing) {
		log.Error.Println("FileWork Complete", group)
		return true
	}

	err := os.Remove(PasswordPath + group + Processing)
	if nil != err {
		log.Error.Println("FileWork Complete", group, err)
	}

	return true
}

func (this *FileWork) ProduceFile() bool {
	result := false
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() {
				group := file.Name()
				if strings.HasSuffix(group, Pwd) {
					group = file.Name()[:len(group)-len(Pwd)]
					if "" != group {
						err := os.Rename(PasswordPath+file.Name(), PasswordPath+group)
						if nil == err {
							log.Info.Println("FileWork CancelAll", group)
						} else {
							log.Error.Println("FileWork CancelAll", group, err)
						}
					}
				}
			}
		}
	} else {
		log.Error.Println("FileWork ProductFile", err)
		time2.Sleep(time2.Second)
	}

	return result
}

func (this *FileWork) CheckAndCancel() {
	ts := time.TimestampNowMs()

	this.mutex.Lock()

	for k, v := range this.data {
		if (ts - v) > FileStateWaitingInterval {
			err := os.Rename(PasswordPath+k+Confirming, PasswordPath+k)
			if nil == err {
				delete(this.data, k)
				log.Info.Println("FileWork Cancel", k)
			} else {
				log.Error.Println("FileWork Cancel", k, err)
			}
		}
	}

	this.mutex.Unlock()
}

func (this *FileWork) CancelAll() {
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() {
				group := file.Name()
				if strings.HasSuffix(group, Confirming) {
					group = file.Name()[:len(group)-len(Confirming)]
					if "" != group {
						err := os.Rename(PasswordPath+file.Name(), PasswordPath+group)
						if nil == err {
							log.Info.Println("FileWork CancelAll", group)
						} else {
							log.Error.Println("FileWork CancelAll", group, err)
						}
					}
				} else if strings.HasSuffix(group, Processing) {
					group = file.Name()[:len(group)-len(Processing)]
					if "" != group {
						err := os.Rename(PasswordPath+file.Name(), PasswordPath+group)
						if nil == err {
							log.Info.Println("FileWork CancelAll", group)
						} else {
							log.Error.Println("FileWork CancelAll", group, err)
						}
					}
				} else if strings.HasSuffix(group, Right) {
					this.mutex.Lock()
					this.result[group] = true
					this.mutex.Unlock()
				}
			}
		}
	} else {
		log.Error.Println("FileWork CancelAll", err)
	}
}

func (this *FileWork) FileInfo() map[string]map[string]int64 {
	waiting := make(map[string]int64)
	confirming := make(map[string]int64)
	processing := make(map[string]int64)
	complete := make(map[string]int64)

	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() {
				if strings.HasSuffix(file.Name(), Waiting) {
					waiting[file.Name()] = file.Size()
				} else if strings.HasSuffix(file.Name(), Confirming) {
					confirming[file.Name()] = file.Size()
				} else if strings.HasSuffix(file.Name(), Processing) {
					processing[file.Name()] = file.Size()
				} else if strings.HasSuffix(file.Name(), Complete) {
					complete[file.Name()] = file.Size()
				}
			}
		}
	} else {
		log.Error.Println("FileWork FileInfo", err)
	}

	data := make(map[string]map[string]int64)
	data["waiting"] = waiting
	data["confirming"] = confirming
	data["processing"] = processing
	data["complete"] = complete

	return data
}

func (this *FileWork) FileInfoOverView() map[string]interface{} {
	waiting := 0
	confirming := 0
	processing := 0
	complete := 0

	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() {
				if strings.HasSuffix(file.Name(), Waiting) {
					waiting++
				} else if strings.HasSuffix(file.Name(), Confirming) {
					confirming++
				} else if strings.HasSuffix(file.Name(), Processing) {
					processing++
				} else if strings.HasSuffix(file.Name(), Complete) {
					complete++
				}
			}
		}
	} else {
		log.Error.Println("FileWork FileInfo", err)
	}

	total := waiting + confirming + processing + complete

	waitingPercent := 0
	confirmingPercent := 0
	processingPercent := 0
	completePercent := 0
	if total > 0 {
		waitingPercent = (waiting * 100) / total
		confirmingPercent = (confirming * 100) / total
		processingPercent = (processing * 100) / total
		completePercent = 100 - waitingPercent - confirmingPercent - processingPercent
	}
	//waitingPercent := (waiting * 100) / total
	//confirmingPercent := (confirming * 100) / total
	//processingPercent := (processing * 100) / total
	//completePercent := 100 - waitingPercent - confirmingPercent - processingPercent

	data := make(map[string]interface{})
	data["total"] = total
	data["waiting"] = waiting
	data["waiting-percent"] = strconv.Itoa(waitingPercent) + "%"
	data["confirming"] = confirming
	data["confirming-percent"] = strconv.Itoa(confirmingPercent) + "%"
	data["processing"] = processing
	data["processing-percent"] = strconv.Itoa(processingPercent) + "%"
	data["complete"] = complete
	data["complete-percent"] = strconv.Itoa(completePercent) + "%"

	return data
}

func (this *FileWork) GenFile() {
	group := uuid.NewV4().String()
	file, err := os.OpenFile(PasswordPath+group+Tmp, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0777)
	if nil == err {
		defer file.Close()

		buf := bufio.NewWriter(file)

		content := conf.Conf.StandardFileSize
		for content > 0 {
			content--
			buf.WriteString("," + strconv.Itoa(content))
			buf.Flush()
		}

		buf.Flush()

		err = os.Rename(PasswordPath+group+Tmp, PasswordPath+group+Waiting)
		if nil != err {
			log.Error.Println("FileWork GenFile", err)
		}
	} else {
		log.Error.Println("FileWork GenFile", err)
	}
}

func (this *FileWork) Discover(group string) bool {
	if "" != group {
		this.mutex.Lock()
		this.result[group] = true
		this.mutex.Unlock()

		log.Info.Println("FileWork Discover", group)

		this.IsExist(PasswordPath + group + Processing)

		text, err := ioutil.ReadFile(PasswordPath + group + Processing)
		if nil == err {
			file, err := os.OpenFile("./"+group+Right, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0777)
			if nil == err {
				defer file.Close()

				buf := bufio.NewWriter(file)
				buf.WriteString("\n")
				buf.Write(text)
				buf.Flush()

				return true
			} else {
				log.Error.Println("FileWork Discover", err)
			}
		} else {
			log.Error.Println("FileWork Discover", err)
		}
	}

	return false
}

func (this *FileWork) Cancel(group string) bool {
	if "" != group {
		err := os.Rename(PasswordPath+group+Processing, PasswordPath+group)
		if nil != err {
			log.Error.Println("FileWork Cancel", group, err)
		}

		err = os.Rename(PasswordPath+group+Confirming, PasswordPath+group)
		if nil != err {
			log.Error.Println("FileWork Cancel", group, err)
		}

		this.mutex.Lock()
		delete(this.data, group)
		this.mutex.Unlock()

		return nil == err
	}

	return false
}

func (this *FileWork) Result() []string {
	result := make([]string, 0)
	this.mutex.Lock()
	for k, _ := range this.result {
		result = append(result, k)
	}
	this.mutex.Unlock()

	return result
}

func (this *FileWork) RARFileMD5() string {
	return this.rar2MD5
}

func (this *FileWork) ProgramFileMD5() string {
	return this.programMD5
}

func (this *FileWork) DownloadFile(md5 string) []byte {
	return this.file[md5]
}

func (this *FileWork) LoadFile() {
	var names []string
	if "linux" == runtime.GOOS {
		names = []string{RARFile, ProgramFileLinux}
	} else {
		names = []string{RARFile, ProgramFileMacOS}
	}
	for _, name := range names {
		content := this.ReadFile(name)
		if nil != content {
			this.file[fmt.Sprintf("%x", md5.Sum(content))] = content
			if name == names[0] {
				this.rar2MD5 = fmt.Sprintf("%x", md5.Sum(content))
			} else if name == names[1] {
				this.programMD5 = fmt.Sprintf("%x", md5.Sum(content))
			}
		}
	}
}

func (this *FileWork) ReadFile(path string) []byte {
	content, err := ioutil.ReadFile(path)
	if err == nil {
		return content
	} else {
		log.Error.Println("FileWork ReadFile", path, err)
		return nil
	}
}

func (this *FileWork) isSmallSize(file os.FileInfo) bool {
	return int64(conf.Conf.StandardFileSize*2) > file.Size()
}

func (this *FileWork) SplitFile() {
	this.fileSplitMutex.Lock()
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err && len(files) > 0 {
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), Waiting) && !this.isSmallSize(file) {
				text, err := ioutil.ReadFile(PasswordPath + file.Name())
				if nil == err {
					char := ([]byte(","))[0]
					l := len(text)
					count := 0
					data := make([]byte, 0, conf.Conf.StandardFileSize)

					for i := 0; i < l; i++ {
						if text[i] == char && len(data) >= conf.Conf.StandardFileSize {
							for {
								if this.WriteFile(PasswordPath+strconv.Itoa(count)+"-"+file.Name(), data) {
									count++
									data = make([]byte, 0, conf.Conf.StandardFileSize)
									break
								} else {
									time2.Sleep(time2.Second)
								}
							}
						} else {
							data = append(data, text[i])
						}
					}
					if 0 < len(data) {
						for {
							if this.WriteFile(PasswordPath+strconv.Itoa(count)+"-"+file.Name(), data) {
								count++
								data = make([]byte, 0, conf.Conf.StandardFileSize)
								break
							} else {
								time2.Sleep(time2.Second)
							}
						}
					}
					os.Remove(PasswordPath + file.Name())
				} else {
					log.Error.Println("FileWork SplitFile", err)
				}

				break
			}
		}
	}
	this.fileSplitMutex.Unlock()
}

func (this *FileWork) WriteFile(filepath string, data []byte) bool {
	file, err := os.OpenFile(filepath+Tmp, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0777)
	if nil != err {
		log.Error.Println("FileWork WriteFile", err)
		return false
	}

	defer file.Close()

	buf := bufio.NewWriter(file)

	_, err = buf.Write(data)
	if nil != err {
		os.Remove(filepath + Tmp)
		log.Error.Println("FileWork WriteFile", err)
		return false
	}

	err = buf.Flush()
	if nil != err {
		os.Remove(filepath + Tmp)
		log.Error.Println("FileWork WriteFile", err)
		return false
	}

	err = file.Close()
	if nil != err {
		os.Remove(filepath + Tmp)
		log.Error.Println("FileWork WriteFile", err)
		return false
	}

	err = os.Rename(filepath+Tmp, filepath)
	if nil != err {
		os.Remove(filepath + Tmp)
		log.Error.Println("FileWork WriteFile", err)
		return false
	}

	return true
}

func (this *FileWork) IsExist(file string) bool {
	_, err := os.Stat(file)
	if nil != err {
		log.Error.Println("FileWork IsExist", file, err)
	}

	return nil == err
}
