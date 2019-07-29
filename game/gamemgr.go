package game

import (
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/msg/smsg"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/pkg/errors"
)

type GameMgr struct {
	//TODO 使用snowflake生成分布式的游戏ID
	gameSeqid   int64
	gameid2Game map[int64]*Game
	userid2Game map[uint64]*Game
	currGame    *Game
	worker      service.Worker
}

func NewGameMgr(worker service.Worker) *GameMgr {
	p := new(GameMgr)
	p.gameid2Game = map[int64]*Game{}
	p.userid2Game = map[uint64]*Game{}
	p.worker = worker
	return p
}

func (p *GameMgr) Start() {
	p.CreateNewGame()
}

func (p *GameMgr) CreateNewGame() {
	p.gameSeqid++
	game := newGame(p.gameSeqid, p.onGameEnd, p.worker)
	p.gameid2Game[p.gameSeqid] = game
	p.currGame = game
	game.Run()
	Server.GetServerById(conf.CenterServerID).Notify(&smsg.GamCeNoticeGameStart{
		Gameid: game.gameID,
	})
	return
}

func (p *GameMgr) onGameEnd(g *Game) {
	for _, entity := range g.userID2entity {
		UMgr.DelUser(entity.u.userID)
	}
	delete(p.gameid2Game, g.gameID)
	Server.GetServerById(conf.CenterServerID).Notify(&smsg.GamCeNoticeGameEnd{
		Gameid: g.gameID,
	})
	p.CreateNewGame()
}

func (p *GameMgr) JoinGame(userID uint64, nickname string, sesGateID int32, sesID int32) (*Game, error) {
	if p.currGame == nil {
		return nil, errors.New("currGame not exist")
	}

	session := rpc.GateSessionID{
		GateID: sesGateID,
		SesID:  sesID,
	}

	u := &User{
		accountType: VISITOR,
		sessionID:   session,
		userID:      userID,
		nickname:    nickname,
	}

	err := p.currGame.Join(u)
	if err != nil {
		return nil, err
	}
	UMgr.AddUser(u)
	return p.currGame, nil
}
