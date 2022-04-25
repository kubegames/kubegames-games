package config

import (
	"common/log"
	"common/rand"
	"encoding/json"
	"game_buyu/renyuchuanshuo/msg"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/bitly/go-simplejson"
)

type config struct {
	ListenerPort int    //web socket 侦听端口
	MaxConn      int32  //最大连接数
	DBUri        string //数据库链接字符串
	LogPath      string //需要制定输出的日志路径
}

var Config = new(config)
var (
	YuConf        = simplejson.New()
	ChangJingConf = simplejson.New()
	PaoConf       = simplejson.New()
	PaoDanConf    = simplejson.New()
	LineConf      = simplejson.New()
	RobotConf     = simplejson.New()
	FormationConf = simplejson.New()
	PoolConf      = simplejson.New()
	Fishes        = make(map[msg.Type][]string, 0)
	MaxBulletLv   = int32(0)
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
	f, err := ioutil.ReadFile("config/yu.json")
	if err == nil {
		YuConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/shuaxin.json")
	if err == nil {
		shuaxin, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("", err)
			return
		}
		s, _ := shuaxin.Map()
		YuConf.Set("2", s)
	}
	f, err = ioutil.ReadFile("config/changjin.json")
	if err == nil {
		ChangJingConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/pao.json")
	if err == nil {
		PaoConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/paodan.json")
	if err == nil {
		PaoDanConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/jiqirAI.json")
	if err == nil {
		RobotConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("robot conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/yuzhen.json")
	if err == nil {
		FormationConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("FormationConf conf err", err)
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
	loadFishLine()
	loadFishType()
	loadBulletLv()
}

func loadFishLine() {
	f, err := ioutil.ReadFile("config/dayu01.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("dayu01", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("dayu01", err)
			return
		}
		LineConf.Set("dayu01", dayu02)
	}
	f, err = ioutil.ReadFile("config/BOSS01.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("BOSS01", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("BOSS01", err)
			return
		}
		LineConf.Set("BOSS01", dayu02)
	}
	f, err = ioutil.ReadFile("config/zy01.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zy01", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zy01", err)
			return
		}
		LineConf.Set("zy01", dayu02)
	}
	f, err = ioutil.ReadFile("config/zy02.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zy02", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zy02", err)
			return
		}
		LineConf.Set("zy02", dayu02)
	}
	f, err = ioutil.ReadFile("config/BOSS1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("BOSS1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("BOSS1", err)
			return
		}
		LineConf.Set("BOSS1", dayu02)
	}
	f, err = ioutil.ReadFile("config/34.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("34", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("34", err)
			return
		}
		LineConf.Set("34", dayu02)
	}
	f, err = ioutil.ReadFile("config/dy1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("dy1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("dy1", err)
			return
		}
		LineConf.Set("dy1", dayu02)
	}
	f, err = ioutil.ReadFile("config/zy1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zy1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zy1", err)
			return
		}
		LineConf.Set("zy1", dayu02)
	}
	f, err = ioutil.ReadFile("config/02.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("02", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("02", err)
			return
		}
		LineConf.Set("02", dayu02)
	}
	f, err = ioutil.ReadFile("config/04.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("04", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("04", err)
			return
		}
		LineConf.Set("04", dayu02)
	}
	f, err = ioutil.ReadFile("config/10.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("10", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("10", err)
			return
		}
		LineConf.Set("10", dayu02)
	}
	f, err = ioutil.ReadFile("config/11.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("11", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("11", err)
			return
		}
		LineConf.Set("11", dayu02)
	}
	f, err = ioutil.ReadFile("config/12.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("12", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("12", err)
			return
		}
		LineConf.Set("12", dayu02)
	}
	f, err = ioutil.ReadFile("config/13.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("13", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("13", err)
			return
		}
		LineConf.Set("13", dayu02)
	}
	f, err = ioutil.ReadFile("config/14.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("14", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("14", err)
			return
		}
		LineConf.Set("14", dayu02)
	}
	f, err = ioutil.ReadFile("config/hdc1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("hdc1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("hdc1", err)
			return
		}
		LineConf.Set("hdc1", dayu02)
	}
	f, err = ioutil.ReadFile("config/yx1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("yx1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("yx1", err)
			return
		}
		LineConf.Set("yx1", dayu02)
	}
	f, err = ioutil.ReadFile("config/HJSY1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("HJSY1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("HJSY1", err)
			return
		}
		LineConf.Set("HJSY1", dayu02)
	}
	f, err = ioutil.ReadFile("config/hjwg1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("hjwg1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("hjwg1", err)
			return
		}
		LineConf.Set("hjwg1", dayu02)
	}
	f, err = ioutil.ReadFile("config/hjwg2.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("hjwg2", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("hjwg2", err)
			return
		}
		LineConf.Set("hjwg2", dayu02)
	}
	f, err = ioutil.ReadFile("config/hjwg3.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("hjwg3", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("hjwg3", err)
			return
		}
		LineConf.Set("hjwg3", dayu02)
	}
	f, err = ioutil.ReadFile("config/JJXJ1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("JJXJ1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("JJXJ1", err)
			return
		}
		LineConf.Set("JJXJ1", dayu02)
	}
	f, err = ioutil.ReadFile("config/jly2.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("jly2", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("jly2", err)
			return
		}
		LineConf.Set("jly2", dayu02)
	}
	f, err = ioutil.ReadFile("config/jly3.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("jly3", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("jly3", err)
			return
		}
		LineConf.Set("jly3", dayu02)
	}
	f, err = ioutil.ReadFile("config/jly4.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("jly4", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("jly4", err)
			return
		}
		LineConf.Set("jly4", dayu02)
	}
	f, err = ioutil.ReadFile("config/yq1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("yq1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("yq1", err)
			return
		}
		LineConf.Set("yq1", dayu02)
	}
	f, err = ioutil.ReadFile("config/xy1.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("xy1", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("xy1", err)
			return
		}
		LineConf.Set("xy1", dayu02)
	}
	f, err = ioutil.ReadFile("config/20.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("20", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("20", err)
			return
		}
		LineConf.Set("20", dayu02)
	}
	f, err = ioutil.ReadFile("config/21.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("21", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("21", err)
			return
		}
		LineConf.Set("21", dayu02)
	}
	f, err = ioutil.ReadFile("config/22.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("22", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("22", err)
			return
		}
		LineConf.Set("22", dayu02)
	}
	f, err = ioutil.ReadFile("config/23.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("23", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("23", err)
			return
		}
		LineConf.Set("23", dayu02)
	}
	f, err = ioutil.ReadFile("config/24.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("24", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("24", err)
			return
		}
		LineConf.Set("24", dayu02)
	}
	f, err = ioutil.ReadFile("config/25.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("25", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("25", err)
			return
		}
		LineConf.Set("25", dayu02)
	}
}

func loadFishType() {
	Fishes[msg.Type_SMALL] = make([]string, 0)
	Fishes[msg.Type_MIDDLE] = make([]string, 0)
	Fishes[msg.Type_BIG] = make([]string, 0)
	Fishes[msg.Type_BOSS] = make([]string, 0)
	Fishes[msg.Type_PILE] = make([]string, 0)
	Fishes[msg.Type_KING] = make([]string, 0)
	all, err := YuConf.Get("1").Map()
	keys := make([]string, 0)
	if err != nil {
		log.Errorf("", err)
		return
	}
	for k, _ := range all {
		keys = append(keys, k)
	}
	for _, k := range keys {
		v := YuConf.Get("1").Get(k)
		tiji, err := v.Get("leixing").Int()
		if err != nil {
			log.Errorf("leixing err", err)
			return
		}
		id, err := v.Get("Id").String()
		if err != nil {
			log.Errorf("Id err", err)
			return
		}
		Fishes[msg.Type(int32(tiji))] = append(Fishes[msg.Type(int32(tiji))], id)
	}
}

func loadBulletLv() {
	all, err := PaoConf.Get("1").Map()
	if err != nil {
		log.Errorf("pao err", err)
		return
	}
	MaxBulletLv = int32(len(all))
}

func GetFishWeight() int {
	length, err := YuConf.Get("maxLength").Int()
	if err != nil {
		log.Errorf("maxLength err", err)
	}
	return length
}

func GetFishTideTime() int {
	length, err := YuConf.Get("fishTideTime").Int()
	if err != nil {
		log.Errorf("fishTideTime err", err)
	}
	return length
}

func GetFishTideForecastTime() int {
	length, err := YuConf.Get("fishTideForecastTime").Int()
	if err != nil {
		log.Errorf("fishTideForecastTime err", err)
	}
	return length
}

func GetFishByType(t msg.Type) string {
	length := len(Fishes[t])
	if length == 0 {
		return ""
	} //Fishes[1][0]
	r := rand.RandInt(0, length)
	return Fishes[t][r]
}

func GetFish(fishId string) *simplejson.Json {
	return YuConf.Get("1").Get(fishId)
}

func GetFishSpeed(fishId string) int32 {
	speed, err := GetFish(fishId).Get("sudu").Int()
	if err != nil {
		log.Errorf("speed err ", err)
	}
	return int32(speed)
}

func GetFishNum(fishId string) int32 {
	speed, err := GetFish(fishId).Get("yuqun").Int()
	if err != nil {
		log.Errorf("Num err ", err)
	}
	return int32(speed)
}

func GetFishName(fishId string) string {
	speed, err := GetFish(fishId).Get("yuming").String()
	if err != nil {
		log.Errorf("fishName err ", err)
	}
	return speed
}

func GetFishOffset(fishId string) []interface{} {
	speed, err := GetFish(fishId).Get("yqpy").Array()
	if err != nil {
		log.Errorf("offset err ", err)
	}
	return speed
}

func GetFishScore(fishId string) int32 {
	speed, err := GetFish(fishId).Get("fenzhi").Int()
	if err != nil {
		log.Errorf("Score err ", err)
	}
	return int32(speed)
}

func GetFishChance(fishId string) int32 {
	speed, err := GetFish(fishId).Get("xittonggailv").Int()
	if err != nil {
		log.Errorf("Chance err ", err)
	}
	return int32(speed)
}

func GetFishTideSustainTime(key string) int {
	speed, err := YuConf.Get("yuchaochixushijian").Map()
	if err != nil {
		log.Errorf("FishTideSustainTime err ", err)
	}
	t, _ := strconv.Atoi(speed[key].(json.Number).String())
	return t
}

func GetFishLineId(fishId string) int32 {
	speed, err := GetFish(fishId).Get("luxianID").Int()
	if err != nil {
		log.Errorf("LineId err ", err)
	}
	return int32(speed)
}

func GetFishHitChance(fishId string, isRobot bool) int32 {
	if isRobot {
		speed, err := GetFish(fishId).Get("jiqigailv").Int()
		if err != nil {
			log.Errorf("robot HitChance err ", err)
		}
		return int32(speed)
	}
	speed, err := GetFish(fishId).Get("gailv").Int()
	if err != nil {
		log.Errorf("HitChance err ", err)
	}
	return int32(speed)
}

func GetBulletId(lv int32) string {
	speed, err := PaoConf.Get("1").Get(strconv.Itoa(int(lv))).Get("BulletId").String()
	if err != nil {
		log.Errorf("BulletId err ", err)
	}
	return speed
}

func GetFishBet(cfg string, lv int32) int32 {
	//bet, err := PaoDanConf.Get(islandId).Get(GetBulletId(lv)).Get("dizhu").Int()
	j, err := simplejson.NewJson([]byte(cfg))
	if err != nil {
		log.Errorf("bet err", err)
		return 0
	}
	bet, _ := j.Get("Bottom_Pouring").Int()
	if err != nil {
		log.Errorf("bet err ", err)
	}
	return int32(bet) * lv
}

func GetSceneFishCap(islandId, key string) map[msg.Type]int {
	fishCap := make(map[msg.Type]int, 0)
	scene := ChangJingConf.Get(islandId)
	small, err := scene.Get(key).Get("xiaoyu").Int()
	if err != nil {
		log.Errorf("small cap nil", err)
	}
	middle, err := scene.Get(key).Get("zhongyu").Int()
	if err != nil {
		log.Errorf("middle cap nil", err)
	}
	big, err := scene.Get(key).Get("dayu").Int()
	if err != nil {
		log.Errorf("big cap nil", err)
	}
	boss, err := scene.Get(key).Get("teshuyu").Int()
	if err != nil {
		log.Errorf("boss cap nil", err)
	}
	pile, err := scene.Get(key).Get("ydxz").Int()
	if err != nil {
		log.Errorf("pile cap nil", err)
	}
	king, err := scene.Get(key).Get("ywxz").Int()
	if err != nil {
		log.Errorf("king cap nil", err)
	}
	fishCap[msg.Type_SMALL] = small
	fishCap[msg.Type_MIDDLE] = middle
	fishCap[msg.Type_BIG] = big
	fishCap[msg.Type_BOSS] = boss
	fishCap[msg.Type_PILE] = pile
	fishCap[msg.Type_KING] = king
	return fishCap
}

func GetSceneNum(key string) int {
	all, err := ChangJingConf.Get(key).Map()
	if err != nil {
		log.Errorf("scene len err", err)
	}
	return len(all) - 1
}

func GetShotoffTime() int {
	t, err := ChangJingConf.Get("shotOffTime").Int()
	if err != nil {
		log.Errorf("ShotoffTime err", err)
	}
	return t
}

func GetFishType(fishId string) msg.Type {
	fishType, err := GetFish(fishId).Get("leixing").Int()
	if err != nil {
		log.Errorf("get fishtype err", err)
	}
	return msg.Type(fishType)
}

func GetSkillHitNum(fishId string, skillId int32) int {
	key := ""
	if skillId == 1 {
		key = "leidiangongji"
	}
	if skillId == 2 {
		key = "baozhagongji"
	}
	if skillId == 3 {
		key = "bingdongshijian"
	}
	if skillId == 4 {
		key = "wangzhagongji"
		return 100
	}
	num, err := GetFish(fishId).Get(key).Int()
	if err != nil {
		log.Errorf("skill hit num err", err)
	}
	return num
}

func GetSkillFishNum(fishId string, skillId int32) int {
	key := ""
	if skillId == 1 {
		key = "ldsl"
	}
	if skillId == 2 {
		key = "bzsl"
	}
	num, err := GetFish(fishId).Get(key).Int()
	if err != nil {
		log.Errorf("skill hit fish num err", err)
	}
	return num
}

func GetSkillId(fishId string) int32 {
	fish := GetFish(fishId)
	num, err := fish.Get("leidiangongji").Int()
	if err != nil {
		log.Errorf("skillId err leidiangongji", err)
	}
	if num > 0 {
		return 1
	}
	num, err = fish.Get("baozhagongji").Int()
	if err != nil {
		log.Errorf("skillId err baozhagongji", err)
	}
	if num > 0 {
		return 2
	}
	num, err = fish.Get("bingdongshijian").Int()
	if err != nil {
		log.Errorf("skillId err bingdongshijian", err)
	}
	if num > 0 {
		return 3
	}
	num, err = fish.Get("wangzhagongji").Int()
	if err != nil {
		log.Errorf("skillId err wangzhagongji", err)
	}
	if num > 0 {
		return 4
	}
	return 0
}

func GetConfLine(fishId string, rr func(limit int, i int64) int, seed int64) ([]interface{}, []interface{}, interface{}, interface{}, []interface{}) {
	fish := GetFish(fishId)
	lineKeys, err := fish.Get("luxianID").Array()
	if err != nil {
		log.Errorf("line key err", err)
		return nil, nil, nil, nil, nil
	}
	//key := lineKeys[rand.RandInt(0, len(lineKeys))]
	key := lineKeys[rr(len(lineKeys), seed)]
	lines, err := LineConf.Get(key.(string)).Array() //.Get("游动路线")
	if err != nil {                                  //key.(string)
		return nil, nil, nil, nil, nil
	}

	//r := rand.RandInt(0, len(lines))
	r := rr(len(lines), seed)
	name := lines[r].(map[string]interface{})["路线名称"]
	line := lines[r].(map[string]interface{})["游动路线"].([]interface{})
	speed := lines[r].(map[string]interface{})["游动速度"].([]interface{})
	t := lines[r].(map[string]interface{})["路线总速度"]
	variant := lines[r].(map[string]interface{})["加速下标"].([]interface{})
	return line, speed, t, name, variant
}

func GetRefreshTime() int {
	t, err := ChangJingConf.Get("shuayusudu").Int()
	if err != nil {
		log.Errorf("refresh time err ", err)
	}
	return t
}

func GetRoundTime() int {
	t, err := ChangJingConf.Get("roundTime").Int()
	if err != nil {
		log.Errorf("round time err ", err)
	}
	return t
}

func GetSaveScoreTime() int {
	t, err := ChangJingConf.Get("saveScoreTime").Int()
	if err != nil {
		log.Errorf("saveScoreTime time err ", err)
	}
	return t
}

func GetRobotInitShootNum() int {
	num, err := RobotConf.Get("1").Get("1").Get("csdj").Int()
	if err != nil {
		log.Errorf("robot init shoot num err", err)
	}
	return num
}

func GetRobotIsFixed() int32 {
	num, err := RobotConf.Get("1").Get("1").Get("gdgd").Int()
	if err != nil {
		log.Errorf("robot isfixed err", err)
	}
	return int32(num)
}

func GetRobotLimit() int32 {
	num, err := RobotConf.Get("1").Get("1").Get("limit").Int()
	if err != nil {
		log.Errorf("robot robot limit err", err)
	}
	return int32(num)
}

func GetRobotChangeTime() int32 {
	num, err := RobotConf.Get("1").Get("1").Get("ptjd1").Int()
	if err != nil {
		log.Errorf("robot shoot time err", err)
	}
	return int32(num)
}

func GetRobotRandChangeTime() int32 {
	num, err := RobotConf.Get("1").Get("1").Get("ptjd2").Int()
	if err != nil {
		log.Errorf("robot rand shoot time err", err)
	}
	return int32(num)
}

func GetRobotWinChangeCoin() int64 {
	num, err := RobotConf.Get("1").Get("1").Get("ptqh1").Int()
	if err != nil {
		log.Errorf("robot win coin change err", err)
	}
	return int64(num)
}

func GetRobotWinChangeLimit() []interface{} {
	num, err := RobotConf.Get("1").Get("1").Get("qhz1").Array()
	if err != nil {
		log.Errorf("robot win limit err", err)
	}
	return num
}

func GetRobotLoseChangeCoin() int64 {
	num, err := RobotConf.Get("1").Get("1").Get("ptqh2").Int()
	if err != nil {
		log.Errorf("robot lose change coin err", err)
	}
	return int64(num)
}

func GetRobotLoseChangeLimit() []interface{} {
	num, err := RobotConf.Get("1").Get("1").Get("qhz2").Array()
	if err != nil {
		log.Errorf("robot lose limit err", err)
	}
	return num
}

func GetRobotLockTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("jdjg").Int()
	if err != nil {
		log.Errorf("robot lock time err", err)
	}
	return num
}

func GetRobotShootLimit() int {
	num, err := RobotConf.Get("1").Get("1").Get("shootLimit").Int()
	if err != nil {
		log.Errorf("robot shoot limit err", err)
	}
	return num
}

func GetRobotRestSpace() int {
	num, err := RobotConf.Get("1").Get("1").Get("jgjd").Array()
	if err != nil {
		log.Errorf("robot rest space err", err)
	}
	min, _ := strconv.Atoi(num[0].(json.Number).String())
	max, _ := strconv.Atoi(num[1].(json.Number).String())
	return rand.RandInt(min, max)
}

func GetRobotRestTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("xxsj").Array()
	if err != nil {
		log.Errorf("robot rest time err", err)
	}
	min := num[0].(int)
	max := num[1].(int)
	return rand.RandInt(min, max)
}

func GetRobotQuitTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("tcsj").Int()
	if err != nil {
		log.Errorf("robot quit time err", err)
	}
	return num
}

func GetRobotQuitCoin() int {
	num, err := RobotConf.Get("1").Get("1").Get("ydbl").Int()
	if err != nil {
		log.Errorf("robot quit coin err", err)
	}
	return num
}

func GetRobotLockChance() int {
	num, err := RobotConf.Get("1").Get("1").Get("tsgl").Int()
	if err != nil {
		log.Errorf("robot lock chance err", err)
	}
	return num
}

func GetRobotLockFishes(fishId string) bool {
	num, err := RobotConf.Get("1").Get("1").Get("jdyID1").Array()
	if err != nil {
		log.Errorf("robot check lock fish err", err)
	}
	for _, fish := range num {
		if fish.(string) == fishId {
			return true
		}
	}
	return false
}

func GetFormation(key string) map[string]interface{} {
	formation, err := FormationConf.Get(key).Map()
	if err != nil {
		log.Errorf("formation err", err)
	}
	return formation
}

func GetAFormationKey() string {
	formation, err := FormationConf.Map()
	if err != nil {
		log.Errorf("formations err", err)
	}
	keys := make([]string, 0)
	for k, _ := range formation {
		keys = append(keys, k)
	}
	return keys[rand.RandInt(0, len(keys))]
}

func GetFormationFishInfo(formationKey, key string) (string, int32, int, int, []interface{}) {
	info := GetFormation(formationKey)[key].(map[string]interface{})
	if info != nil {
		id := info["id"].(json.Number).String()
		speed, _ := strconv.Atoi(info["sd"].(json.Number).String())
		line, _ := info["zb"].([]interface{})
		t, _ := strconv.Atoi(info["scjg"].(json.Number).String())
		num, _ := strconv.Atoi(info["scbs"].(json.Number).String())
		return id, int32(speed), t, num, line
	}
	return "", 0, 0, 0, nil
}

func GetCircleFormationFishInfo(formationKey, key string) (string, int32, int, int, []interface{}, float64, float64, float64, int32) {
	info := GetFormation(formationKey)[key].(map[string]interface{})
	if info != nil {
		id := info["id"].(json.Number).String()
		speed, _ := strconv.Atoi(info["sd"].(json.Number).String())
		line, _ := info["scd"].([]interface{})
		radius, _ := strconv.ParseFloat(info["bjz"].(json.Number).String(), 32)
		overlying, _ := strconv.ParseFloat(info["djz"].(json.Number).String(), 32)
		angle, _ := strconv.ParseFloat(info["jdz"].(json.Number).String(), 32)
		t, _ := strconv.Atoi(info["scjg"].(json.Number).String())
		num, _ := strconv.Atoi(info["scbs"].(json.Number).String())
		time, _ := strconv.Atoi(info["lxzsd"].(json.Number).String())
		return id, int32(speed), t, num, line, radius, overlying, angle, int32(time)
	}
	return "", 0, 0, 0, nil, 0, 0, 0, 0
}

func GetTimeFish() map[string]interface{} {
	timeFish, err := YuConf.Get("2").Map()
	if err != nil {
		log.Errorf("all timeFish err", err)
	}
	t := make(map[string]interface{}, 0)
	index := 0 //rand.RandInt(0, len(timeFish))
	i := 0
	for _, v := range timeFish {
		if i == index {
			t = v.(map[string]interface{})
			break
		}
		i++
	}
	return t
}

func GetTimeFishInfo(key interface{}, r func(limit int, i int64) int, seed int64) (string, int, int, int) {
	info := key.(map[string]interface{})
	if info != nil {
		allId := info["Id"].([]interface{})
		id := allId[0].(string)
		if len(allId) > 1 {
			//id = allId[rand.RandInt(0, len(allId))].(string)
			id = allId[r(len(allId), seed)].(string)
		}
		//start, _ := strconv.Atoi(info["sxsj"].(json.Number).String())sxjg
		space, _ := strconv.Atoi(info["sxsj"].(json.Number).String())
		num, _ := strconv.Atoi(info["sxsl"].(json.Number).String())
		return id, 0, space, num
	}
	return "", 0, 0, 0
}

func GetXueChiChance(fishId string, key int32) int32 {
	if key == -1 {
		return 0
	}
	if key == 0 {
		key = 1000
	}
	chance, err := PoolConf.Get("1").Get(strconv.Itoa(int(key))).Get(fishId).Int()
	if err != nil {
		log.Errorf("xuechi chance err", err)
	}
	return int32(chance)
}

func GetPool() {

}
