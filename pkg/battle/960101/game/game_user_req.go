package game

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// UserBet 用户下注操作
func (game *Blackjack) UserBet(buffer []byte, userInter player.PlayerInterface) {
	// 游戏不在下注状态，退出
	if game.Status != int32(msg.GameStatus_BetStatus) {

		betFailRes := msg.BetFailRes{
			ErrNum: int32(msg.ErrorList_NotBetStatus),
		}

		// 通知用户下注失败信息
		game.SendBetFailInfo(betFailRes, userInter)
		return
	}

	// 下注用户是否在桌上
	user, ok := game.AllUserList[userInter.GetID()]
	if !ok {
		return
	}

	// 用户状态不允许下注
	if user.Status != int32(msg.UserStatus_UserReady) && user.Status != int32(msg.UserStatus_UserBetSuccess) {
		return
	}

	// 下注入参
	req := &msg.BetReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
		return
	}

	log.Debugf("用户 %d 下注入参: %v", user.ID, req)

	// 同时请求重复押注和全压；一般不会出现这种情况
	if req.All && req.RepeatBet {
		return
	}

	// 前端需要记数点当前下注数
	var (
		requestAmount, allBet int64
		userData              data.UserData
	)

	// 重复押注操作
	if req.RepeatBet {

		// 全局user数据
		userData = data.GetUserInterdata(user.User)

		// 第一次下注，并且上一局有下注
		if _, ok := game.UserList[user.ID]; !ok && userData.LastBetAmount > 0 {

			allBet = userData.LastBetAmount
			requestAmount = userData.LastBetAmount

		} else {
			return
		}
	}

	if req.BetIndex < 0 || req.BetIndex > int32(len(game.RoomCfg.ActionOption)-1) {
		log.Warnf("错误到下注下标：%d", req.BetIndex)
		return
	}

	// 全压
	if req.All {
		if (user.BetAmount + user.CurAmount) > game.RoomCfg.MaxAction {
			allBet = game.RoomCfg.MaxAction
		} else {
			allBet = user.BetAmount + user.CurAmount
		}

		requestAmount = allBet - user.BetAmount

		// 全压前已经到达最大押注额度，全压操作无效
		if requestAmount == 0 {
			return
		}
	}

	// 选择筹码
	if !req.All && !req.RepeatBet {
		requestAmount = game.RoomCfg.ActionOption[req.BetIndex]
		allBet = user.BetAmount + requestAmount
	}

	// 超过最大额度判断
	if allBet > game.RoomCfg.MaxAction {
		betFailRes := msg.BetFailRes{
			ErrNum: int32(msg.ErrorList_ExcessMax),
		}

		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBetFail), &betFailRes)
		if err != nil {
			log.Errorf("send bet fail : %v", err)
		}
		return
	}

	// 下注大于持有注，钱不够
	if requestAmount > user.CurAmount {
		betFailRes := msg.BetFailRes{
			ErrNum: int32(msg.ErrorList_LackOfFunds),
		}

		// 通知用户下注失败信息
		game.SendBetFailInfo(betFailRes, userInter)
		return
	}

	user.BetAmount = allBet
	user.Chip = allBet
	user.CurAmount -= requestAmount
	user.Status = int32(msg.UserStatus_UserBetSuccess)

	// 更新玩家列表数据
	game.UserList[user.ID] = user
	game.AllUserList[user.ID] = user

	// 广播下注成功消息
	betSuccessRes := msg.BetSuccessRes{
		ChairId:       user.ChairID,
		BetIndex:      req.BetIndex,
		BetCardsIndex: 0,
		UserId:        user.ID,
		BetNum:        requestAmount,
		IsAll:         req.All,
		IsReatBet:     req.RepeatBet,
	}

	// 广播下注成功信息
	game.SendBetSuccessInfo(betSuccessRes)

}

// UserConfirmBet 用户确认下注
func (game *Blackjack) UserConfirmBet(buffer []byte, userInter player.PlayerInterface) {
	// 游戏不在下注状态，退出
	if game.Status != int32(msg.GameStatus_BetStatus) && game.Status != int32(msg.GameStatus_EndBet) {
		return
	}

	// 下注用户是否在桌上
	user, ok := game.UserList[userInter.GetID()]
	if !ok {
		return
	}

	// 用户已经下注完, 重复操作
	if user.Status != int32(msg.UserStatus_UserBetSuccess) {
		return
	}

	// 确认下注
	user.Status = int32(msg.UserStatus_UserBetOver)

	// 存储user全局数据
	userData := data.UserData{
		LastBetAmount: user.BetAmount,
		NotBetCount:   0,
	}
	data.SetUserInterdata(user.User, userData)

	// 更新玩家列表数据
	game.UserList[user.ID] = user
	game.AllUserList[user.ID] = user

	// 广播下注完毕消息
	res := msg.BetOverRes{
		ChairId: user.ChairID,
		UserId:  user.ID,
	}
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CBetOverRes), &res)

	// 所有玩家都完成下注，进入下注结束
	var betOverCount int
	if game.Status == int32(msg.GameStatus_BetStatus) {
		for _, user := range game.UserList {
			if user.Status == int32(msg.UserStatus_UserBetOver) {
				betOverCount++
			}
		}

		if betOverCount == len(game.AllUserList) {
			game.table.DeleteJob(game.TimerJob)
			game.EndBet()
		}
	}

}

// UserInsure 玩家买保险
func (game *Blackjack) UserInsure(buffer []byte, userInter player.PlayerInterface) {
	log.Debugf(" 用户 %d 买保险", userInter.GetID())

	user, ok := game.UserList[userInter.GetID()]
	if !ok {
		return
	}

	insureRes := msg.BuyInsureRes{
		ChairId: user.ChairID,
		UserId:  userInter.GetID(),
		Status:  0,
	}
	// 判断当前游戏状态
	if game.Status != int32(msg.GameStatus_InsuranceStatus) {
		insureRes.ErrNum = int32(msg.ErrorList_NotInsuranceStatus)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBuyInsure), &insureRes)
		if err != nil {
			log.Errorf("send insure response fail : %v", err)
		}
		return
	}

	// 用户手牌为黑杰克，不能购买
	if user.HoldCards[0].Type == msg.CardsType_BlackJack {
		return
	}

	// 用户已经收到过买保险的请求
	if user.ReceiveInsurance {
		return
	}

	req := &msg.BuyInsureReq{}

	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Debugf("%v", err)
	}

	// 用户拥有筹码不够
	if req.Buy && user.CurAmount < user.BetAmount/2 {
		insureRes.ErrNum = int32(msg.ErrorList_LackOfFunds)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CBuyInsure), &insureRes)
		if err != nil {
			log.Errorf("send insure response fail : %v", err)
		}
		return
	}

	// 回参同时是否买保险
	insureRes.IsBuy = req.Buy

	// 购买保险金
	if req.Buy {
		user.Insurance = user.BetAmount / 2

		user.User.SetScore(game.table.GetGameNum(), -user.Insurance, game.RoomCfg.TaxRate)

		user.InitAmount -= user.BetAmount / 2
		user.CurAmount = user.CurAmount - user.BetAmount/2
		user.IsBuyInsure = true
		user.Chip += user.BetAmount / 2
	}
	user.ReceiveInsurance = true

	game.AllUserList[user.ID] = user
	game.UserList[user.ID] = user

	// 广播购买保险操作
	insureRes.ErrNum = int32(msg.ErrorList_NoError)
	insureRes.Status = 1
	game.table.Broadcast(int32(msg.SendToClientMessageType_S2CBuyInsure), &insureRes)

	if req.Buy {
		// 广播购买保险下注信息
		res := msg.BetSuccessRes{
			ChairId:       user.ChairID,
			BetIndex:      -1,
			BetCardsIndex: 0,
			UserId:        user.ID,
			BetNum:        user.Insurance,
			IsBuyInsure:   true,
		}

		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CBetSuccessMessageID), &res)
	}

	// 所有人都发送了买保险信息
	allSendInsurance := true

	for _, user := range game.UserList {
		// 跳过已经结算的玩家
		if user.Status == int32(msg.UserStatus_UserStopAction) {
			continue
		}

		if !user.ReceiveInsurance {
			allSendInsurance = false
		}
	}

	// 所有玩家发送了买保险信息,停掉买保险阶段，直接进入保险结束
	if allSendInsurance {
		game.table.DeleteJob(game.TimerJob)
		game.InsuranceEnd()
	}

}

// UserDoCmd 用户操作
func (game *Blackjack) UserDoCmd(buffer []byte, userInter player.PlayerInterface) {
	// 不再第二轮发牌操作
	if game.Status != int32(msg.GameStatus_SecondRoundFaPai) {
		return
	}

	userID := userInter.GetID()
	curUser := game.CurActionUser
	user := game.UserList[userID]

	// 可操作开关，防止重复请求
	if !curUser.ActionSwitch {
		return
	}

	req := &msg.AskDoReq{}

	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Debugf("%v", err)
	}

	log.Debugf(" cmd request %v", req)

	// 不是当前发言玩家，返回
	if userID != curUser.UserID || req.BetCardsIndex != curUser.BetCardsIndex {
		return
	}

	betIndex := curUser.BetCardsIndex

	askDoRes := msg.AskDoRes{
		ChairId:       curUser.ChairID,
		Action:        req.CmdType,
		UserId:        curUser.UserID,
		BetCardsIndex: betIndex,
	}

	// 操作不被允许
	if ok := game.CheckUserCmd(req, curUser); !ok {

		// 发送操作失败消息
		askDoRes.ErrNum = int32(msg.ErrorList_ActionFail)
		err := userInter.SendMsg(int32(msg.SendToClientMessageType_S2CAskDo), &askDoRes)
		if err != nil {
			log.Errorf("send action fail : %v", err)
		}

		return
	}

	var (
		reqAmount       int64
		userResultsResp []*msg.UserResultsRes
		faPaiOneRes     []msg.FaPaiOneRes
	)

	// 要牌或者双倍控牌
	if req.CmdType == int32(msg.AskDoType_GetPoker) || req.CmdType == int32(msg.AskDoType_DoubleBet) {
		//game.Poker.CheatCheck(user.TestCardsType, user.HoldCards[betIndex].Cards)
		game.Control(user.HoldCards[betIndex].Cards, user)
	}

	// 闲家操作
	switch req.CmdType {
	// 要牌
	case int32(msg.AskDoType_GetPoker):

		card := game.Poker.DrawCard()
		// 加一张牌
		user.AppendCard(betIndex, card, reqAmount)

		// 张牌信息
		faPaiOneRes = append(faPaiOneRes, msg.FaPaiOneRes{
			ChairId:       curUser.ChairID,
			UserId:        curUser.UserID,
			BetCardsIndex: betIndex,
			Cards:         []byte{card},
			Cardspoint:    poker.ReducePoint(user.HoldCards[betIndex].Point),
			CardType:      int32(user.HoldCards[betIndex].Type),
		})

		break

		// 分牌
	case int32(msg.AskDoType_DepartPoker):

		user.HoldCards[1] = &data.HoldCards{
			Cards:        []byte{user.HoldCards[0].Cards[1]},
			Point:        user.HoldCards[0].Point,
			Type:         user.HoldCards[0].Type,
			BetAmount:    user.HoldCards[0].BetAmount,
			ActionPermit: user.HoldCards[0].ActionPermit,
		}
		user.HoldCards[0].Cards = []byte{user.HoldCards[0].Cards[0]}

		// 闲家筹码变换
		reqAmount = user.HoldCards[betIndex].BetAmount
		user.BetAmount += reqAmount
		user.CurAmount -= reqAmount
		user.Chip += reqAmount

		// 操作两副牌
		for index := 0; index < 2; index++ {

			// 抽选的牌值等于分牌因子，重新抽牌
			var (
				card, cardValue byte
				loopCount       int
			)
			for {
				card = game.Poker.DrawCard()
				cardValue, _ = poker.GetCardValueAndColor(card)
				if cardValue != user.DepartFactor {
					break
				}

				// 防止死循环
				if loopCount >= 16 {
					break
				}
				loopCount++
			}

			user.AppendCard(int32(index), card, 0)

			// 分牌后牌型由黑杰克变21点
			if user.HoldCards[index].Type == msg.CardsType_BlackJack {
				user.HoldCards[index].Type = msg.CardsType_Point21
				user.HoldCards[index].ActionPermit = false
			}

			// 张牌信息
			faPaiOneRes = append(faPaiOneRes, msg.FaPaiOneRes{
				ChairId:       curUser.ChairID,
				UserId:        curUser.UserID,
				BetCardsIndex: int32(index),
				Cards:         []byte{card},
				Cardspoint:    poker.ReducePoint(user.HoldCards[index].Point),
				CardType:      int32(user.HoldCards[index].Type),
			})
		}

		break
		// 双倍
	case int32(msg.AskDoType_DoubleBet):

		card := game.Poker.DrawCard()

		reqAmount = user.HoldCards[betIndex].BetAmount

		user.AppendCard(betIndex, card, reqAmount)

		// 拥有注减去底注，下注池加上底注
		user.CurAmount -= reqAmount
		user.BetAmount += reqAmount
		user.Chip += reqAmount

		// 双倍之后不能再操作
		user.HoldCards[betIndex].ActionPermit = false
		user.HoldCards[betIndex].StopAction = int32(msg.StopAction_ActionDoubleBet)

		// 张牌信息
		faPaiOneRes = append(faPaiOneRes, msg.FaPaiOneRes{
			ChairId:       curUser.ChairID,
			UserId:        curUser.UserID,
			BetCardsIndex: betIndex,
			Cards:         []byte{card},
			Cardspoint:    poker.ReducePoint(user.HoldCards[betIndex].Point),
			CardType:      int32(user.HoldCards[betIndex].Type),
		})

		break

		// 停牌
	case int32(msg.AskDoType_Stand):

		user.HoldCards[betIndex].ActionPermit = false
		user.HoldCards[betIndex].StopAction = int32(msg.StopAction_ActionStand)
		break

		// 认输
	case int32(msg.AskDoType_GiveUp):
		var records []*platform.PlayerRecord

		user.HoldCards[betIndex].ActionPermit = false
		user.HoldCards[betIndex].StopAction = int32(msg.StopAction_ActionGiveUp)
		user.HoldCards[betIndex].EndType = data.EndType_GiveUp

		betAmount := user.HoldCards[betIndex].BetAmount
		// 认输，输一半
		reqAmount = betAmount / 2

		// 归还本金，减去输掉的钱
		game.UserList[userID].CurAmount += betAmount - reqAmount
		profit := game.SettleDivision(userID)

		// 发送打码量
		user.Chip -= reqAmount
		game.SetChip(userID, user.Chip)

		// 发送战绩，计算产出
		if user.User.IsRobot() == false {
			if record := game.TableSendRecord(userID, -reqAmount, profit); record != nil {
				record.Balance = user.User.GetScore()
				records = append(records, record)
			}
		}

		// 编辑用户日志
		game.WriteUserLog(user, profit)

		game.UserList[userID].BetAmount -= betAmount

		// 用户只能单副牌认输
		user.Status = int32(msg.UserStatus_UserStopAction)

		// 结算消息
		userResultsResp = append(userResultsResp, &msg.UserResultsRes{
			UserId:      curUser.UserID,
			UserWinLoss: profit,
			ChairId:     curUser.ChairID,
			Status:      user.Status,
		})

		//发送战绩
		if len(records) > 0 {
			if _, err := game.table.UploadPlayerRecord(records); err != nil {
				log.Errorf("upload player record error %s", err.Error())
			}
		}

		break
	}

	// 检测是否可继续操作
	user.PermitAction(betIndex)

	// 爆牌，有结算消息
	var records []*platform.PlayerRecord

	if user.HoldCards[betIndex].Type == msg.CardsType_Bust {
		user.HoldCards[betIndex].EndType = data.EndType_Boom
		// 第一副牌爆牌并没有第二幅牌 或者 第一，二副牌都爆牌
		if (betIndex == 0 && len(user.HoldCards[1].Cards) == 0) ||
			betIndex == 1 && user.HoldCards[0].Type == msg.CardsType_Bust {

			profit := game.SettleDivision(user.ID)

			// 发送打码量
			game.SetChip(userID, user.Chip)

			// 发送战绩，计算产出
			if user.User.IsRobot() == false {
				if record := game.TableSendRecord(userID, -user.BetAmount, profit); record != nil {
					record.Balance = user.User.GetScore()
					records = append(records, record)
				}
			}

			// 编辑用户日志
			game.WriteUserLog(user, profit)

			user.Status = int32(msg.UserStatus_UserStopAction)

			// 结算消息
			userResultsResp = append(userResultsResp, &msg.UserResultsRes{
				UserId:      curUser.UserID,
				UserWinLoss: profit,
				ChairId:     curUser.ChairID,
				Status:      user.Status,
			})
		}
	}

	if len(records) > 0 {
		if _, err := game.table.UploadPlayerRecord(records); err != nil {
			log.Errorf("upload player record error %s", err.Error())
		}
	}

	// 广播玩家操作结果
	askDoRes.Status = 1
	if user.HoldCards[0].ActionPermit || user.HoldCards[1].ActionPermit {
		askDoRes.IsAction = true
	}

	// 广播用户操作结果
	game.SendAskDoResult(askDoRes, nil)

	// 更新user 数据
	game.AllUserList[userID] = user
	game.UserList[userID] = user

	// 设置一下个 curUser 数据
	game.SetNestCurUser()

	// 停掉定时器任务
	if game.TimerJob != nil {
		game.table.DeleteJob(game.TimerJob)
	}

	// 双倍或者分牌，发送下注消息
	if req.CmdType == int32(msg.AskDoType_DepartPoker) ||
		req.CmdType == int32(msg.AskDoType_DoubleBet) {
		// 广播双倍/分牌下注信息
		res := msg.BetSuccessRes{
			ChairId:       curUser.ChairID,
			BetIndex:      -1,
			BetCardsIndex: betIndex,
			UserId:        user.ID,
			BetNum:        reqAmount,
		}

		switch req.CmdType {
		// 分牌
		case int32(msg.AskDoType_DepartPoker):
			res.IsDepart = true
			res.BetCardsIndex = 1
			break
			// 双倍
		case int32(msg.AskDoType_DoubleBet):
			res.IsDouble = true
			break
		}

		// 广播下注成功信息
		game.SendBetSuccessInfo(res)
	}

	// 有发牌信息
	if len(faPaiOneRes) > 0 {
		for _, v := range faPaiOneRes {

			// 广播发一张牌的信息
			game.SendDealCard(v)
		}
	}

	// 有结算特效动画
	if user.HoldCards[betIndex].Type != msg.CardsType_Other {

		// 有结算飘筹码动画
		if len(userResultsResp) > 0 {

			// 特效动画时间
			game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.SettleAnimation), func() {

				// 广播结算结果
				game.SendSettleInfo(userResultsResp)

				// 玩家间隔操作时间 + 飘筹码动画时间
				duration := int64(game.timeCfg.AdvanceSettle + game.timeCfg.UserActionTimeInterval)
				game.TimerJob, _ = game.table.AddTimer(duration, game.TurnUserAction)
			})

			log.Tracef("添加特效动画时间 定时器")
			return
		}

		// 玩家间隔操作时间 + 特效动画时间
		duration := int64(game.timeCfg.SettleAnimation + game.timeCfg.UserActionTimeInterval)
		game.TimerJob, _ = game.table.AddTimer(duration, game.TurnUserAction)

		log.Tracef("玩家间隔操作时间 + 特效动画时间")
		return
	}

	// 有结算飘筹码动画
	if len(userResultsResp) > 0 {

		// 飘筹码动画时间
		game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.AdvanceSettle), func() {

			// 广播结算结果
			game.SendSettleInfo(userResultsResp)

			// 玩家间隔操作时间
			game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.UserActionTimeInterval), game.TurnUserAction)
		})

		log.Tracef("飘筹码动画时间")
		return
	}

	// 玩家间隔操作时间
	game.TimerJob, _ = game.table.AddTimer(int64(game.timeCfg.UserActionTimeInterval), game.TurnUserAction)
	log.Tracef("玩家间隔操作时间")
	return
}

// UserSetCards 用户设置牌型
func (game *Blackjack) UserSetCards(buffer []byte, userInter player.PlayerInterface) {
	user, ok := game.AllUserList[userInter.GetID()]
	if !ok {
		return
	}

	if user.Status != int32(msg.UserStatus_UserSitDown) {
		return
	}

	req := &msg.TestCardsTypeReq{}

	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Debugf("%v", err)
	}

	log.Debugf("测试牌类型，%d \n", req.Type)

	user.TestCardsType = req.Type

	game.AllUserList[user.ID] = user
}
