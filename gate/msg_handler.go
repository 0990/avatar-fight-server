package gate

import (
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/avatar-fight-server/msg/smsg"
	"github.com/0990/goserver/network"
	"github.com/0990/goserver/rpc"
	"time"
)

func registerHandler() {
	Gate.RegisterNetWorkEvent(onConnect, onDisconnect)
	Gate.RegisterSessionMsgHandler(Login)
	//Gate.RegisterSessionMsgHandler(Test)
	Gate.RegisterServerHandler(NoticeSessionClose)
	Gate.RegisterRequestMsgHandler(Metric)
}

func Login(session network.Session, msg *cmsg.ReqLogin) {
	Gate.GetServerById(conf.CenterServerID).Request(&smsg.GaCeReqLogin{Sesid: session.ID(), Token: msg.Token}, func(cbResp *smsg.GaCeRespLogin, err error) {
		resp := &cmsg.RespLogin{}
		defer session.SendMsg(resp)
		if err != nil {
			resp.Err = cmsg.RespLogin_RPCError
			return
		}

		if cbResp.Err != 0 {
			resp.Err = 1
			return
		}

		userID := cbResp.UserID

		resp.UserID = userID
		resp.Token = cbResp.Token
		resp.InGame = cbResp.InGame

		SMgr.SetSessionLogined(session.ID(), userID)
		return
	})
}

func onConnect(conn network.Session) {
	SMgr.sesID2Session[conn.ID()] = &Session{
		session: conn,
		sesID:   conn.ID(),
	}
}

func onDisconnect(conn network.Session) {
	//TODO 这里先简单处理，理论上没有登录上的玩家不用向中心服务器通知离线事件
	Gate.GetServerById(conf.CenterServerID).Notify(&smsg.GaCeUserDisconnect{
		SessionID: conn.ID(),
	})
	delete(SMgr.sesID2Session, conn.ID())
}

func NoticeSessionClose(server rpc.Server, req *smsg.CeGaCloseSession) {
	s, exist := SMgr.sesID2Session[req.SessionID]
	if !exist {
		return
	}
	s.session.SendMsg(&cmsg.SNoticeKickOut{
		Reason: cmsg.SNoticeKickOut_Relogin,
	})
	//TODO 客户端有问题，暂时服务端过会再关
	Gate.AfterPost(time.Second, func() {
		s.session.Close()
	})
	delete(SMgr.sesID2Session, req.SessionID)
}

func Test(session network.Session, msg *cmsg.ReqJoinGame) {
	Gate.GetServerById(conf.CenterServerID).RouteSession2Server(session.ID(), msg)
}

func Metric(peer rpc.RequestServer, req *smsg.AdReqMetrics) {
	resp := &smsg.AdRespMetrics{}
	defer peer.Answer(resp)
}
