package game

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	rand1 "github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/poker"
)

func (game *Game) StartGame() {

	// 游戏开始，进入摆牌状态
	game.Status = global.TABLE_CUR_STATUS_COMPOSE

	t1 := time.Now()
	if game.GetRoomUserCount() != 4 {
		fmt.Println("人数不为4 怎么就开赛了")
		return
	}
	//给用户发牌
	game.SendCards()

	for _, v := range game.userList {
		cardsSlice := make([]byte, 0)
		for _, v := range v.Cards {
			cardsSlice = append(cardsSlice, v)
		}
		log.Tracef("id : %d, 手牌 : %s", v.User.GetID(), poker.GetCardsCNName(cardsSlice))
	}

	// 记录实际作弊率值
	for i, user := range game.userList {

		// 实际作弊率
		var actualProb int32

		// 获取系统作弊率
		actualProb = game.Table.GetRoomProb()

		// 有点控用点控作弊率
		if user.User.GetProb() != 0 {
			actualProb = user.User.GetProb()
		}

		game.userList[i].Cheat = actualProb

	}

	//分牌
	game.Table.AddTimer(global.TIMER_START_GAME, func() {
		game.splitCardsForAllUser()
	})
	fmt.Println("开始游戏耗时：", time.Now().Sub(t1), game.Table.GetID())
}

//
func (game *Game) splitCardsForAllUser() {
	//给用户分出头中尾三墩
	wg := new(sync.WaitGroup)
	wg.Add(4)
	for _, user := range game.userList {
		go func(userGo *data.User) {
			userGo.SplitCards()
			wg.Done()
		}(user)
	}
	wg.Wait()
	//根据作弊率交换用户的牌
	game.CheatChangeCards()
	//for _,user := range game.userList {
	//	fmt.Println("作弊后玩家的牌 ： ", user.User.GetID(), fmt.Sprintf(`%x,%x,%x`, user.HeadCards, user.MiddleCards, user.TailCards))
	//}
	//fmt.Println("user cards : ",game.userList[0].HeadCards,game.userList[0].HeadCardType,game.userList[0].User.GetID())
	//发好牌通知用户开赛
	for _, user := range game.userList {
		//isbeat, tailType, midType := user.Compare5Cards(user.TailCards, user.MiddleCards)
		//user.TailCardType,user.MidCardType = tailType,midType

		_ = user.User.SendMsg(int32(msg.S2CMsgType_START_GAME), game.NewS2CStartGame(user))
	}

	// 18秒的摆牌倒计时 + 开始游戏动画时间http://192.168.0.46:5288/TenSanShuiScene/?ip=192.168.0.46:9020&gameId=960206
	game.timerJob, _ = game.Table.AddTimer(int64(28000), game.EndCompose)

	//定时结束比赛
	//game.timerJob, _ = game.Table.AddTimer(25*1000, func() {
	//	if game.Status == global.TABLE_CUR_STATUS_ING {
	//		//系统比牌
	//		fmt.Println("定时结束比赛", time.Now(), game.Table.GetID())
	//		game.SystemCompareCards()
	//		game.Table.AddTimer(3*1000, game.EndGame)
	//		//game.EndGame()
	//	}
	//})
}

// EndCompose 摆牌结束
func (game *Game) EndCompose() {

	// 未发送过摆牌请求的用户，系统代替摆牌
	for _, user := range game.userList {
		if !user.IsSettleCards {
			s2cSettleMsg := &msg.S2CSettleCards{
				Uid:         user.User.GetID(),
				ChairId:     user.ChairId,
				SpecialType: user.SpecialCardType,
			}
			game.Table.Broadcast(int32(msg.S2CMsgType_SETTLE_CARDS), s2cSettleMsg)
		}
	}

	game.Compare()
}

// CompareAndSettle 比牌结算流程
func (game *Game) Compare() {
	game.Status = global.TABLE_CUR_STATUS_COMPARE

	// 系统比牌
	game.SystemCompareCards()

	// 3秒后进入结算状态
	game.timerJob, _ = game.Table.AddTimer(int64(3000), game.Settle)
}

// Settle 结算
func (game *Game) Settle() {

	// 结算结算
	game.Status = global.TABLE_CUR_STATUS_SETTLE

	fmt.Println("日志-结束比赛：", time.Now())

	//结算上下分所有用户的输赢
	for i, user := range game.userList {

		//13水因为不知道用户下了多少注，所以就传房间底注
		netProfit, _ := user.User.SetScore(user.Table.GetGameNum(), user.FinalSettle, game.Table.GetRoomRate())

		// 跑马灯触发
		if netProfit > 0 {
			specialWords := ""
			if user.SpecialCardType >= data.SPECIAL_CARD_YTL {
				specialWords = data.GetSpecialCnName(user.SpecialCardType)
			}
			game.PaoMaDeng(netProfit, user.User, specialWords)
		}

		// 战绩
		outputAmount := netProfit

		// 投注
		betsAmount := game.Bottom

		// 抽税
		drawAmount := game.userList[i].FinalSettle - netProfit
		if netProfit < 0 {
			outputAmount = 0
			betsAmount = -netProfit
			drawAmount = 0
		}

		// 发送战绩
		user.User.SendRecord(game.Table.GetGameNum(), netProfit, betsAmount, drawAmount, outputAmount, "") //1月15新加的据结束，用户离开调用

		// 发送打码量
		user.User.SendChip(betsAmount)

		game.userList[i].FinalSettle = netProfit

		// 作弊率来源
		probSource := ProbSourcePoint
		if pointCrl := user.User.GetProb(); pointCrl == 0 {
			probSource = ProbSourceRoom
		}

		effectProb := user.Cheat
		if effectProb == 0 {
			effectProb = 1000
		}

		// 用户日志
		logStr := fmt.Sprintf(`用户id: %d, 角色: %s, 生效作弊率: %d, 获取作弊率: %d, 作弊率来源: %s, `, user.User.GetID(), user.GetSysRole(), effectProb, user.Cheat, probSource)

		// 牌型日志
		logStr += fmt.Sprintf(` 头墩: %s, 头墩牌型: %s, 中墩: %s, 中墩牌型: %s, 尾墩: %s, 尾墩牌型: %s, 特殊牌型: %s, `,
			poker.GetCardsCNName(user.HeadCards),
			poker.GetCard3TypeCNName(user.HeadCardType),
			poker.GetCardsCNName(user.MiddleCards),
			poker.GetCard5TypeCNName(user.MidCardType),
			poker.GetCardsCNName(user.TailCards),
			poker.GetCard5TypeCNName(user.TailCardType),
			data.GetSpecialCnName(user.SpecialCardType))

		// 输赢日志
		logStr += fmt.Sprintf(` 总输赢: %s, 剩余金额 %s `,
			score.GetScoreStr(netProfit),
			score.GetScoreStr(user.User.GetScore()))

		game.Table.WriteLogs(user.User.GetID(), logStr)
	}

	// 广播计算消息
	game.Table.Broadcast(int32(msg.S2CMsgType_END_GAME), game.NewS2CEndGame())

	// 结算状态6s
	game.timerJob, _ = game.Table.AddTimer(int64(6000), game.EndGame)
}

// EndGame 游戏结束
func (game *Game) EndGame() {
	game.Status = global.TABLE_CUR_STATUS_END
	//踢掉用户，重置牌桌
	game.ResetLogicTable()
	//框架结束比赛
	game.Table.EndGame()
}

//根据作弊率换牌
func (game *Game) CheatChangeCards() {
	uidScoreMap := make(map[int64]int)
	for i := 0; i < len(game.userList)-1; i++ {
		for j := i + 1; j < len(game.userList); j++ {
			//头中尾三墩都进行比牌 12 13 14 23 24 34
			if game.userList[i].SpecialCardType != data.SPECIAL_CARD_NO || game.userList[j].SpecialCardType != data.SPECIAL_CARD_NO {
				if game.userList[i].SpecialCardType > game.userList[j].SpecialCardType {
					uidScoreMap[game.userList[i].User.GetID()]++
					uidScoreMap[game.userList[j].User.GetID()]--
				} else if game.userList[i].SpecialCardType < game.userList[j].SpecialCardType {
					uidScoreMap[game.userList[j].User.GetID()]++
					uidScoreMap[game.userList[i].User.GetID()]--
				}
			} else {
				//普通牌型比较
				if res := game.userList[i].CompareHead(game.userList[j]); res == global.COMPARE_WIN {
					uidScoreMap[game.userList[i].User.GetID()]++
					uidScoreMap[game.userList[j].User.GetID()]--
				} else if res == global.COMPARE_LOSE {
					uidScoreMap[game.userList[j].User.GetID()]++
					uidScoreMap[game.userList[i].User.GetID()]--
				}
				//中墩
				if res := game.userList[i].CompareMid(game.userList[j]); res == global.COMPARE_WIN {
					uidScoreMap[game.userList[i].User.GetID()]++
					uidScoreMap[game.userList[j].User.GetID()]--
				} else if res == global.COMPARE_LOSE {
					uidScoreMap[game.userList[j].User.GetID()]++
					uidScoreMap[game.userList[i].User.GetID()]--
				}
				//尾墩
				if res := game.userList[i].CompareTail(game.userList[j]); res == global.COMPARE_WIN {
					uidScoreMap[game.userList[i].User.GetID()]++
					uidScoreMap[game.userList[j].User.GetID()]--
				} else if res == global.COMPARE_LOSE {
					uidScoreMap[game.userList[j].User.GetID()]++
					uidScoreMap[game.userList[i].User.GetID()]--
				}
			}
		}
	}

	// 持有手牌
	type holdCards struct {
		Cards           [13]byte
		HeadCards       []byte
		HeadCardType    int
		MiddleCards     []byte
		MidCardType     int
		TailCards       []byte
		TailCardType    int
		EncodeHead      int
		EncodeMid       int
		EncodeTail      int
		SpecialHead     int32
		SpecialMid      int32
		SpecialTail     int32
		SpecialCardType int32
		SpareArr        []*msg.S2CCardsAndCardType
	}

	type Pair struct {
		Key   int64
		Value int
	}
	var (
		pairList      []Pair      // 积分排序列表
		cardsSequence []holdCards // 手牌序列（按积分从小到大）
	)
	controlledCards := make(map[int64]holdCards)

	for k, v := range uidScoreMap {
		pairList = append(pairList, Pair{k, v})
	}

	// 新起一个切片，按照牌型得分从小到大排序
	sort.Slice(pairList, func(i, j int) bool {
		return pairList[i].Value < pairList[j].Value
	})

	for _, kv := range pairList {
		user := game.GetUserList(kv.Key)
		cardsSequence = append(cardsSequence, holdCards{
			Cards:           user.Cards,
			HeadCards:       user.HeadCards,
			HeadCardType:    user.HeadCardType,
			MiddleCards:     user.MiddleCards,
			MidCardType:     user.MidCardType,
			TailCards:       user.TailCards,
			TailCardType:    user.TailCardType,
			EncodeHead:      user.EncodeHead,
			EncodeMid:       user.EncodeMid,
			EncodeTail:      user.EncodeTail,
			SpecialHead:     user.SpecialHead,
			SpecialMid:      user.SpecialMid,
			SpecialTail:     user.SpecialTail,
			SpecialCardType: user.SpecialCardType,
			SpareArr:        user.SpareArr,
		})
	}

	if len(cardsSequence) != 4 {
		log.Warnf("牌组序列错误")
		return
	}

	// 最大牌，和第二大牌概率分布
	biggestCardsRatePlace := make(map[int64]int)
	secondCardsRatePlace := make(map[int64]int)

	// 确定最大牌 和 第二大牌 概率分布
	for _, user := range game.userList {

		userID := user.User.GetID()

		// 没有点控，用血池
		prob := user.User.GetProb()
		if prob == 0 {
			prob = game.Table.GetRoomProb()
		}

		// 机器人 和 玩家 采用 不同的大牌概率分布
		biggestRateDis := config.CheatConf.PlayerBiggestRate
		secondRateDis := config.CheatConf.PlayerSecondRate
		if user.User.IsRobot() {
			biggestRateDis = config.CheatConf.RobotBiggestRate
			secondRateDis = config.CheatConf.RobotSecondRate
		}

		log.Tracef("用户 %d 作弊率 %d", userID, prob)

		// 检测作弊率
		probIndex := game.checkProb(prob)
		if probIndex == -1 {
			log.Warnf("游戏 %d 错误的作弊率: %d", game.Table.GetID(), prob)
			// 默认 1000 作弊率的 索引
			probIndex = 2
		}

		// 拿牌概率值为0, 不参与大牌分配
		if biggestRateDis[probIndex] != 0 {
			biggestCardsRatePlace[userID] = biggestRateDis[probIndex]
		}

		if secondRateDis[probIndex] != 0 {
			secondCardsRatePlace[userID] = secondRateDis[probIndex]
		}

	}

	// 概率分配最大牌
	if len(biggestCardsRatePlace) != 0 {
		// 总概率值
		var totalRate int

		// 总概率值
		for _, rate := range biggestCardsRatePlace {
			totalRate += rate
		}

		// 未满10000的剩余概率值剩余平均概率
		lessAverageRate := (10000 - totalRate) / len(biggestCardsRatePlace)

		if lessAverageRate < 0 {
			lessAverageRate = 0
		}

		// 更新新概率值，让概率变得更加平缓
		for id, rate := range biggestCardsRatePlace {
			biggestCardsRatePlace[id] = lessAverageRate + rate
			totalRate += lessAverageRate
		}

		// 权重
		weight := rand1.RandInt(0, totalRate+1)

		// 概率累加值
		addRate := 0

		// 最大牌userID
		var biggestCardsUserID int64
		for id, rate := range biggestCardsRatePlace {

			if weight > addRate && weight <= addRate+rate {
				biggestCardsUserID = id
				break
			}
			addRate += rate
		}

		// 权重没有落在概率分布上
		if biggestCardsUserID == 0 {
			log.Errorf("游戏 %d 控牌最大牌 userID 为0，权重没有落在概率分布上", game.Table.GetID())
			return
		}

		log.Tracef("第一大牌用户 %d ", biggestCardsUserID)
		log.Tracef("牌堆有 %d 副", len(cardsSequence))

		// 分配最大牌
		controlledCards[biggestCardsUserID] = cardsSequence[len(cardsSequence)-1]
		cardsSequence = cardsSequence[:len(cardsSequence)-1]
		log.Tracef("牌堆有 %d 副", len(cardsSequence))

		// 第二大牌概率分布提出已拿最大牌的用户
		if _, ok := secondCardsRatePlace[biggestCardsUserID]; ok {
			delete(secondCardsRatePlace, biggestCardsUserID)
		}
	}

	// 概率分配第二大牌
	if len(secondCardsRatePlace) != 0 {
		// 重置总概率值
		var totalRate int

		// 计算拿第二打牌的总概率值
		for _, rate := range secondCardsRatePlace {
			totalRate += rate
		}

		// 剩余平均概率
		lessAverageRate := (10000 - totalRate) / len(secondCardsRatePlace)

		if lessAverageRate < 0 {
			lessAverageRate = 0
		}

		// 更新新概率值，让概率变得更加平缓
		for id, rate := range secondCardsRatePlace {
			secondCardsRatePlace[id] = lessAverageRate + rate
			totalRate += lessAverageRate
		}

		// 权重
		weight := rand1.RandInt(1, totalRate+1)

		// 概率累加值
		addRate := 0

		// 最二大牌userID
		var secondCardsUserID int64
		for id, rate := range secondCardsRatePlace {

			if weight > addRate && weight <= addRate+rate {
				secondCardsUserID = id
				break
			}
			addRate += rate
		}

		// 权重没有落在概率分布上
		if secondCardsUserID == 0 {
			log.Errorf("游戏 %d 控牌最二大牌 userID 为0，权重没有落在概率分布上", game.Table.GetID())
			return
		}

		log.Tracef("第二大牌用户 %d ", secondCardsUserID)
		// 分配最第二大牌
		controlledCards[secondCardsUserID] = cardsSequence[len(cardsSequence)-1]
		cardsSequence = cardsSequence[:len(cardsSequence)-1]
		log.Tracef("牌堆有 %d 副", len(cardsSequence))
	}

	// 分配剩余牌
	for _, holdCards := range cardsSequence {
		for _, user := range game.userList {
			if _, ok := controlledCards[user.User.GetID()]; !ok {
				controlledCards[user.User.GetID()] = holdCards
				break
			}

		}
	}

	for userID, holdCards := range controlledCards {
		for index, user := range game.userList {
			if userID == user.User.GetID() {
				user.Cards = holdCards.Cards
				user.HeadCards = holdCards.HeadCards
				user.HeadCardType = holdCards.HeadCardType
				user.MiddleCards = holdCards.MiddleCards
				user.MidCardType = holdCards.MidCardType
				user.TailCards = holdCards.TailCards
				user.TailCardType = holdCards.TailCardType
				user.EncodeHead = holdCards.EncodeHead
				user.EncodeMid = holdCards.EncodeMid
				user.EncodeTail = holdCards.TailCardType
				user.SpecialHead = holdCards.SpecialHead
				user.SpecialMid = holdCards.SpecialMid
				user.SpecialTail = holdCards.SpecialTail
				user.SpecialCardType = holdCards.SpecialCardType
				user.SpareArr = holdCards.SpareArr
				game.userList[index] = user
				break
			}
		}
	}

}

//系统比牌
func (game *Game) SystemCompareCards() {
	for k, v := range game.userList {
		game.userList[k].IsSettleCards = true
		fmt.Println("userList[i].IsSettleCards : ", v.User.GetID())
	}

	// del by wd in 2020.3.12 系统比牌游戏状态不做更改
	//game.Status = global.TABLE_CUR_STATUS_END

	for i := 0; i < len(game.userList)-1; i++ {
		for j := i + 1; j < len(game.userList); j++ {
			//头中尾三墩都进行比牌 12 13 14 23 24 34
			if game.userList[i].SpecialCardType != data.SPECIAL_CARD_NO || game.userList[j].SpecialCardType != data.SPECIAL_CARD_NO {
				game.userList[i].CompareSpecial(game.userList[j])
			} else {
				if hitRob := game.userList[i].CompareNormal(game.userList[j]); hitRob != nil {
					game.HitRobArr = append(game.HitRobArr, hitRob)
				}
			}
		}
	}
	//全垒打
	game.HomeRun()
	// 结算用户的总输赢
	game.SettleAccounts()
}

func (game *Game) HomeRun() {
	//再看是否全垒打
	//只要有个玩家有特殊牌都不触发全垒打
	for _, user := range game.userList {
		if user != nil && user.SpecialCardType > data.SPECIAL_CARD_NO {
			return
		}
	}

	uidCountMap := make(map[int64]int)
	for _, hitRob := range game.HitRobArr {
		uidCountMap[hitRob.HitUid]++
	}
	for homeRunUid, count := range uidCountMap {
		if count != 3 {
			continue
		}
		fmt.Println("全垒打id：", homeRunUid)
		//可以全垒打
		game.HomeRunUid = homeRunUid
		homeRunUser := game.GetUserList(homeRunUid)
		for _, v := range game.HitRobArr {
			if v.HitUid != homeRunUid {
				continue
			}
			beHitUser := game.GetUserList(v.BeHitUid)
			//fmt.Println("被全垒打玩家：",beHitUser.User.GetID(),v.HitHeadScore,v.HitMidScore,v.HitTailScore)
			beHitUser.HeadPlus -= int(v.HitHeadScore * 2)
			beHitUser.MidPlus -= int(v.HitMidScore * 2)
			beHitUser.TailPlus -= int(v.HitTailScore * 2)
			homeRunUser.HeadPlus += int(v.HitHeadScore * 2)
			homeRunUser.MidPlus += int(v.HitMidScore * 2)
			homeRunUser.TailPlus += int(v.HitTailScore * 2)
			//fmt.Println("全垒打玩家：",homeRunUser.User.GetID(),v.HitHeadScore,v.HitMidScore,v.HitTailScore)
			//fmt.Println("全垒打玩家：",homeRunUser.User.GetID(),homeRunUser.HeadPlus,homeRunUser.MidPlus,homeRunUser.TailPlus)
			score := int(v.HitScore) * 2
			//beHitUser.TotalPlus -= score
			//homeRunUser.TotalPlus += score

			homeRunUser.HomeRunHeadPlus += int(v.HitHeadScore * 2)
			homeRunUser.HomeRunMidPlus += int(v.HitMidScore * 2)
			homeRunUser.HomeRunTailPlus += int(v.HitTailScore * 2)
			homeRunUser.HomeRunTotalPlus += score
			beHitUser.HomeRunHeadPlus -= int(v.HitHeadScore * 2)
			beHitUser.HomeRunMidPlus -= int(v.HitMidScore * 2)
			beHitUser.HomeRunTailPlus -= int(v.HitTailScore * 2)
			beHitUser.HomeRunTotalPlus -= score
		}
		//str := fmt.Sprintf(`全垒打用户id: %d, 全垒打金额: %s`,
		//	homeRunUser.User.GetID(),
		//	score.GetScoreStr(int64(homeRunUser.HomeRunTotalPlus)*game.Bottom))
		//game.Table.WriteLogs(homeRunUid, str)
		return
	}

}

//结算
func (game *Game) SettleAccounts() {
	for _, user := range game.userList {
		if user != nil {
			user.TotalWin += user.HeadWin
			user.TotalWin += user.MidWin
			user.TotalWin += user.TailWin
			user.TotalPlus += user.HeadPlus
			user.TotalPlus += user.MidPlus
			user.TotalPlus += user.TailPlus
			user.FinalSettle = (int64(user.TotalWin) + int64(user.TotalPlus)) * game.Bottom
		}
	}

	// 赢/输 合值
	var (
		winSum        int64 // 赢钱合值
		loseSum       int64 // 输钱合值
		theoryWinSum  int64 // 理论赢钱合值
		theoryLoseSum int64 // 理论输钱合值
	)

	// 赢家结算列表
	winnerList := make(map[int64]*SettleResult)

	// 输家结算列表
	loserList := make(map[int64]*SettleResult)

	// 防止以小博大机制
	for _, user := range game.userList {

		curAmount := user.User.GetScore()

		if user.FinalSettle < 0 {

			loserSettle := &SettleResult{
				TheorySettle: user.FinalSettle,
				ActualSettle: user.FinalSettle,
				CurAmount:    curAmount,
			}

			// 不够输
			if user.FinalSettle < -curAmount {
				loserSettle.ActualSettle = -curAmount
			}

			theoryLoseSum += user.FinalSettle
			loseSum += loserSettle.ActualSettle
			loserList[user.User.GetID()] = loserSettle
		}

		if user.FinalSettle > 0 {

			winnerSettle := &SettleResult{
				TheorySettle: user.FinalSettle,
				ActualSettle: user.FinalSettle,
				CurAmount:    curAmount,
			}

			// 赢太多
			if user.FinalSettle > curAmount {
				winnerSettle.ActualSettle = curAmount
			}

			theoryWinSum += user.FinalSettle
			winSum += winnerSettle.ActualSettle
			winnerList[user.User.GetID()] = winnerSettle
		}

	}

	////// 除法向下取整，等比例减少钱时 最后一个玩家金额 = 总金额 - 前几个玩家等比例缩放金额合值

	// 输的人不够赔，等比例减少赢钱
	if winSum > -loseSum {

		var (
			winCounter int   // 赢家计数器
			winAcc     int64 // 赢钱累加器
		)

		for userID, v := range winnerList {

			// 最后一个赢钱玩家
			if len(winnerList)-winCounter == 1 {
				winnerList[userID].ActualSettle = -loseSum - winAcc
				break
			}

			// 缩减比例后新的结算值 = 应输合值 * 玩家理论赢钱值 / 理论赢钱合值
			winnerList[userID].ActualSettle = -loseSum * v.TheorySettle / theoryWinSum
			winAcc += winnerList[userID].ActualSettle
			winCounter++
		}

		var (
			leftAmount       int64 // 剩余多赢金额
			leftTheoryWinSum int64 // 剩余理论赢钱合值
		)

		for userID, v := range winnerList {

			// 缩减比例后 应赢钱大于携带金额，应赢钱减为携带金额，
			if v.ActualSettle >= v.CurAmount {
				leftAmount += v.ActualSettle - v.CurAmount
				winnerList[userID].ActualSettle = v.CurAmount
			} else {
				leftTheoryWinSum += v.TheorySettle
			}
		}

		// 有剩余多赢金额，重新分配一下剩余金额, 剩余钱 按照 玩家理论赢钱值 / 剩余玩家理论赢钱合值 比例分下去
		if leftAmount > 0 {
			FillWinnerAmount(&leftAmount, &leftTheoryWinSum, winnerList)
			ConvertWinnerAmount(leftAmount, leftTheoryWinSum, winnerList)
		}

	}

	// 赢钱的人赢过多，等比例减少输钱
	if winSum < -loseSum {
		var (
			lossCounter int   // 输家计数器
			lossAcc     int64 // 输钱累加器
		)

		for userID, v := range loserList {

			// 最后一个赢钱玩家
			if len(loserList)-lossCounter == 1 {
				loserList[userID].ActualSettle = (winSum + lossAcc) * -1
				break
			}

			// 缩减比例后新的结算值 = 应赢合值 * 玩家理论输钱值 / 理论输钱合值
			loserList[userID].ActualSettle = winSum * v.TheorySettle / theoryLoseSum * -1
			lossAcc += loserList[userID].ActualSettle
			lossCounter++
		}

		var (
			leftAmount        int64 // 剩余多输金额
			leftTheoryLoseSum int64 // 剩余理论输钱合值
		)

		for userID, v := range loserList {

			// 缩减比例后 应输钱大于携带金额，应输钱减为携带金额，
			if v.ActualSettle <= -v.CurAmount {
				leftAmount += v.ActualSettle + v.CurAmount
				loserList[userID].ActualSettle = -v.CurAmount
			} else {
				leftTheoryLoseSum += v.TheorySettle
			}
		}

		// 有剩余多输金额，重新分配一下剩余金额, 剩余多输钱 按照 玩家理论输钱值 / 剩余玩家理论输钱合值 比例分下去
		if leftAmount < 0 {
			FillLoserAmount(&leftAmount, &leftTheoryLoseSum, loserList)
			ConvertLoserAmount(leftAmount, leftTheoryLoseSum, loserList)
		}

	}

	for i, user := range game.userList {

		// 赢家金额
		for userID, v := range winnerList {
			if user.User.GetID() == userID {
				game.userList[i].FinalSettle = v.ActualSettle
				break
			}
		}

		// 输家金额
		for userID, v := range loserList {
			if user.User.GetID() == userID {
				game.userList[i].FinalSettle = v.ActualSettle
				break
			}
		}
	}
}

//发牌
func (game *Game) SendCards() {
	game.Cards = []byte{
		0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
		0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
		0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
		0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
	}
	//if game.GetRoomUserCount() != 4 {
	//	panic(fmt.Sprintf(`GetRoomUserCount: %d`, game.GetRoomUserCount()))
	//}
	game.ShuffleCards()
	for _, user := range game.userList {
		if user != nil {
			user.Status = global.USER_STATUS_ING
			if !user.IsTest {
				for i := 0; i < 13; i++ {
					user.Cards[i] = game.DealCards()
				}
			}
		}
	}
	//给玩家指定牌，
}

//洗牌
func (game *Game) ShuffleCards() {
	for i := 0; i < len(game.Cards); i++ {
		index1 := rand.Intn(len(game.Cards))
		index2 := rand.Intn(len(game.Cards))
		game.Cards[index1], game.Cards[index2] = game.Cards[index2], game.Cards[index1]
	}
}

//获取牌
func (gp *Game) DealCards() byte {
	card := gp.Cards[0]
	gp.Cards = append(gp.Cards[:0], gp.Cards[1:]...)
	return card
}

//重置桌子
func (game *Game) ResetLogicTable() {
	for _, v := range game.userList {
		if v != nil {
			v.SpecialCardType = 0
			fmt.Println("游戏结束，踢掉用户：", v.User.GetID())
			game.DelChairID(v.ChairId)
			game.DelUserList(v.User.GetID())
			game.Table.KickOut(v.User)
		}
	}
	game.HitRobArr = make([]*msg.S2CHitRob, 0)
	game.HomeRunUid = 0
	game.Status = global.TABLE_CUR_STATUS_WAIT
	game.Cards = make([]byte, 0)

	// add by wd in 2020.3.4
	game.Bottom = 0
}

//判断所有人确定就结束比赛
func (game *Game) AllUserSettleEndGame() {
	if game.Status != global.TABLE_CUR_STATUS_COMPOSE {
		return
	}
	for _, user := range game.userList {
		if user != nil && !user.IsSettleCards {
			return
		}
	}

	game.timerJob.Cancel()
	game.EndCompose()
}
