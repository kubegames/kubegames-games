package main

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/970102/gamelogic"
	"go-game-sdk/example/game_LaBa/970102/msg"
	"go-game-sdk/example/game_LaBa/970102/test/config"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin/json"
)

var (
	Conf              = simplejson.New()
	startcoin         = int64(0)
	endcoin           = int64(0)
	coin              = int64(0)
	winNum            = 0
	loseNum           = 0
	heNum             = 0
	shengLv           = float64(0)
	bet               = int64(0)
	winCoin           = int32(0)
	shouyi            = float64(0)
	taxCoin           = int64(0)
	realRoundNum      = 0
	paiXing           = []int{}
	paiXingGaiLv      = []string{}
	shijipaiXing      = []int{}
	shijipaiXingGaiLv = []string{}
	realWinCoin       = int32(0)
	kingNum           = 0
	huanPaiNum        = 0
	realNum           = 0
	bloodKey          = 0
	shootNum          = 0
	taxNum            = 0
)

func main() {
	loadCfg()
	config.Load()
	game := &gamelogic.Game{}
	rand.Seed(time.Now().UnixNano())
	bloodKey, _ = Conf.Get("bloodKey").Int()
	shootNum, _ = Conf.Get("roundNum").Int()
	diZhu, _ := Conf.Get("diZhu").Int()
	taxNum, _ = Conf.Get("tax").Int()
	startcoin, _ = Conf.Get("startCoin").Int64()
	paiXing = make([]int, 11)
	paiXingGaiLv = make([]string, 11)
	shijipaiXing = make([]int, 11)
	shijipaiXingGaiLv = make([]string, 11)
	for i := 0; i < shootNum; i++ {
		if startcoin+coin <= 0 {
			break
		}
		realRoundNum++
		game.PayKey = getAKey("7")
		game.Poker.InitPoker()
		game.Poker.ShuffleCards()
		game.Cards = make([]byte, 0)
		game.SpareCards = make([]byte, 0)
		chance := config.GetXueChiChance(int32(bloodKey))
		d := int64(diZhu)
		bet += d
		coin -= d
		cards := make([]byte, 0)
		pokerType := msg.PokerType_ZERO
		if chance != 0 {
			pokerType = getPokerType(chance, game.PayKey)
			cards = game.GetACards(pokerType)
			if pokerType != msg.PokerType_ZERO {
				realNum++
			}
			shijipaiXing[pokerType]++
		}
		if chance == 0 {
			cards = game.GetARandomCards()
		}
		//if pokerType == msg.PokerType_DUIZI {
		//	game.Cards = append(game.Cards, cards[:5]...)
		//	game.SpareCards = append(game.SpareCards, cards[5:]...)
		//}
		isFirst := false
		//if pokerType != msg.PokerType_DUIZI {
		length := len(cards)
		for i := 0; i < length; i++ {
			r := rand.Intn(100)
			card := cards[i]
			if (len(game.SpareCards) == 5) || (r < 90 && len(game.Cards) < 5) {
				game.Cards = append(game.Cards, card)
				if card > 240 {
					kingNum++
				}
				continue
			}
			if len(game.Cards) < 5 && pokerType != msg.PokerType_ZERO {
				isFirst = true
				huanPaiNum++
			}
			game.SpareCards = append(game.SpareCards, card)
		}
		//}
		if isFirst && pokerType != msg.PokerType_ZERO {
			realWinCoin += config.GetPokerPay(int(pokerType), game.PayKey) * int32(diZhu)
		}
		pokerType = game.CheckPokerType(game.Cards)
		coinChange := int32(0)
		if pokerType != msg.PokerType_ZERO {
			coinChange = config.GetPokerPay(int(pokerType), game.PayKey) * int32(diZhu)
		}
		paiXing[pokerType]++
		winCoin += coinChange
		if !isFirst {
			realWinCoin += coinChange
		}
		taxedCoin := int64(tax(int(coinChange), taxNum))
		coin += taxedCoin
		taxCoin += taxedCoin
		if taxedCoin-d > 0 {
			winNum++
		}
		if taxedCoin-d == 0 {
			heNum++
		}
		if taxedCoin-d < 0 {
			loseNum++
		}
	}
	shengLv = float64(winNum+heNum) / float64(realRoundNum)
	shouyi = float64(winCoin) / float64(bet)
	taxCoin = int64(winCoin) - taxCoin
	for k, v := range paiXing {
		gailv := float64(v) / float64(winNum+heNum)
		paiXingGaiLv[k] = fmt.Sprintf("%.3f", gailv)
	}
	for k, v := range shijipaiXing {
		gailv := float64(v) / float64(winNum+heNum)
		shijipaiXingGaiLv[k] = fmt.Sprintf("%.3f", gailv)
	}
	writeFile()
}

func getPokerType(chance int32, payKey int) msg.PokerType {
	r := rand.Intn(10000)
	if int32(r) > chance {
		return msg.PokerType(getAKey(strconv.Itoa(payKey + 3)))
	}
	return msg.PokerType_ZERO
}

func getAKey(key string) int {
	r := rand.Intn(10000)
	chance := 0
	for i := 1; i < 11; i++ {
		chance += config.GetValue(key, strconv.Itoa(i))
		if r < chance {
			return i
		}
	}
	return 1
}

func writeFile() {
	f, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	defer f.Close()

	f.WriteString("当前作弊率：")
	b, _ := json.Marshal(bloodKey)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("初始金币：")
	b, _ = json.Marshal(startcoin)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("结束金币：")
	b, _ = json.Marshal(startcoin + coin)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("金币变化：")
	b, _ = json.Marshal(coin)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("应运行局数：")
	b, _ = json.Marshal(shootNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("运行局数：")
	b, _ = json.Marshal(realRoundNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("实际数据：")
	b, _ = json.Marshal(taxCoin)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("收益率：")
	b, _ = json.Marshal(shouyi)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("胜利金币：")
	b, _ = json.Marshal(winCoin)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("税收：")
	b, _ = json.Marshal(taxCoin)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("胜利局数：")
	b, _ = json.Marshal(winNum)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("失败局数：")
	b, _ = json.Marshal(loseNum)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("和局：")
	b, _ = json.Marshal(heNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("平均胜利金额：")
	b, _ = json.Marshal(fmt.Sprintf("%.3f", float64(winCoin)/float64(winNum+heNum)))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("胜率：")
	b, _ = json.Marshal(shengLv)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("牌型：")
	b, _ = json.Marshal(paiXing)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("牌型概率：")
	b, _ = json.Marshal(paiXingGaiLv)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("理论数据：")
	b, _ = json.Marshal(taxCoin)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("收益率：")
	b, _ = json.Marshal(float64(realWinCoin) / float64(bet))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("胜利金币：")
	b, _ = json.Marshal(realWinCoin)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("税收：")
	b, _ = json.Marshal(int(realWinCoin) - tax(int(realWinCoin), taxNum))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("实际应该赢局数：")
	b, _ = json.Marshal(realNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("失败局数：")
	b, _ = json.Marshal(shootNum - realNum - heNum)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("和局：")
	b, _ = json.Marshal(heNum)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("胜率：")
	b, _ = json.Marshal(float64(realNum) / float64(realRoundNum))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("平均胜利金额：")
	b, _ = json.Marshal(fmt.Sprintf("%.3f", float64(realWinCoin)/float64(realNum+heNum)))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("牌型：")
	b, _ = json.Marshal(shijipaiXing)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("牌型概率：")
	b, _ = json.Marshal(shijipaiXingGaiLv)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("大小王概率：")
	b, _ = json.Marshal(float64(kingNum) / float64(winNum+heNum))
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("换牌局数：")
	b, _ = json.Marshal(huanPaiNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("投入：")
	b, _ = json.Marshal(bet)
	f.Write(b)
}

func tax(coinChange, tax int) int {
	if coinChange > 0 {
		coin := coinChange * (10000 - tax) / 10000
		return coin
	}
	return coinChange
}

func loadCfg() {
	f, err := ioutil.ReadFile("test.json")
	if err == nil {
		Conf, err = simplejson.NewJson(f)
		if err != nil {
			fmt.Println("test conf err", err)
			return
		}
	}
}
