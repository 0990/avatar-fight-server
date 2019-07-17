package main

import (
	"flag"
	"fmt"
	"github.com/0990/avatar-fight-server/center"
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/game"
	"github.com/0990/avatar-fight-server/gate"
	"os"
	"os/signal"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

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

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	s := <-c
	fmt.Println("Got signal:", s)
}
