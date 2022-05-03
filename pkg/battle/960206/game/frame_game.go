package game

import (
	"encoding/json"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

//UserReady 用户准备
func (game *Game) UserReady(userInter player.PlayerInterface) bool {
	log.Tracef("用户 %d 在房间 %d 准备", userInter.GetID(), game.Id)
	return true
}

//OnActionUserSitDown 用户坐下
//如果符合条件就坐下，返回true，不符合就不让坐下，返回false
func (game *Game) OnActionUserSitDown(userInter player.PlayerInterface, chairID int, configStr string) table.MatchKind {
	user := game.GetUserList(userInter.GetID())

	// 用户断线重联
	if user != nil {
		log.Tracef("用户断线重联，ID：%d， 座位：%d，状态：%d，房间号：%d", user.User.GetID(), user.ChairId, user.Status, game.Id)
		user.User = userInter
		return table.SitDownOk
	}

	// 游戏不是匹配状态不能坐下
	if game.Status != global.TABLE_CUR_STATUS_WAIT {
		log.Warnf("新用户 %d 非匹配状态允许坐下", userInter.GetID())
		return table.SitDownErrorNomal
	}

	// 桌子满了不允许坐下
	if game.GetRoomUserCount() >= 4 {
		log.Warnf("房间 %d 满员，用户 %d 不允许坐下", game.Id, userInter.GetID())
		return table.SitDownErrorNomal
	}

	game.lock.Lock()
	//chairId := int32(game.GetRoomUserCount()) + 1
	chairId := game.GetChairID(userInter.GetID())
	user = data.NewUser(game.Table, chairId)
	game.lock.Unlock()

	user.Table = game.Table
	user.User = userInter
	game.SetUserList(user)
	if !user.User.IsRobot() {
		log.Traceln("用户正常坐下: ", user.User.GetID(), " 房间id：", game.Table.GetID())
	}
	return table.SitDownOk
}

func (game *Game) BindRobot(ai player.RobotInterface) player.RobotHandler {
	user := game.GetUserList(ai.GetID())
	var robot *Robot
	if user == nil {
		log.Warnf("用户 %d 退出时异常，桌上找不到用户", ai.GetID())
		// 机器人坐下
		robot = NewRobot(game, nil)
	} else {
		// 机器人坐下
		robot = NewRobot(game, user)
	}
	robot.AiUser = ai
	return robot
}

// UserOffline
func (game *Game) UserOffline(userInter player.PlayerInterface) bool {

	// 只能在匹配和结束状态才能退出
	if game.Status != global.TABLE_CUR_STATUS_WAIT && game.Status != global.TABLE_CUR_STATUS_END {
		return false
	}

	// todo 当前玩家退出，检查其他人是否都是机器人
	user := game.GetUserList(userInter.GetID())
	if user == nil {
		log.Errorf("用户 %d 退出时异常，桌上找不到用户", userInter.GetID())
		return true
	}

	log.Tracef("用户 %d 离开游戏 %d", userInter.GetID(), game.Id)

	// 清空座位
	game.DelChairID(user.ChairId)

	// 踢出用户
	game.DelUserList(userInter.GetID())
	game.Table.KickOut(userInter)
	return true
}

// OnGameMessage 游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	user := game.GetUserList(userInter.GetID())
	if user == nil {
		log.Traceln("OnGameMessage user nil ")
		return
	}
	switch subCmd {
	// 摆牌请求
	case int32(msg.C2SMsgType_SET_CARDS):
		game.ProcSetCards(buffer, user)
		break

	// 配牌请求
	case int32(msg.C2SMsgType_USER_SELECT_CARDS):
		//game.ProcUserSelectCards(buffer, user)
		break

	default:
		log.Warnf("用户 %d 发送错误请求，subCmd: %d", userInter.GetID(), subCmd)
	}
}

//GameStart 游戏开始
func (game *Game) GameStart() {
	if game.Table.GetID() < 0 {
		log.Traceln("房间id < 0 ")
		return
	}

	// 添加定时器检查 匹配时间后是否满桌，否则匹配机器人
	if game.robotTimer == nil && game.GetRoomUserCount() < 4 {
		game.robotTimer, _ = game.Table.AddTimer(int64(3000), game.MatchRobot)
	}

	// 满足开赛条件
	if game.GetRoomUserCount() == 4 && game.Status == global.TABLE_CUR_STATUS_WAIT {

		game.Table.StartGame()
		game.StartGame()
		return
	}

	return
}

//TableConfig 牌桌配置 底注
type TableConfig struct {
	Bottom_Pouring int64
}

//SendScene 场景消息
func (game *Game) SendScene(userInter player.PlayerInterface) {
	user := game.GetUserList(userInter.GetID())
	if user == nil {
		log.Traceln("SendScene user nil ")
		return
	}

	// add by wd in 2020.3.4 加载房间底注在发场景消息前

	var tableConfig TableConfig
	if err := json.Unmarshal([]byte(game.Table.GetAdviceConfig()), &tableConfig); err != nil {
		log.Traceln("advice 的 值不对： ", game.Table.GetAdviceConfig())
		return
	}
	game.Bottom = tableConfig.Bottom_Pouring

	roomInfo := game.GetRoomInfo2C(user)
	log.Traceln("房间当前状态：uid : ", userInter.GetID(), game.Status)
	for _, v := range roomInfo.UserArr {
		log.Traceln("vvvvv ", v.Uid, v.IsSettleCards)
	}
	_ = user.User.SendMsg(int32(msg.S2CMsgType_ROOM_INFO_RES), roomInfo)

	// 结算阶段重联回来，发送结算信息
	if game.Status == global.TABLE_CUR_STATUS_SETTLE {
		_ = user.User.SendMsg(int32(msg.S2CMsgType_END_GAME), game.NewS2CEndGame())
	}
	return
}

//ResetTable 重置牌桌
func (game *Game) ResetTable() {
	game.timerJob = nil
	game.robotTimer = nil
}

//关闭桌子
func (game *Game) CloseTable() {

}

//用户在线情况下主动离开
func (game *Game) UserLeaveGame(userInter player.PlayerInterface) bool {
	return game.UserOffline(userInter)
}
