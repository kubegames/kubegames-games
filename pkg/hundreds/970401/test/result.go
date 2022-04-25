package main

import (
	"fmt"
	"game_LaBa/benzbmw/model"
	"os"
	"time"
)

type cars struct {
	BenzRed   int
	BenzGreen int
	BenzBlack int

	BMWRed   int
	BMWGreen int
	BMWBlack int

	LexusRed   int
	LexusGreen int
	LexusBlack int

	VWRed   int
	VWGreen int
	VWBlack int
}

type result struct {
	*cars
	name     string // 游戏名称
	roomProb int    // 作弊率

	// ------------分割线-----------
	destNumTimes   int     // 设定局数
	actualNumTimes int     // 实际游戏局数
	winTimes       int     // 赢次数
	lossTimes      int     // 输次数
	balanceTimes   int     // 和次数
	winProb        float64 // 胜率（赢次数/实际局数*100%，4位小数）

	// ------------分割线-----------
	threeTimes int     // 大三元次数
	threeProb  float64 //  大三元概率
	fourTimes  int     // 大四喜次数
	fourProb   float64 // 大四喜概率

	// ------------分割线-----------
	beginGold   int64   // 初始货币
	endGold     int64   // 结束货币
	changeGold  int64   // 货币变化
	tax         int     // 税点
	winBalance  float64 // 赢局平均赢额(赢局统计中，平均每局赢的额度)
	lossBalance float64 // 输局平均输额(输局统计中，平均每局输的额度)
	allIn       int64   // 总投入(一切出账均纳入记数，包括初始押注、中途追加、结算给出等)
	allWin      int64   // 总收益(一切入账均纳入记数，包括押注退还、结算除税收益等)
	allTax      int64   // 总抽税(所有赢局最终赢额×税点)
	allWinProb  float64 // 总益率（总收益/总投入×100%）

	notInWaitTimes int64 // 不在wait里的次数
}

func (r *result) calc() {
	r.winProb = float64(r.winTimes) / float64(r.actualNumTimes) * 100
	r.endGold += r.allWin
	r.changeGold = r.endGold - r.beginGold
	r.allWinProb = float64(r.allWin) / float64(r.allIn) * 100
	r.threeProb = float64(r.threeTimes) / float64(r.actualNumTimes) * 100
	r.fourProb = float64(r.fourTimes) / float64(r.actualNumTimes) * 100
	r.winBalance = float64(r.allWin) / float64(r.winTimes) * 100
	r.lossBalance = float64(r.allIn) / float64(r.lossTimes) * 100
}
func (r result) println() {
	str := fmt.Sprintf(
		`
游戏名称:%v
作弊率:%v

------------分割线-----------
设定局数:%v
实际游戏局数:%v
赢次数:%v
输次数:%v
和次数:%v
胜率:%v

------------分割线-----------
大三元次数:%v
大三元概率:%v
大四喜次数:%v
大四喜概率:%v

------------分割线-----------
初始货币:%v
结束货币:%v
货币变化:%v
税点:%v
赢局平均赢额:%v
输局平均输额:%v
总投入:%v
总收益:%v
总抽税:%v
总益率:%v

未找到合适返奖率次数:%v
`, r.name,
		r.roomProb,
		r.destNumTimes,
		r.actualNumTimes,
		r.winTimes,
		r.lossTimes,
		r.balanceTimes,
		r.winProb,
		r.threeTimes,

		r.threeProb,
		r.fourTimes,
		r.fourProb,
		r.beginGold,
		r.endGold,
		r.changeGold,
		r.tax,
		r.winBalance,
		r.lossBalance,
		r.allIn,
		r.allWin,
		r.allTax,
		r.allWinProb,
		r.notInWaitTimes,
	)

	str = r.cars.write2log(str)
	tt := time.Now().Format("20060102150405")
	file, err := os.Create(fmt.Sprintf("%v-%s.txt", r.name, tt))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(str)
}

func (tr *result) count(element model.ElemBases) {
	tr.actualNumTimes++
	betGold := getAllBet()
	tr.allIn += betGold
	// 抽税前
	winGold1 := calcWinGold(element)
	// 抽税后
	winGold2 := winGold1 * (100 - int64(tr.tax)) / 100
	tr.allWin += winGold2

	// 总抽税
	tr.allTax += winGold1 - winGold2

	diff := winGold1 - betGold
	switch {
	case diff > 0:
		tr.winTimes++
	case diff < 0:
		tr.lossTimes++
	default:
		tr.balanceTimes++
	}

	switch len(element) {
	case 4: // 大三元
		tr.threeTimes++
	case 5: // 大四喜
		tr.fourTimes++
	default:

	}

	tr.cars.handleCars(element)
}

func (c *cars) handleCars(element model.ElemBases) {
	if len(element) != 1 {
		return
	}
	switch element[0].ElemType {
	case model.BenzRed:
		c.BenzRed++
	case model.BenzGreen:
		c.BenzGreen++
	case model.BenzBlack:
		c.BenzBlack++
	case model.BMWRed:
		c.BMWRed++
	case model.BMWGreen:
		c.BMWGreen++
	case model.BMWBlack:
		c.BMWBlack++
	case model.LexusRed:
		c.LexusRed++
	case model.LexusGreen:
		c.LexusGreen++
	case model.LexusBlack:
		c.LexusBlack++
	case model.VWRed:
		c.VWRed++
	case model.VWGreen:
		c.VWGreen++
	case model.VWBlack:
		c.VWBlack++
	}
}

func (c cars) write2log(str string) string {
	val := fmt.Sprintf(
		`
		
------------分割线-----------
奔驰红:%v,
奔驰绿:%v,
奔驰黑:%v,

宝马红:%v,
宝马绿:%v,
宝马黑:%v,

雷克红:%v,
雷克绿:%v,
雷克黑:%v,

大众红:%v,
大众绿:%v,
大众黑:%v
		`,
		c.BenzRed,
		c.BenzGreen,
		c.BenzBlack,
		c.BMWRed,
		c.BMWGreen,
		c.BMWBlack,
		c.LexusRed,
		c.LexusGreen,
		c.LexusBlack,
		c.VWRed,
		c.VWGreen,
		c.VWBlack,
	)
	return str + val
}
