package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960205/glogic"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	log.Debugf("%v --- 二八杠启动", time.Now())
	rand.Seed(time.Now().UnixNano())
	room := room.NewRoom(&glogic.ErBaGangTable{})
	room.Run()
}
