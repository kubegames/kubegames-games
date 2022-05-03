package main

import (
	"encoding/json"
	game_frame "go-game-sdk"
	"go-game-sdk/example/game_LaBa/970601/config"
	"go-game-sdk/example/game_LaBa/970601/game"
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

	"github.com/sipt/GoJsoner"
)

func main() {
	defer recover_handle.Recover("main recover ")
	log.Traceln("### VER:  1.0.38 ")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	//可修改部分---开始
	initConfig()
	go func() {
		//启动服务
		rand.Seed(time.Now().UnixNano())
		room := game_frame.NewRoom(game.NewLhdbRoom())
		room.Run()
	}()

	//开启pprof
	go func() {
		log.Traceln("pprof start at :9766")
		log.Traceln(http.ListenAndServe(":9766", nil))
	}()

	sig := <-sigs
	log.Warnf("###SIGNAL::%v,PID:%d", sig, os.Getpid())

}

func initConfig() {

	//去除配置文件中的注释
	gameFrameData, err := ioutil.ReadFile("./config/game_frame.json")
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
