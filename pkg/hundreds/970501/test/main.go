package main

import (
	"game_LaBa/brsgj/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitBenzBMWConf("../config/brsgj.json")
	config.LoadRobot("../config/robot.json")
	test()
}

func test() {
	tr := &result{}
	tr.friut = &friut{}
	tr.destNumTimes = int(Conf.TestTimes)
	tr.name = "百人水果机"
	tr.roomProb = int(Conf.RoomProb)
	tr.tax = int(Conf.Tax)
	tr.beginGold = Conf.InitGold
	tr.endGold = tr.beginGold

	for i := int64(0); i < Conf.TestTimes; i++ {
		if tr.endGold < 0 || tr.endGold < Conf.Bottom {
			// 结束测试
			break
		}
		bet(tr)
		result, isRand := shake(tr)
		tr.count(result, isRand)
		reset()
	}

	tr.calc()
	tr.println()
}
