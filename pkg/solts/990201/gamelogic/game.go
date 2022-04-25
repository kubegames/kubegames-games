package gamelogic

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/990201/bibei"
	shz "go-game-sdk/example/game_LaBa/990201/msg"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/example/game_LaBa/labacom/xiaomali"
	"go-game-sdk/inter"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"

	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type LaBaRoom struct {
}

//用户回来的消息
type UserRebackInfo struct {
	LittleGameTimes int
	LittleGameGold  int64
	LastBetGold     int64
}

//初始化桌子
func (lbr *LaBaRoom) InitTable(table table.TableInterface) {
	//log.Tracef("init table num %d", table.GetId())
	g := new(Game)
	g.InitTable(table)
	g.Init(&config.LBConfig, &xiaomali.XMLConfig, &bibei.BBConfig)
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
		if g.XiaoMaLiTimes != 0 {
			str := fmt.Sprintf("%v,%v,%v", g.XiaoMaLiTimes, g.XiaoMaLiGold, g.LastBet)
			g.user.SetTableData(str)

			g.XiaoMaLiTimes = 0
		}
		g.user.SendRecord(g.table.GetGameNum(), g.user.GetScore()-g.curr, g.AllBet*int64(g.lbcfg.LineCount), 0, g.UserTotalWin, "")
		g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(g.user.GetScore())))
		g.curr = g.user.GetScore()
		g.table.KickOut(g.user)
		g.table.EndGame()
	}
}

//用户坐下
func (g *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, config string) int {
	g.XiaoMaLiTimes = 0
	g.LastBet = 0
	g.AllBet = 0
	return 1
}

func (g *Game) UserExit(user player.PlayerInterface) bool {
	if g.XiaoMaLiTimes != 0 {
		str := fmt.Sprintf("%v,%v,%v", g.XiaoMaLiTimes, g.XiaoMaLiGold, g.LastBet)
		user.SetTableData(str)

		g.XiaoMaLiTimes = 0
	}
	user.SendRecord(g.table.GetGameNum(), user.GetScore()-g.curr, g.AllBet*int64(g.lbcfg.LineCount), 0, g.UserTotalWin, "")
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(user.GetScore())))
	g.table.EndGame()
	g.curr = user.GetScore()
	return true
}

func (g *Game) LeaveGame(user player.PlayerInterface) bool {
	if g.XiaoMaLiTimes != 0 {
		str := fmt.Sprintf("%v,%v,%v", g.XiaoMaLiTimes, g.XiaoMaLiGold, g.LastBet)
		user.SetTableData(str)

		g.XiaoMaLiTimes = 0
	}

	user.SendRecord(g.table.GetGameNum(), user.GetScore()-g.curr, g.AllBet*int64(g.lbcfg.LineCount), 0, g.UserTotalWin, "")
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏结束金币:", score.GetScoreStr(user.GetScore())))
	g.table.EndGame()
	g.curr = user.GetScore()
	return true
}

//游戏消息
func (g *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	log.Tracef("开始执行：%v", subCmd)
	switch subCmd {
	case int32(shz.MsgIDC2S_Bet):
		g.OnUserBet(buffer)
		break
	case int32(shz.MsgIDC2S_XiaMaLi):
		g.XiaoMaLi(buffer)
		break
	case int32(shz.MsgIDC2S_BiBei):
		g.UserBiBei(buffer)
		break
	case int32(shz.MsgIDC2S_Test):
		//g.handleTest(buffer)
	}
	log.Tracef("结束执行：%v", subCmd)
}

func (g *Game) UserReady(user player.PlayerInterface) bool {
	return true
}

func (g *Game) SendScene(user player.PlayerInterface) bool {
	g.user = user
	g.UserTotalWin = 0
	g.GetRoomconfig()
	g.GetRebackInfo()
	g.curr = user.GetScore()
	senddata := new(shz.Sence)
	senddata.BetValue = append(senddata.BetValue, g.BetArr...)
	senddata.Gold = user.GetScore()
	if g.XiaoMaLiTimes != 0 {
		senddata.XiaoMaLiTimes = int32(g.XiaoMaLiTimes)
		senddata.XiaoMaLiGold = g.XiaoMaLiGold
		for i := 0; i < len(g.BetArr); i++ {
			if g.BetArr[i] == int32(g.LastBet) {
				senddata.BetIndex = int32(i)
				break
			}
		}
	} else {
		senddata.BetIndex = 0
	}

	user.SendMsg(int32(shz.ReMsgIDS2C_SenceID), senddata)
	g.table.StartGame()
	g.table.WriteLogs(g.user.GetID(), fmt.Sprintln("游戏开始金币:", score.GetScoreStr(user.GetScore())))
	return true
}

func (g *Game) GameStart(user player.PlayerInterface) bool {
	return true
}

func (g *Game) ResetTable() {

}
