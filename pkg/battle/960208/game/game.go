package game

import (
	"github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960208/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// ThreeDollRoom 三公房间
type ThreeDollRoom struct {
}

// ThreeDoll 三公
type ThreeDoll struct {
	Table           table.TableInterface      // 游戏桌子接口
	Chairs          map[int32]int64           // 玩家座位号
	TimerJob        *table.Job                // 主流程定时器
	RobotTimer      *table.Job                // 坐下机器人定时器
	BetSeats        []int64                   // 下注玩家序列
	Status          int32                     // 游戏的状态
	UserList        map[int64]*data.User      // 玩家列表
	Poker           *poker.GamePoker          // 牌堆
	Banker          *data.User                // 庄家
	ControlledCards map[int64]poker.HoldCards // 控制的牌堆
	CardsSequence   []poker.HoldCards         // 牌组序列(从小到大)
	GameCfg         *config.GameConfig        // 时间配置
	TimeCfg         *config.TimeConfig        // 时间配置
	RoomCfg         *config.RoomConfig        // 房间配置
	RobotCfg        *config.RobotConfig       // 机器人配置
	LoadCfg         bool                      // 是否已经加载配置
	ExpectNum       int                       // 期望人数
}

// 系统常量
const (

	// 作弊率来源
	ProbSourceRoom  = "血池" //  血池
	ProbSourcePoint = "点控" // 点控
)

// InitTable 初始化游戏房间
func (room *ThreeDollRoom) InitTable(table table.TableInterface) {
	game := new(ThreeDoll)
	game.Table = table
	game.UserList = make(map[int64]*data.User)
	game.Status = int32(msg.GameStatus_GameReady)
	game.ControlledCards = make(map[int64]poker.HoldCards)

	// 初始化座位座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}
	table.Start(game, nil, nil)
}

// InitConfig 加载配置
func (game *ThreeDoll) InitConfig() {

	// 加载房间配置
	confStr := game.Table.GetAdviceConfig()

	js, err := simplejson.NewJson([]byte(confStr))
	if err != nil {
		log.Errorf("读取游戏配置失败: %v", err)
		return
	}

	betBase, _ := js.Get("Bottom_Pouring").Int64()

	roomCfg := &config.RoomConfig{
		RoomCost: betBase,
		MinLimit: game.Table.GetEntranceRestrictions(),
		TaxRate:  game.Table.GetRoomRate(),
	}

	game.RoomCfg = roomCfg

	// 加载游戏配置；时间配置；控制配置
	tdConf := config.ThreeDollConf

	// 记载机器人配置
	robotConf := config.RobotConf

	game.GameCfg = &tdConf.GameConf
	game.TimeCfg = &tdConf.TimeConf
	game.RobotCfg = &robotConf
	game.LoadCfg = true
}

// OnActionUserSitDown 用户坐下
func (game *ThreeDoll) OnActionUserSitDown(userInter player.PlayerInterface, orderIndex int, config string) table.MatchKind {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 进入房间 %d", userID, game.Table.GetID())

	// 不是重联玩家
	if _, ok := game.UserList[userID]; !ok {

		// 游戏中不能进入/倒计时最后一秒不能进入
		if game.Status >= int32(msg.GameStatus_StartGame) ||
			(game.TimerJob != nil && game.Status == int32(msg.GameStatus_CountDown) && game.TimerJob.GetTimeDifference() < 500) {
			log.Tracef("游戏中不能进入")
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

		user := &data.User{
			ID:         userID,
			User:       userInter,
			Nick:       userInter.GetNike(),
			Head:       userInter.GetHead(),
			Status:     int32(msg.UserStatus_SitDown),
			CurAmount:  userInter.GetScore(),
			InitAmount: userInter.GetScore(),
			ChairID:    chairID,
			ExitPermit: true,
		}

		// 新玩家加入游戏列表
		game.UserList[userID] = user

	} else {

		// 断线用户重新登陆
		game.UserList[userID].ReConnect = true
	}

	return table.SitDownOk
}

func (game *ThreeDoll) BindRobot(ai player.RobotInterface) player.RobotHandler {
	robot := new(Robot)
	if game.RobotCfg == nil {
		log.Errorf("游戏 %v 第一个玩家为机器人", game)
	} else {
		userID := ai.GetID()
		user, ok := game.UserList[userID]
		if !ok {
			log.Warnf("获取玩家异常！！！！")
			robot.Init(ai, game, *game.RobotCfg, user.ChairID)
		} else {
			robot.Init(ai, game, *game.RobotCfg, user.ChairID)
		}
	}
	return robot
}

// SendScene 玩家匹配之后调用这个发送场景消息
func (game *ThreeDoll) SendScene(userInter player.PlayerInterface) {
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return
	}

	// 首个玩家坐下，加载配置文件
	if !game.LoadCfg {
		game.InitConfig()
	}

	// 非断线重联广播玩家坐下
	if !game.UserList[userID].ReConnect {

		// 玩家在游戏倒计时状态进入房间 通知玩家剩余倒计时时间
		if game.Status == int32(msg.GameStatus_CountDown) && game.TimerJob != nil {
			game.SendGameStatus(int32(msg.GameStatus_CountDown), int32(game.TimerJob.GetTimeDifference()/1000), userInter)
		}

		// 广播玩家坐下
		userSitDownResp := msg.UserSitDownRes{
			UserId:   userID,
			ChairId:  user.ChairID,
			UserName: user.Nick,
			Head:     user.Head,
			Coin:     user.CurAmount,
			Address:  user.User.GetCity(),
			Sex:      user.User.GetSex(),
		}
		game.SendUserSitDown(userSitDownResp)

	}

	// 发送场景消息
	game.SendSceneInfo(userInter, game.UserList[userID].ReConnect)

	return
}

// UserReady 用户准备
func (game *ThreeDoll) UserReady(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 在房间 %d 准备，游戏状态为 %d", userID, game.Table.GetID(), game.Status)

	user, ok := game.UserList[userInter.GetID()]
	if !ok {
		log.Errorf("玩家准备时获取玩家异常！！！！")
		return false
	}

	// 玩家重联
	if game.UserList[userID].ReConnect {

		game.UserList[userID].ReConnect = false
	} else {
		// 玩家已经准备
		if user.Status == int32(msg.UserStatus_Ready) {
			return false
		}

		// 用户坐下变准备
		if user.Status == int32(msg.UserStatus_SitDown) {
			user.Status = int32(msg.UserStatus_Ready)
		}

		// 一个玩家准备后三秒内没进入玩家匹配机器人
		if len(game.UserList) == 1 && game.TimeCfg != nil {
			game.RobotTimer, _ = game.Table.AddTimer(int64(game.TimeCfg.RobotSitCheck), game.RobotSit)
		}

	}

	return true
}

// GameStart 玩家准备后调用是否开赛
func (game *ThreeDoll) GameStart() {
	if game.Status == int32(msg.GameStatus_GameReady) {
		// 准备玩家人数
		readyUserCount := 0

		for _, user := range game.UserList {
			if user.Status == int32(msg.UserStatus_Ready) {
				readyUserCount++
			}
		}

		// 准备人数至少有两人就开始游戏
		if readyUserCount >= 2 {
			game.Start()
			return
		}
	}
	return
}

// UserOffline 用户离线
func (game *ThreeDoll) UserOffline(userInter player.PlayerInterface) bool {

	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return false
	}

	exitPermit := user.ExitPermit
	log.Tracef("用户 %d 退出，退出权限 %v", userID, exitPermit)

	if exitPermit {
		// 游戏列表删除用户
		delete(game.UserList, userID)

		// 让出座位
		game.Chairs[user.ChairID] = 0

		// 广播玩家离开信息
		//userExitResult := msg.UserExitRes{
		//	UserId:  user.ID,
		//	ChairId: user.ChairID,
		//}
		//game.SendUserExit(userExitResult)
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_CountDown) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_GameReady)
		game.LoadCfg = false
	}
	return exitPermit
}

// UserLeaveGame 用户正常申请离开
func (game *ThreeDoll) UserLeaveGame(userInter player.PlayerInterface) bool {

	userID := userInter.GetID()

	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return false
	}

	exitPermit := user.ExitPermit
	log.Tracef("用户 %d 退出，退出权限 %v", userID, exitPermit)

	if exitPermit {
		// 游戏列表删除用户
		delete(game.UserList, userID)

		// 让出座位
		game.Chairs[user.ChairID] = 0

		// 广播玩家离开信息
		//userExitResult := msg.UserExitRes{
		//	UserId:  user.ID,
		//	ChairId: user.ChairID,
		//}
		//game.SendUserExit(userExitResult)
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_CountDown) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_GameReady)
		game.LoadCfg = false
	}
	return exitPermit
}

// OnGameMessage 接受用户发送信息
func (game *ThreeDoll) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	log.Debugf(" 收到客户端消息 ：%d", subCmd)

	switch subCmd {
	// 抢庄请求
	case int32(msg.ReceiveMessageType_C2SRobBanker):
		game.UserRobBanker(buffer, userInter)
		break
		// 投注请求
	case int32(msg.ReceiveMessageType_C2SBetChips):
		game.UserBetChips(buffer, userInter)
		break
		// 摊牌请求
	case int32(msg.ReceiveMessageType_C2SShowCards):
		game.UserShowCards(buffer, userInter)
		break
		// 配牌请求
	case int32(msg.ReceiveMessageType_C2SDemandCards):
		//game.UserDemandCards(buffer, userInter)
		break
		// 拉去战绩请求
	case int32(msg.ReceiveMessageType_C2SPullRecords):
		game.UserPullRecords(buffer, userInter)
		break
	}

}

// ResetTable 重置桌子
func (game *ThreeDoll) ResetTable() {

}

func (game *ThreeDoll) CloseTable() {

}
