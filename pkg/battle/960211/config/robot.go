package config

import (
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type robot struct {
	QiangProb []int32 `json:"qiangProb"` // 抢庄概率
	BetProb   []int32 `json:"betProb"`   // 下注概率
	TimeProb  []int32 `josn:"timeProb"`  // 操作时间
}

var Robot robot

func InitRobotConfig(path string) {
	if path == "" {
		path = "./conf/robot.json"
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("LoadRobotConfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	initRobotConfig(result)
}

func initRobotConfig(cfg gjson.Result) {
	qiang := cfg.Get("qiangProb").Array()
	for _, v := range qiang {
		Robot.QiangProb = append(Robot.QiangProb, int32(v.Int()))
	}

	bet := cfg.Get("betProb").Array()
	for _, v := range bet {
		Robot.BetProb = append(Robot.BetProb, int32(v.Int()))
	}

	time := cfg.Get("timeProb").Array()
	for _, v := range time {
		Robot.TimeProb = append(Robot.TimeProb, int32(v.Int()))
	}
}
