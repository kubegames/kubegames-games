package game

import (
	"common/log"
	"fmt"
	"game_frame_v2/game/clock"
	"game_poker/pai9/config"
	"game_poker/pai9/model"
	pai9 "game_poker/pai9/msg"
	"math/rand"
	"sort"
	"time"

	"github.com/bitly/go-simplejson"
	protocol "github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table        table.TableInterface
	userChairMap map[int]*User   // chairID:user
	userIDMap    map[int64]*User // userID:user
	multi        config.Multi

	startTime time.Time
	endTime   time.Time

	startJob *clock.Job
	isStart  bool // 游戏是否开始

	zhuangChairID int     // 庄家的chairid
	zhuangWait    []int32 // 抢庄倍数一样时，随机庄

	setNum int // 游戏局数；2局为一轮

	// BetMultiList []int32 // 下注倍数列表

	dealPokerMsg *pai9.DealPokerMsg
	showPokerMsg *pai9.PokerInfoRespMsg

	status pai9.GameStatus
	job    *clock.Job //时间定时器

	qiangNum int // 已经抢庄了的人数
	betNum   int // 已经下注了的人数

	poker   *model.Cards
	UseProb int32
}

func NewGame(table table.TableInterface) *Game {
	game := new(Game)
	level := table.GetLevel()
	if level == 0 {
		level = 1
	}
	game.table = table
	game.userIDMap = make(map[int64]*User, 0)
	game.userChairMap = make(map[int]*User, 0)
	game.multi = config.Pai9Config.LevelConfig[level-1]
	game.showPokerMsg = new(pai9.PokerInfoRespMsg)
	game.dealPokerMsg = new(pai9.DealPokerMsg)
	game.dealPokerMsg.DiceVal = new(pai9.DiceValRespMsg)
	game.setNum = 0
	// game.poker = model.Init(model.CardsAllType)

	return game
}

func (game *Game) getUser(user player.PlayerInterface) (u *User, chairID int) {
	u = game.userIDMap[user.GetId()]
	if u == nil {
		u = NewUser(user)
		game.userIDMap[user.GetId()] = u
		chairID = game.getChairID()
		if chairID > 0 {
			game.userChairMap[chairID] = u
			u.chairID = chairID
		}
		if user.IsRobot() {
			rb := NewRobot()
			aiUser := user.BindRobot(rb)
			rb.user = aiUser
			rb.chairID = u.chairID
		}
	}

	chairID = u.chairID
	return
}

func (game *Game) sendStatusMsg(time int) {
	msg := new(pai9.StatusMessage)
	msg.Status = int32(game.status)
	msg.StatusTime = int32(time)
	fmt.Printf("状态 ************* %d  当前是第 %d 局\n", msg.Status, game.setNum)
	game.table.Broadcast(int32(pai9.SendToClientMessageType_Status), msg)
}

// 洗牌阶段
func (game *Game) Shuffle() {
	game.isStart = true
	game.table.StartGame()
	game.startTime = time.Now()
	game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.Shuffle), game.QiangZhuang)
	game.status = pai9.GameStatus_Shuffle
	// 发送游戏状态
	game.sendStatusMsg(int(config.Pai9Config.Taketimes.Shuffle))
}

// 抢庄阶段
func (game *Game) QiangZhuang() {
	game.Clear()
	// 洗牌
	game.poker.Shuffle()
	game.setNum++
	game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.QiangZhuang), game.SendWhoZhuang)
	game.status = pai9.GameStatus_QiangZhuang
	// 发送游戏状态
	game.sendStatusMsg(int(5000))
}

func (game *Game) SendWhoZhuang() {
	game.SendNotQiang()
	game.status = pai9.GameStatus_SendZhuang
	t := config.Pai9Config.Taketimes.SendZhuang + len(game.zhuangWait)
	game.sendStatusMsg(t)
	// 抢的数量
	if game.qiangNum < TABLE_NUM {
		game.setZhuangID()
		game.sendZhuangMsg()
	}

	game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.SendZhuang), game.BetStatus)
}

// 下注倍数阶段
func (game *Game) BetStatus() {
	game.CalcUserBetMulti()
	game.status = pai9.GameStatus_Bet
	// 发送游戏状态
	game.sendStatusMsg(int(config.Pai9Config.Taketimes.Bet))
	game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.Bet), game.DealPokerStatus)
}

// 发牌阶段
func (game *Game) DealPokerStatus() {
	game.SendNotBet()
	game.status = pai9.GameStatus_DealPoker
	// 发送游戏状态
	game.sendStatusMsg(config.Pai9Config.Taketimes.DealPoker)
	game.dealPokerMsg.DiceVal = new(pai9.DiceValRespMsg)
	game.dealPokerMsg.DiceVal.First = int32(rand.Intn(6)) + 1
	game.dealPokerMsg.DiceVal.Second = int32(rand.Intn(6)) + 1

	// 发牌顺序
	chairID := game.getDealOrder()
	game.dealPokerMsg.DealFirstChairID = int32(chairID[0])

	// fmt.Println("game.dealPokerMsg   ==========  ", game.dealPokerMsg)
	game.table.Broadcast(int32(pai9.SendToClientMessageType_DealPokerInfo), game.dealPokerMsg)

	game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.DealPoker), game.ShowPoker)
}

// 结算结算/展示牌阶段
func (game *Game) ShowPoker() {
	game.status = pai9.GameStatus_ShowPoker
	if game.setNum == 1 {
		// 发送游戏状态
		game.sendStatusMsg(int(config.Pai9Config.Taketimes.ShowPoker1))
	} else {
		// 发送游戏状态
		game.sendStatusMsg(int(config.Pai9Config.Taketimes.ShowPoker2))
	}

	game.dealPoker()

	bottom := game.getBottom()

	// 获取发牌顺序
	chairID := game.getDealOrder()
	// 比牌
	game.compare()

	zhuangUser := game.userChairMap[game.zhuangChairID]

	for _, chair := range chairID {

		user := game.userChairMap[chair]
		poker := make([]*pai9.Poker, 0, len(user.Cards))
		for _, v := range user.Cards {
			poker = append(poker, (*pai9.Poker)(&v.Poker))
		}

		//fmt.Printf("用户名称【%s】   user.WinGoldActual   =========  %d   当前金额 ====== %d\n ", user.GetUser().GetNike(), user.WinGoldActual, user.GetUser().GetScore())

		beforeSettle := user.GetUser().GetScore()
		output, _ := user.GetUser().SetScore(game.table.GetGameNum(), user.WinGoldActual, game.table.GetRoomRate())
		pval, _ := user.Cards.CalcType()
		user.WinGoldActual = output
		//fmt.Printf("2222用户名称【%s】   user.WinGoldActual   =========  %d   当前金额 ====== %d\n ", user.GetUser().GetNike(), user.WinGoldActual, user.GetUser().GetScore())
		game.showPokerMsg.Info = append(game.showPokerMsg.Info, &pai9.PokerInfoRespDetail{
			ChairID:    int32(chair),
			Poker:      poker,
			PokerVal:   pval,
			GoldChange: user.WinGoldActual,
			GoldNow:    user.GetUser().GetScore(),
		})
		// 每局发送战绩
		user.GetUser().SendRecord(game.table.GetGameNum(), user.WinGold-int64(user.BetMulti)*bottom, int64(user.BetMulti)*bottom, user.WinGold-output, output, "")
		// 发送打码量
		if user.WinGold-int64(user.BetMulti)*bottom > 0 {
			// 表示赢
			user.GetUser().SendChip(bottom)
		} else {
			user.GetUser().SendChip(int64(user.BetMulti) * bottom)
		}

		user.Settle(beforeSettle, int32(bottom), user == zhuangUser)
	}

	// 发送结算牌消息
	game.table.Broadcast(int32(pai9.SendToClientMessageType_PokerInfoResp), game.showPokerMsg)

	game.endTime = time.Now()

	// 每局写日志
	game.writeLog()

	var next func()
	if game.setNum == 2 {
		next = game.SettleAll
	} else {
		next = game.QiangZhuang
	}

	// 第一轮结束后有玩家不满足携带金额，直接进入一轮结算
	if game.isEnd() && game.setNum == 1 {
		log.Debugf("有玩家不满足当前最低携带金额 *****  直接进入一轮结算")
		fmt.Println("有玩家不满足当前最低携带金额 *****  直接进入一轮结算")
		next = game.SettleAll
	}

	if game.setNum == 1 {
		game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.ShowPoker1), next)
	} else {
		game.job, _ = game.table.AddTimer(time.Duration(config.Pai9Config.Taketimes.ShowPoker2), next)
	}
}

// 一轮结算状态
func (game *Game) SettleAll() {
	fmt.Println("一轮结算")
	game.status = pai9.GameStatus_SettleAll
	// 发送游戏状态
	game.sendStatusMsg(0)
	msg := new(pai9.SettleAllRespMsg)
	msg.Info = make([]*pai9.SettleAllRespDetail, 0, 2*TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		user := game.userChairMap[i]
		// 用户的结算下标0肯定存在(第一局满足进入房间条件)
		msg.Info = append(msg.Info, user.GetSettle()[0])
	}
	for i := 1; i <= TABLE_NUM; i++ {
		user := game.userChairMap[i]
		// 用户的结算下标1可能不存在(可能第二局不满足进入房间条件)
		if len(user.GetSettle()) >= 2 {
			msg.Info = append(msg.Info, user.GetSettle()[1])
		}
	}

	fmt.Println("一轮结算 ====== ", msg)

	game.table.Broadcast(int32(pai9.SendToClientMessageType_SettleAllResp), msg)

	game.Reset()
	game.table.EndGame()
}

// 为用户增减金额
func (game *Game) settleForUser() {
	// for i := 1; i <= TABLE_NUM; i++ {
	// 	user := game.userChairMap[i]
	// 	user.EndSettle(game.table.GetGameNum(), game.table.GetRoomRate(), game.getBottom(), game.setNum)
	// 	if user.WinGoldActual > 0 {
	// 		game.PaoMaDeng(user.WinGoldActual, user.GetUser())
	// 	}
	// }
}

// 获取发牌顺序，返回椅子号
func (game *Game) getDealOrder() []int {
	if game.dealPokerMsg.DiceVal == nil {
		game.dealPokerMsg.DiceVal = new(pai9.DiceValRespMsg)
	}
	dice := game.dealPokerMsg.DiceVal.First + game.dealPokerMsg.DiceVal.Second - 1
	offset := dice % TABLE_NUM
	zhuang := game.zhuangChairID

	var result []int
	for i := 0; i < TABLE_NUM; i++ {
		val := int(offset) + zhuang + i
	LOOP:
		if val >= TABLE_NUM+1 {
			val = val - TABLE_NUM
			goto LOOP
		}
		result = append(result, val)
	}
	//fmt.Println("获取到的发牌顺序 == ", result)
	return result
}

// 获取一个椅子号,-1表示没有座位了
func (game *Game) getChairID() int {
	if len(game.userChairMap) >= TABLE_NUM {
		return -1
	}
	chair := make(map[int]struct{}, TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		if _, exist := game.userChairMap[i]; !exist {
			chair[i] = struct{}{}
		}
	}
	for id := range chair {
		return id
	}
	return -1
}

// 比牌，计算牌的输赢
func (game *Game) compare() {
	// 用庄家牌和其他玩家牌进行比较
	zhuangUser := game.userChairMap[game.zhuangChairID]
	bottom := game.getBottom()
	for i := 1; i <= TABLE_NUM; i++ {
		chair := game.zhuangChairID + i
		if chair >= TABLE_NUM+1 {
			chair = chair - TABLE_NUM
		}
		user := game.userChairMap[chair]
		if user == zhuangUser {
			continue
		}

		cmp := zhuangUser.Cards.Compare(user.Cards)

		// 庄家手牌比玩家大或者相同时，庄家赢
		if cmp >= 0 {
			// 庄赢，玩家输
			result := bottom * int64(zhuangUser.QiangMulti) * int64(user.BetMulti)
			user.calcWinGold(-1 * result)
		} else if cmp < 0 {
			// 庄输，玩家赢
			result := bottom * int64(zhuangUser.QiangMulti) * int64(user.BetMulti)
			user.calcWinGold(result)
		}

	}

	// 计算庄家应输/赢金额
	for _, user := range game.userChairMap {
		if user == zhuangUser {
			continue
		}
		zhuangUser.WinGold -= user.WinGold
		zhuangUser.WinGoldActual -= user.WinGoldActual
	}

	zhuangWait := zhuangUser.WinGoldActual
	if zhuangUser.WinGoldActual < 0 {
		zhuangWait *= -1
	}

	var (
		isTrigger bool  // 是否触发以小博大
		diff      int64 // 差值
	)

	if zhuangWait > zhuangUser.user.GetScore() {
		// 触发以小博大机制
		if zhuangUser.WinGoldActual < 0 {
			zhuangUser.WinGoldActual = -1 * zhuangUser.user.GetScore()
			diff = zhuangUser.WinGoldActual - zhuangUser.WinGold
		} else {
			zhuangUser.WinGoldActual = zhuangUser.user.GetScore()
			diff = zhuangUser.WinGold - zhuangUser.WinGoldActual
		}
		isTrigger = true
	}

	if zhuangUser.WinGoldActual > 0 && isTrigger {
		// 庄家赢钱，且触发以小博大机制，重计算闲家的实际输钱
		var allLoss int64
		for _, user := range game.userChairMap {
			if user == zhuangUser {
				continue
			}
			if user.WinGoldActual < 0 {
				allLoss += user.WinGoldActual
			}
		}
		for _, user := range game.userChairMap {
			if user == zhuangUser {
				continue
			}
			if user.WinGoldActual < 0 {
				user.WinGoldActual = user.WinGoldActual - diff*user.WinGoldActual/allLoss
			}
		}
	} else if zhuangUser.WinGoldActual < 0 && isTrigger {
		// 庄家输钱，且触发庄家以小博大机制，重计算闲家的实际赢钱
		var allWin int64
		for _, user := range game.userChairMap {
			if user == zhuangUser {
				continue
			}
			if user.WinGoldActual > 0 {
				allWin += user.WinGoldActual
			}
		}
		for _, user := range game.userChairMap {
			if user == zhuangUser {
				continue
			}
			if user.WinGoldActual > 0 {
				user.WinGoldActual = user.WinGoldActual - diff*user.WinGoldActual/allWin
			}
		}
	}

}

func (game *Game) getBottom() int64 {
	str := game.table.GetAdviceConfig()
	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		log.Errorf("解析房间配置失败 err%v\n", err)
		return 0
	}
	baseBet, _ := js.Get("Bottom_Pouring").Int64()
	return baseBet
}

func (game *Game) GetUserList() []*User {
	list := make([]*User, 0, TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		u := game.userChairMap[i]
		if u != nil {
			list = append(list, u)
		}
	}
	return list
}

// 一轮结束清除
func (game *Game) Reset() {
	game.Clear()
	game.status = pai9.GameStatus_StartStatus
	if game.startJob != nil {
		game.startJob.Cancel()
		game.startJob = nil
	}
	game.isStart = false
	game.setNum = 0
	for _, user := range game.userChairMap {
		user.Reset()
		// 一轮结算剔除所有玩家
		game.table.KickOut(user.GetUser())
	}
	game.userChairMap = make(map[int]*User, 0)
	game.userIDMap = make(map[int64]*User, 0)
}

// 一局结束清除
func (game *Game) Clear() {
	game.zhuangChairID = 0
	game.qiangNum = 0
	game.betNum = 0
	// game.BetMultiList = nil
	for _, user := range game.userChairMap {
		user.Clear()
	}
	game.dealPokerMsg.Reset()
	game.showPokerMsg.Reset()
}

func (game *Game) qiangMsg(bts []byte, user player.PlayerInterface) {

	// 不是抢庄状态，返回
	if game.status != pai9.GameStatus_QiangZhuang {
		return
	}

	u, chairID := game.getUser(user)
	// 已抢庄，返回
	if u.HasQiang {
		return
	}

	qiang := new(pai9.QiangZhuangReqMsg)
	if err := protocol.Unmarshal(bts, qiang); err != nil {
		return
	}
	if qiang.Index < 0 || int(qiang.Index) >= len(game.multi.QiangMulti) {
		return
	}

	u.HasQiang = true
	u.QiangMulti = game.multi.QiangMulti[qiang.Index]
	game.qiangNum++

	// 广播该玩家的抢庄倍数
	msg := new(pai9.QiangZhuangRespMsg)
	msg.ChairID = int32(chairID)
	msg.Val = u.QiangMulti

	// if user.IsRobot() {
	// 	fmt.Printf("收到机器人[%d]--座位号[%d]---抢庄倍数：%v   时间：%v\n", user.GetId(), chairID, game.multi.QiangMulti[qiang.Index], game.job.GetIntervalTime())
	// } else {
	// 	fmt.Printf("收到真人[%d]--座位号[%d]---抢庄倍数：%v    时间：%v\n", user.GetId(), chairID, game.multi.QiangMulti[qiang.Index], game.job.GetIntervalTime())
	// }
	game.table.Broadcast(int32(pai9.SendToClientMessageType_QiangZhuangResp), msg)
	if game.qiangNum != TABLE_NUM {
		return
	}

	// 发送谁是庄
	game.setZhuangID()
	game.sendZhuangMsg()
	// 流程直接进入播放谁是庄的状态
	if game.job != nil {
		game.job.Cancel()
		// 进入谁是庄的状态
		game.SendWhoZhuang()
	}
}

func (game *Game) setZhuangID() {
	if game.zhuangChairID > 0 {
		return
	}

	qiangSort := make(QiangSorts, 0, TABLE_NUM)

	for i := 1; i <= TABLE_NUM; i++ {
		qiangSort = append(qiangSort, QiangSort{
			ChairID:    int32(i),
			QiangMulti: game.userChairMap[i].QiangMulti,
		})
	}
	sort.Sort(qiangSort)
	game.zhuangWait = []int32{qiangSort[0].ChairID}
	for i := 1; i < len(qiangSort); i++ {
		if qiangSort[i].QiangMulti == qiangSort[0].QiangMulti {
			game.zhuangWait = append(game.zhuangWait, qiangSort[i].ChairID)
		}
	}
	game.zhuangChairID = int(game.zhuangWait[rand.Intn(len(game.zhuangWait))])

	if zhuang := game.userChairMap[game.zhuangChairID]; zhuang.QiangMulti == 0 {
		zhuang.QiangMulti = game.multi.QiangMulti[1]
	}
	// fmt.Println("设置后的庄chairid = ", game.zhuangChairID)
	// fmt.Println("设置后的庄zhuangWait = ", game.zhuangWait)
}

func (game *Game) sendZhuangMsg() {
	msg := new(pai9.ZhuangRespMsg)
	msg.ChairID = int32(game.zhuangChairID)
	msg.ZhuangWait = game.zhuangWait
	msg.QiangMulti = int32(game.userChairMap[game.zhuangChairID].QiangMulti)
	game.table.Broadcast(int32(pai9.SendToClientMessageType_ZhuangResp), msg)
}

func (game *Game) betMsg(bts []byte, user player.PlayerInterface) {
	if game.status != pai9.GameStatus_Bet {
		return
	}
	u, chairID := game.getUser(user)
	if u.HasBet {
		return
	}
	bet := new(pai9.BetMultiReqMsg)
	if err := protocol.Unmarshal(bts, bet); err != nil {
		return
	}
	if bet.Index < 0 || int(bet.Index) >= 5 {
		return
	}

	if u == game.userChairMap[game.zhuangChairID] {
		return
	}

	game.betNum++
	u.HasBet = true
	u.BetMulti = u.BetMultiList[bet.Index]
	// 广播下注倍数
	msg := new(pai9.BetMultiRespMsg)
	msg.ChairID = int32(chairID)
	msg.Val = u.BetMulti

	// if user.IsRobot() {
	// 	fmt.Printf("收到机器人[%d]--座位号[%d]---下注倍数：%v   时间：%v\n", user.GetId(), chairID, u.BetMulti, game.job.GetIntervalTime())
	// } else {
	// 	fmt.Printf("收到真人[%d]--座位号[%d]---下注倍数：%v    时间：%v\n", user.GetId(), chairID, u.BetMulti, game.job.GetIntervalTime())
	// }
	game.table.Broadcast(int32(pai9.SendToClientMessageType_BetMultiResp), msg)
	if game.betNum != TABLE_NUM-1 {
		return
	}
	if game.job != nil {
		game.job.Cancel()
		game.DealPokerStatus()
	}

}

// 广播未抢庄的玩家
func (game *Game) SendNotQiang() {

	var notQiangNum int

	// 设置所有未抢庄的玩家的默认抢庄倍数
	for _, user := range game.userChairMap {
		if !user.HasQiang {
			notQiangNum++
			user.QiangMulti = game.multi.QiangMulti[0]
		}
	}

	// 都抢了庄
	if notQiangNum == 0 {
		return
	}

	for chairID, user := range game.userChairMap {
		if !user.HasQiang {
			msg := new(pai9.QiangZhuangRespMsg)
			msg.ChairID = int32(chairID)
			user.HasQiang = true
			msg.Val = user.QiangMulti
			fmt.Println("广播未抢庄的玩家  **** ", msg)
			game.table.Broadcast(int32(pai9.SendToClientMessageType_QiangZhuangResp), msg)
		}
	}

	// 有人没有抢庄
	game.setZhuangID()

}

// 广播未下注的玩家
func (game *Game) SendNotBet() {
	for chairID, user := range game.userChairMap {
		// 不是庄家且未下注
		if !user.HasBet && user != game.userChairMap[game.zhuangChairID] {
			msg := new(pai9.BetMultiRespMsg)
			msg.ChairID = int32(chairID)
			user.BetMulti = user.BetMultiList[0]
			user.HasBet = true
			msg.Val = user.BetMulti
			fmt.Println("广播未下注的玩家  **** ", msg)
			game.table.Broadcast(int32(pai9.SendToClientMessageType_BetMultiResp), msg)
		}
	}
}

// 计算玩家的下注倍数列表
func (game *Game) CalcUserBetMulti() {
	bottom := game.getBottom()
	zhuang := game.userChairMap[game.zhuangChairID]

	if zhuang.QiangMulti == 0 {
		zhuang.QiangMulti = 1
	}

	// 玩家投注倍数规则1：最高倍数 = 庄家携带金额 / 桌面有效玩家数（除开庄家）/ 庄家抢庄倍数 / 底分
	max := zhuang.GetUser().GetScore() / int64(TABLE_NUM-1) / int64(zhuang.QiangMulti) / bottom

	for _, user := range game.userChairMap {
		if user == zhuang {
			continue
		}
		// 玩家投注倍数规则2：闲家携带金额 / 抢庄倍数 / 底分
		multi := user.user.GetScore() / int64(zhuang.QiangMulti) / bottom
		if multi < max {
			max = multi
		}

		if max > 35 {
			max = 35
		} else if max < 5 {
			max = 5
		}

		for i := int64(1); i < 5; i++ {
			user.BetMultiList = append(user.BetMultiList, int32(max*i/5))
		}
		user.BetMultiList = append(user.BetMultiList, int32(max))
	}

	for _, user := range game.userChairMap {
		msg := new(pai9.UserBetMultiMsg)
		if user == zhuang {
			msg.List = []int32{1, 2, 3, 4, 5}
		} else {
			msg.List = user.BetMultiList
		}
		fmt.Printf("玩家【%d】的下注倍数列表 == %v\n", user.GetUser().GetId(), msg.List)

		// 发送下注倍数
		user.GetUser().SendMsg(int32(pai9.SendToClientMessageType_UserBetMulti), msg)
	}

}

func (game *Game) writeLog() {
	str := fmt.Sprintf("参与用户：\n")
	for i := 1; i <= TABLE_NUM; i++ {
		str += fmt.Sprintf("用户%dID:%v,\n", i, game.userChairMap[i].GetUser().GetId())
	}
	str += fmt.Sprintf("开始时间：%s\n", game.startTime.Format("2006-01-02 15:04:05"))
	str += fmt.Sprintf("结束时间：%s\n", game.endTime.Format("2006-01-02 15:04:05"))
	str += fmt.Sprintf("使用作弊值为：%v\n", game.UseProb)

	for i := 1; i <= TABLE_NUM; i++ {
		user := game.userChairMap[i]
		str += fmt.Sprintf("用户%dID:%v   抢庄：%d  下注：%d  \n", i, user.GetUser().GetId(), user.QiangMulti, user.BetMulti)
	}
	str += "开牌结果\n"
	for i := 1; i <= TABLE_NUM; i++ {
		user := game.userChairMap[i]
		_, pvalName := user.Cards.CalcType()
		str += fmt.Sprintf("用户%dID:%v   当前局数：%d  牌型：%s/%s  %s 输赢金额：%d\n", i, user.GetUser().GetId(), game.setNum, user.Cards[0].Name, user.Cards[1].Name, pvalName, user.WinGold)
	}

	game.table.WriteLogs(0, str)
}

// 房间的发牌策略控制
func (game *Game) dealPoker() {

	pokers := make(model.CardsTable, 0, TABLE_NUM)
	for i := 1; i <= TABLE_NUM; i++ {
		pokers = append(pokers, game.poker.DealPoker(2))
	}
	sort.Sort(pokers)

	roomProb, _ := game.table.GetRoomProb()
	// 真人玩家，机器人玩家
	var realUser, robotUser []int
	for chairID, user := range game.userChairMap {
		if user.GetUser().IsRobot() {
			robotUser = append(robotUser, chairID)
		} else {
			realUser = append(realUser, chairID)
		}
	}

	game.UseProb = roomProb
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

	if len(dealOrder) != 4 {
		panic(dealOrder)
	}
	// 给玩家赋值牌
	for i, chairID := range dealOrder {
		game.userChairMap[chairID].calcTestCard(pokers[i])
		if game.userChairMap[chairID].Cards == nil || len(game.userChairMap[chairID].Cards) != 2 {
			panic(game.userChairMap[chairID].Cards)
		}
	}
}

// 一句结束后检查玩家是否满足进入房间条件，如果不满足，则不开第二局
func (game *Game) isEnd() bool {
	for _, user := range game.userChairMap {
		if game.table.GetEntranceRestrictions() > 0 && user.GetUser().GetScore() < game.table.GetEntranceRestrictions() {
			return true
		}
	}
	return false
}

func (game *Game) handleTest(bts []byte, user player.PlayerInterface) {
	// 以下状态不能配牌
	if game.status == pai9.GameStatus_StartStatus ||
		game.status == pai9.GameStatus_DealPoker ||
		game.status == pai9.GameStatus_ShowPoker ||
		game.status == pai9.GameStatus_SettleAll ||
		!game.isStart {
		return
	}

	u, _ := game.getUser(user)
	msg := new(pai9.TestReqMsg)
	if err := protocol.Unmarshal(bts, msg); err != nil {
		return
	}
	if !(msg.Poker1 <= 21 && msg.Poker1 >= 1) {
		return
	}
	if !(msg.Poker2 <= 21 && msg.Poker2 >= 1) {
		return
	}

	u.handleTestMsg(msg)
}

// 获取剩余的牌
func (game *Game) getLeftCards() []int32 {
	result := make([]int32, len(model.CardsAllType))
	resultMap := make(map[int32]int32, len(model.CardsAllType))
	if game.poker == nil {
		return nil
	}
	for _, v := range *game.poker {
		resultMap[v.Sorted-1]++
	}
	for i := 0; i < len(model.CardsAllType); i++ {
		result[i] = resultMap[int32(i)]
	}
	fmt.Println("result = ", result)
	return result
}

// 跑马灯
func (g *Game) PaoMaDeng(Gold int64, user player.PlayerInterface) {
	configs := g.table.GetMarqueeConfig()
	for _, v := range configs {
		if Gold > v.AmountLimit {
			//log.Debugf("创建跑马灯")
			err := g.table.CreateMarquee(user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
		}
	}
}
