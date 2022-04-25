package main

import (
	"game_LaBa/benzbmw/config"
	"game_LaBa/benzbmw/model"
	"math/rand"
)

const (
	PROB_BASE = 10000
)

// 下注区域
var BetArr = [12]int64{}

func bet(tr *result) {
	for i := 0; i < 12; i++ {
		betGold := int64(rand.Intn(10)) * Conf.Bottom
		BetArr[i] += betGold
		tr.endGold -= betGold
	}
}

func shake(tr *result) (element model.ElemBases) {
	roomProbNow := Conf.RoomProb
	winCtrl := config.BenzBMWConf.WinCtrl.Find(int(roomProbNow))

	surpriseProb := model.Rand(PROB_BASE)
	// 中特殊奖项
	if surpriseProb <= winCtrl.Surprise.Prob {
		if model.Rand(PROB_BASE) <= winCtrl.Surprise.Three {
			// 中大三元
			element = model.ElemBases{model.ElemThree}
			// threeElem:=fmt.Sprintf("%x",winCtrl.Surprise.ThreeCtrl.Rand().ElemType)
			element = append(element, model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.ThreeCtrl.Rand().ElemType))...)
		} else {
			// 中大四喜
			element = model.ElemBases{model.ElemFour}
			// fourElem:=fmt.Sprintf("%x",winCtrl.Surprise.FourCtrl.Rand().Color)
			element = append(element, model.ElemShakeProbSlice.FindWithType(model.ElementType(winCtrl.Surprise.FourCtrl.Rand().Color))...)
		}
	} else {
		element = getLastResult(winCtrl, tr)
	}

	// element 就是最后得中奖结果
	return
}

// 摇最后一个元素
func getLastResult(winCtrl *config.WinControl, tr *result) (element model.ElemBases) {
	allBet := getAllBet()
	oddsList := model.GetOdds()
	model.ReverseOdds(oddsList)
	// fmt.Println("oddsList = ", oddsList)
	// fmt.Println("config.BenzBMWConf.Cars = ", config.BenzBMWConf.Cars)
	var waitCheck []int
	var backProb = [12]int64{} // 返奖率
	waitCheck = nil

	if allBet != 0 { // 计算返奖率
		for index, v := range BetArr {
			backProb[index] = int64(float64(v*int64(oddsList[index])) / float64(allBet) * float64(100))
		}
	}

	// fmt.Println("BetArr = ", BetArr)
	// fmt.Println("backProb = ", backProb)
	back := winCtrl.Back.Rand()
	backMin, backMax := int64(back.Min), int64(back.Max)
	// fmt.Printf("backMin = %v ,backMax  = %v \n", backMin, backMax)
	for index, v := range backProb {
		if v >= backMin && v < backMax {
			waitCheck = append(waitCheck, index)
		}
	}
	if len(waitCheck) != 0 {
		// 这里随机选择一个进行返回
		index := waitCheck[rand.Intn(len(waitCheck))]
		element = model.ElemBases{config.BenzBMWConf.Cars.GetByID(12 - index - 1)}
		// fmt.Println("waitCheck = ", waitCheck)
		// fmt.Println()
		tr.notInWaitTimes++
		return
	}
	element = model.ElemBases{config.BenzBMWConf.Cars.RandResult(model.ElementTypeNil, true)}
	return
}

// 获取总下注
func getAllBet() (result int64) {
	for _, v := range BetArr {
		result += v
	}
	return
}

// 计算赢得金币
func calcWinGold(shakeResult model.ElemBases) (winGold int64) {
	if len(shakeResult) > 1 {
		// 中了大三元/大四喜
		shakeResult = shakeResult[1:]
	}
	for _, sr := range shakeResult {
		winGold += BetArr[sr.BetIndex] * int64(sr.Odds)
	}
	return
}

func reset() {
	BetArr = [12]int64{}
}
