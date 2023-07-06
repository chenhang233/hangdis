package database

import "strings"

const (
	flagWrite    = 0
	flagReadOnly = 1
	even         = 2
	odd          = 3
)

type pubSubCommand struct {
	executor SubExecFunc
	flag     int
}

type systemCommand struct {
	executor SysExecFunc
	flag     int
}

type Command struct {
	executor ExecFunc
	prepare  PreFunc
	undo     UndoFunc // Regression function
	arity    int      // arity < 0 means len(args) >= -arity
	parity   int      //  arity even or odd
	flags    int      //  Read only or not
}

var (
	pubSubTable = make(map[string]*pubSubCommand)
	systemTable = make(map[string]*systemCommand)
	cmdTable    = make(map[string]*Command)
)

func checkSysTableExist(name string) {
	_, exist := systemTable[name]
	if exist {
		panic("systemTable[name] exist")
	}
	_, exist = cmdTable[name]
	if exist {
		panic("systemTable[name] exist")
	}
}

func RegisterPubSubCommand(name string, executor SubExecFunc, flag int) {
	name = strings.ToLower(name)
	checkSysTableExist(name)
	pubSubTable[name] = &pubSubCommand{
		executor: executor,
		flag:     flag,
	}
}

func RegisterSystemCommand(name string, executor SysExecFunc) {
	name = strings.ToLower(name)
	checkSysTableExist(name)
	systemTable[name] = &systemCommand{
		executor: executor,
		flag:     flagReadOnly,
	}
}

func RegisterCommand(name string, executor ExecFunc, prepare PreFunc, rollback UndoFunc, arity int, flags int) *Command {
	name = strings.ToLower(name)
	checkSysTableExist(name)
	cmd := &Command{
		executor: executor,
		prepare:  prepare,
		undo:     rollback,
		arity:    arity,
		flags:    flags,
		parity:   -1,
	}
	cmdTable[name] = cmd
	return cmd
}

func (c *Command) addParity(p int) *Command {
	if p != even && p != odd {
		panic("p  must be even or odd")
	}
	c.parity = p
	return c
}
