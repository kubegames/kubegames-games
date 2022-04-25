package poker

import "fmt"

var (
	CardTypeBZ     = 8 //豹子
	CardTypeSJ     = 7 //顺金
	CardTypeSJA23  = 6 //顺金A23
	CardTypeJH     = 5 //金花
	CardTypeSZ     = 4 //顺子
	CardTypeSZA23  = 3 //顺子A23
	CardTypeDZ     = 2 //对子
	CardTypeSingle = 1 //单张
)

//获取金花牌型
//将整个牌型进行sort，返回牌的类型和排序之后的具体值
//比如 3、8、3、3、9，则返回 葫芦(cardType)，33389(排序之后的牌面值)
func GetCardTypeJH(cards []byte) (cardType int, sortRes []byte) {
	if len(cards) != 3 {
		fmt.Println("金花牌型比较只能3张牌")
		return
	}
	sortRes = sortCards(cards)
	//从从大到小

	//豹子 7
	if isCardTypeBZ(cards) {
		cardType = CardTypeBZ
		return
	}

	//顺金 6
	if isCardTypeSJ(cards, true) {
		cardType = CardTypeSJ
		return
	}

	if IsCardTypeSJ123(cards, true) {
		cardType = CardTypeSJA23
		return
	}

	//金花 5
	if isCardTypeJH(cards) {
		cardType = CardTypeJH
		return
	}
	//顺子 4
	if isCardTypeSZ(cards, true) {
		cardType = CardTypeSZ
		return
	}
	//顺子A23 3
	if isCardTypeSZA23(cards, true) {
		cardType = CardTypeSZA23
		return
	}
	//对子 2
	if isCardTypeDZ(cards, true) {
		cardType = CardTypeDZ
		return
	}
	//单张 1
	cardType = CardTypeSingle
	return
}

//判断牌型是否为豹子
func isCardTypeBZ(cards []byte) bool {
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return cv0 == cv1 && cv0 == cv2
}

//判断牌型是否为顺金
func isCardTypeSJ(cards []byte, isSort bool) bool {
	return isCardTypeJH(cards) && isCardTypeSZ(cards, isSort)
}

//判断牌型是否为金花
func isCardTypeJH(cards []byte) bool {
	_, cc0 := GetCardValueAndColor(cards[0])
	_, cc1 := GetCardValueAndColor(cards[1])
	_, cc2 := GetCardValueAndColor(cards[2])
	return cc0 == cc1 && cc0 == cc2
}

//判断牌型是否为顺子
func isCardTypeSZ(cards []byte, isSort bool) bool {
	if !isSort {
		cards = sortCards(cards)
	}
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return (cv0-cv1) == 16 && (cv1-cv2) == 16
}

//判断牌型是否为顺子 A23
func isCardTypeSZA23(cards []byte, isSort bool) bool {
	if !isSort {
		cards = sortCards(cards)
	}
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return cv0 == 0xe0 && cv1 == 0x30 && cv2 == 0x20
}

//判断牌型是否为对子
func isCardTypeDZ(cards []byte, isSort bool) bool {
	if !isSort {
		cards = sortCards(cards)
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
func sortCards(cards []byte) []byte {
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
func IsCardTypeSJ123(cards []byte, isSort bool) bool {
	if !isSort {
		cards = sortCards(cards)
	}

	cv0, cc0 := GetCardValueAndColor(cards[0])
	cv1, cc1 := GetCardValueAndColor(cards[1])
	cv2, cc2 := GetCardValueAndColor(cards[2])
	if cc0 != cc1 || cc0 != cc2 {
		return false
	}

	return cv0 == 0xe0 && cv1 == 0x30 && cv2 == 0x20
}

func GetTypeString(cards []byte) string {
	t, _ := GetCardTypeJH(cards)
	switch t {
	case CardTypeSingle:
		return "单张"
	case CardTypeDZ:
		return "对子"
	case CardTypeSZA23:
		return "顺子A23"
	case CardTypeSZ:
		return "顺子"
	case CardTypeJH:
		return "金花"
	case CardTypeSJA23:
		return "顺金A23"
	case CardTypeSJ:
		return "顺金"
	case CardTypeBZ:
		return "豹子"
	}

	return ""
}
