package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type InitBRZJHConfig struct {
	Num1000   int `json:"1000"`
	Num2000   int `json:"2000"`
	Num3000   int `json:"3000"`
	Num_1000  int `json:"-1000"`
	Num_2000  int `json:"-2000"`
	Num_3000  int `json:"-3000"`
	Taketimes struct {
		Startmove       int `json:"startmove"`
		Startbet        int `json:"startbet"`
		Endmove         int `json:"endmove"`
		ShowPoker       int `json:"showpoker"`
		Endpay          int `json:"endpay"`
		BetGapBroadcast int `json:"betGapBroadcast"` // 下注期间广播消息
	} `json:"taketimes"`
	Robotgold                        [][]int64
	Betmin                           int       `json:"betmin"`
	Unplacebetnum                    int       `json:"unplacebetnum"`
	Chips5times                      [][]int64 //5倍场 筹码
	Shangzhuanglimit5times           []int64   //5倍场 上庄限制
	Singleusersinglespacelimit5times [][]int64 //5倍场 个人玩家单区域限红
	Singleuserallspacelimit5times    []int64   //5倍场 个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64 //5倍场 所有玩家区单区域限红
	Allspacelimit5times              []int64   //5倍场 所有区域总限红
	ShangZhuangPersonNumMax          int       //上庄人数限制
	ShangZhuangMax                   int       //上庄最大局数限制
}

type RoomRules struct {
	BetList                    []int64  //下注列表
	BetLimit                   [4]int64 //下注限制
	UserBetLimit               int64    //个人下注限制
	SitDownLimit               int      //坐下限制
	RobotMaxGold               int64    //机器人带的最多的金币，超过这个数据将被踢掉
	RobotMinGold               int64    //机器人带的最少的金币
	ZhuangLimit                int64    //上庄的最低金币
	Zhuang                     int64    //上庄的金币
	OddsInfo                   int32    //几倍场
	SingleUserSingleSpaceLimit [4]int64 // 个人玩家单区域限红
	SingleUserAllSpaceLimit    int64    //个人玩家所有区域限红
	AllUserSingleSpaceLimit    [4]int64 //所有玩家区单区域限红
	AllSpaceLimit              int64    //所有区域总限红
	BetMinLimit                int64    //最低携带金额限制
}

var BRZJHConfig InitBRZJHConfig

func LoadBRZJHConfig() {
	data, err := ioutil.ReadFile("./conf/BRZJHconfig.json")
	if err != nil {
		log.Println("LoadBRNNConfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	BRZJHConfig.Num1000 = int(cfg.Get("1000").Int())
	BRZJHConfig.Num2000 = int(cfg.Get("2000").Int())
	BRZJHConfig.Num3000 = int(cfg.Get("3000").Int())
	BRZJHConfig.Num_1000 = int(cfg.Get("-1000").Int())
	BRZJHConfig.Num_2000 = int(cfg.Get("-2000").Int())
	BRZJHConfig.Num_3000 = int(cfg.Get("-3000").Int())
	BRZJHConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	BRZJHConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	BRZJHConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	BRZJHConfig.Taketimes.ShowPoker = int(cfg.Get("taketimes.showpoker").Int())
	BRZJHConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())
	BRZJHConfig.Taketimes.BetGapBroadcast = int(cfg.Get("taketimes.betGapBroadcast").Int())
	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}

		BRZJHConfig.Robotgold = append(BRZJHConfig.Robotgold, gold)
	}
	for i := 1; i <= 4; i++ {
		BRZJHConfig.Chips5times = append(BRZJHConfig.Chips5times, loadcfg(cfg, i, "chips5times"))
		Shangzhuanglimit5times := loadcfg(cfg, i, "shangzhuanglimit5times")
		BRZJHConfig.Shangzhuanglimit5times = append(BRZJHConfig.Shangzhuanglimit5times, Shangzhuanglimit5times[0])
		BRZJHConfig.Singleusersinglespacelimit5times = append(BRZJHConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BRZJHConfig.Singleuserallspacelimit5times = append(BRZJHConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		BRZJHConfig.Allusersinglespacelimit5times = append(BRZJHConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BRZJHConfig.Allspacelimit5times = append(BRZJHConfig.Allspacelimit5times, Allspacelimit5times[0])
	}
	BRZJHConfig.Betmin = int(cfg.Get("betmin").Int())
	BRZJHConfig.ShangZhuangPersonNumMax = int(cfg.Get("shangzhuangpersonnummax").Int())
	BRZJHConfig.ShangZhuangMax = int(cfg.Get("shangzhuangmax").Int())
	BRZJHConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
}

func (cfg *InitBRZJHConfig) GetCheatValue(Cheat int) int {
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
