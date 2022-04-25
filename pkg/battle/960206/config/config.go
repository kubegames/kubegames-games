package config

import (
	"encoding/json"
	"os"
)

type config struct {
	ListenerPort int    //web socket 侦听端口
	MaxConn      int32  //最大连接数
	DBUri        string //数据库链接字符串
	LogPath      string //需要制定输出的日志路径
}

var Config = new(config)

type gameFrameConfig struct {
	DbIp   string `json:"db_ip"`
	DbPwd  string `json:"db_pwd"`
	DbUser string `json:"db_user"`
	DbName string `json:"db_name"`
}

var GameFrameConfig = new(gameFrameConfig)

var CheatConf = new(CheatConfig)

type CheatConfig struct {
	ControlRate       []int32 `json:"control_rate"`        // 作弊率等级
	PlayerBiggestRate []int   `json:"player_biggest_rate"` // 玩家最大牌概率分布
	RobotBiggestRate  []int   `json:"robot_biggest_rate"`  // 机器人最大牌概率分布
	PlayerSecondRate  []int   `json:"player_second_rate"`  // 玩家第二大牌概率分布
	RobotSecondRate   []int   `json:"robot_second_rate"`   // 机器人第二大牌概率分布
}

var RobotConf = new(RobotConfig)

type RobotConfig struct {
	ShowdownTime [][]int `json:"showdown_time"` // 延迟摊牌时间
	ShowdownRate []int   `json:"showdown_rate"` // 延迟摊牌概率分布
}

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
