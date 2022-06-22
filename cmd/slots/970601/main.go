package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/slots/970601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/game"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	room "github.com/kubegames/kubegames-sdk/pkg/room/slots"
	"github.com/sipt/GoJsoner"
)

func main() {
	//可修改部分---开始
	initConfig()

	//启动服务
	rand.Seed(time.Now().UnixNano())
	room := room.NewRoom(game.NewLhdbRoom())
	room.Run()
}

func initConfig() {
	//去除配置文件中的注释
	gameFrameData, err := ioutil.ReadFile("./conf/game_frame.json")
	if err != nil {
		log.Traceln("gameFrameData reading error", err)
		panic("")
	}
	gameFrameResult, _ := GoJsoner.Discard(string(gameFrameData))
	err = json.Unmarshal([]byte(gameFrameResult), &config.GameFrameConfig)
	if err != nil {
		log.Errorf("Load game_config.go file err:%s ", err.Error())
		panic("")
	}
	config.LoadLhdbConfig()

}
