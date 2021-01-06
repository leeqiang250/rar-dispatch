package work

import (
	"bufio"
	"crypto/md5"
	"dispatch/log"
	"dispatch/time"
	"fmt"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
	time2 "time"
)

type FileWork struct {
	mutex      sync.Mutex
	data       map[string]int64
	rarFile    []byte
	rarFileMD5 string
}

type File struct {
	Name string `json:"group"`
	Text string `json:"text"`
}

const (
	ResourcePath = "/resource/"
	PasswordPath = ResourcePath + "password/"
	Waiting      = ".waiting"
	Confirming   = ".confirming"
	Processing   = ".processing"
	Complete     = ".complete"
	Tmp          = ".tmp"

	Key     = ".key"
	RARFile = "data.rar"

	FileStateWaitingInterval = int64(1000 * 60)
)

func NewFileWork() *FileWork {
	work := FileWork{}
	work.data = make(map[string]int64)
	return &work
}

func (this *FileWork) Run() {
	this.CancelAll()

	for {
		this.Cancel()

		time2.Sleep(time2.Minute)
	}
}

func (this *FileWork) Get() *File {
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err {
		group := ""
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), Waiting) {
				group = file.Name()
				break
			}
		}
		if "" != group {
			text, err := ioutil.ReadFile(PasswordPath + group)
			if nil == err {
				err = os.Rename(PasswordPath+group, PasswordPath+group+Confirming)
				if nil == err {
					this.mutex.Lock()
					this.data[group] = time.TimestampNowMs()
					this.mutex.Unlock()
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
	err := os.Rename(PasswordPath+group+Processing, PasswordPath+group+Complete)
	if nil == err {
		log.Info.Println("FileWork Complete", group)
	} else {
		log.Error.Println("FileWork Complete", group, err)
	}

	return nil == err
}

func (this *FileWork) Cancel() {
	ts := time.TimestampNowMs()

	this.mutex.Lock()
	defer this.mutex.Unlock()

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
}

func (this *FileWork) CancelAll() {
	files, err := ioutil.ReadDir(PasswordPath)
	if nil == err {
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
	if nil == err {
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
	if nil == err {
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

	waitingPercent := (waiting * 100) / total
	confirmingPercent := (confirming * 100) / total
	processingPercent := (processing * 100) / total
	completePercent := 100 - waitingPercent - confirmingPercent - processingPercent

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
	file, err := os.OpenFile(PasswordPath+group+Tmp, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
	if nil == err {
		defer file.Close()

		buf := bufio.NewWriter(file)

		//content := 1000000
		content := 200000
		for content > 0 {
			content--
			buf.WriteString("," + strconv.Itoa(content))
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

func (this *FileWork) Discover(group string, key string) bool {
	if "" != group && "" != key {
		text := "group:" + group + ",key:" + key
		name := ResourcePath + group + key + Key
		file, err := os.OpenFile(name, os.O_WRONLY|os.O_RDONLY|os.O_CREATE|os.O_APPEND, 0666)
		if nil == err {
			defer file.Close()

			buf := bufio.NewWriter(file)
			buf.WriteString("\n")
			_, err = buf.WriteString(text)
			if nil == err {
				err = buf.Flush()
				if nil == err {
					log.Info.Println("FileWork Discover", text)
					return true
				} else {
					log.Error.Println("FileWork Discover", text, err)
				}
			} else {
				log.Error.Println("FileWork Discover", text, err)
			}
		} else {
			log.Error.Println("FileWork Discover", text, err)
		}
	}

	return false
}

func (this *FileWork) RARFileMD5() string {
	if "" == this.rarFileMD5 {
		content := this.DownloadRARFile()
		if content != nil {
			this.rarFileMD5 = fmt.Sprintf("%x", md5.Sum(content))
		}
	}

	return this.rarFileMD5
}

func (this *FileWork) DownloadRARFile() []byte {
	if nil == this.rarFile {
		content, err := ioutil.ReadFile(ResourcePath + RARFile)
		if err == nil {
			this.rarFile = content
		} else {
			log.Error.Println("FileWork DownloadRARFile", err)
		}
	}

	return this.rarFile
}
