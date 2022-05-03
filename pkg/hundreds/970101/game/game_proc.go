package game

import (
	"common/page"
	"common/score"
	"container/list"
	"fmt"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/golang/protobuf/proto"
)

//发红包
func (game *Game) ProcSendRed(buff []byte, user *data.User) {
	var c2sMsg msg.C2SSendRed
	err := proto.Unmarshal(buff, &c2sMsg)
	if err != nil {
		log.Debugf("procSendRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	if c2sMsg.MineNum > 9 || c2sMsg.MineNum < 0 {
		user.User.SendMsg(global.ERROR_CODE_RED_MINENUM, &msg.C2SIntoGame{})
		return
	}
	if int(c2sMsg.Amount) >= len(game.sendAmount) || c2sMsg.Amount < 0 {
		log.Traceln("c2sMsg.Amount 超过长度： ", c2sMsg.Amount)
		user.User.SendMsg(global.ERROR_CODE_NOT_ENOUGH, user.GetUserMsgInfo())
		return
	}
	if game.IsUserSentRed(user.User.GetId()) {
		//log.Traceln("红包列表中 玩家已经发送过红包了   ",user.User.GetId())
		user.User.SendMsg(global.ERROR_CODE_USER_SENT_RED, user.GetUserMsgInfo())
		return
	}
	amount := game.sendAmount[c2sMsg.Amount]
	//if user.User.IsRobot() {
	//	amount = game.sendAmount[3]
	//}
	if !user.User.IsRobot() {
		log.Traceln("玩家金额：：：：", user.User.GetScore(), "amount：：：", amount, "玩家id：", user.User.GetId())
	}
	if user.User.GetScore() < amount {
		log.Traceln("玩家金额不足： ", amount, "   ", user.User.GetScore())
		user.User.SendMsg(global.ERROR_CODE_NOT_ENOUGH, user.GetUserMsgInfo())
		return
	}
	//log.Traceln("玩家发送金额：",amount)
	_, _ = user.User.SetScore(game.Table.GetGameNum(), -amount, game.Table.GetRoomRate())
	red := NewRed(user, amount, game, c2sMsg.MineNum)
	red.Id = game.getNextRedId()
	//log.Traceln("红包id：", red.Id)
	game.SetRedList(red)
	user.AddSendRedRecord(red.NewSendRedRecord2C(game.Table.GetLevel()))
	user.NotOperateCount = 0
	if !user.User.IsRobot() {
		log.Traceln("user.NotOperateCount = 0 ", user.User.GetId())
	}
	_ = user.User.SendMsg(global.S2C_SEND_RED, red.GetRedInfo2C())
	if !user.User.IsRobot() {
		game.Table.WriteLogs(user.User.GetId(), "扫雷红包用户id:"+fmt.Sprintf(`%d`, user.User.GetID())+" 发红包，发包金额："+score.GetScoreStr(amount)+" 余额："+score.GetScoreStr(user.User.GetScore())+"红包id："+fmt.Sprintf(`%d`, red.Id)+
			"红包抢夺次数：10"+"红包雷号："+fmt.Sprintf(`%d`, c2sMsg.MineNum)+"红包状态：排队待发送")
	}

}

//抢红包
func (game *Game) ProcRobRed(buffer []byte, user *data.User) {
	//log.Traceln("<<<<<<<<<<抢红包---")
	var c2sMsg msg.C2SRobRed
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Tracef("procRobRed err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	red := game.GetRedFromList(c2sMsg.RedId)
	if red == nil {
		//log.Debugf("procRobRed err : ", "该红包不存在", " red id : ", c2sMsg.RedId)
		user.User.SendMsg(global.ERROR_CODE_RED_OVER, &msg.C2SIntoGame{})
		return
	}
	if game.status != global.TABLE_CUR_STATUS_START_ROB {
		log.Traceln("还没开抢")
		user.User.SendMsg(global.ERROR_CODE_NOT_START, &msg.C2SIntoGame{})
		return
	}
	if user.User.GetScore() < red.OriginAmount {
		log.Traceln("用户金额少于红包金额")
		user.User.SendMsg(global.ERROR_CODE_NOT_ENOUGH, user.GetUserMsgInfo())
		return
	}
	if red.RobbedCount >= red.RedFlood || red.Amount <= 0 {
		log.Traceln("红包已被抢完")
		user.User.SendMsg(global.ERROR_CODE_RED_OVER, user.GetUserMsgInfo())
		return
	}
	game.robRed(user, red)
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
	//user.User.SendMsg(global.S2C_GET_SENT_RED, GetUserSentMap(user.Id, int(c2sMsg.PageIndex), int(c2sMsg.PageSize)))
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
			log.Traceln("ProcGetUserList>>,uid : ", v.User.GetId())
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

	//var c2sMsg msg.C2SGetUserList
	//err := proto.Unmarshal(buffer, &msg.C2SIntoGame{})
	//if err != nil {
	//	log.Debugf("procGetUserList err : %v", err.Error())
	//	user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
	//	return
	//}
	//s2cMsg := new(msg.S2CUserInfoArr)
	//s2cMsg.UserArr = make([]*msg.S2CUserInfo, 0)
	//for e := game.userList.Front(); e != nil; e = e.Next() {
	//	v := e.Value.(*data.User)
	//	s2cMsg.UserArr = append(s2cMsg.UserArr, v.GetUserMsgInfo())
	//}
	//s2cMsg.Total = int64(len(s2cMsg.UserArr))
	//pager := page.NewPager(int(c2sMsg.PageIndex), int(c2sMsg.PageSize), len(s2cMsg.UserArr))
	//s2cMsg.Size, s2cMsg.Pages, s2cMsg.Current = int32(pager.Size), int32(pager.Pages), int32(pager.Current)
	//user.User.SendMsg(global.S2C_GET_USER_LIST, s2cMsg)
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
	//log.Traceln("用户抢包记录：：：",fmt.Sprintf(`%+v`,user.GetRobRedRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize)).RobbedArr))
	if err := user.User.SendMsg(global.S2C_GET_ROBBED_INFO, user.GetRobRedRecord(int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {

	}
}

//取消发送红包
func (game *Game) ProcCancelSend(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SCancelSend
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("ProcCancelSend err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	red := game.GetRedFromListByUid(user.User.GetId())
	if red != nil {
		firstRed := game.redList.Front().Value.(*Red)
		if red.Id == firstRed.Id {
			log.Traceln("取消发送红包失败，第一个红包不能取消")
			user.User.SendMsg(global.ERROR_CODE_CANT_CANCEL, user.GetUserMsgInfo())
			return
		}
		_, _ = red.sender.User.SetScore(game.Table.GetGameNum(), red.OriginAmount, 0)
		//red.sender.User.SendRecord(red.sender.User.GetRoomNum(),red.OriginAmount)
		game.Table.WriteLogs(red.sender.User.GetId(), "退还红包金额："+score.GetScoreStr(red.OriginAmount)+" 余额："+score.GetScoreStr(red.sender.User.GetScore()))
		game.Table.Broadcast(global.S2C_CANCEL_SEND, red.GetRedInfo2C())
		red.sender.DelUserSentRedRecord(red.Id)
		game.DelRedListMapNoMsg(red)
	} else {
		log.Traceln("红包不存在：", user.User.GetId(), red)
	}
}

//当前红包列表
func (game *Game) ProcGetCurRedList(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SGetCurRedList
	err := proto.Unmarshal(buffer, &c2sMsg)
	if err != nil {
		log.Debugf("procGetRobbedInfo err : %v", err.Error())
		user.User.SendMsg(global.ERROR_CODE_NOT_INTOROOM, &msg.C2SIntoGame{})
		return
	}
	if err := user.User.SendMsg(global.S2C_GET_CUR_RED_LIST, game.GetCurRedList(user, int(c2sMsg.PageIndex), int(c2sMsg.PageSize))); err != nil {

	}
}
