// 奔驰宝马
package main

import (
	"game_LaBa/benzbmw/config"
	"game_LaBa/benzbmw/gamelogic"
	game_frame "game_frame_v2/game/logic"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitBenzBMWConf("")
	config.LoadRobot("")
	room := game_frame.NewRoom(&gamelogic.BenzBMWRoom{})
	room.Run()
}
