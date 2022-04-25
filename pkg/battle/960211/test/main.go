package main

import (
	"game_poker/pai9/config"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitConfig("../conf/pai9.json")
	config.InitRobotConfig("../conf/robot.json")
	initConf("")

	initUser()

	re := &result{
		name:      "牌九",
		roomProb:  Conf.RoomProb,
		setTimes:  Conf.Times,
		beginGold: Conf.BeginGold,
	}
	for i := 0; i < Conf.Times; i++ {
		qiang()
		bet()
		dealpoker()
		compare()
		re.count()
		reset()
		if !check() {
			break
		}
		// fmt.Println()
	}
	re.endCalc()
	re.println()
}
