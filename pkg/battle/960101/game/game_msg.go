package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// SendGameStatus 发送游戏状态
func (game *Blackjack) SendGameStatus(gameStatus int32, durationTime int32, userInter player.PlayerInterface) {

	resp := msg.StatusMessageRes{
		Status:     gameStatus,
		StatusTime: durationTime,
	}

	if userInter != nil {
		// 发给单个人
		log.Tracef("游戏 %d 发送状态消息：%v", game.table.GetID(), resp)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CStatus), &resp)
		if err != nil {
			log.Errorf("发送状态消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播状态消息：%v", game.table.GetID(), resp)
		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CStatus), &resp)
	}
}

// SendSceneInfo 发送场景消息
func (game *Blackjack) SendSceneInfo(userInter player.PlayerInterface, reConnect bool) {
	var userDatas []*msg.SeatUserInfoRes

	for _, user := range game.AllUserList {

		// 全局用户数据
		userData := data.GetUserInterdata(user.User)

		userDataRes := &msg.SeatUserInfoRes{
			UserName:     user.UserName,
			Head:         user.Head,
			Coin:         user.CurAmount,
			ChairId:      user.ChairID,
			IsBuyInsure:  user.IsBuyInsure,
			UserId:       user.ID,
			CoinPool:     user.BetAmount,
			LastGameBets: userData.LastBetAmount,
			Status:       user.Status,
			Sex:          user.User.GetSex(),  // 性别
			Address:      user.User.GetCity(), // 地址
		}

		// 提前结算的用户 计算结算金额
		if user.Status >= int32(msg.UserStatus_UserStopAction) {
			settleResult := user.CurAmount - user.InitAmount
			userDataRes.SettleResult = settleResult
		}

		// 手牌信息
		for k, item := range user.HoldCards {
			handCards := &msg.HandCardsRes{
				Cards:      item.Cards,
				Cardspoint: poker.ReducePoint(item.Point),
				CardsBet:   item.BetAmount,
				CardType:   int32(item.Type),
				StopAction: item.StopAction,
			}
			if k == 0 {
				userDataRes.Cards0 = handCards
			} else {
				userDataRes.Cards1 = handCards
			}

		}

		userDatas = append(userDatas, userDataRes)
	}

	// 当前状态剩余时间
	statusTimeLeft := int64(0)
	if game.TimerJob != nil {
		statusTimeLeft = game.TimerJob.GetTimeDifference() / 1000
	}

	// 场景消息结构体
	sceneInfo := msg.SceneInfoRes{
		UserData:       userDatas,
		GameStatus:     game.Status,
		StatusTimeLeft: statusTimeLeft,
		MaxBet:         game.RoomCfg.MaxAction,
		BetIndexs:      game.RoomCfg.ActionOption,
		Reconnect:      reConnect,
		RoomID:         int64(game.table.GetRoomID()),
		LimitJoin:      game.RoomCfg.LimitAction,
	}

	// 庄家手牌
	if game.HostCards != nil {
		sceneInfo.HostCards = game.HostCards.Cards
		sceneInfo.HostCardsPoint = poker.ReducePoint(game.HostCards.Point)
		sceneInfo.HostCardsType = int32(game.HostCards.Type)
	}

	// 当前操作玩家
	if game.CurActionUser != nil {
		sceneInfo.CurrentAskSeat = &msg.CurrentSeatRes{
			ChairId:       game.CurActionUser.ChairID,
			BetCardsIndex: game.CurActionUser.BetCardsIndex,
			GetPoker:      game.CurActionUser.GetPoker,
			DepartPoker:   game.CurActionUser.DepartPoker,
			DoubleBet:     game.CurActionUser.DoubleBet,
			Stand:         game.CurActionUser.Stand,
			GiveUp:        game.CurActionUser.GiveUp,
			UserId:        game.CurActionUser.UserID,
			StatusTime:    int32(statusTimeLeft),
		}
	}

	if userInter != nil {
		// 发给单个人
		log.Tracef("游戏 %d 是否是断线重连 %v 发送场景消息：%v", game.table.GetID(), reConnect, sceneInfo)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CSceneMessage), &sceneInfo)
		if err != nil {
			log.Errorf("发送场景消息失败， %v", err.Error())
		}
	} else {
		// 广播
		log.Tracef("游戏 %d 广播场景消息：%v", game.table.GetID(), sceneInfo)
		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CSceneMessage), &sceneInfo)
	}

}

// SendUserSitDown 广播用户坐下信息
func (game *Blackjack) SendUserSitDown(resp msg.UserSitDownRes) {

	log.Tracef("游戏 %d 广播用户坐下：%v", game.table.GetID(), resp)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CUserSitDown), &resp)
}

// SendFirstDeal 广播第一次发牌信息信息
func (game *Blackjack) SendFirstDeal(resp msg.FaPaiRes) {

	log.Tracef("游戏 %d 广播第一次发牌信息，%v", game.table.GetID(), resp)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CFaPai), &resp)
}

// SendSettleInfo 广播结算结果
func (game *Blackjack) SendSettleInfo(resp []*msg.UserResultsRes) {
	settleMsgRes := msg.SettleMsgRes{
		UserInfos: resp,
	}

	log.Tracef("游戏 %d 广播结算结果，%v", game.table.GetID(), settleMsgRes)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CSettle), &settleMsgRes)
}

// SendPocketCard 广播暗牌信息
func (game *Blackjack) SendPocketCard(PocketCard byte) {

	resp := msg.ZhuangJiaAnPaiRes{
		Cards:      []byte{PocketCard},
		Cardspoint: poker.ReducePoint(game.HostCards.Point),
		CardType:   int32(game.HostCards.Type),
	}

	log.Tracef("游戏 %d 广播庄家暗牌信息，%v", game.table.GetID(), resp)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CZhuangJiaAnPai), &resp)

}

// SendInsureResult 广播保险结果信息
func (game *Blackjack) SendInsureResult(resp []*msg.InsureResultRes) {

	insureResultListRes := msg.InsureResultListRes{
		InsureResult: resp,
	}

	log.Tracef("游戏 %d 广播保险结果信息，%v", game.table.GetID(), insureResultListRes)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CInsureResult), &insureResultListRes)
}

// SendCurrentSeat 广播当前玩家信息
func (game *Blackjack) SendCurrentSeat(resp CurUser) {
	// 广播当前操作玩家
	curUserRes := msg.CurrentSeatRes{
		ChairId:       resp.ChairID,
		UserId:        resp.UserID,
		BetCardsIndex: resp.BetCardsIndex,
		GetPoker:      resp.GetPoker,
		DepartPoker:   resp.DepartPoker,
		DoubleBet:     resp.DoubleBet,
		Stand:         resp.Stand,
		GiveUp:        resp.GiveUp,
		StatusTime:    int32(game.timeCfg.UserAction / 1000),
	}
	log.Tracef("游戏 %d 广播当前操作玩家，%v", game.table.GetID(), curUserRes.UserId)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CCurrentSeat), &curUserRes)
}

// SendCurrentSeat 广播发一张牌的信息
func (game *Blackjack) SendDealCard(resp msg.FaPaiOneRes) {

	log.Tracef("游戏 %d 广播发一张牌的信息，%v", game.table.GetID(), resp)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CFaPaiOne), &resp)
}

// SendBetSuccessInfo 广播下注成功信息
func (game *Blackjack) SendBetSuccessInfo(resp msg.BetSuccessRes) {

	log.Tracef("游戏 %d 广播用户 %d 下注成功消息，%v", game.table.GetID(), resp.UserId, resp)
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CBetSuccessMessageID), &resp)
}

// SendBetFailInfo 发送下注失败信息
func (game *Blackjack) SendBetFailInfo(resp msg.BetFailRes, userInter player.PlayerInterface) {

	log.Tracef("游戏 %d 发送用户 %d 下注失败消息，%v", game.table.GetID(), userInter.GetID(), resp)
	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBetFail), &resp)
	if err != nil {
		log.Errorf("游戏 %d 发送下注失败消息错误 : %v", game.table.GetID(), err)
	}
}

// SendAskDoResult 发送用户操作结果
func (game *Blackjack) SendAskDoResult(resp msg.AskDoRes, userInter player.PlayerInterface) {

	if userInter != nil {
		log.Tracef("游戏 %d 发送用户 %d 操作结果，%v", game.table.GetID(), userInter.GetID(), resp)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CAskDo), &resp)
		if err != nil {
			log.Errorf("游戏 %d 发送操作结果错误 : %v", game.table.GetID(), err)
		}
	} else {
		log.Tracef("游戏 %d 广播用户 %d 操作结果，%v", game.table.GetID(), resp.UserId, resp)
		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CAskDo), &resp)
	}

}

// SendControlInfo 发送控牌信息
func (game *Blackjack) SendControlInfo(resp msg.TestControlRes, userInter player.PlayerInterface) {

	err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CTestControl), &resp)
	if err != nil {
		log.Errorf("发送控制信息失败：%v", err.Error())
	}

}
