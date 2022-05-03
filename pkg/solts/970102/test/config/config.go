package config

import (
	"io/ioutil"
	"strconv"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/bitly/go-simplejson"
)

var (
	PokerConf = simplejson.New()
	PoolConf  = simplejson.New()
)

func Load() {
	f, err := ioutil.ReadFile("config/paixing.json")
	if err == nil {
		PokerConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("PokerConf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/xuechi.json")
	if err == nil {
		PoolConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("PoolConf err", err)
			return
		}
	}
}

func GetBet(cfg string) int32 {
	j, err := simplejson.NewJson([]byte(cfg))
	if err != nil {
		log.Errorf("bet err", err)
		return 0
	}
	bet, _ := j.Get("Bottom_Pouring").Int()
	//bet, err := PaoDanConf.Get(strconv.Itoa(int(islandId))).Get(GetBulletId(lv)).Get("dizhu").Int()
	if err != nil {
		log.Errorf("bet err ", err)
	}
	return int32(bet)
}

func GetValue(key, key2 string) int {
	value, err := PokerConf.Get(key).Get(key2).Int()
	if err != nil {
		log.Errorf("value err", err)
	}
	return value
}

func GetPokerPay(pokerType, key int) int32 {
	multiple, err := PokerConf.Get(strconv.Itoa(key)).Get(strconv.Itoa(pokerType)).Int()
	if err != nil {
		log.Errorf("multiple err", err)
		log.Errorf("", pokerType, key)
	}
	return int32(multiple)
}

func GetOperationTime() int {
	t, err := PokerConf.Get("operationTime").Int()
	if err != nil {
		log.Errorf("operationTime err", err)
	}
	return t
}

func GetXueChiChance(key int32) int32 {
	if key == -1 {
		return 0
	}
	if key == 0 {
		key = 1000
	}
	chance, err := PoolConf.Get("1").Get(strconv.Itoa(int(key))).Int()
	if err != nil {
		log.Errorf("xuechi chance err", err)
	}
	return int32(chance)
}

func GetPlayerChance(key int32) int32 {
	if key == -1 {
		return 0
	}
	if key == 0 {
		key = 1000
	}
	chance, err := PoolConf.Get("2").Get(strconv.Itoa(int(key))).Int()
	if err != nil {
		log.Errorf("player chance err", err)
	}
	return int32(chance)
}
