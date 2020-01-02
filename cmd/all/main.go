package main

import (
	"flag"
	"fmt"
	"github.com/0990/avatar-fight-server/center"
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/game"
	"github.com/0990/avatar-fight-server/gate"
	"github.com/0990/goserver"
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var addr = flag.String("addr", "0.0.0.0:9000", "http service address")
var pprofAddr = flag.String("pprofAddr", "0.0.0.0:9900", "http pprof service address")
var gosconf = flag.String("goserver", "", "goserver config file")

//TODO 加woker性能监控和运行时堆栈打印

func main() {
	flag.Parse()
	gosconf, err := goserver.ReadConfig(*gosconf)
	if err != nil {
		logrus.Fatal("readconfig ", err)
	}
	go func() {
		http.ListenAndServe(*pprofAddr, nil)
	}()
	//center
	err = center.Init(conf.CenterServerID, *gosconf)
	if err != nil {
		logrus.WithError(err).Fatal("gosconf", gosconf)
	}
	center.Run()

	//gate
	err = gate.Init(conf.GateServerID, *addr, *gosconf)
	if err != nil {
		logrus.WithError(err).Fatal("gosconf", gosconf)
	}
	gate.Run()

	//game
	err = game.Init(conf.GameServerID, *gosconf)
	if err != nil {
		logrus.WithError(err).Fatal("gosconf", gosconf)
	}
	game.Run()

	logrus.Info("start success...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	logrus.Info("Got signal:", s)
}

func setupSigusr1Trap() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		for range c {
			DumpStacks()
		}
	}()
}
func DumpStacks() {
	buf := make([]byte, 16384)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Printf("=== BEGIN goroutine stack dump ===\n%s\n=== END goroutine stack dump ===", buf)
}
