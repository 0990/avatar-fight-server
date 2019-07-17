package main

import (
	"fmt"
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/game"
	"os"
	"os/signal"
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c)

	err := game.Init(conf.GameServerID)
	if err != nil {
		panic(err)
	}
	game.Run()

	s := <-c
	fmt.Println("Got signal:", s)
}
