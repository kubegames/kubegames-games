package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/hundreds/960302/config"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960302/frameinter"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960302/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/hundreds"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LoadLongHuConfig()
	game.RConfig.LoadLabadCfg()
	room := room.NewRoom(&frameinter.LongHuRoom{})
	room.Run()
}
