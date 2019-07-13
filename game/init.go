package game

import (
	"github.com/0990/goserver/server"
)

var Server *server.Server

func Init(serverID int32) error {
	s, err := server.NewServer(serverID)
	if err != nil {
		return err
	}
	Server = s
	registerHandler()
	return nil
}

func Run() {
	Server.Run()
}
