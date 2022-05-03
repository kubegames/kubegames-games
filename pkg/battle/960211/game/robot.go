package game

import (
	"fmt"
	"math/rand"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-games/pkg/battle/960211/config"
	pai9 "github.com/kubegames/kubegames-games/pkg/battle/960211/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
)

type Robot struct {
	user player.RobotInterface
	job  *player.Job
	//chairID int
}

func (robot *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	if subCmd == int32(pai9.SendToClientMessageType_Status) {
		msg := new(pai9.StatusMessage)
		if err := proto.Unmarshal(buffer, msg); err != nil {
			log.Errorf("机器人解析消息[%d]错误：%v", subCmd, err)
		}
		switch msg.Status {
		case int32(pai9.GameStatus_QiangZhuang):
			go robot.Qiang()
		case int32(pai9.GameStatus_Bet):
			go robot.Bet()
		case int32(pai9.GameStatus_SettleAll):
			robot.leave()
		}
	}
}

func NewRobot() *Robot {
	return &Robot{}
}

// 机器人抢庄
func (robot *Robot) Qiang() {
	gap := robot.getGapTime(true)
	fmt.Printf("机器人【%d】抢庄间隔  %d\n", robot.user.GetID(), gap)
	robot.job, _ = robot.user.AddTimer(gap, robot.realQiang)
}

func (robot *Robot) realQiang() {
	msg := new(pai9.QiangZhuangReqMsg)
	msg.UserID = int32(robot.user.GetID())
	// 随机抢庄率
	qiangProb := rand.Intn(100) + 1
	var qiangIndex int
	for i, v := range config.Robot.QiangProb {
		if int32(qiangProb) <= v {
			qiangIndex = i
			break
		}
		qiangProb -= int(v)
	}
	msg.Index = int32(qiangIndex)
	log.Debugf("机器人ID = %d 抢庄索引=%d \n", robot.user.GetID(), msg.Index)
	if err := robot.user.SendMsgToServer(int32(pai9.ReceiveMessageType_QiangZhuangReq), msg); err != nil {
		log.Errorf("机器人发送抢庄消息错误：%v", err)
	}
}

// 机器人下注
func (robot *Robot) Bet() {
	gap := robot.getGapTime(true)
	fmt.Printf("机器人【%d】下注间隔  %d\n", robot.user.GetID(), gap)
	robot.job, _ = robot.user.AddTimer(gap, robot.realBet)
}

func (robot *Robot) realBet() {
	msg := new(pai9.BetMultiReqMsg)
	msg.UserID = int32(robot.user.GetID())

	// 随机抢庄率
	betProb := rand.Intn(100) + 1
	var betIndex int
	for i, v := range config.Robot.BetProb {
		if int32(betProb) <= v {
			betIndex = i
			break
		}
		betProb -= int(v)
	}
	msg.Index = int32(betIndex)
	log.Debugf("机器人ID = %d 下注索引=%d \n", robot.user.GetID(), msg.Index)
	if err := robot.user.SendMsgToServer(int32(pai9.ReceiveMessageType_BetMultiReq), msg); err != nil {
		log.Errorf("机器人发送下注消息错误：%v", err)
	}
}

// 获取间隔时间,返回ms
func (robot *Robot) getGapTime(isQiang bool) (t int64) {
	defer func() {
		if isQiang {
			t += 2000
			return
		}
	}()
	// var wait int
	// if isQiang {
	// 	wait = config.Pai9Config.Taketimes.QiangZhuang
	// } else {
	// 	wait = config.Pai9Config.Taketimes.Bet
	// }
	timeProb := rand.Intn(100) + 1
	var i int
	var v int32
	for i, v = range config.Robot.TimeProb {
		if int32(timeProb) <= v {
			t = int64((i + 1) * 1000)
			return
		}
		timeProb -= int(v)
	}
	t = int64((i + 1) * 1000)
	return
}

// 离开房间
func (robot *Robot) leave() {

}
