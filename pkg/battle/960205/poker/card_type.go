package poker

import "github.com/kubegames/kubegames-sdk/pkg/log"

var (
	CardTypeTW     = 4 //天王
	CardTypeBao    = 3 //宝
	CardType28     = 2 //28杠
	CardTypeOthers = 1 //普通牌
)

//获取28杠牌型
//将整个牌型进行sort，返回牌的类型和排序之后的具体值
//比如 3、8、3、3、9，则返回 葫芦(cardType)，33389(排序之后的牌面值)
func GetCardType(cards []byte) (cardType int, sortRes []byte) {
	if len(cards) != 2 {
		log.Traceln("牌型比较只能2张牌")
		return
	}
	sortRes = sortCards(cards)
	//从从大到小

	//天王 4
	if isCardTypeTW(cards) {
		cardType = CardTypeTW
		return
	}
	// bao 3
	if isCardTypeBao(cards) {
		cardType = CardTypeBao
		return
	}

	//28 杠 2
	if isCardType28(cards) {
		cardType = CardType28
		return
	}

	//其他牌 1
	cardType = CardTypeOthers
	return
}

//判断牌型是否为天王
func isCardTypeTW(cards []byte) bool {
	cv0 := cards[0]
	cv1 := cards[1]
	return cv0 == 10 && cv1 == 10
}

//判断牌型是否为对子
func isCardTypeBao(cards []byte) bool {
	cv0 := cards[0]
	cv1 := cards[1]
	return cv0 == cv1
}

//判断牌型是否为28杠
func isCardType28(cards []byte) bool {
	cv0 := cards[0]
	cv1 := cards[1]

	return cv0 == 8 && cv1 == 2
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

//比牌结果1 庄赢 0闲赢
func GetCompareCardsRes(bank []byte, xian []byte) int {
	bankType, bankSortRes := GetCardType(bank)
	xianType, xianSortRes := GetCardType(xian)
	//log.Traceln(bankType, bankSortRes, xianType, xianSortRes)
	if bankType > xianType {
		return 1
	} else if bankType == xianType {
		if bankType == CardTypeBao {
			if bankSortRes[0] >= xianSortRes[0] {
				return 1
			}
			return 0
		}
		return compareSameTypeCards(bankSortRes, xianSortRes)
	} else {
		return 0
	}
}
func compareSameTypeCards(bank []byte, xian []byte) int {
	bankSum := (bank[0] + bank[1])% 10
	xianSum := (xian[0] + xian[1])% 10
	if bankSum  > xianSum {
		return 1
	} else if bankSum == xianSum {
		if bank[0] >= xian[0] {
			return 1
		} else {
			return 0
		}

	} else {
		return 0
	}

}
