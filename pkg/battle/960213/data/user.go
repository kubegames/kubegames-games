package data

import "github.com/kubegames/kubegames-sdk/pkg/player"

// User 玩家 斗地主 游戏属性
type User struct {
	ID               int64                  // 框架那边proto定义的int64
	User             player.PlayerInterface // userInter interface
	Nick             string                 // 昵称
	Head             string                 // 头像
	Status           int32                  // 玩家状态
	IsDizhu          bool                   // 是不是地主
	Cards            []byte                 // 手牌
	CurAmount        int64                  // 当前持有数量
	InitAmount       int64                  // 初始金额
	ChairID          int32                  // 椅子id
	ExactControlRate int32                  // 点控用户作弊率
	ReConnect        bool                   // 是否是断线重联登陆上来的
	PutCardsRecords  [][]byte               // 出牌记录
	ExitPermit       bool                   // 离开权限
	RobNum           int64                  // 抢分数 -1, 0, 1, 2, 3 表示 未发送抢分请求/不抢/1分/2分/3分
	AddNum           int64                  // 加倍倍数 0, 1, 2, 4 表示 未发送加倍请求/不加倍/2倍/4倍
	TotalMultiple    int64                  // 总倍数
	SettleResult     int64                  // 结算金额
	Role             int                    // 角色
}

const (
	RoleDizhu       = 0 // 地追角色
	RoleDownPeasant = 1 // 下家农民
	RoleUpPeasant   = 2 // 上家农民
)
