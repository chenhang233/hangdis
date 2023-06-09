package utils

import (
	"math/rand"
	"strings"
	"time"
)

const words = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM"

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
