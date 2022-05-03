package game

import (
	"encoding/json"
	"fmt"
	"go-game-sdk/define"
	"go-game-sdk/example/game_LaBa/970601/data"
	"go-game-sdk/example/game_LaBa/970601/global"
	"go-game-sdk/example/game_LaBa/970601/msg"
	"go-game-sdk/inter"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

//绑定机器人接口
func (g *Game) BindRobot(ai inter.AIUserInter) player.RobotHandler {
	return nil
}

//UserReady 用户准备
func (game *Game) UserReady(user player.PlayerInterface) bool {
	//log.Traceln("frame >>>>>>> UserReady")

	return true
}

//OnActionUserSitDown 用户坐下
//如果符合条件就坐下，返回true，不符合就不让坐下，返回false
func (game *Game) OnActionUserSitDown(userInter player.PlayerInterface, chairID int, configStr string) int {
	if game.user != nil && game.user.User.GetID() != userInter.GetID() {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	if game.user != nil && game.user.User.GetID() == userInter.GetID() {
		log.Traceln("玩家在彩金游戏断线重连")
		return define.SIT_DOWN_OK
	}
	userStr := userInter.GetTableData()
	var user *data.User
	if userStr != "" {
		log.Traceln("userStr : ", userStr)
		if err := json.Unmarshal([]byte(userStr), &user); err != nil {
			log.Traceln("json.Unmarshal([]byte(use err : ", err)
			user = data.NewUser(game.Table)
		} else {
			if time.Now().Sub(user.LastTime) > 5*time.Minute {
				log.Traceln("user : ", userInter.GetID(), " 超过了 24 小时 ", user.LastTime)
				user = data.NewUser(game.Table)
			}
		}
	} else {
		user = data.NewUser(game.Table)
	}

	user.Table = game.Table
	user.User = userInter
	game.CurBoxNum = user.CurBox
	game.IsIntoCaijin = user.IsIntoSmallGame
	game.level = user.Level
	game.TotalInvest = user.TotalInvest
	if game.CurBoxNum == 0 {
		game.CurBoxNum = global.TOTAL_BOX_COUNT
	}
	if game.level == 0 {
		game.level = 1
	}
	game.SetUserList(user)
	game.HoseLampArr = game.Table.GetMarqueeConfig()
	//game.CurBoxNum = global.TOTAL_BOX_COUNT
	log.Traceln("用户坐下，当前转头：", game.CurBoxNum, game.Table.GetMarqueeConfig())
	return define.SIT_DOWN_OK
}

//UserExit
func (game *Game) UserExit(userInter player.PlayerInterface) bool {
	user := game.GetUserList()
	if user == nil {
		log.Traceln("UserExit 用户没在桌子上 ")
		return true
	}

	if user.IsIntoSmallGame {
		return false
	}
	if game.user.TotalWin != 0 {
		log.Traceln("玩家离开游戏，收益率：", fmt.Sprintf(`%.4f`, float64(game.user.TotalInvestForCount)/float64(game.user.TotalWin)), " id: ", userInter.GetID())
	} else {
		log.Traceln("玩家离开游戏，收益率：", 0)
	}
	game.user.TotalInvestForCount = 0
	user.LastTime = time.Now()
	user.TotalInvest = game.TotalInvest
	game.TotalInvest = 0
	user.CurBox = game.CurBoxNum
	user.Level = game.level
	userStr, _ := json.Marshal(user)
	userInter.SetTableData(string(userStr))
	game.Table.KickOut(userInter)
	game.DelUserList()
	game.IsIntoCaijin = false
	game.IsWin = false
	game.level = 1
	game.CacheScore = 0
	game.CurBoxNum = 15

	return true
}

//OnGameMessage 游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	//log.Traceln("frame >>>>>>> OnGameMessage")
	user := game.GetUserList()
	if user == nil {
		log.Traceln("OnGameMessage user nil ")
		return
	}
	switch subCmd {
	case int32(msg.C2SMsgType_ROOM_INFO):
		log.Traceln(">>>>>>>>>获取房间信息+++")
		game.ProcGetRoomInfo(buffer, user)
	case int32(msg.C2SMsgType_START_GAME):
		log.Traceln(">>>>>>>>>开始+++")
		game.ProcStartGame(buffer, user)
	case int32(msg.C2SMsgType_CHOOSE_CAIJIN):
		log.Traceln(">>>>>>>>>选择彩金宝珠+++")
		game.ProcChooseCaijin(buffer, user)
	case int32(msg.C2SMsgType_NORMAL_QUIT):
		log.Traceln(">>>>>>>>>用户正常离开+++")
		game.ProcNormalQuit(buffer, user)
	case int32(msg.C2SMsgType_TEST_TOOL):
		log.Traceln(">>>>>>>>>测试工具+++")
		//game.ProcTestTool(buffer,user)
	}
}

//GameStart 游戏开始
func (game *Game) GameStart(user player.PlayerInterface) bool {
	//log.Traceln("game start ... ")
	if game.Table.GetID() < 0 || game.user == nil {
		log.Traceln("房间id < 0 ")
		return false
	}
	game.user.Cheat = user.GetProb()
	if game.user.Cheat == 0 {
		game.user.Cheat = game.Table.GetRoomProb()
	}
	rc := game.Table.GetRoomProb()
	log.Traceln("用户作弊率：", game.user.Cheat, "房间作弊率：", rc)

	return true
}

//TableConfig 牌桌配置 底注
type TableConfig struct {
	Bottom_Pouring int64
}

//SendScene 场景消息
func (game *Game) SendScene(userInter player.PlayerInterface) bool {
	user := game.GetUserList()
	if user == nil {
		log.Traceln("SendScene user nil ")
		return false
	}

	var tableConfig TableConfig
	if err := json.Unmarshal([]byte(game.Table.GetAdviceConfig()), &tableConfig); err != nil {
		log.Traceln("advice 的 值不对： ", game.Table.GetAdviceConfig())
		return false
	}
	log.Traceln("配置：", tableConfig.Bottom_Pouring)
	game.Bottom = tableConfig.Bottom_Pouring
	game.Bottom2C = []int64{game.Bottom, game.Bottom * 2, game.Bottom * 3, game.Bottom * 4}
	game.BottomCount2C = []int64{1, 2, 3, 4, 5}

	_ = userInter.SendMsg(int32(msg.S2CMsgType_ROOM_INFO_RES), game.GetRoomInfo2C(user))
	log.Traceln("消息0，", fmt.Sprintf(`%+v`, game.GetRoomInfo2C(user)))
	return true
}

func (game *Game) CloseTable() {

}

//ResetTable 重置牌桌
func (game *Game) ResetTable() {

}

//用户在线情况下主动离开
func (game *Game) LeaveGame(userInter player.PlayerInterface) bool {
	return game.UserExit(userInter)
}
