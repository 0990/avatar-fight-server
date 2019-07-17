package center

type GameMgr struct {
	gameSeqid   int64
	gameid2Game map[int64]*Game
}

func NewGameMgr() *GameMgr {
	p := &GameMgr{}
	p.gameid2Game = make(map[int64]*Game)
	return p
}

func (p *GameMgr) AddGame(gameID int64, serverID int32) {
	p.gameid2Game[gameID] = &Game{
		serverID: serverID,
		gameID:   gameID,
	}
}

func (p *GameMgr) GetGame(gameID int64) *Game {
	return p.gameid2Game[gameID]
}

func (p *GameMgr) AddGameUser(gameID int64, userID uint64) {
	game, exist := p.gameid2Game[gameID]
	if !exist {
		return
	}
	game.userIds = append(game.userIds, userID)
}

func (p *GameMgr) RemoveGame(gameID int64) *Game {
	game, exist := p.gameid2Game[gameID]
	if !exist {
		return nil
	}
	delete(p.gameid2Game, gameID)
	return game
}

type Game struct {
	gameID   int64
	serverID int32
	userIds  []uint64
}
