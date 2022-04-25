package game

import (
	"common/rand"
	"fmt"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"game_frame_v2/game/inter"
	"github.com/golang/protobuf/proto"
	"time"
)

type Robot struct {
	game   *Game
	AiUser inter.AIUserInter
	user   *data.User
}

func NewRobot(game *Game, user *data.User) *Robot {
	return &Robot{
		game: game, user: user,
	}
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case global.S2C_START_ROB:
		//fmt.Println("机器人收到抢包信息")
		//return
		time.AfterFunc(500*time.Millisecond, func() {
			robot.procAiRobRed(buffer)
		})
	}
}

func (robot *Robot) procAiRobRed(buffer []byte) {
	var s2cSendRed msg.S2CRedInfo
	if err := proto.Unmarshal(buffer, &s2cSendRed); err != nil {
		fmt.Println("procAiRobRed err : ", err)
		return
	}
	go robot.robRed(s2cSendRed)
}

func (robot *Robot) robRed(red msg.S2CRedInfo) {
	cheat := robot.user.User.GetProb()
	if cheat == 0 {
		cheat = 3000
	}
	robot.user.AiRobConfig = config.AiRobConfigMap[cheat]
	for i := 1; i <= 10; i++ {
		if len(robot.user.AiRobConfig.RobRateArr) <= 1 {
			return
		}
		index := rand.RandInt(0, len(robot.user.AiRobConfig.RobRateArr)-1)
		if robot.user == nil || robot.user.AiRobConfig == nil {
			return
		}
		robConfig := robot.user.AiRobConfig.RobRateArr[index]
		robot.rateRob(i, red, robConfig)
		time.Sleep(time.Second)
	}
}

func (robot *Robot) rateRob(second int, red msg.S2CRedInfo, robConfig *config.RobRate) {
	switch second {
	case 1:
		if rand.RateToExecWithIn(robConfig.AiS1Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
			//game.robRed(ai,red)
		}
		//fmt.Println("第1秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS1Rate)
	case 2:
		if rand.RateToExecWithIn(robConfig.AiS2Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第2秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS2Rate)
	case 3:

		if rand.RateToExecWithIn(robConfig.AiS3Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第3秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS3Rate)
	case 4:

		if rand.RateToExecWithIn(robConfig.AiS4Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第4秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS4Rate)
	case 5:

		if rand.RateToExecWithIn(robConfig.AiS5Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第5秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS5Rate)
	case 6:

		if rand.RateToExecWithIn(robConfig.AiS6Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第6秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS6Rate)
	case 7:

		if rand.RateToExecWithIn(robConfig.AiS7Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第7秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS7Rate)
	case 8:

		if rand.RateToExecWithIn(robConfig.AiS8Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第8秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS8Rate)
	case 9:
		if rand.RateToExecWithIn(robConfig.AiS9Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第9秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS9Rate)
	case 10:
		if rand.RateToExecWithIn(robConfig.AiS10Rate, global.WAN_RATE_TOTAL) {
			if err := robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.RedId, UserId: robot.user.Id}); err != nil {
			}
		}
		//fmt.Println("第10秒 hongbao id : ",red.Id," 剩余血量：",red.RedFlood-red.RobbedCount," 抢包概率：",robConfig.AiS10Rate)
	}
}
