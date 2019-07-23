package game

import (
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/avatar-fight-server/msg/smsg"
	"github.com/0990/goserver/rpc"
	"github.com/sirupsen/logrus"
)

func registerHandler() {
	Server.RegisterRequestMsgHandler(JoinGame)
	Server.RegisterSessionMsgHandler(ReqGameScene)
	Server.RegisterSessionMsgHandler(ReqEnterGame)
	Server.RegisterSessionMsgHandler(ReqMove)
	Server.RegisterSessionMsgHandler(ReqJump)
	Server.RegisterSessionMsgHandler(ReqShoot)
	Server.RegisterServerHandler(UserDisconnect)
	Server.RegisterServerHandler(UserReconnect)
}

func UserDisconnect(server rpc.Server, req *smsg.CeGameUserDisconnect) {
	UMgr.SetUserOffline(req.Userid)
}

func UserReconnect(server rpc.Server, req *smsg.CeGameUserReconnect) {
	session := rpc.GateSessionID{
		GateID: req.GateID,
		SesID:  req.SessionID,
	}
	UMgr.SetUserReconnect(req.Userid, session)
}

func JoinGame(server rpc.RequestServer, req *smsg.CeGamReqJoinGame) {
	resp := &smsg.CeGamRespJoinGame{}
	game, err := GMgr.JoinGame(req.Userid, req.Nickname, req.GateServerid, req.Sesid)
	if err != nil {
		resp.Err = smsg.CeGamRespJoinGame_GameNotExist
		server.Answer(resp)
		return
	}

	resp.Gameid = game.gameID
	server.Answer(resp)
}

func ReqEnterGame(session rpc.Session, req *cmsg.ReqEnterGame) {
	resp := &cmsg.RespEnterGame{}
	id := session.GateSessionID()
	user, ok := UMgr.GetUserBySession(id)
	if !ok {
		logrus.Error("session not existed")
		//resp.Err = cmsg.RespGameScene_GameNotExist
		session.SendMsg(resp)
		return
	}
	user.game.OnReqEnterGame(session, user.userID, req)
}

func ReqGameScene(session rpc.Session, req *cmsg.ReqGameScene) {
	resp := &cmsg.RespGameScene{}
	id := session.GateSessionID()
	user, ok := UMgr.GetUserBySession(id)
	if !ok {
		resp.Err = cmsg.RespGameScene_GameNotExist
		session.SendMsg(resp)
		return
	}
	user.game.OnReqGameScene(session, user.userID, req)
}

//TODO 以下三个消息可以利用reflect复用公共代码
func ReqMove(session rpc.Session, req *cmsg.ReqMove) {
	id := session.GateSessionID()
	user, ok := UMgr.GetUserBySession(id)
	if !ok {
		return
	}
	user.game.OnReqMove(user.userID, req)
}

func ReqJump(session rpc.Session, req *cmsg.ReqJump) {
	id := session.GateSessionID()
	user, ok := UMgr.GetUserBySession(id)
	if !ok {
		return
	}
	user.game.OnReqJump(user.userID, req)
}

func ReqShoot(session rpc.Session, req *cmsg.ReqShoot) {
	id := session.GateSessionID()
	user, ok := UMgr.GetUserBySession(id)
	if !ok {
		return
	}
	user.game.OnReqShoot(user.userID, req)
}
