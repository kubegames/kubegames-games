package frameinter

import (
	"game_poker/pai9/game"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Pai9Room struct {
}

//初始化桌子
func (lhRoom *Pai9Room) InitTable(table table.TableInterface) {
	game := game.NewGame(table)
	table.Start(game, nil, nil)
}

func (lhRoom *Pai9Room) UserExit(user player.PlayerInterface) {
}
