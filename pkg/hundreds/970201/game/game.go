package game

import (
	"common/dynamic"
	"container/list"
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	msg2 "game_frame_v2/msg"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	Id          int64
	Table       table.TableInterface // table interface
	userList    *list.List           //所有的玩家列表
	redList     *list.List           //牌桌上的红包列表
	lockRedMap  map[int64]*Red       //uid => red
	lockRedLock sync.Mutex
	isStarted   bool // 是否启动过
	nowSecond   int  //当前倒计时的秒数，2～7s 每s都有不同概率的处理
	Room        *RedRoom
	sendAmount  []int64 // 发送红包的档位
	*dynamic.Dynamic
	sendRedLock     sync.Mutex
	waitSendRedList *list.List
	curShowRedList  *list.List //当前正在页面的红包
	//hallRecord []*msg.S2CHallRecord
	totalRedCount    int64
	robRedUserArrMap map[int64][]int64
	robRedUserLock   sync.Mutex
	HoseLampArr      []*msg2.MarqueeConfig //跑马灯

	robRedLock sync.Mutex
	RobotScore []int64
	isClosed   bool //是否正在关闭房间
}

func NewGame(id int64, room *RedRoom) (game *Game) {
	if room.Name == "" {
		room.Name = "菜鸟狩猎场"
	}
	game = &Game{
		Id: id, Room: room, userList: list.New(), redList: list.New(), lockRedMap: make(map[int64]*Red),
		sendAmount: make([]int64, 0), waitSendRedList: list.New(), curShowRedList: list.New(),
		robRedUserArrMap: make(map[int64][]int64), HoseLampArr: make([]*msg2.MarqueeConfig, 0),
	}
	game.Dynamic = dynamic.NewDynamic()
	game.AddFunc("ProcSendRed", game, "ProcSendRed")
	game.AddFunc("ProcLockRed", game, "ProcLockRed")
	game.AddFunc("ProcCancelLockRed", game, "ProcCancelLockRed")
	game.AddFunc("ProcRobRed", game, "ProcRobRed")
	game.AddFunc("ProcGetSentRed", game, "ProcGetSentRed")
	game.AddFunc("ProcGetUserList", game, "ProcGetUserList")
	game.AddFunc("ProcGetMineRecord", game, "ProcGetMineRecord")
	game.AddFunc("ProcGetRobbedInfo", game, "ProcGetRobbedInfo")
	return
}

//红包被所抢的用户列表

func (game *Game) SetRobRedUserArr(redId, uid int64) {
	game.robRedUserLock.Lock()
	game.robRedUserArrMap[redId] = append(game.robRedUserArrMap[redId], uid)
	game.robRedUserLock.Unlock()
}
func (game *Game) DelRobRedUserArr(redId int64) {
	game.robRedUserLock.Lock()
	delete(game.robRedUserArrMap, redId)
	game.robRedUserLock.Unlock()
}

//所有人员列表
func (game *Game) SetUserList(userInter *data.User) {
	if game.userList.Len() == 0 {
		game.userList.PushBack(userInter)
		return
	}
	isPushed := false
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if userInter.RobbedAmount > user.RobbedAmount {
			game.userList.InsertBefore(userInter, e)
			isPushed = true
			break
		}
	}
	if !isPushed {
		game.userList.PushBack(userInter)
	}

	//log.Traceln("开头用户：",game.userList.Front().Value.(*data.User).Id)

}
func (game *Game) GetUserList(uid int64) *data.User {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		if e.Value.(*data.User).Id == uid {
			return e.Value.(*data.User)
		}
	}
	return nil
}
func (game *Game) DelUserList(uid int64) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		if e.Value.(*data.User).Id == uid {
			game.userList.Remove(e)
			return
		}
	}
}

//所有红包列表
func (game *Game) SetRedListMap(red *Red) {
	game.redList.PushBack(red)
	game.totalRedCount++
}
func (game *Game) GetRedListMap(redId int64) *Red {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == redId {
			return e.Value.(*Red)
		}
	}
	return nil
}
func (game *Game) DelRedListMap(red *Red) {
	hasRed := false
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == red.Id {
			hasRed = true
			game.redList.Remove(e)
		}
	}
	if hasRed {
		for e := game.userList.Front(); e != nil; e = e.Next() {
			user := e.Value.(*data.User)
			if !user.IsAi {
				//log.Traceln("红包消失id：", red.Id)
				if err := user.User.SendMsg(global.S2C_RED_DISAPPEAR, &msg.S2CRedDisappear{
					RedId: red.Id, Level: red.level, IsRobbed: user.IsRobbedRed(red.Id), IsOtherRobbed: game.IsOtherRobbed(red.Id, user.User.GetID()),
				}); err != nil {

				}
			}
		}
	}
	game.DelCurShowRed(red.Id)
}

func (game *Game) IsOtherRobbed(redId, uid int64) bool {
	game.robRedUserLock.Lock()
	if game.robRedUserArrMap[redId] == nil || len(game.robRedUserArrMap[redId]) == 0 {
		log.Traceln("robRedUserArrMap[redId] 为空 ")
		game.robRedUserLock.Unlock()
		return false
	}
	flag := false
	for _, v := range game.robRedUserArrMap[redId] {
		if v != uid {
			flag = true
			break
		}
	}
	game.robRedUserLock.Unlock()
	return flag
}

//删除当前展示的红包列表
func (game *Game) DelCurShowRed(redId int64) {
	for e := game.curShowRedList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == redId {
			game.curShowRedList.Remove(e)
		}
	}

}

func (game *Game) DelWaitRed(redId int64) {
	for e := game.waitSendRedList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).Id == redId {
			game.waitSendRedList.Remove(e)
		}
	}

}

//锁定的红包列表
func (game *Game) SetLockListMap(userId int64, red *Red) {
	game.lockRedLock.Lock()
	game.lockRedMap[userId] = red
	game.lockRedLock.Unlock()
}

func (game *Game) DelLockListMap(userId int64) {
	game.lockRedLock.Lock()
	delete(game.lockRedMap, userId)
	game.lockRedLock.Unlock()
}

func (game *Game) DelLockListByRedId(redId int64) {
	game.lockRedLock.Lock()
	for uid, red := range game.lockRedMap {
		if red.Id == redId {
			delete(game.lockRedMap, uid)
			break
		}
	}
	game.lockRedLock.Unlock()
}

//离开房间
//func (game *Game) LeaveRoom(user *data.User) {
//	game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo())
//	game.DelUserList(user.Id)
//}

//获取房间基本信息
func (game *Game) GetRoomBaseInfo2C(userSelf *data.User) *msg.S2CRoomBaseInfo {
	info := &msg.S2CRoomBaseInfo{
		Id: game.Id, Name: game.Room.Name, LimitAmount: game.Room.LimitAmount, MaxRed: game.Room.MaxRed,
		UserArr: make([]*msg.S2CUserInfo, 0), SelfInfo: userSelf.GetUserMsgInfo(),
		RedArr: make([]*msg.S2CRedInfo, 0), WaitSendCount: game.GetBeforeSelfRedCount(userSelf.User.GetID()),
	}
	var i int64 = 0
	for e := game.curShowRedList.Front(); e != nil; e = e.Next() {
		v := e.Value.(*Red)
		if i >= 15 {
			break
		}
		info.RedArr = append(info.RedArr, v.GetRedInfo2C())
		i++
	}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		info.UserArr = append(info.UserArr, user.GetUserMsgInfo())
	}
	info.RedConfig = &msg.S2CRedConfigInfo{Odds: config.RedConfig.Odds, Count: config.RedConfig.Count, Amount: game.sendAmount, SpaceAmount: config.RedConfig.SpaceAmount}
	info.UserCount = int64(game.userList.Len())
	//log.Traceln("玩家人数：", info.UserCount)
	//log.Traceln("当前显示红包长度：", game.curShowRedList.Len())
	//log.Traceln("返回给客户端的红包：", info.RedArr)
	return info
}

//给房间里除了自己的其他人广播
func (game *Game) BroadNoAi(subCmd int32, pb proto.Message) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if !user.IsAi {
			//log.Traceln("给玩家：",user.User.GetID(),"发送消息6")
			if err := user.User.SendMsg(subCmd, pb); err != nil {
				//log.Traceln("BroadNoSelf err : ", err)
				//continue
			}
		}
	}
	return
}

//只给机器人发送消息
func (game *Game) BroadAi(subCmd int32, pb proto.Message) {
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() {
			if err := user.User.SendMsg(subCmd, pb); err != nil {
			}
		}
	}
	return
}

//定时器
func (game *Game) goGameTimer() {
	game.lockRedLock.Lock()
	for uid, red := range game.lockRedMap {
		user := game.GetUserList(uid)
		if user != nil && red != nil && red.RobbedCount < red.RedFlood {
			game.sendUserLockRedInfo(user, red)
		}
	}
	game.lockRedLock.Unlock()
	//game.AiSendRedTimer()
	game.nowSecond++
	if game.nowSecond > 7 {
		game.nowSecond = 0
	}
	game.checkData()
	game.tickerCheckKickOut()
	game.tickerCheckOverdueRed()
	//lockRedTicker := time.NewTicker(1 * time.Second)
	//aiSendRedTicker := time.NewTicker(1 * time.Second)
	//checkDataTicker := time.NewTicker(20 * time.Second) //检查数据是否超过3天等
	//checkUserTicker := time.NewTicker(4 * time.Second)
	//for {
	//	select {
	//	case <-lockRedTicker.C:
	//		//log.Traceln("每秒检查lockRedMap : ",game.lockRedMap)
	//		game.lockRedLock.Lock()
	//		for uid, red := range game.lockRedMap {
	//			user := game.GetUserList(uid)
	//			if user != nil && red != nil && red.RobbedCount < red.RedFlood {
	//				game.sendUserLockRedInfo(user, red)
	//			} else {
	//				//log.Traceln(" red : ", fmt.Sprintf(`%v`, red))
	//			}
	//		}
	//		game.lockRedLock.Unlock()
	//	case <-aiSendRedTicker.C:
	//		game.AiSendRedTimer()
	//		game.nowSecond++
	//		if game.nowSecond > 7 {
	//			game.nowSecond = 0
	//		}
	//	case <-checkDataTicker.C:
	//		game.checkData()
	//	case <-checkUserTicker.C:
	//		game.tickerCheckKickOut()
	//		game.tickerCheckOverdueRed()
	//	}
	//}
}

//检查机器人金额不足就踢出
func (game *Game) CheckKickRobot() {
	//log.Traceln(game.Table.GetLevel(),game.RobotScore)
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		if user.User.IsRobot() && (user.User.GetScore() < game.RobotScore[0] || user.User.GetScore() > game.RobotScore[1]) {
			if game.Table.GetLevel() == 1 {
				log.Traceln("初级场 。。。 CheckKickRobot robot,机器人金额： ", user.User.GetScore(), "配置金额：", game.RobotScore)
			}
			red := game.GetRedFromListByUid(user.User.GetID())
			if red != nil {
				game.GiveBackToUser(red)
			}
			game.UserExit(user.User)
		}
	}
}

//检查过期红包
func (game *Game) tickerCheckOverdueRed() {
	for e := game.curShowRedList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if time.Now().Sub(red.FlyTime) > 50*time.Second {
			game.BroadAi(global.S2C_CHECK_OVERDUE_RED, &msg.S2CRedId{RedId: red.Id})
		}
	}
}

//检查踢人
func (game *Game) tickerCheckKickOut() {
	//t1 := time.Now()
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		//log.Traceln("table id : ",game.Table.GetID(),"user id ::: ",user.User.GetID())
		if time.Now().Sub(user.LastRobTime) > 3*time.Minute && game.GetRedFromListByUid(user.User.GetID()) == nil {
			//if !user.User.IsRobot()  {
			//log.Traceln("用户3分钟没发言，踢掉", user.User.GetID(), " 上一次发言：", user.LastRobTime, "是否机器人：", user.User.IsRobot())
			//}
			//game.DelLockListMap(user.User.GetID())
			game.BroadNoAi(global.S2C_LEAVE_TABLE, &msg.S2CKickOutUser{Uid: user.User.GetID(), Reason: global.KICKOUT_TIME_OUT})
			game.Table.KickOut(user.User)
			game.UserExit(user.User)
			//game.DelUserList(user.User.GetID())
		}
	}
	//log.Traceln("table id : ",game.Table.GetID(),"执行时间：",time.Now().Sub(t1))
}

//每秒给锁定了红包的用户发消息
func (game *Game) sendUserLockRedInfo(user *data.User, red *Red) {
	//log.Traceln("给user : ", user.Id, " 发送红包 ：", red.Id, "  血量：", red.RedFlood, "已抢：", red.RobbedCount)
	if err := user.User.SendMsg(global.S2C_RED_INFO, &msg.S2CRedFlood{
		RedId: red.Id, RedFlood: red.RedFlood, RobbedCount: red.RobbedCount, Level: red.level,
	}); err != nil {
	}
}

//判定红包是否需要重新飞入界面
func (game *Game) SendRedFromWaitList() {
	//t1 := time.Now()
	if game.waitSendRedList.Len() >= 1 {
		red := game.waitSendRedList.Front().Value.(*Red)
		red.Route = game.GetRouteAxis()
		game.curShowRedList.PushBack(red)
		red.sender.ChangeUserSendRedStatus(red.Id, global.RED_CUR_STATUS_ING)
		game.Table.Broadcast(global.S2C_SEND_RED, red.GetRedInfo2C())
		if game.waitSendRedList != nil && game.waitSendRedList.Front() != nil {
			game.waitSendRedList.Remove(game.waitSendRedList.Front())
		}

		//log.Traceln("给",red.sender.User.GetID(),"发送消息15   ",game.GetBeforeSelfRedCount(red.sender.User.GetID()))
		if err := red.sender.User.SendMsg(global.S2C_RED_WAIT_SEND, &msg.S2CRedWaitSend{Count: game.GetBeforeSelfRedCount(red.sender.User.GetID())}); err != nil {
			log.Traceln("发送失败，err ： ", err)
		}
	}
	//log.Traceln("------------11-----------",time.Now().Sub(t1))
	for e := game.waitSendRedList.Front(); e != nil; e = e.Next() {
		//log.Traceln("发送消息15 ",red.sender.User.GetID()," 数量：",game.GetBeforeSelfRedCount(red.sender.User.GetID()))
		red := e.Value.(*Red)
		if !red.sender.IsSendWaitMsg {
			if err := red.sender.User.SendMsg(global.S2C_RED_WAIT_SEND, &msg.S2CRedWaitSend{Count: game.GetBeforeSelfRedCount(red.sender.User.GetID())}); err != nil {
			}
			red.sender.IsSendWaitMsg = true
		}
	}
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		user.IsSendWaitMsg = false
	}
	//log.Traceln("------------22-----------",time.Now().Sub(t1))
}

func (game *Game) checkData() {
	//检查数据是否过期
	for e := game.userList.Front(); e != nil; e = e.Next() {
		user := e.Value.(*data.User)
		//log.Traceln("user ：",user.Id,"发包条数：",len(user.SendRedList))
		//log.Traceln("user ：",user.Id,"抢包条数：",len(user.RobbedList))
		//log.Traceln("user ：",user.Id,"中雷条数：",len(user.MineList))
		for i := 0; i < len(user.SendRedList); {
			if time.Now().Unix()-user.SendRedList[i].Time > 60*60*24 {
				user.SendRedList = append(user.SendRedList[:i], user.SendRedList[i+1:]...)
			} else {
				i++
			}
		}

		for i := 0; i < len(user.RobbedList); {
			if time.Now().Unix()-user.RobbedList[i].Time > 60*60*24 {
				user.RobbedList = append(user.RobbedList[:i], user.RobbedList[i+1:]...)
			} else {
				i++
			}
		}

		for i := 0; i < len(user.MineList); {
			if time.Now().Unix()-user.MineList[i].Time > 60*60*24 {
				user.MineList = append(user.MineList[:i], user.MineList[i+1:]...)
			} else {
				i++
			}
		}
	}
}

func (game *Game) GetLevelStr() string {
	switch game.Table.GetLevel() {
	case 1:
		return "初级场"
	case 2:
		return "中级场"
	case 3:
		return "高级场"
	case 4:
		return "大师场"
	}
	return "初级场"
}

func (game *Game) SortUserList(oldList *list.List) (newList *list.List) {
	//t1 := time.Now()
	newList = list.New()
	for v := oldList.Front(); v != nil; v = v.Next() {
		node := newList.Front()
		for nil != node {
			if node.Value.(*data.User).RobbedAmount < v.Value.(*data.User).RobbedAmount {
				newList.InsertBefore(v.Value.(*data.User), node)
				break
			} else if node.Value.(*data.User).RobbedAmount == v.Value.(*data.User).RobbedAmount {
				if node.Value.(*data.User).RobbedAmount < v.Value.(*data.User).RobbedAmount {
					newList.InsertBefore(v.Value.(*data.User), node)
					break
				}
			}
			node = node.Next()
		}
		if node == nil {
			newList.PushBack(v.Value.(*data.User))
		}
	}
	//log.Traceln("排序耗时>>>", time.Now().Sub(t1))
	return newList
}

func (game *Game) UpdateRedSender(user *data.User) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).sender.Id == user.User.GetID() {
			e.Value.(*Red).sender = user
		}
	}
}

//玩家新进来，找寻之前是否就已经有记录在里面，如果有则追加
func (game *Game) CopySendList(user *data.User) {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.sender.Id == user.User.GetID() {
			user.SendRedList = append(user.SendRedList, red.NewSendRedRecord2C(game.GetLevelStr()))
		}
	}
}

//该玩家是否有红包在场
func (game *Game) HasRed(user *data.User) bool {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.sender.Id == user.User.GetID() {
			return true
		}
	}
	return false
}

//没抢完的红包退还给用户
func (game *Game) GiveBackToUser(red *Red) {
	log.Traceln("红包没抢完，退回：", red.sender.User.GetID(), "红包金额：", red.Amount)
	//红包退回 添加打码量 被抢走的金额
	if !red.sender.User.IsRobot() {
		log.Traceln("红包没抢完，退回，打码量：", red.OriginAmount-red.Amount, " 红包id：", red.Id, " 发包者id：", red.sender.User.GetID())
	}
	//robbedAmount := red.OriginAmount - red.Amount
	//red.sender.BetsAmount += robbedAmount
	red.sender.Output += red.Amount
	//red.sender.RobbedAmount += red.Amount
	red.sender.Chip = red.sender.Chip + red.OriginAmount - red.Amount - red.SelfRobbedAmount
	_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.Amount, 0)
	red.sender.ProfitAmount += red.Amount
	//game.Table.WriteLogs(red.sender.User.GetID(), "退还红包金额："+score.GetScoreStr(red.Amount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore())+
	//	"红包剩余次数："+fmt.Sprintf(`%d`, red.RedFlood-red.RobbedCount)+"红包剩余金额："+score.GetScoreStr(red.Amount)+" 红包id："+fmt.Sprintf(`%d`, red.Id))
	red.Amount = 0
	game.DelRedListMap(red)
}

func (game *Game) GetRedFromListByUid(uid int64) *Red {
	for e := game.redList.Front(); e != nil; e = e.Next() {
		if e.Value.(*Red).sender.User.GetID() == uid {
			return e.Value.(*Red)
		}
	}
	return nil
}
