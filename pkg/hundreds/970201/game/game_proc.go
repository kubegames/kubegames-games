package game

import (
	"common/page"
	"fmt"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	msg2 "game_frame_v2/msg"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

//发红包
func (game *Game) ProcSendRed(buff []byte, user *data.User) {
	//defer recover_handle.RecoverHandle("ProcSendRed ... ")
	var c2sMsg msg.C2SSendRed
	err := proto.Unmarshal(buff, &c2sMsg)
	if err != nil {
		log.Debugf("procSendRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	if user.User.GetScore() < c2sMsg.Amount || c2sMsg.Amount < 100 {
		//log.Traceln("user.User.GetScore() ",user.User.GetScore(),c2sMsg.Amount)
		user.User.SendMsg(global.ERROR_CODE_NOT_ENOUGH, &msg.C2SIntoGame{})
		return
	}
	if c2sMsg.MineNum > 9 {
		//log.Traceln("雷号大于9")
		user.User.SendMsg(global.ERROR_CODE_RED_MINENUM, &msg.C2SIntoGame{})
		return
	}

	if game.redList.Len() > 60 || game.waitSendRedList.Len() > 60 {
		log.Traceln("红包长度超过60")
		return
	}

	if c2sMsg.Amount < game.sendAmount[0] || c2sMsg.Amount > game.sendAmount[len(game.sendAmount)-1] {
		//log.Traceln("发送金额 小于/大于 配置的金额：", c2sMsg.Amount, "  ", game.sendAmount[0], "  ", game.sendAmount[len(game.sendAmount)-1])
		if user.User.IsRobot() && user.User.GetScore() >= game.sendAmount[0] {
			c2sMsg.Amount = game.sendAmount[0]
		} else {
			//log.Traceln("发送金额 小于/大于 配置的金额 222 ：", c2sMsg.Amount, "  ", game.sendAmount[0], "  ", game.sendAmount[len(game.sendAmount)-1])
			_ = user.User.SendMsg(global.ERROR_CODE_LESS_MIN_ACTION, &msg.C2SIntoGame{})
			return
		}
	}

	user.BetsAmount += c2sMsg.Amount
	_ = user.User.SendMsg(global.S2C_SEND_RED_SUCCESS, &msg.C2SIntoGame{})
	_, _ = user.User.SetScore(game.Table.GetGameNum(), -c2sMsg.Amount, game.Table.GetRoomRate())
	user.ProfitAmount -= c2sMsg.Amount
	if !user.User.IsRobot() {
		log.Traceln("发送用户日志：：：用户发红包，发包金额")
		user.GameLogs = append(user.GameLogs, &msg2.GameLog{
			UserId: user.User.GetID(),
			Content: "用户id: " + fmt.Sprintf(`%d`, user.User.GetID()) + " 发红包，发包金额：" + score.GetScoreStr(c2sMsg.Amount) + " 余额：" + score.GetScoreStr(user.User.GetScore()) +
				"红包抢夺次数：10" + "红包雷号：" + fmt.Sprintf(`%d`, c2sMsg.MineNum),
		})
		//game.Table.WriteLogs(user.User.GetID(), "用户id: "+fmt.Sprintf(`%d`,user.User.GetID())+" 发红包，发包金额："+score.GetScoreStr(c2sMsg.Amount)+" 余额："+score.GetScoreStr(user.User.GetScore())+
		//	"红包抢夺次数：10"+"红包雷号："+fmt.Sprintf(`%d`,c2sMsg.MineNum))
	}
	red := NewRed(user, c2sMsg.Amount, 10, game, 2, c2sMsg.MineNum, []int32{10, 10, 10})
	red.level = red.GetLevel(game.sendAmount[0])
	red.Id = game.getNextRedId()
	if !user.User.IsRobot() {
		//不发不抢3分钟就被踢出
		user.LastRobTime = time.Now()
	}
	//log.Traceln("红包id：", red.Id)
	game.SetRedListMap(red)
	if game.redList.Len() <= 15 {
		//log.Traceln("消息2 发送的 红包id：",red.Id)
		game.curShowRedList.PushBack(red)
		red.Route = game.GetRouteAxis()
		game.Table.Broadcast(global.S2C_SEND_RED, red.GetRedInfo2C())
	} else {
		red.Status = global.RED_CUR_STATUS_WAITING
		game.waitSendRedList.PushBack(red)
		if err := user.User.SendMsg(global.S2C_RED_WAIT_SEND, &msg.S2CRedWaitSend{Count: game.GetBeforeSelfRedCount(user.User.GetID())}); err != nil {
		}
	}
	user.AddSendRedRecord(red.NewSendRedRecord2C(game.GetLevelStr()))

}

//获取自己的红包前面还有多少个
func (game *Game) GetBeforeSelfRedCount(userId int64) (count int32) {
	hasUserId := false
	for e := game.waitSendRedList.Front(); e != nil; e = e.Next() {
		red := e.Value.(*Red)
		if red.sender.User.GetID() == userId {
			hasUserId = true
			break
		}
	}
	if !hasUserId {
		return
	}

	for e := game.waitSendRedList.Front(); e != nil; e = e.Next() {
		count++
		red := e.Value.(*Red)
		if red.sender.Id == userId {
			return count
		}
	}
	return
}

//锁定红包
func (game *Game) ProcLockRed(buffer []byte, user *data.User) {
	//log.Traceln("<<<<<<<<<<锁定红包---")
	var c2sMsg msg.C2SLockRed
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procSendRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	//log.Traceln("1111>>>>> : ",time.Now().Sub(nowTime))
	red := game.GetRedListMap(c2sMsg.RedId)
	if red == nil {
		user.User.SendMsg(global.ERROR_CODE_RED_OVER, &msg.S2CRedId{RedId: c2sMsg.RedId})
		log.Traceln("procLockRed err : ", "该红包不存在 ", c2sMsg.RedId)
		game.DelCurShowRed(c2sMsg.RedId)
		return
	}
	if red.RobbedCount >= red.RedFlood {
		log.Traceln("procLockRed err : ", "该红包血量没了 ", c2sMsg.RedId)
		_ = user.User.SendMsg(global.ERROR_CODE_RED_OVER, &msg.S2CRedId{RedId: c2sMsg.RedId})
		game.DelCurShowRed(c2sMsg.RedId)
		game.DelRedListMap(red)
		return
	}

	game.SetLockListMap(user.Id, red)
	//log.Traceln("2222>>>>>> : ",time.Now().Sub(nowTime))
	user.User.SendMsg(global.S2C_LOCK_RED, red.GetRedInfo2C())
}

//取消锁定红包
func (game *Game) ProcCancelLockRed(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SLockRed
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procCancelLockRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	red := game.GetRedListMap(c2sMsg.RedId)
	if red == nil {
		//log.Debugf("procCancelLockRed err : %v", "该红包不存在")
		return
	}
	game.DelLockListMap(user.Id)
	_ = user.User.SendMsg(global.S2C_CANCEL_LOCK_RED, red.GetRedInfo2C())
}

//抢红包
func (game *Game) ProcRobRed(buffer []byte, user *data.User) {
	//log.Traceln("<<<<<<<<<<抢红包---")
	var c2sMsg msg.C2SRobRed
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procRobRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	red := game.GetRedListMap(c2sMsg.RedId)
	if red == nil || red.Id == 0 {
		//log.Debugf("procRobRed err : ", "该红包不存在"," red id : ",c2sMsg.RedId)
		return
	}
	//玩家金额小于入场金额，并且场内没有自己发的红包 就踢出去
	if user.User.GetScore() < game.Table.GetEntranceRestrictions() && !game.HasRed(user) {
		log.Traceln("玩家金额小于入场金额：", user.User.GetScore(), game.Table.GetEntranceRestrictions(), "profitAmount: ", user.ProfitAmount)
		game.BroadNoAi(global.S2C_LEAVE_TABLE, &msg.S2CKickOutUser{Uid: user.User.GetID(), Reason: global.KICKOUT_SCORE_NOT_ENOUGH})
		user.User.SendRecord(user.GameNum, user.ProfitAmount,
			user.BetsAmount, user.DrawAmount, user.Output, "")
		user.ProfitAmount = 0
		game.DelUserList(user.User.GetID())
		game.Table.KickOut(user.User)
		return
	}

	game.robRedLock.Lock()
	game.robRed(user, red)
	game.robRedLock.Unlock()

}

//获取发送过的红包列表
func (game *Game) ProcGetSentRed(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetSentRed
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcGetSentRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}

	if err := user.User.SendMsg(global.S2C_GET_SENT_RED, user.GetSendRedRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {
	}
}

//获取房间内用户列表
func (game *Game) ProcGetUserList(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetUserList
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetUserList err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	game.userList = game.SortUserList(game.userList)
	//log.Traceln("c2smsg : ",fmt.Sprintf(`%+v`,c2sMsg))
	s2cMsg := new(msg.S2CUserInfoArr)
	s2cMsg.UserArr = make([]*msg.S2CUserInfo, 0)
	start := c2sMsg.PageIndex * c2sMsg.PageSize
	end := c2sMsg.PageIndex*c2sMsg.PageSize + c2sMsg.PageSize
	//log.Traceln("start : ",start,end,game.userList.Len())
	var i int64 = 0
	for e := game.userList.Front(); e != nil; e = e.Next() {
		if i >= start && i < end {
			v := e.Value.(*data.User)
			s2cMsg.UserArr = append(s2cMsg.UserArr, v.GetUserMsgInfo())
			//log.Traceln("玩家列表信息：",fmt.Sprintf(`%+v`,v.GetUserMsgInfo()))
		}
		if i >= end {
			break
		}
		i++
	}
	//log.Traceln("front ::: ",game.userList.Front().Value.(*data.User))
	//log.Traceln("s2c 用户列表：",s2cMsg.UserArr)
	s2cMsg.Total = int64(game.userList.Len())
	pager := page.NewPager(int(c2sMsg.PageIndex), int(c2sMsg.PageSize), game.userList.Len())
	s2cMsg.Size, s2cMsg.Pages, s2cMsg.Current = int32(pager.Size), int32(pager.Pages), int32(pager.Current)
	log.Traceln("玩家总人数：", s2cMsg.Total)
	user.User.SendMsg(global.S2C_GET_USER_LIST, s2cMsg)
}

//中雷记录
func (game *Game) ProcGetMineRecord(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetMineRecord
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetUserList err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	if err := user.User.SendMsg(global.S2C_GET_MINE_RECORD, user.GetMineRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {
		log.Traceln("err : err : ", err)
	}
}

//抢红包的记录
func (game *Game) ProcGetRobbedInfo(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetRobbedInfo
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetRobbedInfo err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	if err := user.User.SendMsg(global.S2C_GET_ROBBED_INFO, user.GetRobRedRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {

	}
}

//获取大厅所有人的抢包记录
func (game *Game) ProcGetHallRecord(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetHallRecord
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetRobbedInfo err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	//if err := user.User.SendMsg(global.S2C_GET_HALL_RECORD,game.GetHallRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {
	//
	//}
}

//获取大厅所有人的抢包记录
func (game *Game) ProcGetUserCount(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetHallRecord
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetRobbedInfo err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	_ = user.User.SendMsg(global.S2C_USER_COUNT, &msg.S2CUserCount{
		UserCount: int32(game.userList.Len()),
	})
}
