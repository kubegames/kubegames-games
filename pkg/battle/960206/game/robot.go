package game

import (
	"sync"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type Robot struct {
	game   *Game
	AiUser player.RobotInterface
	user   *data.User
}

func NewRobot(game *Game, user *data.User) *Robot {
	return &Robot{
		game: game, user: user,
	}
}

///////////////////////
var redRobCountMap = make(map[int64]int) // redId => 安排来抢包的人数
var redRobCountLock sync.Mutex

func SetRedRobCountMap(redId int64) {
	redRobCountLock.Lock()
	redRobCountMap[redId]++
	redRobCountLock.Unlock()
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(msg.S2CMsgType_START_GAME):

		var (
			showdownLv int // 延迟摊牌level
			addRate    int // 概率累加器
		)

		// 摊牌延迟时间权重
		weight := rand.RandInt(1, 101)

		// 确定延迟摊牌level
		for index, rate := range config.RobotConf.ShowdownRate {
			if weight > rate && weight <= rate+addRate {
				showdownLv = index
				break
			}
			addRate += rate
		}

		// 随机延迟摊牌时间
		randShowdownTime := rand.RandInt(config.RobotConf.ShowdownTime[showdownLv][0], config.RobotConf.ShowdownTime[showdownLv][0]+1)
		log.Tracef("机器人 %d 的摊牌延迟时间：%d s", robot.user.User.GetID(), randShowdownTime)
		//fmt.Println("机器人定时发送确定摆牌")
		robot.game.Table.AddTimer(int64(randShowdownTime*1000), func() {
			_ = robot.AiUser.SendMsgToServer(int32(msg.C2SMsgType_SET_CARDS), &msg.C2SSetCards{
				IsAuto: true,
			})
		})

		//const a = 8
		//robot.game.Table.AddTimer(a*1000, func() {
		//	for _, user := range game.userList {
		//		if user.User.IsRobot() {
		//			game.Table.Broadcast(int32(msg.S2CMsgType_SETTLE_CARDS), &msg.S2CSettleCards{
		//				Uid:         user.User.GetID(),
		//				ChairId:     user.ChairId,
		//				SpecialType: user.SpecialCardType,
		//			})
		//			user.IsSettleCards = true
		//			game.AllUserSettleEndGame()
		//			time.Sleep(2*time.Second)
		//		}
		//	}
		//})
	}
}
