package roomconfig

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

type RoomConfig struct {
	CheatPro     map[int]int
	WildCheatPro map[int]int
}

var CSDConfig RoomConfig

//读取配置文件
func (cfg *RoomConfig) LoadRoomCfg() {
	data, err := ioutil.ReadFile("conf/room.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	cfg.analysiscfg(result)
}

//解析配置文件
func (cfg *RoomConfig) analysiscfg(json_str string) {
	//使用简单json来解析。
	res, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}

	cfg.CheatPro = make(map[int]int)
	cfg.analysischeatvalue(res, 3000)
	cfg.analysischeatvalue(res, 2000)
	cfg.analysischeatvalue(res, 1000)
	cfg.analysischeatvalue(res, -3000)
	cfg.analysischeatvalue(res, -2000)
	cfg.analysischeatvalue(res, -1000)
	cfg.WildCheatPro = make(map[int]int)
	for i := 0; i <= 6; i++ {
		cfg.analysiswildcheatvalue(res, i)
	}
	for i := 11; i <= 17; i++ {
		cfg.analysiswildcheatvalue(res, i)
	}
}

func (cfg *RoomConfig) analysischeatvalue(js *simplejson.Json, cheatvalue int) {
	cheatvaluestr := strconv.FormatInt(int64(cheatvalue), 10)
	pro, _ := js.Get(cheatvaluestr).Int()
	cfg.CheatPro[cheatvalue] = pro
}
func (cfg *RoomConfig) analysiswildcheatvalue(js *simplejson.Json, cheatvalue int) {
	cheatvaluestr := strconv.FormatInt(int64(cheatvalue), 10)
	pro, _ := js.Get(cheatvaluestr).Int()
	cfg.WildCheatPro[cheatvalue] = pro
}
func (cfg *RoomConfig) GetProByCheat(cheatvalue int) int {
	pro, ok := cfg.CheatPro[cheatvalue]

	if !ok {
		pro, _ = cfg.CheatPro[1000]
	}

	return pro
}
