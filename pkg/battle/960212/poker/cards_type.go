package poker

import "game_poker/doudizhu/msg"

// GetCardsType 获取牌型
func GetCardsType(cards []byte) (cardsType msg.CardsType) {

	// 是否单张
	if IsSingleCard(cards) {
		return msg.CardsType_SingleCard
	}

	// 是否对子
	if IsPair(cards) {
		return msg.CardsType_Pair
	}

	// 是否三条
	if IsTriplet(cards) {
		return msg.CardsType_Triplet
	}

	// 是否三带一
	if IsTripletWithSingle(cards) {
		return msg.CardsType_TripletWithSingle
	}

	// 是否三带一对
	if IsTripletWithPair(cards) {
		return msg.CardsType_TripletWithPair
	}

	// 是否顺子
	if IsSequence(cards) {
		return msg.CardsType_Sequence
	}

	// 是否连对
	if IsSerialPair(cards) {
		return msg.CardsType_SerialPair
	}

	// 是否飞机
	if IsSerialTriplet(cards) {
		return msg.CardsType_SerialTriplet
	}

	// 是否飞机带单张
	if IsSerialTripletWithOne(cards) {
		return msg.CardsType_SerialTripletWithOne
	}

	// 是否飞机带对子
	if IsSerialTripletWithWing(cards) {
		return msg.CardsType_SerialTripletWithWing
	}

	// 是否四带二单牌
	if IsQuartetWithTwo(cards) {
		return msg.CardsType_QuartetWithTwo
	}

	// 是否四带二对子
	if IsQuartetWithTwoPair(cards) {
		return msg.CardsType_QuartetWithTwoPair
	}

	// 是否炸弹
	if IsBomb(cards) {
		return msg.CardsType_Bomb
	}

	// 是否火箭
	if IsRocket(cards) {
		return msg.CardsType_Rocket
	}

	return
}

// IsSingleCard 是否是单张牌
func IsSingleCard(cards []byte) bool {

	// 牌数检测
	if len(cards) != 1 {
		return false
	}

	return true
}

// IsPair 是否是对子
func IsPair(cards []byte) bool {

	// 牌数检测
	if len(cards) != 2 {
		return false
	}

	// 牌值检测
	value1, _ := GetCardValueAndColor(cards[0])
	value2, _ := GetCardValueAndColor(cards[1])
	if value1 != value2 {
		return false
	}

	return true
}

// IsTriplet 是否三同张
func IsTriplet(cards []byte) bool {

	if len(cards) != 3 {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 牌值检测
	firstValue, _ := GetCardValueAndColor(cards[0])
	lastValue, _ := GetCardValueAndColor(cards[len(cards)-1])
	if firstValue != lastValue {
		return false
	}

	return true
}

// IsTripletWithSingle 是否三带一
func IsTripletWithSingle(cards []byte) bool {
	if len(cards) != 4 {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == 1 {
		return true
	}

	return false
}

// IsTripletWithPair 是否三带一对
func IsTripletWithPair(cards []byte) bool {

	// 牌数检测
	if len(cards) != 5 {
		return false
	}

	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == 1 && len(repeatedArr[1]) == 1 {
		return true
	}

	return false
}

// IsSequence 是否顺子
func IsSequence(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count < 5 || count > 12 {
		return false
	}

	// 有2或者双王
	if HaveBigCard(cards) {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 第一个值
	firstValue, _ := GetCardValueAndColor(cards[0])

	// 正序排序后判断值是否连续
	for i, card := range cards {
		value, _ := GetCardValueAndColor(card)

		if firstValue+byte(i) != value {
			return false
		}
	}

	return true
}

// IsSerialPair 是否连对
func IsSerialPair(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count%2 != 0 || count < 6 {
		return false
	}

	// 有2或者双王
	if HaveBigCard(cards) {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[1]) == count/2 && int(repeatedArr[1][len(repeatedArr[1])-1]-repeatedArr[1][0]) == count/2-1 {
		return true
	}

	return false
}

// IsSerialTriplet 是否飞机
func IsSerialTriplet(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count%3 != 0 || count < 6 || count > 18 {
		return false
	}

	// 有2或者双王
	if HaveBigCard(cards) {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == count/3 && int(repeatedArr[2][len(repeatedArr[2])-1]-repeatedArr[2][0]) == count/3-1 {
		return true
	}

	return false
}

// IsSerialTripletWithOne 是否飞机带单牌
func IsSerialTripletWithOne(cards []byte) bool {
	// 牌数检测
	count := len(cards)
	if count%4 != 0 && count < 8 {
		return false
	}

	// 双王检测
	if HaveRocket(cards) {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	// 飞机长度
	_, planeLen := GetPlane(cards)

	// 重复 4 张牌个数
	repeatFourCount := len(repeatedArr[3])

	// 飞机个数检测,包含炸弹检测
	if planeLen == 0 || repeatFourCount > 0 || count/planeLen != 4 || count%planeLen != 0 {
		return false
	}

	return true
}

// IsSerialTripletWithWing 是否飞机带对子
func IsSerialTripletWithWing(cards []byte) bool {
	// 牌数检测
	count := len(cards)
	if count%5 != 0 && count < 10 {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	// 飞机有2检测
	for _, value := range repeatedArr[2] {
		if value == 0xd {
			return false
		}
	}

	// 重复个数 2/3 张牌个数
	repeatTwoCount := len(repeatedArr[1])
	repeatThreeCount := len(repeatedArr[2])

	// 飞机和翅膀个数检测
	if repeatTwoCount == 0 || repeatThreeCount == 0 ||
		count/repeatTwoCount != 5 || count%repeatTwoCount != 0 ||
		count/repeatThreeCount != 5 || count%repeatThreeCount != 0 {
		return false
	}

	// 飞机是否连续检测
	if int(repeatedArr[2][repeatThreeCount-1]-repeatedArr[2][0]) != repeatThreeCount-1 {
		return false
	}

	return true
}

// IsQuartetWithTwo 是否四带二
func IsQuartetWithTwo(cards []byte) bool {
	count := len(cards)

	// 牌数检测
	if count != 6 {
		return false
	}

	// 双王检测
	if HaveRocket(cards) {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	// 检测4同张个数
	if len(repeatedArr[3]) != 1 {

		return false
	}

	return true
}

// IsQuartetWithTwoPair 是否四带两对
func IsQuartetWithTwoPair(cards []byte) bool {
	count := len(cards)

	// 牌数检测
	if count != 8 {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	// 检测4同张个数 和 带牌对子个数
	if len(repeatedArr[3]) != 1 || len(repeatedArr[1]) != 2 {

		return false
	}

	return true
}

// IsBomb 是否炸弹
func IsBomb(cards []byte) bool {
	if len(cards) != 4 {
		return false
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 获取第一张和最后一张牌值比较
	valueFirst, _ := GetCardValueAndColor(cards[0])
	valueLast, _ := GetCardValueAndColor(cards[len(cards)-1])

	if valueFirst == valueLast {
		return true
	}

	return false
}

// IsRocket 是否火箭
func IsRocket(cards []byte) bool {

	if len(cards) == 2 && HaveRocket(cards) {
		return true
	}

	return false
}

// 获取飞机部分
// 返回飞机部分，以及它带飞机长度
func GetPlane(inCards []byte) (planeCards []byte, length int) {
	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range inCards {
		cards = append(cards, card)
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	var (
		firstValue byte   // 飞机第一个值
		add        byte   // 轮询递增值
		planeArr   []byte // 飞机组
	)

	// 遍历重复个数为3 的值，寻找飞机
	for _, v := range repeatedArr[2] {
		if v == firstValue+add && v != 0xd {
			planeArr = append(planeArr, v)
			add++
		} else if v > firstValue+add {
			// 飞机长度已经 > 1
			if len(planeArr) > 1 {
				break
			}

			add = 1
			firstValue = v
			planeArr = []byte{firstValue}
		}
	}

	length = len(planeArr)

	// 不满足正常飞机长度
	if length < 2 {
		return
	}

	// 从选定到牌值在牌组中找到对应到牌
	for _, value := range planeArr {
		for i := 0; i < 3; i++ {
			for i, card := range cards {
				cardValue, _ := GetCardValueAndColor(card)
				if value == cardValue {
					planeCards = append(planeCards, card)
					cards = append(cards[:i], cards[i+1:]...)
					break
				}
			}
		}

	}

	return
}
