package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960203/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// SendGameStatus 发送游戏状态
func (game *WatchBanker) SendGameStatus(gameStatus int32, durationTime int32, userInter player.PlayerInterface) {

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
func (game *WatchBanker) SendSceneInfo(userInter player.PlayerInterface, reConnect bool) {

	var userData []*msg.SeatUserInfoRes

	for id, user := range game.UserList {
		SeatUserInfo := msg.SeatUserInfoRes{
			UserName:    user.Nick,           // 名称
			Head:        user.Head,           // 头像
			Coin:        user.CurAmount,      // 资金
			ChairId:     user.ChairID,        // 座位ID
			Status:      user.Status,         // 玩家状态
			UserId:      id,                  // ID
			BetMultiple: user.BetMultiple,    // 玩家投注下标
			RobIndex:    user.RobIndex,       // 玩家抢庄下标
			IsBanker:    user.IsBanker,       // 是否是庄
			Sex:         user.User.GetSex(),  // 性别
			Address:     user.User.GetCity(), // 地址
		}

		if userInter != nil && userInter.GetID() == user.ID && game.Status >= int32(msg.GameStatus_RobBanker) && game.Status < int32(msg.GameStatus_ShowChards) {
			SeatUserInfo.Cards = user.HoldCards.FirstHalfCards
		}

		// 用户已经摊牌，或者游戏状态大于摊牌结束
		if user.Status >= int32(msg.UserStatus_ShowedCards) ||
			game.Status >= int32(msg.GameStatus_EndShow) ||
			(userInter != nil && userInter.GetID() == user.ID && game.Status >= int32(msg.GameStatus_EndBet)) {
			SeatUserInfo.Cards = user.HoldCards.Cards
			SeatUserInfo.CardsType = int32(user.HoldCards.CardsType)
			SeatUserInfo.CardsIndexs = user.HoldCards.SpecialCardIndexs
		}

		// 用户为庄家，返回抢庄下标
		if SeatUserInfo.IsBanker {
			SeatUserInfo.RobIndex = user.RobIndex
		}

		userData = append(userData, &SeatUserInfo)
	}

	messageResp := msg.SceneMessageRes{
		UserData:   userData,                      // 玩家数据
		GameStatus: game.Status,                   // 游戏状态
		RoomCost:   game.RoomCfg.RoomCost,         // 底分
		MinLimit:   game.RoomCfg.MinLimit,         // 入场限制
		RobIndexs:  game.GameCfg.RobOption,        // 抢庄选项
		Reconnect:  reConnect,                     // 是否重联
		RoomID:     int64(game.Table.GetRoomID()), // roomID
	}

	if game.TimerJob != nil {
		messageResp.StatusTimeLeft = game.TimerJob.GetTimeDifference() / 1000
	}

	if userInter != nil {
		// 用户断线重联后跳过下注结束状态，转为摊牌阶段
		//switch game.Status {
		//case int32(msg.GameStatus_EndRob):
		//	messageResp.GameStatus = int32(msg.GameStatus_BetChips)
		//case int32(msg.GameStatus_EndBet):
		//	messageResp.GameStatus = int32(msg.GameStatus_ShowChards)
		//}

		// 断线重联后发送用户可抢庄选项
		if game.Status >= int32(msg.GameStatus_BetChips) {
			user := game.UserList[userInter.GetID()]
			messageResp.BetMultiple = &msg.BetMultipleRes{
				UserId:          user.ID,
				ChairId:         user.ChairID,
				Multiples:       user.BetMultipleOption,
				HighestMultiple: user.HighestMultiple,
			}
		}

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
func (game *WatchBanker) SendDealCardsMsg(dealResp msg.DealRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送发牌信息：%v", game.Table.GetID(), dealResp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CDeal), &dealResp)
	if err != nil {
		log.Errorf("发送场景消息失败， %v", err.Error())
	}
}

// SendShowCardsMsg 广播摊牌信息
func (game *WatchBanker) SendShowCardsMsg(showCardsResp msg.ShowCardsRes) {

	log.Tracef("游戏 %d 广播摊牌信息：%v", game.Table.GetID(), showCardsResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CShowCards), &showCardsResp)
}

// SendRobBankerInfo 广播用户抢庄信息
func (game *WatchBanker) SendRobBankerInfo(robBankerInfoResp msg.RobBankerInfoRes) {

	log.Tracef("游戏 %d 广播用户抢庄信息：%v", game.Table.GetID(), robBankerInfoResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRobBankerInfo), &robBankerInfoResp)
}

// SendRobBankerResult 广播抢庄结果
func (game *WatchBanker) SendRobBankerResult(robBankerResultResp msg.RobBankerResultRes) {

	log.Tracef("游戏 %d 广播抢庄结果：%v", game.Table.GetID(), robBankerResultResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CRobBankerResult), &robBankerResultResp)
}

// SendBetChipsInfo 广播用户投注信息
func (game *WatchBanker) SendBetChipsInfo(betChipsInfoResp msg.BetChipsInfoRes) {

	log.Tracef("游戏 %d 广播用户投注信息：%v", game.Table.GetID(), betChipsInfoResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CBetChipsInfo), &betChipsInfoResp)
}

// SendSettleResult 广播结算结果
func (game *WatchBanker) SendSettleResult(settleResultResp msg.SettleResultRes) {

	log.Tracef("游戏 %d 广播结算结果：%v", game.Table.GetID(), settleResultResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CSettleResult), &settleResultResp)
}

// SendUserExit 广播用户离开信息
func (game *WatchBanker) SendUserExit(userExitResp msg.UserExitRes) {

	log.Tracef("游戏 %d 广播用户离开信息：%v", game.Table.GetID(), userExitResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserExit), &userExitResp)
}

// SendErrMsg 发送错误消息
func (game *WatchBanker) SendErrMsg(errResp msg.ErrRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送错误消息：%v", game.Table.GetID(), errResp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CErrRes), &errResp)
	if err != nil {
		log.Errorf("发送错误消息失败， %v", err.Error())
	}
}

// SendUserSitDown 广播用户坐下信息
func (game *WatchBanker) SendUserSitDown(userSitDownResp msg.UserSitDownRes) {

	log.Tracef("游戏 %d 广播用户坐下：%v", game.Table.GetID(), userSitDownResp)
	game.Table.Broadcast(int32(msg.SendToClientMessageType_S2CUserSitDown), &userSitDownResp)
}

// SendBetMultipleInfo 发送投注倍率信息
func (game *WatchBanker) SendBetMultipleInfo(resp msg.BetMultipleRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 向玩家 %d 发送投注倍率信息：%v", game.Table.GetID(), resp.UserId, resp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBetMultiple), &resp)
	if err != nil {
		log.Errorf("发送错误消息失败， %v", err.Error())
	}
}
