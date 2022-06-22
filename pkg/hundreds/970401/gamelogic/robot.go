package gamelogic

import (
	"game_LaBa/benzbmw/config"
	"game_LaBa/benzbmw/model"
	proto "game_LaBa/benzbmw/msg"
	"game_frame_v2/game/clock"

	protocol "github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"time"
)

var count = [12]int{}

type Robot struct {
	game  *Game
	table table.TableInterface //桌子
	user  player.RobotInterface

	BetGoldThisSet int64 // 本局下注金额

	BetArrGold [BET_AREA_LENGHT]int64 // 下注区域下注的金额，对应0-11

	TimerJob *clock.Job //时间定时器

	BetCount    int32 // 下注次数
	NotBetCount int32 // 未下注局数

	shouldBetTimes int32 // 本局应该下注次数

	// 随机部分
	betNum int
}

func NewRobot(game *Game) *Robot {
	// conf:=deepcopy.Copy(config.RobotConf).(config.RobotConfig)
	return &Robot{
		game:  game,
		table: game.table,
	}
}

func (robot *Robot) BindUser(user player.RobotInterface) {
	robot.user = user
}

//游戏消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(proto.SendToClientMessageType_Status):
		{
			robot.BeforBet(buffer)
		}
		break
	}
}

func (robot *Robot) BeforBet(buf []byte) {

	msg := new(proto.StatusMessage)
	protocol.Unmarshal(buf, msg)
	if msg.Status == int32(proto.GameStatus_BetStatus) {
		robot.rand()
		robot.TimerJob, _ = robot.user.AddTimer(time.Duration(config.RobotConf.RandGap()), robot.DoBet)
	} else if msg.Status == int32(proto.GameStatus_EndBetMovie) {
		robot.Reset()
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
		}
		robot.TimerJob = nil
	}

}

func (robot *Robot) rand() {
	robot.betNum = config.RobotConf.BetTimes.Rand().Times
}

func (robot *Robot) DoBet() {
	if robot.game.Status != int32(proto.GameStatus_BetStatus) {
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
			robot.TimerJob = nil
		}
		return
	}
	if robot.BetCount > int32(robot.betNum) {
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
			robot.TimerJob = nil
		}
		return
	}

	// 3.随机下注筹码区
	betType := config.RobotConf.BetGold.Rand().Index
	betGold := robot.game.BetArr[betType]

	betArr := make([]int64, BET_AREA_LENGHT)
	for i := range robot.game.UserBetInfo {
		betArr[i] = robot.game.UserBetInfo[i] + robot.game.AIBetInfo[i]
	}

	var minID, maxID int
	min := betArr[0]
	var max int64
	for id, v := range betArr {
		if v > max {
			maxID = id
		}
		if v < min {
			minID = id
		}
	}

	var betAreaIndex int
	diff := betArr[maxID] - betArr[minID]
	if diff > config.RobotConf.Diff {
		// 超过差值
		if model.Rand(PROB_BASE) <= config.RobotConf.Small { // 下小概率
			betAreaIndex = minID
		} else {
		RELOOP:
			betAreaIndex = config.RobotConf.BetArea.Rand().Index
			if betAreaIndex == minID {
				goto RELOOP
			}
		}
	} else {
		betAreaIndex = config.RobotConf.BetArea.Rand().Index
	}
	// 4.随机下注区域

	count[betAreaIndex]++

	if robot.BetArrGold[betAreaIndex]+betGold > robot.game.BetLimitInfo.LimitPerUser ||
		robot.BetGoldThisSet+betGold > robot.game.BetLimitInfo.AllLimitPerUser ||
		robot.game.UserBetInfo[betAreaIndex]+robot.game.AIBetInfo[betAreaIndex]+betGold > robot.game.BetLimitInfo.AllLimitPerArea {
		// fmt.Printf("机器人[%v]达到下注上限，不进行下注==========================\n", robot.user.GetID())
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
			robot.TimerJob = nil
		}
		return
	}
	if robot.user.GetScore() < robot.game.BetArr[0] {
		// fmt.Printf("机器人【%v】携带金币不足，不进行下注==========================\n", robot.user.GetID())
		return
	}

	robot.BetArrGold[betAreaIndex] += betGold

	// 给user下注
	if u, ok := robot.game.UserMap[robot.user.GetID()]; ok {

		u.NotBetCount = 0
		u.BetGoldNow += betGold
		// for _, v := range robot.game.topUser {
		// 	if v.user.GetID() == robot.user.GetID() {
		// 		// 如果该机器人是上座机器人，则发送下注消息
		// 		msg := new(proto.UserBet)
		// 		msg.BetType = int32(betAreaIndex)
		// 		msg.BetIndex = int32(betType)
		// 		msg.UserID = u.user.GetID()
		// 		robot.game.table.Broadcast(int32(proto.SendToClientMessageType_BetRet), msg)
		// 	}
		// }

		robot.sendMsg(u.user.GetID(), betAreaIndex, betType)
	}
	robot.BetCount++
	robot.NotBetCount = 0
	robot.AddBetTimer()
}

func (robot *Robot) sendMsg(userId int64, typ, index int) {
	msg := new(proto.UserBet)
	msg.BetType = int32(typ)
	msg.BetIndex = int32(index)
	msg.UserID = userId
	//	log.Tracef("发送机器人下注")
	if err := robot.user.SendMsgToServer(int32(proto.ReceiveMessageType_DoBet), msg); err != nil {
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
			robot.TimerJob = nil
		}
	}
}

func (robot *Robot) AddBetTimer() {
	robot.TimerJob, _ = robot.user.AddTimer(time.Duration(config.RobotConf.RandGap()), robot.DoBet)
}

func (robot *Robot) Reset() {
	robot.BetArrGold = [BET_AREA_LENGHT]int64{}
	robot.BetCount = 0
	robot.BetGoldThisSet = 0
}
