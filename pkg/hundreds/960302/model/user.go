package model

import (
	"common/log"
	"fmt"
	"game_poker/longhu/config"
	"math"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Table          table.TableInterface
	User           player.PlayerInterface
	SceneChairId   int               // 在 场景中玩家的位置
	AllBet         int64             //从开始到现在的总下注
	TotalBet       int64             // 总下注金额
	BetLong        int64             // 下注红方
	BetHu          int64             // 下注黑方
	BetHe          int64             // 下注幸运一击
	WinCount       int               //赢的统计
	RetCount       []bool            //结果统计
	BetCount       []int64           //下注统计
	RetWin         int               //20局赢统计
	RetBet         int64             //20局下注统计
	NoBetCount     int               //为下注统计
	Rule           *config.RoomRules //房间规则
	SettleMsg      *longhu.SettleMsg //结算消息
	RetWinMoneyArr []int64           //20局赢钱统计
	RetWinMoney    int64             //20局赢钱统计
	LastWinMoney   int64             //最后一局赢钱统计
	Winrata        int64             //近20局胜率
	Icon           int32             //用户称号 0无1神算子2大富豪3大富翁
	Time           int64             //用户入场时间
	CruenSorce     int64             //记录用户金币
}

// 发送详细的押注失败消息
func SendBetFailMessage(FailMessageDetail string, user *User) {
	SendBetFailMessage := new(longhu.BetFailMessage)
	SendBetFailMessage.FailMessage = FailMessageDetail

	//log.Tracef("SendBetFailMessage %s", FailMessageDetail)

	err := user.User.SendMsg(int32(longhu.SendToClientMessageType_BetFailID), SendBetFailMessage)
	if err != nil {
		log.Tracef("SendBetFailMessage Error, %s, %s", FailMessageDetail, err.Error())
		return
	}
}

func (user *User) SendBetSuccessMessage(bet *longhu.Bet) {
	SendSuccessMessage := new(longhu.BetSuccessMessage)
	SendSuccessMessage.BetIndex = bet.BetIndex
	SendSuccessMessage.BetType = bet.BetType
	SendSuccessMessage.SeatId = int32(user.SceneChairId)
	SendSuccessMessage.UserID = int64(user.User.GetId())

	user.Table.Broadcast(int32(longhu.SendToClientMessageType_BetSuccessMessageID), SendSuccessMessage)

}

const (
	LONG = 0
	HU   = 1
	HE   = 2
)

// 下注
func (user *User) Bet(bet *longhu.Bet, TableBet [3]int64) bool {
	//判断下注下标和下注区域下标是否超出列表
	if bet.BetIndex < 0 || bet.BetType >= 3 || bet.BetType < 0 || bet.BetIndex >= int32(len(user.Rule.BetList)) {
		SendBetFailMessage("数据异常", user)
		return false
	}
	var TotalBet int64
	for i := 0; i < 3; i++ {
		TotalBet += TableBet[i]
	}

	if bet.BetIndex < 0 {
		return false
	}

	tempSpceAmount := int64(0)
	// 下注总金额
	TotalBetAmount := int64(user.Rule.BetList[bet.BetIndex%int32(len(user.Rule.BetList))])
	//g个人单区域下注总金额
	switch bet.BetType % 3 {
	case LONG:
		tempSpceAmount = user.BetLong
	case HU:
		tempSpceAmount = user.BetHu
	case HE:
		tempSpceAmount = user.BetHe
	}
	TotalUserSingleSpaceAmount := TotalBetAmount + tempSpceAmount
	//个人所有区域下注总金额
	TotalUserAllSpaceAmount := TotalBetAmount + user.TotalBet
	//单区域下注总金额
	TotalSingleSpaceAmount := TotalBetAmount + TableBet[bet.BetType%3]
	//所有区域下注总金额
	TotalAllSpaceAmount := TotalBet + TotalBetAmount
	//个人总区域下注总金额
	// 账户总金额
	TotalAmount := user.User.GetScore()
	// 判断是否可以下注
	// 主要判断总金额与下注金额的关系

	if int64(user.Rule.BetMinLimit) > TotalAmount {
		str := fmt.Sprintf("至少携带%d金币才能下注！", user.Rule.BetMinLimit/100)
		SendBetFailMessage(str, user)
		return false
	} else if TotalBetAmount > TotalAmount {
		//log.Tracef("用户余额为：%v", TotalAmount)
		SendBetFailMessage("您余额不足，请充值！", user)
		return false
	} else if TotalUserAllSpaceAmount > user.Rule.UserBetLimit {
		// 判断和初始设置的个人限制 单人玩家所有区域限红
		SendBetFailMessage("您已达到该房间的下注额度限制！", user)
		return false
	} else if TotalUserSingleSpaceAmount > user.Rule.SingleUserSingleSpaceLimit[bet.BetType%3] {
		//个人玩家单区域限红
		SendBetFailMessage("您已达到该区域的下注额度限制！", user)
		return false
	} else if TotalSingleSpaceAmount > user.Rule.AllUserSingleSpaceLimit[bet.BetType%3] {
		//所有玩家单区域限红
		SendBetFailMessage("该区域的下注已经达到总额度限制！", user)
		return false
	} else if TotalAllSpaceAmount > user.Rule.AllSpaceLimit {
		//所有玩家总区域限红
		SendBetFailMessage("该房间的下注已经达到总额度限制！", user)
		return false
	} else {
		// 下注成功
		user.TotalBet += TotalBetAmount
		if bet.BetType == 0 {
			user.BetLong += TotalBetAmount
		} else if bet.BetType == 1 {
			user.BetHu += TotalBetAmount
		} else if bet.BetType == 2 {
			user.BetHe += TotalBetAmount
		}

		user.AllBet += TotalBetAmount
		user.SendBetSuccessMessage(bet)
		user.NoBetCount = 0
		return true
	}
}

func (user *User) ResetUserData() {
	user.BetHu = 0
	user.BetHe = 0
	user.BetLong = 0
	user.TotalBet = 0
	//用户当前值
	user.CruenSorce = user.User.GetScore()
}

//玩家数据统计
func (u *User) UserCount(bWin bool, currenwin int64) {
	if bWin {
		u.WinCount++
	}

	u.RetCount = append(u.RetCount, bWin)
	u.BetCount = append(u.BetCount, u.TotalBet)
	//赢钱统计
	u.RetWinMoneyArr = append(u.RetWinMoneyArr, currenwin)
	u.LastWinMoney = currenwin
	if len(u.RetCount) > 20 {
		u.RetCount = append(u.RetCount[:0], u.RetCount[1:]...)
		u.BetCount = append(u.BetCount[:0], u.BetCount[1:]...)
		u.RetWinMoneyArr = append(u.RetWinMoneyArr[:0], u.RetWinMoneyArr[1:]...)
	}

	u.RetWin = 0
	for _, v := range u.RetCount {
		if v {
			u.RetWin++
		}
	}

	u.RetBet = 0
	for _, v := range u.BetCount {
		u.RetBet += v
	}
	//20局总赢钱
	u.RetWinMoney = 0
	for _, v := range u.RetWinMoneyArr {
		u.RetWinMoney += v
	}
	u.Winrata = int64(math.Floor(float64(u.RetWin) / (float64(len(u.RetCount))) * 100))
}

type Usercount []*User

func (c Usercount) Len() int {
	return len(c)
}
func (c Usercount) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c Usercount) Less(i, j int) bool {
	if c[i].RetWinMoney > c[j].RetWinMoney {
		return true
	} else if c[i].RetWinMoney < c[j].RetWinMoney {
		return false
	} else {
		return c[i].Time < c[j].Time
	}

}

// 大赢家排序
type BigwinnerUser []*User

func (b BigwinnerUser) Less(i, j int) bool {
	if b[i].LastWinMoney > b[j].LastWinMoney {
		return true
	} else if b[i].LastWinMoney < b[j].LastWinMoney {
		return false
	} else {
		return b[i].Time < b[j].Time
	}
}

func (b BigwinnerUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b BigwinnerUser) Len() int { return len(b) }

// 大富豪排序
type RegalUser []*User

func (b RegalUser) Less(i, j int) bool {
	if b[i].RetWinMoney > b[j].RetWinMoney {
		return true
	} else if b[i].RetWinMoney < b[j].RetWinMoney {
		return false
	} else if b[i].Winrata > b[j].Winrata {
		return true
	} else if b[i].Winrata < b[j].Winrata {
		return false
	} else {
		return b[i].Time < b[j].Time
	}
}

func (b RegalUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b RegalUser) Len() int { return len(b) }

// 神算子排序
type MasterUser []*User

func (b MasterUser) Less(i, j int) bool {
	if b[i].Winrata > b[j].Winrata {
		return true
	} else if b[i].Winrata < b[j].Winrata {
		return false
	} else if b[i].RetWinMoney > b[j].RetWinMoney {
		return true
	} else if b[i].RetWinMoney < b[j].RetWinMoney {
		return false
	} else {
		return b[i].Time < b[j].Time
	}
}

func (b MasterUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b MasterUser) Len() int { return len(b) }
