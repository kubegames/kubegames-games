// 奔驰宝马
package main

import (
	game_frame "go-game-sdk"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/slots/970501/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970501/gamelogic"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.InitBenzBMWConf("")
	config.LoadRobot("")
	room := game_frame.NewRoom(&gamelogic.Room{})
	room.Run()
}
