package game

import (
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type Robot struct {
	game   *Game
	AiUser player.RobotInterface
	user   *data.User
}

func NewRobot(game *Game, user *data.User) *Robot {
	return &Robot{
		game: game, user: user,
	}
}

///////////////////////
var redRobCountMap = make(map[int64]int) // redId => 安排来抢包的人数
var redRobCountLock sync.Mutex

func SetRedRobCountMap(redId int64) {
	redRobCountLock.Lock()
	redRobCountMap[redId]++
	redRobCountLock.Unlock()
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	return
	switch subCmd {
	case global.S2C_SEND_RED:
		//return
		realUserCount := robot.game.getRealUserCount()
		if realUserCount == 0 || realUserCount > 10 {
			if rand.RandInt(1, 100) <= 75 {
				return
			}
		}
		time.AfterFunc(time.Second, func() {
			robot.procAiRobRed(buffer)
		})
	case global.S2C_CHECK_OVERDUE_RED:
		//return
		realUserCount := robot.game.getRealUserCount()
		if realUserCount == 0 || realUserCount > 10 {
			if rand.RandInt(1, 100) <= 50 {
				return
			}
		}
		//log.Traceln("收到 S2C_CHECK_OVERDUE_RED ")
		time.AfterFunc(500*time.Millisecond, func() {
			robot.procCheckOverdueRed(buffer)
		})
	}
}

//定期检查屏幕上存在已久的红包
func (robot *Robot) procCheckOverdueRed(buffer []byte) {
	var s2cSendRed msg.S2CRedId
	if err := proto.Unmarshal(buffer, &s2cSendRed); err != nil {
		//log.Traceln("procAiRobRed err : ", err)
		return
	}
	red := robot.game.GetRedListMap(s2cSendRed.RedId)
	if red == nil {
		//log.Traceln("procCheckOverdueRed red is nil ", s2cSendRed.RedId)
		return
	}
	go robot.robRed(red, 0)

}

//处理机器人抢包
func (robot *Robot) procAiRobRed(buffer []byte) {
	var s2cSendRed msg.S2CRedInfo
	if err := proto.Unmarshal(buffer, &s2cSendRed); err != nil {
		log.Traceln("procAiRobRed err : ", err)
		return
	}
	red := robot.game.GetRedListMap(s2cSendRed.RedId)
	if red == nil {
		log.Traceln("red is nil ", s2cSendRed.RedId)
		return
	}
	redRobCountLock.Lock()
	redRobCount := redRobCountMap[red.Id]
	redRobCountLock.Unlock()

	if config.AiConfig.IsRobotRobOn && redRobCount == 0 && robot.user.User.GetID() != red.sender.User.GetID() {
		SetRedRobCountMap(red.Id)
		robot.robRed(red, 0)

		//检查数据是否过期，过期则删除
		robot.game.Table.AddTimer(3000, func() {
			redRobCountLock.Lock()
			for redId, count := range redRobCountMap {
				if count != 0 {
					//log.Traceln("删除redRobCountMap")
					delete(redRobCountMap, redId)
				}
			}
			redRobCountLock.Unlock()
		})
	}

}

//寻找对应作弊率下的抢红包概率
func (robot *Robot) robRed(red *Red, i int) {
	i++
	if i > 20 || red.RobbedCount >= red.RedFlood {
		return
	}
	robot.user.Cheat, _ = robot.game.Table.GetRoomProb()
	if robot.user.Cheat == 0 {
		robot.user.Cheat = 1000
	}
	for _, cheatJson := range config.AiConfig.RobotConfig {
		if robot.user.Cheat == cheatJson.Cheat {
			robot.aiRob(cheatJson, red)
			break
		}
	}
	//一直抢该红包，直到该红包被抢完
	if red.RobbedCount < red.RedFlood && robot.game.getRealUserCount() >= 1 {
		robot.robRed(red, i)
		time.Sleep(10 * time.Millisecond)
		//time.AfterFunc(10*time.Millisecond, func() {
		//
		//})
	}
}

//机器人抢红包
func (robot *Robot) aiRob(cheatJson *config.CheatJson, red *Red) {
	robConfig := cheatJson.Rob1Config // todo 机器人性格，后面改为随机分配
	isExed := false
	for _, v := range robConfig {
		if rand.RateToExecWithIn(v.Rate, global.WAN_RATE_TOTAL) {
			robot.aiRateRob(v.InterTime, red)
			isExed = true
			break
		}
	}
	if !isExed {
		robot.aiRateRob(robConfig[len(robConfig)-1].InterTime, red)
	}
}

//抢包
func (robot *Robot) aiRateRob(interTime int, red *Red) {
	i := 0
	for {
		i += 100
		if i > 1000 {
			//log.Traceln("超过1000ms，退出")
			return
		}
		if i >= interTime {
			//log.Traceln("jiqiren : ",robot.user.Id," 进行抢包：",red.Id,time.Now())
			_ = robot.AiUser.SendMsgToServer(global.C2S_ROB_RED, &msg.C2SRobRed{RedId: red.Id})
			time.Sleep(100 * time.Millisecond)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}
