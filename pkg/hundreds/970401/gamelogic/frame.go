package gamelogic

import (
	"common/log"
	proto "game_LaBa/benzbmw/msg"
	"game_frame_v2/game/inter"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type BenzBMWRoom struct {
}

//初始化桌子
func (bbr *BenzBMWRoom) InitTable(table table.TableInterface) {
	game := NewGame(table)
	table.BindGame(game)
}

func (bbr *BenzBMWRoom) UserExit(user player.PlayerInterface) {
}

func (bbr *BenzBMWRoom) AIUserLogin(user inter.AIUserInter, game table.TableHandler) {
}

// 实现接口
func (game *Game) ResetTable() {
	game.UserBetInfo = [BET_AREA_LENGHT]int64{}
	game.AIBetInfo = [BET_AREA_LENGHT]int64{}
	game.Status = 0
	game.UserMap = make(map[int64]*User, 0)
	game.TimerJob = nil
	game.loopBetTimer = nil
	game.BigWinner = nil
	game.Trend = nil
	game.topUser = nil
	game.settleElems = nil
	game.BetLimitInfo.BaseBet = 0
	if game.settleMsg != nil {
		game.settleMsg.Reset()
	}
}

func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {

	if oldUser := game.UserMap[user.GetId()]; oldUser != nil {
		oldUser.user = user
	} else {
		u := NewUser(game, user)
		if user.IsRobot() {
			rb := NewRobot(game)
			aiUser := user.BindRobot(rb)
			rb.BindUser(aiUser)
		}
		game.UserMap[user.GetId()] = u
	}

	return 1
}

func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(proto.ReceiveMessageType_DoBet):
		game.UserBet(buffer, user)
	case int32(proto.ReceiveMessageType_GetUserList):
		game.getUserList(buffer, user)
	case int32(proto.ReceiveMessageType_GetTrendHistory):
		game.SendTrendMsg(user)

	case int32(proto.ReceiveMessageType_RoundEnd):
		game.AfterSettle(buffer)

	// TODO: 线上环境删测此行代码
	// case int32(proto.ReceiveMessageType_DoTest):
	// 	game.DoTest(buffer)

	case int32(proto.ReceiveMessageType_BetReptReq):
		game.BetRept(buffer, user)
	}
}

func (game *Game) SendScene(user player.PlayerInterface) bool {
	game.initRule()
	u := game.GetUser(user)
	game.SendSceneMsg(user)
	// game.SendTopUserMsg()
	if game.TimerJob != nil {
		// dur := game
		// game.SendStatusMs
		if game.Status != int32(proto.GameStatus_BetStatus) {
			return true
		}
		u.SendStatusMsg(int32(game.TimerJob.GetTimeDifference()))
	}
	return true
}

func (game *Game) UserReady(player.PlayerInterface) bool {
	return true
}

func (game *Game) GameStart(player.PlayerInterface) bool {
	if game.Status == 0 {
		game.Start()
	}
	return true
}

func (game *Game) UserExit(user player.PlayerInterface) bool {
	log.Tracef("user close %d", user.GetId())
	u, ok := game.UserMap[user.GetId()]
	if !ok {
		return true
	}
	//有下注时不让玩家离开
	if !u.user.IsRobot() && u.BetGoldNow != 0 {
		log.Tracef("有下注时不让玩家离开")
		game.leaveUserID = append(game.leaveUserID, user.GetId())
		return false
	} else {
		log.Tracef("玩家离开 %d", user.GetId())
	}
	game.leaveUserID = append(game.leaveUserID, user.GetId())
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u, ok := game.UserMap[user.GetId()]
	if ok {
		if u.BetGoldNow != 0 {
			game.leaveUserID = append(game.leaveUserID, user.GetId())
			return false
		}
		game.leaveUserID = append(game.leaveUserID, user.GetId())
	}
	return true
}

func (game *Game) CloseTable() {
}
