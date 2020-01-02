package center

import (
	"github.com/0990/goserver"
	"github.com/0990/goserver/server"
)

var Server *server.Server

var UMgr *UserMgr
var GMgr *GameMgr

func Init(serverID int32, config goserver.Config) error {
	s, err := server.NewServer(serverID, config)
	if err != nil {
		return err
	}
	Server = s
	registerHandler()
	GMgr = NewGameMgr()
	UMgr = NewUserMgr()
	return nil
}

func Run() {
	Server.Run()
}
