package config

import (
	"common/log"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/tidwall/gjson"
)

type config struct {
	ListenerPort int    //web socket 侦听端口
	MaxConn      int32  //最大连接数
	DBUri        string //数据库链接字符串
	LogPath      string //需要制定输出的日志路径
}

type aiConfig struct {
	IsRobotRobOn bool
	RobotConfig  []*CheatJson
}

type CheatJson struct {
	Cheat      int32
	AiMine     int
	UserMine   int
	Rob1Config []*InterTimeRate
	Rob2Config []*InterTimeRate
}

//间隔时间概率
type InterTimeRate struct {
	InterTime int
	Rate      int
}

type aiSendConfig struct {
	AiCountMin      int
	AiCountMax      int
	MinLeftRedCount int

	SendCount1    int64 //发送红包数量档位1
	SendCount2    int64 //发送红包数量档位2
	SendCount3    int64 //发送红包数量档位3
	SendAmountMax int64 //发送红包最大金额    分

	S2Send        int //第2s发红包的概率
	S2SendCount1  int //第2s发红包数量1的概率
	S2SendCount2  int //第2s发红包数量2的概率
	S2SendCount3  int //第2s发红包数量1的概率
	S2SendAmount1 int //第2s发红包金额1的概率
	S2SendAmount2 int //第2s发红包金额2的概率
	S2SendAmount3 int //第2s发红包金额3的概率

	S3Send        int //第3s发红包的概率
	S3SendCount1  int //第3s发红包数量1的概率
	S3SendCount2  int //第3s发红包数量2的概率
	S3SendCount3  int //第3s发红包数量1的概率
	S3SendAmount1 int //第3s发红包金额1的概率
	S3SendAmount2 int //第3s发红包金额2的概率
	S3SendAmount3 int //第3s发红包金额3的概率

	S4Send        int //第4s发红包的概率
	S4SendCount1  int //第4s发红包数量1的概率
	S4SendCount2  int //第4s发红包数量2的概率
	S4SendCount3  int //第4s发红包数量1的概率
	S4SendAmount1 int //第4s发红包金额1的概率
	S4SendAmount2 int //第4s发红包金额2的概率
	S4SendAmount3 int //第4s发红包金额3的概率

	S5Send        int //第5s发红包的概率
	S5SendCount1  int //第5s发红包数量1的概率
	S5SendCount2  int //第5s发红包数量2的概率
	S5SendCount3  int //第5s发红包数量1的概率
	S5SendAmount1 int //第5s发红包金额1的概率
	S5SendAmount2 int //第5s发红包金额2的概率
	S5SendAmount3 int //第5s发红包金额3的概率

	S6Send        int //第6s发红包的概率
	S6SendCount1  int //第6s发红包数量1的概率
	S6SendCount2  int //第6s发红包数量2的概率
	S6SendCount3  int //第6s发红包数量1的概率
	S6SendAmount1 int //第6s发红包金额1的概率
	S6SendAmount2 int //第6s发红包金额2的概率
	S6SendAmount3 int //第6s发红包金额3的概率

	S7Send        int //第7s发红包的概率
	S7SendCount1  int //第7s发红包数量1的概率
	S7SendCount2  int //第7s发红包数量2的概率
	S7SendCount3  int //第7s发红包数量1的概率
	S7SendAmount1 int //第7s发红包金额1的概率
	S7SendAmount2 int //第7s发红包金额2的概率
	S7SendAmount3 int
}

type AiRobConfig struct {
	Cheat      int32
	AiMine     int
	UserMine   int
	RobRateArr []*RobRate
}
type RobRate struct {
	AiS1Rate  int //第1个机器人红包出现第1s抢包的概率
	AiS2Rate  int //第1个机器人红包出现第2s抢包的概率
	AiS3Rate  int //第1个机器人红包出现第3s抢包的概率
	AiS4Rate  int //第1个机器人红包出现第4s抢包的概率
	AiS5Rate  int //第1个机器人红包出现第5s抢包的概率
	AiS6Rate  int //第1个机器人红包出现第6s抢包的概率
	AiS7Rate  int //第1个机器人红包出现第7s抢包的概率
	AiS8Rate  int //第1个机器人红包出现第8s抢包的概率
	AiS9Rate  int //第1个机器人红包出现第9s抢包的概率
	AiS10Rate int
}

var AiRobConfigMap = make(map[int32]*AiRobConfig) // cheat => AiRobConfig

type GameConfig struct {
	Id           int64
	Level        string  //初级、中级、高级房
	MinLimit     int64   //最小入场限制
	MinAction    int64   //底分额度
	RaiseAmount  []int64 //加注额度
	MaxRound     int32   //最高轮数
	MaxAllIn     int64   //全押最高额度
	Tax          int64   //税收率 5 就表示 5%
	AiMinLimit   int64
	AIMaxLimit   int64 //
	AiZhengChang int
	AiJiJin      int
	AiWenZhong   int
	AiTouJi      int
}

var Config = new(config)
var AiSendConfig = new(aiSendConfig)
var AiRobConfigArr = make([]*AiRobConfig, 0)
var GameConfigArr = make([]*GameConfig, 0)
var AiConfig = new(aiConfig)

type gameFrameConfig struct {
	DbIp   string `json:"db_ip"`
	DbPwd  string `json:"db_pwd"`
	DbUser string `json:"db_user"`
	DbName string `json:"db_name"`
}

var GameFrameConfig = new(gameFrameConfig)

type redConfig struct {
	Odds        int32
	Count       int32
	MinAmount   int64
	MaxAmount   int64
	SpaceAmount int64
	RedFlood    int64
}

var RedConfig = new(redConfig)

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

//3月5号新增

//发红包
type AiSendNew struct {
	RedCount       []int
	Interval       []int
	SendCount      []int
	SendCountRate  []int
	SendAmountRate []int
}

var AiSendNewArr = make([]*AiSendNew, 0)

func GetAiSendNew(redCount int) *AiSendNew {
	//fmt.Println("AiSendNewArr : ",fmt.Sprintf(`%+v`,AiSendNewArr),len(AiSendNewArr))
	for _, v := range AiSendNewArr {
		//fmt.Println(fmt.Sprintf(`%+v`,v))
		//fmt.Println("red count : ",redCount, v.RedCount)
		if redCount >= v.RedCount[0] && redCount <= v.RedCount[1] {
			return v
		}
	}
	return nil
}

type AiRobNew struct {
	RedCount     []int
	Interval     []int
	RobCount     []int
	RobCountRate []int
	AiCount      int
}

var AiRobNewArr = make([]*AiRobNew, 0)

func GetAiRobNew(redCount int) *AiRobNew {
	//fmt.Println("AiSendNewArr : ",fmt.Sprintf(`%+v`,AiSendNewArr),len(AiSendNewArr))
	for _, v := range AiRobNewArr {
		//fmt.Println(fmt.Sprintf(`%+v`,v))
		//fmt.Println("red count : ",redCount, v.RedCount)
		if redCount >= v.RedCount[0] && redCount <= v.RedCount[1] {
			return v
		}
	}
	return nil
}

//////////////3月16号新增
////////////
func LoadCrazyRedConfig() {
	data, err := ioutil.ReadFile("./config/robot.json")
	if err != nil {
		log.Tracef("baijialeconfig Error %v", err.Error())
		return
	}
	result := gjson.Parse(string(data))
	InitConfig(result)
}

type CrazyRedConfig struct {
	Robotgold [][]int64
}

var crazyRedConfig CrazyRedConfig

func InitConfig(cfg gjson.Result) {
	for i := 1; i <= 4; i++ {
		str := fmt.Sprintf("robotgold.%v", i)
		robotgold := cfg.Get(str).Array()
		var gold []int64
		for _, v := range robotgold {
			gold = append(gold, v.Int())
		}
		crazyRedConfig.Robotgold = append(crazyRedConfig.Robotgold, gold)
	}
	fmt.Println("crazyRedConfig.Robotgold : ", crazyRedConfig.Robotgold)
}

//获取机器人的配置[min,max]
func GetRobotConfByLevel(level int32) []int64 {
	switch level {
	case 1:
		return crazyRedConfig.Robotgold[0]
	case 2:
		return crazyRedConfig.Robotgold[1]
	case 3:
		return crazyRedConfig.Robotgold[2]
	case 4:
		return crazyRedConfig.Robotgold[3]
	default:
		return crazyRedConfig.Robotgold[0]
	}
}
