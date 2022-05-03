package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// SendUserSitDown 广播用户坐下信息
func (game *ThreeDoll) SendUserSitDown(userSitDownResp msg.UserSitDownRes) {

	log.Tracef("游戏 %d 广播用户坐下：%v", game.Table.GetID(), userSitDownResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserSitDown), &userSitDownResp)
}

// SendGameStatus 发送游戏状态
func (game *ThreeDoll) SendGameStatus(gameStatus int32, durationTime int32, userInter player.PlayerInterface) {

	resp := msg.StatusMessageRes{
		Status:     gameStatus,
		StatusTime: durationTime,
	}

	if userInter != nil {
		// 发给单个人
		log.Tracef("游戏 %d 发送状态消息：%v", game.Table.GetID(), resp)
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
func (game *ThreeDoll) SendSceneInfo(userInter player.PlayerInterface, reConnect bool) {

	var userDatas []*msg.UserInfo

	for id, user := range game.UserList {
		userInfo := msg.UserInfo{
			UserName:    user.Nick,           // 名称
			Head:        user.Head,           // 头像
			Coin:        user.CurAmount,      // 资金
			ChairId:     user.ChairID,        // 座位ID
			Status:      user.Status,         // 玩家状态
			UserId:      id,                  // ID
			BetMultiple: user.BetMultiple,    // 玩家投注倍数
			IsRob:       user.IsRob,          // 是否抢庄
			IsBanker:    user.IsBanker,       // 是否是庄家
			Address:     user.User.GetCity(), // 地址
			Sex:         user.User.GetSex(),  // 性别
			Multiples:   user.Multiples,      // 可投注选项
		}

		// 用户已经摊牌，或者游戏状态大于摊牌结束
		if user.Status == int32(msg.UserStatus_ShowedCards) || game.Status >= int32(msg.GameStatus_EndBet) {
			userInfo.Cards = user.HoldCards.Cards
			userInfo.CardsType = int32(user.HoldCards.CardsType)
		}

		userDatas = append(userDatas, &userInfo)
	}

	messageResp := msg.SceneMessageRes{
		GameStatus: game.Status,                   // 游戏状态
		RoomCost:   game.RoomCfg.RoomCost,         // 底分
		MinLimit:   game.RoomCfg.MinLimit,         // 入场限制
		RoomID:     int64(game.Table.GetRoomID()), // roomID
		Reconnect:  reConnect,                     // 是否重联
		UserData:   userDatas,                     // 玩家数据
	}

	if game.TimerJob != nil {
		messageResp.StatusTimeLeft = game.TimerJob.GetTimeDifference() / 1000
	}

	if userInter != nil {

		// 发给单个人
		log.Tracef("游戏 %d 发送场景消息：%v", game.Table.GetID(), messageResp)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CSceneMessage), &messageResp)
		if err != nil {
			log.Errorf("发送场景消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播场景消息：%v", game.Table.GetID(), messageResp)
		game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSceneMessage), &messageResp)
	}
}

// SendDealCardsMsg 发送发牌信息
func (game *ThreeDoll) SendDealCardsMsg(dealResp msg.DealRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送发牌信息：%v", game.Table.GetID(), dealResp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CDealCards), &dealResp)
	if err != nil {
		log.Errorf("发送发牌消息失败， %v", err.Error())
	}
}

// SendShowCardsMsg 广播摊牌信息
func (game *ThreeDoll) SendShowCardsMsg(showCardsResp msg.ShowCardsRes) {

	log.Tracef("游戏 %d 广播摊牌信息：%v", game.Table.GetID(), showCardsResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CShowCards), &showCardsResp)
}

// SendRobBankerInfo 广播用户抢庄信息
func (game *ThreeDoll) SendRobBankerInfo(robInfoResp msg.RobInfoRes) {

	log.Tracef("游戏 %d 广播用户抢庄信息：%v", game.Table.GetID(), robInfoResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRobBankerInfo), &robInfoResp)
}

// SendRobBankerResult 广播抢庄结果
func (game *ThreeDoll) SendRobBankerResult(robResultResp msg.RobResultRes) {

	log.Tracef("游戏 %d 广播抢庄结果：%v", game.Table.GetID(), robResultResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRobBankerResult), &robResultResp)
}

// SendBetChipsInfo 广播用户投注信息
func (game *ThreeDoll) SendBetInfo(betInfoResp msg.BetInfoRes) {

	log.Tracef("游戏 %d 广播用户投注信息：%v", game.Table.GetID(), betInfoResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CBetChipsInfo), &betInfoResp)
}

// SendSettleResult 广播结算结果
func (game *ThreeDoll) SendSettleResult(settleResultResp msg.SettleResultRes) {

	log.Tracef("游戏 %d 广播结算结果：%v", game.Table.GetID(), settleResultResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSettleResult), &settleResultResp)
}

// SendUserExit 广播用户离开信息
//func (game *ThreeDoll) SendUserExit(userExitResp msg.UserExitRes) {
//
//	log.Tracef("游戏 %d 广播用户离开信息：%v", game.Table.GetID(), userExitResp)
//	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserExit), &userExitResp)
//}

// SendErrMsg 发送错误消息
//func (game *ThreeDoll) SendErrMsg(errResp msg.ErrRes, userInter player.PlayerInterface) {
//
//	log.Tracef("游戏 %d 发送错误消息：%v", game.Table.GetID(), errResp)
//	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CErrRes), &errResp)
//	if err != nil {
//		log.Errorf("发送错误消息失败， %v", err.Error())
//	}
//}

// SendBetMultipleInfo 发送投注倍率信息
func (game *ThreeDoll) SendBetMultipleInfo(resp msg.BetMultipleRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 向玩家 %d 发送投注倍率信息：%v", game.Table.GetID(), resp.UserId, resp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBetMultiple), &resp)
	if err != nil {
		log.Errorf("发送错误消息失败， %v", err.Error())
	}
}

// SendBetMultipleInfo 发送投注倍率信息
func (game *ThreeDoll) SendGameRecords(resp msg.GameRecordsRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 向玩家 %d 发送投注倍率信息：%v", game.Table.GetID(), userInter.GetID(), resp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBetMultiple), &resp)
	if err != nil {
		log.Errorf("发送错误消息失败， %v", err.Error())
	}
}
