package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// Control 控制牌
func (game *ThreeDoll) Control() {
	// 设置牌组序列
	game.SetCardsSequence()

	// 处理配牌需求
	game.DealDemandReq()

	// 新控牌
	game.ControlBiggerCards()

	for _, holdCards := range game.CardsSequence {
		for id := range game.UserList {
			if _, ok := game.ControlledCards[id]; !ok {
				game.ControlledCards[id] = holdCards
				break
			}

		}
	}

}

// SetCardsSequence 设置牌组序列
func (game *ThreeDoll) SetCardsSequence() {
	for i := 0; i < len(game.UserList); i++ {
		cards := game.Poker.DrawCard()
		holdCards := poker.HoldCards{
			Cards:     cards,
			CardsType: poker.GetCardsType(cards),
		}

		// 牌组序列为空，加入一张牌组
		if len(game.CardsSequence) == 0 {
			game.CardsSequence = append(game.CardsSequence, holdCards)
			continue
		}

		// 插入排序法从小到大依次排列
		var newSequence []poker.HoldCards
		for k, v := range game.CardsSequence {

			// holdCards < v
			if poker.ContrastCards(&holdCards, &v) {

				rear := append([]poker.HoldCards{}, game.CardsSequence[k:]...)
				newSequence = append(append(game.CardsSequence[:k], holdCards), rear...)
				break
			} else {
				newSequence = append(game.CardsSequence, holdCards)
			}
		}
		game.CardsSequence = newSequence
	}
	log.Tracef("牌组序列：%v", game.CardsSequence)
}

// DealDemandReq 处理配牌需求
func (game *ThreeDoll) DealDemandReq() {
	for _, user := range game.UserList {

		switch user.DemandReq.DemandType {

		// 必赢/通杀，分配给最大牌
		case int32(msg.DemandType_MustWin), int32(msg.DemandType_WinAll):
			game.AssignBiggestCards(user.ID)
			break

			// 必输/通赔，分配给最小牌
		case int32(msg.DemandType_MustLose), int32(msg.DemandType_LoseAll):
			game.AssignLeastCards(user.ID)
			break

			// 爆玖/炸弹/三公
		case int32(msg.DemandType_ExplosionNineCards), int32(msg.DemandType_BoomCards), int32(msg.DemandType_ThreeDollCards), int32(msg.DemandType_PutIn):
			game.AssignInputCards(user.ID, user.DemandReq.DemandCards)
			break
		}
	}
}

// AssignBiggestCards 分配最大牌
func (game *ThreeDoll) AssignBiggestCards(userID int64) {
	if len(game.CardsSequence) != 0 && len(game.ControlledCards[userID].Cards) == 0 {
		game.ControlledCards[userID] = game.CardsSequence[len(game.CardsSequence)-1]
		game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]
	}
}

// AssignLeastCards 分配最小牌
func (game *ThreeDoll) AssignLeastCards(userID int64) {
	if len(game.CardsSequence) != 0 && len(game.ControlledCards[userID].Cards) == 0 {
		game.ControlledCards[userID] = game.CardsSequence[0]
		game.CardsSequence = game.CardsSequence[1:]
	}
}

// AssignInputCards 分配输入牌
func (game *ThreeDoll) AssignInputCards(userID int64, cards []byte) {
	if len(game.ControlledCards[userID].Cards) == 0 {
		cardsType := poker.GetCardsType(cards)
		game.ControlledCards[userID] = poker.HoldCards{
			Cards:     cards,
			CardsType: cardsType,
		}
	}
}

// ControlBiggerCards 控制大牌
func (game *ThreeDoll) ControlBiggerCards() {
	log.Tracef("牌堆有 %d 副", len(game.CardsSequence))
	// 最大牌，和第二大牌概率分布
	biggestCardsRatePlace := make(map[int64]int)
	secondCardsRatePlace := make(map[int64]int)

	for id, user := range game.UserList {
		if len(game.ControlledCards[id].Cards) != 0 {
			continue
		}

		// 没有点控，用血池
		prob := user.User.GetProb()
		if prob == 0 {
			prob = game.Table.GetRoomProb()
		}

		// 机器人 和 玩家 采用 不同的大牌概率分布
		biggestRateDis := game.GameCfg.Control.PlayerBiggestRate
		secondRateDis := game.GameCfg.Control.PlayerSecondRate
		if user.User.IsRobot() {
			biggestRateDis = game.GameCfg.Control.RobotBiggestRate
			secondRateDis = game.GameCfg.Control.RobotSecondRate
		}

		log.Tracef("用户 %d 作弊率 %d", user.ID, prob)

		// 检测作弊率
		probIndex := game.checkProb(prob)
		if probIndex == -1 {
			log.Warnf("游戏 %d 错误的作弊率: %d", game.Table.GetID(), prob)
			// 默认 1000 作弊率的 索引
			probIndex = 2
		}

		// 拿牌概率值为0, 不参与大牌分配
		if biggestRateDis[probIndex] != 0 {
			biggestCardsRatePlace[id] = biggestRateDis[probIndex]
		}

		if secondRateDis[probIndex] != 0 {
			secondCardsRatePlace[id] = secondRateDis[probIndex]
		}
	}

	// 概率分配最大牌
	if len(biggestCardsRatePlace) != 0 {
		// 总概率值
		var totalRate int

		// 总概率值
		for _, rate := range biggestCardsRatePlace {
			totalRate += rate
		}

		// 未满10000的剩余概率值剩余平均概率
		lessAverageRate := (10000 - totalRate) / len(biggestCardsRatePlace)

		if lessAverageRate < 0 {
			lessAverageRate = 0
		}

		// 更新新概率值，让概率变得更加平缓
		for id, rate := range biggestCardsRatePlace {
			biggestCardsRatePlace[id] = lessAverageRate + rate
			totalRate += lessAverageRate
		}

		// 权重
		weight := rand.RandInt(0, totalRate+1)

		// 概率累加值
		addRate := 0

		// 最大牌userID
		var biggestCardsUserID int64
		for id, rate := range biggestCardsRatePlace {

			if weight > addRate && weight <= addRate+rate {
				biggestCardsUserID = id
				break
			}
			addRate += rate
		}

		// 权重没有落在概率分布上
		if biggestCardsUserID == 0 {
			log.Errorf("游戏 %d 控牌最大牌 userID 为0，权重没有落在概率分布上", game.Table.GetID())
			return
		}

		log.Tracef("当前权重 %d，第一大牌用户 %d ", weight, biggestCardsUserID)

		// 分配最大牌
		game.ControlledCards[biggestCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
		game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]

		// 踢出已那最大牌用户的概率
		if _, ok := secondCardsRatePlace[biggestCardsUserID]; ok {
			delete(secondCardsRatePlace, biggestCardsUserID)
		}

	}

	// 概率分配第二大牌
	if len(secondCardsRatePlace) != 0 {
		// 总概率值
		var totalRate int

		// 计算拿第二大牌的总概率值
		for _, rate := range secondCardsRatePlace {
			totalRate += rate
		}

		// 剩余平均概率
		lessAverageRate := (10000 - totalRate) / len(secondCardsRatePlace)

		if lessAverageRate < 0 {
			lessAverageRate = 0
		}

		// 更新新概率值，让概率变得更加平缓
		for id, rate := range secondCardsRatePlace {
			secondCardsRatePlace[id] = lessAverageRate + rate
			totalRate += lessAverageRate
		}

		// 权重
		weight := rand.RandInt(1, totalRate+1)

		// 概率累加值
		addRate := 0

		// 最二大牌userID
		var secondCardsUserID int64
		for id, rate := range secondCardsRatePlace {

			if weight > addRate && weight <= addRate+rate {
				secondCardsUserID = id
				break
			}
			addRate += rate
		}

		// 权重没有落在概率分布上
		if secondCardsUserID == 0 {
			log.Errorf("游戏 %d 控牌最二大牌 userID 为0，权重没有落在概率分布上", game.Table.GetID())
			return
		}

		log.Tracef("当前权重 %d，第二大牌用户 %d ", weight, secondCardsUserID)
		// 分配最第二大牌
		game.ControlledCards[secondCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
		game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]
	}

}

// checkProb 检测作弊率
func (game *ThreeDoll) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range game.GameCfg.Control.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}

// DeleteDemandCards 从牌堆中删除配牌
func (game *ThreeDoll) DeleteDemandCards() {
	for _, user := range game.UserList {
		if len(user.DemandReq.DemandCards) != 3 {
			continue
		}

		for _, card := range user.DemandReq.DemandCards {
			for i, waitCard := range game.Poker.Cards {
				if card == waitCard {
					game.Poker.Cards = append(game.Poker.Cards[:i:i], game.Poker.Cards[i+1:]...)
					break
				}
			}
		}
	}
}
