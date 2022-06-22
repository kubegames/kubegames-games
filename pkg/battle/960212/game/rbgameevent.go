package game

import (
	"reflect"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960212/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func (rb *robot) processPlayerInfos() {
	//rb.User.Cards = rb.robotPlayer.Cards
}

func (rb *robot) processTableInfo() {
	//var bnewstep = rb.gameStep != tableinfo.Gamestep
	//	rb.gameStep = rb.GameLogic.Status
	//	rb.GameLogic.Dizhu.ChairID = rb.GameLogic.Dizhu.ChairID
	//rb.currentSeat = rb.GameLogic.CurrentPlayer.ChairID
	//rb.maxcalltype = rb.GameLogic.CurRobNum
	//rb.threeCards = nil
}

func (rb *robot) processGameStart() {
	rb.processTableInfo()
	rb.processPlayerInfos()
	//	rb.currentSeat = gamestart.Seat
	//	rb.deskLeftCards = CreateCards54Ex()
	//剩余牌减少，减掉自己的手牌
	//_, rb.deskLeftCards = RemoveCards(rb.deskLeftCards, rb.User.Cards)
	//只有第一个叫地主的玩家才有动画时间的延迟
	// rb.gameCfg.SendCardAni = 0
	// if rb.mySeat == gamestart.Seat {
	// 	rb.gameCfg.SendCardAni = 2000
	// }
	rb.checkRunTree()
}

func (rb *robot) processGameRound() {
	//	rb.currentSeat = rb.GameLogic.CurrentPlayer.ChairID
	//每个游戏阶段，不同处理
	switch rb.GameLogic.Status {
	case int32(msg.GameStatus_RobStatus):
		//rb.DealCurrentRobber()
		//rb.checkRunTree()
	case int32(msg.GameStatus_RedoubleStatus):
		//加倍动作在桌子信息里面处理
		rb.checkRunTree()
		/*case int32(msg.GameStatus_PutCardStatus):
		rb.checkRunTree()
		*/
	}
}

func (rb *robot) processCallBanker() {

}
func (rb *robot) processRobBanker() {

}

func (rb *robot) processEnsureBanker() {
	//	rb.GameLogic.Dizhu.ChairID = rb.GameLogic.Dizhu.ChairID
	//rb.maxcalltype = rb.GameLogic.CurRobNum
}

func (rb *robot) processPlayerDouble() {
	//玩家加倍信息
}

func (rb *robot) processStartOutCard() {
	//	rb.currentSeat = rb.GameLogic.CurrentPlayer.ChairID
	//	rb.firstOutSeat = rb.currentSeat
	//当前回合最大出牌玩家
	//rb.maxOutSeat = -1
	//当前回合最大出牌牌组
	//	rb.GameLogic.CurrentCards.Cards = nil
	//玩家pass信息重置
	rb.checkRunTree()
}

func (rb *robot) processOutCard() {
}

func (rb *robot) processGameEnd() {

}

func (rb *robot) doOut3x(t int) {

}
func (rb *robot) doCallBanker(call bool) {
	rb.setRobotTimer(rand.RandInt(500, 5000), func() {
		//TODO
		/*var msg gameddz.CallBanker
		msg.Seat = rb.mySeat
		msg.Yes = call
		event := common.MakeGameEvent(rb.robotID, int32(gameddz.Event_CALL_BANKER_REQ), &msg)
		rb.mgr.PostEvent(*event, false)
		wplog.Tracef(" ---------- robot doCallBanker: ", msg.Seat, msg.Yes)
		*/
	})
}

func (rb *robot) doRobBanker(rob bool) {
	rb.setRobotTimer(rand.RandInt(500, 5000), func() {
		//TODO
		/*
			var msg gameddz.RobBanker
			msg.Seat = rb.mySeat
			msg.Yes = rob
			event := common.MakeGameEvent(rb.robotID, int32(gameddz.Event_ROB_BANKER_REQ), &msg)
			rb.mgr.PostEvent(*event, false)
			wplog.Tracef(" ---------- robot doRobBanker: ", msg.Seat, msg.Yes)
		*/
	})
}
func (rb *robot) doDouble(t int) {
	rb.setRobotTimer(rand.RandInt(500, 5000), func() {
		//TODO
		/*
			var msg gameddz.PlayerDouble
			msg.Seat = rb.mySeat
			if t == 0 {
				msg.Isdouble = false
				msg.IsSuperDouble = false
			} else if t == 1 {
				msg.Isdouble = true
				msg.IsSuperDouble = false
			} else {
				msg.Isdouble = false
				msg.IsSuperDouble = true
			}

			event := common.MakeGameEvent(rb.robotID, int32(gameddz.Event_PLAYER_msg.CardsType_Pair_REQ), &msg)
			rb.mgr.PostEvent(*event, false)
			wplog.Tracef(" ---------- robot doDouble: ", msg.Seat, msg.Isdouble)
		*/
	})
}

func (rb *robot) sendOutCardmsg(option int, cards []byte) {
	rb.sendOutCardmsgEx(rand.RandInt(1200, 2800), option, cards)
}

func (rb *robot) sendOutCardmsgEx(ms int, option int, cards []byte) {
	log.Tracef(" ---------- robot outcard start: %d, %v, %v", ms, option, cards)
	if option == 1 {
		rb.PutCards(cards)
	} else {
		rb.PutCards([]byte{})
	}
	rb.setRobotTimer(ms, func() {
		//TODO
		/*
			var msg gameddz.OutCard
			msg.Seat = rb.mySeat
			msg.Option = option
			log.Tracef(" ---------- robot outcard end : %v", cards)
			for _, card := range cards {
				msg.Outcards = append(msg.Outcards, &gameddz.Card{Number: int32(card>>4), Flower: gameddz.Card_Flower(card.Flower)})
			}
			for _, card := range rb.GameLogic.CurrentCards.Cards {
				msg.EarlyMaxCards = append(msg.EarlyMaxCards, &gameddz.Card{Number: int32(card>>4), Flower: gameddz.Card_Flower(card.Flower)})
			}
			event := common.MakeGameEvent(rb.robotID, int32(gameddz.Event_OUT_CARD_REQ), &msg)
			rb.mgr.PostEvent(*event, false)
		*/
	})
}

func (rb *robot) doPassCard() {
	ErrorCheck(false, 4, "PassCard is illegal !!!")
	rb.sendOutCardmsg(-1, nil)
}

func (rb *robot) doPassCardQuick(ms int) {
	ErrorCheck(false, 4, "PassCard quick is illegal !!!")
	rb.sendOutCardmsgEx(ms, -1, nil)
}

func (rb *robot) isAllSingleOut() bool {
	// 如果手上全是单张
	if IsGroupsAllSingle(rb.cardGroups) {
		length := len(rb.User.Cards)
		ErrorCheck(length > 0, 4, "hand cards len is 0, maybe error occur !!!")
		isOutOrder := (rb.exportData.IsBanker || rb.exportData.IsBankerPrev) && rb.isPlayerAlarm(false, 1)
		if isOutOrder {
			if length <= 2 {
				rb.sendOutCardmsg(1, []byte{rb.User.Cards[length-1]})
				return true
			}
			rb.sendOutCardmsg(1, []byte{rb.User.Cards[length-2]})
			return true
		}
		rb.sendOutCardmsg(1, []byte{rb.User.Cards[0]})
		return true
	}
	return false
}

func (rb *robot) doOutCard() {
	ErrorCheck(false, 4, "OutCard is illegal !!!")
	//是否全单张出牌
	if rb.isAllSingleOut() {
		return
	}
	//按价值排序
	rb.cardGroups = SortCardGroupsByValue(rb.cardGroups)
	//优先级牌组，排在前面的优先被打出
	priorityGroups := rb.getOutCardPriorityGroups()
	//统计大于Q的单张，大于J的对子
	var DaQ int
	var XiaoQ int
	var DaJ int
	var XiaoJ int
	for _, v := range priorityGroups {
		if v.Type == msg.CardsType_SingleCard {
			if v.Cards[0] >= 0xa1 {
				DaQ++
			} else {
				XiaoQ++
			}
		} else if v.Type == msg.CardsType_Pair {
			if v.Cards[0] >= 0x91 {
				DaJ++
			} else {
				XiaoJ++
			}
		}
	}

	//大于Q的单牌
	if DaQ != 0 && XiaoQ != 0 && DaQ >= XiaoQ {
		for _, v := range rb.cardGroups {
			if v.Type == msg.CardsType_SingleCard {
				rb.sendOutCardmsg(1, v.Cards)
				return
			}
		}
	}
	//大于J的对子
	if DaJ != 0 && XiaoJ != 0 && DaJ >= XiaoJ {
		for _, v := range rb.cardGroups {
			if v.Type == msg.CardsType_Pair {
				rb.sendOutCardmsg(1, v.Cards)
				return
			}
		}
	}

	//没找到上面的牌型，就找飞机。
	for _, v := range rb.cardGroups {
		if v.Type == msg.CardsType_SerialPair ||
			v.Type == msg.CardsType_SerialTripletWithOne ||
			v.Type == msg.CardsType_SerialTripletWithWing {
			rb.sendOutCardmsg(1, v.Cards)
			return
		}
	}

	//飞机也没得就找别人没有的牌型
	var cardGroups [2][]*CardGroup
	_, cardGroups[0] = GetMostValueGroup(rb.GameLogic.Chairs[(rb.User.ChairID+1)%3].Cards)
	_, cardGroups[1] = GetMostValueGroup(rb.GameLogic.Chairs[(rb.User.ChairID+2)%3].Cards)
	for _, v := range rb.cardGroups {
		if v.Type == msg.CardsType_Bomb || v.Type == msg.CardsType_Rocket {
			break
		}

		bFind := false
		for i := 0; i < 2; i++ {
			for _, tmp := range cardGroups[i] {
				if tmp.Type == v.Type {
					bFind = true
					break
				}
			}

			if bFind {
				break
			}
		}

		if !bFind {
			rb.sendOutCardmsg(1, v.Cards)
			return
		}
	}

	//别人都有牌的时候，找同种类最大的牌型
	for _, v := range rb.cardGroups {
		if v.Type == msg.CardsType_Bomb || v.Type == msg.CardsType_Rocket {
			break
		}
		temp1 := SearchFirstLargeGroup(cardGroups[0], v.Cards, false)
		temp2 := SearchFirstLargeGroup(cardGroups[1], v.Cards, false)
		AbsBig := false
		if temp1 != nil {
			if SearchFirstLargeGroup(rb.cardGroups, temp1.Cards, false) != nil {
				AbsBig = true
			} else {
				AbsBig = false
			}
		}

		if temp2 != nil && AbsBig == true {
			if SearchFirstLargeGroup(rb.cardGroups, temp2.Cards, false) != nil {
				AbsBig = true
			} else {
				AbsBig = false
			}
		}

		if AbsBig == true {
			rb.sendOutCardmsg(1, v.Cards)
			return
		}
		//rb.sendOutCardmsg(1, group.Cards)

	}

	for _, group := range priorityGroups {
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	ErrorCheck(false, 4, " why can't find groups cards, maybe error occur !!!")
}

func (rb *robot) getOutCardPriorityGroups() []*CardGroup {
	var priorityGroups []*CardGroup
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SerialTripletWithOne || group.Type == msg.CardsType_SerialTripletWithWing {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Sequence ||
			group.Type == msg.CardsType_SerialPair ||
			group.Type == msg.CardsType_SerialTriplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair) && group.Value < 12 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Pair ||
			group.Type == msg.CardsType_Triplet) && group.Value < 12 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_QuartetWithTwo || group.Type == msg.CardsType_QuartetWithTwoPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Bomb {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Rocket {
			priorityGroups = append(priorityGroups, group)
		}
	}
	return priorityGroups
}
func (rb *robot) getOutCardPriorityGroups1() []*CardGroup {
	var priorityGroups []*CardGroup
	rb.cardGroups = SortCardGroupsByValue(rb.cardGroups)
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SerialTripletWithOne || group.Type == msg.CardsType_SerialTripletWithWing {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Sequence || group.Type == msg.CardsType_SerialPair || group.Type == msg.CardsType_SerialTriplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair) && group.Value < 13 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet) && group.Value < 13 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_QuartetWithTwo || group.Type == msg.CardsType_QuartetWithTwoPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Bomb {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Rocket {
			priorityGroups = append(priorityGroups, group)
		}
	}
	return priorityGroups
}
func (rb *robot) doOutAllCard() {
	cardsType := GetCardsType(rb.User.Cards)
	ErrorCheck(cardsType > msg.CardsType_Normal, 4, "OutAllCard is illegal, handcard is not one group !!!")
	if nil != rb.GameLogic.CurrentCards.Cards {
		ErrorCheck(CompareCards(rb.GameLogic.CurrentCards.Cards, rb.User.Cards), 4, "OutAllCard is illegal, handcard can not more than !!!")
	}
	//有可能 4个k + 大小王，或者飞机+ 大小王，也可以判断为一手牌
	//判断手上是否有王炸，有的话先出王炸
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Rocket {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
	}
	//避免33334444这种被当成四带两对, 避免3334445555被当成飞机带两对, 333344445555这种被当成飞机带单张的情况
	bombCnt, exCnt := GetGroupsBombCnt(rb.cardGroups)
	//全炸弹
	if exCnt == 0 {
		for _, group := range rb.cardGroups {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
	} else if exCnt == 1 && bombCnt > 0 {
		for _, group := range rb.cardGroups {
			if group.Type < msg.CardsType_Bomb {
				rb.sendOutCardmsg(1, group.Cards)
				return
			}
		}
	}
	//是否满足4带2且外面没有能打过的炸弹，且玩家手牌不是一对或者一单， 或者单张，对子都是绝对大牌
	if cardsType == msg.CardsType_QuartetWithTwo || cardsType == msg.CardsType_QuartetWithTwoPair {
		var bombGroup *CardGroup
		for _, group := range rb.cardGroups {
			if group.Type == msg.CardsType_Bomb {
				bombGroup = group
				break
			}
		}
		ErrorCheck(nil != bombGroup, 4, " why has no bomb, maybe error occur !!!")
		var mincards []byte
		var maxcards []byte
		for _, group := range rb.cardGroups {
			if group.Type < msg.CardsType_Bomb {
				mincards = CloneCards(group.Cards)
				break
			}
		}
		for i := len(rb.cardGroups) - 1; i >= 0; i-- {
			if rb.cardGroups[i].Type < msg.CardsType_Bomb {
				maxcards = CloneCards(rb.cardGroups[i].Cards)
				break
			}
		}
		ErrorCheck(0 != len(mincards), 4, " mincards len is 0, maybe error occur !!!")
		ErrorCheck(0 != len(maxcards), 4, " maxcards len is 0, maybe error occur !!!")
		//外面没有能大过的牌
		if exist, _ := SearchLargerCardType(bombGroup.Cards, rb.deskLeftCards, true); !exist {
			if rb.isCanSafeOutCards(maxcards, false) {
				rb.sendOutCardmsg(1, maxcards)
				return
			}
		} else {
			if rb.isCanSafeOutCards(mincards, false) {
				rb.sendOutCardmsg(1, mincards)
				return
			}
		}
	}
	//出完所有手牌
	rb.sendOutCardmsg(1, rb.User.Cards)
}

func (rb *robot) doOutAbsBig() {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutAbsBig is illegal !!!")
	//按价值排序
	rb.cardGroups = SortCardGroupsByValue(rb.cardGroups)
	for _, v := range rb.cardGroups {
		log.Debugf("%v", v)
	}

	if len(rb.cardGroups) == 1 {
		//只有一组牌就不找最大的了。直接出
		log.Debugf("只有一组牌就不找最大的了。直接出")
		rb.sendOutCardmsg(1, rb.cardGroups[0].Cards)
		return
	}
	//只剩2组牌
	if len(rb.cardGroups) == 2 {
		//如果外面没任何牌能打过， 例如 4，大王 或者 55，22，或者 10 + 王炸，先出绝对大牌
		for _, group := range rb.cardGroups {
			if exist, _ := SearchLargerCardType(group.Cards, rb.deskLeftCards, true); !exist {
				rb.sendOutCardmsg(1, group.Cards)
				return
			}
		}
		//如果没有绝对大牌，只有相对最大的，并且大牌可以压住小牌，可以先出小再出大, 例如 55，22，或者 10 + 大王，先出小牌
		if CompareCards(rb.cardGroups[0].Cards, rb.cardGroups[1].Cards) && rb.isCanSafeOutCards(rb.cardGroups[0].Cards, false) {
			rb.sendOutCardmsg(1, rb.cardGroups[0].Cards)
			return
		}
	}
	//优先级牌组，排在前面的优先被打出
	var priorityGroups []*CardGroup
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SerialTripletWithOne || group.Type == msg.CardsType_SerialTripletWithWing {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Sequence || group.Type == msg.CardsType_SerialPair || group.Type == msg.CardsType_SerialTriplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet ||
			group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_QuartetWithTwo || group.Type == msg.CardsType_QuartetWithTwoPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Bomb || group.Type == msg.CardsType_Rocket {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range priorityGroups {
		if exist, _ := SearchLargerCardType(group.Cards, rb.deskLeftCards, false); !exist {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
	}
	ErrorCheck(false, 4, " can't find abs big cards, maybe error occur !!!")
	panic("1")
}

func (rb *robot) doGenCard(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "GenCard is illegal !!!")
	log.Debugf(" doGenCard --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//先在现有牌组里搜索
	var group *CardGroup
	if isFromBig {
		group = SearchLastLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
	} else {
		group = SearchFirstLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
	}
	if nil != group {
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	if isForce {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			var out []byte
			if isFromBig {
				out = result[len(result)-1]
			} else {
				out = result[0]
			}
			rb.sendOutCardmsg(1, out)
			return
		}
	}
	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

func (rb *robot) doOutSingle(min int, max int, isFromBig bool, isForce bool) {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutSingle is illegal !!!")
	log.Debugf(" doOutSingle --- ", min, max, isFromBig, isForce)
	//是否全单张出牌
	if rb.isAllSingleOut() {
		return
	}
	//先在拆分牌型里面找
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard && group.Value < 15 {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
	}
	_, counts := GetCardCounts(rb.User.Cards)
	//再找小于J的单张，从最小的找起
	for num, ct := range counts {
		if ct >= 4 {
			continue
		}
		if ct >= 1 && num < 15 {
			var outcards []byte
			for _, card := range rb.User.Cards {
				if card>>4 == byte(num) {
					outcards = append(outcards, card)
					break
				}
			}
			rb.sendOutCardmsg(1, outcards)
			return
		}
	}
	rb.sendOutCardmsg(1, []byte{rb.User.Cards[0]})
}

func (rb *robot) doOutNotSingle(min int, max int, isFromBig bool, isForce bool) {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutNotSingle is illegal !!!")
	log.Debugf(" doOutNotSingle --- ", min, max, isFromBig, isForce)
	//是否全单张出牌
	//if rb.isAllSingleOut() {
	//	return
	//}
	//按价值排序
	rb.cardGroups = SortCardGroupsByValue(rb.cardGroups)
	if isFromBig {
		f := func(s interface{}) {
			n := reflect.ValueOf(s).Len()
			swap := reflect.Swapper(s)
			for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
				swap(i, j)
			}
		}
		f(rb.cardGroups)
	}
	//优先级牌组，排在前面的优先被打出
	var priorityGroups []*CardGroup
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SerialTripletWithOne || group.Type == msg.CardsType_SerialTripletWithWing {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Sequence || group.Type == msg.CardsType_SerialPair || group.Type == msg.CardsType_SerialTriplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet ||
			group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_QuartetWithTwo || group.Type == msg.CardsType_QuartetWithTwoPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Bomb || group.Type == msg.CardsType_Rocket {
			priorityGroups = append(priorityGroups, group)
		}
	}
	//for _, group := range priorityGroups {
	//	rb.sendOutCardmsg(1, group.Cards)
	//	return
	//}
	rb.sendOutCardmsg(1, priorityGroups[0].Cards)
	//rb.doOutCardToStopBanker(priorityGroups)
	//ErrorCheck(false, 4, " why can't find groups cards, maybe error occur !!!")
}

func (rb *robot) doOutDouble(min int, max int, isFromBig bool, isForce bool) {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutDouble is illegal !!!")
	log.Debugf(" doOutDouble --- ", min, max, isFromBig, isForce)
	//先在拆分牌型里面找
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Pair && group.Value < 15 {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
	}
	//牌个数
	_, counts := GetCardCounts(rb.User.Cards)
	for num, ct := range counts {
		if ct >= 2 && num < 15 {
			var outcards []byte
			for _, card := range rb.User.Cards {
				if card>>4 == byte(num) {
					outcards = append(outcards, card)
					if len(outcards) >= 2 {
						break
					}
				}
			}
			rb.sendOutCardmsg(1, outcards)
			return
		}
	}
	//没有找到就默认出牌
	rb.doOutCard()
}

func (rb *robot) doOutNotDouble(min int, max int, isFromBig bool, isForce bool) {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutNotDouble is illegal !!!")
	log.Debugf(" doOutNotDouble --- ", min, max, isFromBig, isForce)
	// 如果手上全是对子
	if IsGroupsAllDouble(rb.cardGroups) {
		_, counts := GetCardCounts(rb.User.Cards)
		//再找小于J的单张，从最小的找起
		for num, ct := range counts {
			if ct >= 4 {
				continue
			}
			if ct >= 1 && num < 15 {
				var outcards []byte
				for _, card := range rb.User.Cards {
					if card>>4 == byte(num) {
						outcards = append(outcards, card)
						break
					}
				}
				rb.sendOutCardmsg(1, outcards)
				return
			}
		}
		//打出最小单张
		rb.sendOutCardmsg(1, []byte{rb.User.Cards[0]})
		return
	}
	//按价值排序
	rb.cardGroups = SortCardGroupsByValue(rb.cardGroups)
	//优先级牌组，排在前面的优先被打出
	var priorityGroups []*CardGroup
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SerialTripletWithOne || group.Type == msg.CardsType_SerialTripletWithWing {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Sequence || group.Type == msg.CardsType_SerialPair || group.Type == msg.CardsType_SerialTriplet {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if (group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Triplet) && group.Value < 15 {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Triplet ||
			group.Type == msg.CardsType_TripletWithSingle || group.Type == msg.CardsType_TripletWithPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_QuartetWithTwo || group.Type == msg.CardsType_QuartetWithTwoPair {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for i := len(rb.cardGroups) - 1; i >= 0; i-- {
		if rb.cardGroups[i].Type == msg.CardsType_Pair {
			priorityGroups = append(priorityGroups, rb.cardGroups[i])
		}
	}
	for _, group := range rb.cardGroups {
		if group.Type == msg.CardsType_Bomb || group.Type == msg.CardsType_Rocket {
			priorityGroups = append(priorityGroups, group)
		}
	}
	for _, group := range priorityGroups {
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	ErrorCheck(false, 4, " why can't find groups cards, maybe error occur !!!")
}

// 被动出牌：出炸弹
func (rb *robot) doOutBomb() {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "OutBomb is illegal !!!")
	if exist, result := SearchAllBombs(rb.User.Cards); exist {
		for _, cards := range result {
			log.Debugf("牌为：%v", cards)
			if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
				log.Debugf("牌为：%v", cards)
				rb.sendOutCardmsg(1, cards)
				return
			}
		}
	}
	if exist, result := SearchRocket(rb.User.Cards); exist {
		rb.sendOutCardmsg(1, result)
		return
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

func (rb *robot) doOutBombMax() {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 3, "OutBomb is illegal !!!")
	if exist, result := SearchAllBombs(rb.User.Cards); exist {
		//for _, cards := range result {
		//	if logic.CompareCards(rb.maxOutCards, cards) {
		//		rb.sendOutCardmsg(gameddz.OptionType_OP_OUT_CARD, cards)
		//		return
		//	}
		//}
		for idx := len(result) - 1; idx >= 0; idx-- {
			if CompareCards(rb.GameLogic.CurrentCards.Cards, result[idx]) {
				rb.sendOutCardmsg(1, result[idx])
				return
			}
		}
	}
	if exist, result := SearchRocket(rb.User.Cards); exist {
		rb.sendOutCardmsg(1, result)
		return
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

func (rb *robot) doOut4With2DoubleOr2Single() {
	c := rb.getCardsByGroup(msg.CardsType_QuartetWithTwoPair)
	if len(c) == 0 {
		c = rb.getCardsByGroup(msg.CardsType_QuartetWithTwo)
	}
	if len(c) > 0 {
		rb.sendOutCardmsg(1, c[0].Cards)
	} else {
		//最后没找到，则过牌
		rb.doPassCard()
	}
}
func (rb *robot) doOutCardToStopBanker(priorityGroups []*CardGroup) {
	var outCards []byte
	if priorityGroups[0].Type != msg.CardsType_SingleCard {
		outCards = priorityGroups[0].Cards
		rb.sendOutCardmsg(1, outCards)
		return
	}

	singles := make([]*CardGroup, 0)
	for _, g := range priorityGroups {
		if g.Type == msg.CardsType_SingleCard {
			singles = append(singles, g)
		}
	}
	for i := len(singles) - 1; i >= 0; i-- {
		g := singles[i]

		if (g.Cards[0]>>4 == byte(C2) || g.Cards[0]>>4 == byte(CBlackJoker) || g.Cards[0]>>4 == byte(CRedJoker)) && i > 0 {
			continue
		}
		if rb.standCard == nil {
			outCards = g.Cards
			break
		} else {
			if g.Cards[0]>>4 >= *rb.standCard>>4 {
				if i > 0 {
					continue
				} else {
					outCards = g.Cards
				}
			} else {
				outCards = g.Cards
				break
			}
		}
	}

	rb.sendOutCardmsg(1, outCards)
}

// 主动出牌：出最小单张
func (rb *robot) doOutMinSingle() {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutMinSingle is illegal !!!")
	_, counts := GetCardCounts(rb.User.Cards)
	//再找小于J的单张，从最小的找起
	for num, ct := range counts {
		if ct >= 4 {
			continue
		}
		if ct >= 1 && num < 15 {
			var outcards []byte
			for _, card := range rb.User.Cards {
				if card>>4 == byte(num) {
					outcards = append(outcards, card)
					break
				}
			}
			rb.sendOutCardmsg(1, outcards)
			return
		}
	}
	rb.sendOutCardmsg(1, []byte{rb.User.Cards[0]})
}

// 主动出牌：出最小对子
func (rb *robot) doOutMinDouble() {
	ErrorCheck(nil == rb.GameLogic.CurrentCards.Cards, 4, "OutMinDouble is illegal !!!")
	//牌个数
	//先找小于J的对子，从最小的找起
	_, counts := GetCardCounts(rb.User.Cards)
	for num, ct := range counts {
		if ct >= 4 {
			continue
		}
		if ct >= 2 && num < 11 {
			var outcards []byte
			for _, card := range rb.User.Cards {
				if card>>4 == byte(num) {
					outcards = append(outcards, card)
					if len(outcards) >= 2 {
						break
					}
				}
			}
			rb.sendOutCardmsg(1, outcards)
			return
		}
	}
	//再找小于J的单张，从最小的找起
	for num, ct := range counts {
		if ct >= 4 {
			continue
		}
		if ct >= 1 && num < 11 {
			var outcards []byte
			for _, card := range rb.User.Cards {
				if card>>4 == byte(num) {
					outcards = append(outcards, card)
					break
				}
			}
			rb.sendOutCardmsg(1, outcards)
			return
		}
	}
	//否则按照默认出牌
	rb.doOutCard()
}
func (rb *robot) followCardAction() []byte {
	var res []byte
	if len(rb.GameLogic.CurrentCards.Cards) == 1 && rb.GameLogic.CurrentCards.Cards[0]>>4 < byte(C2) {
		card2s := rb.getCard(byte(C2))
		if len(card2s) == 3 || len(card2s) == 2 {
			res = card2s[:1]
			return res
		}

		if rb.standCard != nil {
			outCard := [1]*byte{nil}
			for true {
				double := rb.getCardsByGroup(msg.CardsType_Pair)
				if len(double) > 0 {
					if double[0].Cards[0]>>4 >= *rb.standCard>>4 {
						outCard[0] = &double[0].Cards[0]
						break
					}
				}
				t3 := rb.getCardsByGroup(msg.CardsType_Triplet)
				if len(t3) > 0 {
					if t3[0].Cards[0]>>4 >= *rb.standCard>>4 {
						outCard[0] = &t3[0].Cards[0]
						break
					}
				}
				if len(double) > 0 {
					if double[0].Cards[0]>>4 == *rb.standCard>>4-1 {
						outCard[0] = &double[0].Cards[0]
						break
					}
				}
				if len(t3) > 0 {
					if t3[0].Cards[0]>>4 == *rb.standCard>>4-1 {
						outCard[0] = &t3[0].Cards[0]
						break
					}
				}
				break
			}
			if outCard[0] == nil {
				return nil
			} else {
				res = append(res, *outCard[0])
				return res
			}
		}
	}
	if len(rb.GameLogic.CurrentCards.Cards) == 2 && rb.GameLogic.CurrentCards.Cards[0]>>4 < byte(C2) {
		card2s := rb.getCard(byte(C2))
		if len(card2s) == 3 {
			res = card2s[:2]
			return res
		}

		if rb.standCard != nil {
			outCard := [2]*byte{nil}
			for true {
				t3 := rb.getCardsByGroup(msg.CardsType_Triplet)
				if len(t3) > 0 {
					if t3[0].Cards[0]>>4 >= *rb.standCard>>4 {
						outCard[0] = &t3[0].Cards[0]
						outCard[1] = &t3[0].Cards[1]
						break
					}
				}
				break
			}
			if outCard[0] == nil {
				return nil
			} else {
				res = []byte{*outCard[0], *outCard[1]}
			}
		}
	}
	if len(res) == 0 {
		return nil
	}
	return res
}
func (rb *robot) makeSureValueInRange(min int, max int) []*CardGroup {
	c := make([]*CardGroup, len(rb.cardGroups))
	copy(c, rb.cardGroups)
	i := 0
	for _, g := range c {
		if g.Value >= min && g.Value <= max {
			c[i] = g
			i++
		}
	}
	c = c[:i]
	return c
}

// 地主跟上家的出牌
func (rb *robot) doBankerGenPrevOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "BankerGenPrevOut is illegal !!!")
	log.Debugf(" doBankerGenPrevOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//是否稳赢
	if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, false); ok {
		rb.sendOutCardmsg(1, cards)
		return
	}

	cg := rb.makeSureValueInRange(min, max)
	var group *CardGroup

	if isForce {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			var out []byte
			if isFromBig {
				out = result[len(result)-1]
			} else {
				out = result[0]
			}
			rb.sendOutCardmsg(1, out)
			return
		}
	}

	//先在现有牌组里搜索
	if isFromBig {
		group = SearchLastLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	} else {
		group = SearchFirstLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	}
	if nil != group {
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	c := rb.followCardAction()
	if c != nil {
		rb.sendOutCardmsg(1, c)
		return
	}
	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

// 地主跟下家的出牌, 由于逆时针出牌顺序，可以判断地主上家是pass
func (rb *robot) doBankerGenNextOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "BankerGenNextOut is illegal !!!")
	log.Debugf(" doBankerGenNextOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//是否稳赢
	if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, false); ok {
		rb.sendOutCardmsg(1, cards)
		return
	}

	//先在现有牌组里搜索
	var group *CardGroup
	cg := rb.makeSureValueInRange(min, max)
	if isForce {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			var out []byte
			if isFromBig {
				out = result[len(result)-1]
			} else {
				out = result[0]
			}
			rb.sendOutCardmsg(1, out)
			return
		}
	}
	if isFromBig {
		group = SearchLastLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	} else {
		group = SearchFirstLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	}
	if nil != group {
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	c := rb.followCardAction()
	if c != nil {
		rb.sendOutCardmsg(1, c)
		return
	}

	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

// 地主上家跟地主的出牌, 由于逆时针出牌顺序，可以判断地主下家是pass
func (rb *robot) doPrevGenBankerOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "PrevGenBankerOut is illegal !!!")
	log.Debugf(" doPrevGenBankerOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//是否稳赢
	if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, false); ok {
		log.Debugf("doPrevGenBankerOut 稳赢 %v", cards)
		rb.sendOutCardmsg(1, cards)
		return
	}
	//出牌为单张，对子时，特殊处理
	cardstype := GetCardsType(rb.GameLogic.CurrentCards.Cards)
	if cardstype == msg.CardsType_SingleCard {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			//地主只剩一张牌
			if len(rb.GameLogic.Dizhu.Cards) == 1 {
				log.Debugf("1111")
				rb.sendOutCardmsg(1, result[len(result)-1])
				return
			}
			//先在现有牌组里搜索
			largeGroups := SearchAllLargeGroups(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
			//先找k到2，顺序
			for _, group := range largeGroups {
				if group.Value >= 12 && group.Value <= 15 {
					log.Debugf("2222")
					rb.sendOutCardmsg(1, group.Cards)
					return
				}
			}
			//再找2
			for i := 0; i < len(result); i++ {
				if result[i][0]>>4 == 15 {
					log.Debugf("3333")
					rb.sendOutCardmsg(1, result[i])
					return
				}
			}
			// //找大小王
			// for _, group := range largeGroups {
			// 	if group.Value > 15 {
			// 		rb.sendOutCardmsg(1, group.Cards)
			// 		return
			// 	}
			// }
			//找2以下的，逆序
			for i := len(result) - 1; i >= 0; i-- {
				if result[i][0]>>4 < 13 {
					log.Debugf("44444")
					rb.sendOutCardmsg(1, result[i])
					return
				}
			}
			//再找大王，小王
			for i := 0; i < len(result); i++ {
				if result[i][0]>>4 > 15 {
					log.Debugf("555555")
					rb.sendOutCardmsg(1, result[i])
					return
				}
			}
		}
	} else if cardstype == msg.CardsType_Pair {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			//地主只剩两张牌
			if len(rb.GameLogic.Dizhu.Cards) == 2 {
				log.Debugf("66666")
				rb.sendOutCardmsg(1, result[len(result)-1])
				return
			}
			//在现有牌组里搜索
			largeGroups := SearchAllLargeGroups(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
			if len(largeGroups) > 0 {
				log.Debugf("777777")
				rb.sendOutCardmsg(1, largeGroups[0].Cards)
				return
			}
			//出对子
			log.Debugf("88888")
			rb.sendOutCardmsg(1, result[0])
			return
		}
	} else {
		//先在现有牌组里搜索
		var group *CardGroup
		cg := rb.makeSureValueInRange(min, max)
		if isForce {
			//在所有手牌中搜索, 除炸弹之外的
			if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
				var out []byte
				if isFromBig {
					out = result[len(result)-1]
				} else {
					out = result[0]
				}
				log.Debugf("99999")
				rb.sendOutCardmsg(1, out)
				return
			}
		}
		if isFromBig {
			group = SearchLastLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
		} else {
			group = SearchFirstLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
		}
		if nil != group {
			log.Debugf("101010410")
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
		c := rb.followCardAction()
		if c != nil {
			log.Debugf("22221111")
			rb.sendOutCardmsg(1, c)
			return
		}

	}
	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					log.Debugf("213123123")
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			log.Debugf("344234234")
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

// 地主上家跟地主下家的出牌
func (rb *robot) doPrevGenNextOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "PrevGenNextOut is illegal !!!")
	log.Debugf(" doPrevGenNextOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//判断地主能否打过出牌, 地主打不过则过牌
	if rb.isPlayerNotOut(rb.GameLogic.Dizhu.ChairID, false) {
		rb.doPassCard()
		return
	}
	//是否稳赢
	//if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, true); ok {
	//	rb.sendOutCardmsg(1, cards)
	//	return
	//}
	//判断地主能否打过出牌, 地主打不过则过牌
	//if rb.User.CardsCount[(rb.GameLogic.Dizhu.ChairID+1)%3] <= 2 {
	//	rb.doPassCard()
	//	return
	//}
	//出牌为单张，对子时，特殊处理
	cardstype := GetCardsType(rb.GameLogic.CurrentCards.Cards)
	if cardstype == msg.CardsType_SingleCard {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			//地主只剩一张牌
			if len(rb.GameLogic.Dizhu.Cards) == 1 {
				rb.sendOutCardmsg(1, result[len(result)-1])
				return
			}
			//先在现有牌组里搜索
			largeGroups := SearchAllLargeGroups(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
			//先找k到2，顺序
			for _, group := range largeGroups {
				if group.Value >= 12 && group.Value <= 15 {
					rb.sendOutCardmsg(1, group.Cards)
					return
				}
			}
			// //再找2
			// for i := 0; i < len(result); i++ {
			// 	if result[i][0]>>4 == 15 {
			// 		rb.sendOutCardmsg(1, result[i])
			// 		return
			// 	}
			// }
			// //找大小王
			// for _, group := range largeGroups {
			// 	if group.Value > 15 {
			// 		rb.sendOutCardmsg(1, group.Cards)
			// 		return
			// 	}
			// }
			//找2以下的，逆序
			for i := len(result) - 1; i >= 0; i-- {
				if result[i][0]>>4 < 15 {
					rb.sendOutCardmsg(1, result[i])
					return
				}
			}
			// //再找大王，小王
			// for i := 0; i < len(result); i++ {
			// 	if result[i][0]>>4 > 15 {
			// 		rb.sendOutCardmsg(1, result[i])
			// 		return
			// 	}
			// }
		}
	} else if cardstype == msg.CardsType_Pair {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			//地主只剩两张牌
			if len(rb.GameLogic.Dizhu.Cards) == 2 {
				rb.sendOutCardmsg(1, result[len(result)-1])
				return
			}
			//在现有牌组里搜索
			largeGroups := SearchAllLargeGroups(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
			if len(largeGroups) > 0 && largeGroups[0].Value < 12 {
				rb.sendOutCardmsg(1, largeGroups[0].Cards)
				return
			}
			// //出对子
			// rb.sendOutCardmsg(1, result[0])
			// return
		}
	} else {
		//先在现有牌组里搜索
		var group *CardGroup
		if isForce {
			//在所有手牌中搜索, 除炸弹之外的
			if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
				var out []byte
				if isFromBig {
					out = result[len(result)-1]
				} else {
					out = result[0]
				}
				rb.sendOutCardmsg(1, out)
				return
			}
		}
		if isFromBig {
			group = SearchLastLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
		} else {
			group = SearchFirstLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, false)
		}
		if nil != group {
			rb.sendOutCardmsg(1, group.Cards)
			return
		}
		c := rb.followCardAction()
		if c != nil {
			rb.sendOutCardmsg(1, c)
			return
		}

	}
	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

// 地主下家跟地主的出牌
func (rb *robot) doNextGenBankerOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "NextGenBankerOut is illegal !!!")
	log.Debugf(" doNextGenBankerOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//是否稳赢
	if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, false); ok {
		log.Debugf("稳赢出牌 %v", cards)
		rb.sendOutCardmsg(1, cards)
		return
	}
	//isPrevNotOut := rb.isPlayerNotOut((rb.GameLogic.Dizhu.ChairID+2)%3, false)
	//if isPrevNotOut {
	//	isForce = true
	//	//队友是否只剩2张以下
	//	if rb.exportData.BankerPrevCardsCount <= 2 {
	//		group := SearchLastLargeGroup(rb.cardGroups, rb.GameLogic.CurrentCards.Cards, true)
	//		if nil != group {
	//			rb.sendOutCardmsg(1, group.Cards)
	//			return
	//		}
	//		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, true); exist {
	//			rb.sendOutCardmsg(1, result[len(result)-1])
	//			return
	//		}
	//	}
	//}
	//先在现有牌组里搜索
	var group *CardGroup
	cg := rb.makeSureValueInRange(min, max)
	if isForce {
		//在所有手牌中搜索, 除炸弹之外的
		if exist, result := SearchLargerCardType(rb.GameLogic.CurrentCards.Cards, rb.User.Cards, false); exist {
			var out []byte
			if isFromBig {
				out = result[len(result)-1]
			} else {
				out = result[0]
			}
			log.Debugf("%v-----%v----%v", rb.GameLogic.CurrentCards.Cards, rb.User.Cards, result)
			rb.sendOutCardmsg(1, out)
			return
		}
	}
	if isFromBig {
		group = SearchLastLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	} else {
		group = SearchFirstLargeGroup(cg, rb.GameLogic.CurrentCards.Cards, false)
	}
	if nil != group {
		log.Debugf("%v", group.Cards)
		rb.sendOutCardmsg(1, group.Cards)
		return
	}
	c := rb.followCardAction()
	if c != nil {
		log.Debugf("%v", c)
		rb.sendOutCardmsg(1, c)
		return
	}

	if isUseBomb {
		if exist, result := SearchAllBombs(rb.User.Cards); exist {
			for _, cards := range result {
				if CompareCards(rb.GameLogic.CurrentCards.Cards, cards) {
					rb.sendOutCardmsg(1, cards)
					return
				}
			}
		}
	}
	if isUseRocket {
		if exist, result := SearchRocket(rb.User.Cards); exist {
			rb.sendOutCardmsg(1, result)
			return
		}
	}
	//最后没找到，则过牌
	rb.doPassCard()
}

// 地主下家跟地主上家的出牌, 由于逆时针出牌顺序，可以判断地主是pass
func (rb *robot) doNextGenPrevOut(min int, max int, isFromBig bool, isForce bool, isUseBomb bool, isUseRocket bool) {
	ErrorCheck(nil != rb.GameLogic.CurrentCards.Cards, 4, "NextGenPrevOut is illegal !!!")
	log.Debugf(" doNextGenPrevOut --- ", min, max, isFromBig, isForce, isUseBomb, isUseRocket)
	//是否稳赢
	//if ok, cards := GenCardCanAbsWin(rb.cardGroups, rb.User.Cards, rb.GameLogic.CurrentCards.Cards, rb.deskLeftCards, true); ok {
	//	rb.sendOutCardmsg(1, cards)
	//	return
	//}
	//最后没找到，则过牌
	rb.doPassCard()
}
