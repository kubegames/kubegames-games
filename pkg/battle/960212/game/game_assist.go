package game

import (
	"fmt"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960212/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960212/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960212/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
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
		log.Warnf("游戏 %d 申请空座位失败，已经满员, 游戏座位: %v", game.Table.GetID(), game.Chairs)
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
	game.TimerJob, _ = game.Table.AddTimer(int64(game.CurrentPlayer.ActionTime), game.CheckAction)
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

// SetExitPermit 设置用户退出权限
func (game *DouDizhu) SetExitPermit(permit bool) {
	for id := range game.UserList {
		game.UserList[id].ExitPermit = permit
	}
}

type SettleResult struct {
	TheorySettle int64 // 理论结算值
	ActualSettle int64 // 实际结算值
	CurAmount    int64 // 携带金额
}

// SettleDetail 结算细节（防以小博大）
func (game *DouDizhu) SettleDetail() {

	// 输家列表
	loserList := make(map[int64]*SettleResult)

	// 赢家列表
	winnerList := make(map[int64]*SettleResult)

	var (
		theoryWinSum  int64 // 理论赢钱合值
		theoryLoseSum int64 // 理论输钱合值
		actualWinSum  int64 // 实际赢钱合值
		actualLoseSum int64 // 实际输钱合值
	)

	for ID, user := range game.UserList {
		if user.SettleResult >= 0 {

			// 统计理论赢钱合值
			theoryWinSum += user.SettleResult

			winnerList[ID] = &SettleResult{
				TheorySettle: user.SettleResult,
				ActualSettle: user.SettleResult,
				CurAmount:    user.CurAmount,
			}

			// 赢钱大于携带钱
			if user.SettleResult > user.InitAmount {
				game.UserList[ID].SettleResult = user.InitAmount
				winnerList[ID].ActualSettle = user.InitAmount
				game.UserList[ID].SettleResult = user.InitAmount
			}

			// 统计实际赢钱合值
			actualWinSum += game.UserList[ID].SettleResult

		} else {

			// 统计理论输钱合值
			theoryLoseSum += user.SettleResult

			loserList[ID] = &SettleResult{
				TheorySettle: user.SettleResult,
				ActualSettle: user.SettleResult,
				CurAmount:    user.CurAmount,
			}

			// 输钱大于携带钱
			if -user.SettleResult > user.InitAmount {
				game.UserList[ID].SettleResult = -user.InitAmount
				loserList[ID].ActualSettle = -user.InitAmount
				game.UserList[ID].SettleResult = -user.InitAmount
			}

			// 统计实际输钱合值
			actualLoseSum += game.UserList[ID].SettleResult

		}
	}

	// 实际赢钱合值 > 实际输钱合值
	if actualWinSum > -actualLoseSum {

		var (
			winCounter int   // 赢家计数器
			winAcc     int64 // 赢钱累加器
		)

		for ID, v := range winnerList {

			// 最后一个赢钱玩家
			if len(winnerList)-winCounter == 1 {
				winnerList[ID].ActualSettle = (actualLoseSum + winAcc) * -1

				// 更新赢钱玩家
				game.UserList[ID].SettleResult = winnerList[ID].ActualSettle
				break
			}
			// 缩减比例后新的结算值 = 应输合值 * 玩家理论赢钱值 / 理论赢钱合值
			winnerList[ID].ActualSettle = actualLoseSum * v.TheorySettle / theoryWinSum * -1

			// 更新赢钱玩家
			game.UserList[ID].SettleResult = winnerList[ID].ActualSettle
			winAcc += winnerList[ID].ActualSettle
			winCounter++

		}
	}

	// 实际赢钱合值 < 实际输钱合值
	if actualWinSum < -actualLoseSum {

		var (
			loseCounter int   // 输家计数器
			loseAcc     int64 // 输钱累加器
		)

		for ID, v := range loserList {

			// 最后一个赢钱玩家
			if len(loserList)-loseCounter == 1 {
				loserList[ID].ActualSettle = (actualWinSum + loseAcc) * -1

				// 更新输钱玩家
				game.UserList[ID].SettleResult = loserList[ID].ActualSettle
				break
			}
			// 缩减比例后新的结算值 = 应赢合值 * 玩家理论输钱值 / 理论输钱合值
			loserList[ID].ActualSettle = actualWinSum * v.TheorySettle / theoryLoseSum * -1

			// 更新输钱玩家
			game.UserList[ID].SettleResult = loserList[ID].ActualSettle
			loseAcc += loserList[ID].ActualSettle
			loseCounter++

		}
	}

}

// SettleDivision 结算上下分
func (game *DouDizhu) SettleDivision() {
	var records []*platform.PlayerRecord
	for ID, user := range game.UserList {

		profit := user.User.SetScore(game.Table.GetGameNum(), user.SettleResult, game.RoomCfg.TaxRate)

		user.SettleResult = profit

		game.PaoMaDeng(profit, user.User, "")

		// 打码量
		var chip int64

		// 有扣税操作，更新当前金额
		if profit > 0 {
			user.CurAmount = user.InitAmount + profit

			chip = game.RoomCfg.RoomCost
		} else {
			chip = profit
		}

		// 设置打码量
		game.SetChip(ID, chip)

		// 发送战绩，计算产出
		if !user.User.IsRobot() {
			if record := game.TableSendRecord(ID, user.SettleResult, profit); record != nil {
				records = append(records, record)
			}
		}

		// 更新user数据
		game.UserList[ID] = user
	}

	if len(records) > 0 {
		game.Table.UploadPlayerRecord(records)
	}
}

// SetChip 设置码量
func (game *DouDizhu) SetChip(userID int64, chip int64) {
	if chip < 0 {
		chip = -chip
	}
	game.UserList[userID].User.SendChip(chip)
}

// UserSendRecord 发送战绩，计算产出
func (game *DouDizhu) TableSendRecord(userID int64, result int64, netProfit int64) *platform.PlayerRecord {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("发送战绩查询用户 %d 失败", userID)
		return nil
	}
	var (
		profitAmount int64 // 盈利
		betsAmount   int64 // 总下注
		drawAmount   int64 // 总抽水
		outputAmount int64 // 总产出
		//endCards     string // 结算牌
	)

	profitAmount = netProfit

	if netProfit >= 0 {
		betsAmount = game.RoomCfg.RoomCost
		drawAmount = result - netProfit
		outputAmount = netProfit
	} else {
		betsAmount = result
	}

	//user.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, drawAmount, outputAmount, endCards)
	return &platform.PlayerRecord{
		PlayerID:     uint32(user.User.GetID()),
		GameNum:      game.Table.GetGameNum(),
		ProfitAmount: profitAmount,
		BetsAmount:   betsAmount,
		DrawAmount:   drawAmount,
		OutputAmount: outputAmount,
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}
}

// RecordLogOne 记录第一部分日志
func (game *DouDizhu) RecordLogOne() {
	bottomCardsStr := poker.CardsToString(game.bottomCards)

	BottomMultipleStr := "否"
	if game.BottomMultiple > 1 {
		BottomMultipleStr = "是"
	}

	game.Table.WriteLogs(0, " 地主牌： "+bottomCardsStr+
		" 地主牌是否翻倍： "+BottomMultipleStr)

	for _, user := range game.UserList {
		// 角色
		role := user.GetSysRole()

		// 手牌
		cardsStr := poker.CardsToString(poker.ReverseSortCards(user.Cards))
		isDizhuStr := "否"

		// 是否是地主
		if user.IsDizhu {
			isDizhuStr = "是"
		}

		// 是否加倍
		isDoubleStr := "否"
		if user.AddNum == 2 {
			isDoubleStr = "是"
		}

		// 是否超级加倍
		isSuperDoubleStr := "否"
		if user.AddNum == 4 {
			isSuperDoubleStr = "是"
		}

		game.Table.WriteLogs(user.ID, role+"ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
			" 手牌： "+cardsStr+
			" 是否是地主： "+isDizhuStr+
			" 地主倍数： "+fmt.Sprintf(`%d`, game.CurRobNum)+"倍 "+
			" 是否加倍： "+isDoubleStr+
			" 是否超级加倍： "+isSuperDoubleStr)
	}
}

// RecordLogTwo 记录第二部分日志
func (game *DouDizhu) RecordLogTwo() {
	// 记录出牌日志
	game.Table.WriteLogs(0, "出牌日志: "+game.PutCardsLog)
}

// RecordLogThree 记录第三部分日志
func (game *DouDizhu) RecordLogThree() {

	for _, user := range game.UserList {
		// 角色
		role := user.GetSysRole()

		// 获取手牌字符串
		leftCardsStr := poker.CardsToString(poker.ReverseSortCards(user.Cards))

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

		game.Table.WriteLogs(user.ID, role+" ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
			" 角色： "+role+
			" 生效作弊率： "+fmt.Sprintf(`%d`, effectProb)+
			" 获取作弊率： "+fmt.Sprintf(`%d`, getProb)+
			" 作弊率来源： "+probSource+
			" 余牌： "+leftCardsStr+
			" 开始金额： "+score.GetScoreStr(user.InitAmount)+
			" 输赢金额： "+score.GetScoreStr(user.SettleResult)+
			" 剩余金额： "+score.GetScoreStr(user.User.GetScore()))
	}
}

// PaoMaDeng 跑马灯
func (game *DouDizhu) PaoMaDeng(Gold int64, userInter player.PlayerInterface, special string) {
	configs := game.Table.GetMarqueeConfig()

	//for _, conf := range configs {
	//
	//	if len(conf.SpecialCondition) > 0 && len(special) > 0 {
	//		log.Tracef("跑马灯有特殊条件 : %s", conf.SpecialCondition)
	//
	//		specialIndex := game.GetSpecialSlogan(special)
	//		specialArr := strings.Split(conf.SpecialCondition, ",")
	//		for _, specialStr := range specialArr {
	//			specialInt, err := strconv.Atoi(specialStr)
	//			if err != nil {
	//				log.Errorf("解析跑马灯特殊条件出错 : %s", conf.SpecialCondition)
	//				continue
	//			}
	//
	//			// 金额与特殊条件同时满足
	//			if specialInt == specialIndex && Gold >= conf.AmountLimit {
	//				err := game.table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
	//				if err != nil {
	//					log.Errorf("创建跑马灯错误：%v", err)
	//				}
	//				return
	//			}
	//		}
	//	}
	//}

	// 未触发特殊条件
	for _, conf := range configs {
		if len(conf.SpecialCondition) > 0 {
			continue
		}

		// 只需要满足金钱条件
		if Gold >= conf.AmountLimit {
			err := game.Table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
			return
		}
	}
}
