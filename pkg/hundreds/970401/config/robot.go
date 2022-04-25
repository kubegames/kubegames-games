package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/tidwall/gjson"
)

type RobotConfig struct {
	BetGapMax int64 `json:"betGapMax"` // 下注时间间隔大（毫秒）
	BetGapMin int64 `json:"betGapMin"` // 下注时间间隔小（毫秒）

	Diff  int64 `json:"diff"`  // 差值
	Small int   `json:"small"` // 超过差值配置时，下小的区域

	EvictGold [][]int64
	BetTimes  BetTimeCtrls  `json:"betTimes"` // 下注次数控制
	BetGold   BetIndexCtrls `json:"betGold"`  // 下注金额控制(筹码控制)
	BetArea   BetAreaCtrls  `json:"betAreas"` // 下注区域控制
}

type BetTimeCtrl struct {
	Prob  int `json:"prob"`  // 权重
	Times int `json:"times"` // 次数
}

type BetTimeCtrls []BetTimeCtrl

func (b BetTimeCtrls) Rand() BetTimeCtrl {
	var allWeight int
	for _, v := range b {
		allWeight += v.Prob
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range b {
		if randWeight <= v.Prob {
			return v
		}
	}
	return b[rand.Intn(len(b))]
}

type BetIndexCtrl struct {
	Index int `json:"index"` // 下注索引（下标）
	Prob  int `json:"prob"`  // 权重
}

type BetIndexCtrls []*BetIndexCtrl

func (b BetIndexCtrls) Rand() *BetIndexCtrl {
	var allWeight int
	for _, v := range b {
		allWeight += v.Prob
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range b {
		if randWeight <= v.Prob {
			return v
		}
		randWeight -= v.Prob
	}
	return nil
}

type BetAreaCtrl struct {
	ElemType byte `json:"elemType"`
	Index    int  `json:"index"`
	Prob     int  `json:"prob"`
}

type BetAreaCtrls []*BetAreaCtrl

func (b BetAreaCtrls) Rand() *BetAreaCtrl {
	var allWeight int
	for _, v := range b {
		allWeight += v.Prob
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range b {
		if randWeight <= v.Prob {
			return v
		}
		randWeight -= v.Prob
	}
	return nil
}

var RobotConf RobotConfig

func LoadRobot(filepath string) {
	if filepath == "" {
		filepath = "./config/robot.json"
	}
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Println("LoadBenzBMW robot Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	initRobot(result)
	check(RobotConf)
}

func initRobot(cfg gjson.Result) {
	RobotConf.BetGapMax = int64(cfg.Get("betGapMax").Int())
	RobotConf.BetGapMin = int64(cfg.Get("betGapMin").Int())
	RobotConf.Diff = int64(cfg.Get("diff").Int())
	RobotConf.Small = int(cfg.Get("small").Int())

	for _, v := range cfg.Get("betTimes").Array() {
		btc := new(BetTimeCtrl)
		if err := json.Unmarshal([]byte(v.String()), btc); err != nil {
			panic(err)
		}
		RobotConf.BetTimes = append(RobotConf.BetTimes, *btc)
	}

	for _, v := range cfg.Get("betGold").Array() {
		bic := new(BetIndexCtrl)
		if err := json.Unmarshal([]byte(v.String()), bic); err != nil {
			panic(err)
		}
		RobotConf.BetGold = append(RobotConf.BetGold, bic)
	}

	for _, v := range cfg.Get("betAreas").Array() {
		bac := new(BetAreaCtrl)
		if err := json.Unmarshal([]byte(v.String()), bac); err != nil {
			panic(err)
		}
		RobotConf.BetArea = append(RobotConf.BetArea, bac)
	}

	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}
		RobotConf.EvictGold = append(RobotConf.EvictGold, gold)
	}
	if len(RobotConf.EvictGold) == 0 {
		panic("must config robotgold.{{level}}")
	}
}

func (rc RobotConfig) RandGap() int64 {
	min, max := compare(rc.BetGapMax, rc.BetGapMin)
	if min == max {
		return min
	}
	return int64(rand.Intn(int(max)-int(min))) + min
}

func compare(a, b int64) (min, max int64) {
	if a >= b {
		min, max = b, a
	} else {
		min, max = a, b
	}
	return
}

func check(robotConf RobotConfig) {
	for _, v := range robotConf.BetArea {
		if v.Index < 0 || v.Index > 11 {
			panic("index must be within the range of 0 to 11")
		}
	}

}
