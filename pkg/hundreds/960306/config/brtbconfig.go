package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/tidwall/gjson"
)

type InitBRTBConfig struct {
	Num1000   int `json:"1000"`
	Num2000   int `json:"2000"`
	Num3000   int `json:"3000"`
	Num_1000  int `json:"-1000"`
	Num_2000  int `json:"-2000"`
	Num_3000  int `json:"-3000"`
	Taketimes struct {
		StartGame int `json:"startgame"`
		ShakeDice int `json:"shakedice"`
		Startmove int `json:"startmove"`
		Startbet  int `json:"startbet"`
		Endmove   int `json:"endmove"`
		Cardmove  int `json:"cardmove"`
		Endpay    int `json:"endpay"`
	} `json:"taketimes"`
	Robotgold                        [][]int64
	Betmin                           int          `json:"betmin"`
	Unplacebetnum                    int          `json:"unplacebetnum"`
	Chips5times                      [][]int64    // 筹码
	Singleusersinglespacelimit5times [][]int64    //个人玩家单区域限红
	Singleuserallspacelimit5times    []int64      //个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64    //所有玩家区单区域限红
	Allspacelimit5times              []int64      //所有区域总限红
	PolicyTree                       ShakePolicys `json:"policyTree"`
}
type ShakePolicy struct {
	RoomProb int `json:"roomProb"` // 血池值

	Back Backs `json:"back"` // 返奖率
}
type back struct {
	Min  int `json:"min"`
	Max  int `json:"max"`
	Prob int `json:"prob"`
}
type Backs []back

func (bs Backs) Rand() back {
	var allweight int
	for _, v := range bs {
		allweight += v.Prob
	}
	randweight := rand.Intn(allweight) + 1
	for _, v := range bs {
		if v.Prob == 0 {
			continue
		}
		if randweight <= v.Prob {
			return v
		}
		randweight -= v.Prob
	}
	return bs[rand.Intn(len(bs))]
}

type ShakePolicys []*ShakePolicy

func (sps ShakePolicys) Find(roomProb int32) *ShakePolicy {
	for _, v := range sps {
		if v.RoomProb == int(roomProb) {
			return v
		}
	}
	return sps[rand.Intn(len(sps))]
}

type RoomRules struct {
	BetList                    []int64   //下注列表
	BetLimit                   [5]int64  //下注限制
	UserBetLimit               int64     //个人下注限制
	SitDownLimit               int       //坐下限制
	RobotMaxGold               int64     //机器人带的最多的金币，超过这个数据将被踢掉
	RobotMinGold               int64     //机器人带的最少的金币
	SingleUserSingleSpaceLimit [25]int64 // 个人玩家单区域限红
	SingleUserAllSpaceLimit    int64     //个人玩家所有区域限红
	AllUserSingleSpaceLimit    [25]int64 //所有玩家区单区域限红
	AllSpaceLimit              int64     //所有区域总限红
	BetMinLimit                int64     //最低携带金额限制
	RobotLine                  [4]int    //机器人下注平衡线
}

var BRTBConfig InitBRTBConfig

func LoadBaiJiaLeConfig() {
	data, err := ioutil.ReadFile("./conf/brtbconfig.json")
	if err != nil {
		log.Tracef("brtbconfig Error %v", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	BRTBConfig.Num1000 = int(cfg.Get("1000").Int())
	BRTBConfig.Num2000 = int(cfg.Get("2000").Int())
	BRTBConfig.Num3000 = int(cfg.Get("3000").Int())
	BRTBConfig.Num_1000 = int(cfg.Get("-1000").Int())
	BRTBConfig.Num_2000 = int(cfg.Get("-2000").Int())
	BRTBConfig.Num_3000 = int(cfg.Get("-3000").Int())
	BRTBConfig.Taketimes.ShakeDice = int(cfg.Get("taketimes.shakedice").Int())
	BRTBConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	BRTBConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	BRTBConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	BRTBConfig.Taketimes.Cardmove = int(cfg.Get("taketimes.cardmove").Int())
	BRTBConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())
	BRTBConfig.Taketimes.StartGame = int(cfg.Get("taketimes.startgame").Int())
	//获取返奖率

	for _, v := range cfg.Get("policyTree").Array() {
		sp := new(ShakePolicy)
		if err := json.Unmarshal([]byte(v.String()), sp); err != nil {
			panic(err)
		}
		BRTBConfig.PolicyTree = append(BRTBConfig.PolicyTree, sp)
	}

	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}

		BRTBConfig.Robotgold = append(BRTBConfig.Robotgold, gold)
	}
	for i := 1; i <= 4; i++ {
		BRTBConfig.Chips5times = append(BRTBConfig.Chips5times, loadcfg(cfg, i, "chips5times"))
		BRTBConfig.Singleusersinglespacelimit5times = append(BRTBConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BRTBConfig.Singleuserallspacelimit5times = append(BRTBConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		BRTBConfig.Allusersinglespacelimit5times = append(BRTBConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BRTBConfig.Allspacelimit5times = append(BRTBConfig.Allspacelimit5times, Allspacelimit5times[0])
	}
	BRTBConfig.Betmin = int(cfg.Get("betmin").Int())
	BRTBConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
	//log.Traceln(BRTBConfig.PolicyTree.Find(1000).Back.Rand())
	//log.Traceln(BRTBConfig.PolicyTree.Find(2000))
	//log.Traceln(BRTBConfig.PolicyTree.Find(3000))
	//log.Traceln(BRTBConfig.PolicyTree.Find(-1000))

}

func (cfg *InitBRTBConfig) GetCheatValue(Cheat int) int {
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

//func checkConfig() {
//	for _, policy := range BRTBConfig.PolicyTree {
//		if policy.All.AllPay+policy.All.AllKill+policy.All.None != 10000 {
//			panic("policy.all.allPay+policy.all.allKill+policy.all.none != 10000")
//		}
//		if policy.Free.GoldShark+policy.Free.SilverShark != 10000 {
//			panic("policy.all.allPay+policy.all.allKill+policy.all.none != 10000")
//		}
//	}
//}

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
