package frameinter

import (
	"go-game-sdk/example/game_poker/960303/game"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type LongHuRoom struct {
}

//初始化桌子
func (lhRoom *LongHuRoom) InitTable(table table.TableInterface) {
	g := new(game.Game)
	g.Init(table)
	table.BindGame(g)
}

func (lhRoom *LongHuRoom) UserExit(user player.PlayerInterface) {
}
