package game

import (
	"fmt"
	"go-game-sdk/example/game_poker/960304/config"
	"go-game-sdk/example/game_poker/960304/model"
	baijiale "go-game-sdk/example/game_poker/960304/msg"
	"go-game-sdk/lib/clock"
	"math/rand"
	"sort"
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
	Table               table.TableInterface  // table interface
	AllUserList         map[int64]*model.User //所有的玩家列表
	Status              baijiale.GameStatus   // 房间状态1 表示
	Win                 int32                 // 0表示庄胜利，1表示闲胜利，2表示和
	LastWinIsRedOrBlack int                   // 最近一次开龙还是虎
	ZhuangCards         [3]byte               // 庄牌
	XianCards           [3]byte               // 闲牌
	IsLuckWin           bool                  // 幸运一击是否胜利
	BetTotal            [5]int64              //下注统计
	TotalUserBet        [5]int64              //下注统计
	SenceSeat           model.SceneInfo       //下注的玩家列表
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
	sysCheat            string                //
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.SenceSeat.Init()
	game.PokerMsg = new(baijiale.PokerMsg)
	game.PokerMsg.ZhuangPoker = make([]byte, 3)
	game.PokerMsg.XianPoker = make([]byte, 3)
	//test
	//game.Table.AddTimer(1000, game.Start)
}

func (game *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

//用户坐下
func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	game.getUser(user)
	return 1 //business.OnActionUserSitDownHandler()
}

func (game *Game) UserExit(user player.PlayerInterface) bool {
	u := game.getUser(user)
	//有下注时不让玩家离开
	if u.TotalBet != 0 {
		return false
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}
	delete(game.AllUserList, user.GetID())
	//删除用户列表用户
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u := game.getUser(user)
	if u.TotalBet != 0 {
		msg := new(baijiale.ExitFail)
		msg.FailReason = "游戏中不能退出！"
		user.SendMsg(int32(baijiale.SendToClientMessageType_ExitRet), msg)
		return false
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}
	delete(game.AllUserList, user.GetID())
	game.DeleteExitUserFromOnlineUserListSlice(u)
	//game.RandSelectUserSitDownChair()
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	if int32(baijiale.ReceiveMessageType_BetID) == subCmd {
		game.Bet(buffer, user)
	} else if int32(baijiale.ReceiveMessageType_SitDown) == subCmd {
		//game.UserSitDown(buffer, user)
	} else if int32(baijiale.ReceiveMessageType_GetTrend) == subCmd {
		game.SendTrend(user)
	} else if int32(baijiale.ReceiveMessageType_GetUserListInfo) == subCmd {
		game.SendUserListInfo(user)
	} else if int32(baijiale.ReceiveMessageType_StandUp) == subCmd {
		//game.OnUserStanUp(user)
	} else if int32(baijiale.ReceiveMessageType_tempCard) == subCmd {
		//game.testCard(buffer)
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
	game.SendUserBet(u)

	if game.Status >= baijiale.GameStatus_ShowPoker {
		user.SendMsg(int32(baijiale.SendToClientMessageType_PokerInfo), game.PokerMsg)
		if game.Status == baijiale.GameStatus_SettleStatus {
			if u.SettleMsg != nil {
				user.SendMsg(int32(baijiale.SendToClientMessageType_UserComeBack), u.SettleMsg)
			}
		}
	}
	if game.TimerJob != nil {
		game.SendToUserStatusMsg(int(game.TimerJob.GetTimeDifference()), user)
	}
	game.SendTrend(user)

	return true
}

func (game *Game) Start() {
	game.checkUserBet()
	//选择列表中前6个用户上座
	game.SelectUserListInfoBefore6SitDownChair()
	if game.Table.GetRoomID() == -1 {
		game.Status = 0
		return
	}

	//起立不能坐下的人
	//for _, u := range game.SenceSeat.UserSeat {
	//	if u.User.User.GetScore() < int64(game.Rule.SitDownLimit) {
	//		game.OnUserStanUp(u.User.User)
	//	}
	//}

	game.LastMasterBetType = -1
	game.Table.StartGame()

	game.BroadCastPokerCount()
	//if game.RobotTimerJob == nil {
	//	//r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
	//	//game.RobotTimerJob, _ = game.Table.AddTimer(int64(r), game.RobotSitDown)
	//	//r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
	//	//game.Table.AddTimer(int64(r1), game.RobotStandUp)
	//}
	game.Status = baijiale.GameStatus_StartMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startmove), game.StartBet)

	//开始动画消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startmove)
}

func (game *Game) StartBet() {
	game.ResetData()
	game.Status = baijiale.GameStatus_BetStatus
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startbet), game.EndBet)

	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startbet)
}

func (game *Game) EndBet() {
	game.Status = baijiale.GameStatus_EndBetMovie
	//game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endmove), game.Settle)
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endmove), game.getResult)
	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Endmove)
}

//结算
func (game *Game) Settle() {
	/*if len(game.testZhuang) == 0 {
		game.getResult()
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
	}*/

	game.Status = baijiale.GameStatus_SettleStatus
	game.sendSettleMsg()

	endtime := config.LongHuConfig.Taketimes.Endpay

	if game.gp.GetCardsCount()-int(game.LastCardCount) >= 6 {
		game.TimerJob, _ = game.Table.AddTimer(int64(endtime), game.Start)
	} else {
		game.TimerJob, _ = game.Table.AddTimer(int64(endtime), game.InitPoker)
	}

	//game.checkUserBet()
	//发送开始下注消息
	game.SendStatusMsg(endtime)
	game.Table.EndGame()
}

func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(baijiale.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(baijiale.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(baijiale.SendToClientMessageType_Status), msg)
}

func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != baijiale.GameStatus_BetStatus {
		return
	}
	//用户下注
	BetPb := &baijiale.Bet{}
	proto.Unmarshal(buffer, BetPb)
	u := game.getUser(user)
	if u.Bet(BetPb, game.BetTotal) {
		game.BetTotal[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])

		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		}

		u.User.SetScore(game.Table.GetGameNum(), -int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))]), game.Table.GetRoomRate())
		//if game.SenceSeat.UserBet(u) {
		//	game.SendMiPaiSeat()
		//}
	}

	//if game.SenceSeat.GetMaster() == u.SceneChairId {
	//	game.LastMasterBetType = BetPb.BetType
	//}
	//神算子下注区域
	u1, ok := game.SenceSeat.UserSeat[user.GetID()]
	if ok {
		if u1.User.Icon == 1 {
			game.LastMasterBetType = BetPb.BetType
		}

	}
}

func (game *Game) UserSitDown(buffer []byte, user player.PlayerInterface) {
	us := &baijiale.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo)) {
			u.SceneChairId = int(us.ChairNo)

			game.SendSceneMsg(nil)

			//game.SendMiPaiSeat()
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
				game.OnUserStanUp(u.User)
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
	ZhuangMiUser, ok1 := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	XianMiUser, ok2 := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]

	ZhuangBetCount := 0
	XianBetCount := 0
	HeBetCount := 0
	ZhuangDuiBetCount := 0
	XianDuiBetCount := 0
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [5]int64

	ZhuangStr := fmt.Sprintf("庄区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[ZHUANG]), score.GetScoreStr(game.BetTotal[ZHUANG]-game.TotalUserBet[ZHUANG]), score.GetScoreStr(game.TotalUserBet[ZHUANG]))

	XianStr := fmt.Sprintf("闲区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[XIAN]), score.GetScoreStr(game.BetTotal[XIAN]-game.TotalUserBet[XIAN]), score.GetScoreStr(game.TotalUserBet[XIAN]))

	HeStr := fmt.Sprintf("和区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[HE]), score.GetScoreStr(game.BetTotal[HE]-game.TotalUserBet[HE]), score.GetScoreStr(game.TotalUserBet[HE]))

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

		SceneUserInfo := new(baijiale.SceneUserInfo)

		msg := new(baijiale.SettleMsg)
		msg.IsXianDui = IsXianDui
		msg.IsZhuangDui = IsZhuangDui
		var win int64
		var totalTax int64 //总税收
		if u.TotalBet > 0 {
			if game.Win == 0 {
				msg.UserZhuangWin += u.BetArea[ZHUANG]
				win += u.BetArea[ZHUANG] * 2
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], 0)
				Gold += capital
				totalTax += u.BetArea[ZHUANG]*2 - Gold
				msg.TotalWin += Gold
				SceneUserInfo.ZhuangWin = msg.UserZhuangWin
				SceneUserInfo.XianWin = -u.BetArea[XIAN]
				SystemWin[ZHUANG] += Gold
			} else if game.Win == 1 {
				msg.UserXianWin += u.BetArea[XIAN]
				win += u.BetArea[XIAN] * 2
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], 0)
				Gold += capital
				totalTax += u.BetArea[XIAN]*2 - Gold
				msg.TotalWin += Gold
				SceneUserInfo.ZhuangWin -= u.BetArea[ZHUANG]
				SceneUserInfo.XianWin = msg.UserXianWin
				SystemWin[XIAN] += Gold
			} else if game.Win == HE {
				msg.HeWin = u.BetArea[HE] * int64(HEODDS+1)
				win += u.BetArea[HE] * int64(HEODDS+1)
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[HE]*int64(HEODDS), game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[HE], 0)
				Gold += capital
				totalTax += u.BetArea[HE]*int64(HEODDS+1) - Gold
				msg.TotalWin += Gold + u.BetArea[ZHUANG] + u.BetArea[XIAN]
				//把压龙和虎的钱退回
				u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANG], 0)
				u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIAN], 0)
				SystemWin[HE] += Gold + u.BetArea[ZHUANG] + u.BetArea[XIAN]

			}

			if IsZhuangDui {
				win += u.BetArea[ZHUANGDUI] * 12
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANGDUI]*11, game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[ZHUANGDUI], 0)
				Gold += capital
				totalTax += u.BetArea[ZHUANGDUI]*12 - Gold
				msg.TotalWin += Gold
				msg.ZhuangDui = Gold
				SceneUserInfo.ZhuangDui = msg.ZhuangDui
				SystemWin[ZHUANGDUI] += Gold
			}

			if IsXianDui {
				win += u.BetArea[XIANDUI] * 12
				Gold, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIANDUI]*11, game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[XIANDUI], 0)
				Gold += capital
				totalTax += u.BetArea[XIANDUI]*12 - Gold
				msg.TotalWin += Gold
				msg.XianDui = Gold
				SceneUserInfo.XianDui = msg.XianDui
				SystemWin[XIANDUI] += Gold
			}
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
		//统计玩家信息
		if (win) > u.TotalBet {
			u.UserCount(true, msg.TotalWin)
		} else {
			u.UserCount(false, 0)
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
		game.CountUser(u)

		msg.Win = int32(game.Win)
		msg.UserScore = u.User.GetScore()
		msg.ZhuangMi = game.SenceSeat.BetZhuangMaxID
		msg.XianMi = game.SenceSeat.BetXianMaxID
		if ok1 {
			msg.ZhuangMiUserNikeName = ZhuangMiUser.User.User.GetNike()
			msg.ZhuangMiUserHead = ZhuangMiUser.User.User.GetHead()
		}

		if ok2 {
			msg.XiangMiUserNikeName = XianMiUser.User.User.GetNike()
			msg.XiangMiUserHead = XianMiUser.User.User.GetHead()
		}

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
			u.User.SendChip(chip)
		} else {
			u.SettleMsg = nil
		}

		user := game.getUser(u.User)
		betsAmount := u.TotalBet
		profitAmount := u.User.GetScore() - user.CruenSorce
		u.ResetUserData()

		if u.SceneChairId != 0 {
			SceneSettleMsg.UserList = append(SceneSettleMsg.UserList, SceneUserInfo)
		}

		u.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, totalTax, msg.TotalWin, "")
	}
	cou := model.Usercount{}
	cou = game.CountUserList
	sort.Sort(cou)
	game.SetIcon()
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
	game.Table.WriteLogs(0, totalstr)

	game.SenceSeat.BetXianMaxID = 0
	game.SenceSeat.BetZhuangMaxID = 0
	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

func (game *Game) getResult() {
	game.Status = baijiale.GameStatus_ShowPoker
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

	if r < v[0] {
		eat = 1
	} else if r < v[1] {
		out = 1
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
		for i := 0; i < 100; i++ {
			//如果不符合要求，发过的牌还回去
			if bAddCards {
				game.AddCards()
			}

			game.DealPoker()
			bAddCards = true
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
				PayMoney = game.TotalUserBet[ZHUANG]
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
		endtime -= 2100
	}

	if game.XianCards[2] == 0 {
		endtime -= 2100
	}

	game.Table.Broadcast(int32(baijiale.SendToClientMessageType_PokerInfo), game.PokerMsg)
	game.TimerJob, _ = game.Table.AddTimer(int64(endtime), game.Settle)
	game.SendStatusMsg(endtime)
}

func (game *Game) getUser(user player.PlayerInterface) *model.User {
	u, ok := game.AllUserList[user.GetID()]
	if !ok {
		u = new(model.User)
		game.AllUserList[user.GetID()] = u
		u.Table = game.Table
		u.User = user
		u.Rule = &game.Rule
		u.Time = time.Now().UnixNano() / 1e6
		game.CountUserList = append(game.CountUserList, u)
		u.ResetUserData()
	}

	return u
}

//发送场景消息
func (game *Game) SendSceneMsg(u player.PlayerInterface) {
	msg := new(baijiale.SceneMessage)
	//bigwinner := game.SenceSeat.GetBigWinner()
	//master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(baijiale.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetID())
		su.IsMaster = false
		su.IsBigWinner = false
		su.IsMillionaire = false
		switch v.User.Icon {
		case Master:
			su.IsMaster = true
		case bigWinner:
			su.IsBigWinner = true
		case Millionaire:
			su.IsMillionaire = true
		}

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
	//初始化用户称号
	u.Icon = 0
	//uc := len(game.CountUserList)
	//if uc == 0 {
	//	game.CountUserList = append(game.CountUserList, u)
	//	return
	//}
	adduser := u
	//if game.CountUserList[0].RetWin < u.RetWin {
	//	adduser = game.CountUserList[0]
	//	game.CountUserList[0] = u
	//}
	//
	//for i := 1; i < uc; i++ {
	//	if adduser.RetBet > game.CountUserList[i].RetBet {
	//		var tmp []*model.User
	//		tmp = append(tmp, game.CountUserList[i:]...)
	//		game.CountUserList = append(game.CountUserList[0:i], adduser)
	//		game.CountUserList = append(game.CountUserList, tmp...)
	//		return
	//	}
	//}

	game.CountUserList = append(game.CountUserList, adduser)
}

func (game *Game) SendUserListInfo(user player.PlayerInterface) {
	msg := new(baijiale.UserList)
	for _, u := range game.CountUserList {
		userinfo := new(baijiale.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.RetWinMoney
		userinfo.Head = u.User.GetHead()
		userinfo.Icon = u.Icon
		msg.UserList = append(msg.UserList, userinfo)
	}

	user.SendMsg(int32(baijiale.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) ResetData() {
	for i := 0; i < 5; i++ {
		game.TotalUserBet[i] = 0
		game.BetTotal[i] = 0
	}
}

func (game *Game) OnUserStanUp(user player.PlayerInterface) {
	bSendMiPai := false
	if user.GetID() == game.SenceSeat.BetZhuangMaxID || user.GetID() == game.SenceSeat.BetXianMaxID {
		bSendMiPai = true
	}

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
					v.User.GetScore() > int64(game.Rule.SitDownLimit) {
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
					game.OnUserStanUp(user)
				}
			}
		}
	}
}

func (game *Game) BrodCastSceneMsg() {
	msg := new(baijiale.SceneMessage)
	bigwinner := game.SenceSeat.GetBigWinner()
	master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(baijiale.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetID())
		if bigwinner == v.SeatNo {
			su.IsBigWinner = true
		} else {
			su.IsBigWinner = false
		}

		if master == v.SeatNo {
			su.IsMaster = true
		} else {
			su.IsMaster = false
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
	log.Tracef("配置 %v", str)
	js, err := simplejson.NewJson([]byte(str))
	if err != nil {
		fmt.Printf("解析房间配置失败 err%v\n", err)
		fmt.Printf("%v\n", str)
		return
	}

	game.Rule.BetList = make([]int64, 0)
	BetBase, _ := js.Get("Bottom_Pouring").Int()
	betMinLimit, _ := js.Get("betMinLimit").Int()
	game.Rule.BetMinLimit = int64(betMinLimit)
	//game.Rule.UserBetLimit = int64(BetBase) * 5000
	//game.Rule.BetList = append(game.Rule.BetList, BetBase)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*10)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*50)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*100)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*500)
	//game.Rule.BetLimit[ZHUANG] = int64(BetBase) * 5000
	//game.Rule.BetLimit[XIAN] = int64(BetBase) * 5000
	//game.Rule.BetLimit[HE] = int64(BetBase) * 2000
	//game.Rule.BetLimit[ZHUANGDUI] = int64(BetBase) * 3000
	//game.Rule.BetLimit[XIANDUI] = int64(BetBase) * 3000
	game.Rule.SitDownLimit = BetBase * 100
	level := game.Table.GetLevel()
	game.Rule.RobotMinGold = config.LongHuConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.LongHuConfig.Robotgold[level-1][1]
	game.Rule.SingleUserAllSpaceLimit = config.LongHuConfig.Singleuserallspacelimit5times[level-1]
	game.Rule.AllSpaceLimit = config.LongHuConfig.Allspacelimit5times[level-1]
	for i := 0; i < 5; i++ {
		//log.Traceln(config.BRNNConfig.Singleusersinglespacelimit5times[level-1][3],config.BRNNConfig.Allusersinglespacelimit5times[level-1][i])
		game.Rule.SingleUserSingleSpaceLimit[i] = config.LongHuConfig.Singleusersinglespacelimit5times[level-1][i]
		game.Rule.AllUserSingleSpaceLimit[i] = config.LongHuConfig.Allusersinglespacelimit5times[level-1][i]
	}
	for i := 0; i < 5; i++ {
		game.Rule.BetList = append(game.Rule.BetList, config.LongHuConfig.Chips5times[level-1][i])
	}
	game.Rule.UserBetLimit = game.Rule.SingleUserAllSpaceLimit
	//log.Traceln(game.Rule.BetList,":",game.Rule.SingleUserSingleSpaceLimit,":",game.Rule.SingleUserAllSpaceLimit,":",game.Rule.AllUserSingleSpaceLimit,":",game.Rule.AllSpaceLimit)
}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(baijiale.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}

	msg.UserBetLimit = int32(game.Rule.UserBetLimit)
	msg.BetMinLimit = game.Rule.BetMinLimit
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

	msg.RoomID = game.Table.GetRoomID()
	msg.BaseBet = int64(game.Rule.BetList[0])
	msg.UserLimit = game.Rule.UserBetLimit
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

//关闭桌子
func (game *Game) CloseTable() {
}

func (game *Game) DeleteExitUserFromOnlineUserListSlice(user *model.User) {
	//for k, v := range game.OnlineUserList {
	//	if user == v {
	//		game.OnlineUserList = append(game.OnlineUserList[:k], game.OnlineUserList[k+1:]...)
	//		break
	//	}
	//}
	for k, v := range game.CountUserList {
		if user == v {
			game.CountUserList = append(game.CountUserList[:k], game.CountUserList[k+1:]...)
			break
		}
	}
}

//玩家列表中前6用户上座
func (game *Game) SelectUserListInfoBefore6SitDownChair() {
	//game.SenceSeat.Init()
	for _, v := range game.SenceSeat.SenceSeat {
		u, ok := game.AllUserList[v.User.User.GetID()]
		if ok {
			u.SceneChairId = 0
		}
	}
	game.SenceSeat.Init()
	index := len(game.CountUserList)
	if index >= 6 {
		index = 6
	}
	for i := 0; i < index; i++ {
		u := game.CountUserList[i]
		ChairId := game.SenceSeat.GetSceneChairId()
		if ChairId == 0 {
			game.SendSceneMsg(nil)
			return
		}
		//if u.User == game.Zhuang {
		//	continue
		//}
		//检测遍历到的用户是否在椅子上，如无此用户 让用户坐下
		if game.SenceSeat.CheckUserOnChair(u.User.GetID()) {
			if game.SenceSeat.SitScene(u, ChairId) {
				u.SceneChairId = ChairId
			}
		}
	}

	game.SendSceneMsg(nil)
}

const (
	Master      = 1 //神算子
	bigWinner   = 2 //大富豪
	Millionaire = 3 //大赢家

)

//设置称号
func (game *Game) SetIcon() {
	//大赢家3
	bigWinnerid := int64(0)
	Millionaireid := int64(0)
	mastid := int64(0)
	var user []*model.User
	//log.Traceln("chushihua ")
	for k, v := range game.CountUserList {
		if k >= 6 {
			break
		}
		user = append(user, v)
	}
	//大赢家1上一局赢钱最多的优先级最高
	sort.Sort(model.BigwinnerUser(user))
	if len(user) < 1 {
		return
	}
	Millionaireid = user[0].User.GetID()

	u, ok := game.AllUserList[Millionaireid]
	if ok {
		u.Icon = Millionaire
		//log.Traceln("大赢家", u.User.GetID())
	}
	if len(user) == 1 {
		return
	}

	//大富豪2 近20局赢钱最多的 大赢家> 大富豪> 神算子
	sort.Sort(model.RegalUser(user))
	for i := 0; i < len(user); i++ {
		if user[i].User.GetID() != Millionaireid {
			bigWinnerid = user[i].User.GetID()
			break
		}
	}
	u1, ok1 := game.AllUserList[bigWinnerid]
	if ok1 {
		u1.Icon = bigWinner
		//log.Traceln("大富豪",u1.User.GetID())
	}
	//神算子1 胜率最高的 大赢家> 大富豪> 神算子
	sort.Sort(model.MasterUser(user))
	for i := 0; i < len(user); i++ {
		if user[i].User.GetID() != Millionaireid && user[i].User.GetID() != bigWinnerid {
			mastid = user[i].User.GetID()
			break
		}
	}
	u2, ok2 := game.AllUserList[mastid]
	if ok2 {
		u2.Icon = Master
		//log.Traceln("神算子",u2.User.GetID())
	}
}
