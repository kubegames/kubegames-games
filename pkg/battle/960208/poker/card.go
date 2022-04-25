package poker

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
)

const ()

// Deck 基础牌，方梅红黑
var Deck = []byte{
	0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1,
	0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2,
	0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3,
	0x14, 0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4,
}

// GamePoker 牌堆
type GamePoker struct {
	Cards []byte
}

// HoldCards 持有手牌
type HoldCards struct {
	Cards     []byte        // 手牌
	CardsType msg.CardsType // 牌型
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

// DrawCard 抽牌
func (gamePoker *GamePoker) DrawCard() (cards []byte) {

	length := len(gamePoker.Cards)

	cards = gamePoker.Cards[length-3:]
	gamePoker.Cards = gamePoker.Cards[:(length - 3)]

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

// GetCardsType 获取牌型
func GetCardsType(cards []byte) msg.CardsType {

	// 牌长度不为三张
	if len(cards) != 3 {
		return msg.CardsType_Unknown
	}

	// 是否是爆玖牌型
	if IsExplosionNine(cards) {
		return msg.CardsType_ExplosionNine
	}

	// 是否是炸弹牌型
	if IsBoom(cards) {
		return msg.CardsType_Boom
	}

	// 是否是炸弹牌型
	if IsThreeDoll(cards) {
		return msg.CardsType_ThreeDoll
	}

	// 获取点数牌点数
	point := GetPointByCards(cards)
	switch point {
	case 0:
		return msg.CardsType_ZeroPoint
	case 1:
		return msg.CardsType_OnePoint
	case 2:
		return msg.CardsType_TwoPoint
	case 3:
		return msg.CardsType_ThreePoint
	case 4:
		return msg.CardsType_FourPoint
	case 5:
		return msg.CardsType_FivePoint
	case 6:
		return msg.CardsType_SixPoint
	case 7:
		return msg.CardsType_SevenPoint
	case 8:
		return msg.CardsType_EightPoint
	case 9:
		return msg.CardsType_NinePoint
	}

	return msg.CardsType_Unknown
}

// IsExplosionNine 是否是爆玖牌型
func IsExplosionNine(cards []byte) bool {
	for _, card := range cards {
		value, _ := GetCardValueAndColor(card)
		if value != 0x3 {
			return false
		}
	}
	return true
}

// IsBoom 是否是炸弹牌型
func IsBoom(cards []byte) bool {
	boomValue, _ := GetCardValueAndColor(cards[0])
	for _, card := range cards {
		value, _ := GetCardValueAndColor(card)
		if value != boomValue {
			return false
		}
	}
	return true
}

// IsThreeDoll 是否是三公牌型
func IsThreeDoll(cards []byte) bool {
	for _, card := range cards {
		value, _ := GetCardValueAndColor(card)
		if value < 0xb {
			return false
		}
	}
	return true
}

// GetPointByCards 获取点数牌点数
func GetPointByCards(cards []byte) (point int32) {
	for _, card := range cards {
		value, _ := GetCardValueAndColor(card)
		if value < 0xa {
			point += int32(value)
		}
	}
	point = point % 10
	return
}

// GetCardsMultiple 获取牌型倍数
func GetCardsMultiple(cardsType msg.CardsType) (multiple int64) {

	// 零点~六点：1倍
	if cardsType >= msg.CardsType_ZeroPoint && cardsType <= msg.CardsType_SixPoint {
		multiple = 1
	}

	// 七点~九点：2倍
	if cardsType >= msg.CardsType_SevenPoint && cardsType <= msg.CardsType_NinePoint {
		multiple = 2
	}

	// 三公：3倍
	if cardsType == msg.CardsType_ThreeDoll {
		multiple = 3
	}

	// 炸弹：4倍
	if cardsType == msg.CardsType_Boom {
		multiple = 4
	}

	// 爆玖：5倍
	if cardsType == msg.CardsType_ExplosionNine {
		multiple = 5
	}

	return
}

// ContrastCards 比牌
func ContrastCards(bankerCards *HoldCards, playerCards *HoldCards) bool {
	// 闲家赢
	if playerCards.CardsType > bankerCards.CardsType {
		return true
	}

	// 牌型相同
	if playerCards.CardsType == bankerCards.CardsType {
		// 获取庄家和闲家手牌最大牌值
		bankerBiggestCard := GetBiggestFromCards(bankerCards.Cards)
		playerBiggestCard := GetBiggestFromCards(playerCards.Cards)

		// 点数牌先比较公仔牌数
		if playerCards.CardsType < msg.CardsType_ThreeDoll {
			// 公仔数量
			bankerDollCount := getDollCount(bankerCards.Cards)
			playerDollCount := getDollCount(playerCards.Cards)

			if playerDollCount > bankerDollCount {
				return true
			} else if bankerDollCount < playerDollCount {
				return false
			}
		}

		// 其他相同，比较最大牌
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
func GetBiggestFromCards(cards []byte) (biggestCard byte) {

	for _, card := range cards {
		if card > biggestCard {
			biggestCard = card
		}
	}
	return

}

// getDollCount 获取公仔数量
func getDollCount(cards []byte) (count int) {
	for _, card := range cards {
		value, _ := GetCardValueAndColor(card)
		if value > 0xa {
			count++
		}
	}
	return
}

// GetInputCardsType 获取指定牌型手牌
func GetInputCardsType(cardsType msg.CardsType) (cards []byte) {
	gamePoker := new(GamePoker)
	gamePoker.InitPoker()

	switch cardsType {
	// 爆玖
	case msg.CardsType_ExplosionNine:
		// 找到3张3
		for _, card := range gamePoker.Cards {
			value, _ := GetCardValueAndColor(card)
			if value == 0x3 {
				cards = append(cards, card)
			}
			if len(cards) == 3 {
				break
			}
		}
		break
		// 炸弹
	case msg.CardsType_Boom:
		// 找到3张相同牌
		boomValue, _ := GetCardValueAndColor(gamePoker.Cards[0])
		for _, card := range gamePoker.Cards {
			value, _ := GetCardValueAndColor(card)
			if value == boomValue {
				cards = append(cards, card)
			}
			if len(cards) == 3 {
				break
			}
		}

		break
		// 三公
	case msg.CardsType_ThreeDoll:
		// 找到3大于等于J的牌
		for _, card := range gamePoker.Cards {
			value, _ := GetCardValueAndColor(card)
			if value >= 0xb {
				cards = append(cards, card)
			}
			if len(cards) == 3 {
				break
			}
		}
		break
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
	case msg.CardsType_Unknown:
		typeString = "未知牌型"
		break
	case msg.CardsType_ZeroPoint:
		typeString = "0点"
		break
	case msg.CardsType_OnePoint:
		typeString = "1点"
		break
	case msg.CardsType_TwoPoint:
		typeString = "2点"
		break
	case msg.CardsType_ThreePoint:
		typeString = "3点"
		break
	case msg.CardsType_FourPoint:
		typeString = "4点"
		break
	case msg.CardsType_FivePoint:
		typeString = "5点"
		break
	case msg.CardsType_SixPoint:
		typeString = "6点"
		break
	case msg.CardsType_SevenPoint:
		typeString = "7点"
		break
	case msg.CardsType_EightPoint:
		typeString = "8点"
		break
	case msg.CardsType_NinePoint:
		typeString = "9点"
		break
	case msg.CardsType_ThreeDoll:
		typeString = "三公"
		break
	case msg.CardsType_Boom:
		typeString = "炸弹"
		break
	case msg.CardsType_ExplosionNine:
		typeString = "爆玖"
		break
	}
	return
}
