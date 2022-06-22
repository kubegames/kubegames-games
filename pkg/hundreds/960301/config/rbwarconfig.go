package config

import (
	"fmt"
	"io/ioutil"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/tidwall/gjson"
)

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

type InitRBConfig struct {
	Num1000  int `json:"1000"`
	Num2000  int `json:"2000"`
	Num3000  int `json:"3000"`
	Num_1000 int `json:"-1000"`
	Num_2000 int `json:"-2000"`
	Num_3000 int `json:"-3000"`
	Ratio    struct {
		Duizi   int `json:"duizi"`
		Shunzi  int `json:"shunzi"`
		Jinhua  int `json:"jinhua"`
		Shunjin int `json:"shunjin"`
		Baozi   int `json:"baozi"`
	} `json:"ratio"`
	Tax       int `json:"tax"`
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

var RBWarConfig InitRBConfig

func LoadRBWarConfig() {
	data, err := ioutil.ReadFile("./conf/rbwarconfig.json")
	if err != nil {
		log.Errorf("LoadWBWarConfig Error %s", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	RBWarConfig.Num1000 = int(cfg.Get("1000").Int())
	RBWarConfig.Num2000 = int(cfg.Get("2000").Int())
	RBWarConfig.Num3000 = int(cfg.Get("3000").Int())
	RBWarConfig.Num_1000 = int(cfg.Get("-1000").Int())
	RBWarConfig.Num_2000 = int(cfg.Get("-2000").Int())
	RBWarConfig.Num_3000 = int(cfg.Get("-3000").Int())
	RBWarConfig.Ratio.Duizi = int(cfg.Get("ratio.duizi").Int())
	RBWarConfig.Ratio.Shunzi = int(cfg.Get("ratio.shunzi").Int())
	RBWarConfig.Ratio.Jinhua = int(cfg.Get("ratio.jinhua").Int())
	RBWarConfig.Ratio.Shunjin = int(cfg.Get("ratio.shunjin").Int())
	RBWarConfig.Ratio.Baozi = int(cfg.Get("ratio.baozi").Int())

	Betamount := []int{}
	for _, data := range cfg.Get("betamount").Array() {
		Betamount = append(Betamount, int(data.Int()))
	}

	RBWarConfig.Tax = int(cfg.Get("tax").Int())

	RBWarConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	RBWarConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	RBWarConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	RBWarConfig.Taketimes.ShowPoker = int(cfg.Get("taketimes.showpoker").Int())
	RBWarConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())

	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}

		RBWarConfig.Robotgold = append(RBWarConfig.Robotgold, gold)
	}
	for i := 1; i <= 4; i++ {
		//10倍场
		RBWarConfig.Chips5times = append(RBWarConfig.Chips5times, loadcfg(cfg, i, "chips5times"))
		RBWarConfig.Singleusersinglespacelimit5times = append(RBWarConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		RBWarConfig.Singleuserallspacelimit5times = append(RBWarConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		RBWarConfig.Allusersinglespacelimit5times = append(RBWarConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		RBWarConfig.Allspacelimit5times = append(RBWarConfig.Allspacelimit5times, Allspacelimit5times[0])
	}
	RBWarConfig.Betmin = int(cfg.Get("betmin").Int())
	RBWarConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
}

func (cfg *InitRBConfig) GetCheatValue(Cheat int) int {
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
