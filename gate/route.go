package gate

import (
	"github.com/0990/avatar-fight-server/conf"
	"github.com/0990/avatar-fight-server/msg/cmsg"
	"github.com/0990/goserver/network"
	"github.com/golang/protobuf/proto"
)

func registerRoute() {
	//先简化处理，游戏服只有一个
	Route2ServerID((*cmsg.ReqJump)(nil), conf.GameServerID)
	Route2ServerID((*cmsg.ReqMove)(nil), conf.GameServerID)
	Route2ServerID((*cmsg.ReqShoot)(nil), conf.GameServerID)
	Route2ServerID((*cmsg.ReqGameScene)(nil), conf.GameServerID)
	Route2ServerID((*cmsg.ReqJoinGame)(nil), conf.CenterServerID)
	//enter是加入成功后，请求进入游戏的消息
	Route2ServerID((*cmsg.ReqEnterGame)(nil), conf.CenterServerID)
}

func Route2ServerID(msg proto.Message, serverID int32) {
	Gate.RegisterRawSessionMsgHandler(msg, func(session network.Session, message proto.Message) {
		s, exist := SMgr.sesID2Session[session.ID()]
		if exist {
			return
		}
		if !s.logined {
			return
		}
		Gate.GetServerById(serverID).RouteSession2Server(session.ID(), message)
	})
}
