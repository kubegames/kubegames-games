package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// Control 控制牌
func (game *BankerNiuniu) Control() {
	// 设置牌组序列
	game.SetCardsSequence()

	// 新控牌
	game.ControlBiggerCards()

	for _, holdCards := range game.CardsSequence {
		for id := range game.UserList {
			if _, ok := game.ControlledCards[id]; !ok {
				game.ControlledCards[id] = holdCards
				game.ControlList = append(game.ControlList, id)
				break
			}

		}
	}

}

// SetCardsSequence 设置牌组序列
func (game *BankerNiuniu) SetCardsSequence() {
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

// ControlBiggerCards 控制大牌
func (game *BankerNiuniu) ControlBiggerCards() {
	log.Tracef("牌堆有 %d 副", len(game.CardsSequence))
	// 最大牌，和第二大牌概率分布
	biggestCardsRatePlace := make(map[int64]int)
	secondCardsRatePlace := make(map[int64]int)

	for id, user := range game.UserList {

		// 没有点控，用血池
		prob := user.ExactControlRate
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
			log.Tracef("游戏 %d 错误的作弊率: %d", game.Table.GetID(), prob)
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

		game.UserList[biggestCardsUserID].GetBiggest = true
		game.ControlList = append(game.ControlList, biggestCardsUserID)

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

		game.UserList[secondCardsUserID].GetSecond = true
		game.ControlList = append(game.ControlList, secondCardsUserID)

		log.Tracef("当前权重 %d，第二大牌用户 %d ", weight, secondCardsUserID)
		// 分配最第二大牌
		game.ControlledCards[secondCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
		game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]
	}
}

// 检测作弊率
func (game *BankerNiuniu) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range game.GameCfg.Control.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}

// ExchangeCards 换牌控制
func (game *BankerNiuniu) ExchangeControl() {

	for _, user := range game.UserList {

		// 机器人 玩家不是庄 检查
		if !user.IsBanker || user.User.IsRobot() {
			continue
		}

		// 是否拿到最大牌
		getBiggest := game.ControlList[0] == user.ID

		// 最大 最小牌 检查
		if !getBiggest && game.ControlList[len(game.ControlList)-1] != user.ID {
			continue
		}

		// 玩家坐庄 拿最大牌 或者 拿最小牌 触发 换牌机制

		// 确定作弊率
		prob := game.Table.GetRoomProb()
		if user.User.GetProb() != 0 {
			prob = user.User.GetProb()
		}

		// 确定作弊等级
		probIndex := game.checkProb(prob)
		if probIndex == -1 {
			probIndex = 2
		}

		// 权重
		weight := rand.RandInt(1, 101)

		var (
			exchangeIndex int // 换牌索引
			exchangeLv    int // 换牌等级
			addRate       int // 概率累加值
		)

		// 最小牌换牌概率分布
		exchangeDis := game.GameCfg.SmallestExchangeDis[probIndex]

		// 拿到最大牌，用最大牌换牌概率分布
		if getBiggest {
			exchangeDis = game.GameCfg.BiggestExchangeDis[probIndex]
		}

		for id, rate := range exchangeDis {

			if weight > addRate && weight <= addRate+rate {
				exchangeIndex = id
				break
			}
			addRate += rate
		}

		switch exchangeIndex {
		case 1:
			// 最大牌 换第二大牌牌
			if getBiggest {

				exchangeLv = 1

			} else { // 最小牌 换第二大牌牌

				exchangeLv = 1
			}
		case 2:
			// 最大牌 换第三大牌牌
			if getBiggest {

				exchangeLv = 1

			} else { // 最小牌 换最大牌牌

				exchangeLv = 0
			}
		}

		if len(game.ControlList) > exchangeLv && exchangeLv > 0 && user.ID != game.ControlList[exchangeLv] {
			game.ControlledCards[user.ID], game.ControlledCards[game.ControlList[exchangeLv]] =
				game.ControlledCards[game.ControlList[exchangeLv]], game.ControlledCards[user.ID]
		}

	}
}

// ShuffleSlice 切片随机
func ShuffleSlice(group []int64) []int64 {
	rand.Shuffle(len(group), func(i, j int) {
		group[i], group[j] = group[j], group[i]
	})

	return group
}
