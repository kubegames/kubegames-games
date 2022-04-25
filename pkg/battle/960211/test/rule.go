package main

import (
	"game_poker/pai9/config"
	"game_poker/pai9/model"
	"math/rand"
	"sort"
)

const (
	TABLE_NUM = 4
)

var (
	// 抢庄倍数列表
	QIANG_MULTI = []int{1, 2, 3, 4}
	BET_MULTI   []int // 下注倍数列表
)

type user struct {
	chairID  int
	cards    model.Cards
	qiang    int  // 抢庄倍数
	bet      int  // 下注倍数
	isRobot  bool // 是否是机器人
	isZhuang bool // 是否是庄家

	gold          int64 // 携带金额
	winGold       int64 // 理论输赢的钱
	winGoldActual int64 // 以小博大机制后赢得钱
	tax           int64 // 扣税

	paiIndex int // 第n大得牌
}

// 初始化一桌子人
func initUser() {
	userTable = make(map[int]*user, TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		userTable[i] = &user{
			chairID: i,
			gold:    Conf.BeginGold,
		}
	}
}

func (user *user) calcWinGold(win int64) {
	user.winGold = win
	shouldWin := win
	if win < 0 {
		shouldWin = -1 * win
	}

	// 不触发防以小博大机制
	user.winGoldActual = win

	// 触发闲家的防以小博大机制
	if shouldWin > user.gold {
		if win > 0 {
			// 赢钱
			user.winGoldActual = user.gold
		} else {
			// 输钱
			user.winGoldActual = -1 * user.gold
		}
	}
}

// 座位号：玩家
var userTable map[int]*user

type ZhuangUser []*user

func (u ZhuangUser) Less(i, j int) bool { return u[i].qiang > u[j].qiang }
func (u ZhuangUser) Swap(i, j int)      { u[i], u[j] = u[j], u[i] }
func (u ZhuangUser) Len() int           { return len(u) }

// 抢庄阶段
func qiang() {

	for _, v := range userTable {
		if v.isRobot {
			v.qiang = QIANG_MULTI[rand.Intn(len(QIANG_MULTI))]
		} else {
			v.qiang = QIANG_MULTI[rand.Intn(len(QIANG_MULTI))]
		}

	}
	// 计算谁是庄
	var wait []int

	var zhuanglist ZhuangUser
	for _, user := range userTable {
		zhuanglist = append(zhuanglist, user)
	}
	sort.Sort(zhuanglist)
	wait = []int{zhuanglist[0].chairID}
	for i := 1; i < len(zhuanglist); i++ {
		if zhuanglist[i].qiang == zhuanglist[0].qiang {
			wait = append(wait, zhuanglist[i].chairID)
		}
	}

	zhuangChairID := wait[rand.Intn(len(wait))]
	for _, v := range userTable {
		if v.chairID == zhuangChairID {
			v.isZhuang = true
		}
	}
}

// 下注阶段
func bet() {
	var zhuang *user
	for _, v := range userTable {
		if v.isZhuang {
			zhuang = v
			break
		}
	}
	// 计算下注倍数
	max := zhuang.gold / int64(TABLE_NUM-1) / int64(zhuang.qiang) / int64(Conf.Bottom)

	if max > 35 {
		max = 35
	} else if max < 5 {
		max = 5
	}

	for i := int64(1); i < 5; i++ {
		BET_MULTI = append(BET_MULTI, int(max/5*i))
	}
	BET_MULTI = append(BET_MULTI, int(max))

	for _, v := range userTable {
		if v.isZhuang {
			continue
		}
		if v.isRobot {
			v.bet = BET_MULTI[0]
			if Conf.IsBetRand {
				v.bet = BET_MULTI[rand.Intn(len(BET_MULTI))]
			}
		} else {
			v.bet = BET_MULTI[0]
			if Conf.IsBetRand {
				v.bet = BET_MULTI[rand.Intn(len(BET_MULTI))]
			}
		}
	}
}

// 发牌阶段
func dealpoker() {
	// 初始化一副牌
	deckPoker := model.Init(model.CardsAllType)
	deckPoker.Shuffle()

	pokers := make(model.CardsTable, 0, TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		pokers = append(pokers, deckPoker.DealPoker(2))
	}
	sort.Sort(pokers)

	roomProb := int32(Conf.RoomProb)
	// 真人玩家，机器人玩家
	var realUser, robotUser []int
	robotUser = append(robotUser, 2, 3, 4)
	realUser = append(realUser, 1)

	realProb := config.Pai9Config.PokerCtrl[roomProb]
	robotProb := config.Pai9Config.PokerCtrl[-1*roomProb]

	// 定义发牌顺序
	dealOrder := make([]int, 0, len(realUser)+len(robotUser))

CYCLE:
	realUserNum := len(realUser)
	robotUserNum := len(robotUser)
	allProb := int64(realUserNum)*realProb + int64(robotUserNum)*robotProb

	var randProb int64
	// 为0时，则realUserNum和robotUserNum均为0
	if allProb > 0 {
		randProb = rand.Int63n(allProb) + 1
	}

	for i, chairID := range realUser {
		if randProb <= realProb {
			dealOrder = append(dealOrder, chairID)
			realUser = append(realUser[:i], realUser[i+1:]...)
			goto CYCLE
		}
		randProb -= realProb
	}
	for i, chairID := range robotUser {
		if randProb <= robotProb {
			dealOrder = append(dealOrder, chairID)
			robotUser = append(robotUser[:i], robotUser[i+1:]...)
			goto CYCLE
		}
		randProb -= robotProb
	}

	if len(dealOrder) != TABLE_NUM {
		panic(dealOrder)
	}
	// fmt.Println("dealOrder = ", dealOrder)
	// 给玩家赋值牌
	for i, chairID := range dealOrder {
		userTable[chairID].cards = pokers[i]
		userTable[chairID].paiIndex = i + 1
		// val, typ := pokers[i].CalcType()
		// fmt.Printf("val %d  ; typ %v\n", val, typ)
	}

}

// 比牌
func compare() {

	var zhuang *user
	for _, v := range userTable {
		if v.isZhuang {
			zhuang = v
			break
		}
	}

	for _, user := range userTable {
		if user.isZhuang {
			continue
		}

		if zhuang.cards == nil {
			panic("庄的牌为空")
		}
		if user.cards == nil {
			panic("玩家的牌为空")
		}
		// fmt.Printf("zhuang.cards =%v       user.card = %v\n", zhuang.cards, user.cards)
		cmp := zhuang.cards.Compare(user.cards)

		// fmt.Printf("cmp  = %d  user.chairID = %d zhuang.qiang = %d  user.bet = %d \n", cmp, user.chairID, zhuang.qiang, user.bet)
		if cmp > 0 {
			// 庄赢，玩家输
			result := int64(Conf.Bottom) * int64(zhuang.qiang) * int64(user.bet)

			user.calcWinGold(-1 * result)
		} else if cmp < 0 {
			// 庄输，玩家赢
			result := int64(Conf.Bottom) * int64(zhuang.qiang) * int64(user.bet)
			user.calcWinGold(result)
		}
	}

	// 计算庄家应输/赢金额
	for _, user := range userTable {
		if user == zhuang {
			continue
		}
		zhuang.winGold -= user.winGold
		zhuang.winGoldActual -= user.winGoldActual
	}

	// 庄家的以小博大机制

	zhuangWait := zhuang.winGold
	if zhuang.winGold < 0 {
		zhuangWait *= -1
	}

	var (
		isTrigger bool  // 是否触发以小博大
		diff      int64 // 差值
	)

	if zhuangWait > zhuang.gold {
		// 触发以小博大机制
		if zhuang.winGold < 0 {
			zhuang.winGold = -1 * zhuang.gold
			diff = zhuang.winGoldActual - zhuang.winGold
		} else {
			zhuang.winGoldActual = zhuang.gold
			diff = zhuang.winGold - zhuang.winGoldActual
		}
		isTrigger = true
	}

	if zhuang.winGoldActual > 0 && isTrigger {
		// 庄家赢钱，且触发以小博大机制，重计算闲家的实际输钱
		var allLoss int64
		for _, user := range userTable {
			if user == zhuang {
				continue
			}
			if user.winGoldActual < 0 {
				allLoss += user.winGoldActual
			}
		}
		for _, user := range userTable {
			if user == zhuang {
				continue
			}
			if user.winGoldActual < 0 {
				user.winGoldActual = user.winGoldActual - diff*user.winGoldActual/allLoss
			}
		}
	} else if zhuang.winGoldActual < 0 && isTrigger {
		// 庄家输钱，且触发庄家以小博大机制，重计算闲家的实际赢钱
		var allWin int64
		for _, user := range userTable {
			if user == zhuang {
				continue
			}
			if user.winGoldActual > 0 {
				allWin += user.winGoldActual
			}
		}
		for _, user := range userTable {
			if user == zhuang {
				continue
			}
			if user.winGoldActual > 0 {
				user.winGoldActual = user.winGoldActual - diff*user.winGoldActual/allWin
			}
		}
	}

	for _, user := range userTable {
		// 赢钱才扣税
		if user.winGoldActual > 0 {
			user.tax = user.winGoldActual * int64(Conf.Tax) / 100
			user.winGoldActual = user.winGoldActual - user.tax
		}
		// fmt.Printf("玩家【%d】的金币变化为: %d\n", user.chairID, user.winGoldActual)
		// 结算玩家的钱
		user.gold += user.winGoldActual
	}

}

func reset() {
	BET_MULTI = nil
	for _, user := range userTable {
		user.qiang = 0
		user.bet = 0
		user.cards = nil
		user.isZhuang = false
		user.paiIndex = 0
		user.tax = 0
		user.winGold = 0
		user.winGoldActual = 0
	}
}

func check() bool {
	for _, user := range userTable {
		if user.gold < 100 {
			return false
		}
	}
	return true
}
