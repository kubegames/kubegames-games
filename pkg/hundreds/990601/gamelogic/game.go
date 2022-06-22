package gamelogic

import (
	bridanimal "github.com/kubegames/kubegames-games/pkg/slots/990601/msg"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"

	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type LaBaRoom struct {
}

type BetLimit struct {
	UserLimit    int64
	BetAreaLimit [BET_AREA_LENGTH]int64
}

//初始化桌子
func (lbr *LaBaRoom) InitTable(table table.TableInterface) {
	g := new(Game)
	g.InitTable(table)
	table.BindGame(g)
}

func (lbr *LaBaRoom) UserExit(user player.PlayerInterface) {
}

func (lbr *LaBaRoom) AIUserLogin(user player.RobotInterface, game table.TableHandler) {
}

func (g *Game) InitTable(table table.TableInterface) {
	g.Table = table
	// g.BirdAnimals = deepcopy.Copy(config.BirdAnimaConfig.BirdAnimals).(model.Elements)
	// g.rmShark()
	g.UserInfoList = make(map[int64]*UserInfo)
	g.TotalBet = [BET_AREA_LENGTH]int64{}
	g.AITotalBet = [BET_AREA_LENGTH]int64{}
	g.TotalBetTemp = [BET_AREA_LENGTH]int64{}
}

//用户坐下
func (g *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	// g.GetUserByUserID(user.GetID(), user)
	// g.SendUserNumMsg()
	return 1
}

func (g *Game) UserExit(user player.PlayerInterface) bool {

	u := g.GetUserByUserID(user.GetID(), user)
	//有下注时不让玩家离开
	if u.Totol != 0 {
		log.Tracef("有下注时不让玩家离开")
		return false
	} else {
		log.Tracef("玩家离开 %d", user.GetID())
	}

	u.ResetUserData()
	// g.userLock.Loc()
	// defer g.userLock.Unloc()
	g.userLock.Lock()
	delete(g.UserInfoList, user.GetID())
	g.userLock.Unlock()
	// for i, v := range g.UserList {
	// 	if user.GetID() == v.User.GetID() {
	// 		g.UserList = append(g.UserList[:i], g.UserList[i:]...)
	// 	}
	// }
	g.SendUserNumMsg()
	return true
}

func (g *Game) LeaveGame(user player.PlayerInterface) bool {
	u, ok := g.UserInfoList[user.GetID()]
	if ok {
		if u.Totol != 0 {
			// msg := new(bridanimal.BetFailMsg)
			// msg.BetFailInfo = "游戏中不能退出！"
			// TODO: 玩家离开消息
			// user.SendMsg(int32(bridanimal.SendToClientMessageType_UserComeBack), msg)
			return false
		}
		g.userLock.Lock()
		u.ResetUserData()
		delete(g.UserInfoList, user.GetID())
		g.userLock.Unlock()
		g.SendUserNumMsg()

	}
	return true
}

//游戏消息
func (g *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(bridanimal.ReceiveMessageType_BetID):
		g.OnUserBet(buffer, user)
	case int32(bridanimal.ReceiveMessageType_GetUserListInfo):
		g.SendUserListInfo(buffer, user)
	// TODO: 删除
	// case int32(bridanimal.ReceiveMessageType_Test):
	// 	g.handleTestMsg(buffer)
	case int32(bridanimal.ReceiveMessageType_BetReptID):
		g.repeatBet(buffer, user)
	}
}

func (g *Game) UserReady(user player.PlayerInterface) bool {
	g.SendUserNumMsg()
	return true
}
func (game *Game) CloseTable() {
}

func (g *Game) SendScene(user player.PlayerInterface) bool {
	g.InitRoomRule()
	g.SetBetArr()
	g.GetUserByUserID(user.GetID(), user)
	g.SendSceneMsg(user)
	if g.TimerJob != nil {
		if g.Status != int32(bridanimal.GameStatus_BetStatus) {
			return true
		}
		g.SendToUserStatusMsg(int(g.TimerJob.GetTimeDifference()), user)
	}
	g.SendUserNumMsg()
	return true
}

func (g *Game) GameStart(user player.PlayerInterface) bool {
	// g.InitRoomRule()
	// g.SetBetArr()
	if g.Status == 0 {
		g.RandOdds()
	} else if g.TimerJob != nil {
		// g.TimerJob=clock.
	}
	return true
}

func (g *Game) ResetTable() {
	g.UserInfoList = make(map[int64]*UserInfo, 0)
	g.Status = 0
	g.TimerJob = nil
	g.LoopBetTimer = nil
	g.BetArr = nil
	g.TotalBet = [BET_AREA_LENGTH]int64{}
	g.AITotalBet = [BET_AREA_LENGTH]int64{}
	g.TotalBetTemp = [BET_AREA_LENGTH]int64{}
	g.Trend = nil
	g.settleMsg = nil
	g.settleElements = nil
}
