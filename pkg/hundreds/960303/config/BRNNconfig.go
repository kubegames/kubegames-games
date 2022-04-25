package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/tidwall/gjson"
)

type InitBRNNConfig struct {
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
	Robotgold                         [][]int64
	Robotgold10times                  [][]int64 //10倍场 机器人出局
	Betmin                            int       `json:"betmin"`
	Unplacebetnum                     int       `json:"unplacebetnum"`
	Chips5times                       [][]int64 //5倍场 筹码
	Shangzhuanglimit5times            []int64   //5倍场 上庄限制
	Singleusersinglespacelimit5times  [][]int64 //5倍场 个人玩家单区域限红
	Singleuserallspacelimit5times     []int64   //5倍场 个人玩家所有区域限红
	Allusersinglespacelimit5times     [][]int64 //5倍场 所有玩家区单区域限红
	Allspacelimit5times               []int64   //5倍场 所有区域总限红
	Chips10times                      [][]int64 //10倍场 筹码
	Shangzhuanglimit10times           []int64   //10倍场 上庄限制
	Singleusersinglespacelimit10times [][]int64 //10倍场 个人玩家单区域限红
	Singleuserallspacelimit10times    []int64   //10倍场 个人玩家所有区域限红
	Allusersinglespacelimit10times    [][]int64 //10倍场 所有玩家区单区域限红
	Allspacelimit10times              []int64   //10倍场 所有区域总限红
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
	OddsInfo                   int32
	SingleUserSingleSpaceLimit [4]int64 // 个人玩家单区域限红
	SingleUserAllSpaceLimit    int64    //个人玩家所有区域限红
	AllUserSingleSpaceLimit    [4]int64 //所有玩家区单区域限红
	AllSpaceLimit              int64    //所有区域总限红
	BetMinLimit                int64    //最低携带金额限制

}

var BRNNConfig InitBRNNConfig

func LoadBRNNConfig() {
	data, err := ioutil.ReadFile("./conf/BRNNconfig.json")
	if err != nil {
		log.Println("LoadBRNNConfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	BRNNConfig.Num1000 = int(cfg.Get("1000").Int())
	BRNNConfig.Num2000 = int(cfg.Get("2000").Int())
	BRNNConfig.Num3000 = int(cfg.Get("3000").Int())
	BRNNConfig.Num_1000 = int(cfg.Get("-1000").Int())
	BRNNConfig.Num_2000 = int(cfg.Get("-2000").Int())
	BRNNConfig.Num_3000 = int(cfg.Get("-3000").Int())
	BRNNConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	BRNNConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	BRNNConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	BRNNConfig.Taketimes.ShowPoker = int(cfg.Get("taketimes.showpoker").Int())
	BRNNConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())
	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}

		BRNNConfig.Robotgold = append(BRNNConfig.Robotgold, gold)
	}
	for i := 1; i <= 4; i++ {
		//5倍场
		//筹码
		BRNNConfig.Chips5times = append(BRNNConfig.Chips5times, loadcfg(cfg, i, "chips5times"))
		//上庄限制
		Shangzhuanglimit5times := loadcfg(cfg, i, "shangzhuanglimit5times")
		BRNNConfig.Shangzhuanglimit5times = append(BRNNConfig.Shangzhuanglimit5times, Shangzhuanglimit5times[0])
		//用户单区域限制
		BRNNConfig.Singleusersinglespacelimit5times = append(BRNNConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		//用户所有区域限制
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BRNNConfig.Singleuserallspacelimit5times = append(BRNNConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		//所有用户单区域限制
		BRNNConfig.Allusersinglespacelimit5times = append(BRNNConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		//所有用户限制
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BRNNConfig.Allspacelimit5times = append(BRNNConfig.Allspacelimit5times, Allspacelimit5times[0])
		//机器人退场限制
		BRNNConfig.Robotgold10times = append(BRNNConfig.Robotgold10times, loadcfg(cfg, i, "robotgold10times"))
		//10倍场
		BRNNConfig.Chips10times = append(BRNNConfig.Chips10times, loadcfg(cfg, i, "chips10times"))
		Shangzhuanglimit10times := loadcfg(cfg, i, "shangzhuanglimit10times")
		BRNNConfig.Shangzhuanglimit10times = append(BRNNConfig.Shangzhuanglimit10times, Shangzhuanglimit10times[0])
		BRNNConfig.Singleusersinglespacelimit10times = append(BRNNConfig.Singleusersinglespacelimit10times, loadcfg(cfg, i, "singleusersinglespacelimit10times"))
		Singleuserallspacelimit10times := loadcfg(cfg, i, "singleuserallspacelimit10times")
		BRNNConfig.Singleuserallspacelimit10times = append(BRNNConfig.Singleuserallspacelimit10times, Singleuserallspacelimit10times[0])
		BRNNConfig.Allusersinglespacelimit10times = append(BRNNConfig.Allusersinglespacelimit10times, loadcfg(cfg, i, "allusersinglespacelimit10times"))
		Allspacelimit10times := loadcfg(cfg, i, "allspacelimit10times")
		BRNNConfig.Allspacelimit10times = append(BRNNConfig.Allspacelimit10times, Allspacelimit10times[0])
	}

	BRNNConfig.Betmin = int(cfg.Get("betmin").Int())

	BRNNConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
}

func (cfg *InitBRNNConfig) GetCheatValue(Cheat int) int {
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
