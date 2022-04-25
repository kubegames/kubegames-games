package main

import (
	"game_frame_v2/game/logic"
	"game_poker/BRTB/config"
	"game_poker/BRTB/frameinter"
	"game_poker/BRTB/game"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LoadBaiJiaLeConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.BRTBRoom{})
	room.Run()
}
