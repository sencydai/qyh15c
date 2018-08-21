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
	"github.com/sencydai/gameworld/proto/encrypt"
	"github.com/sencydai/gameworld/proto/pack"
)

type connectStatus int

const (
	statusConnecting    connectStatus = 1
	statusChecking      connectStatus = 2
	statusCommunication connectStatus = 3
	statusDisconnect    connectStatus = 4
)

type GlobalConfig struct {
	NamePrefix  string
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
	gConfigs     = &GlobalConfig{}
	serverActors = make(map[int64]int)
)

func startClient(i int) {
	u := url.URL{Scheme: gConfigs.Scheme, Host: gConfigs.Host}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		After(nil, fmt.Sprintf("startClient_%d", i), 15, startClient, i)
		return
	}
	account := &Account{
		conn:        conn,
		connStatus:  statusChecking,
		encrypt:     encrypt.NewEncrypt(),
		accountName: fmt.Sprintf("%s%d", gConfigs.NamePrefix, i),
		data:        make(map[string]interface{}),
	}

	defer func() {
		if err := recover(); err != nil {
			log.Print(account, err)
		}
		account.Close()
		After(nil, fmt.Sprintf("startClient_%d", i), 15, startClient, i)
	}()

	account.onConnect()

	buff := make([]byte, 0)
	reader := bytes.NewReader(buff)
	//读数据
	for {
		_, data, err := account.conn.ReadMessage()
		if err != nil {
			log.Printf(account, "recv error: %s", err.Error())
			break
		}
		if account.connStatus < statusCommunication {
			account.setTargetSalt(data)
			continue
		}
		buff = append(buff, data...)
		if len(buff) < pack.HEAD_SIZE {
			continue
		}
		reader.Reset(buff)
		var tag int
		pack.Read(reader, &tag)
		if tag != pack.DEFAULT_TAG {
			log.Printf(account, "recv error: %s", "error default tag")
			break
		}

		var dataLen int
		pack.Read(reader, &dataLen)
		if dataLen < 2 {
			log.Printf(account, "recv error: %s", "error data len")
			break
		}
		size := pack.HEAD_SIZE + dataLen
		if len(data) < size {
			continue
		}
		data = buff[pack.HEAD_SIZE:size]
		buff = buff[size:]
		reader.Reset(data)
		var sysId byte
		var cmdId byte
		pack.Read(reader, &sysId, &cmdId)

		HandleServereMsg(account, sysId, cmdId, reader)
	}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Print(nil, err)
		}
		log.Close()
	}()

	rand.Seed(time.Now().Unix())
	if data, err := ioutil.ReadFile("config.json"); err != nil {
		log.Print(nil, err.Error())
		return
	} else if err = json.Unmarshal(data, gConfigs); err != nil {
		log.Print(nil, err.Error())
		return
	}

	for i := gConfigs.StartIndex; i < (gConfigs.StartIndex + gConfigs.ClientCount); i++ {
		go startClient(i)
		time.Sleep(time.Millisecond * 200)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, os.Kill)

	<-signalC
}
