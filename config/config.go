package config

import (
	"bufio"
	"hangdis/utils"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type ServerProperties struct {
	RuntimeID         string
	Bind              string `conf:"bind"`
	Port              uint16 `conf:"port"`
	BindAddr          string
	MaxClients        int `conf:"maxClients"`
	AbsPath           string
	AppendOnly        bool   `conf:"appendonly"`
	AppendFilename    string `conf:"appendfilename"`
	AppendFsync       string `conf:"appendfsync"`
	AofUseRdbPreamble bool   `conf:"aof-use-rdb-preamble"`
	RDBFilename       string `conf:"dbfilename"`
	Databases         int    `conf:"databases"`
	Password          string `conf:"requirepass"`
	Dir               string `conf:"dir"`
}

var Properties *ServerProperties

func parse(file *os.File) *ServerProperties {
	config := &ServerProperties{}
	scanner := bufio.NewScanner(file)
	mc := make(map[string]string)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimLeft(line, " ")[0] == '#' {
			continue
		}
		sList := strings.Split(line, " ")
		if len(sList) != 2 {
			continue
		}
		mc[sList[0]] = sList[1]
	}
	vof := reflect.ValueOf(config)
	tof := reflect.TypeOf(config)
	n := tof.Elem().NumField()
	for i := 0; i < n; i++ {
		fieldValue := vof.Elem().Field(i)
		field := tof.Elem().Field(i)
		key, ok := field.Tag.Lookup("conf")
		if !ok || strings.TrimLeft(key, " ") == "" {
			key = field.Name
		}
		v, ok := mc[key]
		if ok {
			switch field.Type.Kind() {
			case reflect.String:
				fieldValue.SetString(v)
			case reflect.Int:
				intValue, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					fieldValue.SetInt(intValue)
				}
			case reflect.Uint16:
				uintValue, err := strconv.ParseUint(v, 10, 16)
				if err == nil {
					fieldValue.SetUint(uintValue)
				}
			case reflect.Bool:
				flag := v == "true"
				fieldValue.SetBool(flag)
			}
		}
	}
	config.BindAddr = config.Bind + ":" + strconv.Itoa(int(config.Port))
	return config
}

func SetupConfig(configName string) {
	file, err := os.Open(configName)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	Properties = parse(file)
	Properties.RuntimeID = utils.RandomUUID()
	abs, err := filepath.Abs(configName)
	if err != nil {
		panic(err)
	}
	Properties.AbsPath = abs
}

func GetTmpDir() string {
	if Properties.Dir == "" {
		return "tmp"
	}
	return Properties.Dir + "/tmp"
}
