package config

import (
	"encoding/json"
	"io/ioutil"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/sipt/GoJsoner"
)

// RoomConfig 游戏配置
type RoomConfig struct {
	RoomCost int64 `json:"room_cost"` // 底注
	MinLimit int64 `json:"min_limit"` // 下注门槛
	TaxRate  int64 `json:"tax_rate"`  // 税收比例
	Level    int32 `json:"level"`     // 房间等级
}

// GameConfig 游戏配置
type GameConfig struct {
	FullTableTimeRate []int `json:"full_table_time_rate"` // 满桌时间概率
	Control           struct {
		ControlRate       []int32 `json:"control_rate"`        // 作弊率等级
		PlayerBiggestRate []int   `json:"player_biggest_rate"` // 玩家最大牌概率分布
		RobotBiggestRate  []int   `json:"robot_biggest_rate"`  // 机器人最大牌概率分布
		ExchangeRate      []int   `json:"exchange_rate"`       // 玩家换牌概率分布
		BuildBoomRate     []int   `json:"build_boom_rate"`     // 机器人造炸弹概率分布
		ExchangeBlockRate []int   `json:"exchange_block_rate"` // 换大小三同张或者对子概率分布
	} `json:"control"` // 控制配置信息
}

// TimeConfig 时间配置
type TimeConfig struct {
	OperationTime int `json:"operation_time"`  // 执行时间
	ExcessiveTime int `json:"excessive_time"`  // 过渡时间
	SpaceTime     int `json:"space_time"`      // 空格时间
	RobotSitCheck int `json:"robot_sit_check"` // 轮询检测添加机器人时间
	OneSecond     int `json:"one_second"`      // 一秒等待时间
}

// RunFasterConfig 牛牛配置
type RunFasterConfig struct {
	GameConfig GameConfig `json:"game_config"`
	TimeConfig TimeConfig `json:"time_config"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	ActionTimePlace     [][2]int `json:"action_time_place"`      // 机器人操作时间分布
	ActionTimeRatePlace []int    `json:"action_time_rate_place"` // 机器人操作时间概率分布
}

// CardsOrderValue 牌型顺序值配置
type CardsOrderValue struct {
	CardsType   int32 `json:"cards_type"`
	Length      int   `json:"length"`
	WeightValue byte  `json:"weight_value"`
	OrderValue  int   `json:"order_value"`
}

// CardsOrderValue 牌型顺序值配置表
type CardsOrderConfig struct {
	CardsOrderForm []CardsOrderValue `json:"cards_order_form"`
}

// RunFasterConf 跑得快配置
var RunFasterConf RunFasterConfig

// RobotConf 机器人配置
var RobotConf RobotConfig

// CardsOrderConf 牌型顺序值配置
var CardsOrderConf CardsOrderConfig

// LoadBlackjackCfg 读取配置文件
func (conf *RunFasterConfig) LoadRunFasterCfg() {
	data, err := ioutil.ReadFile("conf/runFaster.json")
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
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &robotCfg); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}
}

// LoadCardsOrderCfg 读取牌型顺序表配置
func (CardsOrderCfg *CardsOrderConfig) LoadCardsOrderCfg() {
	data, err := ioutil.ReadFile("conf/cardsOrderForm.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &CardsOrderCfg); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}
}

// LoadCardsOrderCfg 读取牌型顺序表配置
func (CardsOrderCfg *CardsOrderConfig) GetCardsOrderValue(cardsType int32, weightValue byte, length int) (orderValue int) {
	for _, cardsOrderValue := range CardsOrderCfg.CardsOrderForm {
		if cardsType == cardsOrderValue.CardsType &&
			weightValue == cardsOrderValue.WeightValue &&
			length == cardsOrderValue.Length {
			orderValue = cardsOrderValue.OrderValue
			break
		}
	}
	return
}
