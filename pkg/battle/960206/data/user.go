package data

import (
	"math/rand"
	"time"

	"github.com/kubegames/kubegames-games/pkg/battle/960206/global"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/msg"
	"github.com/kubegames/kubegames-games/pkg/battle/960206/poker"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	ChairId int32 //椅子id
	Status  int32 //玩家当前状态，0：正在游戏，1：比牌输了，2：弃牌
	Table   table.TableInterface
	User    player.PlayerInterface
	Cards   [13]byte

	HeadCards    []byte //头墩
	MiddleCards  []byte //中墩
	TailCards    []byte //尾墩
	HeadCardType int
	MidCardType  int
	TailCardType int
	EncodeHead   int //经过牌型编码之后的结果
	EncodeMid    int //经过牌型编码之后的结果
	EncodeTail   int //经过牌型编码之后的结果

	SpecialHead     int32 //头墩特殊牌
	SpecialMid      int32 //中墩特殊牌
	SpecialTail     int32 //尾墩特殊牌
	SpecialCardType int32 //特殊牌型
	//ComparedArr     []*BeatUser // 11月12号来改成每个用户比完都互相装上
	IsSettleCards bool // 是否确定摆牌

	HeadWin          int
	MidWin           int
	TailWin          int
	TotalWin         int //普通牌型输赢总
	HeadPlus         int
	MidPlus          int
	TailPlus         int
	TotalPlus        int   //额外输赢总
	FinalSettle      int64 // 最终结算
	HomeRunHeadPlus  int   // 全垒打头墩被打的钱
	HomeRunMidPlus   int
	HomeRunTailPlus  int
	HomeRunTotalPlus int

	SpareArr      []*msg.S2CCardsAndCardType //给用户备选的方案，最佳方案在第一个
	CacheSpareArr []*msg.S2CCardsAndCardType //缓存的用户备选方案

	IsTest bool  // 如果是test状态，就不发牌
	Cheat  int32 //作弊率
	IsAuto bool  //是否为自动摆牌

}

const (

	// 系统角色
	SysRolePlayer = "玩家"  // 玩家
	SysRoleRobot  = "机器人" // 机器人

)

// GetSysRole 获取系统角色
func (user *User) GetSysRole() (SysRole string) {
	SysRole = SysRolePlayer
	if user.User.IsRobot() {
		SysRole = SysRoleRobot
	}

	return
}

func NewUser(table table.TableInterface, chairId int32) *User {
	user := &User{
		Table: table, Status: global.USER_STATUS_WAIT, HeadCards: make([]byte, 3),
		MiddleCards: make([]byte, 5), TailCards: make([]byte, 5), ChairId: chairId,
		SpareArr: make([]*msg.S2CCardsAndCardType, 0), SpecialCardType: 0,
		CacheSpareArr: make([]*msg.S2CCardsAndCardType, 0),
	}
	return user
}

func (user *User) GetUserS2CInfo() *msg.S2CUserInfo {
	return &msg.S2CUserInfo{
		Name: user.User.GetNike(), Uid: user.User.GetID(), Head: user.User.GetHead(),
		Amount: user.User.GetScore(), ChairId: user.ChairId, Status: user.Status,
		IsSettleCards: user.IsSettleCards, SpecialType: user.SpecialCardType,
		Ip: user.User.GetCity(),
	}
}

func (user *User) GetS2CUserEndInfo(bottom int64) *msg.S2CUserEndInfo {
	s2cEndInfo := &msg.S2CUserEndInfo{
		Uid:         user.User.GetID(),
		HeadType:    int32(user.HeadCardType),
		MidType:     int32(user.MidCardType),
		TailType:    int32(user.TailCardType),
		HeadCards:   user.SortCardsSelf(user.HeadCards, user.HeadCardType),
		MidCards:    user.SortCardsSelf(user.MiddleCards, user.MidCardType),
		TailCards:   user.SortCardsSelf(user.TailCards, user.TailCardType),
		SpecialType: user.SpecialCardType,
		//todo 最终的赢钱要乘以倍数
		HeadWin:    int64(user.HeadWin),
		MidWin:     int64(user.MidWin),
		TailWin:    int64(user.TailWin),
		HeadPlus:   int64(user.HeadPlus),
		MidPlus:    int64(user.MidPlus),
		TailPlus:   int64(user.TailPlus),
		TotalWin:   int64(user.TotalWin),
		TotalPlus:  int64(user.TotalPlus),
		ChairId:    user.ChairId,
		FinalScore: user.User.GetScore(),
		WinScore:   user.FinalSettle,
		//全垒打
		HomeRunHeadPlus:  int64(user.HomeRunHeadPlus),
		HomeRunMidPlus:   int64(user.HomeRunMidPlus),
		HomeRunTailPlus:  int64(user.HomeRunTailPlus),
		HomeRunTotalPlus: int64(user.HomeRunTotalPlus),
	}
	if !user.User.IsRobot() {
		//fmt.Println(s2cEndInfo.Uid, " head win:", s2cEndInfo.HeadWin, " m win:", s2cEndInfo.MidWin, " t win:", s2cEndInfo.TailWin)
		//fmt.Println(s2cEndInfo.Uid, " head plus:", s2cEndInfo.HeadPlus, " m plus:", s2cEndInfo.MidPlus, " t plus:", s2cEndInfo.TailPlus)
		//fmt.Println("总输赢：", s2cEndInfo.TotalPlus, s2cEndInfo.TotalWin)
	}
	//fmt.Println("用户：",user.User.GetID(),"牌：头墩中墩尾墩: ",fmt.Sprintf(`%x %x %x`,user.HeadCards,user.MiddleCards,user.TailCards)," special type : ",user.SpecialCardType,"cards : ",fmt.Sprintf(`%x`,user.Cards),time.Now())

	return s2cEndInfo
}

//牌card是否存在于user的牌中
func (user *User) IsCardInUserCards(card byte) bool {
	for _, v := range user.Cards {
		if card == v {
			return true
		}
	}
	return false
}

func (user *User) RandInt(min, max int) int {
	//if min >= max || min == 0 || max == 0 {
	//	return max.
	//}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(max-min) + min
}

//传入指定概率，然后返回是否执行  比如 rate：90 表示有90%的概率要执行
func (user *User) RateToExec(rate int) bool {
	r := user.RandInt(1, 100)
	//fmt.Println("随机数r : ",r)
	if r < rate {
		return true
	}
	return false
}

//从max中随机去一个数，看是否小于rate
func (user *User) RateToExecWithIn(rate, max int) bool {
	r := user.RandInt(1, max)
	//fmt.Println("随机数r : ", r)
	if r < rate {
		return true
	}
	return false
}

func (user *User) ResetUser() {
	user.Status = global.USER_STATUS_END
	user.EncodeHead = 0
	user.EncodeMid = 0
	user.EncodeTail = 0
	user.HeadCards = make([]byte, 3)
	user.MiddleCards = make([]byte, 5)
	user.TailCards = make([]byte, 5)
	user.IsSettleCards = false
}

//特殊牌型比较
func (user *User) CompareSpecial(another *User) {
	//fmt.Println("走特殊牌型比较")
	if user.SpecialCardType > another.SpecialCardType {
		score := SpecialScoreMap[user.SpecialCardType]
		user.TotalPlus += score
		another.TotalPlus -= score
	} else if user.SpecialCardType < another.SpecialCardType {
		score := SpecialScoreMap[another.SpecialCardType]
		another.TotalPlus += score
		user.TotalPlus -= score
	} else {

		// add by wd in 2020.3.10 for 特殊牌同种牌型比较
		// 至尊清龙，一条龙 相同牌型为和，十二皇族只会出现在一家
		if user.SpecialCardType >= SPECIAL_CARD_SEHZ {
			return
		}

		// 倒叙牌组
		userFlashback, anotherFlashBack := poker.SortCards(user.Cards[:]), poker.SortCards(another.Cards[:])

		// 四套三条 和 六对半 的倒叙牌组去除单张牌
		if user.SpecialCardType == SPECIAL_CARD_STST || user.SpecialCardType == SPECIAL_CARD_LDB {

			// 单牌
			userSingle, anotherSingle := poker.GetDzSingleCards(userFlashback)[0], poker.GetDzSingleCards(anotherFlashBack)[0]

			// 去单牌
			for k, card := range userFlashback {
				if card == userSingle {
					userFlashback = append(userFlashback[:k], userFlashback[k+1:]...)
					break
				}
			}
			for k, card := range anotherFlashBack {
				if card == anotherSingle {
					anotherFlashBack = append(anotherFlashBack[:k], anotherFlashBack[k+1:]...)
					break
				}
			}
		}

		// 获取倒叙牌组的编码
		specialCode1, specialCode2 := GetSpecialEncode(userFlashback), GetSpecialEncode(anotherFlashBack)

		if specialCode1 > specialCode2 {
			// 赢
			score := SpecialScoreMap[user.SpecialCardType]
			user.TotalPlus += score
			another.TotalPlus -= score
		} else if specialCode1 < specialCode2 {
			// 输
			score := SpecialScoreMap[another.SpecialCardType]
			another.TotalPlus += score
			user.TotalPlus -= score
		}
	}
}

//普通牌型比较 并返回打枪，如果没有就返回nil
func (user *User) CompareNormal(another *User) *msg.S2CHitRob {
	//fmt.Println("用户：",user.ChairId,"和用户：",another.ChairId," 普通牌型比较")
	hit := 0
	//比较头墩
	//fmt.Println("user 和 ano 比前：",user.User.GetID()," 钱：",user.HeadWin," ",another.User.GetID()," 钱：",another.HeadWin)
	res, headWin := user.ProcCompareHead(another)
	if res == global.COMPARE_WIN {
		hit++
	} else if res == global.COMPARE_LOSE {
		hit--
	}
	//fmt.Println("user 和 ano 比后：",user.User.GetID()," 钱：",user.HeadWin," ",another.User.GetID()," 钱：",another.HeadWin," headWin:",headWin,res)
	//比较中墩
	res, midWin := user.ProcCompareMid(another)
	//fmt.Println(user.ChairId,"中墩：",user.MiddleCards,res)
	//fmt.Println(another.ChairId,"中墩：",another.MiddleCards,res)
	if res == global.COMPARE_WIN {
		hit++
	} else if res == global.COMPARE_LOSE {
		hit--
	}
	//比较尾墩
	res, tailWin := user.ProcCompareTail(another)
	if res == global.COMPARE_WIN {
		hit++
	} else if res == global.COMPARE_LOSE {
		hit--
	}

	var s2cHitRob *msg.S2CHitRob
	//user 打抢 another
	if hit == 3 {
		s2cHitRob = user.HitRob(another, headWin, midWin, tailWin)
	}

	//another 打枪 user
	if hit == -3 {
		s2cHitRob = another.HitRob(user, headWin, midWin, tailWin)
	}
	//fmt.Println("用户：",user.ChairId,"和用户：",another.ChairId," 普通牌型比完")
	//fmt.Println(user.ChairId," ",user.HeadWin," ",user.MidWin," ",user.TailWin)
	//fmt.Println(another.ChairId," ",another.HeadWin," ",another.MidWin," ",another.TailWin)

	return s2cHitRob

}

//user 打枪 another
func (user *User) HitRob(another *User, headWin, midWin, tailWin int) *msg.S2CHitRob {
	//fmt.Println("用户：", user.User.GetID(), "打枪：", another.User.GetID(), "打枪金额：", headWin, midWin, tailWin)
	//str := fmt.Sprintf(`用户id: %d, 打枪用户: %d, 打枪值: %s`,
	//	user.User.GetID(),
	//	another.User.GetID(),
	//	score.GetScoreStr(int64(headWin+midWin+tailWin)))
	//user.Table.WriteLogs(user.User.GetID(), str)

	user.HeadPlus += headWin
	user.MidPlus += midWin
	user.TailPlus += tailWin
	another.HeadPlus -= headWin
	another.MidPlus -= midWin
	another.TailPlus -= tailWin
	//fmt.Println(user.HeadWin,user.MidWin,user.TailWin)
	//fmt.Println(user.HeadPlus,user.MidPlus,user.TailPlus)

	return &msg.S2CHitRob{
		HitUid: user.User.GetID(), BeHitUid: another.User.GetID(), HitScore: int64(headWin + midWin + tailWin),
		HitChairId: user.ChairId, BeHitChairId: another.ChairId, HitHeadScore: int64(headWin), HitMidScore: int64(midWin), HitTailScore: int64(tailWin),
	}
}

//比较头墩 返回 输赢平 headWin 该玩家头墩赢的钱
func (user *User) ProcCompareHead(another *User) (res int, heaWin int) {
	res = user.CompareHead(another)
	if res == global.COMPARE_WIN {
		if user.HeadCardType == poker.Card3TypeBz {
			user.HeadWin += global.HEAD_BZ_SCORE
			another.HeadWin -= global.HEAD_BZ_SCORE
			heaWin += global.HEAD_BZ_SCORE
			user.SpecialHead = poker.Card3TypeBz
		} else {
			user.HeadWin++
			another.HeadWin--
			heaWin++
		}
	} else if res == global.COMPARE_LOSE {
		if another.HeadCardType == poker.Card3TypeBz {
			another.HeadWin += global.HEAD_BZ_SCORE
			user.HeadWin -= global.HEAD_BZ_SCORE
			heaWin += global.HEAD_BZ_SCORE
			another.SpecialHead = poker.Card3TypeBz
		} else {
			another.HeadWin++
			user.HeadWin--
			heaWin++
		}
	}
	return
}

//比较中墩
func (user *User) ProcCompareMid(another *User) (res, midWin int) {
	res = user.CompareMid(another)
	//fmt.Println("res ",res)
	if res == global.COMPARE_WIN {
		if user.MidCardType >= poker.CardTypeHL {
			score := SpecialMidMap[user.MidCardType]
			user.MidWin += score
			another.MidWin -= score
			midWin += score
			user.SpecialMid = int32(user.MidCardType)
		} else {
			user.MidWin++
			another.MidWin--
			midWin++
		}
	} else if res == global.COMPARE_LOSE {
		if another.MidCardType >= poker.CardTypeHL {
			score := SpecialMidMap[another.MidCardType]
			another.MidWin += score
			user.MidWin -= score
			midWin += score
			another.SpecialMid = int32(another.MidCardType)
		} else {
			another.MidWin++
			user.MidWin--
			midWin++
		}
	}
	return
}

//比较尾墩
func (user *User) ProcCompareTail(another *User) (res, tailWin int) {
	res = user.CompareTail(another)
	if res == global.COMPARE_WIN {
		if user.TailCardType >= poker.CardTypeFour {
			score := SpecialTailMap[user.TailCardType]
			user.TailWin += score
			another.TailWin -= score
			tailWin += score
			user.SpecialTail = int32(user.TailCardType)
		} else {
			user.TailWin++
			another.TailWin--
			tailWin++
		}
	} else if res == global.COMPARE_LOSE {
		if another.TailCardType >= poker.CardTypeFour {
			score := SpecialTailMap[another.TailCardType]
			another.SpecialTail = int32(another.TailCardType)
			user.TailWin -= score
			another.TailWin += score
			tailWin += score
		} else {
			user.TailWin--
			another.TailWin++
			tailWin++
		}
	}
	return
}

//和另一个玩家比牌，赢了返回true
func (user *User) CompareHead(another *User) (res int) {
	res, user.HeadCardType, another.HeadCardType = user.Compare3Cards(user.HeadCards, another.HeadCards)
	return
	//user.HeadCardType, user.HeadCards = poker.GetCardType13Water(user.HeadCards)
	//another.HeadCardType, another.HeadCards = poker.GetCardType13Water(another.HeadCards)
	//user.EncodeHead, another.EncodeHead = poker.GetEncodeCard(user.HeadCardType, user.HeadCards), poker.GetEncodeCard(another.HeadCardType, another.HeadCards)
	//if user.HeadCardType == poker.Card3TypeDz && another.HeadCardType == poker.Card3TypeDz {
	//	userCard1, _ := poker.GetCardValueAndColor(user.HeadCards[1])
	//	anotherCard1, _ := poker.GetCardValueAndColor(another.HeadCards[1])
	//	if userCard1 > anotherCard1 {
	//		return global.COMPARE_WIN
	//	} else if userCard1 == anotherCard1 {
	//		return user.compareInt(user.EncodeHead, another.EncodeHead)
	//	} else {
	//		return global.COMPARE_LOSE
	//	}
	//} else {
	//	return user.compareInt(user.EncodeHead, another.EncodeHead)
	//}
}

func (user *User) CompareMid(another *User) (res int) {
	userMidArr := poker.Cards5SliceToArr(user.MiddleCards)
	anotherMidArr := poker.Cards5SliceToArr(another.MiddleCards)
	res, user.MidCardType, another.MidCardType = user.Compare5Cards(userMidArr, anotherMidArr)
	return

}

func (user *User) CompareTail(another *User) (res int) {
	userTailArr := poker.Cards5SliceToArr(user.TailCards)
	anotherTailArr := poker.Cards5SliceToArr(another.TailCards)
	res, user.TailCardType, another.TailCardType = user.Compare5Cards(userTailArr, anotherTailArr)
	return

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

//交换两个玩家牌和得分等
func (user *User) ExchangeEachOther(another *User) {
	user.Cards, another.Cards = another.Cards, user.Cards
	user.HeadCards, another.HeadCards = another.HeadCards, user.HeadCards
	user.HeadCardType, another.HeadCardType = another.HeadCardType, user.HeadCardType
	user.MiddleCards, another.MiddleCards = another.MiddleCards, user.MiddleCards
	user.MidCardType, another.MidCardType = another.MidCardType, user.MidCardType
	user.TailCards, another.TailCards = another.TailCards, user.TailCards
	user.TailCardType, another.TailCardType = another.TailCardType, user.TailCardType
	user.EncodeHead, another.EncodeHead = another.EncodeHead, user.EncodeHead
	user.EncodeMid, another.EncodeMid = another.EncodeMid, user.EncodeMid
	user.EncodeTail, another.TailCardType = another.TailCardType, user.TailCardType
	user.SpecialHead, another.SpecialHead = another.SpecialHead, user.SpecialHead
	user.SpecialMid, another.SpecialMid = another.SpecialMid, user.SpecialMid
	user.SpecialTail, another.SpecialTail = another.SpecialTail, user.SpecialTail
	user.SpecialCardType, another.SpecialCardType = another.SpecialCardType, user.SpecialCardType
}

func (user *User) GetIsAutoStr() string {
	if user.IsAuto {
		return "用户自动摆牌"
	}
	return "用户手动摆牌"
}
