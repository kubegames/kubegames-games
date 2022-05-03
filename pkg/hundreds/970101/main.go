package main

import (
	recover_handle "common/recover"
	"encoding/json"
	"fmt"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/game"
	game_frame "game_frame_v2/game/logic"
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
	defer recover_handle.RecoverHandle("main recover ")
	log.Traceln("VER:  1.0.21  ")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)
	//可修改部分---开始
	initConfig()
	go func() {
		//启动服务
		rand.Seed(time.Now().UnixNano())
		room := game_frame.NewRoom(game.NewRedRoom())
		room.Run()
	}()

	//开启pprof
	go func() {
		//log.Traceln("pprof start at :9877")
		log.Traceln(http.ListenAndServe(":9877", nil))
	}()

	sig := <-sigs
	log.Warnf("###SIGNAL::%v,PID:%d", sig, os.Getpid())
}

func initConfig() {
	//ai配置文件
	aiConfigData, err := ioutil.ReadFile("./config/ai_send_config.json")
	if err != nil {
		panic("ai_send_config reading error" + err.Error())
	}
	//去除配置文件中的注释
	aiConfigResult, _ := GoJsoner.Discard(string(aiConfigData))
	err = json.Unmarshal([]byte(aiConfigResult), &config.AiSendConfig)
	if err != nil {
		log.Errorf("Load ai_send_config.go file err:%s ", err.Error())
		panic("")
	}
	aiRobConfigData, err := ioutil.ReadFile("./config/ai_rob_config.json")
	if err != nil {
		panic("aiRobConfigData reading error" + err.Error())
	}
	//去除配置文件中的注释
	aiRobConfigResult, _ := GoJsoner.Discard(string(aiRobConfigData))
	err = json.Unmarshal([]byte(aiRobConfigResult), &config.AiRobConfigArr)
	if err != nil {
		log.Errorf("Load aiRobConfigResult.go file err:%s ", err.Error())
		panic("")
	}
	//log.Traceln(" ai rob config.json data : ", fmt.Sprintf(`%v`, config.json.AiRobConfigArr[1]))
	for _, arr := range config.AiRobConfigArr {
		config.AiRobConfigMap[arr.Cheat] = arr
	}

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

	//游戏配置文件
	redConfigData, err := ioutil.ReadFile("./config/red.json")
	if err != nil {
		log.Traceln("redConfigData reading error", err)
		panic("")
	}
	//去除配置文件中的注释
	redConfigResult, _ := GoJsoner.Discard(string(redConfigData))
	err = json.Unmarshal([]byte(redConfigResult), &config.RedConfig)
	if err != nil {
		log.Errorf("Load redConfigResult.go file err:%s ", err.Error())
		panic("")
	}
	log.Traceln("red config.json ： ", fmt.Sprintf(`%+v`, config.RedConfig))

	//新增配置文件
	sendNewConfigData, err := ioutil.ReadFile("./config/send_new.json")
	if err != nil {
		log.Traceln("redConfigData reading error", err)
		panic("")
	}
	//去除配置文件中的注释
	sendNewConfigResult, _ := GoJsoner.Discard(string(sendNewConfigData))
	err = json.Unmarshal([]byte(sendNewConfigResult), &config.AiSendNewArr)
	if err != nil {
		log.Errorf("Load send_new.json file err:%s ", err.Error())
		panic("")
	}
	log.Traceln("red config.json ： ", fmt.Sprintf(`%+v`, config.AiSendNewArr[0]))

	robNewConfigData, err := ioutil.ReadFile("./config/rob_new.json")
	if err != nil {
		log.Traceln("redConfigData rob_new error", err)
		panic("")
	}
	//去除配置文件中的注释
	robNewConfigResult, _ := GoJsoner.Discard(string(robNewConfigData))
	err = json.Unmarshal([]byte(robNewConfigResult), &config.AiRobNewArr)
	if err != nil {
		log.Errorf("Load send_new.json file err:%s ", err.Error())
		panic("")
	}
	log.Traceln("red rob_new.json ： ", fmt.Sprintf(`%+v`, config.AiRobNewArr[0]))
	config.LoadCrazyRedConfig()
}
