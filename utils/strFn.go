package utils

import (
	"math/rand"
	"strings"
	"time"
)

const words = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

const defaultConn = 10

func RandomWordsIndex() int {
	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	wLen := len(words)
	return nR.Intn(wLen - 1)
}

func RandomUUID() string {
	sb := strings.Builder{}
	for i := 0; i < 20; i++ {
		u := words[RandomWordsIndex()]
		sb.WriteByte(u)
	}
	return sb.String()
}

func GetConnNum(conn int) int {
	if conn <= 0 {
		return defaultConn
	}
	return conn
}

func ToCmdLine(cmd ...string) [][]byte {
	args := make([][]byte, len(cmd))
	for i, s := range cmd {
		args[i] = []byte(s)
	}
	return args
}

func ToCmdLine3(commandName string, args ...[]byte) [][]byte {
	result := make([][]byte, len(args)+1)
	result[0] = []byte(commandName)
	for i, s := range args {
		result[i+1] = s
	}
	return result
}

func GetExpireTaskName(key string) string {
	return "expire:" + key
}
