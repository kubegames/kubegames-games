package game

import (
	"fmt"
	"game_poker/pai9/model"
	pai9 "game_poker/pai9/msg"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type User struct {
	user          player.PlayerInterface
	settle        []*pai9.SettleAllRespDetail // 玩家的结算信息
	chairID       int                         // 座位号
	QiangMulti    int32                       // 抢庄倍数
	BetMultiList  []int32                     // 下注倍数列表
	BetMulti      int32                       // 下注倍数
	WinGold       int64                       // 当前局赢的钱
	WinGoldActual int64                       // 实际应该赢/输的钱，用作防以小博大机制
	Cards         model.Cards                 // 当前局的牌

	HasQiang bool
	HasBet   bool

	testMsg *pai9.TestReqMsg // 作弊消息
	// BetMultiList []int32 // 下注倍数列表
}

func NewUser(user player.PlayerInterface) *User {
	return &User{
		user: user,
	}
}

// 一轮结束重置
func (user *User) Reset() {
	user.settle = nil
	user.Clear()
}

// 一句结束重置
func (user *User) Clear() {
	user.Cards = nil
	user.QiangMulti = 0
	user.BetMulti = 0
	user.WinGold = 0
	user.HasBet = false
	user.HasQiang = false
	user.testMsg = nil
	user.WinGoldActual = 0
	user.BetMultiList = nil
}

// 为用户结算
func (user *User) Settle(goldBegin int64, bottom int32, isZhuang bool) {
	typ, _ := user.Cards.CalcType()
	user.settle = append(user.settle, &pai9.SettleAllRespDetail{
		Name:       user.user.GetNike(),
		GoldBegin:  goldBegin,
		PokerType:  int32(typ),
		Bottom:     bottom,
		QiangMulti: user.QiangMulti,
		BetMulti:   user.BetMulti,
		GoldChange: user.WinGoldActual,
		IsZhuang:   isZhuang,
	})
}

func (user *User) GetSettle() []*pai9.SettleAllRespDetail {
	return user.settle
}

func (user *User) GetInfo(chairID int32) *pai9.UserInfo {
	return &pai9.UserInfo{
		ChairID: chairID,
		Avatar:  user.user.GetHead(),
		Name:    user.user.GetNike(),
		GoldNow: user.user.GetScore(),
		UserID:  user.user.GetId(),
		IP:      user.user.GetIp(),
	}
}

// 游戏结束时结算
func (user *User) EndSettle(gameNum string, tax int64, bottom int64, setNum int) {
	before := user.settle[setNum-1].GoldChange
	output, _ := user.user.SetScore(gameNum, before, tax)

	bet := int64(user.settle[setNum-1].BetMulti)
	user.user.SendRecord(gameNum,
		user.user.GetScore()-user.settle[setNum-1].GoldBegin,
		bet*bottom,
		before-output,
		output,
		"",
	)
	user.user.SendChip(bet)
}

func (user *User) GetUser() player.PlayerInterface {
	return user.user
}

func (user *User) calcWinGold(win int64) {
	user.WinGold = win
	shouldWin := win
	if win < 0 {
		shouldWin = -1 * win
	}

	// 不触发防以小博大机制
	user.WinGoldActual = win

	// 触发闲家的防以小博大机制
	if shouldWin > user.user.GetScore() {
		if win > 0 {
			// 赢钱
			user.WinGoldActual = user.user.GetScore()
		} else {
			// 输钱
			user.WinGoldActual = -1 * user.user.GetScore()
		}
	}
}

// 计算作弊的牌型
func (user *User) calcTestCard(cards model.Cards) {
	if user.testMsg == nil {
		user.Cards = cards
		return
	}

	// 写入作弊消息的牌型
	user.Cards = nil
	for _, card := range model.CardsAllType {
		if user.testMsg.Poker1 == card.Sorted {
			user.Cards = append(user.Cards, card.Copy())
		}
		if user.testMsg.Poker2 == card.Sorted {
			user.Cards = append(user.Cards, card.Copy())
		}
	}
	if len(user.Cards) != 2 {
		panic(user.Cards)
	}
	fmt.Println("配牌之后的牌 =====", user.Cards)
}

func (user *User) handleTestMsg(msg *pai9.TestReqMsg) {
	user.testMsg = msg
}

// type ZhuangUserSort []*User

// func (z ZhuangUserSort) Less(i, j int) bool { return z[i].QiangMulti > z[j].QiangMulti }

// func (z ZhuangUserSort) Swap(i, j int) { z[i], z[j] = z[j], z[i] }

// func (z ZhuangUserSort) Len() int { return len(z) }

type QiangSort struct {
	ChairID    int32
	QiangMulti int32
}

type QiangSorts []QiangSort

func (q QiangSorts) Less(i, j int) bool { return q[i].QiangMulti > q[j].QiangMulti }

func (q QiangSorts) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q QiangSorts) Len() int { return len(q) }
