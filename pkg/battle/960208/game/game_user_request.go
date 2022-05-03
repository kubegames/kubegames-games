package game

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// UserRobBanker 用户抢庄
func (game *ThreeDoll) UserRobBanker(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是抢庄
	if game.Status != int32(msg.GameStatus_RobBanker) {

		// 发送错误消息
		//errMsg := msg.ErrRes{
		//	ErrNum: int32(msg.ErrorList_ActionTimeOut),
		//}
		//game.SendErrMsg(errMsg, userInter)
		return
	}

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("游戏 %d 获取玩家 %d 异常", game.Table.GetID(), userID)
		return
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

	user.Status = int32(msg.UserStatus_RobAction)
	user.IsRob = req.IsRob

	game.UserList[userID] = user

	// 广播抢庄信息
	robBankerInfo := msg.RobInfoRes{
		UserId:  userID,
		ChairId: user.ChairID,
		IsRob:   req.IsRob,
	}
	game.SendRobBankerInfo(robBankerInfo)

	//// 所有玩家都发送了抢庄信息，进入下一阶段
	allRob := true

	for _, user := range game.UserList {
		if user.Status < int32(msg.UserStatus_RobAction) {
			allRob = false
			break
		}
	}

	if allRob {
		if game.TimerJob != nil {
			game.Table.DeleteJob(game.TimerJob)
		}
		game.EndRob()
	}
}

// UserRobBanker 用户投注
func (game *ThreeDoll) UserBetChips(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是投注
	if game.Status != int32(msg.GameStatus_BetChips) {

		// 发送错误消息
		//errMsg := msg.ErrRes{
		//	ErrNum: int32(msg.ErrorList_ActionTimeOut),
		//}
		//game.SendErrMsg(errMsg, userInter)
		return
	}

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	if user.IsBanker {
		log.Warnf("庄家不能抢庄")
		return
	}

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

	// 检测投注倍数是否允许
	var isExist bool

	for _, oneMultiple := range user.Multiples {
		if oneMultiple == req.BetMultiple {
			isExist = true
		}
	}

	if !isExist {
		log.Tracef("错误的投注倍数: %d", req.BetMultiple)
		return
	}

	// 更新数据
	user.BetMultiple = req.BetMultiple
	user.Status = int32(msg.UserStatus_BetAction)

	game.UserList[userID] = user

	// 广播投注信息
	betInfo := msg.BetInfoRes{
		BetMultiple: req.BetMultiple,
		UserId:      userID,
		ChairId:     user.ChairID,
	}
	game.SendBetInfo(betInfo)

	//// 除开庄家外的玩家都发送了投注信息，进入下一阶段
	allBet := true

	for _, user := range game.UserList {
		if user.ID == game.Banker.ID {
			continue
		}
		if user.Status != int32(msg.UserStatus_BetAction) {
			allBet = false
			break
		}
	}

	if allBet {
		if game.TimerJob != nil {
			game.Table.DeleteJob(game.TimerJob)
		}
		game.EndBet()
	}
}

// UserRobBanker 用户摊牌
func (game *ThreeDoll) UserShowCards(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	// 游戏状态不是摊牌
	if game.Status != int32(msg.GameStatus_ShowCards) {
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
	game.UserList[userID].Status = int32(msg.UserStatus_ShowedCards)

	// 广播摊牌结果
	showCardsResult := msg.ShowCardsRes{
		Cards:     user.HoldCards.Cards,
		UserId:    user.ID,
		ChairId:   user.ChairID,
		CardsType: int32(user.HoldCards.CardsType),
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
		if game.TimerJob != nil {
			game.Table.DeleteJob(game.TimerJob)
		}
		game.EndShow()
		//game.TimerJob, ok = game.Table.AddTimer(int64(game.TimeCfg.StatusSpace), game.EndShow)
		//if !ok {
		//	log.Tracef("定时进入摊牌结束状态失败")
		//}
	}
}

// UserPullRecords 用户拉取战绩
func (game *ThreeDoll) UserPullRecords(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	gameRecords := data.GetUserTabledata(user.User)

	var Records []*msg.Record

	for _, record := range gameRecords {
		Records = append(Records, &msg.Record{
			Time:      record.Time,
			GameNum:   record.GameNum,
			RoomLevel: record.RoomLevel,
			Result:    record.Result,
			Status:    record.Status,
		})
	}

	gameRecordsRes := msg.GameRecordsRes{
		Records: Records,
	}

	game.SendGameRecords(gameRecordsRes, userInter)
}

// UserDemandCards 配牌请求
func (game *ThreeDoll) UserDemandCards(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	if user.Status != int32(msg.UserStatus_SitDown) && user.Status != int32(msg.UserStatus_Ready) {
		log.Tracef("用户 %d 配牌失败，游戏已开始", userID)
		return
	}

	req := &msg.DemandCardsReq{}

	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析用户 %d 要牌入参错误：%v", userID, err)
		return
	}

	log.Tracef("要牌入参：%v", req)

	// 指定手牌
	var demandCards []byte

	switch req.DemandType {
	//
	case int32(msg.DemandType_NoDemand):
		log.Tracef("用户 %d 无配牌需求", userID)
		break

	// 爆玖牌型
	case int32(msg.DemandType_ExplosionNineCards):
		demandCards = poker.GetInputCardsType(msg.CardsType_ExplosionNine)
		break

	// 炸弹牌型
	case int32(msg.DemandType_BoomCards):
		demandCards = poker.GetInputCardsType(msg.CardsType_Boom)
		break

	// 三公牌型
	case int32(msg.DemandType_ThreeDollCards):
		demandCards = poker.GetInputCardsType(msg.CardsType_ThreeDoll)
		break

	// 输入牌型
	case int32(msg.DemandType_PutIn):
		if len(req.Cards) != 3 {
			log.Warnf("用户 %d 错误的要牌数量：%v", userID, req)
			return
		}

		// 存在检测, 重复检测
		for _, card := range req.Cards {
			var (
				sameCount int  // 重复个数
				isExit    bool // 是否存在
			)

			for _, card1 := range req.Cards {
				if card == card1 {
					sameCount++
				}

				if sameCount >= 2 {
					log.Warnf("配牌有重复牌 %v", card)
					return
				}
			}

			for _, waitTakeCard := range poker.Deck {
				if card == waitTakeCard {
					isExit = true
				}
			}

			if !isExit {
				log.Warnf("配牌中有不存在的牌 %v", card)
				return
			}
		}
		demandCards = req.Cards
		break

	}

	log.Tracef("用户 %d 配牌结果 %v", userID, demandCards)
	game.UserList[userID].DemandReq = data.DemandReq{
		DemandType:  req.DemandType,
		DemandCards: demandCards,
	}
}
