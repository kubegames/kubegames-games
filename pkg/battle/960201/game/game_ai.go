package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
)

//机器人通过概率进入分配
func (zjh *Game) AiIntoRoom() {
	//return
	//fmt.Println("分配机器人 ... ")
	zjh.Table.AddTimer(2*1000, func() {
		switch zjh.GetTableUserCount() {
		case 1:
			// 60=>4个人  25=>5个人  15=>3个人
			index := rand.RandInt(0, 100)
			if index < 60 {
				zjh.Table.GetRobot(3, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			if index < 85 {
				zjh.Table.GetRobot(4, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			zjh.Table.GetRobot(2, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
			return
		case 2:
			index := rand.RandInt(0, 100)
			if index < 60 {
				zjh.Table.GetRobot(2, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			if index < 85 {
				zjh.Table.GetRobot(3, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			zjh.Table.GetRobot(1, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
			return
		case 3:
			index := rand.RandInt(0, 100)
			if index < 60 {
				zjh.Table.GetRobot(1, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			if index < 85 {
				zjh.Table.GetRobot(2, zjh.Table.GetConfig().RobotMinBalance, zjh.Table.GetConfig().RobotMaxBalance)
				return
			}
			return
		}
	})
}

//获取各阶段加注金额
func (zjh *Game) getRaiseAmount() (amount int64) {

	var index = 0
	if zjh.MinAction < zjh.GameConfig.RaiseAmount[0] {
		index = rand.RandInt(0, 2)
	} else if zjh.MinAction < zjh.GameConfig.RaiseAmount[1] {
		index = rand.RandInt(1, 2)
	} else {
		index = 2
	}
	amount = zjh.GameConfig.RaiseAmount[index]
	return
}

//获取要比牌的用户
func (zjh *Game) getCompareUser() (userList []*data.User) {
	userList = make([]*data.User, 0)
	for _, v := range zjh.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		if v.CompareWinCount > 0 && v.Id != zjh.CurActionUser.Id {
			userList = append(userList, v)
			return
		}
		if v.IsSawCards && v.Id != zjh.CurActionUser.Id {
			userList = append(userList, v)
			return
		}
	}
	for _, v := range zjh.GetStatusUserList(global.USER_CUR_STATUS_ING) {
		if v.Id != zjh.CurActionUser.Id {
			userList = append(userList, v)
			return
		}
	}
	return
}
