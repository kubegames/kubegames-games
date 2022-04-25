package game

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// UserPutCards 玩家出牌请求
func (game *RunFaster) UserPutCards(buffer []byte, userInter player.PlayerInterface, isSystem bool) {
	// 用户ID
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏状态不是出牌阶段
	if game.Status != int32(msg.GameStatus_PutCardStatus) {
		log.Tracef("玩家 %d 在游戏 %d 非出牌阶段请求出牌", userID)
		return
	}

	putInfoResp := msg.PutInfoRes{
		IsSuccess: false,
		Cards:     nil,
		CardType:  0,
		UserId:    userID,
		ChairId:   user.ChairID,
	}

	// 请求玩家不是当前操作玩家
	if game.CurrentPlayer.UserID != userID {
		log.Tracef("请求玩家不是当前操作玩家")
		putInfoResp.ErrNum = int32(msg.ErrList_TimeOut)
		game.SendPutInfo(putInfoResp)
		return
	}

	// 出牌入参
	req := &msg.PutCardsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析出牌入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 出牌请求：%v", userID, req)

	// 检测出牌请求的牌是否为用户手上的牌
	if !game.CheckCardsExist(user.ID, req.Cards) {
		log.Tracef("用户 %d 出牌请求没有在手牌中找到", userID)
		return
	}

	var nextCards poker.HandCards

	// 判断牌型
	cardsType := poker.GetCardsType(req.Cards)

	sortCards := poker.SortCards(req.Cards)
	transCards := poker.TransformCards(sortCards)

	log.Tracef("用户 %d 手牌：%v, 牌型：%v", user.ID, transCards, cardsType)

	// 没有牌型，错误牌
	if cardsType == msg.CardsType_Normal {
		log.Tracef("出牌没有牌型，错误牌")
		putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
		game.SendPutInfo(putInfoResp)
		return
	}
	var isLegalCardsType bool

	if !(cardsType == msg.CardsType_Triplet || cardsType == msg.CardsType_TripletWithSingle || cardsType == msg.CardsType_IncompleteSerialTripletWithTwo) {
		isLegalCardsType = true
	}

	// 非最后一手牌出不合法牌型
	if !isLegalCardsType && len(req.Cards) != len(user.Cards) {
		log.Tracef("非最后一手牌出不合法牌型")
		putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
		game.SendPutInfo(putInfoResp)
		return
	}

	// 出牌或者接牌
	if game.CurrentPlayer.ActionType == int32(msg.UserActionType_PutCard) {

		// 出牌
		nextCards = poker.HandCards{
			Cards:       req.Cards,
			UserID:      userID,
			WeightValue: poker.GetCardsWeightValue(req.Cards, cardsType),
			CardsType:   int32(cardsType),
		}
	} else {

		// 接牌

		// 合法牌型或不合法牌型
		if isLegalCardsType {
			// 合法牌型

			// 牌型相同，长度不同，pass
			if game.CurrentCards.CardsType == int32(cardsType) && len(req.Cards) != len(game.CurrentCards.Cards) {
				log.Tracef("接牌牌型相同，长度不同，pass")
				putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
				game.SendPutInfo(putInfoResp)
				return
			}

			// 牌型不同，并且出牌不是炸弹，pass
			if game.CurrentCards.CardsType != int32(cardsType) && cardsType != msg.CardsType_Bomb {
				log.Tracef("接牌牌型不同，并且出牌不是炸弹，pass")
				putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
				game.SendPutInfo(putInfoResp)
				return
			}
		} else {

			// 不是合法牌型

			// 出牌为三同张或者三带一，上家不为三带二
			if (cardsType == msg.CardsType_Triplet || cardsType == msg.CardsType_TripletWithSingle) &&
				game.CurrentCards.CardsType != int32(msg.CardsType_TripletWithTwo) {
				log.Tracef("最后一手牌 三带二牌型不对")

				putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
				game.SendPutInfo(putInfoResp)
				return
			}

			// 出牌为不合法的飞机带翅膀，上家不为飞机带翅膀或者飞机长度不一致
			if cardsType == msg.CardsType_IncompleteSerialTripletWithTwo {

				nextRepeatedCards := poker.CheckRepeatedCards(req.Cards)
				CurrentRepeatedCards := poker.CheckRepeatedCards(game.CurrentCards.Cards)

				if game.CurrentCards.CardsType != int32(msg.CardsType_SerialTripletWithTwo) ||
					len(nextRepeatedCards[2]) == len(CurrentRepeatedCards[2]) {
					log.Tracef("最后一手牌 飞机带翅膀牌型不对")
					putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
					game.SendPutInfo(putInfoResp)
					return
				}
			}
		}

		// 出牌牌组
		nextCards = poker.HandCards{
			Cards:       req.Cards,
			UserID:      userID,
			WeightValue: poker.GetCardsWeightValue(req.Cards, cardsType),
			CardsType:   int32(cardsType),
		}

		// 出牌牌型为炸弹，上家牌型不是炸弹 直接大过上家
		if !(game.CurrentCards.CardsType != int32(msg.CardsType_Bomb) && cardsType == msg.CardsType_Bomb) {
			// 比较权重值
			if nextCards.WeightValue <= game.CurrentCards.WeightValue {
				log.Tracef("牌型对，大不过")
				putInfoResp.ErrNum = int32(msg.ErrList_NotGreater)
				game.SendPutInfo(putInfoResp)
				return
			}

		}

	}

	// 更新玩家手牌
	for _, delCard := range req.Cards {

		// 删除玩家要出的手牌
		for k, card := range user.Cards {
			if delCard == card {
				game.UserList[userID].Cards = append(game.UserList[userID].Cards[:k], game.UserList[userID].Cards[k+1:]...)
				break
			}
		}

		// 删除记牌器手牌
		for k, card := range game.LeftCards {
			if delCard == card {
				game.LeftCards = append(game.LeftCards[:k], game.LeftCards[k+1:]...)
				break
			}
		}
	}

	// 添加出牌日志
	game.PutCardsLog += user.GetSysRole() + "ID: " + fmt.Sprintf(`%d`, user.User.GetID()) + " " + poker.CardsToString(req.Cards) + ", "

	// 添加玩家打牌记录
	game.UserList[userID].PutCardsRecords = append(game.UserList[userID].PutCardsRecords, req.Cards)

	// 报单放走包赔检测
	if nextCards.CardsType == int32(msg.CardsType_SingleCard) {

		game.UserList[userID].TakeSingleRisk = game.CheckTakeSingleRisk(user, nextCards.Cards[0])

	}

	// 有玩家逃完所有牌，计算所有玩家输赢
	if len(game.UserList[userID].Cards) == 0 {
		game.UserList[userID].SettleCost = game.RoomCfg.RoomCost

		// 玩家最后逃单张牌获得胜利
		if nextCards.CardsType == int32(msg.CardsType_SingleCard) {
			preChairID := user.ChairID - 1
			nextChairID := user.ChairID + 1
			if user.ChairID == 0 {
				preChairID = 2
			}

			if int(user.ChairID) == len(game.UserList)-1 {
				nextChairID = 0
			}

			preCardsCount := int64(len(game.UserList[game.Seats[preChairID]].Cards))
			nextCardsCount := int64(len(game.UserList[game.Seats[nextChairID]].Cards))

			// 上家承担放走包赔的风险
			if game.UserList[game.Seats[preChairID]].TakeSingleRisk {
				game.UserList[game.Seats[preChairID]].SettleCost = -1 * (preCardsCount + nextCardsCount) * game.RoomCfg.RoomCost
				game.UserList[game.Seats[nextChairID]].SettleCost = 0
			} else {
				game.UserList[game.Seats[preChairID]].SettleCost = -1 * preCardsCount * game.RoomCfg.RoomCost
				game.UserList[game.Seats[nextChairID]].SettleCost = -1 * nextCardsCount * game.RoomCfg.RoomCost
			}

		} else {
			// 玩家最后逃其他牌型获得胜利
			for id, user := range game.UserList {
				if id == userID {
					continue
				}
				game.UserList[id].SettleCost = -1 * int64(len(user.Cards)) * game.RoomCfg.RoomCost
			}
		}
	}

	game.CurrentCards = nextCards
	putInfoResp = msg.PutInfoRes{
		IsSuccess: true,
		Cards:     nextCards.Cards,
		CardType:  nextCards.CardsType,
		UserId:    nextCards.UserID,
		ChairId:   user.ChairID,
	}
	game.SendPutInfo(putInfoResp)

	// 最后一副牌是炸弹，触发炸弹的实时结算
	if cardsType == msg.CardsType_Bomb && len(game.UserList[userID].Cards) == 0 {
		game.BoomSettle(userID)
	}

	// 系统帮忙出牌
	if isSystem && user.Status == int32(msg.UserStatus_UserNormal) {

		if len(game.UserList) == 0 {
			log.Errorf("游戏 %d 系统轮转数 %d, 当前玩家轮转数 %d，当前玩家%d, ", game.Table.GetID(), game.StepCount, game.CurrentPlayer.StepCount, game.CurrentPlayer.UserID)
		}
		game.UserList[user.ID].Status = int32(msg.UserStatus_UserHangUp)

		// 广播玩家进入托管
		userStatusResp := msg.UserStatusRes{
			UserId:   user.ID,
			ChairId:  game.UserList[user.ID].ChairID,
			IsHangUp: true,
		}
		game.SendUserStatus(userStatusResp)
	}

	// 找到下一个执行玩家
	game.CalculateCurrentUser()

	// 中断定时任务
	if game.TimerJob != nil {
		game.Table.DeleteJob(game.TimerJob)
	}

	game.TimerJob, _ = game.Table.AddTimer(int64(game.TimeCfg.OneSecond), game.TurnAction)

}

func (game *RunFaster) UserGetTips(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏状态不是出牌阶段
	if game.Status != int32(msg.GameStatus_PutCardStatus) {
		log.Tracef("玩家 %d 在游戏 %d 非出牌阶段请求提示", userID)
		return
	}

	// 请求玩家不是当前操作玩家
	if game.CurrentPlayer.UserID != userID {
		log.Tracef("请求玩家不是当前操作玩家")
		return
	}

	// 提示入参
	req := &msg.TipsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析提示入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 提示请求：%v", userID, req)

	cardsArr := [][]byte{}

	tipsCards := game.CurrentCards

	// 没有当前牌权或者自己是当前牌权玩家
	if userID == game.CurrentCards.UserID || len(game.CurrentCards.Cards) == 0 {
		cards := poker.CheckPutSingleCard(user.Cards, false)
		cardsType := poker.GetCardsType(cards)
		weight := poker.GetCardsWeightValue(cards, cardsType)

		tipsCards = poker.HandCards{
			Cards:       poker.CheckPutSingleCard(user.Cards, false),
			UserID:      userID,
			WeightValue: weight,
			CardsType:   int32(cardsType),
		}
		cardsArr = append(cardsArr, cards)
	}

	var takeOverCards []byte

	for {
		takeOverCards = poker.CheckTakeOverCards(tipsCards, user.Cards, false)
		if len(takeOverCards) == 0 {
			break
		} else {
			cardsType := poker.GetCardsType(takeOverCards)
			weight := poker.GetCardsWeightValue(takeOverCards, cardsType)
			tipsCards = poker.HandCards{
				Cards:       takeOverCards,
				UserID:      userID,
				WeightValue: weight,
				CardsType:   int32(cardsType),
			}
			cardsArr = append(cardsArr, takeOverCards)
		}

		// todo 防止死循环
	}

	tipsRes := msg.TipsRes{
		Cards: cardsArr,
	}

	game.SendTipsInfo(tipsRes, user.User)

}

// UserCancelHangUp 用户取消托管请求
func (game *RunFaster) UserCancelHangUp(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	if user.Status == int32(msg.UserStatus_UserNormal) {
		log.Tracef("获取玩家 %d 状态正常，不在托管中", userID)
		return
	}

	game.UserList[userID].Status = int32(msg.UserStatus_UserNormal)

	if game.CurrentPlayer.UserID == userID && game.CurrentPlayer.Permission {
		game.CurrentPlayer.ActionTime = game.TimeCfg.OperationTime
	}

	// 广播玩家取消托管
	userStatusResp := msg.UserStatusRes{
		UserId:   userID,
		ChairId:  user.ChairID,
		IsHangUp: false,
	}
	game.SendUserStatus(userStatusResp)
}

func (game *RunFaster) UserDemandCards(buffer []byte, userInter player.PlayerInterface) {
	for id, user := range game.UserList {
		if id == userInter.GetID() {
			continue
		}

		if len(user.Cards) != 0 {
			log.Tracef("其他玩家 %d 已经配牌，不允许用户 %d 配牌", user.ID, userInter.GetID())
			return
		}
	}

	// 用户ID
	userID := userInter.GetID()
	_, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	if game.Status != int32(msg.GameStatus_GameInitStatus) {
		log.Tracef("在桌子 %d 非初始化时，不允许用户 %d 配牌", game.Table.GetID(), userID)
		return
	}

	// 提示入参
	req := &msg.DemandCardsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析提示入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 配牌请求：%v", userID, req)

	if len(req.Cards) != 16 {
		log.Tracef("用户 %d 配牌不满足16张牌", userID)
		return
	}

	// 存在检测, 重复检测
	for _, card := range req.Cards {
		var sameCount int
		for _, waitTakeCard := range poker.Deck {
			if card == waitTakeCard {
				sameCount++
			}
		}

		if sameCount == 0 {
			log.Tracef("用户 %d 配牌有不存在的牌 %v", userID, card)
			return
		}

		if sameCount >= 2 {
			log.Tracef("用户 %d 配牌有重复牌", userID)
			return
		}
	}

	game.UserList[userID].Cards = req.Cards
}
