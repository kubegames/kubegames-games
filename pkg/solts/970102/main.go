package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/970102/config"
	"go-game-sdk/example/game_LaBa/970102/gamelogic"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config.Load()
	room := game_frame.NewRoom(&gamelogic.Server{})
	room.Run()
}
