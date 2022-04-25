//拉霸

package main

import (
	"fmt"
	"game_LaBa/labacom/config"
	"game_LaBa/yhhwd/gamelogic"

	"game_LaBa/yhhwd/config"
	"game_LaBa/yhhwd/test"
	"game_frame_v2/game/logic"
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
	//	maintest()
	room := game_frame.NewRoom(&gamelogic.LaBaRoom{})
	room.Run()
}

//增加功能1，出现特殊玩法时，随机除龙母格子以外随机出3个wild图标SCATER图标不替换
