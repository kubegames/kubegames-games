package glogic

import (
	"go-game-sdk/example/game_MaJiang/960205/config"
	"go-game-sdk/example/game_MaJiang/960205/msg"
	"go-game-sdk/example/game_MaJiang/960205/poker"
	"go-game-sdk/example/game_MaJiang/960205/utils"
	frameMsg "go-game-sdk/sdk/msg"
	"math/rand"
	"strconv"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// SendMsgUSitSetDown 发送用户坐下消息
func (game *ErBaGangGame) SendMsgUSitSetDown(chairId int) {
	log.Tracef("发送用户坐下消息")
	log.Tracef("用户个数是:%d,用户数据是:%", len(game.UserAllList), game.UserAllList)
	if len(game.UserAllList) < 4 {
		return
	}
	reLine := game.checkIsReLine(chairId)

	res := &msg.UserSitDownRes{}
	res.BureauUUID = utils.CreateUUID()
	if len(game.UserAllList) == 4 {
		for k, v := range game.UserAllList {
			info := &msg.SeatUserInfoRes{
				NickName: v.InterUser.GetNike(),
				HeadImg:  v.InterUser.GetHead(),
				Ip:       v.InterUser.GetCity(),
				Gold:     v.InterUser.GetScore(),
			}
			info.MaxBet = game.RobZhuangMultipleList[k]
			if game.State > Game_Zhuang_End && k != game.RobZhuangIndex {
				info.MaxBet = -1
			}
			info.Bet = game.BetMultipleList[k]
			//if game.State > Game_Bet_End {
			//	info.Bet = -1
			//}
			if game.State > Game_DealCard {
				info.Tiles = game.PaiList[k]
			}
			switch v.InterUser.GetChairID() {
			case 0:
				SendUserAllList[0] = info
			case 1:
				SendUserAllList[1] = info
			case 2:
				SendUserAllList[2] = info
			case 3:
				SendUserAllList[3] = info
			}
		}
	}
	res.UserData = SendUserAllList
	res.RoundNo = int32(game.GameCount)
	res.TileAppearedTimes = game.Cards
	if game.State == Game_Zhuang {
		res.MaxBetTime = 5
	}
	if game.State == Game_Bet {
		res.BetTime = 5
	}
	res.Dealer = int32(game.RobZhuangIndex)
	res.Dices = game.DiceNumberArr
	res.IsEndToStart = game.IsEndToStart
	if len(game.PaiList) > 0 {
		res.Dealed = true
	}
	for k, v1 := range game.UserAllList {
		if reLine && k != chairId {
			continue
		}
		if game.State == Game_Bet && k != game.RobZhuangIndex {
			res.Bets = game.BetConfList[k]
		}

		if game.State == Game_Zhuang {
			res.MaxBets = game.RobZhuangConfList[k]
		}

		res.UserIndex = int32(k)
		res.Bet = int32(game.DiZhu)
		res.RoomId = game.InterTable.GetRoomID()
		_ = v1.InterUser.SendMsg(int32(msg.SendToClientMessageType_S2CUserSitDown), res)
	}

	if !reLine && game.checkState(Game_End) {
		// 等待2秒,发送游戏开始
		game.changeState()
		game.InterTable.AddTimer(4*1000, func() {
			game.SendMsgPlayerStart()
		})
	}
}

func (game *ErBaGangGame) checkIsReLine(chairId int) bool {
	for k, v := range game.UserAllList {
		if k == chairId && v.IsDeposit {
			v.IsDeposit = false
			return true
		}
	}
	return false
}

func (game *ErBaGangGame) checkState(state STATE) bool {
	if game.State != state {
		log.Tracef("state err", game.State, state)
		return false
	}
	return true
}

func (game *ErBaGangGame) changeState() {
	game.State++
	if game.State > Game_Count {
		game.State = Game_End
	}
}

func (game *ErBaGangGame) Reset() {
	game.RobZhuangIndex = -1
	game.PaiList = make(map[int][]int32, 0)
	game.RobZhuangConfList = make(map[int][]int32, 0)
	game.BetConfList = make(map[int][]int32, 0)
	game.DiceNumberArr = make([]int32, 0)
	for i := 0; i < 4; i++ {
		game.RobZhuangMultipleList[i] = -1
		game.BetMultipleList[i] = -1
	}
}

// SendMsgPlayerStart 发送游戏开始消息
func (game *ErBaGangGame) SendMsgPlayerStart() {
	log.Tracef("发送游戏开始消息")
	if !game.checkState(Game_Start) {
		return
	}
	game.InterTable.StartGame()
	game.Reset()
	game.IsEndToStart = false
	// 游戏局数+1
	game.GameCount++
	log.Tracef("游戏局数:%d", game.GameCount)
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CPlayerStart), &msg.PlayerStartRes{
		Number: int32(game.GameCount),
	})
	// 等待2秒,发送抢庄按钮配置消息
	game.InterTable.AddTimer(2*1000, func() {
		if game.checkState(Game_Start) {
			game.changeState()
			game.SendMsgRobBtnConf()
		}
	})
}

// SendMsgRobBtnConf 发送抢庄按钮消息
func (game *ErBaGangGame) SendMsgRobBtnConf() {
	log.Tracef("发送抢庄按钮消息")
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_Zhuang) {
		return
	}
	game.BtnCount = game.BtnCount + 1
	for i := 0; i < len(SendUserAllList); i++ {
		// 通过用户索引读取抢庄配置
		confArr := game.ReadRobBtnConf(i)
		game.RobZhuangConfList[i] = confArr
		user := game.UserAllList[i]
		if user == nil {
			continue
		}
		log.Tracef("发送给用户:%s的抢庄配置是:%v", user.InterUser.GetNike(), confArr)
		if err := user.InterUser.SendMsg(int32(msg.SendToClientMessageType_S2CRobZhuangStart), &msg.RobZhuangStartRes{
			Multiples: confArr,
			SleepTime: int32(config.ConfRobZhuangSleepTime()) / 1000,
		}); err != nil {
			log.Tracef("发送用户抢庄配置出错:%s", err.Error())
		}
	}

	game.InterTable.AddTimer(6*1000, func() {
		if game.checkState(Game_Zhuang) {
			game.changeState()
			game.SendMsgRobEnd()
		}
	})

}

// SendMsgRobEnd 抢庄完成消息
// 判断用户是否抢庄,如果没有默认为不抢庄
func (game *ErBaGangGame) SendMsgRobEnd() {
	log.Tracef("用户抢庄完成,")
	log.Tracef("有%d个用户抢庄,抢庄数据是:%v", len(game.RobZhuangMultipleList), game.RobZhuangMultipleList)
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_Zhuang_End) {
		return
	}
	for k, _ := range game.UserAllList {
		// 找出没有抢庄的用户
		if game.RobZhuangMultipleList[k] == -1 {
			log.Tracef("没有抢庄的用户是:%s", game.UserAllList[k].InterUser.GetNike())
			game.RobZhuangMultipleList[k] = 0
			game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2COneRobZhuangEnd), &msg.OneRobZhuangEndRes{
				UserIndex: int32(k),
				Multiple:  0,
			})
		}
	}
	log.Tracef("抢庄列表是:%v", game.RobZhuangMultipleList)
	game.SendMsgZhuangMax()
}

func (game *ErBaGangGame) checkNoZhuang() {
	for _, v := range game.RobZhuangMultipleList {
		if v != 0 {
			return
		}
	}
	index := rand.Intn(4)
	game.RobZhuangMultipleList[index] = 3
}

// SendMsgZhuangMax 发送庄家的索引
func (game *ErBaGangGame) SendMsgZhuangMax() {
	log.Tracef("发送庄家的索引")
	log.Tracef("游戏局数:%d", game.GameCount)
	var (
		//假设第一个元素是最大值，下标为0
		//maxVal   = game.RobZhuangMultipleList[0]
		maxIndex = 0
	)

	game.checkNoZhuang()
	// 如果4个用户都按下了抢庄按钮，就发送全部庄家的索引
	if len(game.RobZhuangMultipleList) == 4 {
		// 取出最大数和对应的用户索引
		//for i := 1; i < len(game.RobZhuangMultipleList); i++ {
		//	//从第二个 元素开始循环比较，如果发现有更大的，则交换
		//	if maxVal < game.RobZhuangMultipleList[i] {
		//		maxVal = game.RobZhuangMultipleList[i]
		//		maxIndex = i
		//	}
		//}
		maxIndex = int(game.getAChance([]int32{int32(game.RobZhuangMultipleList[0]), int32(game.RobZhuangMultipleList[1]),
			int32(game.RobZhuangMultipleList[2]), int32(game.RobZhuangMultipleList[3])}))
		game.RobZhuangIndex = maxIndex
	}
	log.Tracef("庄家id:%v", game.RobZhuangIndex)
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CAllRobZhuangEnd), &msg.AllRobZhuangEndRes{
		UserIndex: int32(game.RobZhuangIndex),
	})

	// 等待3秒发送用户下注按钮配置
	game.InterTable.AddTimer(4*1000, func() {
		if game.checkState(Game_Zhuang_End) {
			game.changeState()
			game.SendMsgBetBtnConf()
		}
	})
}

// SendMsgBetBtnConf 发送下注按钮配置消息
func (game *ErBaGangGame) SendMsgBetBtnConf() {
	log.Tracef("用户开始下注")
	log.Tracef("庄家索引是:%v", game.RobZhuangIndex)
	log.Tracef("玩家列表:%d", len(game.UserAllList))
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_Bet) {
		return
	}
	for k, v := range game.UserAllList {
		confArr := game.ReadBtnBtnConf(k)
		// 将下注配置存入列表
		game.BetConfList[k] = confArr
		if v.InterUser.GetChairID() != game.RobZhuangIndex {
			if err := v.InterUser.SendMsg(int32(msg.SendToClientMessageType_S2CUserBetInfoStart), &msg.UserBetInfoStartRes{
				Multiples: confArr,
				SleepTime: int32(config.ConfBetZhuangSleepTime() / 1000),
			}); err != nil {
				log.Errorf("用户抢庄按钮配置发送失败:%s", err.Error())
			}
		} else {
			if err := v.InterUser.SendMsg(int32(msg.SendToClientMessageType_S2CUserBetInfoStart), &msg.UserBetInfoStartRes{
				SleepTime: int32(config.ConfBetZhuangSleepTime() / 1000),
			}); err != nil {
				log.Errorf("用户抢庄按钮配置发送失败:%s", err.Error())
			}
		}
	}
	// 等待5秒后判断用户是否下注
	game.InterTable.AddTimer(6*1000, func() {
		if game.checkState(Game_Bet) {
			game.changeState()
			game.SendMsgBetEnd()
		}
	})
}

// SendMsgBetEnd 发送下注完成消息
// 判断用户是否下注,如果没有下注，默认为1
func (game *ErBaGangGame) SendMsgBetEnd() {
	if !game.checkState(Game_Bet_End) {
		return
	}
	for k := range game.BetMultipleList {
		if k != game.RobZhuangIndex {
			// 找到用户索引k为-1，且不是庄家索引的用户
			user := game.UserAllList[k]
			if game.BetMultipleList[k] == -1 && k != game.RobZhuangIndex && user != nil {
				log.Tracef("没有下注的用户是:%s", user.InterUser.GetNike())
				game.BetMultipleList[k] = 1
				game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CUserBetInfoEnd), &msg.UserBetInfoEndRes{
					UserIndex: int32(k),
					Multiple:  1,
				})
			}
		}
	}
	game.SendMsgDiceNumber()
}

// SendMsgDiceNumber 发送骰子点数
func (game *ErBaGangGame) SendMsgDiceNumber() {
	log.Tracef("发送骰子数量")
	log.Tracef("游戏局数:%d", game.GameCount)
	// 随机两个骰子数量
	for i := 0; i < 2; i++ {
		game.DiceNumberArr = append(game.DiceNumberArr, rand.Int31n(5)+1)
	}
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CDice), &msg.DiceRes{
		Numbers: game.DiceNumberArr,
	})

	// 等待4秒后发送发牌消息
	game.InterTable.AddTimer(4*1000, func() {
		if game.checkState(Game_Bet_End) {
			game.changeState()
			game.SendMsgDealCard()
		}
	})
}

// SendMsgDealCard 发送发牌消息
func (game *ErBaGangGame) SendMsgDealCard() {
	log.Tracef("发牌")
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_DealCard) {
		return
	}
	var (
		userArr []int32
	)

	for i := 0; i < 4; i++ {
		for j := 0; j < 2; j++ {
			card := int32(game.GamePoker.DealCards())
			userArr = append(userArr, card)
		}
		game.PaiList[i] = userArr
		userArr = nil

	}
	game.control()
	game.GamePoker.ShuffleCards()
	//发送发牌
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CDealCard), &msg.DealCardRes{})
	// 2秒后发送开牌消息
	game.InterTable.AddTimer(2*1000, func() {
		if game.checkState(Game_DealCard) {
			game.changeState()
			game.SendMsgOpenCard()
		}
	})
}

// SendMsgOpenCard 发送开牌消息
func (game *ErBaGangGame) SendMsgOpenCard() {
	log.Tracef("开牌消息")
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_OpenCard) {
		return
	}
	var (
		sendMsg []*msg.OpenCardInfo
	)
	for i := 0; i < len(SendUserAllList); i++ {
		cards := game.PaiList[i]
		sendMsg = append(sendMsg, &msg.OpenCardInfo{
			Number: cards,
		})
		for _, v := range cards {
			if v == 10 {
				game.Cards[0]++
			} else {
				game.Cards[v]++
			}
		}
	}

	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2COpenCard), &msg.OpenCardRes{
		CardNumbers: sendMsg,
	})
	var (
		tw []byte // 天王
	)
	t := 4000
	for i := 0; i < len(game.PaiList); i++ {
		tw = []byte{}
		for j := 0; j < 2; j++ {
			tw = append(tw, byte(game.PaiList[i][j]))
		}
		// 有天王
		a, _ := poker.GetCardType(tw)
		if a == 4 {
			t = 7000
			break
		}
	}
	game.InterTable.AddTimer(int64(t), func() {
		if game.checkState(Game_OpenCard) {
			game.changeState()
			game.SendMsgSmallCloseAnAccount()
		}
	})

}

func (game *ErBaGangGame) control() {
	cards := game.getOrderCards()
	log.Tracef("order card =", cards)
	chance0 := game.getUserWinChance(game.UserAllList[0])
	chance1 := game.getUserWinChance(game.UserAllList[1])
	chance2 := game.getUserWinChance(game.UserAllList[2])
	chance3 := game.getUserWinChance(game.UserAllList[3])
	indexs := game.getOrderIndex([]int32{chance0, chance1, chance2, chance3})
	length := len(indexs)
	for i := 0; i < len(cards); i++ {
		if length == 0 {
			break
		}
		if i < length {
			game.PaiList[int(indexs[i])] = cards[i]
		}
		if i >= length {
			index := game.getAOutIndex(int32(len(cards)), indexs)
			indexs = append(indexs, index)
			game.PaiList[int(index)] = cards[i]
		}

	}
	log.Tracef("paiList =", game.PaiList)
}

func (game *ErBaGangGame) getAOutIndex(limit int32, indexs []int32) int32 {
	find := false
	for i := int32(0); i < limit; i++ {
		find = false
		for _, v := range indexs {
			if v == i {
				find = true
				break
			}
		}
		if !find {
			return i
		}
	}
	return -1
}

func (game *ErBaGangGame) getOrderIndex(chances []int32) []int32 {
	log.Tracef("chances =", chances)
	indexs := make([]int32, 0)
	length := len(chances)
	c := make([]int32, 0)
	c = append(c, chances...)
	for i := 0; i < length; i++ {
		chances1 := game.increment(c)
		index := game.getAChance(chances1)
		if index > -1 {
			c[index] = 0
			indexs = append(indexs, index)
		}
	}
	log.Tracef("indexs = ", indexs)
	return indexs
}

func (game *ErBaGangGame) increment(chances []int32) []int32 {
	tem := make([]int32, 0)
	tem = append(tem, chances...)
	chance := game.sum(tem)
	length := len(tem)
	increment := int32(0)
	if chance < 10000 {
		increment = (10000 - chance) / int32(length)
	}
	if increment > 0 {
		for j := 0; j < length; j++ {
			if tem[j] != 0 {
				tem[j] += increment
			}
		}
	}
	return tem
}

func (game *ErBaGangGame) getAChance(chances []int32) int32 {
	chance := game.sum(chances)
	if chance <= 0 {
		log.Tracef("chances =", chances)
		return -1
	}
	r := int32(rand.Intn(int(chance)))
	k := int32(0)
	if r < chances[k] {
		return k
	}
	if r < chances[k]+chances[k+1] {
		return k + 1
	}
	if r < chances[k]+chances[k+1]+chances[k+2] {
		return k + 2
	}
	return k + 3
}

func (game *ErBaGangGame) sum(chances []int32) int32 {
	length := len(chances)
	chance := int32(0)
	for j := 0; j < length; j++ {
		chance += chances[j]
	}
	return chance
}

func (game *ErBaGangGame) getOrderCards() [][]int32 {
	allCard := make([][]byte, 0)
	for _, v := range game.PaiList {
		cards := make([]byte, 0)
		cards = append(cards, byte(v[0]))
		cards = append(cards, byte(v[1]))
		allCard = append(allCard, cards)
	}
	log.Tracef("allCard =", allCard)
	for i := 0; i < len(allCard)-1; i++ {
		for j := i + 1; j < len(allCard); j++ {
			if poker.GetCompareCardsRes(allCard[i], allCard[j]) == 0 {
				tem := allCard[i]
				allCard[i] = allCard[j]
				allCard[j] = tem
				log.Tracef("allCard eeee =", allCard)
			}
		}
	}
	allCards := make([][]int32, 0)
	for _, v := range allCard {
		cards := make([]int32, 0)
		cards = append(cards, int32(v[0]))
		cards = append(cards, int32(v[1]))
		allCards = append(allCards, cards)
	}
	log.Tracef("allCards =", allCards)
	return allCards
}

func (game *ErBaGangGame) getUserWinChance(user *User) int32 {
	if user == nil {
		return 0
	}
	key := game.InterTable.GetRoomProb()
	if key == 0 {
		key = 1000
	}
	log.Tracef("xuechi = ", key)
	user.IsPoint = "否"
	chance := config.GetPoolChance(strconv.Itoa(int(key)))
	if user.InterUser.IsRobot() {
		chance = config.GetRobotChance(strconv.Itoa(int(key)))
	}
	if user.InterUser.GetProb() != 0 {
		key = user.InterUser.GetProb()
		user.IsPoint = "是"
		log.Tracef("pointKey = ", key)
		chance = config.GetPointChance(strconv.Itoa(int(key)))
	}
	user.ControlKey = key
	return chance
}

func (game *ErBaGangGame) getChance(chance, interval int32) bool {
	if int32(rand.Intn(int(interval))) < chance {
		return true
	}
	return false
}

// SendMsgSmallCloseAnAccount 小局结算
func (game *ErBaGangGame) SendMsgSmallCloseAnAccount() {
	log.Tracef("小局结算")
	log.Tracef("游戏局数:%d", game.GameCount)
	if !game.checkState(Game_Count) {
		return
	}
	var (
		zhuangPai []byte
		xianPai   []byte
		goldArr   []int32 // 玩家的钱
	)

	goldArr = make([]int32, 4)

	for i := 0; i < len(game.PaiList[game.RobZhuangIndex]); i++ {
		zhuangPai = append(zhuangPai, byte(game.PaiList[game.RobZhuangIndex][i]))
	}

	for i := 0; i < len(SendUserAllList); i++ {
		if i != game.RobZhuangIndex {
			for j := 0; j < len(game.PaiList[i]); j++ {
				xianPai = append(xianPai, byte(game.PaiList[i][j]))
			}
			log.Tracef("牌", zhuangPai, xianPai, i)
			biPaiJieGuo := poker.GetCompareCardsRes(zhuangPai, xianPai)
			a := int(game.BetMultipleList[i] * int64(game.DiZhu))
			if biPaiJieGuo == 1 {
				// 庄家盈
				goldArr[i] = int32(-a)
				goldArr[game.RobZhuangIndex] += int32(a)

			} else {
				// 闲家
				goldArr[i] = int32(a)
				goldArr[game.RobZhuangIndex] -= int32(a)
			}
			xianPai = nil
		}
	}
	log.Tracef("%v", goldArr)
	log.Tracef("游戏局数:%d", game.GameCount)
	for i := 0; i < len(goldArr); i++ {
		user := game.UserAllList[i]
		if user == nil {
			continue
		}
		bussType := int32(101401)
		//bet := int64(goldArr[i])
		if goldArr[i] > 0 {
			bussType = 201401
		}
		//if goldArr[i] >= 0 {
		//	bet = int64(game.DiZhu)
		//}
		score, _ := user.InterUser.SetScore(game.InterTable.GetGameNum(), int64(goldArr[i]), game.InterTable.GetRoomRate())
		log.Tracef("user score =", user.InterUser.GetScore(), user.InterUser.GetID())
		chip := int64(0)
		outputAmount := score
		gameNum := game.InterTable.GetGameNum()
		if bussType == 101401 {
			chip = -score
			outputAmount = 0
		}
		if bussType == 201401 {
			chip = int64(game.DiZhu)
		}
		user.InterUser.SendChip(chip)
		user.InterUser.SendRecord(gameNum, score, chip, int64(goldArr[i])-score, outputAmount, "")
		game.createMarquee(user, int64(goldArr[i]))
		goldArr[i] = int32(score)
	}
	game.createOperationLog(goldArr)
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CSmallCloseAnAccount), &msg.SmallCloseAnAccountRes{
		GoldNumbers: goldArr,
	})
	if game.checkState(Game_Count) {
		game.changeState()
	}
	game.InterTable.EndGame()
	if game.checkState(Game_End) {
		game.changeState()
		game.Reset()
	}
	game.IsEndToStart = true
	// 判断局数是否达到5局
	reason := game.checkIsEndGame()
	if reason > -1 {
		// 如果达到5局,等待2秒发送大局结束消息
		game.Dismiss = true
		//if reason == 1 {
		//	game.SendMsgBigCloseAnAccount(reason)
		//	return
		//}
		game.InterTable.AddTimer(7*1000, func() {
			game.SendMsgBigCloseAnAccount(reason)
		})
	} else {
		game.InterTable.AddTimer(7*1000, func() {

			game.SendMsgPlayerStart()
		})
	}

}

func (game *ErBaGangGame) createOperationLog(coinChange []int32) {
	content := "当前局数:" + strconv.Itoa(game.GameCount)
	game.InterTable.WriteLogs(0, content)
	length := len(game.UserAllList)
	for k := 0; k < length; k++ {
		v := game.UserAllList[k]
		if v != nil {
			userIdLog := game.getUserIdLog(k)
			userOperationLog := game.getUserOperationLog(k)
			game.InterTable.WriteLogs(v.InterUser.GetID(), userIdLog+userOperationLog)
		}
		//content += game.getUserLog(k, int64(coinChange[k]))
	}
	for k := 0; k < length; k++ {
		v := game.UserAllList[k]
		if v != nil {
			userIdLog := game.getUserIdLog(k)
			userCountLog := game.getUserCountLog(k, int64(coinChange[k]))
			game.InterTable.WriteLogs(v.InterUser.GetID(), userIdLog+userCountLog)
		}
	}
	//for _, v := range game.UserAllList {
	//	if v != nil {
	//		game.InterTable.WriteLogs(v.InterUser.GetID(), content)
	//	}
	//}
}

func (game *ErBaGangGame) getUserLog(index int, coinChange int64) string {
	content := game.getUserIdLog(index)
	if content != "" {
		return content + game.getUserOperationLog(index) + game.getUserCountLog(index, coinChange)

	}
	return ""
}

func (game *ErBaGangGame) getUserIdLog(index int) string {
	v := game.UserAllList[index]
	if v != nil {
		isRobot := "否"
		if v.InterUser.IsRobot() {
			isRobot = "是"
		}
		return "用户" + strconv.Itoa(index+1) + ":" + strconv.FormatInt(v.InterUser.GetID(), 10) +
			" 作弊值: " + strconv.Itoa(int(v.ControlKey)) + "    " +
			" 是否点控: " + v.IsPoint +
			" 是否机器人: " + isRobot
	}
	return ""
}

func (game *ErBaGangGame) getUserOperationLog(index int) string {
	zhuang := "否    "
	if index == game.RobZhuangIndex {
		zhuang = "是    "
	}
	return "抢庄:" + strconv.FormatInt(game.RobZhuangMultipleList[index], 10) + "倍    " +
		"是否庄家:" + zhuang +
		"下注:" + strconv.FormatInt(game.BetMultipleList[index], 10) + "倍    "
}

func (game *ErBaGangGame) getUserCountLog(index int, coinChange int64) string {
	cards := game.PaiList[index]
	pai := make([]byte, 0)
	for j := 0; j < len(cards); j++ {
		pai = append(pai, byte(cards[j]))
	}
	return "牌型:" + game.getCardString(pai) + " " + game.getCardTypeString(pai) + "    " +
		"输赢金额:" + score.GetScoreStr(coinChange)
}

func (game *ErBaGangGame) getCardString(cards []byte) string {
	cardString := ""
	length := len(cards) - 1
	for k, v := range cards {
		if v < 10 {
			cardString += strconv.Itoa(int(v)) + "筒"
		}
		if v == 10 {
			cardString += "白板"
		}
		if k < length {
			cardString += "/"
		}
	}
	return cardString
}

func (game *ErBaGangGame) getCardTypeString(cards []byte) string {
	cardType, _ := poker.GetCardType(cards)
	if cardType == poker.CardTypeTW {
		return "天王"
	}
	if cardType == poker.CardTypeBao {
		return strconv.Itoa(int(cards[0])) + "宝"
	}
	if cardType == poker.CardType28 {
		return "28杠"
	}
	bankSum := (cards[0] + cards[1]) % 10
	if bankSum == 0 {
		return "鳖十"
	}
	cardString := strconv.Itoa(int(bankSum)) + "点"
	if cards[0] == 10 || cards[1] == 10 {
		cardString += "半"
	}
	return cardString
}

func (game *ErBaGangGame) createMarquee(user *User, coin int64) {
	orderRules := game.orderMarqueeRules(game.InterTable.GetMarqueeConfig())
	//for _, v := range orderRules {
	length := len(orderRules)
	for i := 0; i < length; i++ {
		v := orderRules[i]
		if v.GetAmountLimit() < 0 || coin < v.GetAmountLimit() {
			continue
		}
		game.InterTable.CreateMarquee(user.InterUser.GetNike(), coin, "", v.GetRuleId())
		break
	}
}

func (game *ErBaGangGame) orderMarqueeRules(rules []*frameMsg.MarqueeConfig) []*frameMsg.MarqueeConfig {
	orderRules := make([]*frameMsg.MarqueeConfig, 0)
	orderRules = append(orderRules, rules...)
	length := len(orderRules)
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++ {
			change := false
			if orderRules[i].GetAmountLimit() < orderRules[j].GetAmountLimit() {
				change = true
			}
			if change {
				tem := orderRules[i]
				orderRules[i] = orderRules[j]
				orderRules[j] = tem
			}
		}
	}
	return orderRules
}

func (game *ErBaGangGame) checkIsEndGame() int32 {
	for _, v := range game.UserAllList {
		if v != nil && v.IsDeposit {
			return 1
		}
	}

	if len(game.UserAllList) < 4 {
		return 1
	}

	if len(game.checkNotEnoughCoin()) > 0 {
		return 2
	}
	if game.GameCount == 5 {
		return 0
	}
	return -1
}

func (game *ErBaGangGame) checkNotEnoughCoin() []int32 {
	notEnoughs := make([]int32, 0)
	for k, v := range game.UserAllList {
		if v != nil && v.InterUser.GetScore() < int64(game.DiZhu*3) {
			notEnoughs = append(notEnoughs, int32(k))
		}
	}
	return notEnoughs
}

func (game *ErBaGangGame) SendMsgBigCloseAnAccount(reason int32) {
	log.Tracef("大局结算")
	game.InterTable.Broadcast(int32(msg.SendToClientMessageType_S2CBigCloseAnAccount), &msg.BigCloseAnAccountRes{
		Reason:         reason,
		NotEnoughUsers: game.checkNotEnoughCoin(),
	})
	for k, v := range game.UserAllList {
		log.Tracef("user score =", v.InterUser.GetScore(), v.InterUser.GetID())
		game.InterTable.KickOut(v.InterUser)
		delete(game.UserAllList, k)
	}
	game.Dismiss = false
	game.State = Game_End
	game.IsEndToStart = false
	game.GameCount = 0
	game.DiZhu = 0
	game.Cards = make([]int32, 10)
	game.GamePoker.InitPoker()
	game.GamePoker.ShuffleCards()
}
