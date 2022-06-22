//读取配置文件

package config

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
	Weight      map[int64]int //图片权重
	TotalWeight int           //本位置图片总权重
}

//作弊
type CheatCfg struct {
	Normalicon        map[int64]Iconweight
	Freeicon          map[int64]Iconweight
	SpecialStyleicon  map[int64]Iconweight //特殊玩法 图标配置 权游龙母模式
	Limit             int
	SpecialStyleLimit int //特殊玩法限制
}

//免费游戏配置
type FreeGameCfg struct {
	Times      []int64
	IconId     int
	Type       int
	LimitCount int
}

type WildCfg struct {
	IconId     int //图标id
	Type       int //类型
	LimitCount int
}

type JackpotGame struct {
	IconId int64 //图标id
	Award  int64 //奖励（award）的万分比
	Weight int64 //权重
}

type JackpotCfg struct {
	IconId      int                   //图标id
	Type        int                   //类型
	Award       []int64               //图标个数对应的奖励，奖励奖池中万分比
	Limitcount  int                   //图标出现的个数限制
	AwardType   int                   //获奖以后的类型，1小游戏，2直接按照award里面的比例给钱
	JackpotIcon map[int64]JackpotGame //奖金池小游戏配置
}

type LabaConfig struct {
	Cheat     map[int64]CheatCfg   //一个作弊值中的权重
	Special   map[int64]Iconweight //特殊的作弊配置
	LineCount int                  //线数
	Line      int                  //行
	Row       int                  //列
	AwardType int                  //记奖方式
	FreeGame  FreeGameCfg          //免费游戏
	Wild      WildCfg              //wild图标配置
	Jackpot   JackpotCfg           //奖金池游戏配置
	IconAward map[int64][]int64    //图标配置，对应客户端的图标ID，不能为0//图标对应的奖励，对应1,2,3,4,5个时候的倍数
	BetConfig [][]int64            //下注配置
	LineBall  [][]int64            //二维数组，线的连线，对应轴的排列，线数
	Matrix    [][]int64            //矩形图标排列，左起第一轴从上往下数，如果对应的位置是空格，用0填充，从1开始
}

var LBConfig LabaConfig

//读取配置文件
func (lbcfg *LabaConfig) LoadLabadCfg() {
	data, err := ioutil.ReadFile("conf/laba.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))
	lbcfg.analysiscfg(result)
}

//解析配置文件
func (lbcfg *LabaConfig) analysiscfg(json_str string) {
	//使用简单json来解析。
	res, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}
	lbcfg.getcheatcfg(res)
	lbcfg.getlineinfo(res)
	lbcfg.getfreegameinfo(res)
	lbcfg.getwildcfg(res)
	lbcfg.getjackpotcfg(res)
	lbcfg.geticonaward(res)
	lbcfg.getbetconfig(res)
	lbcfg.getlineball(res)
	lbcfg.getmatrix(res)
}

//获取矩阵
func (lbcfg *LabaConfig) getmatrix(js *simplejson.Json) {
	arr, _ := js.Get("matrix").Array()
	for n := 0; n < len(arr); n++ {
		var mx []int64
		for m := 0; m < len(arr[n].([]interface{})); m++ {
			if v, err := arr[n].([]interface{})[m].(json.Number).Int64(); err == nil {
				mx = append(mx, v)
			}
		}

		lbcfg.Matrix = append(lbcfg.Matrix, mx)
	}

	lbcfg.Line = len(lbcfg.Matrix[0])
}

//获取图标奖励
func (lbcfg *LabaConfig) getbetconfig(js *simplejson.Json) {
	var i int64
	i = 1

	for {
		str := strconv.FormatInt(i, 10)
		betvalue, ok := js.Get("bet").Get(str).Array()
		if ok != nil {
			break
		}
		var arr []int64
		for n := 0; n < len(betvalue); n++ {
			if v, err := betvalue[n].(json.Number).Int64(); err == nil {
				arr = append(arr, v)
			}
		}
		lbcfg.BetConfig = append(lbcfg.BetConfig, arr)
		i++
	}
}

//获取压线的数组
func (lbcfg *LabaConfig) getlineball(js *simplejson.Json) {
	arr, _ := js.Get("lineball").Array()
	for n := 0; n < len(arr); n++ {
		var lb []int64
		for m := 0; m < len(arr[n].([]interface{})); m++ {
			if v, err := arr[n].([]interface{})[m].(json.Number).Int64(); err == nil {
				lb = append(lb, v)
			}
		}

		lbcfg.LineBall = append(lbcfg.LineBall, lb)
	}
}

//获取图标奖励
func (lbcfg *LabaConfig) geticonaward(js *simplejson.Json) {
	lbcfg.IconAward = make(map[int64][]int64)
	var i int64
	i = 0
	for {
		str := strconv.FormatInt(i, 10)
		iconarr, ok := js.Get("iconaward").Get(str).Array()
		if ok != nil {
			break
		}
		var arr []int64
		for n := 0; n < len(iconarr); n++ {
			if v, err := iconarr[n].(json.Number).Int64(); err == nil {
				arr = append(arr, v)
			}
		}
		lbcfg.IconAward[i] = arr
		i++
	}
}

//奖金池游戏配置
func (lbcfg *LabaConfig) getjackpotcfg(js *simplejson.Json) {
	var err error
	lbcfg.Jackpot.IconId, err = js.Get("jackpot").Get("icon").Int()
	if err != nil {
		lbcfg.Jackpot.IconId = -1
		return
	}

	lbcfg.Jackpot.Type, _ = js.Get("jackpot").Get("type").Int()

	award, _ := js.Get("jackpot").Get("award").Array()
	for i := 0; i < len(award); i++ {
		if v, err := award[i].(json.Number).Int64(); err == nil {
			lbcfg.Jackpot.Award = append(lbcfg.Jackpot.Award, v)
		}
	}

	lbcfg.Jackpot.Limitcount, _ = js.Get("jackpot").Get("limitcount").Int()
	lbcfg.Jackpot.AwardType, _ = js.Get("jackpot").Get("awardtype").Int()

	lbcfg.Jackpot.JackpotIcon = make(map[int64]JackpotGame)

	var i int64
	i = 1
	for {
		str := strconv.FormatInt(i, 10)
		iconarr, ok := js.Get("jackpot").Get("jackpoticon").Get(str).Array()
		if ok != nil {
			break
		}

		var jg JackpotGame
		jg.IconId, _ = iconarr[0].(json.Number).Int64()
		jg.Award, _ = iconarr[1].(json.Number).Int64()
		jg.Weight, _ = iconarr[2].(json.Number).Int64()
		lbcfg.Jackpot.JackpotIcon[i] = jg
		i++
	}
}

//获取wild配置
func (lbcfg *LabaConfig) getwildcfg(js *simplejson.Json) {
	var err error
	lbcfg.Wild.IconId, err = js.Get("wild").Get("icon").Int()
	if err != nil {
		lbcfg.Wild.IconId = -1
		return
	}
	lbcfg.Wild.Type, _ = js.Get("wild").Get("type").Int()
	lbcfg.Wild.LimitCount, _ = js.Get("wild").Get("limitcount").Int()
}

//获取免费图标配置
func (lbcfg *LabaConfig) getfreegameinfo(js *simplejson.Json) {
	var err error
	lbcfg.FreeGame.IconId, err = js.Get("freegame").Get("icon").Int()
	if err != nil {
		lbcfg.FreeGame.IconId = -1
		return
	}

	lbcfg.FreeGame.Type, err = js.Get("freegame").Get("type").Int()
	lbcfg.FreeGame.LimitCount, err = js.Get("freegame").Get("limitcount").Int()

	times, _ := js.Get("freegame").Get("times").Array()

	for i := 0; i < len(times); i++ {
		if v, err := times[i].(json.Number).Int64(); err == nil {
			lbcfg.FreeGame.Times = append(lbcfg.FreeGame.Times, v)
		}
	}
}

//获取线
func (lbcfg *LabaConfig) getlineinfo(js *simplejson.Json) {
	lbcfg.LineCount, _ = js.Get("linecount").Int()
	lbcfg.Row, _ = js.Get("row").Int()
	lbcfg.AwardType, _ = js.Get("awardtype").Int()
}

//获取作弊配置
func (lbcfg *LabaConfig) getcheatcfg(js *simplejson.Json) {
	lbcfg.Cheat = make(map[int64]CheatCfg)
	lbcfg.getcheatinfo(js, 3000)
	lbcfg.getcheatinfo(js, 2000)
	lbcfg.getcheatinfo(js, 1000)
	lbcfg.getcheatinfo(js, -1000)
	lbcfg.getcheatinfo(js, -2000)
	lbcfg.getcheatinfo(js, -3000)
	lbcfg.getspecialcheatinfo(js)
}

//获取对应作弊值下的配置
func (lbcfg *LabaConfig) getcheatinfo(js *simplejson.Json, cheatvalue int64) {
	cheatvaluestr := strconv.FormatInt(cheatvalue, 10)
	var c CheatCfg
	c.Normalicon = make(map[int64]Iconweight)
	c.Freeicon = make(map[int64]Iconweight)
	c.SpecialStyleicon = make(map[int64]Iconweight)
	c.Limit, _ = js.Get(cheatvaluestr).Get("limit").Int()
	c.SpecialStyleLimit, _ = js.Get(cheatvaluestr).Get("specialstylelimit").Int()
	var i int64
	i = 1
	for {
		str := strconv.FormatInt(i, 10)
		var j int64
		if _, ok := js.Get(cheatvaluestr).Get("normal").Get(str).Get("0").Int(); ok != nil {
			break
		}
		j = 0
		var iw Iconweight
		var iwfree Iconweight
		var iwspecialstyle Iconweight
		iw.Weight = make(map[int64]int)
		iw.TotalWeight = 0
		iwfree.Weight = make(map[int64]int)
		iwfree.TotalWeight = 0
		iwspecialstyle.Weight = make(map[int64]int)
		iwspecialstyle.TotalWeight = 0
		for {
			str1 := strconv.FormatInt(j, 10)
			if v, ok := js.Get(cheatvaluestr).Get("normal").Get(str).Get(str1).Int(); ok == nil {
				//正常游戏下的
				iw.Weight[j] = v
				iw.TotalWeight += v
				//免费游戏下的
				vfree, _ := js.Get(cheatvaluestr).Get("free").Get(str).Get(str1).Int()
				iwfree.Weight[j] = vfree
				iwfree.TotalWeight += vfree
				//特殊游戏龙母
				vdragonmother, _ := js.Get(cheatvaluestr).Get("specialstyle").Get(str).Get(str1).Int()
				iwspecialstyle.Weight[j] = vdragonmother
				iwspecialstyle.TotalWeight += vdragonmother
				j++
			} else {
				c.Normalicon[i] = iw
				c.Freeicon[i] = iwfree
				c.SpecialStyleicon[i] = iwspecialstyle
				break
			}
		}

		i++
	}
	lbcfg.Cheat[cheatvalue] = c
}

//获取特殊作弊配置下的值
func (lbcfg *LabaConfig) getspecialcheatinfo(js *simplejson.Json) {
	lbcfg.Special = make(map[int64]Iconweight)
	var i int64
	i = 1
	for {
		str := strconv.FormatInt(i, 10)
		var j int64
		if _, ok := js.Get("special").Get(str).Get("0").Int(); ok != nil {
			break
		}
		j = 0
		var iw Iconweight
		iw.Weight = make(map[int64]int)
		iw.TotalWeight = 0
		for {
			str1 := strconv.FormatInt(j, 10)
			if v, ok := js.Get("special").Get(str).Get(str1).Int(); ok == nil {
				iw.Weight[j] = v
				iw.TotalWeight += v
				j++
			} else {
				lbcfg.Special[i] = iw
				break
			}
		}

		i++
	}
}

func (lbcfg *LabaConfig) GetCheatCfg(cheatvalue int64) CheatCfg {
	cfg, ok := lbcfg.Cheat[cheatvalue]
	if ok {
		return cfg
	}

	cfg1, _ := lbcfg.Cheat[1000]
	return cfg1
}
