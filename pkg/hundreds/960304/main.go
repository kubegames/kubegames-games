package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_poker/960304/config"
	"go-game-sdk/example/game_poker/960304/frameinter"
	"go-game-sdk/example/game_poker/960304/game"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LoadBaiJiaLeConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.LongHuRoom{})
	room.Run()
}
