package game

import (
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type LhdbRoom struct {
	Id          int64
	Name        string
	LimitAmount int64 //准入金额，低于该值就不能进入
	MaxRed      int64 //最大红包
	Lock        sync.Mutex
}

func NewLhdbRoom() (redRoom *LhdbRoom) {
	redRoom = &LhdbRoom{}
	return
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
}

//初始化桌子
func (redRoom *LhdbRoom) InitTable(table table.TableInterface) {
	g := NewGame(1, redRoom)
	g.Init(table)
	table.Start(g, nil, nil)
}
