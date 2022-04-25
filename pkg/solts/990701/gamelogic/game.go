package gamelogic

import (
	"fmt"
	powergame "go-game-sdk/example/game_LaBa/990701/msg"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/inter"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

//用户回来的消息
type UserRebackInfo struct {
	FreeGameTimes int
	FreeGameGold  int64
	LastBetGold   int64
}

type LaBaRoom struct {
}

//初始化桌子
func (lbr *LaBaRoom) InitTable(table table.TableInterface) {
	//log.Tracef("init table num %d", table.GetID())
	g := new(Game)
	g.InitTable(table)
	g.Init(&config.LBConfig)
	table.BindGame(g)
}

func (lbr *LaBaRoom) UserExit(user player.PlayerInterface) {
}

func (lbr *LaBaRoom) AIUserLogin(user inter.AIUserInter, game table.TableHandler) {
}

func (g *Game) InitTable(table table.TableInterface) {
	g.table = table
}

func (g *Game) CloseTable() {
	if g.user != nil {
		if g.FreeGameTimes != 0 {
			str := fmt.Sprintf("%v,%v,%v", g.FreeGameTimes, g.FreeGameGold, g.LastBet)

			g.user.SetTableData(str)
			g.FreeGameTimes = 0
		}
		g.user.SendRecord(g.table.GetGameNum(), g.user.GetScore()-g.curr, g.AllBet*int64(g.Line), 0, g.UserTotalWin, "")
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(g.user.GetScore())))
		g.curr = g.user.GetScore()
		g.table.KickOut(g.user)
		g.table.EndGame()
	}
}

//用户坐下
func (g *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	g.FreeGameTimes = 0
	g.LastBet = 0
	g.AllBet = 0
	//g.SendSence()
	return 1
}

func (g *Game) UserExit(user player.PlayerInterface) bool {
	if g.FreeGameTimes != 0 {
		str := fmt.Sprintf("%v,%v,%v", g.FreeGameTimes, g.FreeGameGold, g.LastBet)

		user.SetTableData(str)
		g.FreeGameTimes = 0
	}

	user.SendRecord(g.table.GetGameNum(), user.GetScore()-g.curr, g.AllBet*int64(g.Line), 0, g.UserTotalWin, "")
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(user.GetScore())))
	g.table.EndGame()
	g.curr = user.GetScore()
	return true
}

func (g *Game) LeaveGame(user player.PlayerInterface) bool {
	if g.FreeGameTimes != 0 {
		str := fmt.Sprintf("%v,%v,%v", g.FreeGameTimes, g.FreeGameGold, g.LastBet)

		user.SetTableData(str)
		g.FreeGameTimes = 0
	}
	user.SendRecord(g.table.GetGameNum(), user.GetScore()-g.curr, g.AllBet*int64(g.Line), 0, g.UserTotalWin, "")
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(user.GetScore())))
	g.table.EndGame()
	g.curr = user.GetScore()
	return true
}

//游戏消息
func (g *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(powergame.MsgIDC2S_Bet):
		g.OnUserBet(buffer)
		break
	case int32(powergame.MsgIDC2S_AskSence):
		g.SendScene(user)
		break

	case int32(powergame.MsgIDC2S_Test):
		//fmt.Println("测试", buffer)
		//g.handleTest(buffer)
	}
}

func (g *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

//场景消息
func (g *Game) SendScene(user player.PlayerInterface) bool {
	g.user = user
	g.UserTotalWin = 0
	g.curr = user.GetScore()
	g.GetRoomconfig()
	g.GetRebackInfo()
	senddata := new(powergame.Sence)
	senddata.BetValue = append(senddata.BetValue, g.BetArr...)
	senddata.Gold = user.GetScore()
	if g.FreeGameTimes != 0 {
		senddata.FreeGameTimes = int32(g.FreeGameTimes)
		senddata.FreeGameGold = g.FreeGameGold

		for i := 0; i < len(g.BetArr); i++ {
			if g.BetArr[i] == int32(g.LastBet) {
				senddata.LastBetIndex = int32(i)
				break
			}
		}
	} else {
		senddata.LastBetIndex = 0
	}

	user.SendMsg(int32(powergame.ReMsgIDS2C_SenceID), senddata)
	g.table.StartGame()
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏开始金币:", score.GetScoreStr(user.GetScore())))

	return true
}

func (g *Game) GameStart(user player.PlayerInterface) bool {
	return true
}

func (g *Game) ResetTable() {

}
