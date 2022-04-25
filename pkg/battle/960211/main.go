package main

import (
	game_frame "game_frame_v2/game/logic"
	"game_poker/pai9/config"
	"game_poker/pai9/frameinter"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitConfig("")
	config.InitRobotConfig("")
	// game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.Pai9Room{})
	room.Run()
}
