package frameinter

import (
	"go-game-sdk/example/game_poker/saima/game"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Server struct {
}

//初始化桌子
func (server *Server) InitTable(table table.TableInterface) {
	g := new(game.Game)
	g.Init(table)
	table.BindGame(g)
}

func (server *Server) UserExit(user player.PlayerInterface) {
}
