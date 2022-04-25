package game

import (
	rand2 "math/rand"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960207/poker"
)

// Control 控制牌
func (game *GeneralNiuniu) Control() {
	// 设置牌组序列
	game.SetCardsSequence()

	// 新控牌
	game.ControlBiggerCards()

	for _, holdCards := range game.CardsSequence {
		for id, _ := range game.UserList {
			if _, ok := game.ControlledCards[id]; !ok {
				game.ControlledCards[id] = holdCards
				break
			}

		}
	}

}

// SetCardsSequence 设置牌组序列
func (game *GeneralNiuniu) SetCardsSequence() {
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
func (game *GeneralNiuniu) ControlBiggerCards() {
	log.Tracef("牌堆有 %d 副", len(game.CardsSequence))
	// 最大牌概率分布
	biggestCardsRatePlace := make(map[int64]int)
	for id, user := range game.UserList {

		// 没有点控，用血池
		prob := user.ExactControlRate
		if prob == 0 {
			prob = game.Table.GetRoomProb()
		}

		// 机器人 和 玩家 采用 不同的大牌概率分布
		biggestRateDis := game.GameCfg.Control.PlayerBiggestRate
		if user.User.IsRobot() {
			biggestRateDis = game.GameCfg.Control.RobotBiggestRate
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

	}

	// 可控牌人数为0
	if len(biggestCardsRatePlace) == 0 {
		return
	}

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

	log.Tracef("第一大牌用户 %d ", biggestCardsUserID)
	log.Tracef("牌堆有 %d 副", len(game.CardsSequence))

	// 分配最大牌
	game.ControlledCards[biggestCardsUserID] = game.CardsSequence[len(game.CardsSequence)-1]
	game.CardsSequence = game.CardsSequence[:len(game.CardsSequence)-1]
	log.Tracef("牌堆有 %d 副", len(game.CardsSequence))
}

// 检测作弊率
func (game *GeneralNiuniu) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range game.GameCfg.Control.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}

// ShuffleSlice 切片随机
func ShuffleSlice(group []int64) []int64 {
	rand2.Seed(time.Now().UnixNano())
	rand2.Shuffle(len(group), func(i, j int) {
		group[i], group[j] = group[j], group[i]
	})

	return group
}
