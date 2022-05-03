package gamelogic

import (
	"math/rand"
	"testing"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func TestName(t *testing.T) {
	game := Game{}
	rand.Seed(time.Now().UnixNano())
	//for i := 0; i < 10000; i++ {
	//	game.Poker.InitPoker()
	//	game.Poker.ShuffleCards()
	//	r := rand.Intn(11)
	//	game.getACards(msg.PokerType(r))
	//	//log.Traceln("+++++++++++++++++++++")
	//	//game.transformCardForMsg(cards)
	//}
	card := []byte{161, 193, 242, 241, 129}
	log.Traceln(game.checkPokerType(card))
	//pokerType := game.checkPokerType([]byte{68,100,98,66,84})
	//r := game.getAllCompose(10, 5)
	//log.Traceln(pokerType)
}
