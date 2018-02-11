package main

import (
	"bytes"
	"fmt"
	"github.com/sencydai/qyh15c/pack"
	"github.com/sencydai/utils/timer"
	"math/rand"
	"time"
)

type MsgHandler func(*Socket, *bytes.Reader)

var (
	HandleMgr = make(map[int]MsgHandler)
)

func init() {
	//登陆
	RegHandle(255, 1, HandleLogin)

	//查询角色列表
	RegHandle(255, 4, HandleCheckActorList)
	//随机名称
	RegHandle(255, 6, HandleRandomActorName)
	//创建角色
	RegHandle(255, 2, HandleCreateActor)
	//进入游戏
	RegHandle(255, 5, HandleLoginSuccess)
	//战斗结果
	RegHandle(52, 1, HandleFightResult)

}

func RegHandle(sysId, cmdId byte, handle MsgHandler) {
	mark := (int(sysId) << 8) + int(cmdId)
	HandleMgr[mark] = handle
}

func HandleMsg(sk *Socket, sysId, cmdId byte, reader *bytes.Reader) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("handle Msg %d %d error: %v\n", sysId, cmdId, err)
		}
	}()

	mark := (int(sysId) << 8) + int(cmdId)
	handle, ok := HandleMgr[mark]
	if ok {
		handle(sk, reader)
	}
}

func HandleLogin(sk *Socket, reader *bytes.Reader) {
	var code byte
	pack.Read(reader, &code)
	if code != 0 {
		fmt.Printf("HandleLogin error: %s code(%d)\n", sk.accountName, code)
		sk.conn.Close()
		return
	}
	//查询角色列表
	send2Server(sk, 255, 4, nil)
}

func HandleCheckActorList(sk *Socket, reader *bytes.Reader) {
	var accountId int32
	var code int32
	pack.Read(reader, &accountId, &code)
	if code < 0 {
		return
	}
	sk.accountId = accountId
	if code == 0 {
		randomActorName(sk)
		return
	}

	var actorId float64
	var name string
	var head, sex, level, job, camp int32

	pack.Read(reader, &actorId, &name, &head, &sex, &level, &job, &camp)

	sendLoginGame(sk, actorId)
}

func randomActorName(s *Socket) {
	send2Server(s, 255, 6, pack.GetBytes(rand.Int31n(2)+1))
}

func HandleRandomActorName(s *Socket, reader *bytes.Reader) {
	var code int32
	pack.Read(reader, &code)
	if code != 0 {
		randomActorName(s)
		return
	}

	var sex int32
	var name string
	pack.Read(reader, &sex, &name)

	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, name, sex, int32(1), int32(1), byte(1), "pf_auto")
	send2Server(s, 255, 2, writer.Bytes())
}

func HandleCreateActor(s *Socket, reader *bytes.Reader) {
	var actorId float64
	var code int32
	pack.Read(reader, &actorId, &code)
	if code != 0 {
		s.conn.Close()
		return
	}
	sendLoginGame(s, actorId)
}

func sendLoginGame(s *Socket, actorId float64) {
	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, actorId, "pt_auto")
	s.actorId = actorId
	send2Server(s, 255, 5, writer.Bytes())
}

func HandleLoginSuccess(s *Socket, reader *bytes.Reader) {
	var code int32
	pack.Read(reader, &code)

	fmt.Printf("accountName(%s) login game code(%d),actorId(%d)\n", s.accountName, code, int64(s.actorId))
	if code != 0 {
		s.conn.Close()
		return
	}
	d := time.Duration(*delay)
	timer.After(time.Second*d, enterMainFuben, s)
}

func enterMainFuben(s *Socket) {
	send2Server(s, 181, 2, nil)
}

func HandleFightResult(s *Socket, reader *bytes.Reader) {
	var guid float64
	var ft int32
	pack.Read(reader, &guid, &ft)

	send2Server(s, 52, 2, pack.GetBytes(ft))

	d := time.Duration(*delay)
	timer.After(time.Second*d, enterMainFuben, s)
}
