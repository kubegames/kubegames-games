package main

import (
	"fmt"
	"game_buyu/renyuchuanshuo/config"
	"game_buyu/renyuchuanshuo/msg"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gin-gonic/gin/json"
)

var (
	Conf    = simplejson.New()
	coin    = int64(0)
	winNum  = 0
	loseNum = 0
	heNum   = 0
	shengLv = float64(0)
	bet     = int64(0)
	winCoin = 0
	shouyi  = float64(0)
	taxCoin = int64(0)
)

func main() {
	loadCfg()
	config.Load()
	rand.Seed(time.Now().UnixNano())
	bloodKey, _ := Conf.Get("bloodKey").Int()
	bulletLv, _ := Conf.Get("bulletLv").Int()
	shootNum, _ := Conf.Get("shootNum").Int()
	diZhu, _ := Conf.Get("diZhu").Int()
	taxNum, _ := Conf.Get("tax").Int()
	for i := 0; i < shootNum; i++ {
		fishType := msg.Type(rand.Intn(6) + 1)
		fishId := config.GetFishByType(fishType)
		chance := config.GetFishHitChance(fishId, false)
		chance += config.GetXueChiChance(fishId, int32(bloodKey))
		d := int64(bulletLv * diZhu)
		bet += d
		coin -= d
		r := rand.Intn(10000)
		if int32(r) < chance {
			score := int(config.GetFishScore(fishId))
			coinChange := bulletLv * diZhu * score
			winCoin += coinChange
			taxedCoin := int64(tax(coinChange, taxNum))
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
		} else {
			loseNum++
		}
	}
	shengLv = float64(winNum) / float64(shootNum)
	shouyi = float64(winCoin) / float64(bet)
	taxCoin = int64(winCoin) - taxCoin
	writeFile()
}

func writeFile() {
	f, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	defer f.Close()
	f.WriteString("金币变化：")
	b, _ := json.Marshal(coin)
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
	f.WriteString("胜率：")
	b, _ = json.Marshal(shengLv)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("胜利金币：")
	b, _ = json.Marshal(winCoin)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("收益率：")
	b, _ = json.Marshal(shouyi)
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
