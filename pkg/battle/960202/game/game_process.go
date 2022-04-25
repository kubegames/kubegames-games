package game

import (
	"fmt"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

// Start 游戏开始，进入动画状态
func (game *BankerNiuniu) Start() {
	game.Status = int32(msg.GameStatus_StartMove)
	log.Tracef("游戏 %d 倒计时开始", game.Table.GetID())

	if game.TimerJob != nil {
		game.Table.DeleteJob(game.TimerJob)
	}

	// 匹配机器人
	game.MatchRobot()

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StartMove), game.StartMoveEnd)

	// 广播游戏开始状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.StartMove/1000), nil)
}

// StartMoveEnd 结束倒计时动画
func (game *BankerNiuniu) StartMoveEnd() {
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

		// 提前改变状态防止，用户卡时间退出
		game.Status = int32(msg.GameStatus_RobBanker)
		log.Tracef("游戏 %d 进入抢庄阶段", game.Table.GetID())

		// 更改用户退出权限
		game.SetExitPermit(false)

		// 通知框架开赛
		game.Table.StartGame()

		// 开始游戏动画后开始抢庄
		game.RobotTimer, _ = game.Table.AddTimer(int64(game.TimeCfg.StartAnimation), game.StartRob)

	}

}

// StartRob 开始抢庄
func (game *BankerNiuniu) StartRob() {

	// 洗牌
	game.Poker = new(poker.GamePoker)
	game.Poker.InitPoker()

	game.Control()

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.RobBanker), game.EndRob)

	// 广播抢庄开始状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.RobBanker/1000), nil)

}

// EndRob 抢庄结束
func (game *BankerNiuniu) EndRob() {
	game.Status = int32(msg.GameStatus_EndRob)
	log.Tracef("游戏 %d 进入抢庄结束阶段", game.Table.GetID())

	var (
		candidates    []int32         // 抢庄候选人
		banker        *data.User      // 庄家
		bankerIndex   int32      = -1 // 庄家倍数下标
		bankerChairID int32           // 庄家用户ID
	)

	// 默认不抢发送抢庄信息
	for _, user := range game.UserList {

		if user.RobIndex == bankerIndex {
			candidates = append(candidates, user.ChairID)
		}

		if user.RobIndex > bankerIndex {
			bankerIndex = user.RobIndex
			bankerChairID = user.ChairID
			candidates = []int32{user.ChairID}
		}

		if user.Status != int32(msg.UserStatus_RobAction) {
			robBankerInfo := msg.RobBankerInfoRes{
				RobIndex: -1,
				UserId:   user.ID,
				ChairId:  user.ChairID,
			}
			game.SendRobBankerInfo(robBankerInfo)
		}
	}

	// 有最高抢庄人
	if len(candidates) == 1 {
		candidates = []int32{}
	}

	// 随机庄
	if len(candidates) >= 2 {
		index := rand.RandInt(0, len(candidates))
		bankerChairID = candidates[index]
	}

	// 无人抢庄，随机庄为最低倍数
	if bankerIndex == -1 {
		bankerIndex = 0
	}

	// 更新数据
	for id, user := range game.UserList {
		if user.ChairID == bankerChairID {
			user.IsBanker = true
			banker = user
		} else {
			user.IsBanker = false
		}
		game.UserList[id] = user
	}
	game.Banker = banker

	// 间隔一秒发送消息
	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StatusSpace), func() {
		// 广播抢庄结果
		robBankerResult := msg.RobBankerResultRes{
			RobIndex:      bankerIndex,
			Candidates:    candidates,
			BankerChairId: banker.ChairID,
		}
		game.SendRobBankerResult(robBankerResult)

		// 随机庄有动画
		if len(candidates) >= 2 {
			game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.RobAnimation), game.StartBet)
		} else {
			game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StatusSpace), game.StartBet)
		}
	})

}

// StartBet 开始投注
func (game *BankerNiuniu) StartBet() {
	game.Status = int32(msg.GameStatus_BetChips)
	log.Tracef("游戏 %d 进入投注阶段", game.Table.GetID())

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.BetChips), game.EndBet)

	// 广播投注开始状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.BetChips/1000), nil)

	// 计算玩家可投注倍数
	for id, user := range game.UserList {
		if user.IsBanker {
			continue
		}

		// 庄家可输最高投注倍数 = 庄家携带金额/桌面有效玩家数（除开庄家）/最高牌型倍数3/庄家倍数/底注
		highestMultipleOne := game.Banker.CurAmount / int64(len(game.UserList)-1) / 3 / game.GameCfg.RobOption[game.Banker.RobIndex] / game.RoomCfg.RoomCost

		// 闲家可输最高投注倍数 = 庄家携带金额/最高牌型倍数3/庄家倍数/底注
		highestMultipleTwo := user.CurAmount / 3 / game.GameCfg.RobOption[game.Banker.RobIndex] / game.RoomCfg.RoomCost

		// 去最低的 最高投注倍数
		highestMultiple := highestMultipleOne
		if highestMultipleOne > highestMultipleTwo {
			highestMultiple = highestMultipleTwo
		}

		// 倍率列表
		var multiples []int64
		if highestMultiple >= 20 {
			multiples = []int64{1, 5, 10, 20}
		}

		if highestMultiple <= 7 {
			multiples = []int64{1, 2, 3, highestMultiple}
		}

		if highestMultiple < 4 {
			multiples = []int64{1, 2, 3, 4}
		}

		if highestMultiple > 7 && highestMultiple < 20 {
			multiples = []int64{1, highestMultiple / 4, highestMultiple / 2, highestMultiple}
		}

		// 最高投注倍数 最低为1倍
		if highestMultiple == 0 {
			highestMultiple = 1
		}

		game.UserList[id].HighestMultiple = highestMultiple
		game.UserList[id].BetMultipleOption = multiples

		betMultipleResp := msg.BetMultipleRes{
			UserId:          id,
			ChairId:         user.ChairID,
			Multiples:       multiples,
			HighestMultiple: highestMultiple,
		}

		// 发送投注倍率信息
		game.SendBetMultipleInfo(betMultipleResp, user.User)
	}
}

// EndBet 投注结束
func (game *BankerNiuniu) EndBet() {
	game.Status = int32(msg.GameStatus_EndBet)
	log.Tracef("游戏 %d 进入投注结束阶段", game.Table.GetID())

	// 留点时间间隔发牌
	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.StatusSpace), game.DealCards)

	// 未投注的人默认投最低倍数
	for id, user := range game.UserList {
		if user.IsBanker {
			continue
		}

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

	// 换牌检测
	game.ExchangeControl()

	//game.DealCards()
}

// DealCards 发牌
func (game *BankerNiuniu) DealCards() {

	// 发牌动画时间
	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.DealAnimation*len(game.UserList)), game.StartShow)

	for id, user := range game.UserList {

		cards := game.ControlledCards[id].Cards
		cardsType := game.ControlledCards[id].CardsType

		if len(user.HoldCards.Cards) == 0 {
			user.HoldCards = &poker.HoldCards{
				Cards:             game.ControlledCards[id].Cards,
				CardsType:         game.ControlledCards[id].CardsType,
				SpecialCardIndexs: poker.GetSpecialCardIndexs(cards, cardsType),
			}
		}

		// 更新数据
		game.UserList[id] = user
		if user.IsBanker {
			game.Banker = user
		}

		// 发送发牌信息
		dealResult := msg.DealRes{
			Cards:   user.HoldCards.Cards,
			UserId:  id,
			ChairId: user.ChairID,
		}
		game.SendDealCardsMsg(dealResult, user.User)
	}
}

// StartShow 开始摊牌
func (game *BankerNiuniu) StartShow() {
	game.Status = int32(msg.GameStatus_ShowChards)
	log.Tracef("游戏 %d 进入摊牌阶段", game.Table.GetID())

	// 广播开始摊牌状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.ShowCards/1000), nil)

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.ShowCards), game.EndShow)

}

// EndShow 摊牌结束
func (game *BankerNiuniu) EndShow() {
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

// Settle 结算
func (game *BankerNiuniu) Settle() {
	game.Status = int32(msg.GameStatus_SettleStatus)
	log.Tracef("游戏 %d 进入结算阶段", game.Table.GetID())

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.Settle), game.GameOver)

	// 广播结算状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.Settle/1000), nil)

	// 闲家赢家列表
	winnerList := make(map[int64]*SettleResult)

	// 闲家输家列表
	loserList := make(map[int64]*SettleResult)

	var (
		resultList    []*msg.SettleResult // 结算列表
		bankerWin     int64               // 庄家赢钱
		bankerLose    int64               // 庄家输钱
		theoryWinSum  int64               // 闲家理论赢钱合值
		theoryLoseSum int64               // 闲家理论输钱合值
	)

	// 比牌，结算
	for _, user := range game.UserList {
		if user.IsBanker {
			continue
		}

		// 牌倍数
		var cardsMultiple int64

		if poker.ContrastCards(game.Banker.HoldCards, user.HoldCards) {

			// 闲家赢
			cardsMultiple = poker.GetCardsMultiple(user.HoldCards.CardsType)
		} else {

			// 闲家输
			cardsMultiple = poker.GetCardsMultiple(game.Banker.HoldCards.CardsType) * -1
		}

		// 闲家输赢
		result := game.RoomCfg.RoomCost *
			game.GameCfg.RobOption[game.Banker.RobIndex] *
			user.BetMultiple *
			cardsMultiple

		// 添加输赢列表，计算庄家输/赢合值，计算闲家理论输/赢合值
		if result < 0 {

			loserSettle := &SettleResult{
				TheorySettle: result,
				ActualSettle: result,
				CurAmount:    user.CurAmount,
			}

			theoryLoseSum += result

			// 闲家输钱大于携带本金，只输完携带金额
			if -result > user.CurAmount {
				result = -user.CurAmount
				loserSettle.ActualSettle = -user.CurAmount
			}

			bankerWin += -result
			loserList[user.ID] = loserSettle
		} else {

			winnerSettle := &SettleResult{
				TheorySettle: result,
				ActualSettle: result,
				CurAmount:    user.CurAmount,
			}

			theoryWinSum += result

			// 闲家赢钱大于携带本金，防止以小博大
			if result > user.CurAmount {
				result = user.CurAmount
				winnerSettle.ActualSettle = user.CurAmount
			}
			bankerLose += -result
			winnerList[user.ID] = winnerSettle
		}

	}

	// 庄家赢钱大于携带本金，防止以小博大, 输钱玩家等比例减少输钱金额
	if bankerWin+bankerLose > game.Banker.CurAmount {
		var (
			lossCounter int   // 输家计数器
			lossAcc     int64 // 输钱累加器
		)

		// 输钱闲家的应该输的钱 = 庄家携带金额 + 庄家应输的钱
		bankerWin = game.Banker.CurAmount - bankerLose

		for userID, v := range loserList {

			// 最后一个赢钱玩家
			if len(loserList)-lossCounter == 1 {
				loserList[userID].ActualSettle = (bankerWin + lossAcc) * -1
				break
			}

			// 缩减比例后新的结算值 = 应赢合值 * 玩家理论赢钱值 / 理论赢钱合值
			loserList[userID].ActualSettle = bankerWin * v.TheorySettle / theoryLoseSum * -1
			lossAcc += loserList[userID].ActualSettle
			lossCounter++
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

	// 庄家输钱大于携带本金，不够输，赢钱闲家等比例缩减赢钱金额
	if bankerWin+bankerLose < -game.Banker.CurAmount {
		var (
			winCounter int   // 赢家计数器
			winAcc     int64 // 赢钱累加器
		)

		// 赢钱闲家的应该赢的钱 = 庄家携带金额 + 庄家应赢的钱
		bankerLose = (game.Banker.CurAmount + bankerWin) * -1

		for userID, v := range winnerList {

			// 最后一个输钱玩家
			if len(winnerList)-winCounter == 1 {
				winnerList[userID].ActualSettle = -bankerLose - winAcc
				break
			}

			// 缩减比例后新的结算值 = 应输合值 * 玩家理论赢钱值 / 理论赢钱合值
			winnerList[userID].ActualSettle = -bankerLose * v.TheorySettle / theoryWinSum
			winAcc += winnerList[userID].ActualSettle
			winCounter++

		}

		var (
			leftAmount       int64 // 剩余多赢金额
			leftTheoryWinSum int64 // 剩余理论赢钱合值
		)

		for userID, v := range winnerList {

			// 缩减比例后 应赢钱大于携带金额，应赢钱减为携带金额，
			if v.ActualSettle >= v.CurAmount {
				leftAmount += v.ActualSettle - v.CurAmount
				winnerList[userID].ActualSettle = v.CurAmount
			} else {
				leftTheoryWinSum += v.TheorySettle
			}
		}

		// 有剩余多赢金额，重新分配一下剩余金额, 剩余钱 按照 玩家理论赢钱值 / 剩余玩家理论赢钱合值 比例分下去
		if leftAmount > 0 {
			FillWinnerAmount(&leftAmount, &leftTheoryWinSum, winnerList)
			ConvertWinnerAmount(leftAmount, leftTheoryWinSum, winnerList)
		}
	}

	// 赢家加入结算列表
	for userID, v := range winnerList {
		resultList = append(resultList, &msg.SettleResult{
			UserId:   userID,
			ChairId:  game.UserList[userID].ChairID,
			Result:   v.ActualSettle,
			IsBanker: false,
		})
	}

	// 输家加入结算列表
	for userID, v := range loserList {
		resultList = append(resultList, &msg.SettleResult{
			UserId:   userID,
			ChairId:  game.UserList[userID].ChairID,
			Result:   v.ActualSettle,
			IsBanker: false,
		})
	}

	// 庄家加入结算列表
	resultList = append(resultList, &msg.SettleResult{
		UserId:   game.Banker.ID,
		ChairId:  game.Banker.ChairID,
		Result:   bankerWin + bankerLose,
		IsBanker: true,
	})

	//战绩
	var records []*platform.PlayerRecord

	// 上下分
	for k, v := range resultList {

		// 赢钱下注金额为房间底注，输钱下注金额为输钱金额
		betAmount := game.RoomCfg.RoomCost
		if v.Result < 0 {
			betAmount = v.Result
		}

		_, netProfit := game.UserList[v.UserId].User.SetScore(game.Table.GetGameNum(), v.Result, game.Table.GetRoomRate())

		// 设置打码量
		game.SetChip(v.UserId, betAmount)

		// 更新玩家当前金额
		game.UserList[v.UserId].CurAmount += netProfit

		// 发送战绩，计算产出
		profitAmount, betsAmount, drawAmount, outputAmount := game.TableSendRecord(v.UserId, v.Result, netProfit)
		if !game.UserList[v.UserId].User.IsRobot() {
			records = append(records, &platform.PlayerRecord{
				PlayerID:     uint32(v.UserId),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: profitAmount,
				BetsAmount:   betsAmount,
				DrawAmount:   drawAmount,
				OutputAmount: outputAmount,
				Balance:      game.UserList[v.UserId].User.GetScore(),
				UpdatedAt:    time.Now(),
				CreatedAt:    time.Now(),
			})
		}

		resultList[k].Result = netProfit

		// 跑马灯触发
		if netProfit > 0 {
			specialWords := ""
			if game.UserList[v.UserId].HoldCards.CardsType > msg.CardsType_NiuNine {
				specialWords = poker.TransformCardsType(game.UserList[v.UserId].HoldCards.CardsType)
			}
			game.PaoMaDeng(netProfit, game.UserList[v.UserId].User, specialWords)
		}
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

		// 是否是庄家
		isBanker := "否"
		if user.IsBanker {
			isBanker = "是"
		}
		RobMultiple := "不抢"
		if user.RobIndex >= 0 {
			RobMultiple = fmt.Sprintf(`%d`, game.GameCfg.RobOption[user.RobIndex]) + "倍 "
		}

		// 获取手牌字符串
		cardsStr := poker.CardsToString(user.HoldCards.Cards)

		//作弊率来源
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
			" 抢庄： "+RobMultiple+
			" 是否是庄家： "+isBanker+
			" 投注： "+fmt.Sprintf(`%d`, user.BetMultiple)+"倍 "+
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
func (game *BankerNiuniu) GameOver() {
	game.Status = int32(msg.GameStatus_GameOver)
	log.Tracef("游戏 %d 进入结束阶段", game.Table.GetID())

	// 广播结束状态
	game.SendGameStatus(game.Status, 0, nil)

	// 重置桌面属性
	game.Banker = nil
	game.Poker = nil
	game.TimeCfg = nil
	game.GameCfg = nil
	game.RoomCfg = nil
	game.LoadCfg = false
	game.CardsSequence = []poker.HoldCards{}
	game.ControlledCards = make(map[int64]poker.HoldCards)
	game.ControlList = []int64{}

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}

	// 所有人退出去
	for _, user := range game.UserList {
		game.UserOffline(user.User)
		game.Table.KickOut(user.User)
	}

	game.Table.Close()
}
