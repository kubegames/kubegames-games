package game

import (
	"common/log"
	"common/rand"
	"common/score"
	"container/list"
	"encoding/json"
	"fmt"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	"game_frame_v2/define"
	msg2 "game_frame_v2/msg"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

//UserReady 用户准备
func (game *Game) UserReady(user player.PlayerInterface) bool {
	//fmt.Println("frame >>>>>>> UserReady")

	return true
}

//OnActionUserSitDown 用户坐下
//如果符合条件就坐下，返回true，不符合就不让坐下，返回false
func (game *Game) OnActionUserSitDown(userInter player.PlayerInterface, chairID int, configStr string) int {
	//fmt.Println("frame >>>>>>> OnActionUserSitDown")
	user := game.GetUserList(userInter.GetId())
	if user != nil {
		fmt.Println("分配房间成功,用户掉线重新连回来", user.Id, "chair id : ", user.ChairId, " 用户状态：", user.Status, "房间号：", game.Id)
		user.LastRobTime = time.Now()
		user.User = userInter
		game.UpdateRedSender(user)
		return define.SIT_DOWN_OK
	} else {
		user = data.NewUser(userInter.GetId(), userInter.GetNike(), userInter.IsRobot(), game.Table)
	}
	if userInter.IsRobot() {
		//fmt.Println("性格总数：", len(config.AiRobConfigArr))
		index := rand.RandInt(0, len(config.AiRobConfigArr)-1)
		//fmt.Println("机器人，为机器人分配性格  ", index)
		user.AiRobConfig = config.AiRobConfigArr[index]
		robot := NewRobot(game, user)
		aiUser := userInter.BindRobot(robot)
		robot.AiUser = aiUser
		game.isClosed = false
	}
	user.Table = game.Table
	user.User = userInter
	game.SetUserList(user)
	//更新红包发送者的内存地址
	game.UpdateRedSender(user)
	user.GameNum = game.Table.GetGameNum()
	//if !userInter.IsRobot(){
	//	fmt.Println("分配房间成功： user id : ", user.Id, "chair id : ", user.ChairId, " 用户状态：", user.Status, "房间号：", game.Id)
	//}
	//game.CopySendList(user)
	game.Table.Broadcast(global.S2C_OTHER_INTO_ROOM, user.GetUserMsgInfo())
	return define.SIT_DOWN_OK
}

//UserExit 用户坐下
func (game *Game) UserExit(userInter player.PlayerInterface) bool {
	//if !userInter.IsRobot() || userInter.GetId() > 10000 {
	//fmt.Println("frame >>>>>>> UserExit ",userInter.IsRobot()," ",userInter.GetId())
	//}
	user := game.GetUserList(userInter.GetId())
	if user == nil {
		fmt.Println("UserExit 用户没在桌子上 ")
		return true
	}
	red := game.GetRedFromListByUid(user.User.GetId())
	if red != nil {
		fmt.Println("UserExit 用户还有红包没被抢完，不能离开 ", userInter.GetId())
		_ = userInter.SendMsg(global.ERROR_CODE_CANT_LEAVE, &msg.S2CString{
			Msg: "您发送的红包尚未被抢完，无法退出游戏",
		})
		return false
	}

	user.User.SendChip(user.Chip)
	//新改动需求：玩家离开房间时发送战绩，发送玩家一共抢到的金额
	//userInter.SetBetsAmount(user.BetsAmount)
	if !user.User.IsRobot() {
		fmt.Println("send record userInter.GetRoomNum() : ", user.GameNum)
	}
	userInter.SendRecord(user.GameNum, user.ProfitAmount,
		user.BetsAmount, user.DrawAmount, user.Output, "")
	//fmt.Println(user.User.GetId(),"当前打码量，投入，产出：",user.Chip,user.BetsAmount,user.Output)
	user.Chip = 0
	user.BetsAmount = 0
	user.Output = 0
	//go game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo())
	//game.BackUserAllRed(userInter.GetId())
	game.DelUserList(userInter.GetId())
	game.DelLockListMap(userInter.GetId())
	if !user.User.IsRobot() {
		//fmt.Println("send log ::; userInter.GetRoomNum() ",user.GameNum,"内存地址：",user,&user)
		user.User.SendLogs(user.GameNum, user.GameLogs)
	}
	return true
}

//OnGameMessage 游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	defer log.Trace()()
	//fmt.Println("frame >>>>>>> OnGameMessage")
	user := game.GetUserList(userInter.GetId())
	if user == nil {
		fmt.Println("OnGameMessage user nil ")
		return
	}
	switch subCmd {
	case global.C2S_SEND_RED:
		//fmt.Println(">>>>>>>>>发红包+++")
		game.ProcSendRed(buffer, user)
		//game.Run("ProcSendRed", buffer, user)
	case global.C2S_LOCK_RED:
		//fmt.Println(">>>>>>>>>锁定红包+++")
		game.Run("ProcLockRed", buffer, user)
	case global.C2S_CANCEL_LOCK_RED:
		//fmt.Println(">>>>>>>>>取消锁定红包+++")
		game.Run("ProcCancelLockRed", buffer, user)
	case global.C2S_ROB_RED:
		//if user.User.IsRobot(){
		//fmt.Println(">>>>>>>>>抢红包+++",user.User.GetId())
		//}
		game.ProcRobRed(buffer, user)
		//game.Run("ProcRobRed", buffer, user)
	case global.C2S_GET_SENT_RED:
		fmt.Println(">>>>>>>>>获取发送过的红包列表+++")
		game.ProcGetSentRed(buffer, user)
	case global.C2S_GET_ROBBED_INFO:
		fmt.Println(">>>>>>>>>获取抢过的红包信息列表+++")
		game.Run("ProcGetRobbedInfo", buffer, user)
	case global.C2S_GET_USER_LIST:
		fmt.Println(">>>>>>>>>获取房间内用户列表+++")
		//game.Run("ProcGetUserList", buffer, user)
		game.ProcGetUserList(buffer, user)
	case global.C2S_GET_MINE_RECORD:
		fmt.Println(">>>>>>>>>中雷记录+++")
		game.ProcGetMineRecord(buffer, user)
		//game.Run("ProcGetMineRecord", buffer, user)
	case global.C2S_GET_HALL_RECORD:
		game.ProcGetHallRecord(buffer, user)
	case global.C2S_GET_USER_COUNT:
		game.ProcGetUserCount(buffer, user)
	}
}

//GameStart 游戏开始
func (game *Game) GameStart(user player.PlayerInterface) bool {
	defer log.Trace()()
	if game.Table.GetId() < 0 {
		fmt.Println("房间id < 0 ")
		return false
	}
	//fmt.Println("玩家的作弊率： ",user.GetProb())
	if !game.isStarted {
		fmt.Println("advice : ", game.Table.GetAdviceConfig(), "level :::::::  ", game.Table.GetLevel())
		var tableConfig TableConfig
		if err := json.Unmarshal([]byte(game.Table.GetAdviceConfig()), &tableConfig); err != nil {
			fmt.Println("advice 的 值不对： ", game.Table.GetAdviceConfig())
			return false
		}
		game.RobotScore = config.GetRobotConfByLevel(game.Table.GetLevel())
		game.Table.AddTimerRepeat(2*1000, 0, func() {
			//fmt.Println("红包总长度：",game.redList.Len())
			//fmt.Println("等待红包总长度：",game.waitSendRedList.Len())
			game.goGameTimer()
			game.CheckKickRobot()
		})
		//3月5号新增发包规则
		game.Table.AddTimerRepeat(5*1000, 0, func() {
			game.AiSendNew()
		})
		game.Table.AddTimerRepeat(2*1000, 0, func() {
			game.AiRobNew()
		})
		fmt.Println("底注：", tableConfig.Bottom_Pouring)
		game.HoseLampArr = game.Table.GetMarqueeConfig()
		game.sendAmount = make([]int64, 0)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*2)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*3)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*4)
		//fmt.Println("game.sendAmount 222 ", game.sendAmount, " game id : ", game.Table.GetId(), " level : ", game.Table.GetLevel())
		game.isStarted = true
		//fmt.Println("框架开始游戏")
		game.Table.EndGame() //获取机器人逻辑这儿很奇葩，要先调用endgame才会分配机器人
		game.Table.StartGame()
		game.Table.AddTimerRepeat(1000*60, 0, func() {
			game.HoseLampArr = game.Table.GetMarqueeConfig()
			//fmt.Println("执行end start",time.Now())
			game.Table.EndGame()
			game.Table.StartGame()

		})
	}
	//fmt.Println("frame >>>>>>> GameStart")
	return true
}

//TableConfig 牌桌配置 底注
type TableConfig struct {
	Bottom_Pouring int64
}

//SendScene 场景消息
func (game *Game) SendScene(userInter player.PlayerInterface) bool {
	user := game.GetUserList(userInter.GetId())
	if user == nil {
		fmt.Println("SendScene user nil ")
		return false
	}
	userInter.SendMsg(global.S2C_INTO_ROOM, game.GetRoomBaseInfo2C(user))
	return true
}

//通知关闭桌子
func (game *Game) CloseTable() {
	fmt.Println("清退所有用户的时候，归还红包 22222 ")
	game.isStarted = false
	game.isClosed = true
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		fmt.Println("CloseTable红包没抢完，退回：", red.sender.User.GetId(), "红包金额：", red.Amount)
		//robbedAmount := red.OriginAmount - red.Amount
		//red.sender.BetsAmount += robbedAmount
		//red.sender.RobbedAmount += red.Amount
		//red.sender.RobbedAmount += red.Amount
		red.sender.Output += red.Amount
		red.sender.Chip = red.sender.Chip + red.OriginAmount - red.Amount - red.SelfRobbedAmount
		_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.Amount, 0)
		red.sender.ProfitAmount += red.Amount
		red.sender.GameLogs = append(red.sender.GameLogs, &msg2.GameLog{
			UserId: red.sender.User.GetId(),
			Content: aiRealStr(red.sender.User.IsRobot()) + "用户id：" + fmt.Sprintf(`%d`, red.sender.User.GetID()) + "退还红包金额：" + score.GetScoreStr(red.Amount) + " 余额：" + score.GetScoreStr(red.sender.User.GetScore()) +
				"红包剩余次数：" + fmt.Sprintf(`%d`, red.RedFlood-red.RobbedCount) + "红包剩余金额：" + score.GetScoreStr(red.Amount) + " 红包id：" + fmt.Sprintf(`%d`, red.Id),
		})
		//game.Table.WriteLogs(red.sender.User.GetId(), "退还红包金额："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore())+
		//	"红包剩余次数："+fmt.Sprintf(`%d`, red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+" 红包id："+fmt.Sprintf(`%d`, red.Id))
	}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		fmt.Println(user.User.GetId(), "CloseTable 当前打码量，投入，产出：", user.Chip, user.BetsAmount, user.Output)
		user.User.SendChip(user.Chip)
		//新改动需求：玩家离开房间时发送战绩，发送玩家一共抢到的金额
		//user.User.SetBetsAmount(user.BetsAmount)
		user.User.SendRecord(user.GameNum, user.ProfitAmount,
			user.BetsAmount, user.DrawAmount, user.Output, "")
		game.Table.KickOut(user.User)
	}
	game.userList = list.New()
	game.redList = list.New()
}

//ResetTable 重置牌桌
func (game *Game) ResetTable() {
	game.isClosed = true
}

//用户在线情况下主动离开
func (game *Game) LeaveGame(userInter player.PlayerInterface) bool {
	return game.UserExit(userInter)
}
