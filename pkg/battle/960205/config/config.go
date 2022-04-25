package config

import (
	"common/log"
	"io/ioutil"
	"os"

	"github.com/gogf/gf/encoding/gjson"
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
	b, e := ioutil.ReadFile("./config/erbagang.json")
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
	b, e = ioutil.ReadFile("./config/xuechi.json")
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
	return int64(json.GetInt("sleep_time.players_match_sleep_time"))
}

// ConfGameStartSleepTime 游戏开始延时
func ConfGameStartSleepTime() int64 {
	return int64(json.GetInt("sleep_time.game_start_sleep_time"))
}

// ConfRobZhuangSleepTime 抢庄动画时间
func ConfRobZhuangSleepTime() int64 {
	return int64(json.GetInt("sleep_time.rob_zhuang_sleep_time"))
}

// ConfBetZhuangSleepTime 下注动画时间
func ConfBetZhuangSleepTime() int64 {
	return int64(json.GetInt("sleep_time.bet_sleep_time"))
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
	return poolConf.GetJson("1").GetInt32(key)
}

func GetRobotChance(key string) int32 {
	if key == "0" || key == "-1" {
		key = "1000"
	}
	return poolConf.GetJson("3").GetInt32(key)
}

func GetPointChance(key string) int32 {
	if key == "0" || key == "-1" {
		key = "1000"
	}
	return poolConf.GetJson("2").GetInt32(key)
}
