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
	RobScore          []int64 `json:"rob_score"`            // 抢庄分数
	AddMultiple       []int64 `json:"add_multiple"`         // 加倍倍数
	FullTableTimeRate []int   `json:"full_table_time_rate"` // 满桌时间概率
	Control           struct {
		ControlRate      []int32 `json:"control_rate"`       // 作弊率等级
		BiggestCardsRate []int   `json:"biggest_cards_rate"` // 最大牌概率分布
	} `json:"control"` // 控制配置信息
}

// TimeConfig 时间配置
type TimeConfig struct {
	DealAnimation    int `json:"deal_animation"`     // 发牌动画时间
	RobTime          int `json:"rob_time"`           // 单次抢地主时间
	ConfirmDizhuTime int `json:"confirm_dizhu_time"` // 确认地主时间
	RedoubleTime     int `json:"redouble_time"`      // 加倍时间
	OperationTime    int `json:"operation_time"`     // 执行时间
	ExcessiveTime    int `json:"excessive_time"`     // 过渡时间
	SpaceTime        int `json:"space_time"`         // 空格时间
	HangUpTime       int `json:"hang_up_time"`       // 托管时间
	SettleTime       int `json:"settle_time"`        // 结算时间
	RobotSitCheck    int `json:"robot_sit_check"`    // 轮询检测添加机器人时间
	OneSecond        int `json:"one_second"`         // 一秒等待时间
	RocketTime       int `json:"rocket_time"`        // 火箭动画时间
	BombTime         int `json:"bomb_time"`          // 炸弹动画时间
	SerialPairTime   int `json:"serial_pair_time"`   // 连对动画时间
	SequenceTime     int `json:"sequence_time"`      // 顺子动画时间
	PlaneTime        int `json:"plane_time"`         // 飞机动画时间
	DefaultTime      int `json:"default_time"`       // 普通牌型动画时间
}

// DoudizhuConfig 斗地主配置
type DoudizhuConfig struct {
	GameConfig GameConfig `json:"game_config"`
	TimeConfig TimeConfig `json:"time_config"`
}

// RobotConfig 机器人配置信息
type RobotConfig struct {
	RobScorePlace []int   `json:"rob_score_place"` // 抢庄分值分布
	RobRatePlace  [][]int `json:"rob_rate_place"`  // 抢庄概率分布
	AddScorePlace []int   `json:"add_score_place"` // 加倍分值分布
	AddRatePlace  [][]int `json:"add_rate_place"`  // 加倍概率分布
}

// CardsOrderValue 牌型顺序值配置
type CardsOrderValue struct {
	CardsType   int32 `json:"cards_type"`
	Length      int   `json:"length"`
	WeightValue byte  `json:"weight_value"`
	OrderValue  int   `json:"order_value"`
}

// CardsOrderValue 牌型顺序值配置表
//type CardsOrderConfig struct {
//	CardsOrderForm []CardsOrderValue `json:"cards_order_form"`
//}

// CardsPutScoreConfig 出牌权值配置表
type PutScoreConfig struct {
	PutScoreForm []PutScore `json:"put_score_form"`
}

// 出牌权值配置
type PutScore struct {
	CardsType  int32 `json:"cards_type"`
	FirstValue byte  `json:"first_value"`
	PutScore   int   `json:"put_score"`
}

// DoudizhuConf 斗地主配置
var DoudizhuConf DoudizhuConfig

// RobotConf 机器人配置
var RobotConf RobotConfig

// CardsOrderConf 牌型顺序值配置
//var CardsOrderConf CardsOrderConfig

// PutScoreConf 出牌权值配置
var PutScoreConf PutScoreConfig

// LoadBlackjackCfg 读取配置文件
func (conf *DoudizhuConfig) LoadDoudizhuCfg() {
	data, err := ioutil.ReadFile("config/doudizhu.json")
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

// CardsPutScoreCfg 读取出牌权值配置
func (PutScoreCfg *PutScoreConfig) LoadPutScoreCfg() {
	data, err := ioutil.ReadFile("config/cardsPutScore.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))

	if err := json.Unmarshal([]byte(result), &PutScoreCfg); err != nil {
		log.Errorf("Unmarshal json error : %v", err)
		return
	}
}

// LoadCardsOrderCfg 读取牌型顺序表配置
func (PutScoreCfg *PutScoreConfig) GetPutScore(cardsType int32, firstValue byte) (putScore int) {
	for _, v := range PutScoreCfg.PutScoreForm {
		if cardsType == v.CardsType &&
			firstValue == v.FirstValue {
			putScore = v.PutScore
			break
		}
	}
	return
}
