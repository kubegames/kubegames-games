package gamelogic

import (
	"go-game-sdk/lib/clock"
	"math/rand"
	"sort"

	"github.com/kubegames/kubegames-games/pkg/slots/990601/config"
	bridanimal "github.com/kubegames/kubegames-games/pkg/slots/990601/msg"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
	"github.com/mohae/deepcopy"
)

type Robot struct {
	game           *Game
	Table          table.TableInterface //桌子
	User           player.RobotInterface
	BetGoldThisSet int64 // 本局下注金额

	BetArrNum  int                    // 下注区域数目 (1-12之间随机)
	BetArrGold [BET_AREA_LENGTH]int64 // 下注区域下注的金额，对应0-11

	TimerJob *clock.Job //时间定时器
	Cfg      *config.RobotConfig

	BetCount    int32 // 下注次数
	NotBetCount int32 // 未下注局数
}

func NewRobot(game *Game) *Robot {
	cfg := deepcopy.Copy(config.Robot).(config.RobotConfig)
	return &Robot{
		game:       game,
		Table:      game.Table,
		Cfg:        &cfg,
		BetArrGold: [BET_AREA_LENGTH]int64{},
	}
}

func (r *Robot) BindUser(user player.RobotInterface) {
	r.User = user
}

//游戏消息
func (r *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(bridanimal.SendToClientMessageType_Status):
		{
			r.BeforBet(buffer)
		}
		break
	}
}

func (r *Robot) BeforBet(buf []byte) {
	msg := new(bridanimal.StatusMessage)
	proto.Unmarshal(buf, msg)
	if msg.Status == int32(bridanimal.GameStatus_BetStatus) {
		r.Reset()
		r.Rand()
		r.TimerJob, _ = r.User.AddTimer(int64(r.Cfg.BetGapNow), r.DoBet)
	} else if msg.Status == int32(bridanimal.GameStatus_EndBetMovie) {
		r.Reset()
		if r.TimerJob != nil {
			r.TimerJob.Cancel()
		}
		r.TimerJob = nil
	}

}

func (r *Robot) DoBet() {
	if r.game.Status != int32(bridanimal.GameStatus_BetStatus) {
		return
	}
	base, _ := r.game.getBaseBet()
	randIndex := r.Cfg.BetModChoose.Rand(r.User.GetScore(), base)
	betGold := int64(r.game.BetArr[randIndex])

	// 机器人下注上限
	if r.BetArrGold[randIndex]+betGold > r.game.BetLimitInfo.BetArea || r.BetGoldThisSet+betGold > r.game.BetLimitInfo.AllBetAreaLimit {
		r.TimerJob = nil
		r.AddBetTimer()
		return
	}

	r.BetGoldThisSet += betGold
	r.BetCount++

	// 下注区域的选择
	betTotalInfo := [12]int64{}
	for i := range r.game.AITotalBet {
		betTotalInfo[i] = r.game.AITotalBet[i] + r.game.TotalBet[i]
	}
	betIndex := r.Cfg.BetAreaProb.Rand(betTotalInfo)
	r.BetArrGold[betIndex] += betGold

	// r.game.AITotalBet[betIndex] += betGold
	// 给user下注
	if u, ok := r.game.UserInfoList[r.User.GetID()]; ok {
		u.BetInfo[betIndex] += betGold
		u.Totol += betGold
	}
	r.sendMsg(betIndex, randIndex)
	r.AddBetTimer()
}

func (r *Robot) sendMsg(typ, index int) {
	msg := new(bridanimal.UserBet)
	msg.BetType = int32(typ)
	msg.BetIndex = int32(index)
	r.User.SendMsgToServer(int32(bridanimal.ReceiveMessageType_BetID), msg)
}

func (r *Robot) Rand() {
	r.Cfg.Rand()
	r.BetArrNum = rand.Intn(BET_AREA_LENGTH) + 1
}

func (r *Robot) Reset() {
	if r.TimerJob != nil {
		r.TimerJob.Cancel()
		r.TimerJob = nil
	}
	r.BetCount = 0
	r.NotBetCount = 0
	r.BetArrNum = 0
	r.BetGoldThisSet = 0
	r.BetArrGold = [BET_AREA_LENGTH]int64{}
}

type BetModProb []int

// 从大到小排序
func (b BetModProb) Less(i, j int) bool {
	if b[i] >= b[j] {
		return true
	}
	return false
}

func (b BetModProb) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b BetModProb) Len() int {
	return len(b)
}

func (b BetModProb) Rand() int {
	if !sort.IsSorted(b) {
		sort.Sort(b)
	}

	var allweight int
	for _, v := range b {
		allweight += int(v)
	}
	randweight := rand.Intn(allweight) + 1

	for i, v := range b {
		if randweight <= int(v) {
			return i
		}
		randweight -= int(v)
	}
	return 0
}

func (r *Robot) AddBetTimer() {
	//达到限制条件后不下注
	if r.BetCount >= r.Cfg.BetNumUpline || r.User.GetScore() < int64(r.game.BetArr[0]) {
		if r.TimerJob != nil {
			r.TimerJob.Cancel()
			r.TimerJob = nil
		}
		return
	}
	r.Cfg.Rand()
	r.TimerJob, _ = r.User.AddTimer(int64(r.Cfg.BetGapNow), r.DoBet)
}
