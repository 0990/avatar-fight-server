package center

import (
	"github.com/0990/avatar-fight-server/util"
	"github.com/0990/goserver/rpc"
)

type UserMgr struct {
	token2User map[string]*User
	id2User    map[uint64]*User
	ses2User   map[rpc.GateSessionID]*User //TODO rpc.Session作为key的情况
	userSeqid  uint64
}

func NewUserMgr() *UserMgr {
	p := new(UserMgr)
	p.token2User = make(map[string]*User)
	p.id2User = make(map[uint64]*User)
	p.ses2User = make(map[rpc.GateSessionID]*User)
	return p
}

type User struct {
	userid  uint64
	session rpc.GateSessionID
	token   string
	online  bool
	//TODO token过期处理
	//tokenExpireTime int64
	game *Game
}

func (u *User) SetSession(s rpc.GateSessionID) {
	u.session = s
}

func (p *UserMgr) NewUserID() uint64 {
	p.userSeqid++
	return p.userSeqid
}

func (p *UserMgr) AddUser(s rpc.GateSessionID) *User {
	token := util.RandomStr(10)
	userID := p.NewUserID()

	u := &User{
		userid:  userID,
		session: s,
		token:   token,
		online:  true,
	}
	p.token2User[token] = u
	p.id2User[userID] = u
	p.ses2User[s] = u
	return u
}

func (p *UserMgr) UpdateUserSession(u *User, s rpc.GateSessionID) {
	u.session = s
	p.ses2User[s] = u
}

func (p *UserMgr) FindUserByToken(token string) (*User, bool) {
	v, ok := p.token2User[token]
	return v, ok
}

func (p *UserMgr) FindUserBySession(session rpc.GateSessionID) (*User, bool) {
	v, ok := p.ses2User[session]
	return v, ok
}

func (p *UserMgr) RemoveSession(session rpc.GateSessionID) {
	delete(p.ses2User, session)
}
func (p *UserMgr) FindUserByUserId(userID uint64) (*User, bool) {
	v, ok := p.id2User[userID]
	return v, ok
}
