package game

import (
	"game_buyu/rob_red/config"
	"game_buyu/rob_red/data"
	"game_buyu/rob_red/global"
	"game_buyu/rob_red/msg"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
)

type Red struct {
	Id               int64
	sender           *data.User //红包发送者
	Amount           int64      //红包余额
	Count            int64      //红包数量
	RobbedCount      int64      //已抢红包数量
	Status           int32      // 状态
	MineNum          int64      //雷号
	Route            int32      //[]*msg.S2CAxis //红包路径
	game             *Game
	StartPoint       int32 //红包飞进界面的点，一共5个点
	Time             int64 //红包发送时间
	Lock             sync.Mutex
	Speed            []int32 //红包速度
	RedFlood         int64   //红包血量
	FlyTime          time.Time
	OriginAmount     int64 //发送红包时的金额
	nowSecond        int   //红包出现的秒数
	level            int32 //红包等级
	SelfRobbedAmount int64 //自己抢了的金额
}

func NewRed(sender *data.User, amount, count int64, game *Game, startPoint int32, mineNum int64, speed []int32) *Red {
	red := &Red{
		sender: sender, Amount: amount, Count: count, Status: global.RED_CUR_STATUS_ING, MineNum: mineNum,
		StartPoint: startPoint, Time: time.Now().Unix(), game: game, Speed: speed,
		RedFlood: config.RedConfig.RedFlood, FlyTime: time.Now(), OriginAmount: amount, nowSecond: -10,
	}
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
		RedId: red.Id, Amount: red.Amount, Count: red.Count, Route: red.Route, RobbedCount: red.RobbedCount,
		MineNum: red.MineNum, Status: red.Status, SenderId: red.sender.Id, IsAi: red.sender.IsAi, Time: red.Time,
		RedFlood: config.RedConfig.RedFlood, SenderName: red.sender.Name, Level: red.level,
	}
	if s2cRedInfo.RobbedCount == 0 {
		s2cRedInfo.RobbedCount = -1
	}
	if s2cRedInfo.Status == 0 {
		s2cRedInfo.Status = -1
	}

	return s2cRedInfo
}

//score : 玩家积分
func (red *Red) NewRobbedRed(isMine bool, robbedAmount, score int64) *msg.S2CRobRed {
	s2cRobRed := &msg.S2CRobRed{
		RedId: red.Id, IsMine: isMine, RobbedAmount: robbedAmount, Amount: red.OriginAmount, NotRobbedCount: red.RedFlood - red.RobbedCount,
		SenderName: red.sender.Name, MineNum: red.MineNum, Level: red.level, Score: score,
	}
	return s2cRobRed
}

func (red *Red) NewUserRobbedRedInfo(robbedAmount int64, level string, isMine bool) *msg.S2CRobbedRedInfo {
	s2cRedInfo := &msg.S2CRobbedRedInfo{
		Time: time.Now().Unix(), SenderName: red.sender.Name, RobbedAmount: robbedAmount, Level: level, RedAmount: red.OriginAmount,
		IsMine: isMine, MineNum: red.MineNum, RedId: red.Id,
	}
	return s2cRedInfo
}

func (red *Red) NewMineRecord2C(level string, robbedAmount, mineAmount int64) *msg.S2CMinRecord {
	s2cRobRed := &msg.S2CMinRecord{
		Time: time.Now().Unix(), Level: level, RobbedAmount: robbedAmount, SenderName: red.sender.Name, MineAmount: mineAmount,
		MineNum: red.MineNum,
	}
	return s2cRobRed
}

func (red *Red) NewSendRedRecord2C(level string) *msg.S2CSendRedRecord {
	s2cSend := &msg.S2CSendRedRecord{
		Time: time.Now().Unix(), Level: level, RedAmount: red.OriginAmount, Status: red.Status,
		MineNum: red.MineNum, RedId: red.Id,
	}
	return s2cSend
}

func (red *Red) GetLevel(baseAmount int64) int32 {
	if red.Amount == baseAmount {
		return 1
	}
	if red.Amount == baseAmount*2 {
		return 2
	}
	if red.Amount == baseAmount*3 {
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
	//if red.Amount < 10 {
	//	return red.Amount / 2
	//}
	max := red.Amount / (red.RedFlood - red.RobbedCount) * 2
	if max <= 1 {
		return max
	}
	amount := rand.RandInt(1, int(max))
	if !robUser.User.IsRobot() && !red.sender.User.IsRobot() {
		//log.Traceln("抢包人和发包人都是真实玩家，不做控制",robUser.User.GetID(),red.sender.User.GetID())
		return int64(amount)
	}
	if isMine {
		amount /= 10
		amount *= 10
		amount += int(red.MineNum)
	} else {
		//如果不中雷
		if int64(amount)%10 == red.MineNum {
			amount--
		}
	}
	return int64(amount)
}

func (red *Red) NewRobbedCount2C() *msg.S2CRobbedCount {
	return &msg.S2CRobbedCount{RedId: red.Id, RobbedCount: red.RobbedCount, Level: red.level}
}
