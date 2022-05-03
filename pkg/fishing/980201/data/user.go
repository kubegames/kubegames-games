package data

import (
	"go-game-sdk/example/game_buyu/980201/msg"

	frameMsg "github.com/kubegames/kubegames-sdk/app/message"

	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Table          table.TableInterface
	UserInfo       *msg.UserInfo
	Protect        int32
	LastShootTime  int64
	IsRobot        bool
	InnerUser      player.PlayerInterface
	SkillFishInfos map[int32]SkillFishInfo
	SkillNum       map[int32]int
	SubScore       int64
	AddScore       int64
	TaxedScore     int64
	OutputAmount   int64  //税后赢的钱
	Bet            int64  //下注
	Win            int64  //总共赢的钱
	GameNum        string //局号
	Log            []*frameMsg.GameLog
	BulletNum      int
}

type SkillFishInfo struct {
	StartTime int64
	EndTime   int64
	Mult      int32
	Score     int32
	BulletLv  int32
	Dur       int64
	Shoot     bool
	FishId    string
}

func NewUser(table table.TableInterface) *User {
	return &User{
		Table: table,
	}
}

func (user *User) addProtect(value int32) {
	user.Protect += value
}

func (user *User) WriteLog() {
	user.InnerUser.SendLogs(user.GameNum, user.Log)
	user.Log = make([]*frameMsg.GameLog, 0)
}
