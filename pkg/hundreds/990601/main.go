package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/990601/config"
	"go-game-sdk/example/game_LaBa/990601/gamelogic"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config.LoadBirdAnimalConfig("")
	config.InitRobot("")
	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
