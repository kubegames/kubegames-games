package glogic

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960205/msg"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Robot struct {
	User         player.RobotInterface
	GameLogic    *ErBaGangGame
	BetCount     int         //下注限制
	TimerJob     *player.Job //job
	LastBetPlace int         //机器人上次下注的区域
}

func (r *Robot) Init(User player.RobotInterface, g table.TableHandler) {
	r.User = User
	r.GameLogic = g.(*ErBaGangGame)
}

func (r *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(msg.SendToClientMessageType_S2CRobZhuangStart): // 用户按下抢庄按钮
		r.ReqRobZhuangEnd(buffer)
	case int32(msg.SendToClientMessageType_S2CUserBetInfoStart): // 用户按下下注按钮
		r.ReqUserBetEnd(buffer)
	}
}
