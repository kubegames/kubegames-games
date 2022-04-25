package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"time"

	"github.com/tidwall/gjson"
)

type RobotConfig struct {
	BetGapMin    int32 `json:"betGapMin"` // 最小下注时间间隔（单位：ms）
	BetGapMax    int32 `json:"betGapMax"` // 最大下注时间间隔（单位：ms）
	BetGapNow    int   // 当前的时间间隔
	BetNumUpline int32 `json:"betNumUpline"` // 下注上线，超出不在下注

	BetAreaProb  betAreaProbs  `json:"betAreaProb"`  // 下注区域的选择
	BetModChoose betModChooses `json:"betModChoose"` // 下注筹码区的选择

	EvictGold [][]int64 // 驱逐金额
}

type betModChoose struct {
	ModeMin int64  `json:"min"`     // 倍数小
	ModeMax int64  `json:"max"`     // 倍数大
	BetArea [5]int `json:"betArea"` // 下注筹码区
}

type betModChooses [5]*betModChoose

// 筹码区选择
func (bb betModChooses) Rand(goldNow int64, bottom int64) int {
	var target *betModChoose
	mod := goldNow / bottom
	for i, v := range bb {
		if i != len(bb)-1 {
			if mod >= v.ModeMin && mod < v.ModeMax {
				target = v
			}
		} else {
			if mod >= v.ModeMin {
				target = v
			}
		}
	}
	if target == nil {
		target = bb[rand.Intn(len(bb))]
	}

	var allWeight int
	for _, v := range target.BetArea {
		allWeight += v
	}
	randWeight := rand.Intn(allWeight) + 1
	for i, v := range target.BetArea {
		if randWeight <= v {
			return i
		}
		randWeight -= v
	}
	return rand.Intn(len(target.BetArea))
}

type betAreaProb struct {
	Prob int `json:"prob"`
	Line *struct {
		Norm            int32 `json:"norm"` // 平衡线
		GreaterThanNorm struct {
			LessProb int32 `json:"lessProb"` // 下筹码较少区域权重
			MoreProb int32 `json:"moreProb"` // 下筹码较少区域权重
		} `json:"greaterThanNorm"` // 超过平衡线
		LessThanNorm struct {
			LessProb int32 `json:"lessProb"` // 下筹码较少区域权重
			MoreProb int32 `json:"moreProb"` // 下筹码较少区域权重
		} `json:"lessThanNorm"` //  低于平衡线
	} `json:"line"`
}

type betAreaProbs [3]betAreaProb

// 随机下注区域
func (bb betAreaProbs) Rand(betGold [12]int64) int {
	var allWeight int
	for _, v := range bb {
		allWeight += v.Prob
	}
	randWeight := rand.Intn(allWeight) + 1
	for i, v := range bb {
		if randWeight <= v.Prob {
			return v.handle(i, betGold)
		}
		randWeight -= v.Prob
	}
	return rand.Intn(12)
}

func (b betAreaProb) handle(i int, betGold [12]int64) int {

	switch i {
	default: // 0//普通下注区随机一个返回
		min, max := betGold[0], betGold[0]
		var minIndex int
		for i, v := range betGold {
			if v <= min {
				min = v
				minIndex = i
			}
			if v >= max {
				max = v
			}
		}
		balance := max - min
		if balance > int64(b.Line.Norm) { // 大于
			prob := rand.Intn(int(b.Line.GreaterThanNorm.LessProb) + int(b.Line.GreaterThanNorm.MoreProb))
			if prob <= int(b.Line.GreaterThanNorm.LessProb) {
				// 小区域  选一个
				return minIndex
			} else {
			LOOP1:
				if index := rand.Intn(12); index != minIndex {
					return index
				}
				goto LOOP1
			}
		} else {
			prob := rand.Intn(int(b.Line.LessThanNorm.LessProb) + int(b.Line.LessThanNorm.MoreProb))
			if prob <= int(b.Line.LessThanNorm.LessProb) {
				// 小区域  选一个
				return minIndex
			} else {
			LOOP2:
				if index := rand.Intn(12); index != minIndex {
					return index
				}
				goto LOOP2
			}
		}
	case 1: // 飞禽/走兽
		var min, max int64
		var minIndex, maxIndex int
		if betGold[8] > betGold[9] {
			min, max = betGold[9], betGold[8]
			minIndex = 9
			maxIndex = 8
		} else {
			min, max = betGold[8], betGold[9]
			minIndex = 8
			maxIndex = 9
		}

		balance := max - min
		if balance >= int64(b.Line.Norm) {
			prob := rand.Intn(int(b.Line.GreaterThanNorm.LessProb) + int(b.Line.GreaterThanNorm.MoreProb))
			if prob <= int(b.Line.GreaterThanNorm.LessProb) {
				// 小区域  选一个
				return minIndex
			}
			return maxIndex

		} else {
			prob := rand.Intn(int(b.Line.LessThanNorm.LessProb) + int(b.Line.LessThanNorm.MoreProb))
			if prob <= int(b.Line.LessThanNorm.LessProb) {
				// 小区域  选一个
				return minIndex
			}
			return maxIndex

		}
	case 2: // 鲨鱼
		if time.Now().Unix()%2 == 0 {
			return 2
		}
		return 3
	}

	return rand.Intn(12)
}

var Robot RobotConfig

func InitRobot(filePath string) {
	if filePath == "" {
		filePath = "./config/robot.json"
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("InitRobot Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	initRobot(result)

}

func initRobot(data gjson.Result) {
	Robot.BetNumUpline = int32(data.Get("betNumUpline").Int())
	Robot.BetGapMin = int32(data.Get("betGapMin").Int())
	Robot.BetGapMax = int32(data.Get("betGapMax").Int())

	bap := data.Get("betAreaProb").Array()
	Robot.BetAreaProb = betAreaProbs{}
	for i, v := range bap {
		var bb betAreaProb
		if err := json.Unmarshal([]byte(v.String()), &bb); err != nil {
			panic(err)
		}
		Robot.BetAreaProb[i] = bb
	}

	bmc := data.Get("betModChoose").Array()
	Robot.BetModChoose = betModChooses{}
	for i, v := range bmc {
		var bb betModChoose
		if err := json.Unmarshal([]byte(v.String()), &bb); err != nil {
			panic(err)
		}
		Robot.BetModChoose[i] = &bb
	}

	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold%d", i)
		robotgold := data.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}
		Robot.EvictGold = append(Robot.EvictGold, gold)
	}
	if len(Robot.EvictGold) == 0 {
		panic("must config robotgold.{{level}}")
	}
}

func (r *RobotConfig) Rand() {
	r.RandBetGap()
}

func (r *RobotConfig) RandBetGap() {
	min, max := compare(r.BetGapMin, r.BetGapMax)
	r.BetGapNow = rand.Intn(int(max-min)) + int(min)
}

func compare(a, b int32) (min, max int32) {
	if a >= b {
		min, max = b, a
	} else {
		min, max = a, b
	}
	return
}
