package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// OpenPocketCard 翻开暗牌
func (game *Blackjack) OpenPocketCard() {
	PocketCard := game.HostCards.PocketCard
	if PocketCard != 0 {
		// 翻开暗牌
		game.HostCards.Cards = append(game.HostCards.Cards, PocketCard)
		game.HostCards.Point = poker.GetPoint(game.HostCards.Cards)
		game.HostCards.PocketCard = 0

		// 广播庄家暗牌
		game.SendPocketCard(PocketCard)
	}
}

// 比牌
func (game *Blackjack) ContrastCards(holdCards data.HoldCards, isBuyInsure bool) (result int64) {

	// 闲家手牌牌型为爆牌
	if holdCards.Type == msg.CardsType_Bust {
		result = -holdCards.BetAmount
		return
	}

	// 庄家和闲家手牌牌型 都是普通牌
	if game.HostCards.Type == holdCards.Type && game.HostCards.Type == msg.CardsType_Other {
		hostValue := poker.GetNearPoint21(game.HostCards.Point)
		holdValue := poker.GetNearPoint21(holdCards.Point)
		if hostValue > holdValue {

			// 闲家输
			result = -holdCards.BetAmount
		} else if hostValue < holdValue {

			// 闲家赢
			result = holdCards.BetAmount
		}
		return
	}

	// 庄家手牌牌型 大于 闲家手牌牌型 （牌型倒叙排序）
	if game.HostCards.Type < holdCards.Type {
		result = -holdCards.BetAmount

		// 庄家牌型为黑杰克，并且玩家买过保险
		if game.HostCards.Type == msg.CardsType_BlackJack && isBuyInsure {
			result = 0
		}
		return
	}

	// 庄家手牌牌型 小于 闲家手牌牌型 （牌型倒叙排序）
	if game.HostCards.Type > holdCards.Type {
		result = holdCards.BetAmount

		// 闲家五小龙赢，2赔3
		if holdCards.Type == msg.CardsType_FiveDragon {
			result = holdCards.BetAmount * 3 / 2
		}

		return
	}

	return
}

// CheckUserCmd 检查玩家操作
func (game *Blackjack) CheckUserCmd(cmd *msg.AskDoReq, curUser *CurUser) (isAllow bool) {
	isAllow = false
	switch cmd.CmdType {
	// 要牌
	case int32(msg.AskDoType_GetPoker):
		isAllow = curUser.GetPoker
		// 分牌
	case int32(msg.AskDoType_DepartPoker):
		isAllow = curUser.DepartPoker
		break
		// 双倍
	case int32(msg.AskDoType_DoubleBet):
		isAllow = curUser.DoubleBet
		break
		// 停牌
	case int32(msg.AskDoType_Stand):
		isAllow = curUser.Stand
		break
		// 认输
	case int32(msg.AskDoType_GiveUp):
		isAllow = curUser.GiveUp
		break
	}
	return
}

// SetNestCurUser 找到下一个可执行curUser
func (game *Blackjack) SetNestCurUser() {

	curUser := game.CurActionUser
	betIndex := curUser.BetCardsIndex

	user := game.UserList[curUser.UserID]

	// 下一个curUser
	nextUser := CurUser{}

	//log.Traceln(curUser.UserID)

	// 是否可继续操作
	if user.HoldCards[betIndex].ActionPermit {
		// 可继续操作，更新curUser

		nextUser = CurUser{
			ChairID:       curUser.ChairID,
			UserID:        curUser.UserID,
			BetCardsIndex: betIndex,
			GetPoker:      user.CheckGetPoker(betIndex),
			DepartPoker:   user.CheckDepartPoker(betIndex),
			DoubleBet:     user.CheckDoubleBet(betIndex),
			Stand:         user.CheckStand(betIndex),
			GiveUp:        user.CheckGiveUp(betIndex),
			TurnCounter:   curUser.TurnCounter,
		}

	} else {
		// 分牌，并且有第二副手牌,并且第二副牌可操作
		if betIndex == 0 && len(user.HoldCards[1].Cards) > 0 && user.HoldCards[1].ActionPermit {
			nextUser = CurUser{
				ChairID:       curUser.ChairID,
				UserID:        curUser.UserID,
				BetCardsIndex: 1,
				GetPoker:      user.CheckGetPoker(1),
				DepartPoker:   user.CheckDepartPoker(1),
				DoubleBet:     user.CheckDoubleBet(1),
				Stand:         user.CheckStand(1),
				GiveUp:        user.CheckGiveUp(1),
				TurnCounter:   curUser.TurnCounter,
			}
		} else {

			// 更新curUser
			game.AllUserList[curUser.UserID] = user
			game.UserList[curUser.UserID] = user

			for _, id := range game.BetSeats {
				user, ok := game.UserList[id]

				// 玩家可能提前离开
				if ok {
					if user.ChairID > curUser.ChairID && user.Status != int32(msg.UserStatus_UserStopAction) {
						nextUser = CurUser{
							ChairID:       user.ChairID,
							UserID:        user.ID,
							BetCardsIndex: 0,
							GetPoker:      user.CheckGetPoker(0),
							DepartPoker:   user.CheckDepartPoker(0),
							DoubleBet:     user.CheckDoubleBet(0),
							Stand:         user.CheckStand(0),
							GiveUp:        user.CheckGiveUp(0),
							TurnCounter:   curUser.TurnCounter,
						}
						break
					}
				}
			}

			// 找不到下一个闲家，轮到庄家
			if (CurUser{}) == nextUser {
				nextUser = CurUser{
					UserID:      -1,
					ChairID:     -1,
					TurnCounter: curUser.TurnCounter,
				}
			}

		}
	}

	// 轮转次数 + 1
	nextUser.TurnCounter++

	game.CurActionUser = &nextUser
	//log.Warnf("当前curUser：%v", game.CurActionUser)
	return
}

// KickRobot 检测剩余机器人
func (game *Blackjack) CheckLeftRobot() {
	var robotCount int
	for _, user := range game.AllUserList {

		if user.User.IsRobot() {
			robotCount++
		}
	}

	// 房间只剩下机器人，踢开
	if robotCount == len(game.AllUserList) {
		// 关闭所有定时器
		if game.TimerJob != nil {
			game.table.DeleteJob(game.TimerJob)
		}

		// 踢掉所有机器人
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

			//踢掉用户
			game.table.KickOut(user.User)
		}

		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_StartStatus)
		if game.TimerJob != nil {
			game.table.DeleteJob(game.TimerJob)
		}

		//关闭桌子
		game.table.Close()
	}
}

// MatchRobot 匹配机器人
func (game *Blackjack) MatchRobot() {

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
	sitNum := game.ExpectNum - len(game.AllUserList)

	// 桌子上人数已经达到预期
	if sitNum <= 0 {
		return
	}

	// 定时坐下机器人
	for i := 0; i < sitNum; i++ {
		randTime := rand.RandInt(1, game.timeCfg.StartMove-999)
		log.Tracef("添加机器人定时器%d", randTime)
		_, _ = game.table.AddTimer(int64(randTime), game.RobotSit)
	}
}

// RobotSit 机器人坐下
func (game *Blackjack) RobotSit() {

	// 倒计时最后一秒不匹配
	if game.Status == int32(msg.GameStatus_StartMove) && (game.TimerJob != nil && game.TimerJob.GetTimeDifference() < 1000) {
		return
	}

	// 游戏已经离开匹配状态，停止机器人坐下
	if game.Status != int32(msg.GameStatus_StartMove) {
		return
	}

	// 桌子上的人数已经满足预期坐下人数
	if len(game.AllUserList) >= game.ExpectNum {
		return
	}

	err := game.table.GetRobot(1, game.table.GetConfig().RobotMinBalance, game.table.GetConfig().RobotMaxBalance)
	if err != nil {
		log.Errorf("生成机器人失败：%v", err)
	}
}

// DeleteUser 删除用户
func (game *Blackjack) DeleteUser(userID int64) {
	delete(game.AllUserList, userID)
	delete(game.UserList, userID)

	// 让出座位
	for chairID, id := range game.Chairs {
		if id == userID {
			game.Chairs[chairID] = 0
			break
		}
	}

	// 移除押注序列
	for k, id := range game.BetSeats {
		if id == userID {
			game.BetSeats = append(game.BetSeats[:k], game.BetSeats[k+1:]...)
			break
		}
	}
}

// WriteUserLog 编辑用户日志
func (game *Blackjack) WriteUserLog(user *data.User, netProfit int64) {

	// 双倍牌次数
	DoubleCount := 0

	// 是否分牌
	isDepart := "否"

	// 手牌日志
	var cardsLog string
	for index, handCards := range user.HoldCards {
		if index == 1 {
			if len(handCards.Cards) == 0 {
				continue
			} else {
				isDepart = "是"
			}

		}

		cardsStr := poker.CardsToString(handCards.Cards)
		cardsTypeStr := poker.CardsTypeToString(handCards.Type)
		endTypeStr := data.EndTypeToString(handCards.EndType)

		cardsLog += "手牌" + fmt.Sprintf(`%d`, index+1) + "：" + cardsStr + " " +
			"牌型" + fmt.Sprintf(`%d`, index+1) + "：" + cardsTypeStr + " " +
			"点数" + fmt.Sprintf(`%d`, index+1) + "：" + fmt.Sprintf(`%d`, handCards.Point[0]) + " " +
			"结束方式" + fmt.Sprintf(`%d`, index+1) + "：" + endTypeStr + " "

		if handCards.StopAction == int32(msg.StopAction_ActionDoubleBet) {
			DoubleCount++
		}
	}

	// 是否买保险
	isBuyInsure := "否"
	if user.IsBuyInsure {
		isBuyInsure = "是"
	}

	// 作弊率来源
	probSource := ProbSourcePoint

	// 生效作弊值
	effectProb := user.User.GetProb()
	if effectProb == 0 {
		effectProb = game.table.GetRoomProb()

		probSource = ProbSourceRoom
	}

	// 获取作弊率
	getProb := effectProb

	if effectProb == 0 {
		effectProb = 1000
	}

	// 记录游戏日志
	game.table.WriteLogs(user.ID, " 用户ID： "+fmt.Sprintf(`%d`, user.User.GetID())+
		" 角色： "+user.GetSysRole()+
		" 手牌："+cardsLog+
		" 生效作弊率： "+fmt.Sprintf("%d", effectProb)+
		" 获取作弊率： "+fmt.Sprintf("%d", getProb)+
		" 作弊率来源： "+probSource+
		" 下注额度： "+fmt.Sprintf("%d", user.BetAmount)+
		" 是否买保险： "+isBuyInsure+
		" 是否分牌： "+isDepart+
		" 双倍牌次数： "+fmt.Sprintf("%d", DoubleCount)+
		" 输赢金额： "+fmt.Sprintf("%d", netProfit)+
		" 结束金额： "+fmt.Sprintf("%d", user.User.GetScore()))

}

// PaoMaDeng 跑马灯
func (game *Blackjack) PaoMaDeng(Gold int64, userInter player.PlayerInterface, special string) {
	configs := game.table.GetMarqueeConfig()

	for _, conf := range configs {

		if len(conf.SpecialCondition) > 0 && len(special) > 0 {
			log.Debugf("跑马灯有特殊条件 : %s", conf.SpecialCondition)

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
					err := game.table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
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
			err := game.table.CreateMarquee(userInter.GetNike(), Gold, special, conf.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
			return
		}
	}
}

// GetSpecialSlogan 获取跑马灯触发特殊条件下标
func (game *Blackjack) GetSpecialSlogan(special string) int {
	switch special {
	case "五小龙":
		return 1
	case "黑杰克":
		return 2
	default:
		return 0
	}
}

// SettleDivision 结算上下分
func (game *Blackjack) SettleDivision(userID int64) int64 {
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("用户 %d 上下分查找不到当前用户")
		return 0
	}
	profit := game.UserList[userID].User.SetScore(game.table.GetGameNum(), user.CurAmount-user.InitAmount, game.RoomCfg.TaxRate)

	// 有扣税操作，更新当前金额
	if profit > 0 {
		game.UserList[userID].CurAmount = user.InitAmount + profit
	}

	return profit
}

// SetChip 设置码量
func (game *Blackjack) SetChip(userID int64, chip int64) {
	if chip < 0 {
		chip = -chip
	}
	game.AllUserList[userID].User.SendChip(chip)
}

// UserSendRecord 发送战绩，计算产出（下注之后金额变化）
func (game *Blackjack) TableSendRecord(userID int64, result int64, netProfit int64) *platform.PlayerRecord {
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

	profitAmount = netProfit - user.Insurance
	betsAmount = user.BetAmount + user.Insurance
	outputAmount = netProfit + user.BetAmount

	if netProfit >= 0 {
		drawAmount = result - netProfit
	}

	return &platform.PlayerRecord{
		PlayerID:     uint32(userID),
		GameNum:      game.table.GetGameNum(),
		ProfitAmount: profitAmount,
		BetsAmount:   betsAmount,
		DrawAmount:   drawAmount,
		OutputAmount: outputAmount,
		//EndCards:     endCards,
	}
}
