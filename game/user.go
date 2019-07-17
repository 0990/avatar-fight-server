package game

import (
	"github.com/0990/goserver/rpc"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

type AccountType int8

const (
	_ AccountType = iota
	VISITOR
	WX
	ROBOT
)

type UserMgr struct {
	ses2User    map[rpc.GateSessionID]*User
	userID2User map[uint64]*User
}

func NewUserMgr() *UserMgr {
	p := new(UserMgr)
	p.ses2User = map[rpc.GateSessionID]*User{}
	p.userID2User = map[uint64]*User{}
	return p
}

func (p *UserMgr) GetUserBySession(id rpc.GateSessionID) (*User, bool) {
	v, ok := p.ses2User[id]
	return v, ok
}

func (p *UserMgr) GetUserByUserID(userID uint64) (*User, bool) {
	v, ok := p.userID2User[userID]
	return v, ok
}

func (p *UserMgr) DelUser(userID uint64) {
	u, ok := p.userID2User[userID]
	if !ok {
		return
	}
	delete(p.ses2User, u.sessionID)
	delete(p.userID2User, userID)
}

func (p *UserMgr) SetUserOffline(userID uint64) {
	u, exist := p.GetUserByUserID(userID)
	if !exist {
		logrus.WithField("userid", userID).Error("UserDisconnect user not existed")
		return
	}

	delete(p.ses2User, u.sessionID)
	u.offline = true
	u.sessionID = rpc.GateSessionID{}
	return
}

func (p *UserMgr) SetUserReconnect(userID uint64, sessionID rpc.GateSessionID) {
	u, exist := p.GetUserByUserID(userID)
	if !exist {
		logrus.WithField("userid", userID).Error("UserDisconnect user not existed")
		return
	}

	u.offline = false
	u.sessionID = sessionID
	p.ses2User[sessionID] = u
	return
}

func (p *UserMgr) AddUser(u *User) {
	p.ses2User[u.sessionID] = u
	p.userID2User[u.userID] = u
}

type User struct {
	game        *Game
	sessionID   rpc.GateSessionID
	userID      uint64
	nickname    string
	accountType AccountType
	headImgUrl  string
	offline     bool
}

func (p *User) Send2Client(msg proto.Message) {
	Server.RPCSession(p.sessionID).SendMsg(msg)
}
