package poker

import "fmt"

//随便输入一个数字，得出它的牌值和花色
func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = value & 240 //byte的高4位总和是240
	cardColor = value & 15  //byte的低4位总和是15
	return
}

//将牌 进行牌型编码并返回
func GetEncodeCard(cardType int, cards []byte) (cardEncode int) {

	cardEncode = (cardType) << 20
	for i, card := range cards {
		cardEncode |= (int(card) >> 4) << uint((5-i-1)*4)
	}
	return
}

//获取手中牌的所有组合
//cards 是手中所有的牌，比如13张牌
//num 选出num张，比如选出5张牌
func GetCombineCardsArr(cards []byte, num int) [][]byte {
	index := zuheResult(len(cards), num)
	result := findNumsByIndexs(cards, index)
	return result
}

//删除byte数组中的某个元素
func DelByteSlice(arr []byte, elem byte) []byte {
	for i, v := range arr {
		if v == elem {
			return append(arr[:i], arr[i+1:]...)
		}
	}
	return arr
}

func GetCardsCNName(cards []byte) (name string) {
	for _, v := range cards {
		name += getCardCNName(v)
	}
	return
}

//获取牌的中文名
func getCardCNName(card byte) string {
	switch card {
	case 0x21:
		return "方块2"
	case 0x22:
		return "梅花2"
	case 0x23:
		return "红桃2"
	case 0x24:
		return "黑桃2"
	case 0x31:
		return "方块3"
	case 0x32:
		return "梅花3"
	case 0x33:
		return "红桃3"
	case 0x34:
		return "黑桃3"
	case 0x41:
		return "方块4"
	case 0x42:
		return "梅花4"
	case 0x43:
		return "红桃4"
	case 0x44:
		return "黑桃4"
	case 0x51:
		return "方块5"
	case 0x52:
		return "梅花5"
	case 0x53:
		return "红桃5"
	case 0x54:
		return "黑桃5"
	case 0x61:
		return "方块6"
	case 0x62:
		return "梅花6"
	case 0x63:
		return "红桃6"
	case 0x64:
		return "黑桃6"
	case 0x71:
		return "方块7"
	case 0x72:
		return "梅花7"
	case 0x73:
		return "红桃7"
	case 0x74:
		return "黑桃7"
	case 0x81:
		return "方块8"
	case 0x82:
		return "梅花8"
	case 0x83:
		return "红桃8"
	case 0x84:
		return "黑桃8"
	case 0x91:
		return "方块9"
	case 0x92:
		return "梅花9"
	case 0x93:
		return "红桃9"
	case 0x94:
		return "黑桃9"
	case 0xa1:
		return "方块10"
	case 0xa2:
		return "梅花10"
	case 0xa3:
		return "红桃10"
	case 0xa4:
		return "黑桃10"
	case 0xb1:
		return "方块J"
	case 0xb2:
		return "梅花J"
	case 0xb3:
		return "红桃J"
	case 0xb4:
		return "黑桃J"
	case 0xc1:
		return "方块Q"
	case 0xc2:
		return "梅花Q"
	case 0xc3:
		return "红桃Q"
	case 0xc4:
		return "黑桃Q"
	case 0xd1:
		return "方块K"
	case 0xd2:
		return "梅花K"
	case 0xd3:
		return "红桃K"
	case 0xd4:
		return "黑桃K"
	case 0xe1:
		return "方块A"
	case 0xe2:
		return "梅花A"
	case 0xe3:
		return "红桃A"
	case 0xe4:
		return "黑桃A"
	}
	return "错误牌型"
}

func GetCard3TypeCNName(cardType int) string {
	switch cardType {
	case Card3TypeBz:
		return "豹子"
	case Card3TypeDz:
		return "对子"
	case Card3TypeSingle:
		return "单张"
	}
	return "错误牌型"
}

func GetCard5TypeCNName(cardType int) string {
	switch cardType {
	case CardTypeTHS:
		return "同花顺"
	case CardTypeTHSA2345:
		return "同花顺"
	case CardTypeFour:
		return "四条"
	case CardTypeHL:
		return "葫芦"
	case CardTypeTH:
		return "同花"
	case CardTypeSZ:
		return "顺子"
	case CardTypeSZA2345:
		return "顺子"
	case CardTypeTK:
		return "三条"
	case CardTypeTW:
		return "两对"
	case CardTypeDz:
		return "对子"
	case CardTypeSingle:
		return "单张"
	}
	return "错误牌型"
}

func Cards5SliceToArr(cards []byte) [5]byte {
	arr := [5]byte{cards[0], cards[1], cards[2], cards[3], cards[4]}
	return arr
}
func Cards3SliceToArr(cards []byte) [3]byte {
	arr := [3]byte{cards[0], cards[1], cards[2]}
	return arr
}
func Cards5ArrToSlice(cardsArr [5]byte) []byte {
	slice := make([]byte, 5)
	slice[0] = cardsArr[0]
	slice[1] = cardsArr[1]
	slice[2] = cardsArr[2]
	slice[3] = cardsArr[3]
	slice[4] = cardsArr[4]
	return slice
}
func Cards3ArrToSlice(cardsArr [3]byte) []byte {
	slice := make([]byte, 3)
	slice[0] = cardsArr[0]
	slice[1] = cardsArr[1]
	slice[2] = cardsArr[2]
	return slice
}

func PrintCards(cards []byte) {
	fmt.Println(fmt.Sprintf(`%x`, cards))
}
