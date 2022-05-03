package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/game"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
	"github.com/sipt/GoJsoner"
)

func main() {
	log.Traceln("### VER:  1.0.18 ")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	//可修改部分---开始
	initConfig()

	//启动服务
	rand.Seed(time.Now().UnixNano())
	room := room.NewRoom(game.NewWaterRoom())
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

	//作弊率配置文件
	cheatConfigData, err := ioutil.ReadFile("./config/cheat.json")
	if err != nil {
		log.Traceln("gameConfigData reading error", err)
		panic(err)
	}
	cheatConfigResult, _ := GoJsoner.Discard(string(cheatConfigData))
	err = json.Unmarshal([]byte(cheatConfigResult), &config.CheatConf)
	if err != nil {
		log.Errorf("Load cheat_config.go file err:%s ", err.Error())
		panic(err)
	}

	//作弊率配置文件
	robotConfigData, err := ioutil.ReadFile("./config/robot.json")
	if err != nil {
		log.Traceln("gameConfigData reading error", err)
		panic(err)
	}
	robotConfigResult, _ := GoJsoner.Discard(string(robotConfigData))
	err = json.Unmarshal([]byte(robotConfigResult), &config.RobotConf)
	if err != nil {
		log.Errorf("Load robot_config.go file err:%s ", err.Error())
		panic(err)
	}
}
