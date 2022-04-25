package frameinter

import (
	"game_poker/BRTB/game"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type BRTBRoom struct {
}

//初始化桌子
func (lhRoom *BRTBRoom) InitTable(table table.TableInterface) {
	g := new(game.Game)
	g.Init(table)
	table.BindGame(g)
}

func (lhRoom *BRTBRoom) UserExit(user player.PlayerInterface) {
}
