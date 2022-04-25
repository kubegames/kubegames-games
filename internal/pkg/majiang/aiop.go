package majiang

import (
	"math"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
)

const (
	noOperation = 0
	chitype     = 1
	pengtype    = 2
	gangtype    = 3
)

//定却操作
func dingQueOP() {

}

//出牌操作 输入参数 原手牌，操作后手牌，吃碰杠暗杠，
func PlayCardOP(handCards [MaxCardValue]int, tableCards [MaxCardValue]int) int {
	/* { 打出计算每张手牌权值最小的牌}*/
	temp := handCards
	currentCount := 0
	minCard := 0
	mincardArr, index := getMinWeightCard(temp)
	//log.Debugf("最小值arrmincard:%v,,index,%v", mincardArr, index)
	for i := 0; i < index; i++ {
		if mincardArr[i] >= 1 {
			currentCount++
		}
	}
	if currentCount == 1 {
		minCard = mincardArr[0]
		//log.Debugf("最小值只有一张打出最小牌%v,%v", minCard, mincardArr[0])
		return minCard
	} else {
		if mincardArr[currentCount-1] >= 31 {
			minCard = mincardArr[currentCount-1]
			//log.Debugf("最小牌有多个有字牌打出字牌，字牌%v,%v", minCard, mincardArr[currentCount-1])
			return minCard
		} else {
			//log.Debugf("没有字牌最小牌有%v个", currentCount)
			maxNum := -1
			for i := 0; i <= currentCount-1; i++ {
				if tableCards[mincardArr[i]] > maxNum {
					maxNum = tableCards[mincardArr[i]]
				}
			}
			count := 0
			tempArr := [13]int{}
			for i := 0; i <= currentCount-1; i++ {
				if tableCards[mincardArr[i]] == maxNum {
					minCard = mincardArr[i]
					tempArr[count] = mincardArr[i]
					count++
				}
			}
			//log.Debugf("牌桌上此些牌出现的数量最多且相同的牌%v，个数%v", tempArr, count)
			if count == 1 {
				//log.Debugf("牌桌上此些牌出现的数量最多且为1个牌%v", minCard)
				return minCard
			} else {
				max := 0.0
				for i := 0; i <= count-1; i++ {
					if math.Abs(float64(tempArr[i]%10)-5) > max {
						max = math.Abs(float64(tempArr[i]%10) - 5)
					}
				}
				count1 := 0
				tempArr1 := [13]int{}
				for i := 0; i <= count-1; i++ {
					if math.Abs(float64(tempArr[i]%10)-5) == max {
						tempArr1[count1] = tempArr[i]
						minCard = tempArr[i]
						count1++
					}
				}
				//log.Debugf("牌桌靠边的牌%v，个数%v", tempArr1, count1)
				if count1 == 1 {
					//minCard = tempArr1[0]
					//log.Debugf("牌桌上此些牌出现靠边的牌为1个牌%v", minCard)
					return minCard
				} else {
					a := rand.RandInt(0, count1-1)
					//log.Debugf("%v", count1)
					minCard = tempArr1[a]
					//log.Debugf("随机出牌%v", minCard)
					return minCard
				}

			}
		}
	}
}

//碰牌操作
func PengCardOP(handCards [MaxCardValue]int, opLaterHandCards [MaxCardValue]int, opCard int) (bool, float64) {
	temp1 := handCards
	noOpWeight := 0.0
	oPMaxWeight := 0.0
	noOpWeight = GetCardsWeight(temp1, 0, noOperation)

	minCard, _ := getMinWeightCard(opLaterHandCards)
	temp := opLaterHandCards
	temp[minCard[0]] -= 1
	oPMaxWeight = GetCardsWeight(temp, opCard, pengtype)
	//log.Debugf("操作权值%v",oPMaxWeight)
	//log.Debugf("不操作权值%v",noOpWeight)

	if oPMaxWeight >= noOpWeight {
		return true, oPMaxWeight
	} else {
		return false, 0
	}
}

//吃牌操作
func EatCardOP(handCards [MaxCardValue]int, opLaterHandCards [MaxCardValue]int, opCard int) (bool, float64) {
	//计算每张牌出牌的权值A 选出权值最大的A_max出牌 if 有权值相同的牌{选择场上已经打过的牌}
	tmp := handCards
	noOpWeight := 0.0
	oPMaxWeight := 0.0
	noOpWeight = GetCardsWeight(tmp, 0, noOperation)

	minCard, _ := getMinWeightCard(opLaterHandCards)
	temp := opLaterHandCards
	temp[minCard[0]] -= 1
	oPMaxWeight = GetCardsWeight(temp, opCard, chitype)
	//log.Debugf("操作权值%v",oPMaxWeight)
	//log.Debugf("不操作权值%v",noOpWeight)

	if oPMaxWeight >= noOpWeight {
		return true, oPMaxWeight
	} else {
		return false, 0
	}
}

//杠牌操作
func MoGangCardOP(handCards [MaxCardValue]int, opLaterHandCards [MaxCardValue]int, opCard int, warCardsNum int, tingStatusNum int) bool {
	/* vat ting
	   if 听牌状态{
	   	if 杠牌之后不在是听牌状态{
	   		不杠
	   	}else {杠 }
	   }
	   计算现有手牌权值A 计算杠牌后最佳出牌的权值B*增益风险a（a=a1*a2*a3）a=0.4+s(-0.005)  A>B不杠 反之杠*/
	tem := handCards
	noOpWeight := 0.0
	oPMaxWeight := 0.0
	minCard, _ := getMinWeightCard(tem)
	temp := handCards
	temp[minCard[0]] -= 1
	noOpWeight = GetCardsWeight(temp, 0, noOperation)

	oPMaxWeight = GetCardsWeight(opLaterHandCards, opCard, gangtype)
	oPMaxWeight = oPMaxWeight * (1 + 0.4 + float64(tingStatusNum)*(-0.05) + (-0.005)*(45-float64(warCardsNum)))
	//log.Debugf("操作权值%v",oPMaxWeight)
	//log.Debugf("不操作权值%v",noOpWeight)
	if oPMaxWeight >= noOpWeight {
		return true
	} else {
		return false
	}
}

func MingGangCardOP(handCards [MaxCardValue]int, opLaterHandCards [MaxCardValue]int, opCard int, warCardsNum int, tingStatusNum int) (bool, float64) {
	/* vat ting
	   if 听牌状态{
	   	if 杠牌之后不在是听牌状态{
	   		不杠
	   	}else {杠 }
	   }
	   计算现有手牌权值A 计算杠牌后最佳出牌的权值B*增益风险a（a=a1*a2*a3）a=0.4+s(-0.005)  A>B不杠 反之杠*/
	noOpWeight := 0.0
	oPMaxWeight := 0.0
	temp := handCards
	noOpWeight = GetCardsWeight(temp, 0, noOperation)

	oPMaxWeight = GetCardsWeight(opLaterHandCards, opCard, gangtype)
	//log.Debugf("操作权值%v",oPMaxWeight)
	oPMaxWeight = oPMaxWeight * (1 + 0.4 + float64(tingStatusNum)*(-0.05) + (-0.005)*(45-float64(warCardsNum)))
	//log.Debugf("操作权值%v",oPMaxWeight)
	//log.Debugf("不操作权值%v",noOpWeight)
	if oPMaxWeight >= noOpWeight {
		return true, oPMaxWeight
	} else {
		return false, 0.0
	}
}

//听牌操作
func TingCardOP(cards []int) {

}

//胡牌操作
func HuCardOP(cards []int) {

}

/*
获取各色牌权值，及总权值
返回值万，条，同，字，总。权值
*/
func GetCardsWeight(cards [MaxCardValue]int, opCard int, opType int) float64 {
	//遍历手牌计算每张牌的权值
	totallWeight := 0.0
	wanWeight := 0
	tiaoWeight := 0
	tongWeight := 0
	ziWeight := 0
	wanNum := 0
	tiaoNum := 0
	tongNum := 0
	ziNum := 0
	for i := 1; i <= 9; i++ {
		if cards[i] >= 1 {
			wanWeight += getWeight(cards, i)
			wanNum += cards[i]
		}
	}
	for i := 11; i <= 19; i++ {
		if cards[i] >= 1 {
			tiaoWeight += getWeight(cards, i)
			tiaoNum += cards[i]
		}

	}
	for i := 21; i <= 29; i++ {
		if cards[i] >= 1 {
			tongWeight += getWeight(cards, i)
			tongNum += cards[i]
		}
	}
	for i := 31; i <= 37; i++ {
		if cards[i] >= 1 {
			ziWeight += getWeight(cards, i)
			ziNum += cards[i]
		}

	}
	if opCard > 0 {
		q, w, e, t, y, u, i, o := getCPGNumAndWeight(opCard, opType)
		wanWeight += q
		wanNum += w
		tiaoWeight += e
		tiaoNum += t
		tongWeight += y
		tongNum += u
		ziWeight += i
		ziNum += o
	}
	a := float64(wanWeight) * (1 + float64(wanNum)/14)
	b := float64(tiaoWeight) * (1 + float64(tiaoNum)/14)
	c := float64(tongWeight) * (1 + float64(tongNum)/14)
	d := float64(ziWeight) * (1 + float64(ziNum)/14)
	totallWeight = a + b + c + d
	return float64(totallWeight)
}

func getMinWeightCard(handCards [MaxCardValue]int) ([13]int, int) {
	temp := handCards
	minTemp := 200
	currentIndext := 0
	weightArr := [MaxCardValue]int{}
	cardArr := [13]int{}
	for i := Wan[0]; i <= Bai[0]; i++ {
		if temp[i] >= 1 {
			weight := getWeight(temp, i) + 1
			if minTemp > weight {
				weightArr[i] = weight
				minTemp = weight
			}
		}
	}
	//for i := 11; i <= 19; i++ {
	//	if temp[i] >= 1 {
	//		weight := getWeight(temp, i) + 1
	//		weightArr[i] = weight
	//		if minTemp > weight {
	//			minTemp = weight
	//		}
	//	}
	//
	//}
	//for i := 21; i <= 29; i++ {
	//	if temp[i] >= 1 {
	//		weight := getWeight(temp, i) + 1
	//		weightArr[i] = weight
	//		if minTemp > weight {
	//			minTemp = weight
	//		}
	//	}
	//}
	//for i := 31; i <= 37; i++ {
	//	if temp[i] >= 1 {
	//		weight := getWeight(temp, i) + 1
	//		weightArr[i] = weight
	//		if minTemp > weight {
	//			minTemp = weight
	//		}
	//	}
	//}

	for k, v := range weightArr {
		if v == minTemp {
			cardArr[currentIndext] = k
			currentIndext++
			break
		}
	}
	return cardArr, currentIndext
}

func getWeight(cards [MaxCardValue]int, i int) int {
	weight := 0
	v := cards[i]
	if i < 31 {
		tmp := i % 10
		if tmp == 1 {
			if cards[i+1] >= 1 {
				weight += (cards[i+1] * v) * 2
			}
			if cards[i+2] >= 1 {
				weight += (cards[i+2] * v) * 1
			}
		} else if tmp == 9 {
			if cards[i-1] >= 1 {
				weight += (cards[i-1] * v) * 2
			}
			if cards[i-2] >= 1 {
				weight += (cards[i-2] * v) * 1
			}
		} else {
			if cards[i+1] >= 1 {
				weight += (cards[i+1] * v) * 2
			}
			if cards[i+2] >= 1 {
				weight += (cards[i+2] * v) * 1
			}
			if cards[i-1] >= 1 {
				weight += (cards[i-1] * v) * 2
			}
			if cards[i-2] >= 1 {
				weight += (cards[i-2] * v) * 1
			}
		}
	}

	if cards[i] == 1 {
		return weight
	}
	if cards[i] == 2 {
		weight += 4
	} else if cards[i] == 3 {
		weight += 18
	} else if cards[i] == 4 {
		weight += 48
	}
	return weight
}

//计算吃，杠，碰，牌权值返回万条筒的权值和数量，分别为万权值，万数量以此类推。
func getCPGNumAndWeight(inputCars int, cardsType int) (int, int, int, int, int, int, int, int) {
	wanWeight := 0
	wanNum := 0
	tiaoWeight := 0
	tiaoNum := 0
	tongWeight := 0
	tongNum := 0
	ziWeight := 0
	ziNum := 0
	y := 0
	z := 3
	if cardsType == chitype {
		y = 10
	} else if cardsType == pengtype {
		y = 18
	} else if cardsType == gangtype {
		y = 48
	}

	v := inputCars
	if v >= 1 && v <= 9 {
		wanWeight = y
		wanNum = 3
	} else if v >= 11 && v <= 19 {
		tiaoWeight = y
		tiaoNum = 3
	} else if v >= 21 && v <= 29 {

		tongWeight = y
		tongNum = z
	} else {
		ziWeight = y
		ziNum = z
	}

	return wanWeight, wanNum, tiaoWeight, tiaoNum, tongWeight, tongNum, ziWeight, ziNum
}
