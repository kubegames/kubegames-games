package model

import (
	"fmt"

	"github.com/kubegames/kubegames-games/internal/pkg/poker"
)

// 随机两副牌
func DealRedAndBlack() (Red []byte, Black []byte) {
	Red = poker.GenerateCards()
	Black = poker.GenerateCards()
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

func GetCardString(Card []byte) []string {
	var str []string
	for _, v := range Card {
		tmp := GetColorString(v)
		tmp += GetCardValueString(v)
		str = append(str, tmp)
	}

	return str
}
