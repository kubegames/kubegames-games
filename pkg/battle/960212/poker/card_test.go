package poker

import (
	"common/rand"
	"game_poker/doudizhu/msg"
	"testing"
)

const MaxTurnCount = 1000000

var ()

// 测试检查手牌重复牌
func TestCheckRepeatedCards(t *testing.T) {

	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		repeatedCards := CheckRepeatedCards(gamePoker.Cards)
		t.Log(repeatedCards)

		if count >= MaxTurnCount {
			break
		}

		count++
	}

}

// 测试获取手牌牌型
func TestGetCardsType(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		len := rand.RandInt(1, 18)
		//if count%2 != 0 {
		//	len--
		//}

		cards := gamePoker.Cards[:len]
		cardsType := GetCardsType(cards)

		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)
		typeString := CardsTypeToString(cardsType)

		if cardsType > msg.CardsType_Pair {
			t.Logf("手牌: %v,长度: %v, 牌型: %s", transCards, len, typeString)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}

}

func TestGetCardsType2(t *testing.T) {
	cards := []byte{0x11, 0x12, 0x13, 0x21, 0x22, 0x23, 0x31, 0x32}
	cardsType := GetCardsType(cards)
	t.Log(cardsType)
}

// 测试获取一张牌的牌值和花色
func TestGetCardValueAndColor(t *testing.T) {
	value, color := GetCardValueAndColor(0xf5)
	t.Log(value)
	t.Log(color)
}

// 测试 检查2 或 大小王的
func TestHaveBigCard(t *testing.T) {
	cards := []byte{
		0xc1, 0xf2,
	}
	ok := HaveBigCard(cards)
	t.Log(ok)
}

// 测试 检查有火箭
func TestHaveRocket(t *testing.T) {
	cards := []byte{
		0xc1, 0xf1, 0xd2,
	}
	ok := HaveRocket(cards)
	t.Log(ok)
}

func TestIsSerialTripletWithWing2(t *testing.T) {
	cards := []byte{0x11, 0x12, 0x13, 0x21, 0x22, 0x23, 0x31, 0x32}
	IsSerialTripletWithWing(cards)
}

// 测试比牌
func TestContrastCards(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		len1 := rand.RandInt(3, 18)
		len2 := len1
		cards1, cards2 := gamePoker.Cards[:len1], gamePoker.Cards[len1:len2+len1]
		cardsType1, cardsType2 := GetCardsType(cards1), GetCardsType(cards2)
		if cardsType1 != cardsType2 || cardsType1 < msg.CardsType_Pair {
			continue
		}

		cards1, cards2 = PositiveSortCards(cards1), PositiveSortCards(cards2)
		transCards1, transCards2 := TransformCards(cards1), TransformCards(cards2)
		typeString1, typeString2 := CardsTypeToString(cardsType1), CardsTypeToString(cardsType2)

		t.Logf("手牌1: %v, 牌型: %s", transCards1, typeString1)
		t.Logf("手牌2: %v, 牌型: %s,               %v", transCards2, typeString2, ContrastCards(cards1, cards2))

		t.Log("")

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

// 测试比牌2
func TestContrastCards2(t *testing.T) {
	cards1, cards2 := []byte{0x21, 0xa1, 0xa2, 0xa3}, []byte{0x91, 0x93, 0x94, 0xd1}
	cardsType1, cardsType2 := GetCardsType(cards1), GetCardsType(cards2)

	cards1, cards2 = PositiveSortCards(cards1), PositiveSortCards(cards2)
	transCards1, transCards2 := TransformCards(cards1), TransformCards(cards2)
	typeString1, typeString2 := CardsTypeToString(cardsType1), CardsTypeToString(cardsType2)
	t.Logf("手牌1: %v, 牌型: %s", transCards1, typeString1)
	t.Logf("手牌2: %v, 牌型: %s,               %v", transCards2, typeString2, ContrastCards(cards1, cards2))
}

// 测试获取手牌重复 多少个数的牌
func TestGetRepeatedCards(t *testing.T) {
	cards := []byte{0x21, 0xa1, 0xa2, 0xb1, 0xb2, 0xb3}

	newCards := GetRepeatedCards(cards, 2)

	newCards = PositiveSortCards(newCards)
	transCards := TransformCards(newCards)
	t.Log(transCards)
}

// 获取重复次数超过 目标的数组值
func TestGetRepeatedValueArr(t *testing.T) {
	cards := []byte{0x21, 0xa1, 0xa2, 0xb1, 0xb2, 0xb3}

	valueArr := GetRepeatedValueArr(cards, 4)

	t.Log(valueArr)
}

func TestFindTypeInCards(t *testing.T) {

	userCards := []byte{0x11, 0x21, 0x31, 0x41, 0x51, 0x52, 0x61, 0x62, 0x71, 0x72, 0x73, 0x81, 0x91, 0xa1}

	solutions, _ := FindTypeInCards(msg.CardsType_Sequence, userCards)

	t.Log(solutions)

	for _, v := range solutions {
		transCards := TransformCards(v.Cards)
		t.Log(transCards)
	}
}

func TestSortCards(t *testing.T) {
	cards1 := []byte{0xa1, 0xa2}
	cards2 := []byte{0xb1, 0xb2}

	solutions := []SolutionCards{
		{
			Cards:     cards1,
			CardsType: GetCardsType(cards1),
			PutScore:  1,
		},

		{
			Cards:     cards2,
			CardsType: GetCardsType(cards2),
			PutScore:  2,
		},
	}

	score2Solutions := []SolutionCards{}

	for _, v := range solutions {
		if v.PutScore >= 2 {
			score2Solutions = append(score2Solutions, v)
		}
	}

}

func TestIsTripletWithSingle(t *testing.T) {

	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		cards := gamePoker.Cards[:4]
		if IsTripletWithSingle(cards) {
			//transCards1 := TransformCards(cards)
			//t.Logf("手牌: %v,", transCards1)

		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestIsSequence(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		len := rand.RandInt(5, 13)

		cards := gamePoker.Cards[:len]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		if IsSequence(cards) {
			t.Logf("手牌: %v", transCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestIsSerialTripletWithOne(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		lenArr := []int{8, 12, 16, 20}

		randIndex := rand.RandInt(0, 4)

		cards := gamePoker.Cards[:lenArr[randIndex]]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		if IsSerialTripletWithOne(cards) {
			t.Logf("手牌: %v", transCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestIsSerialTripletWithOne1(t *testing.T) {
	//cards := []byte{0x51, 0x52, 0x53, 0x61, 0x62, 0x63, 0x71, 0x72, 0x73, 0x81, 0x82, 0x83}
	//cards := []byte{0x51, 0x52, 0x53, 0x61, 0x62, 0x63, 0x71, 0x72, 0x73, 0x92, 0x93, 0x94}
	//cards := []byte{0x51, 0x52, 0x53, 0x61, 0x62, 0x63, 0x71, 0x72, 0x73, 0x81, 0x82, 0x83, 0x91, 0x92, 0x93, 0x94}
	//cards := []byte{0x51, 0x52, 0x53, 0x71, 0x72, 0x73, 0x81, 0x82, 0x83, 0x91, 0x92, 0x93}
	cards := []byte{0x51, 0x52, 0x53, 0x71, 0x72, 0x73, 0x81, 0x82, 0x83, 0xa1, 0xa2, 0xa3}
	cards = PositiveSortCards(cards)
	transCards := TransformCards(cards)
	t.Log(transCards)
	if IsSerialTripletWithOne(cards) {
		t.Logf("是飞机带单张")
	} else {
		t.Logf("不是飞机带单张")
	}
}

func TestIsSerialTripletWithWing(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		lenArr := []int{10, 15, 20}

		randIndex := rand.RandInt(0, 3)

		cards := gamePoker.Cards[:lenArr[randIndex]]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		if IsSerialTripletWithWing(cards) {
			t.Logf("手牌: %v", transCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestIsQuartetWithTwo(t *testing.T) {

	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		cards := gamePoker.Cards[:6]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		if IsQuartetWithTwo(cards) {
			t.Logf("手牌: %v", transCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestIsQuartetWithTwoPair(t *testing.T) {

	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		cards := gamePoker.Cards[:8]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		if IsQuartetWithTwoPair(cards) {
			t.Logf("手牌: %v", transCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestGetPlane(t *testing.T) {
	count := 0
	for {
		gamePoker := new(GamePoker)
		gamePoker.InitPoker()
		gamePoker.ShuffleCards()

		cards := gamePoker.Cards[:8]
		cards = PositiveSortCards(cards)
		transCards := TransformCards(cards)

		planeCards, _ := GetPlane(cards)
		planeTransCards := TransformCards(planeCards)
		if len(planeCards) > 0 {
			t.Log(transCards)
			t.Log(planeTransCards)
		}

		if count >= MaxTurnCount {
			break
		}

		count++
	}
}

func TestGetPlane1(t *testing.T) {
	cards := []byte{0xa1, 0xa2, 0xa3, 0xb1, 0xb2, 0xb3, 0xc1, 0xc2, 0xc3, 0xd1, 0xd2, 0xd3}
	planeCards, _ := GetPlane(cards)
	planeTransCards := TransformCards(planeCards)
	t.Log(planeTransCards)
}
