package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kubegames/kubegames-games/pkg/slots/970501/model"

	"github.com/tidwall/gjson"
)

type RoomRule struct {
	BaseBet int64 `json:"Bottom_Pouring"` // 基本低注
	// ChipArea        [5]int64 `json:"chipArea"`       // 筹码区的筹码数
	BetArea         int64 `json:"betArea"` // 每个下注区限制上线
	LimitPerUser    int64 // 单区玩家限红
	AllLimitPerArea int64 // 单个区域总限红
	AllLimitPerUser int64 // 单个玩家总限红
}

type benzBMW struct {
	Taketimes struct {
		Startmove       int `json:"startmove"`
		Startbet        int `json:"startbet"`
		Endmove         int `json:"endmove"`
		Endpay          int `json:"endpay"`
		EndpayAdd       int `json:"endpayadd"`       //
		EndpayAddNormal int `json:"endpayaddnormal"` // 增加时间
		EndpayAddTrain  int `json:"endpayaddtrain"`  // 开火车增加时间
		LoopBetGap      int `json:"loopBetGap"`
	} `json:"taketimes"`
	Unplacebetnum  int             `json:"unplacebetnum"`
	IsOpen3000Ctrl bool            `json:"isOpen3000Ctrl"` // 是否开启3000作弊率下必输配置
	Goodlucks      model.Goodlucks `json:"goodlucks"`

	Betmin                           int64
	Singleusersinglespacelimit5times [][]int64 //5倍场 个人玩家单区域限红
	Singleuserallspacelimit5times    []int64   //5倍场 个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64 //5倍场 所有玩家区单区域限红
	Allspacelimit5times              []int64   //5倍场 所有区域总限红

	Fruits model.Elements `json:"fruits"`
}

// 从./brsgj.json中读取
var BenzBMWConf benzBMW

func InitBenzBMWConf(path string) {
	if path == "" {
		path = "./config/brsgj.json"
	}
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("LoadBenzBMW Error ", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	initBenzBMWConf(result)
}

func initBenzBMWConf(cfg gjson.Result) {
	BenzBMWConf.Taketimes.Startmove = int(cfg.Get("taketimes.startmove").Int())
	BenzBMWConf.Taketimes.Startbet = int(cfg.Get("taketimes.startbet").Int())
	BenzBMWConf.Taketimes.Endmove = int(cfg.Get("taketimes.endmove").Int())
	BenzBMWConf.Taketimes.Endpay = int(cfg.Get("taketimes.endpay").Int())
	BenzBMWConf.Taketimes.EndpayAdd = int(cfg.Get("taketimes.endpayadd").Int())
	BenzBMWConf.Taketimes.EndpayAddNormal = int(cfg.Get("taketimes.endpayaddnormal").Int())
	BenzBMWConf.Taketimes.EndpayAddTrain = int(cfg.Get("taketimes.endpayaddtrain").Int())
	BenzBMWConf.Taketimes.LoopBetGap = int(cfg.Get("taketimes.loopBetGap").Int())
	BenzBMWConf.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
	BenzBMWConf.IsOpen3000Ctrl = cfg.Get("isOpen3000Ctrl").Bool()
	for _, v := range cfg.Get("fruits").Array() {
		ele := new(model.Element)
		if err := json.Unmarshal([]byte(v.String()), ele); err != nil {
			panic(err)
		}
		BenzBMWConf.Fruits = append(BenzBMWConf.Fruits, ele)
	}

	for _, v := range cfg.Get("goodlucks").Array() {
		gl := new(model.Goodluck)
		if err := json.Unmarshal([]byte(v.String()), gl); err != nil {
			panic(err)
		}
		BenzBMWConf.Goodlucks = append(BenzBMWConf.Goodlucks, gl)
	}
	for i := 1; i <= 4; i++ {
		BenzBMWConf.Singleusersinglespacelimit5times = append(BenzBMWConf.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BenzBMWConf.Singleuserallspacelimit5times = append(BenzBMWConf.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		BenzBMWConf.Allusersinglespacelimit5times = append(BenzBMWConf.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BenzBMWConf.Allspacelimit5times = append(BenzBMWConf.Allspacelimit5times, Allspacelimit5times[0])
	}

	model.ElementsAll = BenzBMWConf.Fruits
	model.GoodlucksAll = BenzBMWConf.Goodlucks
	BenzBMWConf.Betmin = cfg.Get("betmin").Int()
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
