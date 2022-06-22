package main

import (
	"github.com/kubegames/kubegames-games/pkg/slots/970102/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970102/gamelogic"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
)

func main() {
	config.Load()
	room := room.NewRoom(&gamelogic.Server{})
	room.Run()
}
