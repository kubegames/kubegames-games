package poker

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// 牌1-9,10为白板
var Deck = []byte{
	0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa,
	0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa,
	0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa,
	0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0xa,
}

type GamePoker struct {
	Cards []byte
}

func (gp *GamePoker) InitPoker() {
	gp.Cards = make([]byte, 0)
	for _, v := range Deck {
		gp.Cards = append(gp.Cards, v)
	}
}

//洗牌
func (gp *GamePoker) ShuffleCards() {
	for i := 0; i < len(gp.Cards); i++ {
		index := rand.Intn(len(gp.Cards))
		gp.Cards[i], gp.Cards[index] = gp.Cards[index], gp.Cards[i]
	}
}

//分牌
func (gp *GamePoker) DealCards() byte {
	card := gp.Cards[0]
	gp.Cards = append(gp.Cards[:0], gp.Cards[1:]...)
	return card
}

func (gp *GamePoker) GetCardsCount() int {
	return len(gp.Cards)
}

//生成2张牌 TODO 暂时先随便生成2张牌，后面要根据相应的策略来生成牌
func GenerateCards() (cards []byte) {
	cards = make([]byte, 2)
	c0 := Deck[rand.Intn(40)]
	c1 := Deck[time.Now().UnixNano()%40]
	cards[0] = c0
	cards[1] = c1
	return
}

func Test(cards []byte) {
	sortRes := sortCards(cards)
	log.Traceln("sortRes : ", sortRes)
	t, _ := GetCardType(sortRes)
	log.Traceln("cardsType : ", t)
}
