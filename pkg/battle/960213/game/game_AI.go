package game

/*
#cgo CFLAGS: -I../
#cgo LDFLAGS: -L../ -lddzall -Wl,-rpath=./
#include "ddzall.h"
*/
import "C"

import (
	"common/rand"
	"game_frame_v2/game/clock"
	"game_poker/ddzall/config"
	"game_poker/ddzall/data"
	"game_poker/ddzall/msg"
	"game_poker/ddzall/poker"
	"strconv"
	"strings"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// Robot 机器人结构体
type Robot struct {
	UserInter    player.RobotInterface
	TimerJob     *clock.Job
	User         *data.User            // 用户
	cards        []byte                // 手牌
	Cfg          config.RobotConfig    // 机器人配置
	GameLogic    *DouDizhu             // 游戏逻辑，只能查看数据，不能修改数据
	BestSolution []poker.SolutionCards // 最优牌解
	RobScore     int                   // 抢庄分值
	playNum      int                   // 难度等级，取值范围暂定0-5000，数值越高，机器人越智能
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, game table.TableHandler, robotCfg config.RobotConfig) {
	robot.UserInter = userInter
	robot.GameLogic = game.(*DouDizhu)
	robot.User = game.(*DouDizhu).UserList[userInter.GetId()]
	robot.Cfg = robotCfg
	robot.playNum = 5000
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {

	switch subCmd {

	// 状态消息
	case int32(msg.SendToClientMessageType_S2CGameStatus):
		robot.DealGameStatus(buffer)
		break

	// 当前抢庄玩家
	case int32(msg.SendToClientMessageType_S2CCurrentRobber):
		robot.DealCurrentRobber(buffer)
		break

	// 当前操作玩家
	case int32(msg.SendToClientMessageType_S2CCurrentPlayer):
		robot.DealCurrentPlayer(buffer)
		break
	}
}

// DealGameStatus 处理游戏状态消息
func (robot *Robot) DealGameStatus(buffer []byte) {
	resp := &msg.StatusMessageRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析发牌信息失败: %v", err)
		return
	}

	switch resp.Status {
	case int32(msg.GameStatus_RedoubleStatus):
		var (
			addLevel    int   // 加倍等级
			acc         int   // 累加器
			addIndex    int   // 加倍索引
			addMultiple int64 // 加倍倍数
		)

		// 获取加倍等级
		for index, score := range robot.Cfg.AddScorePlace {
			if index == 0 && robot.RobScore <= score {
				addLevel = 0
				break
			}

			if index == len(robot.Cfg.RobScorePlace)-1 && robot.RobScore >= score {
				addLevel = len(robot.Cfg.RobScorePlace) - 1
				break
			}

			if robot.RobScore == score {
				addLevel = index
			}
		}

		// 权重值
		weightsValue := rand.RandInt(1, 101)
		log.Tracef("机器人加倍权重：%d", weightsValue)

		// 计算加倍权重
		for k, v := range robot.Cfg.AddRatePlace[addLevel] {
			downLimit := acc
			upLimit := acc + v

			acc = upLimit
			if weightsValue > downLimit && weightsValue <= upLimit {
				addIndex = k
				break
			}
		}

		addMultiple = robot.GameLogic.GameCfg.AddMultiple[addIndex]

		req := msg.RedoubleReq{
			AddNum: addMultiple,
		}

		// 随机时间
		delayTime := rand.RandInt(1000, 5001)

		// 延迟发送消息
		robot.TimerJob, _ = robot.UserInter.AddTimer(time.Duration(delayTime), func() {
			// 请求server加倍
			err := robot.UserInter.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRedouble), &req)
			if err != nil {
				log.Errorf("send server msg fail: %v", err.Error())
			}
		})

	}

}

// DealGameStatus 处理当前抢庄玩家消息
func (robot *Robot) DealCurrentRobber(buffer []byte) {
	resp := &msg.CurrentRobberRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析发牌信息失败: %v", err)
		return
	}

	robotID := robot.User.ID

	// 当前抢庄玩家不是自己
	if resp.UserId != robotID {
		return
	}

	var (
		robLevel int   // 抢地主分数等级
		acc      int   // 累加器
		robIndex int   // 抢地主索引
		robScore int64 // 抢地主分数
	)

	// 设置手牌的抢庄分总值
	robot.SetRobScore(robot.GameLogic.UserList[robotID].Cards)

	log.Warnf("机器人 %d 的抢庄分数 %d", robotID, robot.RobScore)

	// 获取抢地主分数等级
	for index, score := range robot.Cfg.RobScorePlace {
		if index == 0 && robot.RobScore <= score {
			robLevel = 0
			break
		}

		if index == len(robot.Cfg.RobScorePlace)-1 && robot.RobScore >= score {
			robLevel = len(robot.Cfg.RobScorePlace) - 1
			break
		}

		if robot.RobScore == score {
			robLevel = index
		}
	}

	// 权重值
	weightsValue := rand.RandInt(1, 101)
	log.Tracef("机器人抢分权重：%d", weightsValue)

	// 计算抢分权重
	for k, v := range robot.Cfg.RobRatePlace[robLevel] {
		downLimit := acc
		upLimit := acc + v

		acc = upLimit
		if weightsValue > downLimit && weightsValue <= upLimit {
			robIndex = k
			break
		}
	}

	robScore = robot.GameLogic.GameCfg.RobScore[robIndex]

	// 预抢分数已被抢了，就不抢了
	if robScore <= resp.CurrentNum {
		robScore = 0
	}

	req := msg.RobReq{
		RobNum: robScore,
	}

	// 随机时间
	delayTime := rand.RandInt(1000, 5001)

	// 延迟发送消息
	robot.TimerJob, _ = robot.UserInter.AddTimer(time.Duration(delayTime), func() {
		// 请求server抢分
		err := robot.UserInter.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRob), &req)
		if err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
	})

}

//DealGameStatus 处理当前操作玩家消息
func (robot *Robot) DealCurrentPlayer(buffer []byte) {
	resp := &msg.CurrentPlayerRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析发牌信息失败: %v", err)
		return
	}

	// 当前出牌玩家ID不是我
	if resp.UserId != robot.User.ID {
		return
	}

	// 要打出打牌
	var outCards []byte

	// 能够出牌
	if resp.Permission {

		var (
			totalNum    []int
			numSting    []string
			resultNum   [15]int
			nextLastPut []byte
			preLastPut  []byte
		)

		// 地主座位ID
		dizhuChairID := robot.GameLogic.Dizhu.ChairID

		// 地主手牌数字格式
		dizhuNum := GetNumTypeByCards(robot.GameLogic.Chairs[dizhuChairID].Cards)

		// 地主下家座位ID
		dizhuNextChairID := GetNextChairID(dizhuChairID)

		// 地主下家手牌数字格式
		dizhuNextNum := GetNumTypeByCards(robot.GameLogic.Chairs[dizhuNextChairID].Cards)

		// 地主上家座位ID
		dizhuPreChairID := GetPreChairID(dizhuChairID)

		// 地主上家手牌数字格式
		dizhuPreNum := GetNumTypeByCards(robot.GameLogic.Chairs[dizhuPreChairID].Cards)

		// 当前机器人上家座位ID
		preChairID := GetPreChairID(robot.User.ChairID)

		// 当前机器人上家出牌记录
		prePutRecords := robot.GameLogic.Chairs[preChairID].PutCardsRecords
		if len(prePutRecords) != 0 {
			preLastPut = prePutRecords[len(prePutRecords)-1]
		}

		// 当前玩家上家最后一次出牌数字格式
		preLastNum := GetNumTypeByCards(preLastPut)

		// 当前机器人下家座位ID
		nextChairID := GetNextChairID(robot.User.ChairID)

		// 当前机器人下家出牌记录
		nextPutRecords := robot.GameLogic.Chairs[nextChairID].PutCardsRecords
		if len(nextPutRecords) != 0 {
			nextLastPut = nextPutRecords[len(nextPutRecords)-1]
		}

		// 当前玩家上家最后一次出牌数字格式
		nextLastNum := GetNumTypeByCards(nextLastPut)

		// 总和所有数据的数据格式
		totalNum = append(totalNum, dizhuNum[:]...)
		totalNum = append(totalNum, dizhuNextNum[:]...)
		totalNum = append(totalNum, dizhuPreNum[:]...)
		totalNum = append(totalNum, preLastNum[:]...)
		totalNum = append(totalNum, nextLastNum[:]...)

		for _, num := range totalNum {
			numSting = append(numSting, strconv.Itoa(num))
		}

		state := strings.Join(numSting, ",")

		log.Warnf("传入值，state：%s, palyer：%d, playNum：%d", robot.User.Role, robot.playNum)

		result := C.get_action_plus(C.CString(state), C.int(robot.User.Role), C.int(robot.playNum))

		respStr := C.GoString(result)

		log.Warnf("传入值: %s", respStr)

		// 截取前15个字符串
		respStr = respStr[:15]

		// 字符串转切片
		respStrList := strings.Split(respStr, "")
		for i, str := range respStrList {
			num, _ := strconv.Atoi(str)
			resultNum[i] = num
		}

		outCards = GetCardsByNumType(resultNum, robot.User.Cards)

	}

	robot.PutCards(outCards)
}

// PutCards 机器人出牌
func (robot *Robot) PutCards(cards []byte) {
	req := msg.PutCardsReq{
		Cards: cards,
	}

	err := robot.UserInter.SendMsgToServer(int32(msg.ReceiveMessageType_C2SPutCards), &req)
	if err != nil {
		log.Errorf("send server msg fail: %v", err.Error())
	}
}

// 设置抢庄分值
func (robot *Robot) SetRobScore(cards []byte) {
	var robScore int

	// 王牌数量
	kingCount := poker.GetKingCount(cards)

	switch kingCount {

	// 只有一张王
	case 1:
		robScore += 2

	// 有两张王
	case 2:
		robScore += 7

	}

	// 2牌数量
	value2Count := poker.GetValue2Count(cards)

	robScore += value2Count

	// 炸弹列表
	BombList, _ := poker.FindTypeInCards(msg.CardsType_Bomb, cards)

	robScore += len(BombList) * 2

	// 飞机列表
	PlaneList, _ := poker.FindTypeInCards(msg.CardsType_SerialTriplet, cards)

	robScore += len(PlaneList)

	robot.RobScore = robScore
}

// GetNumTypeByCards 获取手牌的数字格式
func GetNumTypeByCards(cards []byte) (numType [15]int) {
	for _, card := range cards {
		value, _ := poker.GetCardValueAndColor(card)
		numType[int(value-1)]++
	}

	return
}

// 从手牌中选出数字格式的牌组
func GetCardsByNumType(numType [15]int, handCards []byte) (outCards []byte) {
	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range handCards {
		cards = append(cards, card)
	}

	for i, count := range numType {
		if count == 0 {
			continue
		}

		for j := 0; j < count; j++ {
			for index, card := range cards {
				value, _ := poker.GetCardValueAndColor(card)
				if byte(i+1) == value {
					outCards = append(outCards, card)
					cards = append(cards[:index], cards[index+1:]...)
					break
				}
			}
		}
	}
	return
}
