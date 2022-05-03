package model

import (
	"fmt"
	"math/rand"
)

// 黑红梅方
var Deck = []byte{
	0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
	0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
	0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
}

//前端显示枚举
//CardTypeBZ     = 6 //豹子
//CardTypeSJ     = 5 //顺金
//CardTypeJH     = 4 //金花
//CardTypeSZ     = 3 //顺子
//CardTypeDZ     = 2 //对子
//CardTypeSingle = 1 //高牌
//CardTypeDL     = 0 //地龙
var (
	CardTypeBZ     = 8 //豹子
	CardTypeSJ     = 7 //顺金
	CardTypeSJA23  = 6 //顺金A23
	CardTypeJH     = 5 //金花
	CardTypeSZ     = 4 //顺子
	CardTypeSZA23  = 3 //顺子A23
	CardTypeDZ     = 2 //对子
	CardTypeSingle = 1 //高牌
	CardTypeDL     = 0 //地龙
)

type JHCard struct {
	Cards [3]byte
}

type GamePoker struct {
	Cards []byte
}

func (gp *GamePoker) InitPoker() {
	for _, v := range Deck {
		gp.Cards = append(gp.Cards, v)
	}
}

func (gp *GamePoker) ShuffleCards() {
	for i := 0; i < len(gp.Cards); i++ {
		index := rand.Intn(len(gp.Cards))
		gp.Cards[i], gp.Cards[index] = gp.Cards[index], gp.Cards[i]
	}
}

func (gp *GamePoker) DealCards() byte {
	card := gp.Cards[0]
	gp.Cards = append(gp.Cards[:0], gp.Cards[1:]...)
	return card
}

func (gp *GamePoker) GetCardsCount() int {
	return len(gp.Cards)
}

//随便输入一个数字，得出它的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = (value & 0xf0) //byte的高4位总和是240
	cardColor = value & 0xf    //byte的低4位总和是15
	return
}

//将牌 进行牌型编码并返回
func GetEncodeCard(cardType int, cards [3]byte) (cardEncode int) {
	cardEncode = (cardType) << 20
	if cardType != CardTypeDZ {
		for i, card := range cards {
			cardEncode |= (int(card) >> 4) << uint((5-i-1)*4)
		}
	} else {
		dui := cards[0]
		dan := cards[0]
		if cards[1]&0xf0 == cards[2]&0xf0 {
			dui = cards[1]
		} else {
			dan = cards[2]
		}

		cardEncode |= (int(dui) >> 4) << uint((4)*4)
		cardEncode |= (int(dui) >> 4) << uint((3)*4)
		cardEncode |= (int(dan) >> 4) << uint((2)*4)
	}

	return
}

//获取金花牌型
//将整个牌型进行sort，返回牌的类型和排序之后的具体值
//比如 3、8、3、3、9，则返回 葫芦(cardType)，33389(排序之后的牌面值)
func GetCardTypeJH(cards [3]byte) (cardType int, sortRes [3]byte) {
	//if len(cards) != 3 {
	//	log.Traceln("金花牌型比较只能3张牌")
	//	return
	//}
	sortRes = SortCards(cards)
	//从大到小
	//if isCardTypeAAA(cards) {
	//	cardType = CardTypeAAA
	//	return
	//}
	//豹子 6
	if isCardTypeBZ(cards) {
		cardType = CardTypeBZ
		return
	}

	//顺金 5
	if isCardTypeSJ(cards, true) {
		cardType = CardTypeSJ
		return
	}
	//顺金A23
	if IsCardTypeSJ123(cards, true) {
		cardType = CardTypeSJA23
		return
	}

	//金花 4
	if isCardTypeJH(cards) {
		cardType = CardTypeJH
		return
	}
	//顺子 3
	if isCardTypeSZ(cards, true) {
		cardType = CardTypeSZ
		return
	}
	//顺子A23
	if isCardTypeSZA23(cards, true) {
		cardType = CardTypeSZA23
		return
	}
	//对子 2
	if isCardTypeDZ(cards, true) {
		cardType = CardTypeDZ
		return
	}
	//地龙
	if isCardTypeDiLong(cards, true) {
		cardType = CardTypeDL
		return
	}
	//单张 1
	cardType = CardTypeSingle
	return
}
func GetCardsType(cards [3]byte) int32 {
	cardsType, _ := GetCardTypeJH(cards)
	return int32(cardsType)

}

//判断牌型是否为豹子
func isCardTypeBZ(cards [3]byte) bool {
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return cv0 == cv1 && cv0 == cv2
}

//判断牌型是否为顺金
func isCardTypeSJ(cards [3]byte, isSort bool) bool {
	return isCardTypeJH(cards) && isCardTypeSZ(cards, isSort)
}

//判断牌型是否为金花
func isCardTypeJH(cards [3]byte) bool {
	_, cc0 := GetCardValueAndColor(cards[0])
	_, cc1 := GetCardValueAndColor(cards[1])
	_, cc2 := GetCardValueAndColor(cards[2])
	return cc0 == cc1 && cc0 == cc2
}

//判断牌型是否为顺子
func isCardTypeSZ(cards [3]byte, isSort bool) bool {
	if !isSort {
		cards = SortCards(cards)
	}
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return (cv0-cv1) == 16 && (cv1-cv2) == 16
}

//判断牌型是否为顺子 A23
func isCardTypeSZA23(cards [3]byte, isSort bool) bool {
	if !isSort {
		cards = SortCards(cards)
	}
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return cv0 == 0xe0 && cv1 == 0x30 && cv2 == 0x20
}

//判断牌型是否为对子
func isCardTypeDZ(cards [3]byte, isSort bool) bool {
	if !isSort {
		cards = SortCards(cards)
	}
	//因为在比较对子之前已经过滤了四条、三条等情况，所以只需判断牌中有相同的就直接返回对子
	for i := 0; i < len(cards)-1; i++ {
		cvi, _ := GetCardValueAndColor(cards[i])
		for j := i + 1; j < len(cards); j++ {
			cvj, _ := GetCardValueAndColor(cards[j])
			if cvi == cvj {
				return true
			}
		}
	}
	return false
}

//将牌进行排序
func SortCards(cards [3]byte) [3]byte {
	for i := 0; i < len(cards)-1; i++ {
		for j := 0; j < (len(cards) - 1 - i); j++ {
			if (cards)[j] < (cards)[j+1] {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
	return cards
}

//判断牌型是否为顺金123
func IsCardTypeSJ123(cards [3]byte, isSort bool) bool {
	if !isSort {
		cards = SortCards(cards)
	}

	cv0, cc0 := GetCardValueAndColor(cards[0])
	cv1, cc1 := GetCardValueAndColor(cards[1])
	cv2, cc2 := GetCardValueAndColor(cards[2])
	if cc0 != cc1 || cc0 != cc2 {
		return false
	}

	return cv0 == 0xe0 && cv1 == 0x30 && cv2 == 0x20
}

//判断是否是地龙
func isCardTypeDiLong(cards [3]byte, isSort bool) bool {
	if !isSort {
		cards = SortCards(cards)
	}

	cv0, cc0 := GetCardValueAndColor(cards[0])
	cv1, cc1 := GetCardValueAndColor(cards[1])
	cv2, cc2 := GetCardValueAndColor(cards[2])
	if cv0 == 0x50 && cv1 == 0x30 && cv2 == 0x20 {
		return cc0 != cc1 || cc0 != cc2
	}
	return false
}

//判断是否是AAA
func isCardTypeAAA(cards [3]byte) bool {
	if !isCardTypeBZ(cards) {
		return false
	} else {
		cv0, _ := GetCardValueAndColor(cards[0])
		return cv0 == 0xe0
	}
}

func GetTypeString(t int32) string {
	//t, _ := GetCardTypeJH(cards)

	switch int(t) {
	case CardTypeSingle:
		return "高牌"
	case CardTypeDL:
		return "地龙"
	case CardTypeDZ:
		return "对子"
	//case CardTypeSZA23:
	//	return "顺子A23"
	case CardTypeSZ:
		return "顺子"
	case CardTypeJH:
		return "金花"
	//case CardTypeSJA23:
	//	return "顺金A23"
	case CardTypeSJ:
		return "顺金"
	case CardTypeBZ:
		return "豹子"
	}

	return ""
}

func GetColorString(v byte) string {
	tmp := v & 0xf
	switch tmp {
	case 1:
		return "方块"
	case 2:
		return "樱花"
	case 3:
		return "红桃"
	case 4:
		return "黑桃"
	}

	return ""
}
func GetCardValueString(v byte) string {
	return fmt.Sprintf("%v", v>>4)
}

func GetCardString(Card [3]byte) []string {
	var str []string
	for _, v := range Card {
		tmp := GetColorString(v)
		tmp += GetCardValueString(v)
		str = append(str, tmp)
	}

	return str
}

//cards1 > cards2 返回1，小于返回2
func ComPareCard(cards1 [3]byte, cards2 [3]byte) int {
	//defer log.Trace()()
	//Red := cards1
	//Black := cards2
	cards1Type, Sortcards1 := GetCardTypeJH(cards1)
	cards2Type, Sortcards2 := GetCardTypeJH(cards2)
	// 地龙 豹子A比牌 地龙大于豹子A
	if (isCardTypeAAA(cards1) && cards2Type == CardTypeDL) || (isCardTypeAAA(cards2) && cards1Type == CardTypeDL) {
		if cards1Type < cards2Type {
			return 1
		} else {
			return 2
		}
	}
	if cards1Type > cards2Type {
		return 1
	} else if cards1Type == cards2Type {
		EncodeRed := GetEncodeCard(cards1Type, Sortcards1)
		EncodeBlack := GetEncodeCard(cards2Type, Sortcards2)
		if EncodeRed > EncodeBlack {
			return 1

		} else if EncodeRed < EncodeBlack {
			return 2

		} else {
			if cards1Type == CardTypeDZ {
				dui1 := Sortcards1[0] & 0xf0
				dui2 := Sortcards2[0] & 0xf0
				dan1 := Sortcards1[0]
				dan2 := Sortcards2[0]
				if Sortcards1[1] == Sortcards1[2] {
					dui1 = Sortcards1[1] & 0xf0
				} else {
					dan1 = Sortcards1[2]
				}

				if Sortcards2[1] == Sortcards2[2] {
					dui2 = Sortcards2[1] & 0xf0
				} else {
					dan2 = Sortcards2[2]
				}

				if dui1 > dui2 {
					return 1
				} else if dui1 < dui2 {
					return 2
				} else if dan1 > dan2 {
					return 1
				} else {
					return 2
				}
			} else if Sortcards1[0] > Sortcards2[0] {
				return 1
			} else {
				return 2
			}
		}
	} else {
		return 2
	}
}
