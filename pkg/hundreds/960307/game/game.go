package game

import (
	"fmt"
	"game_poker/BRZJH/config"
	"game_poker/BRZJH/model"
	"math/rand"
	"sort"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"game_frame_v2/game/clock"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

const (
	QINGLONG = 0
	BAIHU    = 1
	ZHUQUE   = 2
	XUANWU   = 3
)

type Trend struct {
	Win  bool
	Type int32
}

type TableTrend struct {
	t [4]Trend
}

type Game struct {
	Table               table.TableInterface     // table interface
	AllUserList         map[int64]*model.User    //所有的玩家列表
	Status              BRZJH.GameStatus         // 房间状态1 表示
	Win                 []int32                  // 0表示青龙胜利，1表示白虎胜利，2表示朱雀，3表示玄武
	LastWinIsRedOrBlack int                      // 最近一次开龙还是虎
	IsLuckWin           bool                     // 幸运一击是否胜利
	BetTotal            [4]int64                 //总下注统计  局结束重置数据
	TotalUserBet        [4]int64                 //用户各区域下注统计   局结束重置数据
	TotalRobotBet       [4]int64                 //机器人各区域下注统计   局结束重置数据
	SenceSeat           model.SceneInfo          //下注的玩家列表
	TimerJob            *clock.Job               //job
	RobotTimerJob       *clock.Job               //机器人job
	LastMasterBetType   int32                    //最近一次神算子下注的类型
	WinTrend            []TableTrend             //赢的走势
	CountUserList       []*model.User            //统计后的玩家列表
	Rule                config.RoomRules         //房间规则信息
	BaseBet             int64                    //最低下注
	Zhuang              player.PlayerInterface   //庄信息
	Card                [5]model.JHCard          //排序的5个牌
	SendCard            [5]model.JHCard          //发给客户端的5个牌
	OddsInfo            int                      //几倍场
	ZhuangList          []player.PlayerInterface //庄家列表
	LastCount           int                      //本次庄剩余次数
	AllKill             int32                    //通杀次数
	AllPay              int32                    //通赔次数
	HasTest             bool                     //有无测试
	CheatValue          int                      //作弊率
	PokerMsg            *BRZJH.PokerMsg          //翻牌消息
	LoopBetTimer        *clock.Job               // 下注广播消息定时器
	TotalBetTemp        [4]int64                 //总下注统计  局结束重置数据
	BroadTotallBetEnd   *BRZJH.BroadBetEnd       //各区域总下注广播
	OnlineUserList      []*model.User            //所有的玩家列表 用于自动上座，按顺序上坐
	ZhuangTotalWin      Zhuanginfo               //当前庄家赢钱总和,玩家ID 赢钱总和
	sysCheat            string
}
type Zhuanginfo struct {
	UserId   int64
	TotalWin int64
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.OnlineUserList = make([]*model.User, 0)
	game.SenceSeat.Init()
	game.PokerMsg = new(BRZJH.PokerMsg)
	game.PokerMsg.Cards = make([]*BRZJH.Poker, 5)
	for i := 0; i < 5; i++ {
		game.PokerMsg.Cards[i] = new(BRZJH.Poker)
		game.PokerMsg.Cards[i].Cards = make([]byte, 3)
	}

	game.PokerMsg.Type = make([]int32, 5)
	game.PokerMsg.Win = make([]bool, 4)

	game.BroadTotallBetEnd = new(BRZJH.BroadBetEnd)
	game.BroadTotallBetEnd.BetGold = make([]int64, 4)
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
	log.Tracef("user close %d", user.GetID())
	u := game.getUser(user)
	//有下注时不让玩家离开
	if u.TotalBet != 0 {
		log.Tracef("有下注时不让玩家离开")
		return false
	}

	if game.Zhuang == u.User && game.Status >= BRZJH.GameStatus_BetStatus {
		return false
	}

	game.XiaZhuang(u.User)

	if game.Status == BRZJH.GameStatus_SettleStatus && user == game.Zhuang {
		game.Zhuang = nil
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}
	delete(game.AllUserList, user.GetID())
	//删除用户
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if u.TotalBet != 0 {
			msg := new(BRZJH.ExitFail)
			msg.FailReason = "游戏中不能退出！"
			user.SendMsg(int32(BRZJH.SendToClientMessageType_ExitRet), msg)
			return false
		} else if user == game.Zhuang {
			msg := new(BRZJH.ExitFail)
			msg.FailReason = "游戏中不能退出！"
			user.SendMsg(int32(BRZJH.SendToClientMessageType_ExitRet), msg)
			return false
		}
	}

	game.XiaZhuang(u.User)
	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}
	delete(game.AllUserList, user.GetID())
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	if int32(BRZJH.ReceiveMessageType_BetID) == subCmd {
		game.Bet(buffer, user)
	} else if int32(BRZJH.ReceiveMessageType_SitDown) == subCmd {
		//game.UserSitDown(buffer, user)
	} else if int32(BRZJH.ReceiveMessageType_GetTrend) == subCmd {
		game.SendTrend(user)
	} else if int32(BRZJH.ReceiveMessageType_GetUserListInfo) == subCmd {
		game.SendUserListInfo(user)
	} else if int32(BRZJH.ReceiveMessageType_StandUp) == subCmd {
		//game.OnUserStanUp(user)
	} else if int32(BRZJH.ReceiveMessageType_ShangZhuang) == subCmd {
		game.shangZhuang(user)
	} else if int32(BRZJH.ReceiveMessageType_GetZhuangList) == subCmd {
		game.SendZhuangList(user)
	} else if int32(BRZJH.ReceiveMessageType_XiaZhuang) == subCmd {
		game.XiaZhuang(user)
	} else if int32(BRZJH.ReceiveMessageType_tempCard) == subCmd {
		//game.Test(buffer)
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
	game.SendSceneMsg(user)
	game.SendUserBet(u)

	if game.Status >= BRZJH.GameStatus_ShowPoker {
		user.SendMsg(int32(BRZJH.SendToClientMessageType_PokerInfo), game.PokerMsg)
		if game.Status == BRZJH.GameStatus_SettleStatus && u.SettleMsg != nil {
			user.SendMsg(int32(BRZJH.SendToClientMessageType_UserComeBack), u.SettleMsg)
		}
	}

	game.SendZhuangJiaInfo(user)
	if game.TimerJob != nil {
		game.SendToUserStatusMsg(int(game.TimerJob.GetTimeDifference()), user)
	}

	game.SendTrend(user)

	return true
}

//开始游戏
func (game *Game) Start() {
	game.checkUserBet()

	if game.Table.GetRoomID() == -1 {
		game.Status = 0
		return
	}
	game.Status = BRZJH.GameStatus_StartMovie

	game.Table.StartGame()
	for _, u := range game.ZhuangList {
		if u.GetScore() < game.Rule.ZhuangLimit {
			game.XiaZhuang(u)
		}
	}
	game.SetZhuang()
	game.SendZhuangJiaInfo(nil)
	game.SetIcon()
	game.SelectUserListInfoBefore6SitDownChair()
	//game.SendUserBetLimitMultiple()
	game.LastMasterBetType = -1

	if game.RobotTimerJob == nil {
		r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
		//game.RobotTimerJob, _ = game.Table.AddTimer(time.Duration(r), game.RobotSitDown)
		game.Table.AddTimer(time.Duration(r), game.RobotShangZhuang)
		//r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
		//game.Table.AddTimer(time.Duration(r1), game.RobotStandUp)
	}

	//for _, u := range game.SenceSeat.SenceSeat {
	//	if u.User.User.GetScore() < int64(u.User.Rule.SitDownLimit) {
	//		game.OnUserStanUp(u.User.User)
	//	}
	//}

	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRZJHConfig.Taketimes.Startmove), game.StartBet)
	//开始动画消息
	game.SendStatusMsg(config.BRZJHConfig.Taketimes.Startmove)
}

//开始下注
func (game *Game) StartBet() {
	game.ResetData()
	game.Status = BRZJH.GameStatus_BetStatus

	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRZJHConfig.Taketimes.Startbet), game.EndBet)
	//发送开始下注消息
	game.SendStatusMsg(config.BRZJHConfig.Taketimes.Startbet)

}

//结束下注
func (game *Game) EndBet() {
	game.Status = BRZJH.GameStatus_EndBetMovie
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRZJHConfig.Taketimes.Endmove), game.getResult)
	//发送结束下注消息
	game.SendStatusMsg(config.BRZJHConfig.Taketimes.Endmove)
}

//结算
func (game *Game) Settle() {
	game.Status = BRZJH.GameStatus_SettleStatus
	//结算信息
	game.sendSettleMsg()
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRZJHConfig.Taketimes.Endpay), game.Start)
	//发送结算结果
	game.SendStatusMsg(config.BRZJHConfig.Taketimes.Endpay)
	game.Table.EndGame()
}

//发送状态信息
func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(BRZJH.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(BRZJH.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(BRZJH.SendToClientMessageType_Status), msg)
}

//用户下注
func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != BRZJH.GameStatus_BetStatus {
		return
	}
	//庄家不能下注
	if game.Zhuang == user {
		return
	}
	//用户下注
	BetPb := &BRZJH.Bet{}
	proto.Unmarshal(buffer, BetPb)
	u := game.getUser(user)
	if u.Bet(BetPb, game.BetTotal) {
		game.BetTotal[BetPb.BetType%4] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])

		u.User.SetScore(game.Table.GetGameNum(), -int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))]), game.Table.GetRoomRate())
		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType%4] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		} else {
			game.TotalRobotBet[BetPb.BetType%4] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		}
	}

	//if game.SenceSeat.GetMaster() == u.SceneChairId {
	//	game.LastMasterBetType = BetPb.BetType
	//}
	//神算子下注
	u1, ok := game.SenceSeat.UserSeat[user.GetID()]
	if ok {
		if u1.User.Icon == 1 {
			game.LastMasterBetType = BetPb.BetType
		}

	}
}

func (game *Game) UserSitDown(buffer []byte, user player.PlayerInterface) {
	if user == game.Zhuang {
		return
	}
	us := &BRZJH.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo)) {
			u.SceneChairId = int(us.ChairNo)
			game.SendSceneMsg(nil)
		}
	}
}

// 发牌 不出地龙牌型
func (game *Game) DealPoker() {
	var gp model.GamePoker
	gp.InitPoker()
	gp.ShuffleCards()
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			game.Card[i].Cards[j] = gp.DealCards()
		}
		game.Card[i].Cards = model.SortCards(game.Card[i].Cards)
		//如果牌型为地龙 不出当前的牌重新发
		temp, _ := model.GetCardTypeJH(game.Card[i].Cards)
		if temp == model.CardTypeDL {
			i = i - 1
			continue
		}
		game.SendCard[i] = game.Card[i]
	}

}

// 比牌， 并进行排序
func (game *Game) ComparePoker() {
	//对这五服牌进行排序
	for i := 0; i < 4; i++ { //最外层表示一共循环的次数
		for j := 0; j < (4 - i); j++ { //内层表示逐层比较的次数递减
			if model.ComPareCard(game.Card[j].Cards, game.Card[j+1].Cards) == 2 {
				game.Card[j], game.Card[j+1] = game.Card[j+1], game.Card[j]
			}
		}
	}
}

//检查用户是否被踢掉
func (game *Game) checkUserBet() {
	for k, u := range game.AllUserList {
		if u.NoBetCount >= (config.BRZJHConfig.Unplacebetnum+1) ||
			(u.User.IsRobot() &&
				(u.User.GetScore() > game.Rule.RobotMaxGold || u.User.GetScore() < game.Rule.RobotMinGold)) {
			if u.User == game.Zhuang {
				u.NoBetCount = 0
				continue
			}
			//踢掉用户
			u.NoBetCount = 0
			if u.SceneChairId != 0 {
				game.OnUserStanUp(u.User)
			}

			if game.Zhuang == u.User {
				game.Zhuang = nil
			}

			game.XiaZhuang(u.User)
			delete(game.AllUserList, k)
			game.DeleteExitUserFromOnlineUserListSlice(u)

			game.Table.KickOut(u.User)
		}
	}
}

//发送结算消息
func (game *Game) sendSettleMsg() {
	var tmpTrend TableTrend
	WinCount := 0
	PayCount := 0
	//log.Traceln("庄家",model.GetCardString(game.SendCard[0].Cards),"牌型",model.GetTypeString(model.GetCardsType(game.SendCard[0].Cards)))
	for i := 1; i < 5; i++ {
		tmpTrend.t[i-1].Type = model.GetCardsType(game.SendCard[i].Cards)
		//和庄家比牌，记录赔付和赢次数及输赢走势
		if model.ComPareCard(game.SendCard[0].Cards, game.SendCard[i].Cards) == 2 {
			//log.Traceln("赢",model.GetCardString(game.SendCard[i].Cards),"牌型",model.GetTypeString(model.GetCardsType(game.SendCard[i].Cards)))
			tmpTrend.t[i-1].Win = true
			PayCount++
		} else {
			//log.Traceln("输",model.GetCardString(game.SendCard[i].Cards),"牌型",model.GetTypeString(model.GetCardsType(game.SendCard[i].Cards)))
			tmpTrend.t[i-1].Win = false
			WinCount++
		}
	}
	//输赢走势
	game.WinTrend = append(game.WinTrend, tmpTrend)
	winlen := len(game.WinTrend)
	//记录最近20局输赢走势
	if winlen > 20 {
		game.WinTrend = append(game.WinTrend[:(winlen-20-1)], game.WinTrend[(winlen-20):]...)
	}

	game.CountUserList = make([]*model.User, 0)
	//统计庄赢
	var ZhuangWin int64
	var AreaWin []int64

	QingLongBetCount := 0
	BaiHuBetCount := 0
	ZhuQueBetCount := 0
	XuanWuBetCount := 0
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [4]int64

	QingLongStr := fmt.Sprintf(" 黑桃区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[QINGLONG]), score.GetScoreStr(game.BetTotal[QINGLONG]-game.TotalUserBet[QINGLONG]), score.GetScoreStr(game.TotalUserBet[QINGLONG]))

	BaiHuStr := fmt.Sprintf(" 红桃区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[BAIHU]), score.GetScoreStr(game.BetTotal[BAIHU]-game.TotalUserBet[BAIHU]), score.GetScoreStr(game.TotalUserBet[BAIHU]))

	ZhuQueStr := fmt.Sprintf(" 梅花区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[ZHUQUE]), score.GetScoreStr(game.BetTotal[ZHUQUE]-game.TotalUserBet[ZHUQUE]), score.GetScoreStr(game.TotalUserBet[ZHUQUE]))

	XuanWuStr := fmt.Sprintf(" 方块区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[XUANWU]), score.GetScoreStr(game.BetTotal[XUANWU]-game.TotalUserBet[XUANWU]), score.GetScoreStr(game.TotalUserBet[XUANWU]))

	if game.Zhuang != nil {
		var win int64
		tax := int64(0)
		var usertotalwin int64
		//庄家输钱
		usertotallose := int64(0)
		for i := 0; i < 4; i++ {
			var tmpwin int64
			//统计用户庄输赢
			if tmpTrend.t[i].Win {
				//用户庄输钱统计-=下注区域*牌型倍数
				thiswin := game.BetTotal[i]
				win -= thiswin
				usertotallose += thiswin
				ZhuangWin -= thiswin
				tmpwin, _ = game.Zhuang.SetScore(game.Table.GetGameNum(), -thiswin, game.Table.GetRoomRate())
			} else {
				//用户庄赢钱统计+=下注区域*牌型倍数
				thiswin := game.BetTotal[i]
				ZhuangWin += thiswin
				tmpwin, _ = game.Zhuang.SetScore(game.Table.GetGameNum(), thiswin, game.Table.GetRoomRate())
				tax += thiswin - tmpwin
				win += tmpwin
				usertotalwin += tmpwin
			}

			AreaWin = append(AreaWin, tmpwin)
		}
		totalbet := int64(game.Rule.BetList[0])
		if win >= 0 {
			game.Zhuang.SendChip(int64(game.Rule.BetList[0]))
		} else {
			totalbet = usertotallose
			game.Zhuang.SendChip(-win)
		}
		u := game.getUser(game.Zhuang)
		profitAmount := game.Zhuang.GetScore() - u.CruenSorce
		u.ResetUserData()
		game.Zhuang.SendRecord(game.Table.GetGameNum(), profitAmount, totalbet, tax, usertotalwin, "")
		game.PaoMaDeng(win, game.Zhuang)
		//统计庄家上庄期间赢利总和
		if game.Zhuang.GetID() != game.ZhuangTotalWin.UserId {
			game.ZhuangTotalWin.UserId = game.Zhuang.GetID()
			if win > 0 {
				game.ZhuangTotalWin.TotalWin = win
			} else {
				game.ZhuangTotalWin.TotalWin = 0
			}
		} else {
			if win > 0 {
				game.ZhuangTotalWin.TotalWin += win
			}
		}
		log.Traceln(game.ZhuangTotalWin, game.ZhuangTotalWin.UserId, game.Zhuang.GetID(), win)
	} else {
		game.ZhuangTotalWin.TotalWin = 0
		game.ZhuangTotalWin.UserId = 0
	}

	//通杀通赔
	var AllPay bool
	var AllKill bool
	if PayCount == 4 {
		AllPay = true
	} else {
		AllPay = false
	}

	if WinCount == 4 {
		AllKill = true
	} else {
		AllKill = false
	}
	//庄家列表中的用户未下注局数置为0
	game.setZhangListUserNoBetCount()
	SceneSettleMsg := new(BRZJH.SceneUserSettle)
	for _, u := range game.AllUserList {
		u.NoBetCount++
		if u.User == game.Zhuang {
			u.NoBetCount = 0
		}

		if !u.User.IsRobot() && u.User != game.Zhuang {
			if u.NoBetCount >= (config.BRZJHConfig.Unplacebetnum + 1) {
				//发送踢掉用户
				msg := new(BRZJH.KickOutUserMsg)
				msg.KickOutReason = "由于您5局未下注，已被踢出房间！"
				u.User.SendMsg(int32(BRZJH.SendToClientMessageType_KickOutUser), msg)
			}
		}

		SceneUserInfo := new(BRZJH.SceneUserInfo)

		msg := new(BRZJH.SettleMsg)

		var totalwin int64
		var Award int64
		var ResBet int64 //结算后玩家的下注
		var Chip int64
		var totalTax int64 //总税收
		//msg.Type = append(msg.Type, int32(model.GetNiuNiuType(game.SendCard[0])))
		//统计分数
		for i := 0; i < 4; i++ {
			msg.TotalBetInfo = append(msg.TotalBetInfo, game.BetTotal[i])
			msg.UserBetInfo = append(msg.UserBetInfo, u.BetInfo[i])
			msg.Win = append(msg.Win, tmpTrend.t[i].Win)
			//			msg.Type = append(msg.Type, tmpTrend.t[i].Type)
			if u.TotalBet <= 0 {
				continue
			}
			if tmpTrend.t[i].Win {
				//用户赢 = 压注*2
				win := u.BetInfo[i] * 2
				totalwin += win
				msg.UserWin = append(msg.UserWin, win)
				var tax int64
				tax, _ = u.User.SetScore(game.Table.GetGameNum(), u.BetInfo[i], game.Table.GetRoomRate())
				capital, _ := u.User.SetScore(game.Table.GetGameNum(), u.BetInfo[i], 0)
				tax += capital
				totalTax += win - tax
				Award += tax
				msg.TotalWin += tax
				SystemWin[i] += tax
			} else {
				win := -u.BetInfo[i]
				ResBet += u.BetInfo[i]
				Chip += u.BetInfo[i] * 2
				//totalwin += win
				//u.User.SetScore(game.Table.GetGameNum(), win, game.Table.GetRoomRate())

				//win -= u.BetInfo[i]
				totalwin += win
				msg.TotalWin += win
				msg.UserWin = append(msg.UserWin, win)
			}
		}
		//log.Traceln("=========", msg.UserWin, msg.UserBetInfo, msg.TotalBetInfo, msg.TotalWin)
		if u.User != game.Zhuang {
			SceneUserInfo.BetInfo = msg.UserBetInfo
		} else {
			SceneUserInfo.BetInfo = msg.TotalBetInfo
			SceneUserInfo.AreaWin = AreaWin
		}

		SceneUserInfo.TotalWin = msg.TotalWin
		SceneUserInfo.UserID = int64(u.User.GetID())
		SceneUserInfo.SceneSeatID = int32(u.SceneChairId)
		//统计玩家信息
		if (totalwin) > u.TotalBet {
			u.UserCount(true, msg.TotalWin)
		} else {
			if u.User != game.Zhuang {
				u.UserCount(false, 0)
			}
		}

		//写入数据库统计信息
		if MaxWinGold < u.User.GetScore()-u.CruenSorce {
			MaxWinGold = u.User.GetScore() - u.CruenSorce
			MaxWinUserID = u.User.GetID()
		}

		if !u.User.IsRobot() {
			var temp string
			var temp1 string
			if u.TotalBet != 0 {
				temp += fmt.Sprintf("用户ID：%v，开始金币：%v，投注额:", u.User.GetID(), score.GetScoreStr(u.CruenSorce))
				temp1 += fmt.Sprintf(" 输赢：")
			}
			if u.BetInfo[QINGLONG] != 0 {
				Win := msg.UserWin[QINGLONG] - u.BetInfo[QINGLONG]
				if !tmpTrend.t[QINGLONG].Win {
					Win = msg.UserWin[QINGLONG]
				}
				temp += fmt.Sprintf("黑桃：%v ", score.GetScoreStr(u.BetInfo[QINGLONG]))
				temp1 += fmt.Sprintf("黑桃：%v ", score.GetScoreStr(Win))
				QingLongBetCount++
			}

			if u.BetInfo[BAIHU] != 0 {
				Win := msg.UserWin[BAIHU] - u.BetInfo[BAIHU]
				if !tmpTrend.t[BAIHU].Win {
					Win = msg.UserWin[BAIHU]
				}
				temp += fmt.Sprintf("红桃：%v ", score.GetScoreStr(u.BetInfo[BAIHU]))
				temp1 += fmt.Sprintf("红桃：%v ", score.GetScoreStr(Win))
				BaiHuBetCount++
			}

			if u.BetInfo[ZHUQUE] != 0 {
				Win := msg.UserWin[ZHUQUE] - u.BetInfo[ZHUQUE]
				if !tmpTrend.t[ZHUQUE].Win {
					Win = msg.UserWin[ZHUQUE]
				}
				temp += fmt.Sprintf("梅花：%v ", score.GetScoreStr(u.BetInfo[ZHUQUE]))
				temp1 += fmt.Sprintf("梅花：%v ", score.GetScoreStr(Win))
				ZhuQueBetCount++
			}

			if u.BetInfo[XUANWU] != 0 {
				Win := msg.UserWin[XUANWU] - u.BetInfo[XUANWU]
				if !tmpTrend.t[XUANWU].Win {
					Win = msg.UserWin[XUANWU]
				}
				temp += fmt.Sprintf("方块：%v ", score.GetScoreStr(u.BetInfo[XUANWU]))
				temp1 += fmt.Sprintf("方块：%v ", score.GetScoreStr(Win))

				XuanWuBetCount++
			}
			temp1 += fmt.Sprintf(" 总输赢：%v，用户剩余金额：%v \r\n", score.GetScoreStr(u.User.GetScore()-u.CruenSorce), score.GetScoreStr(u.User.GetScore()))
			temp += temp1
			game.Table.WriteLogs(u.User.GetID(), temp)
		}
		game.PaoMaDeng(totalwin-u.TotalBet, u.User)
		game.CountUser(u)

		msg.UserScore = u.User.GetScore()
		SceneUserInfo.UserScore = msg.UserScore
		msg.AllKill = AllKill
		msg.AllPay = AllPay

		u.User.SendMsg(int32(BRZJH.SendToClientMessageType_Settle), msg)

		if u.TotalBet > 0 || game.Zhuang == u.User {
			u.SettleMsg = msg
		} else {
			u.SettleMsg = nil
		}

		if u.SceneChairId != 0 {
			SceneSettleMsg.UserList = append(SceneSettleMsg.UserList, SceneUserInfo)
		}

		if u.User == game.Zhuang {
			SceneUserInfo.SceneSeatID = 7 //庄家座位ID7
			//SceneUserInfo.ZhuangTotalWin = game.ZhuangTotalWin.TotalWin
			SceneSettleMsg.UserList = append(SceneSettleMsg.UserList, SceneUserInfo)
			//log.Traceln("庄信息", SceneUserInfo)
		} else {
			u.User.SendChip(u.TotalBet + ResBet)
			//u.User.SetBetsAmount(u.TotalBet + ResBet)
		}
		//u.ResetUserData()
		user := game.getUser(u.User)
		betsAmount := u.TotalBet + ResBet
		profitAmount := u.User.GetScore() - user.CruenSorce
		u.ResetUserData()
		u.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, totalTax, Award, "")
	}
	//用户列表排序
	cou := model.Usercount{}
	cou = game.CountUserList
	sort.Sort(cou)
	//game.SetIcon()

	//日志信息
	QingLongStr += fmt.Sprintf("真人数量:%v 牌型：%v, %v, 输赢：%v\r\n", QingLongBetCount,
		model.GetCardString(game.SendCard[1].Cards), model.GetTypeString(tmpTrend.t[0].Type),
		score.GetScoreStr(game.BetTotal[QINGLONG]-SystemWin[QINGLONG]))
	BaiHuStr += fmt.Sprintf("真人数量:%v 牌型：%v, %v, 输赢：%v\r\n", BaiHuBetCount,
		model.GetCardString(game.SendCard[2].Cards), model.GetTypeString(tmpTrend.t[1].Type),
		score.GetScoreStr(game.BetTotal[BAIHU]-SystemWin[BAIHU]))
	ZhuQueStr += fmt.Sprintf("真人数量:%v 牌型：%v, %v, 输赢：%v\r\n", ZhuQueBetCount,
		model.GetCardString(game.SendCard[3].Cards), model.GetTypeString(tmpTrend.t[2].Type),
		score.GetScoreStr(game.BetTotal[ZHUQUE]-SystemWin[ZHUQUE]))
	XuanWuStr += fmt.Sprintf("真人数量:%v 牌型：%v, %v, 输赢：%v\r\n", XuanWuBetCount,
		model.GetCardString(game.SendCard[4].Cards), model.GetTypeString(tmpTrend.t[3].Type),
		score.GetScoreStr(game.BetTotal[XUANWU]-SystemWin[XUANWU]))

	var str string
	count := int64(0)
	for k, v := range game.BetTotal {
		count += v - SystemWin[k]
	}
	if game.Zhuang != nil && !game.Zhuang.IsRobot() {
		tname := ""
		if game.Zhuang.IsRobot() {
			tname = "机器人"
		} else {
			tname = "真人"
		}
		str = fmt.Sprintf("坐庄ID：%v 角色: %v 庄区域牌型：牌型：%v %v 输赢：%v\r\n", game.Zhuang.GetID(), tname,
			model.GetCardString(game.SendCard[0].Cards), model.GetTypeString(int32(model.GetCardsType(game.SendCard[0].Cards))),
			score.GetScoreStr(ZhuangWin))
	} else {
		str = fmt.Sprintf("坐庄ID： 角色:系统 庄区域牌型：牌型：%v %v 系统输赢额度：%v\r\n",
			model.GetCardString(game.SendCard[0].Cards), model.GetTypeString(int32(model.GetCardsType(game.SendCard[0].Cards))), score.GetScoreStr(int64(count)))
	}
	str += fmt.Sprintf("%v作弊率：%v \r\n", game.sysCheat, game.CheatValue)
	str += fmt.Sprintf("最高获利用户ID：%v 获得：%v\r\n",
		MaxWinUserID, score.GetScoreStr(MaxWinGold))
	//log.Debugf(str)
	totalstr := QingLongStr + BaiHuStr + ZhuQueStr + XuanWuStr + str
	game.Table.WriteLogs(0, totalstr)

	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

// 发牌 比牌  获取结果
func (game *Game) getResult() {
	game.Status = BRZJH.GameStatus_ShowPoker
	//获取作弊值
	cheat, _ := game.Table.GetRoomProb()
	pro := config.BRZJHConfig.GetCheatValue(int(cheat))
	game.CheatValue = int(cheat)
	if cheat == 0 {
		game.sysCheat = "获取作弊率为0 "
		cheat = 1000
	} else {
		game.sysCheat = ""
	}
	var Eat int
	var Out int
	r := rand.Intn(10000)
	if pro != 0 {
		if r < pro {
			//吃分
			Eat = 1
		} else {
			//吐分
			Out = 1
		}
	}

	log.Debugf("系统吃分:%v %v", Eat, pro)
	//先发五付牌
	if !game.HasTest {
		//先发五付牌
		if Out == 0 && Eat == 0 {
			//不控制情况下随机发牌
			game.DealPoker()
			for i := 0; i < 5; i++ {
				game.SendCard[i] = game.Card[i]
			}
		} else {
			//控制下发牌
			game.DealPoker()
			//对发出的5副牌从大到小排序
			game.ComparePoker()
			if game.Zhuang != nil && !game.Zhuang.IsRobot() {
				//获取用户为庄家时的结果
				game.HasZhuangResult(Eat, Out)
			} else {
				//系统为庄时的结果
				game.SystemResult(Eat, Out)
			}
		}
	} else {
		game.HasTest = false
	}

	//比牌结果
	for i := 1; i < 5; i++ {
		if model.ComPareCard(game.SendCard[0].Cards, game.SendCard[i].Cards) == 2 {
			game.PokerMsg.Win[i-1] = true
		} else {
			game.PokerMsg.Win[i-1] = false
		}
	}
	//牌型
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			game.PokerMsg.Cards[i].Cards[j] = game.SendCard[i].Cards[j]
		}
		tempType := model.GetCardsType(game.SendCard[i].Cards)
		if int(tempType) > model.CardTypeSJA23 {
			tempType -= 2
		} else if int(tempType) > model.CardTypeSZA23 {
			tempType--
		}
		game.PokerMsg.Type[i] = tempType

	}
	//发送比牌信息
	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_PokerInfo), game.PokerMsg)
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(config.BRZJHConfig.Taketimes.ShowPoker), game.Settle)
	game.SendStatusMsg(config.BRZJHConfig.Taketimes.ShowPoker)
}

//控制情况下用户为庄换牌结果
func (game *Game) HasZhuangResult(Eat int, Out int) {
	arr := []int{0, 1, 2, 3}

	var EatArr []int
	var tmppay int64
	for i := 1; i < 5; i++ {
		if model.ComPareCard(game.SendCard[0].Cards, game.SendCard[i].Cards) == 2 {
			tmppay += game.TotalRobotBet[i-1]
			EatArr = append(EatArr, i)
		} else {
			tmppay -= game.TotalRobotBet[i-1]
		}
	}

	if tmppay >= 0 && Eat == 1 {
		return
	}

	if tmppay <= 0 && Out == 1 {
		return
	}

	EatArr = make([]int, 0)

	//100次还不出结果直接出结果
	for n := 0; n < 100; n++ {
		r := rand.Intn(len(arr))
		EatArr = append(EatArr, arr[r])
		arr = append(arr[:r], arr[r+1:]...)
		var TotalPay int64
		for i := 0; i < len(EatArr); i++ {
			TotalPay += game.TotalRobotBet[EatArr[i]]
		}

		for i := 0; i < len(arr); i++ {
			TotalPay -= game.TotalRobotBet[arr[i]]
		}

		if TotalPay >= 0 && Eat == 1 {
			break
		}

		if TotalPay <= 0 && Out == 1 {
			break
		}

		if n == 100-1 {
			break
		}

		if len(arr) == 0 {
			arr = []int{0, 1, 2, 3}
			EatArr = make([]int, 0)
		}
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			game.SendCard[i].Cards[j] = 0
		}
	}

	if len(EatArr) > 0 {
		arr = []int{0, 1, 2, 3, 4}
		game.SendCard[0] = game.Card[len(EatArr)]
		arr = append(arr[:len(EatArr)], arr[len(EatArr)+1:]...)
		for i := 1; i < len(EatArr)+1; i++ {
			game.SendCard[EatArr[i-1]+1] = game.Card[i-1]
			arr = append(arr[:0], arr[1:]...)
		}

		for i := 0; i < 5; i++ {
			if game.SendCard[i].Cards[0] == 0 {
				r := rand.Intn(len(arr))
				game.SendCard[i] = game.Card[arr[r]]
				arr = append(arr[:r], arr[r+1:]...)
			}
		}
	} else {
		arr = []int{0, 1, 2, 3, 4}
		game.SendCard[0] = game.Card[4]
		for i := 0; i < 5; i++ {
			if game.SendCard[i].Cards[0] == 0 {
				r := rand.Intn(len(arr))
				game.SendCard[i] = game.Card[arr[r]]
				arr = append(arr[:r], arr[r+1:]...)
			}
		}
	}

	log.Debugf("玩家庄牌为：%v", game.SendCard)
}

//控制情况下系统为庄换牌结果
func (game *Game) SystemResult(Eat int, Out int) {
	arr := []int{0, 1, 2, 3}

	var EatArr []int
	var tmppay int64
	for i := 1; i < 5; i++ {
		if model.ComPareCard(game.SendCard[0].Cards, game.SendCard[i].Cards) == 2 {
			tmppay -= game.TotalUserBet[i-1]
			EatArr = append(EatArr, i)
		} else {
			tmppay += game.TotalUserBet[i-1]
		}
	}

	if tmppay >= 0 && Eat == 1 {
		return
	}

	if tmppay <= 0 && Out == 1 {
		return
	}

	EatArr = make([]int, 0)

	//100次还不出结果直接出结果
	for n := 0; n < 10000; n++ {
		r := rand.Intn(len(arr))
		EatArr = append(EatArr, arr[r])
		arr = append(arr[:r], arr[r+1:]...)
		var TotalPay int64
		for i := 0; i < len(EatArr); i++ {
			TotalPay -= game.TotalUserBet[EatArr[i]]
		}

		for i := 0; i < len(arr); i++ {
			TotalPay += game.TotalUserBet[arr[i]]
		}

		if TotalPay >= 0 && Eat == 1 {
			break
		}

		if TotalPay <= 0 && Out == 1 {
			break
		}

		if len(arr) == 0 {
			arr = []int{0, 1, 2, 3}
			EatArr = make([]int, 0)
		}
	}

	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			game.SendCard[i].Cards[j] = 0
		}
	}

	if len(EatArr) > 0 {
		arr = []int{0, 1, 2, 3, 4}
		game.SendCard[0] = game.Card[len(EatArr)]
		arr = append(arr[:len(EatArr)], arr[len(EatArr)+1:]...)
		for i := 1; i < len(EatArr)+1; i++ {
			arr = append(arr[:0], arr[1:]...)
			game.SendCard[EatArr[i-1]+1] = game.Card[i-1]
		}

		for i := 0; i < 5; i++ {
			if game.SendCard[i].Cards[0] == 0 {
				r := rand.Intn(len(arr))
				game.SendCard[i] = game.Card[arr[r]]
				arr = append(arr[:r], arr[r+1:]...)
			}
		}
	} else {
		arr = []int{0, 1, 2, 3, 4}
		game.SendCard[0] = game.Card[4]
		for i := 0; i < 5; i++ {
			if game.SendCard[i].Cards[0] == 0 {
				r := rand.Intn(len(arr))
				game.SendCard[i] = game.Card[arr[r]]
				arr = append(arr[:r], arr[r+1:]...)
			}
		}
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
		u.Time = time.Now().UnixNano() / 1e6
		game.OnlineUserList = append(game.OnlineUserList, u)
		game.CountUserList = append(game.CountUserList, u)
		u.ResetUserData()
	}
	return u
}

//发送场景消息
func (game *Game) SendSceneMsg(u player.PlayerInterface) {
	msg := new(BRZJH.SceneMessage)
	//bigwinner := game.SenceSeat.GetBigWinner()
	//master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(BRZJH.SeatUser)
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

	if u == nil {
		game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_SceneID), msg)
	} else {
		u.SendMsg(int32(BRZJH.SendToClientMessageType_SceneID), msg)
	}
}

func (game *Game) SendUserBet(u *model.User) {
	msg := new(BRZJH.SceneBetInfo)

	for i := 0; i < 4; i++ {
		msg.TotalBetInfo = append(msg.TotalBetInfo, game.BetTotal[i])
		msg.UserBetInfo = append(msg.UserBetInfo, u.BetInfo[i])
	}

	msg.UserBetTotal = u.TotalBet
	msg.MasterBetType = game.LastMasterBetType
	u.User.SendMsg(int32(BRZJH.SendToClientMessageType_BetInfo), msg)
}

func (game *Game) SendTrend(u player.PlayerInterface) {
	msg := new(BRZJH.Trend)

	j := 0
	if len(game.WinTrend) >= 20 {
		j = len(game.WinTrend) - 20
	}
	AllKillCount := 0
	AllPayCount := 0
	for ; j < len(game.WinTrend); j++ {
		table := new(BRZJH.TableTrend)
		AllKill := 0
		AllPay := 0
		for i := 0; i < 4; i++ {
			onetrend := new(BRZJH.OneTrend)
			onetrend.Type = game.WinTrend[j].t[i].Type
			onetrend.Win = game.WinTrend[j].t[i].Win
			table.Info = append(table.Info, onetrend)
			if onetrend.Win {
				AllPay++
			} else {
				AllKill++
			}
		}

		if AllPay == 4 {
			AllPayCount++
		} else if AllKill == 4 {
			AllKillCount++
		}
		msg.TableTrendInfo = append(msg.TableTrendInfo, table)
	}

	msg.AllKill = int32(AllKillCount)
	msg.AllPay = int32(AllPayCount)
	u.SendMsg(int32(BRZJH.SendToClientMessageType_TrendInfo), msg)
	log.Tracef("发送走势图:%v", msg)
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
	//if game.CountUserList[0].RetWinMoney < u.RetWinMoney {
	//	adduser = game.CountUserList[0]
	//	game.CountUserList[0] = u
	//}
	//
	//for i := 1; i < uc; i++ {
	//	if adduser.RetWinMoney > game.CountUserList[i].RetWinMoney {
	//		var tmp []*model.User
	//		tmp = append(tmp, game.CountUserList[i:]...)
	//		game.CountUserList = append(game.CountUserList[0:i], adduser)
	//		game.CountUserList = append(game.CountUserList, tmp...)
	//		return
	//	}
	//
	//}

	game.CountUserList = append(game.CountUserList, adduser)
}

func (game *Game) SendUserListInfo(user player.PlayerInterface) {
	msg := new(BRZJH.UserList)
	for _, u := range game.CountUserList {
		if u.User == game.Zhuang {
			userzhuanginfo := new(BRZJH.UserInfo)
			userzhuanginfo.NikeName = u.User.GetNike()
			userzhuanginfo.UserGlod = u.User.GetScore()
			//msg.ZhuangInfo.WinCount = int32(u.RetWin)
			userzhuanginfo.BetGold = game.ZhuangTotalWin.TotalWin
			userzhuanginfo.Head = u.User.GetHead()
			msg.ZhuangInfo = userzhuanginfo
			//msg.ZhuangInfo.Icon = u.Icon
			continue
		}
		userinfo := new(BRZJH.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.RetWinMoney
		userinfo.Head = u.User.GetHead()
		userinfo.Icon = u.Icon
		msg.UserList = append(msg.UserList, userinfo)
	}
	log.Tracef("SendUserListInfo %v", msg)
	user.SendMsg(int32(BRZJH.SendToClientMessageType_UserListInfo), msg)
}

func (game *Game) ResetData() {
	for i := 0; i < 4; i++ {
		game.TotalUserBet[i] = 0
		game.BetTotal[i] = 0
		game.TotalRobotBet[i] = 0
		game.TotalBetTemp[i] = 0
	}
}

func (game *Game) OnUserStanUp(user player.PlayerInterface) {
	if !game.SenceSeat.UserStandUp(user) {
		return
	}
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		u.SceneChairId = 0
	}
	game.SendSceneMsg(nil)
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
					us := &BRZJH.UserSitDown{}
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
	msg := new(BRZJH.SceneMessage)
	bigwinner := game.SenceSeat.GetBigWinner()
	master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(BRZJH.SeatUser)
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

	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_SceneID), msg)
}

func (game *Game) GameStart(user player.PlayerInterface) bool {
	if game.Status == 0 {
		game.Start()
		//定时广播区域下注信息
		game.LoopBetTimer, _ = game.Table.AddTimerRepeat(time.Duration(config.BRZJHConfig.Taketimes.BetGapBroadcast), 0, game.loopBetMsg)
		game.Table.AddTimerRepeat(1000, 0, game.SendRoomInfo)
	} else if game.TimerJob != nil {
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

	game.Rule.BetList = make([]int64, 0)
	BetBase, _ := js.Get("Bottom_Pouring").Int()
	betMinLimit, _ := js.Get("betMinLimit").Int()
	game.Rule.BetMinLimit = int64(betMinLimit)
	//Odds, _ := js.Get("line").Int()
	game.BaseBet = int64(BetBase)
	//game.Rule.UserBetLimit = int64(BetBase) * 5000
	//game.Rule.BetList = append(game.Rule.BetList, BetBase)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*10)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*50)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*100)
	//game.Rule.BetList = append(game.Rule.BetList, BetBase*500)
	//game.Rule.ZhuangLimit = game.BaseBet * 10000
	for i := 0; i < 4; i++ {
		game.Rule.BetLimit[i] = game.BaseBet * 100000
	}
	game.Rule.SitDownLimit = BetBase * 100
	level := game.Table.GetLevel()
	game.Rule.RobotMinGold = config.BRZJHConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.BRZJHConfig.Robotgold[level-1][1]

	//game.OddsInfo = Odds
	//game.Rule.OddsInfo = int32(Odds)
	game.Rule.ZhuangLimit = config.BRZJHConfig.Shangzhuanglimit5times[level-1]
	game.Rule.SingleUserAllSpaceLimit = config.BRZJHConfig.Singleuserallspacelimit5times[level-1]
	game.Rule.AllSpaceLimit = config.BRZJHConfig.Allspacelimit5times[level-1]
	for i := 0; i < 4; i++ {
		//log.Traceln(config.BRNNConfig.Singleusersinglespacelimit5times[level-1][3],config.BRNNConfig.Allusersinglespacelimit5times[level-1][i])
		game.Rule.SingleUserSingleSpaceLimit[i] = config.BRZJHConfig.Singleusersinglespacelimit5times[level-1][i]
		game.Rule.AllUserSingleSpaceLimit[i] = config.BRZJHConfig.Allusersinglespacelimit5times[level-1][i]
	}
	for i := 0; i < 5; i++ {
		game.Rule.BetList = append(game.Rule.BetList, config.BRZJHConfig.Chips5times[level-1][i])
	}

	game.Rule.UserBetLimit = game.Rule.SingleUserAllSpaceLimit
	//log.Traceln(game.OddsInfo,":",game.Rule.BetList,":",game.Rule.ZhuangLimit,":",game.Rule.SingleUserSingleSpaceLimit,":",game.Rule.SingleUserAllSpaceLimit,":",game.Rule.AllUserSingleSpaceLimit,":",game.Rule.AllSpaceLimit)

}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(BRZJH.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}

	msg.UserBetLimit = int32(game.Rule.UserBetLimit)
	msg.OddsInfo = int32(game.OddsInfo)
	msg.ZhuangLimit = game.Rule.ZhuangLimit
	msg.BetMinLimit = game.Rule.BetMinLimit
	u.SendMsg(int32(BRZJH.SendToClientMessageType_RoomRolesInfo), msg)
}

func (game *Game) SendRoomInfo() {
	if game.Status == 0 {
		return
	}
	msg := new(BRZJH.RoomSenceInfoMsg)
	msg.TrendList = new(BRZJH.Trend)
	var AllKillCount int32
	var AllPayCount int32
	for j := 0; j < len(game.WinTrend); j++ {
		table := new(BRZJH.TableTrend)
		KillCount := 0
		PayCount := 0
		for i := 0; i < 4; i++ {
			onetrend := new(BRZJH.OneTrend)
			onetrend.Type = game.WinTrend[j].t[i].Type
			onetrend.Win = game.WinTrend[j].t[i].Win
			table.Info = append(table.Info, onetrend)
			if onetrend.Win {
				PayCount++
			} else {
				KillCount++
			}
		}
		if PayCount == 4 {
			AllPayCount++
		} else if KillCount == 4 {
			AllKillCount++
		}
		msg.TrendList.TableTrendInfo = append(msg.TrendList.TableTrendInfo, table)
	}

	msg.GameStatus = new(BRZJH.StatusMessage)
	msg.GameStatus.Status = int32(game.Status)
	msg.GameStatus.StatusTime = int32(game.TimerJob.GetTimeDifference())
	msg.RoomID = game.Table.GetRoomID()
	msg.BaseBet = int64(game.Rule.BetList[0])
	msg.UserLimit = game.Rule.UserBetLimit
	msg.AllKill = AllKillCount
	msg.AllPei = AllPayCount
	msg.OddsInfo = int32(game.OddsInfo)
	log.Tracef("房间ID：%d", game.Table.GetRoomID())
	//发送给框架
	//b, _ := proto.Marshal(msg)
	//game.Table.BroadcastAll(int32(rbwar.SendToClientMessageType_RoomSenceInfo), b)
	game.Table.BroadcastAll(int32(BRZJH.SendToClientMessageType_RoomSenceInfo), msg)
}

func (game *Game) shangZhuang(user player.PlayerInterface) {
	//log.Traceln("上庄", user.GetID())
	if user == game.Zhuang {
		return
	}
	if user.GetScore() < game.Rule.ZhuangLimit {
		return
	}
	if user.GetScore() < game.Rule.ZhuangLimit {
		return
	}
	if len(game.ZhuangList) >= config.BRZJHConfig.ShangZhuangPersonNumMax {
		u, ok := game.AllUserList[user.GetID()]
		if ok {
			model.SendBetFailMessage("上庄用户已排满", u)
		}
		return
	}
	for _, v := range game.ZhuangList {
		if v.GetID() == user.GetID() {
			return
		}
	}

	game.ZhuangList = append(game.ZhuangList, user)
	msg := new(BRZJH.ZhuangCount)
	msg.Count = int32(len(game.ZhuangList))
	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_ZhuangUserCount), msg)
	str := fmt.Sprintf("用户id：%v， 携带金币：%v，申请上庄：成功", user.GetID(), user.GetScore())
	game.Table.WriteLogs(user.GetID(), str)
	game.SendZhuangList(nil)
	//发送上庄成功
}

//确定庄
func (game *Game) SetZhuang() {
	if game.Zhuang != nil && game.Zhuang.GetScore() < game.Rule.ZhuangLimit {
		game.Zhuang = nil
	}

	if game.Zhuang == nil || game.LastCount == 0 {
		for {
			if len(game.ZhuangList) > 0 {
				u := game.ZhuangList[0]
				game.ZhuangList = append(game.ZhuangList[:0], game.ZhuangList[1:]...)
				if game.HasUser(u) {
					game.Zhuang = u
					game.LastCount = config.BRZJHConfig.ShangZhuangMax
					game.OnUserStanUp(game.Zhuang)
					break
				}
			} else {
				game.Zhuang = nil
				game.LastCount = 0
				break
			}
		}

		return
	}

	game.LastCount--
}

func (game *Game) HasUser(user player.PlayerInterface) bool {
	_, ok := game.AllUserList[user.GetID()]
	return ok
}

func (game *Game) SendZhuangJiaInfo(user player.PlayerInterface) {
	msg := new(BRZJH.CurrZhuangInfo)
	msg.UserInfo = new(BRZJH.ZhuangInfo)
	if game.Zhuang != nil {
		//发送庄家信息
		game.Rule.Zhuang = game.Zhuang.GetScore()
		msg.Lastcount = int32(game.LastCount)
		msg.UserInfo.Gold = game.Zhuang.GetScore()
		msg.UserInfo.NikeName = game.Zhuang.GetNike()
		msg.UserInfo.UserID = game.Zhuang.GetID()
		msg.UserInfo.Head = game.Zhuang.GetHead()
		msg.UserInfo.ZhuangTotalWin = game.ZhuangTotalWin.TotalWin
	} else {
		game.Rule.Zhuang = 0
		msg.UserInfo.Gold = 900000000
	}

	if user == nil {
		game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_CurrZhuang), msg)
	} else {
		user.SendMsg(int32(BRZJH.SendToClientMessageType_CurrZhuang), msg)
	}
}

func (game *Game) SendZhuangList(user player.PlayerInterface) {
	msg := new(BRZJH.ZhuangListInfo)
	if game.Zhuang != nil {
		msg.CurrZhuangUserInfo = new(BRZJH.CurrZhuangInfo)
		msg.CurrZhuangUserInfo.UserInfo = new(BRZJH.ZhuangInfo)
		msg.CurrZhuangUserInfo.Lastcount = int32(game.LastCount)
		msg.CurrZhuangUserInfo.UserInfo.Gold = game.Zhuang.GetScore()
		msg.CurrZhuangUserInfo.UserInfo.NikeName = game.Zhuang.GetNike()
		msg.CurrZhuangUserInfo.UserInfo.UserID = game.Zhuang.GetID()
		msg.CurrZhuangUserInfo.UserInfo.Head = game.Zhuang.GetHead()
	}

	for _, u := range game.ZhuangList {
		tmp := new(BRZJH.ZhuangInfo)
		tmp.Gold = u.GetScore()
		tmp.NikeName = u.GetNike()
		tmp.UserID = u.GetID()
		tmp.Head = u.GetHead()
		msg.List = append(msg.List, tmp)
	}
	if user == nil {
		game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_ZhuangList), msg)
	} else {
		user.SendMsg(int32(BRZJH.SendToClientMessageType_ZhuangList), msg)
	}
	//user.SendMsg(int32(BRZJH.SendToClientMessageType_ZhuangList), msg)
}

func (game *Game) RobotShangZhuang() {
	if len(game.ZhuangList) > 0 {
		return
	}

	var UserRobot *model.User
	Len := len(game.AllUserList) - 1
	if Len <= 0 {
		return
	}

	for _, u := range game.AllUserList {
		if UserRobot == nil && u.User.IsRobot() {
			UserRobot = u
		} else if u.User.IsRobot() {
			r := rand.Intn(Len)
			if r < 1 {
				UserRobot = u
				break
			}
		}
	}

	if UserRobot != nil {
		game.shangZhuang(UserRobot.User)
	}
}

func (game *Game) XiaZhuang(user player.PlayerInterface) {
	if user == game.Zhuang {
		game.LastCount = 0
		game.SendZhuangJiaInfo(user)
		return
	}

	for i, v := range game.ZhuangList {
		if v == user {
			game.ZhuangList = append(game.ZhuangList[:i], game.ZhuangList[i+1:]...)
			break
		}
	}
	msg := new(BRZJH.ZhuangCount)
	msg.Count = int32(len(game.ZhuangList))
	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_ZhuangUserCount), msg)
	game.SendZhuangList(nil)
}

func (game *Game) Test(buffer []byte) {
	game.HasTest = true
	temp := &BRZJH.TempCardMsg{}
	proto.Unmarshal(buffer, temp)
	for i := 0; i < 5; i++ {
		for j := 0; j < 3; j++ {
			game.SendCard[i].Cards[j] = temp.Cards[i][j]
		}
		game.SendCard[i].Cards = model.SortCards(game.SendCard[i].Cards)
	}
}

func (game *Game) ResetTable() {
	game.Status = 0
	game.Rule.UserBetLimit = 0
	game.RobotTimerJob = nil
	game.LoopBetTimer = nil
}

//跑马灯
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
			if u.User == game.Zhuang {
				continue
			}
			//if u.User == user {
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

//发送 用户总下注
//func (game *Game) SendUserTotalBet(user player.PlayerInterface) {
//	u := game.getUser(user)
//	u.User.SetBetsAmount(u.TotalBet)
//}
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
		if u.User == game.Zhuang {
			continue
		}
		//检测遍历到的用户是否在椅子上，如无此用户 让用户坐下
		if game.SenceSeat.CheckUserOnChair(u.User.GetID()) {
			if game.SenceSeat.SitScene(u, i+1) {
				u.SceneChairId = i + 1
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
	index := 6
	for k, v := range game.CountUserList {
		if k >= index {
			break
		}
		if v.User == game.Zhuang {
			index = index - 1
			continue
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

//循环定时广播各区域下注总额，期间无下注不广播 减少包量
func (game *Game) loopBetMsg() {
	//
	if game.Status != BRZJH.GameStatus_BetStatus {
		return
	}

	//msg := new(BRZJH.BroadBetEnd)
	//msg.BetGold = make([]int64, 4)

	var notSend bool
	for index, v := range game.BetTotal {
		game.BroadTotallBetEnd.BetGold[index] = v + game.TotalRobotBet[index]

		if game.TotalBetTemp[index] == game.BetTotal[index]+game.TotalRobotBet[index] {
			notSend = true
			continue
		}
		notSend = false
		game.TotalBetTemp[index] = game.BetTotal[index] + game.TotalRobotBet[index]
	}
	if notSend {
		return
	}
	game.Table.Broadcast(int32(BRZJH.SendToClientMessageType_BroadEndBet), game.BroadTotallBetEnd)
}

//日志信息
func (game *Game) writeLog() {
	//日志信息
}

//庄家列表用户未下注次数为0
func (game *Game) setZhangListUserNoBetCount() {
	for _, u := range game.ZhuangList {
		if u == game.Zhuang {
			continue
		}
		use, ok := game.AllUserList[u.GetID()]
		if ok {
			use.NoBetCount = 0
		}

	}
}
