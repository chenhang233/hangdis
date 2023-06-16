package config

import (
	"fmt"
	"hangdis/utils"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFileNameABS(t *testing.T) {
	abs, err := filepath.Abs("hangdis.conf")
	if err != nil {
		panic(err)
	}
	println(abs)
}

func TestRand(t *testing.T) {
	println(utils.RandomUUID())
}

func TestStrTrimLeft(t *testing.T) {
	left := strings.TrimLeft("     # 你好", " ")
	fmt.Println(left)
}

func TestSys(t *testing.T) {
	fmt.Println(runtime.GOOS, runtime.GOARCH)
}
