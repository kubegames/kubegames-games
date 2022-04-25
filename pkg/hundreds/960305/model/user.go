package model

import (
	"fmt"
	"go-game-sdk/example/game_poker/960305/config"
	baijiale "go-game-sdk/example/game_poker/960305/msg"
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Table table.TableInterface
	User  player.PlayerInterface

	betMux sync.Mutex

	SceneChairId int                 // 在 场景中玩家的位置
	AllBet       int64               //从开始到现在的总下注
	TotalBet     int64               // 总下注金额
	BetArea      [5]int64            //每个区域的下注
	WinCount     int                 //赢的统计
	RetCount     []bool              //结果统计
	BetCount     []int64             //下注统计
	RetWin       int                 //20局赢统计
	RetBet       int64               //20局下注统计
	NoBetCount   int                 //为下注统计
	Rule         *config.RoomRules   //房间规则
	SettleMsg    *baijiale.SettleMsg //结算消息

	WinGold     int64      // 近20局赢得金额
	WinGoldChan chan int64 // 近20局赢得channel

	LastWinGold int64 // 最后一局赢得金额（大赢家排序）
	Title       int   // 头衔

	CurrGold   int64 // 当前金额
	CruenSorce int64 //记录用户金币

	InTime int64 // 进入时间
}

// 发送详细的押注失败消息
func SendBetFailMessage(FailMessageDetail string, user *User) {
	SendBetFailMessage := new(baijiale.BetFailMessage)
	SendBetFailMessage.FailMessage = FailMessageDetail

	//log.Tracef("SendBetFailMessage %s", FailMessageDetail)

	err := user.User.SendMsg(int32(baijiale.SendToClientMessageType_BetFailID), SendBetFailMessage)
	if err != nil {
		log.Tracef("SendBetFailMessage Error, %s, %s", FailMessageDetail, err.Error())
		return
	}
}

func (user *User) SendBetSuccessMessage(bet *baijiale.Bet, zhuangMiSeatID, xianMiSeatID int) {
	SendSuccessMessage := new(baijiale.BetSuccessMessage)
	SendSuccessMessage.BetIndex = bet.BetIndex
	SendSuccessMessage.BetType = bet.BetType
	SendSuccessMessage.SeatId = int32(user.SceneChairId)
	SendSuccessMessage.UserID = int64(user.User.GetID())

	SendSuccessMessage.ZhuangBetGold = user.BetArea[0]
	SendSuccessMessage.XianBetGold = user.BetArea[1]
	// if user.SceneChairId == zhuangMiSeatID {
	// 	   SendSuccessMessage.XianBetGold = -1
	// } else if user.SceneChairId == xianMiSeatID {
	// 	   SendSuccessMessage.ZhuangBetGold = -1
	// }
	user.Table.Broadcast(int32(baijiale.SendToClientMessageType_BetSuccessMessageID), SendSuccessMessage)

}

// 下注
func (user *User) Bet(bet *baijiale.Bet, betGold int64, TableBet [5]int64, betMinLimit int64) bool {
	var TotalBet int64
	for i := range TableBet {
		TotalBet += TableBet[i]
	}
	if bet.BetIndex < 0 {
		return false
	}

	// 下注总金额
	TotalBetAmount := betGold
	//g个人单区域下注总金额
	TotalUserSingleSpaceAmount := TotalBetAmount + user.BetArea[bet.BetType%5]
	//个人所有区域下注总金额
	TotalUserAllSpaceAmount := TotalBetAmount + user.TotalBet
	//单区域下注总金额
	TotalSingleSpaceAmount := TotalBetAmount + TableBet[bet.BetType%5]
	//所有区域下注总金额
	TotalAllSpaceAmount := TotalBet + TotalBetAmount
	//个人总区域下注总金额
	// 账户总金额
	TotalAmount := user.User.GetScore()
	// 判断是否可以下注
	// 主要判断总金额与下注金额的关系
	level := int(user.Table.GetLevel())

	user.betMux.Lock()
	defer user.betMux.Unlock()

	if TotalAmount <= 0 || TotalBetAmount > TotalAmount {
		//log.Tracef("用户余额为：%v", TotalAmount)
		SendBetFailMessage("您余额不足，请充值！", user)
		return false
	} else if betMinLimit != 0 && TotalAmount < betMinLimit {
		str := fmt.Sprintf("至少携带%d金币才能下注！", config.LongHuConfig.Betmin/100)
		SendBetFailMessage(str, user)
		return false
	} else if config.LongHuConfig.Singleusersinglespacelimit5times[level-1][bet.BetType%5] != 0 && TotalUserSingleSpaceAmount > config.LongHuConfig.Singleusersinglespacelimit5times[level-1][bet.BetType%5] {
		//个人玩家单区域限红
		SendBetFailMessage("您已达到该区域的下注额度限制！", user)
		return false
	} else if config.LongHuConfig.Singleuserallspacelimit5times[level-1] != 0 && TotalUserAllSpaceAmount > config.LongHuConfig.Singleuserallspacelimit5times[level-1] {
		// 判断和初始设置的个人限制 单人玩家所有区域限红
		SendBetFailMessage("您已达到该房间的下注额度限制！", user)
		return false
	} else if config.LongHuConfig.Allusersinglespacelimit5times[level-1][bet.BetType%5] != 0 && TotalSingleSpaceAmount > config.LongHuConfig.Allusersinglespacelimit5times[level-1][bet.BetType%5] {
		//所有玩家单区域限红
		SendBetFailMessage("该区域的下注已经达到总额度限制！", user)
		return false
	} else if config.LongHuConfig.Allspacelimit5times[level-1] != 0 && TotalAllSpaceAmount > config.LongHuConfig.Allspacelimit5times[level-1] {
		//所有玩家总区域限红
		SendBetFailMessage("该房间的下注已经达到总额度限制！", user)
		return false
	} else {
		// 下注成功
		user.TotalBet += TotalBetAmount
		user.BetArea[bet.BetType] += TotalBetAmount

		user.AllBet += TotalBetAmount
		user.NoBetCount = 0

		return true
	}
}

func (user *User) ResetUserData() {
	for i := 0; i < 5; i++ {
		user.BetArea[i] = 0
	}

	// user.WinGold = 0
	user.CruenSorce = user.User.GetScore()
	user.TotalBet = 0
}

func (u *User) SyncWinGold(win int64) {
	u.WinGold += win
	if len(u.WinGoldChan) == 20 {
		u.WinGold -= <-u.WinGoldChan
	}
	u.WinGoldChan <- win
}

//玩家数据统计
func (u *User) UserCount(bWin bool) {
	if bWin {
		u.WinCount++
	}

	u.RetCount = append(u.RetCount, bWin)
	u.BetCount = append(u.BetCount, u.TotalBet)

	if len(u.RetCount) > 20 {
		u.RetCount = append(u.RetCount[:0], u.RetCount[1:]...)
		u.BetCount = append(u.BetCount[:0], u.BetCount[1:]...)
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
}

// 大赢家排序（单局赢得最多）
type BigwinnerUser []*User

func (b BigwinnerUser) Less(i, j int) bool { return b[i].LastWinGold > b[j].LastWinGold }

func (b BigwinnerUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b BigwinnerUser) Len() int { return len(b) }

// 大富豪排序(20局赢得最多)
type RegalUser []*User

func (b RegalUser) Less(i, j int) bool { return b[i].WinGold > b[j].WinGold }

func (b RegalUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b RegalUser) Len() int { return len(b) }

// 神算子排序
type MasterUser []*User

func (b MasterUser) Less(i, j int) bool { return b[i].RetWin > b[j].RetWin }

func (b MasterUser) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b MasterUser) Len() int { return len(b) }

type TopUser []*User

func (t TopUser) Less(i, j int) bool {
	// if t[i].Icon > t[j].Icon {
	// 	return true
	// } else if t[i].Icon < t[j].Icon {
	// 	return false
	// } else {
	// 	return t[i].WinCount > t[j].WinCount
	// }
	return t[i].WinGold > t[j].WinGold
}

func (t TopUser) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

func (t TopUser) Len() int { return len(t) }

type LeftUser []*User

func (t LeftUser) Less(i, j int) bool {
	if t[i].WinGold > t[j].WinGold {
		return true
	} else if t[i].WinGold < t[j].WinGold {
		return false
	} else {
		return t[i].InTime < t[j].InTime
	}
}

func (t LeftUser) Swap(i, j int) { t[i], t[j] = t[j], t[i] }

func (t LeftUser) Len() int { return len(t) }
