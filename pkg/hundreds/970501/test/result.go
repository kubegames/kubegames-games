package main

import (
	"fmt"
	"game_LaBa/brsgj/model"
	"os"
	"time"
)

type friut struct {
	// 0:bar;1:77;2:双星;3:西瓜;4:铃铛;5:橘子;6：柠檬;7:苹果
	appleMax      int // 苹果
	orangeMax     int // 橘子
	watermelonMax int
	starMax       int
	bellMax       int
	lemonMax      int
	barMax        int
	sevenMax      int

	appleMin      int // 苹果
	orangeMin     int // 橘子
	watermelonMin int
	starMin       int
	bellMin       int
	lemonMin      int
	barMin        int
	sevenMin      int
}

type result struct {
	*friut

	start    int    // 中开火车时的开始下标
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
	ThreeBig   int // 大三元
	ThreeSmall int // 小三元
	SlamBig    int // 大满贯
	SlamSmall  int // 小满贯
	FoisonBig  int // 大丰收
	Train      int // 开火车

	SpecAllWin int64 // 特殊游戏总盈利

	// ------------分割线-----------
	beginGold   int64   // 初始货币
	endGold     int64   // 结束货币
	changeGold  int64   // 货币变化
	tax         int     // 税点
	winWin      int64   // 赢局赢得金额
	winBalance  float64 // 赢局平均赢额(赢局统计中，平均每局赢的额度)
	lossLoss    int64   // 输局输得金额
	lossBalance float64 // 输局平均输额(输局统计中，平均每局输的额度)
	allIn       int64   // 总投入(一切出账均纳入记数，包括初始押注、中途追加、结算给出等)
	allWin      int64   // 总收益(一切入账均纳入记数，包括押注退还、结算除税收益等)
	allTax      int64   // 总抽税(所有赢局最终赢额×税点)
	allWinProb  float64 // 总益率（总收益/总投入×100%）
	randWinGold int64   // 随机时赢的钱

	notInWaitTimes int64 // 不在wait里的次数
}

func (r *result) calc() {
	r.winProb = float64(r.winTimes) / float64(r.actualNumTimes) * 100
	r.endGold += r.allWin
	r.changeGold = r.endGold - r.beginGold
	r.allWinProb = float64(r.allWin) / float64(r.allIn) * 100
	r.winBalance = float64(r.winWin) / float64(r.winTimes) * 100
	r.lossBalance = float64(r.lossLoss) / float64(r.lossTimes) * 100
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
大满贯次数:%v
大丰收次数:%v
大三元次数:%v
开火车次数:%v
小三元次数:%v
小满贯次数:%v
特殊游戏总次数：%v
特殊游戏总返奖：%v

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

随机开奖赢的钱:%v

未找到合适返奖率次数:%v
`, r.name,
		r.roomProb,
		// ------------分割线-----------
		r.destNumTimes,
		r.actualNumTimes,
		r.winTimes,
		r.lossTimes,
		r.balanceTimes,
		r.winProb,
		// ------------分割线-----------
		r.SlamBig,
		r.FoisonBig,
		r.ThreeBig,
		r.Train,
		r.ThreeSmall,
		r.SlamSmall,

		r.SlamBig+r.FoisonBig+r.ThreeBig+r.Train+r.ThreeSmall+r.SlamSmall,
		r.SpecAllWin,

		// ------------分割线-----------
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
		r.randWinGold,
		r.notInWaitTimes,
	)

	str = r.friut.write2log(str)
	tt := time.Now().Format("20060102150405")
	file, err := os.Create(fmt.Sprintf("%v-%s.txt", r.name, tt))
	if err != nil {
		panic(err)
	}
	defer file.Close()
	file.WriteString(str)
}

func (tr *result) count(element model.Elements, isRand bool) {
	tr.actualNumTimes++
	betGold := getAllBet()
	tr.allIn += betGold
	// 抽税前
	// fmt.Printf("开奖 = %d    大？%v\n", element[0].Id, element[0].IsMax)
	winGold1 := calcWinGold(element)
	tr.friut.handleCars(element)
	if len(element) > 1 {
		tr.SpecAllWin += winGold1
	}
	// 抽税后
	winGold2 := winGold1 * (100 - int64(tr.tax)) / 100
	tr.allWin += winGold2
	if isRand {
		tr.randWinGold += winGold2
	}

	// 总抽税
	tr.allTax += winGold1 - winGold2

	diff := winGold1 - betGold
	switch {
	case diff > 0:
		tr.winTimes++
		tr.winWin += winGold1
	case diff < 0:
		tr.lossTimes++
		tr.lossLoss += -1 * diff
	default:
		tr.balanceTimes++
	}

	switch element[0].Id {
	case model.GoodluckTypeFoisonBigID:
		tr.FoisonBig++
	case model.GoodluckTypeThreeBigID:
		tr.ThreeBig++
	case model.GoodluckTypeThreeSmallID:
		tr.ThreeSmall++
	case model.GoodluckTypeSlamBigID:
		tr.SlamBig++
	case model.GoodluckTypeSlamSmallID:
		tr.SlamSmall++
	case model.GoodluckTypeTrainID:
		tr.Train++
	}

	// tr.cars.handleCars(element)
}

func (c *friut) handleCars(element model.Elements) {
	if len(element) != 1 {
		return
	}
	switch element[0].Id {
	case model.ElementTypeBar:
		if element[0].IsMax {
			c.barMax++
		} else {
			c.barMin++
		}
	case model.ElementTypeSeven2:
		if element[0].IsMax {
			c.sevenMax++
		} else {
			c.sevenMin++
		}
	case model.ElementTypeStar2:
		if element[0].IsMax {
			c.starMax++
		} else {
			c.starMin++
		}
	case model.ElementTypeWatermelon:
		if element[0].IsMax {
			c.watermelonMax++
		} else {
			c.watermelonMin++
		}
	case model.ElementTypeBell:
		if element[0].IsMax {
			c.bellMax++
		} else {
			c.bellMin++
		}
	case model.ElementTypeOrange:
		if element[0].IsMax {
			c.orangeMax++
		} else {
			c.orangeMin++
		}
	case model.ElementTypeLemon:
		if element[0].IsMax {
			c.lemonMax++
		} else {
			c.lemonMin++
		}
	case model.ElementTypeApple:
		if element[0].IsMax {
			c.appleMax++
		} else {
			c.appleMin++
		}
	}
}

func (c friut) write2log(str string) string {
	val := fmt.Sprintf(
		`
		
------------分割线-----------
大bar：%v
大77：%v
大双星:%v
大西瓜:%v
大铃铛:%v
大橘子:%v
大柠檬:%v
大苹果:%v

小bar：%v
小77：%v
小双星:%v
小西瓜:%v
小铃铛:%v
小橘子:%v
小柠檬:%v
小苹果:%v
		`,
		c.barMax,
		c.sevenMax,
		c.starMax,
		c.watermelonMax,
		c.bellMax,
		c.orangeMax,
		c.lemonMax,
		c.appleMax,

		c.barMin,
		c.sevenMin,
		c.starMin,
		c.watermelonMin,
		c.bellMin,
		c.orangeMin,
		c.lemonMin,
		c.appleMin,
	)
	return str + val
}
