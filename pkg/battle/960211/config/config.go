package config

import (
	"encoding/json"
	"game_poker/pai9/model"
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type Config struct {
	Taketimes struct {
		Shuffle     int `json:"shuffle"`     // 洗牌阶段
		QiangZhuang int `json:"qiangzhuang"` // 抢庄阶段
		SendZhuang  int `json:"sendZhuang"`  // 发送谁是庄
		Bet         int `json:"bet"`         // 下注倍数结算
		DealPoker   int `json:"dealpoker"`   // 发牌阶段
		ShowPoker1  int `json:"showpoker1"`  // 第一句结束展示牌
		ShowPoker2  int `json:"showpoker2"`  // 第二句结束展示牌
	} `json:"taketimes"`

	LevelConfig []Multi         `json:"levelConfig"`
	PokerCtrl   map[int32]int64 `json:"pokerCtrl"` // roomProb:prob
}

type Multi struct {
	QiangMulti []int32 `json:"qiangMulti"` // 抢庄倍数
}

var Pai9Config Config

func InitConfig(path string) {
	if path == "" {
		path = "./conf/pai9.json"
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("LoadBirdAminalConfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	initConfig(result)
}

func initConfig(cfg gjson.Result) {
	Pai9Config.Taketimes.Shuffle = int(cfg.Get("taketimes.shuffle").Int())
	Pai9Config.Taketimes.QiangZhuang = int(cfg.Get("taketimes.qiangzhuang").Int())
	Pai9Config.Taketimes.SendZhuang = int(cfg.Get("taketimes.sendZhuang").Int())
	Pai9Config.Taketimes.Bet = int(cfg.Get("taketimes.bet").Int())
	Pai9Config.Taketimes.DealPoker = int(cfg.Get("taketimes.dealpoker").Int())
	Pai9Config.Taketimes.ShowPoker1 = int(cfg.Get("taketimes.showpoker1").Int())
	Pai9Config.Taketimes.ShowPoker2 = int(cfg.Get("taketimes.showpoker2").Int())

	multis := cfg.Get("levelConfig").Array()

	for _, v := range multis {
		var multi Multi
		if err := json.Unmarshal([]byte(v.Raw), &multi); err != nil {
			panic(err)
		}
		Pai9Config.LevelConfig = append(Pai9Config.LevelConfig, multi)
	}

	cards := cfg.Get("cards").Array()
	for _, card := range cards {
		c := new(model.Card)
		if err := json.Unmarshal([]byte(card.Raw), c); err != nil {
			panic(err)
		}
		model.CardsAllType = append(model.CardsAllType, c)
	}
	Pai9Config.PokerCtrl = make(map[int32]int64, 0)
	Pai9Config.PokerCtrl[-3000] = cfg.Get("-3000").Int()
	Pai9Config.PokerCtrl[-2000] = cfg.Get("-2000").Int()
	Pai9Config.PokerCtrl[-1000] = cfg.Get("-1000").Int()
	Pai9Config.PokerCtrl[1000] = cfg.Get("1000").Int()
	Pai9Config.PokerCtrl[2000] = cfg.Get("2000").Int()
	Pai9Config.PokerCtrl[3000] = cfg.Get("3000").Int()
}
