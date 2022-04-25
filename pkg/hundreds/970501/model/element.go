package model

import (
	"math/rand"
)

type ElementType int

const (
	ElementTypeBar ElementType = iota
	ElementTypeSeven2
	ElementTypeStar2
	ElementTypeWatermelon
	ElementTypeBell
	ElementTypeOrange
	ElementTypeLemon
	ElementTypeApple
)

const (
	GoodluckTypeThreeBigID ElementType = iota + 24
	GoodluckTypeThreeSmallID
	GoodluckTypeSlamBigID
	GoodluckTypeSlamSmallID
	GoodluckTypeFoisonBigID
	GoodluckTypeTrainID
)

type Element struct {
	Id      ElementType `json:"id"`      // 0:bar;1:77;2:双星;3:西瓜;4:铃铛;5:橘子;6：柠檬;7:苹果
	OddsMax ElementOdds `json:"oddsMax"` // 大赔率
	OddsMin ElementOdds `json:"oddsMin"` // 小赔率
	IsMax   bool        // 开大赔还是小赔
}

func (e *Element) Copy(isMax *bool) *Element {
	max := e.IsMax
	if isMax != nil {
		max = *isMax
	}
	return &Element{
		Id:      e.Id,
		OddsMax: e.OddsMax,
		OddsMin: e.OddsMin,
		IsMax:   max,
	}
}

type ElementOdds struct {
	Odds   int32 `json:"odds"`   // 赔率
	SubIds []int `json:"subIds"` // 子id（外圈id）
	Prob   int   `json:"prob"`   // 摇中概率
}

type Elements []*Element

// 随机摇出一个结果
func (es Elements) Rand(id ElementType) *Element {
	if id >= ElementTypeApple && id <= ElementTypeBar {
		return es.GetById(id, nil)
	}
	var allWeight int
	for _, v := range es {
		allWeight += (v.OddsMax.Prob)
		allWeight += (v.OddsMin.Prob)
	}
	// bts, _ := json.Marshal(es)
	// fmt.Println("es = ", string(bts))
	prob := rand.Intn(allWeight) + 1

	for _, v := range es {
		// fmt.Println("prob = ", prob)
		if prob <= (v.OddsMin.Prob + v.OddsMax.Prob) {
			if prob <= v.OddsMax.Prob {
				v.IsMax = true
			} else {
				v.IsMax = false
			}
			// fmt.Printf("prob111 = %d  v ===%v \n\n", prob, *v)
			return v
		}
		prob -= (v.OddsMax.Prob + v.OddsMin.Prob)
	}
	panic("未找到")
	// 未找到
	return nil
}

func (es Elements) GetById(id ElementType, isMax *bool) *Element {
	for _, v := range es {
		if v.Id == id {
			if isMax != nil {
				v.IsMax = *isMax
			}
			return v
		}
	}
	return nil
}

func (es Elements) GetOdds() (odds []int32) {
	for _, v := range es {
		odds = append(odds, v.OddsMax.Odds)
	}
	return nil
}

func (e Element) RandSubId() int {
	if e.IsMax {
		return e.OddsMax.SubIds[rand.Intn(len(e.OddsMax.SubIds))]
	}
	return e.OddsMin.SubIds[rand.Intn(len(e.OddsMin.SubIds))]
}

var ElementsAll Elements
