package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_MaJiang/960205/glogic"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func main() {
	log.Debugf("%v --- 二八杠启动", time.Now())
	rand.Seed(time.Now().UnixNano())
	//conf.ConfigInit()
	room := game_frame.NewRoom(&glogic.ErBaGangTable{})
	room.Run()
}
