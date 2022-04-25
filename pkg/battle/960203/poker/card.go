package poker

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960203/msg"
)

var CardCount = map[byte]int{
	0x1: 1,
	0x2: 2,
	0x3: 3,
	0x4: 4,
	0x5: 5,
	0x6: 6,
	0x7: 7,
	0x8: 8,
	0x9: 9,
	0xa: 10,
	0xb: 10,
	0xc: 10,
	0xd: 10,
}

// GamePoker 牌堆
type GamePoker struct {
	Cards []byte
}

// DecodeCard 解析后的牌组
type DecodeCard struct {
	Value byte
	Color byte
}

// HoldCards 持有手牌
type HoldCards struct {
	Cards             []byte        // 手牌
	FirstHalfCards    []byte        // 上半截卡牌
	LowerHalfCards    []byte        // 下半截卡牌
	SpecialCardIndexs []int32       // 特俗处理手牌
	CardsType         msg.CardsType // 牌型
}

// InitPoker 初始化牌组
func (gamePoker *GamePoker) InitPoker() {
	//使用一副牌，去除大小王，共计52张牌 方梅红黑
	gamePoker.Cards = []byte{
		0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1,
		0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2,
		0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3,
		0x14, 0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4,
	}

	// 洗牌
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(gamePoker.Cards), func(i, j int) {
		gamePoker.Cards[i], gamePoker.Cards[j] = gamePoker.Cards[j], gamePoker.Cards[i]
	})
}

// DrawCard 抽5张牌
func (gamePoker *GamePoker) DrawCard() (cards []byte) {

	length := len(gamePoker.Cards)

	cards = gamePoker.Cards[length-5:]

	gamePoker.Cards = gamePoker.Cards[:(length - 5)]

	return
}

// PlugSelectedCard 塞入一张选定牌牌
func (gamePoker *GamePoker) PlugSelectedCard(cardValue byte) {
	for k, v := range gamePoker.Cards {
		value, _ := GetCardValueAndColor(v)

		if value == cardValue {
			gamePoker.Cards = append(gamePoker.Cards[:k], gamePoker.Cards[k+1:]...)
			gamePoker.Cards = append(gamePoker.Cards, v)
			break
		}

	}
}

// PlugRangeCard 塞入一张范围内的牌
func (gamePoker *GamePoker) PlugRangeCard(limitMin, limitMax byte) {
	for k, v := range gamePoker.Cards {
		value, _ := GetCardValueAndColor(v)

		if value >= limitMin && value <= limitMax {
			gamePoker.Cards = append(gamePoker.Cards[:k], gamePoker.Cards[k+1:]...)
			gamePoker.Cards = append(gamePoker.Cards, v)
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

// GetCardByValueAndColor 通过牌值和花色获取牌
func GetCardByValueAndColor(value byte, color byte) (card byte) {
	card = (value << 4) | color
	return
}

// GetCardsType 获取牌的类型
func GetCardsType(cards []byte) (cardsType msg.CardsType) {

	decodeCards := DecodeCards(cards)

	// 判断是五小牛
	if IsFiveSmallerNiu(decodeCards) {
		cardsType = msg.CardsType_FiveSmallerNiu
		return
	}

	// 判断是否是五花牛
	if IsFiveColourfulNiu(decodeCards) {
		cardsType = msg.CardsType_FiveColourfulNiu
		return
	}

	// 判断是否是四炸
	if IsFourBomb(decodeCards) {
		cardsType = msg.CardsType_FourBomb
		return
	}

	cardsType = GetNiuByCards(decodeCards)

	return
}

// IsFiveSmallerNiu 牌型判断：五小牛
func IsFiveSmallerNiu(decodeCards []DecodeCard) bool {
	if len(decodeCards) != 5 {
		return false
	}

	sum := 0

	for _, card := range decodeCards {

		if card.Value >= 0x5 {
			return false
		}

		sum += int(card.Value)
	}

	if sum > 10 {
		return false
	}

	return true
}

// IsFiveColourfulNiu 牌型判断：五花牛
func IsFiveColourfulNiu(decodeCards []DecodeCard) bool {
	if len(decodeCards) != 5 {
		return false
	}

	for _, card := range decodeCards {
		if card.Value < 0xb {
			return false
		}
	}

	return true
}

// IsFourBomb 牌型判断：四炸
func IsFourBomb(decodeCards []DecodeCard) bool {
	if len(decodeCards) != 5 {
		return false
	}

	for _, i := range decodeCards {
		sameCount := 0
		for _, j := range decodeCards {
			if i.Value == j.Value {
				sameCount++
			}
		}

		if sameCount >= 4 {
			return true
		}
	}

	return false
}

// GetNiuByCards 获取牌的牛型
func GetNiuByCards(decodeCards []DecodeCard) msg.CardsType {
	if len(decodeCards) != 5 {
		return msg.CardsType_NotNiu
	}

	var lave, sum int

	for _, card := range decodeCards {
		valueCount := CardCount[card.Value]
		sum += valueCount
	}
	lave = sum % 10

	for i := 0; i < len(decodeCards); i++ {
		for j := i + 1; j < len(decodeCards); j++ {
			if (CardCount[decodeCards[i].Value]+CardCount[decodeCards[j].Value])%10 == lave {
				if lave == 0 {
					lave = 10
				}
				return msg.CardsType(lave + 1)

			}

		}

	}

	return msg.CardsType_NotNiu
}

// DecodeCards 编译牌
func DecodeCards(cards []byte) (decodeCards []DecodeCard) {
	for _, card := range cards {
		value, color := GetCardValueAndColor(card)

		decodeCards = append(decodeCards, DecodeCard{
			Value: value,
			Color: color,
		})
	}

	return
}

// ContrastCards 比牌
func ContrastCards(bankerCards *HoldCards, playerCards *HoldCards) bool {

	// 闲家赢
	if playerCards.CardsType > bankerCards.CardsType {
		return true
	}

	// 牌型相同，比牌的大小
	if playerCards.CardsType == bankerCards.CardsType {

		// 获取庄家和闲家手牌最大牌值
		bankerBiggestCard := GetBiggestFromCards(bankerCards)
		playerBiggestCard := GetBiggestFromCards(playerCards)

		if playerBiggestCard > bankerBiggestCard {
			return true
		}

		// 闲家点数小闲家输
		return false

	}

	// 闲家牌型小闲家输
	return false
}

// GetBiggestFromCards 从牌组中选出最大的牌
func GetBiggestFromCards(holdCards *HoldCards) (biggestCard byte) {

	// 牌型为四炸时，从四张牌值相同的牌中选出最大牌
	if holdCards.CardsType == msg.CardsType_FourBomb {
		for _, i := range holdCards.Cards {

			sameCount := 0
			biggerCard := i

			for _, j := range holdCards.Cards {

				jValue, _ := GetCardValueAndColor(j)
				biggerValue, _ := GetCardValueAndColor(biggerCard)

				if biggerValue == jValue {
					sameCount++

					if j > biggerCard {
						biggerCard = j
					}
				}
			}

			if sameCount >= 4 {
				biggestCard = biggerCard
				break
			}
		}

	} else {

		// 牌型不为四炸时，选出最大牌
		for _, card := range holdCards.Cards {
			if card > biggestCard {
				biggestCard = card
			}
		}
	}

	return

}

// GetSpecialCardIndexs 获取特殊牌
func GetSpecialCardIndexs(cards []byte, cardsType msg.CardsType) (specialCardIndexs []int32) {

	// 四炸弹
	if cardsType == msg.CardsType_FourBomb {
		for i, iCard := range cards {
			sameCount := 0
			iCardValue, _ := GetCardValueAndColor(iCard)

			for _, jCard := range cards {

				jCardValue, _ := GetCardValueAndColor(jCard)

				if iCardValue == jCardValue {
					sameCount++
				}

			}

			if sameCount == 1 {
				specialCardIndexs = append(specialCardIndexs, int32(i))
				return
			}
		}
	}

	// 牛型牌
	if cardsType <= msg.CardsType_NiuNiu && cardsType >= msg.CardsType_NiuOne {

		lave := int(cardsType) - 1
		if cardsType == msg.CardsType_NiuNiu {
			lave = 0
		}

		for i := 0; i < len(cards); i++ {
			iCardValue, _ := GetCardValueAndColor(cards[i])
			for j := i + 1; j < len(cards); j++ {
				jCardValue, _ := GetCardValueAndColor(cards[j])
				if (CardCount[iCardValue]+CardCount[jCardValue])%10 == lave {
					specialCardIndexs = append(specialCardIndexs, int32(i), int32(j))
					return
				}

			}

		}

	}

	return
}

// GetCardsMultiple 获取牌型倍数
func GetCardsMultiple(cardsType msg.CardsType) (multiple int64) {

	// 无牛到牛六 1 倍
	if cardsType >= msg.CardsType_NotNiu && cardsType <= msg.CardsType_NiuSix {
		multiple = 1
	}

	// 无七到牛九 2 倍
	if cardsType >= msg.CardsType_NiuSeven && cardsType <= msg.CardsType_NiuNine {
		multiple = 2
	}

	// 牛牛 3 倍
	if cardsType == msg.CardsType_NiuNiu {
		multiple = 3
	}

	// 四炸 4 倍
	if cardsType == msg.CardsType_FourBomb {
		multiple = 4
	}

	// 五花牛或者五小牛 5 倍
	if cardsType == msg.CardsType_FiveColourfulNiu || cardsType == msg.CardsType_FiveSmallerNiu {
		multiple = 5
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
		case 0x1:
			cardsStr += "-A"
			break
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
		}

		if key == len(cards)-1 {
			continue
		}
		cardsStr += "/"
	}

	return
}

// TransformCardsType 转译牌型
func TransformCardsType(cardsType msg.CardsType) (typeString string) {

	switch cardsType {
	case msg.CardsType_NotNiu:
		typeString = "无牛"
		break
	case msg.CardsType_NiuOne:
		typeString = "牛1"
		break
	case msg.CardsType_NiuTwo:
		typeString = "牛2"
		break
	case msg.CardsType_NiuThree:
		typeString = "牛3"
		break
	case msg.CardsType_NiuFour:
		typeString = "牛4"
		break
	case msg.CardsType_NiuFive:
		typeString = "牛5"
		break
	case msg.CardsType_NiuSix:
		typeString = "牛6"
		break
	case msg.CardsType_NiuSeven:
		typeString = "牛7"
		break
	case msg.CardsType_NiuEight:
		typeString = "牛8"
		break
	case msg.CardsType_NiuNine:
		typeString = "牛9"
		break
	case msg.CardsType_NiuNiu:
		typeString = "牛牛"
		break
	case msg.CardsType_FourBomb:
		typeString = "四炸"
		break
	case msg.CardsType_FiveColourfulNiu:
		typeString = "五花牛"
		break
	case msg.CardsType_FiveSmallerNiu:
		typeString = "五小牛"
		break
	}
	return
}
