package game

import (
	"testing"

	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
)

func TestBankerNiuniu_SetCardsSequence(t *testing.T) {
	game := new(BankerNiuniu)
	game.Poker = &poker.GamePoker{}
	game.Poker.InitPoker()
	game.SetCardsSequence()

}

func TestBankerNiuniu_GameStart(t *testing.T) {
	var a int64 = 1250

	b := a / 5333 / 4 / 100
	t.Log(b)
}
