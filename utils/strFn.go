package utils

import (
	"math/rand"
	"strings"
	"time"
)

const words = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

const defaultConn = 10

func RandomUUID() string {
	nR := rand.New(rand.NewSource(time.Now().UnixNano()))
	wlen := len(words)
	sb := strings.Builder{}
	for i := 0; i < 20; i++ {
		u := words[nR.Intn(wlen-1)]
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
