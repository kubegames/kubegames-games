package data

import (
	"fmt"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/poker"
)

//特殊牌型
const (
	SPECIAL_CARD_ZZQL = 13 //至尊青龙
	SPECIAL_CARD_YTL  = 12 //一条龙
	SPECIAL_CARD_SEHZ = 11 //十二皇族
	SPECIAL_CARD_STHS = 10 //三同花顺
	SPECIAL_CARD_SFTX = 9  //三分天下
	SPECIAL_CARD_QD   = 8  //全大
	SPECIAL_CARD_QX   = 7  //全小
	SPECIAL_CARD_CYS  = 6  //凑一色
	SPECIAL_CARD_STST = 5  //四套三条
	SPECIAL_CARD_WDST = 4  //五对三条
	SPECIAL_CARD_LDB  = 3  //六对半
	SPECIAL_CARD_SSZ  = 2  //三顺子
	SPECIAL_CARD_STH  = 1  //三同花
	SPECIAL_CARD_NO   = 0  //没有特殊牌型
)

func GetSpecialCnName(specialType int32) string {
	switch specialType {
	case SPECIAL_CARD_ZZQL:
		return "至尊青龙"
	case SPECIAL_CARD_YTL:
		return "一条龙"
	case SPECIAL_CARD_SEHZ:
		return "十二皇族"
	case SPECIAL_CARD_STHS:
		return "三同花顺"
	case SPECIAL_CARD_SFTX:
		return "三分天下"
	case SPECIAL_CARD_QD:
		return "全大"
	case SPECIAL_CARD_QX:
		return "全小"
	case SPECIAL_CARD_CYS:
		return "凑一色"
	case SPECIAL_CARD_STST:
		return "四套三条"
	case SPECIAL_CARD_WDST:
		return "五对三条"
	case SPECIAL_CARD_LDB:
		return "六对半"
	case SPECIAL_CARD_SSZ:
		return "三顺子"
	case SPECIAL_CARD_STH:
		return "三同花"
	}
	return "无"
}

const (
	HEAD = 1
	MID  = 2
	TAIL = 3
)

var SpecialScoreMap = make(map[int32]int)
var SpecialMidMap = make(map[int]int)  //中墩的特殊牌型加分
var SpecialTailMap = make(map[int]int) //尾墩的特殊牌型加分

func init() {
	SpecialScoreMap[SPECIAL_CARD_STH] = 3
	SpecialScoreMap[SPECIAL_CARD_SSZ] = 3
	SpecialScoreMap[SPECIAL_CARD_LDB] = 6
	SpecialScoreMap[SPECIAL_CARD_WDST] = 8
	SpecialScoreMap[SPECIAL_CARD_STST] = 8
	SpecialScoreMap[SPECIAL_CARD_CYS] = 10
	SpecialScoreMap[SPECIAL_CARD_QX] = 10
	SpecialScoreMap[SPECIAL_CARD_QD] = 10
	SpecialScoreMap[SPECIAL_CARD_SFTX] = 20
	SpecialScoreMap[SPECIAL_CARD_STHS] = 20
	SpecialScoreMap[SPECIAL_CARD_SEHZ] = 24
	SpecialScoreMap[SPECIAL_CARD_YTL] = 30
	SpecialScoreMap[SPECIAL_CARD_ZZQL] = 32

	SpecialMidMap[poker.CardTypeHL] = 2 //文档上写的2，但玩家赢了已经加了1分，所以是1，以下同理
	SpecialMidMap[poker.CardTypeFour] = 8
	SpecialMidMap[poker.CardTypeTHSA2345] = 10
	SpecialMidMap[poker.CardTypeTHS] = 10

	SpecialTailMap[poker.CardTypeFour] = 4
	SpecialTailMap[poker.CardTypeTHSA2345] = 5
	SpecialTailMap[poker.CardTypeTHS] = 5

}

//设置玩家的特殊牌型 如果是特殊牌型，会将玩家的牌分好
//该函数的性能最差时候为 2ms
func (user *User) SetSpecialCardType() (cards []byte, cardsArr [][]byte) {
	cards = make([]byte, 13)
	for k, v := range user.Cards {
		cards[k] = v
	}
	cards = poker.SortCards(cards)
	cardsArr = poker.GetCombineCardsArr(cards, 5)
	switch {
	case user.isZZQL(cards):
		user.SpecialCardType = SPECIAL_CARD_ZZQL
	case user.isYTL(cards):
		user.SpecialCardType = SPECIAL_CARD_YTL
	case user.isSEHZ(cards):
		user.SpecialCardType = SPECIAL_CARD_SEHZ
	case user.isSTHS(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_STHS
	case user.isSFTX(cards):
		user.SpecialCardType = SPECIAL_CARD_SFTX
	case user.isQD(cards):
		user.SpecialCardType = SPECIAL_CARD_QD
	case user.isQx(cards):
		user.SpecialCardType = SPECIAL_CARD_QX
	case user.isCYS(cards):
		user.SpecialCardType = SPECIAL_CARD_CYS
	case user.isSTST(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_STST
	case user.isWDST(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_WDST
	case user.isLDB(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_LDB
	case user.isSSZ(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_SSZ
	case user.isSTH(cards, cardsArr):
		user.SpecialCardType = SPECIAL_CARD_STH
	default:
		user.SpecialCardType = SPECIAL_CARD_NO
	}
	return
}

//是否为至尊青龙
func (user *User) isZZQL(cards []byte) bool {
	//0x20 0x30
	for i := 0; i < 12; i++ {
		cv, cc := poker.GetCardValueAndColor(cards[i])
		cv2, cc2 := poker.GetCardValueAndColor(cards[i+1])
		if cc != cc2 || cv-cv2 != 16 {
			return false
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为至一条龙
func (user *User) isYTL(cards []byte) bool {
	//0x20 0x30
	for i := 0; i < 12; i++ {
		cv, _ := poker.GetCardValueAndColor(cards[i])
		cv2, _ := poker.GetCardValueAndColor(cards[i+1])
		if cv-cv2 != 16 {
			return false
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为十二皇族
func (user *User) isSEHZ(cards []byte) bool {
	//0x20 0x30
	for i := 0; i < 13; i++ {
		cv, _ := poker.GetCardValueAndColor(cards[i])
		if cv != 0xb0 && cv != 0xc0 && cv != 0xd0 && cv != 0xe0 {
			return false
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为三同花顺
func (user *User) isSTHS(cards []byte, cardsArr [][]byte) bool {
	if user.isSSZ(cards, cardsArr) && user.isSTH(cards, cardsArr) {
		//headType,_ := poker.GetCardType13Water(user.HeadCards)
		midType, _ := poker.GetCardType13Water(user.MiddleCards)
		tailType, _ := poker.GetCardType13Water(user.TailCards)
		if (midType == poker.CardTypeTHS || midType == poker.CardTypeTHSA2345) && (tailType == poker.CardTypeTHS || tailType == poker.CardTypeTHSA2345) {
			return true
		}
	}
	return false
}

//是否为三分天下
func (user *User) isSFTX(cards []byte) bool {
	//不一样的牌只有4张
	cvArr := make([]byte, 0)
	for _, card := range cards {
		cv, _ := poker.GetCardValueAndColor(card)
		cvArr = append(cvArr, cv)
	}
	uniqueArr := RemoveRepeatedByte(cvArr)
	//fmt.Println("去重之后 ", fmt.Sprintf(`%x`, uniqueArr))
	if len(uniqueArr) == 4 {
		user.HeadCards = cards[10:]
		user.MiddleCards = cards[5:10]
		user.TailCards = cards[:5]
		return true
	}
	//fmt.Println("len(uniqueArr) ",len(uniqueArr))
	return false
}

//是否为全大
func (user *User) isQD(cards []byte) bool {
	for _, card := range user.Cards {
		cv, _ := poker.GetCardValueAndColor(card)
		if cv < 0x80 {
			return false
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为全小
func (user *User) isQx(cards []byte) bool {
	for _, card := range user.Cards {
		cv, _ := poker.GetCardValueAndColor(card)
		if cv > 0x80 {
			return false
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为凑一色 13张牌都是同一种颜色（红桃+方片、黑桃+梅花） 13 24
func (user *User) isCYS(cards []byte) bool {
	_, card0Color := poker.GetCardValueAndColor(user.Cards[0])
	if card0Color == 1 || card0Color == 3 {
		for _, card := range user.Cards {
			_, cc := poker.GetCardValueAndColor(card)
			if cc != 1 && cc != 3 {
				return false
			}
		}
	} else {
		for _, card := range user.Cards {
			_, cc := poker.GetCardValueAndColor(card)
			if cc != 2 && cc != 4 {
				return false
			}
		}
	}
	user.HeadCards = cards[10:]
	user.MiddleCards = cards[5:10]
	user.TailCards = cards[:5]
	return true
}

//是否为四套三条 13张牌由4个三条+一个单张组成
func (user *User) isSTST(cardsAll []byte, cardsArr [][]byte) bool {
	for _, cards := range cardsArr {
		cardType, _ := poker.GetCardType13Water(cards)
		if cardType == poker.CardTypeFour {
			return false
		}
	}

	//不一样的牌只有5张
	cvArr := make([]byte, 0)
	for _, card := range cardsAll {
		cv, _ := poker.GetCardValueAndColor(card)
		cvArr = append(cvArr, cv)
	}
	uniqueArr := RemoveRepeatedByte(cvArr)
	if len(uniqueArr) == 5 {
		for _, unique := range uniqueArr {
			//fmt.Println("unique : ",fmt.Sprintf(`%x`,unique))
			uniqueCv, _ := poker.GetCardValueAndColor(unique)
			eqCount := 0
			for _, card := range cardsAll {
				cv, _ := poker.GetCardValueAndColor(card)
				if uniqueCv == cv {
					eqCount++
				}
			}
			if eqCount != 3 && eqCount != 1 {
				return false
			}
		}
		user.HeadCards = cardsAll[10:]
		user.MiddleCards = cardsAll[5:10]
		user.TailCards = cardsAll[:5]
		return true
	}
	return false
}

//是否为五对三条 13张牌由5个对子+一个三条组成
func (user *User) isWDST(cardsAll []byte, cardsArr [][]byte) bool {
	cardsAll = poker.SortCards(cardsAll)
	for _, cards := range cardsArr {
		cardType, _ := poker.GetCardType13Water(cards)
		if cardType == poker.CardTypeFour {
			return false
		}
	}
	//不一样的牌只有6张
	cvArr := make([]byte, 0)
	for _, card := range cardsAll {
		cv, _ := poker.GetCardValueAndColor(card)
		cvArr = append(cvArr, cv)
	}
	uniqueArr := RemoveRepeatedByte(cvArr)
	if len(uniqueArr) == 6 {
		//再看里面是否存在单张的情况
		for _, unique := range uniqueArr {
			//fmt.Println("unique : ",fmt.Sprintf(`%x`,unique))
			uniqueCv, _ := poker.GetCardValueAndColor(unique)
			eqCount := 0
			for _, card := range cardsAll {
				cv, _ := poker.GetCardValueAndColor(card)
				if uniqueCv == cv {
					eqCount++
				}
			}
			if eqCount < 2 {
				return false
			}
		}
		user.HeadCards = cardsAll[10:]
		user.MiddleCards = cardsAll[5:10]
		user.TailCards = cardsAll[:5]
		return true
	}
	return false
}

//是否为六 对 半 13张牌由6个对子+一个单牌组成
func (user *User) isLDB(cardsAll []byte, cardsArr [][]byte) bool {
	for _, cards := range cardsArr {
		cardType, _ := poker.GetCardType13Water(cards)
		if cardType == poker.CardTypeFour || cardType == poker.CardTypeHL || cardType == poker.CardTypeTK {
			return false
		}
	}
	//不一样的牌只有7张
	cvArr := make([]byte, 0)
	for _, card := range cardsAll {
		cv, _ := poker.GetCardValueAndColor(card)
		cvArr = append(cvArr, cv)
	}
	uniqueArr := RemoveRepeatedByte(cvArr)
	if len(uniqueArr) == 7 {
		user.HeadCards = cardsAll[10:]
		user.MiddleCards = cardsAll[5:10]
		user.TailCards = cardsAll[:5]
		return true
	}
	return false
}

//是否为三 顺 子 头墩、中墩、尾墩都是顺子
func (user *User) isSSZ(cardsAll []byte, cardsArr [][]byte) bool {
	if len(cardsAll) != 13 {
		fmt.Println("len(cardsAll) != 13 isSSZ ")
		return false
	}
	//先找出13张牌中可组成的5张顺子可能
	szArr := user.GetSzArr(cardsArr)
	if len(szArr) == 0 {
		return false
	}
	//fmt.Println("三顺子第一手选出的顺子：", fmt.Sprintf(`%x`, szArr))
	leftHasSz := false
	var headCardsShow []byte
	var midCardsShow []byte
	var tailCardsShow []byte
	for _, szCards := range szArr {
		//再找出8张中的顺子
		left8Cards := user.GetDifferentCards(cardsAll, szCards)
		//fmt.Println("三顺子剩下的8张牌：", fmt.Sprintf(`%x`, left8Cards))
		//A2345 678 10jqkA
		leftSzArr := user.GetSzArr(poker.GetCombineCardsArr(left8Cards, 5))
		if len(leftSzArr) == 0 {
			continue
		}
		//fmt.Println("三顺子第二手选出的5张牌：", fmt.Sprintf(`%x`, leftSzArr))
		for _, leftSzCards := range leftSzArr {
			//再看剩余的三张是不是对子
			last3Cards := user.GetDifferentCards(left8Cards, leftSzCards)
			if poker.IsCard3TypeSZ(last3Cards) || poker.IsCard3TypeSZA23(last3Cards) {
				leftHasSz = true
				midCardsShow = szCards
				tailCardsShow = leftSzCards
				headCardsShow = last3Cards
				break
			}
		}
		if leftHasSz {
			break
		}
	}
	if leftHasSz {
		user.HeadCards = headCardsShow
		user.MiddleCards = midCardsShow
		user.TailCards = tailCardsShow
		if res, _, _ := user.Compare5Cards(poker.Cards5SliceToArr(user.MiddleCards), poker.Cards5SliceToArr(user.TailCards)); res == global.COMPARE_WIN {
			user.MiddleCards, user.TailCards = user.TailCards, user.MiddleCards
		}
		return true
	}
	return false

	//fmt.Println("三顺子剩下的3张牌：", fmt.Sprintf(`%x`, last3Cards))
}

// del by wd in 2020.3.9 获取可能的顺子组合 修复是同花顺而导致对顺子的判定错过
//获取可能的顺子组合
//func (user *User) GetSzArr(cardsArr [][]byte) (res [][]byte) {
//	res = make([][]byte, 0)
//	for _, cards := range cardsArr {
//		cardType, sortCards := poker.GetCardType13Water(cards)
//		if cardType == poker.CardTypeSZ || cardType == poker.CardTypeSZA2345 {
//			res = append(res, sortCards)
//		}
//	}
//	return
//}

// add by wd in 2020.3.9 获取可能的顺子组合 修复是同花顺而导致对顺子的判定错过
//获取可能的顺子组合
func (user *User) GetSzArr(cardsArr [][]byte) (res [][]byte) {
	res = make([][]byte, 0)
	for _, cards := range cardsArr {
		if poker.IsCardTypeSz(cards) || poker.IsCardTypeSZA2345(cards) {
			sortCards := poker.SortCards(cards)
			res = append(res, sortCards)
		}
	}
	return
}

//是否为三 同 花 头墩、中墩、尾墩都是同花
func (user *User) isSTH(cardsAll []byte, cardsArr [][]byte) bool {
	hasTh := false
	var first5Cards []byte
	midShow := make([]byte, 0)
	for _, cards := range cardsArr {
		cardType, sortRes := poker.GetCardType13Water(cards)
		if cardType == poker.CardTypeTH || cardType == poker.CardTypeTHS {
			hasTh = true
			midShow = sortRes
			first5Cards = sortRes
			break
		}
	}
	if !hasTh {
		return false
	}
	//if !user.User.IsRobot(){
	//	fmt.Println("midShow：", fmt.Sprintf(`%x`, midShow))
	//}
	tailShow := make([]byte, 0)
	//再找出8张中的同花
	left8Cards := user.GetDifferentCards(cardsAll, first5Cards)
	//fmt.Println("剩下的8张牌：", fmt.Sprintf(`%x`, left8Cards))
	//A2345 678 10jqkA
	leftCardsArr := poker.GetCombineCardsArr(left8Cards, 5)
	hasTh = false
	for _, cards := range leftCardsArr {
		cardType, sortRes2 := poker.GetCardType13Water(cards)
		if cardType == poker.CardTypeTH || cardType == poker.CardTypeTHS {
			hasTh = true
			tailShow = sortRes2
			first5Cards = sortRes2
			break
		}
	}
	if !hasTh {
		return false
	}

	//再看剩余的三张是不是同花
	last3Cards := user.GetDifferentCards(left8Cards, first5Cards)
	if poker.IsCard3TypeJH(last3Cards) {
		fmt.Println("三同花剩下的3张牌：", fmt.Sprintf(`%x`, last3Cards))
		user.HeadCards = last3Cards
		user.MiddleCards = midShow
		user.TailCards = tailShow
		if res, _, _ := user.Compare5Cards(poker.Cards5SliceToArr(user.MiddleCards), poker.Cards5SliceToArr(user.TailCards)); res == global.COMPARE_WIN {
			user.MiddleCards, user.TailCards = user.TailCards, user.MiddleCards
		}
		return true
	}
	return false
}

//找出cards1中不包含card2中的牌 eg: [3,4,5,6] [3,4,5] => 6
func (user *User) GetDifferentCards(cards1 []byte, cards2 []byte) []byte {
	//fmt.Println("1 : ",fmt.Sprintf(`%x`,cards1)," 2 : ",fmt.Sprintf(`%x`,cards2))
	lastCardsAll := make([]byte, 0)
	for _, c1 := range cards1 {
		if !IsByteInArr(c1, cards2) {
			lastCardsAll = append(lastCardsAll, c1)
		}
	}
	return lastCardsAll
}

//byte 元素是否存在于byteArr中
func IsByteInArr(b byte, arr []byte) bool {
	for _, v := range arr {
		if v == b {
			return true
		}
	}
	return false
}

//byteArr去重
func RemoveRepeatedByte(arr []byte) (newArr []byte) {
	newArr = make([]byte, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if !repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

// GetSpecialEncode 获取特殊牌编码
func GetSpecialEncode(cards []byte) (encode int64) {
	for i, card := range cards {
		fmt.Println((int64(card) >> 4) << uint((13-i-1)*4))
		encode |= (int64(card) >> 4) << uint((13-i-1)*4)
	}
	return
}
