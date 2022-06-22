package model

import (
	"fmt"
	"go-game-sdk/example/game_poker/960305/config"
	baijiale "go-game-sdk/example/game_poker/960305/msg"
	"math/rand"
	"sort"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type SceneUser struct {
	User       *User
	SeatNo     int //座位号
	BetMaxArea int
}

type SceneInfo struct {
	SenceSeat      map[int]*SceneUser   //坐下的玩家列表，座位号为索引
	UserSeat       map[int64]*SceneUser //坐下的玩家列表，玩家ID
	BetZhuangMaxID int64                //下注最多的玩家ID
	BetXianMaxID   int64                //下闲最多的ID

	zhuangBetMax int64
	xianBetMax   int64
	IsZhuangMax  bool // 是否是庄下的最多
}

func SendSceneMessage(table table.TableInterface, user *User) {
	// 发送场景数据  SceneSeatDetail
	//table.Broadcast()
}

func (si *SceneInfo) Reset() {
	si.BetZhuangMaxID = -1
	si.BetXianMaxID = -1
	si.zhuangBetMax = 0
	si.xianBetMax = 0
	si.IsZhuangMax = false
}

func (si *SceneInfo) Init() {
	si.SenceSeat = make(map[int]*SceneUser)
	si.UserSeat = make(map[int64]*SceneUser)
}

// 坐下或换座
func (si *SceneInfo) SitScene(user *User, SeatNum int, level int) bool {
	_, ok1 := si.SenceSeat[SeatNum]
	us, ok := si.UserSeat[user.User.GetID()]
	//原来位置上有人,换位置
	if ok1 {
		//发送坐下位置失败
		msg := new(baijiale.UserSitDownFail)
		msg.FailReaSon = "你坐下的位置已经被其他玩家捷足先登了！"

		user.User.SendMsg(int32(baijiale.SendToClientMessageType_SitDownFail), msg)
		return false
	}

	if user.User.GetScore() < int64(config.LongHuConfig.SitDownLimit[level-1]) {
		msg := new(baijiale.UserSitDownFail)
		str := fmt.Sprintf("入座至少需要携带%d金币", config.LongHuConfig.SitDownLimit[level-1]/100)
		msg.FailReaSon = str

		user.User.SendMsg(int32(baijiale.SendToClientMessageType_SitDownFail), msg)
		return false
	}
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
		si.UserSeat[user.User.GetID()] = newuser
		//广播
	}

	return true
}

// 优先级：大赢家>大富豪>神算子
func (si *SceneInfo) GetBigWinner(i int) *User {
	var users []*User
	for _, v := range si.SenceSeat {
		users = append(users, v.User)
	}

	sort.Sort(BigwinnerUser(users))

	if i < 0 {
		if len(users) > 0 && users[0].LastWinGold > 0 {
			return users[0]
		}
		return nil
	}
	if len(users) > i && users[i].LastWinGold > 0 {
		return users[i]
	}
	// if len(users) > 0 && users[0].LastWinGold > 0 {
	// 	return users[0]
	// }
	return nil
}

// 优先级：大赢家>大富豪>神算子
func (si *SceneInfo) GetMaster(i int) *User {
	var users []*User
	for _, v := range si.SenceSeat {
		users = append(users, v.User)
	}

	sort.Sort(MasterUser(users))

	if i < 0 {
		if len(users) > 0 && users[0].RetWin > 0 {
			return users[0]
		}
		return nil
	}
	if len(users) > i && users[i].RetWin > 0 {
		return users[i]
	}
	// if len(users) > 0 && users[0].RetWin > 0 {
	// 	return users[0]
	// }
	return nil
}

// 优先级：大赢家>大富豪>神算子
func (si *SceneInfo) GetRegal(i int) *User {
	var users []*User
	for _, v := range si.SenceSeat {
		users = append(users, v.User)
	}

	sort.Sort(RegalUser(users))

	if i < 0 {
		if len(users) > 0 && users[0].WinGold > 0 {
			return users[0]
		}
		return nil
	}
	if len(users) > i && users[i].WinGold > 0 {
		return users[i]
	}
	// if len(users) > 0 && users[0].WinGold > 0 {
	// 	return users[0]
	// }
	return nil
}

func (si *SceneInfo) UserStandUp(user player.PlayerInterface) {
	v, ok := si.UserSeat[user.GetID()]
	if ok {
		delete(si.SenceSeat, v.SeatNo)
		delete(si.UserSeat, user.GetID())

		betmaxgold := int64(0)
		betmaxid := int64(0)
		if user.GetID() == si.BetZhuangMaxID {
			for id, u := range si.UserSeat {
				if u.User.BetArea[0] > betmaxgold {
					betmaxgold = u.User.BetArea[0]
					betmaxid = id
				}
			}

			si.BetZhuangMaxID = betmaxid
		}

		betmaxgold = 0
		if user.GetID() == si.BetXianMaxID {
			for id, u := range si.UserSeat {
				if u.User.BetArea[1] > betmaxgold {
					betmaxgold = u.User.BetArea[1]
					betmaxid = id
				}
			}

			si.BetXianMaxID = betmaxid
		}
	}
}

func (si *SceneInfo) GetSitDownUserCount() int {
	return len(si.SenceSeat)
}

//随机获取一个空的椅子ID
func (si *SceneInfo) GetSceneChairId() int {
	var chairid []int
	for i := 1; i <= 6; i++ {
		_, ok := si.SenceSeat[i]
		if !ok {
			chairid = append(chairid, i)
		}
	}

	if len(chairid) != 0 {
		index := rand.Intn(len(chairid))
		return chairid[index]
	} else {
		return 0
	}
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

func (si *SceneInfo) IsSitDown(user player.RobotInterface) bool {
	_, ok := si.UserSeat[user.GetID()]
	return ok
}

func (si *SceneInfo) UserBet(user *User) bool {
	_, ok := si.UserSeat[user.User.GetID()]
	if !ok {
		return false
	}

	if user.BetArea[0] > si.zhuangBetMax {
		si.zhuangBetMax = user.BetArea[0]
		si.BetZhuangMaxID = user.User.GetID()
	}
	if user.BetArea[1] > si.xianBetMax {
		si.xianBetMax = user.BetArea[1]
		si.BetXianMaxID = user.User.GetID()
	}

	return true
	// IsChange := true

	// if si.BetZhuangMaxID == 0 {
	// 	if SceneUser.BetMaxArea == 0 {
	// 		si.BetZhuangMaxID = user.User.GetID()
	// 		return IsChange
	// 	}
	// }

	// if si.BetXianMaxID == 0 {
	// 	if SceneUser.BetMaxArea == 1 {
	// 		si.BetXianMaxID = user.User.GetID()
	// 		return IsChange
	// 	}
	// }

	// zhuangbet, ok := si.UserSeat[si.BetZhuangMaxID]

	// if ok {
	// 	if zhuangbet.User.BetArea[0] < user.BetArea[0] {
	// 		si.BetZhuangMaxID = user.User.GetID()
	// 	}
	// }

	// xianbet, ok1 := si.UserSeat[si.BetXianMaxID]

	// if ok1 {
	// 	if xianbet.User.BetArea[1] < user.BetArea[1] {
	// 		si.BetXianMaxID = user.User.GetID()
	// 	}
	// }
	// if ok1 && ok {
	// 	if zhuangbet.User.BetArea[0] > xianbet.User.BetArea[0] {
	// 		si.IsZhuangMax = true
	// 	} else {
	// 		si.IsZhuangMax = false
	// 	}
	// }

	// return IsChange
}

type BetSort struct {
	UserID  int64
	BetGold int64 // 以此排序
}
type BetSorts []BetSort

func (bss BetSorts) Len() int {
	return len(bss)
}
func (bss BetSorts) Swap(i, j int) {
	bss[i], bss[j] = bss[j], bss[i]
}
func (bss BetSorts) Less(i, j int) bool {
	return bss[i].BetGold > bss[j].BetGold
}

func (si *SceneInfo) FindkthZhuang(index int) int64 {
	var bss BetSorts
	for _, v := range si.UserSeat {
		bss = append(bss, BetSort{
			UserID:  v.User.User.GetID(),
			BetGold: v.User.BetArea[0],
		})
	}
	sort.Sort(bss)
	if len(bss) >= index {
		return bss[index-1].UserID
	}
	return -1
}

func (si *SceneInfo) FindkthXian(index int) int64 {
	var bss BetSorts
	for _, v := range si.UserSeat {
		bss = append(bss, BetSort{
			UserID:  v.User.User.GetID(),
			BetGold: v.User.BetArea[1],
		})
	}
	sort.Sort(bss)

	if len(bss) >= index {
		return bss[index-1].UserID
	}
	return -1
}

// 重新计算座位
func (si *SceneInfo) ReCalc() {

	for _, user := range si.UserSeat {
		if user.User.BetArea[0] > si.zhuangBetMax {
			si.zhuangBetMax = user.User.BetArea[0]
		}
		if user.User.BetArea[1] > si.xianBetMax {
			si.xianBetMax = user.User.BetArea[1]
		}
	}

	switch len(si.SenceSeat) {
	case 0:
		si.BetZhuangMaxID = -1
		si.BetXianMaxID = -1
	case 1:
		for _, user := range si.SenceSeat {
			if user.User.BetArea[0] > user.User.BetArea[1] {
				si.BetZhuangMaxID = user.User.User.GetID()
				si.BetXianMaxID = -1
				si.IsZhuangMax = true
			} else if user.User.BetArea[0] < user.User.BetArea[1] {
				si.BetXianMaxID = user.User.User.GetID()
				si.BetZhuangMaxID = -1
				si.IsZhuangMax = false
			} else {
				if si.IsZhuangMax {
					si.BetZhuangMaxID = user.User.User.GetID()
					si.BetXianMaxID = -1
				} else {
					si.BetXianMaxID = user.User.User.GetID()
					si.BetZhuangMaxID = -1
				}
			}
		}
	default:
		if si.zhuangBetMax > si.xianBetMax {
			index := 1
			for si.BetXianMaxID == si.BetZhuangMaxID && si.BetXianMaxID != -1 {
				si.BetXianMaxID = si.FindkthXian(index)
				index++
			}
			si.IsZhuangMax = true
		} else if si.zhuangBetMax < si.xianBetMax {
			index := 1
			for si.BetZhuangMaxID == si.BetXianMaxID && si.BetZhuangMaxID != -1 {
				si.BetZhuangMaxID = si.FindkthZhuang(index)
				index++
			}
			si.IsZhuangMax = false
		} else {
			if si.IsZhuangMax {
				si.BetXianMaxID = si.BetZhuangMaxID
				index := 1
				for si.BetXianMaxID == si.BetZhuangMaxID && si.BetXianMaxID != -1 {
					si.BetXianMaxID = si.FindkthXian(index)
					index++
				}
			} else {
				si.BetZhuangMaxID = si.BetXianMaxID
				index := 1
				for si.BetZhuangMaxID == si.BetXianMaxID && si.BetZhuangMaxID != -1 {
					si.BetZhuangMaxID = si.FindkthZhuang(index)
					index++
				}
			}
		}
	}

	// 如果下注为0，则不让他眯牌
	// if si.zhuangBetMax == 0 {
	// 	si.BetZhuangMaxID = -1
	// }
	// if si.xianBetMax == 0 {
	// 	si.BetXianMaxID = -1
	// }

	if si.UserSeat[si.BetZhuangMaxID] != nil && si.UserSeat[si.BetZhuangMaxID].User.BetArea[0] == 0 {
		si.BetZhuangMaxID = -1
	}

	if si.UserSeat[si.BetXianMaxID] != nil && si.UserSeat[si.BetXianMaxID].User.BetArea[1] == 0 {
		si.BetXianMaxID = -1
	}
}

type ZhuangBetMaxUser []*SceneUser

func (z ZhuangBetMaxUser) Less(i, j int) bool { return z[i].User.BetArea[0] > z[j].User.BetArea[0] }

func (z ZhuangBetMaxUser) Swap(i, j int) { z[i], z[j] = z[j], z[i] }

func (z ZhuangBetMaxUser) Len() int { return len(z) }

type XianBetMaxUser []*SceneUser

func (z XianBetMaxUser) Less(i, j int) bool { return z[i].User.BetArea[1] > z[j].User.BetArea[1] }

func (z XianBetMaxUser) Swap(i, j int) { z[i], z[j] = z[j], z[i] }

func (z XianBetMaxUser) Len() int { return len(z) }
