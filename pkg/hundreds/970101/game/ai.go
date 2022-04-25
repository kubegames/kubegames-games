package game

import (
	"common/rand"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"github.com/golang/protobuf/proto"
	"fmt"
)

//机器人发红包
func (game *Game) AiSendRedTimer() {
	if game.isClosed {
		fmt.Println("房间正在关闭，不再发红包")
		return
	}
	//return
	isSend, _, amount := false, int64(0), int64(0)
	switch game.sendSecond {
	case 2:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S2Send, config.AiSendConfig.S2SendCount1, config.AiSendConfig.S2SendCount2,
			config.AiSendConfig.S2SendAmount1, config.AiSendConfig.S2SendAmount2, config.AiSendConfig.S2SendAmount3)
	case 3:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S3Send, config.AiSendConfig.S3SendCount1, config.AiSendConfig.S3SendCount2,
			config.AiSendConfig.S3SendAmount1, config.AiSendConfig.S3SendAmount2, config.AiSendConfig.S3SendAmount3)
	case 4:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S4Send, config.AiSendConfig.S4SendCount1, config.AiSendConfig.S4SendCount2,
			config.AiSendConfig.S4SendAmount1, config.AiSendConfig.S4SendAmount2, config.AiSendConfig.S4SendAmount3)
	case 5:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S5Send, config.AiSendConfig.S5SendCount1, config.AiSendConfig.S5SendCount2,
			config.AiSendConfig.S5SendAmount1, config.AiSendConfig.S5SendAmount2, config.AiSendConfig.S5SendAmount3)
	case 6:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S6Send, config.AiSendConfig.S6SendCount1, config.AiSendConfig.S6SendCount2,
			config.AiSendConfig.S6SendAmount1, config.AiSendConfig.S6SendAmount2, config.AiSendConfig.S6SendAmount3)
	case 7:
		isSend, _, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S7Send, config.AiSendConfig.S7SendCount1, config.AiSendConfig.S7SendCount2,
			config.AiSendConfig.S7SendAmount1, config.AiSendConfig.S7SendAmount2, config.AiSendConfig.S7SendAmount3)
	}
	if isSend && game.redList.Len() <= config.AiSendConfig.MinLeftRedCount {
		ai := game.randGetAi()
		if ai == nil {
			return
		}
		//fmt.Println("机器人发红包..............当前红包数量：",len(game.redListMap))
		c2sMsg := &msg.C2SSendRed{MineNum: int64(rand.RandInt(0, 9)), Amount: amount, UserId: ai.Id}
		c2sMsgB, _ := proto.Marshal(c2sMsg)
		game.ProcSendRed(c2sMsgB, ai)
	}
}

//获取发送红包的数量和金额
func (game *Game) aiGetSendRedCountAmount(sendRate, count1, c2, amount1, a2, a3 int) (isSend bool, count, amount int64) {
	if rand.RateToExecWithIn(sendRate, global.WAN_RATE_TOTAL) {
		isSend = true
		if rand.RateToExecWithIn(count1, global.WAN_RATE_TOTAL) {
			count = config.AiSendConfig.SendCount1
		} else if rand.RateToExecWithIn(c2, global.WAN_RATE_TOTAL) {
			count = config.AiSendConfig.SendCount2
		} else {
			count = config.AiSendConfig.SendCount3
		}

		if rand.RateToExecWithIn(amount1, global.WAN_RATE_TOTAL) {
			amount = 0
		} else if rand.RateToExecWithIn(a2, global.WAN_RATE_TOTAL) {
			amount = 1
		} else if rand.RateToExecWithIn(a3, global.WAN_RATE_TOTAL) {
			amount = 2
		} else {
			amount = 3
		}
	}
	return
}

//随机获取场上的一个机器人
func (game *Game) randGetAi() *data.User {
	if game.getAiCount() == 0 {
		//fmt.Println("机器人数量为 0 ")
		//game.Table.GetRobot()
		return nil
	}
	index := 0
	aiArr := make([]*data.User, 0)
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() {
			aiArr = append(aiArr, user)
		}
	}
	if len(aiArr) > 1 {
		index = rand.RandInt(0, len(aiArr)-1)
	}
	return aiArr[index]
}

//随机获取场上的几个机器人
func (game *Game) randGetAis(count int) []*data.User {
	if game.getAiCount() == 0 {
		//fmt.Println("机器人数量为 0 ")
		//game.Table.GetRobot()
		return nil
	}
	ais := make([]*data.User,0)
	index := 0
	aiArr := make([]*data.User, 0)
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() {
			aiArr = append(aiArr, user)
		}
	}
	for i:=0;i<count;i++{
		if len(aiArr) > 1 {
			index = rand.RandInt(0, len(aiArr)-1)
			ais = append(ais,aiArr[index])
		}
	}

	return ais
}

func (game *Game) getAiCount() (count int) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() {
			count++
		}
	}
	return
}
