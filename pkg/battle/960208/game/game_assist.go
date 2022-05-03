package game

import (
	"strconv"
	"strings"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// KickCheck 踢人检测
func (game *ThreeDoll) KickCheck() {
	for _, user := range game.UserList {
		if user.Status == int32(msg.UserStatus_Ready) {
			game.UserLeaveGame(user.User)
			game.Table.KickOut(user.User)
		}
	}

	if game.Table.PlayerCount() <= 0 {
		game.Table.Close()
	}
}

// KickRobot 检测剩余机器人
func (game *ThreeDoll) CheckLeftRobot() {
	var robotCount int
	for _, user := range game.UserList {

		if user.User.IsRobot() {
			robotCount++
		}
	}

	// 房间只剩下机器人，踢开
	if robotCount == len(game.UserList) {
		// 停掉所有定时器
		if game.RobotTimer != nil {
			game.Table.DeleteJob(game.RobotTimer)
		}
		if game.TimerJob != nil {
			game.Table.DeleteJob(game.TimerJob)
		}

		// 踢开说有机器人
		for _, user := range game.UserList {
			game.UserLeaveGame(user.User)
			game.Table.KickOut(user.User)
		}

		if game.Table.PlayerCount() <= 0 {
			game.Table.Close()
		}
	}
}

// MatchRobot 匹配机器人
func (game *ThreeDoll) MatchRobot() {

	// 定义桌子期望人数，确定加入机器人策略
	numWeight := rand.RandInt(1, 101)
	var limit int

	for index, rate := range game.GameCfg.NumberRate {
		if numWeight <= limit+rate && numWeight > limit {
			game.ExpectNum = index + 3
			break
		}
		limit += rate
	}

	// 坐下机器人人数
	sitNum := game.ExpectNum - len(game.UserList)

	// 桌子上人数已经达到预期
	if sitNum <= 0 {
		return
	}

	// 定时坐下机器人
	for i := 0; i < sitNum; i++ {
		randTime := rand.RandInt(1, game.TimeCfg.CountDown-999)
		_, _ = game.Table.AddTimer(int64(randTime), game.RobotSit)
	}
}

// RobotSit 机器人坐下
func (game *ThreeDoll) RobotSit() {

	// 倒计时最后一秒不匹配
	if game.Status == int32(msg.GameStatus_CountDown) && (game.TimerJob != nil && game.TimerJob.GetTimeDifference() < 1000) {
		return
	}

	// 游戏已经离开匹配状态，停止机器人坐下
	if game.Status > int32(msg.GameStatus_CountDown) {
		return
	}

	// 桌子上的人数已经满足预期坐下人数
	if len(game.UserList) >= game.GameCfg.NumberRate[len(game.GameCfg.NumberRate)-1] {
		return
	}

	err := game.Table.GetRobot(1, game.Table.GetConfig().RobotMinBalance, game.Table.GetConfig().RobotMaxBalance)
	if err != nil {
		log.Errorf("生成机器人失败：%v", err)
	}
}

// PaoMaDeng 跑马灯
func (game *ThreeDoll) PaoMaDeng(Gold int64, userInter player.PlayerInterface, special string) {
	configs := game.Table.GetMarqueeConfig()

	for _, conf := range configs {

		if len(conf.SpecialCondition) > 0 && len(special) > 0 {
			log.Tracef("跑马灯有特殊条件 : %s", conf.SpecialCondition)

			specialIndex := game.GetSpecialSlogan(special)
			specialArr := strings.Split(conf.SpecialCondition, ",")
			for _, specialStr := range specialArr {
				specialInt, err := strconv.Atoi(specialStr)
				if err != nil {
					log.Errorf("解析跑马灯特殊条件出错 : %s", conf.SpecialCondition)
					continue
				}

				// 金额与特殊条件同时满足
				if specialInt == specialIndex && Gold >= conf.AmountLimit {
					err := game.Table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
					if err != nil {
						log.Errorf("创建跑马灯错误：%v", err)
					}
					return
				}
			}
		}
	}

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

// GetSpecialSlogan 获取跑马灯触发特殊条件下标
func (game *ThreeDoll) GetSpecialSlogan(special string) int {
	switch special {
	case "爆玖":
		return 1
	case "炸弹":
		return 2
	case "三公":
		return 3
	default:
		return 0
	}
}

// SettleDivision 结算上下分
func (game *ThreeDoll) SettleDivision(userID int64) int64 {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("用户 %d 上下分查找不到当前用户")
		return 0
	}

	// 结果
	result := user.CurAmount - user.InitAmount

	profit := game.UserList[userID].User.SetScore(game.Table.GetGameNum(), result, game.RoomCfg.TaxRate)

	// 计算打码量
	chip := profit
	if profit > 0 {
		chip = game.RoomCfg.RoomCost
	}
	game.SetChip(userID, chip)

	// 有扣税操作，更新当前金额
	if profit > 0 {
		game.UserList[userID].CurAmount = user.InitAmount + profit
	}

	return profit
}

// SetGameRecords 创建战绩
func (game *ThreeDoll) SetGameRecords(resultList []*msg.SettleResult) {

	for _, settleResult := range resultList {
		Record := data.GameRecord{
			Time:      0,
			GameNum:   game.Table.GetGameNum(),
			RoomLevel: game.Table.GetLevel(),
			Result:    settleResult.Result,
			Status:    1,
		}

		user := game.UserList[settleResult.UserId]

		gameRecords := data.GetUserTabledata(user.User)

		// 保证记录最多只有30条
		if len(gameRecords) >= 30 {
			gameRecords = gameRecords[len(gameRecords)-29:]
		}

		gameRecords = append(gameRecords, Record)

		data.SetUserTabledata(user.User, gameRecords)
	}

}

// SetChip 设置码量
func (game *ThreeDoll) SetChip(userID int64, chip int64) {
	if chip < 0 {
		chip = -chip
	}
	game.UserList[userID].User.SendChip(chip)
}

// SetExitPermit 设置用户退出权限
func (game *ThreeDoll) SetExitPermit(permit bool) {
	for id := range game.UserList {
		game.UserList[id].ExitPermit = permit
	}
}

// UserSendRecord 发送战绩，计算产出
func (game *ThreeDoll) TableSendRecord(userID int64, result int64, netProfit int64) *platform.PlayerRecord {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("发送战绩查询用户 %d 失败", userID)
		return nil
	}

	if user.User.IsRobot() {
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

	//	user.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, drawAmount, outputAmount, endCards)
	return &platform.PlayerRecord{
		PlayerID:     uint32(user.User.GetID()),
		GameNum:      game.Table.GetGameNum(),
		ProfitAmount: profitAmount,
		BetsAmount:   betsAmount,
		DrawAmount:   drawAmount,
		OutputAmount: outputAmount,
		Balance:      user.User.GetScore(),
		UpdatedAt:    time.Now(),
		CreatedAt:    time.Now(),
	}
}

// GetUserByChairID 通过座位id获取用户
func (game *ThreeDoll) GetUserByChairID(chairID int32) (user *data.User) {
	for _, v := range game.UserList {
		if chairID == v.ChairID {
			user = v
			break
		}
	}

	return user
}

type SettleResult struct {
	TheorySettle int64 // 理论结算值
	ActualSettle int64 // 实际结算值
	CurAmount    int64 // 携带金额
}

// FillWinnerAmount 折算多余赢家金额，补足应赢金额小于携带金额但是 按比例补足会触发防一小博大，则补到携带金额大小
// leftAmount 剩余多赢金额
// leftTheoryWinSum 剩余理论赢钱合值
// WinnerList 赢家列表
func FillWinnerAmount(leftAmount *int64, leftTheoryWinSum *int64, WinnerList map[int64]*SettleResult) {

	newLeftAmount := *leftAmount
	// 按比例折扣 会 触发防止一小博大机制，则先补足金额
	for userID, v := range WinnerList {

		if v.ActualSettle >= v.CurAmount {
			continue
		}

		// 按比例折算金额
		convertAmount := *leftAmount * v.TheorySettle / *leftTheoryWinSum

		if convertAmount+v.ActualSettle > v.CurAmount {
			*leftAmount -= v.CurAmount - v.ActualSettle
			WinnerList[userID].ActualSettle = v.CurAmount
			*leftTheoryWinSum -= v.TheorySettle
			break
		}
	}

	// 剩余钱无变化，不需要再补足，跳出循环
	if newLeftAmount != *leftAmount {
		FillWinnerAmount(leftAmount, leftTheoryWinSum, WinnerList)
	}
}

// ConvertWinnerAmount 按比例折算多余赢家金额
func ConvertWinnerAmount(leftAmount int64, leftTheoryWinSum int64, WinnerList map[int64]*SettleResult) {
	if leftAmount <= 0 || leftTheoryWinSum <= 0 {
		log.Errorf("按比例折算多余赢家金额出现错误，剩余金额：%d，剩余赢钱理论合值：%d", leftAmount, leftTheoryWinSum)
		return
	}

	var (
		convertCount int   // 需要补足多余金额玩家个数
		winCounter   int   // 赢家计数器
		winAcc       int64 // 赢钱累加器
	)

	for _, v := range WinnerList {
		if v.ActualSettle < v.CurAmount {
			convertCount++
		}
	}

	// 按比例折扣
	for userID, v := range WinnerList {

		if v.ActualSettle < v.CurAmount {

			// 最后一个需要补足多余金额到赢家
			if convertCount-winCounter == 1 {
				WinnerList[userID].ActualSettle += leftAmount - winAcc
				break
			}

			// 按比例折算金额
			convertAmount := leftAmount * v.TheorySettle / leftTheoryWinSum
			WinnerList[userID].ActualSettle += convertAmount
			winAcc += convertAmount
			winCounter++
		}
	}
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
