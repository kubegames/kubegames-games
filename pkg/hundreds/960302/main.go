package main

import (
	"game_frame_v2/game/logic"
	"game_poker/longhu/config"
	"game_poker/longhu/frameinter"
	"game_poker/longhu/game"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LoadLongHuConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.LongHuRoom{})
	room.Run()
}
