package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// 控牌
func (game *Blackjack) Control(cards []byte, user *data.User) {
	// 点控
	cheatRate := user.ExactControlRate

	// 血池作弊率
	roomProb := game.table.GetRoomProb()

	log.Warnf("当前血池值：%d", roomProb)

	if cheatRate == 0 {
		cheatRate = int64(roomProb)
	}

	rateIndex := -1

	for k, v := range game.ControlCfg.ExactControlRate {
		if cheatRate == v {
			rateIndex = k
			break
		}
	}

	// 没有点控参数
	if rateIndex < 0 {
		return
	}

	// 最后一张牌
	lastCard := game.Poker.Cards[len(game.Poker.Cards)-1]

	// 最后一张牌牌值
	lastCardValue, _ := poker.GetCardValueAndColor(lastCard)

	// 获取手牌最小点数
	smallestPoint := poker.GetSmallestPoint(poker.GetPoint(cards))

	// 点控权重值
	weightsValue := rand.RandInt(1, 10001)
	log.Debugf("用户 %d 点控权重：%d", user.ID, weightsValue)

	//testControlRes := msg.TestControlRes{
	//	RoomProb:      int64(roomProb),
	//	ExactControl:  user.ExactControlRate,
	//	ActionWeights: int64(weightsValue),
	//}
	//
	//// 发送控牌信息
	//game.SendControlInfo(testControlRes, user.User)

	cardsLen := len(cards)

	// 第一次发牌
	if cardsLen == 0 {
		// 初始黑杰克概率
		blackjackRate := game.ControlCfg.BlackjackPlace[rateIndex]

		log.Debugf("第一次黑杰克权重为 %d, 结果为 %v", weightsValue, weightsValue <= blackjackRate)

		// 给用户安排黑杰克牌
		if weightsValue <= blackjackRate {
			// 插入一张A牌
			game.Poker.PlugSelectedCard(poker.Acard)

			// 插入值为10的牌
			game.Poker.PlugRangeCard(0xa, 0xd)

			return
		}

		////// 随机给牌，不能是21点

		// 牌面是A
		if lastCardValue == poker.Acard {
			// 插入牌值不为10的牌
			game.Poker.PlugUnRangeCard(0xa, 0xd)
			return
		}

		// 牌面是值为10的牌
		if lastCardValue >= 0xa && lastCardValue <= 0xd {
			// 插入值为2 ~ 10 的牌
			game.Poker.PlugRangeCard(0x2, 0xd)
			return
		}

	}

	// 第二次发牌
	if cardsLen == 2 && smallestPoint > 12 {
		// 第一次要牌爆牌概率
		firstBustRate := game.ControlCfg.FirstBustPlace[rateIndex]

		log.Debugf("第一次要牌爆牌权重为 %d, 结果为 %v", weightsValue, weightsValue <= firstBustRate)

		// 安排用户爆牌
		if weightsValue <= firstBustRate {
			minLimit := byte(22 - smallestPoint)

			// 插入值为必定爆牌的牌
			game.Poker.PlugRangeCard(minLimit, 0xd)
			return
		}

		////// 第一次要牌不爆，安排一个不爆的牌

		minLimit := byte(22 - smallestPoint)

		// 插入不会爆炸的牌
		game.Poker.PlugUnRangeCard(minLimit, 0xd)

		return

	}

	// 第三次发牌
	if cardsLen == 3 {

		// 最小点数不大于12点，拿牌后必定大于或等于12点
		if smallestPoint <= 12 {

			minLimit := byte(12 - smallestPoint)

			// 插入必定能与手牌组成大于13的牌
			game.Poker.PlugRangeCard(minLimit, 0xd)
			return
		}

		////// 牌值大于12点，作弊率判定

		// 第二次要牌爆牌概率
		secondBustRate := game.ControlCfg.SecondBustPlace[rateIndex]

		log.Debugf("第二次要牌爆牌权重为 %d, 结果为 %v", weightsValue, weightsValue <= secondBustRate)

		// 安排第二次要牌爆牌
		if weightsValue <= secondBustRate {
			minLimit := byte(22 - smallestPoint)

			// 插入值为必定爆牌的牌
			game.Poker.PlugRangeCard(minLimit, 0xd)
			return
		}

		////// 作弊率判定未中，随机拿牌值不为 10 的牌

		// 插入牌值不为 10 的牌
		game.Poker.PlugUnRangeCard(0xa, 0xd)
		return

	}

	// 第四次发牌
	if cardsLen == 4 {
		// 第三次要牌爆牌概率
		thirdBustPlace := game.ControlCfg.ThirdBustPlace[rateIndex]

		log.Debugf("第三次要牌爆牌权重为 %d, 结果为 %v", weightsValue, weightsValue <= thirdBustPlace)

		// 安排第三次要牌爆牌
		if weightsValue <= thirdBustPlace {
			minLimit := byte(22 - smallestPoint)

			// 插入值为必定爆牌的牌
			game.Poker.PlugRangeCard(minLimit, 0xd)
			return
		}

		///// 不强制爆牌

		// 高作弊率不能拿牌值为10的牌
		if rateIndex > 3 {

			// 插入牌值不为 10 的牌
			game.Poker.PlugUnRangeCard(0xa, 0xd)
			return
		}

	}

}

// BankerControl 庄家控牌
func (game *Blackjack) BankerControl() {

	// 检测庄家控牌开关
	if !game.ControlCfg.BankerControl {
		return
	}

	// 检测当前血池值是否为最高档
	if roomProb := game.table.GetRoomProb(); roomProb != 3000 {
		return
	}

	// 玩家最高点数
	var highestPoint int32

	// 获取玩家最高点数，跳过 21点检测
	for _, user := range game.UserList {

		// 跳过机器人
		if user.User.IsRobot() {
			continue
		}

		// 跳过已经结算的玩家
		if user.Status == int32(msg.UserStatus_UserStopAction) {
			continue
		}

		for _, v := range user.HoldCards {

			// 跳过不存在的手牌（没有经过分牌产生的手牌）
			if len(v.Cards) == 0 {
				continue
			}
			point := poker.GetNearPoint21(v.Point)

			// 跳过21点
			if point >= 21 {
				continue
			}

			if point > highestPoint {
				highestPoint = point
			}

		}

	}

	bankerPoint := poker.GetSmallestPoint(game.HostCards.Point)

	maxPoint := 20 - bankerPoint

	minPoint := highestPoint - bankerPoint

	// 庄家点数在 7点到10点之间, 防止要到大过17点，但是比最高玩家点数小的牌
	if bankerPoint >= 7 && bankerPoint <= 10 {

		maxPoint = 16 - bankerPoint
		minPoint = 2
	}

	if minPoint < 0 {
		minPoint = 2
	}

	if maxPoint < minPoint {
		log.Errorf("庄家控牌，取得错误大小值范围，最小点: %d, 最大点: %d, 庄家点数: %d, 玩家最高点数: %d。", minPoint, maxPoint, bankerPoint, highestPoint)
		return
	}

	// 最小，最大值限制
	var limitMin, limitMax byte

	// 最小点 >= 10, 最小值为 10
	if minPoint >= 10 {
		limitMin = 0xa
	} else {
		limitMin = byte(minPoint)
	}

	// 最大点 >= 10, 最大值为 K
	if maxPoint >= 10 {
		limitMax = 0xd
	} else {
		limitMax = byte(maxPoint)
	}

	game.Poker.PlugRangeCard(limitMin, limitMax)

}
