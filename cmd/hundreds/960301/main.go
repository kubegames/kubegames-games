package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/config"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/frameinter"
	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/hundreds"
)

func main() {
	fmt.Println("### VER:  2.0.6  ")
	rand.Seed(time.Now().UnixNano())
	config.LoadRBWarConfig()
	game.RConfig.LoadLabadCfg()

	room := room.NewRoom(&frameinter.RBWarRoom{})
	room.Run()
}
