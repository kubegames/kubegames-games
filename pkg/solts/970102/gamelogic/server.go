package gamelogic

import (
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Server struct {
}

//初始化桌子
func (server *Server) InitTable(table table.TableInterface) {
	game := NewGame(table)
	game.init()
	table.BindGame(game)
}

func (server *Server) UserExit(user player.PlayerInterface) {
}
