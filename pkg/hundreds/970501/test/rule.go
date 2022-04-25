package main

import (
	"game_LaBa/brsgj/model"
	"math/rand"
)

const (
	PROB_BASE       = 10000
	BET_AREA_LENGHT = 8
)

// 下注区域
var BetArr = [BET_AREA_LENGHT]int64{}

func bet(tr *result) {

	if Conf.IsOpenRand {
		for i := 0; i < BET_AREA_LENGHT; i++ {
			betGold := int64(rand.Intn(10)) * Conf.Bottom
			BetArr[i] += betGold
			tr.endGold -= betGold
		}
	} else {
		for i, v := range Conf.BetArr {
			BetArr[i] += v * Conf.Bottom
		}
	}

}

func shake(tr *result) (element model.Elements, isRand bool) {
	roomProbNow := Conf.RoomProb
	goodLuck := model.GoodlucksAll.Find(int64(roomProbNow))
	element, tr.start = goodLuck.Rand(nil).Handle(nil)
	if element != nil {
		// 中特殊奖项
		return
	}

	// 根据返奖率进行计算

	back := goodLuck.Backs.Rand()
	// fmt.Println("返奖率 ===  ", back)
	backProbMin, backProbMax := int64(back.Min), int64(back.Max)

	backProb := [16]int64{}

	var allBet int64
	for _, v := range BetArr {
		allBet += v
	}

	// fmt.Println("allBet = ", allBet)
	for id, v := range BetArr {
		ele := model.ElementsAll.GetById(model.ElementType(id), nil)
		backProb[id] = int64(float64(v*int64(ele.OddsMax.Odds)) / float64(allBet) * 100)                 // 大倍率
		backProb[id+BET_AREA_LENGHT] = int64(float64(v*int64(ele.OddsMin.Odds)) / float64(allBet) * 100) // 小倍率
	}

	// fmt.Println("backProb = ", backProb)
	var wait []int
	for id, v := range backProb {
		if v >= backProbMin && v < backProbMax {
			wait = append(wait, id)
		}
	}

	// fmt.Println("wait = ", wait)
	if len(wait) != 0 {
		randIndex := wait[rand.Intn(len(wait))]
		var ismax bool
		if randIndex >= BET_AREA_LENGHT {
			randIndex -= BET_AREA_LENGHT
			ismax = false
		} else {
			ismax = true
		}
		element = model.Elements{model.ElementsAll.GetById(model.ElementType(randIndex), &ismax)}
	} else {
		tr.notInWaitTimes++
		element = model.Elements{model.ElementsAll.Rand(model.ElementType(-1))}
		isRand = true
	}

	// fmt.Println()
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
func calcWinGold(shakeResult model.Elements) (winGold int64) {
	temp := shakeResult
	if len(shakeResult) > 1 {
		// 中了大三元/大四喜
		temp = shakeResult[1:]
	}
	for _, sr := range temp {
		if sr.IsMax {
			winGold += BetArr[sr.Id] * int64(sr.OddsMax.Odds)
		} else {
			winGold += BetArr[sr.Id] * int64(sr.OddsMin.Odds)
		}
	}
	// fmt.Printf("winGold =%d \n\n", winGold)
	return
}

func reset() {
	BetArr = [8]int64{}
}
