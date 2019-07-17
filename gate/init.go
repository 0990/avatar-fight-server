package gate

import "github.com/0990/goserver/server"

var Gate *server.Gate

var SMgr *SessionMgr

func Init(serverID int32, addr string) error {
	s, err := server.NewGate(serverID, addr)
	if err != nil {
		return err
	}
	Gate = s
	registerRoute()
	registerHandler()
	SMgr = newSessionMgr()
	return nil
}

func Run() {
	Gate.Run()
}
