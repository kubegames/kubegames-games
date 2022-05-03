package game

import (
	"fmt"
	"game_poker/doudizhu/msg"
	"game_poker/doudizhu/poker"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// UserRobDizhu 玩家抢地主请求
func (game *DouDizhu) UserRobDizhu(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetId()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏状态不是抢地主状态
	if game.Status != int32(msg.GameStatus_RobStatus) {
		log.Tracef("玩家 %d 在非抢分阶段请求抢分", userID)
		return
	}

	// 请求玩家非当前抢分玩家
	if int32(game.curRobberChairID) != user.ChairID {
		log.Tracef("玩家 %d 非当前抢庄玩家", userID)
		return
	}

	// 抢地主请求消息
	req := &msg.RobReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析抢地主请求消息错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 抢地主请求消息：%v", userID, req)

	// 错误的抢分倍数
	if (req.RobNum != 0 && req.RobNum < game.CurRobNum) || req.RobNum > 3 {
		log.Tracef("最高抢分倍数: %d, 请求抢分倍数: %d", game.CurRobNum, req.RobNum)
		return
	}

	game.UserList[userID].RobNum = req.RobNum

	if req.RobNum > game.CurRobNum {
		game.CurRobNum = req.RobNum
	}

	resp := msg.RobResultRes{
		UserId:  userID,
		ChairId: user.ChairID,
		RobNum:  req.RobNum,
	}

	// 广播抢地主请求响应
	game.SendRobResult(resp)

	// 取消原有定时器
	game.TimerJob.Cancel()

	// 玩家抢最高分，直接成为地主
	if req.RobNum == 3 {
		game.Dizhu = user

		// 间隔时间后进入确认地主阶段
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.DefaultTime), game.confirmDizhu)
		return
	}

	// 当前抢庄玩家不是抢庄列表最后一个抢地主玩家
	if game.curRobberChairID != game.RobChairList[len(game.RobChairList)-1] {

		// 指定下一个抢地主玩家
		game.curRobberChairID++
		if game.curRobberChairID > 2 {
			game.curRobberChairID = 0
		}

		// 广播当前抢地主玩家
		game.SendCurrentRobber(int32(game.curRobberChairID))

		// 定时进入轮询抢庄
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.RobTime), game.CheckRob)
		return
	}

	// 抢地主无结果，重新进入发牌，抢地主
	if game.CurRobNum == 0 {

		// 清空玩家手牌 todo 等待控牌流程写入
		for ID, _ := range game.UserList {
			game.UserList[ID].Cards = []byte{}
		}

		game.DealCards()
		return
	}

	// 抢分最高的玩家指定为地主
	for _, player := range game.UserList {
		if player.RobNum == game.CurRobNum {
			game.Dizhu = player

			// 间隔时间后进入确认地主阶段
			game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.DefaultTime), game.confirmDizhu)
			return
		}
	}

}

// UserRedouble 玩家加倍请求
func (game *DouDizhu) UserRedouble(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetId()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏状态不是加倍状态
	if game.Status != int32(msg.GameStatus_RedoubleStatus) {
		log.Tracef("玩家 %d 在非加倍阶段请求加倍", userID)
		return
	}

	// 玩家已经发送过加倍请求
	if user.AddNum > 0 {
		return
	}

	// 抢地主请求消息
	req := &msg.RedoubleReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析加倍请求错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 加倍请求消息：%v", userID, req)

	// 错误的加倍倍数
	if req.AddNum != 1 && req.AddNum != 2 && req.AddNum != 4 {
		log.Tracef("错误加倍倍数 %v", req.AddNum)
		return
	}

	game.UserList[userID].AddNum = req.AddNum

	resp := msg.RedoubleResultRes{
		UserId:  user.ID,
		ChairId: user.ChairID,
		AddNum:  req.AddNum,
	}

	// 广播加倍结果
	game.SendRedoubleResult(resp)

	// 所有人都发送过加倍请求
	allRedouble := true

	for _, user := range game.UserList {
		if user.AddNum == 0 {
			allRedouble = false
		}
	}

	// 所有人都发送过加倍请求, 提前进入加倍结束状态
	if allRedouble {
		game.TimerJob.Cancel()

		// 间隔时间后进入打牌阶段
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.DefaultTime), game.PutCards)

	}

}

// UserPutCards 玩家出牌请求
func (game *DouDizhu) UserPutCards(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetId()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 出牌状态检测
	if game.Status != int32(msg.GameStatus_PutCardStatus) {
		log.Tracef("玩家 %d 在非出牌阶段请求出牌", userID)
		return
	}

	putInfoResp := msg.PutInfoRes{
		IsSuccess: false,
		UserId:    userID,
		ChairId:   user.ChairID,
	}

	// 当前操作玩家检测
	if game.CurrentPlayer.UserID != userID {
		log.Tracef("请求玩家不是当前操作玩家")

		// 错误信息：超时
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
	log.Tracef("用户 %d 出牌请求 %v", userID, req.Cards)

	// 出牌动画中，不允许再出牌
	if game.InAnimation {
		log.Warnf("用户 %d 短时间内多次出牌", user.ID)
		return
	}

	// 出牌存在性检测
	if !game.CheckCardsExist(user.ID, req.Cards) {
		log.Warnf("用户 %d 出牌请求没有在手牌中找到", userID)
		return
	}

	// 出牌重复性性检测
	if game.CheckCardsRepeated(req.Cards) {
		log.Warnf("用户 %d 出牌请求出现重复卡牌 %v", userID, req.Cards)
		return
	}

	// 请求出牌牌型
	cardsType := poker.GetCardsType(req.Cards)

	if len(req.Cards) != 0 {

		// 牌型检测
		if cardsType == msg.CardsType_Normal {

			// 错误信息：牌型错误
			putInfoResp.ErrNum = int32(msg.ErrList_WrongfulType)
			game.SendPutInfo(putInfoResp)

			log.Tracef("错误牌型")
			return
		}

		// 要不起检测
		if game.CurrentPlayer.ActionType == int32(msg.UserActionType_NoPermission) {

			// 管不上
			putInfoResp.ErrNum = int32(msg.ErrList_NotGreater)
			game.SendPutInfo(putInfoResp)

			log.Tracef("要不起不能出牌")
			return
		}

		// 接牌检测
		if game.CurrentPlayer.ActionType == int32(msg.UserActionType_TakeOverCard) &&
			!poker.ContrastCards(req.Cards, game.CurrentCards.Cards) {

			// 管不上
			putInfoResp.ErrNum = int32(msg.ErrList_NotGreater)
			game.SendPutInfo(putInfoResp)

			log.Tracef("不能压过上副牌")
			return
		}
		game.LastOutCardType[game.CurrentPlayer.ChairID] = int32(cardsType)
	} else {

		// 出牌空牌检测
		if game.CurrentPlayer.ActionType == int32(msg.UserActionType_PutCard) {
			log.Errorf("必需出牌")
			return
		}

		game.IsPass[game.CurrentPlayer.ChairID] = true
		game.LastPassCardType[game.CurrentPlayer.ChairID] = game.CurrentCards.CardsType

	}

	// 过掉所有不能出检测，处理出牌

	game.TimerJob.Cancel()

	// 出牌成功，锁住桌子，不允许用户出牌请求
	game.InAnimation = true

	// 出炸弹, 加倍
	if cardsType == msg.CardsType_Bomb {
		game.BoomMultiple *= 2
		var (
			boomToUserID  int64
			bommToChairID int32
		)
		if game.CurrentCards.UserID != 0 && game.CurrentCards.UserID != user.ID {
			boomToUserID = game.CurrentCards.UserID
			bommToChairID = game.UserList[game.CurrentCards.UserID].ChairID
		}
		putInfoResp.BoomToPlayer = &msg.PlayerInfo{
			UserId:  boomToUserID,
			ChairId: bommToChairID,
		}
	}

	// 出火箭, 加倍
	if cardsType == msg.CardsType_Rocket {
		game.RocketMultiple *= 2
	}

	// 更新卡牌变动
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

	log.Warnf("玩家 %d , 打出: %v, 剩余手牌: %v", userID, req.Cards, game.UserList[userID].Cards)

	// 添加玩家打牌记录
	game.UserList[userID].PutCardsRecords = append(game.UserList[userID].PutCardsRecords, req.Cards)

	putcardsStr := poker.CardsToString(poker.ReverseSortCards(req.Cards))

	if len(req.Cards) == 0 {
		putcardsStr = "要不起"
		if game.CurrentPlayer.Permission {
			putcardsStr = "不要"
		}
	}

	// 添加出牌日志
	game.PutCardsLog += user.GetSysRole() + "ID: " + fmt.Sprintf(`%d`, user.User.GetID()) + " " + putcardsStr + ", "

	putInfoResp.IsSuccess = true
	putInfoResp.Cards = req.Cards
	putInfoResp.CardType = int32(cardsType)

	// 广播打牌消息
	game.SendPutInfo(putInfoResp)

	// 游戏轮转计数器自加
	game.StepCount++

	// 牌型时间
	var cardsTypeTime int

	switch cardsType {

	// 火箭
	case msg.CardsType_Rocket:
		cardsTypeTime = game.TimeCfg.RocketTime

	// 炸弹
	case msg.CardsType_Bomb:
		cardsTypeTime = game.TimeCfg.BombTime

	// 连对
	case msg.CardsType_SerialPair:
		cardsTypeTime = game.TimeCfg.SerialPairTime

	// 顺子
	case msg.CardsType_Sequence:
		cardsTypeTime = game.TimeCfg.SequenceTime

	// 飞机
	case msg.CardsType_SerialTriplet:
		cardsTypeTime = game.TimeCfg.PlaneTime

	// 飞机带翅膀
	case msg.CardsType_SerialTripletWithWing:
		cardsTypeTime = game.TimeCfg.PlaneTime

	// 普通牌型
	default:
		cardsTypeTime = game.TimeCfg.DefaultTime

	}

	// 有玩家逃完所有牌, 跳出出牌阶段, 走到结算阶段
	if len(game.UserList[userID].Cards) == 0 {

		// 解锁桌子
		game.InAnimation = false

		// 牌型时间后进入结算
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(cardsTypeTime), game.Settle)
		return
	}

	// 更新当前牌全牌组
	if len(req.Cards) > 0 {
		game.CurrentCards = poker.HandCards{
			Cards:       req.Cards,
			UserID:      userID,
			WeightValue: poker.GetCardsWeightValue(req.Cards, cardsType),
			CardsType:   int32(cardsType),
		}
	}

	// 牌型时间后 指定下一个当前操作玩家
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(cardsTypeTime), game.FindNextPlayer)
}

// UserHangUp 用户托管请求
func (game *DouDizhu) UserHangUp(buffer []byte, userInter player.PlayerInterface) {

	// 用户ID
	userID := userInter.GetId()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏出牌状态检测
	if game.Status != int32(msg.GameStatus_PutCardStatus) {
		log.Tracef("玩家 %d 非出牌阶段请求出牌", userID)
		return
	}

	// 托管入参
	req := &msg.HangUpReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析请求托管错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 托管请求：%v", userID, req)

	// 重复请求检测
	if req.IsHangUp && user.Status == int32(msg.UserStatus_UserHangUp) ||
		!req.IsHangUp && user.Status != int32(msg.UserStatus_UserHangUp) {
		log.Tracef("用户 %d 托管请求重复", userID)
		return
	}

	// 更改玩家托管状态
	if req.IsHangUp {
		game.UserList[userID].Status = int32(msg.UserStatus_UserHangUp)
	} else {
		game.UserList[userID].Status = int32(msg.UserStatus_UserNormal)
	}

	// 广播玩家托管状态
	game.SendHangUpInfo(userID)

	// 当前操作玩家选择托管，跳出定时，直接操作
	if req.IsHangUp && game.CurrentPlayer.UserID == userID {
		game.TimerJob.Cancel()
		game.CheckAction()
	}
}

// UserGetTips 用户提示请求
func (game *DouDizhu) UserGetTips(buffer []byte, userInter player.PlayerInterface) {
	// 用户ID
	userID := userInter.GetId()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	// 游戏状态不是出牌阶段
	if game.Status != int32(msg.GameStatus_PutCardStatus) {
		log.Tracef("玩家 %d 在游戏 %d 非出牌阶段请求提示", userID, game.Table.GetId())
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
	log.Tracef("用户 %d 提示请求: %s", userID, fmt.Sprintf("%+v\n", req))

	cardsArr := [][]byte{}

	tipsCards := game.CurrentCards

	// 没有当前牌权或者自己是当前牌权玩家
	if userID == game.CurrentCards.UserID || len(game.CurrentCards.Cards) == 0 {
		cards := []byte{poker.GetSmallestCard(user.Cards)}
		cardsType := poker.GetCardsType(cards)
		weight := poker.GetCardsWeightValue(cards, cardsType)

		tipsCards = poker.HandCards{
			Cards:       cards,
			UserID:      userID,
			WeightValue: weight,
			CardsType:   int32(cardsType),
		}
		cardsArr = append(cardsArr, cards)
	}

	var takeOverCards []byte

	maxLoop := 16

	for {
		takeOverCards = poker.TakeOverCards(tipsCards, user.Cards)
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

		// 防止死循环
		if len(cardsArr) >= maxLoop {
			break
		}
	}

	tipsRes := msg.TipsRes{
		Cards: cardsArr,
	}

	game.SendTipsInfo(tipsRes, user.User)

}

// UserDemandCards 配牌请求
func (game *DouDizhu) UserDemandCards(buffer []byte, userInter player.PlayerInterface) {
	for id, user := range game.UserList {
		if id == userInter.GetId() {
			continue
		}

		if len(user.Cards) != 0 {
			log.Tracef("其他玩家 %d 已经配牌，不允许用户 %d 配牌", user.ID, userInter.GetId())
			return
		}
	}

	// 用户ID
	userID := userInter.GetId()
	_, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家 %d 异常", userID)
		return
	}

	if game.Status != int32(msg.GameStatus_GameInitStatus) {
		log.Tracef("在桌子 %d 非初始化时，不允许用户 %d 配牌", game.Table.GetId(), userID)
		return
	}

	// 提示入参
	req := &msg.DemandCardsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析提示入参错误: %v", err.Error())
		return
	}
	log.Tracef("用户 %d 配牌请求：%v", userID, req)

	if len(req.Cards) != 17 {
		log.Tracef("用户 %d 配牌不满足17张牌", userID)
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

// UserClean 清桌请求
func (game *DouDizhu) UserClean() {
	if game.TimerJob != nil {
		game.TimerJob.Cancel()
	}

	game.GameOver()
}
