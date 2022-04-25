package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
)

// Robot 机器人结构体
type Robot struct {
	User      player.RobotInterface
	TimerJob  *player.Job
	Cfg       config.RobotConfig
	GameLogic *Blackjack
	HoldCards [2]*data.HoldCards
	CurAmount int64
	BetAmount int64
}

// Init 初始化机器人
func (robot *Robot) Init(userInter player.RobotInterface, game table.TableHandler, robotCfg config.RobotConfig) {
	robot.User = userInter
	robot.Cfg = robotCfg
	robot.GameLogic = game.(*Blackjack)
	robot.CurAmount = userInter.GetScore()
	robot.HoldCards = [2]*data.HoldCards{{}, {}}
}

// OnGameMessage 机器人收到消息
func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {

	switch subCmd {
	// 游戏状态变更消息
	case int32(msg.SendToClientMessageType_S2CStatus):
		robot.OnGameStatus(buffer)
		break

	// 下注成功消息
	case int32(msg.SendToClientMessageType_S2CBetSuccessMessageID):
		robot.ReceiveBetSuccess(buffer)
		break

	// 发牌消息
	case int32(msg.SendToClientMessageType_S2CFaPai):
		robot.ReceiveFisrtDeal(buffer)
		break

	// 发一张牌消息
	case int32(msg.SendToClientMessageType_S2CFaPaiOne):
		robot.ReceiveDeal(buffer)
		break

	// 当前操作玩家
	case int32(msg.SendToClientMessageType_S2CCurrentSeat):
		robot.AskAction(buffer)
		break
	case int32(msg.SendToClientMessageType_S2CUserLeaveRoom):
		req := &msg.UserLeaveRoomRes{}
		if err := proto.Unmarshal(buffer, req); err != nil {
			log.Errorf("proto unmarshal bet request fail: %v", err)
			return
		}
		break
	}
}

// OnGameStatus 机器人收到游戏状态消息
func (robot *Robot) OnGameStatus(buffer []byte) {
	// 状态消息入参
	req := &msg.StatusMessageRes{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("proto unmarshal bet request fail: %v", err)
	}

	switch req.Status {

	// 下注阶段
	case int32(msg.GameStatus_BetStatus):

		robot.Bet()

		break

		//买保险
	case int32(msg.GameStatus_InsuranceStatus):

		robot.Insurance()
		break
	}
}

// Bet 机器人下注
func (robot *Robot) Bet() {

	var (
		// 下注下标; 累加器; 资金档次
		betIndex, accumulator, capitalLevel int

		// 下注数量
		betAmount int64
	)

	// 权重值
	weightsValue := rand.RandInt(1, 101)
	log.Debugf("机器人下注权重：%d", weightsValue)

	// 资金分布 // todo 暂时按照初级房来算
	if len(robot.Cfg.CapitalDivision) == 0 {
		log.Errorf("机器人没有加载到配置")
		return
	}
	capitalDivision := robot.Cfg.CapitalDivision[0]

	// 确定机器人资金档次
	for k, v := range capitalDivision {
		if k == len(capitalDivision)-1 {
			capitalLevel = k
			break
		}

		if robot.CurAmount <= v {
			capitalLevel = k
			break
		}
	}

	log.Debugf("机器人资金档次：%d", capitalLevel)

	// 下注权重分布
	for k, v := range robot.Cfg.BetPlace[capitalLevel] {
		downLimit := accumulator
		upLimit := accumulator + v

		accumulator = upLimit
		if weightsValue > downLimit && weightsValue <= upLimit {
			betIndex = k
			break
		}
	}

	// 下注请求参数
	req := msg.BetReq{}

	for {

		// 全压
		if betIndex == 4 {
			if robot.CurAmount > robot.GameLogic.RoomCfg.MaxAction {
				betAmount = robot.GameLogic.RoomCfg.MaxAction
			} else {
				betAmount = robot.CurAmount
			}

			req.All = true
		} else {

			// 押注选项
			betAmount = robot.GameLogic.RoomCfg.ActionOption[betIndex]

			req.All = false
			req.BetIndex = int32(betIndex)
		}

		// 资金不够以此往前延顺下注
		if robot.CurAmount > betAmount {
			break
		}

		betIndex = betIndex - 1

		// 防止死循环
		if betIndex < 0 {
			break
		}
	}

	// 机器人资金变动
	robot.CurAmount -= betAmount
	robot.BetAmount = betAmount

	// 随机时间
	randomTime := rand.RandInt(robot.Cfg.ActionTime.Shortest, robot.Cfg.ActionTime.Longest)
	log.Debugf("机器人下注时间延迟：%d", randomTime)

	// 延迟发送消息
	robot.User.AddTimer(int64(randomTime), func() {
		// 请求server买保险
		err := robot.User.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBet), &req)
		if err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
		robot.BetConfirm()
	})
}

// BetConfirm 机器人确认下注
func (robot *Robot) BetConfirm() {
	req := msg.BetOverReq{
		IsOk: true,
	}

	robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBetOver), &req)
}

// Insurance 机器人买保险
func (robot *Robot) Insurance() {

	// 确认买保险
	var confirmInsurance bool

	// 手牌为黑杰克，不能买保险
	if robot.HoldCards[0].Type == msg.CardsType_BlackJack {
		return
	}

	// 权重值
	weightsValue := rand.RandInt(1, 101)
	log.Debugf("机器人买保险权重：%d", weightsValue)

	// todo 暂时按照初级房来算
	insuranceDivision := robot.Cfg.InsuranceDivision[0]

	// 持有注 / 初始注 * 100%
	betProportion := int(robot.CurAmount / robot.BetAmount * 100)

	// 持有住与初始注比例档次
	var insuranceLevel int

	for k, v := range insuranceDivision {

		if k == len(insuranceDivision)-1 {
			insuranceLevel = k
			break
		}

		if betProportion <= v {
			insuranceLevel = k
			break
		}
	}

	if weightsValue <= robot.Cfg.InsurancePlace[0][insuranceLevel] {
		confirmInsurance = true
	}

	// 测试机器人都买保险
	//confirmInsurance = true

	// 筹码不够，不能买保险
	if robot.CurAmount < robot.BetAmount/2 {
		confirmInsurance = false
	}

	// 机器人资金变动
	if confirmInsurance {
		robot.CurAmount -= robot.BetAmount / 2
	}

	// 买保险请求参数
	req := msg.BuyInsureReq{
		Buy: confirmInsurance,
	}

	robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SBuyInsure), &req)
}

// AskAction 机器人第二轮操作
func (robot *Robot) AskAction(buffer []byte) {
	// 当前玩家信息
	res := msg.CurrentSeatRes{}
	if err := proto.Unmarshal(buffer, &res); err != nil {
		log.Errorf("proto unmarshal current user response fail: %v", err)
		return
	}

	// 当前操作玩家是机器人
	if res.UserId == robot.User.GetID() {
		if robot.SelectDepartPoker(res) {
			return
		}

		if robot.SelectDouble(res) {
			return
		}

		if robot.SelectGetPoker(res) {
			return
		}

		robot.SelectStand(res)
	}
}

// SelectDepartPoker 机器人考虑分牌
func (robot *Robot) SelectDepartPoker(res msg.CurrentSeatRes) (doAction bool) {
	if res.DepartPoker {
		// 权重值
		weightsValue := rand.RandInt(1, 101)
		log.Debugf("机器人分牌权重：%d", weightsValue)

		// todo 暂时按照初级房来算
		DepartDivision := robot.Cfg.DepartDivision[0]

		// 最接近21点并且比21点小的值
		point := poker.GetNearPoint21(robot.HoldCards[res.BetCardsIndex].Point)
		log.Debugf("机器人当前牌值：%d", point)

		// 持有住与初始注比例档次
		var pointLevel int

		for k, v := range DepartDivision {

			if k == len(DepartDivision)-1 {
				pointLevel = k
				break
			}

			if point <= v {
				pointLevel = k
				break
			}
		}

		// 分牌概率中标
		if weightsValue <= robot.Cfg.DepartPlace[0][pointLevel] {
			// 机器人资金变动
			robot.CurAmount -= robot.HoldCards[0].BetAmount
			robot.BetAmount += robot.HoldCards[0].BetAmount

			doAction = true

			req := msg.AskDoReq{
				CmdType:       int32(msg.AskDoType_DepartPoker),
				BetCardsIndex: res.BetCardsIndex,
			}

			robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SAskDo), &req)
		}
	}
	return
}

// SelectDouble 机器人考虑加倍
func (robot *Robot) SelectDouble(res msg.CurrentSeatRes) (doAction bool) {
	if res.DoubleBet {
		// 权重值
		weightsValue := rand.RandInt(1, 101)
		log.Debugf("机器人加倍权重：%d", weightsValue)

		// 最接近21点并且比21点小的值
		point := poker.GetNearPoint21(robot.HoldCards[res.BetCardsIndex].Point)
		log.Debugf("机器人当前牌值：%d", point)

		weightsDivision := robot.Cfg.DoublePlace.Less9Point

		// 加倍概率计算
		if point < 9 && point > 0 {
			weightsDivision = robot.Cfg.DoublePlace.Less9Point
		} else if point >= 9 && point <= 11 {
			weightsDivision = robot.Cfg.DoublePlace.Less11Point
		} else if point > 11 {
			weightsDivision = robot.Cfg.DoublePlace.More11Point
		}

		if weightsValue <= weightsDivision {
			// 机器人资金变动
			robot.CurAmount -= robot.HoldCards[0].BetAmount
			robot.BetAmount += robot.HoldCards[0].BetAmount

			doAction = true

			req := msg.AskDoReq{
				CmdType:       int32(msg.AskDoType_DoubleBet),
				BetCardsIndex: res.BetCardsIndex,
			}

			robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SAskDo), &req)
		}
	}
	return
}

// SelectGetPoker 机器人考虑加要牌
func (robot *Robot) SelectGetPoker(res msg.CurrentSeatRes) (doAction bool) {
	if res.GetPoker {
		// 权重值
		weightsValue := rand.RandInt(1, 101)

		log.Debugf("机器人要牌权重：%d", weightsValue)

		// 最接近21点并且比21点小的值
		point := poker.GetNearPoint21(robot.HoldCards[res.BetCardsIndex].Point)
		log.Debugf("机器人当前牌值：%d", point)

		// 要牌点数分布 // todo 暂时按照初级房来算
		getPokerDivision := robot.Cfg.GetPokerDivision[0]

		// 点数档次
		var pointLevel int

		for k, v := range getPokerDivision {

			if k == len(getPokerDivision)-1 {
				pointLevel = k
				break
			}

			if point <= v {
				pointLevel = k
				break
			}
		}

		// todo 暂时按照初级房来算
		if weightsValue <= robot.Cfg.GetPokerPlace[0][pointLevel] {
			doAction = true

			req := msg.AskDoReq{
				CmdType:       int32(msg.AskDoType_GetPoker),
				BetCardsIndex: res.BetCardsIndex,
			}

			robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SAskDo), &req)
		}
	}
	return
}

// SelectStand 机器人停牌
func (robot *Robot) SelectStand(res msg.CurrentSeatRes) {
	req := msg.AskDoReq{
		CmdType:       int32(msg.AskDoType_Stand),
		BetCardsIndex: res.BetCardsIndex,
	}

	robot.SendMsgToServer(int32(msg.ReceiveMessageType_C2SAskDo), &req)
}

// ReceiveFisrtDeal 收到第一轮发牌
func (robot *Robot) ReceiveFisrtDeal(buffer []byte) {
	// 发牌信息
	res := msg.FaPaiRes{}
	if err := proto.Unmarshal(buffer, &res); err != nil {
		log.Errorf("proto unmarshal deal response fail: %v", err)
		return
	}

	// 设置牌组信息
	for _, v := range res.UserPaiInfoArr {
		if v.UserId == robot.User.GetID() {
			robot.HoldCards[0] = &data.HoldCards{
				Cards:     v.CardArr,
				Type:      msg.CardsType(v.CardType),
				Point:     v.Cardspoint,
				BetAmount: robot.BetAmount,
			}
		}
	}
}

// ReceiveDeal 收到后续发牌信息
func (robot *Robot) ReceiveDeal(buffer []byte) {
	// 发牌信息
	res := msg.FaPaiOneRes{}
	if err := proto.Unmarshal(buffer, &res); err != nil {
		log.Errorf("proto unmarshal deal response fail: %v", err)
		return
	}

	// ID一致，给机器人发牌了
	if res.UserId == robot.User.GetID() {
		robot.HoldCards[res.BetCardsIndex].Cards = res.Cards
		robot.HoldCards[res.BetCardsIndex].Type = msg.CardsType(res.CardType)
		robot.HoldCards[res.BetCardsIndex].Point = res.Cardspoint
	}
}

// ReceiveBetSuccess 收到成功下注消息
func (robot *Robot) ReceiveBetSuccess(buffer []byte) {
	// 下注成功信息
	res := msg.BetSuccessRes{}
	if err := proto.Unmarshal(buffer, &res); err != nil {
		log.Errorf("proto unmarshal bet Success response fail: %v", err)
		return
	}

	// ID一致，当前机器人注池变化成功
	if res.UserId == robot.User.GetID() {
		robot.HoldCards[res.BetCardsIndex].BetAmount = robot.HoldCards[res.BetCardsIndex].BetAmount + res.BetNum
		robot.BetAmount = robot.BetAmount + res.BetNum
		robot.CurAmount = robot.CurAmount - res.BetNum
	}
}

// ReceiveSettle 收到结算消息
func (robot *Robot) ReceiveSettle(buffer []byte) {
	// 结算信息
	res := msg.SettleMsgRes{}
	if err := proto.Unmarshal(buffer, &res); err != nil {
		log.Errorf("proto unmarshal settle response fail: %v", err)
		return
	}

	// 设置牌组信息
	for _, v := range res.UserInfos {
		if v.UserId == robot.User.GetID() {

			// 赢
			if v.UserWinLoss >= 0 {
				robot.CurAmount += v.UserWinLoss + robot.BetAmount
			}

		}
	}
}

// SendMsgToServer 发送消息给服务端
func (robot *Robot) SendMsgToServer(C2SMessageType int32, pb proto.Message) {

	// 随机时间
	randomTime := rand.RandInt(robot.Cfg.ActionTime.Shortest, robot.Cfg.ActionTime.Longest)

	log.Debugf("机器人操作时间延迟：%d", randomTime)
	// 延迟发送消息
	robot.User.AddTimer(int64(randomTime), func() {
		// 请求server买保险
		err := robot.User.SendMsgToServer(C2SMessageType, pb)
		if err != nil {
			log.Errorf("send server msg fail: %v", err.Error())
		}
	})
}
