package game

import (
	"common/log"
	"common/rand"
	"time"

	//"common/log"
	//"common/rand"
	"fmt"
	"game_frame_v2/game/clock"
	"game_poker/doudizhu/config"
	"game_poker/doudizhu/data"

	"game_poker/doudizhu/msg"
	"game_poker/doudizhu/poker"
	"runtime"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
	b3 "github.com/magicsea/behavior3go"
	b3core "github.com/magicsea/behavior3go/core"
)

//出牌数据
type outData struct {
	//当前出牌回合, 从0开始递增
	round int32
	//玩家当前回合出牌类型
	option int
	//玩家当前回合出牌数据
	cards []byte
}

type robot struct {
	UserInter    player.RobotInterface
	TimerJob     *clock.Job
	GameLogic    *DouDizhu             // 游戏逻辑，只能查看数据，不能修改数据
	BestSolution []poker.SolutionCards // 最优牌解
	RobScore     int                   // 抢庄分值
	//游戏配置
	robotCfg config.RobotConfig // 机器人配置
	//机器人id
	robotID string
	//机器人玩家实例
	User *data.User // 用户
	//机器人座位号
	mySeat int32
	//机器人定时器
	timerID int
	//机器人手牌拆分组合
	cardGroups []*CardGroup
	//当前回合第一个出牌玩家
	firstOutSeat int32

	// there is too fucking much place where leftCards was assigned
	// so update left cards most value group in export data update, this should be ok
	leftCardGroups []*CardGroup
	standCard      *byte
	//当前出牌回合数
	round int32
	//所有玩家的出牌记录
	outCards [][]*outData
	//桌面剩余牌
	deskLeftCards []byte
	//行为树指针，房间持有，保证线程安全
	tree *b3core.BehaviorTree
	//导出给行为树的数据
	exportData *ExportData
	//GameLogic.LastOutCardType  []msg.CardsType
	//GameLogic.LastPassCardType []msg.CardsType
}

//机器人出牌
func (robot *robot) PutCards(cards []byte) {
	req := msg.PutCardsReq{
		Cards: cards,
	}

	err := robot.UserInter.SendMsgToServer(int32(msg.ReceiveMessageType_C2SPutCards), &req)
	if err != nil {
		log.Errorf("send server msg fail: %v", err.Error())
	}
}

func (rb *robot) getCardsByGroup(t msg.CardsType) []CardGroup {
	res := make([]CardGroup, 0)
	for _, g := range rb.cardGroups {
		if g.Type == t {
			res = append(res, *g)
		}
	}
	return res
}
func (rb *robot) getCardCount(n byte) int {
	res := 0
	for _, g := range rb.User.Cards {
		if g>>4 == n {
			res++
		}
	}
	return res
}
func (rb *robot) getCard(n byte) []byte {
	res := make([]byte, 0)
	for _, g := range rb.User.Cards {
		if g>>4 == n {
			res = append(res, g)
		}
	}
	return res
}

// Init 初始化机器人
func (robot *robot) Init(userInter player.RobotInterface, game table.TableHandler, robotCfg config.RobotConfig, tree *b3core.BehaviorTree) {
	robot.UserInter = userInter
	robot.GameLogic = game.(*DouDizhu)
	robot.User = game.(*DouDizhu).UserList[userInter.GetId()]
	robot.robotCfg = robotCfg
	robot.tree = tree
	robot.mySeat = robot.User.ChairID
	robot.resetRobotInfo()
}

//来自游戏框架层的消息
func (rb *robot) OnFrameMsg(eventType int32, event proto.Message) {
	//这里不用处理框架层的消息，所有事件从DealEvent处理
}

//处理具体游戏事件
func (rb *robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	/*case int32(msg.SendToClientMessageType_S2CDeal):
		rb.processGameStart()
	case int32(msg.SendToClientMessageType_S2CRobResult):
		rb.processCallBanker()
	case int32(msg.SendToClientMessageType_S2CConfirmDizhu):
		rb.processEnsureBanker()
	case int32(msg.SendToClientMessageType_S2CRedoubleResult):
		rb.processPlayerDouble()*/
	case int32(msg.SendToClientMessageType_S2CCurrentPlayer):
		rb.processStartOutCard()
	case int32(msg.SendToClientMessageType_S2CGameStatus):
		rb.DealGameStatus(buffer)
		// 当前抢庄玩家
	case int32(msg.SendToClientMessageType_S2CCurrentRobber):
		rb.DealCurrentRobber(buffer)
	default:
	}
}

//定时器，当游戏处于运行状态时，框架层会每一段时间调用一次该接口
//输入： timeId:定时器id, userData:设置定时器时传给框架层的参数
//输出：shouldContinue: 定时器是否继续，如果返回false，框架层将自动撤销该定时器
func (rb *robot) OnTimeout(timerID int, userData interface{}) (shouldContinue bool) {
	userData.(func())()
	return false
}

//获取导出给决策树的导出数据，此接口必须返回结构体的指针
func (rb *robot) GetExportData() interface{} {
	if rb.exportData == nil {
		rb.exportData = &ExportData{}
	}
	return rb.exportData
}

func (rb *robot) updateExportData() {
	rb.exportData.GameStep = int(rb.GameLogic.Status)
	if rb.GameLogic.Dizhu != nil {
		rb.exportData.IsBanker = rb.GameLogic.Dizhu.ChairID == rb.mySeat
	} else {
		rb.exportData.IsBanker = false
	}

	if nil == rb.GameLogic.Dizhu {
		rb.exportData.IsBankerPrev = false
		rb.exportData.IsBankerNext = false
		rb.exportData.BankerCardsCount = 0
		rb.exportData.BankerPrevCardsCount = 0
		rb.exportData.BankerNextCardsCount = 0
	} else {
		bankerseat := rb.GameLogic.Dizhu.ChairID
		bankerprev := (bankerseat + 2) % 3
		bankernext := (bankerseat + 1) % 3
		rb.exportData.IsBankerPrev = bankerprev == rb.mySeat
		rb.exportData.IsBankerNext = bankernext == rb.mySeat
		rb.exportData.BankerCardsCount = len(rb.GameLogic.Dizhu.Cards)
		rb.exportData.BankerPrevCardsCount = len(rb.GameLogic.Chairs[bankerprev].Cards)
		rb.exportData.BankerNextCardsCount = len(rb.GameLogic.Chairs[bankernext].Cards)
	}
	rb.exportData.HandCardsValue, rb.cardGroups = GetMostValueGroup(rb.User.Cards)
	//这些是跟出牌有关的信息，这里做判断，避免一些计算
	if int32(msg.GameStatus_PutCardStatus) == rb.GameLogic.Status {
		rb.exportData.IsLeftOneCards = GetCardsType(rb.User.Cards) > msg.CardsType_Normal
		rb.exportData.IsCanOutAll = CompareCards(rb.GameLogic.CurrentCards.Cards, rb.User.Cards)
		rb.exportData.IsFirstOut = 0 == rb.GameLogic.CurrentCards.UserID
		if nil == rb.GameLogic.Dizhu || 0 == rb.GameLogic.CurrentCards.UserID {
			rb.exportData.IsBankeOut = false
			rb.exportData.IsBankePrevOut = false
			rb.exportData.IsBankeNextOut = false
			rb.exportData.IsBankerPass = false
			rb.exportData.IsBankerPrevPass = false
			rb.exportData.IsBankerNextPass = false
		} else {
			bankerprev := (rb.GameLogic.Dizhu.ChairID + 2) % 3
			bankernext := (rb.GameLogic.Dizhu.ChairID + 1) % 3
			rb.exportData.IsBankeOut = rb.GameLogic.Dizhu.User.GetId() == rb.GameLogic.CurrentCards.UserID
			maxOutSeat := rb.GameLogic.UserList[rb.GameLogic.CurrentCards.UserID]
			rb.exportData.IsBankePrevOut = bankerprev == maxOutSeat.ChairID
			rb.exportData.IsBankeNextOut = bankernext == maxOutSeat.ChairID
			rb.exportData.IsBankerPass = rb.GameLogic.IsPass[rb.GameLogic.Dizhu.ChairID]
			rb.exportData.IsBankerPrevPass = rb.GameLogic.IsPass[bankerprev]
			rb.exportData.IsBankerNextPass = rb.GameLogic.IsPass[bankernext]

			myprev := (rb.User.ChairID + 2) % 3
			mynext := (rb.User.ChairID + 1) % 3
			for _, v := range rb.cardGroups {
				ret, _ := SearchLargerCardType(v.Cards, rb.GameLogic.Chairs[mynext].Cards, true)
				if !ret {
					ret, _ = SearchLargerCardType(v.Cards, rb.GameLogic.Chairs[myprev].Cards, true)
					if !ret {
						rb.exportData.AllBigCount++
					}
				}
			}
		}

		rb.exportData.IsOutNotSigleAndDouble = false
		outType := GetCardsType(rb.GameLogic.CurrentCards.Cards)
		if rb.GameLogic.CurrentCards.UserID == rb.UserInter.GetId() {
			outType = msg.CardsType_Normal
			rb.exportData.IsFirstOut = true
		} else {
			if rb.GameLogic.CurrentCards.CardsType == int32(msg.CardsType_SingleCard) ||
				rb.GameLogic.CurrentCards.CardsType == int32(msg.CardsType_Pair) {
				rb.exportData.IsOutNotSigleAndDouble = true
			}

			if nil != SearchFirstLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false) {
				rb.exportData.IsHaveSametype = true
			} else {
				rb.exportData.IsHaveSametype = false
				IsBig, out := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false)
				if IsBig {
					temp := CloneCards(rb.User.Cards)
					RemoveCards(temp, out[0])
					_, temp1 := GetMostValueGroup(temp)
					cardsnum := 0

					for _, v := range rb.cardGroups {
						if v.Type == msg.CardsType_SingleCard {
							cardsnum++
						}
					}

					for _, v := range temp1 {
						if v.Type == msg.CardsType_SingleCard {
							cardsnum--
						}
					}

					rb.exportData.DemCardMySingleCardsNum = cardsnum
				}
			}
		}

		rb.exportData.IsOutSingle = outType == msg.CardsType_SingleCard
		rb.exportData.IsOutDouble = outType == msg.CardsType_Pair
		rb.exportData.OutCardCount = len(rb.GameLogic.CurrentCards.Cards)
		if rb.exportData.OutCardCount > 0 {
			rb.exportData.OutCardValue = int(rb.GameLogic.CurrentCards.Cards[0] >> 4)
		} else {
			rb.exportData.OutCardValue = 0
		}

		//自己手牌是否满足一组小牌， 其他绝对大牌的赢牌路径
		rb.exportData.IsAllBigWin = IsCanAbsWin(rb.cardGroups, rb.GameLogic.LeftCards, rb.mySeat != rb.GameLogic.Dizhu.ChairID)
		//自己跟牌之后是否能进入一组小牌， 其他绝对大牌的赢牌路径
		//跟牌情况下才计算
		if nil == rb.GameLogic.CurrentCards.Cards {
			rb.exportData.IsCanAllBig = false
		} else {
			rb.exportData.IsCanAllBig, _ = GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.GameLogic.LeftCards, rb.mySeat != rb.GameLogic.Dizhu.ChairID)
		}
	}
	//自己手牌情况
	bombcount, mycounts := GetCardCounts(rb.User.Cards)
	rb.exportData.IsHasBigKing = mycounts[15] > 0
	rb.exportData.IsHasSmallKing = mycounts[14] > 0
	rb.exportData.MyPoker2Count = mycounts[13]
	rb.exportData.MyPoker2More = mycounts[13] + mycounts[14] + mycounts[15]
	rb.exportData.MyPokerACount = mycounts[12]
	rb.exportData.MyPokerAMore = mycounts[12] + mycounts[13] + mycounts[14] + mycounts[15]
	rb.exportData.IsMyRoket = (mycounts[14] + mycounts[15]) == 2
	if mycounts[14] > 0 && mycounts[15] > 0 {
		rb.exportData.MyBombCount = bombcount + 1
	} else {
		rb.exportData.MyBombCount = bombcount
	}
	bombcountMaxValue := 0
	for _, g := range rb.cardGroups {
		if g.Type == msg.CardsType_Bomb || g.Type == msg.CardsType_Rocket {
			bombcountMaxValue++
		}
	}
	rb.exportData.MyBombCountMaxValue = bombcountMaxValue
	rb.exportData.MyNoKingBombCount = bombcount
	rb.exportData.HandCardGroupNum = len(rb.cardGroups)
	_, rb.leftCardGroups = GetMostValueGroup(rb.GameLogic.LeftCards)

	maxValue := -1
	for _, g := range rb.leftCardGroups {
		if g.Type == msg.CardsType_SingleCard {
			if g.Value > maxValue {
				maxValue = g.Value
				rb.standCard = &g.Cards[0]
			}
		}
	}

	var leftMaxBombValue, leftMaxSingleValue, leftMaxDoubleValue int
	for _, g := range rb.leftCardGroups {
		if g.Type == msg.CardsType_Bomb {
			leftMaxBombValue = g.Value
		}
		if g.Type == msg.CardsType_SingleCard {
			leftMaxSingleValue = g.Value
		}
		if g.Type == msg.CardsType_Pair {
			leftMaxDoubleValue = g.Value
		}
	}
	bombs := rb.getCardsByGroup(msg.CardsType_Bomb)
	getMaxValueCardGroup := func(gs []CardGroup) *CardGroup {
		maxValue := 0
		maxIdx := -1
		for i, g := range gs {
			if g.Value > maxValue {
				maxIdx = i
				maxValue = g.Value
			}
		}
		return &gs[maxIdx]
	}
	if len(bombs) > 0 {
		maxGroup := getMaxValueCardGroup(bombs)
		rb.exportData.IsMyBombGreaterThanLeft = maxGroup.Value > leftMaxBombValue
	} else {
		rb.exportData.IsMyBombGreaterThanLeft = false
	}
	singles := rb.getCardsByGroup(msg.CardsType_SingleCard)
	if len(singles) > 0 {
		rb.exportData.MySingleCardsNum = len(singles)
		maxGroup := getMaxValueCardGroup(singles)
		rb.exportData.IsMySingleGreaterThanLeft = maxGroup.Value > leftMaxSingleValue
	} else {
		rb.exportData.MySingleCardsNum = 0
	}
	doubles := rb.getCardsByGroup(msg.CardsType_Pair)
	if len(doubles) > 0 {
		rb.exportData.MyDoubleCardsNum = len(doubles)
		maxGroup := getMaxValueCardGroup(doubles)
		rb.exportData.IsMyDoubleGreaterThanLeft = maxGroup.Value > leftMaxDoubleValue
	} else {
		rb.exportData.MyDoubleCardsNum = 0
	}
	//其他人手牌情况
	otcounts := mycounts
	rb.exportData.OtherBombCount = 0
	for i := 0; i < 3; i++ {
		if i == int(rb.User.ChairID) {
			continue
		}
		bombcount, otcounts = GetCardCounts(rb.GameLogic.Chairs[i].Cards)
		rb.exportData.IsOtherBigKing = otcounts[15] > 0
		rb.exportData.IsOtherSmallKing = otcounts[14] > 0
		rb.exportData.OtherPoker2Count = otcounts[15]
		rb.exportData.OtherPoker2More = otcounts[13] + otcounts[14] + otcounts[15]
		rb.exportData.OtherPokerACount = otcounts[12]
		rb.exportData.OtherPokerAMore = otcounts[12] + otcounts[13] + otcounts[14] + otcounts[15]
		if otcounts[14] > 0 && otcounts[15] > 0 {
			rb.exportData.OtherBombCount += bombcount + 1
		} else {
			rb.exportData.OtherBombCount += bombcount
		}
	}

	for i := len(otcounts) - 1; i >= 3; i-- {
		if otcounts[i] >= 1 {
			rb.exportData.OtherMaxSingle = i
			break
		}
	}
	for i := len(otcounts) - 1; i >= 3; i-- {
		if otcounts[i] >= 2 {
			rb.exportData.OtherMaxDouble = i
			break
		}
	}

	if rb.GameLogic.Status == int32(msg.GameStatus_RedoubleStatus) {
		rb.exportData.MyHandCardsValue = GetThreeCardsHolderInstance().GetCardsValue(int(rb.mySeat))
	} else {
		rb.exportData.MyHandCardsValue = GetThreeCardsHolderInstance().GetCardsValue3(int(rb.mySeat))
	}
	if rb.GameLogic.Status == int32(msg.GameStatus_RedoubleStatus) || rb.GameLogic.Status == int32(msg.GameStatus_RobStatus) {
		{
			myV := GetThreeCardsHolderInstance().GetCardsValue3(int(rb.mySeat)) // self with 3, others not

			a := 0
			b := 0
			other_max := 0
			if rb.mySeat == 0 {
				a = 1
				b = 2
			} else if rb.mySeat == 1 {
				a = 0
				b = 2
			} else {
				a = 0
				b = 1
			}
			av := GetThreeCardsHolderInstance().GetCardsValue(a)
			bv := GetThreeCardsHolderInstance().GetCardsValue(b)
			if av > bv {
				other_max = av
			} else {
				other_max = bv
			}
			rb.exportData.CardsValueDiffWithAnotherMax = myV - other_max
		}
		if rb.GameLogic.Dizhu != nil {
			myV := 0
			if rb.GameLogic.Dizhu.ChairID == rb.mySeat {
				rb.exportData.CardsValueDiffWithBanker = 0
			} else {
				myV = GetThreeCardsHolderInstance().GetCardsValue(int(rb.mySeat))
				rb.exportData.CardsValueDiffWithBanker = myV - GetThreeCardsHolderInstance().GetCardsValue3(int(rb.GameLogic.Dizhu.ChairID))
			}
		}
	}
	//fmt.Println("Seat: ", rb.mySeat, " MyHandCardsValue ", rb.exportData.MyHandCardsValue, " CardsValueDiffWithAnotherMax ", rb.exportData.CardsValueDiffWithAnotherMax, " CardsValueDiffWithBanker ", rb.exportData.CardsValueDiffWithBanker)
	if nil != rb.GameLogic.Dizhu {
		bankerseat := rb.GameLogic.Dizhu.ChairID
		bankerprev := (bankerseat + 2) % 3
		bankernext := (bankerseat + 1) % 3
		rb.exportData.BankerLastCardType = int(rb.GameLogic.LastOutCardType[bankerseat])
		rb.exportData.EarlyBankerLastCardType = int(rb.GameLogic.LastOutCardType[bankerprev])
		rb.exportData.LateBankerLastCardType = int(rb.GameLogic.LastOutCardType[bankernext])
		rb.exportData.BankerPassType = int(rb.GameLogic.LastPassCardType[bankerseat])
		rb.exportData.EarlyBankerPassType = int(rb.GameLogic.LastPassCardType[bankerprev])
		rb.exportData.LateBankerPassType = int(rb.GameLogic.LastPassCardType[bankernext])

		_, nextg := GetMostValueGroup(rb.GameLogic.Chairs[bankernext].Cards)
		_, bankerg := GetMostValueGroup(rb.GameLogic.Dizhu.Cards)
		nextmaxcard := byte(0)
		bankermaxcard := byte(0)
		for _, group := range nextg {
			if group.Type == msg.CardsType_SingleCard {
				if group.Cards[0] > nextmaxcard {
					nextmaxcard = group.Cards[0]
				}
			}
		}

		for _, group := range bankerg {
			if group.Type == msg.CardsType_SingleCard {
				if group.Cards[0] > bankermaxcard {
					bankermaxcard = group.Cards[0]
				}
			}
		}

		rb.exportData.IsBankerNextSingleGreaterThanLeft = nextmaxcard > bankermaxcard

		_, g := GetMostValueGroup(rb.User.Cards)
		rb.exportData.HandCardGroupTypes = rb.exportData.HandCardGroupTypes[:0]
		for _, v := range g {
			rb.exportData.HandCardGroupTypes = append(rb.exportData.HandCardGroupTypes, int(v.Type))
		}

		findChairID := int32(-1)
		if len(rb.GameLogic.Dizhu.Cards) == 2 {
			rb.exportData.IsBankerApair = rb.GameLogic.Dizhu.Cards[0]>>4 == rb.GameLogic.Dizhu.Cards[1]>>4
			if !rb.User.IsDizhu {
				findChairID = bankerseat
			}
		}
		if len(rb.GameLogic.Chairs[bankernext].Cards) == 2 {
			rb.exportData.IsBankerNextApair = rb.GameLogic.Chairs[bankernext].Cards[0]>>4 == rb.GameLogic.Chairs[bankernext].Cards[1]>>4
			if bankernext != rb.mySeat {
				findChairID = bankerseat
			}
		}

		if len(rb.GameLogic.Chairs[bankerprev].Cards) == 2 {
			rb.exportData.IsBankerPrevAPair = rb.GameLogic.Chairs[bankerprev].Cards[0]>>4 == rb.GameLogic.Chairs[bankerprev].Cards[1]>>4
			if bankernext != rb.mySeat {
				findChairID = bankerprev
			}
		}
		rb.exportData.IsMyDoubleGreaterThanRight = false
		rb.exportData.IsMySigleGreaterThanRight = false
		if findChairID != -1 && len(rb.GameLogic.Chairs[findChairID].Cards) <= 2 {
			rb.exportData.MyMinAlertCardCount++
			min := rb.GameLogic.Chairs[findChairID].Cards[0]
			max := min
			if len(rb.GameLogic.Chairs[findChairID].Cards) == 2 {
				max = rb.GameLogic.Chairs[findChairID].Cards[1]
			}

			for _, card := range rb.cardGroups {
				if card.Type == msg.CardsType_Pair {
					if min != max && min>>4 == max>>4 {
						if card.Cards[0] < min {
							rb.exportData.IsMyDoubleGreaterThanRight = true
						}
					}
					continue
				}

				if card.Type != msg.CardsType_SingleCard {
					continue
				}
				if min > card.Cards[0] {
					rb.exportData.MyMinAlertCardCount++
					rb.exportData.IsMySigleGreaterThanRight = true
				} else if max < card.Cards[0] {
					rb.exportData.MyMaxAlertCardCount++
				}
			}
		}
	}

	if len(rb.CheckAllOff()) > 0 {
		rb.exportData.IsHaveOptimalStrategy = true
	} else {
		rb.exportData.IsHaveOptimalStrategy = false
	}

	next := (rb.mySeat + 1) % 3
	_, counts := GetCardCounts(rb.GameLogic.Chairs[next].Cards)
	rb.exportData.IsOtherDoubleKing = (counts[14] + counts[15]) == 2
	_, counts1 := GetCardCounts(rb.GameLogic.Chairs[(next+1)%3].Cards)
	if !rb.exportData.IsOtherDoubleKing {
		rb.exportData.IsOtherDoubleKing = (counts1[14] + counts1[15]) == 2
	}

	rb.exportData.IsOtherRoket = rb.exportData.IsOtherDoubleKing
	for i := 0; i < len(counts); i++ {
		if counts[i] == 4 {
			rb.exportData.OthersBombCount++
		}

		if counts1[i] == 4 {
			rb.exportData.OthersBombCount++
		}
	}
}

func (rb *robot) resetRobotInfo() {
	rb.killRobotTimer()
	rb.GameLogic.Status = int32(msg.GameStatus_GameInitStatus)

	rb.User.Cards = nil
	rb.cardGroups = nil
	//rb.isPassCard = make([]bool, rb.gameCfg.GameCfg.MaxPlayerCnt)
	rb.firstOutSeat = -1
	//rb.GameLogic.CurrentCards.Cards = nil
	//	rb.threeCards = nil
	rb.GameLogic.LeftCards = nil
	rb.round = 0
	//	rb.outCards = make([][]*outData, rb.gameCfg.GameCfg.MaxPlayerCnt)

	rb.exportData = new(ExportData)
}

func (rb *robot) killRobotTimer() {
	//	rb.mgr.RemoveTimer(rb.timerID)
}

func (rb *robot) setRobotTimer(ms int, f func()) {
	rb.killRobotTimer()
	rb.TimerJob, _ = rb.UserInter.AddTimer(time.Duration(ms), f)
	//rb.timerID = rb.mgr.OnceTimer(gameframe.UserData{Obj: rb, UserData: f}, ms)
}

func (rb *robot) checkRunTree() {
	log.Debugf("游戏阶段 %v", rb.GameLogic.Status)
	rb.killRobotTimer()
	//每个游戏阶段，不同处理
	switch rb.GameLogic.Status {
	case int32(msg.GameStatus_RobStatus):
		if rb.GameLogic.CurrentPlayer.ChairID != rb.mySeat {
			return
		}
	case int32(msg.GameStatus_RedoubleStatus):
		{
		}
	case int32(msg.GameStatus_PutCardStatus):
		if rb.GameLogic.CurrentPlayer.ChairID != rb.mySeat {
			return
		}

		if rb.GameLogic.CurrentCards.UserID == rb.UserInter.GetId() {
			//主动出牌且手上只剩一张牌
			if len(rb.User.Cards) == 1 {
				rb.sendOutCardmsgEx(rand.RandInt(300, 600), 1, []byte{rb.User.Cards[0]})
				return
			}
		} else {
			//被动出牌
			if msg.CardsType_Rocket == GetCardsType(rb.GameLogic.CurrentCards.Cards) {
				rb.doPassCardQuick(rand.RandInt(300, 600))
				return
			} else if isCanOut, _ := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, true); !isCanOut {
				rb.doPassCard()
				return
			}
		}
	default:
		return
	}
	//begin := time.Now()
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"机器人手牌 --- before : %v ", rb.User.Cards)
	//更新导出数据
	rb.updateExportData()
	rb.UpDataLeftCards()
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"机器人手牌 --- after : %v ", rb.User.Cards)
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"花费时间 --- before : %v ms", time.Since(begin).Milliseconds())
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"机器人牌值: %d", rb.exportData.HandCardsValue)
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"机器人 exportData: %+v", *rb.exportData)
	//for index, group := range rb.cardGroups {
	//	log.Debugf(rb.robotPlayer.GetDLogPrefix()+" checkRunTree --- index : %d, %v", index, *group)
	//}
	//开始执行行为树

	rb.startRunTree()
	//log.Debugf(rb.robotPlayer.GetDLogPrefix()+"花费时间 --- after : %v ms", time.Since(begin).Milliseconds())
}

func (rb *robot) startRunTree() {
	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("%v", err)
			buf := make([]byte, 2048)
			length := runtime.Stack(buf, false)
			log.Errorf(string(buf[:length]))
		}
	}()

	ErrorCheck(nil != rb.tree, 2, "tree is nil !!!")
	if nil != rb.tree {
		//输入板
		board := b3core.NewBlackboard()
		state := rb.tree.Tick(rb, board)
		ErrorCheck(state == b3.SUCCESS, 2, "behavior tree tick failed !!!")
	}
}

//判断机器人能否安全出此牌型
func (rb *robot) isCanSafeOutCards(cards []byte, isCompare bool) bool {
	length := (len(cards))
	prevseat := (rb.mySeat + 2) % 3
	nextseat := (rb.mySeat + 1) % 3
	otherct1 := len(rb.GameLogic.Chairs[prevseat].Cards)
	otherct2 := len(rb.GameLogic.Chairs[nextseat].Cards)
	var ret = (otherct1 != length && otherct2 != length)
	if ret || !isCompare {
		return ret
	}
	if exist, _ := SearchLargerCardType(cards, rb.GameLogic.LeftCards, false); !exist {
		ret = true
	}
	return ret
}

//敌人,同伴 是否报单, 报双
func (rb *robot) isPlayerAlarm(isPartner bool, count int32) bool {
	prevseat := (rb.GameLogic.Dizhu.ChairID + 2) % 3
	nextseat := (rb.GameLogic.Dizhu.ChairID + 1) % 3
	if isPartner {
		switch rb.mySeat {
		case rb.GameLogic.Dizhu.ChairID:
			return false
		case prevseat:
			return len(rb.GameLogic.Chairs[nextseat].Cards) == int(count)
		case nextseat:
			return len(rb.GameLogic.Chairs[prevseat].Cards) == int(count)
		}
	}
	if rb.mySeat == rb.GameLogic.Dizhu.ChairID {
		return (len(rb.GameLogic.Chairs[prevseat].Cards) == int(count) ||
			len(rb.GameLogic.Chairs[nextseat].Cards) == int(count))
	}
	return len(rb.GameLogic.Chairs[rb.GameLogic.Dizhu.ChairID].Cards) == int(count)
}

//是否是对手，true对手，false队友
func (rb *robot) isOpponent(seat int32) bool {
	if rb.mySeat == seat || rb.mySeat == rb.GameLogic.Dizhu.ChairID || seat == rb.GameLogic.Dizhu.ChairID {
		return false
	}
	return true
}

//某个玩家是否不能大过此出牌
func (rb *robot) isPlayerNotOut(seat int32, isAbs bool) bool {
	outs := rb.GameLogic.CurrentCards.Cards
	if !isAbs && len(rb.GameLogic.Chairs[seat].Cards) < len(outs) {
		return true
	}
	var cards []byte
	if rb.mySeat == seat {
		cards = rb.User.Cards
	} else {
		cards = rb.GameLogic.LeftCards
	}
	if exist, _ := SearchLargerCardType(outs, cards, isAbs); !exist {
		return true
	}
	return false
}

func (rb *robot) doOut(groupType int) bool {
	out := make([]byte, 0)
	ct := msg.CardsType(groupType)
	_, gg := GetMostValueGroup(rb.User.Cards)
	switch ct {
	case msg.CardsType_TripletWithSingle:
		for _, g := range gg {
			if g.Type == msg.CardsType_Triplet || g.Type == msg.CardsType_SingleCard {
				out = append(out, g.Cards...)
			}
		}
		if len(out) != 4 {
			out = out[:0]
			var part3 []byte
			var part2 []byte
			singles := make([]byte, 0)
			for _, g := range gg {
				if g.Type == msg.CardsType_TripletWithPair {
					if g.Cards[0]>>4 == g.Cards[1]>>4 && g.Cards[1]>>4 == g.Cards[2]>>4 {
						part3 = g.Cards[:3]
						part2 = g.Cards[3:]
					} else {
						part3 = g.Cards[2:]
						part2 = g.Cards[:2]
					}
				}
				if g.Type == msg.CardsType_SingleCard {
					singles = append(singles, g.Cards...)
				}
			}
			out = part3 // GetMostValueGroup returned is copied data, so it is ok
			if len(singles) == 0 {
				out = append(out, part2[0])
			} else {
				out = append(out, singles[0])
			}
		}
	case msg.CardsType_SerialTripletWithOne: // AIRPLANEANDmsg.CardsType_Pair ---> AIRPLANEANDmsg.CardsType_SingleCard
		var jet1 []byte
		singles := make([]byte, 0)
		for _, g := range gg {
			if g.Type == msg.CardsType_SerialTripletWithWing {
				jet1 = make([]byte, len(g.Cards))
				copy(jet1, g.Cards)
			}
			if g.Type == msg.CardsType_SingleCard {
				singles = append(singles, g.Cards...)
			}
		}
		if len(jet1) > 0 && len(singles) > 1 {
			jet1 = Sort111(jet1, msg.CardsType_SerialTripletWithWing)
			doubleN := len(jet1) / 5
			doubleParts := jet1[len(jet1)-doubleN*2:]
			tripleParts := jet1[:doubleN*3]
			out = append(out, tripleParts...)
			if len(singles) >= doubleN {
				out = append(out, singles[:doubleN]...)
			} else {
				out = append(out, singles...)
				doubleParts = doubleParts[:len(singles)*2]
				out = append(out, doubleParts...)
			}
		}
	}
	fmt.Println("aaaa", out, " ", groupType)
	rb.sendOutCardmsg(1, out)
	return true
}

func (rb *robot) DealGameStatus(buffer []byte) {
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
		for index, score := range rb.robotCfg.AddScorePlace {
			if index == 0 && rb.RobScore <= score {
				addLevel = 0
				break
			}

			if index == len(rb.robotCfg.RobScorePlace)-1 && rb.RobScore >= score {
				addLevel = len(rb.robotCfg.RobScorePlace) - 1
				break
			}

			if rb.RobScore == score {
				addLevel = index
			}
		}

		// 权重值
		weightsValue := rand.RandInt(1, 101)
		log.Tracef("机器人加倍权重：%d", weightsValue)

		// 计算加倍权重
		for k, v := range rb.robotCfg.AddRatePlace[addLevel] {
			downLimit := acc
			upLimit := acc + v

			acc = upLimit
			if weightsValue > downLimit && weightsValue <= upLimit {
				addIndex = k
				break
			}
		}

		addMultiple = rb.GameLogic.GameCfg.AddMultiple[addIndex]

		req := msg.RedoubleReq{
			AddNum: addMultiple,
		}

		// 随机时间
		delayTime := rand.RandInt(1000, 5001)

		// 延迟发送消息
		rb.TimerJob, _ = rb.UserInter.AddTimer(time.Duration(delayTime), func() {
			// 请求server加倍
			err := rb.UserInter.SendMsgToServer(int32(msg.ReceiveMessageType_C2SRedouble), &req)
			if err != nil {
				log.Errorf("send server msg fail: %v", err.Error())
			}
		})

	}

}

// DealGameStatus 处理当前抢庄玩家消息
func (robot *robot) DealCurrentRobber(buffer []byte) {
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
	for index, score := range robot.robotCfg.RobScorePlace {
		if index == 0 && robot.RobScore <= score {
			robLevel = 0
			break
		}

		if index == len(robot.robotCfg.RobScorePlace)-1 && robot.RobScore >= score {
			robLevel = len(robot.robotCfg.RobScorePlace) - 1
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
	for k, v := range robot.robotCfg.RobRatePlace[robLevel] {
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

// 设置抢庄分值
func (robot *robot) SetRobScore(cards []byte) {
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

//刷新剩余牌
func (rb *robot) UpDataLeftCards() {
	rb.deskLeftCards = make([]byte, 0)
	for _, v := range rb.GameLogic.LeftCards {
		bFind := false
		for _, uv := range rb.User.Cards {
			if uv == v {
				bFind = true
				break
			}
		}

		if !bFind {
			rb.deskLeftCards = append(rb.deskLeftCards, v)
		}
	}
}

func (robot *robot) CheckAllOff() (cardsList [][]byte) {
	// 能被接管索引集合
	var SolutionIndexs []int
	robot.BestSolution = poker.GetBestSolutions(robot.User.Cards)
	for _, user := range robot.GameLogic.UserList {
		if (robot.User.Role == data.RoleDizhu && robot.User.ID == user.ID) ||
			(robot.User.Role != data.RoleDizhu && !user.IsDizhu) {
			continue
		}

		for index, v := range robot.BestSolution {
			handCards := poker.HandCards{
				Cards:       v.Cards,
				WeightValue: poker.GetCardsWeightValue(v.Cards, v.CardsType),
				CardsType:   int32(v.CardsType),
			}

			if len(poker.TakeOverCards(handCards, user.Cards)) > 0 {
				SolutionIndexs = append(SolutionIndexs, index)
			}
		}
	}

	// 只有一手牌能被接管，说明可以实现春天打发
	if len(SolutionIndexs) <= 1 {

		for index, v := range robot.BestSolution {
			if len(SolutionIndexs) > 0 && SolutionIndexs[0] == index {
				continue
			}
			cardsList = append(cardsList, v.Cards)
		}

		if len(SolutionIndexs) > 0 {
			cardsList = append(cardsList, robot.BestSolution[0].Cards)
		}

	}

	return
}
