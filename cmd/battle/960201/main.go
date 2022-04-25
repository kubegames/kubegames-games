package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	_ "net/http/pprof"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/game"
	room "github.com/kubegames/kubegames-sdk/pkg/room/poker"
	"github.com/sipt/GoJsoner"
)

func main() {
	fmt.Println("version ### 2.0.29 ")
	//初始化配置
	initConfig()
	//启动服务
	rand.Seed(time.Now().UnixNano())
	room := room.NewRoom(&game.ZjhRoom{})
	room.Run()
}

func initConfig() {
	err := config.LoadJsonConfig("./conf/config.json", config.Config)
	if err != nil {
		panic(err)
	}
	//ai配置文件
	aiConfigData, err := ioutil.ReadFile("./conf/ai_config.json")
	if err != nil {
		panic(err)
	}
	//去除配置文件中的注释
	aiConfigResult, _ := GoJsoner.Discard(string(aiConfigData))
	err = json.Unmarshal([]byte(aiConfigResult), &config.AiConfigArr)
	if err != nil {
		panic(err)
	}
	fmt.Println("ai config ： ", len(config.AiConfigArr))

	//游戏配置文件
	gameConfigData, err := ioutil.ReadFile("./conf/game.json")
	if err != nil {
		panic(err)
	}
	//去除配置文件中的注释
	gameConfigResult, _ := GoJsoner.Discard(string(gameConfigData))
	err = json.Unmarshal([]byte(gameConfigResult), &config.GameConfigArr)
	if err != nil {
		panic(err)
	}
	fmt.Println("game config ： ", config.GameConfigArr[0].MinAction, config.GameConfigArr[1].MinAction, config.GameConfigArr[2].MinAction, config.GameConfigArr[3].MinAction)

	//作弊率配置文件
	cheatConfigData, err := ioutil.ReadFile("./conf/cheat.json")
	if err != nil {
		panic(err)
	}
	cheatConfigResult, _ := GoJsoner.Discard(string(cheatConfigData))
	err = json.Unmarshal([]byte(cheatConfigResult), &config.CheatConfigArr)
	if err != nil {
		panic(err)
	}
	if len(config.CheatConfigArr) != 4 {
		panic("作弊率的配置文件必须为2个，最大和第二大")
	}
	fmt.Println("作弊率：", config.CheatConfigArr[0].MustLoseRate)

	//作弊率配置文件
	changeConfigData, err := ioutil.ReadFile("./conf/change_cards.json")
	if err != nil {
		panic(err)
	}
	changeConfigResult, _ := GoJsoner.Discard(string(changeConfigData))
	err = json.Unmarshal([]byte(changeConfigResult), &config.ChangeCardsArr)
	if err != nil {
		panic(err)
	}
	fmt.Println("change config json : ", config.ChangeCardsArr)
}
