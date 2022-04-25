package game

import (
	"fmt"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/poker"

	"github.com/golang/protobuf/proto"
)

//获取房间信息
func (game *Game) ProcGetRoomInfo(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SRoomInfo
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetRoomInfo err : %v", err.Error())
		return
	}
	_ = user.User.SendMsg(int32(msg.S2CMsgType_ROOM_INFO_RES), game.GetRoomInfo2C(user))
	return
}

// del by wd in 2020.3.12 删除客户端自定义发给服务端都匹配消息
//开始匹配
//func (game *Game) ProcStartMatch(buffer []byte, user *data.User) {
//	if game.Status == global.TABLE_CUR_STATUS_ING {
//		fmt.Println("游戏正在进行中不能请求机器人")
//		return
//	}
//	var c2sMsg msg.C2SStartMatch
//	err := proto.Unmarshal(buffer, &c2sMsg)
//	if err != nil {
//		log.Debugf("ProcSetCards err : %v", err.Error())
//		return
//	}
//	//game.Table.AddTimer(3*1000, func() {
//	game.Table.AddTimer(3*1000, func() {
//		if game.Status != global.TABLE_CUR_STATUS_ING {
//			_ = game.Table.GetRobot(3)
//		}
//	})
//	_ = user.User.SendMsg(int32(msg.S2CMsgType_START_MATCH_RES), &c2sMsg)
//}

// add by wd in 2020.3.12 添加机器人坐下逻辑
// MatchRobot 机器人坐下
func (game *Game) MatchRobot() {
	if game.Status != global.TABLE_CUR_STATUS_WAIT {
		return
	}

	// 坐下机器人个数
	robotCount := int32(4 - game.GetRoomUserCount())
	log.Tracef("房间 %d 坐下 %d 个机器人", game.Id, robotCount)

	if robotCount != 0 {
		_ = game.Table.GetRobot(robotCount)
	}

	// 补足机器人，防止有人卡时间退出
	game.robotTimer, _ = game.Table.AddTimer(int64(500), game.MatchRobot)
}

//用户确定手动摆牌
func (game *Game) ProcSetCards(buffer []byte, user *data.User) {
	//if game.Status != global.TABLE_CUR_STATUS_ING {
	//	fmt.Println("游戏没在进行中")
	//	return
	//}
	var c2sMsg msg.C2SSetCards
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcSetCards err : %v", err.Error())
		return
	}
	s2cSettleMsg := &msg.S2CSettleCards{
		Uid:         user.User.GetID(),
		ChairId:     user.ChairId,
		SpecialType: user.SpecialCardType,
	}
	if user.IsSettleCards {
		fmt.Println("用户已经确定摆过牌了")
		return
	}
	if c2sMsg.IsAuto || c2sMsg.IsSpecial {
		user.IsAuto = true
		//确定自动摆牌
		game.Table.Broadcast(int32(msg.S2CMsgType_SETTLE_CARDS), s2cSettleMsg)
		//检查如果所有人都手动摆牌ok，就结束比赛
		user.IsSettleCards = true
		game.AllUserSettleEndGame()
		return
	}
	c2sCards := append(c2sMsg.HeadCards, c2sMsg.MidCards...)
	c2sCards = append(c2sCards, c2sMsg.TailCards...)
	//fmt.Println("用户传过来的所有牌: ", fmt.Sprintf(`%x`, c2sCards), " 自己原来的牌：", fmt.Sprintf(`%x`, user.Cards))
	//检查传过来的牌是否符合规则
	if len(c2sMsg.HeadCards) != 3 || len(c2sMsg.MidCards) != 5 || len(c2sMsg.TailCards) != 5 {
		fmt.Println("传过来的牌不符合规则: ", c2sMsg.HeadCards, c2sMsg.MidCards, c2sMsg.TailCards)
		return
	}
	//检查传过来的牌是否有两个完全相同的
	for i := 0; i < len(c2sCards)-1; i++ {
		for j := i + 1; j < len(c2sCards); j++ {
			if c2sCards[i] == c2sCards[j] {
				fmt.Println("用户传过来的牌有两张一样的，", fmt.Sprintf(`%x`, c2sCards[i]))
				return
			}
		}
	}
	//检查传过来的牌是否是用户原来的牌
	for _, card := range c2sCards {
		if !user.IsCardInUserCards(card) {
			fmt.Println("用户传过来的牌没在用户原来的牌中，", fmt.Sprintf(`%x`, card), "  ", fmt.Sprintf(`%x`, user.Cards))
			return
		}
	}
	midCardsArr := poker.Cards5SliceToArr(c2sMsg.MidCards)
	tailCardsArr := poker.Cards5SliceToArr(c2sMsg.TailCards)
	if isBeat, _, _ := user.Compare5Cards(midCardsArr, tailCardsArr); isBeat == global.COMPARE_WIN {
		fmt.Println("中墩比尾墩更大，不合规则：", fmt.Sprintf(`%x , %x`, c2sMsg.MidCards, c2sMsg.TailCards))
		_ = user.User.SendMsg(global.ERROR_CODE_CARD_NOT_RULE, &msg.S2CHitRob{})
		return
	}
	if isBeat := user.Compare5And3Cards(c2sMsg.MidCards, c2sMsg.HeadCards); !isBeat {
		_ = user.User.SendMsg(global.ERROR_CODE_CARD_NOT_RULE, &msg.S2CHitRob{})
		fmt.Println("头墩比中墩更大，不合规则：", fmt.Sprintf(`%x , %x`, c2sMsg.HeadCards, c2sMsg.MidCards))
		return
	}
	user.IsSettleCards = true
	user.HeadCards = c2sMsg.HeadCards
	user.HeadCardType, _ = poker.GetCardType13Water(c2sMsg.HeadCards)

	user.MiddleCards = c2sMsg.MidCards
	user.MidCardType, _ = poker.GetCardType13Water(c2sMsg.MidCards)

	user.TailCards = c2sMsg.TailCards
	user.TailCardType, _ = poker.GetCardType13Water(c2sMsg.TailCards)
	if !c2sMsg.IsSpecial {
		user.SpecialCardType = data.SPECIAL_CARD_NO
	}
	//user.SetSpecialCardType()

	fmt.Println("手动摆牌ok", user.User.GetID())
	game.Table.Broadcast(int32(msg.S2CMsgType_SETTLE_CARDS), s2cSettleMsg)
	//检查如果所有人都手动摆牌ok，就结束比赛
	game.AllUserSettleEndGame()
	return
}

//todo 最后要删除
func (game *Game) ProcUserSelectCards(buffer []byte, user *data.User) {
	//if game.Status != global.TABLE_CUR_STATUS_ING {
	//	fmt.Println("游戏没在进行中")
	//	return
	//}
	var c2sMsg msg.C2SUserSelectCards
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcSetCards err : %v", err.Error())
		return
	}
	if len(c2sMsg.Cards) != 13 {
		fmt.Println("c2sMsg 长度不对 ", len(c2sMsg.Cards))
		return
	}
	deck := []byte{
		0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
		0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
		0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
		0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
	}

	// 存在检测, 重复检测
	for _, card := range c2sMsg.Cards {
		var (
			sameCount int  // 重复个数
			isExit    bool // 是否存在
		)

		for _, card1 := range c2sMsg.Cards {
			if card == card1 {
				sameCount++
			}

			if sameCount >= 2 {
				log.Warnf("配牌有重复牌 %v", card)
				return
			}
		}

		for _, waitTakeCard := range deck {
			if card == waitTakeCard {
				isExit = true
			}
		}

		if !isExit {
			log.Warnf("配牌中有不存在的牌 %v", card)
			return
		}
	}

	for i := 0; i < 13; i++ {
		user.Cards[i] = c2sMsg.Cards[i]
	}
	user.IsTest = true
}
