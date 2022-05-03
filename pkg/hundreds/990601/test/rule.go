package main

import (
	"game_LaBa/birdAnimal/config"
	"game_LaBa/birdAnimal/model"
	bridanimal "game_LaBa/birdAnimal/msg"
	"math/rand"
	"sort"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

const (
	// 下注区飞禽的index
	BIRD_BET_INDEX int = 8
	// 下注区走兽的index
	ANIMAL_BET_INDEX int = 9

	BET_AREA_LENGTH = 12

	PROB_BASE = 10000

	GOLD_SHARK_ID   = 2
	SILVER_SHARK_ID = 3

	ALL_KILL_ID = 12
	ALL_PAY_ID  = 13

	// 历史纪录长度
	TREND_LENGTH = 10
)

var TotalBet = [12]int64{}

var BirdBetArea = []int{0, 1, 6, 7}
var AnimalBetArea = []int{4, 5, 10, 11}

// 游戏下注
func bet(tr *TestResult, i int64) {
	var betArea = -1

	betSharkProb := rand.Intn(PROB_BASE)
	if int32(betSharkProb) < Conf.SharkBetProb {
		// 下鲨鱼区
		if betSharkProb < betSharkProb/2 {
			betArea = GOLD_SHARK_ID
		} else {
			betArea = SILVER_SHARK_ID
		}
		TotalBet[betArea] += Conf.BetGold
		Conf.TakeGold -= Conf.BetGold
		tr.AllBet += Conf.BetGold
	}

	for i, v := range Conf.BetArea {
		// if v <= 0 {
		// 	continue
		// }
		// TODO:下注倍率随机
		v = int64(rand.Intn(16))
		if i <= 1 {
			TotalBet[i] += v * Conf.BetGold
			Conf.TakeGold -= v * Conf.BetGold
			tr.AllBet += v * Conf.BetGold
		} else {
			TotalBet[i+2] += v * Conf.BetGold
			Conf.TakeGold -= v * Conf.BetGold
			tr.AllBet += v * Conf.BetGold
		}
	}
	// betBird(tr)
	// betAnimal(tr)
	// log.Traceln("TotalBet=", TotalBet)
	tr.TestTimes = i
}

// func betBird(tr *TestResult) {
// 	if Conf.BirdAreaNum < 0 {
// 		Conf.BirdAreaNum = 0
// 	} else if Conf.BirdAreaNum > 4 {
// 		Conf.BirdAreaNum = 4
// 	}

// 	var temp []int
// 	temp = BirdBetArea
// 	for i := 0; i < int(Conf.BirdAreaNum); i++ {
// 		if Conf.TakeGold < Conf.BetGold {
// 			return
// 		}
// 		TotalBet[rand.Intn(len(temp))] += Conf.BetGold
// 		Conf.TakeGold -= Conf.BetGold
// 		tr.AllBet += Conf.BetGold
// 		temp = rmSlice(i, temp)
// 	}
// }
// func betAnimal(tr *TestResult) {
// 	if Conf.AnimalAreaNum <= 0 {
// 		Conf.AnimalAreaNum = 0
// 	} else if Conf.AnimalAreaNum > 4 {
// 		Conf.AnimalAreaNum = 4
// 	}

// 	var temp []int
// 	temp = AnimalBetArea
// 	for i := 0; i < int(Conf.AnimalAreaNum); i++ {
// 		if Conf.TakeGold < Conf.BetGold {
// 			return
// 		}
// 		TotalBet[rand.Intn(len(temp))] += Conf.BetGold
// 		Conf.TakeGold -= Conf.BetGold
// 		tr.AllBet += Conf.BetGold
// 		temp = rmSlice(i, temp)
// 	}
// }

func rmSlice(i int, temp []int) []int {
	if i < 0 || i >= len(temp) {
		return temp
	}
	return append(temp[:i], temp[i:]...)
}

// 开奖控制
func shakeRW(tr *TestResult) (result []*model.Element) {
	nowRoomProb := Conf.RoomProb
	shakePolicy := config.BirdAnimaConfig.PolicyTree.Find(nowRoomProb)
	var element, lastElement *model.Element

	var isAll bool
	all := randBase()

	// 1.通杀通赔判定
	if all <= shakePolicy.All.AllPay {
		isAll = true
		// 出通赔
		element = config.BirdAnimaConfig.BirdAnimals.GetByID(ALL_PAY_ID)
	} else if all > shakePolicy.All.AllPay && all <= (shakePolicy.All.AllPay+shakePolicy.All.AllKill) {
		isAll = true
		// 出通杀
		element = config.BirdAnimaConfig.BirdAnimals.GetByID(ALL_KILL_ID)
	}
	if element != nil && isAll {
		result = append(result, element)
		return
	}
	// 2.免费游戏判定
	free := randBase()
	if free <= shakePolicy.Free.Open {
		// 开出免费游戏
		if randBase() <= shakePolicy.Free.GoldShark {
			// 开出金鲨
			element = config.BirdAnimaConfig.BirdAnimals.GetByID(GOLD_SHARK_ID)
		} else {
			// 开出银鲨
			element = config.BirdAnimaConfig.BirdAnimals.GetByID(SILVER_SHARK_ID)
		}
	}
	if element != nil {
		result = append(result, element)
	}
	// 3.开奖结果计算----根据返奖率来结算[第二个]结果
	lastElement, _ = shakeResultRW(tr)
	if lastElement.ID == ALL_PAY_ID ||
		lastElement.ID == ALL_KILL_ID ||
		lastElement.ID == GOLD_SHARK_ID ||
		lastElement.ID == SILVER_SHARK_ID {
		log.Traceln("lastElement.ID=", lastElement.ID)
	}

	if lastElement == nil {
		panic("lastElement is nil")
	}
	result = append(result, lastElement)
	return
}

func reset() {
	TotalBet = [12]int64{}
	config.BirdAnimaConfig.BirdAnimals.Reset()
}

func randBase() int {
	return rand.Intn(PROB_BASE) + 1
}

func shakeResultRW(tr *TestResult) (*model.Element, int) {
	nowRoomProb := Conf.RoomProb
	// RECALC: // 重新赋值nowRoomProb进行计算

	shakePolicy := config.BirdAnimaConfig.PolicyTree.Find(nowRoomProb)
	var allBetGold int64
	for _, v := range TotalBet {
		allBetGold += v
	}

	oddsList := getRandOddsInfo()
	var backProb = [12]int64{} // 返奖率

	if allBetGold != 0 {
		birdPay := TotalBet[BIRD_BET_INDEX] * int64(oddsList[BIRD_BET_INDEX].Odds)
		// fmt.Printf("飞禽   %d&%d=%d\n", TotalBet[BIRD_BET_INDEX], int64(oddsList[BIRD_BET_INDEX].Odds), birdPay)
		animalPay := TotalBet[ANIMAL_BET_INDEX] * int64(oddsList[ANIMAL_BET_INDEX].Odds)
		// fmt.Printf("走兽   %d&%d=%d\n", TotalBet[ANIMAL_BET_INDEX], int64(oddsList[ANIMAL_BET_INDEX].Odds), animalPay)
		for i, v := range TotalBet {
			// 所有的鸟类要加上飞禽的赔付值
			switch i {
			case 0, 1, 6, 7: // 飞禽
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)+birdPay) / float64(allBetGold) * float64(100))
			case 4, 5, 10, 11: // 走兽
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)+animalPay) / float64(allBetGold) * float64(100))
			default:
				backProb[i] = int64(float64(v*int64(oddsList[i].Odds)) / float64(allBetGold) * float64(100))
			}
		}
	}

	// fmt.Printf("返奖率列表=[燕子=%d ; 鸽子=%d ; 兔子=%d ; 猴子=%d ; 孔雀=%d ; 老鹰=%d ; 熊猫=%d ; 狮子=%d ;]\n",
	// 	backProb[0],
	// 	backProb[1],
	// 	backProb[4],
	// 	backProb[5],
	// 	backProb[6],
	// 	backProb[7],
	// 	backProb[10],
	// 	backProb[11],
	// )

	back := shakePolicy.Back.Rand()
	nowBackProbMin, nowBackProbMax := back.Min, back.Max
	var waitCheck []int
	for i, v := range backProb {
		if v >= int64(nowBackProbMin) &&
			v < int64(nowBackProbMax) &&
			i != BIRD_BET_INDEX &&
			i != ANIMAL_BET_INDEX &&
			i != GOLD_SHARK_ID &&
			i != SILVER_SHARK_ID &&
			i != ALL_KILL_ID &&
			i != ALL_PAY_ID {
			// 去除金鲨/银鲨/飞禽/走兽下注区
			waitCheck = append(waitCheck, i)
		}
	}
	// 满足返奖率的待选元素
	if len(waitCheck) != 0 {
		// 这里随机选择一个进行返回
		index := waitCheck[rand.Intn(len(waitCheck))]
		// return config.BirdAnimaConfig.BirdAnimals.GetByID(index)
		ele := config.BirdAnimaConfig.BirdAnimals.GetByIDResult(index)
		return ele, ele.RandSubId()
	}
	tr.NotInBack++
RMLABEL:
	// 此处往上查找都未找到，随机返回一个结果
	ele, id := config.BirdAnimaConfig.BirdAnimals.RandResult(-1, false)
	if ele.ID == BIRD_BET_INDEX ||
		ele.ID == ANIMAL_BET_INDEX ||
		ele.ID == GOLD_SHARK_ID ||
		ele.ID == SILVER_SHARK_ID ||
		ele.ID == ALL_KILL_ID ||
		ele.ID == ALL_PAY_ID {
		goto RMLABEL
	}
	return ele, id
	// 系统开奖返奖率
	// 作弊率	开奖返奖率
	// 3000		0-50
	// 2000		50-70
	// 1000		70-90  -----[中间]
	// -1000	90-110
	// -2000	110-130
	// -3000	130-150
	// if nowRoomProb == 1000 {
	// RMLABEL:
	// 	// 此处往上查找都未找到，随机返回一个结果
	// 	ele, id := config.BirdAnimaConfig.BirdAnimals.RandResult(-1, false)
	// 	if ele.ID == BIRD_BET_INDEX ||
	// 		ele.ID == ANIMAL_BET_INDEX ||
	// 		ele.ID == GOLD_SHARK_ID ||
	// 		ele.ID == SILVER_SHARK_ID ||
	// 		ele.ID == ALL_KILL_ID ||
	// 		ele.ID == ALL_PAY_ID {
	// 		goto RMLABEL
	// 	}
	// 	return ele, id
	// } else if nowRoomProb > 1000 {
	// 	nowRoomProb -= 1000
	// } else {
	// 	nowRoomProb += 1000
	// 	if nowRoomProb == 0 {
	// 		nowRoomProb = 1000
	// 	}
	// }
	// goto RECALC
}

func getRandOddsInfo() model.RandomOddss {
	result := make(model.RandomOddss, 0)

	for _, v := range config.BirdAnimaConfig.BirdAnimals {
		// 返回随机赔率，忽略掉通赔通杀
		if !(v.EType != model.ETypeAllKill && v.EType != model.ETypeAllPay) {
			continue
		}

		result = append(result, &bridanimal.RandomOdds{
			ID:   int32(v.ID),
			Odds: int32(v.OddsNow),
		})
	}

	result = result[:12]
	sort.Sort(result)
	return result
}
