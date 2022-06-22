package gamelogic

import (
	"fmt"
	"sync"

	"github.com/kubegames/kubegames-games/pkg/slots/970501/config"
	proto "github.com/kubegames/kubegames-games/pkg/slots/970501/msg"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type User struct {
	user player.PlayerInterface
	game *Game

	betMux sync.Mutex

	BetChan chan int64 // 近20局赢得纪录
	BetGold int64      //      近20局的赢得次数
	WinChan chan int64 // 近20局的赢纪录
	WinGold int64      // 近20局的赢总金币

	BetInfo [BET_AREA_LENGHT]int64 // 下注区的详细

	NotBetCount int   // 未下注次数
	BetGoldNow  int64 // 当前局下注金额总计
	LastWinGold int64 // 最后一局赢得金币

	taxGold int64 // 当前税
}

func NewUser(game *Game, user player.PlayerInterface) *User {
	return &User{
		game:    game,
		user:    user,
		BetChan: make(chan int64, RECORD_LENGHT),
		WinChan: make(chan int64, RECORD_LENGHT),
		BetInfo: [BET_AREA_LENGHT]int64{},
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

	u.betMux.Lock()
	defer u.betMux.Unlock()

	if TotalAmount <= 0 {
		u.BetFailed("您余额不足，请充值！", false, false)
		return false

		// else if config.BenzBMWConf.Betmin != 0 && TotalAmount < config.BenzBMWConf.Betmin {
		// 	//log.Tracef("用户余额为：%v", TotalAmount)
		// 	// u.BetFailed("您余额不足，请充值！", false, false)
		// 	u.BetFailed("至少携带50金币才能下注！", false, false)
		// 	return false
		// }
	} else if TotalBetAmount > TotalAmount {
		u.BetFailed("您余额不足，请充值！", false, false)
		return false
	} else if betMinLimit != 0 && TotalAmount < betMinLimit {
		u.BetFailed(fmt.Sprintf("至少携带%d金币才能下注！", betMinLimit/100), false, false)
		return false
	} else if config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] != 0 && TotalUserSingleSpaceAmount > config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] {
		//个人玩家单区域限红
		u.BetFailed("您已达到该区域的下注额度限制！", !(TotalUserSingleSpaceAmount == config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT]),
			u.BetGoldNow == config.BenzBMWConf.Singleusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT])
		return false
	} else if config.BenzBMWConf.Singleuserallspacelimit5times[level-1] != 0 && TotalUserAllSpaceAmount > config.BenzBMWConf.Singleuserallspacelimit5times[level-1] {
		// 判断和初始设置的个人限制 单人玩家所有区域限红
		u.BetFailed("您已达到该房间的下注额度限制！", !(TotalUserAllSpaceAmount == config.BenzBMWConf.Singleuserallspacelimit5times[level-1]),
			u.BetGoldNow == config.BenzBMWConf.Singleuserallspacelimit5times[level-1])
		return false
	} else if config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] != 0 && TotalSingleSpaceAmount > config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT] {
		//所有玩家单区域限红
		u.BetFailed("该区域的下注已经达到总额度限制！", !(TotalSingleSpaceAmount == config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT]),
			u.BetGoldNow == config.BenzBMWConf.Allusersinglespacelimit5times[level-1][msg.BetType%BET_AREA_LENGHT])
		return false
	} else if config.BenzBMWConf.Allspacelimit5times[level-1] != 0 && TotalAllSpaceAmount > config.BenzBMWConf.Allspacelimit5times[level-1] {
		//所有玩家总区域限红
		u.BetFailed("该房间的下注已经达到总额度限制！", !(TotalAllSpaceAmount == config.BenzBMWConf.Allspacelimit5times[level-1]),
			(u.BetGoldNow == config.BenzBMWConf.Allspacelimit5times[level-1]))
		return false
	}

	u.BetGoldNow += betGold
	u.BetInfo[msg.BetType] += betGold
	u.user.SetScore(u.game.table.GetGameNum(), -1*betGold, u.game.table.GetRoomRate())
	return true
}

func (u *User) BetFailed(reason string, isNeedDown, isMan bool) {
	msgfailed := new(proto.BetFailMsg)
	msgfailed.BetFailInfo = reason
	msgfailed.IsNeedDown = isNeedDown
	msgfailed.IsMan = isMan
	u.user.SendMsg(int32(proto.SendToClientMessageType_BetFailID), msgfailed)
}

func (u *User) Reset() {
	u.BetInfo = [BET_AREA_LENGHT]int64{}
	u.LastWinGold = 0
	u.BetGoldNow = 0
}

func (u *User) SyncWinData(winGold int64) {
	u.WinGold += winGold
	if len(u.WinChan) == RECORD_LENGHT {
		u.WinGold -= <-u.WinChan
	}
	u.WinChan <- winGold
}

func (u *User) SyncBetGold(betGold int64) {
	u.BetGold += betGold
	if len(u.BetChan) == RECORD_LENGHT {
		u.BetGold -= <-u.BetChan
	}
	u.BetChan <- betGold
}

func (u *User) SendStatusMsg(duration int32) {
	msg := new(proto.StatusMessage)
	msg.Status = u.game.Status
	msg.StatusTime = duration
	u.user.SendMsg(int32(proto.SendToClientMessageType_Status), msg)
}

func (u *User) SendChip() {
	u.user.SendChip(u.BetGoldNow)
}
