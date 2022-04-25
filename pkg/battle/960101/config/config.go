package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

// GameConfig 游戏配置
type GameConfig struct {
	NumberRate  []int      `json:"number_rate"`   // 桌子人数概率
	Chips       [4][]int64 `json:"chips"`         // 下注筹码
	MinBetLimit [4]int64   `json:"min_bet_limit"` // 最低下注限制
	MaxBetLimit [4]int64   `json:"max_bet_limit"` // 最高下注限制
}

// GameConfig 游戏配置
type RoomConfig struct {
	MinAction    int64   // 最小下注
	MaxAction    int64   // 最大下注
	LimitAction  int64   // 下注门槛
	ActionOption []int64 // 可下注选项
	TaxRate      int64   // 税收比例
	RoomCost     int64   // 底注
}

// TimeConfig 时间配置
type TimeConfig struct {
	StartMove              int `json:"start_move"`                // 开始倒计时时间
	BetAnimation           int `json:"bet_animation"`             // 开始下注动画
	BetStatus              int `json:"bet_status"`                // 下注时间
	EndBetStatus           int `json:"end_bet_status"`            // 下注结束时间
	FirstMove              int `json:"first_move"`                // 第一轮发牌单人动画时间
	InsuranceStatus        int `json:"insurance_status"`          // 买保险时间
	EndInsuranceSettle     int `json:"end_insurance_settle"`      // 结束保险，飘筹码动画时间
	EndInsurance           int `json:"end_insurance"`             // 结束保险，到下一阶段时间间隔
	UserAction             int `json:"user_action"`               // 玩家操作时间
	UserActionTimeInterval int `json:"user_action_time_interval"` // 玩家操作时间间隔
	AdvanceSettle          int `json:"advance_settle"`            // 结算飘筹码动画时间
	SettleAnimation        int `json:"settle_animation"`          // 黑杰克，五小龙，21点，爆牌结算动画时间
	HostAction             int `json:"host_action"`               // 庄家操作动画时间
	SettleStatus           int `json:"settle_status"`             // 结算状态
	RobotSitCheck          int `json:"robot_sit_check"`           // 机器人坐下检测轮询时间
	CheckOnlineUser        int `json:"check_online_user"`         // 检测长时间在线不重新匹配玩家
	MsgDelay               int `json:"msg_delay"`                 // 消息延迟间隔
}

// 控制配置
type ExactControl struct {
	BankerControl    bool    `json:"banker_control"`     // 庄家控牌开关
	ExactControlRate []int64 `json:"exact_control_rate"` // 作弊率等级
	BlackjackPlace   []int   `json:"blackjack_place"`    // 初始黑杰克概率分布
	FirstBustPlace   []int   `json:"first_bust_place"`   // 第一次强制爆牌概率分布
	SecondBustPlace  []int   `json:"second_bust_place"`  // 第二次强制爆牌概率分布
	ThirdBustPlace   []int   `json:"third_bust_place"`   // 第三次强制爆牌概率分布
}

// Blackjackconfig 黑杰克配置
type Blackjackconfig struct {
	GameConf     GameConfig   `json:"game_conf"`
	TimeConf     TimeConfig   `json:"time_conf"`
	ExactControl ExactControl `json:"exact_control"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	ActionTime struct {
		Shortest int `json:"shortest"` // 最短等待时间
		Longest  int `json:"longest"`  // 最长等待时间
	} `json:"action_time"` // 操作时间

	CapitalDivision   [][]int64 `json:"capital_division"`   // 资金分布
	BetPlace          [][]int   `json:"bet_place"`          // 下注分布权重
	InsuranceDivision [][]int   `json:"insurance_division"` // 买保险资金分部
	InsurancePlace    [][]int   `json:"insurance_place"`    // 买保险权重
	DepartDivision    [][]int32 `json:"depart_division"`    // 分牌注码分布
	DepartPlace       [][]int   `json:"depart_place"`       // 分牌权重

	DoublePlace struct {
		Less9Point  int `json:"less_9_point"`  // 小于9点
		Less11Point int `json:"less_11_point"` // 大于9点小于11点
		More11Point int `json:"more_11_point"` // 大于11点
	} `json:"double_place"` // 双倍分布权重

	GetPokerDivision [][]int32 `json:"get_poker_division"` // 要牌点数分部
	GetPokerPlace    [][]int   `json:"get_poker_place"`    // 要牌权重
}

// BlackJackConf 黑杰克配置
var BlackJackConf Blackjackconfig

// RobotConf 机器人配置
var RobotConf RobotConfig

// LoadBlackjackCfg 读取配置文件 todo 选择不同桌子类型
func (conf *Blackjackconfig) LoadBlackjackCfg() {
	data, err := ioutil.ReadFile("conf/Blackjackconfig.json")
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
	data, err := ioutil.ReadFile("conf/robot.json")
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
