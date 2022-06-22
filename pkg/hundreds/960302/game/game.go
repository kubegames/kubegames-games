package game

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960302/config"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960302/model"
	longhu "github.com/kubegames/kubegames-games/pkg/hundreds/960302/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

const (
	LONG = 0
	HU   = 1
	HE   = 2
)

const (
	HEODDS = 8
)

type Game struct {
	Table               table.TableInterface  // table interface
	AllUserList         map[int64]*model.User //所有的玩家列表
	Status              longhu.GameStatus     // 房间状态1 表示
	Win                 int32                 // 0表示龙胜利，1表示虎胜利，2表示和
	LastWinIsRedOrBlack int                   // 最近一次开龙还是虎
	LongCards           byte                  // 龙牌
	HuCards             byte                  // 虎牌
	IsLuckWin           bool                  // 幸运一击是否胜利
	BetTotal            [3]int64              //龙虎和的下注统计
	TotalUserBet        [3]int64              //龙虎和的下注统计
	SenceSeat           model.SceneInfo       //下注的玩家列表
	TimerJob            *table.Job            //job
	RobotTimerJob       *table.Job            //机器人job
	LastMasterBetType   int32                 //最近一次神算子下注的类型
	WinTrend            []int32               //赢的走势
	CountUserList       []*model.User         //统计后的玩家列表
	Rule                config.RoomRules      //房间规则信息
	CheatValue          int                   //记录本次作弊值
	PokerMsg            *longhu.PokerMsg      //扑克消息
	HasTest             bool                  //是否有测试牌
	OnlineUserList      []*model.User         //所有的玩家列表 用于自动上座，按顺序上坐
	sysCheat            string
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.OnlineUserList = make([]*model.User, 0)
	game.SenceSeat.Init()
	game.PokerMsg = new(longhu.PokerMsg)
	game.PokerMsg.LongPoker = make([]byte, 1)
	game.PokerMsg.HuPoker = make([]byte, 1)
	//test
	//game.Table.AddTimer(1000, game.Start)
}

func (game *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

//用户坐下
func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) table.MatchKind {
	game.getUser(user)

	return table.SitDownOk
}

func (game *Game) UserOffline(user player.PlayerInterface) bool {
	u := game.getUser(user)
	//有下注时不让玩家离开
	if u.TotalBet != 0 {
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

func (game *Game) UserLeaveGame(user player.PlayerInterface) bool {
	u := game.getUser(user)
	if u.TotalBet != 0 {
		msg := new(longhu.ExitFail)
		msg.FailReason = "游戏中不能退出！"
		user.SendMsg(int32(longhu.SendToClientMessageType_ExitRet), msg)
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
	if int32(longhu.ReceiveMessageType_BetID) == subCmd {
		game.Bet(buffer, user)
	} else if int32(longhu.ReceiveMessageType_SitDown) == subCmd {
		//game.UserSitDown(buffer, user)
	} else if int32(longhu.ReceiveMessageType_GetTrend) == subCmd {
		game.SendTrend(user)
	} else if int32(longhu.ReceiveMessageType_GetUserListInfo) == subCmd {
		game.SendUserListInfo(user)
	} else if int32(longhu.ReceiveMessageType_StandUp) == subCmd {
		//game.OnUserStanUp(user)
	} else if int32(longhu.ReceiveMessageType_tempCard) == subCmd {
		//game.OnTest(buffer)
	}
}

func (game *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	robot.Init(ai, game)
	return robot
}

func (game *Game) SendScene(user player.PlayerInterface) {
	game.GetRoomconfig()
	u := game.getUser(user)
	// if user.IsRobot() {
	// 	robot := new(Robot)
	// 	robotUser := user.BindRobot(robot)
	// 	robot.Init(robotUser, game)
	// 	return true
	// }
	game.SendRuleInfo(user)
	game.SendSceneMsg(user)
	game.SendUserBet(u)

	if game.Status >= longhu.GameStatus_ShowPoker {
		u.User.SendMsg(int32(longhu.SendToClientMessageType_PokerInfo), game.PokerMsg)
		if game.Status == longhu.GameStatus_SettleStatus {
			if u.SettleMsg != nil {
				user.SendMsg(int32(longhu.SendToClientMessageType_UserComeBack), u.SettleMsg)
			}
		}
	}

	if game.TimerJob != nil {
		game.SendToUserStatusMsg(int(game.TimerJob.GetTimeDifference()), user)
	}

	game.SendTrend(user)
}

func (game *Game) Start() {
	//选择列表中前6个用户上座
	game.SelectUserListInfoBefore6SitDownChair()
	// if game.Table.GetRoomID() == -1 {
	// 	game.Status = 0
	// 	return
	// }

	game.LastMasterBetType = -1
	game.Table.StartGame()
	//if game.RobotTimerJob == nil {
	//	//r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
	//	//game.RobotTimerJob, _ = game.Table.AddTimer(int64(r), game.RobotSitDown)
	//	//r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
	//	//game.Table.AddTimer(int64(r1), game.RobotStandUp)
	//}
	//game.RandSelectUserSitDownChair()
	//坐下限制取消
	//for _, u := range game.SenceSeat.SenceSeat {
	//	if u.User.User.GetScore() < int64(u.User.Rule.SitDownLimit) {
	//		game.OnUserStanUp(u.User.User)
	//	}
	//}

	game.Status = longhu.GameStatus_StartMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startmove), game.StartBet)

	//开始动画消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startmove)
}

func (game *Game) StartBet() {
	game.ResetData()
	game.Status = longhu.GameStatus_BetStatus
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Startbet), game.EndBet)
	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Startbet)
}

func (game *Game) EndBet() {
	game.Status = longhu.GameStatus_EndBetMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endmove), game.getResult)
	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Endmove)
}

//结算
func (game *Game) Settle() {
	game.Status = longhu.GameStatus_SettleStatus
	//game.getResult()
	game.sendSettleMsg()

	//发送开始下注消息
	game.SendStatusMsg(config.LongHuConfig.Taketimes.Endpay)

	//结束游戏
	game.Table.EndGame()

	//踢出离线用户
	for _, u := range game.AllUserList {
		if u.User.IsOnline() == false {
			game.Table.KickOut(u.User)
		}
	}

	//检测下注情况
	game.checkUserBet()

	//桌子是否需要关闭
	if game.Table.IsClose() {
		//踢人
		for k, u := range game.AllUserList {
			//踢掉所以用户
			u.NoBetCount = 0
			if u.SceneChairId != 0 {
				game.OnUserStanUp(u.User)
			}
			delete(game.AllUserList, k)
			game.DeleteExitUserFromOnlineUserListSlice(u)
			game.Table.KickOut(u.User)
		}
		if game.Table.PlayerCount() <= 0 {
			game.Table.Close()
			return
		}
	}

	game.TimerJob, _ = game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.Endpay), game.Start)
}

func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(longhu.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	game.Table.Broadcast(int32(longhu.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(longhu.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(longhu.SendToClientMessageType_Status), msg)
}

func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != longhu.GameStatus_BetStatus {
		return
	}
	//用户下注
	BetPb := &longhu.Bet{}
	proto.Unmarshal(buffer, BetPb)
	u := game.getUser(user)
	if u.Bet(BetPb, game.BetTotal) {
		game.BetTotal[BetPb.BetType%3] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])

		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType%3] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		}

		u.User.SetScore(game.Table.GetGameNum(), -int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))]), game.Table.GetRoomRate())
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
	us := &longhu.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo)) {
			u.SceneChairId = int(us.ChairNo)
			for _, v := range game.AllUserList {
				game.SendSceneMsg(v.User)
			}
		}
	}
}

// 发牌
func (game *Game) DealPoker() {
	var gp model.GamePoker
	gp.InitPoker()
	gp.ShuffleCards()

	for i := 0; i < 3; i++ {
		game.LongCards = gp.DealCards()
		game.HuCards = gp.DealCards()
	}
}

// 比牌， 1表示红胜利，2表示黑方胜利
func (game *Game) ComparePoker() {
	long, _ := model.GetCardValueAndColor(game.LongCards)
	hu, _ := model.GetCardValueAndColor(game.HuCards)

	if long == hu {
		game.Win = HE
	} else if long > hu {
		game.Win = LONG
	} else {
		game.Win = HU
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
			game.DeleteExitUserFromOnlineUserListSlice(u)
			game.Table.KickOut(u.User)
		}
	}
}

//发送结算消息
func (game *Game) sendSettleMsg() {
	game.WinTrend = append(game.WinTrend, game.Win)

	winlen := len(game.WinTrend)
	if winlen > 100 {
		game.WinTrend = append(game.WinTrend[:(winlen-100-1)], game.WinTrend[(winlen-100):]...)
	}

	game.CountUserList = make([]*model.User, 0)
	SceneSettleMsg := new(longhu.SceneUserSettle)

	LongBetCount := 0
	HuBetCount := 0
	HeBetCount := 0
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [3]int64
	LongStr := fmt.Sprintf("龙区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[LONG]), score.GetScoreStr(game.BetTotal[LONG]-game.TotalUserBet[LONG]), game.TotalUserBet[LONG])

	HuStr := fmt.Sprintf("虎区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[HU]), score.GetScoreStr(game.BetTotal[HU]-game.TotalUserBet[HU]), game.TotalUserBet[HU])

	HeStr := fmt.Sprintf("和区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[HE]), score.GetScoreStr(game.BetTotal[HE]-game.TotalUserBet[HE]), game.TotalUserBet[HE])

	//战绩
	var records []*platform.PlayerRecord

	for _, u := range game.AllUserList {
		u.NoBetCount++
		if !u.User.IsRobot() {
			if u.NoBetCount >= (config.LongHuConfig.Unplacebetnum + 1) {
				//发送踢掉用户
				msg := new(longhu.KickOutUserMsg)
				msg.KickOutReason = "由于您5局未下注，已被踢出房间！"
				u.User.SendMsg(int32(longhu.SendToClientMessageType_KickOutUser), msg)
			}
		}

		SceneUserInfo := new(longhu.SceneUserInfo)

		msg := new(longhu.SettleMsg)
		var win int64
		var totalTax int64 //总税收
		if u.TotalBet > 0 {
			if game.Win == 0 {
				msg.UserLongWin += u.BetLong
				win += u.BetLong * 2
				Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetLong, game.Table.GetRoomRate())
				capital := u.User.SetScore(game.Table.GetGameNum(), u.BetLong, 0)
				Gold += capital
				totalTax += u.BetLong*2 - Gold
				msg.TotalWin += Gold
				SceneUserInfo.LongWin = msg.UserLongWin
				SceneUserInfo.HuWin = -u.BetHu
				SystemWin[LONG] += Gold
			} else if game.Win == 1 {
				msg.UserHuWin += u.BetHu
				win += u.BetHu * 2
				Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetHu, game.Table.GetRoomRate())
				capital := u.User.SetScore(game.Table.GetGameNum(), u.BetHu, 0)
				Gold += capital
				totalTax += u.BetHu*2 - Gold
				msg.TotalWin += Gold
				SceneUserInfo.LongWin -= u.BetLong
				SceneUserInfo.HuWin = msg.UserHuWin
				SystemWin[HU] += Gold
			} else if game.Win == HE {
				msg.HeWin = u.BetHe * int64(HEODDS)
				win += u.BetHe*int64(HEODDS+1) + u.BetLong + u.BetHu
				Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetHe*int64(HEODDS), game.Table.GetRoomRate())
				capital := u.User.SetScore(game.Table.GetGameNum(), u.BetHe, 0)
				Gold += capital
				totalTax += u.BetHe*int64(HEODDS+1) - Gold
				msg.TotalWin += Gold + u.BetLong + u.BetHu
				//把压龙和虎的钱退回
				u.User.SetScore(game.Table.GetGameNum(), u.BetLong, 0)
				u.User.SetScore(game.Table.GetGameNum(), u.BetHu, 0)
				SystemWin[HE] += Gold + u.BetLong + u.BetHu
			}
		}

		msg.UserBetLong = u.BetLong
		msg.UserBetHu = u.BetHu
		msg.UserBetHe = u.BetHe
		msg.Long = game.BetTotal[LONG]
		msg.Hu = game.BetTotal[HU]
		msg.He = game.BetTotal[HE]

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
			if u.BetLong != 0 {
				temp += fmt.Sprintf("龙：%v ", score.GetScoreStr(u.BetLong))
				temp1 += fmt.Sprintf("龙：%v ", score.GetScoreStr(msg.UserLongWin))
				LongBetCount++
			}

			if u.BetHu != 0 {
				temp += fmt.Sprintf("虎：%v ", score.GetScoreStr(u.BetHu))
				temp1 += fmt.Sprintf("虎：%v ", score.GetScoreStr(msg.UserHuWin))
				HuBetCount++
			}

			if u.BetHe != 0 {
				temp += fmt.Sprintf("和：%v ", score.GetScoreStr(u.BetHe))
				temp1 += fmt.Sprintf("和：%v ", score.GetScoreStr(msg.HeWin))
				HeBetCount++
			}
			temp1 += fmt.Sprintf(" 总输赢：%v，用户剩余金额：%v \r\n", score.GetScoreStr(win-u.TotalBet), score.GetScoreStr(u.User.GetScore()))
			temp += temp1
			game.Table.WriteLogs(u.User.GetID(), temp)
		}

		if u.BetHe != 0 {
			SceneUserInfo.HeWin = msg.HeWin
		} else {
			SceneUserInfo.HeWin -= u.BetHe
		}

		SceneUserInfo.BetLong = msg.UserBetLong
		SceneUserInfo.BetHu = msg.UserBetHu
		SceneUserInfo.BetHe = msg.UserBetHe
		SceneUserInfo.TotalWin = msg.TotalWin
		SceneUserInfo.UserID = int64(u.User.GetID())
		SceneUserInfo.SceneSeatID = int32(u.SceneChairId)
		//统计玩家信息
		if (win) > u.TotalBet {
			u.UserCount(true, msg.TotalWin)
		} else {
			u.UserCount(false, 0)
		}

		game.PaoMaDeng(msg.TotalWin-u.TotalBet, u.User)
		game.CountUser(u)

		//msg.LongPoker = append(msg.LongPoker, game.LongCards)
		//msg.HuPoker = append(msg.HuPoker, game.HuCards)
		msg.Win = int32(game.Win)
		msg.UserScore = u.User.GetScore()
		SceneUserInfo.UserScore = msg.UserScore

		u.User.SendMsg(int32(longhu.SendToClientMessageType_Settle), msg)

		if u.TotalBet > 0 && !u.User.IsRobot() {
			chip := u.BetLong - u.BetHu
			if chip < 0 {
				chip = -chip
			}

			chip += u.BetHe
			if game.Win == HE {
				chip = u.BetHe
			}
			u.User.SendChip(chip)
			u.SettleMsg = msg
		} else {
			u.SettleMsg = nil
		}
		user := game.getUser(u.User)
		betsAmount := u.TotalBet
		profitAmount := u.User.GetScore() - user.CruenSorce
		u.ResetUserData()
		//u.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, totalTax, msg.TotalWin, "")
		if !u.User.IsRobot() {
			records = append(records, &platform.PlayerRecord{
				PlayerID:     uint32(u.User.GetID()),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: profitAmount,
				BetsAmount:   betsAmount,
				DrawAmount:   totalTax,
				OutputAmount: msg.TotalWin,
				Balance:      u.User.GetScore(),
				UpdatedAt:    time.Now(),
				CreatedAt:    time.Now(),
			})
		}

		if u.SceneChairId != 0 {
			SceneSettleMsg.UserList = append(SceneSettleMsg.UserList, SceneUserInfo)
		}
	}
	if len(records) > 0 {
		if _, err := game.Table.UploadPlayerRecord(records); err != nil {
			log.Errorf("upload player record error %s", err.Error())
		}
	}

	cou := model.Usercount{}
	cou = game.CountUserList
	sort.Sort(cou)
	game.SetIcon()
	LongStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", LongBetCount, score.GetScoreStr(game.BetTotal[LONG]-SystemWin[LONG]))
	HuStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", HuBetCount, score.GetScoreStr(game.BetTotal[HU]-SystemWin[HU]))
	HeStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", HeBetCount, score.GetScoreStr(game.BetTotal[HE]-SystemWin[HE]))
	t := ""
	if game.Win == 0 {
		t = "龙赢"
	} else if game.Win == 1 {
		t = "虎赢"
	} else if game.Win == 2 {
		t = "和赢"
	}
	str := fmt.Sprintf("%v作弊率：%v \r\n开局结果 :%v:龙区域牌：%v ,", game.sysCheat, game.CheatValue,
		t, model.GetCardString(game.LongCards))

	str += fmt.Sprintf("虎区域牌型：%v \r\n", model.GetCardString(game.HuCards))

	str += fmt.Sprintf("系统输赢额度：%v \r\n",
		score.GetScoreStr(game.BetTotal[LONG]+game.BetTotal[HU]+game.BetTotal[HE]-SystemWin[LONG]-SystemWin[HU]-SystemWin[HE]))

	str += fmt.Sprintf("最高获利用户ID：%v 获得：%v\r\n",
		MaxWinUserID, score.GetScoreStr(MaxWinGold))
	totalstr := LongStr + HuStr + HeStr + str
	game.Table.WriteLogs(0, totalstr)

	game.Table.Broadcast(int32(longhu.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

func (game *Game) getResult() {
	game.Status = longhu.GameStatus_ShowPoker
	test := game.Table.GetRoomProb()
	if test == 0 {
		game.sysCheat = "获取作弊率为0 "
		test = 1000
	} else {
		game.sysCheat = ""
	}
	game.CheatValue = int(test)
	eat := 0
	out := 0
	r := rand.Intn(10000)
	v := config.LongHuConfig.GetCheatValue(game.CheatValue)

	if r < v {
		eat = 1
	} else {
		out = 1
	}

	log.Debugf("吃分：%v，吐分：%v", eat, out)
	if !game.HasTest {
		for {
			game.DealPoker()
			game.ComparePoker()

			TotalMoney := game.TotalUserBet[LONG] + game.TotalUserBet[HU] + game.TotalUserBet[HE]
			var PayMoney int64
			if game.Win == HU {
				PayMoney = game.TotalUserBet[HU]
			} else if game.Win == LONG {
				PayMoney = game.TotalUserBet[LONG]
			}

			OutMoney := TotalMoney - PayMoney*2
			if out == 1 {
				OutMoney -= game.TotalUserBet[HE]
				if OutMoney <= 0 {
					//吐分
					break
				}
			}

			log.Debugf("结果：%v %v", game.Win, OutMoney)
			if game.Win == HE {
				OutMoney = TotalMoney - game.TotalUserBet[HU] - game.TotalUserBet[LONG] - (HEODDS+1)*game.TotalUserBet[HE]
			}

			if eat == 1 && OutMoney >= 0 {
				//吃分
				break
			} else if eat == 0 && out == 0 {
				//不控制
				break
			}
		}
	} else {
		game.ComparePoker()
		game.HasTest = false
	}

	log.Debugf("获取到的结果")
	game.PokerMsg.LongPoker[0] = game.LongCards
	game.PokerMsg.HuPoker[0] = game.HuCards
	game.PokerMsg.Win = int32(game.Win)
	game.Table.Broadcast(int32(longhu.SendToClientMessageType_PokerInfo), game.PokerMsg)
	game.SendStatusMsg(config.LongHuConfig.Taketimes.ShowPoker)
	game.Table.AddTimer(int64(config.LongHuConfig.Taketimes.ShowPoker), game.Settle)
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
		game.OnlineUserList = append(game.OnlineUserList, u)
		game.CountUserList = append(game.CountUserList, u)
		u.ResetUserData()
	}

	return u
}

//发送场景消息
func (game *Game) SendSceneMsg(u player.PlayerInterface) {
	msg := new(longhu.SceneMessage)
	//bigwinner := game.SenceSeat.GetBigWinner()
	//master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(longhu.SeatUser)
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
		//if v.User.Icon == 0 {
		//	su.IsMaster = false
		//	su.IsMillionaire = false
		//	su.IsBigWinner = false
		//} else if v.User.Icon == Master {
		//	su.IsMaster = true
		//	su.IsMillionaire = false
		//	su.IsBigWinner = false
		//} else if v.User.Icon == Millionaire {
		//	su.IsMaster = false
		//	su.IsMillionaire = true
		//	su.IsBigWinner = false
		//} else if v.User.Icon == bigWinner {
		//	su.IsMaster = false
		//	su.IsMillionaire = false
		//	su.IsBigWinner = true
		//}

		msg.UserData = append(msg.UserData, su)
	}
	if u != nil {
		u.SendMsg(int32(longhu.SendToClientMessageType_SceneID), msg)
	} else {
		game.Table.Broadcast(int32(longhu.SendToClientMessageType_SceneID), msg)
	}

	//u.SendMsg(int32(longhu.SendToClientMessageType_SceneID), msg)
}

func (game *Game) SendUserBet(u *model.User) {
	msg := new(longhu.SceneBetInfo)
	msg.Long = game.BetTotal[LONG]
	msg.Hu = game.BetTotal[HU]
	msg.He = game.BetTotal[HE]
	msg.UserBetLong = u.BetLong
	msg.UserBetHu = u.BetHu
	msg.UserBetHe = u.BetHe
	msg.UserBetTotal = u.TotalBet
	msg.MasterBetType = game.LastMasterBetType
	u.User.SendMsg(int32(longhu.SendToClientMessageType_BetInfo), msg)
}

func (game *Game) SendTrend(u player.PlayerInterface) {
	log.Tracef("发送走势图")
	msg := new(longhu.Trend)
	msg.Win = append(msg.Win, game.WinTrend...)
	u.SendMsg(int32(longhu.SendToClientMessageType_TrendInfo), msg)
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
	msg := new(longhu.UserList)
	for _, u := range game.CountUserList {
		userinfo := new(longhu.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.RetWinMoney
		userinfo.Head = u.User.GetHead()
		userinfo.Icon = u.Icon
		msg.UserList = append(msg.UserList, userinfo)
	}
	log.Tracef("SendUserListInfo %v", msg)
	user.SendMsg(int32(longhu.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) ResetData() {
	for i := 0; i < 3; i++ {
		game.TotalUserBet[i] = 0
		game.BetTotal[i] = 0
	}
}

func (game *Game) OnUserStanUp(user player.PlayerInterface) {
	game.SenceSeat.UserStandUp(user)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		u.SceneChairId = 0
	}
	for _, v := range game.AllUserList {
		game.SendSceneMsg(v.User)
	}
	//game.RobotTimerJob,_=game.Table.AddTimer(2000, func() {
	//	game.RandSelectUserSitDownChair(user)
	//})
	//game.RandSelectUserSitDownChair(user)
}

func (game *Game) OnTest(buffer []byte) {
	game.HasTest = true
	temp := &longhu.TempCardMsg{}
	proto.Unmarshal(buffer, temp)
	game.LongCards = temp.LongPoker[0]
	game.HuCards = temp.HuPoker[0]
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
					us := &longhu.UserSitDown{}
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
	msg := new(longhu.SceneMessage)
	bigwinner := game.SenceSeat.GetBigWinner()
	master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(longhu.SeatUser)
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

	game.Table.Broadcast(int32(longhu.SendToClientMessageType_SceneID), msg)
}

func (game *Game) GameStart() {
	if game.Status == 0 {
		game.Start()
		game.Table.AddTimerRepeat(1000, 0, game.SendRoomInfo)
	}
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
	//game.Rule.BetLimit[LONG] = int64(BetBase) * 5000
	//game.Rule.BetLimit[HU] = int64(BetBase) * 5000
	//game.Rule.BetLimit[HE] = int64(BetBase) * 2000
	game.Rule.SitDownLimit = BetBase * 100
	level := game.Table.GetLevel()
	game.Rule.RobotMinGold = config.LongHuConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.LongHuConfig.Robotgold[level-1][1]
	game.Rule.SingleUserAllSpaceLimit = config.LongHuConfig.Singleuserallspacelimit5times[level-1]
	game.Rule.AllSpaceLimit = config.LongHuConfig.Allspacelimit5times[level-1]
	for i := 0; i < 3; i++ {
		//log.Traceln(config.BRNNConfig.Singleusersinglespacelimit5times[level-1][3],config.BRNNConfig.Allusersinglespacelimit5times[level-1][i])
		game.Rule.SingleUserSingleSpaceLimit[i] = config.LongHuConfig.Singleusersinglespacelimit5times[level-1][i]
		game.Rule.AllUserSingleSpaceLimit[i] = config.LongHuConfig.Allusersinglespacelimit5times[level-1][i]
	}
	for i := 0; i < 5; i++ {
		game.Rule.BetList = append(game.Rule.BetList, config.LongHuConfig.Chips5times[level-1][i])
	}
	game.Rule.UserBetLimit = game.Rule.SingleUserAllSpaceLimit
	//log.Traceln(game.Rule.BetList, ":", game.Rule.SingleUserSingleSpaceLimit, ":", game.Rule.SingleUserAllSpaceLimit, ":", game.Rule.AllUserSingleSpaceLimit, ":", game.Rule.AllSpaceLimit)
}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(longhu.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}

	msg.UserBetLimit = int32(game.Rule.UserBetLimit)
	msg.BetMinLimit = game.Rule.BetMinLimit
	u.SendMsg(int32(longhu.SendToClientMessageType_RoomRolesInfo), msg)
}

func (game *Game) SendRoomInfo() {
	if game.Status == 0 {
		return
	}
	msg := new(longhu.RoomSenceInfoMsg)
	msg.TrendList = new(longhu.Trend)
	msg.TrendList.Win = append(msg.TrendList.Win, game.WinTrend...)
	msg.GameStatus = new(longhu.StatusMessage)
	msg.GameStatus.Status = int32(game.Status)
	msg.GameStatus.StatusTime = int32(game.TimerJob.GetTimeDifference())
	msg.RoomID = int64(game.Table.GetRoomID())
	msg.BaseBet = int64(game.Rule.BetList[0])
	msg.UserLimit = game.Rule.UserBetLimit
	//发送给框架
	//b, _ := proto.Marshal(msg)
	//game.Table.BroadcastAll(int32(rbwar.SendToClientMessageType_RoomSenceInfo), b)
	game.Table.SendToHall(int32(longhu.SendToClientMessageType_RoomSenceInfo), msg)
}

func (game *Game) ResetTable() {
	game.Status = 0
	game.WinTrend = make([]int32, 0)
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
	for k, v := range game.OnlineUserList {
		if user == v {
			game.OnlineUserList = append(game.OnlineUserList[:k], game.OnlineUserList[k+1:]...)
			break
		}
	}
	for k, v := range game.CountUserList {
		if user == v {
			game.CountUserList = append(game.CountUserList[:k], game.CountUserList[k+1:]...)
			break
		}
	}
}

//进入房间顺序选择6个用户坐下
func (game *Game) RandSelectUserSitDownChair() {
	checkChairNum := game.SenceSeat.GetSceneChairId()
	if checkChairNum == 0 {
		return
	} else {
		for _, u := range game.OnlineUserList {
			//获取空位置，如无则返回
			ChairId := game.SenceSeat.GetSceneChairId()
			if ChairId == 0 {
				game.SendSceneMsg(nil)
				return
			}
			//if user==u.User{
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

//神算子1 胜率最高的 大赢家> 大富豪> 神算子
func (game *Game) getMaster() {
	mast := int(0)
	time1 := int64(0)
	time2 := int64(0)
	win1 := int64(0)
	win2 := int64(0)
	var Uindex = 0
	index := len(game.CountUserList)
	if index >= 6 {
		index = 6
	}
	for i := 0; i < index; i++ {
		v := game.CountUserList[i]
		if v.Icon == bigWinner || v.Icon == Millionaire {
			continue
		}
		mastTmp := int(math.Floor(float64(v.RetWin) / (float64(len(v.RetCount))) * 100))
		if mast < mastTmp {
			mast = mastTmp
			time1 = v.Time
			win1 = v.RetWinMoney
			//id = v.SceneChairId
			Uindex = i
		} else if mast == mastTmp {
			mast = mastTmp
			time2 = v.Time
			win2 = v.RetWinMoney
			//id = v.SceneChairId
			if win1 < win2 {
				Uindex = i
			} else if win1 == win2 {
				if time2 < time1 {
					Uindex = i
				} else {
					continue
				}
			} else {
				continue
			}
		}
	}
	game.CountUserList[Uindex].Icon = Master
}

//大富豪2 近20局赢钱最多的 大赢家> 大富豪> 神算子
func (game *Game) getBigWinner() {
	money := int64(0)
	time1 := int64(0)
	time2 := int64(0)
	mast1 := 0
	mast2 := 0
	var Uindex = 0
	index := len(game.CountUserList)
	if index >= 6 {
		index = 6
	}
	for i := 0; i < index; i++ {
		v := game.CountUserList[i]
		if v.Icon == Millionaire {
			continue
		}
		if money < v.RetWinMoney {
			money = v.RetWinMoney
			time1 = v.Time
			mast1 = int(math.Floor(float64(v.RetWin) / (float64(len(v.RetCount))) * 100))
			Uindex = i
		} else if money == v.RetWinMoney {
			money = v.RetWinMoney
			time2 = v.Time
			mast2 = int(math.Floor(float64(v.RetWin) / (float64(len(v.RetCount))) * 100))
			if mast1 < mast2 {
				Uindex = i
			} else if mast1 == mast2 {
				if time2 < time1 {
					Uindex = i
				} else {
					continue
				}
			} else {
				continue
			}
		}
	}
	game.CountUserList[Uindex].Icon = bigWinner
	//return id
}

//大赢家3 上一局赢钱最多的 优先级 大赢家> 大富豪> 神算子
func (game *Game) getMillionaire() {

	//return id
}

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

	// 大富豪
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
	//神算子
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
