package model

import (
	"game_frame_v2/game/inter"
	"math/rand"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type SceneUser struct {
	User   *User
	SeatNo int //座位号
}

type SceneInfo struct {
	SenceSeat map[int]*SceneUser   //坐下的玩家列表，座位号为索引
	UserSeat  map[int64]*SceneUser //坐下的玩家列表，玩家ID
}

func SendSceneMessage(table table.TableInterface, user *User) {
	// 发送场景数据  SceneSeatDetail
	//table.Broadcast()
}

func (si *SceneInfo) Init() {
	si.SenceSeat = make(map[int]*SceneUser)
	si.UserSeat = make(map[int64]*SceneUser)
}

// 坐下或换座
func (si *SceneInfo) SitScene(user *User, SeatNum int) bool {
	_, ok1 := si.SenceSeat[SeatNum]
	us, ok := si.UserSeat[user.User.GetId()]
	//原来位置上有人,换位置
	if ok1 {
		//发送坐下位置失败
		msg := new(longhu.UserSitDownFail)
		msg.FailReaSon = "你坐下的位置已经被其他玩家捷足先登了！"

		user.User.SendMsg(int32(longhu.SendToClientMessageType_SitDownFail), msg)
		return false
	}

	//if user.User.GetScore() < int64(user.Rule.SitDownLimit) {
	//	//msg := new(longhu.UserSitDownFail)
	//	//str := fmt.Sprintf("低于%d不能入座！", user.Rule.SitDownLimit/100)
	//	//msg.FailReaSon = str
	//
	//	//user.User.SendMsg(int32(longhu.SendToClientMessageType_SitDownFail), msg)
	//	return false
	//}
	//原来位置换成新的位置
	if ok {
		si.SenceSeat[SeatNum] = us
		delete(si.SenceSeat, us.SeatNo)
		us.SeatNo = SeatNum
		//广播
	} else {
		newuser := new(SceneUser)
		newuser.User = user
		newuser.SeatNo = SeatNum
		si.SenceSeat[SeatNum] = newuser
		si.UserSeat[user.User.GetId()] = newuser
		//广播
	}

	return true
}

func (si *SceneInfo) GetBigWinner() int {
	money := int64(0)
	id := 0
	var u *SceneUser
	for _, v := range si.SenceSeat {
		if money < v.User.RetBet || id == 0 {
			money = v.User.RetBet
			id = v.User.SceneChairId
			u = v
		} else if money == v.User.RetBet && v.User.User.GetId() > u.User.User.GetId() {
			money = v.User.RetBet
			id = v.User.SceneChairId
			u = v
		}
	}

	return id
}

func (si *SceneInfo) GetMaster() int {
	count := 0
	id := 0
	var u *SceneUser
	for _, v := range si.SenceSeat {
		if count < v.User.RetWin || id == 0 {
			count = v.User.RetWin
			id = v.User.SceneChairId
			u = v
		} else if count == v.User.RetWin && v.User.User.GetId() > u.User.User.GetId() {
			count = v.User.RetWin
			id = v.User.SceneChairId
			u = v
		}
	}

	return id
}

func (si *SceneInfo) UserStandUp(user player.PlayerInterface) {
	v, ok := si.UserSeat[user.GetId()]
	if ok {
		delete(si.SenceSeat, v.SeatNo)
		delete(si.UserSeat, user.GetId())
	}
}

func (si *SceneInfo) GetSitDownUserCount() int {
	return len(si.SenceSeat)
}

//随机获取一个空的椅子ID
func (si *SceneInfo) GetSceneChairId() int {
	//var chairid []int
	for i := 1; i <= 6; i++ {
		_, ok := si.SenceSeat[i]
		if !ok {
			//chairid = append(chairid, i)
			return i
		}
	}
	return 0

	//if len(chairid) != 0 {
	//	index := rand.Intn(len(chairid))
	//	return chairid[index]
	//} else {
	//	return 0
	//}
}

func (si *SceneInfo) GetAiUser() player.PlayerInterface {
	var aiuser []player.PlayerInterface
	for _, v := range si.SenceSeat {
		if v.User.User.IsRobot() {
			aiuser = append(aiuser, v.User.User)
		}
	}

	if len(aiuser) > 0 {
		index := rand.Intn(len(aiuser))
		return aiuser[index]
	}

	return nil
}

func (si *SceneInfo) IsSitDown(user inter.AIUserInter) bool {
	_, ok := si.UserSeat[user.GetId()]
	return ok
}

//获取椅子上所有用户ID
func (si *SceneInfo) CheckUserOnChair(userId int64) bool {
	_, ok := si.UserSeat[userId]
	if ok {
		//用户在椅子上
		return false
	}
	//用户不在椅子上
	return true
}
