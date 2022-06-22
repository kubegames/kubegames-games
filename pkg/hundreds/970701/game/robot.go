package game

import (
	"go-game-sdk/example/game_poker/saima/config"
	"go-game-sdk/example/game_poker/saima/msg"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Robot struct {
	User    player.RobotInterface
	Table   table.TableInterface
	Isleave bool
}

func (robot *Robot) Init(user player.RobotInterface, table table.TableInterface) {
	robot.User = user
	robot.Table = table
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(msg.MsgId_Status_Change_Res):
		robot.statusChange(buffer)
		break
		//case int32(msg.MsgId_Count_Res):
		//	robot.checkLeave()
		//	break
	}
}

//检测离开
func (robot *Robot) checkLeave() {
	score := robot.User.GetScore()
	if score > int64(config.GetRobotMaxCoin(int(robot.Table.GetLevel()))) ||
		score <= int64(config.GetRobotMinCoin(int(robot.Table.GetLevel()))) {
		robot.Isleave = true
	}
	if robot.Isleave {
		robot.User.LeaveRoom()
	}
}

//状态改变
func (robot *Robot) statusChange(buffer []byte) {
	res := &msg.GameStatusChangeRes{}
	proto.Unmarshal(buffer, res)
	if res.GetGameStatus() == msg.GameStatus_game_Start {
		robot.checkLeave()
	}
	if res.GetGameStatus() == msg.GameStatus_game_Bet {
		num := rand.Intn(5)
		if num > 0 {
			t := int(res.GetWaitTime()-1000) / num
			robot.User.AddTimerRepeat(int64(t), uint64(num), robot.bet)
		}
	}
}

//下注
func (robot *Robot) bet() {
	if robot.Isleave {
		return
	}
	index := rand.Intn(5) + 1
	area := rand.Intn(38) + 1
	req := &msg.BetReq{
		UserId:  robot.User.GetID(),
		BetArea: msg.BetArea(area),
		Index:   int32(index),
	}
	robot.User.SendMsgToServer(int32(msg.MsgId_Bet_Req), req)
}
