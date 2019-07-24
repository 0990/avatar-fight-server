package game

import (
	"container/list"
	"fmt"
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/avatar-fight-server/util"
	"github.com/0990/goserver/rpc"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"math"
	"math/rand"
	"sort"
	"time"
)

type OverReason int8

const (
	Invalid OverReason = iota
	Killed
	Normal
)

const (
	ENTITY_SPEED  = 100.0
	ENTITY_RADIUS = 50.0
	BULLET_SPEED  = 400.0

	SHOOT_LIMIT_MS    = 500  //毫秒
	BULLET_LIVE_MS    = 1500 //毫秒
	ENTITY_PROTECT_MS = 2000 //毫秒

	GAME_ROUND_SEC = 60 //秒

	WORLD_WIDTH  = 2560.0
	WORLD_HEIGHT = 1440.0

	TICKER_UPDATE_MS     = 50
	TICKER_NOTICE_POS_MS = 100

	ROTATION_DELTA = 180

	ROBOT_ID_START = 100000000000
)

type Game struct {
	gameID             int64
	entitySeqID        int32
	bulletSeqID        int32
	userSeqID          uint64
	roundReqID         int32
	robotCount         int32
	startTimeSec       int64
	userID2entity      map[uint64]*Entity
	aliveUserID2Entity map[uint64]*Entity
	bulletList         *list.List
	worker             service.Worker
	onGameEnd          func(*Game)

	updateTicker  *time.Ticker
	syncPosTicker *time.Ticker
	printTicker   *time.Ticker
}

func newGame(gameID int64, onGameEnd func(*Game), worker service.Worker) *Game {
	g := new(Game)
	g.userID2entity = map[uint64]*Entity{}
	g.aliveUserID2Entity = map[uint64]*Entity{}
	g.bulletList = list.New()
	g.gameID = gameID
	g.onGameEnd = onGameEnd
	g.worker = worker
	return g
}

type Entity struct {
	id                 int32
	x                  float32
	y                  float32
	rotation           float32
	lastProcessedInput int32
	hp                 int32
	score              int32
	dead               bool
	isProtected        bool //被保护中，玩家创建后一定时间内是受保护状态
	clientSceneReady   bool //如果客户端进入场景了，代表已经数据和界面都准备好了，可以接受任何广播消息了

	createdTime int64

	killCount     int32
	lastShootTime int64

	game *Game
	u    *User
}

//func (p *Entity) Kill(someone *Entity) {
//	p.killCount++
//	if someone.accountType == ROBOT {
//		p.score++
//	} else {
//		p.score += 5
//	}
//}

func (p *Entity) ToCMsg() *cmsg.Entity {
	return &cmsg.Entity{
		Id:          p.id,
		AccountType: int32(p.u.accountType),
		HeadImgUrl:  p.u.headImgUrl,
		Rotation:    p.rotation,
		Hp:          p.hp,
		Score:       p.score,
		KillCount:   p.killCount,
		Dead:        p.dead,
		Protected:   p.isProtected,
		X:           p.x,
		Y:           p.y,
		Nickname:    p.u.nickname,
	}
}

func (p *Entity) KilledBy(killerUserID uint64) {
	var killer *Entity
	killer = p.game.userID2entity[killerUserID]
	//给击杀者相关奖励
	if killer != nil {
		killer.killCount++
		if p.u.accountType == ROBOT {
			killer.score++
		} else {
			killer.score += 5
		}
	}
	p.dead = true
	delete(p.game.aliveUserID2Entity, p.u.userID)

	killerInfo := &cmsg.SNoticeGameOver_Killer{}
	if killer != nil {
		killerInfo.HeadImgUrl = killer.u.headImgUrl
		killerInfo.AccountType = int32(killer.u.accountType)
		killerInfo.Nickname = killer.u.nickname
		killerInfo.Hp = killer.hp
	}

	msg := &cmsg.SNoticeGameOver{
		OverReason:  int32(Killed),
		GameLeftSec: p.game.gameLeftSec(),
		Killer:      killerInfo,
	}
	p.u.Send2Client(msg)
}

type Bullet struct {
	id              int32
	damage          int32
	initCenterX     float32
	initCenterY     float32
	x               float32
	y               float32
	rotation        float32
	creatorUserID   uint64
	creatorEntityID int32
	createdTime     int64
}

func (p *Game) Run() {
	p.startTimeSec = util.GetCurrentSec()
	p.updateTicker = p.worker.NewTicker(TICKER_UPDATE_MS*time.Millisecond, p.onUpdate)
	p.syncPosTicker = p.worker.NewTicker(TICKER_NOTICE_POS_MS*time.Millisecond, p.onNoticeWorldPos)
	p.worker.AfterPost(GAME_ROUND_SEC*time.Second, p.GameNormalEnd)

	p.printTicker = p.worker.NewTicker(time.Second*1, func() {
		//fmt.Println("entity count", len(p.userID2entity))
		//fmt.Println("alive entity count", len(p.aliveUserID2Entity))
		//fmt.Println("workerLen", p.worker.Len())
	})
}

func (p *Game) GameNormalEnd() {
	p.updateTicker.Stop()
	p.syncPosTicker.Stop()
	p.printTicker.Stop()

	//TODO　rankInfo
	rankInfo := &cmsg.Rank{}
	sortEntitys := p.sortRank()
	var rank int32
	for _, e := range sortEntitys {
		rank++
		rankInfo.List = append(rankInfo.List, &cmsg.Rank_Item{
			EntityID:   e.id,
			Score:      e.score,
			Rank:       rank,
			KillCount:  e.killCount,
			HeadImgUrl: e.u.headImgUrl,
			Nickname:   e.u.nickname,
		})
	}
	msg := &cmsg.SNoticeGameOver{
		OverReason: int32(Normal),
		Rank:       rankInfo,
	}
	p.Send2All(msg)

	p.userID2entity = make(map[uint64]*Entity)
	p.aliveUserID2Entity = make(map[uint64]*Entity)
	p.bulletList.Init()
	p.entitySeqID = 0
	p.robotCount = 0
	p.bulletSeqID = 0

	p.onGameEnd(p)
}

//排行
func (p *Game) sortRank() []*Entity {
	entitys := make([]*Entity, 0, len(p.aliveUserID2Entity))

	for _, v := range p.aliveUserID2Entity {
		entitys = append(entitys, v)
	}

	sort.Slice(entitys, func(i, j int) bool {
		a := entitys[i]
		b := entitys[j]
		if a.score > b.score {
			return true
		} else if a.score < b.score {
			return false
		} else {
			return a.id < b.id
		}
	})
	return entitys
}

//刷新 主逻辑
func (p *Game) onUpdate() {
	t := time.Now()
	defer util.PrintElapse("onUpdate", t)
	//创建机器人
	if rand.Int31()%150 < 3 {
		p.createRobot()
	}
	//机器人移动射击
	p.robotMoveShoot()
	//检查碰撞
	p.CheckCollision()
}

func (p *Game) onNoticeWorldPos() {
	t := time.Now()
	defer util.PrintElapse("onNoticeWorldPos", t)
	entities := make([]*cmsg.SNoticeWorldPos_Entity, 0, len(p.aliveUserID2Entity))
	for _, e := range p.aliveUserID2Entity {
		entities = append(entities, &cmsg.SNoticeWorldPos_Entity{
			Id:       e.id,
			X:        e.x,
			Y:        e.y,
			Rotation: e.rotation,
		})
	}

	p.SendPos2All(&cmsg.SNoticeWorldPos{
		Entities: entities,
	})
}

func (p *Game) newUserID() uint64 {
	p.userSeqID++
	return p.userSeqID
}

func (p *Game) newEntityID() int32 {
	p.entitySeqID++
	return p.entitySeqID
}

func (p *Game) createRobot() {
	p.robotCount++
	nickname := fmt.Sprintf("robot%d", p.robotCount)

	//TODO 这里先简化处理，后面要使用唯一ID
	userID := ROBOT_ID_START + uint64(p.robotCount)

	u := &User{
		accountType: ROBOT,
		userID:      userID,
		nickname:    nickname,
	}
	p.createEntity(u)
}

//模拟玩家移动射击
func (p *Game) robotMoveShoot() {
	now := util.GetCurrentMillSec()
	for _, entity := range p.aliveUserID2Entity {
		if entity.u.accountType != ROBOT {
			continue
		}

		targetRotation := entity.rotation
		randInt := rand.Int() % 100
		switch {
		case randInt < 10:
			targetRotation = float32(rand.Int() % 180)
		case randInt < 20:
			targetRotation = float32(rand.Int() % 180)
		default:
			if entity.x < 640 {
				if entity.y < 360 {
					targetRotation = 45
				} else if entity.y > 1080 {
					targetRotation = -45
				} else {
					targetRotation = 0
				}
			} else if entity.x > 1920 {
				if entity.y < 360 {
					targetRotation = 135
				} else if entity.y > 1080 {
					targetRotation = -135
				} else {
					targetRotation = 180
				}
			}
		}

		p.entityMove(entity.u.userID, 0.05, targetRotation, 0)
		if now-entity.lastShootTime > 700 {
			p.entityShoot(entity.u.userID)
		}
	}
}

func (p *Game) createEntity(u *User) *Entity {
	x, y := p.GetRandPosition()

	e := &Entity{}
	e.id = p.newEntityID()
	e.x = x
	e.y = y
	e.u = u
	e.game = p
	e.hp = 6
	e.lastProcessedInput = -1
	e.createdTime = util.GetCurrentSec()
	e.isProtected = true
	e.lastShootTime = util.GetCurrentSec()

	p.worker.AfterPost(ENTITY_PROTECT_MS, func() {
		e.isProtected = false
	})
	p.userID2entity[u.userID] = e
	p.aliveUserID2Entity[u.userID] = e

	p.Send2All(&cmsg.SNoticeNewEntity{
		Entity: e.ToCMsg(),
	})
	return e
}

func (p *Game) entityMove(userID uint64, pressTime, targetRotation float32, inputSeqid int32) {
	entity, exist := p.aliveUserID2Entity[userID]
	if !exist {
		return
	}

	lastRotation := entity.rotation
	newRotation := NewRotation(lastRotation, pressTime, targetRotation)
	newXPos := NewXPos(entity.x, lastRotation, pressTime)
	newYPos := NewYPos(entity.y, lastRotation, pressTime)

	entity.y = newYPos
	entity.x = newXPos
	entity.rotation = newRotation
	entity.lastProcessedInput = inputSeqid
}

func (p *Game) entityJump(userID uint64) {
	entity, exist := p.aliveUserID2Entity[userID]
	if !exist {
		return
	}

	entity.y = NewYPos(entity.y, entity.rotation, 1.0)
	entity.x = NewXPos(entity.x, entity.rotation, 1.0)
}

func (p *Game) entityShoot(userID uint64) {
	entity, exist := p.aliveUserID2Entity[userID]
	if !exist {
		return
	}

	if entity.dead {
		return
	}

	now := util.GetCurrentMillSec()
	if now-entity.lastShootTime < SHOOT_LIMIT_MS {
		return
	}

	entity.lastShootTime = now

	b := &Bullet{}
	b.rotation = entity.rotation
	rotation := float64(b.rotation * math.Pi / 180)
	b.y = entity.y + ENTITY_RADIUS*float32(math.Sin(rotation))
	b.x = entity.x + ENTITY_RADIUS*float32(math.Cos(rotation))
	b.initCenterX = entity.x
	b.initCenterY = entity.y

	b.creatorUserID = entity.u.userID
	b.creatorEntityID = entity.id

	b.id = p.NewBulletID()
	b.createdTime = now
	b.damage = 2
	p.bulletList.PushBack(b)

	//通知所有玩家
	msg := &cmsg.SNoticeShoot{}
	msg.X = b.x
	msg.Y = b.y
	msg.Rotation = b.rotation
	msg.BulletID = b.id
	msg.CreatorEntityID = b.creatorEntityID
	p.Send2All(msg)
}

func (p *Game) NewBulletID() int32 {
	p.bulletSeqID++
	return p.bulletSeqID
}

func (p *Game) GetRandPosition() (x, y float32) {
	var nextRand bool
	var forCount int
	for {
		nextRand = false
		x = float32(rand.Int31()%(WORLD_WIDTH-2*ENTITY_RADIUS) + ENTITY_RADIUS)
		y = float32(rand.Int31()%(WORLD_HEIGHT-2*ENTITY_RADIUS) + ENTITY_RADIUS)

		for _, entity := range p.aliveUserID2Entity {
			if entity.dead {
				continue
			}

			//distance := math.Sqrt(math.Pow(float64(entity.x-x), 2) + math.Pow(entity.y-y, 2))
			distance := util.Distance(entity.x, entity.y, x, y)
			if distance < 5*ENTITY_RADIUS {
				nextRand = true
				break
			}
		}
		if !nextRand {
			break
		}

		forCount++
		if forCount > 10 {
			break
		}
	}
	return
}

func (p *Game) CheckCollision() {
	var delBullets, delEntities []int32
	var dirtyUserIds []uint64

	now := util.GetCurrentMillSec()
	//检测子弹和entity之间碰撞
	for e := p.bulletList.Front(); e != nil; {
		bullet := e.Value.(*Bullet)
		var bulletErase bool
		if now-bullet.createdTime > BULLET_LIVE_MS {
			bulletErase = true
		} else {
			rotation := float64(bullet.rotation * math.Pi / 180)
			bullet.y = bullet.initCenterY + (float32(now-bullet.createdTime)*BULLET_SPEED/1000+ENTITY_RADIUS)*float32(math.Sin(rotation))
			bullet.x = bullet.initCenterX + (float32(now-bullet.createdTime)*BULLET_SPEED/1000+ENTITY_RADIUS)*float32(math.Cos(rotation))

			for _, entity := range p.aliveUserID2Entity {
				if bullet.creatorUserID == entity.u.userID {
					continue
				}
				distance := util.Distance(bullet.x, bullet.y, entity.x, entity.y)
				if distance > ENTITY_RADIUS {
					continue
				}

				if !entity.isProtected {
					entity.hp -= bullet.damage
					if entity.hp <= 0 {
						entity.KilledBy(bullet.creatorUserID)
						delEntities = append(delEntities, entity.id)
					}
					dirtyUserIds = append(dirtyUserIds, entity.u.userID, bullet.creatorUserID)
				}
				bulletErase = true
				break
			}
		}

		next := e.Next()
		if bulletErase {
			delBullets = append(delBullets, bullet.id)
			p.bulletList.Remove(e)
		}
		e = next
	}

	//检测entity之间碰撞
	for userIDA, entityA := range p.aliveUserID2Entity {
		if entityA.isProtected {
			continue
		}
		for userIDB, entityB := range p.aliveUserID2Entity {
			if entityA.id == entityB.id {
				continue
			}

			if entityB.isProtected {
				continue
			}

			distance := util.Distance(entityA.x, entityA.y, entityB.x, entityB.y)

			if distance > 2*ENTITY_RADIUS {
				continue
			}

			damage := util.Int32Min(entityA.hp, entityB.hp)
			entityA.hp -= damage
			entityB.hp -= damage

			if entityB.hp <= 0 {
				entityB.KilledBy(userIDA)
				delEntities = append(delEntities, entityB.id)
			}

			if entityA.hp <= 0 {
				entityA.KilledBy(userIDB)
				delEntities = append(delEntities, entityA.id)
			}
			dirtyUserIds = append(dirtyUserIds, userIDA, userIDB)
			if entityA.dead {
				break
			}
		}
	}

	if len(delBullets) > 0 || len(delEntities) > 0 || len(dirtyUserIds) > 0 {
		changeEntitys := make([]*cmsg.SNoticeWorldChange_Entity, 0, len(dirtyUserIds))
		for _, userID := range dirtyUserIds {
			if entity, exist := p.userID2entity[userID]; exist && !entity.dead {
				changeEntitys = append(changeEntitys, &cmsg.SNoticeWorldChange_Entity{
					Id:        entity.id,
					Score:     entity.score,
					KillCount: entity.killCount,
					Hp:        entity.hp,
				})
			}
		}

		msg := &cmsg.SNoticeWorldChange{
			DeleteBullets:   delBullets,
			DeleteEntities:  delEntities,
			ChangedEntities: changeEntitys,
		}

		p.Send2All(msg)
	}
}

func (p *Game) Send2All(msg proto.Message) {
	for _, v := range p.aliveUserID2Entity {
		if !v.clientSceneReady {
			continue
		}
		v.u.Send2Client(msg)
	}
}

func (p *Game) SendPos2All(msg *cmsg.SNoticeWorldPos) {
	for _, v := range p.aliveUserID2Entity {
		if v.u.accountType == ROBOT || !v.clientSceneReady {
			continue
		}
		msg.LastProcessedInputID = v.lastProcessedInput
		v.u.Send2Client(msg)
	}
}

func (p *Game) Join(u *User) error {
	if u.userID >= ROBOT_ID_START {
		return errors.New("userid too large")
	}
	if _, ok := p.userID2entity[u.userID]; ok {
		return errors.New("user already in game")
	}
	u.game = p
	p.createEntity(u)
	return nil
}

func (p *Game) gameLeftSec() int32 {
	return int32(GAME_ROUND_SEC + p.startTimeSec - util.GetCurrentSec())
}

func (p *Game) OnReqGameScene(session rpc.Session, userID uint64, msg *cmsg.ReqGameScene) {
	resp := &cmsg.RespGameScene{}
	entity, ok := p.userID2entity[userID]
	if !ok {
		resp.Err = 1
		session.SendMsg(resp)
		return
	}
	entity.clientSceneReady = true

	entitys := make([]*cmsg.Entity, 0, len(p.aliveUserID2Entity))
	for _, e := range p.aliveUserID2Entity {
		entitys = append(entitys, e.ToCMsg())
	}

	resp.Entities = entitys
	resp.GameLeftSec = p.gameLeftSec()
	session.SendMsg(resp)
}

func (p *Game) OnReqEnterGame(session rpc.Session, userID uint64, msg *cmsg.ReqEnterGame) {
	resp := &cmsg.RespEnterGame{}
	entity, ok := p.userID2entity[userID]
	if !ok {
		resp.Err = 1
		session.SendMsg(resp)
		return
	}

	if entity.dead {
		resp.Dead = true
		resp.GameLeftSec = p.gameLeftSec()
		session.SendMsg(resp)
		return
	}

	config := &cmsg.RespEnterGame_Config{
		BulletLiveTime:    BULLET_LIVE_MS,
		RotationDelta:     ROTATION_DELTA,
		EntitySpeed:       ENTITY_SPEED,
		BulletSpeed:       BULLET_SPEED,
		NoticePosDuration: TICKER_NOTICE_POS_MS,
		ProtectTime:       ENTITY_PROTECT_MS,
		EntityRadius:      ENTITY_RADIUS,
	}
	resp.EntityID = entity.id
	resp.Config = config
	resp.GameLeftSec = p.gameLeftSec()
	session.SendMsg(resp)
}

func (p *Game) OnReqMove(userID uint64, msg *cmsg.ReqMove) {
	p.entityMove(userID, msg.PressTime, msg.TargetRotation, msg.InputSeqID)
}

func (p *Game) OnReqJump(userID uint64, msg *cmsg.ReqJump) {
	p.entityJump(userID)
}

func (p *Game) OnReqShoot(userID uint64, msg *cmsg.ReqShoot) {
	p.entityShoot(userID)
}
