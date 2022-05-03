package config

import (
	"io/ioutil"
	"os"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

var (
	json     *gjson.Json
	poolConf *gjson.Json
	err      error
)

func init() {
	configInit()
}

func configInit() {
	b, e := ioutil.ReadFile("./conf/erbagang.json")
	if e != nil {
		log.Errorf("%s:%s", "读取配置文件出错:", e.Error())
		// 配置文件读取出错就退出
		os.Exit(0)
	}
	json, err = gjson.DecodeToJson(b)
	if err != nil {
		log.Errorf("%s,%s", "解析配置文件出错", err.Error())
		os.Exit(0)
	}
	b, e = ioutil.ReadFile("./conf/xuechi.json")
	if e != nil {
		log.Errorf("%s:%s", "读取配置文件出错:", e.Error())
		// 配置文件读取出错就退出
		os.Exit(0)
	}
	poolConf, err = gjson.DecodeToJson(b)
	if err != nil {
		log.Errorf("%s,%s", "解析配置文件出错", err.Error())
		os.Exit(0)
	}
}

// ConfPlayersMatchSleepTime 玩家匹配动画时间
func ConfPlayersMatchSleepTime() int64 {
	return json.Get("sleep_time.players_match_sleep_time").Int64()
}

// ConfGameStartSleepTime 游戏开始延时
func ConfGameStartSleepTime() int64 {
	return json.Get("sleep_time.game_start_sleep_time").Int64()
}

// ConfRobZhuangSleepTime 抢庄动画时间
func ConfRobZhuangSleepTime() int64 {
	return json.Get("sleep_time.rob_zhuang_sleep_time").Int64()
}

// ConfBetZhuangSleepTime 下注动画时间
func ConfBetZhuangSleepTime() int64 {
	return json.Get("sleep_time.bet_sleep_time").Int64()
}

// ConfQiangZhuangMultiple 抢庄倍数配置
func ConfQiangZhuangMultiple() *gjson.Json {
	return json.GetJson("qiang_zhuang_multiple")
}

// ConfXiaZhu 下注
func ConfXiaZhu() *gjson.Json {
	return json.GetJson("xia_zhu")
}

func GetPoolChance(key string) int32 {
	if key == "0" || key == "-1" {
		key = "1000"
	}
	return poolConf.GetJson("1").Get(key).Int32()
}

func GetRobotChance(key string) int32 {
	if key == "0" || key == "-1" {
		key = "1000"
	}
	return poolConf.GetJson("3").Get(key).Int32()
}

func GetPointChance(key string) int32 {
	if key == "0" || key == "-1" {
		key = "1000"
	}
	return poolConf.GetJson("2").Get(key).Int32()
}
