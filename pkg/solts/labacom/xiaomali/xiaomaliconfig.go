package xiaomali

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

type Iconweight struct {
	Weight      map[int64][2]int64 //图片权重
	Icon        []int              //图片ID
	TotalWeight int64              //本位置图片总权重
}

type CheatCfg struct {
	Limit       int        //中奖总倍数限制
	Innerweight Iconweight //图片中奖的权重
	Three       int        //三连开关
}

type XiaoMaLiCfg struct {
	Cheat     map[int64]CheatCfg //一个作弊值中的权重
	IconAward map[int64]int      //图标配置，对应客户端的图标ID，不能为0
	Iconrank  []int              //图片在界面上的排列，从左上角开始顺时针排列
}

var XMLConfig XiaoMaLiCfg

//读取文件
func (xml *XiaoMaLiCfg) LoadXiaoMaLiCfg() {
	data, err := ioutil.ReadFile("conf/xiaomali.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))
	xml.analysiscfg(result)
}

//解析配置文件
func (xml *XiaoMaLiCfg) analysiscfg(json_str string) {
	//使用简单json来解析。
	res, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		log.Traceln("XiaoMaLiCfg", json_str)
		fmt.Printf("%v\n", err)
		return
	}

	xml.getcheatcfg(res)
	xml.geticonaward(res)
	xml.geticonrank(res)
}

//获取作弊配置
func (xml *XiaoMaLiCfg) getcheatcfg(js *simplejson.Json) {
	xml.Cheat = make(map[int64]CheatCfg)
	xml.getcheatinfo(js, 3000)
	xml.getcheatinfo(js, 2000)
	xml.getcheatinfo(js, 1000)
	xml.getcheatinfo(js, -1000)
	xml.getcheatinfo(js, -2000)
	xml.getcheatinfo(js, -3000)
}

//获取对应作弊值下的配置
func (xml *XiaoMaLiCfg) getcheatinfo(js *simplejson.Json, cheatvalue int64) {
	cheatvaluestr := strconv.FormatInt(cheatvalue, 10)
	var c CheatCfg
	c.Innerweight.Weight = make(map[int64][2]int64)
	c.Limit, _ = js.Get(cheatvaluestr).Get("limit").Int()
	c.Three, _ = js.Get(cheatvaluestr).Get("three").Int()
	var i int64
	i = -1
	for {
		str := strconv.FormatInt(i, 10)
		arr, ok := js.Get(cheatvaluestr).Get(str).Array()
		if ok != nil {
			break
		}

		award, _ := arr[0].(json.Number).Int64()
		two, _ := arr[1].(json.Number).Int64()
		weight := [2]int64{award, two}
		c.Innerweight.Weight[i] = weight
		c.Innerweight.TotalWeight += award
		if i != -1 {
			c.Innerweight.Icon = append(c.Innerweight.Icon, int(i))
		}
		i++
	}

	xml.Cheat[cheatvalue] = c
}

//获取图标奖励
func (xml *XiaoMaLiCfg) geticonaward(js *simplejson.Json) {
	xml.IconAward = make(map[int64]int)
	var i int64
	i = 0
	for {
		str := strconv.FormatInt(i, 10)
		v, ok := js.Get("iconaward").Get(str).Int()
		if ok != nil {
			break
		}

		xml.IconAward[i] = v
		i++
	}
}

func (xml *XiaoMaLiCfg) geticonrank(js *simplejson.Json) {
	arr, _ := js.Get("iconrank").Array()

	for i := 0; i < len(arr); i++ {
		v, _ := arr[i].(json.Number).Int64()
		xml.Iconrank = append(xml.Iconrank, int(v))
	}
}

func (xml *XiaoMaLiCfg) GetIconWeight(cheatvalue int64) CheatCfg {
	c, err := xml.Cheat[cheatvalue]
	if err {
		return c
	}

	cv, _ := xml.Cheat[-1000]
	return cv
}
