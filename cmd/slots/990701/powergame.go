//拉霸

package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/labacom/config"
	roomconfig "github.com/kubegames/kubegames-games/pkg/slots/990701/config"
	"github.com/kubegames/kubegames-games/pkg/slots/990701/gamelogic"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LBConfig.LoadLabadCfg()
	roomconfig.CSDConfig.LoadRoomCfg()

	room := room.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
