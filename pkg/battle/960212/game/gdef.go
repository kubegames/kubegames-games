package game

import (
	"game_poker/doudizhu/msg"
)

type FlowerType int

const (
	//方块
	DIAMOND FlowerType = iota + 1
	//梅花
	PLUM
	//红桃
	HEART
	//黑桃
	SPADE
)

type CardPoint byte

const (
	C3 CardPoint = iota + 1
	C4
	C5
	C6
	C7
	C8
	C9
	C10
	CJ
	CQ
	CK
	CA
	C2
	CBlackJoker = 0xe1
	CRedJoker   = 0xf1
)

func Sort111(c []byte, cardsType msg.CardsType) []byte {
	var res []byte
	switch cardsType {
	case msg.CardsType_TripletWithSingle:
		if c[0]>>4 != c[1]>>4 {
			res = make([]byte, len(c))
			copy(res, c)
			res[0], res[3] = res[3], res[0]
		}
	case msg.CardsType_TripletWithPair:
		if c[0]>>4 != c[2]>>4 {
			res = make([]byte, 0)
			res = append(res, c[2:5]...)
			res = append(res, c[0:2]...)
		}
	case msg.CardsType_SerialTripletWithOne:
		fallthrough
	case msg.CardsType_SerialTripletWithWing:
		threeIdx := 0
		for i, _ := range c {
			if c[i]>>4 == c[i+1]>>4 && c[i+1]>>4 == c[i+2]>>4 {
				threeIdx = i
				break
			}
		}
		threeEndIdx := 0
		for i := threeIdx; i < len(c)-2; {
			if c[i]>>4 == c[i+1]>>4 && c[i+1]>>4 == c[i+2]>>4 {
				i += 3
				threeEndIdx = i
				continue
			} else {
				break
			}
		}

		res = make([]byte, 0)
		res = append(res, c[threeIdx:threeEndIdx]...)
		if threeIdx == 0 {
			res = append(res, c[threeEndIdx:]...)
		} else if threeIdx > 0 && threeEndIdx < len(c) {
			res = append(res, c[0:threeIdx]...)
			res = append(res, c[threeEndIdx:]...)
		} else if threeIdx > 0 && threeEndIdx == len(c) {
			res = append(res, c[0:threeIdx]...)
		}
	case msg.CardsType_QuartetWithTwo:
		fourIdx := 0
		for i, _ := range c {
			if c[i]>>4 == c[i+1]>>4 && c[i+1]>>4 == c[i+2]>>4 && c[i+2]>>4 == c[i+3]>>4 {
				fourIdx = i
				break
			}
		}
		res = make([]byte, 0)
		res = append(res, c[fourIdx:fourIdx+4]...)
		if fourIdx == 0 {
			res = append(res, c[fourIdx+4:]...)
		} else if fourIdx > 0 && fourIdx < len(c) {
			res = append(res, c[:fourIdx]...)
			res = append(res, c[fourIdx+4:]...)
		} else {
			res = append(res, c[:fourIdx]...)
		}

	default:

		return c
	}

	return res
}

type CardsType int

//取较小的整数
func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

//取较大的整数
func Max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

//取较小的整数
func MinInt64(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

//取较大的整数
func MaxInt64(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}
