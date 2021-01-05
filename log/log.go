package log

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/ioutil"
	"log"
)

var (
	Trace *log.Logger //0
	Debug *log.Logger //2
	Info  *log.Logger //4
	Warn  *log.Logger //6
	Error *log.Logger //8
)

func Init(path string, level int) (ok bool) {
	var traceHandle, debugHandle, infoHandle, warnHandle, errorHandle io.Writer

	if level <= 0 {
		traceHandle = &lumberjack.Logger{
			Filename:   path + "/trace.log",
			MaxSize:    500,
			MaxBackups: 20,
			MaxAge:     7,
			Compress:   false,
		}
	} else {
		traceHandle = ioutil.Discard
	}

	if level <= 2 {
		debugHandle = &lumberjack.Logger{
			Filename:   path + "/debug.log",
			MaxSize:    500,
			MaxBackups: 20,
			MaxAge:     7,
			Compress:   false,
		}
	} else {
		debugHandle = ioutil.Discard
	}

	if level <= 4 {
		infoHandle = &lumberjack.Logger{
			Filename:   path + "/info.log",
			MaxSize:    500,
			MaxBackups: 20,
			MaxAge:     7,
			Compress:   false,
		}
	} else {
		infoHandle = ioutil.Discard
	}

	if level <= 6 {
		warnHandle = &lumberjack.Logger{
			Filename:   path + "/warn.log",
			MaxSize:    500,
			MaxBackups: 20,
			MaxAge:     7,
			Compress:   false,
		}
	} else {
		warnHandle = ioutil.Discard
	}

	if level <= 8 {
		errorHandle = &lumberjack.Logger{
			Filename:   path + "/error.log",
			MaxSize:    500,
			MaxBackups: 20,
			MaxAge:     7,
			Compress:   false,
		}
	} else {
		errorHandle = ioutil.Discard
	}

	Trace = log.New(traceHandle, "trace: ", log.Ldate|log.Ltime|log.Lshortfile)
	Debug = log.New(debugHandle, "debug: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(infoHandle, "info: ", log.Ldate|log.Ltime|log.Lshortfile)
	Warn = log.New(warnHandle, "warn: ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle, "error: ", log.Ldate|log.Ltime|log.Lshortfile)

	return true
}
