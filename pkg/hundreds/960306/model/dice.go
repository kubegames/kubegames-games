package model

import (
	"fmt"
	"math/rand"
)

//类型与下注的区域位置相同
const (
	BigType    = 0
	SmallType  = 1
	SingleType = 2
	DoubleType = 3
	WeiType    = 18
)

// 黑红梅方
var Deck = []byte{
	0x11, 0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1,
	0x12, 0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2,
	0x13, 0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3,
	0x14, 0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4,
}

type GamePoker struct {
	Cards []byte
}

func (gp *GamePoker) InitPoker() {
	gp.Cards = make([]byte, 0)
	for i := 0; i < 8; i++ {
		for _, v := range Deck {
			gp.Cards = append(gp.Cards, v)
		}
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

func (gp *GamePoker) AddCard(value byte) {
	gp.Cards = append(gp.Cards, value)
}

//随便输入一个数字，得出它的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = (value & 0xf0) >> 4 //byte的高4位总和是240
	cardColor = value & 0xf         //byte的低4位总和是15
	return
}

func GetCardValue(card byte) byte {
	return (card & 0xf0) >> 4
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

func GetCardString(Card [3]byte) []string {
	var str []string
	for _, v := range Card {
		if v == 0 {
			continue
		}
		tmp := GetColorString(v)
		tmp += GetCardValueString(v)
		str = append(str, tmp)
	}

	return str
}

func GetTypeString(Card [3]byte) string {
	dian := byte(0)
	for i := 0; i < 3; i++ {
		v := GetCardValue(Card[i])
		if v > 10 {
			v = 10
		}
		dian += v
	}

	dian = dian % 10

	str := fmt.Sprint(dian, "点")

	return str
}

//获取骰子数
func GetDices() (TouZi [3]int32) {
	for i := 0; i <= 2; i++ {
		N := rand.Intn(6) + 1
		TouZi[i] = int32(N)
	}
	return
}

//获取骰子结果 点数 类型1： 大小围0，1，18 类型2： 单双围2，3，18
func GetDiceResult(TouZi [3]int32) (Count int32, BSWType int32, SDWType int32) {
	//计算点数
	for i := 0; i <= 2; i++ {
		Count += TouZi[i]
	}
	if TouZi[0] == TouZi[1] && TouZi[0] == TouZi[2] {
		BSWType = WeiType
		SDWType = WeiType
		return
	}
	// 单双判断
	if Count%2 == 0 {
		SDWType = DoubleType
	} else {
		SDWType = SingleType
	}
	//大小判断
	if Count < 11 {
		BSWType = SmallType
	} else {
		BSWType = BigType
	}
	return
}

//func GetOdds(TouZi [3]int32) (odds int64) {
//	//计算点数
//	Count,bsw,_:=GetDiceResult(TouZi)
//	if bsw==WeiType {
//		return  0
//	}
//	if Count == 4 || Count == 17 {
//		return 51
//	} else if Count == 5 || Count == 16 {
//		return 19
//	} else if Count == 6 || Count == 15 {
//		return 15
//	} else if Count == 7 || Count == 14 {
//		return 13
//	} else if Count == 8 || Count == 13 {
//		return 9
//	} else if Count == 9 || Count == 12 || Count == 10 || Count == 11 {
//		return 7
//	}
//}
