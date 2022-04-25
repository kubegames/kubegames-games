package poker

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960204/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
)

// GamePoker 牌堆
type GamePoker struct {
	Cards []byte
}

//使用一副牌，去掉大小王、三个2（只留下黑桃2）和黑桃A，共48张牌 大小关系：2>A>K>…>4>3 方梅红黑
var Deck = []byte{
	0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
	0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
	0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xf4,
}

const NULL_CARD = 0x21

// HandCards 手牌
type HandCards struct {
	Cards       []byte // 手牌
	UserID      int64  // 持有这幅手牌的用户ID
	WeightValue byte   // 手牌权重值，用与比较大小
	CardsType   int32  // 牌型
}

// SolutionCards 牌解
type SolutionCards struct {
	Cards       []byte // 手牌
	CardsValue  []byte // 牌值组
	CardsType   int32  // 牌型
	WeightValue byte   // 手牌权重值，用与比较大小
	Weight      int    // 一组手牌权值
	OrderValue  int    // 牌型顺序值
}

// InitPoker 初始化牌组
func (gamePoker *GamePoker) InitPoker() {

	for _, v := range Deck {
		gamePoker.Cards = append(gamePoker.Cards, v)
	}

	// 洗牌
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(gamePoker.Cards), func(i, j int) {
		gamePoker.Cards[i], gamePoker.Cards[j] = gamePoker.Cards[j], gamePoker.Cards[i]
	})
}

// DrawCard 抽16张牌
func (gamePoker *GamePoker) DrawCard() (cards []byte) {

	length := len(gamePoker.Cards)

	cards = gamePoker.Cards[length-16:]

	gamePoker.Cards = gamePoker.Cards[:(length - 16)]

	return
}

// GetCardValueAndColor 获取一张牌的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	// 与 1111 1111 进行 & 位运算然后 右移动 4 位
	cardValue = (value & 0xff) >> 4
	// 与 1111 进行 & 位运算
	cardColor = value & 0xf
	return
}

// 将牌按照牌值进行排序
func SortCards(cards []byte) []byte {
	for i := 0; i < len(cards)-1; i++ {
		for j := 0; j < (len(cards) - 1 - i); j++ {
			value1, _ := GetCardValueAndColor(cards[j])
			value2, _ := GetCardValueAndColor(cards[j+1])

			if value1 > value2 || value1 == value2 {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
	return cards
}

// 牌值排序
func SortValue(values []byte) []byte {
	for i := 0; i < len(values)-1; i++ {
		for j := 0; j < (len(values) - 1 - i); j++ {

			if values[j] > values[j+1] {
				values[j], values[j+1] = values[j+1], values[j]
			}
		}
	}
	return values
}

// 查询重复牌个数
// repeatedArr
// index:0 重复一次(单张牌)的牌
// index:1 重复二次(对牌)的牌
// index:2 重复三次(三张)的牌
// index:3 重复四次(炸弹)的牌
func CheckRepeatedCards(cards []byte) (repeatedArr [4][]byte) {
	notRepeatedCards := []byte{}
	for i := 0; i < len(cards); i++ {
		repeatedCount := 0
		value1, _ := GetCardValueAndColor(cards[i])

		// 防止循环检测
		isRepeat := false
		for _, cardValue := range notRepeatedCards {
			if value1 == cardValue {
				isRepeat = true
			}
		}
		if isRepeat {
			continue
		} else {
			notRepeatedCards = append(notRepeatedCards, value1)
		}

		for j := 0; j < len(cards); j++ {
			value2, _ := GetCardValueAndColor(cards[j])

			if value1 == value2 {
				repeatedCount++
			}
		}

		if repeatedCount < 1 || repeatedCount > 4 {
			log.Errorf("错误的重复牌个数 %d ", repeatedCount)
			return
		}
		repeatedArr[repeatedCount-1] = append(repeatedArr[repeatedCount-1], value1)
	}
	return
}

// GetCardsType 获取牌型
func GetCardsType(cards []byte) (cardsType msg.CardsType) {

	// 是否单张
	if IsSingleCard(cards) {
		return msg.CardsType_SingleCard
	}

	// 是否是对子
	if IsPair(cards) {
		return msg.CardsType_Pair
	}

	// 是否是连对
	if IsSerialPair(cards) {
		return msg.CardsType_SerialPair
	}

	// 是否是顺子
	if IsSequence(cards) {
		return msg.CardsType_Sequence
	}

	// 是否三同张
	if IsTriplet(cards) {
		return msg.CardsType_Triplet
	}

	// 是否三带一
	if IsTripletWithSingle(cards) {
		return msg.CardsType_TripletWithSingle
	}

	// 是否是三带二
	if IsTripletWithTwo(cards) {
		return msg.CardsType_TripletWithTwo
	}

	// 是否是三顺
	if IsSerialTriplet(cards) {
		return msg.CardsType_SerialTriplet
	}

	// 是否不合法的飞机带翅膀
	if IsIncompleteSerialTripletWithTwo(cards) {
		return msg.CardsType_IncompleteSerialTripletWithTwo
	}

	// 是否是飞机带翅膀
	if IsSerialTripletWithTwo(cards) {
		return msg.CardsType_SerialTripletWithTwo
	}

	// 是否是四带三
	if IsQuartetWithThree(cards) {
		return msg.CardsType_QuartetWithThree
	}

	// 是否是炸弹
	if IsBomb(cards) {
		return msg.CardsType_Bomb
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

// IsSerialPair 是否是连对
func IsSerialPair(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count%2 != 0 || count < 4 || count > 16 {
		return false
	}

	// 牌值排序
	cards = SortCards(cards)

	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[1]) == count/2 && int(repeatedArr[1][len(repeatedArr[1])-1]-repeatedArr[1][0]) == count/2-1 {
		return true
	}

	return false
}

// IsSequence 是否是顺子
func IsSequence(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count < 5 || count > 13 {
		return false
	}

	// 牌值排序
	cards = SortCards(cards)

	value, _ := GetCardValueAndColor(cards[0])

	for _, v := range cards {

		contrastValue, _ := GetCardValueAndColor(v)
		if value != contrastValue {
			return false
		}

		// 顺子中不能有2
		if value == 0xf {
			return false
		}

		value++
	}

	return true
}

// IsTripletWithTwo 是否是三带二
func IsTripletWithTwo(cards []byte) bool {

	// 牌数检测
	if len(cards) != 5 {
		return false
	}

	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == 1 {
		return true
	}

	return false
}

// IsSerialTriplet 是否是三顺
func IsSerialTriplet(cards []byte) bool {

	// 牌数检测
	count := len(cards)
	if count%3 != 0 || count < 6 || count > 15 {
		return false
	}

	// 牌值排序
	cards = SortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == count/3 && int(repeatedArr[2][len(repeatedArr[2])-1]-repeatedArr[2][0]) == count/3-1 {
		return true
	}

	return false
}

// IsSerialTripletWithTwo 是否是飞机带翅膀
func IsSerialTripletWithTwo(cards []byte) bool {
	// 牌数检测
	count := len(cards)
	if count != 10 && count != 15 {
		return false
	}

	// 牌值排序
	cards = SortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) >= 2 && len(repeatedArr[2]) == count/5 && int(repeatedArr[2][len(repeatedArr[2])-1]-repeatedArr[2][0]) == count/5-1 {
		return true
	}

	return false
}

// IsQuartetWithThree 是否是四带三
func IsQuartetWithThree(cards []byte) bool {

	if len(cards) != 7 {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[3]) == 1 {
		return true
	}

	return false
}

// IsBomb 是否是炸弹
func IsBomb(cards []byte) bool {
	if len(cards) != 4 {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[3]) == 1 {
		return true
	}

	return false
}

// IsBomb 是否三同张
func IsTriplet(cards []byte) bool {

	if len(cards) != 3 {
		return false
	}

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	if len(repeatedArr[2]) == 1 {
		return true
	}

	return false
}

// IsBomb 是否三带一
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

// IsBomb 是否不合法的飞机带翅膀
func IsIncompleteSerialTripletWithTwo(cards []byte) bool {

	count := len(cards)

	// 牌值排序
	cards = SortCards(cards)

	// 重复的值
	repeatedArr := CheckRepeatedCards(cards)

	// 重复三张牌个数
	repeatThreeCount := len(repeatedArr[2])

	if repeatThreeCount >= 2 && count > 3*repeatThreeCount && count < 5*repeatThreeCount &&
		int(repeatedArr[2][len(repeatedArr[2])-1]-repeatedArr[2][0]) == repeatThreeCount-1 {
		return true
	}

	return false
}

// GetCardsWeightValue 获取牌组权重值
func GetCardsWeightValue(userCards []byte, cardsType msg.CardsType) (cardsWeightValue byte) {

	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range userCards {
		cards = append(cards, card)
	}

	// 牌值排序
	cards = SortCards(cards)

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

		// 连对
	case msg.CardsType_SerialPair:
		cardsWeightValue = repeatedArr[1][len(repeatedArr[1])-1]
		break

		// 顺子
	case msg.CardsType_Sequence:
		cardsWeightValue = repeatedArr[0][len(repeatedArr[0])-1]
		break

		// 三同张
	case msg.CardsType_Triplet:
		cardsWeightValue = repeatedArr[2][0]
		break

		// 三带一
	case msg.CardsType_TripletWithSingle:
		cardsWeightValue = repeatedArr[2][0]
		break

		// 三带二
	case msg.CardsType_TripletWithTwo:
		cardsWeightValue = repeatedArr[2][0]
		break

		// 三顺
	case msg.CardsType_SerialTriplet:
		cardsWeightValue = repeatedArr[2][len(repeatedArr[2])-1]
		break

		// 不合法的飞机带翅膀
	case msg.CardsType_IncompleteSerialTripletWithTwo:
		cardsWeightValue = repeatedArr[2][len(repeatedArr[2])-1]
		break

		// 飞机带翅膀
	case msg.CardsType_SerialTripletWithTwo:
		cardsWeightValue = repeatedArr[2][len(repeatedArr[2])-1]
		break

		// 四带三
	case msg.CardsType_QuartetWithThree:
		cardsWeightValue = repeatedArr[3][0]
		break

		// 炸弹
	case msg.CardsType_Bomb:
		cardsWeightValue = repeatedArr[3][0]
		break
	}

	if cardsWeightValue == 0 {
		log.Errorf("出现了为0权重，牌型或者牌组不正确")
	}

	return
}

// CheckTakeOverCards 拆牌接牌检测
func CheckTakeOverCards(handCards HandCards, userCards []byte, isNextSingle bool) (takeCards []byte) {

	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range userCards {
		cards = append(cards, card)
	}

	// 牌值排序
	cards = SortCards(cards)
	repeatedArr := CheckRepeatedCards(cards)

	var valueArr []byte

	// 相同牌型接管判断
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
		lastCardValue, _ := GetCardValueAndColor(cards[len(cards)-1])
		if isNextSingle && lastCardValue > handCards.WeightValue {
			valueArr = []byte{lastCardValue}
			break
		}

		break

		// 对子
	case int32(msg.CardsType_Pair):
		for key, arr := range repeatedArr {
			if key == 0 {
				continue
			}
			for _, value := range arr {
				if value > handCards.WeightValue {
					valueArr = append(valueArr, value, value)
					break
				}
			}
			if len(valueArr) != 0 {
				break
			}
		}
		break

		// 连对
	case int32(msg.CardsType_SerialPair):
		for i := handCards.WeightValue + 1; i <= 0xe; i++ {
			isSelected := true
			for j := 0; j < len(handCards.Cards)/2; j++ {
				hasValue := false
				// 判断 i - j 是否存在
				for key, arr := range repeatedArr {

					// 只能找2同张以上
					if key <= 0 {
						continue
					}
					for _, value := range arr {
						if i-byte(j) == value {
							hasValue = true
							break
						}
					}
					if hasValue {
						break
					}
				}

				// 未找到目标值
				if !hasValue {
					isSelected = false
					break
				}
			}

			// 找到了最小牌组
			if isSelected {
				for j := 0; j < len(handCards.Cards)/2; j++ {
					valueArr = append(valueArr, i-byte(j), i-byte(j))
				}
				break
			}

		}
		break

		// 顺子
	case int32(msg.CardsType_Sequence):
		for i := handCards.WeightValue + 1; i <= 0xe; i++ {
			isSelected := true
			for j := 0; j < len(handCards.Cards); j++ {
				hasValue := false
				// 判断 i - j 是否存在
				for _, arr := range repeatedArr {

					for _, value := range arr {
						if i-byte(j) == value {
							hasValue = true
							break
						}
					}
					if hasValue {
						break
					}
				}

				// 未找到目标值
				if !hasValue {
					isSelected = false
					break
				}
			}

			// 找到了最小牌组
			if isSelected {
				for j := 0; j < len(handCards.Cards); j++ {
					valueArr = append(valueArr, i-byte(j))
				}
				break
			}

		}
		break

		// 三带二
	case int32(msg.CardsType_TripletWithTwo):
		for key, arr := range repeatedArr {

			// 只考虑3同张以上
			if key <= 1 {
				continue
			}
			for _, value := range arr {
				if value > handCards.WeightValue {
					valueArr = append(valueArr, value, value, value)

					for _, card := range cards {
						cardValue, _ := GetCardValueAndColor(card)
						if cardValue != value {
							valueArr = append(valueArr, cardValue)
						}
						if len(valueArr) == 5 {
							break
						}
					}
					break
				}
			}

			if len(valueArr) != 0 {
				break
			}
		}

		//if len(valueArr) < 5 {
		//	valueArr = []byte{}
		//}

		break

		// 三顺
	case int32(msg.CardsType_SerialTriplet):
		for i := handCards.WeightValue + 1; i <= 0xe; i++ {
			isSelected := true
			for j := 0; j < len(handCards.Cards)/3; j++ {
				hasValue := false
				// 判断 i - j 是否存在
				for key, arr := range repeatedArr {

					// 只能找3同张以上
					if key <= 1 {
						continue
					}

					for _, value := range arr {
						if i-byte(j) == value {
							hasValue = true
							break
						}
					}
					if hasValue {
						break
					}
				}

				// 未找到目标值
				if !hasValue {
					isSelected = false
					break
				}
			}

			// 找到了最小牌组
			if isSelected {
				for j := 0; j < len(handCards.Cards)/3; j++ {
					valueArr = append(valueArr, i-byte(j), i-byte(j), i-byte(j))

				}
				break
			}

		}

		if len(valueArr) < len(handCards.Cards) {
			valueArr = []byte{}
		}

		break

		// 飞机带翅膀
	case int32(msg.CardsType_SerialTripletWithTwo):
		// 被比较牌的重复数组
		handRepeatArr := CheckRepeatedCards(handCards.Cards)

		for i := handCards.WeightValue + 1; i <= 0xe; i++ {
			isSelected := true
			for j := 0; j < len(handRepeatArr[2]); j++ {
				hasValue := false
				// 判断 i - j 是否存在
				for key, arr := range repeatedArr {

					// 只能找3同张以上
					if key <= 1 {
						continue
					}

					for _, value := range arr {
						if i-byte(j) == value {
							hasValue = true
							break
						}
					}
					if hasValue {
						break
					}
				}

				// 未找到目标值
				if !hasValue {
					isSelected = false
					break
				}
			}

			// 找到了最小牌组
			if isSelected {

				// 组成飞机
				for j := 0; j < len(handRepeatArr[2]); j++ {
					valueArr = append(valueArr, i-byte(j), i-byte(j), i-byte(j))
				}

				// 翅膀数组
				wingArr := []byte{}

				for _, card := range cards {
					cardValue, _ := GetCardValueAndColor(card)

					allUnEqual := true
					for _, value := range valueArr {
						if cardValue == value {
							allUnEqual = false
						}

					}
					if allUnEqual {
						wingArr = append(wingArr, cardValue)
					}

					if len(wingArr) == len(valueArr)/3*2 {
						break
					}
				}

				if len(wingArr) < len(valueArr)/3*2 {
					valueArr = []byte{}
				} else {
					// 飞机 + 翅膀
					valueArr = append(valueArr, wingArr...)
				}

				break
			}

		}

		break

		// 四带三
	case int32(msg.CardsType_QuartetWithThree):
		for _, value := range repeatedArr[3] {
			if value > handCards.WeightValue {
				valueArr = append(valueArr, value, value, value, value)

				for _, card := range cards {
					cardValue, _ := GetCardValueAndColor(card)
					if cardValue != value {
						valueArr = append(valueArr, cardValue)
					}
					if len(valueArr) == 7 {
						break
					}
				}
				break
			}
		}

		if len(valueArr) < 7 {
			valueArr = []byte{}
		}
		break

		// 炸弹
	case int32(msg.CardsType_Bomb):
		for _, value := range repeatedArr[3] {
			if value > handCards.WeightValue {
				valueArr = append(valueArr, value, value, value, value)
				break
			}
		}
		break
	}

	// 同牌型未找到接管牌，上家不是炸弹时，寻找是否可以用炸弹来接管
	if len(valueArr) == 0 && handCards.CardsType != int32(msg.CardsType_Bomb) && len(repeatedArr[3]) != 0 {
		value := repeatedArr[3][0]
		valueArr = append(valueArr, value, value, value, value)
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

// CheckPutSingleCard 出单张牌检测
func CheckPutSingleCard(cards []byte, isNextSingle bool) (putCards []byte) {
	// 牌值排序
	cards = SortCards(cards)

	if isNextSingle {

		// 下家报单
		putCards = append(putCards, cards[len(cards)-1])

	} else {

		putCards = append(putCards, cards[0])
	}

	return
}

// 从牌堆中找牌型
func GetCardsTypeFromStack(cardsType msg.CardsType, cardsStack []byte) (resultCards []SolutionCards, leftCards []byte) {
	// 牌型顺序表配置
	cardsOrderCfg := &config.CardsOrderConf

	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range cardsStack {
		cards = append(cards, card)
	}

	// 牌值排序
	cards = SortCards(cards)
	repeatedArr := CheckRepeatedCards(cards)

	//var resultCards []SolutionCards

	switch cardsType {
	// 炸弹
	case msg.CardsType_Bomb:

		for _, value := range repeatedArr[3] {
			weight := 0

			switch value {
			case 0x3:
				weight = 20
			case 0x4:
				weight = 30
			case 0x5:
				weight = 40
			case 0x6:
				weight = 50
			case 0x7:
				weight = 60
			case 0x8:
				weight = 70
			case 0x9:
				weight = 80
			case 0xa:
				weight = 90
			case 0xb:
				weight = 100
			case 0xc:
				weight = 110
			case 0xd:
				weight = 120
			}

			resultCards = append(resultCards, SolutionCards{
				CardsValue:  []byte{value, value, value, value},
				CardsType:   int32(msg.CardsType_Bomb),
				WeightValue: value,
				Weight:      weight,
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_Bomb), value, 4),
			})
		}
		break

	// 飞机带翅膀
	case msg.CardsType_SerialTripletWithTwo:

		var (
			firstValue byte
			count      int
			serialArr  [][]byte
			arr        []byte
		)

		for _, value := range repeatedArr[2] {
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

		// 出现最后一个三顺
		if len(arr) >= 2 {
			serialArr = append(serialArr, arr)
		}

		// 获取重复牌小于二，并且小于A的数组
		var lessTwoArr []byte
		for _, value := range repeatedArr[0] {
			if value < 0xe {
				lessTwoArr = append(lessTwoArr, value)
			}
		}
		for _, value := range repeatedArr[1] {
			if value < 0xe {
				lessTwoArr = append(lessTwoArr, value, value)
			}
		}

		// 值排序
		lessTwoArr = SortValue(lessTwoArr)

		// 倒叙排列serialArr
		for i := len(serialArr) - 1; i >= 0; i-- {
			// 寻找翅膀
			var wingArr []byte

			arr := serialArr[i]

			if len(lessTwoArr) >= 2*len(arr) {
				wingArr = append(wingArr, lessTwoArr[:2*len(arr)]...)
				lessTwoArr = lessTwoArr[2*len(arr):]
			} else {
				// 没有多余的单张或者对子来凑成飞机带翅膀
				break
			}

			// 求权值
			weight := 0
			if arr[0] >= 0x3 && arr[0] <= 0x6 {
				weight = 6
			}

			if arr[0] >= 0x7 && arr[0] <= 0xa {
				weight = 8
			}

			if arr[0] >= 0xb && arr[0] <= 0xd {
				weight = 10
			}

			// 每增加一个相连的数字权值加1
			weight = weight + len(arr) - 2

			var cardsValue []byte
			for _, value := range arr {
				cardsValue = append(cardsValue, value, value, value)
			}
			cardsValue = append(cardsValue, wingArr...)
			resultCards = append(resultCards, SolutionCards{
				CardsValue:  cardsValue,
				Weight:      weight,
				CardsType:   int32(msg.CardsType_SerialTripletWithTwo),
				WeightValue: arr[len(arr)-1],
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_SerialTripletWithTwo), arr[len(arr)-1], len(arr)*5),
			})
		}
		break

	// 三顺
	case msg.CardsType_SerialTriplet:
		var (
			firstValue byte
			count      int
			serialArr  [][]byte
			arr        []byte
		)

		for _, value := range repeatedArr[2] {
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

		// 出现最后一个三顺
		if len(arr) >= 2 {
			serialArr = append(serialArr, arr)
		}

		for _, arr := range serialArr {

			// 求权值
			weight := 0
			if arr[0] >= 0x3 && arr[0] <= 0x6 {
				weight = 5
			}

			if arr[0] >= 0x7 && arr[0] <= 0xa {
				weight = 7
			}

			if arr[0] >= 0xb && arr[0] <= 0xd {
				weight = 9
			}

			// 每增加一个相连的数字权值加1
			weight = weight + len(arr) - 2

			var cardsValue []byte
			for _, value := range arr {
				cardsValue = append(cardsValue, value, value, value)
			}

			resultCards = append(resultCards, SolutionCards{
				CardsValue:  cardsValue,
				Weight:      weight,
				CardsType:   int32(msg.CardsType_SerialTriplet),
				WeightValue: arr[len(arr)-1],
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_SerialTriplet), arr[len(arr)-1], len(arr)*3),
			})
		}
		break

	// 三带二
	case msg.CardsType_TripletWithTwo:
		// 获取重复牌数量小于等于二，并且小于A的数组
		var lessTwoArr []byte
		for _, value := range repeatedArr[0] {
			if value < 0xe {
				lessTwoArr = append(lessTwoArr, value)
			}
		}
		for _, value := range repeatedArr[1] {
			if value < 0xe {
				lessTwoArr = append(lessTwoArr, value, value)
			}
		}

		// 值排序
		lessTwoArr = SortValue(lessTwoArr)

		for _, repeatedValue := range repeatedArr[2] {
			// 寻找翅膀
			var wingArr []byte

			if len(lessTwoArr) >= 2 {
				wingArr = append(wingArr, lessTwoArr[:2]...)
				lessTwoArr = lessTwoArr[2:]
			} else {
				// 没有多余的单张或者对子来凑成飞机带翅膀
				break
			}

			// 求权值
			weight := 0
			if repeatedValue >= 0x3 && repeatedValue <= 0x6 {
				weight = 5
			}

			if repeatedValue >= 0x7 && repeatedValue <= 0xa {
				weight = 7
			}

			if repeatedValue >= 0xb && repeatedValue <= 0xe {
				weight = 9
			}

			var cardsValue []byte
			cardsValue = append(cardsValue, repeatedValue, repeatedValue, repeatedValue)

			cardsValue = append(cardsValue, wingArr...)
			resultCards = append(resultCards, SolutionCards{
				CardsValue:  cardsValue,
				Weight:      weight,
				CardsType:   int32(msg.CardsType_TripletWithTwo),
				WeightValue: repeatedValue,
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_TripletWithTwo), repeatedValue, 5),
			})
		}
		break

	// 连对
	case msg.CardsType_SerialPair:
		var (
			firstValue byte
			count      int
			serialArr  [][]byte
			arr        []byte
			valueArr   []byte
		)

		// 获取所有牌值
		for i, array := range repeatedArr {
			if i == 3 || i == 0 {
				continue
			}
			for _, value := range array {
				valueArr = append(valueArr, value)
			}
		}

		// 牌值排序
		sortValue := SortValue(valueArr)

		// 提取联对
		for _, value := range sortValue {
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

		// 出现最后一个连队
		if len(arr) >= 2 {
			serialArr = append(serialArr, arr)
		}

		for _, arr := range serialArr {

			// 求权值
			weight := 0
			if arr[0] >= 0x3 && arr[0] <= 0x5 {
				weight = 5
			}

			if arr[0] >= 0x6 && arr[0] <= 0x9 {
				weight = 6
			}

			if arr[0] >= 0xa && arr[0] <= 0xd {
				weight = 8
			}

			// 每增加一个相连的数字权值加1
			weight = weight + len(arr) - 2

			var cardsValue []byte
			for _, value := range arr {
				cardsValue = append(cardsValue, value, value)
			}

			resultCards = append(resultCards, SolutionCards{
				CardsValue:  cardsValue,
				Weight:      weight,
				CardsType:   int32(msg.CardsType_SerialPair),
				WeightValue: arr[len(arr)-1],
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_SerialPair), arr[len(arr)-1], len(arr)*2),
			})
		}
		break

	// 顺子
	case msg.CardsType_Sequence:
		var (
			firstValue byte
			count      int
			serialArr  [][]byte
			arr        []byte
			valueArr   []byte
		)

		// 获取所有牌值
		for i, array := range repeatedArr {
			if i == 3 {
				continue
			}
			for _, value := range array {
				valueArr = append(valueArr, value)
			}
		}

		// 牌值排序
		sortValue := SortValue(valueArr)

		for _, value := range sortValue {
			// 顺子不能连2
			if value == firstValue+byte(count) && value != 0xf {
				arr = append(arr, value)
				count++
			} else {
				if len(arr) >= 5 {
					serialArr = append(serialArr, arr)
				}

				count = 1
				firstValue = value
				arr = []byte{firstValue}
			}
		}

		// 出现最后一个顺子
		if len(arr) >= 5 {
			serialArr = append(serialArr, arr)
		}

		for _, arr := range serialArr {

			// 求权值
			weight := 0
			if arr[0] >= 0x3 && arr[0] <= 0x5 {
				weight = 5
			}

			if arr[0] >= 0x6 && arr[0] <= 0x8 {
				weight = 7
			}

			if arr[0] >= 0x9 && arr[0] <= 0xa {
				weight = 9
			}

			// 每增加一个相连的数字权值加1
			weight = weight + len(arr) - 2

			var cardsValue []byte
			for _, value := range arr {
				cardsValue = append(cardsValue, value)
			}

			resultCards = append(resultCards, SolutionCards{
				CardsValue:  cardsValue,
				Weight:      weight,
				CardsType:   int32(msg.CardsType_Sequence),
				WeightValue: arr[len(arr)-1],
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_Sequence), arr[len(arr)-1], len(arr)),
			})
		}
		break

	// 对子
	case msg.CardsType_Pair:
		for count, values := range repeatedArr {
			// 跳过单张 和 重复4张
			if count == 0 || count == 3 {
				continue
			}

			for _, value := range values {
				// 求权值
				weight := 0
				if value >= 0x3 && value <= 0x7 {
					weight = 1
				}

				if value >= 0x8 && value <= 0xa {
					weight = 2
				}

				if value >= 0xb && value <= 0xc {
					weight = 3
				}

				if value >= 0xd && value <= 0xe {
					weight = 4
				}

				var cardsValue []byte
				cardsValue = append(cardsValue, value, value)

				resultCards = append(resultCards, SolutionCards{
					CardsValue:  cardsValue,
					Weight:      weight,
					CardsType:   int32(msg.CardsType_Pair),
					WeightValue: value,
					OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_Pair), value, 2),
				})
			}
		}

		break

	// 单张
	case msg.CardsType_SingleCard:

		for _, card := range cards {

			// 取牌值
			value, _ := GetCardValueAndColor(card)

			// 求权值
			weight := 0
			if value >= 0x3 && value <= 0xa {
				weight = 1
			}

			if value >= 0xb && value <= 0xd {
				weight = 2
			}

			if value >= 0xe && value <= 0xf {
				weight = 4
			}

			resultCards = append(resultCards, SolutionCards{
				CardsValue:  []byte{value},
				Weight:      weight,
				CardsType:   int32(msg.CardsType_SingleCard),
				WeightValue: value,
				OrderValue:  cardsOrderCfg.GetCardsOrderValue(int32(msg.CardsType_SingleCard), value, 1),
			})
		}
		break

	}

	// 剔除选中的牌
	for index, cardsList := range resultCards {
		newCards := []byte{}
		for _, value := range cardsList.CardsValue {
			for i, card := range cards {
				cardValue, _ := GetCardValueAndColor(card)
				if value == cardValue {
					newCards = append(newCards, card)
					cards = append(cards[:i], cards[i+1:]...)
					break
				}
			}
		}
		resultCards[index].Cards = newCards
	}
	leftCards = cards

	return
}

// GetOptimalSolutionCards 获取最优解牌型
func GetOptimalSolutionCards(holdCards []byte) []SolutionCards {
	// copy 一份手牌，防止修改外部手牌数据
	cards := []byte{}
	for _, card := range holdCards {
		cards = append(cards, card)
	}

	// 先把炸弹提取出来
	boomResultCards, leftCards := GetCardsTypeFromStack(msg.CardsType_Bomb, cards)

	// 再用穷举法获取所有拿牌序列
	arr := []msg.CardsType{
		msg.CardsType_SerialTripletWithTwo,
		msg.CardsType_SerialTriplet,
		msg.CardsType_TripletWithTwo,
		msg.CardsType_SerialPair,
		msg.CardsType_Sequence,
	}
	arrList := new(ArrList)
	arrList.getAll(arr, 0, len(arr))

	allSolution := [][]SolutionCards{}
	for _, arr := range arrList.list {

		// 一种获牌型排序
		onecSolution := []SolutionCards{}

		var resultCards []SolutionCards

		// 把炸弹牌添进去
		onecSolution = append(onecSolution, boomResultCards...)

		leftCardsWithoutBoom := leftCards

		for _, cardsType := range arr {

			resultCards, leftCardsWithoutBoom = GetCardsTypeFromStack(cardsType, leftCardsWithoutBoom)
			onecSolution = append(onecSolution, resultCards...)

		}
		// 剩余对子
		resultCards, leftCardsWithoutBoom = GetCardsTypeFromStack(msg.CardsType_Pair, leftCardsWithoutBoom)
		onecSolution = append(onecSolution, resultCards...)

		// 剩余单张
		resultCards, leftCardsWithoutBoom = GetCardsTypeFromStack(msg.CardsType_SingleCard, leftCardsWithoutBoom)
		onecSolution = append(onecSolution, resultCards...)
		allSolution = append(allSolution, onecSolution)
	}

	bestSolution := allSolution[0]

	for _, OnceSolution := range allSolution {

		// 手数最少的为最优解
		if len(OnceSolution) < len(bestSolution) {
			bestSolution = OnceSolution
		}

		// 如果最少的手数有多个，则找其中单张最少的方案
		if len(OnceSolution) == len(bestSolution) {
			var bestSingleCount, OnceSingleCount int
			for _, v := range OnceSolution {
				if v.CardsType == int32(msg.CardsType_SingleCard) {
					OnceSingleCount++
				}
			}

			for _, v := range bestSolution {
				if v.CardsType == int32(msg.CardsType_SingleCard) {
					bestSingleCount++
				}
			}

			if OnceSingleCount < bestSingleCount {
				bestSolution = OnceSolution
			}
		}
	}

	return bestSolution
}

// CardsToString 牌组转字符串
func CardsToString(cards []byte) (cardsStr string) {
	for key, card := range cards {

		value, _ := GetCardValueAndColor(card)

		switch value {
		case 0x3:
			cardsStr += "3"
			break
		case 0x4:
			cardsStr += "4"
			break
		case 0x5:
			cardsStr += "5"
			break
		case 0x6:
			cardsStr += "6"
			break
		case 0x7:
			cardsStr += "7"
			break
		case 0x8:
			cardsStr += "8"
			break
		case 0x9:
			cardsStr += "9"
			break
		case 0xa:
			cardsStr += "10"
			break
		case 0xb:
			cardsStr += "J"
			break
		case 0xc:
			cardsStr += "Q"
			break
		case 0xd:
			cardsStr += "K"
			break
		case 0xe:
			cardsStr += "A"
			break
		case 0xf:
			cardsStr += "2"
			break
		}

		if key == len(cards)-1 {
			continue
		}
		cardsStr += "/"
	}

	return
}

// TransformCards 转译牌
func TransformCards(cards []byte) (transCards []string) {

	for _, v := range cards {
		var decodeValue, decodeColor string

		value, color := GetCardValueAndColor(v)

		switch value {
		case 0x1:
			decodeValue = "A"
			break
		case 0x2:
			decodeValue = "2"
			break
		case 0x3:
			decodeValue = "3"
			break
		case 0x4:
			decodeValue = "4"
			break
		case 0x5:
			decodeValue = "5"
			break
		case 0x6:
			decodeValue = "6"
			break
		case 0x7:
			decodeValue = "7"
			break
		case 0x8:
			decodeValue = "8"
			break
		case 0x9:
			decodeValue = "9"
			break
		case 0xa:
			decodeValue = "10"
			break
		case 0xb:
			decodeValue = "J"
			break
		case 0xc:
			decodeValue = "Q"
			break
		case 0xd:
			decodeValue = "K"
			break
		case 0xe:
			decodeValue = "A"
			break
		case 0xf:
			decodeValue = "2"
			break
		}

		switch color {
		case 0x1:
			decodeColor = "♦️"
			break
		case 0x2:
			decodeColor = "♣️"
			break
		case 0x3:
			decodeColor = "♥️"
			break
		case 0x4:
			decodeColor = "♠️"
			break
		}

		transCards = append(transCards, decodeColor+decodeValue)
	}

	return
}

type ArrList struct {
	list [][]msg.CardsType
}

func (arrlist *ArrList) getAll(arr []msg.CardsType, m int, n int) {
	if m == n {

		list := []msg.CardsType{}
		for _, cardsType := range arr {
			list = append(list, cardsType)
		}
		arrlist.list = append(arrlist.list, list)

	} else {

		for i := m; i < n; i++ {

			tamp := arr[m]
			arr[m] = arr[i]
			arr[i] = tamp

			arrlist.getAll(arr, m+1, n)
			tamp = arr[m]
			arr[m] = arr[i]
			arr[i] = tamp
		}
	}

}
