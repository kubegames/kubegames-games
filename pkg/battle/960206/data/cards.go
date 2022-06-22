package data

import (
	"fmt"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

type Compare struct {
	headCards []byte
	midCards  []byte
	tailCards []byte
	headType  int
	midType   int
	tailType  int
	score     int
}

//根据最优配牌法将员工的牌进行分牌
func (user *User) SplitCards() {
	//log.Traceln("开始给用户分牌")
	cards, cardsArr := user.SetSpecialCardType()
	if user.SpecialCardType > SPECIAL_CARD_NO {
		//log.Traceln("玩家特殊牌就直接返回")
		user.addIntoSpareArrNew(true, &msg.S2CCardsAndCardType{
			HeadCards: user.HeadCards,
			MidCards:  user.MiddleCards,
			TailCards: user.TailCards,
			HeadType:  int32(user.HeadCardType),
			MidType:   int32(user.MidCardType),
			TailType:  int32(user.TailCardType),
		})
		//return
	}

	if hasFour, resCards, cardType := user.cardsHasOverFour(cardsArr); hasFour {
		//log.Traceln("走手牌中有四条同花顺逻辑线···,用户id：",user.User.GetID())
		user.splitFourOrThs(cards, resCards, cardType)
	} else {
		//log.Traceln("按照普通的牌型来处理")
		user.splitNormalCards(cards)
	}

	//log.Traceln("作弊前玩家的牌 ： ", user.User.GetID(), fmt.Sprintf(`%x,%x,%x`, user.HeadCards, user.MiddleCards, user.TailCards))
	//log.Traceln("玩家的牌型：", "  ", user.HeadCardType, user.MidCardType, user.TailCardType)

}

//有四条或同花顺并且已经确定尾墩的分法
func (user *User) splitFourOrThs(cardsAll, tailCards []byte, tailType int) {
	if tailType == poker.CardTypeFour {
		log.Traceln("走四条的逻辑")
		user.splitFour(cardsAll, tailCards, tailType)
	} else {
		user.splitNormalCards(cardsAll)
		return
	}
}

func (user *User) addIntoSpareArrNew(isSpecial bool, spare *msg.S2CCardsAndCardType) {
	if isSpecial {
		user.SpareArr = append(user.SpareArr, spare)
		return
	}
	headType, _ := poker.GetCardType13Water(spare.HeadCards)
	midType, _ := poker.GetCardType13Water(spare.MidCards)
	tailType, _ := poker.GetCardType13Water(spare.TailCards)
	spare.HeadType, spare.MidType, spare.TailType = int32(headType), int32(midType), int32(tailType)
	hasSame := false
	for _, v := range user.SpareArr {
		if spare.HeadType == v.HeadType && spare.MidType == v.MidType && spare.TailType == v.TailType {
			hasSame = true
			break
		}
	}
	if !hasSame {
		//if !user.User.IsRobot(){
		//	log.Traceln("非特殊牌型，单独处理 头墩中墩尾墩：",fmt.Sprintf(`%x %x %x`,spare.HeadCards,spare.MidCards,spare.TailCards))
		//}
		user.SpareArr = append(user.SpareArr, spare)
	}
}

//给用户添加备选方案
func (user *User) addIntoSpareArr(compareList []*Compare) {
	var headMax = compareList[0]
	var tailMax = compareList[0]
	for _, v := range compareList {
		v.midType, _ = poker.GetCardType13Water(v.midCards)
		isBeat := global.COMPARE_LOSE
		if isBeat, v.headType, headMax.headType = user.Compare3Cards(v.headCards, headMax.headCards); isBeat == global.COMPARE_WIN {
			v.tailType, _ = poker.GetCardType13Water(v.tailCards)
			headMax = v
		}
		tailArr := poker.Cards5SliceToArr(v.tailCards)
		maxTailArr := poker.Cards5SliceToArr(tailMax.tailCards)
		if isBeat, v.tailType, tailMax.tailType = user.Compare5Cards(tailArr, maxTailArr); isBeat == global.COMPARE_WIN {
			v.headType, _ = poker.GetCardType13Water(v.headCards)
			tailMax = v
		}
	}
	user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
		HeadCards: headMax.headCards,
		MidCards:  headMax.midCards,
		TailCards: headMax.tailCards,
		HeadType:  int32(headMax.headType),
		MidType:   int32(headMax.midType),
		TailType:  int32(headMax.tailType),
	})
	user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
		HeadCards: tailMax.headCards,
		MidCards:  tailMax.midCards,
		TailCards: tailMax.tailCards,
		HeadType:  int32(tailMax.headType),
		MidType:   int32(tailMax.midType),
		TailType:  int32(tailMax.tailType),
	})

}

//传过来的tailCards是最大的四张做为尾墩
func (user *User) splitFour(cardsAllVar, tailCardsVar []byte, tailType int) {
	cardsAll := make([]byte, 0)
	tailCards := make([]byte, 0)
	for _, v := range cardsAllVar {
		cardsAll = append(cardsAll, v)
	}
	for _, v := range tailCardsVar {
		tailCards = append(tailCards, v)
	}
	var leftCards = user.GetDifferentCards(cardsAll, tailCards)
	single := poker.GetDzSingleCards(tailCards)
	leftCards = append(leftCards, single...)
	//再把原来的尾墩中剔除掉落单的牌
	tailCards = poker.DelByteSlice(tailCards, single[0])
	//log.Traceln("剩余的牌有：",fmt.Sprintf(`%x , %x `,leftCards,tailCards))
	compareList := make([]*Compare, 0)
	leftCardsArr := poker.GetCombineCardsArr(leftCards, 5)
	maxCards2Type := user.GetMaxCardType(leftCardsArr)
	if maxCards2Type == poker.CardTypeSingle {
		log.Traceln("==============四条之后剩余的都是单张============")
		leftCards = poker.SortCards(leftCards) // 987654321
		tailCardsNew := append(tailCards, leftCards[8])
		midCards := make([]byte, 0)
		midCards = append(midCards, leftCards[0])
		midCards = append(midCards, leftCards[4:8]...)
		user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
			HeadCards: leftCards[1:4],
			MidCards:  midCards,
			TailCards: tailCardsNew,
			HeadType:  poker.CardTypeSingle,
			MidType:   poker.CardTypeSingle,
			TailType:  poker.CardTypeFour,
		})
		return
	}
	//log.Traceln("maxCards2Type : ",maxCards2Type)
	for _, midCards := range leftCardsArr {
		//log.Traceln("tail cards : ",poker.GetPrintCards(tailCards))
		//目前还剩9张牌在剩余的牌里面
		midType, _ := poker.GetCardType13Water(midCards)
		//剩4张牌
		//如果都是单张
		if maxCards2Type >= poker.CardTypeTW && midType < poker.CardTypeTW {
			continue
		}

		left4Cards := user.GetDifferentCards(leftCards, midCards)
		left4Cards = poker.SortCards(left4Cards)
		//fmt.
		if maxCards2Type == poker.CardTypeDz && midType == poker.CardTypeDz {
			//log.Traceln("四张里面 最大是对=============")
			tailCardsNew := make([]byte, 0)
			for _, v := range tailCards {
				tailCardsNew = append(tailCardsNew, v)
			}
			tailCardsNew = append(tailCardsNew, left4Cards[3])
			//tailCardsNew := append(tailCards, left4Cards[3])
			compareList = user.AppendCompareList(&Compare{
				headCards: left4Cards[:3], midCards: midCards, tailCards: tailCardsNew,
			}, compareList)
			continue
		}

		//如果是葫芦就直接返回
		//var headCards = left4Cards
		if midType >= poker.CardTypeSZA2345 || midType == poker.CardTypeDz {
			//log.Traceln("中墩是大于CardTypeSZA2345，直接就返回 此时的尾墩：",poker.GetPrintCards(tailCards))
			maxHead3 := user.GetMaxCards3(poker.GetCombineCardsArr(left4Cards, 3))
			//log.Traceln("maxHead3 之后的尾墩：",poker.GetPrintCards(tailCards))
			remainHead := user.GetDifferentCards(left4Cards, maxHead3) //剩余的从头墩中选出的一张
			//log.Traceln("GetDifferentCards 之后的尾墩：",poker.GetPrintCards(tailCards))
			//这里有个巨坑，如果直接写成 tailCardsNew := append(tailCards,remain)，之后会将tailCards的值改变，巨坑，要重新make
			tailCardsNew := make([]byte, 0)
			for _, v := range tailCards {
				tailCardsNew = append(tailCardsNew, v)
			}
			tailCardsNew = append(tailCardsNew, remainHead...)

			//log.Traceln("left4Cards: ",poker.GetPrintCards(left4Cards),"remainHead : ",poker.GetPrintCards(remainHead))
			//log.Traceln("头墩：",fmt.Sprintf(`%x`,maxHead3),"添加中墩：",fmt.Sprintf(`%x`,midCards)," 尾墩：",fmt.Sprintf(`%x`,taiCardsNew))
			compareList = user.AppendCompareList(&Compare{
				headCards: maxHead3, midCards: midCards, tailCards: tailCardsNew,
				midType: midType,
			}, compareList)
			//log.Traceln("AppendCompareList 之后的尾墩：",poker.GetPrintCards(tailCards))
			continue
		}
		//如果中墩是两对或者三条，就选出 落单的几张 再去落单的几张中组成最大的头墩
		if midType == poker.CardTypeTK || midType == poker.CardTypeTW {
			//log.Traceln("如果中墩是两对或者三条，就选出 落单的几张 再去-----",poker.GetPrintCards(midCards))
			singleCards := poker.GetDzSingleCards(midCards)
			midRemain := user.GetDifferentCards(midCards, singleCards) // 2233
			//log.Traceln("//取出中墩之后剩余的牌: ",fmt.Sprintf(`%x`,midRemain))
			leftAllCards := make([]byte, 0)
			for _, v := range left4Cards {
				leftAllCards = append(leftAllCards, v)
			}
			leftAllCards = append(leftAllCards, singleCards...) //原来的单张+中墩剔除出来的单张

			//取出最大的头墩
			maxHead3 := user.GetMaxCards3(poker.GetCombineCardsArr(leftAllCards, 3))
			//取出头墩之后剩余的牌
			headRemain := user.GetDifferentCards(leftAllCards, maxHead3) // 4561
			headRemain = poker.SortCards(headRemain)
			//log.Traceln("最后剩余的牌：",fmt.Sprintf(`%x`,leftAllCards),"中墩：",poker.GetPrintCards(midRemain),"尾墩：",poker.GetPrintCards(tailCards))
			//选出了头墩3张和最后一张放到尾墩，剩余的都是中墩的牌
			hasChoose := append(maxHead3, headRemain[len(headRemain)-1])
			remainForMid := user.GetDifferentCards(headRemain, hasChoose)

			//中墩
			midCardsNew := append(midRemain, remainForMid...)
			//尾墩
			tailCardsNew := make([]byte, 0)
			for _, v := range tailCards {
				tailCardsNew = append(tailCardsNew, v)
			}
			tailCardsNew = append(tailCardsNew, headRemain[len(headRemain)-1])
			compareList = user.AppendCompareList(&Compare{headCards: maxHead3, midCards: midCardsNew, tailCards: tailCardsNew}, compareList)
		} else {
			//log.Traceln("不是三条或对子，就选出最大的牌")
			maxHead3 := user.GetMaxCards3(poker.GetCombineCardsArr(left4Cards, 3))
			remainHead := user.GetDifferentCards(left4Cards, maxHead3) //剩余的从头墩中选出的一张
			tailCardsNew := append(tailCards, remainHead...)
			compareList = user.AppendCompareList(&Compare{headCards: maxHead3, midCards: midCards, tailCards: tailCardsNew}, compareList)
		}
	}
	finalMax := user.GetMaxCompareCards(compareList)
	if user.SpecialCardType == SPECIAL_CARD_NO {
		user.HeadCards = finalMax.headCards
		user.MiddleCards = finalMax.midCards
		user.TailCards = finalMax.tailCards
		user.HeadCardType = finalMax.headType
		user.MidCardType = finalMax.midType
		user.TailCardType = finalMax.tailType
	}
	user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
		HeadCards: finalMax.headCards,
		MidCards:  finalMax.midCards,
		TailCards: finalMax.tailCards,
		HeadType:  int32(finalMax.headType),
		MidType:   int32(finalMax.midType),
		TailType:  int32(finalMax.tailType),
	})
	//再添加备选的
	user.addIntoSpareArr(compareList)
}

//没有四条和同花顺的普通牌分法
func (user *User) splitNormalCards(cardsAll []byte) {
	maxCompareList := make([]*Compare, 0)
	resArr := poker.GetCombineCardsArr(cardsAll, 5)
	for _, tailCards := range resArr {
		tailType, tailCards := poker.GetCardType13Water(tailCards)
		if tailType == poker.CardTypeSingle {
			continue
		}
		compareList := user.GetMidAndHeadList(cardsAll, tailCards, tailType)
		maxCompare := user.GetMaxCompareCards(compareList)
		if maxCompare != nil {
			maxCompareList = user.AppendCompareList(maxCompare, maxCompareList)
		}
	}

	finalMax := user.GetMaxCompareCards(maxCompareList)
	if finalMax == nil {
		log.Warnf("finalMax 为空")
		return
	}
	if user.SpecialCardType == SPECIAL_CARD_NO {
		user.HeadCardType, user.HeadCards = poker.GetCardType13Water(finalMax.headCards)
		user.MidCardType, user.MiddleCards = poker.GetCardType13Water(finalMax.midCards)
		user.TailCardType, user.TailCards = poker.GetCardType13Water(finalMax.tailCards)
	}
	user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
		HeadCards: finalMax.headCards,
		MidCards:  finalMax.midCards,
		TailCards: finalMax.tailCards,
		HeadType:  int32(finalMax.headType),
		MidType:   int32(finalMax.midType),
		TailType:  int32(finalMax.tailType),
	})
	//user.PrintUserSpare()
	user.addIntoSpareArr(maxCompareList)
}

//最终得出的最大牌型
func (user *User) GetMaxCompareCards(compareList []*Compare) (maxCompare *Compare) {
	if len(compareList) == 0 {
		return nil
	}
	if len(compareList) == 1 {
		compareList[0].headType, _ = poker.GetCardType13Water(compareList[0].headCards)
		compareList[0].midType, _ = poker.GetCardType13Water(compareList[0].midCards)
		compareList[0].tailType, _ = poker.GetCardType13Water(compareList[0].tailCards)
		return compareList[0]
	}
	maxCompare = compareList[0]
	for i := 1; i < len(compareList); i++ {
		if user.BeatCompare(compareList[i], maxCompare) {
			maxCompare = compareList[i]
		}
	}
	maxCompare.headType, _ = poker.GetCardType13Water(maxCompare.headCards)
	maxCompare.midType, _ = poker.GetCardType13Water(maxCompare.midCards)
	maxCompare.tailType, _ = poker.GetCardType13Water(maxCompare.tailCards)
	//log.Traceln("maxCompare.tailType, ",maxCompare.tailType)
	return
}

//比较两个临时组成的牌，如果c1大于c2则返回true，同时将牌型写入
func (user *User) BeatCompare(c1, c2 *Compare) bool {

	c1.score, c2.score = 0, 0
	beatResHead, c1HeadType, c2HeadType := user.Compare3Cards(c1.headCards, c2.headCards)
	if c1HeadType != c2HeadType {
		if beatResHead == global.COMPARE_WIN {
			c1.score++
			c2.score--
			if c1HeadType == poker.Card3TypeBz {
				c1.score = global.HEAD_BZ_SCORE
				c2.score = -global.HEAD_BZ_SCORE
			}
		}
		if beatResHead == global.COMPARE_LOSE {
			c1.score--
			c2.score++
			if c2HeadType == poker.Card3TypeBz {
				c1.score = -global.HEAD_BZ_SCORE
				c2.score = global.HEAD_BZ_SCORE
			}
		}
	}

	c1MidArr := poker.Cards5SliceToArr(c1.midCards)
	c2MidArr := poker.Cards5SliceToArr(c2.midCards)
	beatResMid, c1MidType, c2MidType := user.Compare5Cards(c1MidArr, c2MidArr)
	if c1MidType != c2MidType {
		if beatResMid == global.COMPARE_WIN {
			c1.score++
			c2.score--
			if c1MidType >= poker.CardTypeHL {
				c1.score += SpecialMidMap[c1MidType]
				c2.score -= SpecialMidMap[c1MidType]
			}
		}
		if beatResMid == global.COMPARE_LOSE {
			c1.score--
			c2.score++
			if c2MidType >= poker.CardTypeHL {
				c1.score -= SpecialMidMap[c2MidType]
				c2.score += SpecialMidMap[c2MidType]
			}
		}
	}

	c1TailArr := poker.Cards5SliceToArr(c1.tailCards)
	c2TailArr := poker.Cards5SliceToArr(c2.tailCards)
	beatResTail, c1TailType, c2TailType := user.Compare5Cards(c1TailArr, c2TailArr)
	if c1TailType != c2TailType {
		if beatResTail == global.COMPARE_WIN {
			c1.score++
			c2.score--
			if c1TailType >= poker.CardTypeHL {
				c1.score += SpecialMidMap[c1TailType]
				c2.score -= SpecialMidMap[c1TailType]
			}
		}
		if beatResTail == global.COMPARE_LOSE {
			c1.score--
			c2.score++
			if c2TailType >= poker.CardTypeHL {
				c1.score -= SpecialMidMap[c2TailType]
				c2.score += SpecialMidMap[c2TailType]
			}
		}
	}

	//if isShow {
	//	log.Traceln("c1.score", c1.score, "c2.score", c2.score, "  ", beatResHead, beatResMid, beatResTail)
	//}

	//
	//如果三墩牌型都相同，则选择头墩最大的，头墩相同，就返回中墩最大的
	if c1TailType == c2TailType && c1HeadType == c2HeadType && c1MidType == c2MidType {
		if beatResHead == global.COMPARE_WIN {
			return true
		} else if beatResHead == global.COMPARE_LOSE {
			return false
		} else {
			//头墩相同，就返回尾墩最小
			return beatResTail == global.COMPARE_LOSE
		}
	}

	//尾墩牌型相同，并且头墩都是单张，则返回中墩牌型大的
	if c1TailType == c2TailType && (c1HeadType == poker.Card3TypeSingle && c2HeadType == poker.Card3TypeSingle) {
		//return c1MidType > c2MidType
	}

	if c1.score > c2.score {
		return true
	} else if c1.score < c2.score {
		return false
	} else {
		//两个积分相等，就选头墩最大
		//如果两个头墩牌型相同，就选小的，否则选大的
		if c1HeadType == c2HeadType {
			return beatResHead == global.COMPARE_LOSE
		}
		return beatResHead == global.COMPARE_WIN
	}

}

//牌中是否含有四条或者同花顺并返回最大的牌
//如果有则返回true 和 最大的牌
func (user *User) cardsHasOverFour(cardsArr [][]byte) (has bool, res []byte, cardTypeRes int) {
	maxSzEncode := 0
	for _, cards := range cardsArr {
		cardType, sortCards := poker.GetCardType13Water(cards)
		if cardType >= poker.CardTypeFour {
			//log.Traceln("cards : ",fmt.Sprintf(`%x`,cards))
			has = true
			encode := poker.GetEncodeCard(cardType, sortCards)
			if encode > maxSzEncode {
				res = sortCards
				cardTypeRes = cardType
				maxSzEncode = encode
			}
		}
	}
	return
}

//根据尾墩 获取所有头墩和中墩的组合，并将尾墩一起，返回完整的三墩牌 集合
func (user *User) GetMidAndHeadList(cardsAll, tailCards []byte, tailType int) (compareList []*Compare) {
	compareList = make([]*Compare, 0)
	left8Cards := user.GetDifferentCards(cardsAll, tailCards)
	left8CardsArr := poker.GetCombineCardsArr(left8Cards, 5)
	maxCards2Type := user.GetMaxCardType(left8CardsArr)

	if maxCards2Type == poker.CardTypeSingle { //98765432
		left8Cards = poker.SortCards(left8Cards)
		compare := &Compare{headCards: left8Cards[1:4], midCards: left8Cards[4:], tailCards: tailCards}
		compare.midCards = append(compare.midCards, left8Cards[0])
		if tailType == poker.CardTypeTHSA2345 || tailType == poker.CardTypeTHS {
			user.addIntoSpareArrNew(false, &msg.S2CCardsAndCardType{
				HeadCards: compare.headCards,
				MidCards:  compare.midCards,
				TailCards: compare.tailCards,
				HeadType:  poker.CardTypeSingle,
				MidType:   poker.CardTypeSingle,
				TailType:  int32(tailType),
			})
		} else {
			user.AppendCompareList(compare, compareList)
		}

		return
	}

	for _, midCards := range left8CardsArr {
		//log.Traceln("中墩：",poker.GetPrintCards(midCards),"尾墩：",poker.GetPrintCards(tailCards))
		tailCardsArr := poker.Cards5SliceToArr(tailCards)
		midCardsArr := poker.Cards5SliceToArr(midCards)
		isBeat, _, c2Type := user.Compare5Cards(tailCardsArr, midCardsArr)
		if isBeat == global.COMPARE_LOSE {
			//中墩比尾墩大，则舍弃
			continue
		}

		left3Cards := user.GetDifferentCards(left8Cards, midCards)

		if !user.Compare5And3Cards(midCards, left3Cards) {
			continue
		}
		//if maxCards2Type >= poker.CardTypeTW && c2Type < poker.CardTypeTW {
		//	continue
		//}

		if maxCards2Type == poker.CardTypeDz && c2Type == poker.CardTypeDz {
			compareList = user.AppendCompareList(&Compare{headCards: left3Cards, midCards: midCards, tailCards: tailCards}, compareList)
			continue
		}
		//如果是葫芦就直接返回
		if c2Type == poker.CardTypeHL {
			//log.Traceln("葫芦：中墩尾墩：",fmt.Sprintf(`%x %x`,midCards,tailCards))
			compareList = user.AppendCompareList(&Compare{headCards: left3Cards, midCards: midCards, tailCards: tailCards}, compareList)
			continue
		}

		//如果中墩是两对或者三条，就选出 落单的几张 再去落单的几张中组成最大的头墩
		if c2Type == poker.CardTypeTK || c2Type == poker.CardTypeTW {
			midCardsNew, left3CardsNew := user.GetSinkSingleMidHead(midCards, left3Cards)
			compareList = user.AppendCompareList(&Compare{headCards: left3CardsNew, midCards: midCardsNew, tailCards: tailCards}, compareList)
			continue
		}
		//满足 尾墩 > 中墩 > 头墩
		compare := &Compare{headCards: left3Cards, midCards: midCards, tailCards: tailCards}
		compareList = user.AppendCompareList(compare, compareList)
	}
	return
}

//获取所有可能牌组合中 最大的牌型
func (user *User) GetMaxCardType(cardsArr [][]byte) (maxCardType int) {
	for _, cards := range cardsArr {
		cardType, _ := poker.GetCardType13Water(cards)
		if cardType > maxCardType {
			maxCardType = cardType
		}
	}
	return
}

//将获取出的结果去重添加进 即将比牌的列表中
func (user *User) AppendCompareList(compare *Compare, compareList []*Compare) []*Compare {

	if !user.Compare5And3Cards(compare.midCards, compare.headCards) {
		return compareList
	}
	tailCardsArr := poker.Cards5SliceToArr(compare.tailCards)
	midCardsArr := poker.Cards5SliceToArr(compare.midCards)
	if isBeat, _, _ := user.Compare5Cards(tailCardsArr, midCardsArr); isBeat == global.COMPARE_LOSE {
		return compareList
	}
	//if !user.User.IsRobot(){
	//	log.Traceln("比较通过：头墩，中墩，尾墩：",fmt.Sprintf(`%x %x %x`,compare.headCards,compare.midCards,compare.tailCards))
	//}
	isExist := false
	for _, v := range compareList {
		if user.IsTwoHeadEq(compare.headCards, v.headCards) && user.IsTwo5CardsEq(compare.midCards, v.midCards) && user.IsTwo5CardsEq(compare.tailCards, v.tailCards) {
			isExist = true
			break
		}
	}
	if !isExist {
		compareList = append(compareList, compare)
		//log.Traceln("不存在就添加：",len(compareList))
	}
	return compareList
}

//判断两个头墩是否相等
func (user *User) IsTwoHeadEq(headCards1, headCards2 []byte) bool {
	if len(headCards1) != 3 || len(headCards2) != 3 {
		return false
	}
	return headCards1[0] == headCards2[0] && headCards1[1] == headCards2[1] && headCards1[2] == headCards2[2]
}

//判断两个5张牌是否相等
func (user *User) IsTwo5CardsEq(headCards1, headCards2 []byte) bool {
	if len(headCards1) != 5 || len(headCards2) != 5 {
		return false
	}
	return headCards1[0] == headCards2[0] && headCards1[1] == headCards2[1] && headCards1[2] == headCards2[2] && headCards1[3] == headCards2[3] && headCards1[4] == headCards2[4]
}

//牌中是否含有四条或者同花顺并返回最大的牌
//如果有则返回true 和 最大的牌
func (user *User) cardsHasCardType(theCardType int, cardsArr [][]byte) (has bool, res []byte) {
	maxSzEncode := 0
	for _, cards := range cardsArr {
		cardType, sortCards := poker.GetCardType13Water(cards)
		if cardType == theCardType {
			//log.Traceln("cards : ",fmt.Sprintf(`%x`,cards))
			has = true
			encode := poker.GetEncodeCard(cardType, sortCards)
			if encode > maxSzEncode {
				res = sortCards
				maxSzEncode = encode
			}
		}
	}
	return
}

//所有牌型组合中，是否为乌龙
func (user *User) cardsIsWuLong(cardsArr [][]byte) bool {
	for _, cards := range cardsArr {
		cardType, _ := poker.GetCardType13Water(cards)
		if cardType > poker.CardTypeSingle {
			//log.Traceln("cards : ",fmt.Sprintf(`%x`,cards))
			return false
		}
	}
	return true
}

//三张或两对落单的牌 重新组合成新的
func (user *User) GetSinkSingleMidHead(midOld, headOld []byte) (midNew, headNew []byte) {
	singleCards := poker.GetDzSingleCards(midOld)
	//log.Traceln("落单的牌: ",fmt.Sprintf(`%x`,singleCards))
	//中墩剩余的牌
	midRemain := user.GetDifferentCards(midOld, singleCards) // 2233
	//log.Traceln("//取出中墩之后剩余的牌: ",fmt.Sprintf(`%x`,midRemain))

	headOld = append(headOld, singleCards...)
	//log.Traceln("最后剩余的牌：",fmt.Sprintf(`%x`,headOld))
	//取出最大的头墩
	maxHead3 := user.GetMaxCards3(poker.GetCombineCardsArr(headOld, 3))
	//log.Traceln("//取出最大的头墩: ",fmt.Sprintf(`%x`,maxHead3))
	//取出头墩之后剩余的牌
	headRemain := user.GetDifferentCards(headOld, maxHead3) // 4561
	//log.Traceln("//中墩剩余的牌: ",fmt.Sprintf(`%x`,midRemain))
	midNew = append(midRemain, headRemain...)
	//log.Traceln("//新的中墩: ",fmt.Sprintf(`%x`,midNew))
	headNew = maxHead3
	return
}

//获取三张牌组合中最大的
func (user *User) GetMaxCards3(cardsArr [][]byte) (maxCards []byte) {
	maxCards = cardsArr[0]
	if len(cardsArr) <= 1 {
		return
	}
	for i := 1; i < len(cardsArr); i++ {
		if res, _, _ := user.Compare3Cards(cardsArr[i], maxCards); res == global.COMPARE_WIN {
			maxCards = cardsArr[i]
		}
	}
	return
}

//中墩5张 和 尾墩3张比较
func (user *User) Compare5And3Cards(cards1, cards2 []byte) bool {
	if len(cards1) != 5 || len(cards2) != 3 {
		log.Warnln("len(cards1) != 5 || len(cards2) != 3 ", fmt.Sprintf(`%x %x`, cards1, cards2))
		panic("Compare5And3Cards ")
		//return false
	}
	c1Type, _ := poker.GetCardType13Water(cards1)
	c2Type, _ := poker.GetCardType13Water(cards2)
	//log.Traceln("c1type : ",c1Type," c2Type : ",c2Type)

	if c1Type == poker.CardTypeTK && c2Type == poker.Card3TypeBz {
		//log.Traceln("都是三个：",)
		c1DzCard, _ := poker.GetDzCard(cards1)
		return c1DzCard > cards2[0]
	}

	if c2Type > c1Type {
		return false
	}
	if c1Type > c2Type {
		return true
	}
	if c1Type == poker.CardTypeTW && c2Type == poker.Card3TypeBz {
		return false
	}

	//找出两组牌中最大的那张牌
	c1Max, c2Max := user.GetMaxCard(cards1), user.GetMaxCard(cards2)
	if c1Type == poker.CardTypeDz && c2Type == poker.CardTypeDz {
		userCardDz, ok := poker.GetDzCard(cards1)
		if !ok {
			log.Traceln("玩家牌不是对子 : ", fmt.Sprintf(`%x`, cards1))
			panic("玩家牌不是对子")
		}
		anotherCardDz, ok := poker.GetDzCard(cards2)
		if !ok {
			log.Traceln("玩家牌不是对子 : ", fmt.Sprintf(`%x`, cards2))
			panic("玩家牌不是对子")
		}
		if userCardDz > anotherCardDz {
			return true
		} else if userCardDz == anotherCardDz {
			return c1Max > c2Max
		} else {
			return false
		}
	} else {
		return c1Max > c2Max
	}
}

func (user *User) GetMaxCard(cards []byte) (max byte) {
	for _, v := range cards {
		if v > max {
			max = v
		}
	}
	return
}

func (user *User) Compare3Cards(cards1, cards2 []byte) (isBeat, c1Type, c2Type int) {
	c1Type, _ = poker.GetCardType13Water(cards1)
	c2Type, _ = poker.GetCardType13Water(cards2)
	c1Encode, c2Encode := poker.GetEncodeCard(c1Type, cards1), poker.GetEncodeCard(c2Type, cards2)
	if c1Type == poker.Card3TypeDz && c2Type == poker.Card3TypeDz {
		userCardDz, ok := poker.GetDzCard(cards1)
		if !ok {
			log.Traceln("玩家牌不是对子 : ", fmt.Sprintf(`%x`, cards1))
			panic("玩家牌不是对子")
		}
		anotherCardDz, ok := poker.GetDzCard(cards2)
		if !ok {
			log.Traceln("玩家牌不是对子 : ", fmt.Sprintf(`%x`, cards2))
			panic("玩家牌不是对子")
		}
		if userCardDz > anotherCardDz {
			return global.COMPARE_WIN, c1Type, c2Type
		} else if userCardDz == anotherCardDz {
			if c1Encode > c2Encode {
				return global.COMPARE_WIN, c1Type, c2Type
			} else if c1Encode == c2Encode {
				return global.COMPARE_EQ, c1Type, c2Type
			} else {
				return global.COMPARE_LOSE, c1Type, c2Type
			}
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}
	} else {
		if c1Encode > c2Encode {
			return global.COMPARE_WIN, c1Type, c2Type
		} else if c1Encode == c2Encode {
			return global.COMPARE_EQ, c1Type, c2Type
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}
	}
}

func (user *User) Compare5Cards(cards1Arr, cards2Arr [5]byte) (isBeat, c1Type, c2Type int) {
	cards1 := poker.Cards5ArrToSlice(cards1Arr)
	cards2 := poker.Cards5ArrToSlice(cards2Arr)
	c1Type, cards1 = poker.GetCardType13Water(cards1)
	c2Type, cards2 = poker.GetCardType13Water(cards2)
	c1Encode, c2Encode := poker.GetEncodeCard(c1Type, cards1), poker.GetEncodeCard(c2Type, cards2)
	if c1Type == poker.CardTypeHL && c2Type == poker.CardTypeHL {
		//两个都是葫芦进行比较
		if cards1[2] > cards2[2] {
			return global.COMPARE_WIN, c1Type, c2Type
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}
	}
	if c1Type == poker.CardTypeDz && c2Type == poker.CardTypeDz || (c1Type == poker.CardTypeTK && c2Type == poker.CardTypeTK) {
		userCardDz, _ := poker.GetDzCard(cards1)
		anotherCardDz, _ := poker.GetDzCard(cards2)
		if userCardDz > anotherCardDz {
			return global.COMPARE_WIN, c1Type, c2Type
		} else if userCardDz == anotherCardDz {
			if c1Encode > c2Encode {
				return global.COMPARE_WIN, c1Type, c2Type
			} else if c1Encode == c2Encode {
				return global.COMPARE_EQ, c1Type, c2Type
			} else {
				return global.COMPARE_LOSE, c1Type, c2Type
			}
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}
	} else if c1Type == poker.CardTypeTW && c2Type == poker.CardTypeTW {
		// del by wd in 2020.3.10 for 双对可能出现合牌
		//log.Traceln("都是两对 ")
		//c1MaxDzCard := user.GetTwoDzMax(cards1)
		//c2MaxDzCard := user.GetTwoDzMax(cards2)
		////log.Traceln("max c1 c2 : ",fmt.Sprintf(`%x %x`,c1MaxDzCard,c2MaxDzCard))
		//if c1MaxDzCard > c2MaxDzCard {
		//	return global.COMPARE_WIN, c1Type, c2Type
		//} else {
		//	return global.COMPARE_LOSE, c1Type, c2Type
		//}

		// add by wd in 2020.3.10 for 双对可能出现合牌
		remainCards1 := user.RemoveSingleInTwoDz(cards1)
		remainCards2 := user.RemoveSingleInTwoDz(cards2)

		remainCode1, remainCode12 := poker.GetEncodeCard(c1Type, remainCards1), poker.GetEncodeCard(c2Type, remainCards2)
		if remainCode1 > remainCode12 {
			return global.COMPARE_WIN, c1Type, c2Type
		} else if remainCode1 == remainCode12 {
			return global.COMPARE_EQ, c1Type, c2Type
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}

	} else {
		if c1Encode > c2Encode {
			return global.COMPARE_WIN, c1Type, c2Type
		} else if c1Encode == c2Encode {
			//log.Traceln("compare 5 相等 ")
			return global.COMPARE_EQ, c1Type, c2Type
		} else {
			return global.COMPARE_LOSE, c1Type, c2Type
		}
	}
}

// add by wd in 2020.3.10
// RemoveSingleInTwoDz 移除双对牌型中的对子牌，返回倒叙排序之后的切片
// 这不对牌型做验证，所以必须传入为双对的牌
func (user *User) RemoveSingleInTwoDz(cards []byte) (remainCards []byte) {
	remainCards = cards
	single := poker.GetDzSingleCards(cards)[0]
	for k, card := range cards {
		if single == card {
			remainCards = append(remainCards[:k], remainCards[k+1:]...)
			break
		}
	}
	return
}

//获取两对中最大的那张对子
func (user *User) GetTwoDzMax(cards []byte) (maxCard byte) {
	single := poker.GetDzSingleCards(cards)[0]
	var dz1 byte
	var dz2 byte
	for _, card := range cards {
		if card != single {
			dz1 = card
			break
		}
	}
	for _, card := range cards {
		if card != single && card != dz1 {
			dz2 = card
			break
		}
	}
	if dz1 > dz2 {
		return dz1
	}
	return dz2
}

//自定义排序手牌数组
func (user *User) SortCardsSelf(cards []byte, cardType int) []byte {
	if cardType == 0 {
		cardType, _ = poker.GetCardType13Water(cards)
	}

	//先进行从小到大的排序
	for i := 0; i < len(cards)-1; i++ {
		for j := 0; j < (len(cards) - 1 - i); j++ {
			if (cards)[j] > (cards)[j+1] {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
	if len(cards) == 3 {
		return cards
	}
	cv0, _ := poker.GetCardValueAndColor(cards[0])
	cv1, _ := poker.GetCardValueAndColor(cards[1])
	cv2, _ := poker.GetCardValueAndColor(cards[2])
	cv3, _ := poker.GetCardValueAndColor(cards[3])
	cv4, _ := poker.GetCardValueAndColor(cards[4])
	if cardType == poker.CardTypeDz {
		//23455,23445,23345 => 22345
		if cv3 == cv4 {
			cards[3], cards[4], cards[0], cards[1] = cards[0], cards[1], cards[3], cards[4] //55423
			cards[2], cards[4] = cards[4], cards[2]                                         //55324
			cards[2], cards[3] = cards[3], cards[2]                                         // 55234
			return cards
		}
		if cv2 == cv3 {
			cards[2], cards[3], cards[0], cards[1] = cards[0], cards[1], cards[2], cards[3] //44235
			return cards
		}
		if cv1 == cv2 {
			cards[0], cards[2] = cards[2], cards[0]
			return cards
		}
		return cards
	}
	if cardType == poker.CardTypeTK {
		if cv0 == cv1 && cv0 == cv2 {
			return cards
		}
		//三条 35556 或 35666 ，需要 55536
		if cv1 == cv2 && cv1 == cv3 {
			cards[0], cards[3] = cards[3], cards[0]
			return cards
		}
		if cv2 == cv4 && cv2 == cv3 {
			cards[0], cards[3] = cards[3], cards[0]
			cards[1], cards[4] = cards[4], cards[1]
			return cards
		}
	}
	//四条
	if cardType == poker.CardTypeFour {
		if cv0 != cv1 {
			cards[0], cards[4] = cards[4], cards[0]
		}
		return cards
	}
	//两对 22334 	23344 22344
	if cardType == poker.CardTypeTW {
		if cv0 != cv1 {
			for i := 0; i < 4; i++ {
				cards[i], cards[i+1] = cards[i+1], cards[i]
			}
			return cards
		}
		if cv2 != cv3 {
			cards[2], cards[4] = cards[4], cards[2]
			return cards
		}
	}
	if cardType == poker.CardTypeTHSA2345 || cardType == poker.CardTypeSZA2345 {
		cards[0], cards[4] = cards[4], cards[0]
		cards[1], cards[4] = cards[4], cards[1]
		cards[2], cards[4] = cards[4], cards[2]
		cards[3], cards[4] = cards[4], cards[3]
		return cards
	}
	//22333 => 22233
	if cardType == poker.CardTypeHL {
		if cv0 != cv2 {
			cards[3], cards[4], cards[0], cards[1] = cards[0], cards[1], cards[3], cards[4]
			return cards
		}
	}

	return cards

}

func (user *User) PrintUserSpare() {
	for _, v := range user.SpareArr {
		log.Traceln("头墩：", v.HeadType, fmt.Sprintf(`%x`, v.HeadCards))
		log.Traceln("中墩：", v.MidType, fmt.Sprintf(`%x`, v.MidCards))
		log.Traceln("尾墩：", v.TailType, fmt.Sprintf(`%x`, v.TailCards))
	}
}
