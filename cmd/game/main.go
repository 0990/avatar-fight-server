package main

import (
	"flag"
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/game"
	"github.com/0990/goserver"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var gosconf = flag.String("goserver", "default", "goserver config file")

func main() {

	gosconf, err := goserver.ReadConfig(*gosconf)
	if err != nil {
		logrus.Fatal("readconf", err)
	}
	c := make(chan os.Signal, 1)
	signal.Notify(c)

	err = game.Init(conf.GameServerID, *gosconf)
	if err != nil {
		logrus.WithError(err).Fatal("gosconf", gosconf)
	}
	game.Run()

	s := <-c
	logrus.Info("Got signal:", s)
}
