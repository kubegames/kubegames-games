package config

import (
	"encoding/json"
	"fmt"
	"game_LaBa/benzbmw/model"
	"io/ioutil"
	"log"
	"math/rand"

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

type WinControl struct {
	RoomProb int `json:"roomProb"` // 血池值
	Surprise struct {
		Prob      int        `json:"prob"`  // 中特殊奖项的概率
		Three     int        `json:"three"` // 大三元概率
		ThreeCtrl ThreeCtrls `json:"threeCtrl"`
		Four      int        `json:"four"` // 大四喜概率
		FourCtrl  FourCtrls  `json:"fourCtrl"`
	} `json:"surprise"` // 大三元/大四喜
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

type WinControls []*WinControl

func (wcs WinControls) Find(roomProb int) *WinControl {
	for _, v := range wcs {
		if v.RoomProb == roomProb {
			return v
		}
	}
	return wcs[rand.Intn(len(wcs))]
}

type ThreeCtrl struct {
	ElemType  int `json:"elemType"`  // 元素类型 0x10：奔驰；0x20：宝马；0x30：雷克萨斯；0x40：大众
	ShakeProb int `json:"shakeProb"` // 选中概率
}

type ThreeCtrls []ThreeCtrl

func (tcs ThreeCtrls) Rand() ThreeCtrl {
	var allWeight int
	for _, v := range tcs {
		allWeight += v.ShakeProb
	}

	randWeight := rand.Intn(allWeight) + 1
	for _, v := range tcs {
		if randWeight <= v.ShakeProb {
			return v
		}
		randWeight -= v.ShakeProb
	}

	return tcs[rand.Intn(len(tcs))]
}

type FourCtrl struct {
	Color     int `json:"color"`     // 颜色类型 0x01:黄金；0x02:白银；0x03:黄铜
	ShakeProb int `json:"shakeProb"` // 选中概率
}

type FourCtrls []FourCtrl

func (fcs FourCtrls) Rand() FourCtrl {
	var allWeight int
	for _, v := range fcs {
		allWeight += v.ShakeProb
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range fcs {
		if v.ShakeProb == 0 {
			continue
		}
		if randWeight <= v.ShakeProb {
			return v
		}
		randWeight -= v.ShakeProb
	}
	return fcs[rand.Intn(len(fcs))]
}

type benzBMW struct {
	Taketimes struct {
		Startmove     int `json:"startmove"`
		Startbet      int `json:"startbet"`
		Endmove       int `json:"endmove"`
		Endpay        int `json:"endpay"`
		EndpayAdd     int `json:"endpayadd"`     // 大三元增加时间
		EndpayAddFour int `json:"endpayaddfour"` // 大四喜增加时间
		LoopBetGap    int `json:"loopBetGap"`
	} `json:"taketimes"`
	Unplacebetnum int `json:"unplacebetnum"`

	IsOpen3000Ctrl bool        `json:"isOpen3000Ctrl"` // 是否开启3000作弊率下必输配置
	WinCtrl        WinControls `json:"winCtrl"`

	Cars model.ElemBases `json:"cars"`

	Betmin                           int64
	Singleusersinglespacelimit5times [][]int64 //5倍场 个人玩家单区域限红
	Singleuserallspacelimit5times    []int64   //5倍场 个人玩家所有区域限红
	Allusersinglespacelimit5times    [][]int64 //5倍场 所有玩家区单区域限红
	Allspacelimit5times              []int64   //5倍场 所有区域总限红
}

// 从./benzbmw.json中读取
var BenzBMWConf benzBMW

func InitBenzBMWConf(filepath string) {
	if filepath == "" {
		filepath = "./config/benzbmw.json"
	}
	data, err := ioutil.ReadFile(filepath)
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
	BenzBMWConf.Taketimes.EndpayAddFour = int(cfg.Get("taketimes.endpayaddfour").Int())
	BenzBMWConf.Taketimes.LoopBetGap = int(cfg.Get("taketimes.loopBetGap").Int())
	BenzBMWConf.Unplacebetnum = int(cfg.Get("unplacebetnum").Int())
	BenzBMWConf.IsOpen3000Ctrl = cfg.Get("isOpen3000Ctrl").Bool()

	for _, v := range cfg.Get("winCtrl").Array() {
		wc := new(WinControl)
		if err := json.Unmarshal([]byte(v.String()), wc); err != nil {
			panic(err)
		}
		BenzBMWConf.WinCtrl = append(BenzBMWConf.WinCtrl, wc)
	}
	for _, v := range cfg.Get("cars").Array() {
		var eb model.ElemBase
		if err := json.Unmarshal([]byte(v.String()), &eb); err != nil {
			panic(err)
		}
		BenzBMWConf.Cars = append(BenzBMWConf.Cars, eb)
	}

	model.ElemShakeProbSlice = BenzBMWConf.Cars

	for i := 1; i <= 4; i++ {
		BenzBMWConf.Singleusersinglespacelimit5times = append(BenzBMWConf.Singleusersinglespacelimit5times, loadcfg(cfg, i, "singleusersinglespacelimit5times"))
		Singleuserallspacelimit5times := loadcfg(cfg, i, "singleuserallspacelimit5times")
		BenzBMWConf.Singleuserallspacelimit5times = append(BenzBMWConf.Singleuserallspacelimit5times, Singleuserallspacelimit5times[0])
		BenzBMWConf.Allusersinglespacelimit5times = append(BenzBMWConf.Allusersinglespacelimit5times, loadcfg(cfg, i, "allusersinglespacelimit5times"))
		Allspacelimit5times := loadcfg(cfg, i, "allspacelimit5times")
		BenzBMWConf.Allspacelimit5times = append(BenzBMWConf.Allspacelimit5times, Allspacelimit5times[0])
	}

	BenzBMWConf.Betmin = cfg.Get("betmin").Int()

	fmt.Println("BenzBMWConf.IsOpen3000Ctrl = ", BenzBMWConf.IsOpen3000Ctrl)

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
