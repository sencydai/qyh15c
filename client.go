package main

import (
	"bytes"
	"time"

	"github.com/sencydai/qyh15c/pack"
	proto "github.com/sencydai/qyh15c/protocol"
	"github.com/sencydai/utils/timer"
)

type ServerMsgHandler func(*AccountData, *bytes.Reader)
type ClientMsgHandler func(*AccountData)

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

func HandleServereMsg(account *AccountData, sysId, cmdId byte, reader *bytes.Reader) {
	defer func() {
		if err := recover(); err != nil {
			Printf(account, "handle Msg %d %d error: %v", sysId, cmdId, err)
		}
	}()

	mark := (int(sysId) << 8) + int(cmdId)
	handle, ok := serverMsgHandles[mark]
	if ok {
		handle(account, reader)
	}
}

func HandleLogin(account *AccountData, reader *bytes.Reader) {
	var code byte
	pack.Read(reader, &code)
	if code != 0 {
		Printf(account, "HandleLogin error: code(%d)", code)
		account.conn.Close()
		return
	}
	//查询角色列表
	send2Server(account, proto.System, proto.SystemCActorList, nil)
}

func HandleCheckActorList(account *AccountData, reader *bytes.Reader) {
	var accountId int32
	var code int32
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
	var head, sex, level, job, camp int32

	pack.Read(reader, &actorId, &name, &head, &sex, &level, &job, &camp)

	sendLoginGame(account, actorId)
}

func randomActorName(account *AccountData) {
	send2Server(account, proto.System, proto.SystemCRandomName, pack.GetBytes(RandInt31n(1, 2)))
}

func HandleRandomActorName(account *AccountData, reader *bytes.Reader) {
	var code int32
	pack.Read(reader, &code)
	if code != 0 {
		randomActorName(account)
		return
	}

	var sex int32
	var name string
	pack.Read(reader, &sex, &name)

	send2Server(account, proto.System, proto.SystemCCreateActor, pack.GetBytes(name, sex, int32(1), int32(1), byte(1), "pf_test"))
}

func HandleCreateActor(account *AccountData, reader *bytes.Reader) {
	var actorId float64
	var code int32
	pack.Read(reader, &actorId, &code)
	if code != 0 {
		account.conn.Close()
		return
	}
	sendLoginGame(account, actorId)
}

func sendLoginGame(account *AccountData, actorId float64) {
	send2Server(account, proto.System, proto.SystemCLoginGame, pack.GetBytes(actorId, "pf_test"))
}

func HandleLoginSuccess(account *AccountData, reader *bytes.Reader) {
	var code int32
	pack.Read(reader, &code)

	Printf(account, "login game code(%d)", code)
	if code != 0 {
		account.conn.Close()
		return
	}

	t := timer.Loop(time.Second*5, time.Second*time.Duration(gConfigs.ChatPeriod), -1, sendChatMsg, account)
	account.timers = append(account.timers, t)

	t = timer.Loop(time.Second*7, time.Second*time.Duration(gConfigs.FightPeriod), -1, randSendFightMsg, account)
	account.timers = append(account.timers, t)

	t = timer.Loop(time.Second*7, time.Second*time.Duration(gConfigs.MsgPeriod), -1, randomSendCommonMsg, account)
	account.timers = append(account.timers, t)
}

func sendChatMsg(account *AccountData) {
	if RandInt31n(0, 100) >= 3 {
		return
	}
	index := RandInt31n(0, int32(len(gConfigs.ChatMsgs)-1))
	msg := gConfigs.ChatMsgs[index]
	send2Server(account, proto.Chat, proto.ChatCSendChatMsg, pack.GetBytes(byte(1), msg, ""))
}

func randSendFightMsg(account *AccountData) {
	if len(fightMsgs) == 0 {
		return
	}
	index := RandInt31n(0, int32(len(fightMsgs)-1))
	fightMsgs[index](account)
}

func randomSendCommonMsg(account *AccountData) {
	if RandInt31n(0, 10000) < 1 {
		account.conn.Close()
		return
	}
	if len(commonMsgs) == 0 {
		return
	}
	index := RandInt31n(0, int32(len(commonMsgs)-1))
	commonMsgs[index](account)
}

func sendEnterMainFuben(account *AccountData) {
	send2Server(account, proto.Fuben, proto.FubenCLoginMainFuben, nil)
}

func HandleFightResult(account *AccountData, reader *bytes.Reader) {
	var guid float64
	var ft int32
	pack.Read(reader, &guid, &ft)

	send2Server(account, proto.Fight, proto.FightCGetAwards, pack.GetBytes(ft))
}
