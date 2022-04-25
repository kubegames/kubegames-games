package game

import (
	simplejson "github.com/bitly/go-simplejson"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960101/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

// BlackjackRoom 黑杰克房间
type BlackjackRoom struct {
}

// InitTable 初始化游戏房间
func (room *BlackjackRoom) InitTable(table table.TableInterface) {
	game := new(Blackjack)
	game.table = table

	game.AllUserList = make(map[int64]*data.User)
	game.UserList = make(map[int64]*data.User)
	game.Status = int32(msg.GameStatus_StartStatus)

	// 初始化座位座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}

	//桌子启动
	table.Start(game, nil, nil)
}

// Blackjack 黑杰克
type Blackjack struct {
	table         table.TableInterface // table interface
	AllUserList   map[int64]*data.User // 所有的玩家列表
	TimerJob      *table.Job           // job
	BetSeats      []int64              // 下注玩家序列
	Status        int32                // 游戏的状态
	UserList      map[int64]*data.User // 下注玩家列表，为下注的玩家不进入游戏
	CurActionUser *CurUser             // 当前发言玩家
	Poker         *poker.GamePoker     // 牌堆
	HostCards     *HostCards           // 庄家手牌
	TurnCounter   int                  // 玩家轮转操作计数器
	ExpectNum     int                  // 期望人数
	Chairs        map[int32]int64      // 玩家座位号
	GameCfg       *config.GameConfig   // 时间配置
	timeCfg       *config.TimeConfig   // 时间配置
	RoomCfg       *config.RoomConfig   // 房间配置
	ControlCfg    *config.ExactControl // 控制配置
	RobotCfg      *config.RobotConfig  // 机器人配置
	LoadCfg       bool                 // 是否已经加载配置
}

// HostCards 庄家手牌
type HostCards struct {
	PocketCard byte          // 暗牌
	Cards      []byte        // 明牌
	Point      []int32       // 点数
	Type       msg.CardsType // 牌类型
}

// CurUser 当前操作玩家数据信息
type CurUser struct {
	UserID        int64
	ChairID       int32
	BetCardsIndex int32
	GetPoker      bool
	DepartPoker   bool
	DoubleBet     bool
	Stand         bool
	GiveUp        bool
	TurnCounter   int
	ActionSwitch  bool
}

// 系统常量
const (

	// 系统角色
	SysRolePlayer = "玩家"  // 玩家
	SysRoleRobot  = "机器人" // 机器人

	// 作弊率来源
	ProbSourceRoom  = "血池" //  血池
	ProbSourcePoint = "点控" // 点控
)

// InitConfig 加载配置
func (game *Blackjack) InitConfig() (err error) {

	// 加载房间配置
	confStr := game.table.GetAdviceConfig()

	js, err := simplejson.NewJson([]byte(confStr))
	if err != nil {
		log.Errorf("读取游戏配置失败: %v", err)
		return err
	}

	betBase, _ := js.Get("Bottom_Pouring").Int64()

	// 获取房间等级
	level := game.table.GetLevel()
	if level > 4 || level < 1 {
		log.Errorf("获取房间错误获取房间等级错误 %d", level)
		panic("获取房间等级错误")
	}

	betChips := config.BlackJackConf.GameConf.Chips[level-1]
	minBetLimit := config.BlackJackConf.GameConf.MinBetLimit[level-1]
	maxBetLimit := config.BlackJackConf.GameConf.MaxBetLimit[level-1]

	roomCfg := &config.RoomConfig{
		MinAction:    minBetLimit,
		MaxAction:    maxBetLimit,
		LimitAction:  game.table.GetEntranceRestrictions(),
		ActionOption: betChips,
		TaxRate:      game.table.GetRoomRate(),
		RoomCost:     betBase,
	}

	game.RoomCfg = roomCfg

	// 加载游戏配置；时间配置；控制配置
	bjConf := config.BlackJackConf

	// 记载机器人配置
	robotConf := config.RobotConf

	game.GameCfg = &bjConf.GameConf
	game.timeCfg = &bjConf.TimeConf
	game.ControlCfg = &bjConf.ExactControl
	game.RobotCfg = &robotConf
	game.LoadCfg = true
	return nil
}
func (game *Blackjack) BindRobot(ai player.RobotInterface) player.RobotHandler {
	// 加入玩家是机器人加载机器人配置
	robot := new(Robot)
	if game.RobotCfg == nil {
		log.Errorf("游戏 %v 第一个玩家为机器人", game)
	} else {
		robot.Init(ai, game, *game.RobotCfg)
	}
	return robot
}

// OnActionUserSitDown 用户坐下
func (game *Blackjack) OnActionUserSitDown(userInter player.PlayerInterface, orderIndex int, config string) table.MatchKind {

	userID := userInter.GetID()
	log.Debugf("玩家 %d 加入游戏 %d", userID, game.table.GetID())

	// 正常进入房间，非断线重联
	if _, ok := game.AllUserList[userID]; !ok {

		// 游戏中不能进入
		if game.Status != int32(msg.GameStatus_StartStatus) &&
			game.Status != int32(msg.GameStatus_StartMove) {
			//log.Tracef("游戏 %d 进行中不能进入", game.table.GetID())
			return table.SitDownErrorNomal
		}

		// 倒计时最后一秒不让进来
		if game.Status == int32(msg.GameStatus_StartMove) && (game.TimerJob != nil && game.TimerJob.GetTimeDifference() < 500) {
			log.Debugf("游戏 %d 最后几秒不能进入", game.table.GetID())
			return table.SitDownErrorNomal
		}

		var chairID int32

		// 随机座位
		for k, v := range game.Chairs {
			if v == 0 {
				chairID = k
				break
			}
		}

		user := &data.User{
			ID:               userID,
			UserName:         userInter.GetNike(),
			Head:             userInter.GetHead(),
			User:             userInter,
			Status:           int32(msg.UserStatus_UserSitDown),
			CurAmount:        userInter.GetScore(),
			InitAmount:       userInter.GetScore(),
			ChairID:          chairID,
			HoldCards:        [2]*data.HoldCards{{}, {}},
			TestCardsType:    int32(msg.TestCardsType_TestNoType),
			ExactControlRate: int64(userInter.GetProb()),
		}
		log.Warnf("玩家点控系数%d", userInter.GetProb())
		// 添加玩家到列表中
		game.Chairs[chairID] = userID
		game.AllUserList[userID] = user
		//log.Tracef("玩家 %d 座位 %d 加入游戏 %d ", userID, chairID, game.table.GetID())

	} else {

		// 断线用户重新登陆
		game.AllUserList[userID].ReConnect = true
	}

	return table.SitDownOk
}

// SendScene 玩家匹配之后调用这个发送场景消息
func (game *Blackjack) SendScene(userInter player.PlayerInterface) {
	userID := userInter.GetID()
	user, ok := game.AllUserList[userID]
	if !ok {
		log.Warnf("获取玩家 %d 异常！！！！", userID)
		return
	}

	// 加载配置文件
	if !game.LoadCfg {
		err := game.InitConfig()
		if err != nil {
			log.Errorf("加载配置文件失败")
			return
		}
	}

	// 非断线重联广播玩家坐下
	if !game.AllUserList[userID].ReConnect {

		// 游戏已经开始，发送开始倒计时状态信息
		if game.Status == int32(msg.GameStatus_StartMove) && game.TimerJob != nil {
			game.SendGameStatus(int32(msg.GameStatus_StartMove), int32(game.TimerJob.GetTimeDifference()/1000), userInter)
		}

		resp := msg.UserSitDownRes{
			UserData: &msg.SeatUserInfoRes{
				UserName: user.UserName,
				Head:     user.Head,
				Coin:     user.CurAmount,
				ChairId:  user.ChairID,
				UserId:   userID,
				Sex:      user.User.GetSex(),  // 性别
				Address:  user.User.GetCity(), // 地址
			},
		}
		game.SendUserSitDown(resp)
	}

	// 发送场景消息
	game.SendSceneInfo(userInter, game.AllUserList[userID].ReConnect)

	game.AllUserList[userID].ReConnect = false
	return
}

// UserReady 用户准备
func (game *Blackjack) UserReady(userInter player.PlayerInterface) bool {
	userID := userInter.GetID()
	//log.Tracef("玩家 %d 准备，游戏 %d 状态为：%d", userID, game.table.GetID(), game.Status)

	user, ok := game.AllUserList[userInter.GetID()]
	if !ok {
		log.Warnf("获取玩家异常！！！！当前用户列表 %v", game.AllUserList)
		return false
	}

	// 玩家已经准备
	if user.Status == int32(msg.UserStatus_UserReady) {
		log.Warnf("玩家重复准备！！")
		return false
	}

	// 用户坐下变准备
	if user.Status == int32(msg.UserStatus_UserSitDown) {
		user.Status = int32(msg.UserStatus_UserReady)
	}
	game.AllUserList[userID] = user

	return true
}

// GameStart 玩家准备后调用是否开赛
func (game *Blackjack) GameStart() {
	// 游戏开始
	if game.Status == int32(msg.GameStatus_StartStatus) {
		game.Start()
		return
	}
	return
}

// UserOffline 用户离线
func (game *Blackjack) UserOffline(userInter player.PlayerInterface) bool {
	var exitPermit bool

	// 初始游戏状态可退出
	if game.Status == int32(msg.GameStatus_StartStatus) {
		exitPermit = true
	}

	// 游戏匹配状态可退出
	if game.Status == int32(msg.GameStatus_StartMove) {
		exitPermit = true
	}

	// 游戏结算状态可退出
	if game.Status == int32(msg.GameStatus_Result) {
		exitPermit = true
	}

	// 游戏结束状态可退出
	if game.Status == int32(msg.GameStatus_GameOver) {
		exitPermit = true
	}

	user, ok := game.AllUserList[userInter.GetID()]
	if ok {
		//设置玩家为断线状态
		user.ReConnect = true

		// 玩家断线，不根据玩家状态进行离开权限判定
		log.Tracef("用户退出，当前游戏状态为：%d, 当前用户状态为：%d", game.Status, user.Status)
		if user.Status != int32(msg.UserStatus_UserBetSuccess) && user.Status != int32(msg.UserStatus_UserBetOver) {
			exitPermit = true
		} else {
			log.Tracef("不能允许玩家退出%d", userInter.GetID())
		}
	} else {
		// 玩家不再游戏桌上
		exitPermit = true
	}

	log.Debugf("用户 %d 退出，退出权限 %v", user.ID, exitPermit)

	// 玩家在桌上并且有退出权限时删除一些列数据
	if ok && exitPermit {
		delete(game.AllUserList, user.ID)
		delete(game.UserList, user.ID)

		// 让出座位
		for chairID, userID := range game.Chairs {
			if user.ID == userID {
				game.Chairs[chairID] = 0
				break
			}
		}

		// 移除押注序列
		for k, userID := range game.BetSeats {
			if user.ID == userID {
				game.BetSeats = append(game.BetSeats[:k], game.BetSeats[k+1:]...)
				break
			}
		}

		// 广播玩家离开信息
		res := msg.UserLeaveRoomRes{
			UserId:  user.ID,
			ChairId: user.ChairID,
		}
		game.table.Broadcast(int32(msg.SendToClientMessageType_S2CUserLeaveRoom), &res)
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_StartMove) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.AllUserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_StartStatus)
		if game.TimerJob != nil {
			game.table.DeleteJob(game.TimerJob)
		}
		game.table.Close()
	}
	return exitPermit
}

// UserLeaveGame 用户正常申请离开
func (game *Blackjack) UserLeaveGame(userInter player.PlayerInterface) bool {
	var exitPermit bool

	// 初始游戏状态可退出
	if game.Status == int32(msg.GameStatus_StartStatus) {
		exitPermit = true
	}

	// 游戏匹配状态可退出
	if game.Status == int32(msg.GameStatus_StartMove) {
		exitPermit = true
	}

	// 游戏结算状态可退出
	if game.Status == int32(msg.GameStatus_Result) {
		exitPermit = true
	}

	// 游戏结束状态可退出
	if game.Status == int32(msg.GameStatus_GameOver) {
		exitPermit = true
	}

	user, ok := game.AllUserList[userInter.GetID()]
	if ok {
		//log.Tracef("用户退出，当前游戏状态为：%d, 当前用户状态为：%d", game.Status, user.Status)
		if user.Status != int32(msg.UserStatus_UserBetSuccess) && user.Status != int32(msg.UserStatus_UserBetOver) {
			exitPermit = true
		}

	} else {
		// 玩家不再游戏桌上
		exitPermit = true
	}

	log.Debugf("用户 %d 退出，退出权限 %v", user.ID, exitPermit)

	// 玩家在桌上并且有退出权限时删除一些列数据
	if ok {
		if exitPermit {
			delete(game.AllUserList, user.ID)
			delete(game.UserList, user.ID)

			// 让出座位
			for chairID, userID := range game.Chairs {
				if user.ID == userID {
					game.Chairs[chairID] = 0
					break
				}
			}

			// 移除押注序列
			for k, userID := range game.BetSeats {
				if user.ID == userID {
					game.BetSeats = append(game.BetSeats[:k], game.BetSeats[k+1:]...)
					break
				}
			}

			// 广播玩家离开信息
			res := msg.UserLeaveRoomRes{
				UserId:  user.ID,
				ChairId: user.ChairID,
			}
			game.table.Broadcast(int32(msg.SendToClientMessageType_S2CUserLeaveRoom), &res)
		}
	}

	// 玩家匹配阶段退出检测
	if game.Status <= int32(msg.GameStatus_StartMove) && !user.User.IsRobot() {
		game.CheckLeftRobot()
	}

	// 所有玩家都离开，重置桌子状态，使其可进入
	if len(game.AllUserList) == 0 {
		// 桌子状态设为等待开始
		game.Status = int32(msg.GameStatus_StartStatus)
		if game.TimerJob != nil {
			game.table.DeleteJob(game.TimerJob)
		}
		game.table.Close()
	}
	return exitPermit
}

// OnGameMessage 接受用户发送信息
func (game *Blackjack) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	log.Debugf("收到客户端消息 ：%d", subCmd)
	switch subCmd {
	// 下注指令
	case int32(msg.ReceiveMessageType_C2SBet):
		game.UserBet(buffer, userInter)
		break
	// 下注完毕
	case int32(msg.ReceiveMessageType_C2SBetOver):
		game.UserConfirmBet(buffer, userInter)
		break
	// 购买保险
	case int32(msg.ReceiveMessageType_C2SBuyInsure):
		game.UserInsure(buffer, userInter)
		break
		// 第二轮操作阶段
	case int32(msg.ReceiveMessageType_C2SAskDo):
		game.UserDoCmd(buffer, userInter)
		break
		// 测试要求牌类型
	case int32(msg.ReceiveMessageType_C2SCardsType):
		//game.UserSetCards(buffer, userInter)
		break
	}
}

// ResetTable 重置桌子
func (game *Blackjack) ResetTable() {
	game.BetSeats = nil
	game.CurActionUser = nil
	game.HostCards = nil
	game.TurnCounter = 0
	game.LoadCfg = false

	// 座位号
	game.Chairs = map[int32]int64{
		0: 0,
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}
}

func (game *Blackjack) CloseTable() {}
