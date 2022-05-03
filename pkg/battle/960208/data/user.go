package data

import (
	"encoding/json"

	"github.com/kubegames/kubegames-games/pkg/battle/960208/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// User 玩家 21点 游戏属性
type User struct {
	ID          int64                  // 框架那边proto定义的int64
	Nick        string                 // 昵称
	Head        string                 // 头像
	User        player.PlayerInterface // userInter interface
	Status      int32                  // 玩家状态
	HoldCards   *poker.HoldCards       // 持有手牌
	CurAmount   int64                  // 当前持有数量
	InitAmount  int64                  // 初始数量
	ChairID     int32                  // 椅子id
	IsRob       bool                   // 是否抢庄
	IsBanker    bool                   // 是否是庄家
	BetMultiple int64                  // 闲家投注倍数
	Multiples   []int64                // 可投注选项
	ReConnect   bool                   // 是否是断线重联登陆上来的
	DemandReq   DemandReq              // 配牌请求
	ExitPermit  bool                   // 离开权限
}

// 战绩信息
type GameRecord struct {
	Time      int32  // 时间
	GameNum   string // 牌局编号
	RoomLevel int32  // 房间类型
	Result    int64  // 盈利结果
	Status    int32  // 状态 1 已结算 2 结算中
}

// 配牌请求
type DemandReq struct {
	DemandType  int32  // 配牌类型
	DemandCards []byte // 配牌手牌
}

const (

	// 系统角色
	SysRolePlayer = "玩家"  // 玩家
	SysRoleRobot  = "机器人" // 机器人

)

// GetSysRole 获取系统角色
func (user *User) GetSysRole() (SysRole string) {
	SysRole = SysRolePlayer
	if user.User.IsRobot() {
		SysRole = SysRoleRobot
	}

	return
}

// GetUserTabledata 获取依赖全局user的游戏数据
func GetUserTabledata(userInter player.PlayerInterface) (data []GameRecord) {

	interData := userInter.GetTableData()
	if len(interData) != 0 {

		err := json.Unmarshal([]byte(interData), &data)
		if err != nil {
			log.Errorf("json unmarshal userInter data fail: %v", err)
			return
		}
	}
	return
}

// SetUserTabledata 设置依赖全局user的游戏数据
func SetUserTabledata(userInter player.PlayerInterface, data []GameRecord) {

	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("json marshal userInter data fail: %v", err)
	}

	userInter.SetTableData(string(b))
}
