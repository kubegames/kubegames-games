package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/poker"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/config"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/model"
	rbwar "github.com/kubegames/kubegames-games/pkg/hundreds/960301/pb"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

const (
	RED    = 0
	BLACK  = 1
	ONEHIT = 2
)

type Game struct {
	Table               table.TableInterface  // table interface
	AllUserList         map[int64]*model.User //所有的玩家列表
	Status              rbwar.GameStatus      // 房间状态1 表示
	Win                 int32                 // 1表示红胜利，2表示黑方胜利
	WinCardType         int                   //赢的牌的类型
	LastWinIsRedOrBlack int                   // 最近一次开红还是黑
	RedCards            []byte                // 红方的三张牌
	BlackCards          []byte                // 黑方的三张牌
	IsLuckWin           bool                  // 幸运一击是否胜利
	BetTotal            [3]int64              //红黑一击的下注统计
	TotalUserBet        [3]int64              //红黑一击的下注统计
	SenceSeat           model.SceneInfo       //下注的玩家列表
	TimerJob            *table.Job            //job
	RobotTimerJob       *table.Job            //机器人job
	LastMasterBetType   int32                 //最近一次神算子下注的类型
	LoseCardType        int                   //输的一方的牌型
	WinTrend            []int32               //赢的走势
	WinCardTypeTrend    []int32               //赢的类型走势
	CountUserList       []*model.User         //统计后的玩家列表
	Rule                config.RoomRules      //房间规则信息
	testRed             []byte                //
	testBlack           []byte                //
	CheatValue          int                   //作弊值
	PokerMsg            *rbwar.PokerMsg       //牌消息
	sysCheat            string
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
	game.AllUserList = make(map[int64]*model.User)
	game.SenceSeat.Init()
	game.SenceSeat.Rule = &game.Rule
	game.PokerMsg = new(rbwar.PokerMsg)
	game.PokerMsg.RedPoker = make([]byte, 3)
	game.PokerMsg.BlackPoker = make([]byte, 3)
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
	//删除用户列表用户
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

//用户离开
func (game *Game) UserLeaveGame(user player.PlayerInterface) bool {
	u := game.getUser(user)
	if u.TotalBet != 0 {
		msg := new(rbwar.ExitFail)
		msg.FailReason = "游戏中不能退出！"
		user.SendMsg(int32(rbwar.SendToClientMessageType_ExitRet), msg)
		return false
	}

	if u.SceneChairId != 0 {
		game.OnUserStanUp(user)
	}

	delete(game.AllUserList, user.GetID())
	game.DeleteExitUserFromOnlineUserListSlice(u)
	return true
}

//关闭桌子
func (game *Game) CloseTable() {
	//不做任何操作
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	if int32(rbwar.ReceiveMessageType_BetID) == subCmd {
		game.Bet(buffer, user)
	} else if int32(rbwar.ReceiveMessageType_SitDown) == subCmd {
		//game.UserSitDown(buffer, user)
	} else if int32(rbwar.ReceiveMessageType_GetTrend) == subCmd {
		game.SendTrend(user)
	} else if int32(rbwar.ReceiveMessageType_GetUserListInfo) == subCmd {
		game.SendUserListInfo(user)
	} else if int32(rbwar.ReceiveMessageType_StandUp) == subCmd {
		//game.OnUserStanUp(user)
	} else if int32(rbwar.ReceiveMessageType_tempCard) == subCmd {
		//game.testCard(buffer)
	}
}

func (game *Game) GameStart() {
	game.GetRoomconfig()
	if game.Status == 0 {
		game.Start()
		game.Table.AddTimerRepeat(1000, 0, game.SendRoomInfo)
	} else if game.TimerJob != nil {
		return
	}
	return
}

func (game *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	robot.Init(ai, game)
	return robot
}

func (game *Game) SendScene(user player.PlayerInterface) {
	game.GetRoomconfig()

	u := game.getUser(user)

	game.SendRuleInfo(user)
	game.SendSceneMsg(user)
	game.SendUserBet(u)

	if game.Status >= rbwar.GameStatus_ShowPoker {
		u.User.SendMsg(int32(rbwar.SendToClientMessageType_PokerInfo), game.PokerMsg)
		if game.Status == rbwar.GameStatus_SettleStatus {
			if u.SettleMsg != nil {
				user.SendMsg(int32(rbwar.SendToClientMessageType_UserComeBack), u.SettleMsg)
			}
		}
	}

	if game.TimerJob != nil {
		game.SendToUserStatusMsg(int(game.TimerJob.GetTimeDifference()), user)
	}

	game.SendTrend(user)

	return
}

func (game *Game) Start() {
	//通知框架游戏开始
	game.Table.StartGame()

	//检测下注情况
	game.checkUserBet()

	//选择列表中前6个用户上座
	game.SelectUserListInfoBefore6SitDownChair()

	if game.Table.GetRoomID() <= 0 {
		game.Status = 0
		return
	}

	game.LastMasterBetType = -1

	game.Status = rbwar.GameStatus_StartMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.Startmove), game.StartBet)
	//开始动画消息
	game.SendStatusMsg(config.RBWarConfig.Taketimes.Startmove)
}

func (game *Game) StartBet() {
	game.ResetData()
	//发送开始下注消息
	game.Status = rbwar.GameStatus_BetStatus
	game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.Startbet), game.EndBet)
	game.SendStatusMsg(config.RBWarConfig.Taketimes.Startbet)
	log.Tracef("发送下注消息%d", game.Status)
}

func (game *Game) EndBet() {
	//结束下注
	game.Status = rbwar.GameStatus_EndBetMovie
	game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.Endmove), game.getResult)
	game.SendStatusMsg(config.RBWarConfig.Taketimes.Endmove)
	log.Tracef("发结束下注消息%d", game.Status)
}

//结算
func (game *Game) Settle() {
	//发送结算信息
	game.sendSettleMsg()
	//结算阶段
	game.Status = rbwar.GameStatus_SettleStatus
	game.SendStatusMsg(config.RBWarConfig.Taketimes.Endpay)
	//结束游戏
	game.Table.EndGame()
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
	//添加一下局 开始
	game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.Endpay), game.Start)
}

func (game *Game) SendStatusMsg(StatusTime int) {
	msg := new(rbwar.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	game.Table.Broadcast(int32(rbwar.SendToClientMessageType_Status), msg)
}

func (game *Game) SendToUserStatusMsg(StatusTime int, user player.PlayerInterface) {
	msg := new(rbwar.StatusMessage)
	msg.Status = int32(game.Status)
	msg.StatusTime = int32(StatusTime)
	user.SendMsg(int32(rbwar.SendToClientMessageType_Status), msg)
}

func (game *Game) Bet(buffer []byte, user player.PlayerInterface) {
	if game.Status != rbwar.GameStatus_BetStatus {
		return
	}

	//用户下注
	BetPb := &rbwar.Bet{}
	proto.Unmarshal(buffer, BetPb)
	u := game.getUser(user)

	if u.Bet(BetPb, game.BetTotal) {
		if !u.User.IsRobot() {
			game.TotalUserBet[BetPb.BetType%3] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
		}
		game.BetTotal[BetPb.BetType%3] += int64(game.Rule.BetList[BetPb.BetIndex%int32(len(game.Rule.BetList))])
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
	us := &rbwar.UserSitDown{}
	proto.Unmarshal(buffer, us)
	u, ok := game.AllUserList[user.GetID()]
	if ok {
		if game.SenceSeat.SitScene(u, int(us.ChairNo)) {
			u.SceneChairId = int(us.ChairNo)
			game.SendSceneMsg(nil)
		}
	}
}

// 发牌
func (game *Game) DealPoker() {
	var gp poker.GamePoker
	gp.InitPoker()
	gp.ShuffleCards()
	game.RedCards = make([]byte, 0)
	game.BlackCards = make([]byte, 0)

	for i := 0; i < 3; i++ {
		game.RedCards = append(game.RedCards, gp.DealCards())
		game.BlackCards = append(game.BlackCards, gp.DealCards())
	}

	/*game.BlackCards[0] = 0xd4
	game.BlackCards[1] = 0x54
	game.BlackCards[2] = 0x24
	*/
}

// 比牌， 1表示红胜利，2表示黑方胜利
func (game *Game) CompareRbPoker() {
	Red := game.RedCards
	Black := game.BlackCards
	RedType, SortRed := poker.GetCardTypeJH(Red)
	BlackType, SortBlack := poker.GetCardTypeJH(Black)

	if RedType > BlackType {
		game.Win = 1
		game.WinCardType = RedType
		game.LoseCardType = BlackType
	} else if RedType == BlackType {
		EncodeRed := poker.GetEncodeCard(RedType, SortRed)
		EncodeBlack := poker.GetEncodeCard(BlackType, SortBlack)
		if EncodeRed > EncodeBlack {
			game.Win = 1
			game.WinCardType = RedType
			game.LoseCardType = BlackType
		} else if EncodeRed < EncodeBlack {
			game.Win = 2
			game.LoseCardType = RedType
			game.WinCardType = BlackType
		} else {
			if RedType == poker.CardTypeDZ {
				dui1 := SortRed[0] & 0xf0
				dui2 := SortBlack[0] & 0xf0
				dan1 := SortRed[0]
				dan2 := SortBlack[0]
				if SortRed[1] == SortRed[2] {
					dui1 = SortRed[1] & 0xf0
				} else {
					dan1 = SortRed[2]
				}

				if SortBlack[1] == SortBlack[2] {
					dui2 = SortBlack[1] & 0xf0
				} else {
					dan2 = SortBlack[2]
				}

				if dui1 > dui2 {
					game.Win = 1
					game.WinCardType = RedType
					game.LoseCardType = BlackType
				} else if dui1 < dui2 {
					game.Win = 2
					game.WinCardType = BlackType
					game.LoseCardType = RedType
				} else if dan1 > dan2 {
					game.Win = 1
					game.WinCardType = RedType
					game.LoseCardType = BlackType
				} else {
					game.Win = 2
					game.WinCardType = BlackType
					game.LoseCardType = RedType
				}
			} else if SortRed[0] > SortBlack[0] {
				game.Win = 1
				game.WinCardType = RedType
				game.LoseCardType = BlackType
			} else {
				game.Win = 2
				game.WinCardType = BlackType
				game.LoseCardType = RedType
			}
		}
	} else {
		game.Win = 2
		game.WinCardType = BlackType
		game.LoseCardType = RedType
	}
}

// 判断是否是幸运一击
func (game *Game) LuckOneHit() {
	if game.WinCardType == 2 {
		WinCards := []byte{}
		if game.Win == 1 {
			WinCards = game.RedCards
		} else {
			WinCards = game.BlackCards
		}

		card0, _ := poker.GetCardValueAndColor(WinCards[0])
		card1, _ := poker.GetCardValueAndColor(WinCards[1])
		card2, _ := poker.GetCardValueAndColor(WinCards[2])

		if card0 == card1 && card1 >= 0x90 {
			game.IsLuckWin = true
		} else if card0 == card2 && card0 >= 0x90 {
			game.IsLuckWin = true
		} else if card1 == card2 && card1 >= 0x90 {
			game.IsLuckWin = true
		} else {
			game.IsLuckWin = false
		}

	} else if game.WinCardType == 1 {
		game.IsLuckWin = false
	} else {
		game.IsLuckWin = true
	}
}

//检查用户是否被踢掉
func (game *Game) checkUserBet() {
	for k, u := range game.AllUserList {
		if u.NoBetCount >= (config.RBWarConfig.Unplacebetnum+1) ||
			(u.User.IsRobot() && (u.User.GetScore() > game.Rule.RobotMaxGold || u.User.GetScore() < game.Rule.RobotMinGold)) {
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
	odds := game.getLuckOneHitOdds()

	game.WinTrend = append(game.WinTrend, game.Win)
	game.WinCardTypeTrend = append(game.WinCardTypeTrend, int32(game.WinCardType))

	winlen := len(game.WinTrend)
	if winlen > 100 {
		game.WinTrend = append(game.WinTrend[:(winlen-100-1)], game.WinTrend[(winlen-100):]...)
		game.WinCardTypeTrend = append(game.WinCardTypeTrend[:(winlen-100-1)], game.WinCardTypeTrend[(winlen-100):]...)
	}

	game.CountUserList = make([]*model.User, 0)
	SceneSettleMsg := new(rbwar.SceneUserSettle)

	RedBetCount := 0
	BlackBetCount := 0
	LuckBetCount := 0
	MaxWinGold := int64(0)
	MaxWinUserID := int64(0)
	var SystemWin [3]int64

	RedStr := fmt.Sprintf("红区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[RED]), score.GetScoreStr(game.BetTotal[RED]-game.TotalUserBet[RED]), score.GetScoreStr(game.TotalUserBet[RED]))

	BlackStr := fmt.Sprintf("黑区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[BLACK]), score.GetScoreStr(game.BetTotal[BLACK]-game.TotalUserBet[BLACK]), score.GetScoreStr(game.TotalUserBet[BLACK]))

	LuckStr := fmt.Sprintf("幸运一击区域：总：%v 机器人：%v 真人：%v ",
		score.GetScoreStr(game.BetTotal[ONEHIT]), score.GetScoreStr(game.BetTotal[ONEHIT]-game.TotalUserBet[ONEHIT]), score.GetScoreStr(game.TotalUserBet[ONEHIT]))

	log.Tracef("RedStr %s", RedStr)
	log.Tracef("BlackStr %s", RedStr)
	log.Tracef("LuckStr %s", RedStr)

	//战绩
	var records []*platform.PlayerRecord

	for _, u := range game.AllUserList {
		u.NoBetCount++
		if !u.User.IsRobot() {
			if u.NoBetCount >= (config.RBWarConfig.Unplacebetnum + 1) {
				//发送踢掉用户
				msg := new(rbwar.KickOutUserMsg)
				msg.KickOutReason = "由于您5局未下注，已被踢出房间！"
				log.Tracef("由于用户%d 5局未下注，已被踢出房间！", u.User.GetID())
				u.User.SendMsg(int32(rbwar.SendToClientMessageType_KickOutUser), msg)
			}
		}

		SceneUserInfo := new(rbwar.SceneUserInfo)

		msg := new(rbwar.SettleMsg)

		var win int64
		var totalTax int64 //总税收
		if u.TotalBet > 0 {
			if game.Win == 1 {
				msg.UserWinRed += u.BetRed
				win += u.BetRed * 2
				SceneUserInfo.RedWin = msg.UserWinRed
				SceneUserInfo.BlackWin = -u.BetBlack
				msg.UserWinBlack = -u.BetBlack
				_, Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetRed, u.Table.GetRoomRate())
				_, capital := u.User.SetScore(game.Table.GetGameNum(), u.BetRed, 0)
				Gold += capital
				totalTax += win - Gold
				msg.UserTotalWin += Gold
				SystemWin[RED] += Gold
			} else {
				msg.UserWinBlack += u.BetBlack
				win += u.BetBlack * 2
				SceneUserInfo.RedWin -= u.BetRed
				SceneUserInfo.BlackWin = msg.UserWinBlack
				msg.UserWinRed = -u.BetRed
				_, Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetBlack, u.Table.GetRoomRate())
				_, capital := u.User.SetScore(game.Table.GetGameNum(), u.BetBlack, 0)
				Gold += capital
				totalTax += win - Gold
				msg.UserTotalWin += Gold
				SystemWin[BLACK] += Gold
			}

			msg.LuckWin = u.BetLuck * int64(odds)
			if odds > 0 {
				win += u.BetLuck * int64(odds+1)
				_, Gold := u.User.SetScore(game.Table.GetGameNum(), u.BetLuck*int64(odds), u.Table.GetRoomRate())
				_, capital := u.User.SetScore(game.Table.GetGameNum(), u.BetLuck, 0)
				Gold += capital
				totalTax += u.BetLuck*int64(odds+1) - Gold
				msg.UserTotalWin += Gold
				SystemWin[ONEHIT] += Gold
			}
		}

		if msg.LuckWin != 0 {
			SceneUserInfo.LuckWin = msg.LuckWin
		} else {
			SceneUserInfo.LuckWin -= u.BetLuck
			msg.LuckWin -= u.BetLuck
		}

		msg.IsLuck = game.IsLuckWin

		SceneUserInfo.UserID = int64(u.User.GetID())
		SceneUserInfo.SceneSeatID = int32(u.SceneChairId)
		SceneUserInfo.UserTotalWin = msg.UserTotalWin
		//统计玩家信息
		if (win) > u.TotalBet {
			u.UserCount(true, msg.UserTotalWin)
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
			if u.BetRed != 0 {
				temp += fmt.Sprintf("红：%v ", score.GetScoreStr(u.BetRed))
				temp1 += fmt.Sprintf("红：%v ", score.GetScoreStr(msg.UserWinRed))
				RedBetCount++
			}

			if u.BetBlack != 0 {
				temp += fmt.Sprintf("黑：%v ", score.GetScoreStr(u.BetBlack))
				temp1 += fmt.Sprintf("黑：%v ", score.GetScoreStr(msg.UserWinBlack))
				BlackBetCount++
			}

			if u.BetLuck != 0 {
				temp += fmt.Sprintf("幸运一击：%v ", score.GetScoreStr(u.BetLuck))
				temp1 += fmt.Sprintf("幸运一击：%v ", score.GetScoreStr(msg.LuckWin))
				LuckBetCount++
			}
			temp1 += fmt.Sprintf(" 总输赢：%v，用户剩余金额：%v \r\n", score.GetScoreStr(win-u.TotalBet), score.GetScoreStr(u.User.GetScore()))
			temp += temp1
			log.Tracef("日志%s", temp)
			game.Table.WriteLogs(u.User.GetID(), temp)
		}

		game.PaoMaDeng(msg.UserTotalWin-u.TotalBet, u.User)
		game.CountUser(u)

		msg.Win = int32(game.Win)
		msg.UserScore = u.User.GetScore()
		msg.Red = game.BetTotal[RED]
		msg.Black = game.BetTotal[BLACK]
		msg.Luck = game.BetTotal[ONEHIT]
		msg.UserBetRed = u.BetRed
		msg.UserBetBlack = u.BetBlack
		msg.UserBetLuck = u.BetLuck

		SceneUserInfo.UserScore = msg.UserScore
		SceneUserInfo.BetRed = msg.UserBetRed
		SceneUserInfo.BetBlack = msg.UserBetBlack
		SceneUserInfo.BetLuck = msg.UserBetLuck
		u.User.SendMsg(int32(rbwar.SendToClientMessageType_Settle), msg)
		if u.TotalBet > 0 && !u.User.IsRobot() {
			chip := u.BetRed - u.BetBlack
			if chip < 0 {
				chip = -chip
			}

			chip += u.BetLuck

			u.User.SendChip(chip)
			u.SettleMsg = msg
		} else {
			u.SettleMsg = nil
		}

		user := game.getUser(u.User)
		betsAmount := u.TotalBet
		profitAmount := u.User.GetScore() - user.CruenSorce
		u.ResetUserData()
		//u.User.SendRecord(game.Table.GetGameNum(), profitAmount, betsAmount, totalTax, msg.UserTotalWin, "")
		if !u.User.IsRobot() {
			records = append(records, &platform.PlayerRecord{
				PlayerID:     uint32(u.User.GetID()),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: profitAmount,
				BetsAmount:   betsAmount,
				DrawAmount:   totalTax,
				OutputAmount: msg.UserTotalWin,
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
	//日志信息
	RedStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", RedBetCount, score.GetScoreStr(game.BetTotal[RED]-SystemWin[RED]))
	BlackStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", BlackBetCount, score.GetScoreStr(game.BetTotal[BLACK]-SystemWin[BLACK]))
	LuckStr += fmt.Sprintf("真人数量:%v 输赢：%v\r\n", LuckBetCount, score.GetScoreStr(game.BetTotal[ONEHIT]-SystemWin[ONEHIT]))
	t := ""
	if game.Win == 1 {
		t = "红赢"
	} else {
		t = "黑赢"
	}
	str := fmt.Sprintf("%v作弊率：%v \r\n开局结果 :%v:红区域牌：%v 牌型：%v,", game.sysCheat, game.CheatValue,
		t, model.GetCardString(game.RedCards), poker.GetTypeString(game.RedCards))

	str += fmt.Sprintf("黑区域牌：%v 牌型：%v \r\n",
		model.GetCardString(game.BlackCards), poker.GetTypeString(game.BlackCards))

	str += fmt.Sprintf("系统输赢额度：%v \r\n",
		score.GetScoreStr(game.BetTotal[RED]+game.BetTotal[BLACK]+game.BetTotal[ONEHIT]-SystemWin[RED]-SystemWin[BLACK]-SystemWin[ONEHIT]),
	)
	str += fmt.Sprintf("最高获利用户ID：%v 获得：%v\r\n",
		MaxWinUserID, score.GetScoreStr(MaxWinGold))
	totalstr := RedStr + BlackStr + LuckStr + str
	game.Table.WriteLogs(0, totalstr)

	log.Tracef("结算日志%s", totalstr)
	game.Table.Broadcast(int32(rbwar.SendToClientMessageType_SceneSettleInfo), SceneSettleMsg)
}

func (game *Game) getResult() {
	//关闭配牌
	if len(game.testRed) > 0 {
		game.Status = rbwar.GameStatus_ShowPoker
		game.RedCards = game.testRed
		game.BlackCards = game.testBlack
		game.testRed = make([]byte, 0)
		game.testBlack = make([]byte, 0)
		game.CompareRbPoker()
		game.LuckOneHit()
		//这里给客户端发送的类型除去顺子123和顺金123
		if game.WinCardType > poker.CardTypeSJA23 {
			game.WinCardType -= 2
		} else if game.WinCardType > poker.CardTypeSZA23 {
			game.WinCardType--
		}

		if game.LoseCardType > poker.CardTypeSJA23 {
			game.LoseCardType -= 2
		} else if game.LoseCardType > poker.CardTypeSZA23 {
			game.LoseCardType--
		}

		for i := 0; i < 3; i++ {
			game.PokerMsg.RedPoker[i] = game.RedCards[i]
			game.PokerMsg.BlackPoker[i] = game.BlackCards[i]
		}

		game.PokerMsg.Win = int32(game.Win)
		game.PokerMsg.WinCardType = int32(game.WinCardType)
		game.PokerMsg.LoseCardType = int32(game.LoseCardType)

		game.Table.Broadcast(int32(rbwar.SendToClientMessageType_PokerInfo), game.PokerMsg)
		game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.ShowPoker), game.Settle)
		game.SendStatusMsg(config.RBWarConfig.Taketimes.ShowPoker)
		return
	}

	game.Status = rbwar.GameStatus_ShowPoker
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
	Luck := 0
	r := rand.Intn(10000)
	v := config.RBWarConfig.GetCheatValue(game.CheatValue)
	if v != 0 {
		if r < v {
			eat = 1
		} else {
			out = 1
		}

		if test == -2000 {
			if r < 1000 {
				Luck = 1
			}
		} else if test == -3000 && r > 1500 {
			Luck = 1
		}
	}
	/*if test > 1000 {
		if (test == 2000 && r < 6000) || (test == 3000 && r < 7000) {
			eat = 1
		}
	} else if test < -1000 {
		if (test == -2000 && r < 6000) || (test == -3000 && r < 7000) {
			out = 1
		}

		if test == -2000 {
			if r < 1000 {
				Luck = 1
			}
		} else if test == -3000 && r > 1500 {
			Luck = 1
		}
	}*/

	for {
		game.DealPoker()
		game.CompareRbPoker()
		game.LuckOneHit()

		//这里给客户端发送的类型除去顺子123和顺金123
		if game.WinCardType > poker.CardTypeSJA23 {
			game.WinCardType -= 2
		} else if game.WinCardType > poker.CardTypeSZA23 {
			game.WinCardType--
		}

		TotalMoney := game.TotalUserBet[RED] + game.TotalUserBet[BLACK] + game.TotalUserBet[ONEHIT]
		PayMoney := game.TotalUserBet[RED]
		if game.Win == 2 {
			PayMoney = game.TotalUserBet[BLACK]
		}

		OneHitOdds := game.getLuckOneHitOdds()
		if OneHitOdds > 0 {
			OneHitOdds += 1
		}
		OutMoney := TotalMoney - PayMoney*2 - int64(OneHitOdds)*game.TotalUserBet[ONEHIT]
		if eat == 1 && OutMoney >= 0 {
			//吃分
			break
		} else if out == 1 && OutMoney <= 0 {
			if Luck == 1 && OneHitOdds < 0 {
				continue
			}
			//吐分
			break
		} else if eat == 0 && out == 0 {
			//不控制
			break
		}
	}

	if game.LoseCardType > poker.CardTypeSJA23 {
		game.LoseCardType -= 2
	} else if game.LoseCardType > poker.CardTypeSZA23 {
		game.LoseCardType--
	}

	for i := 0; i < 3; i++ {
		game.PokerMsg.RedPoker[i] = game.RedCards[i]
		game.PokerMsg.BlackPoker[i] = game.BlackCards[i]
	}

	game.PokerMsg.Win = int32(game.Win)
	game.PokerMsg.WinCardType = int32(game.WinCardType)
	game.PokerMsg.LoseCardType = int32(game.LoseCardType)

	game.Table.Broadcast(int32(rbwar.SendToClientMessageType_PokerInfo), game.PokerMsg)
	game.TimerJob, _ = game.Table.AddTimer(int64(config.RBWarConfig.Taketimes.ShowPoker), game.Settle)
	game.SendStatusMsg(config.RBWarConfig.Taketimes.ShowPoker)
}

func (game *Game) getLuckOneHitOdds() int {
	odds := 0
	if game.IsLuckWin == true {
		switch game.WinCardType {
		case int(rbwar.CardsType_DuiZi):
			{
				odds = config.RBWarConfig.Ratio.Duizi
			}
			break
		case int(rbwar.CardsType_ShunZi):
			{
				odds = config.RBWarConfig.Ratio.Shunzi
			}
			break
		case int(rbwar.CardsType_JinHua):
			{
				odds = config.RBWarConfig.Ratio.Jinhua
			}
			break
		case int(rbwar.CardsType_ShunJin):
			{
				odds = config.RBWarConfig.Ratio.Shunjin
			}
			break
		case int(rbwar.CardsType_BaoZi):
			{
				odds = config.RBWarConfig.Ratio.Baozi
			}
			break
		}
	}

	return odds
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
		if user.IsRobot() {
			/*
				robot := new(Robot)
				robotUser, err := game.Table.GetRobot(robot)
				robot.Init(robotUser, game)
				if err != nil {
					log.Tracef(err.Error())
				}
				r := rand.Intn(RConfig.SitDownTime[1]-RConfig.SitDownTime[0]) + RConfig.SitDownTime[0]
				game.RobotTimerJob, _ = game.Table.AddTimer(int64(r), game.RobotSitDown)
				r1 := rand.Intn(RConfig.StandUpTime[1]-RConfig.StandUpTime[0]) + RConfig.StandUpTime[0]
				game.Table.AddTimer(int64(r1), game.RobotStandUp)
			*/
		}
	}

	return u
}

//发送场景消息
func (game *Game) SendSceneMsg(u player.PlayerInterface) {
	msg := new(rbwar.SceneMessage)
	//bigwinner := game.SenceSeat.GetBigWinner()
	//master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(rbwar.SeatUser)
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
		game.Table.Broadcast(int32(rbwar.SendToClientMessageType_SceneID), msg)
	} else {
		u.SendMsg(int32(rbwar.SendToClientMessageType_SceneID), msg)
	}
}

func (game *Game) SendUserBet(u *model.User) {
	msg := new(rbwar.SceneBetInfo)
	msg.Red = game.BetTotal[RED]
	msg.Black = game.BetTotal[BLACK]
	msg.Luck = game.BetTotal[ONEHIT]
	msg.UserBetRed = u.BetRed
	msg.UserBetBlack = u.BetBlack
	msg.UserBetLuck = u.BetLuck
	msg.UserBetTotal = u.TotalBet
	msg.MasterBetType = game.LastMasterBetType
	u.User.SendMsg(int32(rbwar.SendToClientMessageType_BetInfo), msg)
}

func (game *Game) SendTrend(u player.PlayerInterface) {
	msg := new(rbwar.Trend)
	msg.Win = append(msg.Win, game.WinTrend...)
	msg.WinCardType = append(msg.WinCardType, game.WinCardTypeTrend...)
	u.SendMsg(int32(rbwar.SendToClientMessageType_TrendInfo), msg)
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
	msg := new(rbwar.UserList)
	for _, u := range game.CountUserList {
		userinfo := new(rbwar.UserInfo)
		userinfo.NikeName = u.User.GetNike()
		userinfo.UserGlod = u.User.GetScore()
		userinfo.WinCount = int32(u.RetWin)
		userinfo.BetGold = u.RetWinMoney
		userinfo.Head = u.User.GetHead()
		userinfo.Icon = u.Icon
		msg.UserList = append(msg.UserList, userinfo)
	}

	user.SendMsg(int32(rbwar.SendToClientMessageType_UserListInfo), msg)
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

	game.SendSceneMsg(nil)
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
					us := &rbwar.UserSitDown{}
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
	if count < len(RConfig.StandUpProbability) {
		r = rand.Intn(10000)
		if r < RConfig.StandUpProbability[count].Probability {
			if RConfig.StandUpProbability[count].Max == RConfig.StandUpProbability[count].Min {
				r = RConfig.StandUpProbability[count].Max
			} else {
				r = rand.Intn(RConfig.StandUpProbability[count].Max - RConfig.StandUpProbability[count].Min)
			}

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
	msg := new(rbwar.SceneMessage)
	bigwinner := game.SenceSeat.GetBigWinner()
	master := game.SenceSeat.GetMaster()
	for _, v := range game.SenceSeat.SenceSeat {
		su := new(rbwar.SeatUser)
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

	game.Table.Broadcast(int32(rbwar.SendToClientMessageType_SceneID), msg)
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
		panic("解析房间配置失败")
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
	//game.Rule.BetLimit[RED] = int64(BetBase) * 5000
	//game.Rule.BetLimit[BLACK] = int64(BetBase) * 5000
	//game.Rule.BetLimit[ONEHIT] = int64(BetBase) * 2000
	game.Rule.SitDownLimit = BetBase * 100

	level := game.Table.GetLevel()
	game.Rule.RobotMinGold = config.RBWarConfig.Robotgold[level-1][0]
	game.Rule.RobotMaxGold = config.RBWarConfig.Robotgold[level-1][1]
	log.Infof("RobotMinGold %v %v", game.Rule.RobotMinGold, game.Rule.RobotMaxGold)
	game.Rule.SingleUserAllSpaceLimit = config.RBWarConfig.Singleuserallspacelimit5times[level-1]
	game.Rule.AllSpaceLimit = config.RBWarConfig.Allspacelimit5times[level-1]
	for i := 0; i < 3; i++ {
		//log.Traceln(config.BRNNConfig.Singleusersinglespacelimit5times[level-1][3],config.BRNNConfig.Allusersinglespacelimit5times[level-1][i])
		game.Rule.SingleUserSingleSpaceLimit[i] = config.RBWarConfig.Singleusersinglespacelimit5times[level-1][i]
		game.Rule.AllUserSingleSpaceLimit[i] = config.RBWarConfig.Allusersinglespacelimit5times[level-1][i]
	}
	for i := 0; i < 5; i++ {
		game.Rule.BetList = append(game.Rule.BetList, config.RBWarConfig.Chips5times[level-1][i])
	}
	game.Rule.UserBetLimit = game.Rule.SingleUserAllSpaceLimit
	//log.Traceln(game.Rule.BetList,":",game.Rule.SingleUserSingleSpaceLimit,":",game.Rule.SingleUserAllSpaceLimit,":",game.Rule.AllUserSingleSpaceLimit,":",game.Rule.AllSpaceLimit)
}

func (game *Game) SendRuleInfo(u player.PlayerInterface) {
	msg := new(rbwar.RoomRolesInfoMsg)
	for _, v := range game.Rule.BetList {
		msg.BetArr = append(msg.BetArr, int32(v))
	}

	msg.UserBetLimit = game.Rule.UserBetLimit
	msg.BetMinLimit = game.Rule.BetMinLimit

	u.SendMsg(int32(rbwar.SendToClientMessageType_RoomRolesInfo), msg)
}

func (game *Game) SendRoomInfo() {
	if game.Status == 0 {
		return
	}

	msg := new(rbwar.RoomSenceInfoMsg)
	msg.TrendList = new(rbwar.Trend)
	msg.TrendList.Win = append(msg.TrendList.Win, game.WinTrend...)
	msg.TrendList.WinCardType = append(msg.TrendList.WinCardType, game.WinCardTypeTrend...)
	msg.GameStatus = new(rbwar.StatusMessage)
	msg.GameStatus.Status = int32(game.Status)
	msg.GameStatus.StatusTime = int32(game.TimerJob.GetTimeDifference())
	msg.RoomID = int64(game.Table.GetRoomID())
	msg.BaseBet = int64(game.Rule.BetList[0])
	msg.UserLimit = game.Rule.UserBetLimit

	//发送给框架
	//b, _ := proto.Marshal(msg)
	//game.Table.BroadcastAll(int32(rbwar.SendToClientMessageType_RoomSenceInfo), b)
	game.Table.SendToHall(int32(rbwar.SendToClientMessageType_RoomSenceInfo), msg)
}

func (game *Game) testCard(buffer []byte) {
	tmp := &rbwar.TempCardMsg{}
	proto.Unmarshal(buffer, tmp)
	game.testRed = tmp.RedPoker
	game.testBlack = tmp.BlackPoker
}

func (g *Game) PaoMaDeng(Gold int64, user player.PlayerInterface) {
	configs := g.Table.GetMarqueeConfig()
	//log.Debugf("%v", configs)
	for _, v := range configs {
		//log.Debugf("%v %v", Gold, v)
		if Gold >= v.AmountLimit {
			//log.Debugf("创建跑马灯")
			err := g.Table.CreateMarquee(user.GetNike(), Gold, "", v.RuleId)
			if err != nil {
				log.Errorf("创建跑马灯错误：%v", err)
			}
		}
	}
}

func (game *Game) DeleteExitUserFromOnlineUserListSlice(user *model.User) {
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
