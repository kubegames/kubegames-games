package poker

import "github.com/kubegames/kubegames-sdk/pkg/log"

var (
	CardTypeBZ     = 8 //豹子
	CardTypeSJ     = 7 //顺金
	CardTypeSJ123  = 6 //顺金123
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
		log.Traceln("金花牌型只能3张牌,现在有：", cards)
		return
	}
	sortRes = sortCards(cards)
	//从从大到小

	//豹子 8
	if isCardTypeBZ(cards) {
		cardType = CardTypeBZ
		return
	}
	//顺金 7
	if isCardTypeSJ(cards, true) {
		cardType = CardTypeSJ
		return
	}
	//顺金123 6
	if isCardTypeSJ123(cards, true) {
		cardType = CardTypeSJ123
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

//判断牌型是否为顺金123
func isCardTypeSJ123(cards []byte, isSort bool) bool {
	if !isSort {
		cards = sortCards(cards)
	}

	_, cc0 := GetCardValueAndColor(cards[0])
	_, cc1 := GetCardValueAndColor(cards[1])
	_, cc2 := GetCardValueAndColor(cards[2])
	if cc0 != cc1 || cc0 != cc2 {
		return false
	}
	return isCardTypeSZA23(cards, isSort)
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
	//log.Traceln(fmt.Sprintf(`%x %x %x`,cv0,cv1,cv2))
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

//将几组牌从大到小排序
func SortCardsArrFromBig(cardsArr [][]byte) [][]byte {
	for i := 0; i < len(cardsArr)-1; i++ {
		for j := 0; j < (len(cardsArr) - 1 - i); j++ {
			if !CompareZjhCards(cardsArr[j], cardsArr[j+1]) {
				cardsArr[j], cardsArr[j+1] = cardsArr[j+1], cardsArr[j]
			}
		}
	}
	return cardsArr
}

func IsCardsRepeat(cards []byte, cardsArr [][]byte) bool {
	for _, v := range cardsArr {
		for i := 0; i < len(cards); i++ {
			c1, _ := GetCardValueAndColor(cards[i])
			c2, _ := GetCardValueAndColor(v[i])
			if c1 != c2 {
				return false
			}
			return true
		}
	}
	return false
}
