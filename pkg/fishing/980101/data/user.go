package data

import (
	"github.com/kubegames/kubegames-games/pkg/fishing/980101/msg"
	frameMsg "github.com/kubegames/kubegames-sdk/app/message"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Table         table.TableInterface
	UserInfo      *msg.UserInfo
	Protect       int32
	LastShootTime int64
	IsRobot       bool
	InnerUser     player.PlayerInterface
	SubScore      int64
	AddScore      int64
	TaxedScore    int64
	OutputAmount  int64
	Bet           int64  //下注
	Win           int64  //总共赢的钱
	GameNum       string //局号
	Log           []*frameMsg.GameLog
	BulletNum     int
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
