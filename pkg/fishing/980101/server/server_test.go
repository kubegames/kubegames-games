package server_test

import (
	msg "game_frame/msg"
	"sync"
	"testing"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

func TestClient(t *testing.T) {
	for n := 1; n < 2; n++ {
		waitGroup.Add(1)
		go connect(n)
		time.Sleep(time.Microsecond * 1000)
	}
	waitGroup.Wait()
	log.Tracef("exit")
}

var waitGroup sync.WaitGroup

func connect(num int) {
	defer waitGroup.Done()
	c, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080?token=123", nil)
	if err != nil {
		log.Errorf("dial:%v", err)
		return
	}
	defer c.Close()

	testMsg := &msg.C2SLogin{
		UserId: int64(num),
		//设备码
		EquipmentCode: "112233",
		//游戏名称
		GameName: "1232435",
	} //&msg.Login{UserId: 2, UserName: "哈哈哈哈哈哈"}
	buffer, err := proto.Marshal(testMsg)
	if err != nil {
		log.Warnf(err.Error())
	}
	frameMsg := &msg.FrameMsg{MainCmd: 1, SubCmd: 1, Buff: buffer}
	data, err := proto.Marshal(frameMsg)
	if err != nil {
		log.Warnf(err.Error())
	}

	err = c.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Errorf(err.Error())
		c.Close()
		return
	}
	for {
		_, buffer, err := c.ReadMessage()
		if err != nil {
			log.Errorf(err.Error())
			c.Close()
			return
		}
		//log.Tracef("%d:%s", num, string(msg))
		//time.Sleep(time.Second * 200)
		framMsg := &msg.FrameMsg{}
		err = proto.Unmarshal(buffer, framMsg)
		if err != nil {
			log.Warnf(err.Error())
		}
		log.Tracef("recvice msg %d:%v", framMsg.SubCmd, framMsg)

		testMsg := &msg.C2SLogin{}
		err = proto.Unmarshal(framMsg.Buff, testMsg)
		if err != nil {
			log.Warnf(err.Error())
		}
		log.Tracef("abc %d:%s", testMsg.UserId, testMsg.GameName)

		time.Sleep(time.Second * 100)
		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Errorf(err.Error())
			c.Close()
			return
		}

	}
}
