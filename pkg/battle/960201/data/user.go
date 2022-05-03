package data

import (
	"fmt"

	"sync"
	"time"

	rand2 "github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/msg"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Cards             []byte //用户手里的3张牌
	CardEncode        int    //经过牌型编码之后的结果
	Id                int64  // 框架那边proto定义的int32
	Name              string
	Lock              sync.Locker //会涉及到后面用户并发修改数据，加锁
	ActionTime        int64       //用户发言时间，毫秒级 /  1e6
	ChairId           uint        //椅子id
	IsSawCards        bool        //是否看过牌，true：看了牌
	CurStatus         int         //玩家当前状态，0：正在游戏，1：比牌输了，2：弃牌
	IsActioned        bool        //是否发过言
	Head              string      //头像
	CardType          int         //牌型
	CurAmount         int64       //玩家当前已投注的金额
	ComparedUsers     []*User     //比过牌的玩家
	IsAi              bool        //是否为机器人
	IsFollowAllTheWay bool        //是否跟到底
	CurRaiseAmount    int64       //玩家加注的金额大小（给下一个玩家看 +2 +5 +10）
	TableId           int
	IsTimeOut         bool //玩家如果超时弃牌则下一把倒计时开始时踢出房间
	FollowAmount      int64
	AllInAmount       int64
	InTableCount      int //在该牌桌的场次数， >5 之后要系统对他进行换桌
	IsAllIn           bool
	AiCharacter       *config.AiConfig //机器人性格
	CompareWinCount   int              //比牌胜利次数
	CheatRate         int              //作弊率
	LoseReason        int32            // 输的原因 1：自己弃牌，2：系统比牌
	Table             table.TableInterface
	User              player.PlayerInterface
	IsLastActionUser  bool
	Amount            int64 //12余额13号，更改上下分，先记录，比赛结束再上下分
	Score             int64 //用户当前score，先缓存
	IsLeave           bool
	CardIndexInTable  int //玩家的牌在当前拍桌中是第几大 1：最大 2：第二大
	CheatRateMax      int //最大牌概率
	CheatRateSecond   int //二牌概率
}

func NewUser(uid int64, name string, isAi bool) *User {
	return &User{
		Id: uid, Name: name, ComparedUsers: make([]*User, 0), IsAi: isAi,
		IsTimeOut: false, CurStatus: global.USER_CUR_STATUS_WAIT_START,
	}
}

func (user *User) GetUserMsgInfo(isLastRoundFirstUid bool) *msg.S2CUserInfo {
	return &msg.S2CUserInfo{
		UserName: user.User.GetNike(), UserId: user.User.GetID(), Amount: user.Score, CurActionAmount: user.CurAmount, ChairId: int32(user.ChairId),
		CurRaiseAmount: user.CurRaiseAmount, CurStatus: int32(user.CurStatus), FollowAmount: user.FollowAmount, AllInAmount: user.AllInAmount,
		Ticker: global.USER_ACTION_TIME, Head: user.User.GetHead(), IsLastRoundFirstUid: user.IsLastActionUser,
		Sex: user.User.GetSex(), IsSawCard: user.IsSawCards, IsFollowAllTheWay: user.IsFollowAllTheWay, Ip: user.User.GetCity(),
	}
}

//初始化机器人的性格
func (ai *User) InitAiCharacter(gameConfig config.GameConfig) {
	time.AfterFunc(10*time.Millisecond, func() {
		if rand2.RateToExec(gameConfig.AiJiJin) {
			//log.Traceln("激进行机器人")
			ai.AiCharacter = config.AiConfigArr[1]
			return
		}
		if rand2.RateToExecWithIn(gameConfig.AiTouJi, gameConfig.AiTouJi+gameConfig.AiZhengChang+gameConfig.AiWenZhong) {
			//log.Traceln("投机行机器人")
			ai.AiCharacter = config.AiConfigArr[3]
			return
		}
		if rand2.RateToExecWithIn(gameConfig.AiWenZhong, gameConfig.AiWenZhong+gameConfig.AiZhengChang) {
			//log.Traceln("稳重行机器人")
			ai.AiCharacter = config.AiConfigArr[2]
			return
		} else {
			//log.Traceln("正常行机器人")
			ai.AiCharacter = config.AiConfigArr[0]
			return
		}
	})
}

// func (user *User) RandInt(min, max int) int {
// 	//if min >= max || min == 0 || max == 0 {
// 	//	return max.
// 	//}
// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
// 	return r.Intn(max-min) + min
// }

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func (user *User) RateToExec(rate int) bool {
	r := rand2.RandInt(1, 100)
	if r < rate {
		return true
	}
	return false
}

//从max中随机去一个数，看是否小于rate
func (user *User) RateToExecWithIn(rate, max int) bool {
	r := rand2.RandInt(1, max)
	if r < rate {
		return true
	}
	return false
}

func (user *User) ResetUser() {
	user.CurStatus = global.USER_CUR_STATUS_FINISH_GAME
	user.CurAmount = 0
	user.Cards = make([]byte, 0)
	user.IsActioned = false
	//user.IsBanker = false
	user.IsSawCards = false
	user.CurRaiseAmount = 0
	user.IsFollowAllTheWay = false
	user.ComparedUsers = make([]*User, 0)
	user.InTableCount++
	user.IsAllIn = false
	user.TableId = 0
	user.IsTimeOut = false
	user.CompareWinCount = 0
	user.ChairId = 0
}

//根据座位号设置作弊率
//如果玩家当前没有点控值，则使用血池值
func (user *User) SetCheatByChair(prob int32) {
	if user.User.GetProb() == 0 {
		user.CheatRate = int(prob)
	} else {
		user.CheatRate = int(user.User.GetProb())
	}
	if user.User.IsRobot() {
		user.CheatRate = int(prob)
		//user.CheatRate = -int(prob)
		////6月29日，新增，如果是3000
		//if user.CheatRate == 3000 {
		//	user.CheatRate = 2000
		//}
	}
	user.Table.WriteLogs(user.User.GetID(), "用户id："+fmt.Sprintf(`%d`, user.User.GetID())+
		"用户作弊率："+fmt.Sprintf(`%d`, user.CheatRate)+"系统作弊率："+fmt.Sprintf(`%d`, prob))
	if user.CheatRate == 0 {
		switch user.ChairId {
		case 1:
			user.CheatRate = 1000
		case 2:
			user.CheatRate = 1000
		case 3:
			user.CheatRate = 1000
		case 4:
			user.CheatRate = 1000
		case 5:
			user.CheatRate = 1000
		}
	}

	//log.Traceln("user : ",user.Id,"作弊率： ",user.CheatRate)
}

//是否含有某张牌
func (user *User) HasCard(card byte, cards []byte) (hasKing bool) {
	for _, v := range cards {
		cv, _ := GetCardValueAndColor(v)
		if cv == card {
			hasKing = true
			break
		}
	}
	return
}

func GetCardValueAndColor(value byte) (cardValue, cardColor byte) {
	cardValue = value & 240 //byte的高4位总和是240
	cardColor = value & 15  //byte的低4位总和是15
	return
}
