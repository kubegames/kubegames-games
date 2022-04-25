package config

import (
	"encoding/json"
	"go-game-sdk/example/game_buyu/980301/msg"
	"go-game-sdk/example/game_buyu/980301/tools"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"

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
	SkillConf     = simplejson.New()
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
			log.Errorf("yu conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/shuaxin.json")
	if err == nil {
		shuaxin, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("shuaxin conf err", err)
			return
		}
		s, _ := shuaxin.Map()
		YuConf.Set("2", s)
	}
	f, err = ioutil.ReadFile("config/changjin.json")
	if err == nil {
		ChangJingConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("changjin conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/pao.json")
	if err == nil {
		PaoConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("pao conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/paodan.json")
	if err == nil {
		PaoDanConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("paodan conf err", err)
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
		formation, err := simplejson.NewJson(f)
		f, err := formation.Map()
		if err != nil {
			log.Errorf("FormationConf conf err", err)
			return
		}
		FormationConf.Set("yuzhen", f)
	}
	f, err = ioutil.ReadFile("config/yuzhen2.json")
	if err == nil {
		formation, err := simplejson.NewJson(f)
		f, err := formation.Map()
		if err != nil {
			log.Errorf("FormationConf conf err", err)
			return
		}
		FormationConf.Set("yuzhen2", f)
	}
	f, err = ioutil.ReadFile("config/xuechi.json")
	if err == nil {
		PoolConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("PoolConf conf err", err)
			return
		}
	}
	f, err = ioutil.ReadFile("config/jineng.json")
	if err == nil {
		SkillConf, err = simplejson.NewJson(f)
		if err != nil {
			log.Errorf("SkillConf conf err", err)
			return
		}
	}
	loadFishLine()
	loadFishType()
	loadBulletLv()
}

func loadFishLine() {
	f, err := ioutil.ReadFile("config/quxian.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("quxian", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("quxian", err)
			return
		}
		LineConf.Set("quxian", dayu02)
	}
	f, err = ioutil.ReadFile("config/quxian2.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("quxian2", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("quxian2", err)
			return
		}
		LineConf.Set("quxian2", dayu02)
	}
	f, err = ioutil.ReadFile("config/zhixian(1).json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zhixian(1)", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zhixian(1)", err)
			return
		}
		LineConf.Set("zhixian(1)", dayu02)
	}
	f, err = ioutil.ReadFile("config/zhixian(2).json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zhixian(2)", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zhixian(2)", err)
			return
		}
		LineConf.Set("zhixian(2)", dayu02)
	}
	f, err = ioutil.ReadFile("config/zhaohuan.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("zhaohuan", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("zhaohuan", err)
			return
		}
		LineConf.Set("zhaohuan", dayu02)
	}
	f, err = ioutil.ReadFile("config/boss33luxian.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("boss33luxian", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("boss33luxian", err)
			return
		}
		LineConf.Set("boss33luxian", dayu02)
	}
	f, err = ioutil.ReadFile("config/long.json")
	if err == nil {
		//LineConf, err = simplejson.NewJson(f)
		dayu01, err := simplejson.NewJson(f)
		if err != nil {
			log.Errorf("long", err)
			return
		}
		dayu02, err := dayu01.Array()
		if err != nil {
			log.Errorf("long", err)
			return
		}
		LineConf.Set("long", dayu02)
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
	length, err := YuConf.Get("2").Get("3").Get("chsj").Int()
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
	r := tools.RandInt(0, length, 0)
	return Fishes[t][r]
}

func GetFishByTypeFromFormation(file, key string, t msg.Type) string {
	f := ""
	if t == msg.Type_SMALL {
		f = "idxiaoyu"
	}
	if t == msg.Type_MIDDLE {
		f = "idzhongyu"
	}
	if t == msg.Type_BIG {
		f = "iddayu"
	}
	if f == "" {
		return ""
	}
	all := GetFormation(file, key)[f].([]interface{})
	length := len(all)
	if length == 0 {
		return ""
	} //Fishes[1][0]
	r := tools.RandInt(0, length, 0)
	return all[r].(string)
}

func GetFish(fishId string) *simplejson.Json {
	return YuConf.Get("1").Get(fishId)
}

func GetFishSpeed(fishId string) int32 {
	speed, err := GetFish(fishId).Get("sudu").Int()
	if err != nil {
		log.Errorf("speed err ", err)
		speed = 60
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

func GetFishNumChance(fishId string) int32 {
	speed, err := GetFish(fishId).Get("yqgailv").Int()
	if err != nil {
		log.Errorf("Num chance err ", err)
	}
	return int32(speed)
}

func GetFishTime(fishId string) int {
	speed, err := GetFish(fishId).Get("yqshijian").Int()
	if err != nil {
		log.Errorf("yuqun time err ", err)
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
	speed, err := GetFish(fishId).Get("pzluxiangailv").Int()
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
	//if isRobot {
	//	speed, err := GetFish(fishId).Get("jiqigailv").Int()
	//	if err != nil {
	//		log.Errorf("robot HitChance err ", err)
	//	}
	//	return int32(speed)
	//}
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
	return int32(bet) * lv
}

func GetSceneFishCap(islandId, key string) map[msg.Type]int {
	fishCap := make(map[msg.Type]int, 0)
	scene := ChangJingConf.Get(islandId)
	all, err := scene.Get(key).Get("zuidabeishu").Int()
	if err != nil {
		log.Errorf("zuidabeishu nil", err)
	}
	up, err := scene.Get(key).Get("publ").Int()
	if err != nil {
		log.Errorf("publ nil", err)
	}
	all = all + all*up/100
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
	fishCap[msg.Type_SMALL] = all * small / 100
	fishCap[msg.Type_MIDDLE] = all * middle / 100
	fishCap[msg.Type_BIG] = all * big / 100
	return fishCap
}

func GetSceneBossFishCap(islandId, key string) map[msg.Type]int {
	fishCap := make(map[msg.Type]int, 0)
	scene := ChangJingConf.Get(islandId)
	all, err := scene.Get(key).Get("bosszuidabeishu").Int()
	if err != nil {
		log.Errorf("bosszuidabeishu nil", err)
	}
	up, err := scene.Get(key).Get("bosspubl").Int()
	if err != nil {
		log.Errorf("bosspubl nil", err)
	}
	all = all + all*up/100
	small, err := scene.Get(key).Get("bossxiaoyu").Int()
	if err != nil {
		log.Errorf("boss small cap nil", err)
	}
	middle, err := scene.Get(key).Get("bosszhongyu").Int()
	if err != nil {
		log.Errorf("boss middle cap nil", err)
	}
	big, err := scene.Get(key).Get("bossdayu").Int()
	if err != nil {
		log.Errorf("boss big cap nil", err)
	}
	fishCap[msg.Type_SPECIAL] = int(GetFishScore(GetBossId(islandId, key)))
	fishCap[msg.Type_SMALL] = all * small / 100
	fishCap[msg.Type_MIDDLE] = all * middle / 100
	fishCap[msg.Type_BIG] = all * big / 100
	return fishCap
}

func GetBossId(islandId, sceneId string) string {
	id, err := ChangJingConf.Get(islandId).Get(sceneId).Get("bossID").String()
	if err != nil {
		log.Errorf("bossId err", err)
	}
	return id
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

func GetSkillHitNum(fishId string, skillId int32, duration int64, mult int32) int {
	//key := ""
	//if skillId == 1 {
	//	key = "leidiangongji"
	//}
	//if skillId == 2 {
	//	key = "baozhagongji"
	//}
	//if skillId == 3 {
	//	key = "bingdongshijian"
	//}
	if skillId == 4 {
		return 1000
	}
	num, err := GetFish(fishId).Get("gjcishu").Int()
	if err != nil {
		log.Errorf("skill hit num err", err)
	}
	return num
}

func GetSkillDur(fishId string, skillId int32) int64 {
	num, err := GetFish(fishId).Get("gjshijian").Int()
	if err != nil {
		log.Errorf("skill time err", err)
	}
	return int64(num)
}

func GetSkills() []*msg.Skill {
	skills := make([]*msg.Skill, 0)
	s, err := SkillConf.Map()
	if err != nil {
		log.Errorf("SkillConf err", err)
	}
	for k, _ := range s {
		skill := GetSkillinfo(k)
		skills = append(skills, skill)
	}
	return skills
}

func GetSkillinfo(key string) *msg.Skill {
	info := SkillConf.Get(key)
	skillId, _ := info.Get("jinengid").Int()
	interval, _ := info.Get("cd").Int()
	dur, _ := info.Get("chixushijian").Int()
	chance, _ := info.Get("jinenggailv").Int()
	skill := &msg.Skill{
		SkillId:  int32(skillId),
		Interval: int64(interval),
		Dur:      int64(dur),
		Chance:   int32(chance),
		LastTime: time.Now().UnixNano() / 1e6,
	}
	return skill
}

func GetSkillFishNum(fishId string, skillId int32) int {
	//key := ""
	//if skillId == 1 {
	//	key = "ldsl"
	//}
	//if skillId == 2 {
	//	key = "bzsl"
	//}
	num, err := GetFish(fishId).Get("gjbeishu").Int()
	if err != nil {
		log.Errorf("skill hit fish num err", err)
	}
	return num
}

func GetSkillId(fishId string) int32 {
	if fishId == "28" {
		return 1
	}
	if fishId == "29" {
		return 2
	}
	//if fishId == "30" {
	//	return 1
	//}
	//if fishId == "31" {
	//	return 2
	//}
	//if fishId == "32" {
	//	return 6
	//}
	if fishId == "24" || fishId == "25" || fishId == "26" || fishId == "27" || fishId == "23" {
		return 4
	}
	return 0
}

func GetConfLine(fishId string) ([]interface{}, []interface{}, interface{}, interface{}, int) {
	fish := GetFish(fishId)
	lineKeys, err := fish.Get("pzluxianID").Array()
	if err != nil {
		log.Errorf("line key err", err)
		return nil, nil, nil, nil, -1
	}
	//key := lineKeys[rand.RandInt(0, len(lineKeys))]
	key := lineKeys[tools.RandInt(0, len(lineKeys), 0)]
	lines, err := LineConf.Get(key.(string)).Array() //.Get("游动路线")
	if err != nil {                                  //key.(string)
		log.Errorf("line err", err)
		return nil, nil, nil, nil, -1
	}

	//r := rand.RandInt(0, len(lines))
	return GetLine(lines)
}

func GetLine(lines []interface{}) ([]interface{}, []interface{}, interface{}, interface{}, int) {
	r := tools.RandInt(0, len(lines), 1)
	name := lines[r].(map[string]interface{})["路线名称"]
	line := lines[r].(map[string]interface{})["游动路线"].([]interface{})
	speed := lines[r].(map[string]interface{})["游动速度"].([]interface{})
	t := lines[r].(map[string]interface{})["路线总速度"]
	return line, speed, t, name, r
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
	return tools.RandInt(min, max, 0)
}

func GetRobotRestTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("xxsj").Array()
	if err != nil {
		log.Errorf("robot rest time err", err)
	}
	min := num[0].(int)
	max := num[1].(int)
	return tools.RandInt(min, max, 0)
}

func GetBossAddTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("bossAddTime").Int()
	if err != nil {
		log.Errorf("bossAddTime err", err)
	}
	return num
}

func GetSkillAddTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("skillAddTime").Int()
	if err != nil {
		log.Errorf("skillAddTime err", err)
	}
	return num
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
func GetRobotSkillFrozenChance() int {
	num, err := RobotConf.Get("1").Get("1").Get("skillFrozenChance").Int()
	if err != nil {
		log.Errorf("robot skillFrozenChance err", err)
	}
	return num
}

func GetRobotSkillFrozenCheckTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("skillFrozenCheckTime").Int()
	if err != nil {
		log.Errorf("robot skillFrozenCheckTime err", err)
	}
	return num
}

func GetRobotSkillSummonCheckChance() int {
	num, err := RobotConf.Get("1").Get("1").Get("skillSummonChance").Int()
	if err != nil {
		log.Errorf("robot skillSummonChance err", err)
	}
	return num
}

func GetRobotSkillSummonCheckTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("skillSummonCheckTime").Int()
	if err != nil {
		log.Errorf("robot skillSummonCheckTime err", err)
	}
	return num
}

func GetRobotLockModelChance() int {
	num, err := RobotConf.Get("1").Get("1").Get("lockModelChance").Int()
	if err != nil {
		log.Errorf("robot lockModelChance err", err)
	}
	return num
}

func GetRobotModelCheckTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("modelCheckTime").Int()
	if err != nil {
		log.Errorf("robot modelCheckTime err", err)
	}
	return num
}

func GetRobotLockCheckTime() int {
	num, err := RobotConf.Get("1").Get("1").Get("LockCheckTime").Int()
	if err != nil {
		log.Errorf("robot LockCheckTime err", err)
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

func GetFormation(file, key string) map[string]interface{} {
	formation, err := FormationConf.Get(file).Get(key).Map()
	if err != nil {
		log.Errorf("formation err", err)
	}
	return formation
}

func GetFormationTime(islandId, sceneId string) int {
	t, err := FormationConf.Get(islandId).Get(sceneId).Get("pzshijian").Int()
	if err != nil {
		log.Errorf("Formationtime err", err)
	}
	return t
}

func GetAFormationFile(islandId, sceneId string) string {
	all, err := ChangJingConf.Get(islandId).Get(sceneId).Get("ycpeizhi").Array()
	if err != nil {
		log.Errorf("FormationFile err", err)
		return ""
	}
	return all[tools.RandInt(0, len(all), 0)].(string)
}

func GetAFormationKey(file string) string {
	formation, err := FormationConf.Get(file).Map()
	if err != nil {
		log.Errorf("formations err", err)
		return ""
	}
	keys := make([]string, 0)
	for k, _ := range formation {
		keys = append(keys, k)
	}
	return keys[tools.RandInt(0, len(keys), 0)]
}

func GetFormationFishInfo(formationKey, key, k string) (string, int32, int, int, []interface{}) {
	info := GetFormation(formationKey, key)[k].(map[string]interface{})
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
	info := GetFormation(formationKey, key)
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
	index := tools.RandInt(0, len(timeFish), 0)
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

func GetTimeFishInfo(key string) (string, int, int, int) {
	info := GetTimeFish()[key].(map[string]interface{})
	if info != nil {
		allId := info["Id"].([]interface{})
		id := allId[0].(string)
		if len(allId) > 1 {
			//id = allId[rand.RandInt(0, len(allId))].(string)
			id = allId[tools.RandInt(0, len(allId), 0)].(string)
		}
		//start, _ := strconv.Atoi(info["sxsj"].(json.Number).String())sxjg
		space, _ := strconv.Atoi(info["sxsj"].(json.Number).String())
		num, _ := strconv.Atoi(info["sxsl"].(json.Number).String())
		return id, 0, space, num
	}
	return "", 0, 0, 0
}

func GetXueChiChance(fishId, roomKey string, key int32) int32 {
	if key == -1 {
		return 0
	}
	if key == 0 {
		key = 1000
	}
	chance, err := PoolConf.Get(roomKey).Get(strconv.Itoa(int(key))).Get(fishId).Int()
	if err != nil {
		log.Errorf("xuechi chance err", err)
	}
	return int32(chance)
}

func GetFishTimerInfoByType(t msg.Type) (string, int, int, int) {
	key := getKeyByType(t)
	if key != "" {
		info := YuConf.Get("2").Get(key)
		times, _ := info.Get("sxsj").Array()
		min, _ := strconv.Atoi(times[0].(json.Number).String())
		max, _ := strconv.Atoi(times[1].(json.Number).String())
		t := tools.RandInt(min, max, 0)
		nums, _ := info.Get("sxsl").Array()
		num, _ := strconv.Atoi(nums[0].(json.Number).String())
		gailv, _ := info.Get("gailv").Int()
		if len(nums) == 2 && gailv > tools.RandInt(0, 100, 0) {
			num, _ = strconv.Atoi(nums[1].(json.Number).String())
		}

		ids, _ := info.Get("Id").Array()
		id := ids[tools.RandInt(0, len(ids), 0)].(string)
		spaces, _ := info.Get("space").Array()
		space := 0
		if len(spaces) == 2 {
			min, _ := strconv.Atoi(spaces[0].(json.Number).String())
			max, _ := strconv.Atoi(spaces[1].(json.Number).String())
			space = tools.RandInt(min, max, 0)
		}
		return id, t, num, space
	}
	return "", 0, 0, 0
}

func getKeyByType(t msg.Type) string {
	switch t {
	case msg.Type_SPECIAL:
		return "3"
	case msg.Type_PILE:
		return "2"
	case msg.Type_KING:
		return "1"
	default:
		return ""

	}
}

func GetAssociatedInfo(id string) (string, int) {
	fish := GetFish(id)
	associated, _ := fish.Get("ywshuliang").Int()
	associatedId, _ := fish.Get("ywID").String()
	return associatedId, associated
}

func GetDivider(id string) int {
	fish := GetFish(id)
	reduce, _ := fish.Get("reduce").Int()
	return reduce
}

func GetFishIdByMul(low, up int) string {
	fish := make([]string, 0)
	allFish, _ := YuConf.Get("1").Map()
	for k, v := range allFish {
		mul, _ := v.(*simplejson.Json).Get("fenzhi").Int()
		if mul >= low && mul <= up {
			fish = append(fish, k)
		}
	}
	if len(fish) > 0 {
		return fish[tools.RandInt(0, len(fish), 0)]
	}
	return ""
}

func GetSummonFishId() string {
	all, err := SkillConf.Get("1").Get("yuid").Array()
	if err != nil {
		log.Errorf("summon fishId err", err)
		return ""
	}
	return all[tools.RandInt(0, len(all), 0)].(string)
}

func GetSummonFishLine() ([]interface{}, []interface{}, interface{}, interface{}, int) {
	lines, err := LineConf.Get("zhaohuan").Array()
	if err != nil {
		log.Errorf("summon line err", err)
	}
	return GetLine(lines)
}

func GetBossFishLine() ([]interface{}, []interface{}, interface{}, interface{}, int) {
	lines, err := LineConf.Get("boss33luxian").Array()
	if err != nil {
		log.Errorf("boss line err", err)
	}
	return GetLine(lines)
}

func GetBossTurnInfo(islandId, sceneId string, lastIndex int) ([]interface{}, int) {
	all, err := ChangJingConf.Get(islandId).Get(sceneId).Get("bosschushoupeizhi").Array()
	if err != nil {
		log.Errorf("bosschushoupeizhi err", err)
	}
	index := 0
	for i := 0; i < 100; i++ {
		index = tools.RandInt(0, len(all), 0)
		if index != lastIndex {
			break
		}
	}
	if len(all) <= index {
		return make([]interface{}, 0), 0
	}
	return all[index].([]interface{}), index
}

func GetBossTurnTime(islandId, sceneId string) int {
	t, err := ChangJingConf.Get(islandId).Get(sceneId).Get("bossyuzhuanxiangshijian").Int()
	if err != nil {
		log.Errorf("bossyuzhuanxiangshijian err", err)
	}
	return t
}

func GetBossTurnChance(islandId, sceneId string) int32 {
	chance, err := ChangJingConf.Get(islandId).Get(sceneId).Get("bosszhunxianggailv").Int()
	if err != nil {
		log.Errorf("bosszhunxianggailv err", err)
	}
	return int32(chance)
}

func GetPool() {

}
