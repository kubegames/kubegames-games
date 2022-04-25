package game

import (
	"errors"
	"fmt"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
)

//游戏中倒计时触发相关操作
func (zjh *Game) procTimerEvent() {
	//defer log.Trace()
	//fmt.Println("zhixing procTimerEvent")
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		//fmt.Println("procTimerEvent 没在游戏中")
		//zjh.GameSignal <- global.GAME_SIGNAL_END
		//log.Tracef("结束")
		return
	}
	if zjh.CurActionUser == nil {
		//fmt.Println("zjh.CurActionUser 为 nil ")
		//zjh.GameSignal <- global.GAME_SIGNAL_END
		return
	}
	if zjh.CurActionUser.IsActioned {
		fmt.Println("zjh.CurActionUser 发言过")
		return
	}
	fmt.Println("时间到，玩家: ", zjh.CurActionUser.Id, "没发言，自动弃牌", time.Now())
	zjh.CurActionUser.IsActioned = true
	zjh.CurActionUser.IsTimeOut = true
	zjh.userActionGiveUp(zjh.CurActionUser)
	if zjh.IsSatisfyEnd() {
		zjh.EndGame(false)
		return
	}
	zjh.JudgeAndFinishCurRound()
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		//fmt.Println("游戏没在进行中，退出111")
		return
	}
	if zjh.IsSatisfyEnd() {
		zjh.EndGame(false)
		return
	}
	if zjh.CurStatus != global.TABLE_CUR_STATUS_ING {
		//////fmt.Println("游戏没在进行中，退出222")
		return
	}
	if !zjh.IsAllUserAllIn() {
		//go zjh.SetNextActionUser("")
		zjh.SetNextActionUser("procTimerEvent", "")
	} else {
		userOnlineList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
		zjh.CompareCards(nil, nil, userOnlineList)
		zjh.Table.AddTimer(2*1000, func() {
			zjh.EndGame(true)
		})
	}

}

//以下是用户操作

//弃牌
func (zjh *Game) userActionGiveUp(user *data.User) {
	if user == nil || user.IsLeave {
		////fmt.Println("userActionGiveUp user is nil ")
		return
	}
	if user.User == nil {
		////fmt.Println("userActionGiveUp user.User is nil")
		return
	}
	user.LoseReason = global.LOSE_REASON_GIVE_UP
	//返回给用户自己的牌型
	if !user.IsSawCards {
		selfMsg := &msg.S2CUserSeeCards{UserId: user.Id, CardType: int32(user.CardType), Cards: user.Cards}
		_ = user.User.SendMsg(global.S2C_GIVE_UP_CARDS_FOR_SELF, selfMsg)
		//user.Cli.WriteMessage(msg.PkgS2CMsg(global.S2C_GIVE_UP_CARDS_FOR_SELF, selfMsgB))
	}

	//更新用户当前的状态
	user.CurStatus = global.USER_CUR_STATUS_GIVE_UP
	user.IsActioned = true
	user.IsSawCards = true

	//将消息广播给房间内的其他用户
	zjh.Table.Broadcast(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_GIVE_UP,
		UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})

}

//看牌
func (zjh *Game) userActionSeeCards(user *data.User) []byte {
	user.IsSawCards = true
	s2cUserCards := &msg.S2CUserSeeCards{UserId: user.Id, Cards: user.Cards, CardType: int32(user.CardType)}
	_ = user.User.SendMsg(global.S2C_USER_SEE_CARDS, s2cUserCards)
	zjh.Table.Broadcast(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_SEE_CARDS, UserName: user.Name,
		CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
	return user.Cards
}

//全押
func (zjh *Game) userActionAllIn(user *data.User) int32 {
	if zjh.IsAllIn {
		return global.ERROR_CODE_CAN_NOT_ALL_IN
	}
	if zjh.Round < 3 {
		return global.ERROR_CODE_CAN_NOT_ALL_IN
	}
	minLimitAction := zjh.MinAction
	if user.IsSawCards {
		minLimitAction *= 2
	}
	//if user.User.GetScore() < minLimitAction {
	//	return global.ERROR_CODE_NOT_ENOUGH
	//}

	//选出当前资金最少的玩家进行allIn
	canAllIn := zjh.getCanAllInValue(user)

	zjh.CurActionIsSawCard = user.IsSawCards
	zjh.IsAllIn = true
	user.CurAmount += canAllIn
	zjh.PoolAmount += canAllIn
	minAction := canAllIn
	if user.IsSawCards {
		minAction /= 2
	}
	zjh.ChangeMinAction(minAction)
	zjh.Table.Broadcast(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_ALL_IN,
		Amount: canAllIn, UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
	//其他玩家要取消跟到底
	for _, v := range zjh.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		if v.IsFollowAllTheWay {
			v.IsFollowAllTheWay = false
		}
	}
	user.Amount += canAllIn
	user.Score -= canAllIn
	//if _, err := user.User.SetScore(user.Table.GetGameNum(), -canAllIn, "全押", zjh.Table.GetRoomRate(),
	//	0, global.SET_SCORE_BET, 100105); err != nil {
	//	log.Tracef("user all in set score err : ", err)
	//}
	zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户id："+fmt.Sprintf(`%d`, user.User.GetID())+" 全押 金额："+score.GetScoreStr(canAllIn)+" 余额："+score.GetScoreStr(user.Score))
	return global.ERROR_CODE_OK
}

//跟注
func (zjh *Game) userActionFollow(user *data.User) int32 {
	if user == nil {
		fmt.Println("user follow  nil ")
		return global.ERROR_CODE_NIL
	}
	if zjh.MinAction == 0 {
		fmt.Println("当前最低为0，不能跟注")
		return global.ERROR_CODE_ARG
	}
	amount := zjh.MinAction
	if user.IsSawCards {
		amount *= 2
		//因为全押在做 / 2 的时候会导致缺失1分钱，比如32.75这种，所以加1分钱上去
		if zjh.IsAllIn && user.Score-amount == 1 {
			amount += 1
		}
	}
	if user.Score < amount {
		fmt.Println("user.User.GetScore() < 111 amount111 ", user.User.GetID(), "  ", user.Score, "  ", amount)
		return global.ERROR_CODE_LESS_MIN
	}

	user.CurAmount += amount
	zjh.PoolAmount += amount
	user.Amount += amount
	user.Score -= amount
	//if _, err := user.User.SetScore(user.Table.GetGameNum(), -amount, "跟注", zjh.Table.GetRoomRate(),
	//	0, global.SET_SCORE_BET, 100102); err != nil {
	//	fmt.Println("user follow set score err : ", err)
	//}
	zjh.Table.Broadcast(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_FOLLOW,
		Amount: amount, UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
	zjh.Table.WriteLogs(user.User.GetID(), fmt.Sprintf(`%s用户id：%d 跟注 金额：%d 余额：%d `, aiRealStr(user.User.IsRobot()), user.User.GetID(), amount, user.Score))
	return global.ERROR_CODE_OK
}

//加注  amount 为配置的几个值 +2 +5 +10
func (zjh *Game) userActionRaise(user *data.User, amount int64) int32 {
	originAmount := amount
	if zjh.IsAllIn {
		fmt.Println(" is all in ")
		return global.ERROR_CODE_CAN_NOT_RAISE
	}
	if amount <= 0 {
		fmt.Println("amount <= 0 ", user.User.GetID(), " ", amount)
		return global.ERROR_CODE_ARG
	}
	//todo 判断加注金额只能是给定的三个档位
	fmt.Println("加注的三个档位：", zjh.GameConfig.RaiseAmount)
	if amount != zjh.GameConfig.RaiseAmount[0] && amount != zjh.GameConfig.RaiseAmount[1] && amount != zjh.GameConfig.RaiseAmount[2] {
		fmt.Println("加注金额不符合规则，", amount, zjh.GameConfig.RaiseAmount)
		return global.ERROR_CODE_ARG
	}

	//realAmount := amount
	if user.IsSawCards {
		amount *= 2
	}
	if amount <= zjh.MinAction {
		fmt.Println("amount <= zjh.MinAction ", user.User.GetID(), " ", amount, " ", zjh.MinAction)
		return global.ERROR_CODE_ARG
	}

	if user.Score < amount {
		fmt.Println("user.User.GetScore() < raise amount ", user.User.GetID(), " ", user.Score, " ", amount)
		return global.ERROR_CODE_NOT_ENOUGH
	}

	//如果加注金额大于当前最小玩家余额也不行
	maxRaiseAmount := zjh.getMaxRaiseAmount()
	if originAmount > maxRaiseAmount {
		fmt.Println("当前加注的最大金额：", maxRaiseAmount, "  ", originAmount)
		return global.ERROR_CODE_OVER_USER_MIN_AMOUNT
	}

	zjh.CurActionIsSawCard = user.IsSawCards

	user.CurAmount += amount
	zjh.PoolAmount += amount
	zjh.ChangeMinAction(originAmount)
	user.Amount += amount
	user.Score -= amount
	//if _, err := user.User.SetScore(user.Table.GetGameNum(), -amount, "加注", zjh.Table.GetRoomRate(),
	//	0, global.SET_SCORE_BET, 100103); err != nil {
	//	log.Tracef("user raise set score err : ", err)
	//}
	zjh.Table.Broadcast(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_RAISE,
		Amount: amount, UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
	zjh.Table.WriteLogs(user.User.GetID(), fmt.Sprintf("%s用户id：%d 加注 金额：%d  余额：%d  ", aiRealStr(user.User.IsRobot()), user.User.GetID(), amount, user.Score))
	return global.ERROR_CODE_OK
}

//取消跟到底
func (zjh *Game) userActionCancelFollowAllTheWay(user *data.User) int32 {
	user.IsFollowAllTheWay = false
	_ = user.User.SendMsg(global.S2C_USER_ACTION, &msg.S2CUserActionInfo{UserId: user.Id, Option: global.USER_OPTION_CANCEL_FOLLOW_ALL_THE_WAY,
		UserName: user.Name, CurAmount: user.CurAmount, PoolAmount: zjh.PoolAmount, MinAction: zjh.MinAction, Score: user.Score})
	return global.ERROR_CODE_OK
}

//获取指定状态玩家
func (zjh *Game) getStatusUser(status int) (user *data.User, err error) {
	for _, v := range zjh.userListArr {
		if v == nil {
			continue
		}
		if v.CurStatus == status {
			return v, nil
		}
	}
	return nil, errors.New("没有找到该状态玩家")
}

func (zjh *Game) ChangeMinAction(amount int64) {
	zjh.MinAction = amount
}

//设置牌桌玩家当前可allin的值
func (zjh *Game) setCanAllinValue(user *data.User) {
	canAllIn := zjh.getCanAllInValue(user)
	for _, v := range zjh.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		v.AllInAmount = canAllIn
		v.FollowAmount = zjh.MinAction
		if v.IsSawCards {
			v.FollowAmount *= 2
			if v.Score-v.AllInAmount == 1 {
				v.AllInAmount += 1
				v.FollowAmount += 1
			}
		}
		//fmt.Println("跟注值，玩家：",v.Id," 值：",v.FollowAmount)
		//fmt.Println("allin值，玩家：",v.Id," 值：",v.AllInAmount,"配置：",zjh.GameConfig.MaxAllIn,"advice: ",zjh.Table.GetAdviceConfig())
	}
}

//获取当前玩家可加注的最高值
func (zjh *Game) getMaxRaiseAmount() (minAmount int64) {
	userOnlineArr := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
	if len(userOnlineArr) <= 1 {
		return
	}
	minAmount = userOnlineArr[0].Score
	minUid := userOnlineArr[0].User.GetID()
	for _, v := range userOnlineArr {
		score := v.Score / 2
		if minAmount > score {
			minAmount = score
			minUid = v.User.GetID()
		}
	}
	//如果最小玩家是自己，同时自己又没看牌
	if (zjh.CurActionUser.User.GetID() == minUid && !zjh.CurActionUser.IsSawCards && zjh.Banker != nil && zjh.Banker.User.GetID() != zjh.CurActionUser.User.GetID()) || zjh.Round == 1 {
		minAmount *= 2
	}
	return
}

//获取当前玩家要全押的值
func (zjh *Game) getCanAllInValue(user *data.User) int64 {
	if zjh.IsAllIn {
		allInAmount := zjh.MinAction
		if user.IsSawCards {
			allInAmount *= 2
		}
		return allInAmount
	}
	var minAmount = user.Score
	for _, v := range zjh.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		tmpAmount := v.Score
		if minAmount > tmpAmount {
			minAmount = tmpAmount
		}
	}

	if minAmount > zjh.GameConfig.MaxAllIn {
		minAmount = zjh.GameConfig.MaxAllIn
	}
	if !user.IsSawCards {
		minAmount /= 2
	}

	//if minAmount < 0 {
	//	for _, v := range zjh.userListArr {
	//		if v != nil {
	//			fmt.Println("全押值小于0，panic掉,userId : ", v.Id, " userAmount : ", v.Score, "配置：", zjh.GameConfig.MaxAllIn)
	//		}
	//	}
	//	panic("getCanAllInValue")
	//}

	//fmt.Println("玩家可all in ： ",minAmount)

	return minAmount

}

//判断数组里面是否有该元素
func hasUser(user *data.User, userList []*data.User) bool {
	for _, v := range userList {
		if v.Id == user.Id {
			return true
		}
	}
	return false
}

func (zjh *Game) GetNextActionChair() (nextChair uint) {
	tmpChairId := zjh.CurActionUser.ChairId
	for i := 0; i <= 5; i++ {
		tmpChairId += 1
		if tmpChairId > 5 {
			tmpChairId = 1
		}
		chairUser := zjh.GetUserByChairId(tmpChairId)
		if chairUser == nil {
			////fmt.Println("设置下一个发言时没找到：", tmpChairId)
			continue
		}
		if chairUser.CurStatus != global.USER_CUR_STATUS_ING {
			////fmt.Println("下一个 chair ", chairUser.Id, " 状态 为： ", chairUser.CurStatus)
			continue
		}
		nextChair = tmpChairId
		break
	}
	return
}

//获取看牌人数
func (zjh *Game) getSawCount() (count int) {
	userList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
	for _, v := range userList {
		if v.IsSawCards {
			count++
		}
	}
	return
}

//获取看牌后加注人数
func (zjh *Game) getSawRaiseCount() (count int) {
	if zjh.MinAction > zjh.GameConfig.RaiseAmount[0] {
		userList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
		for _, v := range userList {
			if v.IsSawCards && v.IsActioned {
				count++
			}
		}
	}
	return
}

//获取前两个大牌的用户
func (zjh *Game) getMaxSecondUser() (biggestUser, secondUser *data.User, err error) {
	userList := zjh.GetStatusUserList(global.USER_CUR_STATUS_ING)
	if len(userList) == 0 {
		//fmt.Println("getMaxSecondUser wei 0 ")
		err = errors.New("getMaxSecondUser wei 0 ")
		return
	}
	biggestUser = poker.GetMaxUser(userList)[0]
	//biggestCard := biggestUser.Cards			//最大的牌
	remainUserList := make([]*data.User, 0)
	for _, v := range userList {
		if v.Id != biggestUser.Id {
			remainUserList = append(remainUserList, v)
		}
	}
	secondList := poker.GetMaxUser(remainUserList)
	if len(secondList) >= 1 {
		secondUser = secondList[0]
	}
	//第二大的牌
	//secondCard := secondUser.Cards
	return
}

//将最大牌、第二大牌按概率分配给各个用户
func (zjh *Game) cheatMaxSecond(cardsArr [][]byte, userCount int) {
	cardsIndex := 1 //当前分配的牌处于第几大
	//最大牌概率
	var totalMaxWinRate = 0
	aicheatRateMapMax := make(map[int]int)
	aicheatRateMapMax[-3000] = config.CheatConfigArr[0].MustWinRate
	aicheatRateMapMax[-2000] = config.CheatConfigArr[0].BigWinRate
	aicheatRateMapMax[-1000] = config.CheatConfigArr[0].SmallWinRate
	aicheatRateMapMax[3000] = config.CheatConfigArr[0].MustLoseRate
	aicheatRateMapMax[2000] = config.CheatConfigArr[0].BigLoseRate
	aicheatRateMapMax[1000] = config.CheatConfigArr[0].SmallLoseRate
	usercheatRateMapMax := make(map[int]int)
	usercheatRateMapMax[-3000] = config.CheatConfigArr[2].MustWinRate
	usercheatRateMapMax[-2000] = config.CheatConfigArr[2].BigWinRate
	usercheatRateMapMax[-1000] = config.CheatConfigArr[2].SmallWinRate
	usercheatRateMapMax[3000] = config.CheatConfigArr[2].MustLoseRate
	usercheatRateMapMax[2000] = config.CheatConfigArr[2].BigLoseRate
	usercheatRateMapMax[1000] = config.CheatConfigArr[2].SmallLoseRate
	for _, user := range zjh.userListArr {
		if user != nil {
			if user.User.IsRobot() {
				user.CheatRateMax = aicheatRateMapMax[user.CheatRate]
				totalMaxWinRate += aicheatRateMapMax[user.CheatRate]
			} else {
				user.CheatRateMax = usercheatRateMapMax[user.CheatRate]
				totalMaxWinRate += usercheatRateMapMax[user.CheatRate]
			}
			fmt.Println("玩家", user.Id, "作弊率：", user.CheatRate)
		}
	}
	if totalMaxWinRate < 10000 {
		maxUserCount := userCount
		for _, user := range zjh.userListArr {
			if user != nil && user.CheatRateMax == 0 {
				maxUserCount--
			}
		}
		if maxUserCount <= 0 {
			maxUserCount = 2
			fmt.Println("五个真实玩家？？？？？")
		}

		averageMax := (10000 - totalMaxWinRate) / maxUserCount
		for _, user := range zjh.userListArr {
			if user != nil && user.CheatRateMax != 0 {
				user.CheatRateMax += averageMax
			}
		}
		totalMaxWinRate = 10000
	}
	for _, user := range zjh.userListArr {
		if user != nil {
			fmt.Println(">>>>>>>>玩家：", user.Id, "最大牌概率：", user.CheatRateMax)
		}
	}
	//分配最大的牌
	isDisMax := false
	maxIndex := rand.RandInt(0, totalMaxWinRate)
	tmpMaxRate := 0
	for _, user := range zjh.userListArr {
		if user != nil && user.CheatRateMax > 0 {
			if maxIndex >= tmpMaxRate && maxIndex < tmpMaxRate+user.CheatRateMax {
				//该user放最大牌
				//fmt.Println("最大牌玩家：", user.Id)
				user.CardType, user.Cards = poker.GetCardTypeJH(cardsArr[0]) //最大的牌发给玩家
				user.CardEncode = poker.GetEncodeCard(user.CardType, user.Cards)
				cardsArr = append(cardsArr[:0], cardsArr[1:]...)
				user.CardIndexInTable = cardsIndex
				zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户: "+fmt.Sprintf(`%d`, user.User.GetID())+" 的牌为： "+poker.GetCardsCNName(user.Cards)+" 牌型："+poker.GetCardTypeCNName(user.CardType))
				cardsIndex++
				isDisMax = true
				break
			}
			tmpMaxRate += user.CheatRateMax
		}
		//fmt.Println("tmpSendCountRate  +aiSend.SendCountRate[i] ",tmpSendCountRate,tmpSendCountRate+aiSend.SendCountRate[i])

	}
	if !isDisMax {
		//panic("isDisMax := false")
	}

	//分配第二大的牌
	var totalSecondWinRate = 0
	aicheatRateMapSecond := make(map[int]int)
	aicheatRateMapSecond[-3000] = config.CheatConfigArr[1].MustWinRate
	aicheatRateMapSecond[-2000] = config.CheatConfigArr[1].BigWinRate
	aicheatRateMapSecond[-1000] = config.CheatConfigArr[1].SmallWinRate
	aicheatRateMapSecond[3000] = config.CheatConfigArr[1].MustLoseRate
	aicheatRateMapSecond[2000] = config.CheatConfigArr[1].BigLoseRate
	aicheatRateMapSecond[1000] = config.CheatConfigArr[1].SmallLoseRate
	usercheatRateMapSecond := make(map[int]int)
	usercheatRateMapSecond[-3000] = config.CheatConfigArr[3].MustWinRate
	usercheatRateMapSecond[-2000] = config.CheatConfigArr[3].BigWinRate
	usercheatRateMapSecond[-1000] = config.CheatConfigArr[3].SmallWinRate
	usercheatRateMapSecond[3000] = config.CheatConfigArr[3].MustLoseRate
	usercheatRateMapSecond[2000] = config.CheatConfigArr[3].BigLoseRate
	usercheatRateMapSecond[1000] = config.CheatConfigArr[3].SmallLoseRate
	for _, user := range zjh.userListArr {
		if user != nil && len(user.Cards) == 0 {
			if user.User.IsRobot() {
				user.CheatRateSecond = aicheatRateMapSecond[user.CheatRate]
				totalSecondWinRate += aicheatRateMapSecond[user.CheatRate]
			} else {
				user.CheatRateSecond = usercheatRateMapSecond[user.CheatRate]
				totalSecondWinRate += usercheatRateMapSecond[user.CheatRate]
			}

		}
	}
	if totalSecondWinRate < 10000 {
		averageSecond := (10000 - totalSecondWinRate) / (userCount - 1)
		for _, user := range zjh.userListArr {
			if user != nil && len(user.Cards) == 0 && user.CheatRateSecond != 0 {
				user.CheatRateSecond += averageSecond
				//fmt.Println("玩家：", user.Id, "第二大牌概率：", user.CheatRateSecond)
			}
		}
		totalSecondWinRate = 10000
	}
	//分配第二大的牌
	secondIndex := rand.RandInt(0, totalSecondWinRate)
	fmt.Println("second index : ", secondIndex, totalSecondWinRate)
	tmpsecondRate := 0
	isDissecond := false
	for _, user := range zjh.userListArr {
		if user != nil && len(user.Cards) == 0 && user.CheatRateSecond > 0 {
			//fmt.Println("tmpsecondRate ::::: ", tmpsecondRate, user.Id, user.Cards)
			if secondIndex >= tmpsecondRate && secondIndex <= tmpsecondRate+user.CheatRateSecond && len(user.Cards) == 0 {
				//该user放二大牌
				//fmt.Println("二大牌玩家：", user.Id)
				user.CardType, user.Cards = poker.GetCardTypeJH(cardsArr[0]) //最大的牌发给玩家
				user.CardEncode = poker.GetEncodeCard(user.CardType, user.Cards)
				cardsArr = append(cardsArr[:0], cardsArr[1:]...)
				user.CardIndexInTable = cardsIndex
				cardsIndex++
				isDissecond = true
				zjh.Table.WriteLogs(user.User.GetID(), aiRealStr(user.User.IsRobot())+"用户: "+fmt.Sprintf(`%d`, user.User.GetID())+" 的牌为： "+poker.GetCardsCNName(user.Cards)+" 牌型："+poker.GetCardTypeCNName(user.CardType))
				break
			}
			tmpsecondRate += user.CheatRateSecond
		}
		//fmt.Println("tmpSendCountRate  +aiSend.SendCountRate[i] ",tmpSendCountRate,tmpSendCountRate+aiSend.SendCountRate[i])

	}

	//给剩下的玩家发牌
	for _, v := range zjh.userListArr {
		if v != nil && len(v.Cards) == 0 {
			v.CardType, v.Cards = poker.GetCardTypeJH(cardsArr[0])
			v.CardEncode = poker.GetEncodeCard(v.CardType, v.Cards)
			zjh.Table.WriteLogs(v.User.GetID(), aiRealStr(v.User.IsRobot())+"用户: "+fmt.Sprintf(`%d`, v.User.GetID())+" 的牌为： "+poker.GetCardsCNName(v.Cards)+" 牌型："+poker.GetCardTypeCNName(v.CardType))
			cardsArr = append(cardsArr[:0], cardsArr[1:]...)
			v.CardIndexInTable = cardsIndex
			cardsIndex++
		}
	}

	if !isDissecond {
		fmt.Println("isDissecond := false ======================== +++++++++  ")
		for _, user := range zjh.userListArr {
			if user != nil {
				fmt.Println("玩家：", user.Id, "的牌：：：：：：", fmt.Sprintf(`%x`, user.Cards))
			}
		}
	}

	return
}
