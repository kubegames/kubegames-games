package frameinter

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960211/game"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Pai9Room struct {
}

//初始化桌子
func (lhRoom *Pai9Room) InitTable(table table.TableInterface) {
	game := game.NewGame(table)
	table.Start(game, nil, nil)
}
