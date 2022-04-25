package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/fishing/yaoqianshubuyu/config"
	"github.com/kubegames/kubegames-games/pkg/fishing/yaoqianshubuyu/server"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	config.Load()

	rand.Seed(time.Now().UnixNano())
	room := room.NewRoom(&server.Server{})
	room.Run()
}
