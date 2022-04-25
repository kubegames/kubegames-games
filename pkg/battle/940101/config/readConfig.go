package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bitly/go-simplejson"
	"github.com/sipt/GoJsoner"
)

type InitErRenMaJiangConfig struct {
	OutCardTime        int32              `json:"outCardTime"`
	MaxCardMult        int32              `json:"maxMult"`
	RobotTime          []int64            `json:"robotTime"`
	StartMovie         int32              `json:"startMovie"`
	BuHuaMovie         int32              `json:"buHuaMovie"`
	OptUserTime        int32              `json:"optUserTime"`
	OptOutTime         int32              `json:"optOutTime"`
	BuHuaCurrSleepTime int32              `json:"buHuaCurrSleepTime"`
	WaitTime           int32              `json:"waitTime"`
	AutoOutCardTime    []int64            `json:"autoOutCardTime"`
	CtrlHuMapsUser     map[string]float64 `json:"ctrlHuMapsUser"`
	CtrlHuMapsRobot    map[string]float64 `json:"ctrlHuMapsRobot"`
	CtrlHandMapsUser   map[string]float64 `json:"ctrlHandMapsUser"`
	CtrlHandMapsRobot  map[string]float64 `json:"ctrlHandMapsRobot"`
	CtrlMoMapsUser     map[string]float64 `json:"ctrlMoMapsUser"`
	CtrlMoMapsRobot    map[string]float64 `json:"ctrlMoMapsRobot"`
	OutCardTimeMaps    map[string]int64   `json:"'outCardTimeMaps'"`
	OptCardTimeMaps    map[string]int64   `json:"'optCardTimeMaps'"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	ActionTimePlace     [][2]int `json:"action_time_place"`      // 机器人操作时间分布
	ActionTimeRatePlace []int    `json:"action_time_rate_place"` // 机器人操作时间概率分布
}

var RobotConf RobotConfig
var ErRenMaJiang InitErRenMaJiangConfig

func (rc *InitErRenMaJiangConfig) LoadErRenMaJiangConfig() {
	data, err := ioutil.ReadFile("conf/errenmajiang.json")
	if err != nil {
		log.Fatalf("errenmajaingconfig Error %v", err.Error())
		return
	}

	resault, _ := GoJsoner.Discard(string(data))
	rc.CtrlHuMapsUser = map[string]float64{}
	rc.CtrlHuMapsRobot = map[string]float64{}
	rc.CtrlHandMapsUser = map[string]float64{}
	rc.CtrlHandMapsRobot = map[string]float64{}
	rc.CtrlMoMapsUser = map[string]float64{}
	rc.CtrlMoMapsRobot = map[string]float64{}
	rc.OutCardTimeMaps = map[string]int64{}
	rc.OptCardTimeMaps = map[string]int64{}
	rc.InitConfig(resault)
}

func (rc *InitErRenMaJiangConfig) InitConfig(json_str string) {
	//使用简单json来解析。
	js, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}
	rc.getOutCardTime(js)
	rc.getMaxMult(js)
	rc.getRobotTime(js)
	rc.getStartMovie(js)
	rc.getBuHuaMovie(js)
	rc.getOptUserTime(js)
	rc.getOptOutTime(js)
	rc.getBuHuaCurrSleepTime(js)
	rc.getWaitTime(js)
	rc.getAutoOutCardTime(js)
	rc.getCtrlHuMaps(js)
	rc.getCtrlHandMaps(js)
	rc.getCtrlMoMaps(js)
	rc.getReBotCardTimeMaps(js)
}

func (rc *InitErRenMaJiangConfig) getOutCardTime(data *simplejson.Json) {
	if outTime, err := data.Get("outCardTime").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.OutCardTime = int32(outTime)
	}
}

func (rc *InitErRenMaJiangConfig) getMaxMult(data *simplejson.Json) {
	if maxMult, err := data.Get("maxMult").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.MaxCardMult = int32(maxMult)
	}
}

func (rc *InitErRenMaJiangConfig) getRobotTime(data *simplejson.Json) {
	if robotTime, err := data.Get("robotTime").Array(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for _, val := range robotTime {
			time1, _ := val.(json.Number).Int64()
			rc.RobotTime = append(rc.RobotTime, time1)
		}
	}
}

func (rc *InitErRenMaJiangConfig) getStartMovie(data *simplejson.Json) {
	if startMovie, err := data.Get("startMovie").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.StartMovie = int32(startMovie)
	}
}

func (rc *InitErRenMaJiangConfig) getBuHuaMovie(data *simplejson.Json) {
	if buHuaMovie, err := data.Get("buHuaMovie").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.BuHuaMovie = int32(buHuaMovie)
	}
}

func (rc *InitErRenMaJiangConfig) getOptUserTime(data *simplejson.Json) {
	if optUserTime, err := data.Get("optUserTime").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.OptUserTime = int32(optUserTime)
	}
}

func (rc *InitErRenMaJiangConfig) getOptOutTime(data *simplejson.Json) {
	if optOutTime, err := data.Get("optOutTime").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.OptOutTime = int32(optOutTime)
	}
}
func (rc *InitErRenMaJiangConfig) getBuHuaCurrSleepTime(data *simplejson.Json) {
	if buHuaCurrSleepTime, err := data.Get("buHuaCurrSleepTime").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.BuHuaCurrSleepTime = int32(buHuaCurrSleepTime)
	}
}
func (rc *InitErRenMaJiangConfig) getWaitTime(data *simplejson.Json) {
	if waitTime, err := data.Get("waitTime").Int(); err != nil {
		log.Fatalf("Reading errenmajaingconfig Error %v", err.Error())
		return
	} else {
		rc.WaitTime = int32(waitTime)
	}
}

func (rc *InitErRenMaJiangConfig) getAutoOutCardTime(data *simplejson.Json) {
	if autoOutCardTime, err := data.Get("autoOutCardTime").Array(); err != nil {
		log.Fatalf("autoOutCardTime errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for _, val := range autoOutCardTime {
			time1, _ := val.(json.Number).Int64()
			rc.AutoOutCardTime = append(rc.AutoOutCardTime, time1)
		}
	}
}

func (rc *InitErRenMaJiangConfig) getCtrlHuMaps(data *simplejson.Json) {
	if ctrlHuMaps, err := data.Get("huCtrlProbability").Get("user").Map(); err != nil {
		log.Fatalf("huCtrlProbabilityUser errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlHuMapsUser[key], _ = val.(json.Number).Float64()
		}
	}
	if ctrlHuMaps, err := data.Get("huCtrlProbability").Get("robot").Map(); err != nil {
		log.Fatalf("huCtrlProbabilityRobot errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlHuMapsRobot[key], _ = val.(json.Number).Float64()
		}
	}
	log.Println("===huCtrlProbability===", rc.CtrlHuMapsUser, rc.CtrlHuMapsRobot)
}

func (rc *InitErRenMaJiangConfig) getCtrlHandMaps(data *simplejson.Json) {
	if ctrlHuMaps, err := data.Get("AdvantageCtrlProbability").Get("user").Map(); err != nil {
		log.Fatalf("AdvantageCtrlProbabilityUser errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlHandMapsUser[key], _ = val.(json.Number).Float64()
		}
	}
	if ctrlHuMaps, err := data.Get("AdvantageCtrlProbability").Get("robot").Map(); err != nil {
		log.Fatalf("AdvantageCtrlProbabilityRobot errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlHandMapsRobot[key], _ = val.(json.Number).Float64()
		}
	}
	log.Println("===AdvantageCtrlProbability===", rc.CtrlHandMapsUser, rc.CtrlHandMapsRobot)
}

func (rc *InitErRenMaJiangConfig) getCtrlMoMaps(data *simplejson.Json) {
	if ctrlHuMaps, err := data.Get("MoCardCtrlProbability").Get("user").Map(); err != nil {
		log.Fatalf("MoCardCtrlProbabilityUser errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlMoMapsUser[key], _ = val.(json.Number).Float64()
		}
	}
	if ctrlHuMaps, err := data.Get("MoCardCtrlProbability").Get("robot").Map(); err != nil {
		log.Fatalf("MoCardCtrlProbabilityRobot errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.CtrlMoMapsRobot[key], _ = val.(json.Number).Float64()
		}
	}
	log.Println("===MoCardCtrlProbability===", rc.CtrlMoMapsUser, rc.CtrlMoMapsRobot)
}

func (rc *InitErRenMaJiangConfig) getReBotCardTimeMaps(data *simplejson.Json) {
	if ctrlHuMaps, err := data.Get("RebotOutCardTime").Get("outCard").Map(); err != nil {
		log.Fatalf("RebotOutCardTimeOutCard errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.OutCardTimeMaps[key], _ = val.(json.Number).Int64()
		}
	}
	if ctrlHuMaps, err := data.Get("RebotOutCardTime").Get("OptCard").Map(); err != nil {
		log.Fatalf("RebotOutCardTimeOptCard errenmajaingconfig Error %v", err.Error())
		return
	} else {
		for key, val := range ctrlHuMaps {
			rc.OptCardTimeMaps[key], _ = val.(json.Number).Int64()
		}
	}
}
