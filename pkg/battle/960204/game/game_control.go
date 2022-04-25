package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

type SolutionSequence struct {
	TotalWeight int
	Count       int
	Cards       []byte
}

// ControlPoker 控牌
func (game *RunFaster) ControlPoker() {
	var (
		//牌解序列
		solutionSequence []SolutionSequence

		// 最小手数
		smallestCount int
	)
	for i := 0; i < len(game.UserList); i++ {

		// 最优解
		cards := game.Poker.DrawCard()
		solutionCards := poker.GetOptimalSolutionCards(cards)

		// 一副牌手数
		solutionCount := len(solutionCards)

		// 最小手数
		if smallestCount == 0 {
			smallestCount = solutionCount
		}
		if solutionCount < smallestCount {
			smallestCount = solutionCount
		}

		solutionSequence = append(solutionSequence, SolutionSequence{
			TotalWeight: getTotalWeight(solutionCards),
			Count:       solutionCount,
			Cards:       cards,
		})
	}

	// 以3副牌中手数最少的哪一家为基准，多一手的玩家手牌权值减一分
	for i, solutionCards := range solutionSequence {
		solutionSequence[i].TotalWeight = solutionCards.TotalWeight - solutionCards.Count + smallestCount
	}

	// 对牌解序列按照从小到大进行排序
	for i := 0; i < len(solutionSequence)-1; i++ {
		for j := 0; j < (len(solutionSequence) - 1 - i); j++ {

			if solutionSequence[j].TotalWeight > solutionSequence[j+1].TotalWeight {
				solutionSequence[j], solutionSequence[j+1] = solutionSequence[j+1], solutionSequence[j]
			}
		}
	}

	// 分牌
	game.disposeCard(&solutionSequence)

	// 牌分完了提前返回
	if len(solutionSequence) == 0 {
		return
	}

	// 牌没有分完，随机分配给剩余玩家
	for _, v := range solutionSequence {
		for id := range game.UserList {
			if _, ok := game.ControlledCards[id]; !ok {
				game.ControlledCards[id] = v.Cards
				game.ControlList = append(game.ControlList, id)
				break
			}

		}
	}

}

// disposeCard 配牌
func (game *RunFaster) disposeCard(solutionSequence *[]SolutionSequence) {
	// 牌序列为空，已经完成了配牌
	if len(*solutionSequence) == 0 {
		return
	}

	// 概率分布
	ratePlace := make(map[int64]int)

	for _, id := range game.Seats {
		if len(game.ControlledCards[id]) != 0 {
			continue
		}

		// 没有点控，用血池
		prob := game.UserList[id].User.GetProb()
		if prob == 0 {
			prob = game.Table.GetRoomProb()
		}

		// 机器人 和 玩家 采用 不同的最大牌概率分布
		biggestRateDis := game.GameCfg.Control.PlayerBiggestRate
		if game.UserList[id].User.IsRobot() {
			biggestRateDis = game.GameCfg.Control.RobotBiggestRate
		}

		// 检测血池值, 获取控制等级索引
		probIndex := game.checkProb(prob)
		if probIndex == -1 {
			log.Errorf("游戏 %d 错误的作弊率值: %d", game.Table.GetID(), prob)

			// 默认 1000 作弊率的 索引
			probIndex = 2
		}

		if biggestRateDis[probIndex] != 0 {
			ratePlace[id] = biggestRateDis[probIndex]
		}

	}

	// 可控牌人数为0
	if len(ratePlace) == 0 {
		return
	}

	// 总概率值
	var totalRate int

	// 总概率值
	for _, rate := range ratePlace {
		totalRate += rate
	}

	// 未满10000的剩余概率值剩余平均概率
	lessAverageRate := (10000 - totalRate) / len(ratePlace)

	if lessAverageRate < 0 {
		lessAverageRate = 0
	}

	// 更新新概率值，让概率变得更加平缓
	for id, rate := range ratePlace {
		ratePlace[id] = lessAverageRate + rate
		totalRate += lessAverageRate
	}

	// 权重
	weight := rand.RandInt(1, totalRate+1)
	addRate := 0

	// 把最大牌给权值拥有者
	for id, rate := range ratePlace {
		if weight > addRate && weight <= addRate+rate {
			index := len(*solutionSequence) - 1
			game.ControlledCards[id] = (*solutionSequence)[index].Cards
			*solutionSequence = append((*solutionSequence)[:index], (*solutionSequence)[index+1:]...)
			game.ControlList = append(game.ControlList, id)
			break
		}
		addRate += rate
	}

	game.disposeCard(solutionSequence)
}

// 检测作弊率
func (game *RunFaster) checkProb(prob int32) (probIndex int) {
	probIndex = -1
	for index, rate := range game.GameCfg.Control.ControlRate {
		if prob == rate {
			probIndex = index
		}
	}

	return
}

// exchangeCard 换牌程序
// @ targetID 	目标用户ID
// @ exID 		被替换机器人ID
// @ targetCard 目标牌
// @ exCard 	被替换牌
func (game *RunFaster) exchangeCard(targetID int64, exID int64, targetCard byte, exCard byte) {

	log.Tracef("目标牌 %d，被替换牌 %d 将在目标手牌 %v，被替换手牌 %v 中 替换", targetCard, exCard, game.ControlledCards[targetID], game.ControlledCards[exID])

	// 目标牌，被提换牌在 手牌中的索引
	targetIndex, exIndex := -1, -1

	for index, card := range game.ControlledCards[targetID] {
		if targetCard == card {
			targetIndex = index
		}
	}

	// 目标牌没找到
	if targetIndex < 0 {
		log.Errorf("目标牌 %d 没有在目标手牌 %v 中找到", targetCard, game.ControlledCards[targetID])
		return
	}

	for index, card := range game.ControlledCards[exID] {
		if exCard == card {
			exIndex = index
		}
	}

	// 被提换牌没找到
	if exIndex < 0 {
		log.Errorf("被替换牌 %d 没有在被替换手牌 %v 中找到", exCard, game.ControlledCards[exID])
		return
	}

	// 换牌
	game.ControlledCards[targetID] = append(game.ControlledCards[targetID][:targetIndex], game.ControlledCards[targetID][targetIndex+1:]...)
	game.ControlledCards[targetID] = append(game.ControlledCards[targetID], exCard)

	game.ControlledCards[exID] = append(game.ControlledCards[exID][:exIndex], game.ControlledCards[exID][exIndex+1:]...)
	game.ControlledCards[exID] = append(game.ControlledCards[exID], targetCard)

}

//
func (game *RunFaster) ExchangeControl() {
	// 检查控制列队人数
	if len(game.ControlList) != 3 {
		log.Errorf("控制列队不足三人")
		return
	}

	var targetID, exID int64

	// 拿大牌的玩家为换牌目标
	for _, userID := range game.ControlList {
		if !game.UserList[userID].User.IsRobot() {
			targetID = userID
			break
		}
	}

	// 拿大牌的机器人为被换
	for _, userID := range game.ControlList {
		if game.UserList[userID].User.IsRobot() {
			exID = userID
			break
		}
	}

	// 都是玩家，无换牌操作
	if exID == 0 {
		log.Tracef(" 都是玩家，无换牌操作")
		return
	}

	// 未找到换牌目标，报错返回
	if targetID == 0 {
		log.Errorf("未找到换牌目标")
		return
	}

	// 换牌权重
	exchangeWeight := rand.RandInt(1, 10001)

	// 没有点控，用血池
	prob := game.UserList[targetID].User.GetProb()
	if prob == 0 {
		prob = game.Table.GetRoomProb()
	}

	// 检测血池值, 获取控制等级索引
	probIndex := game.checkProb(prob)
	if probIndex == -1 {
		log.Errorf("游戏 %d 错误的作弊率值: %d", game.Table.GetID(), prob)

		// 默认 1000 作弊率的 索引
		probIndex = 2
	}

	// 权重未落在目标区域, 不换牌
	if exchangeWeight > game.GameCfg.Control.ExchangeRate[probIndex] {
		log.Tracef(" 权重未落在目标区域, 不换牌")
		return
	}

	// 目标牌牌解策略
	targetSolutions := poker.GetOptimalSolutionCards(game.ControlledCards[targetID])

	// 被换牌解策略
	exSolutions := poker.GetOptimalSolutionCards(game.ControlledCards[exID])

	log.Tracef("换牌前，目标牌权重：%d，被换牌权重：%d", getTotalWeight(targetSolutions), getTotalWeight(exSolutions))

	// 目标牌 牌组，炸弹因子
	targetArr, boomCard := FindTargetArr(game.ControlledCards[targetID], targetSolutions)

	// 被换牌 牌组
	exArr := FindExArr(exSolutions)

	// 无牌可换
	if len(targetArr) == 0 || len(exArr) == 0 {
		return
	}

	log.Tracef("目标牌组：%v，被换牌牌组：%v", targetArr, exArr)

	// 将 从 被换牌 牌组 中 选 目标牌组 个数 牌 带入 拆过的 目标牌中 选分值最小的
	if len(exArr) > len(targetArr) {

		// 删除了 目标牌组 后的目标牌
		var delTargetCards []byte
		for _, card := range game.ControlledCards[targetID] {
			delTargetCards = append(delTargetCards, card)
		}

		// 删除元素
		for _, v := range targetArr {
			for index, card := range delTargetCards {

				if v == card {
					delTargetCards = append((delTargetCards)[:index], (delTargetCards)[index+1:]...)
					break
				}
			}
		}

		// 获取索引组合
		indexs := poker.ZuheResult(len(exArr), len(targetArr))

		// 所有组合
		combines := poker.FindNumsByIndexs(exArr, indexs)

		// 最小组合
		var smallestArr []byte

		// 初始权值
		smallestWeight := getTotalWeight(targetSolutions)
		log.Tracef("目标牌初始权重：%d", smallestWeight)

		// 获取最小
		for _, arr := range combines {

			leftTargetCards := append(delTargetCards, arr...)
			newTargetSolution := poker.GetOptimalSolutionCards(leftTargetCards)
			totalWeight := getTotalWeight(newTargetSolution)

			log.Tracef("加入被换牌牌组：%v后，目标牌权重 %d", arr, totalWeight)

			if totalWeight < smallestWeight {
				smallestArr = arr
			}

		}

		// 没有找到最小的
		if len(smallestArr) == 0 {
			return
		}

		exArr = smallestArr
	}

	// 从 目标牌牌组 中 抽取 小牌保留，使 换牌牌组个数 和 目标牌牌组个数一直（不能抽取炸弹因子）
	if len(exArr) < len(targetArr) {

		// 删除
		for i := 0; i < len(targetArr)-len(exArr); i++ {

			smallestIndex := 0
			smallestCard := targetArr[0]
			if smallestCard == boomCard {
				smallestIndex = 1
			}

			for index, card := range targetArr {

				// 不能删除炸弹因子
				if card == boomCard {
					continue
				}

				if card < smallestCard {
					smallestIndex = index
					smallestCard = card
				}
			}

			targetArr = append((targetArr)[:smallestIndex], (targetArr)[smallestIndex+1:]...)
		}
	}

	if len(targetArr) == len(exArr) && len(targetArr) != 0 {
		for i := 0; i < len(targetArr); i++ {
			game.exchangeCard(targetID, exID, targetArr[i], exArr[i])
		}
	}

	game.ExchangeControlTwo(targetID, exID, probIndex)
}

// ExchangeControlTwo 第二步换牌
func (game *RunFaster) ExchangeControlTwo(targetID int64, exID int64, probIndex int) {

	// 目标牌牌解策略
	targetSolutions := poker.GetOptimalSolutionCards(game.ControlledCards[targetID])

	// 被换牌解策略
	exSolutions := poker.GetOptimalSolutionCards(game.ControlledCards[exID])

	// 造炸弹权重
	buildBoomWeight := rand.RandInt(1, 10001)

	// 中了造炸弹权重
	if buildBoomWeight < game.GameCfg.Control.BuildBoomRate[probIndex] {
		targetArr, exArr := BuildBoom(exSolutions, game.ControlledCards[targetID])
		log.Tracef("机器人造炸弹，目标牌组：%v，被换牌牌组：%v", targetArr, exArr)
		// 换牌
		if len(targetArr) == len(exArr) && len(targetArr) != 0 {

			for i := 0; i < len(targetArr); i++ {
				game.exchangeCard(targetID, exID, targetArr[i], exArr[i])
			}
		}

		// 换 大小 对子，大小三同张权重
		exchangeBlockWeight := rand.RandInt(1, 10001)

		// 没中 换 大小 对子，大小三同张权重，返回
		if exchangeBlockWeight > game.GameCfg.Control.ExchangeBlockRate[probIndex] {
			return
		}

		// 目标牌牌解策略
		targetSolutions = poker.GetOptimalSolutionCards(game.ControlledCards[targetID])

		// 被换牌解策略
		exSolutions = poker.GetOptimalSolutionCards(game.ControlledCards[exID])
	}

	targetArr, exArr := FindBlock(targetSolutions, exSolutions)
	log.Tracef("换 大小 对子，大小三同张，目标牌组：%v，被换牌牌组：%v", targetArr, exArr)

	// 换 大小 对子，大小三同张
	if len(targetArr) == len(exArr) && len(targetArr) != 0 {
		for i := 0; i < len(targetArr); i++ {
			game.exchangeCard(targetID, exID, targetArr[i], exArr[i])
		}
	}

	// 目标牌牌解策略
	targetSolutions = poker.GetOptimalSolutionCards(game.ControlledCards[targetID])

	// 被换牌解策略
	exSolutions = poker.GetOptimalSolutionCards(game.ControlledCards[exID])

	log.Tracef("换牌后，目标牌权重：%d，被换牌权重：%d", getTotalWeight(targetSolutions), getTotalWeight(exSolutions))

}

// getTotalWeight 获取总权重
func getTotalWeight(solutions []poker.SolutionCards) (totalWeight int) {
	for _, once := range solutions {
		totalWeight += once.Weight
	}

	return
}

// 找到目标牌 换牌 牌组
func FindTargetArr(targetCards []byte, targetSolutions []poker.SolutionCards) (targetArr []byte, boomCard byte) {

	resultCards := FindBoom(targetSolutions)

	if len(resultCards) != 0 {
		// 找炸弹，拆一张
		targetArr = append(targetArr, resultCards...)

		boomCard = resultCards[0]
	}

	// 找2，无2找A，无A放弃
	targetArr = append(targetArr, FindBigSingle(targetCards)...)

	// 找黑桃3，找到后 50% 概率返回
	targetArr = append(targetArr, FindBlackThree(targetCards)...)

	// 关键牌：找 最大飞机的 中间张
	importCards := FindImportInPlane(targetSolutions)

	// 关键牌为空 找 最大顺子的 中间张
	if len(importCards) == 0 {
		importCards = FindImportInSequence(targetSolutions)
	}

	// 关键牌为空 找 找 权重大于J 的 三同张 或者 三带一，或者三带二的关键牌
	if len(importCards) == 0 {
		importCards = FindImportInTriplet(targetSolutions)
	}

	targetArr = append(targetArr, importCards...)

	// 可能出现重复牌，去重
	newArr := make([]byte, 0)

	for i := 0; i < len(targetArr); i++ {
		repeat := false
		for j := i + 1; j < len(targetArr); j++ {
			if targetArr[i] == targetArr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, targetArr[i])
		}
	}

	targetArr = newArr
	return
}

// 找到被换牌 换牌 牌组
func FindExArr(exSolutions []poker.SolutionCards) (exArr []byte) {

	// 寻找小于J的单张
	exArr = append(exArr, FindLessSingle(exSolutions)...)

	// 寻找小于10的对子
	exArr = append(exArr, FindLessPair(exSolutions)...)

	// 三带一，三带二中寻找 小于 10 的带牌
	exArr = append(exArr, FindLessWith(exSolutions)...)

	// 飞机带翅膀中寻找 小于 10 的带牌
	exArr = append(exArr, FindLessWing(exSolutions)...)

	return
}

// FindLessSingle 寻找小于J的单张
func FindLessSingle(exSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_SingleCard) {
			if v.Cards[0] < 0xb1 {
				resultCards = append(resultCards, v.Cards[0])
			}
		}
	}
	return
}

// FindLessPair 寻找小于10的对子
func FindLessPair(exSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_Pair) {
			if v.Cards[0] < 0xa1 {
				resultCards = append(resultCards, v.Cards...)
			}
		}
	}
	return
}

// FindLessWith 三带一，三带二中寻找 小于 10 的带牌
func FindLessWith(exSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_TripletWithSingle) {
			poker.SortCards(v.Cards)
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if v.Cards[3] < 0xa1 {
					resultCards = append(resultCards, v.Cards[3])
				}
			} else {
				if v.Cards[0] < 0xa1 {
					resultCards = append(resultCards, v.Cards[0])
				}
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithTwo) {
			poker.SortCards(v.Cards)
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if v.Cards[3] < 0xa1 {
					resultCards = append(resultCards, v.Cards[3])
				}

				if v.Cards[4] < 0xa1 {
					resultCards = append(resultCards, v.Cards[4])
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if v.Cards[0] < 0xa1 {
					resultCards = append(resultCards, v.Cards[0])
				}

				if v.Cards[4] < 0xa1 {
					resultCards = append(resultCards, v.Cards[4])
				}
			} else {
				if v.Cards[0] < 0xa1 {
					resultCards = append(resultCards, v.Cards[0])
				}

				if v.Cards[1] < 0xa1 {
					resultCards = append(resultCards, v.Cards[1])
				}
			}
		}
	}
	return
}

// FindLessWing 飞机带翅膀中寻找 小于 10 的带牌
func FindLessWing(exSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_SerialTripletWithTwo) {
			poker.SortCards(v.Cards)
			Len := len(v.Cards)
			for i := 0; (i + 2) < Len; i++ {
				if v.Cards[i]>>4 == v.Cards[i+2]>>4 {
					for j := 0; j < i; j++ {
						if v.Cards[j] < 0xa1 {
							resultCards = append(resultCards, v.Cards[j])
						}
					}
					start := 6 + i
					if Len == 15 {
						start = 9 + i
					}
					for m := start; m < Len; m++ {
						if v.Cards[m] < 0xa1 {
							resultCards = append(resultCards, v.Cards[m])
						} else {
							break
						}
					}
				}
			}
		}
	}
	return
}

// FindBoom 找炸弹，拆一张
func FindBoom(targetSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Bomb) {
			r := rand.RandInt(0, 4)
			resultCards = append(resultCards, v.Cards[r])
		}
	}
	return
}

// FindBigSingle 找2，无2找A，无A放弃
func FindBigSingle(targetCards []byte) (resultCards []byte) {
	temp := byte(0)
	for _, v := range targetCards {
		if v >= 0xe1 && v > temp {
			temp = v
		}
	}

	if temp != 0 {
		resultCards = append(resultCards, temp)
	}
	return
}

// FindBlackThree 找黑桃3，找到后 50% 概率返回
func FindBlackThree(targetCards []byte) (resultCards []byte) {
	for _, v := range targetCards {
		if v == 0x34 {
			//r := rand.RandInt(0, 100)
			//if r < 50 {
			resultCards = append(resultCards, v)
			//}

			break
		}
	}
	return
}

// FindImportInPlane 找 最大飞机的 中间张
func FindImportInPlane(targetSolution []poker.SolutionCards) (resultCards []byte) {
	temp := byte(0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_SerialTripletWithTwo) {
			poker.SortCards(v.Cards)
			Len := len(v.Cards)
			for i := 0; (i + 2) < Len; i++ {
				if v.Cards[i]>>4 == v.Cards[i+2]>>4 {
					if temp < v.Cards[i+3] {
						r := rand.RandInt(0, 3)
						temp = v.Cards[i+3+r]
					}
					break
				}
			}
		} else if v.CardsType == int32(msg.CardsType_SerialTriplet) {
			poker.SortCards(v.Cards)
			if temp < v.Cards[3] {
				r := rand.RandInt(0, 3)
				temp = v.Cards[3+r]
			}
		}
	}
	return
}

// FindImportInSequence 找 最大顺子的 中间张
func FindImportInSequence(targetSolution []poker.SolutionCards) (resultCards []byte) {
	temp := byte(0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Sequence) {
			mid := len(v.Cards) / 2
			ri := (len(v.Cards) + 1) % 2
			r := rand.RandInt(0, ri)
			if temp < v.Cards[mid+r] {
				temp = v.Cards[mid+r]
			}
		}
	}

	if temp != 0 {
		resultCards = append(resultCards, temp)
	}
	return
}

// FindImportInTriplet 找 权重大于J 的 三同张 或者 三带一，或者三带二的关键牌
func FindImportInTriplet(targetSolution []poker.SolutionCards) (resultCards []byte) {
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Triplet) {
			if v.Cards[0] >= 0xb1 {
				r := rand.RandInt(0, 3)
				resultCards = append(resultCards, v.Cards[r])
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithSingle) {
			poker.SortCards(v.Cards)
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if v.Cards[0] >= 0xb1 {
					r := rand.RandInt(0, 3)
					resultCards = append(resultCards, v.Cards[r])
				}
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithTwo) {
			poker.SortCards(v.Cards)
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if v.Cards[0] >= 0xb1 {
					r := rand.RandInt(0, 3)
					resultCards = append(resultCards, v.Cards[r])
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if v.Cards[1] >= 0xb1 {
					r := rand.RandInt(1, 4)
					resultCards = append(resultCards, v.Cards[r])
				}
			} else {
				if v.Cards[2] >= 0xb1 {
					r := rand.RandInt(2, 5)
					resultCards = append(resultCards, v.Cards[r])
				}
			}
		}
	}
	return
}

// FindBlock 对换大小对子或者大小三同张
func FindBlock(targetSolution []poker.SolutionCards, exSolution []poker.SolutionCards) (targetArr []byte, exArr []byte) {
	temp1 := getThree(targetSolution)
	temp2 := getSmallThree(exSolution)
	if len(temp1) > 0 && len(temp2) > 0 {
		if temp1[0] < temp2[0] {
			targetArr = append(targetArr, temp1...)
			exArr = append(exArr, temp2...)
			return
		}
	}

	temp3 := getTwo(targetSolution)
	temp4 := getSmallTwo(exSolution)
	if len(temp3) > 0 && len(temp4) > 0 {
		if temp3[0] < temp4[0] {
			targetArr = append(targetArr, temp3...)
			exArr = append(exArr, temp3...)
			return
		}
	}

	return
}

// BuildBoom 凑炸弹
func BuildBoom(exSolution []poker.SolutionCards, targetCards []byte) (targetArr []byte, exArr []byte) {
	card := byte(0)
	CardOne := GetCardOne(exSolution)
	if CardOne == 0 {
		return
	}

	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_Triplet) {
			card = getCardByValue(targetCards, v.Cards[0]>>4)
			if card != 0 {
				targetArr = append(targetArr, card)
				exArr = append(exArr, CardOne)
				return
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithSingle) {
			card = getCardByValue(targetCards, v.Cards[1]>>4)
			if card != 0 {
				targetArr = append(targetArr, card)
				exArr = append(exArr, CardOne)
				return
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithTwo) {
			card = getCardByValue(targetCards, v.Cards[2]>>4)
			if card != 0 {
				targetArr = append(targetArr, card)
				exArr = append(exArr, CardOne)
				return
			}
		}
	}
	return
}

//获取最大三同张
func getThree(targetSolution []poker.SolutionCards) []byte {
	temp1 := []byte{}
	temp1 = append(temp1, 0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Triplet) {
			if temp1[0] < v.Cards[0] {
				temp1 = make([]byte, 0)
				temp1 = append(temp1, v.Cards...)
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithSingle) {
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if temp1[0] < v.Cards[0] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[0:3]...)
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if temp1[0] < v.Cards[1] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[1:4]...)
				}
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithTwo) {
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if temp1[0] < v.Cards[0] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[0:3]...)
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if temp1[0] < v.Cards[1] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[1:4]...)
				}
			} else {
				if temp1[0] < v.Cards[2] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[2:6]...)
				}
			}
		} else if v.CardsType == int32(msg.CardsType_SerialTriplet) {
			if temp1[0] < v.Cards[len(v.Cards)-3] {
				temp1 = make([]byte, 0)
				temp1 = append(temp1, v.Cards[len(v.Cards)-3:len(v.Cards)]...)
			}
		} else if v.CardsType == int32(msg.CardsType_SerialTripletWithTwo) {
			Len := 4
			if len(v.Cards) == 15 {
				Len = 7
			}
			for i := 0; i < Len; i++ {
				if v.Cards[i]>>4 == v.Cards[i+2]>>4 {
					start := i + 3
					if len(v.Cards) == 15 {
						start = i + 6
					}

					if temp1[0] < v.Cards[start] {
						temp1 = make([]byte, 0)
						temp1 = append(temp1, v.Cards[start:start+4]...)
					}
				}
			}
		}
	}
	if temp1[0] == 0 {
		temp1 = make([]byte, 0)
	}

	return temp1
}

//获取最大的对子
func getTwo(targetSolution []poker.SolutionCards) []byte {
	temp := []byte{}
	temp = append(temp, 0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Pair) {
			if temp[0] < v.Cards[0] {
				temp = make([]byte, 0)
				temp = append(temp, v.Cards...)
			}
		} else if v.CardsType == int32(msg.CardsType_SerialPair) {
			Len := len(v.Cards)
			if temp[0] < v.Cards[Len-2] {
				temp = make([]byte, 0)
				temp = append(temp, v.Cards[Len-2:Len]...)
			}
		}
	}

	if temp[0] == 0 {
		temp = make([]byte, 0)
	}

	return temp
}

//获取最小三同张
func getSmallThree(targetSolution []poker.SolutionCards) []byte {
	temp1 := []byte{}
	temp1 = append(temp1, 0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Triplet) {
			if temp1[0] > v.Cards[0] {
				temp1 = make([]byte, 0)
				temp1 = append(temp1, v.Cards...)
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithSingle) {
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if temp1[0] > v.Cards[0] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[0:3]...)
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if temp1[0] > v.Cards[1] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[1:4]...)
				}
			}
		} else if v.CardsType == int32(msg.CardsType_TripletWithTwo) {
			if v.Cards[0]>>4 == v.Cards[2]>>4 {
				if temp1[0] > v.Cards[0] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[0:3]...)
				}
			} else if v.Cards[1]>>4 == v.Cards[3]>>4 {
				if temp1[0] > v.Cards[1] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[1:4]...)
				}
			} else {
				if temp1[0] > v.Cards[2] {
					temp1 = make([]byte, 0)
					temp1 = append(temp1, v.Cards[2:6]...)
				}
			}
		} else if v.CardsType == int32(msg.CardsType_SerialTriplet) {
			if temp1[0] > v.Cards[len(v.Cards)-3] {
				temp1 = make([]byte, 0)
				temp1 = append(temp1, v.Cards[len(v.Cards)-3:len(v.Cards)]...)
			}
		} else if v.CardsType == int32(msg.CardsType_SerialTripletWithTwo) {
			Len := 4
			if len(v.Cards) == 15 {
				Len = 7
			}
			for i := 0; i < Len; i++ {
				if v.Cards[i]>>4 == v.Cards[i+2]>>4 {
					start := i + 3
					if len(v.Cards) == 15 {
						start = i + 6
					}

					if temp1[0] > v.Cards[start] {
						temp1 = make([]byte, 0)
						temp1 = append(temp1, v.Cards[start:start+4]...)
					}
				}
			}
		}
	}
	if temp1[0] == 0 {
		temp1 = make([]byte, 0)
	}

	return temp1
}

//获取最小的对子
func getSmallTwo(targetSolution []poker.SolutionCards) []byte {
	temp := []byte{}
	temp = append(temp, 0)
	for _, v := range targetSolution {
		if v.CardsType == int32(msg.CardsType_Pair) {
			if temp[0] > v.Cards[0] {
				temp = make([]byte, 0)
				temp = append(temp, v.Cards...)
			}
		} else if v.CardsType == int32(msg.CardsType_SerialPair) {
			Len := len(v.Cards)
			if temp[0] > v.Cards[Len-2] {
				temp = make([]byte, 0)
				temp = append(temp, v.Cards[Len-2:Len]...)
			}
		}
	}

	if temp[0] == 0 {
		temp = make([]byte, 0)
	}

	return temp
}

func getCardByValue(targetCards []byte, targeCard byte) (card byte) {
	for _, v := range targetCards {
		if v&0xf0 == targeCard {
			card = v
			break
		}
	}

	return
}

func GetCardOne(exSolution []poker.SolutionCards) (card byte) {
	for _, v := range exSolution {
		if v.CardsType == int32(msg.CardsType_SingleCard) {
			card = v.Cards[0]
			break
		}
	}

	return
}
