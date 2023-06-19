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

func GetExpireTaskName(key string) string {
	return "expire:" + key
}
