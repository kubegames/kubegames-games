package game

import (
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/dynamic"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/config"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960201/poker"
	msg "github.com/kubegames/kubegames-sdk/app/message"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/platform"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	Table              table.TableInterface // table interface
	TableId            int
	Round              uint         //轮数，每局游戏最多20轮
	PoolAmount         int64        //该局游戏玩家总下注资金池
	CurActionUser      *data.User   //当前发言玩家
	CurStatus          int          //当前游戏状态
	MinAction          int64        //最低下注
	MaxAction          int64        //最高下注
	CurActionIsSawCard bool         //当前下注最高的人是否看过牌
	restartTicker      *time.Ticker //重开的计时器
	lock               sync.Mutex
	IsAllIn            bool       //是否已经有人all in
	Banker             *data.User //庄家
	aiLimitCount       int        //机器人限制个数
	CompareCount       int        //比牌次数
	Cards              []byte     //52张牌
	GameConfig         *config.GameConfig
	userListArr        [5]*data.User
	startGameLock      sync.Mutex //开始游戏的锁
	*dynamic.Dynamic              //
	endGameLock        sync.Mutex //
	timerJob           *table.Job //
	StartTime          time.Time  //游戏准备开始时间，如果一定时间之后没收到玩家发牌动画结束，就踢出所有玩家
	HoseLampArr        []*msg.MarqueeConfig
	//FirstActionUser *data.User //每轮第一个发言的玩家
	IsSetLastRoundUser bool                //是否设置了最后一轮发言玩家
	CardTypeCountMap   map[int]map[int]int // 牌型、数量 => 概率
	CardTypeCountTotal int                 //
	CompareIds         [2]int64            //1月16号加的，因为有比牌阶段玩家断线重连，所以在比牌阶段就禁止比牌的两个玩家弃牌
	isChangeCards      bool                //true：换过牌，就不再换了
}

func (game *Game) Init(table table.TableInterface) {
	game.Table = table
}

func (game *Game) InitCardTypeCountMap() {
	game.CardTypeCountMap = make(map[int]map[int]int)
	game.CardTypeCountMap[poker.CardTypeSingle] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeDZ] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeSZ] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeJH] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeSJ] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeBZ] = make(map[int]int)
	game.CardTypeCountMap[poker.CardTypeSingle][0] = 5005
	game.CardTypeCountMap[poker.CardTypeSingle][1] = 4550
	game.CardTypeCountMap[poker.CardTypeSingle][2] = 4136
	game.CardTypeCountMap[poker.CardTypeSingle][3] = 3760
	game.CardTypeCountMap[poker.CardTypeSingle][4] = 752
	game.CardTypeCountMap[poker.CardTypeDZ][0] = 3603
	game.CardTypeCountMap[poker.CardTypeDZ][1] = 3133
	game.CardTypeCountMap[poker.CardTypeDZ][2] = 2724
	game.CardTypeCountMap[poker.CardTypeDZ][3] = 2369
	game.CardTypeCountMap[poker.CardTypeSZ][0] = 814
	game.CardTypeCountMap[poker.CardTypeSZ][1] = 626
	game.CardTypeCountMap[poker.CardTypeSZ][2] = 482
	game.CardTypeCountMap[poker.CardTypeJH][0] = 505
	game.CardTypeCountMap[poker.CardTypeJH][1] = 316
	game.CardTypeCountMap[poker.CardTypeSJ][0] = 35
	game.CardTypeCountMap[poker.CardTypeSJ][1] = 15
	game.CardTypeCountMap[poker.CardTypeBZ][0] = 38
	game.CardTypeCountTotal = 0
	for _, v := range game.CardTypeCountMap {
		for _, vv := range v {
			game.CardTypeCountTotal += vv
		}
	}
}

func NewGame(tableId int, gameConfig *config.GameConfig) (zjh *Game) {
	zjh = &Game{
		TableId: tableId, HoseLampArr: make([]*msg.MarqueeConfig, 0),
		MinAction: gameConfig.MinAction, aiLimitCount: 2, GameConfig: gameConfig,
		StartTime: time.Now(), //userListArr: make([]*data.User, 0),
		CurStatus: global.TABLE_CUR_STATUS_WAIT,
	}
	for _, v := range poker.Deck {
		zjh.Cards = append(zjh.Cards, v)
	}
	zjh.ShuffleCards()
	zjh.Dynamic = dynamic.NewDynamic()
	zjh.AddFunc("ProcUserStartGame", zjh, "ProcUserStartGame")
	zjh.AddFunc("ProcAction", zjh, "ProcAction")
	zjh.AddFunc("ProcCompare", zjh, "ProcCompare")
	zjh.AddFunc("ProcGetCanCompareList", zjh, "ProcGetCanCompareList")
	zjh.AddFunc("ProcSendCardOver", zjh, "ProcSendCardOver")
	zjh.AddFunc("ProcLeaveGame", zjh, "ProcLeaveGame")
	return
}

func (zjh *Game) SetUserListMap(user *data.User) {
	for i, v := range zjh.userListArr {
		if v == nil {
			zjh.userListArr[i] = user
			return
		}
	}
	log.Traceln("坐满了，set user 失败")
}

func (zjh *Game) GetUserListMap(uid int64) *data.User {
	for _, user := range zjh.userListArr {
		if user != nil && user.User.GetID() == uid {
			return user
		}
	}
	return nil
}

func (zjh *Game) DelUserListMap(uid int64) {
	for i, user := range zjh.userListArr {
		if user != nil && user.User.GetID() == uid && !user.IsLeave && user.CurAmount != 0 {
			log.Traceln("用户提前离开房间 下分", user.User.GetID(), time.Now(), "发送战绩打码量：", user.CurAmount, "总下注：", user.Amount)
			user.User.SetScore(user.Table.GetGameNum(), -user.Amount, zjh.Table.GetRoomRate())
			//user.User.SetEndCards(fmt.Sprintf(`玩家在该局投入%d元，并提前离开`, user.Amount))
			//user.User.SetBetsAmount(user.Amount)
			//user.User.SendRecord(zjh.Table.GetGameNum(), -user.CurAmount, user.Amount, 0, 0, fmt.Sprintf(`玩家在该局投入%d元，并提前离开`, user.Amount))

			//战绩
			var records []*platform.PlayerRecord

			if !user.User.IsRobot() {
				records = append(records, &platform.PlayerRecord{
					PlayerID:     uint32(user.User.GetID()),
					GameNum:      zjh.Table.GetGameNum(),
					ProfitAmount: -user.CurAmount,
					BetsAmount:   user.Amount,
					DrawAmount:   0,
					OutputAmount: 0,
					Balance:      user.User.GetScore(),
					UpdatedAt:    time.Now(),
					CreatedAt:    time.Now(),
				})
				user.User.SendChip(user.CurAmount)
			}
			//user.Amount = 0
			//user.Score = 0
			//zjh.userListArr[i] = nil
			//zjh.userListArr[i].CurStatus = global.USER_CUR_STATUS_GIVE_UP
			user.IsLeave = true
			zjh.userListArr[i].IsLeave = true
			zjh.Table.KickOut(user.User)
			if zjh.CurStatus == global.TABLE_CUR_STATUS_WAIT || zjh.CurStatus == global.TABLE_CUR_STATUS_MATCHING {
				zjh.userListArr[i] = nil
			}

			//发送战绩
			if len(records) > 0 {
				if _, err := zjh.Table.UploadPlayerRecord(records); err != nil {
					log.Warnf("upload player record error %s", err.Error())
				}
			}

			return
		}
	}
}

//获取房间总人数
func (game *Game) GetTableUserCount() (count int) {
	//return len(game.userListArr)
	for _, user := range game.userListArr {
		if user != nil {
			count++
		}
	}
	return
}

func (game *Game) IsAllUserAllIn() bool {
	userOnlineList := game.GetStatusUserList(global.USER_CUR_STATUS_ING)
	for _, v := range userOnlineList {
		if !v.IsAllIn {
			return false
		}
	}
	return true
}

//获取特殊标语内容
func (game *Game) GetSpecialSlogan(cardType int) string {
	switch cardType {
	case poker.CardTypeJH:
		return "金花"
	case poker.CardTypeSJ:
		return "顺金"
	case poker.CardTypeBZ:
		return "豹子"
	}
	return ""
}
