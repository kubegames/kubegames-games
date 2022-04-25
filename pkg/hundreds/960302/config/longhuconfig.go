package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type InitLHConfig struct {
	Num1000   int `json:"1000"`
	Num2000   int `json:"2000"`
	Num3000   int `json:"3000"`
	Num_1000  int `json:"-1000"`
	Num_2000  int `json:"-2000"`
	Num_3000  int `json:"-3000"`
	Taketimes struct {
		Startmove int `json:"startmove"`
		Startbet  int `json:"startbet"`
		Endmove   int `json:"endmove"`
		ShowPoker int `json:"showpoker"`
		Endpay    int `json:"endpay"`
	} `json:"taketimes"`
	Robotgold                        [][]int64
	Betmin                           int       `json:"betmin"`
	Unplacebetnum                    int       `json:"unplacebetnum"`
	Chips5times                      [][]int64 // 筹码
	Singleusersinglespacelimit5times [][]int64 //个人玩家单区域限红
	Singleuserallspacelimit5times    []int64   //个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64 //所有玩家区单区域限红
	Allspacelimit5times              []int64   //所有区域总限红
}

type RoomRules struct {
	BetList                    []int64  //下注列表
	BetLimit                   [3]int64 //下注限制
	UserBetLimit               int64    //个人下注限制
	SitDownLimit               int      //坐下限制
	RobotMaxGold               int64    //机器人带的最多的金币，超过这个数据将被踢掉
	RobotMinGold               int64    //机器人带的最少的金币
	SingleUserSingleSpaceLimit [3]int64 // 个人玩家单区域限红
	SingleUserAllSpaceLimit    int64    //个人玩家所有区域限红
	AllUserSingleSpaceLimit    [3]int64 //所有玩家区单区域限红
	AllSpaceLimit              int64    //所有区域总限红
	BetMinLimit                int64    //最低携带金额限制
}

var LongHuConfig InitLHConfig

func LoadLongHuConfig() {
	data, err := ioutil.ReadFile("./conf/longhuconfig.json")
	if err != nil {
		log.Println("longhuconfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	LongHuConfig.Num1000 = int(cfg.Get("1000").Int())
	LongHuConfig.Num2000 = int(cfg.Get("2000").Int())
	LongHuConfig.Num3000 = int(cfg.Get("3000").Int())
	LongHuConfig.Num_1000 = int(cfg.Get("-1000").Int())
	LongHuConfig.Num_2000 = int(cfg.Get("-2000").Int())
	LongHuConfig.Num_3000 = int(cfg.Get("-3000").Int())
	LongHuConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	LongHuConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	LongHuConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	LongHuConfig.Taketimes.ShowPoker = int(cfg.Get("taketimes.showpoker").Int())
	LongHuConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())

	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}

		LongHuConfig.Robotgold = append(LongHuConfig.Robotgold, gold)
	}
	for i := 1; i <= 4; i++ {
		//10倍场
		LongHuConfig.Chips5times = append(LongHuConfig.Chips5times, loadcfg(cfg, i, "chips5times"))
		LongHuConfig.Singleusersinglespacelimit5times = append(LongHuConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		LongHuConfig.Singleuserallspacelimit5times = append(LongHuConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		LongHuConfig.Allusersinglespacelimit5times = append(LongHuConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		LongHuConfig.Allspacelimit5times = append(LongHuConfig.Allspacelimit5times, Allspacelimit5times[0])
	}

	LongHuConfig.Betmin = int(cfg.Get("betmin").Int())

	LongHuConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
}

func (cfg *InitLHConfig) GetCheatValue(Cheat int) int {
	if Cheat == 1000 {
		return cfg.Num1000
	} else if Cheat == 2000 {
		return cfg.Num2000
	} else if Cheat == 3000 {
		return cfg.Num3000
	} else if Cheat == -1000 {
		return cfg.Num_1000
	} else if Cheat == -2000 {
		return cfg.Num_2000
	} else if Cheat == -3000 {
		return cfg.Num_3000
	}

	return 0
}

//加载[][]int64配置文件
func loadcfg(cfg gjson.Result, i int, jsonConfigField string) []int64 {
	str := fmt.Sprintf(jsonConfigField+".%v", i)
	temp := cfg.Get(str).Array()
	var tempArr []int64
	for _, v := range temp {
		tempArr = append(tempArr, v.Int())
	}
	return tempArr
}
