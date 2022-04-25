package main

import (
	//"game_LaBa/birdAnimal/config"

	"encoding/json"
	"game_LaBa/birdAnimal/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LoadBirdAnimalConfig("../config/birdanimal.json")
	config.BirdAnimaConfig.BirdAnimals.RandOddsAndProb()
	config.InitRobot("../config/robot.json")
	test()
}

func test() {
	tr := &TestResult{}
	for i := int64(0); i < Conf.TestTimes; i++ {
		if tr.TestTimes < i-1 || Conf.TakeGold < 0 || Conf.TakeGold < Conf.BetGold {
			// 测试册书
			break
		}
		config.BirdAnimaConfig.BirdAnimals.RandOddsAndProb()
		bet(tr, i)
		result := shakeRW(tr)

		if len(result) == 1 {
			switch result[0].ID {
			case 0, 1, 4, 5, 6, 7, 10, 11, 12, 13:
			default:
				bts, _ := json.Marshal(result)
				panic(string(bts))
			}
		}
		tr.count(result)
		reset()
	}
	tr.calc()
	tr.write2file()
}
