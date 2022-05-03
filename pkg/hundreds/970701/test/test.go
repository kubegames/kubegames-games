package main

import (
	"encoding/json"
	"fmt"
	"game_poker/saima/config"
	gameLogic "game_poker/saima/game"
	"game_poker/saima/msg"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-sdk/pkg/log"
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
	lossCoin          = int32(0)
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
	xiaZhu            = 0
	xiaZhuNum         = 0

	FirstPay       = 0
	FAndSPay       = 0
	FirstFemalePay = 3.8
	FirstMalePay   = 1.27
)

func main() {
	loadCfg()
	config.Load()
	FirstPay = config.GetFirstPay()
	FAndSPay = config.GetFAndSPay()
	FirstFemalePay = config.GetFirstFemalePay()
	FirstMalePay = config.GetFirstMalePay()
	game := &gameLogic.Game{}
	game.AllResult = game.GetAllResult()
	rand.Seed(time.Now().UnixNano())
	bloodKey, _ = Conf.Get("bloodKey").Int()
	shootNum, _ = Conf.Get("roundNum").Int()
	diZhu, _ := Conf.Get("diZhu").Int()
	xiaZhu, _ = Conf.Get("xiaZhu").Int()
	xiaZhuNum, _ = Conf.Get("xiaZhuNum").Int()
	taxNum, _ = Conf.Get("tax").Int()
	startcoin, _ = Conf.Get("startCoin").Int64()
	paiXing = make([]int, 2)
	paiXingGaiLv = make([]string, 2)
	shijipaiXing = make([]int, 2)
	shijipaiXingGaiLv = make([]string, 2)
	for i := 0; i < shootNum; i++ {
		game.AllBet = 0
		game.Result = make([]int32, 0)
		game.BetInfo = make(map[msg.BetArea]int64, 38)
		xiazhu(game, int64(diZhu))
		if startcoin+coin-game.AllBet <= 0 {
			break
		}
		realRoundNum++
		result := make([]int32, 0)
		tem := make([][]int32, 0)
		min, controlLossRation := config.GetXueChiChance(int32(bloodKey))
		win := int64(0)
		for _, v := range game.AllResult {
			if game.AllBet == 0 {
				break
			}
			win = pay(game, v)
			lossRation := win * 10000 / game.AllBet
			//log.Traceln("lossRation =", win, v, lossRation, min, controlLossRation)
			if lossRation > min && lossRation < controlLossRation {
				//result = append(result, v...)
				tem = append(tem, v)
			}
		}
		if len(tem) > 0 {
			result = append(result, tem[rand.Intn(len(tem))]...)
		}
		if len(tem) == 0 {
			result = game.GetRandResult()
			realNum++
			win = pay(game, result)
		}
		bet += game.AllBet
		coin -= game.AllBet
		if result[0] < 3 {
			paiXing[0]++
		} else {
			paiXing[1]++
		}
		winCoin += int32(win)
		taxedCoin := int64(tax(int(win), taxNum))
		coin += taxedCoin
		taxCoin += taxedCoin
		if taxedCoin-game.AllBet > 0 {
			winNum++
		}
		if taxedCoin-game.AllBet == 0 {
			heNum++
		}
		if taxedCoin-game.AllBet < 0 {
			lossCoin += int32(game.AllBet)
			loseNum++
		}
	}
	shengLv = float64(winNum) / float64(realRoundNum)
	shouyi = float64(taxCoin) / float64(bet)
	taxCoin = int64(winCoin) - taxCoin
	for k, v := range paiXing {
		gailv := float64(v) / float64(realRoundNum)
		paiXingGaiLv[k] = fmt.Sprintf("%.3f", gailv)
	}
	writeFile()
}

func pay(game *gameLogic.Game, result []int32) int64 {
	win := int64(0)
	payArea := game.GetPayArea(result)
	for k, v1 := range game.BetInfo {
		if payArea[k] == 1 {
			if k < msg.BetArea_first_second_12 {
				win += v1 * int64(FirstPay)
			} else if k < msg.BetArea_champion_Man {
				win += v1 * int64(FAndSPay)
			} else if k == msg.BetArea_champion_Man {
				win += int64(float64(v1) * FirstMalePay)
			} else if k == msg.BetArea_champion_Woman {
				win += int64(float64(v1) * FirstFemalePay)
			}
		}
	}
	return win
}

func xiazhu(game *gameLogic.Game, dizhu int64) {
	if xiaZhu == 1 {
		for i := 0; i < xiaZhuNum; i++ {
			area := rand.Intn(38) + 1
			game.AllBet += dizhu
			game.BetInfo[msg.BetArea(area)] += dizhu
		}
	} else {
		for i := 0; i < 38; i++ {
			num := rand.Intn(10)
			for j := 0; j < num; j++ {
				game.AllBet += dizhu
				game.BetInfo[msg.BetArea(i+1)] += dizhu
			}
		}

	}

	//log.Traceln("下注信息 = ", game.BetInfo)
}

func writeFile() {
	f, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Traceln("open file err", err)
		return
	}
	defer f.Close()

	f.WriteString("当前作弊率：")
	b, _ := json.Marshal(bloodKey)
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

	f.WriteString("胜率：")
	b, _ = json.Marshal(shengLv)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("随机结果：")
	b, _ = json.Marshal(realNum)
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("随机结果概率：")
	b, _ = json.Marshal(fmt.Sprintf("%.3f", float64(realNum)/float64(realRoundNum)))
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

	f.WriteString("平均胜利金额：")
	b, _ = json.Marshal(fmt.Sprintf("%.3f", float64(winCoin)/float64(winNum)))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("平均输金额：")
	b, _ = json.Marshal(fmt.Sprintf("%.3f", float64(lossCoin)/float64(loseNum)))
	f.Write(b)
	f.WriteString("\r\n")

	f.WriteString("投入：")
	b, _ = json.Marshal(bet)
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

	f.WriteString("收益率：")
	b, _ = json.Marshal(shouyi)
	f.Write(b)
	f.WriteString("\r\n")

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
			log.Traceln("test conf err", err)
			return
		}
	}
}
