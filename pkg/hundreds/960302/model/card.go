package model

import (
	"fmt"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
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
	for i := 0; i < 8; i++ {
		for _, v := range Deck {
			gp.Cards = append(gp.Cards, v)
		}
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

func GetCardString(Card byte) []string {
	var str []string
	tmp := GetColorString(Card)
	tmp += GetCardValueString(Card)
	str = append(str, tmp)

	return str
}
