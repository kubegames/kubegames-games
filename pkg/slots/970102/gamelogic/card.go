package gamelogic

import (
	"strconv"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
)

// 黑红梅方
var Deck = []byte{
	0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1,
	0x22, 0x32, 0x42, 0x52, 0x62, 0x72, 0x82, 0x92, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	0x23, 0x33, 0x43, 0x53, 0x63, 0x73, 0x83, 0x93, 0xa3, 0xb3, 0xc3, 0xd3, 0xe3,
	0x24, 0x34, 0x44, 0x54, 0x64, 0x74, 0x84, 0x94, 0xa4, 0xb4, 0xc4, 0xd4, 0xe4,
}

var King = []byte{
	0xf1, 0xf2,
}

type GamePoker struct {
	Cards []byte
}

func (gp *GamePoker) InitPoker() {
	gp.Cards = make([]byte, 0)
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

func (gp *GamePoker) swap(index int) bool {
	if index >= len(gp.Cards) {
		return false
	}
	tem := gp.Cards[0]
	gp.Cards[0] = gp.Cards[index]
	gp.Cards[index] = tem
	return true
}

//随便输入一个数字，得出它的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = (value & 0xff) >> 4 //byte的高4位总和是240
	cardColor = value & 0xf         //byte的低4位总和是15
	return
}

func GetDValue(card1 byte, card2 byte) byte {
	value1, _ := GetCardValueAndColor(card1)
	value2, _ := GetCardValueAndColor(card2)
	d := value1 - value2
	if d > 0 {
		return d
	}
	return -d
}

func TransformCard(card byte) string {
	value, huaSe := GetCardValueAndColor(card)
	if value == 15 && huaSe == 0x01 {
		return "小王"
	}
	if value == 15 && huaSe == 0x02 {
		return "大王"
	}
	return GetHuaSeString(huaSe) + GetValueString(value)
}

func GetValueString(value byte) string {
	if value <= 10 {
		return strconv.Itoa(int(value))
	}
	switch value {
	case 0xb:
		return "J"
	case 0xc:
		return "Q"
	case 0xd:
		return "K"
	case 0xe:
		return "A"
	default:
		//log.Errorf("crad value err", value)
		return ""
	}
}

func GetHuaSeString(huaSe byte) string {
	switch huaSe {
	case 0x01:
		return "黑桃-"
	case 0x02:
		return "红桃-"
	case 0x03:
		return "梅花-"
	case 0x04:
		return "方块-"
	default:
		//log.Errorf("crad huaSe err", huaSe)
		return ""
	}
}

func GetCardString(cards []byte) string {
	cardString := ""
	length := len(cards)
	for i := 0; i < length; i++ {
		v := cards[i]
		if v == 0 {
			continue
		}
		cardString += TransformCard(v)
		if i < length-1 {
			cardString += "/"
		}
	}
	return cardString
}
