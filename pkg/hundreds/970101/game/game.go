package game

import (
	"common/dynamic"
	recover_handle "common/recover"
	"container/list"
	"fmt"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"game_frame_v2/game/clock"
	msg2 "game_frame_v2/msg"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

//game就是一个牌桌
type Game struct {
	Id               int64
	Table            table.TableInterface // table interface
	userList         *list.List           //所有的玩家列表
	redList          *list.List
	isStarted        bool // 是否启动过
	status           int  //游戏当前状态
	Room             *RedRoom
	curRobbedUserArr []*msg.S2CCurRobUser
	*dynamic.Dynamic
	timerJob      *clock.Job
	sendSecond    int
	MaxRobUidMap  map[int64]*UserRobStruct //red id =>
	maxRobUidLock sync.Mutex
	totalRedCount int64
	sendAmount    []int64 // 红包发送金额档位
	UserRobbedArr []*UserRobStruct

	UserRobbedCacheMap  map[int64][]*msg.S2CSendRedRecord //玩家发了红包就退出，后续的抢包信息要给他保留
	UserRobbedCacheLock sync.Mutex
	HoseLampArr         []*msg2.MarqueeConfig //跑马灯
	CurRed              *Red                  //当前正在抢的红包
	RobotScore          []int64
	isClosed            bool
}

func NewGame(id int64, room *RedRoom) (game *Game) {
	if room.Name == "" {
		room.Name = "菜鸟狩猎场"
	}
	game = &Game{
		Id: id, Room: room, userList: list.New(), redList: list.New(),
		status: global.TABLE_CUR_STATUS_READY_ROB, curRobbedUserArr: make([]*msg.S2CCurRobUser, 0),
		MaxRobUidMap: make(map[int64]*UserRobStruct), UserRobbedArr: make([]*UserRobStruct, 0),
		UserRobbedCacheMap: make(map[int64][]*msg.S2CSendRedRecord),
	}
	game.Dynamic = dynamic.NewDynamic()
	game.AddFunc("ProcSendRed", game, "ProcSendRed")
	game.AddFunc("ProcRobRed", game, "ProcRobRed")
	game.AddFunc("ProcGetSentRed", game, "ProcGetSentRed")
	game.AddFunc("ProcGetUserList", game, "ProcGetUserList")
	game.AddFunc("ProcGetRobbedInfo", game, "ProcGetRobbedInfo")
	//开始倒计时发送红包
	return
}

//添加暂存用户发过的红包
func (game *Game) AddUserRobbedCacheMap(uid int64, sendRecord *msg.S2CSendRedRecord) {
	game.UserRobbedCacheLock.Lock()
	if game.UserRobbedCacheMap[uid] == nil {
		game.UserRobbedCacheMap[uid] = make([]*msg.S2CSendRedRecord, 0)
	}
	game.UserRobbedCacheMap[uid] = append(game.UserRobbedCacheMap[uid], sendRecord)
	game.UserRobbedCacheLock.Unlock()
}

//删除暂存用户发过的红包
func (game *Game) DelUserRobbedCacheMap(uid int64) {
	game.UserRobbedCacheLock.Lock()
	delete(game.UserRobbedCacheMap, uid)
	game.UserRobbedCacheLock.Unlock()
}

//删除暂存用户发过的红包
func (game *Game) GetUserRobbedCacheMap(uid int64) (sendRecord []*msg.S2CSendRedRecord) {
	game.UserRobbedCacheLock.Lock()
	sendRecord = game.UserRobbedCacheMap[uid]
	game.UserRobbedCacheLock.Unlock()
	return
}

func (game *Game) startTick() {
	game.eventReady()
	//game.timerJob, _ = game.Table.AddTimer(5000, func() {
	//	game.eventStartRob()
	//})
	go game.goGameTimer()
}

//所有人员列表
func (game *Game) SetUserListMap(user *data.User) {
	game.userList.PushBack(user)
}
func (game *Game) GetUserListMap(uid int64) *data.User {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		if e.Value.(*data.User).Id == uid {
			return e.Value.(*data.User)
		}
	}
	return nil
}
func (game *Game) DelUserListMap(uid int64) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		if e.Value.(*data.User).Id == uid {
			game.userList.Remove(e)
			return
		}
	}
}

//所有红包列表
func (game *Game) SetRedList(red *Red) {
	game.redList.PushBack(red)
	game.totalRedCount++
}
func (game *Game) GetRedFromList(redId int64) *Red {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == redId {
			return e.Value.(*Red)
		}
	}
	return nil
}

func (game *Game) GetRedFromListByUid(uid int64) *Red {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).sender.User.GetID() == uid {
			return e.Value.(*Red)
		}
	}
	return nil
}

func (game *Game) GetRedListLen() int {
	return game.redList.Len()
}

func (game *Game) DelRedListMap(red *Red) {
	//抢包结束，在这儿统一上下分，并且写记录
	game.unifyUserRob(red)
	hasRed := false
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == red.Id {
			hasRed = true
			game.redList.Remove(e)
			break
		}
	}
	if hasRed {
		var maxUid int64 = 0
		if game.MaxRobUidMap[red.Id] != nil {
			maxUid = game.MaxRobUidMap[red.Id].User.User.GetID()
		}
		//if maxUid == 0 {
		//	log.Traceln("maxUid = game.MaxRobUidMap[red.Id].Uid  000000 ", fmt.Sprintf(`%+v`, game.MaxRobUidMap[red.Id]))
		//}

		s2cMsg := &msg.S2CRedDisappear{
			RedId: red.Id, Level: red.level, Ticker: 4, MaxUid: maxUid,
			SenderId: red.sender.User.GetID(), SenderScore: red.sender.User.GetScore(),
			RobUserArr: make([]*msg.S2CCurRobUser, 0),
		}
		for _, v := range game.UserRobbedArr {
			s2cMsg.RobUserArr = append(s2cMsg.RobUserArr, &msg.S2CCurRobUser{
				UserId:       v.User.User.GetID(),
				UserName:     v.User.User.GetNike(),
				Head:         v.User.User.GetHead(),
				RobbedAmount: v.RobbedAmount,
				IsMine:       v.IsMine,
				MineAmount:   v.MineAmount,
				Score:        v.User.User.GetScore(),
			})
		}
		game.Table.Broadcast(global.S2C_RED_DISAPPEAR, s2cMsg)
		game.DelMaxRobUidMap(red.Id)
	}
}

//删除红包但是不发送消息
func (game *Game) DelRedListMapNoMsg(red *Red) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == red.Id {
			game.redList.Remove(e)
			return
		}
	}
}

//离开房间
//func (game *Game) LeaveRoom(user *data.User) {
//	game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo())
//	game.DelUserListMap(user.Id)
//}

//获取房间基本信息
func (game *Game) GetRoomBaseInfo2C(userSelf *data.User) *msg.S2CRoomBaseInfo {
	info := &msg.S2CRoomBaseInfo{
		Id: game.Id, Name: game.Room.Name, LimitAmount: game.Room.LimitAmount, MaxRed: game.Room.MaxRed,
		UserArr: make([]*msg.S2CUserInfo, 0), SelfInfo: userSelf.GetUserMsgInfo(), Status: int32(game.status),
		CurRobUserArr: game.curRobbedUserArr,
	}
	if game.timerJob != nil {
		info.Ticker = int32(game.timerJob.GetTimeDifference() / 1000)
		//log.Traceln("场次",game.Table.GetID(),"ticker ::: ",info.Ticker," table ticker : ",game.timerJob.GetTimeDifference())
	}
	//log.Traceln("房间状态：",info.Status)
	//switch game.status {
	//case global.TABLE_CUR_STATUS_READY_ROB:
	//	info.Ticker = global.TICKER_TIME_READY_ROB - game.ticker
	//case global.TABLE_CUR_STATUS_START_ROB:
	//	info.Ticker = global.TICKER_TIME_START_ROB - game.ticker
	//case global.TABLE_CUR_STATUS_END_ROB:
	//	info.Ticker = global.TICKER_TIME_END_ROB - game.ticker
	//}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		info.UserArr = append(info.UserArr, user.GetUserMsgInfo())
	}
	if game.redList.Len() >= 1 {
		curRobbingRed := game.redList.Front().Value.(*Red)
		info.CurRobbingRed = curRobbingRed.GetRedInfo2C()
	}
	info.RedConfig = &msg.S2CRedConfigInfo{Odds: config.RedConfig.Odds, Count: config.RedConfig.Count, Amount: game.sendAmount}
	info.UserCount = int64(game.userList.Len())
	//log.Traceln("user count : ",info.UserCount," uid : ",userSelf.User.GetID())
	return info
}

//给房间里除了自己的其他人广播
func (game *Game) BroadNoSelf(selfId int64, subCmd int32, pb proto.Message) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.Id != selfId {
			if err := user.User.SendMsg(subCmd, pb); err != nil {
				log.Traceln("BroadNoSelf err : ", err)
				continue
			}
		}
	}
	return
}

//定时器
func (game *Game) goGameTimer() {
	recover_handle.RecoverHandle("goGameTimer ... ")
	aiSendRedTicker := time.NewTicker(1 * time.Second)
	kickTiker := time.NewTicker(3 * time.Second)
	checkDataTicker := time.NewTicker(20 * time.Second) //检查数据是否超过3天等
	for {
		select {
		case <-aiSendRedTicker.C:
			game.AiSendRedTimer()
			game.sendSecond++
			if game.sendSecond > 7 {
				game.sendSecond = 0
			}
		case <-checkDataTicker.C:
			game.checkData()
		case <-kickTiker.C:
			//log.Traceln("检查机器人是否金额超出，如果超出则踢")
			game.CheckKickRobot()
		}
	}
}

//检查机器人金额不足就踢出
func (game *Game) CheckKickRobot() {
	//log.Traceln(game.Table.GetLevel(),game.RobotScore)
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() && (user.User.GetScore() < game.RobotScore[0] || user.User.GetScore() > game.RobotScore[1]) {
			//log.Traceln("CheckKickRobot robot ")
			game.Table.KickOut(user.User)
			game.userList.Remove(e)
		}
	}
}

func (game *Game) eventReady() {
	//log.Traceln("时间到，发红包的倒计时...开始计时")
	//log.Traceln("重新获取跑马灯配置--------------------")
	game.HoseLampArr = game.Table.GetMarqueeConfig()
	game.status = global.TABLE_CUR_STATUS_READY_ROB
	game.curRobbedUserArr = make([]*msg.S2CCurRobUser, 0)
	game.UserRobbedArr = make([]*UserRobStruct, 0)
	var redInfo *msg.S2CRedInfo
	if game.redList.Len() >= 1 {
		redInfo = game.redList.Front().Value.(*Red).GetRedInfo2C()
	}
	game.Table.Broadcast(global.S2C_READY_ROB_TICKER, &msg.S2CStartTick{
		Ticker: global.TICKER_TIME_READY_ROB, RedInfo: redInfo,
	})
	game.timerJob, _ = game.Table.AddTimer(5000, func() {
		game.eventStartRob()
	})
	//log.Traceln("场次",game.Table.GetID(),"eventReady game.timerJob ",game.timerJob.GetIntervalTime())
}

//倒计时结束，开抢
func (game *Game) eventStartRob() {
	//log.Traceln("倒计时结束，发包->开始抢包。。。")
	if game.redList.Len() <= 0 {
		game.CurRed = nil
		//log.Traceln("没有红包，重新计时")
		game.timerJob.Cancel()
		game.timerJob, _ = game.Table.AddTimer(5000, func() {
			game.eventReady()
		})
		//log.Traceln("场次",game.Table.GetID(),"eventStartRob game.timerJob ",game.timerJob.GetIntervalTime())
		return
	}

	game.Table.StartGame()
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		user.NotOperateCount++
		if !user.User.IsRobot() {
			//log.Traceln("没发言 + 1  ",user.User.GetID(),user.NotOperateCount)
		}
	}
	red := game.redList.Front().Value.(*Red) //给用户抢的红包
	game.Table.WriteLogs(red.sender.User.GetID(), "扫雷红包用户id:"+fmt.Sprintf(`%d`, red.sender.User.GetID())+" 发红包，发包金额："+score.GetScoreStr(red.OriginAmount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore())+"红包id："+fmt.Sprintf(`%d`, red.Id)+
		"红包抢夺次数：10"+"红包雷号："+fmt.Sprintf(`%d`, red.MineNum)+"红包状态：发出待抢")
	game.CurRed = red
	red.sender.NotOperateCount = 0
	game.status = global.TABLE_CUR_STATUS_START_ROB
	game.BroadCast(global.S2C_START_ROB, red.GetRedInfo2C())
	//game.Table.Broadcast(global.S2C_START_ROB, red.GetRedInfo2C())
	time.AfterFunc(10*time.Second, func() {
		if game.status == global.TABLE_CUR_STATUS_START_ROB {
			log.Traceln("10s之后如果还没抢完就退回红包剩余的钱，开始抢下一个红包 : ", red.Id)
			game.GiveBackToUser(red)
			game.status = global.TABLE_CUR_STATUS_END_ROB
			game.restart()
		}
	})
}

//没抢完的红包退还给用户
func (game *Game) GiveBackToUser(red *Red) {
	log.Traceln("红包没抢完，退回：", red.sender.User.GetID(), "红包金额：", red.Amount)
	//没抢完，发包者打码量
	if !red.sender.User.IsRobot() {
		log.Traceln("红包没抢完，退回，打码量：", red.OriginAmount-red.Amount, " 红包id：", red.Id, " 发包者id：", red.sender.User.GetID())
	}
	_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.Amount, 0)
	red.sender.User.SendRecord(game.Table.GetGameNum(), red.Amount, 0,
		0, 0, "")
	game.Table.WriteLogs(red.sender.User.GetID(), "退还红包金额："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore()))
	game.DelRedListMap(red)
}

func (game *Game) restart() {
	//log.Traceln("抢包结束，进入清场阶段，重新开始")
	game.Table.EndGame()
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		//user.BetsAmount = 0
		if user.NotOperateCount >= 6 && !game.IsUserSentRed(user.User.GetID()) {
			if !user.User.IsRobot() {
				log.Traceln("超过5把，踢掉 ", user.User.GetID(), user.NotOperateCount)
			}
			time.AfterFunc(4*time.Second, func() {
				_ = user.User.SendMsg(global.S2C_KICK_OUT, &msg.S2CKickOutUser{UserId: user.User.GetID(), Reason: "超过5把没操作", ReasonInt: 1})
				user.ResetUser()
				//user.User.SetChip(user.Chip)
				game.Table.KickOut(user.User)
				game.DelUserListMap(user.User.GetID())
			})
		} else if user.User.GetScore() < game.Table.GetEntranceRestrictions() && !game.IsUserSentRed(user.User.GetID()) {
			//log.Traceln("踢掉玩家：",user.User.GetID()," 是否有红包：",game.IsUserSentRed(user.User.GetID()))
			time.AfterFunc(4*time.Second, func() {
				_ = user.User.SendMsg(global.S2C_KICK_OUT, &msg.S2CKickOutUser{UserId: user.User.GetID(), Reason: "金额不足", ReasonInt: 2})
				user.ResetUser()
				//user.User.SetChip(user.Chip)
				game.Table.KickOut(user.User)
				game.DelUserListMap(user.User.GetID())
			})
		}

	}
	game.status = global.TABLE_CUR_STATUS_END_ROB
	game.timerJob.Cancel()
	game.timerJob, _ = game.Table.AddTimer(7000, func() {
		game.eventReady()
	})
	//log.Traceln("场次",game.Table.GetID(),"eventStartRob game.timerJob ",game.timerJob.GetIntervalTime())
}

//统一记录用户抢包信息，进行上下分和写记录
func (game *Game) unifyUserRob(red *Red) {
	var sendRecord *msg.S2CSendRedRecord
	roomProb, _ := game.Table.GetRoomProb()
	//red.sender.User.SetEndCards(fmt.Sprintf(`发包%d元`,red.OriginAmount))
	var senderOutput int64 = red.Amount                                   //红包产出
	var senderChip = red.OriginAmount - red.Amount - red.SelfRobbedAmount //红包发送金额-剩余金额 => 被抢走的金额
	senderBets := red.OriginAmount                                        //红包发送者投入
	//var senderOldScore = red.sender.User.GetScore()
	var senderDrawAmount int64 = 0 //税收
	var senderProfitAmount = -red.OriginAmount
	//log.Traceln("玩家红包剩余金额：",red.sender.User.GetID(),senderTotal,"场次：",game.Table.GetLevel())
	for _, v := range game.UserRobbedArr {
		if v.IsMine {
			mineAmount := v.Red.OriginAmount * (int64(config.RedConfig.Odds) / 100)
			//v.User.BetsAmount += mineAmount
			taxScore := mineAmount * game.Table.GetRoomRate() / 10000
			senderDrawAmount += taxScore
			//log.Traceln("房间税收：",game.Table.GetRoomRate())
			recordScore := mineAmount - taxScore
			senderProfitAmount += recordScore
			//log.Traceln("玩家中雷金额：",mineAmount,"税收：",taxScore,"剩余战绩金额：",recordScore)
			_, _ = v.User.User.SetScore(game.Table.GetGameNum(), -mineAmount+v.RobbedAmount, game.Table.GetRoomRate())

			profitAmount, _ := v.Red.sender.User.SetScore(game.Table.GetGameNum(), mineAmount, game.Table.GetRoomRate())
			senderOutput += recordScore
			game.mineFrameLogSum(red, profitAmount, mineAmount-profitAmount)
			sendRecord = v.Red.sender.AddRedMineAmount(v.Red.Id, mineAmount)
			if !v.User.User.IsRobot() {
				game.Table.WriteLogs(v.User.User.GetID(), "抢红包中雷，赔付金额："+score.GetScoreStr(mineAmount-v.RobbedAmount)+" 余额："+score.GetScoreStr(v.User.User.GetScore())+
					"抢包玩家中雷，作弊率："+fmt.Sprintf(`%d`, v.User.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
					"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
					"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetID())+"发包人是："+aiRealStr(red.sender.User.IsRobot())+" 抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetID())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
					"红包id："+fmt.Sprintf(`%d`, v.Red.Id))
			}

			game.Table.WriteLogs(v.Red.sender.User.GetID(), "收到玩家中雷赔付金额："+score.GetScoreStr(mineAmount)+" 余额："+score.GetScoreStr(v.Red.sender.User.GetScore())+
				"发包玩家作弊率："+fmt.Sprintf(`%d`, v.Red.sender.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
				"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
				"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetID())+"发包人是："+aiRealStr(red.sender.User.IsRobot())+" 抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetID())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
				"红包id："+fmt.Sprintf(`%d`, v.Red.Id))
			game.TriggerHorseLamp(red.sender, mineAmount)
			//输家打码量
			if v.User.User.GetID() != red.sender.User.GetID() {
				if !v.User.User.IsRobot() {
					log.Traceln("中雷打码量：", mineAmount)
				}
				chip := mineAmount
				v.User.User.SendChip(chip)
				//v.User.User.SetBetsAmount(mineAmount) // 1、玩家中雷算投入
			} else {
				senderBets += mineAmount
				senderProfitAmount -= mineAmount
			}

		} else {
			_, _ = v.User.User.SetScore(game.Table.GetGameNum(), v.RobbedAmount, 0)
			if !v.User.User.IsRobot() {
				game.Table.WriteLogs(v.User.User.GetID(), "抢红包金额："+score.GetScoreStr(v.RobbedAmount)+" 余额："+score.GetScoreStr(v.User.User.GetScore())+
					"抢包玩家作弊率："+fmt.Sprintf(`%d`, v.User.Cheat)+" 系统作弊率："+fmt.Sprintf(`%d`, roomProb)+
					"红包被抢次数："+fmt.Sprintf(`%d`, v.RedRobbedCount)+"红包剩余金额："+score.GetScoreStr(v.RedRemainAmount)+
					"发包人id："+fmt.Sprintf(`%d`, red.sender.User.GetID())+"发包人是："+aiRealStr(red.sender.User.IsRobot())+" 抢包人id："+fmt.Sprintf(`%d`, v.User.User.GetID())+"雷号："+fmt.Sprintf(`%d`, v.MineNum)+
					"红包id："+fmt.Sprintf(`%d`, v.Red.Id)+" 玩家状态："+userProbStr(v.User))
			}

		}
		if v.User.User.GetID() == red.sender.User.GetID() {
			senderOutput += v.RobbedAmount
			senderProfitAmount += v.RobbedAmount
		}
		//log.Traceln("红包被抢次数: ",v.RedRobbedCount,"red id : ",v.Red.Id)
		recordAmount := v.RobbedAmount
		if v.IsMine {
			recordAmount = v.RobbedAmount - red.OriginAmount
		}
		if !v.User.User.IsRobot() {
			log.Traceln("玩家显示抢包盈利：", recordAmount, v.RobbedAmount, v.MineAmount, v.IsMine)
		}
		v.User.AddRobRedRecord(v.Red.NewUserRobbedRedInfo(recordAmount, game.Table.GetLevel(), v.IsMine))
		if v.User.User.GetID() != red.sender.User.GetID() {
			var userBetsAmount int64 = 0
			if v.IsMine {
				userBetsAmount = red.OriginAmount
			}
			v.User.User.SendRecord(game.Table.GetGameNum(), v.RobbedAmount-userBetsAmount,
				userBetsAmount, 0, v.RobbedAmount, "抢包")
		}
		game.TriggerHorseLamp(v.User, v.RobbedAmount)
	}
	//发包者单独发送战绩
	//log.Traceln("红包发送者战绩，打码量，总投入：",red.sender.User.GetID(),senderTotal,senderChip,red.sender.BetsAmount)
	red.sender.User.SendChip(senderChip)
	//red.sender.User.SetBetsAmount(red.sender.BetsAmount)
	//red.sender.User.SetBetsAmount(senderBets)	//2、红包发送者投入
	red.sender.User.SendRecord(game.Table.GetGameNum(), senderProfitAmount, senderBets,
		senderDrawAmount, senderOutput, "")
	//log.Traceln("红包发送者 打码量，总投入,战绩(产出)：",red.sender.User.GetID(),senderChip,senderBets,senderOutput)

	sender := game.GetUserListMap(red.sender.Id)
	if sender == nil && sendRecord != nil {
		log.Traceln("给玩家添加暂存", red.sender.Id)
		game.AddUserRobbedCacheMap(red.sender.Id, sendRecord)
	}

}

func userProbStr(user *data.User) string {
	if user.User.GetProb() == 0 {
		return "使用血池"
	}
	return "被点控"
}

func (game *Game) checkData() {
	//检查数据是否过期
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		for i := 0; i < len(user.SendRedList); {
			if time.Now().Unix()-user.SendRedList[i].Time > 60*60*24 {
				user.SendRedList = append(user.SendRedList[:i], user.SendRedList[i+1:]...)
				break
			} else {
				i++
			}
		}

		for i := 0; i < len(user.RobbedList); {
			if time.Now().Unix()-user.RobbedList[i].Time > 60*60*24 {
				user.RobbedList = append(user.RobbedList[:i], user.RobbedList[i+1:]...)
				break
			} else {
				i++
			}
		}

	}
}

func (game *Game) BroadCast(subCmd int32, pb proto.Message) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		//log.Traceln("给user：",user.User.GetID(),"发送 ",subCmd)
		if err := user.User.SendMsg(subCmd, pb); err != nil {
			log.Traceln("user.User.SendMsg(sub err : ", err)
		}
	}
}

//玩家新进来，找寻之前是否就已经有记录在里面，如果有则追加
func (game *Game) CopySendList(user *data.User) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.sender.Id == user.User.GetID() {
			user.SendRedList = append(user.SendRedList, red.NewSendRedRecord2C(game.Table.GetLevel()))
		}
	}

	//再将原来有的发包信息添加进去
	oldSendRecord := game.GetUserRobbedCacheMap(user.Id)
	if oldSendRecord != nil && len(oldSendRecord) != 0 {
		user.SendRedList = append(user.SendRedList, oldSendRecord...)
	}
	game.DelUserRobbedCacheMap(user.Id)
}

//更新红包发送者的内存地址
func (game *Game) UpdateRedSender(user *data.User) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).sender.Id == user.User.GetID() {
			e.Value.(*Red).sender = user
		}
	}
}

//是否有玩家已经发送过
func (game *Game) IsUserSentRed(uid int64) bool {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.sender.Id == uid {
			return true
		}
	}
	return false
}

//true : 返回 "机器人"
func aiRealStr(flag bool) string {
	if flag {
		return "【机器人】"
	}
	return "【真实玩家】"
}
