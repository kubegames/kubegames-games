package main

import (
	"encoding/json"
	"fmt"
	game_frame "go-game-sdk"
	recover_handle "go-game-sdk/lib/recover"
	"io/ioutil"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/game"

	"github.com/sipt/GoJsoner"
)

func main() {

	defer recover_handle.Recover("main recover ")
	fmt.Println("### VER:  1.0.18 ")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	//可修改部分---开始
	initConfig()
	go func() {
		//启动服务
		rand.Seed(time.Now().UnixNano())
		room := game_frame.NewRoom(game.NewWaterRoom())
		room.Run()
	}()

	//开启pprof
	go func() {
		fmt.Println("pprof start at :9766")
		fmt.Println(http.ListenAndServe(":9766", nil))
	}()

	sig := <-sigs
	log.Warnf("###SIGNAL::%v,PID:%d", sig, os.Getpid())

}

func initConfig() {

	//去除配置文件中的注释
	gameFrameData, err := ioutil.ReadFile("./config/game_frame.json")
	if err != nil {
		fmt.Println("gameFrameData reading error", err)
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
		fmt.Println("gameConfigData reading error", err)
		panic("")
		return
	}
	cheatConfigResult, _ := GoJsoner.Discard(string(cheatConfigData))
	err = json.Unmarshal([]byte(cheatConfigResult), &config.CheatConf)
	if err != nil {
		log.Errorf("Load cheat_config.go file err:%s ", err.Error())
		panic("")
		return
	}

	//作弊率配置文件
	robotConfigData, err := ioutil.ReadFile("./config/robot.json")
	if err != nil {
		fmt.Println("gameConfigData reading error", err)
		panic("")
		return
	}
	robotConfigResult, _ := GoJsoner.Discard(string(robotConfigData))
	err = json.Unmarshal([]byte(robotConfigResult), &config.RobotConf)
	if err != nil {
		log.Errorf("Load robot_config.go file err:%s ", err.Error())
		panic("")
		return
	}
}
