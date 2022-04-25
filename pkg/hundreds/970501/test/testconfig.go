package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type testConfig struct {
	RoomProb   int32    `json:"roomProb"`   // 血池值
	TestTimes  int64    `json:"testTimes"`  // 测试次数
	InitGold   int64    `json:"initGold"`   // 初始货币
	Tax        int64    `json:"tax"`        // 税点
	Bottom     int64    `json:"bottom"`     // 底注
	BetArr     [8]int64 `json:"betArr"`     // 下注
	IsOpenRand bool     `json:"isOpenRand"` // 是否开启随机
}

var Conf testConfig

func init() {
	data, err := ioutil.ReadFile("./testconfig.json")
	if err != nil {
		log.Println("open file error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	json.Unmarshal([]byte(result.String()), &Conf)
	Conf.RoomProb = int32(result.Get("roomProb").Int())
	Conf.TestTimes = int64(result.Get("testTimes").Int())
	Conf.InitGold = int64(result.Get("initGold").Int())
	Conf.Tax = int64(result.Get("tax").Int())
	Conf.Bottom = int64(result.Get("bottom").Int())

	betArr := result.Get("betArr").Array()
	if len(betArr) != 8 {
		panic("betArr长度必须为8")
	}

	for i, v := range betArr {
		Conf.BetArr[i] = v.Int()
	}
	Conf.IsOpenRand = result.Get("isOpenRand").Bool()

	checkTestConfig()
}

// 检查config文件
func checkTestConfig() {
	if Conf.TestTimes <= 0 {
		Conf.TestTimes = 10000
	}
	if Conf.InitGold <= 0 {
		Conf.InitGold = 100000
	}
	if Conf.Tax < 0 {
		Conf.Tax = 5
	}
	switch Conf.RoomProb {
	case -3000, -2000, -1000, 1000, 2000, 3000:
	default:
		log.Fatalln("roomProb must be one of [-3000,-2000,-1000,1000,2000,3000]")
	}
}
