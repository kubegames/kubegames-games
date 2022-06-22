//拉霸

package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/xiaomali"
	"github.com/kubegames/kubegames-games/pkg/slots/990101/gamelogic"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LBConfig.LoadLabadCfg()
	xiaomali.XMLConfig.LoadXiaoMaLiCfg()

	room := room.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
