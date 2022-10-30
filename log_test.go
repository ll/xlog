package xlog

import (
	"testing"
	"time"
)

func TestInfof(t *testing.T) {
	InitLog("./log")
	for i := 0; i < 100; i++ {
		Infof("This is msg: %v", i)
	}
	time.Sleep(time.Second * 2)
	Infof("end")
}
