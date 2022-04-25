package config

import (
	"encoding/json"
	"fmt"
	"go-game-sdk/example/game_LaBa/990601/model"
	bridanimal "go-game-sdk/example/game_LaBa/990601/msg"
	"io/ioutil"
	"log"
	"math/rand"

	"github.com/tidwall/gjson"
)

type RoomRule struct {
	BaseBet         int64    `json:"Bottom_Pouring"` // 基本低注
	ChipArea        [5]int64 `json:"chipArea"`       // 筹码区的筹码数
	BetArea         int64    `json:"betArea"`        // 每个下注区限制上线
	AllBetAreaLimit int64    // 所有下注区的上线
}

type ShakePolicy struct {
	RoomProb int `json:"roomProb"` // 血池值

	All struct {
		AllPay  int `json:"allPay"`  // 通赔
		AllKill int `json:"allKill"` // 通杀
		None    int `json:"none"`    // 不触发通杀通赔
	} `json:"all"` // 通杀通赔配置
	Free struct {
		Open        int `json:"open"`        // 开出鲨鱼概率
		GoldShark   int `json:"goldShark"`   // 金鲨
		SilverShark int `json:"silverShark"` // 银鲨
	} `json:"free"` // 免费游戏
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

type InitBirdAnimaConfig struct {
	Taketimes struct {
		StartRandOdds   int `json:"startrandodds"`
		Startmove       int `json:"startmove"`
		Startbet        int `json:"startbet"`
		Endmove         int `json:"endmove"`
		Endpay          int `json:"endpay"`
		EndpayAdd       int `json:"endpayadd"`
		BetGapBroadcast int `json:"betGapBroadcast"` // 下注期间广播消息
	} `json:"taketimes"`
	Unplacebetnum int `json:"unplacebetnum"`

	BetBottomLine                    int64     `json:"betBottomLine"`  // 金额低于此值时，不能下注
	IsOpen3000Ctrl                   bool      `json:"isOpen3000Ctrl"` // 是否开启3000作弊率下必输配置
	Singleusersinglespacelimit5times [][]int64 //5倍场 个人玩家单区域限红
	Singleuserallspacelimit5times    []int64   //5倍场 个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64 //5倍场 所有玩家区单区域限红
	Allspacelimit5times              []int64   //5倍场 所有区域总限红
	// 开奖策略树
	PolicyTree  ShakePolicys   `json:"policyTree"`
	BirdAnimals model.Elements `json:"birdanimals"`
}

var BirdAnimaConfig InitBirdAnimaConfig

var TaketimesMap = make(map[bridanimal.GameStatus]int, 0)

func LoadBirdAnimalConfig(filePath string) {
	if filePath == "" {
		filePath = "./config/birdanimal.json"
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("LoadBirdAminalConfig Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

func InitConfig(cfg gjson.Result) {
	Betamount := []int{}
	for _, data := range cfg.Get("betamount").Array() {
		Betamount = append(Betamount, int(data.Int()))
	}

	BirdAnimaConfig.Taketimes.StartRandOdds = int(cfg.Get("taketimes.startrandodds").Int())
	BirdAnimaConfig.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	BirdAnimaConfig.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	BirdAnimaConfig.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	BirdAnimaConfig.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())
	BirdAnimaConfig.Taketimes.EndpayAdd = int(cfg.Get("taketimes.endpayadd").Int())
	BirdAnimaConfig.Taketimes.BetGapBroadcast = int(cfg.Get("taketimes.betGapBroadcast").Int())
	BirdAnimaConfig.IsOpen3000Ctrl = cfg.Get("isOpen3000Ctrl").Bool()
	BirdAnimaConfig.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
	BirdAnimaConfig.BetBottomLine = int64(cfg.Get("betBottomLine").Int())

	// betMods := cfg.Get("betMod").Array()
	// BirdAnimaConfig.BetMod = make([]int, len(betMods))
	// for i, v := range betMods {
	// 	BirdAnimaConfig.BetMod[i] = int(v.Int())
	// }

	//BirdAnimaConfig.BirdAnimals = cfg.Get("birdanimals").Value().(model.Elements)

	for _, v := range cfg.Get("birdanimals").Array() {
		e := new(model.Element)

		if err := json.Unmarshal([]byte(v.String()), e); err != nil {
			panic(err)
		}
		BirdAnimaConfig.BirdAnimals = append(BirdAnimaConfig.BirdAnimals, e)
	}

	for _, v := range cfg.Get("policyTree").Array() {
		sp := new(ShakePolicy)

		if err := json.Unmarshal([]byte(v.String()), sp); err != nil {
			panic(err)
		}
		BirdAnimaConfig.PolicyTree = append(BirdAnimaConfig.PolicyTree, sp)
	}

	for i := 1; i <= 4; i++ {
		BirdAnimaConfig.Singleusersinglespacelimit5times = append(BirdAnimaConfig.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BirdAnimaConfig.Singleuserallspacelimit5times = append(BirdAnimaConfig.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		BirdAnimaConfig.Allusersinglespacelimit5times = append(BirdAnimaConfig.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BirdAnimaConfig.Allspacelimit5times = append(BirdAnimaConfig.Allspacelimit5times, Allspacelimit5times[0])
	}

	// 初始化map
	TaketimesMap[bridanimal.GameStatus_RandOdds] = BirdAnimaConfig.Taketimes.StartRandOdds
	TaketimesMap[bridanimal.GameStatus_StartMovie] = BirdAnimaConfig.Taketimes.Startmove
	TaketimesMap[bridanimal.GameStatus_BetStatus] = BirdAnimaConfig.Taketimes.Startbet
	TaketimesMap[bridanimal.GameStatus_EndBetMovie] = BirdAnimaConfig.Taketimes.Endmove
	TaketimesMap[bridanimal.GameStatus_SettleStatus] = BirdAnimaConfig.Taketimes.Endpay
	checkConfig()
}

func checkConfig() {
	for _, policy := range BirdAnimaConfig.PolicyTree {
		if policy.All.AllPay+policy.All.AllKill+policy.All.None != 10000 {
			panic("policy.all.allPay+policy.all.allKill+policy.all.none != 10000")
		}
		if policy.Free.GoldShark+policy.Free.SilverShark != 10000 {
			panic("policy.all.allPay+policy.all.allKill+policy.all.none != 10000")
		}
	}
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

//加载[]int64 配置文件
func loadcfg2(cfg gjson.Result, jsonConfigField string, ConfigField [][]int64) {
	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf(jsonConfigField+".%v", i)
		temp := cfg.Get(str).Array()
		var tempArr []int64
		for _, v := range temp {
			tempArr = append(tempArr, v.Int())
		}
		ConfigField = append(ConfigField, tempArr)
	}
}
