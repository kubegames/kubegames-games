package poker_test

import (
	"testing"

	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
)

const MaxTurnCount = 1000

func TestGetCardValueAndColor(t *testing.T) {
	specialWords := ""
	t.Log(len(specialWords))
}

// TestGamePoker_DrawCard 发牌测试
func TestGamePoker_DrawCard(t *testing.T) {

	gamePoker := new(poker.GamePoker)
	gamePoker.InitPoker()
	t.Log(gamePoker.Cards)
	cards := gamePoker.DrawCard()
	t.Log(cards)
	t.Log(gamePoker.Cards)
}

// TestGetCardsType 测试获取牌型
func TestGetCardsType(t *testing.T) {
	count := 0
	for {
		gamePoker := new(poker.GamePoker)
		gamePoker.InitPoker()

		cards := gamePoker.DrawCard()
		t.Log("")

		transCards := poker.TransformCards(cards)
		t.Log(transCards)

		cardsType := poker.GetCardsType(cards)

		typeString := poker.TransformCardsType(cardsType)
		t.Log(typeString)

		count++
		if count >= MaxTurnCount {
			break
		}
	}

}

// TestGetSpecialCardIndexs 获取牌型特殊牌下标
func TestGetSpecialCardIndexs(t *testing.T) {
	count := 0
	for {
		gamePoker := new(poker.GamePoker)
		gamePoker.InitPoker()

		cards := gamePoker.DrawCard()
		t.Log("")

		transCards := poker.TransformCards(cards)
		t.Log(transCards)

		cardsType := poker.GetCardsType(cards)

		typeString := poker.TransformCardsType(cardsType)
		t.Log(typeString)

		indexs := poker.GetSpecialCardIndexs(cards, cardsType)
		t.Log(indexs)

		count++
		if count >= MaxTurnCount {
			break
		}
	}
}

// TestGetCardsMultiple 测试获取牌型对应的倍数
func TestGetCardsMultiple(t *testing.T) {
	count := 0
	for {
		gamePoker := new(poker.GamePoker)
		gamePoker.InitPoker()

		cards := gamePoker.DrawCard()
		t.Log("")

		cardsType := poker.GetCardsType(cards)

		typeString := poker.TransformCardsType(cardsType)
		t.Log(typeString)

		multiple := poker.GetCardsMultiple(cardsType)
		t.Log(multiple)

		count++
		if count >= MaxTurnCount {
			break
		}
	}
}

// GetBiggestFromCards 从牌组中选出最大的牌
func TestGetBiggestFromCards(t *testing.T) {
	count := 0
	for {
		gamePoker := new(poker.GamePoker)
		gamePoker.InitPoker()

		cards := gamePoker.DrawCard()
		t.Log("")

		transCards := poker.TransformCards(cards)
		t.Log(transCards)

		cardsType := poker.GetCardsType(cards)

		holdCards := &poker.HoldCards{
			Cards:     cards,
			CardsType: cardsType,
		}

		typeString := poker.TransformCardsType(cardsType)
		t.Log(typeString)

		biggestCard := poker.GetBiggestFromCards(holdCards)
		biggestCards := poker.TransformCards([]byte{biggestCard})
		t.Log(biggestCards)

		count++
		if count >= MaxTurnCount {
			break
		}
	}
}

// TestContrastCards 测试比牌
func TestContrastCards(t *testing.T) {
	count := 0
	for {
		gamePoker := new(poker.GamePoker)
		gamePoker.InitPoker()
		t.Log("")

		bankerCards := gamePoker.DrawCard()
		bankerCardsType := poker.GetCardsType(bankerCards)

		bankerHoldCards := &poker.HoldCards{
			Cards:     bankerCards,
			CardsType: bankerCardsType,
		}

		bankerTransCards := poker.TransformCards(bankerCards)
		bankerTransType := poker.TransformCardsType(bankerCardsType)

		bankerBiggestCard := poker.GetBiggestFromCards(bankerHoldCards)
		bankerBiggestCards := poker.TransformCards([]byte{bankerBiggestCard})
		t.Log("庄家牌：", bankerTransCards, ", 牌型为：", bankerTransType, ", 最大牌值：", bankerBiggestCards)

		playerCards := gamePoker.DrawCard()
		playerCardsType := poker.GetCardsType(playerCards)

		playerHoldCards := &poker.HoldCards{
			Cards:     playerCards,
			CardsType: playerCardsType,
		}

		playerTransCards := poker.TransformCards(playerCards)
		playerTransType := poker.TransformCardsType(playerCardsType)

		playerBiggestCard := poker.GetBiggestFromCards(playerHoldCards)
		playerBiggestCards := poker.TransformCards([]byte{playerBiggestCard})
		t.Log("闲家牌：", playerTransCards, ", 牌型为：", playerTransType, ", 最大牌值：", playerBiggestCards)

		if poker.ContrastCards(bankerHoldCards, playerHoldCards) {
			t.Log("闲家赢")
		} else {
			t.Log("庄家赢")
		}

		count++
		if count >= MaxTurnCount {
			break
		}
	}
}
