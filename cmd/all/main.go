package main

import (
	"flag"
	"fmt"
	"github.com/0990/avatar-fight-server/center"
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/game"
	"github.com/0990/avatar-fight-server/gate"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "0.0.0.0:9000", "http service address")

func main() {
	flag.Parse()

	//center
	err := center.Init(conf.CenterServerID)
	if err != nil {
		panic(err)
	}
	center.Run()

	//gate
	err = gate.Init(conf.GateServerID, *addr)
	if err != nil {
		panic(err)
	}
	gate.Run()

	//game
	err = game.Init(conf.GameServerID)
	if err != nil {
		panic(err)
	}
	game.Run()

	logrus.Info("start success...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	s := <-c
	fmt.Println("Got signal:", s)
}
