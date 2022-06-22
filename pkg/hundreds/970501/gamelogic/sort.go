package gamelogic

import (
	"math/rand"
	"sort"

	"github.com/kubegames/kubegames-games/pkg/slots/970501/model"
	proto "github.com/kubegames/kubegames-games/pkg/slots/970501/msg"
)

type UserSettleInfos []*proto.UserSettleInfo

// 从大到小排序
func (u UserSettleInfos) Less(i, j int) bool {
	if u[i].WinGold <= u[j].WinGold {
		return false
	}
	return true
}

func (u UserSettleInfos) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u UserSettleInfos) Len() int {
	return len(u)
}

type UserList []*User

// 携带金额从高到低
func (u UserList) Less(i, j int) bool {
	return u[i].WinGold > u[j].WinGold
}

func (u UserList) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u UserList) Len() int {
	return len(u)
}

func (u UserList) sort() {
	sort.Sort(u)
}

type TopUsers []*User

// 从大到小排序
func (u TopUsers) Less(i, j int) bool {
	if u[i].user.GetScore() <= u[j].user.GetScore() {
		return false
	}
	return true
}

func (u TopUsers) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}

func (u TopUsers) Len() int {
	return len(u)
}

func (u TopUsers) sort() {
	sort.Sort(u)
}

type AreaGold struct {
	ElemType model.ElementType // 元素类型
	PayGold  int64             // 赔付金币
	BetIndex int               // 下注索引
}
type AreaGolds []AreaGold

func (a AreaGolds) Less(i, j int) bool {
	if a[i].PayGold <= a[j].PayGold {
		return true
	}
	return false
}

func (a AreaGolds) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a AreaGolds) Len() int {
	return len(a)
}

func (a AreaGolds) GetMax() AreaGolds {
	if len(a) == 1 {
		return a
	}
	return a[len(a)/2:]
}

func (a AreaGolds) GetMin() AreaGolds {
	if len(a) == 1 {
		return a
	}
	return a[:len(a)/2]
}

func (a AreaGolds) Rand() AreaGold {
	return a[rand.Intn(len(a))]
}
