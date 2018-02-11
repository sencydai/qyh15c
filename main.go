package main

import (
	"bytes"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/sencydai/qyh15c/encrypt"
	"github.com/sencydai/qyh15c/pack"
	"github.com/sencydai/utils/timer"
	"math"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"time"
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

var (
	accountNamePrefix = "test"
	sendData          = make(chan *SendInfo, 1000)

	startIndex = flag.Int("s", 1, "测试账号起始索引")
	endIndex   = flag.Int("e", 1, "测试账号结束索引")
	scheme     = flag.String("w", "ws", "ws or wss")
	host       = flag.String("h", "localhost:9401", "服务器地址")
	serverId   = flag.Int("i", 1, "服务器id")
	delay      = flag.Int("d", 30, "主线副本挑间隔")
)

type Socket struct {
	socketStatus connectStatus
	pid          uint32
	conn         *websocket.Conn
	isClose      bool

	accountName string
	accountId   int32
	actorId     float64

	encrypt *encrypt.Encrypt
	timers  []*timer.Timer
	data    map[string]interface{}
}

type SendInfo struct {
	socket *Socket
	data   []byte
}

func startClient(i int) {
	u := url.URL{Scheme: *scheme, Host: *host}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err.Error())
		timer.After(time.Second*30, startClient, i)
		return
	}
	addr := c.LocalAddr().String()
	sk := &Socket{conn: c}
	sk.accountName = fmt.Sprintf("%s%d", accountNamePrefix, i)
	sk.data = make(map[string]interface{})
	sk.encrypt = encrypt.NewEncrypt()
	sk.timers = make([]*timer.Timer, 0)

	//读数据
	go func(sk *Socket) {
		for {
			mt, data, err := sk.conn.ReadMessage()
			if err != nil {
				fmt.Println(addr, " recv error ", err.Error())
				break
			}
			if mt != websocket.BinaryMessage {
				fmt.Println(addr, " recv error message type")
				break
			}
			if sk.socketStatus < statusCommunication {
				sendKey2Server(sk, data)
				continue
			}
			if len(data) < HEAD_SIZE {
				fmt.Println(addr, " recv error head size")
				break
			}
			reader := bytes.NewReader(data)
			var tag int32
			pack.Read(reader, &tag)
			if tag != DEFAULT_TAG {
				fmt.Println(addr, " recv error default tag")
				break
			}

			var dataLen int32
			pack.Read(reader, &dataLen)
			if dataLen < 2 {
				fmt.Println(addr, " recv error data len")
				break
			}
			data = data[HEAD_SIZE:]
			if int(dataLen) != len(data) {
				fmt.Println(addr, " recv error data len")
				break
			}
			reader.Reset(data)
			var sysId byte
			var cmdId byte
			pack.Read(reader, &sysId, &cmdId)

			HandleMsg(sk, sysId, cmdId, reader)
		}

		sk.isClose = true
		sk.conn.Close()
	}(sk)

	onConnected(sk)

	for {
		select {
		case <-time.After(time.Millisecond * 50):
			if sk.isClose {
				fmt.Println(addr, " close ", sk.accountId, sk.accountName, int64(sk.actorId))
				for _, t := range sk.timers {
					t.Stop()
				}
				timer.After(time.Second*30, startClient, i)
				return
			}
		}
	}
}

func send2Server(s *Socket, sysId, cmdId byte, data []byte) {
	if s.isClose {
		return
	}

	s.pid++
	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, DEFAULT_TAG, int32(0), int16(0), DEFAULT_CRC_KEY, s.pid, sysId, cmdId)
	if len(data) > 0 {
		pack.Write(writer, data)
	}
	data = writer.Bytes()
	Len := pack.GetBytes(int32(len(data) - HEAD_SIZE))
	for i := 0; i < len(Len); i++ {
		data[i+4] = Len[i]
	}
	msgCK := pack.GetBytes(s.encrypt.GetCRC16ByPos(data, HEAD_SIZE, 0))
	for i := 0; i < len(msgCK); i++ {
		data[i+8] = msgCK[i]
	}
	headerCRC := pack.GetBytes(s.encrypt.GetCRC16(data, HEAD_SIZE))
	for i := 0; i < len(headerCRC); i++ {
		data[i+10] = headerCRC[i]
	}
	s.encrypt.Encode(data, 8, 4)

	sendData <- &SendInfo{s, data}
}

func onConnected(sk *Socket) {
	sk.socketStatus = statusChecking
	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, sk.encrypt.GetSelfSalt())
	sendData <- &SendInfo{sk, writer.Bytes()}
}

func sendKey2Server(sk *Socket, data []byte) {
	reader := bytes.NewReader(data)
	var value uint32
	binary.Read(reader, binary.LittleEndian, &value)
	sk.encrypt.SetTargetSalt(value)

	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, sk.encrypt.GetCheckKey())
	sendData <- &SendInfo{sk, writer.Bytes()}

	sk.socketStatus = statusCommunication

	sendCheckAccount(sk, sk.accountName, "e10adc3949ba59abbe56e057f20f883e")
}

func sendCheckAccount(sk *Socket, user, pwd string) {
	writer := bytes.NewBuffer([]byte{})
	pack.Write(writer, int32(*serverId), user, pwd)

	//登陆
	send2Server(sk, 255, 1, writer.Bytes())
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().Unix())

	go func() {
		for sd := range sendData {
			sk, data := sd.socket, sd.data
			if !sk.isClose {
				sk.conn.WriteMessage(websocket.BinaryMessage, data)
			}
		}
	}()

	start := math.Max(1, float64(*startIndex))
	start = math.Min(start, 10000)

	end := math.Max(1, float64(*endIndex))
	end = math.Min(end, 10000)

	for i := int(start); i <= int(end); i++ {
		go startClient(i)
		time.Sleep(time.Millisecond * 200)
	}

	signalC := make(chan os.Signal, 1)
	signal.Notify(signalC, os.Interrupt, os.Kill)

	<-signalC
}
