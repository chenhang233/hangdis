package logs

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

const (
	logsFilePath  = "utils/logs"
	ServerLogPath = "ServerLogPath"
)

var LogPathList []string

type LogConf struct {
	ServiceName string
	Debug       *log.Logger
	Info        *log.Logger
	Warn        *log.Logger
	Error       *log.Logger
}

var LOG *LogConf

func init() {
	l, err := LoadLog(ServerLogPath)
	if err != nil {
		panic(err)
	}
	LOG = l
}

func LoadLog(sName string) (*LogConf, error) {
	LogPathList = append([]string{}, ServerLogPath)
	if !LogPathInclude(sName) {
		return nil, errors.New("LoadLog error")
	}
	conf := &LogConf{
		ServiceName: sName,
	}
	err := handleDir()
	if err != nil {
		return nil, err
	}
	format := time.Now().Format("2006_01_02")
	logFile, err := os.OpenFile(logsFilePath+"/"+conf.ServiceName+"/"+format+".txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		return nil, err
	}
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	conf.Debug = log.New(multiWriter, "[debug]", log.Ldate|log.Ltime|log.Llongfile)
	conf.Info = log.New(multiWriter, "[info]", log.Ldate|log.Ltime|log.Llongfile)
	conf.Warn = log.New(multiWriter, "[warn]", log.Ldate|log.Ltime|log.Llongfile)
	conf.Error = log.New(multiWriter, "[error]", log.Ldate|log.Ltime|log.Llongfile)
	return conf, nil
}

func handleDir() error {
	var err error
	_, err = os.Stat(logsFilePath)
	if err != nil {
		err = os.MkdirAll(logsFilePath, os.ModePerm)
	}
	logPathMap := make(map[int]string, len(LogPathList))
	for i, item := range LogPathList {
		logPathMap[i] = logsFilePath + "/" + item
	}
	for _, m := range logPathMap {
		_, err = os.Stat(m)
		if err != nil {
			err = os.MkdirAll(m, os.ModePerm)
		}
	}
	return err
}

func LogPathInclude(path string) bool {
	for _, s := range LogPathList {
		if s == path {
			return true
		}
	}
	return false
}

func ReadLogs(path string, date string) error {
	if !LogPathInclude(path) {
		return errors.New("log path error")
	}
	p := logsFilePath + "/" + path + "/" + date + ".txt"
	ReadAllLogs(p)
	return nil
}

func ReadAllLogs(p string) {
	file, err := os.ReadFile(p)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(file))
}
