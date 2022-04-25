package config

import (
	"common/log"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"

	"github.com/bitly/go-simplejson"
)

var (
	RoomConf = simplejson.New()
	PoolConf = simplejson.New()
	RunConf  = simplejson.New()
)

func LoadJsonConfig(_filename string, _config interface{}) (err error) {
	f, err := os.Open(_filename)
	if err == nil {
		defer f.Close()
		var fileInfo os.FileInfo
		fileInfo, err = f.Stat()
		if err == nil {
			bytes := make([]byte, fileInfo.Size())
			_, err = f.Read(bytes)
			if err == nil {
				BOM := []byte{0xEF, 0xBB, 0xBF}

				if bytes[0] == BOM[0] && bytes[1] == BOM[1] && bytes[2] == BOM[2] {
					bytes = bytes[3:]
				}
				err = json.Unmarshal(bytes, _config)
			}
		}
	}
	return
}

func Load() { //(/\*([^*]|[\r\n]|(\*+([^*/]|[\r\n])))*\*+/|[ \t]*//.*)  去掉注释的正则
	f, err := ioutil.ReadFile("config/room.json")
	if err == nil {
		RoomConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("room conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/xuechi.json")
	if err == nil {
		PoolConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("PoolConf conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/run.json")
	if err == nil {
		RunConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("RunConf conf err", err)
			return
		}
	}
}

//下注列表
func GetBetList(level int) []int64 {
	t, err := RoomConf.Get(strconv.Itoa(level)).Get("BetList").Array()
	if err != nil {
		log.Errorf("gameStartTime err", err)
	}
	betList := make([]int64, 0)
	for _, v := range t {
		bet, _ := v.(json.Number).Int64()
		betList = append(betList, bet)
	}
	log.Tracef("betlist", betList)
	return betList
}

//开始动画时间
func GetGameStartTime() int {
	t, err := RoomConf.Get("gameStartTime").Int()
	if err != nil {
		log.Errorf("gameStartTime err", err)
	}
	return t
}

//下注时间
func GetBetTime() int {
	t, err := RoomConf.Get("betTime").Int()
	if err != nil {
		log.Errorf("betTime err", err)
	}
	return t
}

//下注通知时间
func GetBetNoticeTime() int {
	t, err := RoomConf.Get("betNoticeTime").Int()
	if err != nil {
		log.Errorf("betNoticeTime err", err)
	}
	return t
}

//下注结束时间
func GetBetEndTime() int {
	t, err := RoomConf.Get("betEndTime").Int()
	if err != nil {
		log.Errorf("betEndTime err", err)
	}
	return t
}

//游戏时间
func GetGameTime() int {
	t, err := RoomConf.Get("gameTime").Int()
	if err != nil {
		log.Errorf("gameTime err", err)
	}
	return t
}

//结算时间
func GetCountTime() int {
	t, err := RoomConf.Get("countTime").Int()
	if err != nil {
		log.Errorf("countTime err", err)
	}
	return t
}

//冠军女赔付
func GetFirstFemalePay() float64 {
	t, err := RoomConf.Get("firstFemalePay").Float64()
	if err != nil {
		log.Errorf("firstFemalePay err", err)
	}
	return t
}

//冠军男赔付
func GetFirstMalePay() float64 {
	t, err := RoomConf.Get("firstMalePay").Float64()
	if err != nil {
		log.Errorf("firstMalePay err", err)
	}
	return t
}

//冠军赔付
func GetFirstPay() int {
	t, err := RoomConf.Get("firstPay").Int()
	if err != nil {
		log.Errorf("firstPay err", err)
	}
	return t
}

//一二名赔付
func GetFAndSPay() int {
	t, err := RoomConf.Get("fAndSPay").Int()
	if err != nil {
		log.Errorf("fAndSPay err", err)
	}
	return t
}

//下注上限
func GetBetLimit(level int) int {
	t, err := RoomConf.Get(strconv.Itoa(level)).Get("PersonalBetToplimit").Int()
	if err != nil {
		log.Errorf("PersonalBetToplimit err", err)
	}
	return t
}

//所有下注上限
func GetAllBetLimit(level int) int {
	t, err := RoomConf.Get(strconv.Itoa(level)).Get("AllPersonalBetToplimit").Int()
	if err != nil {
		log.Errorf("AllPersonalBetToplimit err", err)
	}
	return t
}

//机器人金币上限
func GetRobotMaxCoin(level int) int {
	t, err := RoomConf.Get(strconv.Itoa(level)).Get("RobotMaxCoin").Int()
	if err != nil {
		log.Errorf("RobotMaxCoin err", err)
	}
	return t
}

//机器人金币下限
func GetRobotMinCoin(level int) int {
	t, err := RoomConf.Get(strconv.Itoa(level)).Get("RobotMinCoin").Int()
	if err != nil {
		log.Errorf("RobotMinCoin err", err)
	}
	return t
}

//获取血池概率
func GetXueChiChance(key int32) (int64, int64) {
	if key == -1 {
		return 0, 0
	}
	if key == 0 {
		key = 1000
	}
	r := rand.Intn(10000)
	chanceArray, err := PoolConf.Get("1").Get(strconv.Itoa(int(key))).Array()
	if err != nil {
		log.Errorf("xuechi chance err", err)
	}
	lossRationArray, err := PoolConf.Get("2").Array()
	if err != nil {
		log.Errorf("lossRation err", err)
	}
	chance := 0
	lossRation := 0
	min := 0
	for k, v := range chanceArray {
		c, _ := strconv.Atoi(v.(json.Number).String())
		chance += c
		temLossRation, _ := strconv.Atoi(lossRationArray[k].(json.Number).String())
		min = lossRation
		lossRation += temLossRation
		if r < chance {
			break
		}
	}
	return int64(min), int64(lossRation)
}

//获取时间列表
func GetTimeList() []interface{} {
	timeList := make([]interface{}, 0)
	runConfArray, err := RunConf.Array()
	if err != nil {
		log.Errorf("get run conf length err", err)
		return nil
	}
	length := len(runConfArray)
	r := rand.Intn(length)
	runConf := runConfArray[r].(map[string]interface{})
	timeList = append(timeList, runConf["first"])
	timeList = append(timeList, runConf["second"])
	other := runConf["other"].([]interface{})
	rand.Shuffle(len(other), func(i, j int) {
		other[i], other[j] = other[j], other[i]
	})
	timeList = append(timeList, other...)
	return timeList
}
