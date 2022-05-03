package model

import (
	"math/rand"
	"sort"

	pai9 "github.com/kubegames/kubegames-games/pkg/battle/960211/msg"
)

const (
	DEFAULT_CARD_LEN = 32 // 默认的牌数量
)

var (
	// 牌的所有类型(单牌所有类型)
	CardsAllType Cards
)

type Card struct {
	pai9.Poker
	Name string `json:"Name"`
}

func (c *Card) Copy() *Card {
	return &Card{
		Poker: c.Poker,
		Name:  c.Name,
	}
}

func (c *Card) Equal(target *Card) bool {
	return c.Sorted == target.Sorted && c.Val == target.Val
}

type Cards []*Card

// 单排从大到小排序
func (c Cards) Less(i, j int) bool { return c[i].Sorted > c[j].Sorted }
func (c Cards) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c Cards) Len() int           { return len(c) }

// 初始化{{DEFAULT_CARD_LEN}}长度的一副牌
func Init(list []*Card) *Cards {
	deck := new(Cards)
	for i := 0; i < DEFAULT_CARD_LEN; i++ {
		*deck = append(*deck, list[i%len(list)])
	}
	deck.Shuffle()
	return deck
}

// 发n张牌
func (c *Cards) DealPoker(n int) Cards {
	if n > len(*c) {
		defer func() { c = nil }()
		return *c
	}
	deal := (*c)[len(*c)-n:]
	*c = (*c)[:len(*c)-n]
	return deal
}

// 洗牌
func (c *Cards) Shuffle() {
	for i := 0; i < len(*c); i++ {
		index := rand.Intn(len(*c))
		(*c)[i], (*c)[index] = (*c)[index], (*c)[i]
	}
}

// 1:c>target
// -1: c<target
// 0: c == target
// 只支持2张牌的比较
func (c Cards) Compare(target Cards) int {
	if len(c) != 2 || len(target) != 2 {
		return 0
	}
	ct, _ := c.CalcType()
	tt, _ := target.CalcType()
	if ct > tt {
		return 1
	} else if ct < tt {
		return -1
	} else {
		// 相等，计算大点
		sort.Sort(c)
		sort.Sort(target)

		// 计算第一张牌
		cmp := c[0].compare(target[0])
		if cmp == 0 {
			cmp = c[1].compare(target[1])
		}
		return cmp
	}
}

// 单牌比较
// 1:c>target
// 0:c==target
// -1:c<target
func (c *Card) compare(target *Card) int {
	diff := c.Sorted - target.Sorted
	if diff < 0 {
		diff *= -1
	}

	// 相邻
	if diff == 1 {

		// 特殊相邻相等情况
		if (c.Sorted == 1 && target.Sorted == 2) ||
			(c.Sorted == 2 && target.Sorted == 1) ||
			(c.Sorted == 3 && target.Sorted == 4) ||
			(c.Sorted == 4 && target.Sorted == 3) ||
			(c.Sorted == 5 && target.Sorted == 6) ||
			(c.Sorted == 6 && target.Sorted == 5) ||
			(c.Sorted == 7 && target.Sorted == 8) ||
			(c.Sorted == 8 && target.Sorted == 7) ||
			(c.Sorted == 9 && target.Sorted == 10) ||
			(c.Sorted == 10 && target.Sorted == 9) {
			return 0
		} else if c.Sorted > target.Sorted {
			return 1
		} else {
			return -1
		}
	}

	// 不相邻
	if c.Sorted > target.Sorted {
		return 1
	} else if c.Sorted < target.Sorted {
		return -1
	} else {
		return 0
	}
}

// 计算牌组合成的类型（目前只支持2张牌）
func (c Cards) CalcType() (pai9.PokerType, string) {
	if len(c) != 2 {
		return pai9.PokerType_Zero, "鳖十"
	}
	switch {
	case c.IsZhiZun():
		return pai9.PokerType_ZhiZun, "至尊"
	case c.IsShaungTian():
		return pai9.PokerType_ShuangTian, "双天"
	case c.IsShaungDi():
		return pai9.PokerType_ShuangDi, "双地"
	case c.IsShaungRen():
		return pai9.PokerType_ShuangRen, "双人"
	case c.IsShaungE():
		return pai9.PokerType_ShuangE, "双鹅"
	case c.IsShaungMei():
		return pai9.PokerType_ShuangMei, "双梅"
	case c.IsShaungChangSan():
		return pai9.PokerType_ShuangChangSan, "双长三"
	case c.IsShaungBanDeng():
		return pai9.PokerType_ShuangBanDeng, "双板凳"
	case c.IsShaungFuTou():
		return pai9.PokerType_ShuangFuTou, "双斧头"
	case c.IsShaungHongTou():
		return pai9.PokerType_ShuangHongTou, "双红头"
	case c.IsShaungGaoJiao():
		return pai9.PokerType_ShuangGaoJiao, "双高脚"
	case c.IsShaungLingLin():
		return pai9.PokerType_ShuangLingLin, "双零霖"
	case c.IsZaJiu():
		return pai9.PokerType_ZaJiu, "杂九"
	case c.IsZaBa():
		return pai9.PokerType_ZaBa, "杂八"
	case c.IsZaQi():
		return pai9.PokerType_ZaQi, "杂七"
	case c.IsZaWu():
		return pai9.PokerType_ZaWu, "杂五"
	case c.IsTianWang():
		return pai9.PokerType_TianWang, "天王"
	case c.IsDiWang():
		return pai9.PokerType_DiWang, "地王"
	case c.IsTianGang():
		return pai9.PokerType_TianGang, "天杠"
	case c.IsDiGang():
		return pai9.PokerType_DiGang, "地杠"
	case c.IsTianGaoJiu():
		return pai9.PokerType_TianGaoJiu, "天高九"
	case c.IsDiGaoJiu():
		return pai9.PokerType_DiGaoJiu, "地高九"
	}
	// 不是对牌组合，计算点数
	var val int32
	for _, v := range c {
		val += v.Val
	}
	return pai9.PokerType(val % 10), getPoint(int(val) % 10)
}

// 是否是至尊牌
func (c Cards) IsZhiZun() bool {
	return (c[0].Equal(CardsAllType[20]) && c[1].Equal(CardsAllType[19])) ||
		(c[1].Equal(CardsAllType[20]) && c[0].Equal(CardsAllType[19]))
}

// 是否是双天
func (c Cards) IsShaungTian() bool {
	return (c[0].Equal(CardsAllType[0]) && c[1].Equal(CardsAllType[0]))
}

// 是否是双地
func (c Cards) IsShaungDi() bool {
	return (c[0].Equal(CardsAllType[1]) && c[1].Equal(CardsAllType[1]))
}

// 是否是双人
func (c Cards) IsShaungRen() bool {
	return (c[0].Equal(CardsAllType[2]) && c[1].Equal(CardsAllType[2]))
}

// 是否是双鹅
func (c Cards) IsShaungE() bool {
	return (c[0].Equal(CardsAllType[3]) && c[1].Equal(CardsAllType[3]))
}

// 是否是双梅
func (c Cards) IsShaungMei() bool {
	return (c[0].Equal(CardsAllType[4]) && c[1].Equal(CardsAllType[4]))
}

// 是否是双长三
func (c Cards) IsShaungChangSan() bool {
	return (c[0].Equal(CardsAllType[5]) && c[1].Equal(CardsAllType[5]))
}

// 是否是双板凳
func (c Cards) IsShaungBanDeng() bool {
	return (c[0].Equal(CardsAllType[6]) && c[1].Equal(CardsAllType[6]))
}

// 是否是双斧头
func (c Cards) IsShaungFuTou() bool {
	return (c[0].Equal(CardsAllType[7]) && c[1].Equal(CardsAllType[7]))
}

// 是否是双红头
func (c Cards) IsShaungHongTou() bool {
	return (c[0].Equal(CardsAllType[8]) && c[1].Equal(CardsAllType[8]))
}

// 是否是双高脚
func (c Cards) IsShaungGaoJiao() bool {
	return (c[0].Equal(CardsAllType[9]) && c[1].Equal(CardsAllType[9]))
}

// 是否是双零霖
func (c Cards) IsShaungLingLin() bool {
	return (c[0].Equal(CardsAllType[10]) && c[1].Equal(CardsAllType[10]))
}

// 是否是杂九
func (c Cards) IsZaJiu() bool {
	return (c[0].Equal(CardsAllType[11]) && c[1].Equal(CardsAllType[12])) ||
		(c[1].Equal(CardsAllType[11]) && c[0].Equal(CardsAllType[12]))
}

// 是否是杂八
func (c Cards) IsZaBa() bool {
	return (c[0].Equal(CardsAllType[13]) && c[1].Equal(CardsAllType[14])) ||
		(c[1].Equal(CardsAllType[13]) && c[0].Equal(CardsAllType[14]))
}

// 是否是杂七
func (c Cards) IsZaQi() bool {
	return (c[0].Equal(CardsAllType[15]) && c[1].Equal(CardsAllType[16])) ||
		(c[1].Equal(CardsAllType[15]) && c[0].Equal(CardsAllType[16]))
}

// 是否是杂五
func (c Cards) IsZaWu() bool {
	return (c[0].Equal(CardsAllType[17]) && c[1].Equal(CardsAllType[18])) ||
		(c[1].Equal(CardsAllType[17]) && c[0].Equal(CardsAllType[18]))
}

// 是否是天王
func (c Cards) IsTianWang() bool {
	return (c[0].Equal(CardsAllType[0]) && c[1].Val == 9) ||
		(c[1].Equal(CardsAllType[0]) && c[0].Val == 9)
}

// 是否是地王
func (c Cards) IsDiWang() bool {
	return (c[0].Equal(CardsAllType[1]) && c[1].Val == 9) ||
		(c[1].Equal(CardsAllType[1]) && c[0].Val == 9)
}

// 是否是天杠
func (c Cards) IsTianGang() bool {
	return (c[0].Equal(CardsAllType[0]) && (c[1].Val == 8)) ||
		(c[1].Equal(CardsAllType[0]) && (c[0].Val == 8))
}

// 是否是地杠
func (c Cards) IsDiGang() bool {
	return ((c[0].Val == 8) && c[1].Equal(CardsAllType[1])) ||
		(c[0].Equal(CardsAllType[1]) && c[1].Val == 8)
}

// 是否是天高九
func (c Cards) IsTianGaoJiu() bool {
	return (c[0].Equal(CardsAllType[16]) && c[1].Equal(CardsAllType[0])) ||
		(c[1].Equal(CardsAllType[16]) && c[0].Equal(CardsAllType[0]))
}

// 是否是地高九
func (c Cards) IsDiGaoJiu() bool {
	return (c[0].Equal(CardsAllType[9]) && c[1].Equal(CardsAllType[1])) ||
		(c[1].Equal(CardsAllType[9]) && c[0].Equal(CardsAllType[1]))
}

func getPoint(val int) string {
	var prefix string
	switch val {
	case 0:
		prefix = "零"
	case 1:
		prefix = "一"
	case 2:
		prefix = "二"
	case 3:
		prefix = "三"
	case 4:
		prefix = "四"
	case 5:
		prefix = "五"
	case 6:
		prefix = "六"
	case 7:
		prefix = "七"
	case 8:
		prefix = "八"
	case 9:
		prefix = "九"
	}
	return prefix + "点"
}

// 桌子上已发出的牌
type CardsTable []Cards

// 大牌在前
func (ct CardsTable) Less(i, j int) bool {
	cmp := ct[i].Compare(ct[j])
	return cmp > 0
}

func (ct CardsTable) Swap(i, j int) { ct[i], ct[j] = ct[j], ct[i] }

func (ct CardsTable) Len() int { return len(ct) }
