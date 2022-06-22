//拉霸

package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	"github.com/kubegames/kubegames-games/internal/pkg/labacom/xiaomali"
	"github.com/kubegames/kubegames-games/pkg/slots/990201/bibei"
	"github.com/kubegames/kubegames-games/pkg/slots/990201/gamelogic"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	gamelogic.LoadConfig()
	config.LBConfig.LoadLabadCfg()
	xiaomali.XMLConfig.LoadXiaoMaLiCfg()
	bibei.BBConfig.LoadBiBeiCfg()

	room := room.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
