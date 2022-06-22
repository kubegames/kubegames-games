package model

import (
	"math/rand"
	"sort"
	"sync"
	"time"

	bridanimal "github.com/kubegames/kubegames-games/pkg/slots/990601/msg"
)

var (
	// 跑马灯区域
	// 猴子：1,15
	// 熊猫：2,6,9
	// 燕子：3,10,19
	// 兔子：4,8,13
	// 金鲨：5
	// 鸽子：7,
	Marquee = [...]int{
		1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12,
		13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
	}

	// 待下注区域
	// 分别为：燕子，鸽子，金鲨，银鲨，兔子，猴子，孔雀，老鹰，飞禽，走兽，熊猫，狮子
	WaitBet = []*Element{
		// Element{ID:1,Odd}
	}

	// 内层圈押注区与跑马灯的映射关系
	WaitBetMap = map[int][]int{
		1:  []int{},
		2:  []int{},
		3:  []int{},
		4:  []int{},
		5:  []int{},
		6:  []int{},
		7:  []int{},
		8:  []int{},
		9:  []int{},
		10: []int{},
		11: []int{},
		12: []int{},
	}

	// 待随机的元素，包括待下注区+通杀+通赔
	RandElement = func() []*Element {

		var result = make([]*Element, 0)
		result = append(result, WaitBet[:]...)

		result = append(result, &Element{
			ID: 13,
			OddsList: []*Odds{
				&Odds{
					OddsMax: 1,
					OddsMin: 1,
					Weight:  1},
			},

			// ProbMax: ,
		})

		return result

	}()
)

type EType int

const (
	ETypeBird        EType = iota + 1 // 飞禽
	ETypeAnimal                       // 走兽
	ETypeGoldShark                    // 金鲨
	ETypeSilverShark                  // 银鲨
	ETypeAllKill                      // 通杀
	ETypeAllPay                       // 通赔
)

type Element struct {
	ID       int   `json:"id"`       // id
	OddsList Oddss `json:"oddsList"` // 赔率权重列表
	EType    EType `json:"etype"`    // 元素类型

	OddsNow int // 当前赔率
	ProbMax int `json:"probMax"` // 最大被摇中概率
	ProbMin int `json:"probMin"` // 最小被摇中概率
	ProbNow int // 当前被摇中概率

	BaseID int32 `json:"baseID"` // 基本id

	SubIds []int `json:"subIds"` // 周围跑马灯隶属id
}

// 赔率
type Odds struct {
	OddsMax int `json:"oddsMax"` // 最大赔率
	OddsMin int `json:"oddsMin"` // 最小赔率
	Weight  int `json:"weight"`  // 权重
}

// 随机赔率值
func (o Odds) randOddsValue() int {
	if o.OddsMax == o.OddsMin {
		return o.OddsMax
	}
	max, min := o.splitOddsValue()
	return rand.Intn(max-min) + min
}

func (o Odds) splitOddsValue() (max int, min int) {
	if o.OddsMax >= o.OddsMin {
		return o.OddsMax, o.OddsMin
	}
	return o.OddsMin, o.OddsMax
}

type Oddss []*Odds

// 从大到小排序
func (o Oddss) Less(i, j int) bool {
	if o[i].Weight > o[j].Weight {
		return true
	}
	return false
}

func (o Oddss) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
}

func (o Oddss) Len() int {
	return len(o)
}

func (o Oddss) randOdds() *Odds {
	var allWeight int
	for _, v := range o {
		allWeight += v.Weight
	}
	if len(o) == 0 {
		return nil
	}
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range o {
		if randWeight <= v.Weight {
			return v
		}
		randWeight = randWeight - v.Weight
	}
	return o[0]
}

// 随机当前赔率
func (e *Element) randOddsNow() {
	if !sort.IsSorted(e.OddsList) {
		sort.Sort(e.OddsList)
	}
	if e.OddsNow == 0 {
		// 随机当前赔率
		if randOdds := e.OddsList.randOdds(); randOdds != nil {
			e.OddsNow = randOdds.randOddsValue()
		}
	}
}

// 随机当前被摇中的概率
func (e *Element) randProbNow() {

	if e.ProbMax == e.ProbMin {
		e.ProbNow = e.ProbMax
		return
	}
	if e.ProbNow == 0 {
		e.ProbNow = rand.Intn(e.ProbMax-e.ProbMin) + e.ProbMin
	}
}

func (e Element) RandSubId() int {
	if e.SubIds != nil && len(e.SubIds) != 0 {
		return e.SubIds[rand.Intn(len(e.SubIds))]
	}
	return 0
}

func (e *Element) reset() {
	e.ProbNow = 0
	e.OddsNow = 0
}

type Elements []*Element

// func (e Elements) Rand() int {
// 	e.RandOddsAndProb()
// 	if !sort.IsSorted(e) {
// 		sort.Sort(e)
// 	}
// 	// 随机选中结果
// 	id := e.randResult()
// 	// 根据结果随机出跑马灯的选项
// 	// 外层跑马灯下标
// 	result := WaitBetMap[id][index]
// 	return result
// }

func (e Elements) RandOddsAndProb() {
	wg := &sync.WaitGroup{}

	wg.Add(len(e))
	for _, v := range e {
		go func(swg *sync.WaitGroup, e *Element) {
			defer swg.Done()
			e.randOddsNow()
			e.randProbNow()
		}(wg, v)
	}
	wg.Wait()
}

func (e Elements) Reset() {
	for _, v := range e {
		v.reset()
	}
}

// 从大到小排序
func (e Elements) Less(i, j int) bool {
	if e[i].ProbNow < e[j].ProbNow {
		return false
	}
	return true
}

func (e Elements) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func (e Elements) Len() int {
	return len(e)
}

// 根据摇中概率随机摇中结果
// param
// rmShark 是否去除摇中鲨鱼
// response
// 1:element元素
// 2:int 外层跑马灯的id（转圈圈转的结果）
func (e Elements) RandResult(id int, rmShark bool) (*Element, int) {
	if !sort.IsSorted(e) {
		sort.Sort(e)
	}

	var allWeight int
	for _, v := range e {
		allWeight += v.ProbNow
		// 返回固定id的元素
		if id >= 0 && id < 14 && v.ID == id {
			if rmShark && (v.EType == ETypeGoldShark || v.EType == ETypeSilverShark) {
				// 移除金鲨/银鲨

			}

			return v, v.RandSubId()
		}
	}
	rand.Seed(time.Now().UnixNano())
LABEL:
	randWeight := rand.Intn(allWeight) + 1
	for _, v := range e {
		if randWeight <= v.ProbNow {
			if rmShark && (v.EType == ETypeGoldShark || v.EType == ETypeSilverShark) {
				goto LABEL
			}
			return v, v.RandSubId()
		}
		randWeight = randWeight - v.ProbNow
	}
	return e[0], e[0].RandSubId()
}

type Heads []*bridanimal.UserSettleInfo

// 从大到小排序
func (h Heads) Less(i, j int) bool {
	if h[i].WinGold >= h[j].WinGold {
		return true
	}
	return false
}

func (h Heads) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h Heads) Len() int {
	return len(h)
}

func (e Elements) GetByID(id int) *Element {
	for _, v := range e {
		if id == v.ID {
			return v
		}
	}
	return nil
}

func (e Elements) GetByIDResult(id int) *Element {
LABEL:
	for _, v := range e {
		if id == v.ID {
			return v
		}
	}
	prepare := []int{0, 1, 4, 5, 6, 7, 10, 11}
	id = rand.Intn(len(prepare))
	goto LABEL
}

type RandomOddss []*bridanimal.RandomOdds

func (r RandomOddss) Less(i, j int) bool {
	if r[i].ID <= r[j].ID {
		return true
	}
	return false
}

func (r RandomOddss) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RandomOddss) Len() int {
	return len(r)
}
