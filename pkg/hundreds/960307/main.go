package main

import (
	"game_poker/BRZJH/config"
	"game_poker/BRZJH/frameinter"
	"game_poker/BRZJH/game"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func main() {
	log.Traceln("### VER:  1.0.2  ")
	rand.Seed(time.Now().UnixNano())
	config.LoadBRZJHConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.BaiRenZhaJinHuaRoom{})
	room.Run()
}
