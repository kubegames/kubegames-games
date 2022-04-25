package main

import (
	"fmt"
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_poker/960303/config"
	"go-game-sdk/example/game_poker/960303/frameinter"
	"go-game-sdk/example/game_poker/960303/game"
	"math/rand"
	"time"
)

func main() {

	fmt.Println("### VER:  2.0.6  ")
	rand.Seed(time.Now().UnixNano())
	config.LoadBRNNConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.LongHuRoom{})
	room.Run()
}
