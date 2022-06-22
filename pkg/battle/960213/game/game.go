package game

import (
	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960213/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// DouDizhu 斗地主逻辑
type DouDizhu struct {
	Table                 table.TableInterface
	Chairs                [3]*data.User        // 座位
	UserList              map[int64]*data.User // 所有的玩家列表
	CurrentPlayer         CurrentPlayer        // 当前可执行玩家
	curRobberChairID      int                  // 当前抢地主玩家座位ID
	RobChairList          [3]int               // 抢地主座位列表
	bottomCards           []byte               // 底牌
	Poker                 *poker.GamePoker     // 牌堆
	Dizhu                 *data.User           // 地主
	TimerJob              *table.Job           // job
	RobotTimer            *table.Job           // 轮训机器人检测
	Status                int32                // 游戏的状态
	TimeCfg               *config.TimeConfig   // 时间配置
	GameCfg               *config.GameConfig   // 游戏配置
	RoomCfg               *config.RoomConfig   // 游戏配置
	RobotCfg              *config.RobotConfig  // 机器人配置
	LoadCfg               bool                 // 是否已经加载配置
	CurrentCards          poker.HandCards      // 当前有牌权的牌组
	TipsCards             poker.HandCards      // 提示牌组
	LeftCards             []byte               // 剩余牌组
	StepCount             int                  // 操作计数器
	ControlledCards       map[int64][]byte     // 控制的牌堆
	RobCount              int                  // 抢分次数
	CurRobNum             int64                // 当前最高抢庄分数
	BottomMultiple        int64                // 底牌倍数
	BoomMultiple          int64                // 炸弹倍数
	RocketMultiple        int64                // 火箭倍数
	AllOffMultiple        int64                // 春天倍数
	BeAllOffMultiple      int64                // 反春倍数
	TotalPeasantsMultiple int64                // 农民总倍数
	InAnimation           bool                 // 是否动画时间中
}

type CurrentPlayer struct {
	UserID     int64 // 用户ID
	ChairID    int32 // 作为ID
	ActionTime int   // 行动时间
	Permission bool  // 是否有出牌权
	StepCount  int   // 操作计数器
	ActionType int32 // 操作类型
}

// DouDizhuRoom 斗地主房间
type DouDizhuRoom struct{}

// InitTable 初始化游戏房间
func (room *DouDizhuRoom) InitTable(table table.TableInterface) {

	game := new(DouDizhu)
	game.Table = table
	game.UserList = make(map[int64]*data.User)
	game.Status = int32(msg.GameStatus_GameInitStatus)
	// 初始化控制牌组
	game.ControlledCards = make(map[int64][]byte)

	table.Start(game, nil, nil)
}

// UserExit 用户退出游戏房间
func (room *DouDizhuRoom) UserExit(userInter player.PlayerInterface) {

}

// InitConfig 加载配置文件
func (game *DouDizhu) InitConfig() {
	// 加载房间配置
	confStr := game.Table.GetAdviceConfig()

	js, err := simplejson.NewJson([]byte(confStr))
	if err != nil {
		log.Errorf("读取游戏配置失败: %v", err)
	}
	betBase, _ := js.Get("Bottom_Pouring").Int64()

	RoomCfg := &config.RoomConfig{
		RoomCost: betBase,
		MinLimit: game.Table.GetEntranceRestrictions(),
		TaxRate:  game.Table.GetRoomRate(),
		Level:    game.Table.GetLevel(),
	}

	game.RoomCfg = RoomCfg

	// 加载游戏配置；时间配置；游戏配置；机器人配置
	game.TimeCfg = &config.DoudizhuConf.TimeConfig
	game.GameCfg = &config.DoudizhuConf.GameConfig
	game.RobotCfg = &config.RobotConf

	game.LoadCfg = true
}

// OnActionUserSitDown 用户坐下
func (game *DouDizhu) OnActionUserSitDown(userInter player.PlayerInterface, orderIndex int, config string) table.MatchKind {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 进入房间 %d", userID, game.Table.GetID())

	// 用户不再玩家列表中，
	if _, ok := game.UserList[userID]; !ok {

		// 游戏中不能进入
		if game.Status != int32(msg.GameStatus_GameInitStatus) {
			return table.SitDownErrorNomal
		}

		// 获取一个空座位
		chairID := game.GetEmptyChair()
		if chairID < 0 {
			log.Warnf("游戏 %d 玩家 %d 获取座位失败", game.Table.GetID(), userID)
			return table.SitDownErrorOver
		}

		user := &data.User{
			ID:               userID,
			User:             userInter,
			Nick:             userInter.GetNike(),
			Head:             userInter.GetHead(),
			Status:           int32(msg.UserStatus_UserNormal),
			CurAmount:        userInter.GetScore(),
			InitAmount:       userInter.GetScore(),
			ChairID:          int32(chairID),
			ExactControlRate: userInter.GetProb(),
			ExitPermit:       true,
		}

		// 新玩家加入游戏列表
		game.UserList[userID] = user

		// 新玩家加入游戏列表
		game.Chairs[chairID] = user

		// 加入玩家是机器人加载机器人配置
		// if userInter.IsRobot() {
		// 	robot := new(Robot)
		// 	robotUser := userInter.BindRobot(robot)
		// 	if game.RobotCfg == nil {
		// 		log.Errorf("游戏 %v 第一个玩家为机器人", game)
		// 	} else {
		// 		robot.Init(robotUser, game, *game.RobotCfg)

		// 	}
		// }
	} else {
		// 断线用户重新登陆
		game.UserList[userID].ReConnect = true
	}
	return table.SitDownOk
}

//BindRobot 绑定机器人
func (game *DouDizhu) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	if game.RobotCfg == nil {
		log.Errorf("游戏 %v 第一个玩家为机器人", game)
	} else {
		robot.Init(ai, game, *game.RobotCfg)
	}
	return robot
}

// SendScene 发送场景消息
func (game *DouDizhu) SendScene(userInter player.PlayerInterface) {
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	// 第一个玩家进入加载配置文件
	if !game.LoadCfg {
		game.InitConfig()
	}

	// todo 断线重连玩家是否取消托管
	if user.ReConnect {

	}

	// 发送场景消息
	game.SendSceneInfo(userInter, game.UserList[userID].ReConnect)

	if game.Status >= int32(msg.GameStatus_SettleStatus) {
		game.SendSettleInfo()
	}

	game.UserList[userID].ReConnect = false
	return
}

// UserReady 用户准备
func (game *DouDizhu) UserReady(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 在房间 %d 准备，游戏状态为 %d", userID, game.Table.GetID(), game.Status)

	//game.UserList[userID].Status = int32(msg.UserStatus_UserNormal)
	// 第一个玩家进入，预加载机器人
	if game.RobotTimer == nil {
		// 满桌时间
		fullTableTime := 1
		// 满桌时间权重
		fullTableWeight := rand.RandInt(0, 101)
		if game.GameCfg == nil {
			log.Errorf("第一个玩家准备时配置文件加载失败")
			return false
		}
		lastRate := 0
		for index, rate := range game.GameCfg.FullTableTimeRate {
			if fullTableWeight > lastRate && fullTableWeight <= rate {
				fullTableTime = index + 1
			}
			lastRate = rate
		}
		game.RobotTimer, _ = game.Table.AddTimer(int64(fullTableTime*1000), game.RobotSitCheck)
	}

	return true
}

// GameStart 框架询问是否开赛
func (game *DouDizhu) GameStart() {

	if len(game.UserList) == 3 && game.Status == int32(msg.GameStatus_GameInitStatus) {
		allReady := true
		for _, user := range game.UserList {
			if user.Status != int32(msg.UserStatus_UserNormal) {
				allReady = false
			}
		}

		if allReady {
			game.Start()
			return
		}
	}
}

// UserOffline 用户离线
func (game *DouDizhu) UserOffline(userInter player.PlayerInterface) bool {

	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
	}

	exitPermit := user.ExitPermit
	log.Tracef("用户 %d 退出，退出权限 %v", userID, exitPermit)

	if exitPermit {

		// 游戏列表删除用户
		delete(game.UserList, userID)

		// 让出座位
		game.Chairs[user.ChairID] = nil

	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {

		game.LoadCfg = false
		switch game.Status {
		// 游戏已经结束，重置桌子状态
		case int32(msg.GameStatus_GameOver):
			game.Status = int32(msg.GameStatus_GameInitStatus)
			break

			// 游戏还未开始，停下所有定时器
		case int32(msg.GameStatus_GameInitStatus):
			log.Tracef("重置了定时器")
			if game.TimerJob != nil {
				game.Table.DeleteJob(game.TimerJob)
				game.TimerJob = nil

			}

			if game.RobotTimer != nil {
				game.Table.DeleteJob(game.RobotTimer)
				game.RobotTimer = nil
			}
			break
		}

	}
	return exitPermit
}

// UserLeaveGame 用户正常申请离开
func (game *DouDizhu) UserLeaveGame(userInter player.PlayerInterface) bool {

	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
	}

	exitPermit := user.ExitPermit
	log.Tracef("用户 %d 退出，退出权限 %v", userID, exitPermit)

	if exitPermit {

		// 游戏列表删除用户
		delete(game.UserList, userID)

		// 让出座位
		game.Chairs[user.ChairID] = nil

	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {

		game.LoadCfg = false
		switch game.Status {
		// 游戏已经结束，重置桌子状态
		case int32(msg.GameStatus_GameOver):
			game.Status = int32(msg.GameStatus_GameInitStatus)
			break

			// 游戏还未开始，停下所有定时器
		case int32(msg.GameStatus_GameInitStatus):
			log.Tracef("重置了定时器")
			if game.TimerJob != nil {
				game.Table.DeleteJob(game.TimerJob)
				game.TimerJob = nil

			}

			if game.RobotTimer != nil {
				game.Table.DeleteJob(game.RobotTimer)
				game.RobotTimer = nil
			}
			break
		}

	}
	return exitPermit
}

// OnGameMessage 接受用户发送信息
func (game *DouDizhu) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	switch subCmd {
	// 抢地主请求
	case int32(msg.ReceiveMessageType_C2SRob):
		game.UserRobDizhu(buffer, userInter)
		break
	// 加倍请求
	case int32(msg.ReceiveMessageType_C2SRedouble):
		game.UserRedouble(buffer, userInter)
		break
	// 提示请求
	case int32(msg.ReceiveMessageType_C2STips):
		game.UserGetTips(buffer, userInter)
		break
	// 出牌请求
	case int32(msg.ReceiveMessageType_C2SPutCards):
		game.UserPutCards(buffer, userInter)
		break
	// 托管请求
	case int32(msg.ReceiveMessageType_C2SHangUp):
		game.UserHangUp(buffer, userInter)
		break
	// 配牌请求
	case int32(msg.ReceiveMessageType_C2SDemandCards):
		//game.UserDemandCards(buffer, userInter)
		break
	// 清桌请求
	case int32(msg.ReceiveMessageType_C2SClean):
		//game.UserClean()
		break
	}
}

// ResetTable 重置桌子
func (game *DouDizhu) ResetTable() {}

func (game *DouDizhu) CloseTable() {}
