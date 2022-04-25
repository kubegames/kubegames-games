package gamelogic

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/970501/config"
	"go-game-sdk/example/game_LaBa/970501/model"
	proto "go-game-sdk/example/game_LaBa/970501/msg"
	"go-game-sdk/inter"
	"go-game-sdk/lib/clock"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/bitly/go-simplejson"
	protocol "github.com/golang/protobuf/proto"
)

type Game struct {
	table     table.TableInterface
	startTime time.Time
	endTime   time.Time

	roundMux     sync.Mutex
	roundEnd     bool // 是否转圈结束
	deleteUserID []int64

	userMux     sync.RWMutex
	UserBetInfo [BET_AREA_LENGHT]int64 // 玩家下注信息

	AIBetInfo    [BET_AREA_LENGHT]int64 // AI下注信息
	Status       int32                  //游戏状态
	UserMap      map[int64]*User        // 玩家列表
	TimerJob     *clock.Job             // 定时
	loopBetTimer *clock.Job

	Trend        []int32         // 走势图
	BetArr       []int64         // 下注区筹码
	BetLimitInfo config.RoomRule // 下注限制

	wait        sync.WaitGroup
	settleElems model.Elements // 结算时摇中元素
	// settleMsg   *proto.SettleMsg // 结算时信息

	outID   int32
	start   int           // 开火车初始值
	testMsg *proto.TestIn // 测试的结算结果

	allElems model.Elements
}

func NewGame(table table.TableInterface) *Game {
	return &Game{
		table:       table,
		UserBetInfo: [BET_AREA_LENGHT]int64{},
		AIBetInfo:   [BET_AREA_LENGHT]int64{},
		Trend:       []int32{},
		BetArr:      []int64{},
		// settleMsg:   new(proto.SettleMsg),
		UserMap: make(map[int64]*User),
	}
}

func (game *Game) BindRobot(ai inter.AIUserInter) player.RobotHandler {
	rb := NewRobot(game)
	rb.BindUser(ai)
	return rb
}

func (game *Game) GetUser(user player.PlayerInterface) *User {
	game.userMux.RLock()
	defer game.userMux.RUnlock()

	u, ok := game.UserMap[user.GetID()]
	if !ok {
		u = NewUser(game, user)
		game.UserMap[user.GetID()] = u
	} else {
		game.UserMap[user.GetID()].user = user
	}

LOOP:

	// 重新进入检查一下是否在移除队列中
	for i, v := range game.deleteUserID {
		if v == user.GetID() {
			game.deleteUserID = append(game.deleteUserID[:i], game.deleteUserID[i+1:]...)
			goto LOOP
		}
	}
	u.user = user
	return u
}

// 发送场景消息
func (game *Game) SendSceneMsg(user player.PlayerInterface) {
	game.GetUser(user)
	msg := new(proto.RoomSceneInfo)
	msg.Bets = game.BetArr
	msg.Gold = user.GetScore()
	msg.OnlineCount = int32(len(game.UserMap))

	u := game.GetUser(user)
	msg.MyBets = make([]*proto.SceneMyBet, BET_AREA_LENGHT)
	for index, v := range u.BetInfo {
		msg.MyBets[index] = &proto.SceneMyBet{
			AllGold:  game.UserBetInfo[index] + game.AIBetInfo[index],
			UserID:   user.GetID(),
			UserGold: v,
		}
	}
	msg.Trend = game.getTrendTop()
	msg.Odds = model.ElementsAll.GetOdds()
	user.SendMsg(int32(proto.SendToClientMessageType_RoomSence), msg)
}

func (game *Game) SendStatusMsg(duration int32) {
	msg := new(proto.StatusMessage)
	msg.Status = game.Status
	msg.StatusTime = duration

	log.Debugf("状态 ****************** %d", msg.Status)
	game.table.Broadcast(int32(proto.SendToClientMessageType_Status), msg)
}

func (game *Game) UserBet(buf []byte, user player.PlayerInterface) {
	if game.Status != int32(proto.GameStatus_BetStatus) {
		return
	}
	msg := new(proto.UserBet)
	if err := protocol.Unmarshal(buf, msg); err != nil {
		log.Errorf("用户下注时解析消息错误%v", err)
	}
	if msg.BetIndex < 0 || msg.BetIndex >= int32(len(game.BetArr)) || msg.BetType < 0 || msg.BetType >= BET_AREA_LENGHT {
		return
	}

	totalBet := [BET_AREA_LENGHT]int64{}
	for i := range game.UserBetInfo {
		totalBet[i] = game.UserBetInfo[i] + game.AIBetInfo[i]
	}

	u := game.GetUser(user)
	_, betMinLimit := game.getBaseBet()
	if u.DoBet(msg, totalBet, int64(betMinLimit)) {
		betGold := game.BetArr[msg.BetIndex]
		// 增加玩家戏下注
		if u.user.IsRobot() {
			game.AIBetInfo[msg.BetType] += betGold
		} else {
			game.UserBetInfo[msg.BetType] += betGold
		}
		// u.BetInfo[msg.BetType] += betGold
		msg.UserID = u.user.GetID()
		u.user.SendMsg(int32(proto.SendToClientMessageType_BetRet), msg)
	}
}

func (game *Game) getBaseBet() (int64, int) {
	str := game.table.GetAdviceConfig()
	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		log.Errorf("解析房间配置失败 err%v\n", err)
		return 0, 0
	}
	baseBet, _ := js.Get("Bottom_Pouring").Int64()
	betMinLimit, _ := js.Get("betMinLimit").Int()
	return baseBet, betMinLimit
}

// 开始游戏
func (game *Game) Start() {
	if game.table.GetRoomID() == -1 {
		log.Debugf("房间ID是负一")
		game.Status = 0
		return
	}
	game.Reset()

	game.roundMux.Lock()
	game.roundEnd = false
	game.roundMux.Unlock()

	game.table.StartGame()
	game.startTime = time.Now()
	game.goldNotice()
	game.Status = int32(proto.GameStatus_StartMovie)
	game.TimerJob, _ = game.table.AddTimer(int64(config.BenzBMWConf.Taketimes.Startmove), game.StartBet)
	//  发送游戏状态
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Startmove))

}

// 开始动画
func (game *Game) StartBet() {
	game.Status = int32(proto.GameStatus_BetStatus)
	game.TimerJob, _ = game.table.AddTimer(int64(config.BenzBMWConf.Taketimes.Startbet), game.EndBetMovie)
	//  发送游戏状态
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Startbet))
	game.loopBetTimer, _ = game.table.AddTimerRepeat(int64(config.BenzBMWConf.Taketimes.LoopBetGap), 0, game.loopBroadcastBetInfo)
}

//结束动画
func (game *Game) EndBetMovie() {
	// 发送结算消息
	game.TimerJob, _ = game.table.AddTimer(int64(config.BenzBMWConf.Taketimes.Endmove), game.SettleMsg)
	game.Status = int32(proto.GameStatus_EndBetMovie)
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Endmove))
	game.sendEndMsg()
	game.settleElems = nil
	game.wait.Add(1)
	go game.PreSettleRW()
}

// 发送结算消息
func (game *Game) SettleMsg() {
	game.wait.Wait()
	toRandTime := config.BenzBMWConf.Taketimes.Endpay
	switch {
	case len(game.settleElems) > 1 && game.settleElems[0].Id == model.GoodluckTypeTrainID:
		toRandTime += config.BenzBMWConf.Taketimes.EndpayAddTrain
	case len(game.settleElems) > 1:
		toRandTime += config.BenzBMWConf.Taketimes.EndpayAddNormal
	}

	game.TimerJob, _ = game.table.AddTimer(int64(toRandTime-1000), game.RoundEnd2)
	game.Status = int32(proto.GameStatus_SettleStatus)
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Endpay))

	game.userMux.RLock()
	defer game.userMux.RUnlock()

	for _, v := range game.UserMap {

		// 计算税后的金额
		taxMoney := v.LastWinGold * game.table.GetRoomRate() / 10000
		y := v.LastWinGold * game.table.GetRoomRate() % 10000
		if y > 0 {
			taxMoney += 1
		}
		output := v.LastWinGold - taxMoney
		v.taxGold = taxMoney
		// 设置每个玩家的分数
		// output, _ := v.user.SetScore(game.table.GetGameNum(), v.LastWinGold, game.table.GetRoomRate())
		// v.user.SendRecord(game.table.GetGameNum(), output-v.BetGoldNow, v.BetGoldNow, v.LastWinGold-output, output, "")
		// 税后计算
		v.LastWinGold = output

		mymsg := new(proto.SettleMsg)

		var train bool
		var self *proto.UserSettleInfo

		// 特殊奖项
		if len(game.settleElems) > 1 {
			self = &proto.UserSettleInfo{
				UserId:       v.user.GetID(),
				WinGold:      v.LastWinGold,
				GoodluckType: int32(game.settleElems[0].Id),
				OutId:        game.outID,
			}
			if game.settleElems[0].Id == model.GoodluckTypeTrainID {
				train = true
			}
		} else {
			self = &proto.UserSettleInfo{
				UserId:  v.user.GetID(),
				WinGold: v.LastWinGold,
				OutId:   game.outID,
			}
		}
		mymsg.Self = self
		if train {
			mymsg.Begin = int32(game.start)
		}
		mymsg.GoldNow = v.user.GetScore() + v.LastWinGold
		v.user.SendMsg(int32(proto.SendToClientMessageType_Settle), mymsg)
		// 发送打码量
		// v.user.SetChip(v.BetGoldNow)

	}

	game.checkUser()
	// game.Reset()
}

func (game *Game) setTrend() {
	game.Trend = append(game.Trend, game.outID)

	if len(game.Trend) > TREND_LENGTH {
		temp := game.Trend[len(game.Trend)-TREND_LENGTH:]
		game.Trend = make([]int32, TREND_LENGTH)
		copy(game.Trend, temp)
	}
}

func (game Game) getTrendTop() []int32 {
	if len(game.Trend) >= TREND_LENGTH {
		return game.Trend[len(game.Trend)-TREND_LENGTH:]
	}
	return game.Trend
}

func (game Game) getBetAllUser() (allBet int64) {
	for _, v := range game.UserBetInfo {
		allBet += v
	}
	return
}

func (game Game) SendTrendMsg(user player.PlayerInterface) {
	msg := new(proto.TrendMsg)
	msg.Trend = game.Trend
	user.SendMsg(int32(proto.SendToClientMessageType_TrendRet), msg)
}

func (game *Game) initRule() {
	if game.BetLimitInfo.BaseBet != 0 {
		return
	}
	betBase, _ := game.getBaseBet()
	game.BetLimitInfo.BaseBet = betBase
	game.BetArr = []int64{
		betBase * 1,
		betBase * 10,
		betBase * 50,
		betBase * 100,
		betBase * 500,
	}
	game.BetLimitInfo.LimitPerUser = betBase * 1000
	game.BetLimitInfo.AllLimitPerArea = betBase * 25000
	game.BetLimitInfo.AllLimitPerUser = betBase * 2500
}

func (game *Game) Reset() {
	// if game.settleMsg != nil {
	// 	game.settleMsg.Reset()
	// 	game.settleElems = nil
	// }
	game.UserBetInfo = [BET_AREA_LENGHT]int64{}
	game.AIBetInfo = [BET_AREA_LENGHT]int64{}
	for _, v := range game.UserMap {
		v.Reset()
	}
	game.deleteUserID = nil
	game.testMsg = nil
}

func (game Game) getUserList(buf []byte, user player.PlayerInterface) {

	args := new(proto.UserListReq)
	if err := protocol.Unmarshal(buf, args); err != nil {
		log.Errorf("获取用户列表消息解析错误：", err)
		return
	}

	// if args.PageIndex <= 0 {
	// 	return
	// }
	// if int(args.PageIndex)-USER_PAGE_LIMIT < 0 {
	// 	return
	// }

	var temp = make(UserList, 0, len(game.UserMap))

	game.userMux.RLock()
	defer game.userMux.RUnlock()

	for _, user := range game.UserMap {
		temp = append(temp, user)
	}

	sort.Sort(temp)
	msg := new(proto.UserListResp)

	// 取消分页
	// if len(temp) >= int(args.PageIndex)+USER_PAGE_LIMIT {
	// 	temp = temp[int(args.PageIndex)-USER_PAGE_LIMIT : args.PageIndex]
	// } else if len(temp) >= int(args.PageIndex) {
	// 	temp = temp[int(args.PageIndex)-USER_PAGE_LIMIT : args.PageIndex]
	// } else if len(temp) < int(args.PageIndex)-USER_PAGE_LIMIT {
	// 	temp = nil
	// } else if len(temp) < int(args.PageIndex) {
	// 	temp = temp[int(args.PageIndex)-USER_PAGE_LIMIT:]
	// }

	for _, v := range temp {
		ui := new(proto.UserInfo)
		ui.NickName = v.user.GetNike()
		ui.Gold = v.user.GetScore()
		ui.ID = v.user.GetID()
		ui.WinGold = v.WinGold
		ui.BetGold = v.BetGold
		ui.Avatar = v.user.GetHead()
		msg.UserList = append(msg.UserList, ui)
	}
	user.SendMsg(int32(proto.SendToClientMessageType_UserList), msg)
}

func (game *Game) UserOut(user player.PlayerInterface) {
	game.userMux.Lock()
	defer game.userMux.Unlock()
	delete(game.UserMap, user.GetID())
}

// 计算赢的金额
func (game *Game) calcWinGold() {

	var tmp model.Elements
	if len(game.settleElems) > 1 {
		tmp = game.settleElems[1:]
	} else {
		tmp = game.settleElems
	}
	for _, v := range game.UserMap {
		for _, sr := range tmp {
			if sr.IsMax {
				v.LastWinGold += v.BetInfo[int(sr.Id)] * int64(sr.OddsMax.Odds)
			} else {
				v.LastWinGold += v.BetInfo[int(sr.Id)] * int64(sr.OddsMin.Odds)
			}
		}
	}

}

func (game *Game) checkUser() {
	for _, v := range game.UserMap {
		level := game.table.GetLevel()
		min, max := config.RobotConf.EvictGold[level-1][0], config.RobotConf.EvictGold[level-1][1]
		if v.user.IsRobot() && (v.user.GetScore() < min || v.user.GetScore() > max) {
			// 机器人直接剔除
			game.table.KickOut(v.user)
			v.Reset()
			game.deleteUserID = append(game.deleteUserID, v.user.GetID())
			continue
		}

		var betVal int64
		for _, val := range v.BetInfo {
			betVal |= val
		}
		if betVal == 0 {
			v.NotBetCount++
		} else {
			v.NotBetCount = 0
			continue
		}

		if v.NotBetCount >= config.BenzBMWConf.Unplacebetnum {
			// n次未下注，提出玩家
			if v.user.IsRobot() {
				// 机器人直接剔除
				v.Reset()
				game.table.KickOut(v.user)
				game.deleteUserID = append(game.deleteUserID, v.user.GetID())
				// delete(game.UserMap, k)
				continue
			}
			// 发送剔除用户消息
			msg := new(proto.BetFailMsg)
			msg.BetFailInfo = "最近五局未下注，已被移出房间。"
			msg.IsKickOut = true
			//game.Table.KickOut(v.User)
			v.Reset()
			v.NotBetCount = 0
			game.deleteUserID = append(game.deleteUserID, v.user.GetID())
			v.user.SendMsg(int32(proto.SendToClientMessageType_BetFailID), msg)
		}
	}

}

// 循环播放下注消息
func (game *Game) loopBroadcastBetInfo() {
	if game.Status != int32(proto.GameStatus_BetStatus) {
		return
	}
	// 小于2s的时候不再循环发送广播
	if game.TimerJob.GetTimeDifference() < 2000 {
		if game.loopBetTimer != nil {
			game.loopBetTimer.Cancel()
			game.loopBetTimer = nil
		}
		return
	}
	msg := new(proto.LoopBetNoticeMsg)
	msg.BetGold = make([]int64, BET_AREA_LENGHT)
	for index, v := range game.UserBetInfo {
		msg.BetGold[index] = v + game.AIBetInfo[index]
	}
	game.table.Broadcast(int32(proto.SendToClientMessageType_LoopBetNotice), msg)
}

// 单独播放结束消息
func (game *Game) sendEndMsg() {
	msg := new(proto.LoopBetNoticeMsg)
	msg.BetGold = make([]int64, BET_AREA_LENGHT)
	for index, v := range game.UserBetInfo {
		msg.BetGold[index] = v + game.AIBetInfo[index]
	}
	game.table.Broadcast(int32(proto.SendToClientMessageType_LoopBetNotice), msg)

}

func (game *Game) DoTest(bts []byte) {
	if game.Status != int32(proto.GameStatus_BetStatus) {
		return
	}
	var msg proto.TestIn
	err := protocol.Unmarshal(bts, &msg)
	if err != nil {
		return
	}
	msg.OutID %= int32(model.GoodluckTypeTrainID) + 1
	game.testMsg = &msg
}

// func (game *Game) TestPreSettle() {
// 	if !(game.testIn >= 0 && game.testIn <= 23) {
// 		return
// 	}

// 	game.settleMsg.Reset()

// 	var element model.Elements
// 	game.settleElems = nil
// 	roomProbNow, _ := game.table.GetRoomProb()
// 	winCtrl := config.BenzBMWConf.WinCtrl.Find(int(roomProbNow))
// 	if game.testIn == 3 {
// 		element = model.ElemBases{model.ElemThree}
// 		// threeElem:=fmt.Sprintf("%x",winCtrl.Surprise.ThreeCtrl.Rand().ElemType)
// 		element = append(element, model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.ThreeCtrl.Rand().ElemType))...)
// 		for _, v := range game.settleElems {
// 			game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, int32(v.RandSubId()))
// 		}
// 	} else if game.testIn == 4 {
// 		// 中大四喜
// 		element = model.ElemBases{model.ElemFour}
// 		// fourElem:=fmt.Sprintf("%x",winCtrl.Surprise.FourCtrl.Rand().Color)
// 		element = append(element, model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.FourCtrl.Rand().Color))...)

// 		for _, v := range game.settleElems {
// 			game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, int32(v.RandSubId()))
// 		}
// 	} else {
// 		game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, game.testIn)
// 	}
// }

// 结算规则
func (game *Game) PreSettleRW() {
	// game.TimerJob, _ = game.table.AddTimer(int64(config.BenzBMWConf.Taketimes.Endmove-1000), game.SettleMsg)
	defer game.wait.Done()
	defer func() {
		if game.testMsg != nil {
			switch model.ElementType(game.testMsg.OutID) {
			case model.GoodluckTypeTrainID,
				model.GoodluckTypeThreeBigID,
				model.GoodluckTypeThreeSmallID,
				model.GoodluckTypeSlamBigID,
				model.GoodluckTypeSlamSmallID,
				model.GoodluckTypeFoisonBigID: // 作弊开出goodluck

				gl := []int{9, 21}
				game.outID = int32(gl[rand.Intn(len(gl))])
			default:
				game.outID = game.testMsg.OutID
			}
		} else {
			game.outID = int32(game.settleElems[0].RandSubId())
		}
	}()
	roomProbNow := game.table.GetRoomProb()

	if roomProbNow == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProbNow)
		roomProbNow = 1000
	}

	goodLuck := model.GoodlucksAll.Find(int64(roomProbNow))
	var element model.Elements
	var start int
	element, start = goodLuck.Rand(game.testMsg, config.BenzBMWConf.IsOpen3000Ctrl, int64(roomProbNow)).Handle(game.testMsg)
	game.start = start
	if element != nil {
		// 中特殊奖项
		game.settleElems = element
		// game.TestPreSettle()
		game.calcWinGold()
		return
	}

	// 根据返奖率进行计算

	back := goodLuck.Backs.Rand()
	backProbMin, backProbMax := int64(back.Min), int64(back.Max)

	backProb := [16]int64{}

	var allBet int64
	for _, v := range game.UserBetInfo {
		allBet += v
	}

	for id, v := range game.UserBetInfo {
		ele := model.ElementsAll.GetById(model.ElementType(id), nil)
		backProb[id] = int64(float64(v*int64(ele.OddsMax.Odds)) / float64(allBet) * 100)                 // 大倍率
		backProb[id+BET_AREA_LENGHT] = int64(float64(v*int64(ele.OddsMin.Odds)) / float64(allBet) * 100) // 小倍率
	}

	var wait []int
	for id, v := range backProb {
		if v >= backProbMin && v < backProbMax {
			wait = append(wait, id)
		}
	}

	// 3000作弊率下必输控制
	if config.BenzBMWConf.IsOpen3000Ctrl && roomProbNow == 3000 {
		// 重置wait，选出返奖率小于100的值
		wait = nil
		for id, v := range backProb {
			if v < 100 {
				wait = append(wait, id)
			}
		}

		if len(wait) == 0 {
			indexMin, backMin := 0, backProb[0]
			for i, v := range backProb {
				if v < backMin {
					backMin = v
					indexMin = i
				}
			}
			wait = append(wait, indexMin)
		}

	}

	if len(wait) != 0 {
		randIndex := wait[rand.Intn(len(wait))]
		var ismax bool
		if randIndex >= BET_AREA_LENGHT {
			randIndex -= BET_AREA_LENGHT
			ismax = false
		} else {
			ismax = true
		}
		element = model.Elements{model.ElementsAll.GetById(model.ElementType(randIndex), &ismax)}
	} else {
		element = model.Elements{model.ElementsAll.Rand(model.ElementType(-1))}
	}

	game.settleElems = element
	game.handleTestMsg()
	// game.TestPreSettle()
	game.calcWinGold()
}

func (game *Game) getEndGameResult() (element model.Elements) {
	allBet := game.getBetAllUser()
	oddsList := model.ElementsAll.GetOdds()
	roomProbNow := game.table.GetRoomProb()

	if roomProbNow == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProbNow)
		roomProbNow = 1000
	}

RECALC: // 重新计算摇奖结果
	var waitCheck []int
	var backProb = [12]int64{} // 返奖率
	waitCheck = nil

	if allBet != 0 {
		for index, v := range game.UserBetInfo {
			backProb[index] = int64(float64(v*int64(oddsList[index])) / float64(allBet) * float64(100))
		}
	}

	var backMin, backMax int64

	for index, v := range backProb {
		if v >= backMin && v < backMax {
			waitCheck = append(waitCheck, index)
		}
	}
	if len(waitCheck) != 0 {
		// 这里随机选择一个进行返回
		index := waitCheck[rand.Intn(len(waitCheck))]
		element = model.Elements{model.ElementsAll.GetById(model.ElementType(index), nil)}
		return
	}
	if roomProbNow == 1000 {
		element = model.Elements{model.ElementsAll.Rand(model.ElementType(-1))}
		return
	} else if roomProbNow > 1000 {
		roomProbNow -= 1000
	} else if roomProbNow == 0 {
		roomProbNow = 1000
	} else {
		roomProbNow += 1000
	}
	goto RECALC
}

func (game *Game) BetRept(bts []byte, user player.PlayerInterface) {
	msg := new(proto.BetReptMsg)
	if err := protocol.Unmarshal(bts, msg); err != nil {
		return
	}
	if len(msg.BetGold) != 8 {
		return
	}

	var allBet int64
	for _, v := range msg.BetGold {
		if v < 0 {
			return
		}
		allBet += v
	}

	uu := game.GetUser(user)
	if user.GetScore() < allBet {
		uu.BetFailed("金币不足", false, false)
		return
	}

	for i := range msg.BetGold {
		uu.BetInfo[i] += msg.BetGold[i]
		game.UserBetInfo[i] += msg.BetGold[i]
	}
	uu.BetGoldNow += allBet
	user.SetScore(game.table.GetGameNum(), -1*allBet, 0)
	user.SendMsg(int32(proto.SendToClientMessageType_BetReptResp), msg)
}

func (game *Game) RoundEnd() {
	if game.Status != int32(proto.GameStatus_SettleStatus) {
		return
	}
	game.roundMux.Lock()
	defer game.roundMux.Unlock()
	if game.roundEnd {
		return
	}
	game.roundEnd = true
	game.afterSettle()
	game.clearUser()
	game.setTrend()
	game.endTime = time.Now()
	game.Reset()
	game.table.EndGame()
}

func (game *Game) RoundEnd2() {
	game.TimerJob, _ = game.table.AddTimer(int64(1000), game.Start)
	game.roundMux.Lock()
	defer game.roundMux.Unlock()
	if game.roundEnd {
		return
	}
	game.roundEnd = true
	game.afterSettle()
	game.clearUser()
	game.setTrend()
	game.endTime = time.Now()
	game.Reset()
	game.table.EndGame()
}

// 清除用户
func (game *Game) clearUser() {
	for _, k := range game.deleteUserID {
		if user := game.UserMap[k]; user != nil {
			user.Reset()
		}
		delete(game.UserMap, k)
	}
}

// 为每个用户设置金额
func (game *Game) afterSettle() {
	game.userMux.Lock()
	defer game.userMux.Unlock()
	for _, v := range game.UserMap {
		if v.LastWinGold > 0 {
			game.PaoMaDeng(v.LastWinGold, v.user)
		}
		output, _ := v.user.SetScore(game.table.GetGameNum(), v.LastWinGold, 0)
		v.user.SendRecord(game.table.GetGameNum(), v.LastWinGold-v.BetGoldNow, v.BetGoldNow, v.taxGold, output, "")
		var allBet int64
		for _, val := range v.BetInfo {
			allBet += val
		}
		v.SyncBetGold(allBet)
		v.SyncWinData(v.LastWinGold)
		v.SendChip()
	}
	game.writeLog()
}

// 处理作弊消息
func (game *Game) handleTestMsg() {
	if game.testMsg == nil {
		return
	}
	isMax := true
	isSmall := false
	for _, ele := range model.ElementsAll {
		for _, max := range ele.OddsMax.SubIds {
			if game.testMsg.OutID == int32(max) {
				game.settleElems = model.Elements{model.ElementsAll.GetById(ele.Id, nil).Copy(&isMax)}
			}
		}

		for _, min := range ele.OddsMin.SubIds {
			if game.testMsg.OutID == int32(min) {
				game.settleElems = model.Elements{model.ElementsAll.GetById(ele.Id, nil).Copy(&isSmall)}
			}
		}
	}
}

func (game *Game) goldNotice() {
	msg := new(proto.GoldNowMsg)
	for _, user := range game.UserMap {
		msg.Reset()
		msg.GoldNow = user.user.GetScore()
		msg.UserID = user.user.GetID()
		user.user.SendMsg(int32(proto.SendToClientMessageType_GoldNow), msg)
	}
}

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

func (game *Game) writeLog() {

	game.endTime = time.Now()

	roomProb := game.table.GetRoomProb()

	if roomProb == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProb)
		roomProb = 1000
	}

	var isNormal bool
	if len(game.settleElems) == 1 {
		isNormal = true
	}
	var normal, spec, normalResult, specResult, specType string
	if isNormal {
		normal = "是"
		spec = "否"
	} else {
		spec = "是"
		normal = "否"
	}

	getNormalResult := func(val *model.Element) string {
		if val == nil {
			return ""
		}
		var normalResult string
		switch {
		case val.Id == 0 && val.IsMax:
			normalResult = "大双BAR"
		case val.Id == 1 && val.IsMax:
			normalResult = "大双7"
		case val.Id == 2 && val.IsMax:
			normalResult = "大双星"
		case val.Id == 3 && val.IsMax:
			normalResult = "大西瓜"
		case val.Id == 4 && val.IsMax:
			normalResult = "大铃铛"
		case val.Id == 5 && val.IsMax:
			normalResult = "大橘子"
		case val.Id == 6 && val.IsMax:
			normalResult = "大柠檬"
		case val.Id == 7 && val.IsMax:
			normalResult = "大苹果"

		case val.Id == 0 && !val.IsMax:
			normalResult = "小双BAR"
		case val.Id == 1 && !val.IsMax:
			normalResult = "小双7"
		case val.Id == 2 && !val.IsMax:
			normalResult = "小双星"
		case val.Id == 3 && !val.IsMax:
			normalResult = "小西瓜"
		case val.Id == 4 && !val.IsMax:
			normalResult = "小铃铛"
		case val.Id == 5 && !val.IsMax:
			normalResult = "小橘子"
		case val.Id == 6 && !val.IsMax:
			normalResult = "小柠檬"
		case val.Id == 7 && !val.IsMax:
			normalResult = "小苹果"
		}
		return normalResult
	}

	if isNormal {
		normalResult = getNormalResult(game.settleElems[0])
	} else {
		switch game.settleElems[0].Id {
		case model.GoodluckTypeThreeBigID:
			specResult = "大三元"
			specType = "大西瓜、大双星、大铃铛"
		case model.GoodluckTypeThreeSmallID:
			specResult = "小三元"
			specType = "大柠檬、大橘子、大苹果"
		case model.GoodluckTypeSlamBigID:
			specResult = "大满贯"
			specType = "不带x2的图标（不包含BAR）"
		case model.GoodluckTypeSlamSmallID:
			specResult = "小满贯"
			specType = "所有x2的图标都中奖"
		case model.GoodluckTypeFoisonBigID:
			specResult = "大丰收"
			specType = "大西瓜、大柠檬、大苹果、大橙子"
		case model.GoodluckTypeTrainID:
			specResult = "开火车"
			isMax := true
			isSmall := false
			var ele *model.Element
			for _, v := range model.ElementsAll {
				for _, max := range v.OddsMax.SubIds {
					if game.start == max {
						ele = model.ElementsAll.GetById(v.Id, &isMax)
					}
				}
				for _, min := range v.OddsMin.SubIds {
					if game.start == min {
						ele = model.ElementsAll.GetById(v.Id, &isSmall)
					}
				}
			}
			specType = "起始点：" + getNormalResult(ele)
		}

	}

	tmpl := fmt.Sprintf(
		`
开始时间：%s
结束时间：%s
当前作弊率：%v
开奖结果：
普通中奖：%s 开奖符号：%s。
特殊中奖：%s 中奖：%s 开奖符号：%v。`,
		game.startTime.Format("2006-01-02 15:04:05"),
		game.endTime.Format("2006-01-02 15:04:05"),
		roomProb,
		normal, normalResult,
		spec, specResult, specType,
	)

	for _, user := range game.UserMap {
		if user.user.IsRobot() {
			continue
		}
		userTmpl := fmt.Sprintf(
			`
用户ID：%v 总投注金额: %v 总获利金额：%v 用户剩余金额：%v
（投注区域：金额）：BAR：%v 77：%v 双星：%v 西瓜：%v 铃铛：%v 橘子：%v 柠檬：%v 苹果：%v`,
			user.user.GetID(),
			score.GetScoreStr(user.BetGoldNow),
			score.GetScoreStr(user.LastWinGold),
			score.GetScoreStr(user.user.GetScore()),
			score.GetScoreStr(user.BetInfo[0]),
			score.GetScoreStr(user.BetInfo[1]),
			score.GetScoreStr(user.BetInfo[2]),
			score.GetScoreStr(user.BetInfo[3]),
			score.GetScoreStr(user.BetInfo[4]),
			score.GetScoreStr(user.BetInfo[5]),
			score.GetScoreStr(user.BetInfo[6]),
			score.GetScoreStr(user.BetInfo[7]),
		)
		// tmpl += userTmpl
		game.table.WriteLogs(user.user.GetID(), userTmpl)
	}
	game.table.WriteLogs(0, tmpl)
}

func (game *Game) SendTop3User(user player.PlayerInterface) {
	msg := new(proto.TopUserRespMsg)
	var temp = make(UserList, 0, len(game.UserMap))

	game.userMux.RLock()
	defer game.userMux.RUnlock()

	for _, user := range game.UserMap {
		temp = append(temp, user)
	}

	sort.Sort(temp)

	l := 3
	if len(temp) < 3 {
		l = len(temp)
	}
	for i := 0; i < l; i++ {
		msg.UserList = append(msg.UserList, &proto.UserInfo{
			ID:       temp[i].user.GetID(),
			NickName: temp[i].user.GetNike(),
			Avatar:   temp[i].user.GetHead(),
			Gold:     temp[i].user.GetScore(),
		})
	}
	user.SendMsg(int32(proto.SendToClientMessageType_TopUserResp), msg)
}

func (game *Game) leftTime(user player.PlayerInterface) {
	msg := new(proto.BackInRespMsg)
	msg.LeftTime = int32(game.TimerJob.GetTimeDifference())
	user.SendMsg(int32(proto.SendToClientMessageType_BackInResp), msg)
}
