package game

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

// Start 游戏开始，进入动画状态
func (game *Blackjack) Start() {

	if game.table.GetRoomID() < 0 {
		log.Debugf("房间ID为空")
		return
	}

	game.Status = int32(msg.GameStatus_StartMove)

	// 初始化牌，洗牌
	var gamePoker poker.GamePoker
	gamePoker.InitPoker()
	game.Poker = &gamePoker
	game.CurActionUser = nil

	// 匹配机器人
	game.MatchRobot()

	// 定时进入开始下注
	log.Tracef("定时进入开始下注")
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.StartMove), game.StartBet)

	// 广播开始游戏
	log.Tracef("广播开始游戏 %d", game.table.GetID())
	game.SendGameStatus(game.Status, int32(game.timeCfg.StartMove/1000), nil)
}

// StartBet 开始下注，进入下注状态
func (game *Blackjack) StartBet() {
	game.Status = int32(msg.GameStatus_BetStatus)
	game.table.StartGame()
	log.Tracef("游戏 %d 进入下注阶段", game.table.GetID())

	// 下注开始动画时间
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.BetAnimation), func() {
		// 广播发送场景消息
		game.SendSceneInfo(nil, false)

		game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.MsgDelay), func() {
			game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.BetStatus), game.EndBet)
			log.Tracef("EndBet 结束下注 定时消息 %v", game.TimerJob)

			// 发送开始下注消息
			game.SendGameStatus(game.Status, int32(game.timeCfg.BetStatus/1000), nil)
		})

	})

}

// EndBet 结束下注
func (game *Blackjack) EndBet() {
	game.Status = int32(msg.GameStatus_EndBet)
	log.Tracef("游戏 %d 下注结束了", game.table.GetID())

	userListSeats := make([]int64, 5)

	for k, user := range game.AllUserList {

		// 将未下注的玩家 记数 +
		if user.Status == int32(msg.UserStatus_UserReady) {

			user.Status = int32(msg.UserStatus_UserWatch)

			// 全局user 数据
			userData := data.GetUserInterdata(user.User)

			userData.NotBetCount++

			data.SetUserInterdata(user.User, userData)

		}

		game.AllUserList[k] = user

		// 已经下注，转为下注完毕
		if user.Status == int32(msg.UserStatus_UserBetSuccess) {
			game.UserConfirmBet([]byte{}, user.User)
		}

		userListSeats[user.ChairID] = user.ID
	}

	// 下注玩家序列
	var betSeats []int64
	for _, userID := range userListSeats {
		if _, ok := game.UserList[userID]; ok {
			betSeats = append(betSeats, userID)
		}
	}

	game.BetSeats = betSeats

	// 广播发送场景消息
	game.SendSceneInfo(nil, false)

	// 无人押注直接进入结算阶段
	if len(game.UserList) < 1 {
		game.Settle()
		return
	}

	// 下注结束到后第一轮发牌
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.EndBetStatus), game.FirstDealCards)

	// 广播下注结束状态
	game.SendGameStatus(game.Status, int32(game.timeCfg.EndBetStatus/1000), nil)
}

// FirstDealCards 第一轮发牌
func (game *Blackjack) FirstDealCards() {
	game.Status = int32(msg.GameStatus_FirstRoundFaPai)
	log.Tracef("游戏 %d 第一轮发牌开始", game.table.GetID())

	// 牌组;第一组牌出参
	var firstCardsList []*msg.UserPaiInfoRes

	var hasBlackJack bool

	// 给下注用户发牌
	for k, user := range game.UserList {

		//game.Poker.CheatCheck(user.TestCardsType, []byte{})

		game.Control([]byte{}, user)

		// 用户手牌
		cards := game.Poker.DrawTwoCards()

		// 所有人拿对子
		//var card1 byte
		//card1 = 0xe1
		//cardValue, _ := poker.GetCardValueAndColor(card1)
		//game.Poker.PlugSelectedCard(cardValue)
		//card2 := game.Poker.DrawCard()
		//cards = []byte{card1, card2}

		cardsType := poker.GetCardsType(cards)

		// 用户持有的第一副手牌
		holdCards := &data.HoldCards{
			Cards:        cards,
			Point:        poker.GetPoint(cards),
			Type:         cardsType,
			BetAmount:    user.BetAmount,
			ActionPermit: true,
		}

		// 牌型为黑杰克，关闭牌组操作权限
		if cardsType == msg.CardsType_BlackJack {
			hasBlackJack = true
			holdCards.ActionPermit = false
			holdCards.EndType = data.EndType_Blackjack
		}
		user.HoldCards[0] = holdCards

		// 分牌因子
		card, ok := poker.IsPair(cards)
		if ok {
			user.DepartFactor = card
		}

		game.AllUserList[k] = user
		game.UserList[k] = user

		// 第一组牌出参
		firstCardsRes := msg.UserPaiInfoRes{
			ChairId:    user.ChairID,
			UserId:     user.ID,
			CardArr:    holdCards.Cards,
			CardType:   int32(holdCards.Type),
			Cardspoint: poker.ReducePoint(holdCards.Point),
		}

		firstCardsList = append(firstCardsList, &firstCardsRes)
	}

	var firstCard, secondCard, firstValue, secondValue byte

	// 庄家手牌
	hostCards := HostCards{}

	// 第一张牌
	firstCard = game.Poker.DrawCard()

	// 保险测试
	//firstCard = 0xe1

	// 第一张牌牌值
	firstValue, _ = poker.GetCardValueAndColor(firstCard)

	// 第一张牌放入明牌
	hostCards.Cards = append(hostCards.Cards, firstCard)

	// 抽到重复牌放弃
	var loopCount int
	for {
		secondCard = game.Poker.DrawCard()
		secondValue, _ = poker.GetCardValueAndColor(secondCard)

		if firstValue != secondValue {
			break
		}

		// 测试庄家必定是黑杰克
		//if secondValue >= 10 && firstValue != secondValue {
		//	break
		//}

		// 防止死循环
		if loopCount >= 16 {
			break
		}
		loopCount++
	}

	// 第二张牌如果为A，放入明牌，第一张牌放入暗牌
	if secondValue == poker.Acard {
		hostCards.PocketCard = firstCard
		hostCards.Cards[0] = secondCard
	} else {
		hostCards.PocketCard = secondCard
	}

	// 庄家牌值和牌型
	cards := []byte{firstCard, secondCard}
	hostCards.Point = poker.GetPoint(hostCards.Cards)
	hostCards.Type = poker.GetCardsType(cards)

	// 更新庄家手牌，以及游戏牌组信息
	game.HostCards = &hostCards

	resp := msg.FaPaiRes{
		Cards:          hostCards.Cards,
		UserPaiInfoArr: firstCardsList,
	}

	// 广播第一次发牌信息
	game.SendFirstDeal(resp)

	// 动画时间根据人数调整
	timeC := len(game.UserList) * game.timeCfg.FirstMove

	// 玩家是黑杰克提前结算
	if hasBlackJack {
		// 加上黑杰克动画结算时间
		timeC += game.timeCfg.SettleAnimation
	}

	game.TimerJob, _ = game.table.AddTimer(int64(timeC), game.FirstDealCardsEnd)

	// 广播第一轮发牌阶段
	game.SendGameStatus(game.Status, int32(timeC/1000), nil)
}

// FirstDealCardsEnd 第一轮发牌阶段结束
func (game *Blackjack) FirstDealCardsEnd() {
	log.Tracef("游戏 %d 第一轮发牌结束了", game.table.GetID())
	var hasBlackJackCount int

	// 结算回参
	var userResultsList []*msg.UserResultsRes
	var records []*platform.PlayerRecord
	for userID, user := range game.UserList {

		// 是否是黑杰克
		if user.HoldCards[0].Type == msg.CardsType_BlackJack {
			hasBlackJackCount++

			// 赢利
			profit := user.BetAmount * 3 / 2

			// 变更筹码
			game.UserList[userID].CurAmount += user.BetAmount + profit
			netProfit := game.SettleDivision(userID)

			// 发送打码量
			game.SetChip(userID, user.Chip)

			// 发送战绩，计算产出
			if !user.User.IsRobot() {
				if record := game.TableSendRecord(userID, profit, netProfit); record != nil {
					record.Balance = user.User.GetScore()
					records = append(records, record)
				}
			}

			// 注池清空
			game.UserList[userID].BetAmount -= user.BetAmount

			// 编辑用户日志
			game.WriteUserLog(user, netProfit)

			game.PaoMaDeng(netProfit, user.User, "黑杰克")

			// 停止操作状态
			user.Status = int32(msg.UserStatus_UserStopAction)

			game.AllUserList[userID] = user
			game.UserList[userID] = user

			userResults := msg.UserResultsRes{
				UserId:      user.ID,
				UserWinLoss: netProfit,
				ChairId:     user.ChairID,
				Status:      user.Status,
			}
			userResultsList = append(userResultsList, &userResults)
		}
	}

	if len(records) > 0 {
		if _, err := game.table.UploadPlayerRecord(records); err != nil {
			log.Errorf("upload player record error %s", err.Error())
		}
	}

	// 闲家有黑杰克，
	if hasBlackJackCount > 0 {

		// 广播结算结果
		game.SendSettleInfo(userResultsList)

		// 飘筹码动画时间
		duration := int64(game.timeCfg.AdvanceSettle)

		// 闲家都是黑杰克，提前结算
		if hasBlackJackCount == len(game.UserList) {
			game.TimerJob, _ = game.table.AddTimer(duration, game.Settle)
			return
		}

		// 庄家明牌为A，进入保险阶段
		if (game.HostCards.Cards[0]&0xff)>>4 == poker.Acard {
			game.TimerJob, _ = game.table.AddTimer(duration, game.Insurance)
		} else {
			// 进入第二轮发牌阶段
			game.TimerJob, _ = game.table.AddTimer(duration, game.SecondDealCards)
		}

	} else {
		//庄家明牌为A，进入保险阶段
		if (game.HostCards.Cards[0]&0xff)>>4 == poker.Acard {
			game.Insurance()
		} else {
			// 进入第二轮发牌阶段
			game.SecondDealCards()
		}
	}

}

// Insurance 进入保险阶段
func (game *Blackjack) Insurance() {
	game.Status = int32(msg.GameStatus_InsuranceStatus)
	log.Tracef("游戏 %d 进入买保险", game.table.GetID())

	// 定时进入保险结束
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.InsuranceStatus), game.InsuranceEnd)

	// 广播买保险阶段
	game.SendGameStatus(game.Status, int32(game.timeCfg.InsuranceStatus/1000), nil)
	return
}

// InsuranceEnd 保险阶段结束
func (game *Blackjack) InsuranceEnd() {
	game.Status = int32(msg.GameStatus_EndInsurance)
	log.Tracef("游戏 %d 买保险结束", game.table.GetID())

	var listRes []*msg.InsureResultRes

	// 通知保险结束状态
	game.SendGameStatus(game.Status, 0, nil)

	for _, user := range game.UserList {

		if user.IsBuyInsure {

			// 保险结果回参数
			insureResultRes := msg.InsureResultRes{
				UserId:      user.ID,
				ChairId:     user.ChairID,
				UserWinLoss: -user.Insurance,
			}
			listRes = append(listRes, &insureResultRes)

		}
	}

	// 庄家为黑杰克,翻开暗牌
	if game.HostCards.Type == msg.CardsType_BlackJack {
		game.OpenPocketCard()
	}

	// 飘保险结果动画时间
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.EndInsuranceSettle), func() {

		// 广播保险结果
		game.SendInsureResult(listRes)

		// 保险结束，隔一秒进入下一状态
		game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.EndInsurance), func() {

			if game.HostCards.Type == msg.CardsType_BlackJack {

				// 游戏进入结算阶段
				game.Settle()
			} else {

				// 进入第二轮发牌阶段
				game.SecondDealCards()
			}
		})
	})

}

// SecondDealCards 第二轮发牌阶段
func (game *Blackjack) SecondDealCards() {
	game.Status = int32(msg.GameStatus_SecondRoundFaPai)
	log.Tracef("游戏 %d 第二轮发牌开始了", game.table.GetID())

	// 发送第二轮发牌阶段信息
	game.SendGameStatus(game.Status, 0, nil)

	// 玩家轮换操作
	game.TurnUserAction()
}

// TurnUserAction 轮转玩家操作
func (game *Blackjack) TurnUserAction() {
	// 当前操作玩家
	curUser := game.CurActionUser

	if curUser != nil && curUser.UserID == 0 {
		log.Errorf("curUser出现了错误的id， curUser: %v", curUser)
		curUser = nil
	}

	log.Debugf("当前操作玩家：%v", curUser)

	// 当前操作玩家不为空
	if curUser != nil {

		// 所有玩家操作完，庄家操作
		if curUser.ChairID == -1 {

			game.HostAction(true)

			return
		}

		// 轮转计数器不一致：用户未在时限内操作，默认停牌
		if game.TurnCounter != curUser.TurnCounter {
			log.Tracef("轮转计数器不一致，玩家 %d 默认执行停牌", curUser.UserID)

			AskDoReq := &msg.AskDoReq{
				CmdType:       int32(msg.AskDoType_Stand),
				BetCardsIndex: curUser.BetCardsIndex,
			}

			buffer, err := proto.Marshal(AskDoReq)
			if err != nil {
				log.Errorf("proto marshal fail :", err.Error())
				return
			}
			game.UserDoCmd(buffer, game.UserList[curUser.UserID].User)
			return
		}

		// 游戏轮转计数器+1
		game.TurnCounter++

	} else {
		for _, id := range game.BetSeats {
			user, ok := game.UserList[id]
			if !ok {
				log.Warnf("查询玩家 %d 异常", id)
			}

			// 未结算的玩家
			if user.Status != int32(msg.UserStatus_UserStopAction) {
				curUser = &CurUser{
					ChairID:       user.ChairID,
					UserID:        user.ID,
					BetCardsIndex: 0,
					GetPoker:      user.CheckGetPoker(0),
					DepartPoker:   user.CheckDepartPoker(0),
					DoubleBet:     user.CheckDoubleBet(0),
					Stand:         user.CheckStand(0),
					GiveUp:        user.CheckGiveUp(0),
					TurnCounter:   0,
				}
				break
			}
		}

		// 所有玩家都已经提前结算,庄家操作
		if curUser == nil {
			game.HostAction(true)
			return
		}
		game.TurnCounter = 1

	}

	// 打开当前玩家可操作开关
	curUser.ActionSwitch = true

	game.CurActionUser = curUser

	// 定时执行当前玩家检测
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.UserAction), game.TurnUserAction)

	// 广播当前操作玩家
	game.SendCurrentSeat(*curUser)
}

// HostAction 庄家操作
func (game *Blackjack) HostAction(actionPermit bool) {
	// 结算玩家个数
	var settlePlayerCount int
	for _, user := range game.UserList {
		if user.Status == int32(msg.UserStatus_UserStopAction) {
			settlePlayerCount++
		}
	}

	// 所有玩家都结算完，直接进入结算
	if settlePlayerCount == len(game.UserList) {
		game.Settle()
		return
	}
	// 庄家手牌
	hostCards := game.HostCards

	// 暗牌未翻开
	if hostCards.PocketCard != 0 {

		// 翻开暗牌
		game.OpenPocketCard()

		// 牌值在 17 ～ 21 之间，停止要牌
		clearPoint := poker.GetNearPoint21(hostCards.Point)
		if clearPoint >= poker.Point17 && clearPoint <= poker.Point21 {
			actionPermit = false
		}

		// 定时轮训执行庄家操作
		game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.HostAction), func() {
			game.HostAction(actionPermit)
		})
		return
	}

	// 庄家操作权限
	if actionPermit {

		// 庄家控牌
		game.BankerControl()

		// 庄家要牌
		card := game.Poker.DrawCard()
		hostCards.Cards = append(hostCards.Cards, card)
		hostCards.Point = poker.GetPoint(hostCards.Cards)
		hostCards.Type = poker.GetCardsType(hostCards.Cards)

		// 牌值在 17 ～ 21 之间，停止要牌
		clearPoint := poker.GetNearPoint21(hostCards.Point)
		if clearPoint >= poker.Point17 && clearPoint <= poker.Point21 {
			actionPermit = false
		}

		// 非普通牌，停止要牌
		if hostCards.Type != msg.CardsType_Other {
			actionPermit = false
		}

		// 手牌不能超过5张
		if len(hostCards.Cards) >= 5 {
			actionPermit = false
		}

		game.HostCards = hostCards

		faPaiOneRes := msg.FaPaiOneRes{
			ChairId:    -1,
			Cards:      []byte{card},
			CardType:   int32(hostCards.Type),
			Cardspoint: poker.ReducePoint(hostCards.Point),
		}

		// 广播发一张牌的信息
		game.SendDealCard(faPaiOneRes)

		game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.HostAction), func() {
			game.HostAction(actionPermit)
		})

	} else {
		duration := int64(game.timeCfg.HostAction)
		if game.HostCards.Type != msg.CardsType_Other {
			duration = int64(game.timeCfg.SettleAnimation)
		}

		// 庄家操作完成，进入结算
		game.TimerJob, _ = game.table.AddTimer(duration, game.Settle)
	}

}

// Settle 结算阶段
func (game *Blackjack) Settle() {
	game.Status = int32(msg.GameStatus_Result)
	log.Tracef("游戏 %d 进入结算阶段", game.table.GetID())

	// 广播结算状态
	game.SendGameStatus(game.Status, int32(game.timeCfg.SettleStatus), nil)

	// 结算回参
	var userResultsList []*msg.UserResultsRes
	var records []*platform.PlayerRecord
	for userID, user := range game.UserList {

		// 跳过已经结算的玩家
		if user.Status == int32(msg.UserStatus_UserStopAction) {
			continue
		}

		// 单人总结算,盈利,净盈利
		var allResult, netProfit int64

		for index, holdCards := range user.HoldCards {

			// 没有手牌跳过
			if len(holdCards.Cards) == 0 {
				continue
			}
			result := game.ContrastCards(*holdCards, user.IsBuyInsure)

			if result > 0 {
				game.UserList[userID].HoldCards[index].EndType = data.EndType_BankerLoss
			} else if result < 0 {
				game.UserList[userID].HoldCards[index].EndType = data.EndType_BankerWin
			} else {
				game.UserList[userID].HoldCards[index].EndType = data.EndType_draw
			}

			allResult += result
		}

		game.UserList[userID].CurAmount += user.BetAmount + allResult
		netProfit = game.SettleDivision(userID)

		// 发送打码量
		game.SetChip(userID, user.Chip)

		// 发送战绩，计算产出
		if !user.User.IsRobot() {
			if record := game.TableSendRecord(userID, allResult, netProfit); record != nil {
				record.Balance = user.User.GetScore()
				records = append(records, record)
			}
		}

		// 编辑用户日志
		game.WriteUserLog(user, netProfit)

		game.PaoMaDeng(netProfit, user.User, "")

		user.Status = int32(msg.UserStatus_UserStopAction)

		game.AllUserList[userID] = user
		game.UserList[userID] = user

		userResultsList = append(userResultsList, &msg.UserResultsRes{
			UserId:      userID,
			UserWinLoss: netProfit,
			Status:      user.Status,
			ChairId:     user.ChairID,
		})

	}

	//上传战绩
	if len(records) > 0 {
		if _, err := game.table.UploadPlayerRecord(records); err != nil {
			log.Errorf("upload player record error %s", err.Error())
		}
	}

	// 广播结算信息
	game.SendSettleInfo(userResultsList)

	// 广播场景消息
	game.SendSceneInfo(nil, false)

	// 庄家有手牌几率庄家手牌
	if game.HostCards != nil {
		bankerCards := poker.CardsToString(game.HostCards.Cards)
		bankerCardsType := poker.CardsTypeToString(game.HostCards.Type)
		game.table.WriteLogs(0, "庄家手牌："+bankerCards+" 牌型： "+bankerCardsType+" 点数： "+fmt.Sprintf(`%d`, game.HostCards.Point))
	}

	// 通知框架游戏结束
	game.table.EndGame()

	// 游戏结束
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.SettleStatus), game.GameOver)
}

// GameOver 游戏结束状态
func (game *Blackjack) GameOver() {
	log.Tracef("游戏结束")

	game.Status = int32(msg.GameStatus_GameOver)

	// 广播结束状态
	game.SendGameStatus(game.Status, 0, nil)

	// 重置桌面属性
	game.BetSeats = nil
	game.CurActionUser = nil
	game.HostCards = nil
	game.TurnCounter = 0
	game.LoadCfg = false

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}

	// 踢出所有人
	for _, user := range game.AllUserList {
		delete(game.AllUserList, user.ID)
		delete(game.UserList, user.ID)

		// 让出座位
		for chairID, userID := range game.Chairs {
			if user.ID == userID {
				game.Chairs[chairID] = 0
				break
			}
		}

		// 移除押注序列
		for k, userID := range game.BetSeats {
			if user.ID == userID {
				game.BetSeats = append(game.BetSeats[:k], game.BetSeats[k+1:]...)
				break
			}
		}

		// 广播玩家离开信息
		res := msg.UserLeaveRoomRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CUserLeaveRoom), &res)

		//踢出用户
		game.table.KickOut(user.User)
	}

	// 桌子状态设为等待开始
	game.Status = int32(msg.GameStatus_StartStatus)
	if game.TimerJob != nil {
		game.table.DeleteJob(game.TimerJob)
	}

	//关闭桌子
	log.Tracef("游戏结束关闭桌子")
	game.table.Close()

}
