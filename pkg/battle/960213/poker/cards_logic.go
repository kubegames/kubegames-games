package poker

import (
	"fmt"

	"github.com/kubegames/kubegames-games/pkg/battle/960213/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// GetCardsWeightValue 获取牌组权重值
func GetCardsWeightValue(userCards []byte, cardsType msg.CardsType) (cardsWeightValue byte) {

	// 牌值排序
	cards := PositiveSortCards(userCards)

	repeatedArr := CheckRepeatedCards(cards)

	switch cardsType {
	// 单张
	case msg.CardsType_SingleCard:
		cardsWeightValue = repeatedArr[0][0]
		break

	// 对子
	case msg.CardsType_Pair:
		cardsWeightValue = repeatedArr[1][0]
		break

	// 三条
	case msg.CardsType_Triplet:
		cardsWeightValue = repeatedArr[2][0]
		break

	// 三带一
	case msg.CardsType_TripletWithSingle:
		cardsWeightValue = repeatedArr[2][0]
		break

	// 三带二
	case msg.CardsType_TripletWithPair:
		cardsWeightValue = repeatedArr[2][0]
		break

	// 顺子
	case msg.CardsType_Sequence:
		cardsWeightValue = repeatedArr[0][len(repeatedArr[0])-1]
		break

	// 连对
	case msg.CardsType_SerialPair:
		cardsWeightValue = repeatedArr[1][len(repeatedArr[1])-1]
		break

	// 飞机
	case msg.CardsType_SerialTriplet:
		cardsWeightValue = repeatedArr[2][len(repeatedArr[2])-1]
		break

	// 飞机带单张
	case msg.CardsType_SerialTripletWithOne:
		planeCards, _ := GetPlane(cards)
		cardsWeightValue, _ = GetCardValueAndColor(planeCards[len(planeCards)-1])
		break

	// 飞机带对子
	case msg.CardsType_SerialTripletWithWing:
		cardsWeightValue = repeatedArr[2][len(repeatedArr[2])-1]
		break

	// 四带二单张
	case msg.CardsType_QuartetWithTwo:
		cardsWeightValue = repeatedArr[3][0]
		break

	// 四带二对子
	case msg.CardsType_QuartetWithTwoPair:
		cardsWeightValue = repeatedArr[3][0]
		break

	// 炸弹
	case msg.CardsType_Bomb:
		cardsWeightValue = repeatedArr[3][0]
		break

	// 火箭
	case msg.CardsType_Rocket:
		cardsWeightValue = repeatedArr[0][0]
		break
	}

	if cardsWeightValue == 0 {
		log.Errorf("出现了为0权重，牌型或者牌组不正确")
	}

	return
}

// TakeOverCards 最小接牌
func TakeOverCards(handCards HandCards, userCards []byte) (takeCards []byte) {

	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range userCards {
		cards = append(cards, card)
	}

	// 正序排序
	cards = PositiveSortCards(cards)

	// 被接手牌的长度
	handLen := len(handCards.Cards)

	var valueArr []byte

	// 相同牌型接管
	switch handCards.CardsType {

	// 单张
	case int32(msg.CardsType_SingleCard):

		for _, card := range cards {
			cardValue, _ := GetCardValueAndColor(card)
			if cardValue > handCards.WeightValue {
				valueArr = []byte{cardValue}
				break
			}
		}

	// 对子
	case int32(msg.CardsType_Pair):

		// 所有的 重复次数超过 2次的 值数组
		Repeated2ValueArr := GetRepeatedValueArr(cards, 2)

		for _, value := range Repeated2ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value}
				break
			}
		}

	// 三条
	case int32(msg.CardsType_Triplet):

		// 所有的 重复次数超过 3次的 值数组
		Repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		for _, value := range Repeated3ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value}
				break
			}
		}

	// 三带一
	case int32(msg.CardsType_TripletWithSingle):

		// 所有的 重复次数超过 3次的 值数组
		Repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		for _, value := range Repeated3ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value}
				break
			}
		}

		// 找到 三同张, 找翅膀
		if len(valueArr) == 3 {

			// 找不同于 三同张 的 单张
			for _, card := range cards {
				value, _ := GetCardValueAndColor(card)
				if !InByteArr(value, valueArr) {
					valueArr = append(valueArr, value)
					break
				}
			}

		}

		// 没找到同长度，至空
		if len(valueArr) != handLen {
			valueArr = []byte{}
		}

	// 三带一对
	case int32(msg.CardsType_TripletWithPair):

		// 所有的 重复次数超过 3次的 值数组
		Repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		for _, value := range Repeated3ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value}
				break
			}
		}

		// 找到 三同张, 找翅膀
		if len(valueArr) == 3 {

			// 所有的 重复次数超过 2次的 值数组
			Repeated2ValueArr := GetRepeatedValueArr(cards, 2)

			// 找不同于 三同张 的 对子
			for _, value := range Repeated2ValueArr {

				if !InByteArr(value, valueArr) {
					valueArr = append(valueArr, value, value)
					break
				}
			}

		}

		// 没找到同长度，至空
		if len(valueArr) != handLen {
			valueArr = []byte{}
		}

	// 顺子
	case int32(msg.CardsType_Sequence):

		// 所有的 值数组
		repeated1ValueArr := GetRepeatedValueArr(cards, 1)

		// 从 被接牌顺子最大值+1 开始 寻找 同长度 顺子
		for i := handCards.WeightValue + 1; i <= 0xc; i++ {

			allSelected := true
			for j := 0; j < handLen; j++ {
				value := i - byte(j)

				if InByteArr(value, repeated1ValueArr) {
					valueArr = append(valueArr, value)
				} else {
					allSelected = false
					valueArr = []byte{}
					break
				}
			}

			// 找到同长度 比 被接牌 大的 顺子
			if allSelected {
				break
			}
		}

	// 连对
	case int32(msg.CardsType_SerialPair):

		// 所有的 重复次数超过 2次的 值数组
		repeated2ValueArr := GetRepeatedValueArr(cards, 2)

		for i := handCards.WeightValue + 1; i <= 0xc; i++ {

			allSelected := true
			for j := 0; j < handLen/2; j++ {

				value := i - byte(j)

				if InByteArr(value, repeated2ValueArr) {
					valueArr = append(valueArr, value, value)
				} else {
					allSelected = false
					valueArr = []byte{}
					break
				}
			}

			// 找到同长度 比 被接牌 大的 连对
			if allSelected {
				break
			}

		}

	// 飞机
	case int32(msg.CardsType_SerialTriplet):

		// 所有的 重复次数超过 3次的 值数组
		repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		for i := handCards.WeightValue + 1; i <= 0xc; i++ {

			allSelected := true
			for j := 0; j < handLen/3; j++ {

				value := i - byte(j)

				if InByteArr(value, repeated3ValueArr) {
					valueArr = append(valueArr, value, value, value)
				} else {
					allSelected = false
					valueArr = []byte{}
					break
				}
			}

			// 找到同长度 比 被接牌 大的 飞机
			if allSelected {
				break
			}

		}

	// 飞机带单张
	case int32(msg.CardsType_SerialTripletWithOne):

		// 所有的 重复次数超过 3次的 值数组
		repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		// 飞机长度 (被接牌 重复次数 超过 3次 的 值数组 长度)
		planeLen := len(GetRepeatedValueArr(handCards.Cards, 3))

		for i := handCards.WeightValue + 1; i <= 0xc; i++ {

			allSelected := true
			for j := 0; j < planeLen; j++ {

				value := i - byte(j)

				if InByteArr(value, repeated3ValueArr) {
					valueArr = append(valueArr, value, value, value)
				} else {
					allSelected = false
					valueArr = []byte{}
					break
				}
			}

			// 找到同长度 比 被接牌 大的 飞机, 找翅膀
			if allSelected {

				for _, card := range cards {
					value, _ := GetCardValueAndColor(card)

					if !InByteArr(value, valueArr) {
						valueArr = append(valueArr, value)
					}

					if len(valueArr) == 4*planeLen {
						break
					}

				}

				break
			}

		}

	// 飞机带对子
	case int32(msg.CardsType_SerialTripletWithWing):

		// 所有的 重复次数超过 3次的 值数组
		repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		// 飞机长度 (被接牌 重复次数 超过 3次 的 值数组 长度)
		planeLen := len(GetRepeatedValueArr(handCards.Cards, 3))

		for i := handCards.WeightValue + 1; i <= 0xc; i++ {

			allSelected := true
			for j := 0; j < planeLen; j++ {

				value := i - byte(j)

				if InByteArr(value, repeated3ValueArr) {
					valueArr = append(valueArr, value, value, value)
				} else {
					allSelected = false
					valueArr = []byte{}
					break
				}
			}

			// 找到同长度 比 被接牌 大的 飞机, 找翅膀
			if allSelected {

				// 所有的 重复次数超过 2次的 值数组
				repeated2ValueArr := GetRepeatedValueArr(cards, 2)

				for _, value := range repeated2ValueArr {

					if !InByteArr(value, valueArr) {
						valueArr = append(valueArr, value, value)
					}

					if len(valueArr) == 5*planeLen {
						break
					}
				}

				break
			}

		}

	// 四带二单张
	case int32(msg.CardsType_QuartetWithTwo):

		// 所有的 重复次数超过 3次的 值数组
		repeated4ValueArr := GetRepeatedValueArr(cards, 4)

		for _, value := range repeated4ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value, value}
				break
			}
		}

		// 找到四同张, 找翅膀
		if len(valueArr) == 4 {

			// 两张 单张 可以相同，用一个翅膀数组来
			wingArr := []byte{}
			for _, card := range cards {
				value, _ := GetCardValueAndColor(card)

				if !InByteArr(value, valueArr) {
					wingArr = append(wingArr, value)
				}

				if len(wingArr) == 2 {
					break
				}
			}

			valueArr = append(valueArr, wingArr...)

		}

		// 没找到同长度，至空
		if len(valueArr) != handLen {
			valueArr = []byte{}
		}

	// 四带二对子
	case int32(msg.CardsType_QuartetWithTwoPair):

		// 所有的 重复次数超过 3次的 值数组
		repeated4ValueArr := GetRepeatedValueArr(cards, 4)

		for _, value := range repeated4ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value, value}
				break
			}
		}

		// 找到四同张, 找翅膀
		if len(valueArr) == 4 {

			// 所有的 重复次数超过 2次的 值数组
			repeated2ValueArr := GetRepeatedValueArr(cards, 2)

			for _, value := range repeated2ValueArr {

				if !InByteArr(value, valueArr) {
					valueArr = append(valueArr, value, value)
				}

				if len(valueArr) == 8 {
					break
				}
			}

		}

		// 没找到同长度，至空
		if len(valueArr) != handLen {
			valueArr = []byte{}
		}

	// 炸弹
	case int32(msg.CardsType_Bomb):

		// 所有的 重复次数超过 4次的 值数组
		repeated4ValueArr := GetRepeatedValueArr(cards, 4)

		for _, value := range repeated4ValueArr {
			if value > handCards.WeightValue {
				valueArr = []byte{value, value, value, value}
				break
			}
		}

	}

	// 同牌型压牌找不到，找炸弹
	if len(valueArr) == 0 {

		repeatedArr := CheckRepeatedCards(cards)

		if handCards.CardsType < int32(msg.CardsType_Bomb) && len(repeatedArr[3]) > 0 {
			value := repeatedArr[3][0]
			valueArr = []byte{value, value, value, value}
		}
	}

	// 炸弹也接不上，但是有火箭
	if len(valueArr) == 0 && HaveRocket(cards) {
		valueArr = []byte{0xe, 0xf}
	}

	// 从选定到牌值在牌组中找到对应到牌
	for _, value := range valueArr {
		for i, card := range cards {
			cardValue, _ := GetCardValueAndColor(card)
			if value == cardValue {
				takeCards = append(takeCards, card)
				cards = append(cards[:i], cards[i+1:]...)
				break
			}
		}
	}

	return
}

// HangUpPutCards 托管出牌
func HangUpPutCards(handCards HandCards, user *data.User, handIsDizhu bool) (putCards []byte) {
	actionType := msg.UserActionType_TakeOverCard
	if handCards.UserID == 0 || user.ID == handCards.UserID {
		actionType = msg.UserActionType_PutCard
	}

	switch actionType {

	// 有牌权出牌：出最小单张牌
	case msg.UserActionType_PutCard:

		putCards = []byte{GetSmallestCard(user.Cards)}
		log.Tracef("托管, 出牌, 出最小单张牌: %s", fmt.Sprintf("%+v\n", putCards))

	// 接牌, 拆牌出最小压对手牌, 盟友牌不压
	case msg.UserActionType_TakeOverCard:

		// 接牌玩家 和 被街拍玩家都是农民时，不接
		if !handIsDizhu && !user.IsDizhu {
			return
		}

		putCards = TakeOverCards(handCards, user.Cards)

		log.Tracef("托管, 接牌, 出牌: %s", fmt.Sprintf("%+v\n", putCards))

	}

	return
}

// FindTypeInCards 从手牌中找到对应牌型
func FindTypeInCards(cardsType msg.CardsType, cardsStack []byte) (solutions []SolutionCards, leftCards []byte) {

	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range cardsStack {
		cards = append(cards, card)
	}

	// 值数组列表
	var valueArrList [][]byte

	switch cardsType {

	// 单张牌
	case msg.CardsType_SingleCard:
		// 获取所有牌的值
		values := GetAllValue(cards)

		for _, v := range values {
			arr := []byte{v}
			valueArrList = append(valueArrList, arr)
		}

	// 对牌
	case msg.CardsType_Pair:

		// 所有的 重复次数超过 2次的 值数组
		Repeated2ValueArr := GetRepeatedValueArr(cards, 2)

		for _, value := range Repeated2ValueArr {
			arr := []byte{value, value}
			valueArrList = append(valueArrList, arr)
		}

	// 三同张
	case msg.CardsType_Triplet:

		// 所有的 重复次数超过 3次的 值数组
		Repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		for _, value := range Repeated3ValueArr {
			arr := []byte{value, value, value}
			valueArrList = append(valueArrList, arr)
		}

	// 顺子(先提取5连顺，然后扩展5连，最后合并顺子检查)
	case msg.CardsType_Sequence:

		// 获取所有牌的值
		values := GetAllValue(cards)

		// 正序排序
		values = PositiveSortCards(values)

		// 提取 5连顺
		for {

			var (
				firstValue byte   // 顺子第一个值
				count      byte   // 轮询递增值
				arr        []byte // 不重复的值数组
			)

			// 遍历有序的所有值，找出一个五连顺
			for _, v := range values {
				if v == firstValue+count {
					arr = append(arr, v)
					if len(arr) == 5 {
						break
					}
					count++
				} else if v > firstValue+count {
					count = 1
					firstValue = v
					arr = []byte{firstValue}
				}
			}

			if len(arr) != 5 {
				arr = []byte{}
			}

			if len(arr) == 0 {
				break
			}

			// 从所有值数组中删除已经选中的值
			for _, value := range arr {
				for i, v := range values {
					if value == v {
						values = append(values[:i], values[i+1:]...)
						break
					}
				}
			}

			valueArrList = append(valueArrList, arr)

		}

		// 扩展五连顺
		for index, arr := range valueArrList {

			for {
				var addValue byte

				for i, v := range values {
					if v == arr[len(arr)-1]+1 {
						addValue = v
						values = append(values[:i], values[i+1:]...)
						break
					}
				}

				if addValue == 0 {
					break
				}

				arr = append(arr, addValue)

				valueArrList[index] = arr

			}
		}

		// 合并顺子检查, 最多只能两个顺子相连
		for index, arr := range valueArrList {
			if index < 1 {
				continue
			}
			preArr := valueArrList[index-1]

			// 头一个顺子的最后一个 等于 下一个顺子的第一个 - 1
			if preArr[len(preArr)-1] == arr[0]-1 {
				for _, v := range arr {
					preArr = append(preArr, v)
				}

				// 把下一个顺子加入到上一个顺子里，删除一个顺子
				valueArrList[index-1] = preArr
				valueArrList = append(valueArrList[:index], valueArrList[index+1:]...)
				break
			}
		}

	// 连对
	case msg.CardsType_SerialPair:
		var (
			firstValue byte     // 连对第一个值
			count      int      // 轮询递增值
			serialArr  [][]byte // 不重复的值数组列表
			arr        []byte   // 不重复的值数组
		)

		// 所有的 重复次数超过 2次的 值数组
		Repeated2ValueArr := GetRepeatedValueArr(cards, 2)

		// 正序排序
		Repeated2ValueArr = PositiveSortCards(Repeated2ValueArr)

		for _, value := range Repeated2ValueArr {
			if value == firstValue+byte(count) {
				arr = append(arr, value)
				count++
			} else {
				if len(arr) >= 2 {
					serialArr = append(serialArr, arr)
				}

				count = 1
				firstValue = value
				arr = []byte{firstValue}
			}
		}

		// 最后出现的一个连对
		if len(arr) >= 2 {
			serialArr = append(serialArr, arr)
		}

		for _, arr := range serialArr {
			var valueArr []byte
			for _, value := range arr {
				valueArr = append(valueArr, value, value)
			}
			valueArrList = append(valueArrList, valueArr)
		}

	// 飞机
	case msg.CardsType_SerialTriplet:

		var (
			firstValue byte     // 飞机第一个值
			count      int      // 轮询递增值
			serialArr  [][]byte // 不重复的值数组列表
			arr        []byte   // 不重复的值数组
		)

		// 所有的 重复次数超过 3次的 值数组
		Repeated3ValueArr := GetRepeatedValueArr(cards, 3)

		// 正序排序
		Repeated3ValueArr = PositiveSortCards(Repeated3ValueArr)

		for _, value := range Repeated3ValueArr {
			if value == firstValue+byte(count) {
				arr = append(arr, value)
				count++
			} else {
				if len(arr) >= 2 {
					serialArr = append(serialArr, arr)
				}

				count = 1
				firstValue = value
				arr = []byte{firstValue}
			}
		}

		// 最后出现的一个三顺
		if len(arr) >= 2 {
			serialArr = append(serialArr, arr)
		}

		for _, arr := range serialArr {
			var valueArr []byte
			for _, value := range arr {
				valueArr = append(valueArr, value, value, value)
			}
			valueArrList = append(valueArrList, valueArr)
		}

	// 炸弹
	case msg.CardsType_Bomb:

		// 所有的 重复次数超过 4次的 值数组
		Repeated4ValueArr := GetRepeatedValueArr(cards, 4)

		for _, value := range Repeated4ValueArr {
			valueArrList = append(valueArrList, []byte{value, value, value, value})
		}
	}

	putScoreCfg := &config.PutScoreConf

	// 遍历值数组列表，从手牌中找到等同值的手牌，组装 手牌列表
	for _, valueArr := range valueArrList {
		var handCards []byte

		putScore := putScoreCfg.GetPutScore(int32(cardsType), valueArr[0])
		for _, value := range valueArr {

			for i, card := range cards {
				v, _ := GetCardValueAndColor(card)
				if value == v {
					handCards = append(handCards, card)
					cards = append(cards[:i], cards[i+1:]...)
					break
				}

			}

		}

		solutions = append(solutions, SolutionCards{
			Cards:     handCards,
			CardsType: cardsType,
			PutScore:  putScore,
		})

	}

	return
}

// GetBestSolutions 获取最优牌解
func GetBestSolutions(cards []byte) (bestSolutions []SolutionCards) {

	// 拆牌顺序：火箭, 炸弹, 飞机, 顺子, 连对, 三张, 对子, 单张
	typeOrder := []msg.CardsType{
		msg.CardsType_Rocket,
		msg.CardsType_Bomb,
		msg.CardsType_SerialTriplet,
		msg.CardsType_Sequence,
		msg.CardsType_SerialPair,
		msg.CardsType_Triplet,
		msg.CardsType_Pair,
		msg.CardsType_SingleCard,
	}

	var (
		solutions, bestSolution []SolutionCards // 单牌型的多牌解/最优牌解集合
		singleLen, pairLen      int             // 单张/对子 个数
	)

	for _, cardsType := range typeOrder {
		solutions, cards = FindTypeInCards(cardsType, cards)

		// 获取对子个数
		if cardsType == msg.CardsType_Pair {
			pairLen = len(solutions)
		}

		// 获取单张个数
		if cardsType == msg.CardsType_SingleCard {
			singleLen = len(solutions)
		}

		bestSolution = append(bestSolution, solutions...)
	}

	// todo 寻找带牌
	for _, v := range bestSolution {
		// 三同张个数
		var tripletCount int

		if v.CardsType == msg.CardsType_SerialTripletWithWing {
			tripletCount = len(v.Cards) / 3

			// 单张个数 >= 飞机三条个数,
			if singleLen >= tripletCount {

			}
			if tripletCount <= pairLen {

			}

		}
	}

	return
}

// SolutionFindWing 策略解寻找带牌
//func SolutionFindWing(bestSolution []SolutionCards) {
//
//	var (
//		maxTripletIndex int // 最长三同张索引
//		tripletCount    int // 三同张个数
//		singleCount     int // 单张个数
//		pairCount       int // 对子个数
//	)
//
//	// 寻找三同张最长的飞机或者三条
//	for _, v := range bestSolution {
//
//		// 跳过 非三条 和 非飞机
//		if v.CardsType != msg.CardsType_Triplet && v.CardsType != msg.CardsType_SerialTriplet {
//
//		}
//
//	}
//
//}

// TakeOverNotSolution 拆牌接牌
// @targetCards 目标手牌（被比较的手牌）
// @cardsStack 牌堆
// @takeOverCards 能接的手牌
// @leftCards 牌堆剩余牌
func TakeOverNotSolution(targetCards HandCards, cardsStack []byte) (takeOverList [][]byte) {
	var takeOverCards []byte
	for {

		// 最小拆牌接牌
		takeOverCards = TakeOverCards(targetCards, cardsStack)
		if len(takeOverCards) == 0 {
			break
		}
		cardsType := GetCardsType(takeOverCards)
		targetCards = HandCards{
			Cards:       takeOverCards,
			WeightValue: GetCardsWeightValue(takeOverCards, cardsType),
			CardsType:   int32(cardsType),
		}

		takeOverList = append(takeOverList, takeOverCards)
	}
	return
}
