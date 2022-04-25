package game

import (
	"testing"

	"github.com/kubegames/kubegames-games/pkg/battle/960207/poker"
)

func TestGeneralNiuniu_SetCardsSequence(t *testing.T) {
	game := new(GeneralNiuniu)
	game.Poker = poker.GamePoker{}
	game.Poker.InitPoker()
	game.SetCardsSequence()

}

func TestGeneralNiuniu_GameStart(t *testing.T) {
	var a int64 = 1250

	b := a / 5333 / 4 / 100
	t.Log(b)
}
