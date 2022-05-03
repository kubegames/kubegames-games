package glogic

import (
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type ErBaGangTable struct {
}

// InitTable 初始化桌子
func (EBGroom *ErBaGangTable) InitTable(table table.TableInterface) {
	g := new(ErBaGangGame)
	g.InitTable(table)
	table.Start(g, nil, nil)
}

// UserExit 用户退出桌子
func (EBGroom *ErBaGangTable) UserExit(user player.PlayerInterface) {
	log.Tracef("用户退出桌子")
}
