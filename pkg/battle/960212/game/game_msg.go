package game

import (
	"fmt"

	"github.com/kubegames/kubegames-games/pkg/battle/960212/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960212/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// SendGameStatus 发送游戏状态
func (game *DouDizhu) SendGameStatus(gameStatus int32, durationTime int32, userInter player.PlayerInterface) {
	resp := msg.StatusMessageRes{
		Status:     gameStatus,
		StatusTime: durationTime,
	}

	if userInter != nil {
		// 发给单个人
		log.Tracef("游戏 %d 发送玩家 %d 状态消息：%v", game.Table.GetID(), userInter.GetID(), fmt.Sprintf("%+v\n", resp))
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CGameStatus), &resp)

		if err != nil {
			log.Errorf("发送状态消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播状态消息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
		game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CGameStatus), &resp)
	}
}

// SendSceneInfo 发送场景消息
func (game *DouDizhu) SendSceneInfo(userInter player.PlayerInterface, reConnect bool) {

	var userData []*msg.SeatUserInfoRes

	for id, user := range game.UserList {
		SeatUserInfo := msg.SeatUserInfoRes{
			UserName: user.Nick,
			Head:     user.Head,
			Coin:     user.CurAmount,
			ChairId:  user.ChairID,
			Status:   user.Status,
			UserId:   id,
			Sex:      user.User.GetSex(),
			Address:  user.User.GetCity(),
			IsDizhu:  user.IsDizhu,
			RobNum:   user.RobNum,
			AddNum:   user.AddNum,
			CardsLen: int64(len(user.Cards)),
		}
		if userInter != nil && userInter.GetID() == id {
			SeatUserInfo.Cards = user.Cards
		}

		// 用户操作过，发送上一次操作结果
		if len(user.PutCardsRecords) != 0 {
			SeatUserInfo.LastCards = user.PutCardsRecords[len(user.PutCardsRecords)-1]
			SeatUserInfo.IsActioned = true
		}

		userData = append(userData, &SeatUserInfo)
	}

	// 公共倍数 = 抢分倍数 * 地分倍数 * 炸弹倍数 * 火箭倍数 * 春天倍数 * 反春倍数
	commonMultiple := game.CurRobNum *
		game.BottomMultiple *
		game.BoomMultiple *
		game.RocketMultiple *
		game.AllOffMultiple *
		game.BeAllOffMultiple

	resp := msg.SceneMessageRes{
		UserData:       userData,
		GameStatus:     game.Status,
		RoomCost:       game.RoomCfg.RoomCost,
		MinLimit:       game.RoomCfg.MinLimit,
		Reconnect:      reConnect,
		RoomID:         int64(game.Table.GetRoomID()),
		RoomLevel:      game.RoomCfg.Level,
		CommonMultiple: commonMultiple,
		CurrentPlayer: &msg.CurrentPlayerRes{
			UserId:     game.CurrentPlayer.UserID,
			ChairId:    game.CurrentPlayer.ChairID,
			ActionTime: int32(game.CurrentPlayer.ActionTime) / 1000,
			Permission: game.CurrentPlayer.Permission,
			ActionType: game.CurrentPlayer.ActionType,
		},
	}

	if game.Chairs[game.curRobberChairID] != nil {
		resp.CurrentRobber = &msg.CurrentRobberRes{
			UserId:     game.Chairs[game.curRobberChairID].ID,
			ChairId:    int32(game.curRobberChairID),
			ActionTime: int32(game.TimeCfg.RobTime) / 1000,
			CurrentNum: game.CurRobNum,
		}
	}

	// 当前状态大于或等于确认地主，发送底牌
	if game.Status >= int32(msg.GameStatus_confirmDizhuStatus) {
		resp.BottomCards = game.bottomCards
	}

	if game.TimerJob != nil {
		resp.StatusTimeLeft = game.TimerJob.GetTimeDifference() / 1000
	}

	if userInter != nil {

		// 重联时获取记牌器剩余牌
		leftCards := []*msg.LeftCards{}
		repeatedArr := poker.CheckRepeatedCards(game.LeftCards)
		for i, arr := range repeatedArr {
			for _, value := range arr {
				leftCards = append(leftCards, &msg.LeftCards{
					Count:     int32(i) + 1,
					CardValue: []byte{value},
				})
			}
		}
		resp.LeftCards = leftCards

		// 发给单个人
		log.Tracef("游戏 %d 发送玩家 %d 场景消息：%v", game.Table.GetID(), userInter.GetID(), fmt.Sprintf("%+v\n", resp))
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CSceneMessage), &resp)
		if err != nil {
			log.Errorf("发送场景消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播场景消息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
		game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSceneMessage), &resp)
	}
}

// SendCurrentRobber 广播当前抢地主玩家
func (game *DouDizhu) SendCurrentRobber(chairID int32) {
	user := game.Chairs[chairID]

	resp := msg.CurrentRobberRes{
		UserId:     user.ID,
		ChairId:    user.ChairID,
		ActionTime: int32(game.TimeCfg.RobTime) / 1000,
		CurrentNum: game.CurRobNum,
	}

	log.Tracef("游戏 %d 广播当前抢庄玩家：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CCurrentRobber), &resp)
}

// SendRobResult 广播抢地主请求响应
func (game *DouDizhu) SendRobResult(resp msg.RobResultRes) {

	log.Tracef("游戏 %d 广播玩家抢地主反馈：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRobResult), &resp)
}

// SendConfirmDizhu 广播确认地主消息
func (game *DouDizhu) SendConfirmDizhu(resp msg.ConfirmDizhuRes) {

	log.Tracef("游戏 %d 广播确认地主消息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CConfirmDizhu), &resp)
}

// SendRedoubleResult 广播加倍结果
func (game *DouDizhu) SendRedoubleResult(resp msg.RedoubleResultRes) {

	log.Tracef("游戏 %d 广播加倍结果消息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRedoubleResult), &resp)
}

// SendSceneInfo 发送发牌消息
func (game *DouDizhu) SendDealInfo(dealResp msg.DealRes, userInter player.PlayerInterface) {

	//log.Tracef("游戏 %d 发送玩家 %d 发牌信息：%v", game.Table.GetID(), userInter.GetID(), dealResp)
	//err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CDeal), &dealResp)
	//if err != nil {
	//	log.Errorf("发送发牌消息失败， %v", err.Error())
	//}

	log.Tracef("游戏 %d 广播玩家 %d 发牌信息：%v", game.Table.GetID(), userInter.GetID(), dealResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CDeal), &dealResp)
}

// SendCurrentPlayer 广播当前玩家信息
func (game *DouDizhu) SendCurrentPlayer() {

	resp := msg.CurrentPlayerRes{
		UserId:     game.CurrentPlayer.UserID,
		ChairId:    game.CurrentPlayer.ChairID,
		ActionTime: int32(game.CurrentPlayer.ActionTime / 1000),
		Permission: game.CurrentPlayer.Permission,
		ActionType: game.CurrentPlayer.ActionType,
	}

	log.Tracef("游戏 %d 广播当前操作玩家：%s", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CCurrentPlayer), &resp)
}

// SendPutInfo 广播出牌结果信息
func (game *DouDizhu) SendPutInfo(resp msg.PutInfoRes) {

	log.Tracef("游戏 %d 广播出牌结果信息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CPutInfo), &resp)
}

// SendHangUpInfo 广播玩家托管操作
func (game *DouDizhu) SendHangUpInfo(userID int64) {
	user := game.UserList[userID]
	var isHangUp bool

	switch user.Status {
	case int32(msg.UserStatus_UserHangUp):
		isHangUp = true
		break
	case int32(msg.UserStatus_UserNormal):
		isHangUp = false
		break
	}

	resp := msg.HangUpInfoRes{
		UserId:   user.ID,
		ChairId:  user.ChairID,
		IsHangUp: isHangUp,
	}

	log.Tracef("游戏 %d 广播玩家托管操作信息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", resp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CHangUpInfo), &resp)
}

// SendTipsInfo 发送提示信息
func (game *DouDizhu) SendTipsInfo(resp msg.TipsRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送玩家 %d 提示信息：%v", game.Table.GetID(), userInter.GetID(), fmt.Sprintf("%+v\n", resp))
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CTips), &resp)
	if err != nil {
		log.Errorf("发送提示消息失败， %v", err.Error())
	}
}

// SendSettleInfo 广播结算信息
func (game *DouDizhu) SendSettleInfo() {
	var settleList []*msg.SettleResult // 结算列表

	for _, user := range game.UserList {

		// 农民倍数
		peasantsMultiple := user.AddNum
		if user.IsDizhu {
			peasantsMultiple = game.TotalPeasantsMultiple
		}

		settleList = append(settleList, &msg.SettleResult{
			UserId:           user.ID,
			ChairId:          user.ChairID,
			IsDizhu:          user.IsDizhu,
			Result:           user.SettleResult,
			PeasantsMultiple: peasantsMultiple,
			TotalMultiple:    user.TotalMultiple,
			LeftCards:        user.Cards,
			UserName:         user.Nick,
		})
	}

	// 公共倍数 = 抢分倍数 * 地分倍数 * 炸弹倍数 * 火箭倍数 * 春天倍数 * 反春倍数
	commonMultiple := game.CurRobNum *
		game.BottomMultiple *
		game.BoomMultiple *
		game.RocketMultiple *
		game.AllOffMultiple *
		game.BeAllOffMultiple

	SettleInfoResp := msg.SettleInfoRes{
		RobMultiple:      game.CurRobNum,
		AllOffMultiple:   game.AllOffMultiple,
		BeAllOffMultiple: game.BeAllOffMultiple,
		BottomMultiple:   game.BottomMultiple,
		BoomMultiple:     game.BoomMultiple,
		RocketMultiple:   game.RocketMultiple,
		CommonMultiple:   commonMultiple,
		DizhuMultiple:    game.Dizhu.AddNum,
		ResultList:       settleList,
	}

	log.Tracef("游戏 %d 广播结算信息：%v", game.Table.GetID(), fmt.Sprintf("%+v\n", SettleInfoResp))
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSettleInfo), &SettleInfoResp)
}
