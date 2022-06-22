package model

import (
	"math/rand"

	proto "github.com/kubegames/kubegames-games/pkg/slots/970501/msg"
)

type GoodluckType int

const (
	GoodluckTypeNil GoodluckType = iota
	GoodluckTypeThreeBig
	GoodluckTypeThreeSmall
	GoodluckTypeSlamBig
	GoodluckTypeSlamSmall
	GoodluckTypeFoisonBig
	GoodluckTypeTrain
)

type Goodluck struct {
	RoomProb   int64 `json:"roomProb"`   // 作弊率
	Open       int64 `json:"open"`       // 开goodluck概率
	ThreeBig   int64 `json:"threeBig"`   // 大三元(下注大西瓜、大双星、大铃铛都中奖)
	ThreeSmall int64 `json:"threeSmall"` // 小三元(下注大柠檬、大橘子、大苹果都中奖)
	SlamBig    int64 `json:"slamBig"`    // 大满贯(所有下注的不带x2的图标（不包含BAR）都中奖)
	SlamSmall  int64 `json:"slamSmall"`  // 小满贯(所有下注的x2的图标都中奖)
	FoisonBig  int64 `json:"foisonBig"`  // 大丰收(所有下注的水果图案都中奖，即大西瓜、大柠檬、大苹果、大橙子)
	Train      int64 `json:"train"`      // 开火车（随机指定起点，从起点位置（格子）顺时针走四个格子，如这四个格子中有任意下注，即可获奖。（具体表现为：四个格子依次被光标圈中，后这个五个相邻格子闪光，提示中奖。））
	Backs      Backs `json:"backs"`      // 返奖率控制
}

type back struct {
	Min  int `json:"min"`
	Max  int `json:"max"`
	Prob int `json:"prob"`
}

type Backs []back

func (bs Backs) Rand() back {
	var allweight int
	for _, v := range bs {
		allweight += v.Prob
	}
	randweight := rand.Intn(allweight) + 1
	for _, v := range bs {
		if v.Prob == 0 {
			continue
		}
		if randweight <= v.Prob {
			return v
		}
		randweight -= v.Prob
	}
	return bs[rand.Intn(len(bs))]
}

type Goodlucks []*Goodluck

var GoodlucksAll Goodlucks

func (gs Goodlucks) Find(roomProb int64) *Goodluck {
	for _, v := range gs {
		if v.RoomProb == roomProb {
			return v
		}
	}

	// 随机返回一个
	return gs[rand.Intn(len(gs))]
}

func (g Goodluck) Rand(testIn *proto.TestIn, isOpenl3000Ctrl bool, roomProb int64) GoodluckType {

	prob := int64(RandProb())
	if testIn != nil {
		switch testIn.OutID {
		case int32(GoodluckTypeTrainID):
			return GoodluckTypeTrain
		case int32(GoodluckTypeThreeBigID):
			return GoodluckTypeThreeBig
		case int32(GoodluckTypeThreeSmallID):
			return GoodluckTypeThreeSmall
		case int32(GoodluckTypeSlamBigID):
			return GoodluckTypeSlamBig
		case int32(GoodluckTypeSlamSmallID):
			return GoodluckTypeSlamSmall
		case int32(GoodluckTypeFoisonBigID):
			return GoodluckTypeFoisonBig
		case 9, 21: // 作弊开出goodluck
			prob = 0
		default:
			return GoodluckTypeNil
		}
	}

	// 3000作弊率下必输控制
	if isOpenl3000Ctrl && roomProb == 3000 {
		// 不能开出goodluck
		return GoodluckTypeNil
	}
	if prob > g.Open {
		// if prob > 5000 {
		// 没开goodluck
		return GoodluckTypeNil
	}
	// 开某种goodluck
	prob = int64(RandProb())
	// 开大三元
	if prob <= g.ThreeBig {
		return GoodluckTypeThreeBig
	}

	prob -= g.ThreeBig
	// 开小三元
	if prob <= g.ThreeSmall {
		return GoodluckTypeThreeSmall
	}

	prob -= g.ThreeSmall
	// 开大满元
	if prob <= g.SlamBig {
		return GoodluckTypeSlamBig
	}
	prob -= g.SlamBig
	// 开小满贯
	if prob <= g.SlamSmall {
		return GoodluckTypeSlamSmall
	}

	prob -= g.SlamSmall
	// 开大丰收
	if prob <= g.FoisonBig {
		return GoodluckTypeFoisonBig
	}
	// 开火车
	return GoodluckTypeTrain
}

// 根据开奖类型算出结果
func (t GoodluckType) Handle(testIn *proto.TestIn) (result []*Element, start int) {
	l := t

	big := true
	small := false
	switch l {
	case GoodluckTypeThreeBig: // 大三元
		result = append(result, &Element{Id: GoodluckTypeThreeBigID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		result = append(result, ElementsAll.GetById(ElementTypeWatermelon, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeStar2, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeBell, nil).Copy(&big))

	case GoodluckTypeThreeSmall: // 小三元
		result = append(result, &Element{Id: GoodluckTypeThreeSmallID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		result = append(result, ElementsAll.GetById(ElementTypeLemon, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeOrange, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeApple, nil).Copy(&big))

	case GoodluckTypeSlamSmall: // 小满贯
		result = append(result, &Element{Id: GoodluckTypeSlamSmallID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		result = append(result, ElementsAll.GetById(ElementTypeLemon, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeOrange, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeApple, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeBell, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeWatermelon, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeStar2, nil).Copy(&small))
		result = append(result, ElementsAll.GetById(ElementTypeSeven2, nil).Copy(&small))
		// result = append(result, ElementsAll.GetById(ElementTypeBar, &small))

	case GoodluckTypeFoisonBig: // 大丰收
		result = append(result, &Element{Id: GoodluckTypeFoisonBigID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		result = append(result, ElementsAll.GetById(ElementTypeLemon, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeOrange, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeApple, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeWatermelon, nil).Copy(&big))

	case GoodluckTypeSlamBig: // 大满贯
		result = append(result, &Element{Id: GoodluckTypeSlamBigID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		result = append(result, ElementsAll.GetById(ElementTypeLemon, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeOrange, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeApple, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeWatermelon, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeBell, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeStar2, nil).Copy(&big))
		result = append(result, ElementsAll.GetById(ElementTypeSeven2, nil).Copy(&big))

	case GoodluckTypeTrain: // 开火车
		result = append(result, &Element{Id: GoodluckTypeTrainID, OddsMax: ElementOdds{SubIds: []int{9, 21}}, OddsMin: ElementOdds{SubIds: []int{9, 21}}})
		var train []int
		start, train = GetTrain()
		result = append(result, HandleTrain(train)...)
	default:
		// type是nil，进行随机
		result = nil
	}
	return
}

func RandProb() int {
	return rand.Intn(10000) + 1
}

// 获取开火车的尾号
func GetTrain() (int, []int) {
	start := rand.Intn(24) + 1
	train := []int{start}
LOOP:
	switch start {
	case 9 + 1, 21 + 1:
		start = rand.Intn(24) + 1
		train = []int{start}
		goto LOOP
	}
	for i := 1; i <= 5; i++ {
		if start+i != 9+1 && start+i != 21+1 {
			// 外圈id超过最大值
			val := start + i
			if val >= 25 {
				val = val - 24
			}
			train = append(train, val)
		}
		if len(train) >= 5 {
			break
		}
	}
	return start, train
}

// 处理开火车
func HandleTrain(train []int) (result Elements) {
	max := true
	small := false
	for _, v := range train {
		for _, ele := range ElementsAll {
			for _, subId := range ele.OddsMax.SubIds {
				if subId == (v - 1) {
					result = append(result, ele.Copy(&max))
				}
			}
			for _, subId := range ele.OddsMin.SubIds {
				if subId == (v - 1) {
					result = append(result, ele.Copy(&small))
				}
			}
		}
	}
	return
}
