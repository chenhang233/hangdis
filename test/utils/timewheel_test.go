package utils

import (
	"fmt"
	"hangdis/utils"
	"hangdis/utils/timewheel"
	"testing"
	"time"
)

func TestAt(t *testing.T) {
	times := make([]time.Time, 100)
	names := make([]string, 100)
	for i := 0; i < 10; i++ {
		time.Sleep(time.Millisecond * 200)
		add := time.Now().Add(time.Second * 5)
		name := utils.GetExpireTaskName(utils.RandomUUID())
		names[i] = name
		times[i] = add
		fmt.Println("run: ", i)
	}
	fmt.Println("--------------------------------------------------------")
	for i := 0; i < 10; i++ {
		name := names[i]
		t2 := times[i]
		fmt.Println(name, t2, "--test")
		go func() {
			timewheel.At(t2, name, func() {
				fmt.Println("my name is ", name)
			})
		}()

	}

	for {
		time.Sleep(time.Second)
	}
}
