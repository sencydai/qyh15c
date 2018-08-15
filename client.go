package main

import (
	"bytes"

	"github.com/sencydai/gameworld/base"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

type ServerMsgHandler func(*Account, *bytes.Reader)
type ClientMsgHandler func(*Account)

var (
	serverMsgHandles = make(map[int]ServerMsgHandler)
	fightMsgs        = make([]ClientMsgHandler, 0)
	commonMsgs       = make([]ClientMsgHandler, 0)
)

func serverMsgInit() {
	//登陆
	RegServerHandle(proto.System, proto.SystemSLogin, HandleLogin)
	//查询角色列表
	RegServerHandle(proto.System, proto.SystemSActorLists, HandleCheckActorList)
	//随机名称
	RegServerHandle(proto.System, proto.SystemSRandomName, HandleRandomActorName)
	//创建角色
	RegServerHandle(proto.System, proto.SystemSCreateActor, HandleCreateActor)
	//进入游戏
	RegServerHandle(proto.System, proto.SystemSLoginGame, HandleLoginSuccess)
	//战斗结果
	RegServerHandle(proto.Fight, proto.FightSResult, HandleFightResult)
	//系统tips
	RegServerHandle(proto.Chat, proto.ChatSTips, HandleChatTips)
}

func clientMsgInit() {
	RegFightMsg(sendEnterMainFuben)
}

func init() {
	serverMsgInit()
	clientMsgInit()
}

func RegServerHandle(sysId, cmdId byte, handle ServerMsgHandler) {
	mark := (int(sysId) << 8) + int(cmdId)
	serverMsgHandles[mark] = handle
}

func RegFightMsg(handle ClientMsgHandler) {
	fightMsgs = append(fightMsgs, handle)
}

func RegCommonMsg(handle ClientMsgHandler) {
	commonMsgs = append(commonMsgs, handle)
}

func HandleServereMsg(account *Account, sysId, cmdId byte, reader *bytes.Reader) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf(account, "handle Msg %d %d error: %v", sysId, cmdId, err)
		}
	}()

	mark := (int(sysId) << 8) + int(cmdId)
	handle, ok := serverMsgHandles[mark]
	if ok {
		//log.Printf(account, "recv %d %d", sysId, cmdId)
		handle(account, reader)
	}
}

func HandleLogin(account *Account, reader *bytes.Reader) {
	var code byte
	pack.Read(reader, &code)
	if code != 0 {
		log.Printf(account, "HandleLogin error: code(%d)", code)
		account.conn.Close()
		return
	}
	//查询角色列表
	account.send(proto.System, proto.SystemCActorList)
}

func HandleCheckActorList(account *Account, reader *bytes.Reader) {
	var accountId int
	var code int
	pack.Read(reader, &accountId, &code)
	if code < 0 {
		return
	}
	account.accountId = accountId
	if code == 0 {
		randomActorName(account)
		return
	}

	var actorId float64
	var name string
	var head, sex, level, job, camp int

	pack.Read(reader, &actorId, &name, &head, &sex, &level, &job, &camp)

	sendLoginGame(account, actorId)
}

func randomActorName(account *Account) {
	account.send(proto.System, proto.SystemCCreateActor, "", 1, 0, 1, "pf_test")
}

func HandleRandomActorName(account *Account, reader *bytes.Reader) {
	var code int
	pack.Read(reader, &code)
	if code != 0 {
		randomActorName(account)
		return
	}

	var sex int
	var name string
	pack.Read(reader, &sex, &name)
}

func HandleCreateActor(account *Account, reader *bytes.Reader) {
	var actorId float64
	var code int
	pack.Read(reader, &actorId, &code)
	if code != 0 {
		account.conn.Close()
		return
	}
	sendLoginGame(account, actorId)
}

func sendLoginGame(account *Account, actorId float64) {
	account.actorId = int64(actorId)
	account.send(proto.System, proto.SystemCLoginGame, actorId, "pf_test")
}

func HandleLoginSuccess(account *Account, reader *bytes.Reader) {
	var code int
	pack.Read(reader, &code)

	log.Printf(account, "login game code(%d)", code)
	if code != 0 {
		account.conn.Close()
		return
	}
	serverActors[int64(account.actorId)] = gConfigs.ServerId

	Loop(account, "sendChatMsg", gConfigs.ChatPeriod, gConfigs.ChatPeriod, -1, sendChatMsg)

	After(account, "randSendFightMsg", gConfigs.FightPeriod, randSendFightMsg)

	Loop(account, "randSendCommonMsg", gConfigs.MsgPeriod, gConfigs.MsgPeriod, -1, randSendCommonMsg)
}

func sendChatMsg(account *Account) {
	if base.Rand(0, 100) >= 3 {
		return
	}
	index := base.Rand(0, len(gConfigs.ChatMsgs)-1)
	msg := gConfigs.ChatMsgs[index]
	account.send(proto.Chat, proto.ChatCSendChatMsg, byte(1), msg, "")
}

func randSendFightMsg(account *Account) {
	if len(fightMsgs) == 0 {
		return
	}
	index := base.Rand(0, len(fightMsgs)-1)
	fightMsgs[index](account)
}

func randSendCommonMsg(account *Account) {
	// if base.Rand(0, 10000) < 1 {
	// 	account.conn.Close()
	// 	return
	// }
	if len(commonMsgs) == 0 {
		return
	}
	index := base.Rand(0, len(commonMsgs)-1)
	commonMsgs[index](account)
}

func sendEnterMainFuben(account *Account) {
	account.send(proto.Fuben, proto.FubenCLoginMainFuben)
}

func HandleFightResult(account *Account, reader *bytes.Reader) {
	var guid float64
	var ft int
	pack.Read(reader, &guid, &ft)

	account.send(proto.Fight, proto.FightCGetAwards, ft, 0)

	After(account, "randSendFightMsg", gConfigs.FightPeriod, randSendFightMsg)
}

func HandleChatTips(account *Account, reader *bytes.Reader) {
	var t int
	var tips string
	pack.Read(reader, &t, &tips)

	//Print(account, "chat tips: %s", tips)
}
