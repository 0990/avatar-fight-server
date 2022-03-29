package util

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"runtime/debug"
)

func Recover() {
	if e := recover(); e != nil {
		stack := debug.Stack()
		logrus.WithFields(logrus.Fields{
			"err":   e,
			"stack": string(stack),
		}).Error("Recover")

		os.Stderr.Write([]byte(fmt.Sprintf("%v\n", e)))
		os.Stderr.Write(stack)
	}
}

// SafeGo go
func SafeGo(f func()) {
	if f != nil {
		go func() {
			defer Recover()
			f()
		}()
	}
}
