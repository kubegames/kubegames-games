package game

import (
	"go-game-sdk/lib/clock"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960208/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
)

// Robot 机器人结构体
type Robot struct {
	User      player.RobotInterface
	TimerJob  *clock.Job
	Cfg       config.RobotConfig
	GameLogic *ThreeDoll
	IsBanker  bool
	ChairID   int32
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, gameInter table.TableHandler, robotCfg config.RobotConfig, chairID int32) {
	robot.User = userInter
	robot.Cfg = robotCfg
	robot.GameLogic = gameInter.(*ThreeDoll)
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
	case int32(msg.GameStatus_ShowCards):
		robot.RequestShowCards()
		break
	}
}

// ReceiveRobResult 机器人收到抢庄结果
func (robot *Robot) ReceiveRobResult(buffer []byte) {
	// 状态消息入参
	resp := &msg.RobResultRes{}
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

	var isRob bool
	// 匹配抢庄权重
	robWeight := rand.RandInt(0, 101)
	log.Tracef("机器人 %d 抢庄权重 %d", robot.User.GetID(), robWeight)

	if robWeight <= 50 {
		isRob = true
	}

	req := msg.RobBankerReq{
		UserId: robot.User.GetID(),
		IsRob:  isRob,
	}

	waitTime := robot.GetWaitTime()

	switch waitTime {
	case 0:
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRobBanker), &req)
		if err != nil {
			log.Errorf("机器人抢庄错误: %v", err.Error())
		}
		break
	case 5000:
		break
	default:
		robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {
			// 请求server买保险
			err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRobBanker), &req)
			if err != nil {
				log.Errorf("机器人抢庄错误: %v", err.Error())
			}
		})
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
		betIndex    int
		betMultiple int64
	)

	// 投注权重
	betWeight := rand.RandInt(0, 101)
	log.Tracef("机器人 %d 投注权重 %d", robot.User.GetID(), betWeight)
	rateLimit := 0
	for index, rate := range robot.Cfg.BetRate {
		if betWeight >= rateLimit && betWeight < rate+rateLimit {
			betIndex = index
			break
		}
		rateLimit += rate
	}

	if betIndex >= len(resp.Multiples) {
		betMultiple = resp.Multiples[len(resp.Multiples)-1]
	} else {
		betMultiple = resp.Multiples[betIndex]
	}

	req := msg.BetChipsReq{
		BetMultiple: betMultiple,
	}

	waitTime := robot.GetWaitTime()

	switch waitTime {
	case 0:
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBetChips), &req)
		if err != nil {
			log.Errorf("机器人投注错误: %v", err.Error())
		}
		break
	case 5000:
		break
	default:
		robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {

			err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBetChips), &req)
			if err != nil {
				log.Errorf("机器人投注错误: %v", err.Error())
			}
		})
	}

}

// RequestShowCards 机器人请求摊牌
func (robot *Robot) RequestShowCards() {

	req := msg.ShowCardsReq{
		UserId: robot.User.GetID(),
	}

	waitTime := robot.GetWaitTime()

	switch waitTime {
	case 0:
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SShowCards), &req)
		if err != nil {
			log.Errorf("机器人摊牌错误: %v", err.Error())
		}
		break
	case 5000:
		break
	default:
		robot.TimerJob, _ = robot.User.AddTimer(int64(waitTime), func() {

			err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SShowCards), &req)
			if err != nil {
				log.Errorf("机器人摊牌错误: %v", err.Error())
			}
		})
	}

}

// GetWaitTime 获取等待时间
func (robot *Robot) GetWaitTime() (randTime int) {
	// 操作时间权重
	matchWeight := rand.RandInt(0, 101)

	rateIndex := -1

	rateLimit := 0

	for index, rate := range robot.Cfg.RobotActionRate {
		if matchWeight >= rateLimit && matchWeight < rate+rateLimit {
			rateIndex = index
			break
		}
		rateLimit += rate
	}

	switch rateIndex {
	// 立即执行
	case 0:
		randTime = 0
		break
		// 1000~2000豪秒
	case 1:
		// 操作时间权重
		randTime = rand.RandInt(1000, 2001)
		break
		// 3000~4000豪秒
	case 2:
		randTime = rand.RandInt(3000, 4001)
		break
		// 超时
	case 3:
		randTime = 5000
		break
		// 默认超时
	default:
		randTime = 5000
		break
	}

	return
}
