package game

import (
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type WaterRoom struct {
	Id          int64
	Name        string
	LimitAmount int64 //准入金额，低于该值就不能进入
	MaxRed      int64 //最大红包
	Lock        sync.Mutex
}

func NewWaterRoom() (redRoom *WaterRoom) {
	redRoom = &WaterRoom{}
	return
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
}

//初始化桌子
func (redRoom *WaterRoom) InitTable(table table.TableInterface) {
	g := NewGame(int64(table.GetID()), redRoom)
	g.Init(table)
	table.Start(g, nil, nil)
}
