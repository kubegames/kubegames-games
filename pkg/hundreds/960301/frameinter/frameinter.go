package frameinter

import (
	"github.com/kubegames/kubegames-games/pkg/hundreds/960301/game"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type RBWarRoom struct {
}

//初始化桌子
func (rbWarRoom *RBWarRoom) InitTable(table table.TableInterface) {
	g := new(game.Game)
	g.Init(table)
	//桌子启动
	table.Start(g, func() {
		//百人游戏，开启桌子就启动游戏
		g.GameStart()
	}, nil)
}
