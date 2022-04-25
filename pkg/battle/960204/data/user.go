package data

import "github.com/kubegames/kubegames-sdk/pkg/player"

// User 玩家 跑得快 游戏属性
type User struct {
	ID               int64                  // 框架那边proto定义的int64
	User             player.PlayerInterface // userInter interface
	Nick             string                 // 昵称
	Head             string                 // 头像
	Status           int32                  // 玩家状态
	Cards            []byte                 // 手牌
	CurAmount        int64                  // 当前持有数量
	InitAmount       int64                  // 初始金额
	ChairID          int32                  // 椅子id
	ExactControlRate int32                  // 点控用户作弊率
	ReConnect        bool                   // 是否是断线重联登陆上来的
	BoomSettle       int64                  // 炸弹结算积分
	BoomCount        int                    // 炸弹结算个数
	TakeSingleRisk   bool                   // 是否承担放走包赔的风险
	SettleCost       int64                  // 结算基数
	PutCardsRecords  [][]byte               // 出牌记录
	ExitPermit       bool                   // 离开权限
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
