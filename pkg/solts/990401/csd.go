//拉霸

package main

import (
	"fmt"
	game_frame "go-game-sdk"
	roomconfig "go-game-sdk/example/game_LaBa/990401/config"
	"go-game-sdk/example/game_LaBa/990401/gamelogic"
	"go-game-sdk/example/game_LaBa/990401/test"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"math/rand"
	"time"
)

func maintest() {
	fmt.Println("开始", time.Now())
	var g gamelogic.Game
	g.Init(&config.LBConfig)
	//这里调用测试工具，如果是正式版本需要屏蔽这个结果
	test.Test(&config.LBConfig)
	fmt.Println("结束", time.Now())
}

func main() {
	rand.Seed(time.Now().UnixNano())
	config.LBConfig.LoadLabadCfg()
	roomconfig.CSDConfig.LoadRoomCfg()
	maintest()

	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}
