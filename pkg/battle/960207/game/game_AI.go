package game

import (
	"go-game-sdk/lib/clock"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	conf "github.com/kubegames/kubegames-games/pkg/battle/960207/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/msg"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
)

// Robot 机器人结构体
type Robot struct {
	User      player.RobotInterface
	TimerJob  *clock.Job
	Cfg       conf.RobotConfig
	GameLogic *GeneralNiuniu
	IsBanker  bool
	ChairID   int32
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, gameInter table.TableHandler, robotCfg conf.RobotConfig, chairID int32) {
	robot.User = userInter
	robot.Cfg = robotCfg
	robot.GameLogic = gameInter.(*GeneralNiuniu)
	robot.ChairID = chairID
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {

	switch subCmd {
	// 游戏状态变更消息
	case int32(msg.SendToClientMessageType_S2CGameStatus):
		robot.OnGameStatus(buffer)
		break

		// 收到投注倍率信息
	case int32(msg.SendToClientMessageType_S2CBetMultiple):
		robot.RequestBetChips(buffer)
		break
	}

}

// OnGameStatus 机器人收到游戏状态改变
func (robot *Robot) OnGameStatus(buffer []byte) {
	// 状态消息入参
	resp := &msg.StatusMessageRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析游戏状态消息错误: %v", err)
		return
	}

	// 摊牌状态
	if resp.Status == int32(msg.GameStatus_ShowCards) {
		robot.RequestShowCards()
	}
}

// RequestBetChips 机器人请求投注
func (robot *Robot) RequestBetChips(buffer []byte) {
	// 机器人是庄家，不能投注
	if robot.IsBanker {
		return
	}

	resp := &msg.BetMultipleRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析投注倍率消息错误: %v", err)
		return
	}

	var (
		betRate, betIndex int
		betMultiple       int64
	)

	// 匹配投注权重
	betWeight := rand.RandInt(1, 101)
	log.Tracef("机器人投注权重：%d", betWeight)

	// 根据手上牌是否是大牌 获取对应到 投注概率分布
	betRateDis := robot.Cfg.SmallCardsBetRateDis
	if robot.GameLogic.UserList[robot.User.GetID()].GetBiggest {
		betRateDis = robot.Cfg.BigCardsBetRateDis
	}

	for index, rate := range betRateDis {
		if betWeight > betRate && betWeight <= betRate+rate {
			betIndex = index
			break
		}
		betRate += rate
	}

	betMultiple = int64(betIndex + 1)

	if betMultiple > resp.HighestMultiple {
		betMultiple = 1
	}

	log.Tracef("机器人投注倍数：%d", betMultiple)

	req := msg.BetChipsReq{
		BetMultiple: betMultiple,
	}

	waitTime := robot.GetWaitTime()

	robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {

		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBetChips), &req)
		if err != nil {
			log.Errorf("机器人投注错误: %v", err.Error())
		}
	})
}

// RequestShowCards 机器人请求摊牌
func (robot *Robot) RequestShowCards() {

	req := msg.ShowCardsReq{
		Ok: true,
	}

	waitTime := robot.GetWaitTime()

	robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {

		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SShowCards), &req)
		if err != nil {
			log.Errorf("机器人摊牌错误: %v", err.Error())
		}
	})
}

// GetWaitTime 获取等待时间
func (robot *Robot) GetWaitTime() int {
	// 随机时间
	randomTime := rand.RandInt(robot.Cfg.ActionTime.Shortest, robot.Cfg.ActionTime.Longest)
	return randomTime
}
