package gamelogic

import (
	"fmt"
	"go-game-sdk/example/game_LaBa/labacom/config"
	"go-game-sdk/example/game_LaBa/labacom/iconlogic"
	"io/ioutil"
	"math/rand"

	"github.com/tidwall/gjson"
)

type Config struct {
	Wild1       int
	Wild3       int
	BarAnd7     int
	Any7        int
	WhiteBar    int
	BlueBar     int
	RedBar      int
	White7      int
	Blue7       int
	Red7        int
	UnAward     int
	TotalWeight int
}

type ConfigLogic struct {
	icon iconlogic.Iconinfo //图形算法逻辑
}

var GameConfig []*Config

func LoadConfig() {
	data, err := ioutil.ReadFile("conf/777.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result := gjson.Parse(string(data))
	Analysiscfg(result, 3000)
	Analysiscfg(result, 2000)
	Analysiscfg(result, 1000)
	Analysiscfg(result, -1000)
	Analysiscfg(result, -2000)
	Analysiscfg(result, -3000)
	//InitConfig(result)
}

func Analysiscfg(cfg gjson.Result, cheatvalue int) {
	tmp := new(Config)
	str := fmt.Sprintf("%v.wild_1", cheatvalue)
	tmp.Wild1 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.wild_3", cheatvalue)
	tmp.Wild3 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.barand7", cheatvalue)
	tmp.BarAnd7 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.any7", cheatvalue)
	tmp.Any7 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.whitebar", cheatvalue)
	tmp.WhiteBar = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.bluebar", cheatvalue)
	tmp.BlueBar = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.redbar", cheatvalue)
	tmp.RedBar = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.white7", cheatvalue)
	tmp.White7 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.blue7", cheatvalue)
	tmp.Blue7 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.red7", cheatvalue)
	tmp.Red7 = int(cfg.Get(str).Int())
	str = fmt.Sprintf("%v.unaward", cheatvalue)
	tmp.UnAward = int(cfg.Get(str).Int())
	tmp.TotalWeight = tmp.BarAnd7 + tmp.Any7 +
		tmp.WhiteBar + tmp.BlueBar + tmp.RedBar +
		tmp.White7 + tmp.Blue7 + tmp.Red7 + tmp.UnAward

	GameConfig = append(GameConfig, tmp)
}

var BarIconID = [3]int{0, 1, 2}
var SevenIconID = [3]int{3, 4, 5}

func (cfg *ConfigLogic) GetIconArr(cheatvalue int, bFree bool, WildIndex int) [3]int {
	index := 2
	switch cheatvalue {
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

	return cfg.GetIconArrRet(GameConfig[index], bFree, WildIndex)
}

var count int

func (cfg *ConfigLogic) GetIconArrRet(gcf *Config, bFree bool, WildIndex int) [3]int {
	r := rand.Intn(10000)

	var IconArr [3]int
	HasWild := false
	TmpIndex := WildIndex
	//非免费的情况下才判断是否有wild
	if !bFree {
		//有三个wild
		if r < gcf.Wild3 {
			for i := 0; i < 3; i++ {
				IconArr[i] = config.LBConfig.Wild.IconId
			}

			return IconArr
		}
		//有一个wild
		r = rand.Intn(10000)
		if r < gcf.Wild1 {
			HasWild = true
			TmpIndex = rand.Intn(3)

			count++
		} else {
			TmpIndex = -1
		}
	} else {
		HasWild = true
	}

	r = rand.Intn(gcf.TotalWeight)
	//是bar和7的这种情况
	if r < gcf.BarAnd7 {
		cfg.BarAnd7Ret(&IconArr, HasWild, TmpIndex)
		cfg.icon.Iconarr = make([]int32, 0)
		for i := 0; i < 3; i++ {
			cfg.icon.Iconarr = append(cfg.icon.Iconarr, int32(IconArr[i]))
		}

		cfg.icon.Gettotalodds(&config.LBConfig)
		cfg.ExterIcon(1)
		return IconArr
	}

	r -= gcf.BarAnd7
	if r < gcf.Any7 {
		cfg.Any7Ret(&IconArr, HasWild, TmpIndex)
		return IconArr
	}

	r -= gcf.Any7
	if r < gcf.WhiteBar {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 0
			}
		}
		return IconArr
	}

	r -= gcf.WhiteBar
	if r < gcf.BlueBar {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 1
			}
		}
		return IconArr
	}

	r -= gcf.BlueBar
	if r < gcf.RedBar {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 2
			}
		}
		return IconArr
	}

	r -= gcf.RedBar
	if r < gcf.White7 {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 3
			}
		}
		return IconArr
	}

	r -= gcf.White7
	if r < gcf.Blue7 {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 4
			}
		}
		return IconArr
	}

	r -= gcf.Blue7
	if r < gcf.Red7 {
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = 5
			}
		}
		return IconArr
	}
	//最后剩下未中奖的算法
	for {
		cfg.icon.Iconarr = make([]int32, 0)
		for i := 0; i < 3; i++ {
			if i == TmpIndex {
				IconArr[i] = config.LBConfig.Wild.IconId
			} else {
				IconArr[i] = rand.Intn(6)
			}

			cfg.icon.Iconarr = append(cfg.icon.Iconarr, int32(IconArr[i]))
		}
		cfg.icon.Gettotalodds(&config.LBConfig)
		cfg.ExterIcon(0)
		if cfg.icon.Odds == 0 {
			break
		}
	}

	return IconArr
}

//计算出同色Bar和7的结果
func (cfg *ConfigLogic) BarAnd7Ret(IconArr *[3]int, HasWild bool, WildIndex int) {
	index := rand.Intn(3)
	iconid := BarIconID[index]
	for {
		index = rand.Intn(3)
		if index != WildIndex {
			break
		}
	}

	IconArr[index] = iconid
	index2 := index
	for {
		tmp := rand.Intn(3)
		if tmp != index && tmp != WildIndex {
			index2 = tmp
			break
		}
	}

	IconArr[index2] = iconid + 3
	iconid3 := iconid
	if HasWild {
		iconid3 = config.LBConfig.Wild.IconId
	} else {
		r := rand.Intn(2)
		if r == 1 {
			iconid3 = iconid + 3
		}
	}

	for i := 0; i < 3; i++ {
		if i != index && i != index2 {
			IconArr[i] = iconid3
			break
		}
	}
}

//计算出任意7的结果
func (cfg *ConfigLogic) Any7Ret(IconArr *[3]int, HasWild bool, WildIndex int) {
	index := rand.Intn(3)
	iconid := SevenIconID[index]
	iconindex := index
	for {
		index = rand.Intn(3)
		if index != WildIndex {
			break
		}
	}
	IconArr[index] = iconid
	index2 := index
	for {
		tmp := rand.Intn(3)
		if tmp != index && tmp != WildIndex {
			index2 = tmp
			break
		}
	}

	iconid2 := iconid
	for {
		tmp := rand.Intn(3)
		if tmp != iconindex {
			iconid2 = SevenIconID[tmp]
			break
		}
	}

	IconArr[index2] = iconid2
	r := rand.Intn(3)
	iconid3 := SevenIconID[r]
	if HasWild {
		iconid3 = config.LBConfig.Wild.IconId
	}

	for i := 0; i < 3; i++ {
		if i != index && i != index2 {
			IconArr[i] = iconid3
			break
		}
	}
}

//额外的图片情形
func (cfg *ConfigLogic) ExterIcon(v int64) {
	if cfg.icon.Odds > 0 {
		return
	}

	isarr := true
	for j := 0; j < 3; j++ {
		if cfg.icon.Iconarr[j] != 3 && cfg.icon.Iconarr[j] != 4 && cfg.icon.Iconarr[j] != 5 &&
			cfg.icon.Iconarr[j] != 6 {
			isarr = false
			break
		}
	}

	if isarr {
		cfg.icon.Odds = 2
		return
	}

	bar := -1
	seven := -1
	three := -1
	for j := 0; j < 3; j++ {
		if cfg.icon.Iconarr[j] <= int32(BarIconID[2]) && bar == -1 {
			bar = int(cfg.icon.Iconarr[j])
		} else if cfg.icon.Iconarr[j] <= int32(SevenIconID[2]) &&
			cfg.icon.Iconarr[j] > int32(BarIconID[2]) && seven == -1 {
			seven = int(cfg.icon.Iconarr[j])
		} else {
			three = int(cfg.icon.Iconarr[j])
		}
	}

	if bar+3 == seven && (bar == three || three == seven || three == 6) {
		cfg.icon.Odds = 1
	} else if v == 1 {
		fmt.Println(cfg.icon.Iconarr, bar, seven, three)
	}
}
