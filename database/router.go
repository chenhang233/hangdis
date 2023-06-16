package database

import "strings"

const (
	flagWrite    = 0
	flagReadOnly = 1
)

var (
	systemTable = make(map[string]*systemCommand)
)

type systemCommand struct {
	executor SysExecFunc
	flag     int
}

func RegisterSystemCommand(name string, executor SysExecFunc) {
	name = strings.ToLower(name)
	systemTable[name] = &systemCommand{
		executor: executor,
	}
}
