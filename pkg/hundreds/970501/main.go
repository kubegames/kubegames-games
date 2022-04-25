// 奔驰宝马
package main

import (
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/970501/config"
	"go-game-sdk/example/game_LaBa/970501/gamelogic"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitBenzBMWConf("")
	config.LoadRobot("")
	room := game_frame.NewRoom(&gamelogic.Room{})
	room.Run()
}
