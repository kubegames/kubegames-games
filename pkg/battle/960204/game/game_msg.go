package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// SendGameStatus 发送游戏状态
func (game *RunFaster) SendGameStatus(gameStatus int32, durationTime int32, userInter player.PlayerInterface) {
	resp := msg.StatusMessageRes{
		Status:     gameStatus,
		StatusTime: durationTime,
	}

	if userInter != nil {
		// 发给单个人
		log.Tracef("游戏 %d 发送玩家 %d 状态消息：%v", game.Table.GetID(), userInter.GetID(), resp)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CGameStatus), &resp)
		if err != nil {
			log.Errorf("发送状态消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播状态消息：%v", game.Table.GetID(), resp)
		game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CGameStatus), &resp)
	}
}

// SendSceneInfo 发送场景消息
func (game *RunFaster) SendSceneInfo(userInter player.PlayerInterface, reConnect bool) {

	var userData []*msg.SeatUserInfoRes

	for id, user := range game.UserList {
		SeatUserInfo := msg.SeatUserInfoRes{
			UserName: user.Nick,
			Head:     user.Head,
			Coin:     user.CurAmount,
			ChairId:  user.ChairID,
			Status:   user.Status,
			UserId:   id,
			CardsLen: int32(len(user.Cards)),
			Sex:      user.User.GetSex(),
			Address:  user.User.GetCity(),
		}
		if userInter != nil && userInter.GetID() == id {
			SeatUserInfo.Cards = user.Cards
		}
		if len(user.PutCardsRecords) != 0 {
			SeatUserInfo.LastCards = user.PutCardsRecords[len(user.PutCardsRecords)-1]
		}

		userData = append(userData, &SeatUserInfo)
	}

	resp := msg.SceneMessageRes{
		UserData:         userData,
		GameStatus:       game.Status,
		RoomCost:         game.RoomCfg.RoomCost,
		MinLimit:         game.RoomCfg.MinLimit,
		Reconnect:        reConnect,
		RoomID:           int64(game.Table.GetRoomID()),
		RoomLevel:        game.RoomCfg.Level,
		StandardNullCard: []byte{poker.NULL_CARD},
		CurrentPlayer: &msg.CurrentPlayerRes{
			UserId:     game.CurrentPlayer.UserID,
			ChairId:    game.CurrentPlayer.ChairID,
			ActionTime: int32(game.CurrentPlayer.ActionTime) / 1000,
			Permission: game.CurrentPlayer.Permission,
			ActionType: game.CurrentPlayer.ActionType,
		},
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
		log.Tracef("游戏 %d 发送玩家 %d 场景消息：%v", game.Table.GetID(), userInter.GetID(), resp)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CSceneMessage), &resp)
		if err != nil {
			log.Errorf("发送场景消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播场景消息：%v", game.Table.GetID(), resp)
		game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSceneMessage), &resp)
	}
}

// SendSceneInfo 发送发牌消息
func (game *RunFaster) SendDealInfo(dealResp msg.DealRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送玩家 %d 发牌信息：%v", game.Table.GetID(), userInter.GetID(), dealResp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CDeal), &dealResp)
	if err != nil {
		log.Errorf("发送发牌消息失败， %v", err.Error())
	}

	//log.Tracef("游戏 %d 广播玩家 %d 发牌信息：%v", game.Table.GetID(), userInter.GetID(), dealResp)
	//game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CDeal), &dealResp)

}

// SendCurrentPlayer 广播当前玩家信息
func (game *RunFaster) SendCurrentPlayer(resp msg.CurrentPlayerRes) {

	log.Tracef("游戏 %d 广播当前操作玩家：%v", game.Table.GetID(), resp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CCurrentPlayer), &resp)
}

// SendTipsInfo 发送提示信息
func (game *RunFaster) SendTipsInfo(resp msg.TipsRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送玩家 %d 提示信息：%v", game.Table.GetID(), userInter.GetID(), resp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CTips), &resp)
	if err != nil {
		log.Errorf("发送提示消息失败， %v", err.Error())
	}
}

// SendPutInfo 广播出牌结果信息
func (game *RunFaster) SendPutInfo(resp msg.PutInfoRes) {

	log.Tracef("游戏 %d 广播出牌结果信息：%v", game.Table.GetID(), resp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CPutInfo), &resp)
}

// SendSettleInfo 广播结算信息
func (game *RunFaster) SendSettleInfo(resp msg.SettleInfoRes) {

	log.Tracef("游戏 %d 广播结算信息：%v", game.Table.GetID(), resp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSettleInfo), &resp)
}

// SendShowCardsInfo 广播摊牌信息
func (game *RunFaster) SendShowCardsInfo(resp msg.ShowCardsRes) {

	log.Tracef("游戏 %d 广播摊牌信息：%v", game.Table.GetID(), resp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CShowCards), &resp)
}

// SendUserExitInfo 广播玩家离开信息
func (game *RunFaster) SendUserExitInfo(resp msg.UserExitRes) {

	log.Tracef("游戏 %d 广播玩家离开信息：%v", game.Table.GetID(), resp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserExit), &resp)
}

// SendUserSitDown 广播用户坐下信息
func (game *RunFaster) SendUserSitDown(userSitDownResp msg.UserSitDownRes) {

	log.Tracef("游戏 %d 广播用户坐下：%v", game.Table.GetID(), userSitDownResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserSitDown), &userSitDownResp)
}

// SendUserStatus 广播用户状态状态
func (game *RunFaster) SendUserStatus(userStatusResp msg.UserStatusRes) {

	log.Tracef("游戏 %d 广播用户状态消息：%v", game.Table.GetID(), userStatusResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserStatus), &userStatusResp)
}
