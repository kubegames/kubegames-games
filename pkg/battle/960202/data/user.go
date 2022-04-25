package data

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960202/poker"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// User 玩家 抢庄牛牛 游戏属性
type User struct {
	ID                int64                  // 框架那边proto定义的int64
	User              player.PlayerInterface // userInter interface
	Nick              string                 // 昵称
	Head              string                 // 头像
	Status            int32                  // 玩家状态
	CurAmount         int64                  // 当前持有数量
	InitAmount        int64                  // 初始金额
	ChairID           int32                  // 椅子id
	HoldCards         *poker.HoldCards       // 持有手牌
	IsBanker          bool                   // 是不是庄家
	RobIndex          int32                  // 抢庄下标
	BetMultiple       int64                  // 闲家投注倍数
	BetMultipleOption []int64                // 闲家可投注选项
	HighestMultiple   int64                  // 闲家最高可投注倍数
	ExactControlRate  int32                  // 点控用户作弊率
	ReConnect         bool                   // 是否是断线重联登陆上来的
	ExitPermit        bool                   // 离开权限
	GetBiggest        bool                   // 拿最大牌
	GetSecond         bool                   // 拿第二大牌
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
