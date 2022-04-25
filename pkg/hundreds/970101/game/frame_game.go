package game

import (
	"common/rand"
	"common/score"
	"container/list"
	"encoding/json"
	"fmt"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"game_frame_v2/define"

	"github.com/kubegames/kubegames-sdk/pkg/player"
)

func (game *Game) UserReady(user player.PlayerInterface) bool {
	//fmt.Println("frame >>>>>>> UserReady")
	return true
}

//用户坐下
//如果符合条件就坐下，返回true，不符合就不让坐下，返回false
func (game *Game) OnActionUserSitDown(userInter player.PlayerInterface, chairId int, configStr string) int {
	game.isClosed = false
	//fmt.Println("frame >>>>>>> OnActionUserSitDown")
	user := game.GetUserListMap(userInter.GetId())
	if user != nil {
		fmt.Println("分配房间成功,用户掉线重新连回来", user.Id, "chair id : ", user.ChairId, " 用户状态：", user.Status, "房间号：", game.Id)
		user.User = userInter
		game.UpdateRedSender(user)
		return define.SIT_DOWN_OK
	} else {
		user = data.NewUser(userInter.GetId(), userInter.GetNike(), userInter.IsRobot(), game.Table)
	}
	if userInter.IsRobot() {
		//fmt.Println("性格总数：", len(config.json.AiRobConfigArr))
		index := rand.RandInt(0, len(config.AiRobConfigArr)-1)
		//fmt.Println("机器人，为机器人分配性格  ", index)
		user.AiRobConfig = config.AiRobConfigArr[index]
		robot := NewRobot(game, user)
		aiUser := userInter.BindRobot(robot)
		robot.AiUser = aiUser
	}
	user.Table = game.Table
	user.User = userInter
	game.SetUserListMap(user)
	game.UpdateRedSender(user)
	//game.CopySendList(user)
	if userInter.IsRobot() {
		game.isClosed = false
	}

	return define.SIT_DOWN_OK
}

func (game *Game) UserExit(userInter player.PlayerInterface) bool {
	//fmt.Println("frame >>>>>>> UserExit")
	user := game.GetUserListMap(userInter.GetId())
	if user == nil {
		fmt.Println("UserExit 用户没在桌子上 ")
		return true
	}
	red := game.GetRedFromListByUid(user.User.GetId())
	if red != nil {
		fmt.Println("UserExit 用户还有红包没被抢完，不能离开 ")
		_ = userInter.SendMsg(global.ERROR_CODE_CANT_LEAVE, &msg.S2CString{
			Msg: "您发送的红包尚未被抢完，无法退出游戏",
		})
		return false
	}
	if game.CurRed != nil && game.CurRed.sender.User.GetId() == userInter.GetId() && game.CurRed.RobbedCount < 10 {
		fmt.Println("UserExit 当前红包正在被抢，不能离开 ")
		_ = userInter.SendMsg(global.ERROR_CODE_CANT_LEAVE, &msg.S2CString{
			Msg: "您发送的红包正在被抢，暂时无法退出游戏",
		})
		return false
	}
	//game.BroadCast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo())
	game.DelUserListMap(userInter.GetId())
	userCount := game.userList.Len()
	game.BroadCast(global.S2C_CUR_USER_COUNT, &msg.S2CUserCount{
		Count: int64(userCount),
	})

	return true
}

//游戏消息
func (game *Game) OnGameMessage(subCmd int32, buffer []byte, userInter player.PlayerInterface) {
	//fmt.Println("frame >>>>>>> OnGameMessage")
	user := game.GetUserListMap(userInter.GetId())
	if user == nil {
		fmt.Println("OnGameMessage user nil ")
		return
	}
	switch subCmd {
	case global.C2S_SEND_RED:
		//fmt.Println(">>>>>>>>>发红包+++")
		game.ProcSendRed(buffer, user)
	case global.C2S_ROB_RED:
		//fmt.Println(">>>>>>>>>抢红包+++")
		game.ProcRobRed(buffer, user)
	case global.C2S_GET_SENT_RED:
		fmt.Println(">>>>>>>>>获取发送过的红包列表+++")
		game.ProcGetSentRed(buffer, user)
	case global.C2S_GET_ROBBED_INFO:
		fmt.Println(">>>>>>>>>获取抢过的红包信息列表+++")
		game.ProcGetRobbedInfo(buffer, user)
	case global.C2S_GET_USER_LIST:
		fmt.Println(">>>>>>>>>获取房间内用户列表+++")
		game.ProcGetUserList(buffer, user)
	case global.C2S_CANCEL_SEND:
		fmt.Println(">>>>>>>>>取消发送+++")
		game.ProcCancelSend(buffer, user)
	case global.C2S_GET_CUR_RED_LIST:
		//fmt.Println(">>>>>>>>>C2S_GET_CUR_RED_LIST+++")
		game.ProcGetCurRedList(buffer, user)
	}
}

type TableConfig struct {
	Bottom_Pouring int64
}

func (game *Game) GameStart(user player.PlayerInterface) bool {
	if game.Table.GetId() < 0 {
		fmt.Println("房间id < 0 ")
		return false
	}
	//fmt.Println("玩家的作弊率： ",user.GetProb())
	if !game.isStarted {
		game.isStarted = true
		var tableConfig TableConfig
		if err := json.Unmarshal([]byte(game.Table.GetAdviceConfig()), &tableConfig); err != nil {
			fmt.Println("advice 的 值不对： ", game.Table.GetAdviceConfig())
			return false
		}
		fmt.Println("框架开始游戏,配置：", fmt.Sprintf(`%+v`, tableConfig))
		game.HoseLampArr = game.Table.GetMarqueeConfig()
		game.sendAmount = make([]int64, 0)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*2)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*3)
		game.sendAmount = append(game.sendAmount, tableConfig.Bottom_Pouring*4)
		game.RobotScore = config.GetRobotConfByLevel(game.Table.GetLevel())
		//game.Table.EndGame() //获取机器人逻辑这儿很奇葩，要先调用endgame才会分配机器人
		game.Table.StartGame()
		game.startTick()
	}
	//fmt.Println("frame >>>>>>> GameStart")
	return true
}

func (game *Game) SendScene(userInter player.PlayerInterface) bool {
	//fmt.Println("frame >>>>>>> SendScene")
	//fmt.Println("table level : ",game.Table.GetLevel())
	//fmt.Println(config.GetRobotConfByLevel(game.Table.GetLevel()))
	user := game.GetUserListMap(userInter.GetId())
	if user == nil {
		fmt.Println("SendScene user nil ")
		return false
	}
	tableInfo := game.GetRoomBaseInfo2C(user)
	//fmt.Println("table Info : ",tableInfo)
	userInter.SendMsg(global.S2C_INTO_ROOM, tableInfo)
	//game.BroadNoSelf(userInter.GetId(), global.S2C_OTHER_INTO_ROOM, user.GetUserMsgInfo())
	//game.Table.Broadcast()
	userCount := game.userList.Len()
	game.BroadCast(global.S2C_CUR_USER_COUNT, &msg.S2CUserCount{
		Count: int64(userCount),
	})
	return true
}

//清退正在抢的红包
func (game *Game) ClearRobbingRed(red *Red) {
	var sendRecord *msg.S2CSendRedRecord
	roomProb, _ := game.Table.GetRoomProb()
	var senderOutput = red.Amount                                         //发包者战绩获得的钱 ，中雷+自己抢的
	senderBets := red.OriginAmount                                        //红包发送者投入
	var senderDrawAmount int64 = 0                                        //税收
	var senderChip = red.OriginAmount - red.Amount - red.SelfRobbedAmount //红包发送金额-剩余金额 => 被抢走的金额
	for _, v := range game.UserRobbedArr {
		if v.IsMine {
			mineAmount := v.Red.OriginAmount * (int64(config.RedConfig.Odds) / 100)
			taxScore := mineAmount * game.Table.GetRoomRate() / 10000
			senderDrawAmount += taxScore
			recordScore := mineAmount - taxScore
			_, _ = v.User.User.SetScore(game.Table.GetGameNum(), -mineAmount+v.RobbedAmount, game.Table.GetRoomRate())
			profitAmount, _ := v.Red.sender.User.SetScore(game.Table.GetGameNum(), mineAmount, game.Table.GetRoomRate())
			senderOutput += recordScore
			game.mineFrameLogSum(red, profitAmount, mineAmount-profitAmount)
			sendRecord = v.Red.sender.AddRedMineAmount(v.Red.Id, mineAmount)
			game.Table.WriteLogs(v.User.User.GetId(), "抢红包中雷，赔付金额："+score.GetScoreStr(mineAmount-v.RobbedAmount)+" 余额："+score.GetScoreStr(v.User.User.GetScore())+
				"抢包玩家中雷，作弊率："+fmt.Sprintf(`%d`, v.User.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
				"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
				"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetID())+"抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetID())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
				"红包id："+fmt.Sprintf(`%d`, v.Red.Id))
			game.Table.WriteLogs(v.Red.sender.User.GetId(), "收到玩家中雷赔付金额："+score.GetScoreStr(mineAmount)+" 余额："+score.GetScoreStr(v.Red.sender.User.GetScore())+
				"发包玩家作弊率："+fmt.Sprintf(`%d`, v.Red.sender.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
				"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
				"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetID())+"抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetID())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
				"红包id："+fmt.Sprintf(`%d`, v.Red.Id))
			game.TriggerHorseLamp(red.sender, mineAmount)
			//输家打码量
			if v.User.User.GetId() != red.sender.User.GetId() {
				if !v.User.User.IsRobot() {
					fmt.Println("中雷打码量：", mineAmount)
				}
				chip := mineAmount
				v.User.User.SendChip(chip)
				//v.User.User.SetBetsAmount(mineAmount)
			} else {
				senderBets += mineAmount
			}
		} else {
			_, _ = v.User.User.SetScore(game.Table.GetGameNum(), v.RobbedAmount, 0)
			game.Table.WriteLogs(v.User.User.GetId(), "抢红包金额："+score.GetScoreStr(v.RobbedAmount)+" 余额："+score.GetScoreStr(v.User.User.GetScore())+
				"抢包玩家作弊率："+fmt.Sprintf(`%d`, v.User.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
				"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
				"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetId())+"抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetId())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
				"红包id："+fmt.Sprintf(`%d`, v.Red.Id))
		}
		if v.User.User.GetId() == red.sender.User.GetId() {
			senderOutput += v.RobbedAmount
		} else {
			var userBetsAmount int64 = 0
			if v.IsMine {
				userBetsAmount = red.OriginAmount
			}
			v.User.User.SendRecord(game.Table.GetGameNum(), v.RobbedAmount-userBetsAmount, userBetsAmount, 0, v.RobbedAmount, "抢包")
		}
		//fmt.Println("红包被抢次数: ",v.RedRobbedCount,"red id : ",v.Red.Id)
		recordAmount := v.RobbedAmount
		if v.IsMine {
			recordAmount = v.RobbedAmount - v.MineAmount
		}
		v.User.AddRobRedRecord(v.Red.NewUserRobbedRedInfo(recordAmount, game.Table.GetLevel(), v.IsMine))
		game.TriggerHorseLamp(v.User, v.RobbedAmount)
	}
	//发包者单独发送战绩
	//fmt.Println("clear robbing red 红包发送者战绩，打码量，总投入：",red.sender.User.GetId(),senderTotal,senderChip,red.sender.BetsAmount)
	red.sender.User.SendChip(senderChip)
	//red.sender.User.SetBetsAmount(senderBets)
	red.sender.User.SendRecord(game.Table.GetGameNum(), senderOutput, senderBets, senderDrawAmount, senderOutput, "")
	sender := game.GetUserListMap(red.sender.Id)
	if sender == nil && sendRecord != nil {
		fmt.Println("给玩家添加暂存", red.sender.Id)
		game.AddUserRobbedCacheMap(red.sender.Id, sendRecord)
	}
	game.UserRobbedArr = make([]*UserRobStruct, 0)
}

//通知关闭桌子
func (game *Game) CloseTable() {
	game.isClosed = true
	fmt.Println("清退所有用户的时候，归还红包 222 ", game.redList.Len(), game.Table.GetId())
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.OriginAmount != red.Amount {
			fmt.Println("red 正在抢", red.Id)
			game.ClearRobbingRed(red)
		} else {
			fmt.Println("关闭桌子 ，退回：", red.sender.User.GetId(), "红包金额：", red.OriginAmount)
			//red.sender.User.SetBetsAmount(red.OriginAmount)
			_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.OriginAmount, 0)
			//red.sender.User.SendRecord(red.sender.User.GetRoomNum(), red.OriginAmount,red.OriginAmount,0,red.OriginAmount,"关桌子退红包金额")
			game.Table.WriteLogs(red.sender.User.GetId(), "退还红包金额："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore()))
		}
	}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		game.Table.KickOut(user.User)
	}
	game.userList = nil
	game.redList = nil
	game.userList = list.New()
	game.redList = list.New()
	fmt.Println("清空玩家和红包列表,清退所有用户的时候，归还红包 222 ", game.redList.Len(), game.Table.GetId())
	game.Table.StartGame()
	game.Table.AddTimer(3000, func() {
		fmt.Println("执行end game")
		game.Table.EndGame()
		game.timerJob.Cancel()
	})

}

//重置逻辑
func (game *Game) ResetTable() {
	game.isClosed = true
	//if game.isClosed {
	//	fmt.Println("房间已关闭，无须reset ")
	//	return
	//}
	//fmt.Println("清退所有用户的时候，归还红包 1111 红包长度：",,game.redList.Len())
	//game.isStarted = false
	//for e := game.redList.Front(); e != nil; e = e.Next() {
	//	red := e.Value.(*Red)
	//	fmt.Println("退还红包 ： ",red.Id)
	//	game.GiveBackToUser(red)
	//
	//}
	//game.redList = list.New()
	fmt.Println("清退所有用户的时候，归还红包 111 红包列表：房间id：人数：", game.redList.Len(), game.Table.GetId(), game.userList.Len())
	game.isStarted = false
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.OriginAmount != red.Amount {
			fmt.Println("red 正在抢", red.Id)
			game.ClearRobbingRed(red)
		} else {
			fmt.Println("关闭桌子 ，退回：", red.sender.User.GetId(), "红包金额：", red.OriginAmount)
			//red.sender.User.SetBetsAmount(red.OriginAmount)
			_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.OriginAmount, 0)
			//red.sender.User.SendRecord(red.sender.User.GetRoomNum(), red.OriginAmount,red.OriginAmount,0,red.OriginAmount,"关桌子退红包金额")
			game.Table.WriteLogs(red.sender.User.GetId(), "退还红包金额："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore()))
		}
	}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		game.Table.KickOut(user.User)
	}
	game.userList = nil
	game.redList = nil
	game.userList = list.New()
	game.redList = list.New()
	fmt.Println("清空两个列表，清退所有用户的时候，归还红包 111 ", game.redList.Len(), game.Table.GetId())
	//game.Table.EndGame()
}

//用户在线情况下主动离开
func (game *Game) LeaveGame(userInter player.PlayerInterface) bool {
	return game.UserExit(userInter)
}
