package game

import (
	"container/list"
	"fmt"
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/avatar-fight-server/util"
	"github.com/0990/goserver/service"
	"github.com/golang/protobuf/proto"
	"math"
	"math/rand"
	"time"
)

type AccountType int8

const (
	_ AccountType = iota
	VISITOR
	WX
	ROBOT
)

type OverReason int8

const (
	Killed OverReason = iota
	Normal
)

const (
	ROTATION_DELTA      = 100
	ENTITY_SPEED        = 100.0
	ENTITY_RADIUS       = 100.0
	WORLD_WIDTH         = 1280.0
	WORLD_HEIGHT        = 1280.0
	SHOOT_LIMIT_TIME    = 500 //毫秒
	ENTITY_PROTECT_TIME = 3 * time.Second
	BULLET_LIVE_TIME    = 5 //秒
	BULLET_SPEED        = 1.0
	GAME_ROUND_TIME     = 180
)

type Game struct {
	entitySeqID    int32
	bulletSeqID    int32
	roundReqID     int32
	robotCount     int32
	startTime      int64
	entityMap      map[uint64]*Entity
	aliveEntityMap map[uint64]*Entity
	bulletList     *list.List
	worker         service.Worker
}

type Entity struct {
	userID      uint64
	accountType AccountType
	nickname    string
	headImgUrl  string
	sesID       int32

	entityID           int32
	x                  float64
	y                  float64
	rotation           float64
	lastProcessedInput int32
	bulletSeqID        int32
	hp                 int32
	score              int32
	dead               bool
	isProtected        bool //被保护中，玩家创建后一定时间内是受保护状态

	createdTime int64

	killCount     int32
	lastShootTime int64

	game *Game
}

//func (p *Entity) Kill(someone *Entity) {
//	p.killCount++
//	if someone.accountType == ROBOT {
//		p.score++
//	} else {
//		p.score += 5
//	}
//}

func (p *Entity) KilledBy(killerUserID uint64) {
	var killer *Entity
	killer = p.game.entityMap[killerUserID]
	//给击杀者相关奖励
	if killer != nil {
		killer.killCount++
		if p.accountType == ROBOT {
			killer.score++
		} else {
			killer.score += 5
		}
	}
	p.dead = true
	delete(p.game.aliveEntityMap, p.userID)

	killerInfo := &cmsg.SNoticeGameOver_Killer{}
	if killer != nil {
		killerInfo.HeadImgUrl = killer.headImgUrl
		killerInfo.AccountType = int32(killer.accountType)
		killerInfo.Nickname = killer.nickname
		killerInfo.Hp = killer.hp
	}

	leftTime := GAME_ROUND_TIME + p.game.startTime - util.GetCurrentSec()

	msg := &cmsg.SNoticeGameOver{
		OverReason:  int32(Killed),
		GameLeftSec: int32(leftTime),
		Killer:      killerInfo,
	}
	p.game.SendMsg(p.sesID, msg)
}

type Bullet struct {
	bulletID        int32
	damage          int32
	initCenterX     float64
	initCenterY     float64
	x               float64
	y               float64
	rotation        float64
	creatorUserID   uint64
	creatorEntityID int32
	createdTime     int64
}

//刷新
func (p *Game) onUpdate() {

}

func (p *Game) noticeAllWorldState() {

}

func (p *Game) CreateRobot() {
	p.robotCount++
	nickname := fmt.Sprintf("robot%d", p.robotCount)
	p.CreateEntity(ROBOT, nickname, 0, 0)
}

//模拟玩家移动射击
func (p *Game) RobotMoveShoot() {
	now := util.GetCurrentMillSec()
	for _, entity := range p.entityMap {
		if entity.dead {
			continue
		}
		if entity.accountType != ROBOT {
			continue
		}

		targetRotation := entity.rotation
		randInt := rand.Int() % 100
		switch {
		case randInt < 10:
			targetRotation = float64(rand.Int() % 180)
		case randInt < 20:
			targetRotation = float64(rand.Int() % 180)
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

		p.EntityMove(entity.userID, 0.05, targetRotation, 0)
		if now-entity.lastShootTime > 700 {
			p.EntityShoot(entity.userID)
		}
	}
}

func (p *Game) CreateEntity(accountType AccountType, nickName string, userID uint64, sesID int32) *Entity {
	x, y := p.GetRandPosition()

	e := &Entity{}
	e.x = x
	e.y = y
	e.sesID = sesID
	e.userID = userID
	e.accountType = accountType
	e.createdTime = util.GetCurrentSec()
	e.isProtected = true
	e.lastShootTime = util.GetCurrentSec()

	p.worker.AfterPost(ENTITY_PROTECT_TIME, func() {
		e.isProtected = false
	})
	return e
}

func (p *Game) EntityMove(userID uint64, pressTime, targetRotation float64, inputNumber int32) {
	entity, exist := p.entityMap[userID]
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
	entity.lastProcessedInput = inputNumber
}

func (p *Game) EntityJump(userID uint64) {
	entity, exist := p.entityMap[userID]
	if !exist {
		return
	}

	entity.y = NewYPos(entity.y, entity.rotation, 1.0)
	entity.x = NewXPos(entity.x, entity.rotation, 1.0)
}

func (p *Game) EntityShoot(userID uint64) {
	entity, exist := p.entityMap[userID]
	if !exist {
		return
	}

	if entity.dead {
		return
	}

	now := util.GetCurrentMillSec()
	if now-entity.lastShootTime < SHOOT_LIMIT_TIME {
		return
	}

	entity.lastShootTime = now

	b := &Bullet{}
	b.rotation = entity.rotation
	rotation := b.rotation * math.Pi / 180
	b.y = entity.y + ENTITY_RADIUS*math.Sin(rotation)
	b.x = entity.x + ENTITY_RADIUS*math.Cos(rotation)
	b.initCenterX = entity.x
	b.initCenterY = entity.y

	b.creatorUserID = entity.userID
	b.creatorEntityID = entity.entityID

	b.bulletID = p.NewBulletID()
	b.createdTime = now
	b.damage = 2

	p.bulletList.PushBack(b)
	//TODO　sendShootmsg

}

func (p *Game) NewBulletID() int32 {
	p.bulletSeqID++
	return p.bulletSeqID
}

func (p *Game) GetRandPosition() (x, y float64) {
	var nextRand bool
	for {
		nextRand = false
		x = rand.Int63()%(WORLD_WIDTH-2*ENTITY_RADIUS) + ENTITY_RADIUS
		y = rand.Int63()%(WORLD_HEIGHT-2*ENTITY_RADIUS) + ENTITY_RADIUS

		for _, entity := range p.entityMap {
			if entity.dead {
				continue
			}

			distance := math.Sqrt(math.Pow(entity.x-x, 2) + math.Pow(entity.y-y, 2))
			if distance < 5*ENTITY_RADIUS {
				nextRand = true
				break
			}
		}
		if !nextRand {
			break
		}
	}
	return
}

func (p *Game) CheckCollision() {
	var delBullets, delEntitys []int32
	var dirtyUserids []uint64

	now := util.GetCurrentMillSec()
	//检测子弹和entity之间碰撞
	for e := p.bulletList.Front(); e != nil; {
		bullet := e.Value.(*Bullet)
		var bulletErase bool
		if now-bullet.createdTime > BULLET_LIVE_TIME {
			bulletErase = true
		} else {
			rotation := bullet.rotation * math.Pi / 180
			bullet.y = bullet.initCenterY + (float64(now-bullet.createdTime)*BULLET_SPEED/1000+ENTITY_RADIUS)*math.Sin(rotation)
			bullet.x = bullet.initCenterX + (float64(now-bullet.createdTime)*BULLET_SPEED/1000+ENTITY_RADIUS)*math.Cos(rotation)

			for _, entity := range p.aliveEntityMap {
				if bullet.creatorUserID == entity.userID {
					continue
				}
				distance := util.Distance(bullet.x, entity.x, bullet.y, entity.y)
				if distance > ENTITY_RADIUS {
					continue
				}

				if !entity.isProtected {
					entity.hp -= bullet.damage
					if entity.hp <= 0 {
						entity.KilledBy(bullet.creatorUserID)
						delEntitys = append(delEntitys, entity.entityID)
					}
					dirtyUserids = append(dirtyUserids, entity.userID, bullet.creatorUserID)
				}
				bulletErase = true
				break
			}
		}

		next := e.Next()
		if bulletErase {
			delBullets = append(delBullets, bullet.bulletID)
			p.bulletList.Remove(e)
		}
		e = next
	}

	//检测entity之间碰撞
	for userIDA, entityA := range p.aliveEntityMap {
		if entityA.isProtected {
			continue
		}
		for userIDB, entityB := range p.aliveEntityMap {
			if entityA.entityID == entityB.entityID {
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
				delEntitys = append(delEntitys, entityB.entityID)
			}

			if entityA.hp <= 0 {
				entityA.KilledBy(userIDB)
				delEntitys = append(delEntitys, entityA.entityID)
			}
			dirtyUserids = append(dirtyUserids, userIDA, userIDB)
			if entityA.dead {
				break
			}
		}
	}

	if len(delBullets) > 0 || len(delEntitys) > 0 || len(dirtyUserids) > 0 {
		changeEntitys := make([]*cmsg.SNoticeWorldChange_Entity, 0, len(dirtyUserids))
		for _, userID := range dirtyUserids {
			if entity, exist := p.entityMap[userID]; exist {
				changeEntitys = append(changeEntitys, &cmsg.SNoticeWorldChange_Entity{
					Id:        entity.entityID,
					Score:     entity.score,
					KillCount: entity.killCount,
					Hp:        entity.hp,
				})
			}
		}

		msg := &cmsg.SNoticeWorldChange{
			DeleteBullets:  delBullets,
			DeleteEntitys:  delEntitys,
			ChangedEntitys: changeEntitys,
		}

		p.Send2All(msg)
	}
}

func (p *Game) Send2All(msg proto.Message) {

}

func (p *Game) SendMsg(sesID int32, msg proto.Message) {

}
