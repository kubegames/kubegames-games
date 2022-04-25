package gamelogic

import (
	proto "go-game-sdk/example/game_LaBa/970501/msg"
	"go-game-sdk/inter"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"

	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Room struct {
}

//初始化桌子
func (bbr *Room) InitTable(table table.TableInterface) {
	game := NewGame(table)
	table.BindGame(game)
}

func (bbr *Room) UserExit(user player.PlayerInterface) {
}

func (bbr *Room) AIUserLogin(user inter.AIUserInter, game table.TableHandler) {
}

// 实现接口
func (game *Game) ResetTable() {
	game.UserBetInfo = [BET_AREA_LENGHT]int64{}
	game.AIBetInfo = [BET_AREA_LENGHT]int64{}
	game.Status = 0
	game.UserMap = make(map[int64]*User, 0)
	game.TimerJob = nil
	game.loopBetTimer = nil
	game.Trend = nil
	game.BetArr = nil
	game.settleElems = nil
	// game.settleMsg = nil
}

func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	u := game.GetUser(user)
	u.user = user
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

	// TODO: 线上环境删测此行代码
	case int32(proto.ReceiveMessageType_DoTest):
		// game.DoTest(buffer)
		// 重复下注
	case int32(proto.ReceiveMessageType_BetReptReq):
		game.BetRept(buffer, user)

	case int32(proto.ReceiveMessageType_RoundEnd):
		game.RoundEnd()
	case int32(proto.ReceiveMessageType_TopUserReq):
		game.SendTop3User(user)
	case int32(proto.ReceiveMessageType_BackInReq):
		game.leftTime(user)
	}
}

func (game *Game) SendScene(user player.PlayerInterface) bool {
	game.initRule()
	u := game.GetUser(user)
	u.user = user
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
	log.Tracef("user close %d", user.GetID())
	u := game.GetUser(user)
	game.deleteUserID = append(game.deleteUserID, user.GetID())
	//有下注时不让玩家离开
	if u.BetGoldNow != 0 {
		log.Tracef("有下注时不让玩家离开")
		return false
	} else {
		log.Tracef("玩家离开 %d", user.GetID())
	}
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	u, ok := game.UserMap[user.GetID()]
	game.deleteUserID = append(game.deleteUserID, user.GetID())
	if ok {
		if u.BetGoldNow != 0 {
			// msg := new(bridanimal.BetFailMsg)
			// msg.BetFailInfo = "游戏中不能退出！"
			// TODO: 玩家离开消息
			// user.SendMsg(int32(bridanimal.SendToClientMessageType_UserComeBack), msg)
			return false
		}
	}
	return true
}

func (game *Game) CloseTable() {

}
