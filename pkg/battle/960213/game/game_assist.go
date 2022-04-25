package game

import (
	"common/log"
	"common/rand"
	"game_poker/ddzall/data"
	"game_poker/ddzall/msg"
	"game_poker/ddzall/poker"
	"time"
)

// InitTable 初始化桌子
func (game *DouDizhu) InitTable() {
	// 初始化倍率数据
	game.BottomMultiple = 1
	game.BoomMultiple = 1
	game.RocketMultiple = 1
	game.AllOffMultiple = 1
	game.BeAllOffMultiple = 1
	game.RobCount = 0
	game.InAnimation = false

	// 初始化记牌器
	for _, v := range poker.Deck {
		game.LeftCards = append(game.LeftCards, v)
	}

	game.StepCount = 0
	game.ControlledCards = make(map[int64][]byte)
}

// GetEmptyChair 获取空椅子
func (game *DouDizhu) GetEmptyChair() (chairID int) {

	// 剩余空座位
	var emptySeats []int

	for chairID, user := range game.Chairs {

		if user == nil {
			emptySeats = append(emptySeats, chairID)
		}
	}

	// 没有空座位
	if len(emptySeats) == 0 {
		log.Warnf("游戏 %d 申请空座位失败，已经满员, 游戏座位: %v", game.Table.GetId(), game.Chairs)
		return -1
	}

	// 随机椅子索引
	randChair := rand.RandInt(0, len(emptySeats))

	chairID = emptySeats[randChair]

	return
}

// SetRobChairList 生成抢地主列表
func (game *DouDizhu) SetRobChairList() {

	switch game.RobCount {
	case 0:
		game.RobChairList = [3]int{0, 1, 2}
	case 1:
		game.RobChairList = [3]int{1, 2, 0}
	case 2:
		game.RobChairList = [3]int{2, 0, 1}
	}
}

// FindNextPlayer 寻找下一个玩家
func (game *DouDizhu) FindNextPlayer() {

	// 解锁桌子
	game.InAnimation = false

	// 下一个玩家座位id
	nextChairID := game.CurrentPlayer.ChairID + 1

	if int(game.CurrentPlayer.ChairID) == len(game.UserList)-1 {
		nextChairID = 0
	}

	// 下一个玩家
	nextUser := game.Chairs[nextChairID]

	// 操作时间
	actionTime := game.TimeCfg.ExcessiveTime

	// 操作权限和类型
	permission, actionType := game.GetUserActionPermission(nextUser)

	if permission {
		actionTime = game.TimeCfg.OperationTime
	}

	if nextUser.Status == int32(msg.UserStatus_UserHangUp) {
		actionTime = game.TimeCfg.HangUpTime
	}

	game.CurrentPlayer = CurrentPlayer{
		UserID:     nextUser.ID,
		ChairID:    nextUser.ChairID,
		ActionTime: actionTime,
		Permission: permission,
		StepCount:  game.CurrentPlayer.StepCount + 1,
		ActionType: int32(actionType),
	}

	// 广播当前玩家
	game.SendCurrentPlayer()

	// 定时检查
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.CurrentPlayer.ActionTime), game.CheckAction)
}

// GetUserActionPermission 获取用户的权限
func (game *DouDizhu) GetUserActionPermission(user *data.User) (permission bool, actionType msg.UserActionType) {

	// 判断当前牌权到玩家是自己
	if game.CurrentCards.UserID == user.ID {
		permission = true
		actionType = msg.UserActionType_PutCard
		return
	}

	//takeOverCards := poker.TakeOverCards(game.CurrentCards, user.Cards)

	if len(poker.TakeOverCards(game.CurrentCards, user.Cards)) != 0 {
		permission = true
		actionType = msg.UserActionType_TakeOverCard
	}

	return
}

// CheckCardsExist 检查牌是否存在
func (game *DouDizhu) CheckCardsExist(userID int64, cards []byte) bool {

	for _, card := range cards {

		if !poker.InByteArr(card, game.UserList[userID].Cards) {
			return false
		}

	}

	return true
}

// CheckCardsRepeated 卡牌重复性检测
func (game *DouDizhu) CheckCardsRepeated(cards []byte) bool {

	for i := 0; i < len(cards); i++ {

		for j := i + 1; j < len(cards); j++ {

			if cards[i] == cards[j] {
				return true
			}
		}
	}

	return false
}

// IsAllOff 春天检测
func (game *DouDizhu) IsAllOff() bool {

	for _, user := range game.UserList {
		if user.IsDizhu {
			continue
		}

		for _, cards := range user.PutCardsRecords {
			if len(cards) != 0 {
				return false
			}
		}

	}

	return true
}

// IsBeAllOff 反春检测
func (game *DouDizhu) IsBeAllOff() bool {
	// 出牌次数
	var putCardsCount int

	for _, cards := range game.Dizhu.PutCardsRecords {
		if len(cards) != 0 {
			putCardsCount++
		}
	}

	// 地主只出了一首，并且并不是第一手出完
	if putCardsCount == 1 && len(game.Dizhu.PutCardsRecords[0]) != 20 {
		return true
	}

	return false
}

// RobotSitCheck 机器人坐下检测
func (game *DouDizhu) RobotSitCheck() {

	log.Tracef("请求机器人坐满桌子")
	// 现有人数
	count := len(game.UserList)
	if count >= 3 {
		return
	}

	// 游戏状态检测
	if game.Status != int32(msg.GameStatus_GameInitStatus) {
		log.Errorf("游戏 %d 在状态为 %d 请求机器人", game.Table.GetId(), game.Status)
		return
	}

	if 3-count > 0 {
		err := game.Table.GetRobot(int32(3 - count))
		if err != nil {
			log.Errorf("游戏 %d 请求机器人失败：%v", game.Table.GetId(), err)
		}
	}
}

// GetNextChairID 获取下家座位ID
func GetNextChairID(chairID int32) int32 {
	return (chairID + 1) % 3
}

// GetPreChairID 获取上家座位ID
func GetPreChairID(chairID int32) int32 {
	return (chairID + 2) % 3
}

// SettleDivision 结算上下分
func (game *DouDizhu) SettleDivision(userID int64) int64 {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("用户 %d 上下分查找不到当前用户", userID)
		return 0
	}

	// 结果
	result := user.CurAmount - user.InitAmount

	profit, err := game.UserList[userID].User.SetScore(game.Table.GetGameNum(), result, game.RoomCfg.TaxRate)
	if err != nil {
		log.Errorf("用户 %d 上下分失败：%v", user.ID, err.Error())
		return 0
	}

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

	return profit
}

// SetChip 设置码量
func (game *DouDizhu) SetChip(userID int64, chip int64) {
	if chip < 0 {
		chip = -chip
	}
	game.UserList[userID].User.SendChip(chip)
}

// SetExitPermit 设置用户退出权限
func (game *DouDizhu) SetExitPermit(permit bool) {
	for id, _ := range game.UserList {
		game.UserList[id].ExitPermit = permit
	}
}

// UserSendRecord 发送战绩，计算产出
func (game *DouDizhu) TableSendRecord(userID int64, result int64, netProfit int64) {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("发送战绩查询用户 %d 失败", userID)
		return
	}
	var (
		profitAmount int64  // 盈利
		betsAmount   int64  // 总下注
		drawAmount   int64  // 总抽水
		outputAmount int64  // 总产出
		endCards     string // 结算牌
	)

	profitAmount = netProfit

	if netProfit >= 0 {
		betsAmount = game.RoomCfg.RoomCost
		drawAmount = result - netProfit
		outputAmount = netProfit
	} else {
		betsAmount = result
	}

	user.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, drawAmount, outputAmount, endCards)
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
