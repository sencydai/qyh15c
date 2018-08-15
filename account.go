package main

import (
	"bytes"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sencydai/gameworld/proto/encrypt"
	"github.com/sencydai/gameworld/proto/pack"
	proto "github.com/sencydai/gameworld/proto/protocol"
)

const (
	keyBag  = "bag"
	keyLord = "lord"
	keyHero = "hero"
)

type Account struct {
	conn       *websocket.Conn
	connStatus connectStatus
	encrypt    *encrypt.Encrypt
	pid        uint32
	closed     bool
	lock       sync.RWMutex

	accountName string
	accountId   int
	actorId     int64

	data map[string]interface{}
}

func (account *Account) Close() {
	account.lock.Lock()
	defer account.lock.Unlock()

	if account.closed {
		return
	}
	account.closed = true
	account.conn.Close()

	StopAccountTimers(account)
}

func (account *Account) IsClose() bool {
	account.lock.RLock()
	defer account.lock.RUnlock()

	return account.closed
}

func (account *Account) onConnect() {
	account.conn.WriteMessage(websocket.BinaryMessage, pack.GetBytes(account.encrypt.GetSelfSalt()))
}

func (account *Account) setTargetSalt(data []byte) {
	reader := bytes.NewReader(data)
	var value uint32
	pack.Read(reader, &value)
	account.encrypt.SetTargetSalt(value)

	account.conn.WriteMessage(websocket.BinaryMessage, pack.GetBytes(account.encrypt.GetCheckKey()))
	account.connStatus = statusCommunication

	account.send(proto.System, proto.SystemCLogin, gConfigs.ServerId, account.accountName, "e10adc3949ba59abbe56e057f20f883e")
}

var headerData = pack.GetBytes(pack.DEFAULT_TAG, 0, int16(0), pack.DEFAULT_CRC_KEY)

func (account *Account) send(sysId, cmdId byte, datas ...interface{}) {
	account.lock.Lock()
	defer account.lock.Unlock()
	if account.closed {
		return
	}

	account.pid++
	writer := pack.NewWriter(headerData, account.pid, sysId, cmdId)
	pack.Write(writer, datas...)

	data := writer.Bytes()
	Len := pack.GetBytes(len(data) - pack.HEAD_SIZE)
	copy(data[4:], Len)

	msgCK := pack.GetBytes(account.encrypt.GetCRC16ByPos(data, pack.HEAD_SIZE, 0))
	copy(data[8:], msgCK)

	headerCRC := pack.GetBytes(account.encrypt.GetCRC16(data, pack.HEAD_SIZE))
	copy(data[10:], headerCRC)

	account.encrypt.Encode(data, 8, 4)

	if account.conn.WriteMessage(websocket.BinaryMessage, data) == nil {
		//log.Printf(account, "send %d %d", sysId, cmdId)
	}
}
