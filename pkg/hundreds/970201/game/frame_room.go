package game

import (
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type RedRoom struct {
	Id          int64
	Name        string
	LimitAmount int64 //准入金额，低于该值就不能进入
	MaxRed      int64 //最大红包
	Lock        sync.Mutex
}

func NewRedRoom() (redRoom *RedRoom) {
	redRoom = &RedRoom{}
	return
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
}

//初始化桌子
func (redRoom *RedRoom) InitTable(table table.TableInterface) {
	g := NewGame(int64(table.GetID()), redRoom)
	g.Init(table)
	table.BindGame(g)
}

func (redRoom *RedRoom) UserExit(user player.PlayerInterface) {
}
