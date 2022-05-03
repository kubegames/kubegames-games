package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960211/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960211/frameinter"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitConfig("")
	config.InitRobotConfig("")
	room := room.NewRoom(&frameinter.Pai9Room{})
	room.Run()
}
