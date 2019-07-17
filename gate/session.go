package gate

import "github.com/0990/goserver/network"

type SessionMgr struct {
	sesID2Session map[int32]*Session
}

type Session struct {
	sesID   int32
	userID  uint64
	session network.Session
	logined bool
}

func newSessionMgr() *SessionMgr {
	return &SessionMgr{
		sesID2Session: map[int32]*Session{},
	}
}

func (p *SessionMgr) SetSessionLogined(sesID int32, userID uint64) {
	if session, exist := p.sesID2Session[sesID]; exist {
		session.userID = userID
		session.logined = true
	}
}
