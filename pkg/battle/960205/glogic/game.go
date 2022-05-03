package glogic

import (
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/kubegames/kubegames-games/pkg/battle/960205/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960205/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type ErBaGangGame struct {
	InterTable            table.TableInterface // 桌子
	DiZhu                 int                  // 游戏底注
	UserAllList           map[int]*User        // 玩家列表
	UserIndex             int                  // 玩家索引
	GameCount             int                  // 游戏局数
	RobZhuangConfList     map[int][]int32      // 用户抢庄配置列表 [用户id] = []抢庄配置
	RobZhuangMultipleList map[int]int64        // 用户抢庄倍数列表 [用户id] = 抢庄倍数
	RobZhuangIndex        int                  // 庄家的索引
	BetConfList           map[int][]int32      // 用户下注配置列表 [用户id] = []下注配置
	BetMultipleList       map[int]int64        // 用户下注倍数列表 [用户id] = 抢庄倍数
	DiceNumberArr         []int32              // 骰子数组
	PaiList               map[int][]int32      // 发给用户的牌的列表 [用户索引] = []牌组数据
	GamePoker             poker.GamePoker      // 牌
	BtnCount              int                  // 按钮计数
	TimerJob              *table.Job           //job
	State                 STATE
	Cards                 []int32 //所有牌出现的次数
	Dismiss               bool
	IsEndToStart          bool
}

type STATE int

const (
	Game_End        STATE = iota
	Game_Start      STATE = 1
	Game_Zhuang     STATE = 2
	Game_Zhuang_End STATE = 3
	Game_Bet        STATE = 4
	Game_Bet_End    STATE = 5
	Game_DealCard   STATE = 6
	Game_OpenCard   STATE = 7
	Game_Count      STATE = 8
)

var (
	// 初始化有序4用户列表
	SendUserAllList = []*msg.SeatUserInfoRes{nil, nil, nil, nil}
)

// InitTable 初始化桌子
func (game *ErBaGangGame) InitTable(table table.TableInterface) {
	log.Tracef("======= %s =======", "初始化桌子")

	game.InterTable = table
	game.UserIndex = 0 // 将玩家索引置为0
	game.GameCount = 0 // 游戏局数置为0
	game.DiZhu = 0
	game.State = Game_End
	game.Dismiss = false

	game.UserAllList = make(map[int]*User) // 初始化玩家列表
	//game.RobZhuangConfList = make(map[int][]int32)   // 初始化用户抢庄配置列表
	game.RobZhuangMultipleList = make(map[int]int64) // 初始化用户抢庄倍数列表
	//game.BetConfList = make(map[int][]int32)         // 初始化用户下注列表
	game.BetMultipleList = make(map[int]int64) // 初始化用户下注倍数列表
	//game.PaiList = make(map[int][]int32)             // 初始化牌列表
	game.Cards = make([]int32, 10)
	game.Reset()

	game.GamePoker.InitPoker()
	game.GamePoker.ShuffleCards()
}

// GameStart 游戏开始
func (game *ErBaGangGame) GameStart() {
	log.Tracef("游戏开始")
}

// OnActionUserSitDown 用户坐下
// 但是没有在桌子上
func (game *ErBaGangGame) OnActionUserSitDown(userInter player.PlayerInterface, chairId int, config string) table.MatchKind {
	log.Tracef("用户坐下")
	if game.DiZhu == 0 {
		jsonObj, err := gjson.DecodeToJson([]byte(config))
		if err != nil {
			log.Errorf("获取底注配置出错:%s", err.Error())
		}
		// 游戏底注
		game.DiZhu = jsonObj.Get("Bottom_Pouring").Int()
	}

	entranceRestrictions := game.InterTable.GetEntranceRestrictions()
	if entranceRestrictions != -1 && userInter.GetScore() < entranceRestrictions {
		return table.SitDownErrorOver
	}

	if game.checkReline(userInter) {
		//user.IsDeposit = false
		return table.SitDownOk
	}
	if game.Dismiss {
		return table.SitDownErrorNomal
	}
	if game.checkUser(userInter) {
		return table.SitDownErrorOver
	}

	if len(game.UserAllList) < 4 {
		uInfo := &User{
			InterUser: userInter,
		}
		game.UserAllList[chairId] = uInfo
		//game.UserIndex = chairId
	} else {
		//game.UserIndex = 0
		return table.SitDownErrorNomal
	}

	log.Tracef("用户列表:%v,用户个数是:%d", game.UserAllList, len(game.UserAllList))
	return table.SitDownOk
}

func (game *ErBaGangGame) checkUser(userInter player.PlayerInterface) bool {
	for _, v := range game.UserAllList {
		if v != nil && v.InterUser.GetID() == userInter.GetID() {
			return true
		}
	}
	return false
}

func (game *ErBaGangGame) checkReline(userInter player.PlayerInterface) bool {
	user := game.UserAllList[userInter.GetChairID()]
	if user != nil && user.InterUser.GetID() == userInter.GetID() {
		return true
	}
	return false
}

func (game *ErBaGangGame) UserLeaveGame(userInter player.PlayerInterface) bool {
	if game.State != Game_End {
		return false
	}
	delete(game.UserAllList, userInter.GetChairID())
	return true
}

// OnGameMessage 接收客户端发来的消息
func (game *ErBaGangGame) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(msg.ReceiveMessageType_C2SRobZhuangEnd): // 用户按下抢庄按钮
		game.ReceiveMsgRobBtnEnd(buffer, user)
	case int32(msg.ReceiveMessageType_C2SUserBetEnd):
		game.ReceiveMsgBetBtnEnd(buffer, user)
	case int32(msg.ReceiveMessageType_C2STest):
		game.test(buffer, user)
	case int32(msg.ReceiveMessageType_C2SDeposit):
		game.deposit(buffer, user)
	}
}

//BindRobot 绑定机器人
func (game *ErBaGangGame) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	robot.Init(ai, game)
	return robot
}

// SendScene 发送场景消息
func (game *ErBaGangGame) SendScene(userInter player.PlayerInterface) {
	log.Tracef("发送场景消息")
	// if userInter.IsRobot() {
	// 	robot := new(Robot)
	// 	robotUser := userInter.BindRobot(robot)
	// 	robot.Init(robotUser, game)
	// }
	if len(game.UserAllList) == 4 {
		game.SendMsgUSitSetDown(userInter.GetChairID())
	}
	if len(game.UserAllList) == 1 {
		game.InterTable.AddTimer(5*1000, func() {
			if len(game.UserAllList) > 0 && len(game.UserAllList) < 4 {
				// 匹配机器人
				if err := game.InterTable.GetRobot(uint32(4-len(game.UserAllList)), game.InterTable.GetConfig().RobotMinBalance, game.InterTable.GetConfig().RobotMaxBalance); err != nil {
					log.Errorf("分配机器人失败:%s", err.Error())
				}
				//game.SendMsgUSitSetDown()
			}
		})
	}
	return
}

// UserReady 用户准备好了
// 坐在桌子上
func (game *ErBaGangGame) UserReady(userInter player.PlayerInterface) bool {
	return true
}

// ResetTable 重新开始桌子
func (game *ErBaGangGame) ResetTable() {
	log.Tracef("重新开始桌子")
	game.InitTable(game.InterTable)
}

func (game *ErBaGangGame) CloseTable() {

}

// UserOffline 用户掉线
func (game *ErBaGangGame) UserOffline(userInter player.PlayerInterface) bool {
	log.Tracef("用户退出")
	if game.State != Game_End {
		user := game.UserAllList[userInter.GetChairID()]
		if user != nil {
			user.IsDeposit = true
		}
		return false
	}
	delete(game.UserAllList, userInter.GetChairID())
	//if len(game.UserAllList) == 0 {
	//	game.SendMsgBigCloseAnAccount()
	//}
	//for _, v := range game.UserAllList {
	//	game.InterTable.KickOut(v.InterUser)
	//}
	log.Tracef("用户退出后的用户列表是:%v,用户个数是:%d", game.UserAllList, len(game.UserAllList))
	return true
}
