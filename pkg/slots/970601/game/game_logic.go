package game

import (
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/internal/pkg/score"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/global"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
)

//单独写一个方法，将game.icons 赋值出去单独运算一边

func (game *Game) StartGame() {
	//game.Table.StartGame()
	//t1 := time.Now()
	if game.IsShuzhi {

	} else {
		game.user.Cheat = game.user.User.GetProb()
		if game.user.Cheat == 0 {
			game.user.Cheat = game.Table.GetRoomProb()
		}
	}
	game.user.TotalInvestForCount += game.Bottom * game.BottomCount
	rc := game.Table.GetRoomProb()
	if rc == 0 {
		game.Table.WriteLogs(0, "系统作弊率为0，使用1000作为系统作弊率")
		rc = 1000
	}
	log.Traceln("用户作弊率222222：", game.user.Cheat, "房间作弊率：", rc, "用户投入：", game.user.TotalInvestForCount, "收益：", game.user.TotalWin)
	game.CheatConfig = config.GetLhdbConfig(game.user.Cheat)
	//log.Traceln("游戏作弊率配置：",fmt.Sprintf(`%+v`,game.CheatConfig))
	game.user.CaijinBase += game.Bottom * game.BottomCount
	//log.Traceln("StartGame ::: ",fmt.Sprintf(`%+v`,game.CheatConfig))
	//初始化图标数组
	if !game.IsTest {
		game.InitIcons(true)
	}
	game.IsTest = false
	game.InitTopIcons(true)
	//game.Icons[2][2] = game.GetRandIcon(true,0,0)
	//game.Icons = [][]int32{
	//	{3,2,4,2},
	//	{2,2,2,2},
	//	{2,1,1,2},
	//	{4,4,4,4},
	//}
	game.InitEndGame2C()
	//开始检测钥匙、是否中奖等
	game.Ergodic(1)
	if game.CurBoxNum == 0 {
		if game.level == 1 {
			game.user.Level1Count++
		}
		if game.level == 2 {
			game.user.Level2Count++
		}
		if game.level == 3 {
			game.user.Level3Count++
		}
		game.level++
		if game.level > 3 {
			game.EndGameInfo.IsIntoSmallGame = true
			game.user.IsIntoSmallGame = true
			game.IsIntoCaijin = true
			if game.IsShuzhi {
				buf, _ := proto.Marshal(&msg.C2SChooseCaijin{Index: 1})
				game.ProcChooseCaijin(buf, game.user)
			}
			game.level = 1
		}
		game.CurBoxNum = global.TOTAL_BOX_COUNT
	}
	game.EndGameInfo.NextLevel = game.level
	//log.Traceln("开始游戏耗时：", time.Now().Sub(t1))
	//log.Traceln("-----------------")
	//log.Traceln("topIcons: ", game.EndGameInfo.TopIcons)
	//log.Traceln("score : ", game.EndGameInfo.Score)
	game.EndGame()
	//game.PrintIcons()
	//log.Traceln("arr topIcons: ",game.EndGameInfo.AllWinInfoArr[0].TopIcons)
}

//结算玩家赢钱，并上分
func (game *Game) EndGame() {
	winSerial := 0
	logStr := fmt.Sprintf(`第%d关，用户id：%d，投入金额：%.2f`, game.level, game.user.User.GetID(), float64(game.Bottom*game.BottomCount)/100)
	for _, v := range game.EndGameInfo.AllWinInfoArr {
		//log.Traceln("-----------第", v.Serial, "中奖-----------")
		if v.DisappearInfoArr[0].Count >= 4 {
			winSerial++
		}
		for _, dInfo := range v.DisappearInfoArr {
			//log.Traceln("        -----------中奖数量：", dInfo.Count, "中奖金额：", dInfo.WinScore, "消除的图标：", dInfo.WinAxis[0].Value)
			logStr += fmt.Sprintf(`中奖物品：%s，相连%d，赢得金额:%.2f / `, game.getIconName(dInfo.WinAxis[0].Value), dInfo.Count, float64(dInfo.WinScore)/100)
		}
	}
	if game.EndGameInfo.AllWinInfoArr == nil || len(game.EndGameInfo.AllWinInfoArr) == 0 {
		logStr += " 没有中奖 "
	}
	logStr += fmt.Sprintf(`用户作弊率:%d `, game.user.Cheat)
	if game.user.User.GetProb() == 0 {
		logStr += " 使用血池"
	} else {
		logStr += " 被点控"
	}

	//log.Traceln("总中奖金额：",game.EndGameInfo.WinScore,"中奖次数：",winSerial,totalSerial)
	//game.user.TotalWinNoKey = game.user.TotalWinNoKey + winSerial
	game.user.TotalWinSerial += winSerial
	//---------统计信息---------
	switch {
	case winSerial == 1:
		game.user.WinSerial1++
	case winSerial == 2:
		game.user.WinSerial2++
	case winSerial == 3:
		game.user.WinSerial3++
	case winSerial == 4:
		game.user.WinSerial4++
	case winSerial == 5:
		game.user.WinSerial5++
	case winSerial > 5:
		game.user.WinSerialOver5++
	}
	//if game.user.Times > game.CheatConfig.MaxTimes {
	//	game.user.Times1150++
	//}else {
	switch {
	case game.user.Times >= 1 && game.user.Times <= 10:
		game.user.Times110++
	case game.user.Times >= 11 && game.user.Times <= 50:
		game.user.Times1150++
	case game.user.Times >= 51 && game.user.Times <= 200:
		game.user.Times51200++
	case game.user.Times >= 201 && game.user.Times <= 500:
		game.user.Times201500++
	case game.user.Times >= 501 && game.user.Times <= 1000:
		game.user.Times5011000++
	case game.user.Times >= 1001 && game.user.Times <= 2000:
		game.user.Times10012000++
	case game.user.Times >= 2001 && game.user.Times <= 5000:
		game.user.Times20015000++
	case game.user.Times >= 5001:
		game.user.TimesOver5000++
	}
	//}

	//---------统计信息---------
	if game.EndGameInfo.WinScore > game.user.Invest {
		game.user.WinNum++
	} else if game.EndGameInfo.WinScore < game.user.Invest {
		game.user.LoseNum++
	} else {
		game.user.PeaceNum++
	}
	//game.user.Invest = 0

	game.user.TotalWin += game.EndGameInfo.WinScore
	if game.EndGameInfo.WinScore > 0 {
		game.user.WinCount++
		taxScore := game.EndGameInfo.WinScore * game.Table.GetRoomRate() / 10000
		game.user.User.SetScore(game.Table.GetGameNum(), game.EndGameInfo.WinScore, game.Table.GetRoomRate())
		//game.user.User.SendRecord(game.Table.GetGameNum(), game.EndGameInfo.WinScore-game.BottomCount*game.Bottom, game.BottomCount*game.Bottom, taxScore, game.EndGameInfo.WinScore, "")
		game.Table.UploadPlayerRecord([]*platform.PlayerRecord{
			{
				PlayerID:     uint32(game.user.User.GetID()),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: game.EndGameInfo.WinScore - game.BottomCount*game.Bottom,
				BetsAmount:   game.BottomCount * game.Bottom,
				DrawAmount:   taxScore,
				OutputAmount: game.EndGameInfo.WinScore,
				Balance:      game.user.User.GetScore(),
				UpdatedAt:    time.Now(),
				CreatedAt:    time.Now(),
			},
		})
	} else {
		//game.user.User.SendRecord(game.Table.GetGameNum(), -game.BottomCount*game.Bottom, game.BottomCount*game.Bottom, 0, 0, "")
		game.Table.UploadPlayerRecord([]*platform.PlayerRecord{
			{
				PlayerID:     uint32(game.user.User.GetID()),
				GameNum:      game.Table.GetGameNum(),
				ProfitAmount: -game.BottomCount * game.Bottom,
				BetsAmount:   game.BottomCount * game.Bottom,
				DrawAmount:   0,
				OutputAmount: 0,
				Balance:      game.user.User.GetScore(),
				UpdatedAt:    time.Now(),
				CreatedAt:    time.Now(),
			},
		})
	}
	logStr += " 玩家余额：" + score.GetScoreStr(game.user.User.GetScore())
	//log.Traceln("日志：：：", logStr)
	game.Table.WriteLogs(game.user.User.GetID(), logStr)
	game.TriggerHorseLamp(game.EndGameInfo.WinScore)

	game.EndGameInfo.Score = game.user.User.GetScore()
	game.EndGameInfo.TotalInvest = game.TotalInvest
	//log.Traceln("投入：",game.EndGameInfo.TotalInvest,"产出：",game.EndGameInfo.WinScore,game.EndGameInfo.WinScore,game.Bottom,game.BottomCount)
	game.user.User.SendMsg(int32(msg.S2CMsgType_END_GAME_RES), game.EndGameInfo)
	//踢掉用户，重置牌桌
	game.EndGameInfo = nil
	game.CacheScore = 0
	game.AllWinInfoCache = nil
	game.MaxTimes = 0
	game.MaxDisCount = 0
	game.user.Times = 0
	//框架结束比赛
	game.Table.EndGame()
}

//发送当局游戏结束时产生的日志信息
func (game *Game) writeEndGameLog() {

}
