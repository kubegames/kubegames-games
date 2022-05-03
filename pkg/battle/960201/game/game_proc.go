package game

import (
	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//玩家点击开始游戏-请求机器人
func (game *Game) ProcUserStartGame(buffer []byte, user *data.User) {
	if user == nil {
		//log.Traceln("ProcUserStartGame user nil ")
		return
	}
	//game.userStartGameLock.Lock()
	//defer game.userStartGameLock.Unlock()
	var c2sMsg msg.C2SUserStartGame
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("procUserStartGame proto err : ", err)
		return
	}

	if game.GetTableUserCount() == 1 {
		if err := user.User.SendMsg(global.S2C_WAIT_START, &msg.S2CRoomInfo{}); err != nil {
			log.Traceln("User.SendMsg err : ", err)
			return
		}
	}
	game.AiIntoRoom()

}

//用户发言
func (game *Game) ProcAction(buffer []byte, user *data.User) {
	if user == nil {
		////log.Traceln("ProcAction user nil ")
		return
	}
	var c2sMsg msg.C2SUserAction
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	game.Action(user, c2sMsg.Option, c2sMsg.Amount)
}

//比牌
func (game *Game) ProcCompare(buffer []byte, user *data.User) {
	if user == nil {
		//log.Traceln("ProcCompare user nil ")
		return
	}
	//log.Traceln("倒计时时间：",game.timerJob.GetTimeDifference() )
	if game.timerJob.GetTimeDifference() <= 1000 {
		log.Traceln("倒计时时间 小于 1000 ")
		return
	}
	var c2sMsg msg.C2SCompareCards
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	user1 := game.GetUserListMap(c2sMsg.FirstUserId)
	user2 := game.GetUserListMap(c2sMsg.SecondUserId)
	if user1 == nil || user2 == nil {
		//log.Traceln("user1 : ", user1, " first : ", c2sMsg.FirstUserId)
		//log.Traceln("user2 : ", user2, " second : ", c2sMsg.SecondUserId)
		return
	}

	//如果发起者玩家余额不足，则不能比牌
	needAmount := game.MinAction * 2
	if user1.IsSawCards {
		needAmount *= 2
	}
	if user1.Score < needAmount {
		log.Traceln("钱不够，不能比牌")
		user1.User.SendMsg(global.ERROR_CODE_COMPARE_NOT_ENOUGH, user1.GetUserMsgInfo(false))
		return
	}
	game.CompareIds[0], game.CompareIds[1] = c2sMsg.FirstUserId, c2sMsg.SecondUserId

	//开始比牌
	game.SetTicker(0, 100000)
	userList := make([]*data.User, 2)
	userList[0] = user1
	userList[1] = user2
	game.CompareCards(user1, user2, userList)
	if game.IsSatisfyEnd() {
		game.Table.AddTimer(4*1000, func() {
			//log.Traceln("比牌结束，结束比赛")
			game.EndGame(false)
		})
	}
}

//用户获取可比牌列表
func (game *Game) ProcGetCanCompareList(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SIntoGame
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	userInfoArr := new(msg.S2CUserInfoArr)
	userInfoArr.UserInfoArr = make([]*msg.S2CUserInfo, 0)
	for _, v := range game.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		if v.Id != user.Id {
			userInfoArr.UserInfoArr = append(userInfoArr.UserInfoArr, v.GetUserMsgInfo(false))
		}
	}
	user.User.SendMsg(global.S2C_GET_CAN_COMPARE_LIST, userInfoArr)
	game.Table.Broadcast(global.S2C_PUB_COMPARE, &msg.S2CPubCompare{Uid: user.User.GetID()})
}

//客户端发牌动画结束，通知服务器开始倒计时
func (game *Game) ProcSendCardOver(buffer []byte, user *data.User) {
	if user == nil {
		//log.Traceln("ProcSendCardOver user nil ")
		return
	}
	var c2sMsg msg.C2SIntoGame
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	game.SendCardOver(user)
}

func (game *Game) ProcSetCardType(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SSetCardType
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	if len(c2sMsg.Cards) != 3 {
		log.Traceln("长度不为3 ", len(c2sMsg.Cards))
		return
	}
	//log.Traceln("用户设置的牌型为：", c2sMsg.CardType, "  牌为： ", fmt.Sprintf(`%x`, cards))
	user.CardType, user.Cards = poker.GetCardTypeJH(c2sMsg.Cards)
}

func (game *Game) ProcSeeOtherCards(buffer []byte, user *data.User) {
	var c2sMsg msg.C2SSeeOtherCards
	if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
		//log.Traceln("Action proto err : ", err)
		return
	}
	res := &msg.S2CSeeOtherCards{
		UserCards: make([]*msg.S2CUserSeeCards, 0),
	}
	for _, v := range game.userListArr {
		if v != nil {
			res.UserCards = append(res.UserCards, &msg.S2CUserSeeCards{
				UserId: v.User.GetID(), ChairId: int32(v.ChairId),
				CardType: int32(v.CardType), Cards: v.Cards,
			})
		}
	}
	user.User.SendMsg(global.S2C_SEE_OTHER_CARDS, res)
}

//客户端发牌动画结束，通知服务器开始倒计时
func (game *Game) ProcLeaveGame(buffer []byte, user *data.User) {
	//var c2sMsg msg.C2SIntoGame
	//if err := proto.Unmarshal(buffer, &c2sMsg); err != nil {
	//	//log.Traceln("procLeaveGame proto err : ", err)
	//	return
	//}
	//if game.CurStatus == global.TABLE_CUR_STATUS_WAIT_SEND_CARDS {
	//	//log.Traceln("发牌阶段不能退出...")
	//	user.User.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.C2SIntoGame{})
	//	return
	//}
	//if user.CurStatus == global.USER_CUR_STATUS_ING && game.CurStatus == global.TABLE_CUR_STATUS_ING {
	//	//log.Traceln("用户未弃牌，不能离开房间")
	//	user.User.SendMsg(global.ERROR_CODE_CANNOT_LEAVE, &msg.C2SIntoGame{})
	//	return
	//}
	//
	////log.Traceln("user : ", user.Id, " 离开牌桌")
	//game.Table.Broadcast(global.S2C_LEAVE_TABLE, user.GetUserMsgInfo())
	//game.DelUserListMap(user.Id)
	//if len(game.userListMap) <= 1 {
	//	game.EndGame()
	//}
}

//获取玩家的下一位玩家
func (game *Game) GetNextUser(user *data.User) *data.User {
	var nextChair uint = 0
	tmpChairId := user.ChairId
	for i := 0; i <= 5; i++ {
		tmpChairId += 1
		if tmpChairId > 5 {
			tmpChairId = 1
		}
		chairUser := game.GetUserByChairId(tmpChairId)
		if chairUser == nil {
			////log.Traceln("设置下一个发言时没找到：", tmpChairId)
			continue
		}
		if chairUser.CurStatus != global.USER_CUR_STATUS_ING {
			////log.Traceln("下一个 chair ", chairUser.Id, " 状态 为： ", chairUser.CurStatus)
			continue
		}
		nextChair = tmpChairId
		break
	}
	return game.GetUserByChairId(nextChair)
}
