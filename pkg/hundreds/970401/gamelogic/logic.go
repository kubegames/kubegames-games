package gamelogic

import (
	"common/score"
	"fmt"
	"game_LaBa/benzbmw/config"
	"game_LaBa/benzbmw/model"
	proto "game_LaBa/benzbmw/msg"
	"game_frame_v2/game/clock"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/bitly/go-simplejson"
	protocol "github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	table table.TableInterface

	startTime time.Time
	endTime   time.Time

	UserMap map[int64]*User // 玩家列表

	UserBetInfo [BET_AREA_LENGHT]int64 // 玩家下注信息

	AIBetInfo   [BET_AREA_LENGHT]int64 // AI下注信息
	BetInfoTemp [BET_AREA_LENGHT]int64
	WinInfo     [BET_AREA_LENGHT]int64

	Status int32 //游戏状态

	TimerJob     *clock.Job // 定时
	loopBetTimer *clock.Job
	// MasterUser  *User                  // 神算子
	BigWinner *User   // 大赢家
	Trend     []int32 // 走势图
	BetArr    []int64 // 下注区筹码

	BetLimitInfo config.RoomRule // 下注限制

	topUser []*User // 上座的6位玩家

	wait        sync.WaitGroup
	specType    model.ElementType //特殊奖项类型，颜色/车型
	settleElems model.ElemBases   // 结算时摇中元素
	settleMsg   *proto.SettleMsg  // 结算时信息
	testIn      *proto.TestIn     // 测试的结算结果

	leaveUserID       []int64 // 离开用户id
	hasAfterSettle    bool
	hasAfterSettleMux sync.Mutex
}

func NewGame(table table.TableInterface) *Game {
	return &Game{
		table:       table,
		UserBetInfo: [BET_AREA_LENGHT]int64{},
		AIBetInfo:   [BET_AREA_LENGHT]int64{},
		Trend:       []int32{},
		BetArr:      []int64{},
		settleMsg:   new(proto.SettleMsg),
		UserMap:     make(map[int64]*User),
	}
}

func (game *Game) GetUser(user player.PlayerInterface) *User {

	u, ok := game.UserMap[user.GetId()]
	if !ok {
		u = NewUser(game, user)
		if user.IsRobot() {
			rb := NewRobot(game)
			aiUser := user.BindRobot(rb)
			rb.BindUser(aiUser)
		}
		game.UserMap[user.GetId()] = u
	} else {
		game.UserMap[user.GetId()].user = user
	}
	// 重新进入检查一下是否在移除队列中
	for i, v := range game.leaveUserID {
		if v == user.GetId() {
			game.leaveUserID = append(game.leaveUserID[:i], game.leaveUserID[i+1:]...)
		}
	}
	return u
}

// 发送场景消息
func (game *Game) SendSceneMsg(user player.PlayerInterface) {
	game.GetUser(user)
	msg := new(proto.RoomSceneInfo)
	msg.Bets = game.BetArr
	msg.Gold = user.GetScore()
	msg.OnlineCount = int32(len(game.UserMap))

	var topUsers []*proto.SettleTopUserInfo
	game.topUser = game.GetTopUser()
	for _, v := range game.topUser {
		var isBigWinner bool
		if game.BigWinner != nil && game.BigWinner.user.GetId() == v.user.GetId() {
			isBigWinner = true
		}
		topUsers = append(topUsers, &proto.SettleTopUserInfo{
			UserID:      v.user.GetId(),
			WinGold:     v.LastWinGold,
			TakeGold:    v.user.GetScore(),
			Avatar:      v.user.GetHead(),
			NickName:    v.user.GetNike(),
			IsBigWinner: isBigWinner,
		})
	}
	u := game.GetUser(user)

	msg.MyBets = make([]*proto.SceneMyBet, BET_AREA_LENGHT)
	for index, v := range u.BetInfo {
		msg.MyBets[index] = &proto.SceneMyBet{
			AllGold:  game.UserBetInfo[index] + game.AIBetInfo[index],
			UserID:   user.GetId(),
			UserGold: v,
		}
	}
	msg.TopUsers = topUsers

	game.hasAfterSettleMux.Lock()
	if game.hasAfterSettle {
		msg.Trend = game.getTrendTop()
	} else {
		msg.Trend = game.getTrendTop()
		if game.settleMsg != nil && len(game.settleMsg.ShakeResult) > 0 {
			msg.Trend = append(msg.Trend, game.settleMsg.ShakeResult[0])
		}
	}
	game.hasAfterSettleMux.Unlock()
	msg.Odds = model.GetOdds()

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

	u := game.GetUser(user)

	totalBet := [BET_AREA_LENGHT]int64{}
	for i := range game.UserBetInfo {
		totalBet[i] = game.UserBetInfo[i] + game.AIBetInfo[i]
	}

	_, betMinLimit := game.getBaseBet()
	if u.DoBet(msg, totalBet, int64(betMinLimit)) {
		betGold := game.BetArr[msg.BetIndex]
		// 增加玩家戏下注
		if u.user.IsRobot() {
			game.AIBetInfo[msg.BetType] += betGold
		} else {
			game.UserBetInfo[msg.BetType] += betGold
		}
		u.BetInfo[msg.BetType] += betGold
		msg.UserID = u.user.GetId()

		// 如果在座位上，下注消息进行广播
		for _, topUser := range game.topUser {
			if topUser.user.GetId() == u.user.GetId() {
				game.table.Broadcast(int32(proto.SendToClientMessageType_BetRet), msg)
				return
			}
		}
		// 不在，则返回
		u.user.SendMsg(int32(proto.SendToClientMessageType_BetRet), msg)
	}
}

func (game Game) getBaseBet() (int64, int) {
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
	game.hasAfterSettle = false
	game.SendTopUserMsg()
	game.table.StartGame()
	game.startTime = time.Now()
	game.goldNowNotice()
	game.Status = int32(proto.GameStatus_StartMovie)
	game.TimerJob, _ = game.table.AddTimer(time.Duration(config.BenzBMWConf.Taketimes.Startmove), game.StartBet)
	//  发送游戏状态
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Startmove))

}

// 开始动画
func (game *Game) StartBet() {
	game.Status = int32(proto.GameStatus_BetStatus)
	game.TimerJob, _ = game.table.AddTimer(time.Duration(config.BenzBMWConf.Taketimes.Startbet), game.EndBetMovie)
	//  发送游戏状态
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Startbet))
	game.loopBetTimer, _ = game.table.AddTimerRepeat(time.Duration(config.BenzBMWConf.Taketimes.LoopBetGap), 0, game.loopBroadcastBetInfo)
}

//结束动画
func (game *Game) EndBetMovie() {
	// 发送结算消息
	game.TimerJob, _ = game.table.AddTimer(time.Duration(config.BenzBMWConf.Taketimes.Endmove), game.SettleMsg)
	game.Status = int32(proto.GameStatus_EndBetMovie)
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Endmove))
	game.sendEndMsg()
	game.settleElems = nil
	game.wait.Add(1)
	go game.PreSettleRW()
}

// 此阶段同步中奖历史和玩家的输赢金额
func (game *Game) AfterSettle(_ []byte) {
	// log.Traceln("游戏转圈结束")
	// game.hasAfterSettleMux.Lock()
	// defer game.hasAfterSettleMux.Unlock()
	// if game.hasAfterSettle {
	// 	return
	// }
	// game.hasAfterSettle = true
	// game.checkUser()
	// game.setTrend()
	// game.endTime = time.Now()
	// game.settle4UserAfterGame()
	// game.clearUser()
	// game.findBigWinner()
	// game.Reset()
	// game.table.EndGame()

}

// 客户端可能不发信息
func (game *Game) AfterSettle2() {
	game.TimerJob, _ = game.table.AddTimer(time.Duration(1000), game.Start)
	game.hasAfterSettleMux.Lock()
	defer game.hasAfterSettleMux.Unlock()
	if !game.hasAfterSettle {
		game.hasAfterSettle = true
		game.checkUser()
		game.setTrend()
		game.endTime = time.Now()
		game.settle4UserAfterGame()
		game.clearUser()
		game.findBigWinner()
		game.Reset()
		game.table.EndGame()
	}
}

// 发送结算消息
func (game *Game) SettleMsg() {
	game.wait.Wait()
	toRandTime := config.BenzBMWConf.Taketimes.Endpay

	if len(game.settleElems) > 1 {
		if len(game.settleElems) == 4 {
			// 大三元
			toRandTime += config.BenzBMWConf.Taketimes.EndpayAdd
		} else {
			// 大四喜增加时间
			toRandTime += config.BenzBMWConf.Taketimes.EndpayAddFour
		}
	}
	game.TimerJob, _ = game.table.AddTimer(time.Duration(toRandTime-1000), game.AfterSettle2)

	game.Status = int32(proto.GameStatus_SettleStatus)
	game.SendStatusMsg(int32(config.BenzBMWConf.Taketimes.Endpay))

	// TODO：结算消息待定
	msg := new(proto.SettleMsg)

	userSettleInfos := make(UserSettleInfos, len(game.UserMap))

	var index int
	// 为每个用户结算
	for _, v := range game.UserMap {

		// 计算税后的金额
		taxMoney := v.LastWinGold * game.table.GetRoomRate() / 10000
		y := v.LastWinGold * game.table.GetRoomRate() % 10000
		if y > 0 {
			taxMoney += 1
		}
		v.taxGold = taxMoney
		output := v.LastWinGold - taxMoney
		fmt.Printf("发送结算消息时；用户id = %d , 赢 = %d , 税 = %d , output = %d , 税率 = %v\n", v.user.GetId(), v.LastWinGold, taxMoney, output, game.table.GetRoomRate())

		// 税后
		v.LastWinGold = output
		userSettleInfos[index] = &proto.UserSettleInfo{
			UserID:   v.user.GetId(),
			Avatar:   v.user.GetHead(),
			NickName: v.user.GetNike(),
			WinGold:  v.LastWinGold, // 最后一局赢得金额
		}
		index++
		// 在SendRecord之后发送
		v.sendChip()
	}

	sort.Sort(userSettleInfos)
	if len(userSettleInfos) >= 3 {
		msg.Head = userSettleInfos[:3]
	} else {
		msg.Head = userSettleInfos
	}

	var allZero bool
	for _, v := range msg.Head {
		if v.WinGold == 0 {
			allZero = true
		} else {
			allZero = false
			break
		}
	}
	// 如果获胜金额均为0，则结算排名前3按玩家携带金额前3进行排序
	if allZero {
		msg.Head = nil
		topUser := game.GetTopUser()
		userSettleInfos = nil
		for _, v := range topUser {
			userSettleInfos = append(userSettleInfos, &proto.UserSettleInfo{
				UserID:   v.user.GetId(),
				Avatar:   v.user.GetHead(),
				NickName: v.user.GetNike(),
				WinGold:  v.LastWinGold, // 最后一局赢得金额
			})
		}
		if len(userSettleInfos) >= 3 {
			msg.Head = userSettleInfos[:3]
		} else {
			msg.Head = userSettleInfos
		}
	}

	for _, v := range game.UserMap {
		// 设置bigWinner的值
		mymsg := new(proto.SettleMsg)
		mymsg.Self = &proto.UserSettleInfo{
			UserID:   v.user.GetId(),
			Avatar:   v.user.GetHead(),
			NickName: v.user.GetNike(),
			WinGold:  v.LastWinGold,
		}
		mymsg.Head = msg.Head
		mymsg.ShakeResult = game.settleMsg.ShakeResult
		mymsg.GoldNow = v.user.GetScore() + v.LastWinGold // 此时还未调用SetScore()增加金额

		v.user.SendMsg(int32(proto.SendToClientMessageType_Settle), mymsg)
		// 发送打码量

	}

	game.SendTopUserWin()
}

func (game *Game) setTrend() {
	game.Trend = append(game.Trend, game.settleMsg.ShakeResult[0])
	if len(game.Trend) > TREND_LENGTH {
		temp := game.Trend[len(game.Trend)-TREND_LENGTH:]
		game.Trend = make([]int32, TREND_LENGTH)
		copy(game.Trend, temp)
	}
}

func (game Game) getTrendTop() []int32 {
	if len(game.Trend) >= TREND_TOP_LENGTH {
		return game.Trend[len(game.Trend)-TREND_TOP_LENGTH:]
	}
	return game.Trend
}

func (game Game) getBetAllUser() (allBet int64) {
	for _, v := range game.UserBetInfo {
		allBet += v
	}
	return
}

// 获取上座的6位玩家（包括1位大赢家和5位携带金币最多的玩家）
func (game *Game) GetTopUser() []*User {
	return game.topUser
}

// 重新计算上座的6位玩家（包括1位大赢家和5位携带金币最多的玩家）
func (game *Game) CalcTopUser() []*User {
	topUsers := make(TopUsers, 0)
	var result []*User

	if game.BigWinner == nil {
		for _, v := range game.UserMap {
			topUsers = append(topUsers, v)
		}
		sort.Sort(topUsers)
	} else {
		topUsers = append(topUsers, game.BigWinner)
		for _, v := range game.UserMap {
			if v.user.GetId() == game.BigWinner.user.GetId() {
				continue
			}
			topUsers = append(topUsers, v)
		}
		if len(topUsers) > 1 {
			sort.Sort(topUsers[1:])
		} else {
			sort.Sort(topUsers)
		}
	}

	if len(topUsers) >= TOP_USER_LENGTH {
		result = topUsers[:TOP_USER_LENGTH]
	} else {
		result = topUsers
	}

	return result
}

func (game Game) SendTrendMsg(user player.PlayerInterface) {
	msg := new(proto.TrendMsg)
	msg.Trend = game.Trend
	user.SendMsg(int32(proto.SendToClientMessageType_TrendRet), msg)
}

// 发送上座玩家的消息
func (game *Game) SendTopUserMsg() {
	msg := new(proto.SettleTopUser)

	game.topUser = game.CalcTopUser()
	for _, v := range game.topUser {
		var isBigWinner bool
		if game.BigWinner != nil && game.BigWinner.user.GetId() == v.user.GetId() {
			isBigWinner = true
		}

		msg.List = append(msg.List, &proto.SettleTopUserInfo{
			UserID:      v.user.GetId(),
			TakeGold:    v.user.GetScore(),
			Avatar:      v.user.GetHead(),
			NickName:    v.user.GetNike(),
			IsBigWinner: isBigWinner,
		})
	}

	log.Traceln("发送6位上座玩家的结算信息=======================", msg)
	game.table.Broadcast(int32(proto.SendToClientMessageType_TopUserList), msg)
}

// 发送上座6位玩家赢得信息
func (game *Game) SendTopUserWin() {
	msg := new(proto.SettleTopUser)

	var isBigWinner bool
	users := game.GetTopUser()
	for _, v := range users {
		if v == game.BigWinner {
			isBigWinner = true
		} else {
			isBigWinner = false
		}
		msg.List = append(msg.List, &proto.SettleTopUserInfo{
			UserID:      v.user.GetId(),
			WinGold:     v.LastWinGold,
			TakeGold:    v.user.GetScore(),
			Avatar:      v.user.GetHead(),
			NickName:    v.user.GetNike(),
			IsBigWinner: isBigWinner,
		})
	}
	game.table.Broadcast(int32(proto.SendToClientMessageType_TopUserList), msg)
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
	if game.settleMsg != nil {
		game.settleMsg.Reset()
		game.settleElems = nil
	}
	game.UserBetInfo = [BET_AREA_LENGHT]int64{}
	game.AIBetInfo = [BET_AREA_LENGHT]int64{}
	game.BetInfoTemp = [BET_AREA_LENGHT]int64{}
	game.WinInfo = [BET_AREA_LENGHT]int64{}
	game.testIn = nil
	game.specType = model.ElementTypeNil
	for _, v := range game.UserMap {
		v.Reset()
	}
	count = [12]int{}
}

func (game *Game) getUserList(buf []byte, user player.PlayerInterface) {

	// args := new(proto.UserListReq)
	// if err := protocol.Unmarshal(buf, args); err != nil {
	// 	log.Errorf("获取用户列表消息解析错误：", err)
	// 	return
	// }

	// if args.PageIndex <= 0 {
	// 	return
	// }
	// if int(args.PageIndex)-USER_PAGE_LIMIT < 0 {
	// 	return
	// }

	var temp = make(UserList, 0, len(game.UserMap))
	msg := new(proto.UserListResp)

	if game.BigWinner == nil {
		for _, v := range game.UserMap {
			// temp = append(temp, v)
			temp = append(temp, v)
		}
		sort.Sort(temp)
	} else {
		temp = append(temp, game.BigWinner)
		for _, v := range game.UserMap {
			// temp = append(temp, v)
			if v.user.GetId() == game.BigWinner.user.GetId() {
				continue
			}
			temp = append(temp, v)
		}
		sort.Sort(temp[1:])
	}

	// fmt.Printf("请求玩家列表   ===== %v   ====%v  ", args.PageIndex, len(temp))
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
		ui.ID = v.user.GetId()
		ui.WinGold = v.WinGold
		ui.WinTimes = v.WinTimes
		ui.Avatar = v.user.GetHead()
		msg.UserList = append(msg.UserList, ui)
	}

	user.SendMsg(int32(proto.SendToClientMessageType_UserList), msg)
}

func (game *Game) UserOut(user player.PlayerInterface) {
	delete(game.UserMap, user.GetId())
}

// 计算赢的金额
func (game *Game) calcWinGold() {
	for _, v := range game.UserMap {
		var shakeResult model.ElemBases
		if len(game.settleElems) > 1 {
			// 中了大三元/大四喜
			shakeResult = game.settleElems[1:]
		} else {
			//
			shakeResult = game.settleElems
		}
		for _, sr := range shakeResult {
			v.LastWinGold += v.BetInfo[sr.BetIndex] * int64(sr.Odds)

			taxMoney := v.BetInfo[sr.BetIndex] * int64(sr.Odds) * game.table.GetRoomRate() / 10000
			y := v.LastWinGold * game.table.GetRoomRate() % 10000
			if y > 0 {
				taxMoney += 1
			}

			v.WinInfo[sr.BetIndex] += v.BetInfo[sr.BetIndex]*int64(sr.Odds) - taxMoney
			game.WinInfo[sr.BetIndex] += v.BetInfo[sr.BetIndex] * int64(sr.Odds)
		}
	}
}

// 剔除不在额定金额内的机器人，剔除n局未下注的玩家
func (game *Game) checkUser() {

	var deleteQueue []int64

	// 读map
	for k, v := range game.UserMap {

		level := game.table.GetLevel()
		min, max := config.RobotConf.EvictGold[level-1][0], config.RobotConf.EvictGold[level-1][1]
		if v.user.IsRobot() && (v.user.GetScore() < min || v.user.GetScore() > max) {
			// 机器人直接剔除
			game.table.KickOut(v.user)
			v.Reset()
			deleteQueue = append(deleteQueue, k)
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
				// delete(game.UserMap, k)
				deleteQueue = append(deleteQueue, k)
				continue
			}
			// 发送剔除用户消息
			msg := new(proto.BetFailMsg)
			msg.BetFailInfo = "最近五局未下注，已被移出房间。"
			msg.IsKickOut = true
			//game.Table.KickOut(v.User)
			v.Reset()
			v.NotBetCount = 0
			delete(game.UserMap, k)
			deleteQueue = append(deleteQueue, k)
			v.user.SendMsg(int32(proto.SendToClientMessageType_BetFailID), msg)
		}
	}
	game.leaveUserID = append(game.leaveUserID, deleteQueue...)
}

// 找出大赢家，近20局赢钱最多的玩家，每局结算后调用
func (game *Game) findBigWinner() {
	var mostWinnerID, mostWinnerGold int64
	for id, v := range game.UserMap {
		if v.WinGold > mostWinnerGold {
			mostWinnerGold = v.WinGold
			mostWinnerID = id
		}
	}
	game.BigWinner = game.UserMap[mostWinnerID]
}

// 循环播放下注消息
func (game *Game) loopBroadcastBetInfo() {
	if game.Status != int32(proto.GameStatus_BetStatus) {
		if game.loopBetTimer != nil {
			game.loopBetTimer.Cancel()
			game.loopBetTimer = nil
		}
		return
	}
	// 小于2s的时候不再循环发送广播
	// if game.TimerJob.GetTimeDifference() < 2000 {
	// 	game.loopBetTimer.Cancel()
	// 	game.loopBetTimer = nil
	// 	return
	// }

	// 7个座位上玩家的下注金额
	// var topUserBet = make([]int64, BET_AREA_LENGHT)
	// for _, v := range game.topUser {
	// 	for index, gold := range v.BetInfo {
	// 		topUserBet[index] += gold
	// 	}
	// }

	msg := new(proto.LoopBetNoticeMsg)
	msg.BetGold = make([]int64, BET_AREA_LENGHT)
	for index, v := range game.UserBetInfo {
		msg.BetGold[index] = v + game.AIBetInfo[index] //  - topUserBet[index]
		// game.BetInfoTemp[index] = v + game.AIBetInfo[index]
	}

	// log.Traceln("循环广播 ===========  ", msg)

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
	var msg proto.TestIn
	err := protocol.Unmarshal(bts, &msg)
	if err != nil {
		return
	}

	log.Traceln("测试消息 ====== ", msg)
	game.testIn = &msg

}

func (game *Game) TestPreSettle() {
	if game.testIn == nil {
		return
	}
	if !(game.testIn.OutID >= 0 && game.testIn.OutID <= 25) {
		return
	}

	game.settleMsg.Reset()

	var element model.ElemBases
	game.settleElems = nil
	roomProbNow, _ := game.table.GetRoomProb()
	if roomProbNow == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProbNow)
		roomProbNow = 1000
	}
	winCtrl := config.BenzBMWConf.WinCtrl.Find(int(roomProbNow))
	if game.testIn.OutID == 0 {
		// 大三元
		element = model.ElemBases{model.ElemThree}
		var eles model.ElemBases
		// threeElem:=fmt.Sprintf("%x",winCtrl.Surprise.ThreeCtrl.Rand().ElemType)
		eles, game.specType = model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.ThreeCtrl.Rand().ElemType))
		element = append(element, eles...)
		game.settleElems = element
		for _, v := range game.settleElems {
			game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, int32(v.RandSubId()))
		}
	} else if game.testIn.OutID == 13 {
		// 中大四喜
		element = model.ElemBases{model.ElemFour}
		var eles model.ElemBases
		eles, game.specType = model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.FourCtrl.Rand().Color))
		element = append(element, eles...)
		game.settleElems = element
		for _, v := range game.settleElems {
			game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, int32(v.RandSubId()))
		}
	} else {
		var eleIndex model.ElementType
		for _, car := range config.BenzBMWConf.Cars {
			for _, subId := range car.SubIds {
				if int32(subId) == game.testIn.OutID {
					eleIndex = car.ElemType
				}
			}
		}
		game.settleElems, game.specType = config.BenzBMWConf.Cars.FindWithType(eleIndex)
		game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, game.testIn.OutID)
	}
	log.Traceln("处理测试消息得结果 =====", game.settleMsg.ShakeResult)
}

// 结算规则
func (game *Game) PreSettleRW() {
	// game.TimerJob, _ = game.table.AddTimer(time.Duration(config.BenzBMWConf.Taketimes.Endmove-1000), game.SettleMsg)
	defer game.wait.Done()
	roomProbNow, _ := game.table.GetRoomProb()
	if roomProbNow == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProbNow)
		roomProbNow = 1000
	}
	winCtrl := config.BenzBMWConf.WinCtrl.Find(int(roomProbNow))
	var element model.ElemBases

	surpriseProb := model.Rand(PROB_BASE)

	if config.BenzBMWConf.IsOpen3000Ctrl && roomProbNow == 3000 {
		// 不能中特殊奖项
		surpriseProb = PROB_BASE * 2
	}
	// 中特殊奖项
	if surpriseProb <= winCtrl.Surprise.Prob {
		if model.Rand(PROB_BASE) <= winCtrl.Surprise.Three {
			// 中大三元
			element = model.ElemBases{model.ElemThree}
			var eles model.ElemBases
			eles, game.specType = model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.ThreeCtrl.Rand().ElemType))
			// threeElem:=fmt.Sprintf("%x",winCtrl.Surprise.ThreeCtrl.Rand().ElemType)
			element = append(element, eles...)
		} else {
			// 中大四喜
			element = model.ElemBases{model.ElemFour}
			var eles model.ElemBases
			eles, game.specType = model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.FourCtrl.Rand().Color))
			// fourElem:=fmt.Sprintf("%x",winCtrl.Surprise.FourCtrl.Rand().Color)
			element = append(element, eles...)
		}
	} else {
		element = game.getEndGameResult(winCtrl)
	}
	game.settleElems = element
	for _, v := range game.settleElems {
		game.settleMsg.ShakeResult = append(game.settleMsg.ShakeResult, int32(v.RandSubId()))
	}
	game.TestPreSettle()
	game.calcWinGold()
}

func (game *Game) getEndGameResult(winCtrl *config.WinControl) (element model.ElemBases) {
	roomProbNow, _ := game.table.GetRoomProb()
	if roomProbNow == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProbNow)
		roomProbNow = 1000
	}
	allBet := game.getBetAllUser()
	oddsList := model.GetOdds()
	model.ReverseOdds(oddsList)
	var waitCheck []int
	var backProb = [12]int64{} // 返奖率
	waitCheck = nil

	if allBet != 0 { // 计算返奖率
		for index, v := range game.UserBetInfo {
			backProb[index] = int64(float64(v*int64(oddsList[index])) / float64(allBet) * float64(100))
		}
	}

	back := winCtrl.Back.Rand()
	backMin, backMax := int64(back.Min), int64(back.Max)

	for index, v := range backProb {
		if v >= backMin && v < backMax {
			waitCheck = append(waitCheck, index)
		}
	}

	// 对于3000作弊率开启必输的处理
	if config.BenzBMWConf.IsOpen3000Ctrl && roomProbNow == 3000 {
		// 重置waitCheck，找出返奖率小于100的
		waitCheck = nil
		for i, v := range backProb {
			if v < 100 {
				waitCheck = append(waitCheck, i)
			}
		}

		// 如果没有返奖率小于100的值，则找出最小返奖率值
		if len(waitCheck) == 0 {
			indexMin, backMin := 0, backProb[0]
			for i, v := range backProb {
				if v < backMin {
					indexMin = i
					backMin = v
				}
			}
			waitCheck = append(waitCheck, indexMin)
		}
	}

	if len(waitCheck) != 0 {
		// 这里随机选择一个进行返回
		index := waitCheck[rand.Intn(len(waitCheck))]
		element = model.ElemBases{config.BenzBMWConf.Cars.GetByID(11 - index)}
		return
	}
	element = model.ElemBases{config.BenzBMWConf.Cars.RandResult(model.ElementTypeNil, true)}
	return
}

func (game *Game) settle4UserAfterGame() {
	log.Traceln("为每个用户进行结算")
	for _, v := range game.UserMap {
		if v.LastWinGold > 0 {
			game.PaoMaDeng(v.LastWinGold, v.user)
		}
		// 设置每个玩家的分数
		// 此处lastWinGold是已经扣过税了
		fmt.Printf("用户id == %d , 赢  == %d\n", v.user.GetId(), v.LastWinGold)
		v.user.SetScore(game.table.GetGameNum(), v.LastWinGold, 0)
		v.user.SendRecord(game.table.GetGameNum(), v.LastWinGold-v.BetGoldNow, v.BetGoldNow, v.taxGold, v.LastWinGold, "")
		var allBet int64
		for _, val := range v.BetInfo {
			allBet += val
		}
		if v.LastWinGold > allBet {
			v.SyncWinTimes(1)
			v.SyncWinData(v.LastWinGold - allBet)
		} else {
			v.SyncWinTimes(0)
			v.SyncWinData(0)
		}
	}
	log.Traceln("为每个用户进行结算结束")
	game.writeLog()
}

func (game *Game) BetRept(bts []byte, user player.PlayerInterface) {

	msg := new(proto.BetReptMsg)
	if err := protocol.Unmarshal(bts, msg); err != nil {
		return
	}
	if len(msg.BetGold) != BET_AREA_LENGHT {
		return
	}

	u := game.GetUser(user)

	var allBet int64
	for _, v := range msg.BetGold {
		if v < 0 {
			return
		}
		allBet += v
	}
	// 钱不够
	if allBet > user.GetScore() {
		u.SendBetFailed("余额不足")
	}
	for i := 0; i < BET_AREA_LENGHT; i++ {
		u.BetInfo[i] += msg.BetGold[i]
		if user.IsRobot() {
			game.AIBetInfo[i] += msg.BetGold[i]
		} else {
			game.UserBetInfo[i] += msg.BetGold[i]
		}
	}

	u.BetGoldNow += allBet
	u.user.SetScore(u.game.table.GetGameNum(), -1*allBet, 0)

	for _, user := range game.topUser {
		if user.user.GetId() == u.user.GetId() {
			mymsg := new(proto.BetReptRespNoticeMsg)
			mymsg.UserID = u.user.GetId()
			mymsg.BetGold = msg.BetGold
			game.table.Broadcast(int32(proto.SendToClientMessageType_BetReptRespNotice), mymsg)
		}
	}
	user.SendMsg(int32(proto.SendToClientMessageType_BetReptResp), msg)
}

// 清除离开用户，每局游戏结束时清楚
func (game *Game) clearUser() {

	for _, id := range game.leaveUserID {
		if user, ok := game.UserMap[id]; ok {
			user.Reset()
		}
		delete(game.UserMap, id)
		if game.BigWinner != nil && game.BigWinner.user.GetId() == id {
			game.BigWinner = nil
		}
	}
	// 清空列表
	game.leaveUserID = nil
}

func (game *Game) goldNowNotice() {
	for _, user := range game.UserMap {
		msg := new(proto.GoldNowMsg)
		msg.UserID = user.user.GetId()
		msg.GoldNow = user.user.GetScore()
		user.user.SendMsg(int32(proto.SendToClientMessageType_GoldNowNotice), msg)
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

var AreaName = []string{"红色奔驰", "蓝色奔驰", "灰色奔驰", "红色宝马", " 蓝色宝马", " 灰色宝马", " 红色雷克萨斯", " 蓝色雷克萨斯", " 灰色雷克萨斯", " 红色大众", " 蓝色大众", " 灰色大众"}

func (game Game) writeLog() {
	roomProb, _ := game.table.GetRoomProb()

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

	if isNormal {
		switch game.settleElems[0].ElemType {
		case model.BenzRed:
			normalResult = "红色奔驰"
		case model.BenzGreen:
			normalResult = "蓝色奔驰"
		case model.BenzBlack:
			normalResult = "灰色奔驰"
		case model.BMWRed:
			normalResult = "红色宝马"
		case model.BMWGreen:
			normalResult = "蓝色宝马"
		case model.BMWBlack:
			normalResult = "灰色宝马"
		case model.LexusRed:
			normalResult = "红色雷克萨斯"
		case model.LexusGreen:
			normalResult = "蓝色雷克萨斯"
		case model.LexusBlack:
			normalResult = "灰色雷克萨斯"
		case model.VWRed:
			normalResult = "红色大众"
		case model.VWGreen:
			normalResult = "蓝色大众"
		case model.VWBlack:
			normalResult = "灰色大众"
		}
	} else {
		switch game.settleElems[0].ElemType {
		case model.BigThreeElem:
			specResult = "大三元"
		case model.BigFourElem:
			specResult = "大四喜"
		}

		switch game.specType {
		case model.ElementTypeColorRed:
			specType = "红色"
		case model.ElementTypeColorGreed:
			specType = "蓝色"
		case model.ElementTypeColorBlack:
			specType = "灰色"
		case model.ElementTypeCarBenz:
			specType = "奔驰"
		case model.ElementTypeCarBMW:
			specType = "宝马"
		case model.ElementTypeCarLexus:
			specType = "雷克萨斯"
		case model.ElementTypeCarVW:
			specType = "大众"
		}
	}

	var BetAreaCount [BET_AREA_LENGHT]int
	var tmpl string

	var SysWin int64
	for i := 0; i < BET_AREA_LENGHT; i++ {
		SysWin += game.UserBetInfo[i] + game.AIBetInfo[i]
		SysWin -= game.WinInfo[i]
	}
	tmpl += fmt.Sprintf("系统输赢额度:%v;", SysWin)

	var WinMostUserID int64
	var WinMostGold int64
	for _, user := range game.UserMap {
		if user.user.IsRobot() {
			continue
		}

		userTmpl := fmt.Sprintf("用户ID：%v 投注金额:[", user.user.GetID())
		for i, v := range user.BetInfo {
			if v != 0 {
				BetAreaCount[i]++
				userTmpl += fmt.Sprintf("%v:%v ", AreaName[i], score.GetScoreStr(user.BetInfo[i]))
			}
		}
		userTmpl += fmt.Sprintf("]，输赢:[")

		for i, v := range user.BetInfo {
			if v != 0 {
				Win := -v
				if user.WinInfo[i] > 0 {
					Win = user.WinInfo[i]
				}
				userTmpl += fmt.Sprintf("%v:%v ", AreaName[i], score.GetScoreStr(Win))
			}
		}

		userTmpl += fmt.Sprintf("]，总输赢：%v;", user.LastWinGold)

		game.table.WriteLogs(user.user.GetId(), userTmpl)
		if user.LastWinGold > WinMostGold {
			WinMostGold = user.LastWinGold
			WinMostUserID = user.user.GetID()
		}
	}

	tmpl = fmt.Sprintf(`红色奔驰：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	蓝色奔驰：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v；  
	灰色奔驰：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	红色宝马：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
 	蓝色宝马：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	灰色宝马：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	红色雷克萨斯：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	蓝色雷克萨斯：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
 	灰色雷克萨斯：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	红色大众：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	蓝色大众：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v； 
	灰色大众：[总：%v 机器人：%v 真人：%v 下注真人人数：%v]，输赢：%v；`,
		score.GetScoreStr(game.UserBetInfo[0]+game.AIBetInfo[0]), score.GetScoreStr(game.AIBetInfo[0]), score.GetScoreStr(game.UserBetInfo[0]), BetAreaCount[0], score.GetScoreStr(game.UserBetInfo[0]+game.AIBetInfo[0]-game.WinInfo[0]),
		score.GetScoreStr(game.UserBetInfo[1]+game.AIBetInfo[1]), score.GetScoreStr(game.AIBetInfo[1]), score.GetScoreStr(game.UserBetInfo[1]), BetAreaCount[1], score.GetScoreStr(game.UserBetInfo[1]+game.AIBetInfo[1]-game.WinInfo[1]),
		score.GetScoreStr(game.UserBetInfo[2]+game.AIBetInfo[2]), score.GetScoreStr(game.AIBetInfo[2]), score.GetScoreStr(game.UserBetInfo[2]), BetAreaCount[2], score.GetScoreStr(game.UserBetInfo[2]+game.AIBetInfo[2]-game.WinInfo[2]),
		score.GetScoreStr(game.UserBetInfo[3]+game.AIBetInfo[3]), score.GetScoreStr(game.AIBetInfo[3]), score.GetScoreStr(game.UserBetInfo[3]), BetAreaCount[3], score.GetScoreStr(game.UserBetInfo[3]+game.AIBetInfo[3]-game.WinInfo[3]),
		score.GetScoreStr(game.UserBetInfo[4]+game.AIBetInfo[4]), score.GetScoreStr(game.AIBetInfo[4]), score.GetScoreStr(game.UserBetInfo[4]), BetAreaCount[4], score.GetScoreStr(game.UserBetInfo[4]+game.AIBetInfo[4]-game.WinInfo[4]),
		score.GetScoreStr(game.UserBetInfo[5]+game.AIBetInfo[5]), score.GetScoreStr(game.AIBetInfo[5]), score.GetScoreStr(game.UserBetInfo[5]), BetAreaCount[5], score.GetScoreStr(game.UserBetInfo[5]+game.AIBetInfo[5]-game.WinInfo[5]),
		score.GetScoreStr(game.UserBetInfo[6]+game.AIBetInfo[6]), score.GetScoreStr(game.AIBetInfo[6]), score.GetScoreStr(game.UserBetInfo[6]), BetAreaCount[6], score.GetScoreStr(game.UserBetInfo[6]+game.AIBetInfo[6]-game.WinInfo[6]),
		score.GetScoreStr(game.UserBetInfo[7]+game.AIBetInfo[7]), score.GetScoreStr(game.AIBetInfo[7]), score.GetScoreStr(game.UserBetInfo[7]), BetAreaCount[7], score.GetScoreStr(game.UserBetInfo[7]+game.AIBetInfo[7]-game.WinInfo[7]),
		score.GetScoreStr(game.UserBetInfo[8]+game.AIBetInfo[8]), score.GetScoreStr(game.AIBetInfo[8]), score.GetScoreStr(game.UserBetInfo[8]), BetAreaCount[8], score.GetScoreStr(game.UserBetInfo[8]+game.AIBetInfo[8]-game.WinInfo[8]),
		score.GetScoreStr(game.UserBetInfo[9]+game.AIBetInfo[9]), score.GetScoreStr(game.AIBetInfo[9]), score.GetScoreStr(game.UserBetInfo[9]), BetAreaCount[9], score.GetScoreStr(game.UserBetInfo[9]+game.AIBetInfo[9]-game.WinInfo[9]),
		score.GetScoreStr(game.UserBetInfo[10]+game.AIBetInfo[10]), score.GetScoreStr(game.AIBetInfo[10]), score.GetScoreStr(game.UserBetInfo[10]), BetAreaCount[10], score.GetScoreStr(game.UserBetInfo[10]+game.AIBetInfo[10]-game.WinInfo[10]),
		score.GetScoreStr(game.UserBetInfo[11]+game.AIBetInfo[11]), score.GetScoreStr(game.AIBetInfo[11]), score.GetScoreStr(game.UserBetInfo[11]), BetAreaCount[11], score.GetScoreStr(game.UserBetInfo[11]+game.AIBetInfo[11]-game.WinInfo[11]))

	tmpl += fmt.Sprintf(`
	当前作弊率：%v
	开奖结果：
	普通中奖：%s 开奖符号：%s。
	特殊中奖：%s 中奖：%s 开奖符号：%v。`,
		roomProb,
		normal, normalResult,
		spec, specResult, specType)
	tmpl += fmt.Sprintf("最高获利用户ID：%v 获得:%v;", WinMostUserID, score.GetScoreStr(WinMostGold))
	game.table.WriteLogs(0, tmpl)
}
