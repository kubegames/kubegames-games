package game

import (
	"common/rand"
	"game_buyu/crazy_red/config"
	"game_buyu/crazy_red/data"
	"game_buyu/crazy_red/global"
	"game_buyu/crazy_red/msg"
	"sync"
	"time"

	"github.com/kubegames/kubegames-sdk/pkg/log"
)

type Red struct {
	Id               int64
	sender           *data.User //红包发送者
	Amount           int64      //红包余额
	RobbedCount      int64      //已抢红包数量
	Status           int32      // 状态
	MineNum          int64      //雷号
	game             *Game
	Time             int64 //红包发送时间
	Lock             *sync.Mutex
	RedFlood         int64 //红包血量
	FlyTime          time.Time
	OriginAmount     int64 //发送红包时的金额
	nowSecond        int   //红包出现的秒数
	level            int32 //红包等级
	SelfRobbedAmount int64
}

func NewRed(sender *data.User, amount int64, game *Game, mineNum int64) *Red {
	red := &Red{
		sender: sender, Amount: amount, Status: global.RED_CUR_STATUS_READY, MineNum: mineNum,
		Time: time.Now().Unix(), Lock: new(sync.Mutex), game: game,
		RedFlood: config.RedConfig.RedFlood, FlyTime: time.Now(), OriginAmount: amount, nowSecond: -10,
	}
	red.level = red.GetLevel()
	return red
}

//红包金额减少
func (red *Red) SubAmount(amount int64) {
	red.Amount -= amount
}

//红包被抢次数增加
func (red *Red) AddRobbedCount() {
	red.RobbedCount++
}

func (red *Red) GetRedInfo2C() *msg.S2CRedInfo {
	s2cRedInfo := &msg.S2CRedInfo{
		RedId: red.Id, Amount: red.Amount, RobbedCount: red.RobbedCount,
		MineNum: red.MineNum, Status: red.Status, SenderId: red.sender.Id, Time: red.Time,
		RedFlood: config.RedConfig.RedFlood, SenderName: red.sender.Name, Level: red.level,
		Ticker: global.TICKER_TIME_START_ROB, SenderHead: red.sender.User.GetHead(),
	}
	if s2cRedInfo.RobbedCount == 0 {
		s2cRedInfo.RobbedCount = -1
	}
	if s2cRedInfo.Status == 0 {
		s2cRedInfo.Status = -1
	}

	return s2cRedInfo
}

func (red *Red) GetCurRed2C() *msg.S2CCurRed {
	return &msg.S2CCurRed{
		RedId: red.Id, UserName: red.sender.Name, Amount: red.OriginAmount, MineNum: red.MineNum,
		SenderId: red.sender.User.GetId(),
	}
}

//score : 玩家积分
func (red *Red) NewRobbedRed(isMine bool, robAmount int64, robUser *data.User) *msg.S2CRobRed {
	if !robUser.User.IsRobot() {
		log.Traceln("发送金额：", robUser.User.GetId(), "  ", robUser.User.GetScore())
	}
	s2cRobRed := &msg.S2CRobRed{
		RedId: red.Id, IsMine: isMine, RobbedAmount: robAmount, Amount: red.OriginAmount, NotRobbedCount: red.RedFlood - red.RobbedCount,
		SenderName: red.sender.Name, MineNum: red.MineNum, Level: red.level, Score: robUser.User.GetScore(), UserId: robUser.User.GetId(),
		RobberHead: robUser.User.GetHead(), RobberName: robUser.User.GetNike(),
	}
	return s2cRobRed
}

func (red *Red) NewUserRobbedRedInfo(amount int64, level int32, isMine bool) *msg.S2CRobbedRedInfo {
	s2cRedInfo := &msg.S2CRobbedRedInfo{
		Time: time.Now().Unix(), SenderName: red.sender.Name, Amount: amount, Level: level, RedAmount: red.Amount,
		IsMine: isMine, MineNum: red.MineNum,
	}
	return s2cRedInfo
}

func (red *Red) NewSendRedRecord2C(level int32) *msg.S2CSendRedRecord {
	s2cSend := &msg.S2CSendRedRecord{
		Time: time.Now().Unix(), Level: level, RedAmount: red.OriginAmount, Status: red.Status,
		RedId: red.Id,
	}
	return s2cSend
}

func GetRouteAxis() int32 {
	//return 1
	return int32(rand.RandInt(1, 4))
}

func (red *Red) GetLevel() int32 {
	if red.Amount >= 500 && red.Amount < 1000 {
		return 1
	}
	if red.Amount >= 1000 && red.Amount < 1500 {
		return 2
	}
	return 3
	if red.Amount >= 1500 && red.Amount < 2000 {
		return 3
	}
	return 4
}

//获取该次抢包金额
func (red *Red) GetRobAmount(robUser *data.User, isMine bool) int64 {
	if red.RobbedCount == red.RedFlood-1 {
		//log.Traceln("最后一个血量：",red.Amount)
		return red.Amount
	}
	if red.Amount < 10 {
		return red.Amount / 2
	}
	max := red.Amount / (red.RedFlood - red.RobbedCount) * 2
	if max <= 1 {
		return max
	}
	amount := rand.RandInt(1, int(max))
	if !robUser.User.IsRobot() && !red.sender.User.IsRobot() {
		log.Traceln("抢包人和发包人都是真实玩家，不做控制", robUser.User.GetId(), red.sender.User.GetId())
		return int64(amount)
	}
	if isMine {
		amount /= 10
		amount *= 10
		amount += int(red.MineNum)
	} else {
		if amount%10 == int(red.MineNum) {
			amount -= 1
		}
	}
	return int64(amount)
}
