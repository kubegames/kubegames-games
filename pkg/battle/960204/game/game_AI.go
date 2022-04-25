package game

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// Robot 机器人结构体
type Robot struct {
	User         player.RobotInterface
	cards        []byte
	TimerJob     *player.Job
	Cfg          config.RobotConfig
	GameLogic    *RunFaster            // 游戏逻辑，只能查看数据，不能修改数据
	BestSolution []poker.SolutionCards // 最优牌解
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, game table.TableHandler, robotCfg config.RobotConfig) {
	robot.User = userInter
	robot.Cfg = robotCfg
	robot.GameLogic = game.(*RunFaster)
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {

	switch subCmd {
	// 当前操作权玩家
	case int32(msg.SendToClientMessageType_S2CCurrentPlayer):
		robot.GetCurrentPlayer(buffer)
		break
		// 发牌信息
	case int32(msg.SendToClientMessageType_S2CDeal):
		robot.GetDeal(buffer)
		break
		// 出牌结果
	case int32(msg.SendToClientMessageType_S2CPutInfo):
		robot.HandlePutInfo(buffer)
		break
	}
}

// GetDeal 机器人获取发牌信息
func (robot *Robot) GetDeal(buffer []byte) {
	// 状态消息入参
	resp := &msg.DealRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析发牌信息失败: %v", err)
		return
	}

	if resp.UserId == robot.User.GetID() {
		robot.cards = resp.Cards
	}

}

// GetDeal 机器人处理出牌结果
func (robot *Robot) HandlePutInfo(buffer []byte) {
	// 状态消息入参
	resp := &msg.PutInfoRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析出牌结果失败: %v", err)
		return
	}

	// 删除自己出的牌
	if resp.IsSuccess && resp.UserId == robot.User.GetID() {
		for _, putCard := range resp.Cards {
			for i, card := range robot.cards {
				if putCard == card {
					robot.cards = append(robot.cards[:i], robot.cards[i+1:]...)
				}
			}
		}

	}
}

// GetCurrentPlayer 机器人处理当前操作玩家信息
func (robot *Robot) GetCurrentPlayer(buffer []byte) {

	// 状态消息入参
	resp := &msg.CurrentPlayerRes{}
	if err := proto.Unmarshal(buffer, resp); err != nil {
		log.Errorf("机器人解析当前操作玩家信息失败: %v", err)
		return
	}

	// 当前操作玩家是自己，并且有牌权
	if resp.UserId == robot.User.GetID() && resp.Permission {

		if resp.IsFinalEnd {
			log.Warnf("最后一手牌系统自动帮机器人 %d 出掉", robot.User.GetID())
			return
		}

		robot.BestSolution = poker.GetOptimalSolutionCards(robot.cards)

		if len(robot.BestSolution) == 0 {
			log.Errorf("机器人最优牌解 %v，手牌 %v 出现了错误", robot.BestSolution, robot.cards)
			return
		}

		var (
			solution          poker.SolutionCards // 解决方案
			preUser, nextUser *data.User          // 上家/下家
		)

		// 上家/下家 座位ID
		preChairID := robot.GameLogic.UserList[robot.User.GetID()].ChairID - 1
		nextChairID := robot.GameLogic.UserList[robot.User.GetID()].ChairID + 1
		if nextChairID == 3 {
			nextChairID = 0
		}
		if preChairID == -1 {
			preChairID = 2
		}

		// 上家
		preUser = robot.GameLogic.UserList[robot.GameLogic.Seats[preChairID]]

		// 下家
		nextUser = robot.GameLogic.UserList[robot.GameLogic.Seats[nextChairID]]

		switch resp.ActionType {
		// 出牌
		case int32(msg.UserActionType_PutCard):

			solution = robot.TakeOutCards(preUser, nextUser)
			break
			// 接牌
		case int32(msg.UserActionType_TakeOverCard):

			solution = robot.TakeOverCards(preUser, nextUser)
			break
		}

		if len(solution.Cards) != 0 {

			log.Warnf("机器人 %d, 出牌策略 %v", robot.User.GetID(), solution)
			robot.putCards(solution.Cards)
		}
	}

}

// TakeOutCards 机器人出牌
func (robot *Robot) TakeOutCards(preUser *data.User, nextUser *data.User) (solution poker.SolutionCards) {
	switch len(nextUser.Cards) {

	// 下家报单
	case 1:
		for _, once := range robot.BestSolution {
			if once.CardsType == int32(msg.CardsType_SingleCard) {
				continue
			}

			// 找到第一解决方案
			if len(solution.Cards) == 0 {
				solution = once
			}

			// 找到顺序值最小的解决方案
			if once.OrderValue < solution.OrderValue {
				solution = once
			}
		}

		log.Warnf("出牌，下家报单，机器人 %d 最优解中找手牌 %v", robot.User.GetID(), solution)

		// 只有单张牌型
		if len(solution.Cards) == 0 {
			for _, once := range robot.BestSolution {
				if once.WeightValue > solution.WeightValue {
					solution = once
				}
			}
			log.Warnf("出牌，下家报单，机器人 %d 最优解中没找到，找最大单张手牌 %v", robot.User.GetID(), solution)
		}
	// 下家报双
	case 2:
		nestCardsType := poker.GetCardsType(nextUser.Cards)

		for _, once := range robot.BestSolution {
			// 真双不出双
			if nestCardsType == msg.CardsType_Pair && once.CardsType == int32(msg.CardsType_Pair) {
				continue
			}

			// 假双不出单
			if nestCardsType != msg.CardsType_Pair && once.CardsType == int32(msg.CardsType_SingleCard) {
				continue
			}

			// 找到第一解决方案
			if len(solution.Cards) == 0 {
				solution = once
			}

			// 找到顺序值最小的解决方案
			if once.OrderValue < solution.OrderValue {
				solution = once
			}
		}

		log.Warnf("出牌，下家报双，机器人 %d 最优解中找手牌 %v", robot.User.GetID(), solution)

		// 没有出牌值最小的非双
		if len(solution.Cards) == 0 {
			for _, once := range robot.BestSolution {
				if once.WeightValue > solution.WeightValue {
					solution = once
				}
			}
			log.Warnf("出牌，下家报双，机器人 %d 最优解中没找到，找最大单张手牌 %v", robot.User.GetID(), solution)
		}

	default:

		// 下家不报单，查看自己手牌数是否为2
		if len(robot.BestSolution) == 2 {

			// 其他玩家接不上的解决方案
			var biggestSolution []poker.SolutionCards
			var singleBiggestSolution []poker.SolutionCards

			//遍历其他玩家手牌，如果有一手牌打出去其他玩家都接不起，则优先打该手牌
			for _, once := range robot.BestSolution {
				handCards := poker.HandCards{
					Cards:       once.Cards,
					WeightValue: once.WeightValue,
					CardsType:   once.CardsType,
				}

				// 检查其他玩家是否能接得起这副牌
				preTakeOverCards := poker.CheckTakeOverCards(handCards, preUser.Cards, false)
				nextTakeOverCards := poker.CheckTakeOverCards(handCards, nextUser.Cards, false)

				if len(preTakeOverCards) == 0 && len(nextTakeOverCards) == 0 {
					biggestSolution = append(biggestSolution, once)
				}

				var haveSingleBiggest bool

				// 比较单张牌值大小，获取 单张比其他人都大的牌组
				for _, cardValue := range once.CardsValue {
					biggerThanPre := true
					biggerThanNext := true

					// 遍历上家的手牌找到大过上家的牌
					for _, card := range preUser.Cards {
						preCardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue < preCardValue {
							biggerThanPre = false
						}
					}

					// 遍历下家的手牌找到大过下家的牌
					for _, card := range nextUser.Cards {
						nextCardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue < nextCardValue {
							biggerThanNext = true
						}
					}

					// 找到同时大过上家和下家大牌
					if biggerThanPre && biggerThanNext {
						haveSingleBiggest = true
						break
					}
				}

				if haveSingleBiggest {
					singleBiggestSolution = append(singleBiggestSolution, once)
				}

			}

			// 有其他玩家大不起的手牌
			if len(biggestSolution) != 0 {
				solution = biggestSolution[0]
				log.Warnf("出牌，手牌数为2，机器人 %d 最优解中有其他玩家大不起的手牌 %v", robot.User.GetID(), solution)
			} else {

				// 没有其他玩家大不起的手牌，有单张比其他人都大的牌组
				if len(singleBiggestSolution) == 1 {
					for _, once := range robot.BestSolution {

						// 跳过 有单张比其他人都大的牌组，选择另外一个牌组
						if reflect.DeepEqual(once, singleBiggestSolution[0]) {
							continue
						}
						solution = once
					}
					log.Warnf("出牌，手牌数为2，机器人 %d 最优解中没有其他玩家大不起的手牌，有单张比其他人都大的牌组手牌，保留单张大手牌，出另外一手牌 %v", robot.User.GetID(), solution)

				} else if len(singleBiggestSolution) == 2 {
					// 两幅牌都是 单张比其他人都大的牌组 中 取长度最长的手牌
					for _, once := range singleBiggestSolution {
						if len(once.Cards) > len(solution.Cards) {
							solution = once
						}
					}
					log.Warnf("出牌，手牌数为2，机器人 %d 最优解中没有其他玩家大不起的手牌，有2副单张比其他人都大的牌组手牌，取长度最长的手牌 %v", robot.User.GetID(), solution)

				} else if len(singleBiggestSolution) == 0 {
					// 既没有其他玩家大不起的手牌，也没有单张比其他人都大的牌组， 出长度较长的手牌
					for _, once := range robot.BestSolution {
						if len(once.Cards) > len(solution.Cards) {
							solution = once
						}
					}
					log.Warnf("出牌，手牌数为2，机器人 %d 最优解中既没有其他玩家大不起的手牌，也没有单张比其他人都大的手牌 出长度较长的手牌 %v", robot.User.GetID(), solution)

				}
			}

		} else {

			// 下家既不报单，自己的手牌数也不为2手， 考虑自己的剩余牌数量
			if len(robot.cards) <= 5 && len(robot.cards) >= 2 {
				// 压过下家的手牌
				var beOverSolution []poker.SolutionCards

				for _, once := range robot.BestSolution {
					handCards := poker.HandCards{
						Cards:       once.Cards,
						WeightValue: once.WeightValue,
						CardsType:   once.CardsType,
					}

					takeOverCards := poker.CheckTakeOverCards(handCards, nextUser.Cards, false)
					if len(takeOverCards) == 0 {
						beOverSolution = append(beOverSolution, once)
					}
				}

				// 没有能压过下家的手牌
				if len(beOverSolution) == 0 {
					// 按照出牌顺序值从小到大出牌
					for _, once := range robot.BestSolution {
						if len(solution.Cards) == 0 {
							solution = once
						}
						if once.OrderValue < solution.OrderValue {
							solution = once
						}
					}
					log.Warnf("出牌，手牌长度为2～5，机器人 %d 最优解中没有能压过下家的手牌，出顺序值最小的手牌 %v", robot.User.GetID(), solution)
				} else {
					// 最小可压过下家的手牌
					var smallestOverSolution poker.SolutionCards
					for _, once := range beOverSolution {
						if len(smallestOverSolution.Cards) == 0 {
							smallestOverSolution = once
						}
						if once.OrderValue < smallestOverSolution.OrderValue {
							smallestOverSolution = once
						}
					}

					// 是否能拿回牌权
					canGetPower := true

					// 除开最小可压过下家的手牌 最大的牌值牌
					var biggestCardValue byte

					for _, card := range robot.cards {
						// 等待删除的牌
						var waitDealCard bool
						for _, solutionCard := range smallestOverSolution.Cards {
							if card == solutionCard {
								waitDealCard = true
								break
							}
						}
						if waitDealCard {
							continue
						}
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							biggestCardValue = cardValue
						}
					}

					if biggestCardValue == 0 {
						log.Errorf("机器人手牌 %v 没找到最大牌", robot.cards)
						return
					}

					for _, card := range preUser.Cards {
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							canGetPower = false
						}
					}
					for _, card := range nextUser.Cards {
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							canGetPower = false
						}
					}

					// 除开最小可压过下家的手牌，有其他牌有牌值比其他人的牌都大的牌，选择压
					if canGetPower {
						solution = smallestOverSolution
						log.Warnf("出牌，手牌长度为2～5，机器人 %d 最优解中有能压过下家的手牌，并且可拿回牌权，出最小可压过下家的手牌 %v", robot.User.GetID(), solution)
					} else {

						// 拿不到牌权， 按照出牌顺序值从小到大出牌
						for _, once := range robot.BestSolution {
							if len(solution.Cards) == 0 {
								solution = once
							}
							if once.OrderValue < solution.OrderValue {
								solution = once
							}
						}
						log.Warnf("出牌，手牌长度为2～5，机器人 %d 最优解中有能压过下家的手牌，但拿不回牌权，出顺序值最小的手牌 %v", robot.User.GetID(), solution)
					}

				}

			} else if len(robot.cards) > 5 {

				// 先寻找是否有顺序值小于0 的手牌
				for _, once := range robot.BestSolution {
					if once.OrderValue >= 0 {
						continue
					}
					if len(solution.Cards) == 0 {
						solution = once
					}
					if once.OrderValue < solution.OrderValue {
						solution = once
					}
				}
				log.Warnf("出牌，手牌长度 >5，机器人 %d 最优解中寻找顺序值小于0的最小手牌 %v", robot.User.GetID(), solution)

				// 没有顺序值小于0的手牌，找同种牌型顺序差值最大 中 最小的手牌
				if len(solution.Cards) == 0 {
					haveSameTypeSolution := map[int32][]poker.SolutionCards{}
					for _, once := range robot.BestSolution {
						if once.CardsType == int32(msg.CardsType_Bomb) {
							continue
						}
						haveSameTypeSolution[once.CardsType] = append(haveSameTypeSolution[once.CardsType], once)
					}

					for _, solutionArr := range haveSameTypeSolution {
						if len(solutionArr) >= 2 {

							// 最大顺序值牌组， 最小顺序值牌组
							var biggestOrderSolution, smallestOrderSolution poker.SolutionCards

							for _, once := range solutionArr {
								if len(biggestOrderSolution.Cards) == 0 {
									biggestOrderSolution = once
								}
								if len(smallestOrderSolution.Cards) == 0 {
									smallestOrderSolution = once
								}
								if once.OrderValue > biggestOrderSolution.OrderValue {
									biggestOrderSolution = once
								}
								if once.OrderValue < smallestOrderSolution.OrderValue {
									smallestOrderSolution = once
								}
							}

							// 同种牌型差值大于100，出最小牌
							if biggestOrderSolution.OrderValue-smallestOrderSolution.OrderValue > 100 {
								solution = smallestOrderSolution
								break
							}
						}
					}
					log.Warnf("出牌，手牌长度 >5，机器人 %d 最优解中寻找同种牌型有差值并且差值最大的小手牌 %v", robot.User.GetID(), solution)
				}

				// 最优解中没有同种牌型， 按照顺序值出牌
				if len(solution.Cards) == 0 {
					for _, once := range robot.BestSolution {
						if len(solution.Cards) == 0 {
							solution = once
						}
						if once.OrderValue < solution.OrderValue {
							solution = once
						}
					}
					log.Warnf("出牌，手牌长度 >5，机器人 %d 最优解中无同种牌型，出顺序值最小的小手牌 %v", robot.User.GetID(), solution)

				}
			}

		}

	}

	// 手上有牌，但是计算不出出牌策略
	if len(solution.Cards) == 0 && len(robot.cards) != 0 {
		log.Errorf("机器人 出牌 最优牌解 %v，手牌 %v 出现了错误", robot.BestSolution, robot.cards)
	}

	return
}

// TakeOverCards 机器人接牌
func (robot *Robot) TakeOverCards(preUser *data.User, nextUser *data.User) (solution poker.SolutionCards) {

	nextUserCardsLen := len(nextUser.Cards)

	// 在最优牌解中找到能接的起的手牌
	var takeOverSolutions []poker.SolutionCards
	for _, once := range robot.BestSolution {

		// 同种牌型，长度相同，找牌权重值大的，或者用炸弹大过其他牌型
		if (len(robot.GameLogic.CurrentCards.Cards) == len(once.Cards) &&
			robot.GameLogic.CurrentCards.CardsType == once.CardsType &&
			robot.GameLogic.CurrentCards.WeightValue < once.WeightValue) ||
			(robot.GameLogic.CurrentCards.CardsType != int32(msg.CardsType_Bomb) &&
				once.CardsType == int32(msg.CardsType_Bomb)) {
			takeOverSolutions = append(takeOverSolutions, once)
		}
	}

	if len(takeOverSolutions) > 0 {
		// 下家报单
		if nextUserCardsLen == 1 {

			// 下家报单 如果是接单张，则需要按照牌值从大到小顺序接牌
			if robot.GameLogic.CurrentCards.CardsType == int32(msg.CardsType_SingleCard) {
				cards := poker.SortCards(robot.cards)

				solution = poker.SolutionCards{
					Cards: []byte{cards[len(cards)-1]},
				}
				log.Warnf("接牌，下家报单，机器人 %d 接单牌，出牌值最大的手牌 %v", robot.User.GetID(), solution)
			} else {

				// 接非单张牌 按照出牌顺序值从小到大出牌
				for _, once := range takeOverSolutions {
					if len(solution.Cards) == 0 {
						solution = once
					}
					if once.OrderValue < solution.OrderValue {
						solution = once
					}
				}
				log.Warnf("接牌，下家报单，机器人 %d 接非单牌，出顺序值最小的手牌 %v", robot.User.GetID(), solution)
			}

		} else if nextUserCardsLen >= 2 && nextUserCardsLen <= 5 {

			// 压过下家的手牌
			var beOverSolution []poker.SolutionCards

			for _, once := range takeOverSolutions {
				handCards := poker.HandCards{
					Cards:       once.Cards,
					WeightValue: once.WeightValue,
					CardsType:   once.CardsType,
				}

				takeOverCards := poker.CheckTakeOverCards(handCards, nextUser.Cards, false)
				if len(takeOverCards) == 0 {
					beOverSolution = append(beOverSolution, once)
				}
			}

			// 没有能压过下家的手牌
			if len(beOverSolution) == 0 {
				// 按照出牌顺序值从小到大出牌
				for _, once := range takeOverSolutions {
					if len(solution.Cards) == 0 {
						solution = once
					}
					if once.OrderValue < solution.OrderValue {
						solution = once
					}
				}
				log.Warnf("接牌，下家牌组长度为2～5，机器人 %d 没压过下家的手牌，出顺序值最小的手牌 %v", robot.User.GetID(), solution)
			} else {

				// 最小可压过下家的手牌
				var smallestOverSolution poker.SolutionCards
				for _, once := range beOverSolution {
					if len(smallestOverSolution.Cards) == 0 {
						smallestOverSolution = once
					}
					if once.OrderValue < smallestOverSolution.OrderValue {
						smallestOverSolution = once
					}
				}
				nextCardsType := poker.GetCardsType(nextUser.Cards)

				// 下家一手牌就可以出完的情况下,选择压牌
				if nextCardsType != msg.CardsType_Normal {
					solution = smallestOverSolution
				} else {

					// 是否能拿回牌权
					canGetPower := true

					// 除开最小可压过下家的手牌 最大的牌值牌
					var biggestCardValue byte

					for _, card := range robot.cards {
						// 等待删除的牌
						var waitDealCard bool
						for _, solutionCard := range smallestOverSolution.Cards {
							if card == solutionCard {
								waitDealCard = true
								break
							}
						}
						if waitDealCard {
							continue
						}
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							biggestCardValue = cardValue
						}
					}

					if biggestCardValue == 0 {
						log.Errorf("机器人手牌 %v 没找到最大牌", robot.cards)
						return
					}

					for _, card := range preUser.Cards {
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							canGetPower = false
						}
					}
					for _, card := range nextUser.Cards {
						cardValue, _ := poker.GetCardValueAndColor(card)
						if cardValue > biggestCardValue {
							canGetPower = false
						}
					}

					// 除开最小可压过下家的手牌，有其他牌有牌值比其他人的牌都大的牌，选择压
					if canGetPower {
						solution = smallestOverSolution
						log.Warnf("接牌，下家牌组长度为2～5，机器人 %d 有压过下家的手牌，并且可以拿回出牌权，出最小可压过下家的手牌 %v", robot.User.GetID(), solution)
					} else {

						// 拿不到牌权， 按照出牌顺序值从小到大出牌
						for _, once := range takeOverSolutions {
							if len(solution.Cards) == 0 {
								solution = once
							}
							if once.OrderValue < solution.OrderValue {
								solution = once
							}
						}
						log.Warnf("接牌，下家牌组长度为2～5，机器人 %d 有压过下家的手牌，拿不回出牌权，出顺序值最小的手牌 %v", robot.User.GetID(), solution)
					}

				}

			}

		} else if nextUserCardsLen > 5 {

			// 按照出牌顺序值从小到大出牌
			for _, once := range takeOverSolutions {
				if len(solution.Cards) == 0 {
					solution = once
				}
				if once.OrderValue < solution.OrderValue {
					solution = once
				}
			}
			log.Warnf("接牌，下家牌组长度>5，机器人 %d 有压过下家的手牌，拿不回出牌权，出顺序值最小的手牌 %v", robot.User.GetID(), solution)

		}
	} else {
		// 拆解最优解出牌
		solution = poker.SolutionCards{
			Cards: poker.CheckTakeOverCards(robot.GameLogic.CurrentCards, robot.cards, false),
		}
		log.Warnf("接牌，机器人 %d 最优牌解中无牌解，出拆解后的手牌 %v", robot.User.GetID(), solution)

	}

	if len(solution.Cards) == 0 {
		log.Errorf("机器人 接牌 最优牌解 %v，手牌 %v 出现了错误", robot.BestSolution, robot.cards)
	}

	return
}

// GetCurrentPlayer 机器人处理当前操作玩家信息
func (robot *Robot) putCards(cards []byte) {
	// 机器人操作时间权重
	actionTimeWeight := rand.RandInt(0, 101)

	// 概率基数
	rateBase := 0

	// 机器人操作时间区域索引
	actionTimeRegionIndex := 0
	for index, rate := range robot.Cfg.ActionTimeRatePlace {
		downLimit := rateBase
		upLimit := rateBase + rate

		rateBase = upLimit
		if actionTimeWeight > downLimit && actionTimeWeight <= upLimit {
			actionTimeRegionIndex = index
			break
		}
	}

	// 所及延迟操作时间
	randomTime := 1

	randomTime = rand.RandInt(robot.Cfg.ActionTimePlace[actionTimeRegionIndex][0], robot.Cfg.ActionTimePlace[actionTimeRegionIndex][1]+1)

	// 防止随机到 0
	if randomTime == 0 {
		randomTime = 1
	}

	req := msg.PutCardsReq{
		Cards: cards,
	}

	// 延迟发送消息
	robot.TimerJob, _ = robot.User.AddTimer(int64(randomTime*1000), func() {
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SPutCards), &req)
		if err != nil {
			log.Errorf("机器人发送出牌请求失败: %v", err.Error())
		}
	})

}
