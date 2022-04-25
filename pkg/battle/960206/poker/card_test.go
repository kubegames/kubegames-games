package poker_test

import (
	"common/test/poker"
	"container/list"
	"fmt"
	poker3 "game_poker/13water/poker"
	"game_poker/zjh/data"
	poker2 "game_poker/zjh/poker"
	"testing"
)

//测试三个用户牌型比较
func TestCompareCards(t *testing.T) {

	userList := make([]*poker.UserCards, 0)
	//获取要比较的用户的牌型
	user1 := &poker.UserCards{
		Uid: 1, Cards: []byte{0x61, 0x81, 0x31},
	}
	user2 := &poker.UserCards{
		Uid: 2, Cards: []byte{0x72, 0x22, 0x32},
	}
	user3 := &poker.UserCards{
		Uid: 3, Cards: []byte{0x12, 0x22, 0x33},
	}
	userList = append(userList, user1)
	userList = append(userList, user2)
	userList = append(userList, user3)
	ul := poker.GetMaxUser(userList)
	for _, v := range ul {
		fmt.Println("获取的ul： ", v)
	}
}

func TestCard(t *testing.T) {
	//cardsType ,cards := poker3.GetCardType13Water([]byte{0xe3,0x22,0x32,0x42,0x52}) //12345
	//cardsType ,cards := poker3.GetCardType13Water([]byte{0xe3,0xe2,0xe2,0x42,0x52}) //三条
	//cardsType ,cards := poker3.GetCardType13Water([]byte{0xb3,0xb2,0xe2,0x52,0x52}) //两队
	cardsType1 ,cards1 := poker3.GetCardType13Water([]byte{0xb3,0xb2,0xe2,0x42,0x52}) //对子
	cardsType ,cards := poker3.GetCardType13Water([]byte{0xa3,0xb2,0xe2,0x42,0x52}) //单张
	code1,code2 := poker3.GetEncodeCard(cardsType1,cards1),poker3.GetEncodeCard(cardsType,cards)
	fmt.Println(code1,code2)
	//fmt.Println(cardsType," ",fmt.Sprintf(`0x%x 0x%x 0x%x 0x%x 0x%x`,cards[0],cards[1],cards[2],cards[3],cards[4]))
}

//测试发牌
func TestSendCards(t *testing.T) {
	for i:=0;i<100;i++ {
		a := 0
		b := 0
		a += b
		switch  {
		case a < -1:
			fmt.Println("case a < 1")
		case a < 2:
			fmt.Println("case a < 2")
		}
	}
}

func TestCompare(t *testing.T) {
	user1 := &data.User{Id: 1, Cards: []byte{0x11, 0x22, 0x32}, Name: "用户1"}
	user2 := &data.User{Id: 2, Cards: []byte{0x61, 0x32, 0x62}, Name: "用户2"}
	userList := make([]*data.User, 0)
	userList = append(userList, user1)
	userList = append(userList, user2)
	fmt.Println(poker2.GetMaxUser(userList)[0])
}

func TestMagu(t *testing.T) {
	ergodic([]byte{0x31, 0x51, 0x61, 0x71})
}

func ergodic(arr []byte) {
	for i := 0; i < len(arr)-2; i++ {
		for j := i + 1; j < len(arr)-1; j++ {
			for k := j + 1; k < len(arr); k++ {
				fmt.Println(arr[i], arr[j], arr[k])
			}
		}
	}
}




//从13张牌中选出5张最大的牌
func TestChooseMax(t *testing.T) {
	testList := list.New()
	for i:=0;i<1000;i++{
		testList.PushBack(i)
	}
	fmt.Println(testList.Len())
	for e := testList.Front(); e != nil; e = e.Next() {
		fmt.Println(e.Value)
		//testList.Remove(e)
	}
	fmt.Println(testList.Len())
}

func GetScoreStr(score int64) string {
	yuan := score / 100
	remain := score % 100
	jiao := remain / 10
	fen := remain % 10
	return fmt.Sprintf(`%d元%d角%d分`,yuan,jiao,fen)
}

















