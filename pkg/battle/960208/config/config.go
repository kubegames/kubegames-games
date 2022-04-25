package config

import (
	"common/log"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sipt/GoJsoner"
)

// GameConfig 游戏配置
type GameConfig struct {
	NumberRate []int `json:"number_rate"` // 桌子人数概率
	Control    struct {
		ControlRate       []int32 `json:"control_rate"`        // 作弊率等级
		PlayerBiggestRate []int   `json:"player_biggest_rate"` // 玩家最大牌概率分布
		RobotBiggestRate  []int   `json:"robot_biggest_rate"`  // 机器人最大牌概率分布
		PlayerSecondRate  []int   `json:"player_second_rate"`  // 玩家第二大牌概率分布
		RobotSecondRate   []int   `json:"robot_second_rate"`   // 机器人第二大牌概率分布
	} `json:"control"` // 控制配置信息
}

// GameConfig 游戏配置
type RoomConfig struct {
	RoomCost int64 // 底注
	MinLimit int64 // 最小限制
	TaxRate  int64 // 税收比例
}

// TimeConfig 时间配置
type TimeConfig struct {
	CountDown        int `json:"count_down"`         // 倒计时时间
	DealCards        int `json:"deal_cards"`         // 开始动画时间,发牌动画时间
	RobBanker        int `json:"rob_banker"`         // 抢庄时间
	RobAnimation     int `json:"rob_animation"`      // 抢庄动画
	BetChips         int `json:"bet_chips"`          // 投注时间
	ShowCards        int `json:"show_cards"`         // 摊牌时间
	EndShowAnimation int `json:"end_show_animation"` // 摊牌结束动画
	RobotSitCheck    int `json:"robot_sit_check"`    // 轮询检测添加机器人时间
}

// ThreeDollConfig 三公配置
type ThreeDollConfig struct {
	GameConf GameConfig `json:"game_conf"`
	TimeConf TimeConfig `json:"time_conf"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	RobotActionRate []int `json:"robot_action_rate"` // 机器人操作时间概率分布
	BetRate         []int `json:"bet_rate"`          // 投注概率分布
}

// ThreeDollConf 三公配置
var ThreeDollConf ThreeDollConfig

// RobotConf 机器人配置
var RobotConf RobotConfig

// LoadBlackjackCfg 读取配置文件
func (conf *ThreeDollConfig) LoadThreeDollCfg() {
	data, err := ioutil.ReadFile("config/three_doll.json")
	if err != nil {
		log.Errorf("File reading error : %v", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &conf); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}

}

// LoadRobotCfg 读取机器人配置文件
func (robotCfg *RobotConfig) LoadRobotCfg() {
	data, err := ioutil.ReadFile("config/robot.json")
	if err != nil {
		fmt.Println("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &robotCfg); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}
}
