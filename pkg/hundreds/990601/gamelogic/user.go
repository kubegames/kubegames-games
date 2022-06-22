package gamelogic

import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/kubegames/kubegames-games/pkg/slots/990601/config"
	bridanimal "github.com/kubegames/kubegames-games/pkg/slots/990601/msg"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

const (
	USER_CHAN_LENGTH   int = 20     // 记录玩家最近n局的输赢金币
	USER_PAGE_LIMIT    int = 10     // 用户页数限制条数
	BET_DOWN_SCORE         = 101901 // 下注下分
	SETTLE_UP_SCORE        = 201901 // 结算上分
	FREE_GAME_UP_SCORE     = 201902 // 免费游戏上分
)

type UserInfo struct {
	User player.PlayerInterface

	betMux      sync.RWMutex
	BetInfo     [BET_AREA_LENGTH]int64
	BetInfoTemp [BET_AREA_LENGTH]int64 // 玩家下注 n 毫秒之前的缓存
	Totol       int64
	g           *Game

	NotBetNum int32 // 未投注次数

	WinGold int64      // 获胜总金币（近20局）// 以此排序
	BetGold int64      // 下注总金币（近20局）
	WinChan chan int64 // 近20局的输赢情况
	BetChan chan int64 // 近20局的下注情况

	BetGoldNow  int64 // 当前局下注金额总计
	LastWinGold int64 // 最后一局赢得金币

	currGold int64
	taxGold  int64

	betMsg chan *bridanimal.UserBet
}

func (u *UserInfo) Bet(pb *bridanimal.UserBet, TableBet [BET_AREA_LENGTH]int64, betMinLimit int64) bool {

	betGold := int64(u.g.BetArr[pb.BetIndex])

	var TotalBet int64
	for i := 0; i < BET_AREA_LENGTH; i++ {
		TotalBet += TableBet[i]
	}

	if pb.BetIndex < 0 {
		return false
	}

	// 下注总金额
	TotalBetAmount := int64(u.g.BetArr[pb.BetIndex%int32(len(u.g.BetArr))])
	//g个人单区域下注总金额
	TotalUserSingleSpaceAmount := TotalBetAmount + u.BetInfo[pb.BetType%BET_AREA_LENGTH]
	//个人所有区域下注总金额
	TotalUserAllSpaceAmount := TotalBetAmount + u.BetGoldNow
	//单区域下注总金额
	TotalSingleSpaceAmount := TotalBetAmount + TableBet[pb.BetType%BET_AREA_LENGTH]
	//所有区域下注总金额
	TotalAllSpaceAmount := TotalBet + TotalBetAmount
	//个人总区域下注总金额
	// 账户总金额
	TotalAmount := u.User.GetScore()
	// 判断是否可以下注
	// 主要判断总金额与下注金额的关系
	level := int(u.g.Table.GetLevel())

	u.betMux.Lock()
	defer u.betMux.Unlock()

	if TotalAmount <= 0 {
		u.SendBetFail("您余额不足，请充值！")
		return false
	} else if TotalBetAmount > TotalAmount {
		//log.Tracef("用户余额为：%v", TotalAmount)
		u.SendBetFail("您余额不足，请充值！")
		return false
	} else if betMinLimit != 0 && TotalAmount < betMinLimit {
		u.SendBetFail(fmt.Sprintf("至少携带%d金币才能下注！", betMinLimit/100))
		return false
	} else if config.BirdAnimaConfig.Singleusersinglespacelimit5times[level-1][pb.BetType%BET_AREA_LENGTH] != 0 && TotalUserSingleSpaceAmount > config.BirdAnimaConfig.Singleusersinglespacelimit5times[level-1][pb.BetType%BET_AREA_LENGTH] {
		//个人玩家单区域限红
		u.SendBetFail("您已达到该区域的下注额度限制！")
		return false
	} else if config.BirdAnimaConfig.Singleuserallspacelimit5times[level-1] != 0 && TotalUserAllSpaceAmount > config.BirdAnimaConfig.Singleuserallspacelimit5times[level-1] {
		// 判断和初始设置的个人限制 单人玩家所有区域限红
		u.SendBetFail("您已达到该房间的下注额度限制！")
		return false
	} else if config.BirdAnimaConfig.Allusersinglespacelimit5times[level-1][pb.BetType%BET_AREA_LENGTH] != 0 && TotalSingleSpaceAmount > config.BirdAnimaConfig.Allusersinglespacelimit5times[level-1][pb.BetType%BET_AREA_LENGTH] {
		//所有玩家单区域限红
		u.SendBetFail("该区域的下注已经达到总额度限制！")
		return false
	} else if config.BirdAnimaConfig.Allspacelimit5times[level-1] != 0 && TotalAllSpaceAmount > config.BirdAnimaConfig.Allspacelimit5times[level-1] {
		//所有玩家总区域限红
		u.SendBetFail("该房间的下注已经达到总额度限制！")
		return false
	}

	u.BetInfo[pb.BetType] += betGold
	u.Totol += betGold
	u.BetGoldNow += betGold

	u.User.SetScore(u.g.Table.GetGameNum(), -int64(u.g.BetArr[pb.BetIndex]), u.g.Table.GetRoomRate())
	u.NotBetNum = 0
	return true
}

func (u *UserInfo) SendBetFail(FailStr string) {
	data := new(bridanimal.BetFailMsg)
	data.BetFailInfo = FailStr
	u.User.SendMsg(int32(bridanimal.SendToClientMessageType_BetFailID), data)
}

func (u *UserInfo) setChip() {
	abs := math.Abs(float64(u.BetInfo[BIRD_BET_INDEX] - u.BetInfo[ANIMAL_BET_INDEX]))
	u.User.SendChip(u.BetGoldNow - u.BetInfo[BIRD_BET_INDEX] - u.BetInfo[ANIMAL_BET_INDEX] + int64(abs))
}

// 重置用户的下注信息
func (u *UserInfo) ResetUserData() {
	u.BetInfo = [12]int64{}
	u.BetInfoTemp = [12]int64{}
	u.Totol = 0
	u.BetGoldNow = 0
	u.LastWinGold = 0
	u.currGold = u.User.GetScore()
}

// 同步用户赢取的金额
func (u *UserInfo) SyncDataWin(winGold int64) {
	u.WinGold += winGold
	if len(u.WinChan) == USER_CHAN_LENGTH {
		temp := <-u.WinChan
		u.WinGold -= temp
	}
	u.WinChan <- winGold
}

// 同步用户下注金额
func (u *UserInfo) SyncDataBet(betGold int64) {
	u.BetGold += betGold
	if len(u.BetChan) == USER_CHAN_LENGTH {
		temp := <-u.BetChan
		u.BetGold -= temp
	}
	u.BetChan <- betGold
}

func (u UserInfo) GetAllBetGold() int64 {
	var val int64
	for _, v := range u.BetInfo {
		val += v
	}
	return val
}

type UserInfos []*UserInfo

// 从大到小排序
func (u UserInfos) Less(i, j int) bool {
	if u[i].WinGold >= u[j].WinGold {
		return true
	}
	return false
}

func (u UserInfos) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u UserInfos) Len() int {
	return len(u)
}

func (u UserInfos) Sort() {
	sort.Sort(u)
}
