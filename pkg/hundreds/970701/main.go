package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_poker/saima/config"
	"go-game-sdk/example/game_poker/saima/frameinter"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.Load()
	room := game_frame.NewRoom(&frameinter.Server{})
	room.Run()
}
