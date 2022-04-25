package game

import (
	"common/log"
	"fmt"
	"game_poker/doudizhu/msg"
	"game_poker/doudizhu/poker"
	"sort"
)

type SplitMode int

const (
	Split_4321 SplitMode = iota
	Split_432S1
	Split_43S21
	Split_4S321
	Split_S4321
	Split_Max
)

const (
	// NewRoundValue 一个回合的价值等于大王的牌值, 理论+实际经验值
	NewRoundValue = 0xf
	// GroupsValueOffset 手上所有牌组的总价值偏移量, 理论+实际经验值
	GroupsValueOffset = 77
)

type CardGroup struct {
	Cards []byte
	Type  msg.CardsType
	Value int
}

var CardTypeScoreMap map[string]float64

func InitCardTypeScoreMap() {
	//cardTypeScoreRes, err := session.GetCardTypeScore(&roomsrv.GetCardTypeScoreReq{GameId: int32(global.GameType)})
	//if err != nil {
	//	log.Errorf("GetCardTypeScore  error!,%v ", err)
	//}
	//CardTypeScoreMap = cardTypeScoreRes.Ddz
}
func (pGroup *CardGroup) IsEuqal(group *CardGroup) bool {
	if pGroup.Type != group.Type {
		return false
	}
	if pGroup.Value != group.Value {
		return false
	}
	if len(pGroup.Cards) != len(group.Cards) {
		return false
	}
	for i := 0; i < len(pGroup.Cards); i++ {
		if pGroup.Cards[i]&0xf0 != group.Cards[i]&0xf0 || pGroup.Cards[i]&0xf != group.Cards[i]&0xf {
			return false
		}
	}
	return true
}

func (pGroup *CardGroup) IsValueEuqal(group *CardGroup) bool {
	if pGroup.Type != group.Type {
		return false
	}
	if pGroup.Value != group.Value {
		return false
	}
	if len(pGroup.Cards) != len(group.Cards) {
		return false
	}
	for i := 0; i < len(pGroup.Cards); i++ {
		if pGroup.Cards[i]>>4 != group.Cards[i]>>4 {
			return false
		}
	}
	return true
}

func GetCardsValue(cards []byte) int {
	poker.SortCards(cards)
	var cardstype = poker.GetCardsType(cards)
	var cardslen = len(cards)
	var value int

	//log.Debugf("-------------- %v", cardstype)
	switch cardstype {
	case msg.CardsType_SingleCard:
		fallthrough
	case msg.CardsType_Pair:
		fallthrough
	case msg.CardsType_Triplet:
		value = int(cards[0] >> 4)
	case msg.CardsType_Bomb:
		value = int(cards[0]>>4 + NewRoundValue)
	case msg.CardsType_TripletWithSingle:
		fallthrough
	case msg.CardsType_TripletWithPair:
		value = int(cards[2] >> 4)
	case msg.CardsType_Sequence:
		value = int(cards[cardslen-1]>>4 + 1)
	case msg.CardsType_SerialPair:
		value = int(cards[cardslen-1]>>4 + 2)
	case msg.CardsType_SerialTriplet:
		value = int(cards[cardslen-1]>>4 + 3)
	case msg.CardsType_SerialTripletWithOne:
		fallthrough
	case msg.CardsType_SerialTripletWithWing:
		_, number := GetAirPlaneStartEnd(cards)
		value = (number + 3)
	case msg.CardsType_QuartetWithTwo:
		fallthrough
	case msg.CardsType_QuartetWithTwoPair:
		_, number := IsHasBomb(cards)
		value = number + NewRoundValue - 1
	case msg.CardsType_Rocket:
		value = int(cards[0]>>4 + cards[1]>>4)
	}
	return value
}

//获取牌型分数
func GetCardsScore(cards []byte, t msg.CardsType) int {
	//sort.Sort(CardSlice(cards))
	c := Sort111(cards, t)
	//cardsType := GetCardsType(c)
	if c == nil {
		c = cards
	}
	score := CalculateScore(t, c...)
	return int(score)
}

////计算牌型分数
func CalculateScore(cardType msg.CardsType, cards ...byte) (score float64) {

	switch cardType {
	case msg.CardsType_TripletWithSingle:
		tripleScore := CalculateScore(msg.CardsType_Triplet, cards[:3]...)
		singleScore := CalculateScore(msg.CardsType_SingleCard, cards[3])
		score = tripleScore + singleScore
	case msg.CardsType_TripletWithPair:
		tripleScore := CalculateScore(msg.CardsType_Triplet, cards[:3]...)
		doubleScore := CalculateScore(msg.CardsType_Pair, cards[3:]...)
		score = tripleScore + doubleScore
	case msg.CardsType_SerialTripletWithOne:
		lenth := len(cards) / 4
		airplaneScore := CalculateScore(msg.CardsType_SerialTriplet, cards[:lenth*3]...)
		score += airplaneScore
		for i := 0; i < lenth; i++ {
			singleScore := CalculateScore(msg.CardsType_SingleCard, cards[3*lenth+i])
			score += singleScore
		}
	case msg.CardsType_SerialTripletWithWing:
		lenth := len(cards) / 5
		// TODO 算分数这个算三顺。
		airplaneScore := CalculateScore(msg.CardsType_Triplet, cards[:lenth*3]...)
		score += airplaneScore
		for i := 0; i < lenth; i++ {
			doubleScore := CalculateScore(msg.CardsType_Pair, cards[lenth*3+i*2:lenth*3+i*2+2]...)
			score += doubleScore
		}

	case msg.CardsType_QuartetWithTwo:
		quadraScore := CalculateScore(msg.CardsType_Bomb, cards[:4]...)
		doubleScore := CalculateScore(msg.CardsType_Pair, cards[4:]...)
		score = quadraScore + doubleScore
	case msg.CardsType_QuartetWithTwoPair:
		quadraScore := CalculateScore(msg.CardsType_Bomb, cards[:4]...)
		doubleScore1 := CalculateScore(msg.CardsType_Pair, cards[4:6]...)
		doubleScore2 := CalculateScore(msg.CardsType_Pair, cards[6:]...)
		score = quadraScore + doubleScore1 + doubleScore2
	default:
		card := CardSliceToString(cards)
		tempScore, ok := CardTypeScoreMap[card]
		if !ok {
			log.Errorf("找不到对应的牌: %s\n", card)
		}

		score += tempScore
	}

	//log.Printf("手牌:%+v\n总分: %f\n", *temp, scoreTemp)

	//case <-stopCh:
	//fmt.Printf("2手牌:%+v\n总分: %f\n", *result, score)
	return
}
func GetGroupsValue(groups []*CardGroup) int {
	var sum = 0
	for _, group := range groups {
		sum += group.Value
	}
	return sum - len(groups)*NewRoundValue + GroupsValueOffset
}
func SumCardsGroupValue(groups []*CardGroup) int {
	var sum = 0
	for _, group := range groups {
		sum += group.Value
	}
	return sum
}

func GetMostValueGroup(cards []byte) (int, []*CardGroup) {
	var maxGroups []*CardGroup
	var maxValue = -1000
	if !isHasSingleStraight(cards) {
		maxGroups = splitCards(Split_4321, cards)
		maxValue = GetGroupsValue(maxGroups)
		return maxValue, maxGroups
	}
	var vecGroups [][]*CardGroup
	for i := Split_4321; i < Split_Max; i++ {
		vecGroups = append(vecGroups, splitCards(i, cards))
	}
	ErrorCheck(len(vecGroups) > 0, 0, "vecGroups is 0 !!!")
	// //打印
	// for k, groups := range vecGroups {
	// 	groupValue := GetGroupsValue(groups)
	// 	fmt.Printf(" >>>>>>>>>>>>> mode : %v \n", k)
	// 	fmt.Println(" >>>>>>>>>>>>> groupValue : ", groupValue)
	// 	for index, group := range groups {
	// 		fmt.Println(" index : ", index, *group)
	// 	}
	// 	fmt.Println()
	// }
	vecGroups = sortGroupsVec(vecGroups, false)
	maxGroups = vecGroups[0]
	maxValue = GetGroupsValue(maxGroups)
	return maxValue, maxGroups
}

//手牌拆分
func splitCards(mode SplitMode, cards []byte) []*CardGroup {
	//重新创建切片，不改变原切片的数据
	var tempCards = CloneCards(cards)
	//排序，由小到大
	//sort.Sort(CardSlice(tempCards))
	poker.SortCards(tempCards)
	//牌个数
	_, counts := GetCardCounts(tempCards)
	//牌型组合
	var cardGroups []*CardGroup
	var pGroup *CardGroup
	var tempGroups []*CardGroup
	//拆王炸
	if tempCards, pGroup = splitRocket(tempCards, counts); pGroup != nil {
		cardGroups = append(cardGroups, pGroup)
	}

	//拆2，2不能组成顺子，单2，对2都是较大牌
	if tempCards, pGroup = splitPoker2(tempCards, counts); pGroup != nil {
		cardGroups = append(cardGroups, pGroup)
	}

	switch mode {
	case Split_4321:
		//拆炸弹
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 4); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆三张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 3); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆对子
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 2); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆单张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 1); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
	case Split_432S1:
		//拆炸弹
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 4); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆三张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 3); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
			for _, v := range tempGroups {
				log.Debugf("1 %v", *v)
			}
		}
		//拆对子
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 2); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆顺子
		if tempCards, tempGroups = splitStraights(tempCards, counts); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆单张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 1); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
	case Split_43S21:
		//拆炸弹
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 4); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆三张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 3); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆顺子
		if tempCards, tempGroups = splitStraights(tempCards, counts); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆对子
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 2); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆单张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 1); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
	case Split_4S321:
		//拆炸弹
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 4); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆顺子
		if tempCards, tempGroups = splitStraights(tempCards, counts); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆三张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 3); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆对子
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 2); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆单张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 1); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
			for _, v := range tempGroups {
				log.Debugf("1 %v", *v)
			}
		}
	case Split_S4321:
		//拆顺子
		if tempCards, tempGroups = splitStraights(tempCards, counts); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)

		}
		//拆炸弹
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 4); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆三张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 3); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆对子
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 2); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
		//拆单张
		if tempCards, tempGroups = splitSameCards(tempCards, counts, 1); tempGroups != nil {
			cardGroups = append(cardGroups, tempGroups...)
		}
	}
	ErrorCheck(len(tempCards) == 0, 1, "tempCards is remained !!!")
	//顺子组合
	cardGroups = combineStraightGroups(cardGroups)
	//飞机，三带一，三带二组合
	cardGroups = combineThreeGroups(cardGroups)

	//校验数据
	checkCards := CloneCards(cards)
	checkLen := len(checkCards)
	var count int
	var ok bool
	for _, group := range cardGroups {
		count += len(group.Cards)
		if GetCardsValue(group.Cards) != group.Value {
			log.Debugf("%v---%v", GetCardsValue(group.Cards), group.Value)
			log.Debugf("%v", group.Cards)
		}

		ErrorCheck(GetCardsValue(group.Cards) == group.Value, 1, " cards value must equal to group value, error ocuur !!!")
		ok, checkCards = RemoveCards(checkCards, group.Cards)
		ErrorCheck(ok, 1, " group cards must within checkCards, error ocuur !!!")
	}
	ErrorCheck(checkLen == count, 1, " checkCards len must equal to count, error ocuur !!!")
	return sortCardGroups(cardGroups)
}

func splitRocket(cards []byte, counts []int) ([]byte, *CardGroup) {
	//拆王炸
	var group CardGroup
	if counts[14] == 1 && counts[15] == 1 {
		group.Cards = append(group.Cards, 0xe1, 0xf1)
		group.Type = msg.CardsType_Rocket
		group.Value = 29
		_, cards = RemoveCards(cards, group.Cards)
		counts[14] = 0
		counts[15] = 0
	} else if counts[14] == 1 {
		group.Cards = append(group.Cards, 0xe1)
		group.Type = msg.CardsType_SingleCard
		group.Value = 14
		_, cards = RemoveCards(cards, group.Cards)
		counts[14] = 0
	} else if counts[15] == 1 {
		group.Cards = append(group.Cards, 0xf1)
		group.Type = msg.CardsType_SingleCard
		group.Value = 15
		_, cards = RemoveCards(cards, group.Cards)
		counts[15] = 0
	} else {
		return cards, nil
	}
	return cards, &group
}

func splitPoker2(cards []byte, counts []int) ([]byte, *CardGroup) {
	//拆2，2不能组成顺子，单2，对2都是较大牌
	if counts[13] > 0 {
		var group CardGroup
		for _, card := range cards {
			if card>>4 == 13 {
				group.Cards = append(group.Cards, card)
			}
		}
		switch counts[13] {
		case 1:
			group.Type = msg.CardsType_SingleCard
			group.Value = 13
		case 2:
			group.Type = msg.CardsType_Pair
			group.Value = 13
		case 3:
			group.Type = msg.CardsType_Triplet
			group.Value = 13
		case 4:
			group.Type = msg.CardsType_Bomb
			group.Value = 13 + NewRoundValue
		}
		_, cards = RemoveCards(cards, group.Cards)
		counts[13] = 0
		return cards, &group
	}
	return cards, nil
}

func splitSameCards(cards []byte, counts []int, samecount int) ([]byte, []*CardGroup) {
	//拆相同点数的牌
	var groups []*CardGroup
	for num, ct := range counts {
		if ct == samecount {
			var group CardGroup
			for _, card := range cards {
				if card>>4 == byte(num) {
					group.Cards = append(group.Cards, card)
				}
			}
			switch samecount {
			case 1:
				group.Type = msg.CardsType_SingleCard
				group.Value = num
			case 2:
				group.Type = msg.CardsType_Pair
				group.Value = num
			case 3:
				group.Type = msg.CardsType_Triplet
				group.Value = num
			case 4:
				group.Type = msg.CardsType_Bomb
				group.Value = num + NewRoundValue
			}
			groups = append(groups, &group)
			_, cards = RemoveCards(cards, group.Cards)
			counts[num] = 0
		}
	}
	return cards, groups
}

func splitStraights(cards []byte, counts []int) ([]byte, []*CardGroup) {
	//拆顺子，这里只拆出5张单顺子，后面会延长,合并单顺子，双顺子和三顺子可以由单顺子合成
	var groups []*CardGroup
	var start = 0
	var end = 0

	pickStraight := func(start int, end int) {
		var group CardGroup
		for i := start; i <= end; i++ {
			counts[i]--
			for _, card := range cards {
				if card>>4 == byte(i) {
					group.Cards = append(group.Cards, card)
					break
				}
			}
		}
		group.Type = msg.CardsType_Sequence
		group.Value = end + 1
		groups = append(groups, &group)
		_, cards = RemoveCards(cards, group.Cards)
	}

	for num := 1; num < 13; {
		if counts[num] > 0 {
			if start == 0 {
				start = num
			}
			end = num
			if end-start+1 >= 5 {
				pickStraight(start, end)
				num = start
				start = 0
				end = 0
				continue
			}
		} else {
			if end-start+1 >= 5 {
				pickStraight(start, end)
				num = start
				start = 0
				end = 0
				continue
			}
			start = 0
			end = 0
		}
		num++
	}

	//延长所有找到的单顺子
	if len(groups) > 0 {
		//排序cards
		for _, group := range groups {
			poker.SortCards(group.Cards)
		}

		extendStraight := func(group *CardGroup, start int, end int) {
			var tempcards []byte
			for i := start; i <= end; i++ {
				counts[i]--
				for _, card := range cards {
					if card>>4 == byte(i) {
						tempcards = append(tempcards, card)
						break
					}
				}
			}
			group.Cards = append(group.Cards, tempcards...)
			group.Type = msg.CardsType_Sequence
			group.Value = end + 1
			_, cards = RemoveCards(cards, tempcards)
		}

		start = 0
		end = 0
		for _, group := range groups {
			start = int(group.Cards[len(group.Cards)-1]>>4) + 1
			for num := start; num < 13; num++ {
				if counts[num] > 0 {
					end = num
				} else {
					break
				}
			}
			if end-start+1 > 0 {
				extendStraight(group, start, end)
				start = 0
				end = 0
			}
		}
	}

	//合并所有找到的单顺子, 最多只有2个单顺子能合并成一个
	if len(groups) > 1 {

		mergeStraight := func(groups []*CardGroup, first int, second int) []*CardGroup {
			var group CardGroup
			group.Cards = append(group.Cards, groups[first].Cards...)
			group.Cards = append(group.Cards, groups[second].Cards...)
			group.Type = msg.CardsType_Sequence
			group.Value = groups[second].Value
			var tempGroups []*CardGroup
			for i := 0; i < len(groups); i++ {
				if i != first && i != second {
					tempGroups = append(tempGroups, groups[i])
				}
			}
			tempGroups = append(tempGroups, &group)
			return tempGroups
		}

		//循环合并
		for {
			var change = false
			//排序groups
			groups = sortCardGroups(groups)
		ins_loop:
			for i := 0; i < len(groups)-1; i++ {
				for j := i + 1; j < len(groups); j++ {
					if groups[i].Cards[len(groups[i].Cards)-1]>>4+1 == groups[j].Cards[0]>>4 {
						change = true
						groups = mergeStraight(groups, i, j)
						break ins_loop
					}
				}
			}
			if !change {
				break
			}
		}
	}
	return cards, groups
}

func isHasSingleStraight(cards []byte) bool {
	//牌个数
	_, counts := GetCardCounts(cards)
	var startnum = 0
	var endnum = 0
	for num := 1; num < 13; num++ {
		if counts[num] > 0 {
			if startnum == 0 {
				startnum = num
			}
			endnum = num
		} else {
			if endnum-startnum+1 >= 5 {
				break
			}
			startnum = 0
			endnum = 0
		}
	}
	if endnum-startnum+1 >= 5 {
		return true
	}
	return false
}

//排序元素为groups的容器
func sortGroupsVec(vec [][]*CardGroup, isRise bool) [][]*CardGroup {
	//长度小于2，直接返回
	if len(vec) < 2 {
		return vec
	}
	//对玩家进行排序
	sort.Slice(vec, func(i int, j int) bool {
		vi := GetGroupsValue(vec[i])
		vj := GetGroupsValue(vec[j])
		if isRise {
			if vi != vj {
				return vi < vj
			}
			return len(vec[i]) > len(vec[j])
		}
		if vi != vj {
			return vi > vj
		}
		return len(vec[i]) < len(vec[j])
	})
	return vec
}

//排序groups
func sortCardGroups(groups []*CardGroup) []*CardGroup {
	//按类型, 牌值, 长度排序
	sort.Slice(groups, func(i int, j int) bool {
		if groups[i].Type != groups[j].Type {
			return groups[i].Type < groups[j].Type
		}
		if groups[i].Value != groups[j].Value {
			return groups[i].Value < groups[j].Value
		}
		return len(groups[i].Cards) < len(groups[j].Cards)
	})
	//cards排序
	for _, group := range groups {
		poker.SortCards(group.Cards)
		//sort.Sort(CardSlice(group.Cards))
	}
	return groups
}

//排序groups
func SortCardGroupsByType(groups []*CardGroup) []*CardGroup {
	return sortCardGroups(groups)
}

//排序groups
func SortCardGroupsByValue(groups []*CardGroup) []*CardGroup {
	//按类型, 牌值, 长度排序
	sort.Slice(groups, func(i int, j int) bool {
		if groups[i].Value != groups[j].Value {
			return groups[i].Value < groups[j].Value
		}
		return len(groups[i].Cards) < len(groups[j].Cards)
	})
	//cards排序
	for _, group := range groups {
		poker.SortCards(group.Cards)
		//sort.Sort(CardSlice(group.Cards))
	}
	return groups
}

//cloneGroups, 浅拷贝
func CloneCardGroups(groups []*CardGroup) []*CardGroup {
	retGroups := make([]*CardGroup, len(groups))
	for k, v := range groups {
		retGroups[k] = v
	}
	return retGroups
}

func deleteCardsGroup(groups []*CardGroup, cp *CardGroup) []*CardGroup {
	for k, group := range groups {
		if group.IsEuqal(cp) {
			groups = append(groups[:k], groups[k+1:]...)
			break
		}
	}
	return groups
}

func combineStraightGroups(groups []*CardGroup) []*CardGroup {
	//先排序
	groups = sortCardGroups(groups)
	//组合顺子成双顺子，三顺子
	groups = combineSingleStraightToMore(groups)
	//分解顺子, 顺子的头尾如果能拿出来和其他牌组成炸弹，三张，对子
	groups = analysisStraight(groups)
	//组合三张成三顺子, 优先组飞机牌型
	groups = combineThreeToStraight(groups)
	//分解顺子, 顺子的头尾如果能拿出来和其他牌组成炸弹，三张，对子
	groups = analysisStraight(groups)
	//组合对子成双顺子
	groups = combineDoubleToStraight(groups)
	//分解顺子, 顺子的头尾如果能拿出来和其他牌组成炸弹，三张，对子
	groups = analysisStraight(groups)
	return sortCardGroups(groups)
}

//组合顺子成双顺子，三顺子
func combineSingleStraightToMore(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_Sequence]) >= 2 {
		//组合顺子成双顺子，三顺子
		var length = len(groupsMap[msg.CardsType_Sequence])
		for i := 0; i < length-1; i++ {
			for j := i + 1; j < length; j++ {
				gi := groupsMap[msg.CardsType_Sequence][i]
				gj := groupsMap[msg.CardsType_Sequence][j]
				if gi.IsValueEuqal(gj) {
					var isthree = false
					var gk *CardGroup
					k := j + 1
					for ; k < length; k++ {
						gk = groupsMap[msg.CardsType_Sequence][k]
						if gj.IsValueEuqal(gk) {
							isthree = true
							break
						}
					}
					if isthree {
						var group CardGroup
						group.Cards = append(group.Cards, gi.Cards...)
						group.Cards = append(group.Cards, gj.Cards...)
						group.Cards = append(group.Cards, gk.Cards...)
						poker.SortCards(group.Cards)
						//sort.Sort(CardSlice(group.Cards))
						group.Type = msg.CardsType_SerialTriplet
						group.Value = gi.Value + 2
						groups = deleteCardsGroup(groups, gi)
						groups = deleteCardsGroup(groups, gj)
						groups = deleteCardsGroup(groups, gk)
						groups = append(groups, &group)
						i = k
					} else {
						var group CardGroup
						group.Cards = append(group.Cards, gi.Cards...)
						group.Cards = append(group.Cards, gj.Cards...)
						poker.SortCards(group.Cards)
						//sort.Sort(CardSlice(group.Cards))
						group.Type = msg.CardsType_SerialPair
						group.Value = gi.Value + 1
						groups = deleteCardsGroup(groups, gi)
						groups = deleteCardsGroup(groups, gj)
						groups = append(groups, &group)
						i = j
					}
					break
				}
			}
		}
	}
	return sortCardGroups(groups)
}

//组合对子成双顺子
func combineDoubleToStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		if (group.Type != msg.CardsType_Pair && group.Type != msg.CardsType_Triplet) || group.Value >= 12 {
			continue
		}
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	var cards []byte
	for _, group := range groupsMap[msg.CardsType_Pair] {
		cards = append(cards, group.Cards...)
	}
	for _, group := range groupsMap[msg.CardsType_Triplet] {
		cards = append(cards, group.Cards...)
	}
	cards = SortCards(cards)

	if len(cards) >= 3*2 {
		_, counts := GetCardCounts(cards)
		var bflag bool
		pickStraight := func(start int, end int) {
			bflag = true
			var group CardGroup
			for i := start; i <= end; i++ {
				counts[i] -= 2
				findCnt := 0
				for _, card := range cards {
					if card>>4 == byte(i) {
						group.Cards = append(group.Cards, card)
						findCnt++
						if findCnt >= 2 {
							break
						}
					}
				}
			}
			group.Type = msg.CardsType_SerialPair
			group.Value = end + 2
			groups = append(groups, &group)
			_, cards = RemoveCards(cards, group.Cards)
		}
		var start = 0
		var end = 0
		for num := 1; num < 13; {
			if counts[num] >= 2 {
				if start == 0 {
					start = num
				}
				end = num
			} else {
				if end-start+1 >= 3 {
					pickStraight(start, end)
					num = start
					start = 0
					end = 0
					continue
				}
				start = 0
				end = 0
			}
			num++
		}
		if end-start+1 >= 3 {
			pickStraight(start, end)
		}
		if bflag {
			//删除所有对子，三张
			for _, group := range groupsMap[msg.CardsType_Pair] {
				groups = deleteCardsGroup(groups, group)
			}
			for _, group := range groupsMap[msg.CardsType_Triplet] {
				groups = deleteCardsGroup(groups, group)
			}
			//重新分解三张，对子，单张
			var tempGroups []*CardGroup
			//拆三张
			if cards, tempGroups = splitSameCards(cards, counts, 3); tempGroups != nil {
				groups = append(groups, tempGroups...)
			}
			//拆对子
			if cards, tempGroups = splitSameCards(cards, counts, 2); tempGroups != nil {
				groups = append(groups, tempGroups...)
			}
			//拆单张
			if cards, tempGroups = splitSameCards(cards, counts, 1); tempGroups != nil {
				groups = append(groups, tempGroups...)
			}
			ErrorCheck(len(cards) == 0, 1, "cards is remained, maybe error occur !!!")
		}
	}
	return sortCardGroups(groups)
}

//组合三张成三顺子
func combineThreeToStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		if group.Type != msg.CardsType_Triplet || group.Value >= 12 {
			continue
		}
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_Triplet]) >= 2 {
		//组合三张成三顺子
		var length = len(groupsMap[msg.CardsType_Triplet])
		for i := 0; i < length-1; i++ {
			gi1 := groupsMap[msg.CardsType_Triplet][i]
			gi2 := groupsMap[msg.CardsType_Triplet][i+1]
			if gi1.Value+1 == gi2.Value {
				j := i + 1
				for j < length-1 {
					gj1 := groupsMap[msg.CardsType_Triplet][j]
					gj2 := groupsMap[msg.CardsType_Triplet][j+1]
					if gj1.Value+1 == gj2.Value {
						j++
					} else {
						break
					}
				}
				var group CardGroup
				for k := i; k <= j; k++ {
					group.Cards = append(group.Cards, groupsMap[msg.CardsType_Triplet][k].Cards...)
					groups = deleteCardsGroup(groups, groupsMap[msg.CardsType_Triplet][k])
				}
				poker.SortCards(group.Cards)
				//sort.Sort(CardSlice(group.Cards))
				group.Type = msg.CardsType_SerialTriplet
				group.Value = groupsMap[msg.CardsType_Triplet][j].Value + 3
				groups = append(groups, &group)
				i = j
			}
		}
	}
	return sortCardGroups(groups)
}

//判断某张相同点数的牌是否在其他牌组中存在
func isSameNumExsitOthers(groups []*CardGroup, num byte) (bool, *CardGroup) {
	for _, group := range groups {
		if group.Type == msg.CardsType_SingleCard || group.Type == msg.CardsType_Pair || group.Type == msg.CardsType_Triplet {
			if group.Cards[0]>>4 == byte(num) {
				return true, group
			}
		}
	}
	return false, nil
}

//分解顺子
func analysisStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_Sequence]) > 0 {
		//单顺子的头尾如果能够拿出来和剩余的牌组成 炸弹，三张，对子
		groups = analysisSingleStraight(groups)
	}

	if len(groupsMap[msg.CardsType_SerialPair]) > 0 {
		//双顺子的头尾如果能够拿出来和剩余的牌组成 炸弹，三张
		groups = analysisDoubleStraight(groups)
	}

	if len(groupsMap[msg.CardsType_SerialTriplet]) > 0 {
		//三顺子的头尾如果能够拿出来和剩余的牌组成炸弹
		groups = analysisThreeStraight(groups)
	}
	return sortCardGroups(groups)
}

//分解单顺子，//单顺子的头尾如果能够拿出来和剩余的牌组成 炸弹，三张，对子
func analysisSingleStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_Sequence]) <= 0 {
		return groups
	}
	//递归分解
	var divideStraight func(group *CardGroup)
	divideStraight = func(group *CardGroup) {
		poker.SortCards(group.Cards)
		//sort.Sort(CardSlice(group.Cards))
		length := len(group.Cards)
		if length < 6 {
			return
		}
		//优化从尾部取，尾部的数较大，取出来价值更大
		//从尾取
		if exsit, gp := isSameNumExsitOthers(groups, (group.Cards[length-1] >> 2)); exsit {
			gp.Cards = append(gp.Cards, group.Cards[length-1])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Pair
			} else if gp.Type == msg.CardsType_Pair {
				gp.Type = msg.CardsType_Triplet
			} else if gp.Type == msg.CardsType_Triplet {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[length-1])
			ErrorCheck(ok, 1, " remove not exsit element, maybe serious error occur !!!")
			group.Value--
			divideStraight(group)
		}
		if len(group.Cards) < 6 {
			return
		}
		//从头取
		if exsit, gp := isSameNumExsitOthers(groups, group.Cards[0]>>4); exsit {
			gp.Cards = append(gp.Cards, group.Cards[0])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Pair
			} else if gp.Type == msg.CardsType_Pair {
				gp.Type = msg.CardsType_Triplet
			} else if gp.Type == msg.CardsType_Triplet {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[0])
			ErrorCheck(ok, 1, " remove not exsit element, maybe serious error occur !!!")
			divideStraight(group)
		}
	}

	for _, group := range groupsMap[msg.CardsType_Sequence] {
		length := len(group.Cards)
		if length < 6 {
			continue
		}
		divideStraight(group)
	}
	return sortCardGroups(groups)
}

//分解双顺子，//双顺子的头尾如果能够拿出来和剩余的牌组成 炸弹，三张
func analysisDoubleStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_SerialPair]) <= 0 {
		return groups
	}
	//递归分解
	var divideStraight func(group *CardGroup)
	divideStraight = func(group *CardGroup) {
		poker.SortCards(group.Cards)
		//sort.Sort(CardSlice(group.Cards))
		length := len(group.Cards)
		if length < 4*2 {
			return
		}
		//优化从尾部取，尾部的数较大，取出来价值更大
		//从尾取
		if exsit, gp := isSameNumExsitOthers(groups, group.Cards[length-1]>>4); exsit {
			gp.Cards = append(gp.Cards, group.Cards[length-1], group.Cards[length-2])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Triplet
			} else if gp.Type == msg.CardsType_Pair {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			for i := 0; i < 2; i++ {
				ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[len(group.Cards)-1])
				ErrorCheck(ok, 1, " remove not exsit element [len-1] , maybe serious error occur !!!")
			}
			group.Value--
			divideStraight(group)
		}
		if len(group.Cards) < 4*2 {
			return
		}
		//从头取
		if exsit, gp := isSameNumExsitOthers(groups, group.Cards[0]>>4); exsit {
			gp.Cards = append(gp.Cards, group.Cards[0], group.Cards[1])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Triplet
			} else if gp.Type == msg.CardsType_Pair {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			for i := 0; i < 2; i++ {
				ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[0])
				ErrorCheck(ok, 1, " remove not exsit element [0], maybe serious error occur !!!")
			}
			divideStraight(group)
		}
	}
	for _, group := range groupsMap[msg.CardsType_SerialPair] {
		length := len(group.Cards)
		if length < 4*2 {
			continue
		}
		divideStraight(group)
	}
	return sortCardGroups(groups)
}

//分解三顺子，//三顺子的头尾如果能够拿出来和剩余的牌组成 炸弹
func analysisThreeStraight(groups []*CardGroup) []*CardGroup {
	groupsMap := make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	if len(groupsMap[msg.CardsType_SerialTriplet]) <= 0 {
		return groups
	}
	//递归分解
	var divideStraight func(group *CardGroup)
	divideStraight = func(group *CardGroup) {
		poker.SortCards(group.Cards)
		length := len(group.Cards)
		if length < 3*3 {
			return
		}
		//优化从尾部取，尾部的数较大，取出来价值更大
		//从尾取
		if exsit, gp := isSameNumExsitOthers(groups, group.Cards[length-1]>>4); exsit {
			gp.Cards = append(gp.Cards, group.Cards[length-1], group.Cards[length-2], group.Cards[length-3])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			for i := 0; i < 3; i++ {
				ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[len(group.Cards)-1])
				ErrorCheck(ok, 1, " remove not exsit element [len-1], maybe serious error occur !!!")
			}
			group.Value--
			divideStraight(group)
		}
		if len(group.Cards) < 3*3 {
			return
		}
		//从头取
		if exsit, gp := isSameNumExsitOthers(groups, group.Cards[0]>>4); exsit {
			gp.Cards = append(gp.Cards, group.Cards[0], group.Cards[1], group.Cards[2])
			if gp.Type == msg.CardsType_SingleCard {
				gp.Type = msg.CardsType_Bomb
				gp.Value += NewRoundValue
			} else {
				ErrorCheck(false, 1, " card count more than 4, maybe serious error occur !!!")
			}
			var ok bool
			for i := 0; i < 3; i++ {
				ok, group.Cards = RemoveOneCard(group.Cards, group.Cards[0])
				ErrorCheck(ok, 1, " remove not exsit element [0], maybe serious error occur !!!")
			}
			divideStraight(group)
		}
	}
	for _, group := range groupsMap[msg.CardsType_SerialTriplet] {
		length := len(group.Cards)
		if length < 3*3 {
			continue
		}
		divideStraight(group)
	}
	return sortCardGroups(groups)
}

//组合三张和三顺子成为 三带一，三带二，飞机
func combineThreeGroups(groups []*CardGroup) []*CardGroup {
	//排序现有牌组
	groups = sortCardGroups(groups)
	var groupsMap = make(map[msg.CardsType][]*CardGroup)
	for _, group := range groups {
		groupsMap[group.Type] = append(groupsMap[group.Type], group)
	}
	//如果没有三顺子和三张，直接返回
	if len(groupsMap[msg.CardsType_SerialTriplet]) == 0 && len(groupsMap[msg.CardsType_Triplet]) == 0 {
		return groups
	}
	var sllen = len(groupsMap[msg.CardsType_SingleCard])
	var dllen = len(groupsMap[msg.CardsType_Pair])
	//如果没有多余的单张，对子，直接返回, 这里不在从顺子中提取
	if sllen == 0 && dllen == 0 {
		return groups
	}
	//三顺子组合成飞机带单张，飞机带对子
	if len(groupsMap[msg.CardsType_SerialTriplet]) > 0 {
		gi := groupsMap[msg.CardsType_SerialTriplet][0]
		var count = len(gi.Cards) / 3
		tgs := selectGroupsEx(groupsMap[msg.CardsType_SingleCard], groupsMap[msg.CardsType_Pair], count)
		if nil == tgs {
			return groups
		}
		var tempct = 0
		for _, group := range tgs {
			tempct += len(group.Cards)
			gi.Cards = append(gi.Cards, group.Cards...)
		}
		if tempct == count {
			gi.Type = msg.CardsType_SerialTripletWithOne
		} else {
			ErrorCheck(tempct == 2*count, 1, " algorithm is error !!!")
			gi.Type = msg.CardsType_SerialTripletWithWing
		}
		for _, group := range tgs {
			groups = deleteCardsGroup(groups, group)
		}
		return combineThreeGroups(groups)
	}
	//三张组合成三带一，三带二
	if len(groupsMap[msg.CardsType_Triplet]) > 0 {
		var minSingle *CardGroup
		if sllen > 0 {
			minSingle = groupsMap[msg.CardsType_SingleCard][0]
		}
		var minDouble *CardGroup
		if dllen > 0 {
			minDouble = groupsMap[msg.CardsType_Pair][0]
		}
		tg := selectGroups(minSingle, minDouble)
		if nil == tg {
			return groups
		}
		gi := groupsMap[msg.CardsType_Triplet][0]
		gi.Cards = append(gi.Cards, tg.Cards...)
		if tg.Type == msg.CardsType_SingleCard {
			gi.Type = msg.CardsType_TripletWithSingle
		} else {
			gi.Type = msg.CardsType_TripletWithPair
		}
		groups = deleteCardsGroup(groups, tg)
		return combineThreeGroups(groups)
	}
	return sortCardGroups(groups)
}

func selectGroups(sl *CardGroup, dl *CardGroup) *CardGroup {
	if nil == sl && nil == dl {
		return nil
	}
	if nil == sl {
		return dl
	}
	if nil == dl {
		return sl
	}
	if sl.Value < dl.Value {
		return sl
	}
	return dl
}

func selectGroupsEx(sls []*CardGroup, dls []*CardGroup, count int) []*CardGroup {
	var sllen = len(sls)
	var dllen = len(dls)
	//单张和对子不够带的情况下
	if sllen+2*dllen < count {
		//原本设计从顺子，双顺子里面取，这里暂时选择不带
		return nil
	}
	var groups []*CardGroup
	if sllen+2*dllen == count {
		groups = append(groups, sls...)
		groups = append(groups, dls...)
		return groups
	}

	//下面的逻辑必满足条件 sllen+2*dllen > count
	if dllen == 0 {
		ErrorCheck(sllen >= count, 1, " algorithm is error !!!")
		for i := 0; i < count; i++ {
			groups = append(groups, sls[i])
		}
		return groups
	}

	//ErrorCheck(count == 2 || count == 3 || count == 4, 1, " algorithm is error !!!")
	//threelen最大为5, 如果为5，则sllen+2*dllen <= count, 前面已经过滤掉, 走到这里, 2x + y == 2，2x + y == 3 或者 2x + y == 4
	unpickDouble := func(src *CardGroup) *CardGroup {
		ErrorCheck(nil != src, 3, " src is nil !!!")
		ErrorCheck(src.Type == msg.CardsType_Pair, 3, " src type is not double !!!")
		//拆对子
		var group CardGroup
		group.Cards = src.Cards[:1]
		group.Type = msg.CardsType_SingleCard
		group.Value = src.Value
		//修改被拆对子
		src.Cards = src.Cards[1:]
		src.Type = msg.CardsType_SingleCard
		return &group
	}

	switch count {
	case 2:
		if sllen == 0 {
			ErrorCheck(2*dllen >= count, 1, " algorithm is error !!!")
			//带对子
			if dllen >= count {
				for i := 0; i < count; i++ {
					groups = append(groups, dls[i])
				}
				return groups
			}
			//带单张
			groups = append(groups, dls[0])
			return groups
		}
		ErrorCheck(sllen > 0 && dllen > 0, 1, " algorithm is error !!!")
		p1 := sllen > 1 && sls[1].Value < 13
		p2 := dllen > 1 && dls[1].Value < 13
		//带单张
		if p1 && p2 {
			//带单张
			if sls[0].Value+sls[1].Value < dls[0].Value+dls[1].Value {
				for i := 0; i < count; i++ {
					groups = append(groups, sls[i])
				}
				return groups
			}
			//带对子
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		//带单张
		if p1 {
			for i := 0; i < count; i++ {
				groups = append(groups, sls[i])
			}
			return groups
		}
		//带对子
		if p2 {
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		if sllen <= 1 {
			//带单张
			if dls[0].Value <= sls[0].Value {
				groups = append(groups, dls[0])
				return groups
			}
			//取单张
			groups = append(groups, sls[0])
			//拆对子，从中拿一张与单张一起
			groups = append(groups, unpickDouble(dls[0]))
			return groups
		}
		if dllen <= 1 {
			//带单张
			if sls[0].Value+sls[1].Value < 2*dls[0].Value {
				for i := 0; i < count; i++ {
					groups = append(groups, sls[i])
				}
				return groups
			}
			//带单张
			groups = append(groups, dls[0])
			return groups
		}
		//带单张
		if sls[0].Value+sls[1].Value < dls[0].Value+dls[1].Value {
			for i := 0; i < count; i++ {
				groups = append(groups, sls[i])
			}
			return groups
		}
		//带对子
		for i := 0; i < count; i++ {
			groups = append(groups, dls[i])
		}
		return groups
	case 3:
		if sllen == 0 {
			ErrorCheck(2*dllen >= count, 1, " algorithm is error !!!")
			//带对子
			if dllen >= count {
				for i := 0; i < count; i++ {
					groups = append(groups, dls[i])
				}
				return groups
			}
			ErrorCheck(dllen == 2, 1, " algorithm is error !!!")
			//带单张
			groups = append(groups, dls[0])
			//拆对子，从中拿一张与单张一起
			groups = append(groups, unpickDouble(dls[1]))
			return groups
		}
		ErrorCheck(sllen > 0 && dllen > 0, 1, " algorithm is error !!!")
		p1 := sllen > 2 && sls[2].Value < 13
		p2 := dllen > 2 && dls[2].Value < 13
		//带单张
		if p1 && p2 {
			//带单张
			if sls[0].Value+sls[1].Value+sls[2].Value < dls[0].Value+dls[1].Value+dls[2].Value {
				for i := 0; i < count; i++ {
					groups = append(groups, sls[i])
				}
				return groups
			}
			//带对子
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		//带单张
		if p1 {
			for i := 0; i < count; i++ {
				groups = append(groups, sls[i])
			}
			return groups
		}
		//带对子
		if p2 {
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		if sllen <= 2 || dllen <= 2 {
			//带单张
			groups = append(groups, sls[0])
			groups = append(groups, dls[0])
			return groups
		}
		//带单张
		if sls[0].Value+sls[1].Value+sls[2].Value < dls[0].Value+dls[1].Value+dls[2].Value {
			for i := 0; i < count; i++ {
				groups = append(groups, sls[i])
			}
			return groups
		}
		//带对子
		for i := 0; i < count; i++ {
			groups = append(groups, dls[i])
		}
		return groups
	case 4:
		if sllen == 0 {
			ErrorCheck(2*dllen >= count, 1, " algorithm is error !!!")
			//带对子
			if dllen >= count {
				for i := 0; i < count; i++ {
					groups = append(groups, dls[i])
				}
				return groups
			}
			//带单张
			groups = append(groups, dls[0])
			groups = append(groups, dls[1])
			return groups
		}
		ErrorCheck(sllen > 0 && dllen > 0, 1, " algorithm is error !!!")
		p1 := sllen > 3 && sls[3].Value < 13
		p2 := dllen > 3 && dls[3].Value < 13
		//带单张
		if p1 && p2 {
			//带单张
			if sls[0].Value+sls[1].Value+sls[2].Value+sls[3].Value < dls[0].Value+dls[1].Value+dls[2].Value+dls[3].Value {
				for i := 0; i < count; i++ {
					groups = append(groups, sls[i])
				}
				return groups
			}
			//带对子
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		//带单张
		if p1 {
			for i := 0; i < count; i++ {
				groups = append(groups, sls[i])
			}
			return groups
		}
		//带对子
		if p2 {
			for i := 0; i < count; i++ {
				groups = append(groups, dls[i])
			}
			return groups
		}
		if sllen <= 1 {
			//带单张
			if sls[0].Value < dls[1].Value {
				groups = append(groups, sls[0])
				groups = append(groups, dls[0])
				groups = append(groups, unpickDouble(dls[1]))
				return groups
			}
			//带单张
			groups = append(groups, dls[0])
			groups = append(groups, dls[1])
			return groups
		}

		if dllen <= 1 {
			//带单张
			if dls[0].Value < sls[2].Value {
				groups = append(groups, sls[0])
				groups = append(groups, sls[1])
				groups = append(groups, dls[0])
				return groups
			}
			//带单张
			groups = append(groups, sls[0])
			groups = append(groups, sls[1])
			groups = append(groups, sls[2])
			groups = append(groups, unpickDouble(dls[0]))
			return groups
		}
		ErrorCheck(sllen > 1 && dllen > 1, 1, " algorithm is error !!!")
		//带单张
		groups = append(groups, sls[0])
		groups = append(groups, sls[1])
		groups = append(groups, dls[0])
		return groups
	}
	return nil
}

func SearchLargerCardType(cardsDst []byte, cardsSrc []byte, isAppendBomb bool) (bool, [][]byte) {
	poker.SortCards(cardsDst)
	poker.SortCards(cardsSrc)
	//sort.Sort(CardSlice(cardsDst))
	//sort.Sort(CardSlice(cardsSrc))
	var exist bool
	var result [][]byte
	var cardstype = GetCardsType(cardsDst)
	//log.Debugf("目标牌型：%v %v", cardstype, cardsDst)
	switch cardstype {
	case msg.CardsType_Normal:
		for _, card := range cardsSrc {
			var cards []byte
			cards = append(cards, card)
			result = append(result, cards)
		}
		return true, result
	case msg.CardsType_SingleCard:
		fallthrough
	case msg.CardsType_Pair:
		fallthrough
	case msg.CardsType_Triplet:
		exist, result = SearchLargerSameCard(cardsDst, cardsSrc)
	case msg.CardsType_TripletWithSingle:
		fallthrough
	case msg.CardsType_TripletWithPair:
		exist, result = SearchLargerThreeAndEx(cardsDst, cardsSrc)
	case msg.CardsType_Sequence:
		exist, result = SearchLargerSameStraight(cardsDst, cardsSrc, 1)
	case msg.CardsType_SerialPair:
		exist, result = SearchLargerSameStraight(cardsDst, cardsSrc, 2)
	case msg.CardsType_SerialTriplet:
		exist, result = SearchLargerSameStraight(cardsDst, cardsSrc, 3)
	case msg.CardsType_SerialTripletWithOne:
		fallthrough
	case msg.CardsType_SerialTripletWithWing:
		exist, result = SearchLargerAirplane(cardsDst, cardsSrc)
	case msg.CardsType_QuartetWithTwo:
		exist, result = SearchLargerFourAndTwo(cardsDst, cardsSrc)
	case msg.CardsType_Bomb:
		if isAppendBomb {
			exist, result = SearchLargerSameCard(cardsDst, cardsSrc)
		} else {
			return false, nil
		}
	case msg.CardsType_Rocket:
		return false, nil
	}

	if isAppendBomb {
		if cardstype > msg.CardsType_Normal && cardstype < msg.CardsType_Bomb {
			if exb, bombs := SearchAllBombs(cardsSrc); exb {
				exist = true
				result = append(result, bombs...)
			}
			if exr, rocket := SearchRocket(cardsSrc); exr {
				exist = true
				result = append(result, rocket)
			}
		} else if cardstype == msg.CardsType_Bomb {
			if exr, rocket := SearchRocket(cardsSrc); exr {
				exist = true
				result = append(result, rocket)
			}
		}
	}
	//特殊处理，防止飞机带单张，遇到JJJJQQQQ这种情况
	if cardstype == msg.CardsType_SerialTripletWithOne {
		var realResult [][]byte
		for _, cards := range result {
			bombct, _ := GetCardCounts(cards)
			ctype := GetCardsType(cards)
			if ctype < msg.CardsType_Bomb && bombct*4 == len(cards) {
				continue
			}
			realResult = append(realResult, cards)
		}
		result = realResult
		exist = len(result) > 0
	}
	//排序
	for _, cards := range result {
		poker.SortCards(cards)
		//sort.Sort(CardSlice(cards))
	}
	//返回结果检测
	for _, cards := range result {
		temp := CloneCards(cardsSrc)
		ok, _ := RemoveCards(temp, cards)
		ErrorCheck(ok, 1, fmt.Sprintf("check result -> cards must within temp, maybe algorithm error, %v, %v !!!", cardsDst, cards))
		ErrorCheck(CompareCards(cardsDst, cards), 1, fmt.Sprintf("check result -> cards must larger than dest, maybe algorithm error, %v, %v !!!", cardsDst, cards))
	}
	if exist {
		ErrorCheck(len(result) > 0, 1, fmt.Sprintf("check result -> result len must larger than 0, maybe algorithm error, %v !!!", result))
	} else {
		ErrorCheck(len(result) == 0, 1, fmt.Sprintf("check result -> result len must equal to 0, maybe algorithm error, %v !!!", result))
	}
	return exist, result
}

//查找较大的单张，对子，三张， 炸弹
func SearchLargerSameCard(cardsDst []byte, cardsSrc []byte) (bool, [][]byte) {
	var length = len(cardsDst)
	if len(cardsSrc) < length {
		return false, nil
	}
	var exist bool
	var result [][]byte
	_, counts := GetCardCounts(cardsSrc)
	for num, ct := range counts {
		if ct >= length && byte(num) > cardsDst[0]>>4 {
			var rtcards []byte
			for _, card := range cardsSrc {
				if card>>4 == byte(num) {
					rtcards = append(rtcards, card)
					if len(rtcards) >= length {
						break
					}
				}
			}
			exist = true
			result = append(result, rtcards)
		}
	}
	return exist, result
}

//查找较大的三带一, 三带二
func SearchLargerThreeAndEx(cardsDst []byte, cardsSrc []byte) (bool, [][]byte) {
	var length = len(cardsDst)
	if len(cardsSrc) < length {
		return false, nil
	}
	var exist bool
	var result [][]byte
	_, counts := GetCardCounts(cardsSrc)
	for num, ct := range counts {
		if ct >= 3 && byte(num) > cardsDst[2]>>4 {
			var rtcards []byte
			for _, card := range cardsSrc {
				if card>>4 == byte(num) {
					rtcards = append(rtcards, card)
					if len(rtcards) >= 3 {
						break
					}
				}
			}

			var getct = 0
			for num2, ct2 := range counts {
				if num2 != num && ct2 >= length-3 {
					for _, card := range cardsSrc {
						if card>>4 == byte(num2) {
							rtcards = append(rtcards, card)
							getct++
							if getct >= length-3 {
								break
							}
						}
					}
				}
				if getct >= length-3 {
					break
				}
			}

			if len(rtcards) == length {
				exist = true
				result = append(result, rtcards)
			}
		}
	}
	return exist, result
}

//查找较大的单顺子, 双顺子，三顺子
func SearchLargerSameStraight(cardsDst []byte, cardsSrc []byte, samecount int) (bool, [][]byte) {
	var length = len(cardsDst)
	if len(cardsSrc) < length {
		return false, nil
	}
	//到A的顺子直接返回
	if cardsDst[length-1]>>4 >= 12 {
		return false, nil
	}
	var exist bool
	var result [][]byte
	_, counts := GetCardCounts(cardsSrc)
	var startnum byte
	var endnum byte
	for num := cardsDst[0]>>4 + 1; num < 13; num++ {
		if counts[num] >= samecount {
			if startnum == 0 {
				startnum = num
			}
			endnum = num
			if endnum-startnum+1 >= byte(length/samecount) {
				var rtcards []byte
				for i := startnum; i <= endnum; i++ {
					var getct = 0
					for _, card := range cardsSrc {
						if card>>4 == i {
							rtcards = append(rtcards, card)
							getct++
							if getct >= samecount {
								break
							}
						}
					}
				}
				exist = true
				result = append(result, rtcards)
				num = startnum
				startnum = 0
				endnum = 0
				continue
			}
		} else {
			startnum = 0
			endnum = 0
		}
	}
	return exist, result
}

//查找比较大的飞机
func SearchLargerAirplane(cardsDst []byte, cardsSrc []byte) (bool, [][]byte) {
	var length = len(cardsDst)
	if len(cardsSrc) < length {
		return false, nil
	}
	var exist bool
	var result [][]byte
	_, dstCounts := GetCardCounts(cardsDst)
	start, end := GetAirPlaneStartEnd(cardsDst)
	var tempcards []byte
	for num := start; num <= end; num++ {
		var getct = 0
		for _, card := range cardsDst {
			if card>>4 == byte(num) {
				tempcards = append(tempcards, card)
				getct++
				if getct >= 3 {
					break
				}
			}
		}
		dstCounts[num] -= 3
	}
	var totalcount = 0
	for _, ct := range dstCounts {
		totalcount += ct
	}
	var excount = totalcount / (end - start + 1)
	if flag, straights := SearchLargerSameStraight(tempcards, cardsSrc, 3); flag {
		_, srcCounts := GetCardCounts(cardsSrc)
		var ok bool
		if excount == 1 {
			for _, cards := range straights {
				counts := CloneInts(srcCounts)
				for num := cards[0] >> 2; num <= cards[len(cards)-1]>>4; num++ {
					counts[num] -= 3
				}
				var allcount = 0
				for _, ct := range counts {
					allcount += ct
				}
				var templen = cards[len(cards)-1]>>4 - cards[0]>>4 + 1
				if allcount >= int(templen) {
					varCards := CloneCards(cardsSrc)
					ok, varCards = RemoveCards(varCards, cards)
					ErrorCheck(ok, 1, " straights cards must within varCards 1, maybe error occur !!!")

					var rtcards []byte
					rtcards = append(rtcards, cards...)

					var getct = 0
					for _, card := range varCards {
						rtcards = append(rtcards, card)
						getct++
						if getct >= int(templen) {
							break
						}
					}
					exist = true
					result = append(result, rtcards)
				}
			}
		} else if excount == 2 {
			//这里不找 6667778888 等带炸的飞机带对子
			for _, cards := range straights {
				counts := CloneInts(srcCounts)
				for num := cards[0] >> 2; num <= cards[len(cards)-1]>>4; num++ {
					counts[num] -= 3
				}
				var doublecount = 0
				for _, ct := range counts {
					if ct >= 2 {
						doublecount++
					}
				}
				var templen = cards[len(cards)-1]>>4 - cards[0]>>4 + 1
				if doublecount >= int(templen) {
					varCards := CloneCards(cardsSrc)
					ok, varCards = RemoveCards(varCards, cards)
					ErrorCheck(ok, 1, " straights cards must within varCards 2, maybe error occur !!!")

					var rtcards []byte
					rtcards = append(rtcards, cards...)

					var getct = 0
					for num, ct := range counts {
						if ct >= 2 {
							getct++
							var findct = 0
							for _, card := range varCards {
								if card>>4 == byte(num) {
									rtcards = append(rtcards, card)
									findct++
									if findct >= 2 {
										break
									}
								}
							}
						}
						if getct >= int(templen) {
							break
						}
					}
					exist = true
					result = append(result, rtcards)
				}
			}
		}
	}
	return exist, result
}

//查找比较大的四带二单，四带二对
func SearchLargerFourAndTwo(cardsDst []byte, cardsSrc []byte) (bool, [][]byte) {
	var length = len(cardsDst)
	if len(cardsSrc) < length {
		return false, nil
	}
	var exct = (length - 4) / 2
	var exist bool
	var result [][]byte
	_, num := IsHasBomb(cardsDst)
	var tempcards []byte
	tempcards = append(tempcards, byte(num<<4)|byte(DIAMOND), byte(num)<<4|byte(PLUM), byte(num)<<4|byte(HEART), byte(num)<<4|byte(SPADE))
	if flag, bombs := SearchLargerSameCard(tempcards, cardsSrc); flag {
		_, counts := GetCardCounts(cardsSrc)
		if exct == 1 {
			for _, cards := range bombs {
				var tempct = 0
				for num, ct := range counts {
					if byte(num) != cards[0]>>4 {
						tempct += ct
					}
				}
				if tempct >= 2 {
					var rtcards []byte
					rtcards = append(rtcards, cards...)
					var getct = 0
					for _, card := range cardsSrc {
						if card>>4 != cards[0]>>4 {
							rtcards = append(rtcards, card)
							getct++
							if getct >= 2 {
								break
							}
						}
					}
					exist = true
					result = append(result, rtcards)
				}
			}
		} else if exct == 2 {
			//这里不找 77778888 等连炸牌型的4带2
			for _, cards := range bombs {
				var tempct = 0
				for num, ct := range counts {
					if byte(num) != cards[0]>>4 && ct >= 2 {
						tempct++
					}
				}
				if tempct >= 2 {
					var rtcards []byte
					rtcards = append(rtcards, cards...)
					var getct = 0
					for num, ct := range counts {
						if byte(num) != cards[0]>>4 && ct >= 2 {
							getct++
							var findct = 0
							for _, card := range cardsSrc {
								if card>>4 == byte(num) {
									rtcards = append(rtcards, card)
									findct++
									if findct >= 2 {
										break
									}
								}
							}
						}
						if getct >= 2 {
							break
						}
					}
					exist = true
					result = append(result, rtcards)
				}
			}
		}
	}
	return exist, result
}

//查找所有炸弹
func SearchAllBombs(cards []byte) (bool, [][]byte) {
	var exist bool
	var result [][]byte
	_, counts := GetCardCounts(cards)
	for num, ct := range counts {
		if ct == 4 {
			exist = true
			var rtcards []byte
			rtcards = append(rtcards, byte(num<<4)|byte(DIAMOND), byte(num)<<4|byte(PLUM), byte(num)<<4|byte(HEART), byte(num)<<4|byte(SPADE))
			log.Debugf("%v %v %v", rtcards, (byte(num) << 4), (PLUM))
			result = append(result, rtcards)
		}
	}
	return exist, result
}

//查找火箭
func SearchRocket(cards []byte) (bool, []byte) {
	var exist bool
	var result []byte
	_, counts := GetCardCounts(cards)
	if counts[14] >= 1 && counts[15] >= 1 {
		exist = true
		result = append(result, 0xe1, 0xf1)
	}
	return exist, result
}

//查找牌组中小于目标牌的个数
func SearchLessGroupCount(groups []*CardGroup, cards []byte) (count int) {
	ErrorCheck(len(groups) > 0, 3, " SearchFirstMoreGroup param is illegal !!!")
	//浅拷贝
	tempGroups := CloneCardGroups(groups)
	//按价值排序，不按牌型
	tempGroups = SortCardGroupsByValue(tempGroups)

	cardstype := GetCardsType(cards)
	for _, group := range tempGroups {
		if cardstype == group.Type && !CompareCards(cards, group.Cards) {
			count++
		}
	}

	return
}

//查找牌组中第一个(价值最小)能打过的牌
func SearchFirstLargeGroup(groups []*CardGroup, cards []byte, isContainBomb bool) *CardGroup {
	ErrorCheck(len(groups) > 0, 3, " SearchFirstMoreGroup param is illegal !!!")
	//浅拷贝
	tempGroups := CloneCardGroups(groups)
	//按价值排序，不按牌型
	tempGroups = SortCardGroupsByValue(tempGroups)
	var rtGroup *CardGroup
	cardstype := GetCardsType(cards)
	if cardstype == msg.CardsType_Rocket {
		return nil
	} else if cardstype == msg.CardsType_Bomb {
		if isContainBomb {
			for _, group := range tempGroups {
				if group.Type >= msg.CardsType_Bomb && CompareCards(cards, group.Cards) {
					rtGroup = group
					break
				}
			}
		}
	} else {
		for _, group := range tempGroups {
			if isContainBomb && group.Type >= msg.CardsType_Bomb {
				rtGroup = group
				break
			} else if group.Type == cardstype && CompareCards(cards, group.Cards) {
				rtGroup = group
				break
			}
		}
	}
	return rtGroup
}

//查找牌组中最后一个(价值最大)能打过的牌
func SearchLastLargeGroup(groups []*CardGroup, cards []byte, isContainBomb bool) *CardGroup {
	ErrorCheck(len(groups) > 0, 3, " SearchLastMoreGroup param is illegal !!!")
	//浅拷贝
	tempGroups := CloneCardGroups(groups)
	//按价值排序，不按牌型
	tempGroups = SortCardGroupsByValue(tempGroups)
	var rtGroup *CardGroup
	cardstype := GetCardsType(cards)
	if cardstype == msg.CardsType_Rocket {
		return nil
	} else if cardstype == msg.CardsType_Bomb {
		if isContainBomb {
			for i := len(tempGroups) - 1; i >= 0; i-- {
				group := tempGroups[i]
				if group.Type >= msg.CardsType_Bomb && CompareCards(cards, group.Cards) {
					rtGroup = group
					break
				}
			}
		}
	} else {
		for i := len(tempGroups) - 1; i >= 0; i-- {
			group := tempGroups[i]
			if isContainBomb && group.Type >= msg.CardsType_Bomb {
				rtGroup = group
				break
			} else if group.Type == cardstype && CompareCards(cards, group.Cards) {
				rtGroup = group
				break
			}
		}
	}
	return rtGroup
}

//查找牌组中所有能打过当前牌的牌组
func SearchAllLargeGroups(groups []*CardGroup, cards []byte, isContainBomb bool) []*CardGroup {
	//浅拷贝
	tempGroups := CloneCardGroups(groups)
	//按价值排序，不按牌型
	tempGroups = SortCardGroupsByValue(tempGroups)
	var rtGroups []*CardGroup

	cardstype := GetCardsType(cards)
	if cardstype == msg.CardsType_Rocket {
		return nil
	} else if cardstype == msg.CardsType_Bomb {
		if isContainBomb {
			for _, group := range tempGroups {
				if group.Type >= msg.CardsType_Bomb && CompareCards(cards, group.Cards) {
					rtGroups = append(rtGroups, group)
					break
				}
			}
		}
	} else {
		for _, group := range tempGroups {
			if isContainBomb && group.Type >= msg.CardsType_Bomb {
				rtGroups = append(rtGroups, group)
			} else if group.Type == cardstype && CompareCards(cards, group.Cards) {
				rtGroups = append(rtGroups, group)
			}
		}
	}
	return rtGroups
}

//查找能大过当前牌组的数量
func SearchMoreThanCount(groups []*CardGroup, cards []byte, isContainBomb bool) int {
	var ct = 0
	for _, group := range groups {
		if !isContainBomb && group.Type >= msg.CardsType_Bomb {
			continue
		}
		if exist, _ := SearchLargerCardType(group.Cards, cards, isContainBomb); exist {
			ct++
		}
	}
	return ct
}

//获取group的索引号
func GetGroupIndex(groups []*CardGroup, cp *CardGroup) int {
	for k, group := range groups {
		if group.IsEuqal(cp) {
			return k
		}
	}
	return -1
}

//获取groups中炸弹的数量和非炸弹的数量
func GetGroupsBombCnt(groups []*CardGroup) (int, int) {
	var bombCnt, exCnt int
	for _, group := range groups {
		if group.Type >= msg.CardsType_Bomb {
			bombCnt++
		} else {
			exCnt++
		}
	}
	return bombCnt, exCnt
}

//是否所有牌均是单牌
func IsGroupsAllSingle(groups []*CardGroup) bool {
	for _, group := range groups {
		if group.Type != msg.CardsType_SingleCard {
			return false
		}
	}
	return true
}

//是否所有牌均是对子
func IsGroupsAllDouble(groups []*CardGroup) bool {
	for _, group := range groups {
		if group.Type != msg.CardsType_Pair {
			return false
		}
	}
	return true
}

//是否有王炸
func IsHasRocket(groups []*CardGroup) bool {
	for _, group := range groups {
		if group.Type == msg.CardsType_Rocket {
			return true
		}
	}
	return false
}

//判断手牌是否符合一组小牌，其他绝对大牌的赢牌路径
func IsCanAbsWin(groups []*CardGroup, lefts []byte, isAbs bool) bool {
	ErrorCheck(len(groups) > 0 && len(lefts) > 0, 3, " IsCanAbsWin param is illegal !!!")
	return SearchMoreThanCount(groups, lefts, isAbs) <= 1
}

//判断跟牌之后是否能稳赢, isAbs为true，如果返回true绝对能赢， isAbs为false，如果返回true相对能赢，如果别人手上没有更多的炸弹，绝对能赢
func GenCardCanAbsWin(groups []*CardGroup, hands []byte, outs []byte, lefts []byte, isAbs bool) (bool, []byte) {
	ErrorCheck(len(groups) > 0 && len(hands) > 0 && len(outs) > 0 && len(lefts) > 0, 3, " GenCardCanAbsWin param is illegal !!!")
	var bRet bool
	var rtCards []byte
	largeGroups := SearchAllLargeGroups(groups, outs, true)
	for _, group := range largeGroups {
		found, _ := SearchLargerCardType(group.Cards, lefts, isAbs)
		if found == true {
			return false, nil
		}
	}
	c := 0
	for _, group := range groups {
		found, _ := SearchLargerCardType(group.Cards, lefts, isAbs)
		if found == true {
			c++
		}
	}
	if c > 1 {
		return false, nil
	}

	for _, group := range largeGroups {
		k := GetGroupIndex(groups, group)
		ErrorCheck(k >= 0, 1, " k must more than 0, maybe error occur !!!")
		//新建一个切片，不改变原有的
		var tempGroups []*CardGroup
		tempGroups = append(tempGroups, groups[:k]...)
		tempGroups = append(tempGroups, groups[k+1:]...)
		exist, _ := SearchLargerCardType(group.Cards, lefts, isAbs)
		bombCnt, _ := GetGroupsBombCnt(tempGroups)
		moreCnt := SearchMoreThanCount(tempGroups, lefts, isAbs)
		if moreCnt <= 1 || bombCnt >= moreCnt || (!exist && bombCnt >= moreCnt-1) {
			bRet = true
			rtCards = group.Cards
			break
		}
	}
	if bRet {
		return bRet, rtCards
	}
	//然后打破牌组，在所有手牌中搜索
	var exist bool
	var result [][]byte
	if exist, result = SearchLargerCardType(outs, hands, true); exist {
		for _, cards := range result {
			tempCards := CloneCards(hands)
			var ok bool
			ok, tempCards = RemoveCards(tempCards, cards)
			ErrorCheck(ok, 1, " cards must within tempcards, maybe error occur !!!")
			_, cardGroups := GetMostValueGroup(tempCards)
			exist, _ := SearchLargerCardType(cards, lefts, isAbs)
			bombCnt, _ := GetGroupsBombCnt(cardGroups)
			moreCnt := SearchMoreThanCount(cardGroups, lefts, isAbs)
			if moreCnt <= 1 || bombCnt >= moreCnt || (!exist && bombCnt >= moreCnt-1) {
				bRet = true
				rtCards = cards
				break
			}
		}
	}
	return bRet, rtCards
}

//拆最小对子为单张出牌
func ChaiLessPairs(groups []*CardGroup) byte {
	var temp byte
	for _, group := range groups {
		if group.Type == msg.CardsType_Pair {
			temp = group.Cards[0]
			break
		}
	}
	return temp
}

//拆最大对子为单张出牌
func ChaiLargePairs(groups []*CardGroup) byte {
	var temp byte
	for _, group := range groups {
		if group.Type == msg.CardsType_Pair {
			temp = group.Cards[0]
		}
	}
	return temp
}
