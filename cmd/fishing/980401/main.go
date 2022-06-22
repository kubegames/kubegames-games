package main

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/fishing/980401/config"
	"github.com/kubegames/kubegames-games/pkg/fishing/980401/server"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
)

func main() {
	config.Load()
	rand.Seed(time.Now().UnixNano())
	//roomConfig := game_frame.RoomConfig{
	//	Port:      8080, //端口号
	//	Count:     5000, //最大连接数
	//	HeartBeat: 0,    //心跳 0（表示不设置心跳）
	//	TableNum:  500,  //房间桌子数量
	//	Max:       4,    //房间桌子的最大容纳人数
	//	Min:       1,    //房间桌子的最小开赛数
	//}
	room := room.NewRoom(&server.Server{})
	room.Run()
}
