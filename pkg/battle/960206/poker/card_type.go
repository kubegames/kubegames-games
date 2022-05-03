package poker

import (
	"fmt"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

const (
	CardTypeTHS      = 11 //同花顺
	CardTypeTHSA2345 = 10
	CardTypeFour     = 9 //四张 1
	CardTypeHL       = 8 //葫芦 1
	CardTypeTH       = 7 //同花
	CardTypeSZ       = 6 //顺子
	CardTypeSZA2345  = 5 //
	CardTypeTK       = 4 //三条	//Three kind
	CardTypeTW       = 3 //两对	//Two pairs
	CardTypeDz       = 2 //对子
	CardTypeSingle   = 1 //单牌

	//三张牌的比较
	Card3TypeBz     = 3 //豹子
	Card3TypeDz     = 2 //对子
	Card3TypeSingle = 1 //单张
)

func GetPrintCards(cards []byte) string {
	return fmt.Sprintf(`%x`, cards)
}

//获取金花牌型
//将整个牌型进行sort，返回牌的类型和排序之后的具体值
//比如 3、8、3、3、9，则返回 葫芦(cardType)，33389(排序之后的牌面值)
func GetCardType13Water(cards []byte) (cardType int, sortRes []byte) {
	if len(cards) == 3 {
		return getCardType3(cards)
	}
	if len(cards) == 5 {
		return getCardType5(cards)
	}
	return
}

//头墩3张，获取牌型
func getCardType3(cards []byte) (cardType int, sortRes []byte) {
	sortRes = SortCards(cards)
	switch {
	case isCardTypeBZ(sortRes):
		cardType = Card3TypeBz
	case isCardTypeDZ(sortRes):
		cardType = Card3TypeDz
	default:
		cardType = Card3TypeSingle
	}
	return
}

//中墩和尾墩5张，获取牌型
func getCardType5(cards []byte) (cardType int, sortRes []byte) {
	sortRes = SortCards(cards)
	switch {
	case isCardTypeThs(cards):
		cardType = CardTypeTHS
	case isCardTypeThsA2345(cards):
		cardType = CardTypeTHSA2345
	case isCardTypeFour(cards):
		cardType = CardTypeFour
	case isCardTypeHl(cards):
		cardType = CardTypeHL
	case isCardTypeTH(cards):
		cardType = CardTypeTH
	case IsCardTypeSz(cards):
		cardType = CardTypeSZ
	case IsCardTypeSZA2345(cards):
		cardType = CardTypeSZA2345
	case isCardTypeTk(cards):
		cardType = CardTypeTK
	case isCardTypeTw(cards):
		cardType = CardTypeTW
	case isCardTypeDZ(cards):
		cardType = CardTypeDz
	default:
		cardType = CardTypeSingle
	}
	return
}

//判断牌型是否为豹子
func isCardTypeBZ(cards []byte) bool {
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return cv0 == cv1 && cv0 == cv2
}

//判断牌型是否为对子
func isCardTypeDZ(cards []byte) bool {
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

//获取对子牌型的对子的那张 eg: 223 返回 2
func GetDzCard(cards []byte) (byte, bool) {
	for i := 0; i < len(cards)-1; i++ {
		cvi, _ := GetCardValueAndColor(cards[i])
		for j := i + 1; j < len(cards); j++ {
			cvj, _ := GetCardValueAndColor(cards[j])
			if cvi == cvj {
				return cvi, true
			}
		}
	}
	return 0, false
}

//获取对子、三条中，单张的牌
func GetDzSingleCards(arr []byte) (newArr []byte) {
	newArr = make([]byte, 0)
	repeatedArr := make([]byte, 0)
	for i := 0; i < len(arr)-1; i++ {
		cvi, _ := GetCardValueAndColor(arr[i])
		if IsInByteArr(cvi, repeatedArr) {
			continue
		}
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			cvj, _ := GetCardValueAndColor(arr[j])
			if cvi == cvj {
				repeat = true
				repeatedArr = append(repeatedArr, cvi)
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	lastCv, _ := GetCardValueAndColor(arr[len(arr)-1])
	if !IsInByteArr(lastCv, repeatedArr) {
		newArr = append(newArr, arr[len(arr)-1])
	}
	return
}

func IsInByteArr(num byte, arr []byte) bool {
	//log.Traceln(fmt.Sprintf(`%x`,num)," arr : ",fmt.Sprintf(`%x`,arr))
	for _, v := range arr {
		if num == v {
			return true
		}
	}
	return false
}

//将牌进行排序 从大到小 0x40 0x30 0x20
func SortCards(cards []byte) []byte {
	for i := 0; i < len(cards)-1; i++ {
		for j := 0; j < (len(cards) - 1 - i); j++ {
			if (cards)[j] < (cards)[j+1] {
				cards[j], cards[j+1] = cards[j+1], cards[j]
			}
		}
	}
	return cards
}

////////-----------------以下是牌型----------------

//判断牌型是否为同花顺A2345 11
func isCardTypeThs(cards []byte) bool {
	return isCardTypeTH(cards) && IsCardTypeSz(cards)
}

//判断牌型是否为同花顺A2345 10
func isCardTypeThsA2345(cards []byte) bool {
	return isCardTypeTH(cards) && IsCardTypeSZA2345(cards)
}

//判断牌型是否为四条 9
func isCardTypeFour(cards []byte) bool {
	//判断方式：只有 32222和33332两种情况是
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	cv3, _ := GetCardValueAndColor(cards[3])
	cv4, _ := GetCardValueAndColor(cards[4])
	if (cv1 == cv2 && cv1 == cv3 && cv1 == cv4) ||
		(cv1 == cv2 && cv1 == cv3 && cv1 == cv0) {
		return true
	}
	return false
}

//判断牌型是否为葫芦 8
func isCardTypeHl(cards []byte) bool {
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	cv3, _ := GetCardValueAndColor(cards[3])
	cv4, _ := GetCardValueAndColor(cards[4])
	//判断方式：只有 33322和33222两种情况是
	if (cv0 == cv1 && cv0 == cv2 && cv3 == cv4) ||
		(cv0 == cv1 && cv2 == cv3 && cv2 == cv4) {
		return true
	}
	return false
}

//判断牌型是否为同花 7
func isCardTypeTH(cards []byte) bool {
	_, cc0 := GetCardValueAndColor(cards[0])
	for _, card := range cards {
		_, cc := GetCardValueAndColor(card)
		if cc != cc0 {
			return false
		}
	}
	return true
}

//判断牌型是否为顺子 6
func IsCardTypeSz(cards []byte) bool {
	//将牌进行大小排序
	cards = SortCards(cards)
	cv0, _ := GetCardValueAndColor(cards[len(cards)-1])
	//log.Traceln("cv0:: ",cv0)
	var value byte = 64
	for i := 0; i < len(cards)-1; i++ {
		cv, _ := GetCardValueAndColor(cards[i])
		//log.Traceln("cv",i," : ",cv)
		if (cv - cv0) != value {
			//log.Traceln(" cv : ",cv, " cv0 : " ,cv0," value : ",value)
			//比如1、2、3、4、5，2-1等于1，3-1等于2，以此类推
			return false
		}
		value -= 16
	}
	return true
}

//判断牌型是否为顺子 A2345 5
func IsCardTypeSZA2345(cards []byte) bool {
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	cv3, _ := GetCardValueAndColor(cards[3])
	cv4, _ := GetCardValueAndColor(cards[4])
	//fmt.Sprintf(`%x ,%x ,%x ,%x ,%x `, cv0, cv1, cv2, cv3, cv4)
	return cv0 == 0xe0 && cv1 == 0x50 && cv2 == 0x40 && cv3 == 0x30 && cv4 == 0x20
}

//判断牌型是否为三条 4
func isCardTypeTk(cards []byte) bool {
	//判断方式：只有 33321 32221 32111三种情况 并且先判断是葫芦，再判断3条 才会没问题
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	cv3, _ := GetCardValueAndColor(cards[3])
	cv4, _ := GetCardValueAndColor(cards[4])
	if (cv0 == cv1 && cv0 == cv2) ||
		(cv1 == cv2 && cv1 == cv3) ||
		(cv2 == cv3 && cv2 == cv4) {
		return true
	}
	return false
}

//判断牌型是否为两对 3
func isCardTypeTw(cards []byte) bool {
	//判断方式：只有 33221 33211 32211 三种情况
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	cv3, _ := GetCardValueAndColor(cards[3])
	cv4, _ := GetCardValueAndColor(cards[4])
	if (cv0 == cv1 && cv2 == cv3) ||
		(cv0 == cv1 && cv3 == cv4) ||
		(cv1 == cv2 && cv3 == cv4) {
		return true
	}
	return false
}

//判断牌型是否为顺子
func IsCard3TypeSZ(cards []byte) bool {
	if len(cards) != 3 {
		log.Traceln("len(cards) != 3 ", fmt.Sprintf(`%x`, cards))
		return false
	}
	cards = SortCards(cards)
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	return (cv0-cv1) == 16 && (cv1-cv2) == 16
}

//判断牌型是否为顺子 A23
func IsCard3TypeSZA23(cards []byte) bool {
	cards = SortCards(cards)
	cv0, _ := GetCardValueAndColor(cards[0])
	cv1, _ := GetCardValueAndColor(cards[1])
	cv2, _ := GetCardValueAndColor(cards[2])
	//log.Traceln(fmt.Sprintf(`%x %x %x`,cv0,cv1,cv2))
	return cv0 == 0xe0 && cv1 == 0x30 && cv2 == 0x20
}

//判断牌型是否为金花
func IsCard3TypeJH(cards []byte) bool {
	_, cc0 := GetCardValueAndColor(cards[0])
	_, cc1 := GetCardValueAndColor(cards[1])
	_, cc2 := GetCardValueAndColor(cards[2])
	return cc0 == cc1 && cc0 == cc2
}
