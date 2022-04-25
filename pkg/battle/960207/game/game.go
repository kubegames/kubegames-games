package game

import (
	"go-game-sdk/define"
	"go-game-sdk/lib/clock"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	conf "github.com/kubegames/kubegames-games/pkg/battle/960207/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960207/poker"
	"github.com/kubegames/kubegames-sdk/pkg/player"

	"github.com/bitly/go-simplejson"
)

type GeneralNiuniu struct {
	Table           table.TableInterface
	Chairs          map[int32]int64           // 玩家座位号
	UserList        map[int64]*data.User      // 所有的玩家列表
	Poker           *poker.GamePoker          // 牌堆
	TimerJob        *clock.Job                // job
	RobotTimer      *clock.Job                // 轮训机器人检测
	ControlledCards map[int64]poker.HoldCards // 控制的牌堆
	CardsSequence   []poker.HoldCards         // 牌组序列
	Status          int32                     // 游戏的状态
	TimeCfg         *conf.TimeConfig          // 时间配置
	GameCfg         *conf.GameConfig          // 游戏配置
	RoomCfg         *conf.RoomConfig          // 游戏配置
	RobotCfg        *conf.RobotConfig         // 机器人配置
	LoadCfg         bool                      // 是否已经加载配置
	ExpectNum       int                       // 期望人数
}

// GeneralNiuniuRoom 通比牛牛房间
type GeneralNiuniuRoom struct {
}

// 系统常量
const (

	// 作弊率来源
	ProbSourceRoom  = "血池" //  血池
	ProbSourcePoint = "点控" // 点控
)

// InitTable 初始化游戏房间
func (room *GeneralNiuniuRoom) InitTable(table table.TableInterface) {
	//log.Tracef("init table num %d", table.GetID())
	game := new(GeneralNiuniu)
	game.Table = table
	game.UserList = make(map[int64]*data.User)
	game.Status = int32(msg.GameStatus_ReadyStatus)
	game.ControlledCards = make(map[int64]poker.HoldCards)

	// 初始化座位座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
	table.Start(game, nil, nil)
}

// UserExit 用户退出游戏房间
func (room *GeneralNiuniuRoom) UserExit(userInter player.PlayerInterface) {

}

// InitConfig 加载配置文件
func (game *GeneralNiuniu) InitConfig() {
	// 加载房间配置
	confStr := game.Table.GetAdviceConfig()

	js, err := simplejson.NewJson([]byte(confStr))
	if err != nil {
		log.Errorf("读取游戏配置失败: %v", err)
	}
	betBase, _ := js.Get("Bottom_Pouring").Int64()

	RoomCfg := &conf.RoomConfig{
		RoomCost: betBase,
		MinLimit: game.Table.GetEntranceRestrictions(),
		TaxRatio: game.Table.GetRoomRate(),
	}

	game.RoomCfg = RoomCfg

	// 加载游戏配置；时间配置；控制配置
	bnnConf := conf.GeneralNiuniuConf

	game.TimeCfg = &bnnConf.TimeConfig
	game.GameCfg = &bnnConf.GameConfig

	// 记载机器人配置
	robotConf := conf.RobotConf
	robotConf.LoadRobotCfg()

	game.RobotCfg = &robotConf

	game.LoadCfg = true
}

// OnActionUserSitDown 用户坐下
func (game *GeneralNiuniu) OnActionUserSitDown(userInter player.PlayerInterface, orderIndex int, config string) int {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 进入房间 %d", userID, game.Table.GetID())

	// 用户不再玩家列表中，
	if _, ok := game.UserList[userID]; !ok {

		// 游戏中不能进入
		if game.Status != int32(msg.GameStatus_ReadyStatus) &&
			game.Status != int32(msg.GameStatus_StartMove) {
			log.Tracef("游戏中不能进入")
			return define.SIT_DOWN_ERROR_NORMAL
		}

		// 倒计时最后一秒不让进来
		if game.Status == int32(msg.GameStatus_StartMove) && game.TimerJob.GetTimeDifference() < 500 {
			log.Tracef("游戏 %d 最后0.5秒不能进入", game.Table.GetID())
			return define.SIT_DOWN_ERROR_NORMAL
		}

		////// 随机一个无人座位
		var chairID int32

		// 椅子个数
		chairSize := len(game.Chairs)

		// 随机椅子索引
		randChair := rand.RandInt(0, chairSize)

		i := 0

		for k, _ := range game.Chairs {
			if i == randChair {
				chairID = k
				break
			}
			i++
		}
		delete(game.Chairs, chairID)

		user := &data.User{
			ID:               userID,
			User:             userInter,
			Nick:             userInter.GetNike(),
			Head:             userInter.GetHead(),
			Status:           int32(msg.UserStatus_SitDown),
			CurAmount:        userInter.GetScore(),
			InitAmount:       userInter.GetScore(),
			ChairID:          chairID,
			HoldCards:        &poker.HoldCards{},
			ExactControlRate: userInter.GetProb(),
			ExitPermit:       true,
		}

		// 新玩家加入游戏列表
		game.UserList[userID] = user

	} else {

		// 断线用户重新登陆
		game.UserList[userID].ReConnect = true
		log.Tracef("重联，玩家 %d走到坐下", userID)
	}

	return define.SIT_DOWN_OK
}

func (game *GeneralNiuniu) BindRobot(ai player.RobotInterface) player.RobotHandler {

	robot := new(Robot)
	if game.RobotCfg == nil {
		log.Errorf("游戏 %v 第一个玩家为机器人", game)
	} else {
		userID := ai.GetID()
		user, ok := game.UserList[userID]
		if !ok {
			log.Errorf("获取玩家异常！！！！")
			robot.Init(ai, game, *game.RobotCfg, -1)
		} else {
			robot.Init(ai, game, *game.RobotCfg, user.ChairID)
		}
	}
	return robot
}

// SendScene 发送场景消息
func (game *GeneralNiuniu) SendScene(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	user, ok := game.UserList[userID]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
		return false
	}

	if !game.LoadCfg {
		game.InitConfig()
	}

	// 非断线重联广播玩家坐下
	if !game.UserList[userID].ReConnect {

		// 游戏已经开始，发送开始倒计时状态信息
		if game.Status == int32(msg.GameStatus_StartMove) {
			game.SendGameStatus(int32(msg.GameStatus_StartMove), int32(game.TimerJob.GetTimeDifference()/1000), userInter)
		}

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

	return true
}

// UserReady 用户准备
func (game *GeneralNiuniu) UserReady(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	log.Tracef("玩家 %d 在房间 %d 准备，游戏状态为 %d", userID, game.Table.GetID(), game.Status)

	user, ok := game.UserList[userInter.GetID()]
	if !ok {
		log.Errorf("获取玩家异常！！！！")
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
		if len(game.UserList) == 1 {
			if game.TimeCfg != nil {
				game.RobotTimer, _ = game.Table.AddTimer(int64(game.TimeCfg.DelayCheckMatch), game.RobotSit)
			} else {
				log.Errorf("用户准备时配置文件未加载完")
			}
		}
	}

	return true
}

// GameStart 框架询问是否开赛
func (game *GeneralNiuniu) GameStart() {
	if game.Status == int32(msg.GameStatus_ReadyStatus) {
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
			return true
		}
	}
	return false
}

// UserExit 用户离线
func (game *GeneralNiuniu) UserExit(userInter player.PlayerInterface) bool {

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
		userExitResult := msg.UserExitRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.SendUserExit(userExitResult)
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_StartMove) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_ReadyStatus)
		game.LoadCfg = false
	}
	return exitPermit
}

// LeaveGame 用户正常申请离开
func (game *GeneralNiuniu) LeaveGame(userInter player.PlayerInterface) bool {

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
		userExitResult := msg.UserExitRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.SendUserExit(userExitResult)
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_StartMove) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.UserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_ReadyStatus)
		game.LoadCfg = false
	}
	return exitPermit
}

// OnGameMessage 接受用户发送信息
func (game *GeneralNiuniu) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	log.Tracef(" 收到客户端消息 ：%d", subCmd)

	switch subCmd {

	// 下注请求
	case int32(msg.ReceiveMessageType_C2SBetChips):
		game.UserBetChips(buffer, userInter)
		break
		// 摊牌请求
	case int32(msg.ReceiveMessageType_C2SShowCards):
		game.UserShowCards(buffer, userInter)
		break
		// 要牌请求
	case int32(msg.ReceiveMessageType_C2SDemandCards):
		//game.UserDemandCards(buffer, userInter)
		break
	}
}

// ResetTable 重置桌子
func (game *GeneralNiuniu) ResetTable() {

	// 重置桌面属性
	game.Poker = nil
	game.TimeCfg = nil
	game.GameCfg = nil
	game.RoomCfg = nil
	game.LoadCfg = false
	game.CardsSequence = []poker.HoldCards{}
	game.ControlledCards = make(map[int64]poker.HoldCards)

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
}

func (game *GeneralNiuniu) CloseTable() {}
