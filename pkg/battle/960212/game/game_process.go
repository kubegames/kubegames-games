package game

import (
	"common/log"
	"fmt"
	"game_poker/doudizhu/msg"
	"game_poker/doudizhu/poker"
	"time"

	"github.com/golang/protobuf/proto"
)

// Start 开始游戏逻辑
func (game *DouDizhu) Start() {

	game.Status = int32(msg.GameStatus_ReadyStatus)
	log.Tracef("游戏 %d 开始", game.Table.GetId())

	// 更改用户退出权限
	game.SetExitPermit(false)

	// 通知框架开赛
	game.Table.StartGame()

	// 初始化桌子
	game.InitTable()

	// 发送场景消息
	game.SendSceneInfo(nil, false)

	game.DealCards()
}

// DealCards 发牌阶段
func (game *DouDizhu) DealCards() {

	game.Status = int32(msg.GameStatus_DealStatus)
	log.Tracef("游戏 %d 开始发牌", game.Table.GetId())

	// 初始化牌组
	game.Poker = new(poker.GamePoker)
	game.Poker.InitPoker()

	// 如果玩家已经配牌，从牌库中删除已经配置的牌
	for _, user := range game.UserList {
		if len(user.Cards) != 0 {
			for _, card := range user.Cards {
				for i, waitDelCard := range game.Poker.Cards {
					if card == waitDelCard {
						game.Poker.Cards = append(game.Poker.Cards[:i:i], game.Poker.Cards[i+1:]...)
						break
					}
				}
			}
		}
	}

	// 洗牌
	game.Poker.ShuffleCards()

	// 选出底牌
	game.bottomCards = game.Poker.DrawCard(3)

	// 控牌
	//game.ControlPoker()

	// todo 如果玩家已经配牌，从牌库中删除已经配置的牌
	//for _, user := range game.UserList {
	//	if len(user.Cards) != 0 {
	//		for _, card := range user.Cards {
	//			for i, waitDelCard := range game.Poker.Cards {
	//				if card == waitDelCard {
	//					game.Poker.Cards = append(game.Poker.Cards[:i:i], game.Poker.Cards[i+1:]...)
	//					break
	//				}
	//			}
	//		}
	//	}
	//}

	for id, user := range game.UserList {
		if len(user.Cards) == 0 {
			cards := game.Poker.DrawCard(17)

			game.UserList[id].Cards = cards
		}

		//cardsStr := poker.CardsToString(user.Cards)

		// 记录玩家日志
		//game.Table.WriteLogs(user.ID, " 用户ID： "+fmt.Sprintf(`%d`, user.ID)+
		//	" 手牌： "+cardsStr)

		dealResp := msg.DealRes{
			Cards:   game.UserList[id].Cards,
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.SendDealInfo(dealResp, user.User)
	}

	// 生成抢地主座位列表
	game.SetRobChairList()

	// 抢地主两局无结果, 跳过抢地主, 指定首家为地主
	if game.RobCount >= 2 {
		game.Dizhu = game.Chairs[game.RobChairList[0]]
		game.CurRobNum = 1
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.DealAnimation), game.confirmDizhu)
	} else {
		game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.DealAnimation), game.RobDizhu)
	}

}

// RobDizhu 抢地主阶段
func (game *DouDizhu) RobDizhu() {

	game.Status = int32(msg.GameStatus_RobStatus)
	log.Tracef("游戏 %d 开始抢地主", game.Table.GetId())

	// 广播抢地主状态
	game.SendGameStatus(game.Status, 0, nil)

	// 重置玩家抢地主数值
	for ID, _ := range game.UserList {
		game.UserList[ID].RobNum = -1
	}

	// 重置当前最高抢庄倍数
	game.CurRobNum = 0

	// 游戏抢地主次数自加
	game.RobCount++

	// 广播当前抢地主玩家，
	game.curRobberChairID = game.RobChairList[0]
	game.SendCurrentRobber(int32(game.RobChairList[0]))

	// 定时进入抢庄检查
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.RobTime), game.CheckRob)
}

// turnRob 抢庄检查
func (game *DouDizhu) CheckRob() {

	// 当前抢地主玩家
	user := game.Chairs[game.curRobberChairID]

	// 当前抢地主玩家未发送抢地主请求
	if user.RobNum < 0 {

		// 封装入参，系统代替发送不抢
		robReq := &msg.RobReq{
			RobNum: 0,
		}
		buffer, err := proto.Marshal(robReq)
		if err != nil {
			log.Errorf("proto marshal fail : %v", err.Error())
			return
		}

		game.UserRobDizhu(buffer, user.User)
	}
}

// confirmDizhu 确认地主阶段
func (game *DouDizhu) confirmDizhu() {

	game.Status = int32(msg.GameStatus_confirmDizhuStatus)
	log.Tracef("游戏 %d 确认地主", game.Table.GetId())

	// 底牌加入地主手牌
	game.Dizhu.Cards = append(game.Dizhu.Cards, game.bottomCards...)

	// 确认地主
	game.Dizhu.IsDizhu = true

	// 王牌数量
	kingCount := poker.GetKingCount(game.bottomCards)

	// 底牌倍数
	if kingCount != 0 {
		game.BottomMultiple = int64(kingCount) * 2
	}

	resp := msg.ConfirmDizhuRes{
		UserId:      game.Dizhu.ID,
		ChairId:     game.Dizhu.ChairID,
		BottomCards: game.bottomCards,
		RobNum:      game.CurRobNum,
	}

	// 广播确认地主消息
	game.SendConfirmDizhu(resp)

	// 记录第一部分日志
	game.RecordLogOne()

	// 确认地主动画后进入加倍阶段
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.ConfirmDizhuTime), game.Redouble)

}

// Redouble 加倍阶段
func (game *DouDizhu) Redouble() {

	game.Status = int32(msg.GameStatus_RedoubleStatus)
	log.Tracef("游戏 %d 进入加倍阶段", game.Table.GetId())

	// 广播加倍状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.RedoubleTime)/1000, nil)

	// 加倍时间后进入加倍结束
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.RedoubleTime), game.EndRedouble)
}

// EndRedouble 结束加倍阶段
func (game *DouDizhu) EndRedouble() {

	// 未发送加倍请求玩家默认发送不加倍请求
	for _, user := range game.UserList {
		if user.AddNum == 0 {
			redoubleReq := &msg.RedoubleReq{
				AddNum: 1,
			}

			buffer, err := proto.Marshal(redoubleReq)
			if err != nil {
				log.Errorf("proto marshal fail : %v", err.Error())
				return
			}

			game.UserRedouble(buffer, user.User)
		}
	}

}

// PutCard 出牌阶段
func (game *DouDizhu) PutCards() {

	game.Status = int32(msg.GameStatus_PutCardStatus)
	log.Tracef("游戏 %d 进入出牌阶段", game.Table.GetId())

	game.SendGameStatus(game.Status, 0, nil)

	// 游戏轮转计数器自加
	game.StepCount++

	// 更新当前玩家
	game.CurrentPlayer = CurrentPlayer{
		UserID:     game.Dizhu.ID,
		ChairID:    game.Dizhu.ChairID,
		ActionTime: game.TimeCfg.OperationTime,
		Permission: true,
		StepCount:  0,
		ActionType: int32(msg.UserActionType_PutCard),
	}

	// 广播当前玩家
	game.SendCurrentPlayer()

	// 定时检查
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.OperationTime), game.CheckAction)
}

// CheckAction 检查操作
func (game *DouDizhu) CheckAction() {

	log.Tracef("系统轮转数 %d, 当前玩家轮转数 %d，当前玩家%d", game.StepCount, game.CurrentPlayer.StepCount, game.CurrentPlayer.UserID)
	//  轮转计数器不一值, 当前玩家超时, 系统代替操作, 并且进入托管状态
	if game.StepCount != game.CurrentPlayer.StepCount {

		curUser := game.UserList[game.CurrentPlayer.UserID]

		putCardsReq := &msg.PutCardsReq{
			Cards: []byte{},
		}

		// 变成托管状态
		if curUser.Status != int32(msg.UserStatus_UserHangUp) {

			// 更新玩家状态为托管
			game.UserList[curUser.ID].Status = int32(msg.UserStatus_UserHangUp)

			// 广播玩家进入托管
			game.SendHangUpInfo(curUser.ID)

			// 操作类型为出牌默认出最小的牌
			if game.CurrentPlayer.ActionType == int32(msg.UserActionType_PutCard) {
				putCardsReq.Cards = []byte{poker.GetSmallestCard(curUser.Cards)}
			}

		} else {
			var handIsDizhu bool

			// 有当前牌权玩家，并且当前牌权玩家是地主
			if game.CurrentCards.UserID != 0 && game.UserList[game.CurrentCards.UserID].IsDizhu {
				handIsDizhu = true
			}

			putCardsReq.Cards = poker.HangUpPutCards(game.CurrentCards, curUser, handIsDizhu)

			log.Tracef("玩家托管出牌: %s", fmt.Sprintf("%+v\n", putCardsReq.Cards))
		}

		buffer, err := proto.Marshal(putCardsReq)
		if err != nil {
			log.Errorf("proto marshal fail : %v", err.Error())
			return
		}

		game.UserPutCards(buffer, curUser.User)
	}

}

// Settle 结算阶段
func (game *DouDizhu) Settle() {
	game.Status = int32(msg.GameStatus_SettleStatus)
	log.Tracef("游戏 %d 进入结算阶段", game.Table.GetId())

	// 记录第二部分游戏日志
	game.RecordLogTwo()

	// 广播结算状态
	game.SendGameStatus(game.Status, int32(game.TimeCfg.SettleTime)/1000, nil)

	var (
		DizhuWin              bool      // 是否地主赢
		allOffMultiple        int64 = 1 // 春天倍数
		beAllOffMultiple      int64 = 1 // 反春倍数
		commonMultiple        int64     // 公共倍数
		totalPeasantsMultiple int64     // 农民总倍数
	)

	// 定输赢
	for _, user := range game.UserList {

		if user.IsDizhu {
			if len(user.Cards) == 0 {
				DizhuWin = true
			}
		} else {
			totalPeasantsMultiple += user.AddNum
		}
	}

	game.TotalPeasantsMultiple = totalPeasantsMultiple

	// 春天检测
	if game.IsAllOff() {
		allOffMultiple = 2
		game.AllOffMultiple = 2
	}

	// 反春检测
	if game.IsBeAllOff() {
		beAllOffMultiple = 2
		game.BeAllOffMultiple = 2
	}

	// 公共倍数 = 抢分倍数 * 地分倍数 * 炸弹倍数 * 火箭倍数 * 春天倍数 * 反春倍数
	commonMultiple = game.CurRobNum *
		game.BottomMultiple *
		game.BoomMultiple *
		game.RocketMultiple *
		allOffMultiple *
		beAllOffMultiple

	//game.CommonMultiple = commonMultiple

	for ID, user := range game.UserList {
		if user.IsDizhu {
			game.UserList[ID].TotalMultiple = commonMultiple * user.AddNum * totalPeasantsMultiple
		} else {
			game.UserList[ID].TotalMultiple = commonMultiple * user.AddNum * game.Dizhu.AddNum
		}
	}

	for ID, user := range game.UserList {
		result := user.TotalMultiple * game.RoomCfg.RoomCost

		if (user.IsDizhu && !DizhuWin) || (!user.IsDizhu && DizhuWin) {
			result *= -1
		}

		game.UserList[ID].SettleResult = result
	}

	//  结算细节（防以小博大）
	game.SettleDetail()

	// 结算上下分
	game.SettleDivision()

	// 记录第三部分游戏日志
	game.RecordLogThree()

	// 广播结算消息
	game.SendSettleInfo()

	// 结算时间后游戏结束
	game.TimerJob, _ = game.Table.AddTimer(time.Duration(game.TimeCfg.SettleTime), game.GameOver)
}

// GameOver 游戏结束
func (game *DouDizhu) GameOver() {

	// 重置桌面属性
	game.Status = int32(msg.GameStatus_GameOver)
	log.Tracef("游戏 %d 结束", game.Table.GetId())

	// 更改用户退出权限
	game.SetExitPermit(true)

	// 通知框架游戏结束
	game.Table.EndGame()

	game.SendGameStatus(game.Status, 0, nil)

	game.Poker = nil
	game.TimerJob = nil
	game.RobotTimer = nil
	game.TimeCfg = nil
	game.GameCfg = nil
	game.RoomCfg = nil
	game.LoadCfg = false
	game.CurrentPlayer = CurrentPlayer{}
	game.CurrentCards = poker.HandCards{}
	game.LeftCards = []byte{}

	// 所有人退出去
	for _, user := range game.UserList {

		game.UserExit(user.User)
		game.Table.KickOut(user.User)
	}
}
