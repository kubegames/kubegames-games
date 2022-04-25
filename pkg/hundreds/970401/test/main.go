package main

import (
	"game_LaBa/benzbmw/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitBenzBMWConf("../config/benzbmw.json")
	config.LoadRobot("../config/robot.json")
	test()
}

func test() {
	tr := &result{}
	tr.cars = &cars{}
	tr.destNumTimes = int(Conf.TestTimes)
	tr.name = "奔驰宝马"
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
		result := shake(tr)
		tr.count(result)
		reset()
	}

	tr.calc()
	tr.println()
}
