package admin

import (
	"github.com/0990/goserver"
	"github.com/0990/goserver/server"
)

var Server *server.Server

func Init(serverID int32, addr string, config goserver.Config) error {
	s, err := server.NewServer(serverID, config)
	if err != nil {
		return err
	}
	Server = s
	startMetrics()

	hs := newHTTPServer(addr)
	err = hs.Run()
	if err != nil {
		return err
	}

	return nil
}

func Run() {
	Server.Run()
}
