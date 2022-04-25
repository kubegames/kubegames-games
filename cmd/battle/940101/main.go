package main

import (
	"math/rand"
	_ "net/http/pprof"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/940101/config"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/gamelogic"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.ErRenMaJiang.LoadErRenMaJiangConfig()
	config.TestErRenMaJiang.TestLoadErRenMaJiangConfig()
	room := room.NewRoom(&gamelogic.LogicInterFace{})
	room.Run()
}
