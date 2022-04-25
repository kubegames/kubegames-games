package game

import (
	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960204/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// RunFaster 跑得快逻辑
type RunFaster struct {
	Table           table.TableInterface
	Chairs          map[int32]int64      // 玩家座位号
	UserList        map[int64]*data.User // 所有的玩家列表
	CurrentPlayer   CurrentPlayer        // 当前可执行玩家
	Poker           *poker.GamePoker     // 牌堆
	TimerJob        *table.Job           // job
	RobotTimer      *table.Job           // 轮训机器人检测
	Status          int32                // 游戏的状态
	TimeCfg         *config.TimeConfig   // 时间配置
	GameCfg         *config.GameConfig   // 游戏配置
	RoomCfg         *config.RoomConfig   // 游戏配置
	RobotCfg        *config.RobotConfig  // 机器人配置
	LoadCfg         bool                 // 是否已经加载配置
	CurrentCards    poker.HandCards      // 当前有牌权的牌组
	TipsCards       poker.HandCards      // 提示牌组
	Seats           [3]int64             // 座位
	LeftCards       []byte               // 剩余牌组
	StepCount       int                  // 操作计数器
	ControlledCards map[int64][]byte     // 控制的牌堆
	ControlList     []int64              // 控制列表, 按顺序来代表获得从大到小的手牌
	PutCardsLog     string               // 出牌日志
}

type CurrentPlayer struct {
	UserID     int64 // 用户ID
	ChairID    int32 // 作为ID
	ActionTime int   // 行动时间
	Permission bool  // 是否有出牌权
	StepCount  int   // 操作计数器
	ActionType int32 // 操作类型
	IsFinalEnd bool  // 是否最后一手牌可终结游戏
}

// RunFasterRoom 跑得快房间
type RunFasterRoom struct {
}

// 系统常量
const (

	// 作弊率来源
	ProbSourceRoom  = "血池" //  血池
	ProbSourcePoint = "点控" // 点控
)

// InitTable 初始化游戏房间
func (room *RunFasterRoom) InitTable(table table.TableInterface) {
	//log.Tracef("init table num %d", table.GetID())
	game := new(RunFaster)
	game.Table = table
	game.UserList = make(map[int64]*data.User)
	game.Status = int32(msg.GameStatus_GameInitStatus)
	game.ControlledCards = make(map[int64][]byte)

	// 初始化座位座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
	}
	table.Start(game, nil, nil)
}

// InitConfig 加载配置文件
func (game *RunFaster) InitConfig() {
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
	game.TimeCfg = &config.RunFasterConf.TimeConfig
	game.GameCfg = &config.RunFasterConf.GameConfig
	game.RobotCfg = &config.RobotConf

	game.LoadCfg = true
}

// OnActionUserSitDown 用户坐下
func (game *RunFaster) OnActionUserSitDown(userInter player.PlayerInterface, orderIndex int, config string) table.MatchKind {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 进入房间 %d", userID, game.Table.GetID())

	// 用户不再玩家列表中，
	if _, ok := game.UserList[userID]; !ok {

		// 游戏中不能进入
		if game.Status != int32(msg.GameStatus_GameInitStatus) {
			return table.SitDownErrorNomal
		}

		////// 随机一个无人座位
		var chairID int32

		// 椅子个数
		chairSize := len(game.Chairs)

		// 随机椅子索引
		randChair := rand.RandInt(0, chairSize)

		i := 0

		for k := range game.Chairs {
			if i == randChair {
				chairID = k
				break
			}
			i++
		}
		delete(game.Chairs, chairID)

		game.Seats[chairID] = userID

		user := &data.User{
			ID:               userID,
			User:             userInter,
			Nick:             userInter.GetNike(),
			Head:             userInter.GetHead(),
			Status:           int32(msg.UserStatus_UserInitStatus),
			CurAmount:        userInter.GetScore(),
			InitAmount:       userInter.GetScore(),
			ChairID:          chairID,
			ExactControlRate: userInter.GetProb(),
			ExitPermit:       true,
		}

		// 新玩家加入游戏列表
		game.UserList[userID] = user

	} else {

		// 断线用户重新登陆
		game.UserList[userID].ReConnect = true
	}

	return table.SitDownOk
}

//BindRobot 绑定机器人
func (game *RunFaster) BindRobot(ai player.RobotInterface) player.RobotHandler {
	// 加入玩家是机器人加载机器人配置
	robot := new(Robot)
	if game.RobotCfg == nil {
		log.Errorf("游戏 %v 第一个玩家为机器人", game)
	} else {
		robot.Init(ai, game, *game.RobotCfg)
	}
	return robot
}

// SendScene 发送场景消息
func (game *RunFaster) SendScene(userInter player.PlayerInterface) {
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

	// 非断线重联广播玩家坐下
	if !game.UserList[userID].ReConnect {
		userSitDownResp := msg.UserSitDownRes{
			UserId:   userID,
			ChairId:  user.ChairID,
			UserName: user.Nick,
			Head:     user.Head,
			Coin:     user.CurAmount,
			Sex:      user.User.GetSex(),
			Address:  user.User.GetCity(),
		}
		game.SendUserSitDown(userSitDownResp)
	}

	// 发送场景消息
	game.SendSceneInfo(userInter, game.UserList[userID].ReConnect)

	game.UserList[userID].ReConnect = false

	//用户状态改变为正常
	game.UserList[userID].Status = int32(msg.UserStatus_UserNormal)
	return
}

// UserReady 用户准备
func (game *RunFaster) UserReady(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	_, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return false
	}
	log.Tracef("玩家 %d 在房间 %d 准备，游戏状态为 %d", userID, game.Table.GetID(), game.Status)

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
func (game *RunFaster) GameStart() {

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

	return
}

// UserExit 用户离线
func (game *RunFaster) UserOffline(userInter player.PlayerInterface) bool {
	user, ok := game.UserList[userInter.GetID()]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return true
	}

	//设置断开链接
	game.UserList[userInter.GetID()].ReConnect = true

	exitPermit := user.ExitPermit
	log.Tracef("用户 %d 退出，退出权限 %v", userInter.GetID(), exitPermit)

	if exitPermit {

		// 游戏列表删除用户
		delete(game.UserList, userInter.GetID())

		// 让出座位
		game.Chairs[user.ChairID] = 0
		game.Seats[user.ChairID] = 0

		// 广播玩家离开信息
		userExitResult := msg.UserExitRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.SendUserExitInfo(userExitResult)
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
func (game *RunFaster) UserLeaveGame(userInter player.PlayerInterface) bool {

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
		game.Chairs[user.ChairID] = 0
		game.Seats[user.ChairID] = 0

		// 广播玩家离开信息
		userExitResult := msg.UserExitRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.SendUserExitInfo(userExitResult)
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
func (game *RunFaster) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	log.Tracef(" 收到客户端消息 ：%d", subCmd)

	switch subCmd {
	// 通知请求
	case int32(msg.ReceiveMessageType_C2STips):
		game.UserGetTips(buffer, userInter)
		break
		// 出牌请求
	case int32(msg.ReceiveMessageType_C2SPutCards):
		game.UserPutCards(buffer, userInter, false)
		break
		// 出牌请求
	case int32(msg.ReceiveMessageType_C2SCancelHangUp):
		game.UserCancelHangUp(buffer, userInter)
		break
		// 配牌请求
	case int32(msg.ReceiveMessageType_C2SDemandCards):
		//game.UserDemandCards(buffer, userInter)
		break
	}
}

// ResetTable 重置桌子
func (game *RunFaster) ResetTable() {}

func (game *RunFaster) CloseTable() {}
