package game

import (
	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960204/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
)

// CalculateCurrentUser 计算当前操作玩家
func (game *RunFaster) CalculateCurrentUser() {

	// 已经有人逃完牌，不寻找下一个玩家
	if len(game.UserList[game.CurrentPlayer.UserID].Cards) == 0 {
		game.CurrentPlayer = CurrentPlayer{
			UserID:     -1,
			ChairID:    -1,
			ActionTime: 0,
			Permission: false,
			StepCount:  game.CurrentPlayer.StepCount + 1,
			ActionType: int32(msg.UserActionType_NoPermission),
		}
	} else {
		currentUserID := game.CurrentPlayer.UserID

		// 下一个玩家座位id
		nextChairID := game.CurrentPlayer.ChairID + 1

		if int(game.CurrentPlayer.ChairID) == len(game.UserList)-1 {
			nextChairID = 0
		}

		// 下一个操作玩家
		nextUser := game.UserList[game.Seats[nextChairID]]
		game.CurrentPlayer = CurrentPlayer{
			UserID:    nextUser.ID,
			ChairID:   nextUser.ChairID,
			StepCount: game.CurrentPlayer.StepCount + 1,
		}

		// 判断当前牌权到玩家是否是自己
		if game.CurrentCards.UserID == game.CurrentPlayer.UserID {

			// 炸弹获得的牌权，进行一次实时计算
			if game.CurrentCards.CardsType == int32(msg.CardsType_Bomb) {
				game.BoomSettle(game.CurrentCards.UserID)
			}

			// 是则直接获得出牌权限
			game.CurrentPlayer.ActionTime = game.TimeCfg.OperationTime
			game.CurrentPlayer.Permission = true
			game.CurrentPlayer.ActionType = int32(msg.UserActionType_PutCard)

			// 最后一手牌可直接出完,并且不能是4带3
			cardsType := poker.GetCardsType(nextUser.Cards)
			if cardsType != msg.CardsType_Normal && cardsType != msg.CardsType_QuartetWithThree {
				game.CurrentPlayer.IsFinalEnd = true
			}
		} else {

			// 不是则进行接牌判断
			takeOverCards := poker.CheckTakeOverCards(game.CurrentCards, nextUser.Cards, false)

			if len(takeOverCards) != 0 {
				game.CurrentPlayer.ActionTime = game.TimeCfg.OperationTime
				game.CurrentPlayer.Permission = true
				game.CurrentPlayer.ActionType = int32(msg.UserActionType_TakeOverCard)

				// 最后一手牌可以直接出完,排除4带3的情况
				if len(takeOverCards) == len(nextUser.Cards) && game.CurrentCards.CardsType != int32(msg.CardsType_QuartetWithThree) {
					game.CurrentPlayer.IsFinalEnd = true
				}

			} else {
				game.CurrentPlayer.ActionTime = game.TimeCfg.ExcessiveTime
				game.CurrentPlayer.Permission = false
				game.CurrentPlayer.ActionType = int32(msg.UserActionType_NoPermission)

				// 上家如果承担放走包赔风险又不能接牌，停止让上家承担风险
				if game.UserList[currentUserID].TakeSingleRisk {
					game.UserList[currentUserID].TakeSingleRisk = false
				}
			}
		}

		if nextUser.Status == int32(msg.UserStatus_UserHangUp) {
			game.CurrentPlayer.ActionTime = 0
		}

	}

}

// CheckCardsExist 检查牌是否存在
func (game *RunFaster) CheckCardsExist(userID int64, cards []byte) bool {

	for _, card := range cards {
		isExist := false
		for _, userCard := range game.UserList[userID].Cards {
			if card == userCard {
				isExist = true
				break
			}
		}

		if !isExist {
			return false
		}
	}

	return true
}

// CheckTakeSingleRisk 放走包赔检测
func (game *RunFaster) CheckTakeSingleRisk(user *data.User, putCard byte) bool {
	// 下一个玩家座位id
	nextChairID := user.ChairID + 1

	if int(user.ChairID) == len(game.UserList)-1 {
		nextChairID = 0
	}

	// 下家只有一张手牌
	if len(game.UserList[game.Seats[nextChairID]].Cards) == 1 {

		// 找到了比出牌更大到单张手牌
		putCardValue, _ := poker.GetCardValueAndColor(putCard)
		for _, card := range user.Cards {
			cardValue, _ := poker.GetCardValueAndColor(card)
			if cardValue > putCardValue {
				return true
			}
		}
	}

	return false
}

// BoomSettle 炸弹实时结算
func (game *RunFaster) BoomSettle(userID int64) {
	settleResultList := []*msg.SettleResult{}
	winner, ok := game.UserList[userID]
	if !ok {
		log.Errorf("游戏 %d 炸弹结算寻找赢家 %d 错误", game.Table.GetID(), userID)
		return
	}

	// 防止以小博大
	winResult := 20 * game.RoomCfg.RoomCost
	if winner.CurAmount < 10*game.RoomCfg.RoomCost {
		winResult = 2 * winner.CurAmount
	}

	var totalLoseResult int64
	for id, user := range game.UserList {
		if id == userID {
			continue
		}
		loseResult := -10 * game.RoomCfg.RoomCost

		// 防止不够赔
		if user.CurAmount < winResult/2 {
			loseResult = -user.CurAmount
		}
		totalLoseResult += loseResult

		//if loseResult != 0 {
		//_, err := user.User.SetScore(game.Table.GetGameNum(), loseResult, "炸弹结算下分", game.Table.GetRoomRate(), 0, SET_SCORE_SETTLE, game.GameCfg.BoomDownID)
		//if err != nil {
		//	log.Errorf("输家下分失败：%v", err.Error())
		//}
		//}

		settleResultList = append(settleResultList, &msg.SettleResult{
			UserId:  id,
			ChairId: user.ChairID,
			Result:  loseResult,
		})
		game.UserList[id].BoomSettle += loseResult
		game.UserList[id].CurAmount += loseResult
	}

	winResult = -totalLoseResult

	//if winResult != 0 {
	//_, err := winner.User.SetScore(game.Table.GetGameNum(), winResult, "炸弹结算上分", game.Table.GetRoomRate(), 0, SET_SCORE_SETTLE, game.GameCfg.BoomUpID)
	//if err != nil {
	//	log.Errorf("赢家上分失败：%v", err.Error())
	//}
	//}

	settleResultList = append(settleResultList, &msg.SettleResult{
		UserId:  userID,
		ChairId: winner.ChairID,
		Result:  winResult,
	})
	game.UserList[userID].BoomCount++
	game.UserList[userID].BoomSettle += winResult
	game.UserList[userID].CurAmount += winResult

	settleInfoRes := msg.SettleInfoRes{
		ResultList: settleResultList,
		SettleType: 1,
	}
	game.SendSettleInfo(settleInfoRes)
}

// RobotSitCheck 机器人坐下检测
func (game *RunFaster) RobotSitCheck() {

	log.Tracef("请求机器人坐满桌子")
	// 现有人数
	count := len(game.UserList)
	if count >= 3 {
		return
	}

	// 游戏状态检测
	if game.Status != int32(msg.GameStatus_GameInitStatus) {
		log.Errorf("游戏 %d 在状态为 %d 请求机器人", game.Table.GetID(), game.Status)
		return
	}

	if 3-count > 0 {
		err := game.Table.GetRobot(uint32(3-count), game.Table.GetConfig().RobotMinBalance, game.Table.GetConfig().RobotMaxBalance)
		if err != nil {
			log.Errorf("游戏 %d 请求机器人失败：%v", game.Table.GetID(), err)
		}
	}
}

// IsNextReportSingle 下家是否报单
func (game *RunFaster) IsNextReportSingle(userID int64) (isNextSingle bool) {
	nextChairID := game.UserList[userID].ChairID + 1
	if nextChairID == int32(len(game.UserList)) {
		nextChairID = 0
	}

	if len(game.UserList[game.Seats[nextChairID]].Cards) == 1 {
		isNextSingle = true
	}

	return
}

// SettleDivision 结算上下分
func (game *RunFaster) SettleDivision(userID int64) (result, profit int64) {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("用户 %d 上下分查找不到当前用户")
		return
	}

	// 结果
	result = user.CurAmount - user.InitAmount

	_, profit = game.UserList[userID].User.SetScore(game.Table.GetGameNum(), result, game.RoomCfg.TaxRate)

	// 打码量
	var chip int64

	// 有扣税操作，更新当前金额
	if profit > 0 {
		game.UserList[userID].CurAmount = user.InitAmount + profit

		chip = game.RoomCfg.RoomCost
	} else {
		chip = profit
	}

	// 设置打码量
	game.SetChip(userID, chip)

	return result, profit
}

// SetChip 设置码量
func (game *RunFaster) SetChip(userID int64, chip int64) {
	if chip < 0 {
		chip = -chip
	}
	game.UserList[userID].User.SendChip(chip)
}

// SetExitPermit 设置用户退出权限
func (game *RunFaster) SetExitPermit(permit bool) {
	for id := range game.UserList {
		game.UserList[id].ExitPermit = permit
	}
}

// UserSendRecord 发送战绩，计算产出
func (game *RunFaster) TableSendRecord(userID int64, result int64, netProfit int64) (profitAmount int64, betsAmount int64, drawAmount int64, outputAmount int64) {
	// user, ok := game.UserList[userID]
	// if !ok {
	// 	log.Errorf("发送战绩查询用户 %d 失败", userID)
	// 	return
	// }
	// var (
	// 	profitAmount int64  // 盈利
	// 	betsAmount   int64  // 总下注
	// 	drawAmount   int64  // 总抽水
	// 	outputAmount int64  // 总产出
	// 	endCards     string // 结算牌
	// )

	profitAmount = netProfit

	if netProfit >= 0 {
		betsAmount = game.RoomCfg.RoomCost
		drawAmount = result - netProfit
		outputAmount = netProfit
	} else {
		betsAmount = netProfit
	}

	log.Tracef("用户 %d 盈利: %d, 总下注: %d, 总抽水: %d, 总产出: %v", userID, profitAmount, betsAmount, drawAmount, outputAmount)
	//user.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, drawAmount, outputAmount, endCards)
	return profitAmount, betsAmount, drawAmount, outputAmount
}

type SettleResult struct {
	TheorySettle int64 // 理论结算值
	ActualSettle int64 // 实际结算值
	CurAmount    int64 // 携带金额
}

// FillLoserAmount 折算多余输家金额，补足应输金额小于携带金额但是 按比例补足会触发防一小博大，则补到携带金额大小
// leftAmount 剩余多输金额
// leftTheoryLoseSum 剩余理论输钱合值
// LoserList 输家列表
func FillLoserAmount(leftAmount *int64, leftTheoryLoseSum *int64, LoserList map[int64]*SettleResult) {

	newLeftAmount := *leftAmount
	// 按比例折扣 会 触发防止一小博大机制，则先补足金额
	for userID, v := range LoserList {

		if v.ActualSettle <= -v.CurAmount {
			continue
		}

		// 按比例折算金额
		convertAmount := *leftAmount * v.TheorySettle / *leftTheoryLoseSum

		if convertAmount+v.ActualSettle < -v.CurAmount {
			*leftAmount += v.CurAmount + v.ActualSettle
			LoserList[userID].ActualSettle = -v.CurAmount
			*leftTheoryLoseSum -= v.TheorySettle
			break
		}
	}

	// 剩余钱无变化，不需要再补足，跳出循环
	if newLeftAmount != *leftAmount {
		FillLoserAmount(leftAmount, leftTheoryLoseSum, LoserList)
	}
}

// ConvertLoserAmount 按比例折算多余输家金额
func ConvertLoserAmount(leftAmount int64, leftTheoryLoseSum int64, LoserList map[int64]*SettleResult) map[int64]*SettleResult {
	if leftAmount >= 0 || leftTheoryLoseSum >= 0 {
		log.Errorf("按比例折算多余输家金额出现错误，剩余金额：%d，剩余输钱理论合值：%d", leftAmount, leftTheoryLoseSum)
		return LoserList
	}

	var (
		convertCount int   // 需要补足多余金额玩家个数
		loseCounter  int   // 输家计数器
		loseAcc      int64 // 输钱累加器
	)

	for _, v := range LoserList {
		if v.ActualSettle > -v.CurAmount {
			convertCount++
		}
	}

	// 按比例折扣
	for userID, v := range LoserList {

		if v.ActualSettle > -v.CurAmount {

			// 最后一个需要补足多余金额到输家
			if convertCount-loseCounter == 1 {
				LoserList[userID].ActualSettle += leftAmount - loseAcc
				break
			}

			// 按比例折算金额
			convertAmount := leftAmount * v.TheorySettle / leftTheoryLoseSum
			LoserList[userID].ActualSettle += convertAmount
			loseAcc += convertAmount
			loseCounter++
		}
	}

	return LoserList
}
