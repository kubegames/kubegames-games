package game

import (
	"common/log"
	"game_poker/doudizhu/msg"
	"game_poker/doudizhu/poker"
	"math/rand"
	"strings"
	"time"
)

func ExcludeOne(srcCard []byte, ddz []uint32) ([]byte, []byte) {

	cardSlice := make([]byte, 0)

	//for _, v := range ddz {
	//	for _,card:=range srcCard{
	//		if v== card.Number {
	//			cardSlice=append(cardSlice,card)
	//
	//			RemoveOneCardEx(srcCard,card)
	//			delete(cardSet,card)
	//			continue
	//		}
	//	}
	//}

	for _, v := range ddz {
		for k, card := range srcCard {
			if card>>4 == byte(v) {
				cardSlice = append(cardSlice, card)
				srcCard = append(srcCard[:k], srcCard[k+1:]...)
				break
			}
		}
	}

	return srcCard, cardSlice
}

//获取一组牌的牌型
func GetCardsType(cards []byte) msg.CardsType {
	//重新创建切片，不改变原切片的数据
	var tempcards = CloneCards(cards)
	//排序，由小到大
	//sort.Sort(CardSlice(tempcards))
	poker.SortCards(tempcards)
	var cardstype msg.CardsType
	var length = len(tempcards)

	switch length {
	case 0:
		//无类型
		cardstype = msg.CardsType_Normal
		return cardstype
	case 1:
		//单张
		cardstype = msg.CardsType_SingleCard
		return cardstype
	case 2:
		if tempcards[0]>>4 == tempcards[1]>>4 {
			//对子
			cardstype = msg.CardsType_Pair
			return cardstype
		} else if tempcards[0] == 0xe1 && tempcards[1] == 0xf1 {
			//火箭
			cardstype = msg.CardsType_Rocket
			return cardstype
		}
	case 3:
		if tempcards[0]>>4 == tempcards[2]>>4 {
			//三张
			cardstype = msg.CardsType_Triplet
			return cardstype
		}
	case 4:
		if tempcards[0]>>4 == tempcards[3]>>4 {
			//炸弹
			cardstype = msg.CardsType_Bomb
			return cardstype
		} else if tempcards[0]>>4 == tempcards[2]>>4 || tempcards[1]>>4 == tempcards[3]>>4 {
			//三带一
			cardstype = msg.CardsType_TripletWithSingle
			return cardstype
		}
	case 5:
		if (tempcards[0]>>4 == tempcards[2]>>4 && tempcards[3]>>4 == tempcards[4]>>4) ||
			(tempcards[0]>>4 == tempcards[1]>>4 && tempcards[2]>>4 == tempcards[4]>>4) {
			//三带二
			cardstype = msg.CardsType_TripletWithPair
			return cardstype
		}
	}
	//顺子
	if bflag, cardstype := IsStraight(tempcards); bflag {
		return cardstype
	}
	//四带二
	if bflag, cardstype := IsFourAndTwo(tempcards); bflag {
		return cardstype
	}
	//飞机
	if bflag, cardstype := IsAirPlane(tempcards); bflag {
		return cardstype
	}
	return cardstype
}

//是否是顺子
func IsStraight(cards []byte) (bool, msg.CardsType) {
	var length = len(cards)
	if length < 5 {
		return false, msg.CardsType_Normal
	}
	//最大值不能大于2
	if cards[length-1]>>4 >= 13 {
		return false, msg.CardsType_Normal
	}
	_, counts := GetCardCounts(cards)
	var interval int
	for _, ct := range counts {
		if ct > 0 {
			if interval == 0 {
				interval = ct
			} else if interval > 0 && ct != interval {
				return false, msg.CardsType_Normal
			}
		}
	}
	if interval <= 0 {
		return false, msg.CardsType_Normal
	}
	var startnum, endnum int
	for num, ct := range counts {
		if ct > 0 {
			if startnum == 0 {
				startnum = num
			}
			endnum = num
		} else if startnum > 0 {
			break
		}
	}
	cnt := endnum - startnum + 1
	if cnt*interval != length {
		return false, msg.CardsType_Normal
	}
	switch interval {
	case 1:
		if cnt >= 5 {
			return true, msg.CardsType_Sequence
		}
	case 2:
		if cnt >= 3 {
			return true, msg.CardsType_SerialPair
		}
	case 3:
		if cnt >= 2 {
			return true, msg.CardsType_SerialTriplet
		}
	}
	return false, msg.CardsType_Normal
}

//是否是四带二
func IsFourAndTwo(cards []byte) (bool, msg.CardsType) {
	length := len(cards)
	if length != 6 && length != 8 {
		return false, msg.CardsType_Normal
	}
	ishas, number := IsHasBomb(cards)
	log.Debugf("%v", ishas)
	if !ishas {
		return false, msg.CardsType_Normal
	}
	var counts = make([]int, 18)
	for _, card := range cards {
		if card>>4 != byte(number) {
			counts[card>>4]++
		}
	}
	var tempct1 = 0
	var tempct2 = 0
	for _, count := range counts {
		tempct1 += count
		if count == 2 {
			tempct2++
		} else if count == 4 {
			tempct2 += 2
		}
	}
	log.Debugf("%v", tempct1)
	if tempct1 == 2 {
		return true, msg.CardsType_QuartetWithTwo
	} else if tempct2 == 2 {
		return true, msg.CardsType_QuartetWithTwoPair
	}
	return false, msg.CardsType_Normal
}

//是否是飞机带翅膀
func IsAirPlane(cards []byte) (bool, msg.CardsType) {
	if bflag, startnum, endnum := IsHasThreeStraight(cards); bflag {
		length := len(cards)
		_, counts := GetCardCounts(cards)
		//由大到小遍历，例如555666777888，可以取较大的飞机 666777888
		for s := endnum - 1; s >= startnum; s-- {
			for e := endnum; e > s; e-- {
				threeCnt := e - s + 1
				//长度校验，要么飞机单带张，要么飞机带对子
				if length != threeCnt*(3+1) && length != threeCnt*(3+2) {
					continue
				}
				temps := CloneInts(counts)
				for i := range temps {
					if i >= s && i <= e {
						temps[i] -= 3
					}
				}
				var tempct1 = 0
				var tempct2 = 0
				for _, count := range temps {
					tempct1 += count
					if count == 2 {
						tempct2++
					} else if count == 4 {
						tempct2 += 2
					}
				}
				if tempct1 == threeCnt {
					return true, msg.CardsType_SerialTripletWithOne
				} else if tempct2 == threeCnt {
					return true, msg.CardsType_SerialTripletWithWing
				}
			}
		}
	}
	return false, msg.CardsType_Normal
}

//是否有三顺子
func IsHasThreeStraight(cards []byte) (bool, int, int) {
	var length = len(cards)
	if length < 6 {
		return false, 0, 0
	}
	_, counts := GetCardCounts(cards)
	var result = false
	var startnum = 0
	var endnum = 0
	for num := 1; num < 13; num++ {
		if counts[num] >= 3 && counts[num+1] >= 3 {
			result = true
			if startnum == 0 {
				startnum = num
			}
			endnum = num + 1
		} else if result {
			break
		}
	}
	return result, startnum, endnum
}

//获取飞机牌型中的三顺子头尾数字, 参数只支持飞机带翅膀牌型, 不带翅膀的飞机是三顺子
func GetAirPlaneStartEnd(cards []byte) (int, int) {
	length := len(cards)
	_, counts := GetCardCounts(cards)
	bflag, startnum, endnum := IsHasThreeStraight(cards)
	ErrorCheck(bflag, 0, " cards type maybe illegal !!!")
	var rtStart, rtEnd int
	//由大到小遍历，例如555666777888，可以取较大的飞机 666777888
outloop:
	for s := endnum - 1; s >= startnum; s-- {
		for e := endnum; e > s; e-- {
			threeCnt := e - s + 1
			//长度校验，要么飞机单带张，要么飞机带对子
			if length != threeCnt*(3+1) && length != threeCnt*(3+2) {
				continue
			}
			temps := CloneInts(counts)
			for i := range temps {
				if i >= s && i <= e {
					temps[i] -= 3
				}
			}
			var tempct1 = 0
			var tempct2 = 0
			for _, count := range temps {
				tempct1 += count
				if count == 2 {
					tempct2++
				} else if count == 4 {
					tempct2 += 2
				}
			}
			if tempct1 == threeCnt || tempct2 == threeCnt {
				rtStart = s
				rtEnd = e
				break outloop
			}
		}
	}
	return rtStart, rtEnd
}

//是否有炸弹
func IsHasBomb(cards []byte) (bool, int) {
	_, counts := GetCardCounts(cards)
	for i := 13; i >= 1; i-- {
		if counts[i] == 4 {
			return true, i
		}
	}
	return false, 0
}

//比较2组牌的大小,cards2 > cards1 返回true, 否则false
func CompareCards(cards1 []byte, cards2 []byte) bool {
	//先排序，无论传入参数是否有序，保证返回正确结果
	poker.SortCards(cards1)
	poker.SortCards(cards2)
	var cardstype1 = GetCardsType(cards1)
	var cardstype2 = GetCardsType(cards2)

	if cardstype1 == msg.CardsType_Normal {
		return cardstype2 > msg.CardsType_Normal
	}

	if cardstype2 == msg.CardsType_Normal {
		return false
	}

	if cardstype1 == msg.CardsType_Rocket {
		return false
	}

	if cardstype2 == msg.CardsType_Rocket {
		return true
	}

	if cardstype1 == msg.CardsType_Bomb && cardstype2 < msg.CardsType_Bomb {
		return false
	}

	if cardstype2 == msg.CardsType_Bomb && cardstype1 < msg.CardsType_Bomb {
		return true
	}
	var length1 = len(cards1)
	var length2 = len(cards2)
	if length1 != length2 {
		return false
	}

	if cardstype2 == msg.CardsType_SerialTriplet && cardstype1 == msg.CardsType_SerialTripletWithOne {
		//特殊情况，防止JJJQQQKKKAAA打不过333444555678的情况
	} else if cardstype1 != cardstype2 {
		return false
	}
	switch cardstype1 {
	case msg.CardsType_SingleCard:
		fallthrough
	case msg.CardsType_Pair:
		fallthrough
	case msg.CardsType_Triplet:
		fallthrough
	case msg.CardsType_Bomb:
		return cards2[0]>>4 > cards1[0]>>4
	case msg.CardsType_TripletWithSingle:
		return cards2[1]>>4 > cards1[1]>>4
	case msg.CardsType_TripletWithPair:
		return cards2[2]>>4 > cards1[2]>>4
	case msg.CardsType_Sequence:
		fallthrough
	case msg.CardsType_SerialPair:
		fallthrough
	case msg.CardsType_SerialTriplet:
		return cards2[length2-1]>>4 > cards1[length1-1]>>4
	case msg.CardsType_SerialTripletWithOne:
		fallthrough
	case msg.CardsType_SerialTripletWithWing:
		start1, end1 := GetAirPlaneStartEnd(cards1)
		start2, end2 := GetAirPlaneStartEnd(cards2)
		if end1-start1 != end2-start2 {
			return false
		}
		return start2 > start1
	case msg.CardsType_QuartetWithTwo:
		_, number1 := IsHasBomb(cards1)
		_, number2 := IsHasBomb(cards2)
		return number2 > number1
	}
	return false
}

//获取牌数目
func GetCardCounts(cards []byte) (int, []int) {
	//牌个数
	var counts = make([]int, 16)
	for _, card := range cards {
		counts[card>>4]++
	}

	//炸弹数量
	var bombct = 0
	for _, ct := range counts {
		if ct == 4 {
			bombct++
		}
	}
	return bombct, counts
}

//获取牌在切片中的索引
func GetCardIndex(cards []byte, card byte) int {
	for i := 0; i < len(cards); i++ {
		if cards[i] == card {
			return i
		}
	}
	return -1
}

//获取出牌
func GetOutCards(cards []byte, outs []byte) []byte {
	cardstype := GetCardsType(cards)
	if cardstype == msg.CardsType_Rocket {
		return cards
	}
	if nil == outs || GetCardsType(outs) <= msg.CardsType_Normal {
		getMinCards := func(counts []int) []byte {
			var minNum int
			for num, ct := range counts {
				if ct > 0 {
					minNum = num
					break
				}
			}
			var minCards []byte
			for _, card := range cards {
				if card>>4 == byte(minNum) {
					minCards = append(minCards, card)
				}
			}
			return minCards
		}
		bombct, counts := GetCardCounts(cards)
		//只剩最后一手牌，且手上有王炸，防止王炸+炸弹被当成4带2单打出
		if cardstype == msg.CardsType_QuartetWithTwo && counts[14] > 0 && counts[15] > 0 {
			return getMinCards(counts)
		}
		if cardstype > msg.CardsType_Normal && bombct*4 != len(cards) {
			//只剩一手牌，且手上不是全炸弹，防止55556666，被当成4带2对打出
			return cards
		}
		return getMinCards(counts)
	}
	exist, result := SearchLargerCardType(outs, cards, true)
	if !exist || 0 == len(result) {
		return nil
	}
	return result[0]
}

//反序字符串
func ReverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

//反序一个切片
func ReverseCards(cards []byte) []byte {
	for i, j := 0, len(cards)-1; i < j; i, j = i+1, j-1 {
		cards[i], cards[j] = cards[j], cards[i]
	}
	return cards
}

//排序
func SortCards(cards []byte) []byte {
	poker.SortCards(cards)
	return cards
}

//clone
func CloneCards(cards []byte) []byte {
	//重新创建切片，不改变原切片的数据
	tempcards := make([]byte, len(cards))
	copy(tempcards, cards)
	return tempcards
}

//clone int
func CloneInts(ints []int) []int {
	//重新创建切片，不改变原切片的数据
	temps := make([]int, len(ints))
	copy(temps, ints)
	return temps
}

//增加一个元素
func AddOneCard(cardsDst []byte, card byte) []byte {
	cardsDst = append(cardsDst, card)
	poker.SortCards(cardsDst)
	return cardsDst
}

//增加一个切片
func AddCards(cardsDst []byte, cards []byte) []byte {
	cardsDst = append(cardsDst, cards...)
	poker.SortCards(cardsDst)
	return cardsDst
}

//删除一个元素并排序
func RemoveOneCard(cardsDst []byte, card byte) (bool, []byte) {
	for k, v := range cardsDst {
		if card == v {
			cardsDst = append(cardsDst[:k], cardsDst[k+1:]...)
			poker.SortCards(cardsDst)
			return true, cardsDst
		}
	}
	return false, cardsDst
}

//删除一个元素, 不排序
func RemoveOneCardEx(cardsDst []byte, card byte) (bool, []byte) {
	for k, v := range cardsDst {
		if card == v {
			cardsDst = append(cardsDst[:k], cardsDst[k+1:]...)
			return true, cardsDst
		}
	}
	return false, cardsDst
}

//删除一个切片并排序
func RemoveCards(cardsDst []byte, cards []byte) (bool, []byte) {
	//安全删除检测，先检查cardsDst里面是否包含cards
	var tempcards []byte
	for _, v := range cards {
		var bExist = false
		for _, v2 := range cardsDst {
			var bRepeat = false
			for _, v3 := range tempcards {
				if v2 == v3 {
					bRepeat = true
					break
				}
			}
			if bRepeat {
				continue
			}
			if v == v2 {
				bExist = true
				tempcards = append(tempcards, v2)
				break
			}
		}
		if !bExist {
			return false, cardsDst
		}
	}

	if len(tempcards) != len(cards) {
		return false, cardsDst
	}

	for _, v := range cards {
		for k2, v2 := range cardsDst {
			if v == v2 {
				cardsDst = append(cardsDst[:k2], cardsDst[k2+1:]...)
				break
			}
		}
	}

	poker.SortCards(cardsDst)
	return true, cardsDst
}

//查询切片里面是否包含某张牌
func IsContainNumber(cards []byte, number byte) bool {
	var rt bool
	for _, card := range cards {
		if card>>4 == number {
			rt = true
			break
		}
	}
	return rt
}

//随机打乱牌组
func ShuffleCards(cards []byte) []byte {
	// 时间随机
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 打乱旧牌盒下标
	perm := r.Perm(len(cards))
	//新的牌堆
	newcards := make([]byte, len(cards))
	// 牌盒遍历获取新的牌盒
	for i, rindex := range perm {
		newcards[i] = cards[rindex]
	}
	// 返回洗过的牌
	return newcards
}

//随机打乱整型
func ShuffleInts(nums []int) []int {
	// 时间重置随机种子
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 无序
	perm := r.Perm(len(nums))
	// 新的切片
	newInts := make([]int, len(nums))
	// 遍历获取
	for i, rindex := range perm {
		newInts[i] = nums[rindex]
	}
	// 返回
	return newInts
}

//生成count个[start,end]结束的不重复的随机数
func GetRandInts(start int, end int, count int) []int {
	//范围检查
	if end < start || (end-start+1) < count {
		return nil
	}
	//存放结果的slice
	nums := make([]int, 0)
	//随机数生成器，加入时间戳保证每次生成的随机数不一样
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for len(nums) < count {
		//生成随机数
		num := r.Intn(end-start+1) + start
		//查重
		exist := false
		for _, v := range nums {
			if v == num {
				exist = true
				break
			}
		}
		if !exist {
			nums = append(nums, num)
		}
	}
	return nums
}

//随机获取[0,total)范围内的一个整数
func GetRandInt(total int) int {
	ErrorCheck(total > 0, 0, "total should more than 0 !!!")
	// 时间随机
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 打乱下标, 生成[0,n)的随机整型切片
	perm := r.Perm(total)
	// Intn以int形式返回[0，n）中的非负伪随机数。
	index := r.Intn(total)
	return perm[index]
}

func CardSliceToString(self []byte) string {
	sb := strings.Builder{}
	lenth := len(self)
	for k, v := range self {
		if k < lenth-1 {
			sb.WriteString(String(int32(v>>4)) + ",")
		} else if k == lenth-1 {
			sb.WriteString(String(int32(v >> 2)))
		}

	}
	return sb.String()
}

func String(n int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}
