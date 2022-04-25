package poker

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

const (

	// Tcard 牌值10
	Tcard byte = 0xa

	// Acard 牌值A
	Acard byte = 0xe

	// Point21 牌值21点
	Point21 int32 = 21

	// Point17 牌值17点
	Point17 int32 = 17
)

// Deck 基础牌，方梅红黑
var Deck = []byte{
	0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
	0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
	0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
}

// GamePoker 牌堆
type GamePoker struct {
	Cards []byte
}

// InitPoker 初始化牌组
func (poker *GamePoker) InitPoker() {
	//总共使用8副牌，去除8副牌的大小王，共计416张牌
	for i := 0; i < 8; i++ {
		for _, v := range Deck {
			poker.Cards = append(poker.Cards, v)
		}
	}

	// 洗牌
	rand.Seed(time.Now().UnixNano())

	rand.Shuffle(len(poker.Cards), func(i, j int) {
		poker.Cards[i], poker.Cards[j] = poker.Cards[j], poker.Cards[i]
	})
}

// DrawCard 抽牌
func (poker *GamePoker) DrawCard() (card byte) {

	length := len(poker.Cards)

	card = poker.Cards[length-1]
	poker.Cards = poker.Cards[:(length - 1)]

	return
}

// DrawTwoCards 抽取两张牌
func (poker *GamePoker) DrawTwoCards() (cards []byte) {
	c0 := poker.DrawCard()
	c1 := poker.DrawCard()
	cards = append(cards, c0, c1)
	return

}

// CheatCheck 作弊检测
func (poker *GamePoker) CheatCheck(cheatType int32, cards []byte) {
	if cheatType == int32(msg.TestCardsType_TestNoType) {
		return
	}
	log.Debugf("配牌 %d", cheatType)

	length := len(cards)

	// 最后一张牌
	lastCard := poker.Cards[len(poker.Cards)-1]

	// 最后一张牌
	//lastCardValue, _ := GetCardValueAndColor(lastCard)

	switch cheatType {

	// 黑杰克
	case int32(msg.TestCardsType_TestBlackJack):
		if length != 0 {
			return
		}

		// 插入一张A牌
		poker.PlugSelectedCard(Acard)

		// 插入值为10的牌
		poker.PlugRangeCard(0xa, 0xd)

		break
		// 对子
	case int32(msg.TestCardsType_TestPairs):
		if length != 0 {
			return
		}

		// 获取最后一张牌值
		cardValue, _ := GetCardValueAndColor(poker.Cards[len(poker.Cards)-1])

		// 插入一张相同牌值的牌
		poker.PlugSelectedCard(cardValue)

		break
		// 3张21点
	case int32(msg.TestCardsType_TestThreeCards21Point):
		if length == 0 {

			// 获取这张牌最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint([]byte{lastCard}))

			var limitMin byte = 0x2
			var limitMax byte = 0xe

			// 最后一张牌值和塞入牌值相加 <= 20 && >= 10
			if clearPoint < 8 {
				limitMin = byte(10 - clearPoint)
			}

			if clearPoint == 10 {
				limitMax = 0xd
			}

			if clearPoint == 11 {
				limitMax = 0x9
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return

		}

		if length == 2 {

			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			var selectedValue byte

			if clearPoint == 11 {
				// 抽一张牌值为10的牌
				poker.PlugRangeCard(0xa, 0xd)

				return
			}

			if clearPoint == 20 || clearPoint == 10 {
				selectedValue = 0xe
			}

			if clearPoint > 11 && clearPoint <= 19 {
				selectedValue = byte(21 - clearPoint)
			}

			poker.PlugSelectedCard(selectedValue)
		}
		break

		// 4张21点
	case int32(msg.TestCardsType_TestFourCards21Point):
		var limitMin byte = 0x2
		var limitMax byte = 0xe

		if length == 0 {

			// 获取这张牌最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint([]byte{lastCard}))

			// 最后一张牌值和塞入牌值相加 <= 19
			if clearPoint == 9 {
				limitMax = 0xd
			}

			if clearPoint > 9 {
				limitMax = byte(19 - clearPoint)
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return
		}

		if length == 2 {
			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			// 最后一张牌值和塞入牌值相加 <= 20 && >= 10
			if clearPoint < 8 {
				limitMin = byte(10 - clearPoint)
			}

			if clearPoint == 10 {
				limitMax = 0xd
			}

			if clearPoint > 10 && clearPoint < 19 {
				limitMax = byte(20 - clearPoint)
			}

			if clearPoint == 19 {
				poker.PlugSelectedCard(0xe)
				return
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return
		}

		if length == 3 {
			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			var selectedValue byte

			if clearPoint == 11 {
				// 抽一张牌值为10的牌
				poker.PlugRangeCard(0xa, 0xd)

				return
			}

			if clearPoint == 20 || clearPoint == 10 {
				selectedValue = 0xe
			}

			if clearPoint > 11 && clearPoint <= 19 {
				selectedValue = byte(21 - clearPoint)
			}

			poker.PlugSelectedCard(selectedValue)
		}
		break

		// 五小龙
	case int32(msg.TestCardsType_TestFiveDragon):
		var limitMin byte = 0x2
		var limitMax byte = 0xe

		if length == 0 {
			// 获取这张牌最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint([]byte{lastCard}))

			// 最后一张牌值和塞入牌值相加 <= 18
			if clearPoint == 8 {
				limitMax = 0xd
			}

			if clearPoint > 8 {
				limitMax = byte(18 - clearPoint)
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return
		}

		if length == 2 {

			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			// 最后一张牌值和塞入牌值相加 <= 19
			if clearPoint == 9 {
				limitMax = 0xd
			}

			if clearPoint > 9 && clearPoint < 18 {
				limitMax = byte(19 - clearPoint)
			}

			if clearPoint == 18 {
				poker.PlugSelectedCard(0xe)
				return
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return
		}

		if length == 3 {
			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			// 最后一张牌值和塞入牌值相加 <= 20
			if clearPoint == 10 {
				limitMax = 0xd
			}

			if clearPoint > 10 && clearPoint < 19 {
				limitMax = byte(20 - clearPoint)
			}

			if clearPoint == 19 {
				poker.PlugSelectedCard(0xe)
				return
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
			return
		}

		if length == 4 {
			// 获取最靠近21点的值
			clearPoint := GetNearPoint21(GetPoint(cards))

			// 最后一张牌值和塞入牌值相加 <= 21
			if clearPoint == 11 {
				limitMax = 0xd
			}

			if clearPoint > 11 && clearPoint < 20 {
				limitMax = byte(21 - clearPoint)
			}

			if clearPoint == 20 {
				poker.PlugSelectedCard(0xe)
				return
			}

			// 在范围内抽一张牌塞入牌尾
			poker.PlugRangeCard(limitMin, limitMax)
		}
		break

	}
}

// PlugSelectedCard 塞入一张选定牌牌
func (poker *GamePoker) PlugSelectedCard(cardValue byte) {
	for k, v := range poker.Cards {
		value, _ := GetCardValueAndColor(v)

		if value == cardValue {
			poker.Cards = append(poker.Cards[:k], poker.Cards[k+1:]...)
			poker.Cards = append(poker.Cards, v)
			break
		}

	}
}

// PlugRangeCard 塞入一张范围内的牌
func (poker *GamePoker) PlugRangeCard(limitMin, limitMax byte) {
	for k, v := range poker.Cards {
		value, _ := GetCardValueAndColor(v)

		if value >= limitMin && value <= limitMax {
			poker.Cards = append(poker.Cards[:k], poker.Cards[k+1:]...)
			poker.Cards = append(poker.Cards, v)
			break
		}

	}
}

// PlugUnRangeCard 塞入一张范围外的牌
func (poker *GamePoker) PlugUnRangeCard(limitMin, limitMax byte) {
	for k, v := range poker.Cards {
		value, _ := GetCardValueAndColor(v)

		if value > limitMax || value < limitMin {
			poker.Cards = append(poker.Cards[:k], poker.Cards[k+1:]...)
			poker.Cards = append(poker.Cards, v)
			break
		}

	}
}

// GetCardValueAndColor 获取一张牌的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	// 与 1111 1111 进行 & 位运算然后 右移动 4 位
	cardValue = (value & 0xff) >> 4
	// 与 1111 进行 & 位运算
	cardColor = value & 0xf
	return
}

// GetPoint 获取牌点
func GetPoint(cards []byte) (point []int32) {
	var (
		aNumber     int
		normalPoint int32
	)

	// 算出A有几张; 得到除A牌点其他牌点数
	for _, v := range cards {
		value, _ := GetCardValueAndColor(v)

		if value == Acard {
			aNumber++
		} else {
			if value >= Tcard {
				value = 10
			}
			normalPoint += int32(value)
		}

	}

	// 点数出现情况为 aNumber + 1 种
	for i := 0; i < aNumber+1; i++ {
		totalPoint := normalPoint + int32(i*10+aNumber)
		point = append(point, totalPoint)
	}

	return
}

// GetCardsType 获取牌的类型
func GetCardsType(Cards []byte) (Type msg.CardsType) {
	if IsBlackJack(Cards) {
		Type = msg.CardsType_BlackJack
		return
	}

	if IsFiveDragon(Cards) {
		Type = msg.CardsType_FiveDragon
		return
	}

	if IsTwentyOnePoint(Cards) {
		Type = msg.CardsType_Point21
		return
	}

	if IsOther(Cards) {
		Type = msg.CardsType_Other
		return
	}

	if IsBust(Cards) {
		Type = msg.CardsType_Bust
		return
	}

	return msg.CardsType_Bust
}

// IsBlackJack 是否是黑杰克
func IsBlackJack(Cards []byte) bool {

	// 点数
	points := GetPoint(Cards)

	// 靠近21点
	clearPoint := GetNearPoint21(points)

	if len(Cards) == 2 && clearPoint == Point21 {
		return true
	}
	return false
}

// IsFiveDragon 是否是五小龙
func IsFiveDragon(Cards []byte) bool {

	// 点数
	points := GetPoint(Cards)

	// 靠近21点
	clearPoint := GetNearPoint21(points)

	if len(Cards) == 5 && clearPoint <= Point21 {
		return true
	}
	return false
}

// IsTwentyOnePoint 是否是二十一点
func IsTwentyOnePoint(Cards []byte) bool {

	// 点数
	points := GetPoint(Cards)

	// 靠近21点
	clearPoint := GetNearPoint21(points)

	if len(Cards) < 5 && len(Cards) > 2 && clearPoint == Point21 {

		return true
	}
	return false
}

// IsOther 是否是普通牌
func IsOther(Cards []byte) bool {

	// 点数
	points := GetPoint(Cards)

	// 靠近21点
	clearPoint := GetNearPoint21(points)

	if len(Cards) < 5 && clearPoint < Point21 {
		return true
	}

	return false
}

// IsBust 是否爆牌
func IsBust(Cards []byte) bool {

	// 点数
	points := GetPoint(Cards)

	// 靠近21点
	clearPoint := GetNearPoint21(points)

	if clearPoint > Point21 {
		return true
	}

	return false
}

// IsPair 判断是不是对子
func IsPair(Cards []byte) (card byte, isPair bool) {
	if len(Cards) == 2 {
		vauleOne, _ := GetCardValueAndColor(Cards[0])
		vauleTwo, _ := GetCardValueAndColor(Cards[1])
		if vauleOne == vauleTwo {
			card = vauleOne
			isPair = true
			return
		}
	}
	return
}

// DecodeCards 解析牌
func DecodeCards(cards []byte) (decodeCards []string) {

	for _, v := range cards {
		var decodeValue string

		value, _ := GetCardValueAndColor(v)
		if value <= Tcard {

		}

		switch value {
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
		}

		decodeCards = append(decodeCards, decodeValue)
	}

	return
}

// ReducePoint 精简点数
func ReducePoint(point []int32) (newPoint []int32) {
	for k, v := range point {
		if k == 0 || v <= Point21 {
			newPoint = append(newPoint, v)
		}
	}
	return
}

// GetNearPoint21 获取靠近21点的牌值
func GetNearPoint21(point []int32) (value int32) {

	for _, v := range point {
		if v > value && v <= 21 {
			value = v
		}
	}

	if value == 0 {
		value = point[0]
	}

	return
}

// GetSmallestPoint 获取最小都点
func GetSmallestPoint(point []int32) (value int32) {
	return point[0]
}

// CardsToString 牌组转字符串
func CardsToString(cards []byte) (cardsStr string) {
	for key, card := range cards {

		value, color := GetCardValueAndColor(card)

		switch color {
		case 0x1:
			cardsStr += "方块"
			break
		case 0x2:
			cardsStr += "梅花"
			break
		case 0x3:
			cardsStr += "红桃"
			break
		case 0x4:
			cardsStr += "黑桃"
			break
		}

		switch value {
		case 0x2:
			cardsStr += "-2"
			break
		case 0x3:
			cardsStr += "-3"
			break
		case 0x4:
			cardsStr += "-4"
			break
		case 0x5:
			cardsStr += "-5"
			break
		case 0x6:
			cardsStr += "-6"
			break
		case 0x7:
			cardsStr += "-7"
			break
		case 0x8:
			cardsStr += "-8"
			break
		case 0x9:
			cardsStr += "-9"
			break
		case 0xa:
			cardsStr += "-10"
			break
		case 0xb:
			cardsStr += "-J"
			break
		case 0xc:
			cardsStr += "-Q"
			break
		case 0xd:
			cardsStr += "-K"
			break
		case 0xe:
			cardsStr += "-A"
			break
		}

		if key == len(cards)-1 {
			continue
		}
		cardsStr += "/"
	}

	return
}

// CardsTypeToString 牌型转字符串
func CardsTypeToString(cardsType msg.CardsType) (cardsTypeStr string) {
	switch cardsType {
	case msg.CardsType_BlackJack:
		cardsTypeStr = "黑杰克"
		break
	case msg.CardsType_FiveDragon:
		cardsTypeStr = "五小龙"
		break
	case msg.CardsType_Point21:
		cardsTypeStr = "21点"
		break
	case msg.CardsType_Other:
		cardsTypeStr = "普通牌"
		break
	case msg.CardsType_Bust:
		cardsTypeStr = "爆牌"
		break
	}
	return
}
