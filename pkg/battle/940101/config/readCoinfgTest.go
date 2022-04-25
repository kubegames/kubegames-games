package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/bitly/go-simplejson"
	"github.com/sipt/GoJsoner"
)

type TestInitErRenMaJiangConfig struct {
	Cheat      []int64 `json:"cheat"`
	Round      int32   `json:"round"`
	StartMoney int     `json:"startMoney"`
	Tax        float64 `json:"tax"`
	BaseBet    int32   `json:"baseBet"`
	Path       string  `json:"path"`
	Number     int     `json:"number"`
}

var TestErRenMaJiang TestInitErRenMaJiangConfig

func (rc *TestInitErRenMaJiangConfig) TestLoadErRenMaJiangConfig() {
	data, err := ioutil.ReadFile("conf/testConfig.json")
	if err != nil {
		log.Fatalf("testConfig Error %v", err.Error())
		return
	}

	resault, _ := GoJsoner.Discard(string(data))
	rc.InitConfig(resault)
}
func (rc *TestInitErRenMaJiangConfig) InitConfig(json_str string) {
	//使用简单json来解析。
	js, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}
	rc.getCheatPoint(js)
	rc.getRoundPoint(js)
	rc.getStartMoneyPoint(js)
	rc.getTaxPoint(js)
	rc.getBaseBetPoint(js)
	rc.getPath(js)
	rc.getNumber(js)
}

func (rc *TestInitErRenMaJiangConfig) getCheatPoint(data *simplejson.Json) {
	if cheat, err := data.Get("cheat").Array(); err != nil {
		log.Fatalf("Reading testConfig Cheat Error %v", err.Error())
		return
	} else {
		for _, val := range cheat {
			tmpNum, err := val.(json.Number).Int64()
			if err != nil {
				log.Fatalf("Reading testConfig Cheat Error %v", err.Error())
			}
			rc.Cheat = append(rc.Cheat, tmpNum)
		}
		log.Println("==========rc.Path==========", rc.Cheat)
	}
}
func (rc *TestInitErRenMaJiangConfig) getRoundPoint(data *simplejson.Json) {
	if round, err := data.Get("round").Int(); err != nil {
		log.Fatalf("Reading testConfig Error %v", err.Error())
		return
	} else {
		rc.Round = int32(round)
	}
	log.Println("==========rc.Round==========", rc.Round)
}
func (rc *TestInitErRenMaJiangConfig) getStartMoneyPoint(data *simplejson.Json) {
	if startMoney, err := data.Get("startMoney").Int(); err != nil {
		log.Fatalf("Reading testConfig Error %v", err.Error())
		return
	} else {
		rc.StartMoney = startMoney
	}
}
func (rc *TestInitErRenMaJiangConfig) getTaxPoint(data *simplejson.Json) {
	if tax, err := data.Get("tax").Float64(); err != nil {
		log.Fatalf("Reading testConfig Error %v", err.Error())
		return
	} else {
		rc.Tax = tax
	}
}
func (rc *TestInitErRenMaJiangConfig) getBaseBetPoint(data *simplejson.Json) {
	if baseBet, err := data.Get("baseBet").Int(); err != nil {
		log.Fatalf("Reading testConfig Error %v", err.Error())
		return
	} else {
		rc.BaseBet = int32(baseBet)
	}
}
func (rc *TestInitErRenMaJiangConfig) getPath(data *simplejson.Json) {
	if path, err := data.Get("testLogPath").String(); err != nil {
		log.Fatalf("Reading testLogPath Error %v", err.Error())
		return
	} else {
		rc.Path = path
	}
	log.Println("==========rc.Path==========", rc.Path)
}
func (rc *TestInitErRenMaJiangConfig) getNumber(data *simplejson.Json) {
	if num, err := data.Get("testRobotNumber").Int(); err != nil {
		log.Fatalf("Reading testRobotNumber Error %v", err.Error())
		return
	} else {
		rc.Number = num
	}
	log.Println("==========rc.Number==========", rc.Number)
}
