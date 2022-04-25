package game

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/970601/config"
	"go-game-sdk/example/game_LaBa/970601/data"
	"go-game-sdk/example/game_LaBa/970601/msg"

	score2 "github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

//获取房间信息
func (game *Game) ProcGetRoomInfo(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SRoomInfo
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}
	_ = user.User.SendMsg(int32(msg.S2CMsgType_ROOM_INFO_RES), game.GetRoomInfo2C(user))
	return
}

//开始
func (game *Game) ProcStartGame(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SStartGame
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}

	if c2sMsg.Bottom < 0 || int(c2sMsg.Bottom) >= len(game.Bottom2C) {
		fmt.Println("用户传来数据不在数组内：", c2sMsg)
		return
	}
	if c2sMsg.Count < 0 || int(c2sMsg.Count) >= len(game.BottomCount2C) {
		fmt.Println("用户传来数据不在数组内222：", c2sMsg)
		return
	}

	// 检查用户金额等信息
	score := game.Bottom2C[int(c2sMsg.Bottom)] * game.BottomCount2C[int(c2sMsg.Count)]
	fmt.Println("c2sMsg.Bottom,c2sMsg.Count : ", c2sMsg.Bottom, c2sMsg.Count, game.Bottom2C, game.BottomCount2C)
	if score > user.User.GetScore() {
		fmt.Println("用户余额不足：", c2sMsg.Bottom, c2sMsg.Count, user.User.GetScore(), game.Bottom2C[int(c2sMsg.Bottom)], game.Bottom2C[int(c2sMsg.Count)])
		return
	}
	//fmt.Println("投入：",score)
	game.TotalInvest += score
	game.Table.StartGame()
	game.Table.WriteLogs(user.User.GetID(), "开始游戏用户余额："+score2.GetScoreStr(user.User.GetScore()))
	game.Bottom, game.BottomCount = game.Bottom2C[int(c2sMsg.Bottom)], game.BottomCount2C[int(c2sMsg.Count)]
	//fmt.Println("用户原来的金额：",user.User.GetScore(),"扣除的score：",score)
	_, _ = user.User.SetScore(user.Table.GetGameNum(), -score, game.Table.GetRoomRate())
	//fmt.Println("现在的金额：",user.User.GetScore())
	_ = user.User.SendMsg(int32(msg.S2CMsgType_START_GAME_RES), &msg.S2CUserScore{
		Score: user.User.GetScore(),
	})
	user.User.SendChip(score)
	//user.User.SetBetsAmount(score)
	game.StartGame()
	return
}

//彩金游戏
func (game *Game) ProcChooseCaijin(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SChooseCaijin
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}
	if game.user == nil || !game.user.IsIntoSmallGame {
		fmt.Println("玩家为空或者IsIntoSmallGame为false")
		return
	}
	if !rand.RateToExec(game.CheatConfig.IntoCaijin) {
		fmt.Println("没触发彩金，", game.CheatConfig.IntoCaijin)
		return
	}
	game.user.IntoCaijinCount++
	rate0 := game.CheatConfig.Caijin1
	rate1 := rate0 + game.CheatConfig.Caijin2
	rate2 := rate1 + game.CheatConfig.Caijin3
	rate3 := rate2 + game.CheatConfig.Caijin4
	rate4 := rate3 + game.CheatConfig.Caijin5
	index := rand.RandInt(0, 10000)
	fmt.Println("caijin and index : ", rate0, rate1, rate2, rate3, rate4, index, " score：", config.IconConfig.CaijinScore)
	score := game.user.CaijinBase
	if index < rate0 {
		game.user.Caijin1Count++
		score = score * config.IconConfig.CaijinScore[0] / 10000
	}
	if index >= rate0 && index < rate1 {
		game.user.Caijin2Count++
		score = score * config.IconConfig.CaijinScore[1] / 10000
	}
	if index >= rate1 && index < rate2 {
		game.user.Caijin3Count++
		score = score * config.IconConfig.CaijinScore[2] / 10000
	}
	if index >= rate2 && index < rate3 {
		game.user.Caijin4Count++
		score = score * config.IconConfig.CaijinScore[3] / 10000
	}
	if index >= rate3 && index < rate4 {
		game.user.Caijin5Count++
		score = score * config.IconConfig.CaijinScore[4] / 10000
	}
	//fmt.Println("彩金金额：",score)
	game.user.TotalCaijin += score
	_, _ = game.user.User.SetScore(game.Table.GetGameNum(), score, game.Table.GetRoomRate())
	_ = user.User.SendMsg(int32(msg.S2CMsgType_RES_CAIJIN), &msg.S2CCaijin{
		Score: score,
	})
	game.user.TotalWin += score
	game.user.IsIntoSmallGame = false
	game.user.CaijinBase = 0

	if !game.IsShuzhi {
		game.TotalInvest = 0
	}
	game.user.TotalInvest = 0
}

//正常退出游戏
func (game *Game) ProcNormalQuit(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SChooseCaijin
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}
	user.IsNormalQuit = true
	_ = user.User.SendMsg(int32(msg.S2CMsgType_NORMAL_QUIT_RES), &msg.S2CQuit{})
}

//测试工具
func (game *Game) ProcTestTool(buffer []byte, user *data.User) {
	var c2sMsg msg.C2STestTool
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}
	game.CheatConfig = config.GetLhdbConfig(game.user.Cheat)
	fmt.Println("测试工具传过来的：", c2sMsg.Icon, c2sMsg.Count)
	game.IsTest = true
	var i int32
	game.InitIcons(true)
	for y, _ := range game.Icons {
		for x, _ := range game.Icons[y] {
			game.Icons[x][y] = 0
		}
	}
	for y, _ := range game.Icons {
		for x, _ := range game.Icons[y] {
			game.Icons[x][y] = c2sMsg.Icon
			i++
			if i >= c2sMsg.Count {
				goto BreakLoop
			}
		}
	}
BreakLoop:
	fmt.Println("跳出循环")
	game.PrintIcons()
	for y, _ := range game.Icons {
		for x, _ := range game.Icons[y] {
			if game.Icons[x][y] == 0 {
				game.Icons[x][y] = game.GetDifferentIcon(x, y)
			}
		}
	}

}
