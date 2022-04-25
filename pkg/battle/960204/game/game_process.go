package game

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

// Start 开始游戏逻辑
func (game *RunFaster) Start() {
	game.Status = int32(msg.GameStatus_ReadyStatus)
	log.Tracef("游戏 %d 开始", game.Table.GetID())

	// 更改用户退出权限
	game.SetExitPermit(false)

	// 初始化出牌日志
	game.PutCardsLog = ""

	// 通知框架开赛
	game.Table.StartGame()

	// 洗牌
	game.Poker = new(poker.GamePoker)
	game.Poker.InitPoker()

	// 如果玩家已经配牌，从牌库中删除已经配置的牌
	for _, user := range game.UserList {
		if len(user.Cards) != 0 {
			for _, card := range user.Cards {
				for i, waitDelCard := range game.Poker.Cards {
					if card == waitDelCard {
						game.Poker.Cards = append(game.Poker.Cards[:i:i], game.Poker.Cards[i+1:]...)
						break
					}
				}
			}
		}
	}

	for _, v := range poker.Deck {
		game.LeftCards = append(game.LeftCards, v)
	}

	game.DealCards()
}

// DealCards 发牌
func (game *RunFaster) DealCards() {
	game.Status = int32(msg.GameStatus_DealStatus)
	log.Tracef("游戏 %d 开始发牌", game.Table.GetID())

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.SpaceTime), game.PutCards)

	// 广播游戏发牌状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.SpaceTime/1000), nil)

	// 没有配牌了 控牌
	isConfCard := true
	if len(game.Poker.Cards) == 48 {

		// 控牌
		game.ControlPoker()

		fmt.Printf("换牌前：%v\n", time.Now())
		// 换牌控制
		game.ExchangeControl()
		fmt.Printf("换牌后：%v\n", time.Now())

		isConfCard = false
	}

	for id, user := range game.UserList {
		var cards []byte

		if isConfCard {
			if len(user.Cards) == 0 {
				cards = game.Poker.DrawCard()
				game.UserList[id].Cards = cards
			} else {
				cards = user.Cards
			}
		} else {
			cards = game.ControlledCards[id]
			game.UserList[id].Cards = cards
		}

		if game.CurrentPlayer.UserID != 0 {
			continue
		}

		// 找到黑桃3为第一个可执行玩家
		for _, card := range cards {
			if card == 0x34 {
				game.CurrentPlayer = CurrentPlayer{
					UserID:     user.ID,
					ChairID:    user.ChairID,
					ActionTime: game.TimeCfg.OperationTime,
					Permission: true,
					StepCount:  0,
					ActionType: int32(msg.UserActionType_PutCard),
				}
				break
			}
		}
	}

	for _, user := range game.UserList {

		cardsStr := poker.CardsToString(user.Cards)

		// 记录玩家日志
		game.Table.WriteLogs(user.ID, " 用户ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
			" 手牌： "+cardsStr)

		dealResp := msg.DealRes{
			Cards:              user.Cards,
			FirstActionChairId: game.CurrentPlayer.ChairID,
			FirstActionUserId:  game.CurrentPlayer.UserID,
			UserId:             user.ID,
			ChairId:            user.ChairID,
		}
		game.SendDealInfo(dealResp, user.User)
	}

}

// PutCards 出牌
func (game *RunFaster) PutCards() {
	game.Status = int32(msg.GameStatus_PutCardStatus)
	log.Tracef("游戏 %d 进入出牌阶段", game.Table.GetID())

	game.StepCount = 0

	game.TurnAction()
}

// TurnAction 循环操作动作
func (game *RunFaster) TurnAction() {

	// 轮转流程
	if game.CurrentPlayer.UserID != -1 {
		user, ok := game.UserList[game.CurrentPlayer.UserID]
		if !ok {
			log.Errorf("轮转操作获取当前玩家 %d 出现错误，用户列表 %v", game.CurrentPlayer.UserID, game.UserList)
			return
		}

		log.Tracef("系统轮转数 %d, 当前玩家轮转数 %d，当前玩家%d", game.StepCount, game.CurrentPlayer.StepCount, game.CurrentPlayer.UserID)
		if game.StepCount != game.CurrentPlayer.StepCount {
			log.Tracef("操作计数器不一致，有玩家未操作")

			putCardsReq := &msg.PutCardsReq{}

			// 下家是否报单
			isNextSingle := game.IsNextReportSingle(user.ID)

			// todo系统代替玩家出牌
			switch game.CurrentPlayer.ActionType {
			case int32(msg.UserActionType_PutCard):

				putCards := poker.CheckPutSingleCard(user.Cards, isNextSingle)
				putCardsReq.Cards = putCards
				break
			case int32(msg.UserActionType_TakeOverCard):
				takeOverCards := poker.CheckTakeOverCards(game.CurrentCards, user.Cards, isNextSingle)
				if len(takeOverCards) == 0 {
					log.Errorf("玩家 %d 出现错误了不能接牌", user.ID)
				}

				putCardsReq.Cards = takeOverCards
				break
			}

			log.Tracef("系统代替操作入参：%v", putCardsReq)
			buffer, err := proto.Marshal(putCardsReq)
			if err != nil {
				log.Errorf("proto marshal fail :", err.Error())
				return
			}

			game.UserPutCards(buffer, user.User, true)

		} else {

			// 广播当前玩家
			currentPlayerResp := msg.CurrentPlayerRes{
				UserId:     game.CurrentPlayer.UserID,
				ChairId:    game.CurrentPlayer.ChairID,
				ActionTime: int32(game.CurrentPlayer.ActionTime / 1000),
				Permission: game.CurrentPlayer.Permission,
				ActionType: game.CurrentPlayer.ActionType,
				IsFinalEnd: game.CurrentPlayer.IsFinalEnd,
			}
			game.SendCurrentPlayer(currentPlayerResp)

			if game.CurrentPlayer.ActionType == int32(msg.UserActionType_TakeOverCard) {
				game.UserGetTips([]byte{}, game.UserList[game.CurrentPlayer.UserID].User)
			}

			// 定时执行下一次循环操作
			duration := int64(game.CurrentPlayer.ActionTime)
			//game.TimerJob, _ = game.Table.AddTimer(duration, game.TurnAction)

			// 有牌权 并且玩家最后一手牌可直接终结比赛
			if game.CurrentPlayer.IsFinalEnd {
				putCardsReq := &msg.PutCardsReq{}
				putCardsReq.Cards = game.UserList[game.CurrentPlayer.UserID].Cards

				log.Tracef("最后一手牌，系统代替操作入参：%v", putCardsReq)
				buffer, err := proto.Marshal(putCardsReq)
				if err != nil {
					log.Errorf("proto marshal fail :", err.Error())
					return
				}

				game.UserPutCards(buffer, game.UserList[game.CurrentPlayer.UserID].User, false)
			} else {
				// 没有出牌权，直接寻找下一个操作玩家
				if !game.CurrentPlayer.Permission {
					// 更新玩家打牌记录打出一张空牌
					game.UserList[game.CurrentPlayer.UserID].PutCardsRecords = append(game.UserList[game.CurrentPlayer.UserID].PutCardsRecords, []byte{poker.NULL_CARD})

					// 添加出牌日志
					game.PutCardsLog += game.UserList[game.CurrentPlayer.UserID].GetSysRole() + "ID: " + fmt.Sprintf(`%d`, user.User.GetID()) + " 要不起, "

					// 寻找下一个操作玩家
					game.CalculateCurrentUser()
				}
				game.StepCount++
				if duration == 0 {
					game.TurnAction()
				} else {
					game.TimerJob, _ = game.Table.AddTimer(duration, game.TurnAction)
				}
			}

		}

	} else {

		// 流程结束, 发送摊牌信息

		userList := []*msg.SeatUserInfoRes{}

		for _, user := range game.UserList {
			userList = append(userList, &msg.SeatUserInfoRes{
				ChairId: user.ChairID,
				Cards:   user.Cards,
				UserId:  user.ID,
			})
		}

		showCardsRes := msg.ShowCardsRes{
			ShowList: userList,
		}
		game.SendShowCardsInfo(showCardsRes)

		game.Settle()
	}

}

// Settle 结算
func (game *RunFaster) Settle() {

	game.Status = int32(msg.GameStatus_SettleStatus)
	log.Tracef("游戏 %d 进入结算阶段", game.Table.GetID())

	// 闲家输家列表
	loserList := make(map[int64]*SettleResult)

	var (
		settleResultList []*msg.SettleResult // 结算结果列表
		winnerResult     int64               // 赢家结算结果
		theoryLoseSum    int64               // 输家理论输钱合值
		winner           *data.User          // 赢家
	)

	// 输家结算统计
	for id, user := range game.UserList {
		if user.SettleCost > 0 {
			winner = user
			continue
		}
		cardCount := len(user.Cards)
		result := user.SettleCost

		// 被全关
		if cardCount == 16 {
			result = result * 2
		}

		// 被另外一个输家放走包赔，不加入输家列表
		if result == 0 {
			continue
		}

		// 输家加入输家列表
		loserList[id] = &SettleResult{
			TheorySettle: result,
			ActualSettle: result,
			CurAmount:    user.CurAmount,
		}

		// 统计理论输钱合值
		theoryLoseSum += result

		// 输家防止不够输
		if -result > user.CurAmount {
			result = -user.CurAmount
			loserList[id].ActualSettle = -user.CurAmount
		}

		winnerResult += -result

	}

	if winner == nil {
		log.Errorf("游戏 %d 结算时寻找赢家错误", game.Table.GetID())
		return
	}

	// 赢钱大于携带本金，防止以小博大, 输钱玩家等比例减少输钱金额
	if winnerResult+winner.BoomSettle > winner.InitAmount {
		var (
			loseCounter int   // 输家计数器
			loseAcc     int64 // 输钱累加器
		)

		winnerResult = winner.InitAmount - winner.BoomSettle

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

	// 构建结算列表
	for id, user := range game.UserList {
		settleResult := user.SettleCost
		if id == winner.ID {
			settleResult = winnerResult
		}

		for loserID, loser := range loserList {
			if id == loserID {
				settleResult = loser.ActualSettle
			}
		}

		settleResultList = append(settleResultList, &msg.SettleResult{
			UserId:         id,
			ChairId:        user.ChairID,
			Result:         settleResult,
			CardCount:      int32(len(user.Cards)),
			BoomSettle:     user.BoomSettle,
			TakeSingleRisk: user.TakeSingleRisk,
		})
	}

	log.Tracef("结算列表 %s", fmt.Sprintf("%+v\n", settleResultList))

	//战绩
	var records []*platform.PlayerRecord

	// 上下分结算，发送战绩
	for index, userResult := range settleResultList {
		user, ok := game.UserList[userResult.UserId]
		if !ok {
			log.Errorf("用户 %d 上下分查找不到当前用户")
			continue
		}

		game.UserList[userResult.UserId].CurAmount += userResult.Result

		// 净赢钱
		result, netProfit := game.SettleDivision(userResult.UserId)

		// 发送战绩，计算产出
		profitAmount, betsAmount, drawAmount, outputAmount := game.TableSendRecord(userResult.UserId, result, netProfit)
		if !user.User.IsRobot() {
			records = append(records, &platform.PlayerRecord{
				PlayerID:     uint32(userResult.UserId),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: profitAmount,
				BetsAmount:   betsAmount,
				DrawAmount:   drawAmount,
				OutputAmount: outputAmount,
				Balance:      user.User.GetScore(),
				UpdatedAt:    time.Now(),
				CreatedAt:    time.Now(),
			})
		}
		settleResultList[index].Result = netProfit
	}

	//发送战绩
	if len(records) > 0 {
		if _, err := game.Table.UploadPlayerRecord(records); err != nil {
			log.Warnf("upload player record error %s", err.Error())
		}
	}
	// 记录出牌日志
	game.Table.WriteLogs(0, "出牌日志: "+game.PutCardsLog)

	// 记录游戏日志
	for _, settleResult := range settleResultList {
		user := game.UserList[settleResult.UserId]

		// 获取手牌字符串
		cardsStr := poker.CardsToString(user.Cards)

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

		game.Table.WriteLogs(user.ID, " 用户ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
			" 角色： "+user.GetSysRole()+
			" 生效作弊率： "+fmt.Sprintf(`%d`, effectProb)+
			" 获取作弊率： "+fmt.Sprintf(`%d`, getProb)+
			" 作弊率来源： "+probSource+
			" 余牌： "+cardsStr+
			" 炸弹： "+fmt.Sprintf(`%d`, user.BoomCount)+
			" 输赢金额： "+score.GetScoreStr(settleResult.Result)+
			" 剩余金额： "+score.GetScoreStr(game.UserList[settleResult.UserId].User.GetScore()))
	}

	settleInfoRes := msg.SettleInfoRes{
		ResultList: settleResultList,
		SettleType: 2,
	}
	game.SendSettleInfo(settleInfoRes)

	// 更改用户退出权限
	game.SetExitPermit(true)

	// 通知框架游戏结束
	game.Table.EndGame()

	game.End()
}

// GameOver 游戏结束
func (game *RunFaster) End() {
	// 重置桌面属性
	game.Status = int32(msg.GameStatus_GameOver)
	log.Tracef("游戏 %d 结束", game.Table.GetID())

	// 所有机器人退出去
	for _, user := range game.UserList {
		game.UserOffline(user.User)
		game.Table.KickOut(user.User)
	}

	game.Poker = nil
	game.TimerJob = nil
	game.RobotTimer = nil
	game.TimeCfg = nil
	game.GameCfg = nil
	game.RoomCfg = nil
	game.LoadCfg = false
	game.CurrentPlayer = CurrentPlayer{}
	game.CurrentCards = poker.HandCards{}
	game.LeftCards = []byte{}
	game.ControlledCards = make(map[int64][]byte)
	game.ControlList = []int64{}
	//game.StepCount = 0

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
	}
	log.Tracef("游戏 %d 桌子 %d 重置数据完成：%v", game.Table.GetID(), game.Table.GetRoomID(), game)

	//关闭
	game.Table.Close()
}
