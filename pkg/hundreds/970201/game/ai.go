package game

import (
	"common/rand"
	"fmt"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	"github.com/golang/protobuf/proto"
)

//机器人发红包
func (game *Game) AiSendRedTimer() {
	//fmt.Println("准备发送红包，房间：",game.Table.GetLevel())
	isSend, count, amount := false, int64(0), int64(0)
	switch game.nowSecond {
	case 2:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S2Send, config.AiSendConfig.S2SendCount1, config.AiSendConfig.S2SendCount2,
			config.AiSendConfig.S2SendAmount1, config.AiSendConfig.S2SendAmount2, config.AiSendConfig.S2SendAmount3)
	case 3:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S3Send, config.AiSendConfig.S3SendCount1, config.AiSendConfig.S3SendCount2,
			config.AiSendConfig.S3SendAmount1, config.AiSendConfig.S3SendAmount2, config.AiSendConfig.S3SendAmount3)
	case 4:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S4Send, config.AiSendConfig.S4SendCount1, config.AiSendConfig.S4SendCount2,
			config.AiSendConfig.S4SendAmount1, config.AiSendConfig.S4SendAmount2, config.AiSendConfig.S4SendAmount3)
	case 5:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S5Send, config.AiSendConfig.S5SendCount1, config.AiSendConfig.S5SendCount2,
			config.AiSendConfig.S5SendAmount1, config.AiSendConfig.S5SendAmount2, config.AiSendConfig.S5SendAmount3)
	case 6:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S6Send, config.AiSendConfig.S6SendCount1, config.AiSendConfig.S6SendCount2,
			config.AiSendConfig.S6SendAmount1, config.AiSendConfig.S6SendAmount2, config.AiSendConfig.S6SendAmount3)
	case 7:
		isSend, count, amount = game.aiGetSendRedCountAmount(config.AiSendConfig.S7Send, config.AiSendConfig.S7SendCount1, config.AiSendConfig.S7SendCount2,
			config.AiSendConfig.S7SendAmount1, config.AiSendConfig.S7SendAmount2, config.AiSendConfig.S7SendAmount3)
	}
	if isSend && game.redList.Len() <= config.AiSendConfig.MinLeftRedCount {
		ai := game.randGetAi()
		if ai == nil {
			//fmt.Println("ai == nil ",game.Table.GetLevel())
			return
		}
		//fmt.Println("机器人发红包..............房间等级：",game.Table.GetLevel(),"机器人id：",ai.User.GetId())
		c2sMsg := &msg.C2SSendRed{MineNum: int64(rand.RandInt(0, 9)), Amount: amount, Count: int32(count), UserId: ai.Id}
		c2sMsgB, _ := proto.Marshal(c2sMsg)
		game.ProcSendRed(c2sMsgB, ai)
	}
}

//获取发送红包的数量和金额
func (game *Game) aiGetSendRedCountAmount(sendRate, count1, c2, amount1, a2, a3 int) (isSend bool, count, amount int64) {
	//fmt.Println("sendRate, count1, c2, amount1, a2, a3 ",sendRate, count1, c2, amount1, a2, a3,"table level : ",game.Table.GetLevel())
	if len(game.sendAmount) != 4 {
		fmt.Println("len(game.sendAmount) != 4 ",game.Table.GetLevel(),game.sendAmount)
		return
	}
	if rand.RateToExecWithIn(sendRate, global.WAN_RATE_TOTAL) {
		//fmt.Println("game.sendAmount 111 ",game.sendAmount," game id : ",game.Table.GetId())
		isSend = true
		if rand.RateToExecWithIn(count1, global.WAN_RATE_TOTAL) {
			count = config.AiSendConfig.SendCount1
		} else if rand.RateToExecWithIn(c2, global.WAN_RATE_TOTAL) {
			count = config.AiSendConfig.SendCount2
		} else {
			count = config.AiSendConfig.SendCount3
		}

		if rand.RateToExecWithIn(amount1, global.WAN_RATE_TOTAL) {
			amount = game.sendAmount[0]
			//amount = config.AiSendConfig.SendAmount[0]
		} else if rand.RateToExecWithIn(a2, global.WAN_RATE_TOTAL) {
			amount = game.sendAmount[1]
		} else if rand.RateToExecWithIn(a3, global.WAN_RATE_TOTAL) {
			amount = game.sendAmount[2]
		} else {
			amount = game.sendAmount[3]
		}
	}
	return
}

//机器人抢红包定时触发事件
func (game *Game) AiRobRedTimer() {
	//count := 0
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		//if count > 10 {
		//	return
		//}
		if user.User.IsRobot() {
			go game.aiFuncRob(user)
			//count ++
		}
	}
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		red.nowSecond++
	}
}

//机器人抢包
func (game *Game) aiFuncRob(ai *data.User) {
	//return
	//获取机器人的作弊率
	ai.Cheat ,_ = game.Table.GetRoomProb()
	if ai.Cheat == 0 {
		ai.Cheat = 1000
	}
	ai.AiRobConfig = config.AiRobConfigMap[ai.Cheat]
	robConfig := ai.AiRobConfig.RobRateArr[rand.RandInt(0, len(ai.AiRobConfig.RobRateArr)-1)]
	//fmt.Println("robConfig : ",fmt.Sprintf(`%+v`,robConfig))
	game.rateRob(ai, robConfig)

}

func (game *Game) rateRob(ai *data.User, robConfig *config.RobRate) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		switch red.nowSecond {
		case 1:
			if rand.RateToExecWithIn(robConfig.AiS1Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第1秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS1Rate)
		case 2:
			if rand.RateToExecWithIn(robConfig.AiS2Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第2秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS2Rate)
		case 3:
			if rand.RateToExecWithIn(robConfig.AiS3Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第3秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS3Rate)
		case 4:

			if rand.RateToExecWithIn(robConfig.AiS4Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第4秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS4Rate)
		case 5:

			if rand.RateToExecWithIn(robConfig.AiS5Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第5秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS5Rate)
		case 6:

			if rand.RateToExecWithIn(robConfig.AiS6Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第6秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS6Rate)
		case 7:

			if rand.RateToExecWithIn(robConfig.AiS7Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第7秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS7Rate)
		case 8:

			if rand.RateToExecWithIn(robConfig.AiS8Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第8秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS8Rate)
		case 9:
			if rand.RateToExecWithIn(robConfig.AiS9Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第9秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS9Rate)
		case 10:
			if rand.RateToExecWithIn(robConfig.AiS10Rate, global.WAN_RATE_TOTAL) {
				game.robRed(ai, red)
			}
			//fmt.Println("第10秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS10Rate)
		}
	}
}

//随机获取场上的一个机器人
func (game *Game) randGetAi() *data.User {
	if game.getAiCount() == 0 {
		fmt.Println("机器人数量为 0 ",game.Table.GetLevel())
		//game.Table.GetRobot()
		return nil
	}
	aiArr := make([]*data.User, 0)
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() {
			aiArr = append(aiArr, user)
		}
	}
	if len(aiArr) > 1 {
		return aiArr[rand.RandInt(0, len(aiArr)-1)]
	}else {
		return nil
	}
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

func (game *Game) getRealUserCount() (count int) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if !user.User.IsRobot() {
			count++
		}
	}
	return
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