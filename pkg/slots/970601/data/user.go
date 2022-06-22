package data

import (
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/global"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Status              int32                  //玩家当前状态，0：正在游戏，1：比牌输了，2：弃牌
	Table               table.TableInterface   `json:"-"`
	User                player.PlayerInterface `json:"-"`
	IsIntoSmallGame     bool
	IsNormalQuit        bool
	Cheat               int32
	LastTime            time.Time
	CurBox              int32 //用户当前消了多少个
	Level               int32 //保留24小时用户当前关卡
	TotalInvest         int64 //玩家总下注，要累计，退出才清零
	TotalInvestForCount int64 //统计用-用户总投入，直到退出才清零
	TotalWin            int64 // 统计-总收益
	WinCount            int   // 统计-中奖次数
	KeyCount            int   // 统计-钻头掉落次数
	Level1Count         int   // 统计-通1关次数
	Level2Count         int   // 统计-通2关次数
	Level3Count         int   // 统计-通3关次数
	TotalCaijin         int64 //统计-彩金池总金额
	Caijin1Count        int   //统计-彩金1档次数
	Caijin2Count        int   //统计-彩金2档次数
	Caijin3Count        int   //统计-彩金3档次数
	Caijin4Count        int   //统计-彩金4档次数
	Caijin5Count        int   //统计-彩金5档次数
	WinSerial1          int   //统计-单次中奖1次
	WinSerial2          int   //统计-单次中奖2次
	WinSerial3          int   //统计-单次中奖3次
	WinSerial4          int   //统计-单次中奖4次
	WinSerial5          int   //统计-单次中奖5次
	WinSerialOver5      int   //统计-单次中奖>5次
	IntoCaijinCount     int   //统计-触发彩金次数
	Times110            int   //统计-中奖倍数1～5
	Times1150           int   //统计-中奖倍数6～20
	Times51200          int   //统计-中奖倍数21～50
	Times201500         int   //统计-中奖倍数51～100
	Times5011000        int   //统计-中奖倍数101～500
	Times10012000       int   //统计-中奖倍数101～500
	Times20015000       int   //统计-中奖倍数101～500
	TimesOver5000       int   //统计-中奖倍数>=500
	Times               int64 //玩家这把中奖倍数
	WinNum              int   // 赢局数
	LoseNum             int   // 输局数
	PeaceNum            int   // 和局数
	Invest              int64 // 当局投入
	TotalWinSerial      int   //总共消除次数
	//TotalWinNoKey  int   //总共消除次数，不加钻头
	CaijinBase  int64 // 用户计算彩金的基数，每次彩金游戏之后要清零
	WinCountNew int   //除开钻头中奖次数

}

func NewUser(table table.TableInterface) *User {
	user := &User{
		Table: table, Status: global.USER_STATUS_WAIT,
	}
	return user
}

//func (user *User) GetUserS2CInfo() *msg.S2CUserInfo {
//	return &msg.S2CUserInfo{
//		Name: user.User.GetNike(), Uid: user.User.GetID(), Head: user.User.GetHead(),
//		Amount: user.User.GetScore(),  Status: user.Status,
//	}
//}

func (user *User) RandInt(min, max int) int {
	return rand.Intn(max-min) + min
}

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func (user *User) RateToExec(rate int) bool {
	r := user.RandInt(1, 100)
	//log.Traceln("随机数r : ",r)
	if r < rate {
		return true
	}
	return false
}

//从max中随机去一个数，看是否小于rate
func (user *User) RateToExecWithIn(rate, max int) bool {
	r := user.RandInt(1, max)
	//log.Traceln("随机数r : ", r)
	if r < rate {
		return true
	}
	return false
}

func (user *User) ResetUser() {
	user.Status = global.USER_STATUS_END

}

func (user *User) compareInt(i1, i2 int) int {
	if i1 > i2 {
		return global.COMPARE_WIN
	}
	if i1 < i2 {
		return global.COMPARE_LOSE
	}
	return global.COMPARE_EQ
}
