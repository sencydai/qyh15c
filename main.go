package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sencydai/qyh15c/encrypt"
	"github.com/sencydai/qyh15c/pack"
	proto "github.com/sencydai/qyh15c/protocol"
	"github.com/sencydai/utils/timer"
)

type connectStatus int

const (
	statusConnecting    connectStatus = 1
	statusChecking      connectStatus = 2
	statusCommunication connectStatus = 3
	statusDisconnect    connectStatus = 4

	DEFAULT_TAG     int32 = 0xccee
	DEFAULT_CRC_KEY int16 = 0x765d

	HEAD_SIZE = 12
)

type GlobalConfig struct {
	StartIndex  int
	ClientCount int
	Scheme      string
	Host        string
	ServerId    int
	FightPeriod int
	ChatPeriod  int
	MsgPeriod   int
	ChatMsgs    []string `json:chatMsgs`
}

var (
	gConfigs          = &GlobalConfig{}
	accountNamePrefix = "test"
	sendData          = make(chan *SendInfo, 1000)
)

type SendInfo struct {
	account *AccountData
	data    []byte
}

func startClient(i int) {
	u := url.URL{Scheme: gConfigs.Scheme, Host: gConfigs.Host}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		timer.After(time.Second*30, startClient, i)
		return
	}
	account := &AccountData{conn: c}
	account.accountName = fmt.Sprintf("%s%d", accountNamePrefix, i)
	account.data = make(map[string]interface{})
	account.encrypt = encrypt.NewEncrypt()
	account.timers = make([]*timer.Timer, 0)

	//读数据
	go func(account *AccountData) {
		for {
			mt, data, err := account.conn.ReadMessage()
			if err != nil {
				Printf(account, "recv error: %s", err.Error())
				break
			}
			if mt != websocket.BinaryMessage {
				Printf(account, "recv error: %s", "error message type")
				break
			}
			if account.connStatus < statusCommunication {
				sendKey2Server(account, data)
				continue
			}
			if len(data) < HEAD_SIZE {
				Printf(account, "recv error: %s", "error head size")
				break
			}
			reader := bytes.NewReader(data)
			var tag int32
			pack.Read(reader, &tag)
			if tag != DEFAULT_TAG {
				Printf(account, "recv error: %s", "error default tag")
				break
			}

			var dataLen int32
			pack.Read(reader, &dataLen)
			if dataLen < 2 {
				Printf(account, "recv error: %s", "error data len")
				break
			}
			data = data[HEAD_SIZE:]
			if int(dataLen) != len(data) {
				Printf(account, "recv error: %s", "error data len")
				break
			}
			reader.Reset(data)
			var sysId byte
			var cmdId byte
			pack.Read(reader, &sysId, &cmdId)

			HandleServereMsg(account, sysId, cmdId, reader)
		}

		account.isClose = true
		account.conn.Close()
	}(account)

	onConnected(account)

	for {
		select {
		case <-time.After(time.Millisecond * 50):
			if account.isClose {
				for _, t := range account.timers {
					t.Stop()
				}
				timer.After(time.Second*30, startClient, i)
				return
			}
		}
	}
}

func send2Server(account *AccountData, sysId, cmdId byte, data []byte) {
	if account.isClose {
		return
	}

	account.pid++
	writer := pack.NewWriter(DEFAULT_TAG, int32(0), int16(0), DEFAULT_CRC_KEY, account.pid, sysId, cmdId)
	if len(data) > 0 {
		pack.Write(writer, data)
	}
	data = writer.Bytes()
	Len := pack.GetBytes(int32(len(data) - HEAD_SIZE))
	for i := 0; i < len(Len); i++ {
		data[i+4] = Len[i]
	}
	msgCK := pack.GetBytes(account.encrypt.GetCRC16ByPos(data, HEAD_SIZE, 0))
	for i := 0; i < len(msgCK); i++ {
		data[i+8] = msgCK[i]
	}
	headerCRC := pack.GetBytes(account.encrypt.GetCRC16(data, HEAD_SIZE))
	for i := 0; i < len(headerCRC); i++ {
		data[i+10] = headerCRC[i]
	}
	account.encrypt.Encode(data, 8, 4)

	sendData <- &SendInfo{account, data}
}

func onConnected(account *AccountData) {
	account.connStatus = statusChecking
	sendData <- &SendInfo{account, pack.GetBytes(account.encrypt.GetSelfSalt())}
}

func sendKey2Server(account *AccountData, data []byte) {
	reader := bytes.NewReader(data)
	var value uint32
	pack.Read(reader, &value)
	account.encrypt.SetTargetSalt(value)

	sendData <- &SendInfo{account, pack.GetBytes(account.encrypt.GetCheckKey())}

	account.connStatus = statusCommunication

	sendCheckAccount(account, account.accountName, "e10adc3949ba59abbe56e057f20f883e")
}

func sendCheckAccount(account *AccountData, user, pwd string) {
	//登陆
	send2Server(account, proto.System, proto.SystemCLogin, pack.GetBytes(int32(gConfigs.ServerId), user, pwd))
}

func main() {
	rand.Seed(time.Now().Unix())
	if data, err := ioutil.ReadFile("config.json"); err != nil {
		fmt.Println(err.Error())
		return
	} else if err = json.Unmarshal(data, gConfigs); err != nil {
		fmt.Println(err.Error())
		return
	}

	go func() {
		for sd := range sendData {
			account, data := sd.account, sd.data
			if !account.isClose {
				account.conn.WriteMessage(websocket.BinaryMessage, data)
			}
		}
	}()

	for i := gConfigs.StartIndex; i < (gConfigs.StartIndex + gConfigs.ClientCount); i++ {
		go startClient(i)
		time.Sleep(time.Millisecond * 200)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, os.Kill)

	<-signalC
}
