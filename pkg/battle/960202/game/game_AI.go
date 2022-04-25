package game

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	conf "github.com/kubegames/kubegames-games/pkg/battle/960202/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// Robot 机器人结构体
type Robot struct {
	User      player.RobotInterface
	TimerJob  *player.Job
	Cfg       conf.RobotConfig
	GameLogic *BankerNiuniu
	IsBanker  bool
	ChairID   int32
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, gameInter table.TableHandler, robotCfg conf.RobotConfig, chairID int32) {
	robot.User = userInter
	robot.Cfg = robotCfg
	robot.GameLogic = gameInter.(*BankerNiuniu)
	robot.ChairID = chairID
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {

	switch subCmd {
	// 游戏状态变更消息
	case int32(msg.SendToClientMessageType_S2CGameStatus):
		robot.OnGameStatus(buffer)
		break

	// 收到抢庄结果
	case int32(msg.SendToClientMessageType_S2CRobBankerResult):
		robot.ReceiveRobResult(buffer)
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

	switch resp.Status {
	// 抢庄状态
	case int32(msg.GameStatus_RobBanker):
		robot.RequestRobBanker()
		break

		// 摊牌状态
	case int32(msg.GameStatus_ShowChards):
		robot.RequestShowCards()
		break
	}
}

// ReceiveRobResult 机器人收到抢庄结果
func (robot *Robot) ReceiveRobResult(buffer []byte) {
	// 状态消息入参
	resp := &msg.RobBankerResultRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析抢庄结果消息错误: %v", err)
		return
	}

	if resp.BankerChairId == robot.ChairID {
		robot.IsBanker = true
	}
}

// RequestRobBanker 机器人请求抢庄
func (robot *Robot) RequestRobBanker() {

	var robRate, RobIndex int
	// 匹配抢庄权重
	robWeight := rand.RandInt(1, 101)
	log.Tracef("机器人抢庄权重：%d", robWeight)

	// 根据手上牌是否是大牌 获取对应到 抢庄概率分布
	robRateDis := robot.Cfg.SmallCardsRobRateDis
	if robot.GameLogic.UserList[robot.User.GetID()].GetBiggest || robot.GameLogic.UserList[robot.User.GetID()].GetSecond {
		robRateDis = robot.Cfg.BigCardsRobRateDis
	}

	// 吐分状态采用 吐分抢庄概率分布
	if roomProb := robot.GameLogic.Table.GetRoomProb(); roomProb < 0 {
		robRateDis = robot.Cfg.HighWinRobRateDis
	}

	for index, rate := range robRateDis {
		if robWeight > robRate && robWeight <= robRate+rate {
			RobIndex = index - 1
			break
		}
		robRate += rate
	}
	log.Tracef("机器人抢庄下标：%d", RobIndex)

	req := msg.RobBankerReq{
		RobIndex: int32(RobIndex),
	}

	waitTime := robot.GetWaitTime()

	robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {
		// 请求server买保险
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRobBanker), &req)
		if err != nil {
			log.Errorf("机器人抢庄错误: %v", err.Error())
		}
	})
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

	var betRate, betIndex int

	// 匹配抢庄权重
	betWeight := rand.RandInt(1, 101)
	log.Tracef("机器人投注权重：%d", betWeight)

	// 根据手上牌是否是大牌 获取对应到 投注概率分布
	betRateDis := robot.Cfg.SmallCardsBetRateDis
	if robot.GameLogic.UserList[robot.User.GetID()].GetBiggest || robot.GameLogic.UserList[robot.User.GetID()].GetSecond {
		betRateDis = robot.Cfg.BigCardsBetRateDis
	}

	// 吐分状态采用 吐分投注概率分布
	if roomProb := robot.GameLogic.Table.GetRoomProb(); roomProb < 0 {
		betRateDis = robot.Cfg.HighWinBetRateDis
	}

	for index, rate := range betRateDis {
		if betWeight > betRate && betWeight <= betRate+rate {
			betIndex = index
			break
		}
		betRate += rate
	}

	log.Tracef("机器人投注倍数：%d", resp.Multiples[betIndex])

	req := msg.BetChipsReq{
		BetMultiple: resp.Multiples[betIndex],
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
