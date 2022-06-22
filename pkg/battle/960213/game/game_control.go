package game

//
//import (
//	"github.com/kubegames/kubegames-sdk/pkg/log"
//	"github.com/kubegames/kubegames-games/internal/pkg/rand"
//	"github.com/kubegames/kubegames-games/pkg/battle/960213/poker"
//)
//
//type SolutionSequence struct {
//	TotalWeight int
//	Count       int
//	Cards       []byte
//}
//
//// ControlPoker 控牌
//func (game *DouDizhu) ControlPoker() {
//	var (
//		//牌解序列
//		solutionSequence []SolutionSequence
//
//		// 最小手数
//		smallestCount int
//	)
//	for i := 0; i < len(game.UserList); i++ {
//
//		// 最优解
//		cards := game.Poker.DrawCard()
//		solutionCards := poker.GetOptimalSolutionCards(cards)
//
//		// 一副牌手数
//		solutionCount := len(solutionCards)
//
//		// 最小手数
//		if smallestCount == 0 {
//			smallestCount = solutionCount
//		}
//		if solutionCount < smallestCount {
//			smallestCount = solutionCount
//		}
//
//		// 一副牌总权重
//		var totalWeight int
//		for _, once := range solutionCards {
//			totalWeight += once.Weight
//		}
//
//		solutionSequence = append(solutionSequence, SolutionSequence{
//			TotalWeight: totalWeight,
//			Count:       solutionCount,
//			Cards:       cards,
//		})
//	}
//
//	// 以3副牌中手数最少的哪一家为基准，多一手的玩家手牌权值减一分
//	for i, solutionCards := range solutionSequence {
//		solutionSequence[i].TotalWeight = solutionCards.TotalWeight - solutionCards.Count + smallestCount
//	}
//
//	// 对牌解序列按照从小到大进行排序
//	for i := 0; i < len(solutionSequence)-1; i++ {
//		for j := 0; j < (len(solutionSequence) - 1 - i); j++ {
//
//			if solutionSequence[j].TotalWeight > solutionSequence[j+1].TotalWeight {
//				solutionSequence[j], solutionSequence[j+1] = solutionSequence[j+1], solutionSequence[j]
//			}
//		}
//	}
//
//	game.disposeCard(&solutionSequence)
//
//}
//
//// disposeCard 配牌
//func (game *DouDizhu) disposeCard(solutionSequence *[]SolutionSequence) {
//	// 牌序列为空，已经完成了配牌
//	if len(*solutionSequence) == 0 {
//		return
//	}
//
//	roomProb, err := game.Table.GetRoomProb()
//	if err != nil {
//		log.Errorf("游戏 %d 获取血池作弊率错误：%v", game.Table.GetID(), err.Error())
//		return
//	}
//	log.Tracef("游戏 %d 获取血池值 %d", game.Table.GetID(), roomProb)
//
//	// 血池值为0， 默认为1000
//	if roomProb == 0 {
//		roomProb = 1000
//	}
//
//	// 概率分布
//	ratePlace := make(map[int64]int)
//
//	for _, id := range game.Seats {
//		if len(game.ControlledCards[id]) != 0 {
//			continue
//		}
//
//		matchProb := roomProb
//
//		// 点控
//		pointProb := game.UserList[id].User.GetProb()
//
//		if pointProb != 0 {
//			matchProb = pointProb
//		}
//
//		// 机器人 血池值取反
//		if game.UserList[id].User.IsRobot() {
//			matchProb = -roomProb
//		}
//
//		// 检测血池值
//		probIndex := game.checkProb(matchProb)
//		if probIndex == -1 {
//			log.Errorf("游戏 %d 错误的作弊率值: %d", game.Table.GetID(), matchProb)
//			// 默认 1000 作弊率的 索引
//			probIndex = 2
//			if game.UserList[id].User.IsRobot() {
//				probIndex = 4
//			}
//		}
//
//		ratePlace[id] = game.GameCfg.Control.BiggestCardsRate[probIndex]
//	}
//
//	// 可控牌人数为0
//	if len(ratePlace) == 0 {
//		return
//	}
//
//	// 总概率值
//	var totalRate int
//
//	// 总概率值
//	for _, rate := range ratePlace {
//		totalRate += rate
//	}
//
//	// 未满10000的剩余概率值剩余平均概率
//	lessAverageRate := (10000 - totalRate) / len(ratePlace)
//
//	if lessAverageRate < 0 {
//		lessAverageRate = 0
//	}
//
//	// 更新新概率值，让概率变得更加平缓
//	for id, rate := range ratePlace {
//		ratePlace[id] = lessAverageRate + rate
//		totalRate += lessAverageRate
//	}
//
//	// 权重
//	weight := rand.RandInt(1, totalRate+1)
//	addRate := 0
//
//	// 把最大牌给权值拥有者
//	for id, rate := range ratePlace {
//		if weight > addRate && weight <= addRate+rate {
//			index := len(*solutionSequence) - 1
//			game.ControlledCards[id] = (*solutionSequence)[index].Cards
//			*solutionSequence = append((*solutionSequence)[:index:index], (*solutionSequence)[index+1:]...)
//			break
//		}
//		addRate += rate
//	}
//
//	game.disposeCard(solutionSequence)
//}
//
//// 检测作弊率
//func (game *DouDizhu) checkProb(prob int32) (probIndex int) {
//	probIndex = -1
//	for index, rate := range game.GameCfg.Control.ControlRate {
//		if prob == rate {
//			probIndex = index
//		}
//	}
//
//	return
//}
