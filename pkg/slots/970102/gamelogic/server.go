package gamelogic

import (
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Server struct {
}

//初始化桌子
func (server *Server) InitTable(table table.TableInterface) {
	game := NewGame(table)
	game.init()
	table.Start(game, nil, nil)
}
