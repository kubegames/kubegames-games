package majiang

import (
	"math/rand"
)

//麻将牌类型
const (
	CardType     = 0
	CardTypeWan  = 0x1
	CardTypeTiao = 0x2
	CardTypeTong = 0x4
	CardTypeZi   = 0x8
	CardTypeBai  = 0x10
	CardTypeHua  = 0x20
)

//操作类型
const (
	OptTypeOutCard  = 0x1
	OptTypeZuoChi   = 0x2
	OptTypeZhongChi = 0x4
	OptTypeYouChi   = 0x8
	OptTypePeng     = 0x10
	OptTypeMingGang = 0x20
	OptTypeGang     = 0x40
	OptTypeAnGang   = 0x80
	OptTypeHu       = 0x100
)

//万条同
var Wan = []int{1, 2, 3, 4, 5, 6, 7, 8, 9}

var Tiao = []int{11, 12, 13, 14, 15, 16, 17, 18, 19}

var Tong = []int{21, 22, 23, 24, 25, 26, 27, 28, 29}

//东南西北中发
var Zi = []int{31, 32, 33, 34, 35, 36}

//白板
var Bai = []int{37}

//花牌
var Hua = []int{38, 39, 40, 41, 42, 43, 44, 45}

var CardStrings = []string{"", "一万", "二万", "三万", "四万", "五万", "六万", "七万", "八万", "九万",
	"", "幺鸡", "二条", "三条", "四条", "五条", "六条", "七条", "八条", "九条",
	"", "一筒", "二筒", "三筒", "四筒", "五筒", "六筒", "七筒", "八筒", "九筒",
	"", "东", "南", "西", "北", "中", "发", "白", "梅", "兰", "竹", "菊", "春", "夏", "秋", "冬"}

//最大牌花+1，作为数组下标
const MaxCardValue = 46

var AllCards = [][]int{Wan, Tiao, Tong, Zi, Bai, Hua}
var CardsCount = []int{4, 4, 4, 4, 4, 1}

//有哪些牌0x1万，0x2条,0x4同，0x8字牌，0x10白板，0x20花牌
var HasCards = []int{0x1, 0x2, 0x4, 0x8, 0x10, 0x20}

type MaJiang struct {
	Cards        [144]int
	MaxCardIndex int //最多牌索引
	CurrIndex    int //当前牌的索引
}

//v,有哪些牌,初始牌调用一次就可以了
func (mj *MaJiang) InitMaJiang(v int) {
	mj.CurrIndex = 0
	mj.MaxCardIndex = 0
	mj.Cards = [144]int{}
	for i := 0; i < 6; i++ {
		if (HasCards[i] & v) != 0 {
			for _, cardvalue := range AllCards[i] {
				for m := 0; m < CardsCount[i]; m++ {
					mj.Cards[mj.MaxCardIndex] = cardvalue
					mj.MaxCardIndex++
				}
			}
		}
	}
}

//洗牌
func (mj *MaJiang) FlushCards() {
	mj.CurrIndex = 0
	for i := 0; i < mj.MaxCardIndex; i++ {
		r := rand.Intn(mj.MaxCardIndex)
		mj.Cards[i], mj.Cards[r] = mj.Cards[r], mj.Cards[i]
	}

	//最后一张不是花牌，前端显示会有问题
	if mj.Cards[mj.MaxCardIndex] >= Hua[0] {
		r := rand.Intn(mj.MaxCardIndex - 1)
		mj.Cards[mj.MaxCardIndex], mj.Cards[r] = mj.Cards[r], mj.Cards[mj.MaxCardIndex]
	}
	/*
		mj.Cards = [144]int{1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4, 4, 37,
			5, 5, 5, 6, 6, 6, 7, 7, 7, 8, 8, 8, 37,
			1, 2, 3, 4, 5, 6, 7, 8,
			9, 9, 9, 9, 31, 31, 31, 31, 32, 32, 32, 32,
			33, 33, 33, 33, 34, 34, 34, 34, 35, 35, 35, 35,
			38, 39, 40, 41, 42, 43, 44, 45,
			36, 36, 36, 36, 37, 37}
	*/
}

//发牌
func (mj *MaJiang) DealCard() int {
	if mj.MaxCardIndex <= mj.CurrIndex {
		return 0
	}

	var temp [MaxCardValue]int
	for i := mj.CurrIndex; i < mj.MaxCardIndex; i++ {
		temp[mj.Cards[i]]++
	}

	mj.CurrIndex++
	return mj.Cards[mj.CurrIndex-1]
}

//获取剩余牌中特定的牌
func (mj *MaJiang) GetCardValue(CardValue int) int {
	for i := mj.CurrIndex; i < mj.MaxCardIndex; i++ {
		if CardValue == mj.Cards[i] {
			mj.Cards[i], mj.Cards[mj.CurrIndex] = mj.Cards[mj.CurrIndex], mj.Cards[i]
			mj.CurrIndex++
			return CardValue
		}
	}

	return 0
}

//获取剩余牌中特定的牌的张数
func (mj *MaJiang) GetCardCount(CardValue int) int {
	count := 0
	for i := mj.CurrIndex; i < mj.MaxCardIndex; i++ {
		if CardValue == mj.Cards[i] {
			count++
		}
	}

	return count
}

func (mj *MaJiang) ResaultCard(card int) {
	if mj.Cards[mj.CurrIndex-1] == card {
		mj.Cards[mj.CurrIndex-2], mj.Cards[mj.CurrIndex-1] = mj.Cards[mj.CurrIndex-1], mj.Cards[mj.CurrIndex-2]
	}

	mj.CurrIndex--
	var temp [MaxCardValue]int
	for i := mj.CurrIndex; i < mj.MaxCardIndex; i++ {
		temp[mj.Cards[i]]++
	}
}

//剩余牌张数
func (mj *MaJiang) GetLastCardsCount() int {
	return mj.MaxCardIndex - mj.CurrIndex
}

//吃牌类型,左吃，中吃，右吃
func CanChi(Cards [MaxCardValue]int, OutCard int) int {
	//字牌以上直接返回
	if OutCard >= Zi[0] {
		return 0
	}

	t := 0
	if OutCard >= 0x3 && Cards[OutCard-2] >= 1 && Cards[OutCard-1] >= 1 {
		t |= OptTypeZuoChi
	}

	if Cards[OutCard-1] >= 1 && Cards[OutCard+1] >= 1 {
		t |= OptTypeZhongChi
	}

	if Cards[OutCard+1] >= 1 && Cards[OutCard+2] >= 1 {
		t |= OptTypeYouChi
	}

	return t
}

//碰和杠
func CanPengAndGang(Cards [MaxCardValue]int, CardValue int) int {
	tmp := 0
	if Cards[CardValue] >= 2 {
		tmp |= OptTypePeng
	}

	if Cards[CardValue] >= 3 {
		tmp |= OptTypeGang
	}

	return tmp
}

//明杠
func CanMingGang(Cards [4]int, CardValue int, cards [MaxCardValue]int, GangCard [4]int) int {
	j := 0
	OptType := 0
	for _, v := range Cards {
		if v <= 0 {
			continue
		}
		if v == CardValue || cards[v] >= 1 {
			GangCard[j] = v
			j++
			OptType = OptTypeMingGang
		}
	}

	return OptType
}

//暗杠
func CanAnGang(Cards [MaxCardValue]int, CardValue int, GangCard [4]int) int {
	tmp := 0
	tempCards := Cards
	j := 0
	tempCards[CardValue]++
	for i := 1; i < Hua[0]; i++ {
		if tempCards[i] == 4 {
			tmp |= OptTypeAnGang
			GangCard[j] = i
			j++
		}
	}
	return tmp
}

func CanHu(Cards [MaxCardValue]int, CardValue int32) (int, int32) {
	tmp := Cards
	tmp[CardValue] += 1

	//这里不遍历花牌
	//麻将是2个的统计
	duicount := 0
	//开始的位置
	startindex := 0
	//结束的位置记录开始和结束可以减少循环次数
	endindex := 0
	//最多7对将牌
	var jiang [7]int
	jiangcount := 0
	for i := 1; i <= Bai[0]; i++ {
		if tmp[i] >= 1 {
			if tmp[i] >= 2 {
				jiang[jiangcount] = i
				jiangcount++
				if tmp[i] == 4 {
					duicount++
				}
			}
			if startindex == 0 {
				startindex = i
			}

			endindex = i
		} else if tmp[i] < 0 {
			return 0, CardValue
		}
	}

	//7对胡牌
	if (jiangcount + duicount) == 7 {
		return OptTypeHu, CardValue
	}

	sub := 0
	bFu := true
	for m := 0; m < jiangcount; m++ {
		tmpfu := tmp
		bFu = true
		//处理将牌的情况
		tmpfu[jiang[m]] -= 2
		//分两步处理，减少循环次数
		for i := startindex; i < Zi[0]; i++ {
			if tmpfu[i] == 0 || tmpfu[i] == 3 {
				continue
			}
			sub = tmpfu[i] % 3

			tmpfu[i+1] -= sub
			tmpfu[i+2] -= sub

			if tmpfu[i+1] < 0 || tmpfu[i+2] < 0 {
				bFu = false
				break
			}
		}

		if bFu {
			for i := Zi[0]; i <= endindex; i++ {
				if tmpfu[i] == 0 || tmpfu[i] == 3 {
					continue
				}

				bFu = false
				break
			}

			if bFu {
				return OptTypeHu, CardValue
			}
		}
	}
	return 0, CardValue
}

//找将对
func CanJiangHu(Cards [MaxCardValue]int, CardValue int32, JiangValue int32) (int, int32) {
	tmp := Cards
	tmp[CardValue] += 1

	//这里不遍历花牌
	//麻将是2个的统计

	bFu := true
	//处理将牌的情况
	tmp[JiangValue] -= 2
	//分两步处理，减少循环次数
	for i := Wan[0]; i < Zi[0]; i++ {
		if tmp[i] == 0 || tmp[i] == 3 {
			continue
		}
		sub := tmp[i] % 3

		tmp[i+1] -= sub
		tmp[i+2] -= sub

		if tmp[i+1] < 0 || tmp[i+2] < 0 {
			bFu = false
			break
		}
	}

	if bFu {
		for i := Zi[0]; i <= Bai[0]; i++ {
			if tmp[i] == 0 || tmp[i] == 3 {
				continue
			}

			bFu = false
			break
		}

		if bFu {
			return OptTypeHu, CardValue
		}
	}

	return 0, CardValue
}

//排序后的牌。
func GetHandCards(Cards [MaxCardValue]int) (ret []int32) {
	for i := 1; i < MaxCardValue; i++ {
		for n := 0; n < Cards[i]; n++ {
			ret = append(ret, int32(i))
		}
	}

	return ret
}

//按照出牌无序的
func GetCards(Cards [MaxCardValue]int) (ret []int32) {
	for i := 1; i < MaxCardValue; i++ {
		for n := 0; n < Cards[i]; n++ {
			ret = append(ret, int32(i))
		}
	}

	return ret
}

// 获取当前剩余的牌
func (mj *MaJiang) GetLastCardArray() []int {
	return mj.Cards[mj.CurrIndex:]
}

func GetHandCardString(Cards [MaxCardValue]int) string {
	totalString := ""
	for i := 1; i < MaxCardValue; i++ {
		for n := 0; n < Cards[i]; n++ {
			if len(totalString) != 0 {
				totalString += "-"
			}
			totalString += CardStrings[i]
		}
	}
	return totalString
}

func GetOutCardString(Cards [144]int32) string {
	totalString := ""
	for i := 0; i < 144; i++ {
		if Cards[i] == 0 {
			break
		}
		totalString += CardStrings[Cards[i]] + "-"
	}
	return totalString
}

func InitTestCards(Cards []int32) [MaxCardValue]int {
	var temp [MaxCardValue]int
	for _, val := range Cards {
		temp[val] += 1
	}
	return temp
}
func InitTestCardsInt(Cards []int) [MaxCardValue]int {
	var temp [MaxCardValue]int
	for _, val := range Cards {
		temp[val] += 1
	}
	return temp
}
