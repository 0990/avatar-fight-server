package center

import (
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/avatar-fight-server/msg/smsg"
	"github.com/0990/goserver/rpc"
)

func registerHandler() {
	Server.RegisterRequestMsgHandler(Login)
	Server.RegisterSessionMsgHandler(JoinGame)
	Server.RegisterServerHandler(NoticeGameStart)
	Server.RegisterServerHandler(NoticeGameEnd)
	Server.RegisterServerHandler(UserDisconnect)
}

func Login(peer rpc.RequestServer, req *smsg.GaCeReqLogin) {
	u, exist := UMgr.FindUserByToken(req.Token)

	gateSessionID := rpc.GateSessionID{
		GateID: peer.ID(),
		SesID:  req.Sesid,
	}
	//这里先简化处理，没找到就新建个玩家
	if exist {
		if u.online {
			Server.GetServerById(conf.GateServerID).Notify(&smsg.CeGaCloseSession{
				SessionID: u.session.SesID,
			})
			UMgr.RemoveSession(u.session)
		}

		if u.game != nil {
			if u.online {
				Server.GetServerById(conf.GameServerID).Notify(&smsg.CeGameUserDisconnect{
					Userid: u.userID,
				})
			}
			Server.GetServerById(conf.GameServerID).Notify(&smsg.CeGameUserReconnect{
				Userid:    u.userID,
				GateID:    gateSessionID.GateID,
				SessionID: gateSessionID.SesID,
			})
		}

		u.online = true
		UMgr.UpdateUserSession(u, gateSessionID)
	} else {
		u = UMgr.AddUser(gateSessionID)
	}
	resp := &smsg.GaCeRespLogin{
		UserID: u.userID,
		Token:  u.token,
		InGame: u.game != nil,
	}
	peer.Answer(resp)
}

func JoinGame(session rpc.Session, req *cmsg.ReqJoinGame) {
	resp := &cmsg.RespJoinGame{}

	u, exist := UMgr.FindUserBySession(session.GateSessionID())
	if !exist {
		resp.Err = cmsg.RespJoinGame_UserNotExisted
		session.SendMsg(resp)
		return
	}

	if u.game != nil {
		resp.Err = cmsg.RespJoinGame_AlreadyInGame
		session.SendMsg(resp)
		return
	}

	Server.GetServerById(conf.GameServerID).Request(&smsg.CeGamReqJoinGame{
		Userid:       u.userID,
		Nickname:     req.Nickname,
		GateServerid: u.session.GateID,
		Sesid:        u.session.SesID,
	}, func(cbResp *smsg.CeGamRespJoinGame, err error) {
		if err != nil {
			resp.Err = 2
			session.SendMsg(cbResp)
			return
		}
		gameID := cbResp.Gameid
		u.game = GMgr.GetGame(gameID)
		GMgr.AddGameUser(gameID, u.userID)
		//TODO 多游戏服时，要在gate绑定game服
		//Server.GetServerById(100).Send(&smsg.CeGaBindGameServer{
		//	Sesid:        0,
		//	Gameserverid: 102,
		//})
		resp.Nickname = req.Nickname
		session.SendMsg(resp)
		return
	})
}

func NoticeGameStart(server rpc.Server, req *smsg.GamCeNoticeGameStart) {
	GMgr.AddGame(req.Gameid, server.ID())
}

func NoticeGameEnd(server rpc.Server, req *smsg.GamCeNoticeGameEnd) {
	game := GMgr.RemoveGame(req.Gameid)
	if game != nil {
		for _, userID := range game.userIds {
			if u, exist := UMgr.FindUserByUserId(userID); exist {
				u.game = nil
			}
		}
	}
}

func UserDisconnect(server rpc.Server, req *smsg.GaCeUserDisconnect) {
	session := rpc.GateSessionID{
		GateID: server.ID(),
		SesID:  req.SessionID,
	}
	u, exist := UMgr.FindUserBySession(session)
	if !exist {
		return
	}
	u.online = false
	UMgr.RemoveSession(session)
	if u.game != nil {
		Server.GetServerById(conf.GameServerID).Notify(&smsg.CeGameUserDisconnect{
			Userid: u.userID,
		})
	}
}
