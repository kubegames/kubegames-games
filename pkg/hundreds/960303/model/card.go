package model

import (
	"fmt"
	"math/rand"
)

// 黑红梅方
var Deck = []byte{
	0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1,
	0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2,
	0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3,
	0x14, 0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4,
}

const (
	NIU0      = 0
	NIU1      = 1
	NIU2      = 2
	NIU3      = 3
	NIU4      = 4
	NIU5      = 5
	NIU6      = 6
	NIU7      = 7
	NIU8      = 8
	NIU9      = 9
	NIUNIU    = 10
	ZHADANNIU = 11
	WUHUANIU  = 12
	WUXIAONIU = 13
)

type NiuNiuCard struct {
	Cards [5]byte
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
		index1 := rand.Intn(len(gp.Cards))
		index2 := rand.Intn(len(gp.Cards))
		gp.Cards[index1], gp.Cards[index2] = gp.Cards[index2], gp.Cards[index1]
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

func IsWuXiaoNiu(cards NiuNiuCard) bool {
	var Total byte
	for i := 0; i < len(cards.Cards); i++ {
		cardvalue, _ := GetCardValueAndColor(cards.Cards[i])
		cardvalue = cardvalue >> 4
		Total += cardvalue
		if cardvalue >= 5 {
			return false
		}
	}

	if Total >= 10 {
		return false
	}

	return true
}

func IsWuHuaNiu(cards NiuNiuCard) bool {
	for i := 0; i < len(cards.Cards); i++ {
		cardvalue, _ := GetCardValueAndColor(cards.Cards[i])
		cardvalue = cardvalue >> 4
		if cardvalue <= 10 {
			return false
		}
	}

	return true
}

func IsZhaDanNiu(cards NiuNiuCard) bool {
	count := 0
	var cardvalue0 byte
	cardvalue1, _ := GetCardValueAndColor(cards.Cards[0])
	cardvalue2, _ := GetCardValueAndColor(cards.Cards[1])
	cardvalue3, _ := GetCardValueAndColor(cards.Cards[2])
	if cardvalue1 == cardvalue2 || cardvalue1 == cardvalue3 {
		cardvalue0 = cardvalue1
	} else if cardvalue2 == cardvalue3 {
		cardvalue0 = cardvalue2
	} else {
		return false
	}

	for i := 0; i < len(cards.Cards); i++ {
		cardvalue, _ := GetCardValueAndColor(cards.Cards[i])
		if cardvalue0 == cardvalue {
			count++
		}
	}

	if count != 4 {
		return false
	}

	return true
}

func NiuDian(cards NiuNiuCard) int {
	var arr []int
	for i := 0; i < 3; i++ {
		cardvalue1, _ := GetCardValueAndColor(cards.Cards[i])
		cardvalue1 = cardvalue1 >> 4
		if cardvalue1 > 10 {
			cardvalue1 = 10
		}
		for j := i + 1; j < 4; j++ {
			cardvalue2, _ := GetCardValueAndColor(cards.Cards[j])
			cardvalue2 = cardvalue2 >> 4
			if cardvalue2 > 10 {
				cardvalue2 = 10
			}
			for m := j + 1; m < 5; m++ {
				cardvalue3, _ := GetCardValueAndColor(cards.Cards[m])
				cardvalue3 = cardvalue3 >> 4
				if cardvalue3 > 10 {
					cardvalue3 = 10
				}
				if (cardvalue1+cardvalue2+cardvalue3)%10 == 0 {
					arr = append(arr, i)
					arr = append(arr, j)
					arr = append(arr, m)
					break
				}
			}

			if len(arr) == 3 {
				break
			}
		}

		if len(arr) == 3 {
			break
		}
	}

	if len(arr) == 3 {
		var Total byte
		for i := 0; i < len(cards.Cards); i++ {
			bFind := false
			for j := 0; j < 3; j++ {
				if i == arr[j] {
					bFind = true
				}
			}

			if bFind {
				continue
			}

			cardvalue, _ := GetCardValueAndColor(cards.Cards[i])
			cardvalue = cardvalue >> 4
			if cardvalue > 10 {
				Total += 10
			} else {
				Total += cardvalue
			}
		}

		ret := Total % 10
		if ret == 0 {
			return 10
		}

		return int(Total % 10)
	}

	return 0
}

func IsNiuNiu(cards NiuNiuCard) bool {
	return (NiuDian(cards) == 10)
}

func GetNiuNiuType(cards NiuNiuCard) int {
	if IsWuXiaoNiu(cards) {
		return WUXIAONIU
	}

	if IsWuHuaNiu(cards) {
		return WUHUANIU
	}

	if IsZhaDanNiu(cards) {
		return ZHADANNIU
	}

	if IsNiuNiu(cards) {
		return NIUNIU
	}

	return NiuDian(cards)
}

func GetZhaDan(cards NiuNiuCard) byte {
	cardvalue1, _ := GetCardValueAndColor(cards.Cards[0])
	cardvalue2, _ := GetCardValueAndColor(cards.Cards[1])
	cardvalue3, _ := GetCardValueAndColor(cards.Cards[2])
	if cardvalue1 == cardvalue2 || cardvalue1 == cardvalue3 {
		return cards.Cards[0]
	} else {
		return cards.Cards[1]
	}
}

func GetMaxCard(cards NiuNiuCard) byte {
	var ret byte
	for i := 0; i < len(cards.Cards); i++ {
		if ret < cards.Cards[i] {
			ret = cards.Cards[i]
		}
	}

	return ret
}

//cards1 > cards2 返回，小于返回-1
func ComPareCard(cards1 NiuNiuCard, cards2 NiuNiuCard) int {
	type1 := GetNiuNiuType(cards1)
	type2 := GetNiuNiuType(cards2)
	if type1 > type2 {
		return 1
	} else if type1 < type2 {
		return -1
	}

	if type1 == ZHADANNIU {
		if GetZhaDan(cards1) > GetZhaDan(cards2) {
			return 1
		}

		return -1
	}

	if type1 != NIU0 {
		if GetMaxCard(cards1) > GetMaxCard(cards2) {
			return 1
		}

		return -1
	}

	dian1 := GetMaxCard(cards1) >> 4
	dian2 := GetMaxCard(cards2) >> 4
	if dian1 > dian2 {
		return 1
	} else if dian1 < dian2 {
		return -1
	}

	return 0
}

func GetCardValueString(v byte) string {
	return fmt.Sprintf("%v", v>>4)
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

func GetCardString(Card NiuNiuCard) []string {
	var str []string
	for _, v := range Card.Cards {
		tmp := GetColorString(v)
		tmp += GetCardValueString(v)
		str = append(str, tmp)
	}

	return str
}

var typestring = []string{"无牛", "牛一", "牛二", "牛三", "牛四", "牛五", "牛六", "牛七",
	"牛八", "牛九", "牛牛", "炸弹牛", "五花牛", "五小牛"}

func GetTypeString(t int32) string {
	return typestring[t]
}
