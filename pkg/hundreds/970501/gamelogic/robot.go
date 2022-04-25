package gamelogic

import (
	"go-game-sdk/example/game_LaBa/970501/config"
	proto "go-game-sdk/example/game_LaBa/970501/msg"
	"go-game-sdk/inter"
	"go-game-sdk/lib/clock"
	"math/rand"

	protocol "github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Robot struct {
	game  *Game
	table table.TableInterface //桌子
	user  inter.AIUserInter

	BetGoldThisSet int64 // 本局下注金额

	BetArrGold [BET_AREA_LENGHT]int64 // 下注区域下注的金额，对应0-11

	TimerJob *clock.Job //时间定时器

	BetCount    int32 // 下注次数
	NotBetCount int32 // 未下注局数

	shouldBetTimes int32 // 本局应该下注次数

	// 随机部分
	betAreaNum int
	betNum     int
}

func NewRobot(game *Game) *Robot {
	// conf:=deepcopy.Copy(config.RobotConf).(config.RobotConfig)
	return &Robot{
		game:  game,
		table: game.table,
	}
}

func (robot *Robot) BindUser(user inter.AIUserInter) {
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
		robot.TimerJob, _ = robot.user.AddTimer(int64(config.RobotConf.RandGap()), robot.DoBet)
	} else if msg.Status == int32(proto.GameStatus_EndBetMovie) {
		robot.Reset()
		if robot.TimerJob != nil {
			robot.TimerJob.Cancel()
		}
		robot.TimerJob = nil
	}

}

func (robot *Robot) rand() {
	robot.betAreaNum = rand.Intn(BET_AREA_LENGHT)
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

	// 4.随机下注区域
	betAreaIndex := config.RobotConf.BetArea.Rand().Index

	if robot.BetArrGold[betAreaIndex]+betGold > robot.game.BetLimitInfo.LimitPerUser ||
		robot.BetGoldThisSet+betGold > robot.game.BetLimitInfo.AllLimitPerUser ||
		robot.game.UserBetInfo[betAreaIndex]+robot.game.AIBetInfo[betAreaIndex]+betGold > robot.game.BetLimitInfo.AllLimitPerArea ||
		robot.user.GetScore() < robot.game.BetArr[0] {
		robot.AddBetTimer()
		return
	}

	robot.BetArrGold[betAreaIndex] += betGold
	robot.game.AIBetInfo[betAreaIndex] += betGold

	// 给user下注
	if u, ok := robot.game.UserMap[robot.user.GetID()]; ok {
		u.BetInfo[betAreaIndex] += betGold
		u.NotBetCount = 0
		u.BetGoldNow += betGold
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
	robot.user.SendMsgToServer(int32(proto.ReceiveMessageType_DoBet), msg)
}

func (robot *Robot) AddBetTimer() {
	robot.TimerJob, _ = robot.user.AddTimer(int64(config.RobotConf.RandGap()), robot.DoBet)
}

func (robot *Robot) Reset() {
	robot.BetArrGold = [BET_AREA_LENGHT]int64{}
	robot.BetCount = 0
	robot.BetGoldThisSet = 0
}
