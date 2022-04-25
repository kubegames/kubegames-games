package poker_test

import (
	"fmt"
	poker2 "game_MaJiang/erbagang/poker"
	"testing"
	"time"
)

//测试三个用户牌型比较
func TestCompareCards(t *testing.T) {
	zhuangcards := []byte{0x7, 0x7}
	xiancards := []byte{0xa, 0x9}
	res := poker2.GetCompareCardsRes(zhuangcards, xiancards)
	fmt.Println("zhuang,xiantype", res, zhuangcards, xiancards)

}
func TestCard(t *testing.T) {
	//var a byte = 0XE0
	//ca,cc := poker.GetCardValueAndColor(a)
	//fmt.Println(ca)
	//fmt.Println(fmt.Sprintf(`%x`,ca))
	//fmt.Println(fmt.Sprintf(`%X`,cc))
	//return

	cards := []byte{0xa, 0xa}
	ct, res := poker2.GetCardType(cards)
	fmt.Println(res)
	fmt.Println("card type : ", ct)
	//fmt.Println(fmt.Sprintf("%x", poker2.GetEncodeCard(ct, res)))
	cards1 := []byte{9, 9}
	ct1, res1 := poker2.GetCardType(cards1)
	fmt.Println(res1)
	fmt.Println("111card type : ", ct1)
	//fmt.Println(fmt.Sprintf("%x", poker2.GetEncodeCard(ct1, res1)))
}

////测试发牌
func TestSendCards(t *testing.T) {
	fmt.Println(poker2.GenerateCards())
	time.Sleep(time.Millisecond * 500)
	fmt.Println(poker2.GenerateCards())
	time.Sleep(time.Millisecond * 500)
}
