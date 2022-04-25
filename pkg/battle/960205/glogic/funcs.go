package glogic

import (
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

// ReadRobBtnConf 读取抢庄配置
func (game *ErBaGangGame) ReadRobBtnConf(userIndex int) []int32 {
	log.Tracef("ReadRobBtnConf", userIndex, game.UserAllList)
	user := game.UserAllList[userIndex]
	n := int64(0)
	if user != nil {
		n = user.InterUser.GetScore() / int64(game.DiZhu)
	}

	if n >= 200 {
		n = 200
	}

	n0 := 3

	n1 := n - 2*(n/3)
	if n1 <= 3 || n1 > n {
		n1 = 0
	}

	n2 := n - n/3
	if n2 <= 3 || n2 > n {
		n2 = 0
	}

	n3 := n
	if n3 <= 3 || n3 > n {
		n3 = 0
	}

	if n < 3 {
		return []int32{0, 0, 0, 0}
	} else {
		return []int32{int32(n0), int32(n1), int32(n2), int32(n3)}
	}

}

// ReadBtnBtnConf 读取下注配置
func (game *ErBaGangGame) ReadBtnBtnConf(userIndex int) []int32 {
	// 庄家倍数
	n := game.RobZhuangMultipleList[game.RobZhuangIndex]
	user := game.UserAllList[userIndex]
	n4 := int64(0)
	if user != nil {
		n4 = user.InterUser.GetScore() / int64(game.DiZhu)
	}
	n0 := n / 12
	if n0 > n4 || n0 == 1 {
		n0 = 0
	}
	n1 := n / 6
	if n1 > n4 || n1 == 1 {
		n1 = 0
	}
	n2 := n / 4
	if n2 > n4 || n2 == 1 {
		n2 = 0
	}
	n3 := n / 3
	if n3 > n4 || n3 == 1 {
		n3 = 0
	}
	return []int32{int32(n0), int32(n1), int32(n2), int32(n3)}
}
