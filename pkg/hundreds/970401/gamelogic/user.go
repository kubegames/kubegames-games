package gamelogic

import (
	"fmt"
	"game_LaBa/benzbmw/config"
	proto "game_LaBa/benzbmw/msg"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type User struct {
	user player.PlayerInterface
	game *Game

	WinTimesChan chan int64 // 近20局赢得纪录
	WinTimes     int64      //      近20局的赢得次数
	WinChan      chan int64 // 近20局的赢纪录
	WinGold      int64      // 近20局的赢总金币

	BetInfo     [BET_AREA_LENGHT]int64 // 下注区的详细
	WinInfo     [BET_AREA_LENGHT]int64 // 赢的下注区的详细
	NotBetCount int                    // 未下注次数

	BetGoldNow int64 // 当前局下注金额总计

	LastWinGold int64 // 最后一局赢得金币

	currGold int64 // 当前金币
	taxGold  int64 // 当前税
}

func NewUser(game *Game, user player.PlayerInterface) *User {
	return &User{
		game:         game,
		user:         user,
		WinTimesChan: make(chan int64, RECORD_LENGHT),
		WinChan:      make(chan int64, RECORD_LENGHT),
		BetInfo:      [BET_AREA_LENGHT]int64{},
		currGold:     user.GetScore(),
	}
}

// 初始化user
func (u User) Init(game *Game) {
	u.game = game

}

// 下注
func (u *User) DoBet(msg *proto.UserBet, TableBet [BET_AREA_LENGHT]int64, betMinLimit int64) bool {
	betGold := u.game.BetArr[msg.BetIndex]

	var TotalBet int64
	for i := 0; i < BET_AREA_LENGHT; i++ {
		TotalBet += TableBet[i]
	}

	if msg.BetIndex < 0 {
		return false
	}

	// 下注总金额
	TotalBetAmount := int64(u.game.BetArr[msg.BetIndex%int32(len(u.game.BetArr))])
	//g个人单区域下注总金额
	TotalUserSingleSpaceAmount := TotalBetAmount + u.BetInfo[msg.BetType%BET_AREA_LENGHT]
	//个人所有区域下注总金额
	TotalUserAllSpaceAmount := TotalBetAmount + u.BetGoldNow
	//单区域下注总金额
	TotalSingleSpaceAmount := TotalBetAmount + TableBet[msg.BetType%BET_AREA_LENGHT]
	//所有区域下注总金额
	TotalAllSpaceAmount := TotalBet + TotalBetAmount
	//个人总区域下注总金额
	// 账户总金额
	TotalAmount := u.user.GetScore()
	// 判断是否可以下注
	// 主要判断总金额与下注金额的关系
	level := int(u.game.table.GetLevel())

	if TotalAmount <= 0 {
		u.SendBetFailed("您余额不足，请充值！")
		return false
	} else if TotalBetAmount > TotalAmount {
		//log.Tracef("用户余额为：%v", TotalAmount)
		u.SendBetFailed("您余额不足，请充值！")
		return false
	} else if betMinLimit != 0 && TotalAmount < betMinLimit {
		u.SendBetFailed(fmt.Sprintf("至少携带%d金币才能下注！", betMinLimit/100))
		return false
	} else if config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] != 0 && TotalUserSingleSpaceAmount > config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] {
		//个人玩家单区域限红
		u.SendBetFailed("您已达到该区域的下注额度限制！")
		return false
	} else if config.BenzBMWConf.Singleuserallspacelimit5times[level-1] != 0 && TotalUserAllSpaceAmount > config.BenzBMWConf.Singleuserallspacelimit5times[level-1] {
		// 判断和初始设置的个人限制 单人玩家所有区域限红
		u.SendBetFailed("您已达到该房间的下注额度限制！")
		return false
	} else if config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] != 0 && TotalSingleSpaceAmount > config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] {
		//所有玩家单区域限红
		u.SendBetFailed("该区域的下注已经达到总额度限制！")
		return false
	} else if config.BenzBMWConf.Allspacelimit5times[level-1] != 0 && TotalAllSpaceAmount > config.BenzBMWConf.Allspacelimit5times[level-1] {
		//所有玩家总区域限红
		u.SendBetFailed("该房间的下注已经达到总额度限制！")
		return false
	}

	u.BetGoldNow += betGold
	u.user.SetScore(u.game.table.GetGameNum(), -1*betGold, 0)
	return true
}

func (u *User) SendBetFailed(reason string) {
	msgfailed := new(proto.BetFailMsg)
	msgfailed.BetFailInfo = reason
	u.user.SendMsg(int32(proto.SendToClientMessageType_BetFailID), msgfailed)
}

func (u *User) Reset() {
	u.BetInfo = [BET_AREA_LENGHT]int64{}
	u.WinInfo = [BET_AREA_LENGHT]int64{}
	u.LastWinGold = 0
	u.BetGoldNow = 0
	u.currGold = u.user.GetScore()
}

func (u *User) SyncWinData(winGold int64) {
	u.WinGold += winGold
	if len(u.WinChan) == RECORD_LENGHT {
		u.WinGold -= <-u.WinChan
	}
	u.WinChan <- winGold
}

func (u *User) SyncWinTimes(times int64) {
	if times != 0 && times != 1 {
		return
	}
	u.WinTimes += times
	if len(u.WinTimesChan) == RECORD_LENGHT {
		u.WinTimes -= <-u.WinTimesChan
	}
	u.WinTimesChan <- times
}

func (u *User) SendStatusMsg(duration int32) {
	msg := new(proto.StatusMessage)
	msg.Status = u.game.Status
	msg.StatusTime = duration
	u.user.SendMsg(int32(proto.SendToClientMessageType_Status), msg)
}

func (u *User) sendChip() {
	u.user.SendChip(u.BetGoldNow)
}
