package main

import (
	"fmt"
	"game_frame_v2/game/logic"
	"game_poker/BRZJH/config"
	"game_poker/BRZJH/frameinter"
	"game_poker/BRZJH/game"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("### VER:  1.0.2  ")
	rand.Seed(time.Now().UnixNano())
	config.LoadBRZJHConfig()
	game.RConfig.LoadLabadCfg()
	room := game_frame.NewRoom(&frameinter.BaiRenZhaJinHuaRoom{})
	room.Run()
}
