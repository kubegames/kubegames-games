package game

import (
	"fmt"

	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

// Start 游戏开始，进入动画状态
func (game *GeneralNiuniu) Start() {
	game.Status = int32(msg.GameStatus_StartMove)
	log.Tracef("游戏 %d 倒计时开始", game.Table.GetID())

	if game.TimerJob != nil {
		game.Table.DeleteJob(game.TimerJob)
	}

	game.MatchRobot()

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StartMove), game.StartMoveEnd)

	// 广播游戏开始状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.StartMove/1000), nil)
}

// StartMoveEnd 结束倒计时动画
func (game *GeneralNiuniu) StartMoveEnd() {
	log.Tracef("enter start")
	if len(game.UserList) < 2 {

		// 重新进入匹配
		game.Status = int32(msg.GameStatus_ReadyStatus)
		log.Tracef("游戏 %d 人数不满足开始，回到匹配状态", game.Table.GetID())

		// 检测剩余机器人
		game.CheckLeftRobot()

		// 广播状态变更为匹配中
		game.SendGameStatus(game.Status, 0, nil)
	} else {
		// 重新进入匹配
		game.Status = int32(msg.GameStatus_DealCards)
		log.Tracef("游戏 %d 进入发牌阶段，开始游戏", game.Table.GetID())

		// 更改用户退出权限
		game.SetExitPermit(false)

		// 通知框架开赛
		game.Table.StartGame()

		// 开始游戏动画后开始投注
		game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StartAnimation), game.FirstDealCards)

	}

}

// FirstDealCards 第一次发牌
func (game *GeneralNiuniu) FirstDealCards() {
	// 洗牌
	game.Poker = new(poker.GamePoker)
	game.Poker.InitPoker()

	// 控牌
	game.Control()

	// 发脾
	for id, user := range game.UserList {

		cards := game.ControlledCards[id].Cards
		cardsType := game.ControlledCards[id].CardsType

		// 有配牌，优先取配牌
		if len(user.HoldCards.Cards) > 0 {
			cards = user.HoldCards.Cards
			cardsType = poker.GetCardsType(cards)
		}

		user.HoldCards = &poker.HoldCards{
			Cards:             cards,
			CardsType:         cardsType,
			FirstHalfCards:    cards[:3],
			LowerHalfCards:    cards[3:],
			SpecialCardIndexs: poker.GetSpecialCardIndexs(cards, cardsType),
		}

		// 更新数据
		game.UserList[id] = user

		// 发送发牌信息
		dealResult := msg.DealRes{
			Cards:   user.HoldCards.FirstHalfCards,
			UserId:  id,
			ChairId: user.ChairID,
		}
		game.SendDealCardsMsg(dealResult, user.User)
	}

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.DealAnimation*len(game.UserList)), game.StartBet)
}

// StartBet 开始投注
func (game *GeneralNiuniu) StartBet() {

	game.Status = int32(msg.GameStatus_BetChips)
	log.Tracef("游戏 %d 进入投注阶段", game.Table.GetID())

	// 广播投注开始状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.BetChips/1000), nil)

	// 计算玩家可投注倍数
	for id, user := range game.UserList {

		// 最高可投注倍数 = 玩家携带金额/玩家人数/牌的最高倍数（倍数按3倍算）/底注
		highestBet := user.CurAmount / int64(len(game.UserList)-1) / 3 / game.RoomCfg.RoomCost

		// 最高可投注倍数不能超过4倍
		if highestBet > 4 {
			highestBet = 4
		}

		// 最低可投1倍
		if highestBet == 0 {
			highestBet = 1
		}

		game.UserList[id].HighestMultiple = highestBet

		betMultipleResp := msg.BetMultipleRes{
			UserId:          id,
			ChairId:         user.ChairID,
			HighestMultiple: highestBet,
		}

		// 发送投注倍率信息
		game.SendBetMultipleInfo(betMultipleResp, user.User)
	}

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.BetChips), game.EndBet)

}

// EndBet 投注结束
func (game *GeneralNiuniu) EndBet() {
	game.Status = int32(msg.GameStatus_EndBet)
	log.Tracef("游戏 %d 进入投注结束阶段", game.Table.GetID())

	// 未投注的人默认投最低倍数
	for id, user := range game.UserList {

		if user.BetMultiple == 0 {

			// 更新数据
			user.BetMultiple = 1
			game.UserList[id] = user

			// 广播投注信息
			betChipsInfo := msg.BetChipsInfoRes{
				BetMultiple: user.BetMultiple,
				UserId:      id,
				ChairId:     user.ChairID,
			}
			game.SendBetChipsInfo(betChipsInfo)
		}
	}

	game.SendGameStatus(game.Status, 0, nil)

	game.DealCards()
}

// DealCards 发牌
func (game *GeneralNiuniu) DealCards() {

	// 发牌动画时间
	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.SecondDealAnimation), game.StartShow)

	for id, user := range game.UserList {

		// 发送发牌信息
		dealResult := msg.DealRes{
			Cards:   user.HoldCards.LowerHalfCards,
			UserId:  id,
			ChairId: user.ChairID,
		}
		game.SendDealCardsMsg(dealResult, user.User)
	}
}

// StartShow 开始摊牌
func (game *GeneralNiuniu) StartShow() {
	game.Status = int32(msg.GameStatus_ShowCards)
	log.Tracef("游戏 %d 进入摊牌阶段", game.Table.GetID())

	// 广播开始摊牌状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.ShowCards/1000), nil)

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.ShowCards), game.EndShow)

}

// EndShow 摊牌结束
func (game *GeneralNiuniu) EndShow() {
	game.Status = int32(msg.GameStatus_EndShow)
	log.Tracef("游戏 %d 进入摊牌结束阶段", game.Table.GetID())

	// 留点时间间隔进入结算阶段
	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.CheckCardsType), game.Settle)

	// 未摊牌的人默认摊牌
	for _, user := range game.UserList {
		if user.Status != int32(msg.UserStatus_ShowedCards) {

			// 广播摊牌结果
			showCardsResult := msg.ShowCardsRes{
				Cards:       user.HoldCards.Cards,
				UserId:      user.ID,
				ChairId:     user.ChairID,
				CardsType:   int32(user.HoldCards.CardsType),
				CardsIndexs: user.HoldCards.SpecialCardIndexs,
			}
			game.SendShowCardsMsg(showCardsResult)
		}
	}
}

// Settle 结算(赢家通吃原则)
func (game *GeneralNiuniu) Settle() {
	game.Status = int32(msg.GameStatus_SettleStatus)
	log.Tracef("游戏 %d 进入结算阶段", game.Table.GetID())

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.Settle), game.GameOver)

	// 广播结算状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.Settle/1000), nil)

	// 闲家输家列表
	loserList := make(map[int64]*SettleResult)

	var (
		resultList    []*msg.SettleResult // 结算结果列表
		cardsMultiple int64               // 牌倍数
		winnerResult  int64               // 赢家结算结果
		theoryLoseSum int64               // 闲家理论输钱合值
		winner        *data.User          // 赢家
	)

	// 确定赢家
	for _, user := range game.UserList {
		// 假定一个赢家
		winner = user
		for _, waitUser := range game.UserList {
			if winner.ID == waitUser.ID {
				continue
			}
			if poker.ContrastCards(winner.HoldCards, waitUser.HoldCards) {
				winner = waitUser
			}
		}
		break
	}

	if winner == nil {
		log.Errorf("没有找到赢家")
		return
	}

	// 赢家牌倍数
	cardsMultiple = poker.GetCardsMultiple(winner.HoldCards.CardsType)

	// 初步结算
	for id, user := range game.UserList {
		if id == winner.ID {
			continue
		}

		// 结算额度 = 房间底分 * 赢家下注倍数 * 赢家牌型对应的倍数 * 输家下注倍数
		result := game.RoomCfg.RoomCost * game.UserList[winner.ID].BetMultiple * cardsMultiple * user.BetMultiple * -1

		loserList[id] = &SettleResult{
			TheorySettle: result,
			ActualSettle: result,
			CurAmount:    user.CurAmount,
		}
		theoryLoseSum += result

		// 输家防止不够输
		if -result > user.CurAmount {
			result = -user.CurAmount
			loserList[id].ActualSettle = -user.CurAmount
		}

		winnerResult += -result
	}

	// 赢钱大于携带本金，防止以小博大, 输钱玩家等比例减少输钱金额
	if winnerResult > winner.CurAmount {
		var (
			loseCounter int   // 输家计数器
			loseAcc     int64 // 输钱累加器
		)

		winnerResult = winner.CurAmount

		for userID, v := range loserList {

			// 最后一个赢钱玩家
			if len(loserList)-loseCounter == 1 {
				loserList[userID].ActualSettle = (winnerResult + loseAcc) * -1
				break
			}

			// 缩减比例后新的结算值 = 应赢合值 * 玩家理论赢钱值 / 理论赢钱合值
			loserList[userID].ActualSettle = winnerResult * v.TheorySettle / theoryLoseSum * -1
			loseAcc += loserList[userID].ActualSettle
			loseCounter++
		}

		var (
			leftAmount        int64 // 剩余多输金额
			leftTheoryLoseSum int64 // 剩余理论输钱合值
		)

		for userID, v := range loserList {

			// 缩减比例后 应输钱大于携带金额，应输钱减为携带金额，
			if v.ActualSettle <= -v.CurAmount {
				leftAmount += v.ActualSettle + v.CurAmount
				loserList[userID].ActualSettle = -v.CurAmount
			} else {
				leftTheoryLoseSum += v.TheorySettle
			}
		}

		// 有剩余多输金额，重新分配一下剩余金额, 剩余多输钱 按照 玩家理论输钱值 / 剩余玩家理论输钱合值 比例分下去
		if leftAmount < 0 {
			FillLoserAmount(&leftAmount, &leftTheoryLoseSum, loserList)
			ConvertLoserAmount(leftAmount, leftTheoryLoseSum, loserList)
		}
	}

	// 输家加入结算列表
	for id, v := range loserList {
		resultList = append(resultList, &msg.SettleResult{
			UserId:  id,
			ChairId: game.UserList[id].ChairID,
			Result:  v.ActualSettle,
		})
	}

	// 赢家加入结算列表
	resultList = append(resultList, &msg.SettleResult{
		UserId:  winner.ID,
		ChairId: winner.ChairID,
		Result:  winnerResult,
	})

	//战绩
	var records []*platform.PlayerRecord

	// 上下分
	for i, v := range resultList {
		netProfit := game.UserList[v.UserId].User.SetScore(game.Table.GetGameNum(), v.Result, game.Table.GetRoomRate())

		// 计算打码量,触发跑马灯
		chip := netProfit
		if netProfit > 0 {
			chip = game.RoomCfg.RoomCost

			// 跑马灯触发
			specialWords := ""
			if winner.HoldCards.CardsType > msg.CardsType_NiuNine {
				specialWords = poker.TransformCardsType(winner.HoldCards.CardsType)
			}
			game.PaoMaDeng(netProfit, winner.User, specialWords)
		}

		// 发送打码量
		game.SetChip(v.UserId, chip)

		// 更新玩家当前金额
		game.UserList[v.UserId].CurAmount += netProfit

		// 发送战绩，计算产出
		if record := game.TableSendRecord(v.UserId, resultList[i].Result, netProfit); record != nil {
			records = append(records, record)
		}

		resultList[i].Result = netProfit
	}

	//发送战绩
	if len(records) > 0 {
		if _, err := game.Table.UploadPlayerRecord(records); err != nil {
			log.Warnf("upload player record error %s", err.Error())
		}
	}

	// 记录游戏日志
	for _, settleResult := range resultList {
		user := game.UserList[settleResult.UserId]

		// 获取手牌字符串
		cardsStr := poker.CardsToString(user.HoldCards.Cards)

		// 作弊率来源
		probSource := ProbSourcePoint

		// 生效作弊值
		effectProb := user.User.GetProb()
		if effectProb == 0 {
			effectProb = game.Table.GetRoomProb()

			probSource = ProbSourceRoom
		}

		// 获取作弊率
		getProb := effectProb

		if effectProb == 0 {
			effectProb = 1000
		}

		// 获取手牌牌型字符串
		cardsTypeStr := poker.TransformCardsType(user.HoldCards.CardsType)
		game.Table.WriteLogs(user.ID, " 用户ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
			" 开始金币： "+score.GetScoreStr(user.InitAmount)+
			" 下注："+fmt.Sprintf(`%d`, user.BetMultiple)+"倍 "+
			" 角色： "+user.GetSysRole()+
			" 生效作弊率： "+fmt.Sprintf(`%d`, effectProb)+
			" 获取作弊率： "+fmt.Sprintf(`%d`, getProb)+
			" 作弊率来源： "+probSource+
			" 手牌："+cardsStr+
			" 牌型："+cardsTypeStr+
			" 输赢金额： "+score.GetScoreStr(settleResult.Result)+
			" 结束金额： "+score.GetScoreStr(game.UserList[settleResult.UserId].User.GetScore()))
	}

	// 广播结算结果
	settleResult := msg.SettleResultRes{
		ResultList: resultList,
	}
	game.SendSettleResult(settleResult)

	// 更改用户退出权限
	game.SetExitPermit(true)

	// 通知框架游戏结束
	game.Table.EndGame()

}

// GameOver 游戏结束
func (game *GeneralNiuniu) GameOver() {
	game.Status = int32(msg.GameStatus_GameOver)
	log.Tracef("游戏 %d 进入结束阶段", game.Table.GetID())

	// 广播结束状态
	game.SendGameStatus(game.Status, 0, nil)

	// 重置桌面属性
	game.Poker = nil
	game.TimeCfg = nil
	game.GameCfg = nil
	game.RoomCfg = nil
	game.LoadCfg = false
	game.CardsSequence = []poker.HoldCards{}
	game.ControlledCards = make(map[int64]poker.HoldCards)

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}

	// 所有人退出去
	for _, user := range game.UserList {
		game.UserLeaveGame(user.User)
		game.Table.KickOut(user.User)
	}

	game.Table.Close()
}
