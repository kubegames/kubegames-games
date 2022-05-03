package server

import (
	"go-game-sdk/example/game_buyu/980201/data"
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

var waitGroup sync.WaitGroup

var userOnlineMap = make(map[int64]*data.User)

func GetUser(uid int64) *data.User {
	return userOnlineMap[uid]
}
func SetUser(user *data.User) {
	//userOnlineMap[user.Userinfo.UserId] = user
}

type Server struct {
	Conf int32
}

func (self *Server) InitTable(table table.TableInterface) {
	tableLogic := new(TableLogic)
	tableLogic.init(table)
	table.BindGame(tableLogic)
}

func (self *Server) UserExit(user player.PlayerInterface) {

}

//func (self *Server) Accept(cli inter.InteClient) {
//	cli.Start(func(mainCmd int32, subCmd int32, buffer []byte) {
//		log.Traceln("maincmd : ", mainCmd, "sub cmd : ", subCmd, " buffer : ", string(buffer))
//		if mainCmd != 2 {
//			log.Traceln("框架消息，不处理")
//			return
//		}
//		switch msg2.MsgId(subCmd) {
//		//case msg2.MsgId_INTO_ROOM_Req:
//		//	log.Traceln("进入房间")
//		//	testMsg := &msg2.EnterRoomReq{}
//		//	err := proto.Unmarshal(buffer, testMsg)
//		//	if err != nil {
//		//		log.Warnf(err.Error())
//		//		cli.SendMsgToClinet([]byte(err.Error()))
//		//	} else {
//		//		procIntoRoom(cli, testMsg)
//		//	}
//		case msg2.MsgId_SHOOT_Req:
//			log.Traceln("射击")
//			testMsg := &msg2.ShootReq{}
//			if err := proto.Unmarshal(buffer, testMsg); err != nil {
//				cli.SendMsgToClinet([]byte(err.Error()))
//				return
//			}
//			shoot(cli, testMsg)
//		//case msg2.MsgId_EXIST_ROOM_Req:
//		//	log.Traceln("退出")
//		//	testMsg := &msg2.ExistRoomReq{}
//		//	if err := proto.Unmarshal(buffer, testMsg); err != nil {
//		//		cli.SendMsgToClinet([]byte(err.Error()))
//		//		return
//		//	}
//		//	procLeaveTable(cli, testMsg)
//
//		case msg2.MsgId_HIT_Req:
//			log.Traceln("击中")
//			testMsg := &msg2.HitReq{}
//			if err := proto.Unmarshal(buffer, testMsg); err != nil {
//				cli.SendMsgToClinet([]byte(err.Error()))
//				return
//			}
//			hit(cli, testMsg)
//
//		default:
//			log.Traceln("非法游戏协议: ", subCmd)
//			cli.SendMsgToClinet([]byte("非法游戏协议"))
//		}
//
//	}, func() {
//		log.Tracef("exit")
//	})
//}
//
//func (self *Server) Exit() {
//	log.Tracef("server exit")
//}
//
///*
//*测试服务端代码
// */
//func StartServer(port int, maxConn int32, heartBeat int) {
//	log.Traceln("ws监听端口：", port)
//	ser := &Server{}
//	server := libnet.NewServer(port, maxConn, heartBeat, ser)
//	go func() {
//		for {
//			//log.Tracef("client:%d", server.GetOnlineNum())
//			time.Sleep(time.Second * 5)
//		}
//	}()
//	server.Run()
//}
//
///*
//*测试客户端
// */
//func Client() {
//	for n := 0; n < 1; n++ {
//		waitGroup.Add(1)
//		go connect(n)
//		time.Sleep(time.Microsecond * 1000)
//	}
//	waitGroup.Wait()
//	log.Tracef("exit")
//}
//
//func connect(num int) {
//	defer waitGroup.Done()
//	c, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8082", nil)
//	if err != nil {
//		log.Errorf("dial:%v", err)
//		return
//	}
//	defer c.Close()
//
//	testMsg := &msg.Login{UserId: 1, UserName: "哈哈哈哈哈哈"}
//	buffer, err := proto.Marshal(testMsg)
//	if err != nil {
//		log.Warnf(err.Error())
//	}
//	frameMsg := &msg.FrameMsg{MainCmd: 2, SubCmd: 2, Buff: buffer}
//	data, err := proto.Marshal(frameMsg)
//	if err != nil {
//		log.Warnf(err.Error())
//	}
//
//	err = c.WriteMessage(websocket.TextMessage, data)
//	if err != nil {
//		log.Errorf(err.Error())
//		c.Close()
//		return
//	}
//	for {
//		_, buffer, err := c.ReadMessage()
//		if err != nil {
//			log.Errorf(err.Error())
//			c.Close()
//			return
//		}
//		//log.Tracef("%d:%s", num, string(msg))
//		//time.Sleep(time.Second * 200)
//		framMsg := &msg.FrameMsg{}
//		err = proto.Unmarshal(buffer, framMsg)
//		if err != nil {
//			log.Warnf(err.Error())
//		}
//		log.Tracef("recvice msg %d:%v", framMsg.MainCmd, framMsg)
//
//		testMsg := &msg.Login{}
//		err = proto.Unmarshal(framMsg.Buff, testMsg)
//		if err != nil {
//			log.Warnf(err.Error())
//		}
//		log.Tracef("%d:%s", testMsg.UserId, testMsg.UserName)
//
//		time.Sleep(time.Second * 100)
//		err = c.WriteMessage(websocket.TextMessage, data)
//		if err != nil {
//			log.Errorf(err.Error())
//			c.Close()
//			return
//		}
//
//	}
//}
