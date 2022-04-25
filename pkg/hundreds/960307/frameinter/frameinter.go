package frameinter

import (
	"game_poker/BRZJH/game"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type BaiRenZhaJinHuaRoom struct {
}

//初始化桌子
func (lhRoom *BaiRenZhaJinHuaRoom) InitTable(table table.TableInterface) {
	g := new(game.Game)
	g.Init(table)
	table.BindGame(g)
}

func (lhRoom *BaiRenZhaJinHuaRoom) UserExit(user player.PlayerInterface) {
}
