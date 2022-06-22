package gamelogic

import (
	"fmt"
	"go-game-sdk/lib/clock"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/pkg/slots/990601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/990601/model"
	bridanimal "github.com/kubegames/kubegames-games/pkg/slots/990601/msg"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
)

const (
	// 下注区飞禽的index
	BIRD_BET_INDEX int = 8
	// 下注区走兽的index
	ANIMAL_BET_INDEX int = 9

	BET_AREA_LENGTH = 12

	PROB_BASE = 10000

	GOLD_SHARK_ID   = 2
	SILVER_SHARK_ID = 3

	ALL_KILL_ID = 12
	ALL_PAY_ID  = 13

	// 历史纪录长度
	TREND_LENGTH = 10
)

type Game struct {
	Table        table.TableInterface //桌子
	UserInfoList map[int64]*UserInfo  //用户列表(map)
	userLock     sync.RWMutex
	wait         sync.WaitGroup

	timeStart    time.Time
	timeEnd      time.Time
	Status       int32      //游戏状态
	TimerJob     *clock.Job //时间定时器
	LoopBetTimer *clock.Job // 下注广播消息定时器

	BetArr []int32 // 下注的筹码基数

	TotalBet     [BET_AREA_LENGTH]int64 // 玩家下注
	AITotalBet   [BET_AREA_LENGTH]int64 // AI下注
	TotalBetTemp [BET_AREA_LENGTH]int64 // 前 n 毫秒下注金额缓存
	BetLimitInfo config.RoomRule        // 下注限制

	Trend []int32 // 走势图

	// BirdAnimals model.Elements

	// for settle
	settleMsg      *bridanimal.SettleMsg // 结算消息
	settleElements []*model.Element      // 结算时摇中的元素

	testMsg *bridanimal.TestMsg // 作弊消息
}

//玩家下注
func (g *Game) OnUserBet(buffer []byte, user player.PlayerInterface) {

	if g.Status != int32(bridanimal.GameStatus_BetStatus) {
		return
	}

	BetInfo := &bridanimal.UserBet{}
	if err := proto.Unmarshal(buffer, BetInfo); err != nil {
		log.Errorf("下注消息解析失败：", err)
	}

	if BetInfo.BetIndex < 0 || BetInfo.BetIndex >= int32(len(g.BetArr)) || BetInfo.BetType >= BET_AREA_LENGTH || BetInfo.BetType < 0 {
		return
	}

	u := g.GetUserByUserID(user.GetID(), user)
	totalBet := [BET_AREA_LENGTH]int64{}
	for i := 0; i < BET_AREA_LENGTH; i++ {
		totalBet[i] = g.TotalBet[i] + g.AITotalBet[i]
	}

	_, betMinLimit := g.getBaseBet()

	if u.Bet(BetInfo, totalBet, int64(betMinLimit)) {
		if !u.User.IsRobot() {
			g.TotalBet[BetInfo.BetType] += int64(g.BetArr[BetInfo.BetIndex])
		} else {
			g.AITotalBet[BetInfo.BetType] += int64(g.BetArr[BetInfo.BetIndex])
		}
		u.User.SendMsg(int32(bridanimal.SendToClientMessageType_BetRet), BetInfo)
	}
}

func (g *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	rb := NewRobot(g)
	rb.BindUser(ai)
	return rb
}

//获取玩家
func (g *Game) GetUserByUserID(userid int64, user player.PlayerInterface) *UserInfo {
	u, ok := g.UserInfoList[userid]

	if !ok {
		u = new(UserInfo)
		u.g = g

		u.betMsg = make(chan *bridanimal.UserBet, 2000)
		u.User = user
		u.Totol = user.GetScore()
		u.NotBetNum = 0
		u.WinChan = make(chan int64, USER_CHAN_LENGTH)
		u.BetChan = make(chan int64, USER_CHAN_LENGTH)
		u.BetInfo = [BET_AREA_LENGTH]int64{}
		u.BetInfoTemp = [BET_AREA_LENGTH]int64{}
		u.ResetUserData()
		u.currGold = user.GetScore()

		g.userLock.Lock()
		defer g.userLock.Unlock()
		g.UserInfoList[userid] = u

	}
	return u
}

func (g *Game) SendStatusMsg(StatusTime int) {
	msg := new(bridanimal.StatusMessage)
	msg.Status = int32(g.Status)
	msg.StatusTime = int32(StatusTime)
	g.Table.Broadcast(int32(bridanimal.SendToClientMessageType_Status), msg)
}

// 随机各种动物的赔率
func (g *Game) RandOdds() {
	if g.Table.GetRoomID() == -1 {
		log.Debugf("房间ID是负一")
		g.Status = 0
		return
	}
	g.timeStart = time.Now()
	g.Table.StartGame()
	g.noticeGoldNow()
	// g.BirdAnimals.RandOddsAndProb()
	config.BirdAnimaConfig.BirdAnimals.RandOddsAndProb()
	// 随机之后发送场景信息
	g.TimerJob, _ = g.Table.AddTimer(int64(config.BirdAnimaConfig.Taketimes.StartRandOdds), g.Start)
	g.Status = int32(bridanimal.GameStatus_RandOdds)
	// 发送游戏状态
	g.SendStatusMsg(int(config.BirdAnimaConfig.Taketimes.StartRandOdds))
	// 发送一个RandomOdds消息
	g.SendOddsMsg()
}

//开始动画
func (g *Game) Start() {
	g.TimerJob, _ = g.Table.AddTimer(int64(config.BirdAnimaConfig.Taketimes.Startmove), g.StartBet)
	g.Status = int32(bridanimal.GameStatus_StartMovie)
	g.SendStatusMsg(int(config.BirdAnimaConfig.Taketimes.Startmove))
}

//开始下注
func (g *Game) StartBet() {
	g.TimerJob, _ = g.Table.AddTimer(int64(config.BirdAnimaConfig.Taketimes.Startbet), g.EndBetMovie)
	g.Status = int32(bridanimal.GameStatus_BetStatus)
	g.SendStatusMsg(int(config.BirdAnimaConfig.Taketimes.Startbet))
	g.LoopBetTimer, _ = g.Table.AddTimerRepeat(int64(config.BirdAnimaConfig.Taketimes.BetGapBroadcast), 0, g.loopBetMsg)
}

//结束动画
func (g *Game) EndBetMovie() {
	g.TimerJob, _ = g.Table.AddTimer(int64(config.BirdAnimaConfig.Taketimes.Endmove), g.SendSettleMsg)
	g.Status = int32(bridanimal.GameStatus_EndBetMovie)
	g.SendStatusMsg(int(config.BirdAnimaConfig.Taketimes.Endmove))
	g.wait.Add(1)
	go g.PreSettleRW()
	g.betEndMsg()
}

// 发送结算消息
func (g *Game) SendSettleMsg() {
	g.wait.Wait()
	toRandTime := config.BirdAnimaConfig.Taketimes.Endpay

	if len(g.settleMsg.Award) == 2 {
		toRandTime += config.BirdAnimaConfig.Taketimes.EndpayAdd
	}
	// 发送消息
	g.TimerJob, _ = g.Table.AddTimer(int64(toRandTime-1000), g.ReSettleStatus)
	g.Status = int32(bridanimal.GameStatus_SettleStatus)

	g.SendStatusMsg(int(toRandTime))
	g.setTrend(nil)
	g.settleForUser(g.settleElements, g.settleMsg)
	g.timeEnd = time.Now()
}

func (g *Game) ReSettleStatus() {
	g.TimerJob, _ = g.Table.AddTimer(int64(1000), g.RandOdds)
	// 先同步玩家的赢金额
	g.ReSettle()
	// 检查玩家下注信息
	g.checkUserBet()
	g.writeLog()
	g.Reset()
	g.Table.EndGame()
}

func (game *Game) SendUserListInfo(buf []byte, user player.PlayerInterface) {

	// args := new(bridanimal.UserListReq)
	// if err := proto.Unmarshal(buf, args); err != nil {
	// 	log.Errorf("获取用户列表参数错误 %v\n", err)
	// 	return
	// }
	// if args.PageIndex <= 0 {
	// 	return
	// }
	// // args.PageIndex += 1
	// if int(args.PageIndex)-USER_PAGE_LIMIT < 0 {
	// 	return
	// }

	var temp = make(UserInfos, len(game.UserInfoList))
	msg := new(bridanimal.UserListResp)

	var index int

	game.userLock.RLock()
	defer game.userLock.RUnlock()
	for _, v := range game.UserInfoList {
		// temp = append(temp, v)
		temp[index] = v
		index++
	}

	temp.Sort()

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
		ui := new(bridanimal.UserInfo)
		ui.NickName = v.User.GetNike()
		ui.Gold = v.User.GetScore()
		ui.ID = v.User.GetID()
		ui.WinGold = v.WinGold
		ui.BetGold = v.BetGold
		ui.Avatar = v.User.GetHead()
		msg.UserList = append(msg.UserList, ui)
	}
	user.SendMsg(int32(bridanimal.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) getRandResult(id int, rmShark bool) (*model.Element, int32) {
	element, subId := config.BirdAnimaConfig.BirdAnimals.RandResult(id, rmShark)
	game.Trend = append(game.Trend, int32(subId))
	return element, int32(subId)
}

func (game *Game) SendSceneMsg(user player.PlayerInterface) {
	msg := new(bridanimal.RoomSceneInfo)

	msg.Trend = game.getTrend()
	// 获取每个元素的赔率
	value := game.getRandOddsInfo()
	sort.Sort(value)
	msg.RandOdds = value
	// 设置
	var err error
	msg.Bets, err = game.getBetMod()
	if err != nil {
		log.Errorf("获取下注数组失败 err%v\n", err)
	}
	game.GetUserByUserID(user.GetID(), user)
	msg.OnlineCount = int32(len(game.UserInfoList))
	msg.SceneBets = make([]*bridanimal.SceneAllBets, len(game.TotalBet))
	if user != nil {
		u := game.UserInfoList[user.GetID()]
		msg.Gold = user.GetScore()

		for i, val := range game.TotalBet {
			msg.SceneBets[i] = &bridanimal.SceneAllBets{
				AllGold:  val + game.AITotalBet[i],
				UserID:   user.GetID(),
				UserGold: u.BetInfo[i],
			}
		}

		user.SendMsg(int32(bridanimal.SendToClientMessageType_RoomSenceInfo), msg)
	} else {

		for userID, v := range game.UserInfoList {
			msg.Gold = v.User.GetScore()
			for i, val := range game.TotalBet {
				msg.SceneBets[i] = &bridanimal.SceneAllBets{
					AllGold:  val + game.AITotalBet[i],
					UserID:   userID,
					UserGold: v.BetInfo[i],
				}
			}
			// 为每个用户发送下注的信息
			user.SendMsg(int32(bridanimal.SendToClientMessageType_RoomSenceInfo), msg)
		}
	}
}

// 获取随机倍率

func (game *Game) getRandOddsInfo() model.RandomOddss {
	result := make(model.RandomOddss, 0)

	for _, v := range config.BirdAnimaConfig.BirdAnimals {
		// 返回随机赔率，忽略掉通赔通杀
		if !(v.EType != model.ETypeAllKill && v.EType != model.ETypeAllPay) {
			continue
		}

		result = append(result, &bridanimal.RandomOdds{
			ID:   int32(v.ID),
			Odds: int32(v.OddsNow),
		})
	}

	result = result[:BET_AREA_LENGTH]
	sort.Sort(result)
	return result
}

func (game Game) getTrend() []int32 {
	if len(game.Trend) >= 10 {
		return game.Trend[len(game.Trend)-10:]
	}
	return game.Trend
}

func (game Game) getBetMod() ([]int64, error) {
	// 设置
	var result = make([]int64, len(game.BetLimitInfo.ChipArea))
	for i, v := range game.BetLimitInfo.ChipArea {
		result[i] = int64(v)
	}
	return result, nil
}

func (game Game) getSceneBet(user player.PlayerInterface) []*bridanimal.SceneAllBets {

	result := make([]*bridanimal.SceneAllBets, len(game.BetArr))
	return result
}

// 获取总下注金币
func (game Game) getAllBetGold() (all int64) {
	for _, v := range game.TotalBet {
		all += v
	}
	return
}

func (game *Game) checkUserBet() {
	var deleteUserID []int64

	for k, v := range game.UserInfoList {
		level := game.Table.GetLevel()
		min, max := config.Robot.EvictGold[level-1][0], config.Robot.EvictGold[level-1][1]
		if v.User.IsRobot() && (v.User.GetScore() < min || v.User.GetScore() > max) {
			v.ResetUserData()
			game.Table.KickOut(v.User)
			deleteUserID = append(deleteUserID, k)
			// game.userLock.Lock()
			// delete(game.UserInfoList, k)
			// game.userLock.Unlock()
			game.SendUserNumMsg()
			continue
		}

		// 重置未下注局数
		var orVal int64
		for _, val := range v.BetInfo {
			orVal += val
		}
		if orVal == 0 { // 未下注
			v.NotBetNum++
		} else {
			v.NotBetNum = 0
			continue
		}
		if int(v.NotBetNum) > config.BirdAnimaConfig.Unplacebetnum {

			// 直接踢出机器人
			if v.User.IsRobot() {
				v.ResetUserData()
				game.Table.KickOut(v.User)
				deleteUserID = append(deleteUserID, k)
				// game.userLock.Lock()
				// delete(game.UserInfoList, k)
				// game.userLock.Unlock()
				game.SendUserNumMsg()
				continue
			}
			// 发送剔除用户消息
			msg := new(bridanimal.BetFailMsg)
			msg.BetFailInfo = "最近五局未下注，已被移出房间。"
			msg.IsKickOut = true
			v.ResetUserData()
			v.NotBetNum = 0
			game.userLock.Lock()
			deleteUserID = append(deleteUserID, k)
			game.userLock.Unlock()
			v.User.SendMsg(int32(bridanimal.SendToClientMessageType_BetFailID), msg)
			game.SendUserNumMsg()

		}
		// 重置用户数据
		v.ResetUserData()
	}

	for _, v := range deleteUserID {
		delete(game.UserInfoList, v)
	}
}

func (game Game) SendOddsMsg() {
	msg := new(bridanimal.OddsMsg)
	value := game.getRandOddsInfo()
	sort.Sort(value)
	msg.RandOdds = value
	game.Table.Broadcast(int32(bridanimal.SendToClientMessageType_OddsInfo), msg)
}

func (g *Game) settleForUser(elements []*model.Element, msg *bridanimal.SettleMsg) {

	head := make(model.Heads, len(g.UserInfoList))

	getWinGold := func(g *Game, user *UserInfo, element *model.Element, award *bridanimal.AwardSettleInfo) (winGold int64) {

		switch element.EType {
		case model.ETypeAllKill: // 通杀
			winGold = 0
		case model.ETypeAllPay: // 通赔
			for id, v := range user.BetInfo {
				ele := config.BirdAnimaConfig.BirdAnimals.GetByID(id)
				winGold += int64(ele.OddsNow) * v
			}
		default:

			// 飞禽/走兽下注区赔付
			if element.EType == model.ETypeAnimal {
				// 走兽
				animalElement := config.BirdAnimaConfig.BirdAnimals.GetByID(ANIMAL_BET_INDEX)
				if animalElement != nil {
					winGold += user.BetInfo[ANIMAL_BET_INDEX] * int64(animalElement.OddsNow)
				}
			} else if element.EType == model.ETypeBird {
				// 飞禽
				birdElement := config.BirdAnimaConfig.BirdAnimals.GetByID(BIRD_BET_INDEX)
				if birdElement != nil {
					winGold += user.BetInfo[BIRD_BET_INDEX] * int64(birdElement.OddsNow)
				}
			}

			winGold += user.BetInfo[element.ID] * int64(award.Odds)
		}
		return
	}

	g.userLock.RLock()
	defer g.userLock.RUnlock()
	var i int
	for _, v := range g.UserInfoList {
		// allBetGold:=v.GetAllBetGold()
		// 中奖金币
		var awardGold int64
		// var err error
		if len(msg.Award) == 1 {
			awardGold += getWinGold(g, v, elements[0], msg.Award[0])

			taxMoney := awardGold * g.Table.GetRoomRate() / 10000
			y := awardGold * g.Table.GetRoomRate() % 10000
			if y > 0 {
				taxMoney += 1
			}
			v.taxGold = taxMoney
			v.LastWinGold = awardGold - v.taxGold
			// 结算上分
			// outputAmount, _ := v.User.SetScore(g.Table.GetGameNum(), awardGold, g.Table.GetRoomRate())
			// v.User.SendRecord(g.Table.GetGameNum(), outputAmount-v.BetGoldNow, v.BetGoldNow, awardGold-outputAmount, outputAmount, "")
			// awardGold = outputAmount
		} else if len(msg.Award) == 2 {
			// 暂存免费游戏上分值
			sharkWinGold := getWinGold(g, v, elements[0], msg.Award[0])
			// sharkWinGold:=awardGold
			var awardSettle int64
			if elements[1].EType == model.ETypeAllKill {
				awardSettle = 0
				sharkWinGold = 0
			} else {
				awardSettle = getWinGold(g, v, elements[1], msg.Award[1])
			}

			gold := sharkWinGold + awardSettle
			taxMoney := gold * g.Table.GetRoomRate() / 10000
			y := gold * g.Table.GetRoomRate() % 10000
			if y > 0 {
				taxMoney += 1
			}
			v.taxGold = taxMoney
			v.LastWinGold = gold - v.taxGold
			// 结算上分(抽税，鲨鱼包含在里面一起抽)
			// outputAmount, err = v.User.SetScore(g.Table.GetGameNum(), gold, g.Table.GetRoomRate())
			// if err != nil {
			// 	outputAmount, _ = v.User.SetScore(g.Table.GetGameNum(), gold, g.Table.GetRoomRate())
			// }
			// v.User.SendRecord(g.Table.GetGameNum(), outputAmount-v.BetGoldNow, v.BetGoldNow, gold-outputAmount, outputAmount, "")
			// awardGold = outputAmount
		}

		// 最后一次赢得金币=总赢得金币-总下注金币
		// v.LastWinGold = awardGold
		head[i] = &bridanimal.UserSettleInfo{
			Avatar:   v.User.GetHead(),
			NickName: v.User.GetNike(),
			WinGold:  v.LastWinGold,
		}

		// v.SyncDataWin(v.LastWinGold)
		i++
	}

	sort.Sort(head)

	// 发送消息时锁定用户

	for _, v := range g.UserInfoList {
		mymsg := new(bridanimal.SettleMsg)
		mymsg.Award = msg.Award

		if len(head) >= 3 {
			mymsg.Head = head[:3]
		} else {
			mymsg.Head = head
		}
		mymsg.Self = &bridanimal.UserSettleInfo{
			Avatar:   v.User.GetHead(),
			NickName: v.User.GetNike(),
			WinGold:  v.LastWinGold,
		}
		mymsg.GoldNow = v.User.GetScore() + v.LastWinGold
		// 跑马灯
		if v.LastWinGold > 0 {
			g.PaoMaDeng(v.LastWinGold, v.User)
		}

		// 飞禽和走兽下注区返回
		if len(msg.Award) == 1 {
			if elements[0].EType == model.ETypeAnimal && v.BetInfo[ANIMAL_BET_INDEX] != 0 {
				mymsg.BAType = bridanimal.BirdAnimalType_Animal
			} else if elements[0].EType == model.ETypeBird && v.BetInfo[BIRD_BET_INDEX] != 0 {
				mymsg.BAType = bridanimal.BirdAnimalType_Bird
			}
		} else if len(msg.Award) == 2 {
			if elements[1].EType == model.ETypeAnimal && v.BetInfo[ANIMAL_BET_INDEX] != 0 {
				mymsg.BAType = bridanimal.BirdAnimalType_Animal
			} else if elements[1].EType == model.ETypeBird && v.BetInfo[BIRD_BET_INDEX] != 0 {
				mymsg.BAType = bridanimal.BirdAnimalType_Bird
			}
		}

		// v.SyncDataBet(v.Totol)

		v.User.SendMsg(int32(bridanimal.SendToClientMessageType_Settle), mymsg)
		// 打码量
		// v.setChip()
	}
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(bridanimal.StatusMessage)
	msg.Status = int32(game.Status)

	// 四舍五入
	StatusTime += 500
	StatusTime = int(StatusTime/1000) * 1000
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(bridanimal.SendToClientMessageType_Status), msg)
}

func (game *Game) SetBetArr() {
	if len(game.BetArr) == 0 {
		game.BetArr = make([]int32, 0)
		for _, v := range game.BetLimitInfo.ChipArea {
			game.BetArr = append(game.BetArr, int32(v))
		}
	}
}

func (g *Game) PaoMaDeng(Gold int64, user player.PlayerInterface) {
	configs := g.Table.GetMarqueeConfig()
	for _, v := range configs {
		if Gold > v.AmountLimit {
			//log.Debugf("创建跑马灯")
			err := g.Table.CreateMarquee(user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
		}
	}
}

// 游戏结算重置游戏
func (game *Game) Reset() {
	// game.BirdAnimals.Reset()
	config.BirdAnimaConfig.BirdAnimals.Reset()
	for _, v := range game.UserInfoList {
		v.ResetUserData()
	}
	game.TotalBet = [BET_AREA_LENGTH]int64{}
	game.AITotalBet = [BET_AREA_LENGTH]int64{}
	game.TotalBetTemp = [BET_AREA_LENGTH]int64{}
	game.settleMsg = nil
	game.settleElements = nil
	game.testMsg = nil
}

func (game *Game) setTrend(elements []*model.Element) {
	if game.settleMsg == nil {
		return
	}
	for _, v := range game.settleMsg.Award {
		game.Trend = append(game.Trend, int32(v.AwardType))
	}
	if len(game.Trend) > TREND_LENGTH {
		temp := game.Trend[len(game.Trend)-TREND_LENGTH:]
		game.Trend = make([]int32, TREND_LENGTH)
		copy(game.Trend, temp)
	}
}

func (game Game) SendUserNumMsg() {
	msg := new(bridanimal.UserNumMsg)
	msg.Num = int32(len(game.UserInfoList))
	game.Table.Broadcast(int32(bridanimal.SendToClientMessageType_UserNum), msg)
}

type AreaGold struct {
	bridanimal.RandomOdds
	PayGold int64 // 赔付金币
	EType   model.EType
}
type AreaGolds []AreaGold

func (a AreaGolds) Less(i, j int) bool {
	if a[i].PayGold <= a[j].PayGold {
		return true
	}
	return false
}

func (a AreaGolds) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AreaGolds) Len() int {
	return len(a)
}

func (a AreaGolds) GetMax() AreaGolds {
	if len(a) == 1 {
		return a
	}
	return a[len(a)/2:]
}

func (a AreaGolds) GetMin() AreaGolds {
	if len(a) == 1 {
		return a
	}
	return a[:len(a)/2]
}

func (game *Game) InitRoomRule() {
	baseBet, _ := game.getBaseBet()
	game.BetLimitInfo.BaseBet = baseBet
	game.BetLimitInfo.ChipArea = [5]int64{
		baseBet * 1,
		baseBet * 10,
		baseBet * 50,
		baseBet * 100,
		baseBet * 500,
	}
	game.BetLimitInfo.BetArea = baseBet * 2000
	game.BetLimitInfo.AllBetAreaLimit = baseBet * 5000
}
func (game Game) getBaseBet() (int64, int) {
	str := game.Table.GetAdviceConfig()
	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		log.Errorf("解析房间配置失败 err%v\n", err)
		return 0, 0
	}
	baseBet, _ := js.Get("Bottom_Pouring").Int()
	betMinLimit, _ := js.Get("betMinLimit").Int()
	return int64(baseBet), betMinLimit
}

// 获取游戏玩家总下注
func (game Game) GetBetAll() int64 {
	var allBet int64
	for _, v := range game.TotalBet {
		allBet += v
	}
	return allBet
}

// // 开奖控制
// func (g *Game) PreSettleRW() {

// 	g.SendUserNumMsg()
// 	// 发送结算消息（减去之前预结算的1s）
// 	g.TimerJob, _ = g.Table.AddTimer(int64(config.BirdAnimaConfig.Taketimes.Endmove-1000), g.SendSettleMsg)

// 	// 预先结算出结果
// 	nowRoomProb, _ := g.Table.GetRoomProb()

// 	shakePolicy := config.BirdAnimaConfig.PolicyTree.Find(nowRoomProb)
// 	var element, lastElement *model.Element
// 	var id int
// 	var hasShark bool

// 	settleMsg := new(bridanimal.SettleMsg)
// 	prob := rand.Intn(PROB_BASE) + 1
// 	if prob <= shakePolicy.OpenShark {
// 		// 开出鲨鱼
// 		sharkProb := rand.Intn(PROB_BASE) + 1

// 		if sharkProb <= shakePolicy.GoldShark {
// 			element, id = g.BirdAnimals.RandResult(GOLD_SHARK_ID, false)
// 		} else {
// 			element, id = g.BirdAnimals.RandResult(SILVER_SHARK_ID, false)
// 		}

// 		award1 := new(bridanimal.AwardSettleInfo)
// 		award1.AwardType = int32(id)
// 		award1.AwardBase = element.BaseID
// 		award1.Odds = int32(element.OddsNow)

// 		settleMsg.Award = append(settleMsg.Award, award1)
// 		g.settleElements = []*model.Element{element}
// 		hasShark=true
// 	}

// 	if shakePolicy.WinProb == 0 {
// NOT_CTRL:
// 		// 系统不控制输赢
// 		lastElement, id = g.BirdAnimals.RandResult(-1, true)
// 		// 如果第一次出现鲨鱼，移除通杀通赔
// 		if hasShark&&(lastElement.EType==model.ETypeAllKill||lastElement.EType==model.ETypeAllPay){
// 			goto NOT_CTRL
// 		}
// 		award2 := new(bridanimal.AwardSettleInfo)
// 		award2.AwardType = int32(id)
// 		award2.AwardBase = lastElement.BaseID
// 		award2.Odds = int32(lastElement.OddsNow)
// 		settleMsg.Award = append(settleMsg.Award, award2)
// 		g.settleMsg = settleMsg
// 		g.settleElements = append(g.settleElements, lastElement)
// 		return
// 	}

// 	oddsList := g.getRandOddsInfo() // 赔率列表
// 	areaGold := make(AreaGolds, 0)
// 	allBet := g.GetBetAll()
// 	winProb := rand.Intn(PROB_BASE) + 1
// 	if winProb <= shakePolicy.WinProb {
// LABEL1:
// 		// 系统赢
// 		allKillProb := rand.Intn(PROB_BASE) + 1
// 		if allKillProb <= shakePolicy.Win.AllKill {
// 			if hasShark {
// 				goto LABEL1  // 如果第一次得出鲨鱼，第二次不转出鲨鱼，通杀通赔
// 			}
// 			// 通杀
// 			lastElement, id = g.BirdAnimals.RandResult(ALL_KILL_ID, false)
// 		} else {
// 			// 开奖
// 			for index, v := range oddsList {
// 				if v.ID == 2 || v.ID == 3 || v.ID == 12 || v.ID == 13 || v.ID == 8 || v.ID == 9 {
// 					// 去除金鲨/银鲨/通杀/通赔/飞禽/走兽
// 					continue
// 				}
// 				// 赔钱<=总下注
// 				if int64(v.Odds)*g.TotalBet[index] <= allBet {
// 					areaGold = append(areaGold, AreaGold{
// 						RandomOdds: bridanimal.RandomOdds{
// 							ID:   v.ID,
// 							Odds: v.Odds,
// 						},
// 						PayGold: int64(v.Odds) * g.TotalBet[index],
// 					})
// 				}
// 			}

// 			sort.Sort(areaGold)
// 			if len(areaGold) == 0 {
// 				// 全部随机一个
// 				lastElement, id = g.BirdAnimals.RandResult(-1, true)
// 			} else {
// 				if rand.Intn(PROB_BASE)+1 <= shakePolicy.OpenCommon.Big {
// 					// 开大奖
// 					max := areaGold.GetMax()
// 					lastElement, id = g.BirdAnimals.RandResult(int(max[rand.Intn(len(max))].ID), false)
// 				} else {
// 					// 开小奖
// 					min := areaGold.GetMin()
// 					lastElement, id = g.BirdAnimals.RandResult(int(min[rand.Intn(len(min))].ID), false)
// 				}
// 			}
// 		}
// 	} else {
// LABEL2:
// 		// 系统输
// 		allPayProb := rand.Intn(PROB_BASE) + 1
// 		if allPayProb <= shakePolicy.Lost.AllPay {
// 			if hasShark {
// 				goto LABEL2  // 如果第一次得出鲨鱼，第二次不转出鲨鱼，通杀通赔
// 			}
// 			// 通赔
// 			lastElement, id = g.BirdAnimals.RandResult(ALL_PAY_ID, false)
// 		} else {
// 			for index, v := range oddsList {
// 				if v.ID == 2 || v.ID == 3 || v.ID == 12 || v.ID == 13 || v.ID == 8 || v.ID == 9 {
// 					// 去除金鲨/银鲨/通杀/通赔/飞禽/走兽
// 					continue
// 				}
// 				// 赔钱>=总下注
// 				if int64(v.Odds)*g.TotalBet[index] >= allBet {
// 					areaGold = append(areaGold, AreaGold{
// 						RandomOdds: bridanimal.RandomOdds{
// 							ID:   v.ID,
// 							Odds: v.Odds,
// 						},
// 						PayGold: int64(v.Odds) * g.TotalBet[index],
// 					})
// 				}
// 			}

// 			sort.Sort(areaGold)
// 			if len(areaGold) == 0 {
// 				// 全部随机一个
// 				lastElement, id = g.BirdAnimals.RandResult(-1, true)
// 			} else {
// 				if rand.Intn(PROB_BASE)+1 <= shakePolicy.OpenCommon.Big {
// 					// 开大奖
// 					max := areaGold.GetMax()
// 					lastElement, id = g.BirdAnimals.RandResult(int(max[rand.Intn(len(max))].ID), false)
// 				} else {
// 					// 开小奖
// 					min := areaGold.GetMin()
// 					lastElement, id = g.BirdAnimals.RandResult(int(min[rand.Intn(len(min))].ID), false)
// 				}
// 			}
// 		}
// 	}
// 	award2 := new(bridanimal.AwardSettleInfo)
// 	award2.AwardType = int32(id)
// 	award2.AwardBase = lastElement.BaseID
// 	award2.Odds = int32(lastElement.OddsNow)
// 	settleMsg.Award = append(settleMsg.Award, award2)
// 	g.settleMsg = settleMsg
// 	g.settleElements = append(g.settleElements, lastElement)
// }

// 循环下注消息
func (game *Game) loopBetMsg() {
	if game.Status != int32(bridanimal.GameStatus_BetStatus) {
		return
	}

	if game.TimerJob.GetTimeDifference() < 2000 {
		if game.LoopBetTimer != nil {
			game.LoopBetTimer.Cancel()
		}
		game.LoopBetTimer = nil
		return
	}

	msg := new(bridanimal.BroadBetEnd)
	msg.BetGold = make([]int64, BET_AREA_LENGTH)

	var notSend bool
	for index, v := range game.TotalBet {
		msg.BetGold[index] = v + game.AITotalBet[index]
		// msg.BetGold[index] = 1100
		// if game.TotalBetTemp[index] == game.TotalBet[index]+game.AITotalBet[index] {
		// 	notSend = true
		// 	continue
		// }
		// notSend = false
		// game.TotalBetTemp[index] = game.TotalBet[index] + game.AITotalBet[index]
	}
	if notSend {
		return
	}
	game.Table.Broadcast(int32(bridanimal.SendToClientMessageType_BroadEndBet), msg)
}

// 结束下注后的消息（10）
func (game *Game) betEndMsg() {
	msg := new(bridanimal.BroadBetEnd)
	msg.BetGold = make([]int64, BET_AREA_LENGTH)
	for index, v := range game.TotalBet {
		msg.BetGold[index] = v + game.AITotalBet[index]
	}
	game.Table.Broadcast(int32(bridanimal.SendToClientMessageType_BroadEndBet), msg)
}

// 结算日志
func (game *Game) writeLog() {

	getName := func(id int) string {
		var name string
		switch id {
		case 0:
			name = "燕子"
		case 1:
			name = "鸽子"
		case 2:
			name = "金鲨"
		case 3:
			name = "银鲨"
		case 4:
			name = "兔子"
		case 5:
			name = "猴子"
		case 6:
			name = "孔雀"
		case 7:
			name = "老鹰"
		case 8:
			name = "飞禽"
		case 9:
			name = "走兽"
		case 10:
			name = "熊猫"
		case 11:
			name = "狮子"
		case 12:
			name = "通杀"
		case 13:
			name = "通赔"
		}
		return name
	}

	var logStr string
	roomProb := game.Table.GetRoomProb()
	if roomProb == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProb)
		roomProb = 1000
	}
	logStr += fmt.Sprintf("当前作弊率：%v\n 开奖结果：", roomProb)

	for _, v := range game.settleElements {
		name := getName(v.ID)
		logStr += fmt.Sprintf("%v\n  ", name)
	}

	if len(game.settleElements) > 1 {
		logStr += fmt.Sprintf("是否中了金鲨/银鲨：%s；", "是")
	} else {
		logStr += fmt.Sprintf("是否中了金鲨/银鲨：%s；\n", "否")
	}

	{
		resultID := game.settleElements[len(game.settleElements)-1].ID
		//name := getName(resultID)

		var all bool
		switch resultID {
		case 12, 13:
			all = true
		default:
			all = false
		}
		if all {
			logStr += fmt.Sprintf("是否通杀/通赔：%s；\n", "是")
		} else {
			logStr += fmt.Sprintf("是否通杀/通赔：%s；\n", "否")
		}
	}

	var WinMostUserID int64
	var WinMostGold int64
	game.userLock.Lock()
	defer game.userLock.Unlock()
	for _, v := range game.UserInfoList {
		if v.User.IsRobot() {
			continue
		}

		var userLog string
		var allBetGold int64
		userLog += fmt.Sprintf("用户ID：%d，开始金币：%v，投注额:[", v.User.GetID(), score.GetScoreStr(v.currGold))
		for index, betGold := range v.BetInfo {
			allBetGold += betGold
			name := getName(index)
			ele, _ := config.BirdAnimaConfig.BirdAnimals.RandResult(index, false)
			odds := ele.OddsNow
			if ele.ID == BIRD_BET_INDEX || ele.ID == ANIMAL_BET_INDEX {
				odds = 2
			}
			userLog += fmt.Sprintf("%s%d倍：%v，\n",
				name, odds, score.GetScoreStr(betGold))
		}
		userLog += "],输赢：["
		for index, betGold := range v.BetInfo {
			name := getName(index)
			ele, _ := config.BirdAnimaConfig.BirdAnimals.RandResult(index, false)
			odds := ele.OddsNow
			if ele.ID == BIRD_BET_INDEX || ele.ID == ANIMAL_BET_INDEX {
				odds = 2
			}

			bFind := false
			for _, v := range game.settleElements {
				if (v.EType == model.ETypeAnimal && ele.ID == ANIMAL_BET_INDEX) ||
					(v.EType == model.ETypeBird && ele.ID == BIRD_BET_INDEX) {
					log.Debugf("找到 %v", v.EType)
					bFind = true
					break
				}

				if v.BaseID == ele.BaseID {
					bFind = true
					break
				}
			}

			var tmp string

			win := betGold * int64(odds)
			if !bFind {
				tmp = "-"
				win = betGold
			} else {
				taxMoney := win * game.Table.GetRoomRate() / 10000
				y := win * game.Table.GetRoomRate() % 10000
				if y > 0 {
					taxMoney += 1
				}

				win -= taxMoney
			}
			userLog += fmt.Sprintf("%s%d倍：%v%v，\n",
				name, odds, tmp, score.GetScoreStr(win))
		}

		userLog += "]，总输赢："
		userLog += fmt.Sprintf("%v，\n", score.GetScoreStr(v.LastWinGold-v.BetGoldNow))

		resultID := game.settleElements[len(game.settleElements)-1].ID
		name := getName(resultID)

		name = ""
		if len(game.settleElements) == 1 {
			name += getName(resultID)
		} else if len(game.settleElements) == 2 {
			if game.settleElements[0].ID == GOLD_SHARK_ID {
				name = "金鲨+" + name
			} else {
				name = "银鲨+" + name
			}
		}

		if game.settleElements[len(game.settleElements)-1].EType == model.ETypeBird && v.BetInfo[BIRD_BET_INDEX] != 0 {
			name += "+飞禽"
		} else if game.settleElements[len(game.settleElements)-1].EType == model.ETypeAnimal && v.BetInfo[ANIMAL_BET_INDEX] != 0 {
			name += "+走兽"
		}

		userLog += fmt.Sprintf("用户剩余金额：%v\n", score.GetScoreStr(v.User.GetScore()))
		game.Table.WriteLogs(v.User.GetID(), userLog)

		if v.LastWinGold > WinMostGold {
			WinMostUserID = v.User.GetID()
			WinMostGold = v.LastWinGold
		}
	}

	logStr += fmt.Sprintf("最高获利用户ID：%v 获得:%v;\n", WinMostUserID, WinMostGold)
	game.Table.WriteLogs(0, logStr)

}

// 开奖控制  重写于@2020.01.16 by youngbloood
func (g *Game) PreSettleRW() {
	defer g.wait.Done()
	g.SendUserNumMsg()
	// 发送结算消息（减去之前预结算的1s）
	settleMsg := new(bridanimal.SettleMsg)
	// 预先结算出结果
	nowRoomProb := g.Table.GetRoomProb()

	if nowRoomProb == 0 {
		log.Debugf("获取到系统作弊率：%d", nowRoomProb)
		nowRoomProb = 1000
	}

	shakePolicy := config.BirdAnimaConfig.PolicyTree.Find(nowRoomProb)
	var element, lastElement *model.Element

	var id int
	var isAll bool
	all := randBase()

	// 1.通杀通赔判定
	if all <= shakePolicy.All.AllPay {
		isAll = true
		// 出通赔
		log.Debugf("开出通赔")
		element = config.BirdAnimaConfig.BirdAnimals.GetByID(ALL_PAY_ID)
	} else if all > shakePolicy.All.AllPay && all <= (shakePolicy.All.AllPay+shakePolicy.All.AllKill) {
		isAll = true
		// 出通杀
		log.Debugf("开出通杀")
		element = config.BirdAnimaConfig.BirdAnimals.GetByID(ALL_KILL_ID)
	}

	if eleTemp := g.dealwithTest(); eleTemp != nil {
		element = eleTemp
		if eleTemp.ID == ALL_PAY_ID || eleTemp.ID == ALL_KILL_ID {
			isAll = true
		} else {
			isAll = false
		}
	}

	// 3000作弊率下必输控制
	if config.BirdAnimaConfig.IsOpen3000Ctrl && nowRoomProb == 3000 {
		// 不开通杀或通赔
		element = nil
	}

	if isAll && element != nil {
		//
		log.Debugf("开出通杀或者通赔")
		award1 := new(bridanimal.AwardSettleInfo)
		id = element.RandSubId()
		if g.testMsg != nil {
			id = int(g.testMsg.Result)
		}
		award1.AwardType = int32(id)
		award1.AwardBase = element.BaseID
		award1.Odds = int32(element.OddsNow)
		settleMsg.Award = append(settleMsg.Award, award1)
		g.settleMsg = settleMsg
		g.settleElements = []*model.Element{element}
		if id == GOLD_SHARK_ID || id == SILVER_SHARK_ID {
			log.Errorf("出现错误，出现金鲨和银鲨")
		}
		return
	}
	// 2.免费游戏判定
	free := randBase()

	// 3000作弊率下必输控制
	if config.BirdAnimaConfig.IsOpen3000Ctrl && nowRoomProb == 3000 {
		// 不能开出免费游戏
		free = PROB_BASE * 2
	}

	if free <= shakePolicy.Free.Open {
		// 开出免费游戏
		log.Debugf("开出免费游戏")
		if randBase() <= shakePolicy.Free.GoldShark {
			// 开出金鲨
			log.Debugf("开出金鲨")
			element = config.BirdAnimaConfig.BirdAnimals.GetByID(GOLD_SHARK_ID)
		} else {
			// 开出银鲨
			log.Debugf("开出银鲨")
			element = config.BirdAnimaConfig.BirdAnimals.GetByID(SILVER_SHARK_ID)
		}
	}

	if eleTemp := g.dealwithTest(); eleTemp != nil {
		element = eleTemp
		log.Debugf("开出来的奖励:%v", *element)
		if eleTemp.ID != GOLD_SHARK_ID &&
			eleTemp.ID != SILVER_SHARK_ID &&
			eleTemp.ID != ALL_PAY_ID &&
			eleTemp.ID != ALL_KILL_ID {
			// 普通元素
			award1 := new(bridanimal.AwardSettleInfo)
			id = element.RandSubId()
			if g.testMsg != nil {
				id = int(g.testMsg.Result)
			}
			award1.AwardType = int32(id)
			award1.AwardBase = element.BaseID
			award1.Odds = int32(element.OddsNow)
			settleMsg.Award = append(settleMsg.Award, award1)
			g.settleMsg = settleMsg
			g.settleElements = []*model.Element{element}
			return
		}

		// 通杀/通赔
		if eleTemp.ID == ALL_PAY_ID ||
			eleTemp.ID == ALL_KILL_ID {
			// 通杀/通赔
			award1 := new(bridanimal.AwardSettleInfo)
			id = element.RandSubId()
			if g.testMsg != nil {
				id = int(g.testMsg.Result)
			}
			award1.AwardType = int32(id)
			award1.AwardBase = element.BaseID
			award1.Odds = int32(element.OddsNow)
			settleMsg.Award = append(settleMsg.Award, award1)
			g.settleMsg = settleMsg
			g.settleElements = []*model.Element{element}
			return
		}

		// 开出鲨鱼，往下走
	}

	// 开出免费游戏
	if element != nil {
		log.Debugf("%v", *element)
		award1 := new(bridanimal.AwardSettleInfo)
		id = element.RandSubId()
		award1.AwardType = int32(id)
		award1.AwardBase = int32(element.BaseID)
		award1.Odds = int32(element.OddsNow)
		settleMsg.Award = append(settleMsg.Award, award1)
		g.settleMsg = settleMsg
		g.settleElements = []*model.Element{element}
	}
	log.Debugf("开出来第二个奖励")
	// 3.开奖结果计算----根据返奖率来结算[第二个]结果
	lastElement, _ = g.shakeResult()
	log.Debugf("%v", *lastElement)
	award2 := new(bridanimal.AwardSettleInfo)
	award2.AwardType = int32(lastElement.RandSubId())
	award2.AwardBase = int32(lastElement.BaseID)
	award2.Odds = int32(lastElement.OddsNow)
	log.Debugf("第二个奖励：%v", award2)
	log.Debugf("lastElement.BaseID = %v", lastElement.BaseID)
	settleMsg.Award = append(settleMsg.Award, award2)
	g.settleMsg = settleMsg
	g.settleElements = append(g.settleElements, lastElement)
	if (settleMsg.Award[0].AwardBase == GOLD_SHARK_ID || settleMsg.Award[0].AwardBase == SILVER_SHARK_ID) && len(settleMsg.Award) == 1 {
		log.Errorf("出现错误，出现金鲨和银鲨")
	}
}

func randBase() int {
	return rand.Intn(PROB_BASE) + 1
}

func (g *Game) shakeResult() (*model.Element, int) {
	nowRoomProb := g.Table.GetRoomProb()
	// RECALC: // 重新赋值nowRoomProb进行计算

	if nowRoomProb == 0 {
		log.Debugf("获取到系统作弊率：%d", nowRoomProb)
		nowRoomProb = 1000
	}

	shakePolicy := config.BirdAnimaConfig.PolicyTree.Find(nowRoomProb)
	var allBetGold int64
	for _, v := range g.TotalBet {
		allBetGold += v
	}
	oddsList := g.getRandOddsInfo()
	var backProb = [12]int64{} // 返奖率

	if allBetGold != 0 {
		birdPay := g.TotalBet[BIRD_BET_INDEX] * int64(oddsList[BIRD_BET_INDEX].Odds)
		animalPay := g.TotalBet[ANIMAL_BET_INDEX] * int64(oddsList[ANIMAL_BET_INDEX].Odds)
		for i, v := range g.TotalBet {
			// 所有的鸟类要加上飞禽的赔付值
			switch i {
			case 0, 1, 6, 7: // 飞禽
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)+birdPay) / float64(allBetGold) * float64(100))
			case 4, 5, 10, 11: // 走兽
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)+animalPay) / float64(allBetGold) * float64(100))
			default:
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)) / float64(allBetGold) * float64(100))
			}
		}
	}

	back := shakePolicy.Back.Rand()
	nowBackProbMin, nowBackProbMax := back.Min, back.Max
	var waitCheck []int
	for i, v := range backProb {
		if v >= int64(nowBackProbMin) &&
			v < int64(nowBackProbMax) &&
			i != BIRD_BET_INDEX &&
			i != ANIMAL_BET_INDEX &&
			i != GOLD_SHARK_ID &&
			i != SILVER_SHARK_ID &&
			i != ALL_KILL_ID &&
			i != ALL_PAY_ID {
			// 去除金鲨/银鲨/飞禽/走兽下注区
			waitCheck = append(waitCheck, i)
		}
	}

	// 3000作弊率下玩家必输控制
	if config.BirdAnimaConfig.IsOpen3000Ctrl && nowRoomProb == 3000 {
		// 重置waitCheck，选出小于100返奖率的进行开奖
		waitCheck = nil

		for i, v := range backProb {
			switch i {
			case 0, 1, 4, 5, 6, 7, 10, 11: // 排除金鲨/银鲨/通杀/通赔
				if v < 100 {
					waitCheck = append(waitCheck, i)
				}
			}
		}
		if len(waitCheck) == 0 {
			indexMin, backMin := 0, backProb[0]
			for i, v := range backProb {
				switch i {
				case 0, 1, 4, 5, 6, 7, 10, 11: // 排除金鲨/银鲨/通杀/通赔
					if v < backMin {
						indexMin = i
						backMin = v
					}
				}
			}
			waitCheck = append(waitCheck, indexMin)
		}
	}

	// 满足返奖率的待选元素
	if len(waitCheck) != 0 {
		// 这里随机选择一个进行返回
		log.Debugf("这里随机选择一个进行返回")
		index := waitCheck[rand.Intn(len(waitCheck))]
		// return config.BirdAnimaConfig.BirdAnimals..RandResult(index, false)
		ele := config.BirdAnimaConfig.BirdAnimals.GetByIDResult(index)
		return ele, ele.RandSubId()
	}

RMLABEL:
	// 此处往上查找都未找到，随机返回一个结果
	log.Debugf("此处往上查找都未找到，随机返回一个结果")
	ele, _ := config.BirdAnimaConfig.BirdAnimals.RandResult(-1, false)
	if ele.BaseID == int32(BIRD_BET_INDEX) ||
		ele.BaseID == int32(ANIMAL_BET_INDEX) ||
		ele.BaseID == GOLD_SHARK_ID ||
		ele.BaseID == SILVER_SHARK_ID ||
		ele.BaseID == ALL_KILL_ID ||
		ele.BaseID == ALL_PAY_ID {
		goto RMLABEL
	}
	return ele, ele.RandSubId()

	// 系统开奖返奖率
	// 作弊率	开奖返奖率
	// 3000		0-50
	// 2000		50-70
	// 1000		70-90  -----[中间]
	// -1000	90-110
	// -2000	110-130
	// -3000	130-150
	// if nowRoomProb == 1000 {
	// RMLABEL:
	// 	// 此处往上查找都未找到，随机返回一个结果
	// 	ele, id := config.BirdAnimaConfig.BirdAnimals.RandResult(-1, false)
	// 	if ele.ID == BIRD_BET_INDEX ||
	// 		ele.ID == ANIMAL_BET_INDEX ||
	// 		ele.ID == GOLD_SHARK_ID ||
	// 		ele.ID == SILVER_SHARK_ID ||
	// 		ele.ID == ALL_KILL_ID ||
	// 		ele.ID == ALL_PAY_ID {
	// 		goto RMLABEL
	// 	}
	// 	if len(g.Trend) != 0 && int32(id) == g.Trend[len(g.Trend)-1] {
	// 		ele, id = config.BirdAnimaConfig.BirdAnimals.RandResult(-1, false)
	// 	}
	// 	return ele, id
	// } else if nowRoomProb > 1000 {
	// 	nowRoomProb -= 1000
	// } else {
	// 	nowRoomProb += 1000
	// 	if nowRoomProb == 0 {
	// 		nowRoomProb = 1000
	// 	}
	// }
	// goto RECALC
}

func (game *Game) dealwithTest() *model.Element {
	if game.testMsg == nil {
		return nil
	}

	if game.testMsg.Result < 0 || game.testMsg.Result > 23 {
		return nil
	}

	result := game.testMsg.Result
	var id int
	for _, v := range config.BirdAnimaConfig.BirdAnimals {
		if v.ID == 8 || v.ID == 9 {
			continue
		}
		for _, subId := range v.SubIds {
			if subId == int(result) {
				id = v.ID
				break
			}
		}
	}
	return config.BirdAnimaConfig.BirdAnimals.GetByID(id)
}

func (game *Game) handleTestMsg(bts []byte) {
	if game.Status == int32(bridanimal.GameStatus_SettleStatus) || game.Status == int32(bridanimal.GameStatus_EndBetMovie) {
		return
	}
	msg := new(bridanimal.TestMsg)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	game.testMsg = msg
}

func (game *Game) repeatBet(bts []byte, user player.PlayerInterface) {
	msg := new(bridanimal.BetRept)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	if len(msg.BetArea) != BET_AREA_LENGTH {
		return
	}
	var allBet int64
	for _, v := range msg.BetArea {
		if v < 0 {
			return
		}
		allBet += v
	}

	userInfo := game.GetUserByUserID(user.GetID(), user)

	if user.GetScore() < allBet {
		userInfo.SendBetFail("重复下注金额不足")
	}

	for index, v := range msg.BetArea {
		userInfo.BetInfo[index] += v
		game.TotalBet[index] += v
	}
	user.SetScore(game.Table.GetGameNum(), -1*allBet, game.Table.GetRoomRate())
	// 同步局数
	userInfo.Totol += allBet
	userInfo.BetGoldNow += allBet
	user.SendMsg(int32(bridanimal.SendToClientMessageType_BetReptRet), msg)
}

func (game *Game) ReSettle() {
	for _, user := range game.UserInfoList {
		user.User.SetScore(game.Table.GetGameNum(), user.LastWinGold, 0)
		user.User.SendRecord(game.Table.GetGameNum(), user.LastWinGold-user.BetGoldNow, user.BetGoldNow, user.taxGold, user.LastWinGold, "")
		user.setChip()
		user.SyncDataWin(user.LastWinGold)
		user.SyncDataBet(user.Totol)
	}
}

func (game *Game) noticeGoldNow() {
	msg := new(bridanimal.GoldNowNoticeMsg)
	for userID, user := range game.UserInfoList {
		msg.UserID = userID
		msg.GoldNow = user.User.GetScore()
		user.User.SendMsg(int32(bridanimal.SendToClientMessageType_GoldNowNotice), msg)
		msg.Reset()
	}
}
