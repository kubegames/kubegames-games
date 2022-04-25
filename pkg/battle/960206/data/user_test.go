package data

import (
	"fmt"
	"game_poker/13water/poker"
	"testing"
	"time"
)

//测试用户的特殊牌型

func TestUser_ZZQL(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//一条龙 12
func TestUser_YTL(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x21, 0x31, 0x41, 0x51, 0x62, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//十二皇族 11
func TestUser_SEHZ(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0xd2, 0xd1, 0xd1, 0xd1, 0xb2, 0xb1, 0xb1, 0xb1, 0xc1, 0xc1, 0xc1, 0xc1, 0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//三同花顺
func TestUser_STHS(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x83,0x93,0xa3,0x74,0x84,0x94,0xa4,0xb4,0xa1,0xb1,0xc1,0xd1,0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//三分天下 9
func TestUser_SFTX(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0xd2, 0xd1, 0xd3, 0xd4, 0x91, 0x92, 0x93, 0x94, 0x21, 0x22, 0x23, 0x24, 0x61},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//全大 8
func TestUser_QD(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0xd2, 0xd1, 0xd3, 0xd4, 0x91, 0x92, 0x93, 0x94, 0xa1, 0xa2, 0xb3, 0xb4, 0xc1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//全小 7
func TestUser_QX(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x72, 0x71, 0x73, 0x74, 0x61, 0x62, 0x63, 0x64, 0x21, 0x22, 0x33, 0x34, 0x41},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//凑一色 6
func TestUser_CYS(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x72, 0x74, 0xa2, 0xa4, 0x64, 0x62, 0xb2, 0xb4, 0x52, 0x54, 0xe2, 0xe4, 0x82},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//四套三条 5
func TestUser_STST(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x71,0x72,0x81,0x82,0x84,0xa1,0xa2,0xa4,0xc1,0xc4,0xd1,0xd3,0xd4},
		//Cards: [13]byte{0x22, 0x24, 0x23, 0x84, 0x81, 0x92, 0x91, 0x94, 0xc2, 0xc1, 0xc3, 0xe4, 0xe1},
		//Cards: [13]byte{0x72, 0x74, 0x73, 0xa4, 0x64, 0x62, 0x61, 0xb4, 0xb2, 0xb1, 0xe2, 0xe4, 0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//五对三条 4
func TestUser_WDST(t *testing.T) {
	t1 := time.Now()
	user := &User{
		//Cards: [13]byte{0x22, 0x24, 0x23, 0x94, 0x54, 0x92, 0x41, 0x93, 0xb2, 0xb1, 0xc2, 0xc4, 0xc1},
		Cards: [13]byte{0x22, 0x31, 0x32, 0x84, 0x91, 0x92, 0xc1, 0xc2, 0x71, 0x72, 0x74, 0x81, 0x83},
		//Cards: [13]byte{0x72, 0x74, 0xa3, 0xa4, 0x64, 0x62, 0x81, 0x84, 0xb2, 0xb1, 0xe2, 0xe4, 0xe1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//六对半 3
func TestUser_LDB(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x72, 0x74, 0xa3, 0xa4, 0x64, 0x62, 0x81, 0x84, 0xb2, 0xb1, 0xe2, 0xe4, 0x21},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//三顺子 2
func TestUser_SSZ(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x72, 0x64, 0x53, 0x44, 0x34, 0xc2, 0xd1, 0xb4, 0xa2, 0x91, 0xe1, 0xd4, 0xc1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//三同花 1
func TestUser_STH(t *testing.T) {
	t1 := time.Now()
	user := &User{
		Cards: [13]byte{0x72, 0x62, 0x52, 0x42, 0x32, 0xc2, 0xd2, 0xe2, 0xa2, 0x92, 0xe1, 0xd1, 0xc1},
	}
	user.SetSpecialCardType()
	fmt.Println("头墩：", fmt.Sprintf(`%x`, user.HeadCards))
	fmt.Println("头墩：", fmt.Sprintf(`%x`, user.MiddleCards))
	fmt.Println("头墩：", fmt.Sprintf(`%x`, user.TailCards))
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

//没有特殊牌 0
func TestUser_NO(t *testing.T) {
	t1 := time.Now()

	user := &User{
		Cards: [13]byte{0x73, 0x62, 0x52, 0x42, 0x32, 0xc2, 0xd2, 0xe2, 0xa2, 0x92, 0xe1, 0xd1, 0xc1},
	}
	user.SetSpecialCardType()
	fmt.Println(user.SpecialCardType, "时间：", time.Now().Sub(t1))
}

func TestUser_Card(t *testing.T) {
	user := &User{}
	//cards1 := []byte{0x21,0x23,0x83,0x84,0xd3}
	//cards2 := []byte{0x52,0x53,0x74,0xc2,0xc4}
	//fmt.Println(user.Compare5Cards(cards1,cards2))
	cards1 := []byte{0x43,0x41,0x82,0xb2,0xb1}
	cards2 := []byte{0xd1,0xd2,0xd3}
	fmt.Println(user.Compare5And3Cards(cards1,cards2))
	return
}

//type Compare struct {
//	headCards []byte
//	midCards  []byte
//	tailCards []byte
//	score     int
//}

//没有特殊牌型分牌
func TestUser_SplitCards(t *testing.T) {
	//arr1 := []byte{1,2}
	//arr2 := append(arr1,[]byte{3,4}...)
	//fmt.Println(arr1)
	//fmt.Println(arr2)
	//return
	user := &User{
		Cards: [13]byte{0x34,0xa4,0x81,0x84,0xe2,0x41,0x51,0x63,0x71,0x72,0x73,0x74,0x91},	//有问题牌型
		//Cards: [13]byte{0x21, 0x31, 0x41, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa1, 0xb1, 0xc1, 0xd1, 0xe1},	//有问题牌型
		//Cards: [13]byte{0xe2, 0xd1, 0xc4, 0xb1, 0xa1, 0x91, 0x84, 0x72, 0x61, 0x51, 0x41, 0x31, 0x21},	//特殊牌型
		//Cards: [13]byte{0xc2, 0xc1, 0xa4, 0xa1, 0x71, 0x91, 0x84, 0xb1, 0x61, 0x51, 0x41, 0x31, 0x21},	//同花顺
		//Cards: [13]byte{0xc2, 0xc1, 0xa4, 0xa1, 0x61, 0x91, 0x84, 0xb1, 0x72, 0x23, 0x22, 0x24, 0x21},	//四条
		//Cards: [13]byte{0xc2, 0x61, 0xa4, 0xa1, 0xc1, 0x91, 0x84, 0xb1, 0x42, 0x83, 0x82, 0xc4, 0xb4},
		//Cards: [13]byte{0xd3,0xb4,0x81,0xa3,0x72,0x92,0x83,0x52,0x22,0x33,0x32,0x73,0x24},
		//Cards: [13]byte{0xd3,0xc3,0x81,0xe4,0xa3,0x83,0x72,0x61,0xb1,0x41,0x34,0x31,0x21},
		//Cards: [13]byte{0x33,0xb3,0xb2,0x91,0xa4,0x23,0xa1,0x61,0x63,0xd1,0xd4,0xd2,0x34},
		//Cards: [13]byte{0xe4,0xe3,0xb3,0x53,0x51,0x31,0x23,0x21,0xc4,0xc2,0xa3,0xa2,0xa1},
		//Cards: [13]byte{0x62,0x51,0x43,0xe2,0xe1,0xd4,0xb1,0x84,0xd2,0xc3,0xc2,0x23,0x21},
	}
	t2 := time.Now()
	user.SplitCards()
	fmt.Println("spare len : ",len(user.SpareArr))
	user.PrintUserSpare()
	fmt.Println(time.Now().Sub(t2))
	return

}

func TestGetMaxCompareCards(t *testing.T) {
	maxCompareList := make([]*Compare, 2)
	maxCompareList[0] = &Compare{
		headCards: []byte{0x62, 0x51, 0x43},
		midCards:  []byte{0xe2, 0xe1, 0xd4, 0xb1, 0x84},
		tailCards: []byte{0xd2, 0xc3, 0xc2, 0x23, 0x21},
	}
	maxCompareList[1] = &Compare{
		headCards: []byte{0xd4, 0xd2, 0x43},
		midCards:  []byte{0xe2, 0xe1, 0xb1, 0x84, 0x62},
		tailCards: []byte{0xc3, 0xc2, 0x51, 0x23, 0x21},
	}
	user := new(User)
	fmt.Println(user.BeatCompare(maxCompareList[1], maxCompareList[0]))
	finalMax := user.GetMaxCompareCards(maxCompareList)
	fmt.Println("最终得到的牌：", fmt.Sprintf(`%x,%x,%x`, finalMax.headCards, finalMax.midCards, finalMax.tailCards))
}

func TestBeatCompare(t *testing.T) {
	user := new(User)
	c1 := &Compare{
		headCards: []byte{0xd3, 0xc3, 0x81},
		midCards:  []byte{0xe4, 0xa3, 0x83, 0x72, 0x61},
		tailCards: []byte{0xb1, 0x41, 0x34, 0x31, 0x21},
	}
	c2 := &Compare{
		headCards: []byte{0x83, 0x72, 0x34},
		midCards:  []byte{0xe4, 0xd3, 0xc3, 0xb1, 0xa3},
		tailCards: []byte{0x81, 0x61, 0x41, 0x31, 0x21},
	}
	maxCompare := c1
	if !user.BeatCompare(c1, c2) {
		maxCompare = c2
	}

	fmt.Println(fmt.Sprintf(`%x,%x,%x`, maxCompare.headCards, maxCompare.midCards, maxCompare.tailCards))
}

func TestGetNotDz(t *testing.T) {
	user := new(User)
	cards5 := []byte{0xe2, 0xe1, 0xb1, 0x84, 0x62}
	cards3 := []byte{0xd4, 0xd2, 0x43}
	fmt.Println(user.Compare5And3Cards(cards5, cards3))
}

func TestCompare5And3Cards(t *testing.T) {

	user := new(User)
	cards1 := [5]byte{0x21,0x31,0x41,0x51,0xc1}
	//card ,_ := poker.GetDzCard(cards1)
	//fmt.Println(fmt.Sprintf(`%x`,card))
	//return
	//
	//fmt.Println(fmt.Sprintf(`%x`,user.GetTwoDzMax(cards1)))
	//return
	cards2 := [5]byte{0x61,0x71,0x81,0x91,0xb1}
	a, b, c := user.Compare5Cards(cards1, cards2)
	fmt.Println(a, b, c)
}

func TestCompare3Cards(t *testing.T) {
	t1 := time.Now()
	slice := make([]int,100000)
	for i:=0;i<100000;i++{
		slice[i] = i
	}
	fmt.Println("slice 时间：",time.Now().Sub(t1))
	t2 := time.Now()
	arr := [100000]int{}
	for i:=0;i<100000;i++{
		arr[i] = i
	}
	fmt.Println("arr 时间：",time.Now().Sub(t2))
	return

	user := &User{}
	cards := user.SortCardsSelf([]byte{0x93,0x83,0x91,0x43,0x94},poker.CardTypeTK)
	fmt.Println(fmt.Sprintf(`%x`,cards))
}

func TestExchange(t *testing.T) {
	user1 := &User{
		Cards:           [13]byte{0xc2, 0x61, 0xa4, 0xa1, 0xc1, 0x91, 0x84, 0xb1, 0x42, 0x83, 0x82, 0xc4, 0xb4},
		HeadCards:       []byte{0x11,0x11,0x11},
		MiddleCards:     []byte{0x11,0x11,0x11},
		TailCards:       []byte{0x11,0x11,0x11},
		HeadCardType:    0,
		MidCardType:     0,
		TailCardType:    0,
		EncodeHead:      0,
		EncodeMid:       0,
		EncodeTail:      0,
		SpecialHead:     0,
		SpecialMid:      0,
		SpecialTail:     0,
		SpecialCardType: 0,
	}
	user2 := &User{
		Cards:           [13]byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x42, 0x83, 0x82, 0xc4, 0xb4},
		HeadCards:       []byte{0x22,0x22,0x22},
		MiddleCards:     []byte{0x22,0x22,0x22},
		TailCards:       []byte{0x22,0x22,0x22},
		HeadCardType:    1,
		MidCardType:     1,
		TailCardType:    1,
		EncodeHead:      1,
		EncodeMid:       1,
		EncodeTail:      1,
		SpecialHead:     1,
		SpecialMid:      1,
		SpecialTail:     1,
		SpecialCardType: 1,
	}
	fmt.Println("换牌钱：user1 user2: ",fmt.Sprintf(`%+v  %+v`,user1,user2))
	user1.ExchangeEachOther(user2)
	fmt.Println("换牌后：user1 user2: ",fmt.Sprintf(`%+v  %+v`,user1,user2))
}

func TestNewUser1(t *testing.T) {
	cards := []byte{
		0x54, 0x64, 0x74, 0x51, 0x61, 0x71, 0x81, 0x91, 0xa2, 0xb2, 0xc2, 0xd2, 0xe2,
	}
	cardsArr := poker.GetCombineCardsArr(cards, 5)
	t.Log(len(cardsArr))

	user := &User{}
	if user.isSSZ(cards, cardsArr){
		t.Log("是三顺子")
	}else {
		t.Log("不是三顺子")
	}
}

func TestNewUser2(t *testing.T) {
	cards := []byte{
		145, 129, 116, 100, 81,
	}
	cardType, sortCards := poker.GetCardType13Water(cards)
	if cardType == poker.CardTypeSZ || cardType == poker.CardTypeSZA2345 {
		t.Log("是顺子")
	}else {
		t.Logf("不是顺子，%v",sortCards)
	}


}