package game

import (
	"encoding/json"

	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

func (game *Game) UserReady(user player.PlayerInterface) bool {
	//if game.CurStatus == global.TABLE_CUR_STATUS_ING {
	//	return false
	//}
	return true
}

//用户坐下
//如果符合条件就坐下，返回true，不符合就不让坐下，返回false
func (game *Game) OnActionUserSitDown(userInter player.PlayerInterface, chairId int, config string) table.MatchKind {
	user := game.GetUserListMap(userInter.GetID())
	if user != nil && !user.IsLeave {
		//if user != nil && !user.IsLeave && user.CurStatus != global.USER_CUR_STATUS_GIVE_UP {
		log.Traceln("分配房间成功,用户掉线重新连回来", user.Id, "chair id : ", user.ChairId, " 用户状态：", user.CurStatus, "房间号：", game.TableId)
		user.User = userInter
		if !userInter.IsRobot() {
			_ = userInter.SendMsg(global.S2C_INTO_ROOM, game.GetTableInfo(user))
		}
		return table.SitDownOk
	}
	if game.CurStatus == global.TABLE_CUR_STATUS_ING || game.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD {
		//userInter.SendMsg(global.S2C_INTO_ROOM, game.GetTableInfo(user))
		return table.SitDownErrorNomal
	}
	if game.GetTableUserCount() >= 5 {
		return table.SitDownErrorNomal
	}
	var tableConfig TableConfig
	if err := json.Unmarshal([]byte(config), &tableConfig); err != nil {
		log.Warnln("advice 的 值不对 1111 ： ", game.Table.GetAdviceConfig())
		return table.SitDownErrorOver
	}
	if userInter.GetScore() < tableConfig.Bottom_Pouring {
		log.Traceln("玩家金币不够底注：", userInter.GetScore(), "  ", tableConfig.Bottom_Pouring)
		//return define.SIT_DOWN_ERROR_NORMAL
		return table.SitDownErrorOver
	}
	//log.Traceln("user id : ", userInter.GetID(), " 坐下的牌桌：：： ", game.Table.GetID(), "房间人数：", len(game.userListMap))
	user = data.NewUser(userInter.GetID(), userInter.GetNike(), userInter.IsRobot())
	user.User = userInter
	user.Score = userInter.GetScore()
	if !user.User.IsRobot() {
		//log.Traceln("用户 ",userInter.GetID(),"钱：",user.Score)
	}
	user.Table = game.Table
	game.AddUserIntoTable(user)

	return table.SitDownOk
}

func (game *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	user := game.GetUserListMap(ai.GetID())
	user.InitAiCharacter(*game.GameConfig)
	robot := NewRobot(game, user)
	robot.AiUser = ai
	return robot
}

func (game *Game) UserOffline(userInter player.PlayerInterface) bool {
	user := game.GetUserListMap(userInter.GetID())
	if user == nil {
		userInter.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.C2SIntoGame{})
		return true
	}
	if user.CurStatus == global.USER_CUR_STATUS_WAIT_START && (game.CurStatus == global.TABLE_CUR_STATUS_WAIT || game.CurStatus == global.TABLE_CUR_STATUS_MATCHING) {
		game.Table.KickOut(user.User)
		game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo(false))
		for i, v := range game.userListArr {
			if v != nil && v.User.GetID() == user.User.GetID() {
				//log.Traceln("用户在匹配阶段离开")
				game.userListArr[i] = nil
				return true
			}
		}
	}
	//if game.timerJob.GetTimeDifference() <= 1000 {
	//	log.Traceln("倒计时时间 小于 1000 不退出 ")
	//	return false
	//}
	if game.CurStatus == global.TABLE_CUR_STATUS_ING || game.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD || game.CurStatus == global.TABLE_CUR_STATUS_SYSTEM_COMPARE {
		//if user.CurStatus == global.USER_CUR_STATUS_ING && (game.CurStatus == global.TABLE_CUR_STATUS_ING || game.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD || game.CurStatus == global.TABLE_CUR_STATUS_SYSTEM_COMPARE) {
		log.Traceln("用户未弃牌，不能离开房间")
		userInter.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.S2CRoomInfo{})
		return false
	}

	game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo(false))
	//if !user.IsAllIn{
	//game.DelUserListMap(user.Id)
	//}
	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	//log.Traceln("on message : ",subCmd)
	user := game.GetUserListMap(userInter.GetID())
	if user == nil && subCmd != global.C2S_LEAVE_TABLE {
		//log.Traceln("OnGameMessage user nil ", userInter.GetID())
		return
	}
	switch subCmd {
	case global.C2S_START_GAME:
		//log.Traceln("玩家点击开始游戏+++")
		game.ProcUserStartGame(buffer, user)
	case global.C2S_USER_ACTION:
		//////log.Traceln("用户发言+++")
		game.ProcAction(buffer, user)
	case global.C2S_COMPARE_CARDS:
		//log.Traceln("比牌+++")
		game.ProcCompare(buffer, user)
	case global.C2S_GET_CAN_COMPARE_LIST:
		//log.Traceln("获取可比牌用户列表+++")
		game.ProcGetCanCompareList(buffer, user)
	case global.C2S_SENDCARD_OVER:
		////log.Traceln("客户端发牌动画结束+++")
		game.ProcSendCardOver(buffer, user)
	case global.C2S_SET_CARD_TYPE:
		return
		//log.Traceln("------------设置牌型------------")
		//game.ProcSetCardType(buffer, user)
	case global.C2S_SEE_OTHTER_CARDS:
		return
		//log.Traceln("------------工具，查看其他玩家的牌------------")
		//game.ProcSeeOtherCards(buffer, user)
	case global.C2S_LEAVE_TABLE:
		//log.Traceln("客户端离开房间、重开比赛、继续游戏+++")
		if user == nil {
			//log.Traceln("房间没在游戏中，不退出 111 ")
			userInter.SendMsg(global.ERROR_CODE_GAME_NOT_ING, &msg.C2SIntoGame{})
			return
		}
		game.ProcLeaveGame(buffer, user)
	}
}

type TableConfig struct {
	Bottom_Pouring int64
}

func (game *Game) GameStart() {
	countUserWait := 0
	for _, user := range game.userListArr {
		if user != nil && user.CurStatus != global.USER_CUR_STATUS_ING {
			countUserWait++
		}
	}
	if countUserWait >= 2 && game.CurStatus == global.TABLE_CUR_STATUS_WAIT {
		game.CurStatus = global.TABLE_CUR_STATUS_MATCHING
		game.Table.AddTimer(2*1000, func() {
			game.InitStartGameConfig()
			game.InitCardTypeCountMap()
			game.StartGame()
		})
		return
	}
	return

}

func (game *Game) SendScene(userInter player.PlayerInterface) {
	user := game.GetUserListMap(userInter.GetID())
	if user == nil {
		log.Warnf("user send scene 没在场景里面")
		return
	}

	//初始化用户
	if game.CurStatus != global.TABLE_CUR_STATUS_ING {
		var tableConfig TableConfig
		if err := json.Unmarshal([]byte(game.Table.GetAdviceConfig()), &tableConfig); err != nil {
			log.Warnf("advice 的 值不对： %s", game.Table.GetAdviceConfig())
			return
		}
		//level := int(game.Table.GetLevel())
		game.GameConfig = new(config.GameConfig)

		game.GameConfig.MinAction = tableConfig.Bottom_Pouring
		game.GameConfig.RaiseAmount = []int64{tableConfig.Bottom_Pouring * 2, tableConfig.Bottom_Pouring * 5, tableConfig.Bottom_Pouring * 10}
		game.GameConfig.AiTouJi = 10      //config.GameConfigArr[level-1].AiTouJi
		game.GameConfig.AiWenZhong = 30   //config.GameConfigArr[level-1].AiWenZhong
		game.GameConfig.AiJiJin = 10      //config.GameConfigArr[level-1].AiJiJin
		game.GameConfig.AiZhengChang = 50 //config.GameConfigArr[level-1].AiZhengChang
		game.GameConfig.MaxAllIn = tableConfig.Bottom_Pouring * 400
		game.GameConfig.MaxRound = 20
		game.MinAction = game.GameConfig.MinAction
	}
	tableInfo := game.GetTableInfo(user)
	if !userInter.IsRobot() {
		_ = userInter.SendMsg(global.S2C_INTO_ROOM, tableInfo)
	}
	if game.GetTableUserCount() >= 2 && game.CurStatus == global.TABLE_CUR_STATUS_MATCHING {
		_ = userInter.SendMsg(global.S2C_WAIT_START, &msg.S2CTickerStart{Ticker: 1})
	}
	//给其他玩家发送有玩家进来的消息
	if game.CurStatus == global.TABLE_CUR_STATUS_WAIT || game.CurStatus == global.TABLE_CUR_STATUS_MATCHING {
		for _, v := range game.userListArr {
			if v != nil && v.User.GetID() != userInter.GetID() {
				_ = v.User.SendMsg(global.S2C_OTHER_INTO_ROOM, user.GetUserMsgInfo(false))
			}
		}
	}
	return
}

//通知关闭桌子
func (game *Game) CloseTable() {

}

//用户在线情况下主动离开
func (game *Game) UserLeaveGame(userInter player.PlayerInterface) bool {
	user := game.GetUserListMap(userInter.GetID())
	if user == nil {
		log.Warnf("user exit 用户为空")
		userInter.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.C2SIntoGame{})
		return true
	}
	if user.CurStatus == global.USER_CUR_STATUS_WAIT_START && (game.CurStatus == global.TABLE_CUR_STATUS_WAIT || game.CurStatus == global.TABLE_CUR_STATUS_MATCHING) {
		game.Table.KickOut(user.User)
		game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo(false))
		for i, v := range game.userListArr {
			if v != nil && v.User.GetID() == user.User.GetID() {
				log.Warnf("用户在匹配阶段离开")
				game.userListArr[i] = nil
				return true
			}
		}
	}

	if user.CurStatus == global.USER_CUR_STATUS_ING && (game.CurStatus == global.TABLE_CUR_STATUS_ING || game.CurStatus == global.TABLE_CUR_STATUS_START_SEND_CARD || game.CurStatus == global.TABLE_CUR_STATUS_SYSTEM_COMPARE) {
		log.Warnf("用户未弃牌，不能离开房间")
		userInter.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.S2CRoomInfo{})
		return false
	}

	game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo(false))
	log.Warnf("玩家：", user.Id, " 主动离开房间")
	//if !user.IsAllIn{
	game.DelUserListMap(user.Id)
	return true
}
