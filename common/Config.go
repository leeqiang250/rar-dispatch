package common

import (
	"dispatch/log"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

const (
	Path = "./conf/config.yaml"
)

type Config struct {
	App struct {
		Port string `yaml:"port"`
	}

	CoreThreadCount  int `yaml:"core-thread-count"`
	ReportInterval   int `yaml:"report-interval"`
	StandardFileSize int `yaml:"standard-file-size"`
}

func ConfigInit() *Config {
	file, err := ioutil.ReadFile(Path)
	if err != nil {
		log.Error.Println("Config File", Path, err)
		os.Exit(0)
	}

	conf := Config{}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		log.Error.Println("Config File", Path, err)
		os.Exit(0)
	}

	if conf.StandardFileSize < 1 {
		conf.StandardFileSize = 10240
	}

	return &conf
}
