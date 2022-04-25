package roomconfig

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/sipt/GoJsoner"
)

type FreeGameWeight struct {
	ExtraOdds                 [3]int
	FreeGameTimes_Weight      [3]int
	FreeGameTimes_TotalWeight int
	FreeGameTimes             int
}

type FreeGameWeightCfg struct {
	Fgw [5]FreeGameWeight
}

type RoomConfig struct {
	Cheat map[int64]*FreeGameWeightCfg
}

var RoomCfg RoomConfig

func (rc *RoomConfig) LoadRoomConfig() {
	data, err := ioutil.ReadFile("conf/room.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))
	rc.analysiscfg(result)
}

func (rc *RoomConfig) analysiscfg(json_str string) {
	res, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}

	rc.Cheat = make(map[int64]*FreeGameWeightCfg)
	rc.analysisweight(3000, res)
	rc.analysisweight(2000, res)
	rc.analysisweight(1000, res)
	rc.analysisweight(-1000, res)
	rc.analysisweight(-2000, res)
	rc.analysisweight(-3000, res)
}

func (rc *RoomConfig) analysisweight(cheatvalue int64, js *simplejson.Json) {
	cheatvaluestr := strconv.FormatInt(cheatvalue, 10)
	fgwcfg := new(FreeGameWeightCfg)

	fgwcfg.Fgw[0].FreeGameTimes = 5
	fgwcfg.Fgw[0].ExtraOdds[0] = 10
	fgwcfg.Fgw[0].ExtraOdds[1] = 15
	fgwcfg.Fgw[0].ExtraOdds[2] = 30
	fgwcfg.Fgw[0].FreeGameTimes_Weight[0], _ = js.Get(cheatvaluestr).Get("5").Get("10").Int()
	fgwcfg.Fgw[0].FreeGameTimes_Weight[1], _ = js.Get(cheatvaluestr).Get("5").Get("15").Int()
	fgwcfg.Fgw[0].FreeGameTimes_Weight[2], _ = js.Get(cheatvaluestr).Get("5").Get("30").Int()
	fgwcfg.Fgw[0].FreeGameTimes_TotalWeight = fgwcfg.Fgw[0].FreeGameTimes_Weight[0] +
		fgwcfg.Fgw[0].FreeGameTimes_Weight[1] + fgwcfg.Fgw[0].FreeGameTimes_Weight[2]

	fgwcfg.Fgw[1].FreeGameTimes = 8
	fgwcfg.Fgw[1].ExtraOdds[0] = 8
	fgwcfg.Fgw[1].ExtraOdds[1] = 10
	fgwcfg.Fgw[1].ExtraOdds[2] = 15
	fgwcfg.Fgw[1].FreeGameTimes_Weight[0], _ = js.Get(cheatvaluestr).Get("8").Get("8").Int()
	fgwcfg.Fgw[1].FreeGameTimes_Weight[1], _ = js.Get(cheatvaluestr).Get("8").Get("10").Int()
	fgwcfg.Fgw[1].FreeGameTimes_Weight[2], _ = js.Get(cheatvaluestr).Get("8").Get("15").Int()
	fgwcfg.Fgw[1].FreeGameTimes_TotalWeight = fgwcfg.Fgw[1].FreeGameTimes_Weight[0] +
		fgwcfg.Fgw[1].FreeGameTimes_Weight[1] + fgwcfg.Fgw[1].FreeGameTimes_Weight[2]

	fgwcfg.Fgw[2].FreeGameTimes = 10
	fgwcfg.Fgw[2].ExtraOdds[0] = 5
	fgwcfg.Fgw[2].ExtraOdds[1] = 8
	fgwcfg.Fgw[2].ExtraOdds[2] = 10
	fgwcfg.Fgw[2].FreeGameTimes_Weight[0], _ = js.Get(cheatvaluestr).Get("10").Get("5").Int()
	fgwcfg.Fgw[2].FreeGameTimes_Weight[1], _ = js.Get(cheatvaluestr).Get("10").Get("8").Int()
	fgwcfg.Fgw[2].FreeGameTimes_Weight[2], _ = js.Get(cheatvaluestr).Get("10").Get("10").Int()
	fgwcfg.Fgw[2].FreeGameTimes_TotalWeight = fgwcfg.Fgw[2].FreeGameTimes_Weight[0] +
		fgwcfg.Fgw[2].FreeGameTimes_Weight[1] + fgwcfg.Fgw[2].FreeGameTimes_Weight[2]

	fgwcfg.Fgw[3].FreeGameTimes = 15
	fgwcfg.Fgw[3].ExtraOdds[0] = 3
	fgwcfg.Fgw[3].ExtraOdds[1] = 5
	fgwcfg.Fgw[3].ExtraOdds[2] = 8
	fgwcfg.Fgw[3].FreeGameTimes_Weight[0], _ = js.Get(cheatvaluestr).Get("15").Get("3").Int()
	fgwcfg.Fgw[3].FreeGameTimes_Weight[1], _ = js.Get(cheatvaluestr).Get("15").Get("5").Int()
	fgwcfg.Fgw[3].FreeGameTimes_Weight[2], _ = js.Get(cheatvaluestr).Get("15").Get("8").Int()
	fgwcfg.Fgw[3].FreeGameTimes_TotalWeight = fgwcfg.Fgw[3].FreeGameTimes_Weight[0] +
		fgwcfg.Fgw[3].FreeGameTimes_Weight[1] + fgwcfg.Fgw[3].FreeGameTimes_Weight[2]

	fgwcfg.Fgw[4].FreeGameTimes = 20
	fgwcfg.Fgw[4].ExtraOdds[0] = 2
	fgwcfg.Fgw[4].ExtraOdds[1] = 3
	fgwcfg.Fgw[4].ExtraOdds[2] = 5
	fgwcfg.Fgw[4].FreeGameTimes_Weight[0], _ = js.Get(cheatvaluestr).Get("20").Get("2").Int()
	fgwcfg.Fgw[4].FreeGameTimes_Weight[1], _ = js.Get(cheatvaluestr).Get("20").Get("3").Int()
	fgwcfg.Fgw[4].FreeGameTimes_Weight[2], _ = js.Get(cheatvaluestr).Get("20").Get("5").Int()
	fgwcfg.Fgw[4].FreeGameTimes_TotalWeight = fgwcfg.Fgw[4].FreeGameTimes_Weight[0] +
		fgwcfg.Fgw[4].FreeGameTimes_Weight[1] + fgwcfg.Fgw[4].FreeGameTimes_Weight[2]

	rc.Cheat[cheatvalue] = fgwcfg
}

func (rc *RoomConfig) GetFreeGameWeight(cheatvalue int32) *FreeGameWeightCfg {
	fgw, ok := rc.Cheat[int64(cheatvalue)]
	if ok {
		return fgw
	}

	fgw, _ = rc.Cheat[1000]

	return fgw
}
