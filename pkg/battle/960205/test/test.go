package main

import (
	"encoding/json"
	"fmt"
	"game_MaJiang/erbagang/glogic"
	"game_MaJiang/erbagang/poker"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/bitly/go-simplejson"
)

var Conf = simplejson.New()
var coin = []int32{0, 0, 0, 0}
var zhuangNum = []int32{0, 0, 0, 0}
var winNum = []int32{0, 0, 0, 0}
var loseNum = []int32{0, 0, 0, 0}
var heNum = []int32{0, 0, 0, 0}
var winCoin = []int32{0, 0, 0, 0}
var bet = []int32{0, 0, 0, 0}
var firstNum = []int32{0, 0, 0, 0}
var secondNum = []int32{0, 0, 0, 0}
var shouyi = []float64{0, 0, 0, 0}
var shengLv = []float64{0, 0, 0, 0}

func main() {
	loadCfg()
	game := Init()
	roundNum, _ := Conf.Get("roundNum").Int()
	userArr := []int32{}
	chances, _ := Conf.Get("blood").Array()
	index := make([]int32, 0)
	for _, v := range chances {
		s, _ := strconv.Atoi(v.(json.Number).String())
		index = append(index, int32(s))
	}
	z, _ := Conf.Get("qiangZhuang").Array()
	zhuang := make([]int32, 0)
	for _, v := range z {
		s, _ := strconv.Atoi(v.(json.Number).String())
		zhuang = append(zhuang, int32(s))
	}
	g, _ := Conf.Get("qiangZhuangGaiLv").Array()
	gaiLv := make([]int32, 0)
	for _, v := range g {
		s, _ := strconv.Atoi(v.(json.Number).String())
		gaiLv = append(gaiLv, int32(s))
	}
	taxNum, _ := Conf.Get("tax").Int()
	for i := 0; i < roundNum; i++ {
		if i%5 == 0 {
			game.GamePoker.InitPoker()
			game.GamePoker.ShuffleCards()
		}
		z := zhuang
		z = checkQiangZhuang(z, gaiLv)
		maxIndex := int(game.GetAChance(z))
		if maxIndex == -1 {
			maxIndex = checkNoZhuang(z)
		}
		game.RobZhuangIndex = maxIndex
		zhuangNum[maxIndex]++
		for i := 0; i < 4; i++ {
			for j := 0; j < 2; j++ {
				card := int32(game.GamePoker.DealCards())
				userArr = append(userArr, card)
			}
			game.PaiList[i] = userArr
			userArr = nil
		}
		cards := game.GetOrderCards()
		indexs := game.GetOrderIndex(index)
		length := len(indexs)
		for i := 0; i < len(cards); i++ {
			if length == 0 {
				break
			}
			if i < length {
				game.PaiList[int(indexs[i])] = cards[i]
				if i == 0 {
					firstNum[int(indexs[i])]++
				}
				if i == 1 {
					secondNum[int(indexs[i])]++
				}
			}
			if i >= length {
				index := game.GetAOutIndex(int32(len(cards)), indexs)
				indexs = append(indexs, index)
				game.PaiList[int(index)] = cards[i]
				if i == 1 {
					secondNum[index]++
				}
			}
		}
		zhuangPai := make([]byte, 0)
		xianPai := make([]byte, 0)
		for i := 0; i < len(game.PaiList[game.RobZhuangIndex]); i++ {
			zhuangPai = append(zhuangPai, byte(game.PaiList[game.RobZhuangIndex][i]))
		}
		zhuangCoin := 0

		for i := 0; i < 4; i++ {
			if i != game.RobZhuangIndex {
				for j := 0; j < len(game.PaiList[i]); j++ {
					xianPai = append(xianPai, byte(game.PaiList[i][j]))
				}
				fmt.Println("牌", zhuangPai, xianPai, i)
				biPaiJieGuo := poker.GetCompareCardsRes(zhuangPai, xianPai)
				a := game.DiZhu
				if biPaiJieGuo == 1 {
					// 庄家盈
					coin[i] -= int32(a)
					zhuangCoin += a
					winCoin[game.RobZhuangIndex] += int32(a)
					bet[i] += int32(a)
					loseNum[i]++
				} else {
					// 闲家
					coin[i] += int32(tax(a, taxNum))
					zhuangCoin -= a
					winCoin[i] += int32(a)
					winNum[i]++
					bet[game.RobZhuangIndex] += int32(a)
				}
				xianPai = nil
			}
		}
		if zhuangCoin > 0 {
			winNum[game.RobZhuangIndex]++
		}
		if zhuangCoin == 0 {
			heNum[game.RobZhuangIndex]++
		}
		if zhuangCoin < 0 {
			loseNum[game.RobZhuangIndex]++
		}
		coin[game.RobZhuangIndex] += int32(tax(zhuangCoin, taxNum))
	}
	for k, v := range winCoin {
		shouyi[k] = float64(v) / float64(bet[k])
	}
	for k, v := range winNum {
		shengLv[k] = float64(v) / float64(roundNum)
	}
	fmt.Println("coin = ", coin)
	//fmt.Println("bet = ", bet)
	writeFile()
}

func checkNoZhuang(zhuang []int32) int {
	for _, v := range zhuang {
		if v != 0 {
			return -1
		}
	}
	return rand.Intn(4)
}

func checkQiangZhuang(zhuang, gailv []int32) []int32 {
	for k, _ := range zhuang {
		if !checkChance(int(gailv[k])) {
			zhuang[k] = 0
		}
	}
	return zhuang
}

func checkChance(chance int) bool {
	r := rand.Intn(100)
	if r < chance {
		return true
	}
	return false
}

func writeFile() {
	f, err := os.OpenFile("./result.txt", os.O_APPEND|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("open file err", err)
		return
	}
	defer f.Close()
	f.WriteString("金币变化：")
	b, _ := json.Marshal(coin)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("庄家局数：")
	b, _ = json.Marshal(zhuangNum)
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
	f.WriteString("最大牌次数：")
	b, _ = json.Marshal(firstNum)
	f.Write(b)
	f.WriteString("\r\n")
	f.WriteString("二大牌次数：")
	b, _ = json.Marshal(secondNum)
	f.Write(b)
}

func tax(coinChange, tax int) int {
	if coinChange > 0 {
		coin := coinChange * (10000 - tax) / 10000
		return coin
	}
	return coinChange
}

func Init() glogic.ErBaGangGame {
	game := glogic.ErBaGangGame{}
	dizhu, err := Conf.Get("diZhu").Int()
	if err != nil {
		fmt.Println("dizhu err", err)
	}
	game.DiZhu = dizhu
	game.PaiList = make(map[int][]int32, 0)
	return game
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
