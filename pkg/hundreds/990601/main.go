package main

import (
	game_frame "go-game-sdk"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/slots/990601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/990601/gamelogic"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	config.LoadBirdAnimalConfig("")
	config.InitRobot("")
	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
