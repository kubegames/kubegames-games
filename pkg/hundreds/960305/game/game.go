package game

import (
	"fmt"
	"go-game-sdk/example/game_poker/960305/config"
	"go-game-sdk/example/game_poker/960305/model"
	baijiale "go-game-sdk/example/game_poker/960305/msg"
	"go-game-sdk/lib/clock"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
)

const (
	ZHUANG    = 0
	XIAN      = 1
	HE        = 2
	ZHUANGDUI = 3
	XIANDUI   = 4
)

const (
	HEODDS = 8
)

type Trend struct {
	Win         int32
	IsZhuangDui bool
	IsXianDui   bool
}

type Game struct {
	Table table.TableInterface // table interface

	startTime time.Time
	endTime   time.Time

	AllUserList         map[int64]*model.User //所有的玩家列表
	Status              baijiale.GameStatus   // 房间状态1 表示
	Win                 int32                 // 0表示庄胜利，1表示闲胜利，2表示和
	LastWinIsRedOrBlack int                   // 最近一次开龙还是虎
	ZhuangCards         [3]byte               // 庄牌
	XianCards           [3]byte               // 闲牌
	IsLuckWin           bool                  // 幸运一击是否胜利
	BetTotal            [5]int64              //下注统计
	TotalUserBet        [5]int64              //下注统计
	SenceSeat           *model.SceneInfo      //下注的玩家列表
	TimerJob            *clock.Job            //job
	RobotTimerJob       *clock.Job            //机器人job
	LastMasterBetType   int32                 //最近一次神算子下注的类型
	WinTrend            []*baijiale.OneTrend  //赢的走势
	CountUserList       []*model.User         //统计后的玩家列表
	Rule                config.RoomRules      //房间规则信息
	gp                  model.GamePoker       //牌
	LastCardCount       int32                 //剩余牌的张数
	testZhuang          []byte                //测试用庄牌
	testXian            []byte                //测试用闲牌
	CheatValue          int                   //作弊值
	PokerMsg            *baijiale.PokerMsg    //牌消息

	userMiMsg      *baijiale.MiPaiSynPosMsg // 用户咪牌动作(30次/s)
	userMiMsgTimer *clock.Job               //job

	RegalUser     *model.User // 大富豪玩家
	BigwinnerUser *model.User // 大赢家玩家
	MasterUser    *model.User // 神算子玩家
	sysCheat      string      //

	// ZhuangOpen []bool // 庄家开牌
	// XianOpen   []bool // 闲家开牌
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.SenceSeat = new(model.SceneInfo)
	game.SenceSeat.Init()
	game.PokerMsg = new(baijiale.PokerMsg)
	game.PokerMsg.ZhuangPoker = make([]byte, 3)
	game.PokerMsg.XianPoker = make([]byte, 3)
	game.PokerMsg.IsOpen = &baijiale.IsOpen{
		ZhuangOpen: make([]bool, 3),
		XianOpen:   make([]bool, 3),
	}
}

func (game *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

func (game *Game) CloseTable() {

}

//用户坐下
func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	return 1 //business.OnActionUserSitDownHandler()
}

func (game *Game) UserExit(user player.PlayerInterface) bool {
	u := game.getUser(user)
	//有下注时不让玩家离开
	if u.TotalBet != 0 {
		return false
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user, true)
	}
	delete(game.AllUserList, user.GetID())
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if u.TotalBet != 0 {
			msg := new(baijiale.ExitFail)
			msg.FailReason = "游戏中不能退出！"
			user.SendMsg(int32(baijiale.SendToClientMessageType_ExitRet), msg)
			return false
		}
		if u.SceneChairId != 0 {
			game.OnUserStanUp(user, true)
		}
		delete(game.AllUserList, user.GetID())
	}

	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(baijiale.ReceiveMessageType_BetID):
		game.Bet(buffer, user)
	case int32(baijiale.ReceiveMessageType_SitDown):
		game.UserSitDown(buffer, user)
	case int32(baijiale.ReceiveMessageType_GetTrend):
		game.SendTrend(user)
	case int32(baijiale.ReceiveMessageType_GetUserListInfo):
		game.SendUserListInfo(user)
	case int32(baijiale.ReceiveMessageType_StandUp):
		game.OnUserStanUp(user, false)
	case int32(baijiale.ReceiveMessageType_tempCard):
		// TODO: 注释掉如下块
		// game.testCard(buffer)
	case int32(baijiale.ReceiveMessageType_RecMiPaiSynPos):
		game.SendMiPaiSynPos(buffer)
	case int32(baijiale.ReceiveMessageType_RecMiEndMsg):
		game.MiEnd(buffer)
	}
}
func (game *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	robot.Init(ai, game)
	return robot
}

func (game *Game) SendScene(user player.PlayerInterface) bool {
	game.GetRoomconfig()
	u := game.getUser(user)

	game.SendRuleInfo(user)
	game.SendSceneMsg(user)
	game.SendMiPaiSeat()
	msg := new(baijiale.SceneGoldMsg)
	msg.Gold = user.GetScore()
	user.SendMsg(int32(baijiale.SendToClientMessageType_SceneGold), msg)

	game.SendUserBet(u)

	// if game.Status >= baijiale.GameStatus_EndBetMovie {
	user.SendMsg(int32(baijiale.SendToClientMessageType_PokerInfo), game.PokerMsg)
	// log.Traceln("进入场景发送PokerMsg =========== ", game.PokerMsg)
	if game.Status == baijiale.GameStatus_SettleStatus {
		if u.SettleMsg != nil {
			user.SendMsg(int32(baijiale.SendToClientMessageType_UserComeBack), u.SettleMsg)
		}
	}
	// }
	if game.TimerJob != nil {
		game.SendToUserStatusMsg(int(game.TimerJob.GetTimeDifference()), user)
	}

	game.SendTrend(user)

	return true
}

func (game *Game) Start() {
	if game.Table.GetRoomID() == -1 {
		game.Status = 0
		return
	}
	game.RefreshGold()
	//起立不能坐下的人
	for _, u := range game.SenceSeat.UserSeat {
		level := int(game.Table.GetLevel())
		if u.User.User.GetScore() < int64(config.LongHuConfig.SitDownLimit[game.Table.GetLevel()-1]) {
			game.OnUserStanUp(u.User.User, true)
			msg := new(baijiale.UserSitDownFail)
			str := fmt.Sprintf("低于%d不能入座！", config.LongHuConfig.SitDownLimit[level-1]/100)
			msg.FailReaSon = str
			u.User.User.SendMsg(int32(baijiale.SendToClientMessageType_SitDownFail), msg)
		}
	}

	game.LastMasterBetType = -1
	game.Table.StartGame()
	game.startTime = time.Now()

	game.BroadCastPokerCount()
	if game.RobotTimerJob == nil {
		r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
		game.RobotTimerJob, _ = game.Table.AddTimer(int64(r), game.RobotSitDown)
		r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
		game.Table.AddTimer(int64(r1), game.RobotStandUp)
	}
	game.Status = baijiale.GameStatus_StartMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startmove), game.StartBet)

	//开始动画消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startmove)
}

func (game *Game) StartBet() {
	game.ResetData()
	game.Status = baijiale.GameStatus_BetStatus
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startbet), game.getResult)

	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startbet)
}

// func (game *Game) EndBet() {
// 	game.Status = baijiale.GameStatus_EndBetMovie
// 	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endmove), game.getResult)
// 	//发送开始下注消息
// 	game.SendStatusMsg(config.LongHuConfig.Taketimes.Endmove)
// }

func (game *Game) showPoker2() {
	game.Status = baijiale.GameStatus_ShowPoker2
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker2), game.showPoker4)
	game.PokerMsg.IsOpen.XianOpen[0], game.PokerMsg.IsOpen.XianOpen[1] = true, true
	XianMiUser, ok := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]
	if ok && XianMiUser.User.User.IsRobot() {
		mi := game.genRobotMi(1, 2)
		game.sendMi(mi)
	}
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPoker2)
}
func (game *Game) showPoker4() {
	game.Status = baijiale.GameStatus_ShowPoker4
	game.PokerMsg.IsOpen.ZhuangOpen[0], game.PokerMsg.IsOpen.ZhuangOpen[1] = true, true
	if game.ZhuangCards[2] == 0 && game.XianCards[2] == 0 {
		// 只有4张牌
		game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker4), game.showPokerEnd)
	} else {
		game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker4), game.showPoker5)
	}
	ZhuangMiUser, ok := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	if ok && ZhuangMiUser.User.User.IsRobot() {
		mi := game.genRobotMi(3, 4)
		game.sendMi(mi)
	}
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPoker4)
}
func (game *Game) showPoker5() {
	game.Status = baijiale.GameStatus_ShowPoker5

	if game.ZhuangCards[2] != 0 {
		game.PokerMsg.IsOpen.ZhuangOpen[2] = true
	}
	if game.XianCards[2] != 0 {
		game.PokerMsg.IsOpen.XianOpen[2] = true
	}

	if (game.ZhuangCards[2] != 0 && game.XianCards[2] == 0) || (game.ZhuangCards[2] == 0 && game.XianCards[2] != 0) {
		// 只有5张牌
		game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker5), game.showPokerEnd)
	} else {
		game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker5), game.showPoker6)
	}

	XianMiUser, ok := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]
	if ok && XianMiUser.User.User.IsRobot() && game.XianCards[2] != 0 {
		mi := game.genRobotMi(5)
		game.sendMi(mi)
	}
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPoker5)
}
func (game *Game) showPoker6() {
	game.Status = baijiale.GameStatus_ShowPoker6
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker6), game.showPokerEnd)
	game.PokerMsg.IsOpen.ZhuangOpen[2], game.PokerMsg.IsOpen.XianOpen[2] = true, true

	ZhuangMiUser, ok := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	if ok && ZhuangMiUser.User.User.IsRobot() && game.ZhuangCards[2] != 0 {
		mi := game.genRobotMi(6)
		game.sendMi(mi)
	}
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPoker6)
}

func (game *Game) showPokerEnd() {
	if game.userMiMsgTimer != nil {
		game.userMiMsgTimer.Cancel()
		game.userMiMsgTimer = nil
	}
	game.Status = baijiale.GameStatus_ShowPokerEnd
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPokerEnd), game.Settle)
	game.PokerMsg.IsOpen.ZhuangOpen[2], game.PokerMsg.IsOpen.XianOpen[2] = true, true
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPokerEnd)
}

//结算
func (game *Game) Settle() {
	// TODO: 注释掉如下if块
	if len(game.testZhuang) == 0 {
		// game.getResult()
	} else {
		game.ZhuangCards[2] = 0
		for i := 0; i < len(game.testZhuang); i++ {
			game.ZhuangCards[i] = game.testZhuang[i]
		}

		game.testZhuang = make([]byte, 0)
		game.XianCards[2] = 0
		for i := 0; i < len(game.testXian); i++ {
			game.XianCards[i] = game.testXian[i]
		}
		game.testXian = make([]byte, 0)
		game.ComparePoker()
	}

	game.Status = baijiale.GameStatus_SettleStatus
	game.endTime = time.Now()
	game.sendSettleMsg()

	endtime := config.LongHuConfig.Taketimes.Endpay

	if game.gp.GetCardsCount()-int(game.LastCardCount) >= 6 {
		game.TimerJob, _ = game.Table.AddTimer(int64(endtime), game.Start)
	} else {
		game.TimerJob, _ = game.Table.AddTimer(int64(endtime), game.InitPoker)
	}

	game.checkUserBet()
	game.SendSceneMsg(nil)
	//发送开始下注消息
	game.SendStatusMsg(endtime)

	game.SenceSeat.Reset()
	game.Table.EndGame()
}

func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(baijiale.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	msg.TotalStatusTime = int32(game.TimerJob.GetIntervalTime() / time.Millisecond)
	// log.Traceln("状态*******************", msg)
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(baijiale.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	msg.TotalStatusTime = int32(game.TimerJob.GetIntervalTime() / time.Millisecond)
	user.SendMsg(int32(baijiale.SendToClientMessageType_Status), msg)
}

func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != baijiale.GameStatus_BetStatus {
		return
	}

	u := game.getUser(user)
	//用户下注
	BetPb := &baijiale.Bet{}
	proto.Unmarshal(buffer, BetPb)
	if BetPb.BetType < 0 || BetPb.BetType >= 5 || BetPb.BetIndex < 0 || int(BetPb.BetIndex) >= len(game.Rule.BetList) {
		model.SendBetFailMessage("数据异常", u)
		return
	}

	if u.Bet(BetPb, int64(game.Rule.BetList[BetPb.BetIndex]), game.BetTotal, int64(game.Rule.BetMinLimit)) {
		game.BetTotal[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])

		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		}

		u.User.SetScore(game.Table.GetGameNum(), -int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))]), game.Table.GetRoomRate())

		// 设置最大下注区域

		seatUser, ok := game.SenceSeat.UserSeat[u.User.GetID()]
		if ok {
			switch {
			case seatUser.User.BetArea[0] > seatUser.User.BetArea[1]:
				seatUser.BetMaxArea = 0
			case seatUser.User.BetArea[0] > seatUser.User.BetArea[1]:
				seatUser.BetMaxArea = 1
			default:
				switch BetPb.BetType {
				case 0:
					seatUser.BetMaxArea = 0
				case 1:
					seatUser.BetMaxArea = 1
				}
			}
		}
		if game.SenceSeat.UserBet(u) {
			game.SendMiPaiSeat()
		}

		zhuangMiSeatID, xianMiSeatID := -1, -1
		if game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID] != nil {
			zhuangMiSeatID = game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID].SeatNo
		}
		if game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID] != nil {
			xianMiSeatID = game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID].SeatNo
		}
		u.SendBetSuccessMessage(BetPb, zhuangMiSeatID, xianMiSeatID)
	}

	if game.MasterUser == u {
		game.LastMasterBetType = BetPb.BetType
	}
}

func (game *Game) UserSitDown(buffer []byte, user player.PlayerInterface) {

	// 这几种状态不接受用户坐下
	if game.Status == baijiale.GameStatus_EndBetMovie ||
		game.Status == baijiale.GameStatus_ShowPoker2 ||
		game.Status == baijiale.GameStatus_ShowPoker4 ||
		game.Status == baijiale.GameStatus_ShowPoker5 ||
		game.Status == baijiale.GameStatus_ShowPoker6 {
		return
	}

	us := &baijiale.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo), int(game.Table.GetLevel())) {
			u.SceneChairId = int(us.ChairNo)

			game.SendSceneMsg(nil)
			game.SendMiPaiSeat()
		}
	}
}

func (game *Game) InitPoker() {
	game.BroadCastPokerCount()
	game.Status = baijiale.GameStatus_FlushPoker
	game.gp.InitPoker()
	game.gp.ShuffleCards()
	game.LastCardCount = rand.Int31n(208) + 1

	game.WinTrend = make([]*baijiale.OneTrend, 0)
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Fulshpoker), game.Start)
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Fulshpoker)
}

// 发牌
func (game *Game) DealPoker() {
	game.ZhuangCards[2] = 0
	game.XianCards[2] = 0
	var zhuangdian byte
	var xiandian byte
	for i := 0; i < 2; i++ {
		game.ZhuangCards[i] = game.gp.DealCards()
		game.XianCards[i] = game.gp.DealCards()
		v1 := model.GetCardValue(game.ZhuangCards[i])
		v2 := model.GetCardValue(game.XianCards[i])
		if v1 > 10 {
			v1 = 10
		}

		if v2 > 10 {
			v2 = 10
		}

		zhuangdian += v1
		xiandian += v2
	}

	//game.XianCards[0] = 0x21
	//game.XianCards[1] = 0xd1
	//game.ZhuangCards[0] = 0x81
	//game.ZhuangCards[1] = 0x82
	//zhuangdian = 16
	//xiandian = 12

	/*
	   |	前两张牌合计点数	 |		闲家	 	 |	庄家
	   |   			0		    |		补牌 		|	补牌
	   |			1			|		补牌		|	补牌
	   |			2			|		补牌		|	补牌
	   |			3			|		补牌		|	闲家第三张牌：8，庄家不补牌
	   |			4			|		补牌		|	闲家第三张牌：1,8,9,0庄家不补牌
	   |			5			|		补牌		|	闲家第三张牌：1,2,3,8,9,0庄家不补牌
	   |			6			|		无需补牌	|	闲家第三张牌：6,7庄家要补牌
	   |			7			|		无需补牌	|	无需补牌
	   |			8			|	天牌，无需补牌   |	天牌，无需补牌
	   |			9			|	天牌，无需补牌	 |	天牌，无需补牌
	*/

	zhuangdian = zhuangdian % 10
	xiandian = xiandian % 10

	if zhuangdian >= 8 || xiandian >= 8 {
		return
	}

	if xiandian >= 6 && zhuangdian >= 6 {
		return
	}

	if zhuangdian == 7 && xiandian < 6 {
		game.XianCards[2] = game.gp.DealCards()
		return
	}

	if xiandian < 6 {
		game.XianCards[2] = game.gp.DealCards()
	}

	if zhuangdian < 3 {
		game.ZhuangCards[2] = game.gp.DealCards()
		return
	}
	//game.XianCards[2] = 0x31
	thirdCardValue := model.GetCardValue(game.XianCards[2])
	//庄家不补牌
	if thirdCardValue == 8 && zhuangdian == 3 {
		return
	}

	//庄家不补牌
	if (thirdCardValue >= 8 || thirdCardValue == 1) && zhuangdian == 4 {
		return
	}

	//庄家不补牌
	if (thirdCardValue >= 8 || (thirdCardValue >= 1 && thirdCardValue <= 3)) && zhuangdian == 5 {
		return
	}

	if (thirdCardValue != 6 && thirdCardValue != 7) && zhuangdian == 6 {
		return
	}

	game.ZhuangCards[2] = game.gp.DealCards()
}

// 比牌
func (game *Game) ComparePoker() {
	var zhuangdian byte
	var xiandian byte
	for i := 0; i < 3; i++ {
		v1 := model.GetCardValue(game.ZhuangCards[i])
		v2 := model.GetCardValue(game.XianCards[i])
		if v1 > 10 {
			v1 = 10
		}

		if v2 > 10 {
			v2 = 10
		}

		zhuangdian += v1
		xiandian += v2
	}

	zhuangdian = zhuangdian % 10
	xiandian = xiandian % 10

	if zhuangdian == xiandian {
		game.Win = HE
	} else if zhuangdian > xiandian {
		game.Win = ZHUANG
	} else {
		game.Win = XIAN
	}
}

//检查用户是否被踢掉
func (game *Game) checkUserBet() {
	for k, u := range game.AllUserList {
		if u.NoBetCount >= (config.LongHuConfig.Unplacebetnum+1) ||
			(u.User.IsRobot() &&
				(u.User.GetScore() > game.Rule.RobotMaxGold || u.User.GetScore() < game.Rule.RobotMinGold)) {
			//踢掉用户
			u.NoBetCount = 0
			if u.SceneChairId != 0 {
				game.OnUserStanUp(u.User, true)
			}

			delete(game.AllUserList, k)
			game.Table.KickOut(u.User)
		}
	}
}

//发送结算消息
func (game *Game) sendSettleMsg() {
	IsZhuangDui := model.GetCardValue(game.ZhuangCards[0]) == model.GetCardValue(game.ZhuangCards[1])
	IsXianDui := model.GetCardValue(game.XianCards[0]) == model.GetCardValue(game.XianCards[1])

	wt := &baijiale.OneTrend{}
	wt.Win = game.Win
	winlen := len(game.WinTrend)
	wt.IsZhuangDui = IsZhuangDui
	wt.IsXianDui = IsXianDui

	game.WinTrend = append(game.WinTrend, wt)
	if winlen > 100 {
		game.WinTrend = append(game.WinTrend[:(winlen-100-1)], game.WinTrend[(winlen-100):]...)
	}

	game.CountUserList = make([]*model.User, 0)
	SceneSettleMsg := new(baijiale.SceneUserSettle)
	// ZhuangMiUser, ok1 := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	// XianMiUser, ok2 := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]

	ZhuangBetCount := 0
	XianBetCount := 0
	HeBetCount := 0
	ZhuangDuiBetCount := 0
	XianDuiBetCount := 0
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [5]int64

	roomProb := game.Table.GetRoomProb()

	if roomProb == 0 {
		log.Debugf("获取到系统作弊率：%d", roomProb)
		roomProb = 1000
	}
	_ = roomProb
	//	titleTmpl := fmt.Sprintf(`
	//基础信息
	//场次名称：眯牌百家乐-房间等级：%v
	//开始时间：%v
	//结束时间：%v
	//		`,
	//		game.Table.GetLevel(),
	//		game.startTime.Format("2006-01-02 15:04:05"),
	//		game.endTime.Format("2006-01-02 15:04:05"),
	//		// roomProb,
	//	)

	//var zhuangBetInfo, xianBetInfo, heBetInfo, zhuangDuiBetInfo, xianDuiBetInfo string
	ZhuangStr := fmt.Sprintf("庄区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[ZHUANG]), score.GetScoreStr(game.BetTotal[ZHUANG]-game.TotalUserBet[ZHUANG]),
		score.GetScoreStr(game.TotalUserBet[ZHUANG]))

	XianStr := fmt.Sprintf("闲区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[XIAN]), score.GetScoreStr(game.BetTotal[XIAN]-game.TotalUserBet[XIAN]),
		score.GetScoreStr(game.TotalUserBet[XIAN]))

	HeStr := fmt.Sprintf("和区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[HE]), score.GetScoreStr(game.BetTotal[HE]-game.TotalUserBet[HE]),
		score.GetScoreStr(game.TotalUserBet[HE]))

	ZhuangDuiStr := fmt.Sprintf("庄对区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[ZHUANGDUI]),
		score.GetScoreStr(game.BetTotal[ZHUANGDUI]-game.TotalUserBet[ZHUANGDUI]), score.GetScoreStr(game.TotalUserBet[ZHUANGDUI]))

	XianDuiStr := fmt.Sprintf("闲对区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[XIANDUI]),
		score.GetScoreStr(game.BetTotal[XIANDUI]-game.TotalUserBet[XIANDUI]), score.GetScoreStr(game.TotalUserBet[XIANDUI]))

	for _, u := range game.AllUserList {
		u.NoBetCount++
		if !u.User.IsRobot() {
			if u.NoBetCount >= (config.LongHuConfig.Unplacebetnum + 1) {
				//发送踢掉用户
				msg := new(baijiale.KickOutUserMsg)
				msg.KickOutReason = "由于您5局未下注，已被踢出房间！"
				u.User.SendMsg(int32(baijiale.SendToClientMessageType_KickOutUser), msg)
			}
		}

		// 		if !u.User.IsRobot() {
		// 			titleTmpl += fmt.Sprintf(`
		// 用户ID：%v  投注金额 %v
		// 		`, u.User.GetID(), score.GetScoreStr(u.TotalBet))
		// 		}

		SceneUserInfo := new(baijiale.SceneUserInfo)

		msg := new(baijiale.SettleMsg)
		msg.IsXianDui = IsXianDui
		msg.IsZhuangDui = IsZhuangDui
		var win int64

		var before int64
		//startGold := score.GetScoreStr(u.User.GetScore() + u.TotalBet)
		var beforeTax int64

		var totalTax int64 //总税收
		if u.TotalBet > 0 {
			if game.Win == 0 {
				beforeTax += u.BetArea[ZHUANG]
				msg.UserZhuangWin += u.BetArea[ZHUANG]
				win += u.BetArea[ZHUANG] * 2
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], 0)
				Gold += capital
				msg.TotalWin += Gold
				before += msg.TotalWin
				totalTax += u.BetArea[ZHUANG]*2 - Gold
				SceneUserInfo.ZhuangWin = msg.UserZhuangWin
				SceneUserInfo.XianWin = -u.BetArea[XIAN]
				SystemWin[ZHUANG] += Gold
			} else if game.Win == 1 {
				msg.UserXianWin += u.BetArea[XIAN]
				beforeTax += u.BetArea[XIAN]
				win += u.BetArea[XIAN] * 2
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], 0)
				msg.TotalWin += Gold + capital
				before += msg.TotalWin
				totalTax += u.BetArea[XIAN]*2 - Gold
				SceneUserInfo.ZhuangWin -= u.BetArea[ZHUANG]
				SceneUserInfo.XianWin = msg.UserXianWin
				SystemWin[XIAN] += Gold
			} else if game.Win == HE {
				msg.HeWin = u.BetArea[HE] * int64(HEODDS)
				beforeTax += msg.HeWin + u.BetArea[ZHUANG] + u.BetArea[XIAN]
				win += u.BetArea[HE] * int64(HEODDS+1)

				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[HE]*int64(HEODDS), game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[HE], 0)
				Gold += capital
				msg.TotalWin += Gold + u.BetArea[ZHUANG] + u.BetArea[XIAN]

				totalTax += u.BetArea[HE]*int64(HEODDS+1) - Gold
				// 此处只计算和赢的钱
				before += Gold
				//把压龙和虎的钱退回
				u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], 0)
				u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], 0)
				SystemWin[HE] += Gold + u.BetArea[ZHUANG] + u.BetArea[XIAN]
			}

			if IsZhuangDui {
				win += u.BetArea[ZHUANGDUI] * 12
				beforeTax += u.BetArea[ZHUANGDUI] * 12
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANGDUI]*11, game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANGDUI], 0)
				Gold += capital
				msg.TotalWin += Gold
				before += msg.TotalWin
				msg.ZhuangDui = Gold

				totalTax += u.BetArea[ZHUANGDUI]*12 - Gold
				SceneUserInfo.ZhuangDui = msg.ZhuangDui
				SystemWin[ZHUANGDUI] += Gold
			}

			if IsXianDui {
				win += u.BetArea[XIANDUI] * 12
				beforeTax += u.BetArea[XIANDUI] * 12
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIANDUI]*11, game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIANDUI], 0)
				Gold += capital
				msg.TotalWin += Gold
				before += msg.TotalWin
				msg.XianDui = Gold

				totalTax += u.BetArea[XIANDUI]*12 - Gold
				SceneUserInfo.XianDui = msg.XianDui
				SystemWin[XIANDUI] += Gold
			}
		}
		// 发送战绩
		// u.User.SendRecord(game.Table.GetGameNum(), u.User.GetScore()-u.CurrGold, u.TotalBet, win-msg.TotalWin, Gold, "")

		// 发送打码量
		switch game.Win {
		case HE: // 开和不计算庄和闲差值，只计算和的下注和对的下注
			u.User.SendChip(u.BetArea[2] + u.BetArea[3] + u.BetArea[4])
		default: // 庄与闲下注金额的差值的绝对值+和的下注+庄对+闲对
			diff := int64(math.Abs(float64(u.BetArea[0]) - float64(u.BetArea[1])))
			u.User.SendChip(diff + u.BetArea[2] + u.BetArea[3] + u.BetArea[4])
		}

		for i := 0; i < 5; i++ {
			msg.BetArea = append(msg.BetArea, game.BetTotal[i])
			msg.UserBet = append(msg.UserBet, u.BetArea[i])
		}

		if u.BetArea[HE] != 0 {
			SceneUserInfo.HeWin = msg.HeWin
		} else {
			SceneUserInfo.HeWin -= u.BetArea[HE]
		}

		SceneUserInfo.BetArea = append(SceneUserInfo.BetArea, msg.BetArea...)
		SceneUserInfo.UserBet = append(SceneUserInfo.UserBet, msg.UserBet...)

		SceneUserInfo.TotalWin = msg.TotalWin
		SceneUserInfo.UserID = int64(u.User.GetID())
		SceneUserInfo.SceneSeatID = int32(u.SceneChairId)

		// 退的钱要算在赢得钱里面（如：下和，结果开和，则下的庄和闲得钱都会退回来，退回来得钱算赢得）
		//统计玩家信息
		if (msg.TotalWin) > u.TotalBet {
			// 赢
			u.UserCount(true)
			u.SyncWinGold(msg.TotalWin)
		} else {
			// 输
			u.UserCount(false)
			u.SyncWinGold(0)
		}

		//写入数据库统计信息
		if MaxWinGold < win-u.TotalBet {
			MaxWinGold = win - u.TotalBet
			MaxWinUserID = u.User.GetID()
		}

		if !u.User.IsRobot() {
			var temp string
			var temp1 string
			if u.TotalBet != 0 {
				temp += fmt.Sprintf("用户ID：%v，开始金币：%v，投注额:", u.User.GetID(), score.GetScoreStr(u.CruenSorce))
				temp1 += fmt.Sprintf(" 输赢：")
			}
			if u.BetArea[ZHUANG] != 0 {
				temp += fmt.Sprintf("庄：%v ", score.GetScoreStr(u.BetArea[ZHUANG]))
				temp1 += fmt.Sprintf("庄：%v ", score.GetScoreStr(msg.UserZhuangWin*2-u.BetArea[ZHUANG]))
				ZhuangBetCount++
			}
			if u.BetArea[XIAN] != 0 {
				temp += fmt.Sprintf("闲：%v ", score.GetScoreStr(u.BetArea[XIAN]))
				temp1 += fmt.Sprintf("闲：%v ", score.GetScoreStr(msg.UserXianWin*2-u.BetArea[XIAN]))
				XianBetCount++
			}

			if u.BetArea[HE] != 0 {
				temp += fmt.Sprintf("和：%v ", score.GetScoreStr(u.BetArea[HE]))
				temp1 += fmt.Sprintf("和：%v ", score.GetScoreStr(msg.HeWin-u.BetArea[HE]))
				HeBetCount++
			}

			if u.BetArea[ZHUANGDUI] != 0 {
				temp += fmt.Sprintf("庄对：%v ", score.GetScoreStr(u.BetArea[ZHUANGDUI]))
				temp1 += fmt.Sprintf("庄对：%v ", score.GetScoreStr(msg.ZhuangDui-u.BetArea[ZHUANGDUI]))
				ZhuangDuiBetCount++
			}

			if u.BetArea[XIANDUI] != 0 {
				temp += fmt.Sprintf("闲对：%v ", score.GetScoreStr(u.BetArea[XIANDUI]))
				temp1 += fmt.Sprintf("闲对：%v ", score.GetScoreStr(msg.XianDui-u.BetArea[XIANDUI]))
				XianDuiBetCount++
			}
			temp1 += fmt.Sprintf(" 总输赢：%v，用户剩余金额：%v \r\n", score.GetScoreStr(win-u.TotalBet), score.GetScoreStr(u.User.GetScore()))
			temp += temp1
			game.Table.WriteLogs(u.User.GetID(), temp)
		}

		game.PaoMaDeng(msg.TotalWin-u.TotalBet, u.User)
		u.LastWinGold = msg.TotalWin
		game.CountUser(u)

		msg.Win = int32(game.Win)
		msg.UserScore = u.User.GetScore()
		// if ok1 {
		// 	msg.ZhuangMiUserNikeName = ZhuangMiUser.User.User.GetNike()
		// 	msg.ZhuangMiUserHead = ZhuangMiUser.User.User.GetHead()
		// }

		// if ok2 {
		// 	msg.XiangMiUserNikeName = XianMiUser.User.User.GetNike()
		// 	msg.XiangMiUserHead = XianMiUser.User.User.GetHead()
		// }

		SceneUserInfo.UserScore = msg.UserScore

		u.User.SendMsg(int32(baijiale.SendToClientMessageType_Settle), msg)

		if u.TotalBet > 0 && !u.User.IsRobot() {
			u.SettleMsg = msg
			chip := u.BetArea[ZHUANG] - u.BetArea[XIAN]

			if chip < 0 {
				chip = -chip
			}

			for m := HE; m <= XIANDUI; m++ {
				chip += u.BetArea[m]
			}
			if game.Win == HE {
				chip = u.BetArea[HE] + u.BetArea[ZHUANGDUI] + u.BetArea[XIANDUI]
			}
			// u.User.SetChip(chip)
		} else {
			u.SettleMsg = nil
		}

		if u.SceneChairId != 0 {
			SceneSettleMsg.UserList = append(SceneSettleMsg.UserList, SceneUserInfo)
		}
		u.User.SendRecord(game.Table.GetGameNum(), msg.TotalWin-u.TotalBet, u.TotalBet,
			totalTax, msg.TotalWin, "")
		u.ResetUserData()
	}

	//日志信息
	ZhuangStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", ZhuangBetCount, score.GetScoreStr(game.BetTotal[ZHUANG]-SystemWin[ZHUANG]))
	XianStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", XianBetCount, score.GetScoreStr(game.BetTotal[XIAN]-SystemWin[XIAN]))
	HeStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", HeBetCount, score.GetScoreStr(game.BetTotal[HE]-SystemWin[HE]))
	ZhuangDuiStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", ZhuangDuiBetCount, score.GetScoreStr(game.BetTotal[ZHUANGDUI]-SystemWin[ZHUANGDUI]))
	XianDuiStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", XianDuiBetCount, score.GetScoreStr(game.BetTotal[XIANDUI]-SystemWin[XIANDUI]))
	t := ""
	if game.Win == 0 {
		t = "庄赢"
	} else if game.Win == 1 {
		t = "闲赢"
	} else if game.Win == 2 {
		t = "和赢"
	}
	str := fmt.Sprintf("%v作弊率：%v \r\n开局结果 :%v:庄区域牌：%v 点数：%v,", game.sysCheat, game.CheatValue,
		t, model.GetCardString(game.ZhuangCards), model.GetTypeString(game.ZhuangCards))

	str += fmt.Sprintf("闲区域牌：%v 点数：%v \r\n",
		model.GetCardString(game.XianCards), model.GetTypeString(game.XianCards))
	count := int64(0)
	for k, v := range game.BetTotal {
		count += v - SystemWin[k]
	}
	str += fmt.Sprintf("系统输赢额度：%v \r\n",
		score.GetScoreStr(int64(count)))

	str += fmt.Sprintf("最高获利用户ID：%v 获得：%v\r\n",
		MaxWinUserID, score.GetScoreStr(MaxWinGold))
	totalstr := ZhuangStr + XianStr + HeStr + ZhuangDuiStr + XianDuiStr + str
	go game.Table.WriteLogs(0, totalstr)

	game.SenceSeat.BetXianMaxID = 0
	game.SenceSeat.BetZhuangMaxID = 0
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

func (game *Game) getResult() {
	game.Status = baijiale.GameStatus_EndBetMovie
	test := game.Table.GetRoomProb()

	if test == 0 {
		game.sysCheat = "获取作弊率为0 "
		test = 1000
	} else {
		game.sysCheat = ""
	}
	log.Debugf("使用作弊率为：%v", test)
	game.CheatValue = int(test)
	eat := 0
	out := 0
	r := rand.Intn(10000)
	v := config.LongHuConfig.GetCheatValue(game.CheatValue)
	if v != 0 {
		if r < v {
			eat = 1
		} else {
			out = 1
		}
	}

	if len(game.testXian) > 0 {
		game.ZhuangCards[2] = 0
		game.XianCards[2] = 0
		for i := 0; i < len(game.testZhuang); i++ {
			game.ZhuangCards[i] = game.testZhuang[i]
		}
		for i := 0; i < len(game.testXian); i++ {
			game.XianCards[i] = game.testXian[i]
		}
		game.testXian = make([]byte, 0)
		game.testZhuang = make([]byte, 0)
		game.ComparePoker()
	} else {
		TotalMoney := game.TotalUserBet[ZHUANG] + game.TotalUserBet[XIAN] + game.TotalUserBet[HE]
		bAddCards := false
		for i := 0; i < 50; i++ {
			//如果不符合要求，发过的牌还回去
			if bAddCards {
				game.AddCards()
			}

			game.DealPoker()
			bAddCards = true

			// if len(game.testZhuang) != 0 {
			// 	for i := 0; i < len(game.testZhuang); i++ {
			// 		game.ZhuangCards[i] = game.testZhuang[i]
			// 	}
			// 	game.XianCards[2] = 0
			// 	for i := 0; i < len(game.testXian); i++ {
			// 		game.XianCards[i] = game.testXian[i]
			// 	}
			// }

			game.ComparePoker()
			//log.Debugf("结果为：%v", game.ZhuangCards, game.XianCards)
			//第一把不出和
			if len(game.WinTrend) == 0 && game.Win == HE {
				i--
				continue
			}

			if eat == 0 && out == 0 {
				//不控制
				break
			}

			IsZhuangDui := model.GetCardValue(game.ZhuangCards[0]) == model.GetCardValue(game.ZhuangCards[1])
			IsXianDui := model.GetCardValue(game.XianCards[0]) == model.GetCardValue(game.XianCards[1])

			var PayMoney int64
			if game.Win == XIAN {
				PayMoney = game.TotalUserBet[XIAN]
			} else if game.Win == ZHUANG {
				PayMoney = game.TotalUserBet[ZHUANG] * 95 / 100
			}

			OutMoney := TotalMoney - PayMoney*2
			if IsZhuangDui {
				OutMoney -= game.TotalUserBet[ZHUANGDUI] * 12
			}

			if IsXianDui {
				OutMoney -= game.TotalUserBet[XIANDUI] * 12
			}

			if out == 1 {
				OutMoney -= game.TotalUserBet[HE]
				if OutMoney <= 0 {
					//吐分
					break
				}
			}

			if game.Win == HE {
				OutMoney = TotalMoney - game.TotalUserBet[XIAN] - game.TotalUserBet[ZHUANG] - (HEODDS+1)*game.TotalUserBet[HE]
			}

			if eat == 1 && OutMoney >= 0 {
				//吃分
				break
			}
		}
	}
	for i := 0; i < 3; i++ {
		game.PokerMsg.ZhuangPoker[i] = game.ZhuangCards[i]
		game.PokerMsg.XianPoker[i] = game.XianCards[i]
	}

	game.PokerMsg.Win = int32(game.Win)
	endtime := config.LongHuConfig.Taketimes.Cardmove
	if game.ZhuangCards[2] == 0 {
		endtime -= 5000
	}

	if game.XianCards[2] == 0 {
		endtime -= 5000
	}

	// 重计算眯牌的玩家座位号
	game.SenceSeat.ReCalc()

	ZhuangMiUser, ok1 := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	XianMiUser, ok2 := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]
	if ok1 {
		game.PokerMsg.ZhuangMiUserNikeName = ZhuangMiUser.User.User.GetNike()
		game.PokerMsg.ZhuangMiUserHead = ZhuangMiUser.User.User.GetHead()
		game.PokerMsg.ZhuangMiUserSeatId = int32(ZhuangMiUser.SeatNo)
	} else {
		game.PokerMsg.ZhuangMiUserNikeName = ""
		game.PokerMsg.ZhuangMiUserHead = ""
		game.PokerMsg.ZhuangMiUserSeatId = 0
	}
	if ok2 {
		game.PokerMsg.XiangMiUserNikeName = XianMiUser.User.User.GetNike()
		game.PokerMsg.XiangMiUserHead = XianMiUser.User.User.GetHead()
		game.PokerMsg.XianMiUserSeatId = int32(XianMiUser.SeatNo)
	} else {
		game.PokerMsg.XiangMiUserNikeName = ""
		game.PokerMsg.XiangMiUserHead = ""
		game.PokerMsg.XianMiUserSeatId = 0
	}

	game.PokerMsg.ZhuangMi = game.SenceSeat.BetZhuangMaxID
	game.PokerMsg.XianMi = game.SenceSeat.BetXianMaxID
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endmove), game.showPoker2)
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_PokerInfo), game.PokerMsg)

	game.SendStatusMsg(config.LongHuConfig.Taketimes.Endmove)

	// 30帧/s
	game.userMiMsgTimer, _ = game.Table.AddTimerRepeat(int64(33), 0, game.loopMiOpt)
}

// 重复发送咪牌动作
func (game *Game) loopMiOpt() {
	if game.userMiMsg != nil {
		game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SendMiPaiSynPos), game.userMiMsg)
	}
}

func (game *Game) getUser(user player.PlayerInterface) *model.User {
	u, ok := game.AllUserList[user.GetID()]
	if !ok {
		u = new(model.User)
		game.AllUserList[user.GetID()] = u
		u.Table = game.Table
		u.User = user
		u.Rule = &game.Rule
		u.CurrGold = user.GetScore()
		u.WinGoldChan = make(chan int64, 20)
		u.InTime = time.Now().Unix()
		u.ResetUserData()
	}

	return u
}

//发送场景消息
func (game *Game) SendSceneMsg(u player.PlayerInterface) {
	// 重新计算眯牌玩家
	game.SenceSeat.ReCalc()

	msg := new(baijiale.SceneMessage)
	bigwinner := game.BigwinnerUser
	master := game.MasterUser
	regal := game.RegalUser
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(baijiale.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetID())
		if bigwinner == v.User {
			su.IsBigWinner = true
		}

		if master == v.User {
			su.IsMaster = true
		}
		if regal == v.User {
			su.IsMillionaire = true
		}

		su.ZhuangBetGold = v.User.BetArea[0]
		su.XianBetGold = v.User.BetArea[1]

		// if game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID] != nil && su.SeatId == int32(game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID].SeatNo) {
		// 	su.XianBetGold = -1
		// } else if game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID] != nil && su.SeatId == int32(game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID].SeatNo) {
		// 	su.ZhuangBetGold = -1
		// }

		msg.UserData = append(msg.UserData, su)
	}

	msg.PokerCount = int32(game.gp.GetCardsCount()) - game.LastCardCount
	if u != nil {
		u.SendMsg(int32(baijiale.SendToClientMessageType_SceneID), msg)
	} else {
		game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SceneID), msg)
	}
}

func (game *Game) SendUserBet(u *model.User) {
	msg := new(baijiale.SceneBetInfo)
	for i := 0; i < 5; i++ {
		msg.BetArea = append(msg.BetArea, game.BetTotal[i])
		msg.UserBet = append(msg.UserBet, u.BetArea[i])
	}

	msg.UserBetTotal = u.TotalBet
	msg.MasterBetType = game.LastMasterBetType
	u.User.SendMsg(int32(baijiale.SendToClientMessageType_BetInfo), msg)
}

func (game *Game) SendTrend(u player.PlayerInterface) {
	log.Tracef("发送走势图")
	msg := new(baijiale.Trend)
	msg.Info = append(msg.Info, game.WinTrend...)

	u.SendMsg(int32(baijiale.SendToClientMessageType_TrendInfo), msg)
}

func (game *Game) CountUser(u *model.User) {
	uc := len(game.CountUserList)
	if uc == 0 {
		game.CountUserList = append(game.CountUserList, u)
		return
	}

	adduser := u
	if game.CountUserList[0].RetWin < u.RetWin {
		adduser = game.CountUserList[0]
		game.CountUserList[0] = u
	}

	for i := 1; i < uc; i++ {
		if adduser.RetBet > game.CountUserList[i].RetBet {
			var tmp []*model.User
			tmp = append(tmp, game.CountUserList[i:]...)
			game.CountUserList = append(game.CountUserList[0:i], adduser)
			game.CountUserList = append(game.CountUserList, tmp...)
			return
		}
	}

	game.CountUserList = append(game.CountUserList, adduser)

	game.BigwinnerUser = game.SenceSeat.GetBigWinner(-1)
	for i := 0; i < 6; i++ {
		if regal := game.SenceSeat.GetRegal(i); regal != game.BigwinnerUser {
			game.RegalUser = regal
			break
		}
	}
	for i := 0; i < 6; i++ {
		if master := game.SenceSeat.GetMaster(i); master != game.BigwinnerUser && master != game.RegalUser {
			game.MasterUser = master
			break
		}
	}
}

func (game *Game) SendUserListInfo(user player.PlayerInterface) {
	msg := new(baijiale.UserList)

	topUser := make(model.TopUser, 0)
	topUserID := make([]int64, 6)
	// 玩家列表
	for i := 1; i <= 6; i++ {
		uu, ok := game.SenceSeat.SenceSeat[i]
		if ok {
			topUser = append(topUser, uu.User)
			topUserID[i-1] = uu.User.User.GetID()
		}
	}
	// 前6个玩家排序
	sort.Sort(topUser)
	for _, u := range topUser {
		userinfo := new(baijiale.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.WinGold
		userinfo.Head = u.User.GetHead()
		switch u {
		case game.BigwinnerUser:
			userinfo.Icon = 3
		case game.MasterUser:
			userinfo.Icon = 1
		case game.RegalUser:
			userinfo.Icon = 2
		}
		msg.UserList = append(msg.UserList, userinfo)
	}

	leftUser := make(model.LeftUser, 0)
	for _, u := range game.CountUserList {
		// 排除上座玩家
		if u.User.GetID() == topUserID[0] ||
			u.User.GetID() == topUserID[1] ||
			u.User.GetID() == topUserID[2] ||
			u.User.GetID() == topUserID[3] ||
			u.User.GetID() == topUserID[4] ||
			u.User.GetID() == topUserID[5] {
			continue
		}
		leftUser = append(leftUser, u)

	}
	// 剩余玩家排序
	sort.Sort(leftUser)
	for _, u := range leftUser {
		userinfo := new(baijiale.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.WinGold
		userinfo.Head = u.User.GetHead()
		msg.UserList = append(msg.UserList, userinfo)
	}
	user.SendMsg(int32(baijiale.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) ResetData() {
	for i := 0; i < 5; i++ {
		game.TotalUserBet[i] = 0
		game.BetTotal[i] = 0
	}
	game.testZhuang = make([]byte, 0)
	game.testXian = make([]byte, 0)
	game.PokerMsg.IsOpen.ZhuangOpen = make([]bool, 3)
	game.PokerMsg.IsOpen.XianOpen = make([]bool, 3)
	game.userMiMsg = nil
}

func (game *Game) OnUserStanUp(user player.PlayerInterface, isJump bool) {

	// 这几种状态不接受用户站起
	if !isJump {
		if game.Status == baijiale.GameStatus_EndBetMovie ||
			game.Status == baijiale.GameStatus_ShowPoker2 ||
			game.Status == baijiale.GameStatus_ShowPoker4 ||
			game.Status == baijiale.GameStatus_ShowPoker5 ||
			game.Status == baijiale.GameStatus_ShowPoker6 {
			return
		}
	}

	bSendMiPai := true
	// if user.GetID() == game.SenceSeat.BetZhuangMaxID || user.GetID() == game.SenceSeat.BetXianMaxID {
	// 	bSendMiPai = true
	// }

	game.SenceSeat.UserStandUp(user)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		u.SceneChairId = 0
	}

	game.SendSceneMsg(nil)

	if bSendMiPai {
		game.SendMiPaiSeat()
	}
}

func (game *Game) RobotSitDown() {
	r := 0
	if RConfig.SitDownTime[1] == RConfig.SitDownTime[0] {
		r = RConfig.SitDownTime[1]
	} else {
		r = rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
	}
	game.RobotTimerJob, _ = game.Table.AddTimer(int64(r), game.RobotSitDown)

	count := game.SenceSeat.GetSitDownUserCount()
	if count < len(RConfig.SitDownProbability) {
		r = rand.Intn(10000)
		if r < RConfig.SitDownProbability[count].Probability {
			sitdowncount := 0
			r = RConfig.SitDownProbability[count].Max
			if RConfig.SitDownProbability[count].Max != RConfig.SitDownProbability[count].Min {
				r = rand.Intn(RConfig.SitDownProbability[count].Max-RConfig.SitDownProbability[count].Min) + RConfig.SitDownProbability[count].Min
			}

			for _, v := range game.AllUserList {
				if v.User.IsRobot() && v.SceneChairId == 0 &&
					v.User.GetScore() > int64(config.LongHuConfig.SitDownLimit[game.Table.GetLevel()-1]) {
					us := &baijiale.UserSitDown{}
					us.ChairNo = int32(game.SenceSeat.GetSceneChairId())
					if us.ChairNo != 0 {
						pb, _ := proto.Marshal(us)
						game.UserSitDown(pb, v.User)
						sitdowncount++
					}
				}

				if sitdowncount >= r {
					break
				}
			}
		}
	}
}

func (game *Game) RobotStandUp() {
	r := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
	game.Table.AddTimer(int64(r), game.RobotStandUp)

	count := game.SenceSeat.GetSitDownUserCount()
	//log.Tracef("有多少个人坐下%v", count)
	if count < len(RConfig.StandUpProbability) {
		//log.Tracef("机器人站立")
		r = rand.Intn(10000)
		if r < RConfig.StandUpProbability[count].Probability {
			//log.Tracef("机器人站立1")
			if RConfig.StandUpProbability[count].Max == RConfig.StandUpProbability[count].Min {
				r = RConfig.StandUpProbability[count].Max
			} else {
				r = rand.Intn(RConfig.StandUpProbability[count].Max - RConfig.StandUpProbability[count].Min)
			}

			//log.Tracef("机器人站立1%v", r)
			for i := 0; i < r; i++ {
				user := game.SenceSeat.GetAiUser()
				if user != nil {
					game.OnUserStanUp(user, true)
				}
			}
		}
	}
}

func (game *Game) BrodCastSceneMsg() {
	msg := new(baijiale.SceneMessage)
	bigwinner := game.BigwinnerUser
	master := game.MasterUser
	regal := game.RegalUser
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(baijiale.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetID())
		if bigwinner == v.User {
			su.IsBigWinner = true
		}
		if master == v.User {
			su.IsMaster = true
		}
		if regal == v.User {
			su.IsMillionaire = true
		}

		msg.UserData = append(msg.UserData, su)
	}

	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SceneID), msg)
}

func (game *Game) GameStart(user player.PlayerInterface) bool {
	game.GetRoomconfig()
	if game.Status == 0 {
		game.InitPoker()
		game.Table.AddTimerRepeat(1000, 0, game.SendRoomInfo)
	}
	return true
}

func (game *Game) GetRoomconfig() {
	if game.Rule.UserBetLimit != 0 {
		return
	}
	str := game.Table.GetAdviceConfig()

	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		fmt.Printf("解析房间配置失败 err%v\n", err)
		fmt.Printf("%v\n", str)
		return
	}

	level := game.Table.GetLevel()
	game.Rule.BetList = make([]int, 0)
	BetBase, _ := js.Get("Bottom_Pouring").Int()
	_ = BetBase

	game.Rule.BetMinLimit, _ = js.Get("betMinLimit").Int()
	// game.Rule.UserBetLimit = int64(BetBase) * 5000
	game.Rule.UserBetLimit = config.LongHuConfig.Singleuserallspacelimit5times[level-1]
	// game.Rule.BetList = append(game.Rule.BetList, BetBase)
	// game.Rule.BetList = append(game.Rule.BetList, BetBase*10)
	// game.Rule.BetList = append(game.Rule.BetList, BetBase*50)
	// game.Rule.BetList = append(game.Rule.BetList, BetBase*100)
	// game.Rule.BetList = append(game.Rule.BetList, BetBase*500)

	for _, v := range config.LongHuConfig.Chips5times[level-1] {
		game.Rule.BetList = append(game.Rule.BetList, int(v))

	}
	// game.Rule.BetLimit = config.LongHuConfig.BetLimit
	// game.Rule.BetLimit[ZHUANG] = int64(BetBase) * 5000
	// game.Rule.BetLimit[XIAN] = int64(BetBase) * 5000
	// game.Rule.BetLimit[HE] = int64(BetBase) * 2000
	// game.Rule.BetLimit[ZHUANGDUI] = int64(BetBase) * 3000
	// game.Rule.BetLimit[XIANDUI] = int64(BetBase) * 3000
	// game.Rule.SitDownLimit = BetBase * 100

	game.Rule.RobotMinGold = config.LongHuConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.LongHuConfig.Robotgold[level-1][1]
}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(baijiale.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}
	msg.BetMinLimit = int64(game.Rule.BetMinLimit)
	msg.UserBetLimit = int32(game.Rule.UserBetLimit)
	u.SendMsg(int32(baijiale.SendToClientMessageType_RoomRolesInfo), msg)
}

func (game *Game) SendRoomInfo() {
	if game.Status == 0 {
		return
	}

	msg := new(baijiale.RoomSenceInfoMsg)
	msg.TrendList = new(baijiale.Trend)
	msg.TrendList.Info = append(msg.TrendList.Info, game.WinTrend...)
	msg.GameStatus = new(baijiale.StatusMessage)
	msg.GameStatus.Status = int32(game.Status)
	if game.TimerJob != nil {
		msg.GameStatus.StatusTime = int32(game.TimerJob.GetTimeDifference())
		msg.GameStatus.TotalStatusTime = int32(game.TimerJob.GetIntervalTime() / time.Millisecond)
	}

	level := int(game.Table.GetLevel())
	msg.RoomID = game.Table.GetRoomID()
	// msg.BaseBet = int64(game.Rule.BetList[0])
	// msg.UserLimit = game.Rule.UserBetLimit
	msg.BaseBet = int64(config.LongHuConfig.Chips5times[int(level)-1][0])
	msg.UserLimit = config.LongHuConfig.Singleuserallspacelimit5times[int(level)-1]
	//发送给框架
	//b, _ := proto.Marshal(msg)
	//game.Table.BroadcastAll(int32(rbwar.SendToClientMessageType_RoomSenceInfo), b)
	game.Table.BroadcastAll(int32(baijiale.SendToClientMessageType_RoomSenceInfo), msg)
}

func (game *Game) testCard(buffer []byte) {
	tmp := &baijiale.TempCardMsg{}
	proto.Unmarshal(buffer, tmp)
	game.testZhuang = tmp.ZhuangPoker
	game.testXian = tmp.XianPoker
	log.Debugf("收到的牌型为：%v", tmp)
	log.Debugf("测试牌为：%v， %v", game.testZhuang, game.testXian)
}

func (game *Game) BroadCastPokerCount() {
	msg := new(baijiale.PokerCount)
	msg.Count = int32(game.gp.GetCardsCount()) - game.LastCardCount
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_GamePokerCount), msg)
}

func (game *Game) AddCards() {
	for i := 0; i < 2; i++ {
		game.gp.AddCard(game.ZhuangCards[i])
		game.gp.AddCard(game.XianCards[i])
		game.ZhuangCards[i] = 0
		game.XianCards[i] = 0
	}

	if game.ZhuangCards[2] != 0 {
		game.gp.AddCard(game.ZhuangCards[2])
	}

	if game.XianCards[2] != 0 {
		game.gp.AddCard(game.XianCards[2])
	}

	game.gp.ShuffleCards()
}

func (g *Game) SendMiPaiSeat() {
	g.SenceSeat.ReCalc()
	msg := new(baijiale.MiPaiUserInfo)

	ZhuangMiUser, ok1 := g.SenceSeat.UserSeat[g.SenceSeat.BetZhuangMaxID]
	XianMiUser, ok2 := g.SenceSeat.UserSeat[g.SenceSeat.BetXianMaxID]
	if ok1 {
		msg.ZhuangSeatID = int32(ZhuangMiUser.SeatNo)
	}

	if ok2 {
		msg.XianSeatID = int32(XianMiUser.SeatNo)
	}
	g.Table.Broadcast(int32(baijiale.SendToClientMessageType_UpdateMiSeatID), msg)
}

func (game *Game) ResetTable() {
	game.Status = 0
	game.Rule.UserBetLimit = 0
	game.RobotTimerJob = nil
}

func (g *Game) PaoMaDeng(Gold int64, user player.PlayerInterface) {
	configs := g.Table.GetMarqueeConfig()
	for _, v := range configs {
		if Gold >= v.AmountLimit {
			err := g.Table.CreateMarquee(user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Debugf("创建跑马灯错误：%v", err)
			}
		}
	}
}

func (game *Game) SendMiPaiSynPos(buf []byte) {
	switch game.Status {
	case baijiale.GameStatus_ShowPoker2,
		baijiale.GameStatus_ShowPoker4,
		baijiale.GameStatus_ShowPoker5,
		baijiale.GameStatus_ShowPoker6:
	default:
		return
	}
	msg := new(baijiale.MiPaiSynPosMsg)
	if err := proto.Unmarshal(buf, msg); err != nil {
		return
	}
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SendMiPaiSynPos), msg)
}

var milock sync.Mutex

func (game *Game) MiEnd(buf []byte) {
	var msg baijiale.MiEnd
	if err := proto.Unmarshal(buf, &msg); err != nil {
		return
	}
	if !msg.IsEnd {
		return
	}

	if game.TimerJob != nil && game.TimerJob.GetTimeDifference() < 2000 {
		return
	}

	milock.Lock()
	defer milock.Unlock()

	if msg.Status != int32(game.Status) {
		return
	}

	switch msg.Status {
	case int32(baijiale.GameStatus_ShowPoker2):
		if game.TimerJob != nil {
			game.TimerJob.Cancel()
		}
		game.PokerMsg.IsOpen.XianOpen[0], game.PokerMsg.IsOpen.XianOpen[1] = true, true
		game.showPoker4()

	case int32(baijiale.GameStatus_ShowPoker4):
		if game.TimerJob != nil {
			game.TimerJob.Cancel()
		}
		game.PokerMsg.IsOpen.ZhuangOpen[0], game.PokerMsg.IsOpen.ZhuangOpen[1] = true, true
		if game.ZhuangCards[2] == 0 && game.XianCards[2] == 0 {
			game.showPokerEnd()
		} else {
			game.showPoker5()
		}

	case int32(baijiale.GameStatus_ShowPoker5):
		if game.TimerJob != nil {
			game.TimerJob.Cancel()
		}
		if game.ZhuangCards[2] != 0 {
			game.PokerMsg.IsOpen.ZhuangOpen[2] = true
		}
		if game.XianCards[2] != 0 {
			game.PokerMsg.IsOpen.XianOpen[2] = true
		}

		if (game.ZhuangCards[2] != 0 && game.XianCards[2] == 0) || (game.ZhuangCards[2] == 0 && game.XianCards[2] != 0) {
			game.showPokerEnd()
		} else {
			game.showPoker6()
		}
	case int32(baijiale.GameStatus_ShowPoker6):
		if game.TimerJob != nil {
			game.TimerJob.Cancel()
		}
		game.PokerMsg.IsOpen.ZhuangOpen[2], game.PokerMsg.IsOpen.XianOpen[2] = true, true
		game.showPokerEnd()
	}
}

func (game *Game) sendMi(mi *baijiale.MiPai) {
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SendRobotMiPai), mi)
}

func (game *Game) genRobotMi(podIndex ...int) *baijiale.MiPai {
	result := new(baijiale.MiPai)
	result.Pai = make([]*baijiale.RobotMiMsgDan, 0)
	for i := 0; i < len(podIndex); i++ {
		result.Pai = append(result.Pai, &baijiale.RobotMiMsgDan{
			PosIndex: int32(podIndex[i]),
			Action:   int32(rand.Intn(10000)),
		})
	}
	return result
}

func (game *Game) RefreshGold() {
	for _, user := range game.AllUserList {
		msg := new(baijiale.SceneGoldMsg)
		msg.Gold = user.User.GetScore()
		user.User.SendMsg(int32(baijiale.SendToClientMessageType_SceneGold), msg)
	}
}
