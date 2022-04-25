package data

import (
	"encoding/json"

	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

// User 玩家 21点 游戏属性
type User struct {
	ID               int64                  // 框架那边proto定义的int64
	UserName         string                 // 名称
	Head             string                 // 头像
	User             player.PlayerInterface // userInter interface
	Status           int32                  // 玩家状态
	BetAmount        int64                  // 下注数量
	CurAmount        int64                  // 当前持有数量
	InitAmount       int64                  // 初始金额
	HoldCards        [2]*HoldCards          // 持有手牌
	Insurance        int64                  // 保险金
	IsBuyInsure      bool                   // 是否购买保险
	ChairID          int32                  // 椅子id
	DepartFactor     byte                   // 分牌因子
	ReceiveInsurance bool                   // 收到买保险信息
	TestCardsType    int32                  // 测试想要点牌型
	ExactControlRate int64                  // 点控用户作弊率
	ReConnect        bool                   // 是否是断线重联登陆上来的
	Chip             int64                  // 打码量
}

// HoldCards 持有手牌
type HoldCards struct {
	Cards        []byte        // 手牌
	Point        []int32       // 点数
	Type         msg.CardsType // 牌类型
	BetAmount    int64         // 下注数量
	ActionPermit bool          // 可操作权限
	StopAction   int32         // 停止操作原因
	EndType      int           // 结束类型
}

// 结束类型
const (
	EndType_Unknow     = 0 // 未知
	EndType_Blackjack  = 1 // 黑杰克
	EndType_Boom       = 3 // 爆牌
	EndType_GiveUp     = 4 // 认输
	EndType_BankerWin  = 5 // 庄家大
	EndType_BankerLoss = 6 // 庄家小
	EndType_draw       = 7 // 打平

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

// UserData 存在框架user.data中，作为全局数据
type UserData struct {
	LastBetAmount int64 // 上次下注数量
	NotBetCount   int   // 未下注局数
}

// AppendCard 添加一张牌
func (user *User) AppendCard(index int32, card byte, betAmount int64) {
	holdCards := user.HoldCards[index]
	cards := append(holdCards.Cards, card)

	user.HoldCards[index] = &HoldCards{
		Cards:        cards,
		Point:        poker.GetPoint(cards),
		Type:         poker.GetCardsType(cards),
		BetAmount:    holdCards.BetAmount + betAmount,
		ActionPermit: holdCards.ActionPermit,
	}
}

// CheckGetPoker 检查要牌操作
func (user *User) CheckGetPoker(index int32) bool {
	holdCards := user.HoldCards[index]

	// 玩家状态为停止操作状态
	if user.Status == int32(msg.UserStatus_UserStopAction) {
		return false
	}

	// 牌组无操作权限
	if !holdCards.ActionPermit {
		return false
	}

	// 牌型不是普通牌
	if holdCards.Type != msg.CardsType_Other {
		return false
	}

	return true
}

// CheckDepartPoker 检查分牌操作
func (user *User) CheckDepartPoker(index int32) bool {
	holdCards := user.HoldCards[index]

	// 玩家状态为停止操作状态
	if user.Status == int32(msg.UserStatus_UserStopAction) {
		return false
	}

	// 牌组无操作权限
	if !holdCards.ActionPermit {
		return false
	}

	// 手牌数量不是2张
	if len(holdCards.Cards) != 2 {
		return false
	}

	// 没有分牌因子
	if user.DepartFactor == 0 {
		return false
	}

	if _, isPair := poker.IsPair(holdCards.Cards); !isPair {
		return false
	}

	// 筹码不足
	if user.CurAmount < user.HoldCards[index].BetAmount {
		return false
	}

	return true
}

// CheckDoubleBet 检查双倍操作
func (user *User) CheckDoubleBet(index int32) bool {
	holdCards := user.HoldCards[index]

	// 玩家状态为停止操作状态
	if user.Status == int32(msg.UserStatus_UserStopAction) {
		return false
	}

	// 牌组无操作权限
	if !holdCards.ActionPermit {
		return false
	}

	// 牌型不是普通牌
	if holdCards.Type != msg.CardsType_Other {
		return false
	}

	// 手牌数量不是2张
	if len(holdCards.Cards) != 2 {
		return false
	}

	// 筹码不足
	if user.CurAmount < user.HoldCards[index].BetAmount {
		return false
	}

	return true
}

// CheckStand 检查停牌操作
func (user *User) CheckStand(index int32) bool {
	holdCards := user.HoldCards[index]

	// 玩家状态为停止操作状态
	if user.Status == int32(msg.UserStatus_UserStopAction) {
		return false
	}

	// 牌组无操作权限
	if !holdCards.ActionPermit {
		return false
	}

	// 牌型不是普通牌
	if holdCards.Type != msg.CardsType_Other {
		return false
	}

	return true
}

// CheckGiveUp 检查认输操作
func (user *User) CheckGiveUp(index int32) bool {
	holdCards := user.HoldCards[index]

	// 玩家状态为停止操作状态
	if user.Status == int32(msg.UserStatus_UserStopAction) {
		return false
	}

	// 分牌之后不能认输
	if len(user.HoldCards[1].Cards) != 0 {
		return false
	}

	// 牌组无操作权限
	if !holdCards.ActionPermit {
		return false
	}

	// 要牌后不能再认输
	if len(holdCards.Cards) >= 3 {
		return false
	}

	return true
}

// PermitAction 操作权限变更
func (user *User) PermitAction(index int32) {
	if user.Status == int32(msg.UserStatus_UserStopAction) || user.HoldCards[index].Type != msg.CardsType_Other {
		user.HoldCards[index].ActionPermit = false
	}
}

// GetUserInterdata 获取依赖全局user的游戏数据
func GetUserInterdata(userInter player.PlayerInterface) (data UserData) {

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

// SetUserInterdata 设置依赖全局user的游戏数据
func SetUserInterdata(userInter player.PlayerInterface, data UserData) {

	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("json marshal userInter data fail: %v", err)
	}

	userInter.SetTableData(string(b))
}

// EndTypeToString 结束方式转字符串
func EndTypeToString(endType int) (endTypeStr string) {
	switch endType {
	case EndType_Unknow:
		endTypeStr = "未知"
		break
	case EndType_Blackjack:
		endTypeStr = "黑杰克"
		break
	case EndType_Boom:
		endTypeStr = "爆牌"
		break
	case EndType_GiveUp:
		endTypeStr = "认输"
		break
	case EndType_BankerWin:
		endTypeStr = "庄家大"
		break
	case EndType_BankerLoss:
		endTypeStr = "庄家小"
		break
	case EndType_draw:
		endTypeStr = "打平"
		break

	}
	return
}
