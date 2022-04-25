package model

import (
	"math/rand"
)

type ElementType int

const (
	ElementTypeNil ElementType = 0x00
	BigThreeElem   ElementType = 0x01 // 大三元（同一种车型的不同颜色）
	BigFourElem    ElementType = 0x02 // 大四喜（同一种颜色的不同车型）

	// 0x01:黄金；0x02:白银；0x03:黄铜
	ElementTypeColorRed   ElementType = 0x01
	ElementTypeColorGreed ElementType = 0x02
	ElementTypeColorBlack ElementType = 0x03
	//  0x10：奔驰；0x20：宝马；0x30：雷克萨斯；0x40：大众
	ElementTypeCarBenz  ElementType = 0x10
	ElementTypeCarBMW   ElementType = 0x20
	ElementTypeCarLexus ElementType = 0x30
	ElementTypeCarVW    ElementType = 0x40

	// 奔驰
	BenzRed   ElementType = 0x11 //
	BenzGreen ElementType = 0x12 //
	BenzBlack ElementType = 0x14 //
	// 宝马
	BMWRed   ElementType = 0x21
	BMWGreen ElementType = 0x22
	BMWBlack ElementType = 0x24
	// 雷克
	LexusRed   ElementType = 0x41
	LexusGreen ElementType = 0x42
	LexusBlack ElementType = 0x44
	// 大众
	VWRed   ElementType = 0x81
	VWGreen ElementType = 0x82
	VWBlack ElementType = 0x84
)

type ElemBase struct {
	ElemType  ElementType `json:"elemType"`  // 基本元素
	ShakeProb int         `json:"shakeProb"` // 摇中概率
	Odds      int         `json:"odds"`      // 赔率
	BetIndex  int         `json:"betIndex"`  // 下注索引
	SubIds    []int       `json:"subIds"`    // 外圈子id
}

var ElemShakeProbSlice ElemBases

// // 按下注区排序
// var ElemShakeProbSlice = ElemBases{
// 	ElemBase{ElemType: VWBlack, ShakeProb: 1800, Odds: 4,BetIndex:11, SubIds: []int{6, 17}},     // 大众黑
// 	ElemBase{ElemType: VWGreen, ShakeProb: 1550, Odds: 5,BetIndex:10, SubIds: []int{5, 18}},     // 大众绿
// 	ElemBase{ElemType: VWRed, ShakeProb: 1200, Odds: 6,BetIndex:9, SubIds: []int{4, 19}},       // 大众红
// 	ElemBase{ElemType: LexusBlack, ShakeProb: 1000, Odds: 7,BetIndex:8, SubIds: []int{9, 20}},  // 雷克萨斯黑
// 	ElemBase{ElemType: LexusGreen, ShakeProb: 1000, Odds: 10,BetIndex:7, SubIds: []int{8, 21}}, // 雷克萨斯绿
// 	ElemBase{ElemType: LexusRed, ShakeProb: 800, Odds: 12,BetIndex:6, SubIds: []int{7, 22}},    // 雷克萨斯红
// 	ElemBase{ElemType: BMWBlack, ShakeProb: 750, Odds: 13,BetIndex:5, SubIds: []int{3, 14}},    // 宝马黑
// 	ElemBase{ElemType: BMWGreen, ShakeProb: 650, Odds: 15,BetIndex:4, SubIds: []int{2, 15}},    // 宝马绿
// 	ElemBase{ElemType: BMWRed, ShakeProb: 450, Odds: 22 ,BetIndex:3,SubIds: []int{1, 16}},      // 宝马红
// 	ElemBase{ElemType: BenzBlack, ShakeProb: 350, Odds: 26 ,BetIndex:2,SubIds: []int{12, 23}},  // 奔驰黑
// 	ElemBase{ElemType: BenzGreen, ShakeProb: 250, Odds: 35,BetIndex:1, SubIds: []int{11, 24}},  // 奔驰绿
// 	ElemBase{ElemType: BenzRed, ShakeProb: 200, Odds: 45,BetIndex:0, SubIds: []int{10, 25}},    // 奔驰红
// 	// // TODO: 大三元和四喜
// 	// ElemBase{ElemType: BigThreeElem, ShakeProb: 1000, SubIds: []int{15},
// 	// ElemBase{ElemType: BigFourElem, ShakeProb: 1000, SubIds: []int{25},
// }

// 大三元
var ElemThree = ElemBase{
	ElemType: BigThreeElem,
	BetIndex: -1, // 表示没有下注区
	SubIds:   []int{0},
}

// 大四喜
var ElemFour = ElemBase{
	ElemType: BigFourElem,
	BetIndex: -1, // 表示没有下注区
	SubIds:   []int{13},
}

func Rand(base int) int {
	return rand.Intn(base) + 1
}

var r []int32

func GetOdds() []int32 {
	if r != nil {
		return r
	}
	var result = make([]int, len(ElemShakeProbSlice))

	for index := len(ElemShakeProbSlice) - 1; index >= 0; index-- {
		result[len(ElemShakeProbSlice)-1-index] = int(ElemShakeProbSlice[index].Odds)
	}
	// sort.Ints(result)
	// sort.Reverse(sort.IntSlice(result))
	for _, v := range result {
		r = append(r, int32(v))
	}
	return r
}

func ReverseOdds(r []int32) []int32 {

	result := make([]int32, len(r))
	for index := 0; index < len(r); index++ {
		result[len(r)-1-index] = r[index]
	}
	return result
}
