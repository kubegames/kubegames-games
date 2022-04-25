package game

import (
	"fmt"
	"go-game-sdk/lib/clock"
	"sync"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/data"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
)

type Game struct {
	Id         int64
	Table      table.TableInterface // table interface
	userList   [4]*data.User        //所有的玩家列表
	Room       *WaterRoom
	Status     int32
	Cards      []byte
	HitRobArr  []*msg.S2CHitRob //打枪列表
	HomeRunUid int64            // 全垒打的用户id，为0则没有全垒打
	Bottom     int64            // 底注
	lock       sync.Mutex       // 锁
	Level      int32            // 游戏当前关卡
	timerJob   *clock.Job       // 定时器
	robotTimer *clock.Job       // 机器人定时器
	Charis     [4]int64         // 座位
}

// 系统常量
const (

	// 作弊率来源
	ProbSourceRoom  = "血池" //  血池
	ProbSourcePoint = "点控" // 点控
)

func NewGame(id int64, room *WaterRoom) (game *Game) {
	if room.Name == "" {
		room.Name = "菜鸟狩猎场"
	}
	game = &Game{
		Id: id, Room: room, userList: [4]*data.User{}, HitRobArr: make([]*msg.S2CHitRob, 0),
	}
	//go game.goGameTimer()
	return
}

func (game *Game) SetUserList(user *data.User) {
	for k, v := range game.userList {
		if v == nil {
			game.userList[k] = user
			break
		}
	}
}
func (game *Game) GetUserList(uid int64) *data.User {
	for _, v := range game.userList {
		if v != nil && v.User.GetID() == uid {
			return v
		}
	}
	return nil
}
func (game *Game) DelUserList(uid int64) {
	for k, v := range game.userList {
		if v != nil && v.User.GetID() == uid {
			v = nil
			game.userList[k] = nil
			break
		}
	}
}

//获取房间基本信息
func (game *Game) GetRoomInfo2C(userSelf *data.User) *msg.S2CRoomInfo {
	info := &msg.S2CRoomInfo{
		Bottom: game.Bottom, TableStatus: game.Status, UserArr: make([]*msg.S2CUserInfo, 0),
		SelfCards: &msg.S2CUserEndInfo{
			HeadCards:   userSelf.HeadCards,
			MidCards:    userSelf.MiddleCards,
			TailCards:   userSelf.TailCards,
			HeadType:    int32(userSelf.HeadCardType),
			MidType:     int32(userSelf.MidCardType),
			TailType:    int32(userSelf.TailCardType),
			SpecialType: userSelf.SpecialCardType,
		},
		SpareArr:    userSelf.SpareArr,
		SpecialType: userSelf.SpecialCardType,
		RoomId:      game.Table.GetRoomID(),
		MinLimit:    game.Table.GetEntranceRestrictions(),
	}
	if game.timerJob != nil {
		info.Ticker = int32(game.timerJob.GetTimeDifference() / 1000)
	}
	for _, v := range game.userList {
		if v != nil {
			info.UserArr = append(info.UserArr, v.GetUserS2CInfo())
		}
	}

	log.Tracef("发送场景消息：%s", fmt.Sprintf("%+v\n", info))
	return info
}

//发送给客户端的开始游戏 消息结构体
func (game *Game) NewS2CStartGame(user *data.User) (s2cStartGame *msg.S2CStartGame) {
	s2cStartGame = &msg.S2CStartGame{
		UserArr: make([]*msg.S2CUserInfo, 0), Ticker: 25, SpecialType: user.SpecialCardType,
		SpareArr: user.SpareArr,
	}
	for k, v := range s2cStartGame.SpareArr {
		s2cStartGame.SpareArr[k].HeadCards = user.SortCardsSelf(v.HeadCards, int(v.HeadType))
		s2cStartGame.SpareArr[k].MidCards = user.SortCardsSelf(v.MidCards, int(v.MidType))
		s2cStartGame.SpareArr[k].TailCards = user.SortCardsSelf(v.TailCards, int(v.TailType))
	}
	if !user.User.IsRobot() {
		fmt.Println("s2cStartGame.SpareArr len 111 : ", len(s2cStartGame.SpareArr))
		for _, v := range s2cStartGame.SpareArr {
			fmt.Println("发给用户的备选：", fmt.Sprintf(`%x %x %x`, v.HeadCards, v.MidCards, v.TailCards))
		}
	}
	for _, v := range game.userList {
		if v != nil {
			s2cStartGame.UserArr = append(s2cStartGame.UserArr, v.GetUserS2CInfo())
		}
	}
	//fmt.Println("s2cStartGame : uid :  ", user.User.GetID(), fmt.Sprintf(`%x`, s2cStartGame.SpareArr[0].HeadCards), "用户头墩：", fmt.Sprintf(`%x`, user.HeadCards), user.HeadCardType, user.MidCardType, user.TailCardType)
	//fmt.Println("s2cStartGame : uid :  ", user.User.GetID(), fmt.Sprintf(`%x`, s2cStartGame.SpareArr[0].HeadCards), "用户中墩：", fmt.Sprintf(`%x`, user.TailCards))
	//fmt.Println("s2cStartGame : uid :  ", user.User.GetID(), fmt.Sprintf(`%x`, s2cStartGame.SpareArr[0].HeadCards), "用户尾蹲：", fmt.Sprintf(`%x`, user.TailCards))
	//fmt.Println(s2cStartGame.SpareArr[0].HeadType,s2cStartGame.SpareArr[0].MidType,s2cStartGame.SpareArr[0].TailType)
	return
}

//发送给客户端的结束游戏 消息结构体
func (game *Game) NewS2CEndGame() (s2cEndGame *msg.S2CEndGame) {
	s2cEndGame = &msg.S2CEndGame{
		HitArr: game.HitRobArr, UserArr: make([]*msg.S2CUserEndInfo, 0), HomeRunUid: game.HomeRunUid,
	}
	if game.HomeRunUid != 0 {
		homeRunUser := game.GetUserList(game.HomeRunUid)
		s2cEndGame.HomeRunChairId = homeRunUser.ChairId
	}
	for _, user := range game.userList {
		if user != nil {
			s2cEndGame.UserArr = append(s2cEndGame.UserArr, user.GetS2CUserEndInfo(game.Bottom))
		}
	}

	//fmt.Println("结束游戏：打枪用户：")
	//for _, v := range s2cEndGame.HitArr {
	//	fmt.Println(fmt.Sprintf(`%+v`, v))
	//}

	return
}

//获取房间人数
func (game *Game) GetRoomUserCount() (count int) {
	for _, v := range game.userList {
		if v != nil {
			count++
		}
	}
	return
}

// GetChairID 获取座位
func (game *Game) GetChairID(userID int64) (chairID int32) {

	for k, v := range game.Charis {

		if v == 0 {
			chairID = int32(k) + 1
			game.Charis[k] = userID
			break
		}
	}

	return
}

// DelChairID 移除座位
func (game *Game) DelChairID(chairID int32) {
	if game.Charis[chairID-1] != 0 {
		game.Charis[chairID-1] = 0
	}
}

//自定义排序手牌数组
//func (game *Game) SortCardsSelf(cards []byte, cardType int32) []byte {
//	for i := 0; i < len(cards)-1; i++ {
//		for j := 0; j < (len(cards) - 1 - i); j++ {
//			if (cards)[j] > (cards)[j+1] {
//				cards[j], cards[j+1] = cards[j+1], cards[j]
//			}
//		}
//	}
//	if cardType == poker.CardTypeTK {
//		cv0 ,_ := poker.GetCardValueAndColor(cards[0])
//		cv1 ,_ := poker.GetCardValueAndColor(cards[1])
//		cv2 ,_ := poker.GetCardValueAndColor(cards[2])
//		cv3 ,_ := poker.GetCardValueAndColor(cards[3])
//		//fmt.Println(cv0,cv1,cv2,cv3)
//		cv4 ,_ := poker.GetCardValueAndColor(cards[4])
//		if cv1 == cv2 && cv1 == cv3 {
//			return cards
//		}
//		//三条 35556
//		if cv0 == cv2 {
//			cards[0], cards[3] = cards[3], cards[0]
//			return cards
//		}
//		if cv2 ==cv4 {
//			cards[1], cards[4] = cards[4], cards[1]
//			return cards
//		}
//	}
//	if cardType == poker.CardTypeTHSA2345 || cardType == poker.CardTypeSZA2345 {
//		cards[0],cards[4] = cards[4],cards[0]
//		cards[1],cards[4] = cards[4],cards[1]
//		cards[2],cards[4] = cards[4],cards[2]
//		cards[3],cards[4] = cards[4],cards[3]
//	}
//
//	return cards
//
//}
