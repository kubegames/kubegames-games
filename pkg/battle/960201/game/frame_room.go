package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type ZjhRoom struct {
}

//初始化桌子
func (rbWarRoom *ZjhRoom) InitTable(table table.TableInterface) {
	g := NewGame(int(table.GetID()), config.GameConfigArr[0])
	g.Init(table)
	table.Start(g, nil, nil)
}

func (rbWarRoom *ZjhRoom) UserExit(user player.PlayerInterface) {
}
