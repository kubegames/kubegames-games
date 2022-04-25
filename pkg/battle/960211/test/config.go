package main

import (
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type conf struct {
	RoomProb  int   `json:"roomProb"`  // 作弊率
	Times     int   `json:"times"`     // 设定局数
	BeginGold int64 `json:"beginGold"` // 初始金额
	Tax       int   `json:"tax"`       // 税率
	Bottom    int   `json:"bottom"`    // 下注底注
	IsBetRand bool  `json:"isBetRand"` // 下注倍数是否随机,true：随机；false：默认第一个
}

var Conf conf

func initConf(path string) {
	if path == "" {
		path = "./test.json"
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("加载测试配置文件失败 Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))

	Conf.RoomProb = int(result.Get("roomProb").Int())
	Conf.Times = int(result.Get("times").Int())
	Conf.BeginGold = result.Get("beginGold").Int()
	Conf.Tax = int(result.Get("tax").Int())
	Conf.Bottom = int(result.Get("bottom").Int())
	Conf.IsBetRand = result.Get("isBetRand").Bool()
}
