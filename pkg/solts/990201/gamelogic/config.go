package gamelogic

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"io/ioutil"
	"math/rand"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
	"github.com/tidwall/gjson"
)

type ProConfig struct {
	HumanPro int
	ArmsPro  int
}

type Config struct {
	HumanOdd int
	ArmsOdd  int
	CheatCfg []*ProConfig
	Human    map[int64]config.Iconweight //特殊的作弊配置
	Arms     map[int64]config.Iconweight //特殊的作弊配置
}

type FullIcon struct {
	Iconarr []int32
	Odd     int
	Type    int //测试用全图类型
}

var GameConfig Config

func LoadConfig() {
	data, err := ioutil.ReadFile("conf/shz.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}

	result := gjson.Parse(string(data))
	GameConfig.GetOdds(result)
	GameConfig.GetPro(result, 3000)
	GameConfig.GetPro(result, 2000)
	GameConfig.GetPro(result, 1000)
	GameConfig.GetPro(result, -1000)
	GameConfig.GetPro(result, -2000)
	GameConfig.GetPro(result, -3000)
	json_str, _ := GoJsoner.Discard(string(data))
	js, err := simplejson.NewJson([]byte(json_str))
	GameConfig.GetArmsIconWeight(js)
	GameConfig.GetHumanIconWeight(js)

	log.Traceln(GameConfig)
}

func (cfg *Config) GetOdds(res gjson.Result) {
	cfg.HumanOdd = int(res.Get("humanodd").Int())
	cfg.ArmsOdd = int(res.Get("armsodd").Int())
}

func (cfg *Config) GetPro(res gjson.Result, v int) {
	tmp := new(ProConfig)
	str := fmt.Sprintf("%v.human", v)
	tmp.HumanPro = int(res.Get(str).Int())
	str = fmt.Sprintf("%v.arms", v)
	tmp.ArmsPro = int(res.Get(str).Int())

	cfg.CheatCfg = append(cfg.CheatCfg, tmp)
}

func (cfg *Config) GetArmsIconWeight(js *simplejson.Json) {
	cfg.Arms = make(map[int64]config.Iconweight)
	var i int64
	i = 1
	for m := 0; m < 5; m++ {
		str := strconv.FormatInt(i, 10)
		var j int64
		j = 0
		var iw config.Iconweight
		iw.Weight = make(map[int64]int)
		iw.TotalWeight = 0
		for {
			str1 := strconv.FormatInt(j, 10)
			if v, ok := js.Get("arms").Get(str).Get(str1).Int(); ok == nil {
				iw.Weight[j] = v
				iw.TotalWeight += v
				j++
			} else {
				cfg.Arms[i] = iw
				break
			}
		}

		i++
	}
}

func (cfg *Config) GetHumanIconWeight(js *simplejson.Json) {
	cfg.Human = make(map[int64]config.Iconweight)
	var i int64
	i = 1
	for m := 0; m < 5; m++ {
		str := strconv.FormatInt(i, 10)
		var j int64
		j = 0
		var iw config.Iconweight
		iw.Weight = make(map[int64]int)
		iw.TotalWeight = 0
		for {
			str1 := strconv.FormatInt(j, 10)
			if v, ok := js.Get("human").Get(str).Get(str1).Int(); ok == nil {
				iw.Weight[j] = v
				iw.TotalWeight += v
				j++
			} else {
				cfg.Human[i] = iw
				break
			}
		}

		i++
	}
}

func (fi *FullIcon) GetFullIcon(v int) (bFull bool) {
	index := 2
	switch v {
	case 3000:
		index = 0
	case 2000:
		index = 1
	case 1000:
		index = 2
	case -1000:
		index = 3
	case -2000:
		index = 4
	case -3000:
		index = 5
	default:
		index = 2
	}

	var iconweightcfg map[int64]config.Iconweight
	bFull = false
	fi.Odd = 0
	fi.Type = 0
	r := rand.Intn(10000)
	if r < GameConfig.CheatCfg[index].ArmsPro {
		bFull = true
		fi.Odd = GameConfig.ArmsOdd
		iconweightcfg = GameConfig.Arms
		fi.Type = 1
	} else {
		r := rand.Intn(10000)
		if r < GameConfig.CheatCfg[index].HumanPro {
			bFull = true
			fi.Odd = GameConfig.HumanOdd
			iconweightcfg = GameConfig.Human
			fi.Type = 2
		}
	}

	if bFull {
		fi.geticonbyweight(&config.LBConfig, iconweightcfg)
	}

	return bFull
}

//通过权重计算出对应的图标
func (fi *FullIcon) geticonbyweight(lbcfg *config.LabaConfig, iconweightcfg map[int64]config.Iconweight) {
	fi.Iconarr = make([]int32, 0)
	for i := 1; i <= 5; i++ {
		for j := 0; j < 3; j++ {

			w := iconweightcfg[int64(i)].TotalWeight
			r := rand.Intn(w)
			for k, v := range iconweightcfg[int64(i)].Weight {
				if r < v {
					//将生成的图标加入到列表中去
					fi.Iconarr = append(fi.Iconarr, int32(k))

					break
				}

				r -= v
			}
		}
	}
}

func (fi *FullIcon) IsFullIcon(Iconarr []int32) (bFull bool) {
	HasArms := false
	HasHuman := false
	bFull = true
	for _, v := range Iconarr {
		if v <= 2 {
			HasArms = true
		} else if v <= 4 {
			HasHuman = true
		} else if v != int32(config.LBConfig.Wild.IconId) {
			bFull = false
			break
		}

		if HasArms && HasHuman {
			bFull = false
			break
		}
	}

	return bFull
}
