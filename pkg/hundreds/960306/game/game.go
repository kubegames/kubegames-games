package game

import (
	"common/log"
	"common/score"
	"fmt"
	"game_poker/BRTB/config"
	"game_poker/BRTB/model"
	"math/rand"
	"sort"
	"time"

	"game_frame_v2/game/clock"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

const (
	Big       = 0
	Small     = 1
	Single    = 2
	Double    = 3
	Four      = 4
	Five      = 5
	Six       = 6
	Seven     = 7
	Eight     = 8
	Nine      = 9
	Ten       = 10
	Eleven    = 11
	Twelve    = 12
	Thirteen  = 13
	Fourteen  = 14
	Fifteen   = 15
	Sixteen   = 16
	Seventeen = 17
	Wei       = 18
	WeiOne    = 19
	WeiTwo    = 20
	WeiThree  = 21
	WeiFour   = 22
	WeiFive   = 23
	WeiSix    = 24
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
	Status              BRTB.GameStatus       // 房间状态1 表示
	LastWinIsRedOrBlack int                   // 最近一次开龙还是虎
	Dices               [3]int32              // 骰子3个
	BetTotal            [25]int64             //下注统计
	TotalUserBet        [25]int64             //下注统计
	SenceSeat           model.SceneInfo       //下注的玩家列表
	TimerJob            *clock.Job            //job
	RobotTimerJob       *clock.Job            //机器人job
	LastMasterBetType   int32                 //最近一次神算子下注的类型
	WinTrend            []*BRTB.OneTrend      //赢的走势
	CountUserList       []*model.User         //统计后的玩家列表
	Rule                config.RoomRules      //房间规则信息
	gp                  model.GamePoker       //牌
	//LastCardCount       int32                 //剩余牌的张数
	testDices  []int32        //测试骰子数
	CheatValue int            //作弊值
	PokerMsg   *BRTB.PokerMsg //骰子消息
	sysCheat   string         //
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.SenceSeat.Init()
	game.PokerMsg = new(BRTB.PokerMsg)
	game.PokerMsg.Dices = make([]int32, 3)
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
	delete(game.AllUserList, user.GetId())
	//删除用户列表用户
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u := game.getUser(user)
	if u.TotalBet != 0 {
		msg := new(BRTB.ExitFail)
		msg.FailReason = "游戏中不能退出！"
		user.SendMsg(int32(BRTB.SendToClientMessageType_ExitRet), msg)
		return false
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}
	delete(game.AllUserList, user.GetId())
	game.DeleteExitUserFromOnlineUserListSlice(u)
	//game.RandSelectUserSitDownChair()
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	if int32(BRTB.ReceiveMessageType_BetID) == subCmd {
		//客户端下注消息
		game.Bet(buffer, user)
	} else if int32(BRTB.ReceiveMessageType_SitDown) == subCmd {
		//game.UserSitDown(buffer, user)
	} else if int32(BRTB.ReceiveMessageType_GetTrend) == subCmd {
		game.SendTrend(user)
	} else if int32(BRTB.ReceiveMessageType_GetUserListInfo) == subCmd {
		game.SendUserListInfo(user)
	} else if int32(BRTB.ReceiveMessageType_StandUp) == subCmd {
		//game.OnUserStanUp(user)
	} else if int32(BRTB.ReceiveMessageType_tempCard) == subCmd {
		game.testCard(buffer)
	} else if int32(BRTB.ReceiveMessageType_BetReptID) == subCmd {
		game.repeatBet(buffer, user)
	}
}

func (game *Game) SendScene(user player.PlayerInterface) bool {
	game.GetRoomconfig()
	u := game.getUser(user)
	if user.IsRobot() {
		robot := new(Robot)
		robotUser := user.BindRobot(robot)
		robot.Init(robotUser, game)
		return true
	}

	game.SendRuleInfo(user)
	//game.SendSceneMsg(user)
	game.SendUserBet(u)

	if game.Status >= BRTB.GameStatus_ShowPoker {
		user.SendMsg(int32(BRTB.SendToClientMessageType_PokerInfo), game.PokerMsg)
		if game.Status == BRTB.GameStatus_SettleStatus {
			if u.SettleMsg != nil {
				user.SendMsg(int32(BRTB.SendToClientMessageType_UserComeBack), u.SettleMsg)
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
	//game.SelectUserListInfoBefore6SitDownChair()
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

	//game.BroadCastPokerCount()
	//if game.RobotTimerJob == nil {
	//	//r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
	//	//game.RobotTimerJob, _ = game.Table.AddTimer(time.Duration(r), game.RobotSitDown)
	//	//r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
	//	//game.Table.AddTimer(time.Duration(r1), game.RobotStandUp)
	//}
	game.Status = BRTB.GameStatus_StartMovie
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.Startmove), game.StartBet)

	//开始动画消息
	game.SendStatusMsg(config.BRTBConfig.Taketimes.Startmove)
}

func (game *Game) StartBet() {
	game.ResetData()
	game.Status = BRTB.GameStatus_BetStatus
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.Startbet), game.EndBet)

	//发送开始下注消息
	game.SendStatusMsg(config.BRTBConfig.Taketimes.Startbet)
}

func (game *Game) EndBet() {
	game.Status = BRTB.GameStatus_EndBetMovie
	//game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.Endmove), game.Settle)
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.Endmove), game.getResult)
	//发送开始下注消息
	game.SendStatusMsg(config.BRTBConfig.Taketimes.Endmove)
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

	game.Status = BRTB.GameStatus_SettleStatus
	game.sendSettleMsg()

	endtime := config.BRTBConfig.Taketimes.Endpay
	//结算完成游戏开始
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(endtime), game.StartGame)
	//if game.gp.GetCardsCount()-int(game.LastCardCount) >= 6 {
	//	game.TimerJob, _ = game.Table.AddTimer(time.Duration(endtime), game.Start)
	//} else {
	//	game.TimerJob, _ = game.Table.AddTimer(time.Duration(endtime), game.InitPoker)
	//}

	//game.checkUserBet()
	//发送开始下注消息
	game.SendStatusMsg(endtime)
	//
	game.Table.EndGame()
}

//下注消息通知前端 msg 游戏阶段 下注时间
func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(BRTB.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	game.Table.Broadcast(int32(BRTB.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(BRTB.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(BRTB.SendToClientMessageType_Status), msg)
}

//客户端下注
func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != BRTB.GameStatus_BetStatus {
		return
	}
	//用户下注
	BetPb := &BRTB.Bet{}
	proto.Unmarshal(buffer, BetPb)

	u := game.getUser(user)
	if u.Bet(BetPb, game.BetTotal) {
		game.BetTotal[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex])

		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType] += int64(game.Rule.BetList[BetPb.BetIndex])
		}

		u.User.SetScore(game.Table.GetGameNum(), -int64(game.Rule.BetList[BetPb.BetIndex]), game.Table.GetRoomRate())
		//if game.SenceSeat.UserBet(u) {
		//	game.SendMiPaiSeat()
		//}
	}

	//if game.SenceSeat.GetMaster() == u.SceneChairId {
	//	game.LastMasterBetType = BetPb.BetType
	//}
	//神算子下注区域
	u1, ok := game.SenceSeat.UserSeat[user.GetId()]
	if ok {
		if u1.User.Icon == 1 {
			game.LastMasterBetType = BetPb.BetType
		}

	}
}

func (game *Game) UserSitDown(buffer []byte, user player.PlayerInterface) {
	us := &BRTB.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetId()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo)) {
			u.SceneChairId = int(us.ChairNo)
			game.SendSceneMsg(nil)
		}
	}
}

//摇骰子阶段
func (game *Game) InitPoker() {
	//game.BroadCastPokerCount()
	game.Status = BRTB.GameStatus_ShakeDice
	//game.gp.InitPoker()
	//game.gp.ShuffleCards()
	//game.LastCardCount = rand.Int31n(208) + 1

	game.WinTrend = make([]*BRTB.OneTrend, 0)
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.ShakeDice), game.Start)
	game.SendStatusMsg(config.BRTBConfig.Taketimes.ShakeDice)
}

//开始游戏动画
func (game *Game) StartGame() {
	game.Status = BRTB.GameStatus_StartGame
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.ShakeDice), game.ShakeDice)
	game.SendStatusMsg(config.BRTBConfig.Taketimes.StartGame)
}

//摇骰子动画
func (game *Game) ShakeDice() {
	game.Status = BRTB.GameStatus_ShakeDice
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.ShakeDice), game.Start)
	game.SendStatusMsg(config.BRTBConfig.Taketimes.ShakeDice)
}

// 骰子获取
func (game *Game) Dicing() {
	temp := model.GetDices()
	for i := 0; i <= 2; i++ {
		game.Dices[i] = temp[i]
	}
}

// 比牌
func (game *Game) ComparePoker() {
	model.GetDiceResult(game.Dices)
}

//检查用户是否被踢掉
func (game *Game) checkUserBet() {
	for k, u := range game.AllUserList {
		if u.NoBetCount >= (config.BRTBConfig.Unplacebetnum+1) ||
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
	//出将结果 点数 大小豹子，单双豹子，
	Count, BswType, SDwType := model.GetDiceResult(game.Dices)

	wt := &BRTB.OneTrend{}
	wt.Count = Count
	wt.SingleDouble = SDwType
	wt.BigSmall = BswType
	for i := 0; i < 3; i++ {
		wt.Dices = append(wt.Dices, game.Dices[i])
	}

	winlen := len(game.WinTrend)

	game.WinTrend = append(game.WinTrend, wt)
	if winlen > 100 {
		game.WinTrend = append(game.WinTrend[:(winlen-100-1)], game.WinTrend[(winlen-100):]...)
	}

	game.CountUserList = make([]*model.User, 0)
	SceneSettleMsg := new(BRTB.SceneUserSettle)
	//ZhuangMiUser, ok1 := game.SenceSeat.UserSeat[game.SenceSeat.BetZhuangMaxID]
	//XianMiUser, ok2 := game.SenceSeat.UserSeat[game.SenceSeat.BetXianMaxID]
	var BetCount [25]int64     //各区域下注人数统计0-25 大小单双4 5 6...围6
	var SpaceLogStr [25]string //各区域下注日志
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [25]int64
	//第一条区域日志加上作弊率
	//区域日志消息
	for i := 0; i <= 24; i++ {
		SpaceLogStr[i] = fmt.Sprintf("%s区域：总：%v 机器人：%v 真人：%v ",
			GetStringNameById(i), score.GetScoreStr(game.BetTotal[i]), score.GetScoreStr(game.BetTotal[i]-game.TotalUserBet[i]), score.GetScoreStr(game.TotalUserBet[i]))
	}

	for _, u := range game.AllUserList {
		u.NoBetCount++
		if !u.User.IsRobot() {
			if u.NoBetCount >= (config.BRTBConfig.Unplacebetnum + 1) {
				//发送踢掉用户
				msg := new(BRTB.KickOutUserMsg)
				msg.KickOutReason = "由于您5局未下注，已被踢出房间！"
				u.User.SendMsg(int32(BRTB.SendToClientMessageType_KickOutUser), msg)
			}
		}

		SceneUserInfo := new(BRTB.SceneUserInfo)

		msg := new(BRTB.SettleMsg)
		msg.Win = make([]bool, 25)
		if BswType == Wei {
			msg.Win[Wei] = true
			msg.Win[Wei+game.Dices[0]] = true
		} else {
			msg.Win[BswType] = true
			msg.Win[SDwType] = true
			msg.Win[Count] = true
		}
		var win int64
		var totalTax int64 //总税收
		var totalwin int64
		//var Award int64  //总产出
		var ResBet int64 //结算后玩家的下注
		var Chip int64   //

		/*
			用户结算信息
		*/
		for i := 0; i < 25; i++ {
			//所有区域总的下注
			msg.BetArea = append(msg.BetArea, game.BetTotal[i])
			//个人所有区域下注
			msg.UserBet = append(msg.UserBet, u.BetArea[i])
			//msg.Win = append(msg.Win, tmpTrend.t[i].Win)
			//			msg.Type = append(msg.Type, tmpTrend.t[i].Type)
			if u.TotalBet <= 0 {
				continue
			}
			//如果是围结果不出大小单双。
			if BswType == Wei {
				if i == Wei {
					win = u.BetArea[i] * 25
					totalwin += win
					msg.UserWin = append(msg.UserWin, win)
					var tax int64
					tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i]*24, game.Table.GetRoomRate())
					capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], 0)
					tax += capital
					totalTax += win - tax
					msg.TotalWin += tax
					SystemWin[i] += tax
				} else if i == Wei+int(game.Dices[0]) {
					win = u.BetArea[i] * 151
					totalwin += win
					msg.UserWin = append(msg.UserWin, win)
					var tax int64
					tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i]*150, game.Table.GetRoomRate())
					capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], 0)
					tax += capital
					totalTax += win - tax
					msg.TotalWin += tax
					SystemWin[i] += tax
				} else {
					win = -u.BetArea[i]
					ResBet += u.BetArea[i]
					Chip += u.BetArea[i] * 2
					totalwin += win
					msg.TotalWin += win
					msg.UserWin = append(msg.UserWin, win)
					continue
				}
			} else {
				//结果不是围 结算大小单双围
				if i == int(BswType) {
					//大小
					win = u.BetArea[i] * 2
					totalwin += win
					msg.UserWin = append(msg.UserWin, win)
					var tax int64
					tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], game.Table.GetRoomRate())
					capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], 0)
					tax += capital
					totalTax += win - tax
					msg.TotalWin += tax
					SystemWin[i] += tax
				} else if i == int(SDwType) {
					//单双
					win = u.BetArea[i] * 2
					totalwin += win
					msg.UserWin = append(msg.UserWin, win)
					var tax int64
					tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], game.Table.GetRoomRate())
					capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], 0)
					tax += capital
					totalTax += win - tax
					msg.TotalWin += tax
					SystemWin[i] += tax
				} else if i == int(Count) {
					//点数
					win = u.BetArea[i] * (game.GetOdds(Count) + 1)
					totalwin += win
					msg.UserWin = append(msg.UserWin, win)
					var tax int64
					tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i]*game.GetOdds(Count), game.Table.GetRoomRate())
					capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetArea[i], 0)
					tax += capital
					totalTax += win - tax
					msg.TotalWin += tax
					SystemWin[i] += tax
				} else {
					win = -u.BetArea[i]
					ResBet += u.BetArea[i]
					Chip += u.BetArea[i] * 2
					totalwin += win
					//msg.TotalWin += win
					msg.UserWin = append(msg.UserWin, win)
					continue
				}
			}

		}
		//
		SceneUserInfo.BetArea = append(SceneUserInfo.BetArea, msg.BetArea...)
		SceneUserInfo.UserBet = append(SceneUserInfo.UserBet, msg.UserBet...)

		SceneUserInfo.TotalWin = msg.TotalWin
		SceneUserInfo.UserID = int64(u.User.GetId())
		SceneUserInfo.SceneSeatID = int32(u.SceneChairId)
		//统计玩家信息
		if (win) > u.TotalBet {
			u.UserCount(true, msg.TotalWin)
		} else {
			u.UserCount(false, 0)
		}

		//写入数据库统计信息
		if MaxWinGold < u.User.GetScore()-u.CruenSorce {
			MaxWinGold = u.User.GetScore() - u.CruenSorce
			MaxWinUserID = u.User.GetID()
		}
		//用户输赢日志
		if !u.User.IsRobot() {
			var temp string
			var temp1 string
			if u.TotalBet != 0 {
				temp += fmt.Sprintf("用户ID：%v，开始金币：%v，投注额:", u.User.GetID(), score.GetScoreStr(u.CruenSorce))
				temp1 += fmt.Sprintf(" 输赢：")
			}
			for i := 0; i <= 24; i++ {
				sapceName := GetStringNameById(i)
				if u.BetArea[i] != 0 {
					twin := msg.UserWin[i] - u.BetArea[i]
					if msg.UserWin[i] < 0 {
						twin = msg.UserWin[i]
					}
					temp += fmt.Sprintf("%s：%v;", sapceName,
						score.GetScoreStr(u.BetArea[i]))
					temp1 += fmt.Sprintf("%s：%v;", sapceName,
						score.GetScoreStr(twin))
					BetCount[i]++
				}

			}
			temp1 += fmt.Sprintf(" 总输赢：%v，用户剩余金额：%v \r\n", score.GetScoreStr(u.User.GetScore()-u.CruenSorce), score.GetScoreStr(u.User.GetScore()))
			temp += temp1
			game.Table.WriteLogs(u.User.GetId(), temp)
		}
		//跑马灯"玩家中豹子触发"
		if SDwType == Wei {
			game.PaoMaDeng(msg.TotalWin-u.TotalBet, u.User)
		}
		game.CountUser(u)
		msg.UserScore = u.User.GetScore()

		u.User.SendMsg(int32(BRTB.SendToClientMessageType_Settle), msg)

		if u.TotalBet > 0 && !u.User.IsRobot() {
			u.SettleMsg = msg
			chip := u.BetArea[Big] - u.BetArea[Small]

			if chip < 0 {
				chip = -chip
			}
			if u.BetArea[Single]-u.BetArea[Double] < 0 {
				chip += -(u.BetArea[Big] - u.BetArea[Small])
			}
			for m := Four; m <= WeiSix; m++ {
				chip += u.BetArea[m]
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
	sysallwin := int64(0) //系统总输赢
	for i := 0; i <= 24; i++ {
		SpaceLogStr[i] += fmt.Sprintf("真人数量:%v 输赢：%v;\r\n", BetCount[i], score.GetScoreStr(game.BetTotal[i]-SystemWin[i]))
		sysallwin += game.BetTotal[i] - SystemWin[i]
	}

	str := fmt.Sprintf("%v作弊率：%v \r\n开局结果:骰子:%v\r\n", game.sysCheat, game.CheatValue,
		game.Dices)
	if SDwType == Wei {
		str += fmt.Sprintf("任意围；指定围（%v）, 系统输赢额度：%v\r\n",
			game.Dices[0], score.GetScoreStr(sysallwin))
	} else {
		str += fmt.Sprintf("点数：%v 大小：%s 单双：%s 系统输赢额度：%v\r\n",
			Count, GetStringNameById(int(BswType)), GetStringNameById(int(SDwType)), score.GetScoreStr(sysallwin))
	}
	str += fmt.Sprintf("最高获利用户ID：%v 获得：%v\r\n",
		MaxWinUserID, score.GetScoreStr(MaxWinGold))
	var totalstr string
	for i := 0; i <= 24; i++ {
		totalstr += SpaceLogStr[i]
	}
	totalstr += str
	game.Table.WriteLogs(0, totalstr)

	game.SenceSeat.BetXianMaxID = 0
	game.SenceSeat.BetZhuangMaxID = 0
	game.Table.Broadcast(int32(BRTB.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

func (game *Game) getResult() {
	game.Status = BRTB.GameStatus_ShowPoker
	test, _ := game.Table.GetRoomProb()
	if test == 0 {
		game.sysCheat = "获取作弊率为0 "
		test = 1000
	} else {
		game.sysCheat = ""
	}
	log.Debugf("使用作弊率为：%v", test)
	game.CheatValue = int(test)
	//shakePolicy := config.BRTBConfig.PolicyTree.Find(test)
	//back := shakePolicy.Back.Rand()
	eat := 0
	out := 0
	tempi := 100 //循环次数如果系统为输循环10次出结果,为赢循环100次
	r := rand.Intn(10000)
	//根据作弊率是否是吃分还是吐分
	v := config.BRTBConfig.GetCheatValue(game.CheatValue)
	if v != 0 {
		if r < v {
			eat = 1
		} else {
			out = 1
			tempi = 10
		}
	}
	//fmt.Println("循环次数",tempi,"吃分",eat,"吐分",out)
	Count := int32(0)
	BSWType := int32(0)
	SDWType := int32(0)
	if len(game.testDices) > 0 {
		//测试
		for i := 0; i < len(game.testDices); i++ {
			game.Dices[i] = game.testDices[i]
		}
		game.testDices = make([]int32, 0)
		game.ComparePoker()
		Count, BSWType, SDWType = model.GetDiceResult(game.Dices)
	} else {
		TotalMoney := int64(0) //总压注金额
		for i := 0; i <= 24; i++ {
			TotalMoney += game.TotalUserBet[i]
		}

		//循环出结果
		bAddCards := false
		backrate := 0
		for i := 0; i < tempi; i++ {
			//如果不符合要求，骰子重新出
			if bAddCards {
				game.Dicing()
			}

			game.Dicing()
			bAddCards = true
			//点数 大小 单双 围（指定围下注区域=围n+4）
			Count, BSWType, SDWType = model.GetDiceResult(game.Dices)
			//game.ComparePoker()
			//log.Debugf("结果为：%v", game.ZhuangCards, game.XianCards)
			//第一把不出和
			//if len(game.WinTrend) == 0 && game.Win == HE {
			//	i--
			//	continue
			//}
			if eat == 0 && out == 0 {
				//不控制
				break
			}
			//赔付的钱
			//赔付的钱
			var PayMoney int64
			if BSWType == Wei {
				PayMoney = game.TotalUserBet[BSWType] * 25
				PayMoney += game.TotalUserBet[game.Dices[0]+Wei] * 151
			} else {
				//大小单双赔付
				PayMoney = game.TotalUserBet[BSWType] * 2
				PayMoney += game.TotalUserBet[SDWType] * 2
				//计算点数赔付
				PayMoney += game.TotalUserBet[Count] * (game.GetOdds(Count) + 1)
			}
			if TotalMoney == 0 {
				break
			} else {
				backrate = int(PayMoney / TotalMoney * 100)
			}

			//总钱1赔付的钱
			OutMoney := TotalMoney - PayMoney
			if eat == 1 && OutMoney >= 0 {
				//吃分
				break
			} else if out == 1 && OutMoney <= 0 {
				//吐分
				if game.CheatValue == 1000 || game.CheatValue == 2000 || game.CheatValue == 3000 {
					if backrate < 500 {
						break
					}
				} else {
					break
				}

			}

		}
	}

	/*
	  结算信息骰子点数类型
	*/
	for i := 0; i <= 2; i++ {
		//骰子信息
		game.PokerMsg.Dices[i] = game.Dices[i]
	}
	//骰子点数
	game.PokerMsg.Count = Count
	//骰子类型
	game.PokerMsg.BigSmall = BSWType
	game.PokerMsg.SingleDouble = SDWType

	endtime := config.BRTBConfig.Taketimes.Cardmove

	game.Table.Broadcast(int32(BRTB.SendToClientMessageType_PokerInfo), game.PokerMsg)
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(endtime), game.Settle)
	game.SendStatusMsg(endtime)
}

func (game *Game) getUser(user player.PlayerInterface) *model.User {
	u, ok := game.AllUserList[user.GetId()]
	if !ok {
		u = new(model.User)
		game.AllUserList[user.GetId()] = u
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
	msg := new(BRTB.SceneMessage)
	//bigwinner := game.SenceSeat.GetBigWinner()
	//master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(BRTB.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetId())
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

	if u != nil {
		u.SendMsg(int32(BRTB.SendToClientMessageType_SceneID), msg)
	} else {
		game.Table.Broadcast(int32(BRTB.SendToClientMessageType_SceneID), msg)
	}
}

func (game *Game) SendUserBet(u *model.User) {
	msg := new(BRTB.SceneBetInfo)
	for i := 0; i < 25; i++ {
		msg.BetArea = append(msg.BetArea, game.BetTotal[i])
		msg.UserBet = append(msg.UserBet, u.BetArea[i])
	}

	msg.UserBetTotal = u.TotalBet
	msg.MasterBetType = game.LastMasterBetType
	u.User.SendMsg(int32(BRTB.SendToClientMessageType_BetInfo), msg)
}

func (game *Game) SendTrend(u player.PlayerInterface) {
	log.Tracef("发送走势图")
	msg := new(BRTB.Trend)
	msg.Info = append(msg.Info, game.WinTrend...)

	u.SendMsg(int32(BRTB.SendToClientMessageType_TrendInfo), msg)
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
	msg := new(BRTB.UserList)
	for _, u := range game.CountUserList {
		userinfo := new(BRTB.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.RetWinMoney
		userinfo.Head = u.User.GetHead()
		userinfo.Icon = u.Icon
		msg.UserList = append(msg.UserList, userinfo)
	}

	user.SendMsg(int32(BRTB.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) ResetData() {
	for i := 0; i < 25; i++ {
		game.TotalUserBet[i] = 0
		game.BetTotal[i] = 0
	}
}

func (game *Game) OnUserStanUp(user player.PlayerInterface) {
	bSendMiPai := false
	if user.GetId() == game.SenceSeat.BetZhuangMaxID || user.GetId() == game.SenceSeat.BetXianMaxID {
		bSendMiPai = true
	}

	game.SenceSeat.UserStandUp(user)
	u, ok := game.AllUserList[user.GetId()]
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
	game.RobotTimerJob, _ = game.Table.AddTimer(time.Duration(r), game.RobotSitDown)

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
					us := &BRTB.UserSitDown{}
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
	game.Table.AddTimer(time.Duration(r), game.RobotStandUp)

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
	msg := new(BRTB.SceneMessage)
	bigwinner := game.SenceSeat.GetBigWinner()
	master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(BRTB.SeatUser)
		su.Head = v.User.User.GetHead()
		su.Nick = v.User.User.GetNike()
		su.Score = v.User.User.GetScore()
		su.SeatId = int32(v.SeatNo)
		su.UserID = int64(v.User.User.GetId())
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

	game.Table.Broadcast(int32(BRTB.SendToClientMessageType_SceneID), msg)
}

func (game *Game) GameStart(user player.PlayerInterface) bool {
	if game.Status == 0 {
		game.Status = BRTB.GameStatus_StartGame
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRTBConfig.Taketimes.StartGame), game.InitPoker)
		game.SendStatusMsg(config.BRTBConfig.Taketimes.StartGame)
		//game.InitPoker()
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
	game.Rule.RobotMinGold = config.BRTBConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.BRTBConfig.Robotgold[level-1][1]
	game.Rule.SingleUserAllSpaceLimit = config.BRTBConfig.Singleuserallspacelimit5times[level-1]
	game.Rule.AllSpaceLimit = config.BRTBConfig.Allspacelimit5times[level-1]
	for i := 0; i < 25; i++ {
		//fmt.Println(config.BRNNConfig.Singleusersinglespacelimit5times[level-1][3],config.BRNNConfig.Allusersinglespacelimit5times[level-1][i])
		game.Rule.SingleUserSingleSpaceLimit[i] = config.BRTBConfig.Singleusersinglespacelimit5times[level-1][i]
		game.Rule.AllUserSingleSpaceLimit[i] = config.BRTBConfig.Allusersinglespacelimit5times[level-1][i]
	}
	for i := 0; i < 5; i++ {
		game.Rule.BetList = append(game.Rule.BetList, config.BRTBConfig.Chips5times[level-1][i])

	}
	//fmt.Println(level, RConfig.Line)
	game.Rule.RobotLine[0] = RConfig.Line[level-1][0]
	game.Rule.RobotLine[1] = RConfig.Line[level-1][1]
	game.Rule.RobotLine[2] = RConfig.Line[level-1][2]
	game.Rule.RobotLine[3] = RConfig.Line[level-1][3]
	game.Rule.UserBetLimit = game.Rule.SingleUserAllSpaceLimit
	//fmt.Println("=====", game.Rule.RobotLine)
	//fmt.Println(game.Rule.BetList,":",game.Rule.SingleUserSingleSpaceLimit,":",game.Rule.SingleUserAllSpaceLimit,":",game.Rule.AllUserSingleSpaceLimit,":",game.Rule.AllSpaceLimit)
}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(BRTB.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}

	msg.UserBetLimit = int32(game.Rule.UserBetLimit)
	msg.BetMinLimit = game.Rule.BetMinLimit
	u.SendMsg(int32(BRTB.SendToClientMessageType_RoomRolesInfo), msg)
}

func (game *Game) SendRoomInfo() {
	if game.Status == 0 {
		return
	}

	msg := new(BRTB.RoomSenceInfoMsg)
	msg.TrendList = new(BRTB.Trend)
	msg.TrendList.Info = append(msg.TrendList.Info, game.WinTrend...)
	msg.GameStatus = new(BRTB.StatusMessage)
	msg.GameStatus.Status = int32(game.Status)
	if game.TimerJob != nil {
		msg.GameStatus.StatusTime = int32(game.TimerJob.GetTimeDifference())
		msg.GameStatus.TotalStatusTime = int32(game.TimerJob.GetIntervalTime() / time.Millisecond)
	}

	msg.RoomID = game.Table.GetRoomID()
	msg.BaseBet = int64(game.Rule.BetList[0])
	msg.UserLimit = game.Rule.UserBetLimit
	msg.OnlineNumber = int64(len(game.AllUserList))
	//发送给框架
	//b, _ := proto.Marshal(msg)
	//game.Table.BroadcastAll(int32(rbwar.SendToClientMessageType_RoomSenceInfo), b)
	game.Table.BroadcastAll(int32(BRTB.SendToClientMessageType_RoomSenceInfo), msg)
}

func (game *Game) testCard(buffer []byte) {
	tmp := &BRTB.TempCardMsg{}
	proto.Unmarshal(buffer, tmp)
	game.testDices = tmp.Dices
	log.Debugf("收到的牌型为：%v", tmp)
	log.Debugf("测试牌为：%v， %v", game.testDices)
}

func (game *Game) BroadCastPokerCount() {
	//msg := new(BRTB.PokerCount)
	//msg.Count = int32(game.gp.GetCardsCount()) - game.LastCardCount
	//game.Table.Broadcast(int32(BRTB.SendToClientMessageType_GamePokerCount), msg)
}

func (game *Game) AddCards() {
	temp := model.GetDices()
	for i := 0; i <= 2; i++ {
		game.Dices[i] = temp[i]

	}
}

func (g *Game) SendMiPaiSeat() {
	//msg := new(BRTB.MiPaiUserInfo)
	//ZhuangMiUser, ok1 := g.SenceSeat.UserSeat[g.SenceSeat.BetZhuangMaxID]
	//XianMiUser, ok2 := g.SenceSeat.UserSeat[g.SenceSeat.BetXianMaxID]
	//if ok1 {
	//	msg.ZhuangSeatID = int32(ZhuangMiUser.SeatNo)
	//}
	//
	//if ok2 {
	//	msg.XianSeatID = int32(XianMiUser.SeatNo)
	//}
	//
	//g.Table.Broadcast(int32(BRTB.SendToClientMessageType_UpdateMiSeatID), msg)
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
		u, ok := game.AllUserList[v.User.User.GetId()]
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
		if game.SenceSeat.CheckUserOnChair(u.User.GetId()) {
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
	//fmt.Println("chushihua ")
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
	Millionaireid = user[0].User.GetId()

	u, ok := game.AllUserList[Millionaireid]
	if ok {
		u.Icon = Millionaire
		//fmt.Println("大赢家", u.User.GetId())
	}
	if len(user) == 1 {
		return
	}

	//大富豪2 近20局赢钱最多的 大赢家> 大富豪> 神算子
	sort.Sort(model.RegalUser(user))
	for i := 0; i < len(user); i++ {
		if user[i].User.GetId() != Millionaireid {
			bigWinnerid = user[i].User.GetId()
			break
		}
	}
	u1, ok1 := game.AllUserList[bigWinnerid]
	if ok1 {
		u1.Icon = bigWinner
		//fmt.Println("大富豪",u1.User.GetId())
	}
	//神算子1 胜率最高的 大赢家> 大富豪> 神算子
	sort.Sort(model.MasterUser(user))
	for i := 0; i < len(user); i++ {
		if user[i].User.GetId() != Millionaireid && user[i].User.GetId() != bigWinnerid {
			mastid = user[i].User.GetId()
			break
		}
	}
	u2, ok2 := game.AllUserList[mastid]
	if ok2 {
		u2.Icon = Master
		//fmt.Println("神算子",u2.User.GetId())
	}
}

//获取赔付倍数，倍数
func (game *Game) GetOdds(Count int32) (odds int64) {
	//model.GetDiceResult(game.Dices)
	if Count == 4 || Count == 17 {
		return 50
	} else if Count == 5 || Count == 16 {
		return 18
	} else if Count == 6 || Count == 15 {
		return 14
	} else if Count == 7 || Count == 14 {
		return 12
	} else if Count == 8 || Count == 13 {
		return 8
	} else if Count == 9 || Count == 12 || Count == 10 || Count == 11 {
		return 6
	} else {
		//如果出了3 和18
		return 0
	}
}
func (game *Game) repeatBet(bts []byte, user player.PlayerInterface) {
	msg := new(BRTB.BetRept)
	if err := proto.Unmarshal(bts, msg); err != nil {
		return
	}
	userInfo := game.getUser(user)

	var allBet int64
	var totallBet int64
	if len(msg.BetArea) != 25 {
		return
	}
	for k, v := range msg.BetArea {
		if (game.BetTotal[k] + v) > game.Rule.AllUserSingleSpaceLimit[k] {
			model.SendBetFailMessage(fmt.Sprintf("%s区域的下注已经达到总额度限制！重复下注失败", GetStringNameById(k)), userInfo)
			return
		}
		totallBet += game.BetTotal[k]
		if v < 0 {
			continue
		}
		allBet += v
	}
	if totallBet+allBet > game.Rule.AllSpaceLimit {
		model.SendBetFailMessage("您已达到该房间的下注额度限制！重复下注失败", userInfo)
		return
	} else if userInfo.User.GetScore() < allBet {
		model.SendBetFailMessage("您的余额不足,重复下注失败", userInfo)
		return
	} else {
		for index, v := range msg.BetArea {
			userInfo.BetArea[index] += v
			game.BetTotal[index] += v
			game.TotalUserBet[index] += v
			if !userInfo.User.IsRobot() {
				game.TotalUserBet[index] += v
			}
		}
		user.SetScore(game.Table.GetGameNum(), -1*allBet, game.Table.GetRoomRate())
		// 同步局数
		userInfo.AllBet += allBet
		userInfo.TotalBet += allBet
		user.SendMsg(int32(BRTB.SendToClientMessageType_BetReptRet), msg)

	}

}

//获取区域名字
func GetStringNameById(id int) string {
	switch id {
	case Big:
		return "大"
	case Small:
		return "小"
	case Single:
		return "单"
	case Double:
		return "双"
	case Four:
		return "4点"
	case Five:
		return "5点"
	case Six:
		return "6点"
	case Seven:
		return "7点"
	case Eight:
		return "8点"
	case Nine:
		return "9点"
	case Ten:
		return "10点"
	case Eleven:
		return "11点"
	case Twelve:
		return "12点"
	case Thirteen:
		return "13点"
	case Fourteen:
		return "14点"
	case Fifteen:
		return "15点"
	case Sixteen:
		return "16点"
	case Seventeen:
		return "17点"
	case Wei:
		return "任意围"
	case WeiOne:
		return "围(111)"
	case WeiTwo:
		return "围(222)"
	case WeiThree:
		return "围(333)"
	case WeiFour:
		return "围(444)"
	case WeiFive:
		return "围(555)"
	case WeiSix:
		return "围(666)"

	}
	return "错误"
}
