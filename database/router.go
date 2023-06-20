package database

import "strings"

const (
	flagWrite    = 0
	flagReadOnly = 1
)

type systemCommand struct {
	executor SysExecFunc
	flag     int
}

type command struct {
	executor ExecFunc
	prepare  PreFunc
	undo     UndoFunc // Regression function
	arity    int      // arity < 0 means len(args) >= -arity
	flags    int      //  Read only or not
}

var (
	systemTable = make(map[string]*systemCommand)
	cmdTable    = make(map[string]*command)
)

func RegisterSystemCommand(name string, executor SysExecFunc) {
	name = strings.ToLower(name)
	systemTable[name] = &systemCommand{
		executor: executor,
	}
}

func RegisterCommand(name string, executor ExecFunc, prepare PreFunc, rollback UndoFunc, arity int, flags int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		executor: executor,
		prepare:  prepare,
		undo:     rollback,
		arity:    arity,
		flags:    flags,
	}
}
