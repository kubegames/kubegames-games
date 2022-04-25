package game

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// UserRobBanker 用户抢庄
func (game *BankerNiuniu) UserRobBanker(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是抢庄
	if game.Status != int32(msg.GameStatus_RobBanker) {

		// 发送错误消息
		errMsg := msg.ErrRes{
			ErrNum: int32(msg.ErrorList_ActionTimeOut),
		}
		game.SendErrMsg(errMsg, userInter)
		return
	}

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
	}

	// 玩家已经发送过抢庄信息
	if user.Status == int32(msg.UserStatus_RobAction) {
		return
	}

	// 抢庄入参
	req := &msg.RobBankerReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析抢庄入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 抢庄请求：%v", userID, req)

	// index out of range
	if int(req.RobIndex) > len(game.GameCfg.RobOption)-1 || int(req.RobIndex) < -1 {
		log.Tracef("错误的抢注选项: %d", req.RobIndex)
		return
	}

	if req.RobIndex != -1 {
		// 抢庄倍数
		robMultiple := game.GameCfg.RobOption[req.RobIndex]

		// 最高抢庄倍数 = 携带金额/桌面有效玩家数（除开自己）/最高牌型3 / 底注
		maxMultiple := user.CurAmount / int64(len(game.UserList)-1) / 3 / game.RoomCfg.RoomCost

		if maxMultiple == 0 {
			maxMultiple = 1
		}

		// 资金不足
		if robMultiple > maxMultiple {

			// 发送错误消息
			errMsg := msg.ErrRes{
				ErrNum: int32(msg.ErrorList_LackOfFunds),
			}
			game.SendErrMsg(errMsg, userInter)
			return
		}
	}

	// 更新数据
	user.RobIndex = req.RobIndex
	user.Status = int32(msg.UserStatus_RobAction)

	game.UserList[userID] = user

	// 广播抢庄信息
	robBankerInfo := msg.RobBankerInfoRes{
		RobIndex: req.RobIndex,
		UserId:   userID,
		ChairId:  user.ChairID,
	}
	game.SendRobBankerInfo(robBankerInfo)

	//// 所有玩家都发送了抢庄信息，进入下一阶段
	allRob := true

	for _, user := range game.UserList {
		if user.Status != int32(msg.UserStatus_RobAction) {
			allRob = false
			break
		}
	}

	if allRob {
		game.Table.DeleteJob(game.TimerJob)
		game.EndRob()
	}
}

// UserRobBanker 用户投注
func (game *BankerNiuniu) UserBetChips(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是投注
	if game.Status != int32(msg.GameStatus_BetChips) {

		// 发送错误消息
		errMsg := msg.ErrRes{
			ErrNum: int32(msg.ErrorList_ActionTimeOut),
		}
		game.SendErrMsg(errMsg, userInter)
		return
	}

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
	}

	// 庄家不能投注
	if user.IsBanker {
		return
	}

	log.Tracef("用户状态为 %v", user.Status)
	// 用户已投注
	if user.Status == int32(msg.UserStatus_BetAction) {
		log.Tracef("用户 %v 已投注", user.ID)
		return
	}

	// 投注入参
	req := &msg.BetChipsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析投注入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 投注请求：%v", userID, req)

	if req.BetMultiple <= 0 {
		log.Tracef("错误的投注倍数: %d", req.BetMultiple)
		return
	}

	// 资金不足
	if req.BetMultiple > user.HighestMultiple {

		// 发送错误消息
		errMsg := msg.ErrRes{
			ErrNum: int32(msg.ErrorList_LackOfFunds),
		}
		game.SendErrMsg(errMsg, userInter)
		return
	}

	// 更新数据
	user.BetMultiple = req.BetMultiple
	user.Status = int32(msg.UserStatus_BetAction)

	game.UserList[userID] = user

	// 广播投注信息
	betChipsInfo := msg.BetChipsInfoRes{
		BetMultiple: req.BetMultiple,
		UserId:      userID,
		ChairId:     user.ChairID,
	}
	game.SendBetChipsInfo(betChipsInfo)

	//// 所有玩家都发送了投注信息，进入下一阶段
	allBet := true

	for _, user := range game.UserList {
		if user.IsBanker {
			continue
		}
		if user.Status != int32(msg.UserStatus_BetAction) {
			allBet = false
			break
		}
	}

	if allBet {
		game.Table.DeleteJob(game.TimerJob)

		game.EndBet()
	}
}

// UserRobBanker 用户摊牌
func (game *BankerNiuniu) UserShowCards(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是摊牌
	if game.Status != int32(msg.GameStatus_ShowChards) {

		// 发送错误消息
		errMsg := msg.ErrRes{
			ErrNum: int32(msg.ErrorList_ActionTimeOut),
		}
		game.SendErrMsg(errMsg, userInter)
		return
	}

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
	}

	// 用户已摊牌
	if user.Status == int32(msg.UserStatus_ShowedCards) {
		return
	}

	// 更新数据
	user.Status = int32(msg.UserStatus_ShowedCards)
	game.UserList[userID] = user

	// 广播摊牌结果
	showCardsResult := msg.ShowCardsRes{
		Cards:       user.HoldCards.Cards,
		UserId:      user.ID,
		ChairId:     user.ChairID,
		CardsType:   int32(user.HoldCards.CardsType),
		CardsIndexs: user.HoldCards.SpecialCardIndexs,
	}
	game.SendShowCardsMsg(showCardsResult)

	//// 所有玩家都发送了摊牌信息，进入下一阶段
	allShow := true

	for _, user := range game.UserList {

		if user.Status != int32(msg.UserStatus_ShowedCards) {
			allShow = false
			break
		}
	}

	if allShow {
		game.Table.DeleteJob(game.TimerJob)

		game.TimerJob, ok = game.Table.AddTimer(int64(game.TimeCfg.StatusSpace), game.EndShow)
		if !ok {
			log.Tracef("定时进入摊牌结束状态失败")
		}
	}
}

// UserDemandCards 要牌请求
func (game *BankerNiuniu) UserDemandCards(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	if user.Status != int32(msg.UserStatus_SitDown) {
		log.Tracef("玩家不在坐下状态")
		return
	}

	req := &msg.DemandCardsReq{}

	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析要牌入参错误：%v", err)
		return
	}

	log.Tracef("要牌入参：%v", req)

	if len(req.Cards) != 5 {
		log.Tracef("错误的要牌数量：%v", req)
		return
	}

	cards := req.Cards
	cardsType := poker.GetCardsType(cards)
	game.UserList[userID].HoldCards = &poker.HoldCards{
		Cards:             cards,
		CardsType:         cardsType,
		SpecialCardIndexs: poker.GetSpecialCardIndexs(cards, cardsType),
	}

}
