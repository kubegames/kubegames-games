package game

import (
	"game_poker/ddzall/msg"
	"game_poker/ddzall/poker"
	"strconv"
	"strings"
	"testing"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

func TestDouDizhu_GameStart(t *testing.T) {
	putCardsReq := &msg.PutCardsReq{
		Cards: []byte{51},
	}
	buffer, err := proto.Marshal(putCardsReq)
	if err != nil {
		log.Errorf("proto marshal fail : %v", err.Error())
		return
	}

	req := &msg.PutCardsReq{}
	if err := proto.Unmarshal(buffer, req); err != nil {
		log.Errorf("解析出牌入参错误: %v", err.Error())
		return
	}

	t.Log(putCardsReq.Cards[0])
	t.Log(req.Cards[0])

	//t.Log(putCardsReq.Cards[0] == req.Cards[0])
}

func TestRobot_Init(t *testing.T) {
	var a [][]int
	a = append(a, []int{})
	t.Log(a)
	t.Log(len(a))
}

func TestGetNumTypeByCards(t *testing.T) {

	gamePoker := new(poker.GamePoker)
	gamePoker.InitPoker()
	gamePoker.ShuffleCards()

	//cards := gamePoker.Cards[:20]
	cards := []byte{}
	cards = poker.PositiveSortCards(cards)
	transCards := poker.TransformCards(cards)
	t.Log(transCards)
	numType := GetNumTypeByCards(cards)
	t.Log(numType)
}

func TestGetCardsByNumType(t *testing.T) {
	gamePoker := new(poker.GamePoker)
	gamePoker.InitPoker()
	gamePoker.ShuffleCards()

	cards := gamePoker.Cards[:5]
	cards = poker.PositiveSortCards(cards)
	transCards := poker.TransformCards(cards)
	t.Log(transCards)

	numType := [15]int{4, 4, 1, 1, 1, 1, 1, 1, 2, 1, 1, 3, 2, 0, 1}
	outcards := GetCardsByNumType(numType, cards)
	t.Log(numType)
	transOutcards := poker.TransformCards(outcards)
	t.Log(transOutcards)
	t.Log(len(cards))
}

func TestRobot_Init2(t *testing.T) {
	var totalNum []int
	var numSting []string
	numType1 := [15]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	numType2 := [15]int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}

	totalNum = append(totalNum, numType1[:]...)
	totalNum = append(totalNum, numType2[:]...)
	for _, num := range totalNum {
		numSting = append(numSting, strconv.Itoa(num))
	}

	result := strings.Join(numSting, ",")
	t.Log(result)
}

func TestRobot_Init3(t *testing.T) {
	var resultNum [15]int
	str := "100110333000000>"
	content := str[:15]
	t.Log(content)
	s := strings.Split(content, "")
	t.Log(s)
	for i, str := range s {
		tttt, _ := strconv.Atoi(str)
		resultNum[i] = tttt
	}
	t.Log(resultNum)

	gamePoker := new(poker.GamePoker)
	gamePoker.InitPoker()
	gamePoker.ShuffleCards()

	cards := gamePoker.Cards[:54]
	outcards := GetCardsByNumType(resultNum, cards)
	transOutcards := poker.TransformCards(outcards)
	t.Log(transOutcards)
}
