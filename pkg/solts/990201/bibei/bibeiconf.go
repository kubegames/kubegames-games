package bibei

import (
	"encoding/json"
	"fmt"
	"go-service/lib/log"
	"io/ioutil"
	"math/rand"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/sipt/GoJsoner"
)

const (
	HE    int = 0
	SMALL int = 1
	BIG   int = 2
)

type CheatCfg struct {
	resultHe int //开和的概率，万分比
	userWin  int //玩家赢的概率，万分比
	pairs    int //玩家赢然后开对子的概率
}

type BiBeiConf struct {
	cheat map[int]CheatCfg
	odds  [4]int
}

var BBConfig BiBeiConf

//读取文件
func (bbc *BiBeiConf) LoadBiBeiCfg() {
	data, err := ioutil.ReadFile("conf/bibei.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))
	bbc.analysiscfg(result)
}

//解析配置文件
func (bbc *BiBeiConf) analysiscfg(json_str string) {
	//使用简单json来解析。
	res, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	bbc.getcheatcfg(res)
	bbc.getOdds(res)
}

//获取作弊配置
func (bbc *BiBeiConf) getcheatcfg(js *simplejson.Json) {
	bbc.cheat = make(map[int]CheatCfg)
	bbc.getCheatInfo(js, 3000)
	bbc.getCheatInfo(js, 2000)
	bbc.getCheatInfo(js, 1000)
	bbc.getCheatInfo(js, -1000)
	bbc.getCheatInfo(js, -2000)
	bbc.getCheatInfo(js, -3000)
}

//获取对应作弊值下的配置
func (bbc *BiBeiConf) getCheatInfo(js *simplejson.Json, cheatvalue int) {
	cheatvaluestr := strconv.FormatInt(int64(cheatvalue), 10)
	var c CheatCfg

	c.resultHe, _ = js.Get(cheatvaluestr).Get("resulthe").Int()
	c.userWin, _ = js.Get(cheatvaluestr).Get("userwin").Int()
	c.pairs, _ = js.Get(cheatvaluestr).Get("pairs").Int()

	bbc.cheat[cheatvalue] = c
}

func (bbc *BiBeiConf) getOdds(js *simplejson.Json) {
	arr, _ := js.Get("odds").Array()
	for i := 0; i < len(arr); i++ {

		v, _ := arr[i].(json.Number).Int64()
		bbc.odds[i] = int(v)
	}
}

func (bbc *BiBeiConf) GetCheatConf(v int) CheatCfg {
	c, err := bbc.cheat[v]
	if err {
		return c
	}

	c1, _ := bbc.cheat[1000]
	return c1
}

func GetOdds(bbc *BiBeiConf, d1 int, d2 int) int {
	if d1 == d2 {
		return bbc.odds[3]
	} else if (d1 + d2) == 7 {
		return bbc.odds[2]
	}

	return bbc.odds[0]
}

//获取比倍的结果t玩家压的结果
func GetBiBeiRes(bbc *BiBeiConf, CheatVaule int, t int) (int, int, bool) {
	log.Tracef("作弊值为：%v", CheatVaule)
	retype, win := GetDiceType(bbc, CheatVaule, t)
	d1, d2 := GetDice(retype)
	return d1, d2, win
}

func GetDiceType(bbc *BiBeiConf, CheatVaule int, t int) (int, bool) {
	//先判断是否开和
	r := int(rand.Int31n(10000))
	c := bbc.GetCheatConf(CheatVaule)
	if r < c.resultHe {
		return 0, t == HE
	}

	if t == HE {
		if rand.Int31n(2) == 0 {
			return 5, false
		} else {
			return 6, false
		}
	}

	r = int(rand.Int31n(10000))
	if r < c.userWin {
		r = int(rand.Int31n(10000))
		if r < c.pairs {
			return t + 2, true
		} else {
			return t, true
		}
	}

	if t == SMALL {
		return 6, false
	} else {
		return 5, false
	}
}

func GetDice(t int) (int, int) {
	log.Tracef("获取结果：%v", t)
	r := 0
	r2 := 0
	switch t {
	case 0: //结果是和
		r, r2 = GetDiceHe()
		break
	case 1: //小不开对子
		r, r2 = GetDiceXiao()
		break
	case 2: //大不开对子
		r, r2 = GetDiceBig()
		break
	case 3: //小开对子
		r, r2 = GetDiceXiaoPairs()
		break
	case 4: //大开对子
		r, r2 = GetDiceBigPairs()
		break
	case 5: //小随机开
		r, r2 = GetDiceXiaoRand()
		break
	case 6: //大随机开
		r, r2 = GetDiceBigRand()
		break
	}
	//log.Tracef("获取结果：%v,%v", r, r2)
	return r, r2
}

func GetDiceHe() (int, int) {
	r := int(rand.Int31n(6) + 1)

	return r, 7 - r
}

//取得小不开对子
func GetDiceXiao() (int, int) {
	r := 0
	r2 := 0
	for {
		r = int(rand.Int31n(5) + 1)
		r2 = int(rand.Int31n(5) + 1)
		if r2 != r && (r+r2) < 7 {
			break
		}
	}

	return r, r2
}

//取得大不开对子
func GetDiceBig() (int, int) {
	r := 0
	r2 := 0
	for {
		r = int(rand.Int31n(6) + 1)
		r2 = int(rand.Int31n(6) + 1)
		if r2 != r && (r+r2) > 7 {
			break
		}
	}

	return r, r2
}

//取得小开对子
func GetDiceXiaoPairs() (int, int) {
	r := int(rand.Int31n(3) + 1)

	return r, r
}

//取得大开对子
func GetDiceBigPairs() (int, int) {
	r := int(rand.Int31n(3) + 4)

	return r, r
}

//取得小随机开
func GetDiceXiaoRand() (int, int) {
	r := 0
	r2 := 0

	for {
		r = int(rand.Int31n(5) + 1)
		r2 = int(rand.Int31n(5) + 1)
		if (r + r2) < 7 {
			break
		}
	}

	return r, r2
}

//取得大随机开
func GetDiceBigRand() (int, int) {
	r := 0
	r2 := 0
	for {
		r = int(rand.Int31n(6) + 1)
		r2 = int(rand.Int31n(6) + 1)
		if (r + r2) > 7 {
			break
		}
	}

	return r, r2
}
