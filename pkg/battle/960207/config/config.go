package conf

import (
	"common/log"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sipt/GoJsoner"
)

// RoomConfig 游戏配置
type RoomConfig struct {
	RoomCost int64 `json:"room_cost"` // 底注
	MinLimit int64 `json:"min_limit"` // 下注门槛
	TaxRatio int64 `json:"tax_ratio"` // 税收比例
}

// GameConfig 游戏配置
type GameConfig struct {
	RobOption  []int64 `json:"rob_option"`  // 抢庄选项
	NumberRate []int   `json:"number_rate"` // 桌子人数概率
	Control    struct {
		ControlRate       []int32 `json:"control_rate"`        // 作弊率等级
		PlayerBiggestRate []int   `json:"player_biggest_rate"` // 玩家最大牌概率分布
		RobotBiggestRate  []int   `json:"robot_biggest_rate"`  // 机器人最大牌概率分布
	} `json:"control"` // 控制配置信息
	BetOption []int64 `json:"bet_option"` // 投注选项
}

// TimeConfig 时间配置
type TimeConfig struct {
	StartMove           int `json:"start_move"`            // 开始倒计时时间
	StartAnimation      int `json:"start_animation"`       // 开始动画时间
	BetChips            int `json:"bet_chips"`             // 投注时间
	DealAnimation       int `json:"deal_animation"`        // 单人发牌动画时间
	SecondDealAnimation int `json:"second_deal_animation"` // 第二次发牌动画时间
	ShowCards           int `json:"show_cards"`            // 摊派时间
	Settle              int `json:"settle"`                // 结算时间
	StatusSpace         int `json:"status_space"`          // 状态间隔时间
	DelayCheckMatch     int `json:"delay_check_match"`     // 玩家准备延迟检测匹配机器人时间
	LoopCheckMatch      int `json:"loop_check_match"`      // 循环检测匹配机器人时间
	CheckCardsType      int `json:"check_cards_type"`      // 查看牌型时间
}

// ExactControl 点控
type ExactControl struct {
	ExactControlRate []int `json:"exact_control_rate"` // 用户点控作弊率
}

// GeneralNiuniuConfig 牛牛配置
type GeneralNiuniuConfig struct {
	GameConfig GameConfig `json:"game_config"`
	TimeConfig TimeConfig `json:"time_config"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	ActionTime struct {
		Shortest int `json:"shortest"` // 最短等待时间
		Longest  int `json:"longest"`  // 最长等待时间
	} `json:"action_time"` // 延迟操作时间
	BigCardsBetRateDis   []int `json:"big_cards_bet_rate_dis"`   // 大牌投注概率分布
	SmallCardsBetRateDis []int `json:"small_cards_bet_rate_dis"` // 小牌投注概率分布
}

// GeneralNiuniuConf 通比牛牛配置
var GeneralNiuniuConf GeneralNiuniuConfig

// RobotConf 机器人配置
var RobotConf RobotConfig

// LoadBlackjackCfg 读取配置文件
func (conf *GeneralNiuniuConfig) LoadGeneralNiuniuCfg() {
	data, err := ioutil.ReadFile("config/general_niuniu.json")
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
