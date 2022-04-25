package gamelogic

import (
	majiangcom "github.com/kubegames/kubegames-games/internal/pkg/majiang"
	"github.com/kubegames/kubegames-games/pkg/battle/940101/def"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func GetCardsOdds(userData *UserData) (mask, noMask int64) {
	andCard := userData.HandCards
	userData.GetHanCardWANNumber(andCard)
	userData.GetHanCardZINumber(andCard)
	//userData.GetHanCardZINumber(andCard)
	gangNumber := userData.GangPaiNumber + userData.AnGangPaiNumber
	pengNumber := userData.PengPaiNumber
	chiNumber := userData.ChiPaiNumber

	if len(majiangcom.GetHandCards(userData.HandCards)) == 2 {
		// 边张或坎张
		mask |= def.DanDiao
	}

	if userData.AnGangPaiNumber == 2 {
		mask |= def.ShuangAnGang
	} else if userData.AnGangPaiNumber == 1 {
		mask |= def.AnGang
	}
	if userData.GangPaiNumber >= 1 {
		mask |= def.MingGang
	}

	temp1 := andCard // 杠加手牌
	temp2 := andCard // 吃加手牌
	temp3 := andCard // 吃碰杠加手牌
	temp4 := andCard // 吃碰加手牌
	temp5 := andCard // 碰杠加手牌
	temp6 := andCard // 碰加手牌

	for _, val := range userData.GangPai { // and 杠牌
		temp1[val] = 4
		temp3[val] = 4
		temp5[val] = 4
		if val == 1 || val == 9 {
			//幺九刻的特殊情况
			mask |= def.YaoJiuKe
		}
	}

	for _, val := range userData.AnGangPai { // and 杠牌
		temp1[val] = 4
		temp3[val] = 4
		temp5[val] = 4
		if val == 1 || val == 9 {
			//幺九刻的特殊情况
			mask |= def.YaoJiuKe
		}
	}

	for _, val := range userData.PengCards { // and 碰牌
		temp3[val] += 3
		temp4[val] += 3
		temp5[val] += 3
		temp6[val] += 3
		if val == 1 || val == 9 || (val >= majiangcom.Zi[0] && val <= majiangcom.Zi[3]) {
			//幺九刻的特殊情况
			mask |= def.YaoJiuKe
		}
	}

	for _, val := range userData.ChiPai { // and 吃牌
		temp2[val] += 1
		temp3[val] += 1
		temp4[val] += 1
	}
	//统计四归一
	siGuiYiCount := 0
	for key, val := range temp4 {
		if val == 4 && key > 0 {
			siGuiYiCount++
			log.Debugf("四归一统计 %v", key)
		}
	}

	log.Debugf("%v", majiangcom.GetHandCardString(temp4))
	if siGuiYiCount > 0 {
		mask |= def.SiGuiYi
	}

	userData.SiGuiYiNumber = siGuiYiCount
	userData.GetHanCardWANNumber(temp3)
	userData.GetHanCardZINumber(temp3)

	if userData.WANNumber != 0 {
		IsHunYaoJiu(temp3, &mask, &noMask)
		IsPengPengHe(temp5, &mask, &noMask)
	}

	if pengNumber == 0 && gangNumber == 0 && chiNumber == 0 { // 吃、碰、杠 都没有
		IsJiuBaoLianDeng(andCard, &mask, &noMask)
		if mask&def.PengPengHe != 0 {
			mask |= def.SiAnKe
			noMask |= def.MengQianQing | def.PengPengHe | def.SanAnKe | def.ShuangAnGang | def.BuQiuRen
		} else {
			IsSiAnKe(andCard, &mask, &noMask)
		}

		IsQiDui(andCard, &mask, &noMask)
	} else {
		if userData.GangPaiNumber == 2 {
			mask |= def.ShuangMingGang
			noMask |= def.MingGang
		}

		switch userData.AnGangPaiNumber + userData.GangPaiNumber {
		case 3:
			mask |= def.SanGang
			noMask |= def.ShuangMingGang | def.ShuangAnGang | def.MingGang | def.AnGang
		case 4:
			mask |= def.SiGang
			noMask |= def.SanGang | def.ShuangAnGang | def.ShuangMingGang | def.MingGang | def.AnGang | def.DanDiao
		}
	}

	IsAnKe(andCard, &mask, &noMask)
	IsDaSiXi(temp3, &mask, &noMask)
	if userData.WANNumber == 0 {
		mask |= def.ZiYiSe
		noMask |= def.PengPengHe | def.HunYaoJiu | def.QuanDaiYao | def.YaoJiuKe | def.HunYiSe
	}
	IsDaSanYuan(temp3, &mask, &noMask)
	IsXiaoSiXi(temp5, &mask, &noMask)
	IsXiaoSanYuan(temp3, &mask, &noMask)
	IsQingYiSe(temp3, &mask, &noMask)
	IsYiSeSanJieGao(temp6, &mask, &noMask)
	IsShuangJianKe(temp5, &mask, &noMask)
	if mask&def.YiSeSanJieGao == 0 {
		IsYiSeSanTongShun(andCard, userData.ChiPai, &mask, &noMask)
	}
	IsYiSeSanBuGao(temp2, &mask, &noMask)
	IsYiSeSiBuGao(temp2, &mask, &noMask)
	IsYiSeSiTongShun(andCard, userData.ChiPai, &mask, &noMask)
	IsQingLong(temp2, &mask, &noMask)
	if userData.ZiPengNumber >= 3 {
		IsSanFengKe(temp3, &mask, &noMask)
	}

	IsJianKe(temp6, &mask, &noMask)

	if mask&def.QingYiSe != 0 {
		IsPingHu(temp2, &mask, &noMask)
	}

	IsQuanDaiYao(temp3, &mask, &noMask)
	IsDuanYao(temp3, &mask, &noMask)
	IsYiBanGao(temp2, &mask, &noMask)
	IsLaoShaoFu(temp2, &mask, &noMask)
	IsLianLiu(temp2, &mask, &noMask)
	IsYaoJiuKe(andCard, &mask, &noMask)
	IsBianZhangOrKangZhang(andCard, userData.HuOptCard, &mask, &noMask)

	return
}

func IsDaSiXi(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Zi[0]; i <= majiangcom.Zi[3]; i++ {
		if Cards[i] < 3 {
			return
		}
	}

	*mask |= def.DaSiXi
	*noMask |= def.SanFengKe | def.PengPengHe | def.YaoJiuKe
}

func IsDaSanYuan(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Zi[4]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] < 3 {
			return
		}
	}
	*mask |= def.DaSanYuan
	*noMask |= def.JianKe | def.ShuangJianKe | def.YaoJiuKe
	return
}
func IsJiuBaoLianDeng(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	arr := [9]int{3, 1, 1, 1, 1, 1, 1, 1, 3}
	for index, val := range arr {
		if Cards[index+1] < val {
			return
		}
	}
	*mask |= def.JiuBaoLianDeng
	*noMask |= def.QingYiSe | def.BuQiuRen | def.MengQianQing | def.YaoJiuKe
	return
}
func IsSiGang(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Wan[0]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] == 4 {
			count++
		}
	}
	if count < 4 {
		return
	} else {
		*mask |= def.SiGang
		*noMask |= def.SanGang | def.ShuangAnGang | def.ShuangMingGang | def.MingGang | def.AnGang | def.DanDiao
		return
	}
}

func IsLianQiDui(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	Count := 0
	var arr [7]int
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if Cards[i] == 2 {
			arr[Count] = i
			Count++
		}
	}
	if Count == 7 && arr[6]-arr[0] == 6 {
		*mask |= def.LianQiDui
		*noMask |= def.QingYiSe | def.BuQiuRen | def.MengQianQing |
			def.QiDui | def.DanDiao
		return
	}
	return
}

func IsXiaoSiXi(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Zi[0]; i <= majiangcom.Zi[3]; i++ {
		if Cards[i] < 2 {
			return
		} else if Cards[i] == 2 {
			count++
			if count > 1 {
				return
			}
		}
	}

	if count != 1 {
		return
	}

	*mask |= def.XiaoSiXi
	*noMask |= def.SanFengKe | def.YaoJiuKe
}

func IsXiaoSanYuan(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	flag := false
	for i := majiangcom.Zi[4]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] >= 3 {
			count++
		} else if Cards[i] == 2 {
			flag = true
		}
	}

	if count < 2 || !flag {
		return
	}
	*mask |= def.XiaoSanYuan
	*noMask |= def.JianKe | def.ShuangJianKe | def.YaoJiuKe
	return
}

func IsZiYiSe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if Cards[i] > 0 {
			return
		}
	}
	*mask |= def.ZiYiSe
	*noMask |= def.PengPengHe | def.HunYaoJiu | def.QuanDaiYao | def.YaoJiuKe | def.HunYiSe
	return
}

func IsSiAnKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	temp := Cards
	count := 0
	for key, val := range Cards {
		if val >= 3 && key > 0 {
			temp[key] -= 3
			count++
		}
	}

	if count >= 4 {
		ret, _ := majiangcom.CanHu(temp, 0)
		if ret != 0 {
			*mask |= def.SiAnKe
			*noMask |= def.MengQianQing | def.PengPengHe | def.SanAnKe | def.ShuangAnGang | def.BuQiuRen
		}
	}
}

func IsYiSeShuangLongHui(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	if Cards[majiangcom.Wan[4]] != 2 {
		return
	}
	var arr []int
	arr = append(append(arr, majiangcom.Wan[:3]...), majiangcom.Wan[6:]...)
	for _, val := range arr {
		if Cards[val] < 1 {
			return
		}
	}
	*mask |= def.YiSeShuangLongHun
	*noMask |= def.PingHu | def.QiDui | def.QingYiSe | def.YiBanGao | def.LaoShaoFu
	return
}
func IsYiSeSiTongShun(Cards [majiangcom.MaxCardValue]int, ChiPai [12]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[6]; i++ {
		count := 0
		for j := 0; j < 12; j += 3 {
			if ChiPai[j] == 0 {
				break
			}

			if ChiPai[j] == i {
				count++
			}
		}

		if Cards[i]+count == 4 && Cards[i+1]+count == 4 && Cards[i+2]+count == 4 {
			*mask |= def.YiSeSiTongShun
			*noMask |= def.YiSeSanJieGao | def.YiBanGao |
				def.SiGuiYi | def.YiSeSanTongShun | def.QiDui
			return
		}
	}
}

func IsYiSeSiJieGao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[5]; i++ {
		if Cards[i] >= 3 && Cards[i+1] >= 3 && Cards[i+2] >= 3 && Cards[i+3] >= 3 {
			temp := Cards
			temp[i] -= 3
			temp[i+1] -= 3
			temp[i+2] -= 3
			temp[i+3] -= 3
			ret, _ := majiangcom.CanHu(temp, 0)
			if ret != 0 {
				*mask |= def.YiSeSiJieGao
				*noMask |= def.YiSeSanJieGao | def.YiSeSanTongShun | def.PengPengHe
				return
			}
		}
	}
}

func IsYiSeSiBuGao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[3]; i++ {
		tempCards := Cards
		if tempCards[i] > 0 &&
			tempCards[i+1] >= 2 &&
			tempCards[i+2] >= 3 &&
			tempCards[i+3] >= 3 &&
			tempCards[i+4] >= 2 &&
			tempCards[i+5] > 0 {
			tempCards[i] -= 1
			tempCards[i+1] -= 2
			tempCards[i+2] -= 3
			tempCards[i+3] -= 3
			tempCards[i+4] -= 2
			tempCards[i+5] -= 1
			tempOpt, _ := majiangcom.CanHu(tempCards, 0)
			if tempOpt&majiangcom.OptTypeHu != 0 {
				*mask |= def.YiSeSiBuGao
				*noMask |= def.YiSeSanBuGao
				return
			}
		}

		tempCards = Cards
		if tempCards[i] >= 1 &&
			tempCards[i+1] >= 1 &&
			tempCards[i+2] >= 2 &&
			tempCards[i+3] >= 1 &&
			tempCards[i+4] >= 2 &&
			tempCards[i+5] >= 1 &&
			tempCards[i+6] >= 2 &&
			tempCards[i+7] >= 1 &&
			tempCards[i+8] >= 1 {
			tempCards[i] -= 1
			tempCards[i+1] -= 1
			tempCards[i+2] -= 2
			tempCards[i+3] -= 1
			tempCards[i+4] -= 2
			tempCards[i+5] -= 1
			tempCards[i+6] -= 2
			tempCards[i+7] -= 1
			tempCards[i+8] -= 1
			tempOpt, _ := majiangcom.CanHu(tempCards, 0)
			if tempOpt&majiangcom.OptTypeHu != 0 {
				*mask |= def.YiSeSiBuGao
				*noMask |= def.YiSeSanBuGao
				return
			}
		}
	}
}
func IsYiSeSanBuGao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[4]; i++ {
		tempCards := Cards
		if tempCards[i] > 0 &&
			tempCards[i+1] >= 2 &&
			tempCards[i+2] >= 3 &&
			tempCards[i+3] >= 2 &&
			tempCards[i+4] > 0 {
			tempCards[i] -= 1
			tempCards[i+1] -= 2
			tempCards[i+2] -= 3
			tempCards[i+3] -= 2
			tempCards[i+4] -= 1
			tempOpt, _ := majiangcom.CanHu(tempCards, 0)
			if tempOpt&majiangcom.OptTypeHu != 0 {
				*mask |= def.YiSeSanBuGao
				return
			}
		}

		tempCards = Cards
		if tempCards[i] >= 1 &&
			tempCards[i+1] >= 1 &&
			tempCards[i+2] >= 2 &&
			tempCards[i+3] >= 1 &&
			tempCards[i+4] >= 2 &&
			tempCards[i+5] >= 1 &&
			tempCards[i+6] >= 1 {
			tempCards[i] -= 1
			tempCards[i+1] -= 1
			tempCards[i+2] -= 2
			tempCards[i+3] -= 1
			tempCards[i+4] -= 2
			tempCards[i+5] -= 1
			tempCards[i+6] -= 1
			tempOpt, _ := majiangcom.CanHu(tempCards, 0)
			if tempOpt&majiangcom.OptTypeHu != 0 {
				*mask |= def.YiSeSanBuGao
				return
			}
		}
	}
}

func IsSanGang(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for key, val := range Cards {
		if val == 4 && key > 0 {
			count++
		}
	}
	switch count {
	case 3:
		*mask |= def.SanGang
		*noMask |= def.ShuangMingGang | def.ShuangAnGang | def.MingGang | def.AnGang
	case 4:
		*mask |= def.SiGang
		*noMask |= def.SanGang | def.ShuangAnGang | def.ShuangMingGang | def.MingGang | def.AnGang | def.DanDiao
	}
}

func IsHunYaoJiu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	if Cards[majiangcom.Wan[0]] != 3 || Cards[majiangcom.Wan[8]] != 3 {
		return
	}

	for i := majiangcom.Wan[1]; i <= majiangcom.Wan[7]; i++ {
		if Cards[i] > 0 {
			return
		}
	}

	*mask |= def.HunYaoJiu
	*noMask |= def.PengPengHe | def.YaoJiuKe | def.QuanDaiYao
	return
}

func IsQiDui(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Wan[0]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] == 2 {
			count++
		} else if Cards[i] == 4 {
			count += 2
		}
	}
	if count < 7 {
		return
	}
	IsLianQiDui(Cards, mask, noMask)
	IsYiSeShuangLongHui(Cards, mask, noMask)
	*mask |= def.QiDui
	*noMask |= def.DanDiao | def.MengQianQing
	return
}

func IsQingYiSe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Zi[0]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] > 0 {
			return
		}
	}
	*mask |= def.QingYiSe
	*noMask |= def.HunYiSe
	if Cards[majiangcom.Wan[4]] == 0 {
		IsXiaoYuWu(Cards, mask, noMask)
		IsDaYuWu(Cards, mask, noMask)
	}
}

func IsYiSeSanTongShun(Cards [majiangcom.MaxCardValue]int, ChiPai [12]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[6]; i++ {
		count := 0
		for j := 0; j < 12; j += 3 {
			if ChiPai[j] == 0 {
				break
			}

			if ChiPai[j] == i {
				count++
			}
		}

		if Cards[i]+count >= 3 && Cards[i+1]+count >= 3 && Cards[i+2]+count >= 3 {
			temp := Cards
			temp[i] -= 3 - count
			temp[i+1] -= 3 - count
			temp[i+2] -= 3 - count
			ret, _ := majiangcom.CanHu(temp, 0)
			if ret != 0 {
				*mask |= def.YiSeSanTongShun
				*noMask |= def.YiSeSanJieGao | def.YiBanGao
				return
			}
		}
	}
}

func IsYiSeSanJieGao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[6]; i++ {
		if Cards[i] >= 3 && Cards[i+1] >= 3 && Cards[i+2] >= 3 {
			temp := Cards
			temp[i] -= 3
			temp[i+1] -= 3
			temp[i+2] -= 3
			ret, _ := majiangcom.CanHu(temp, 0)
			if ret != 0 {
				*mask |= def.YiSeSanJieGao
				*noMask |= def.YiSeSanTongShun | def.YiBanGao
				return
			}
		}
	}
}

func IsQingLong(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	tmp := Cards
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if tmp[i] == 0 {
			return
		}
		tmp[i] -= 1
	}
	tempOpt, _ := majiangcom.CanHu(tmp, 0)
	if tempOpt&majiangcom.OptTypeHu != 0 {
		*mask |= def.QingLong
		*noMask |= def.LaoShaoFu | def.Lian6
	}
	//for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
	//	if Cards[i] < 1 {
	//		return
	//	}
	//}
	//*mask |= def.QingLong
	//*noMask |= def.LaoShaoFu | def.Lian6
}

func IsSanFengKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Zi[0]; i <= majiangcom.Zi[3]; i++ {
		if Cards[i] >= 3 {
			count++
		}
	}
	if count < 3 {
		return
	}
	*mask |= def.SanFengKe
	return
}

func IsDaYuWu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[3]; i++ {
		if Cards[i] > 0 {
			return
		}
	}

	*mask |= def.DaYuWu
	return
}

func IsXiaoYuWu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[5]; i <= majiangcom.Wan[8]; i++ {
		if Cards[i] > 0 {
			return
		}
	}
	*mask |= def.XiaoYuWu
	return
}

func IsShuangAnGang(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Wan[0]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] == 4 {
			count++
		}
	}
	if count < 2 {
		return
	}
	*mask |= def.ShuangAnGang
}

func IsPengPengHe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	hasJiang := false
	for key, val := range Cards {
		if val >= 3 && key > 0 {
			count++
		} else if val == 1 {
			return
		} else if val == 2 {
			//这是将
			hasJiang = true
		}
	}

	if count >= 4 && hasJiang {
		*mask |= def.PengPengHe
		IsYiSeSiJieGao(Cards, mask, noMask)
	}
}
func IsShuangJianKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	for i := majiangcom.Zi[4]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] >= 3 {
			count++
		}
	}
	if count < 2 {
		return
	}
	*mask |= def.ShuangJianKe
	*noMask |= def.JianKe
	return
}

func IsQuanDaiYao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Bai[0]; i++ {
		if i == majiangcom.Wan[0] || i == majiangcom.Wan[8] || i >= majiangcom.Zi[0] && Cards[i] < 1 {
			return
		}
	}
	*mask |= def.QuanDaiYao
}

func IsJianKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	count := 0
	arr := []int{35, 36, 37}
	for _, val := range arr {
		if Cards[val] >= 3 {
			count++
		}
	}
	switch count {
	case 1:
		*mask |= def.JianKe
	case 2:
		*mask |= def.ShuangJianKe
		*noMask |= def.JianKe
	}
}

func IsPingHu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	jiangcount := 0
	jiang := [7]int{}
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if Cards[i] >= 2 {
			jiang[jiangcount] = i
			jiangcount++
		}
	}
	count := 0
	for i := 0; i < jiangcount; i++ {
		temp := Cards
		hu := true
		temp[jiang[i]] -= 2
		count = 0
		for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
			temp[i+1] -= temp[i]
			temp[i+2] -= temp[i]
			if temp[i+1] < 0 || temp[i+2] < 0 {
				hu = false
				break
			}

			count += temp[i]
		}

		if hu && count >= 4 {
			break
		}
	}

	if count >= 4 {
		*mask |= def.PingHu
		*noMask |= def.MingGang
		return
	}
}

func IsDuanYao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	//没有1和9
	if Cards[majiangcom.Wan[0]] > 0 || Cards[majiangcom.Wan[8]] > 0 {
		return
	}
	//没有字牌
	for i := majiangcom.Zi[0]; i <= majiangcom.Bai[0]; i++ {
		if Cards[i] > 0 {
			return
		}
	}
	*mask |= def.DuanYao
	return
}

func IsYiBanGao(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		if Cards[i] >= 2 && Cards[i+1] >= 2 && Cards[i+2] >= 2 {
			temp := Cards
			temp[i] -= 2
			temp[i+1] -= 2
			temp[i+2] -= 2
			tempOpt, _ := majiangcom.CanHu(temp, 0)
			if tempOpt&majiangcom.OptTypeHu != 0 {
				*mask |= def.YiBanGao
				break
			}
		}
	}
}

func IsLianLiu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[3]; i++ {
		if Cards[i] > 0 && Cards[i+1] > 0 && Cards[i+2] > 0 &&
			Cards[i+3] > 0 && Cards[i+4] > 0 && Cards[i+5] > 0 {
			temp := Cards
			temp[i]--
			temp[i+1]--
			temp[i+2]--
			temp[i+3]--
			temp[i+4]--
			temp[i+5]--
			ret, _ := majiangcom.CanHu(temp, 0)
			if ret != 0 {
				*mask |= def.Lian6
				return
			}
		}
	}
}

func IsLaoShaoFu(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	var arr = []int{majiangcom.Wan[0], majiangcom.Wan[1], majiangcom.Wan[2], majiangcom.Wan[6], majiangcom.Wan[7], majiangcom.Wan[8]}

	temp := Cards
	for _, val := range arr {
		if Cards[val] < 1 {
			return
		}

		temp[val]--
	}

	ret, _ := majiangcom.CanHu(temp, 0)
	if ret != 0 {
		*mask |= def.LaoShaoFu
	}
}

func IsYaoJiuKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	temp1 := [2]int{1, 9}
	for i := 0; i < 2; i++ {
		if Cards[temp1[i]] >= 3 {
			temp := Cards
			temp[temp1[i]] -= 3
			ret, _ := majiangcom.CanHu(temp, 0)
			if ret != 0 {
				*mask |= def.YaoJiuKe
				return
			}
		}
	}

	for i := majiangcom.Zi[0]; i <= majiangcom.Zi[3]; i++ {
		if Cards[i] == 3 {
			*mask |= def.YaoJiuKe
			return
		}
	}
}

func IsBianZhangOrKangZhang(Cards [majiangcom.MaxCardValue]int, HuOptCard int32, mask, noMask *int64) {
	if len(majiangcom.GetHandCards(Cards)) == 2 {
		return
	}

	cards := Cards
	huCard := HuOptCard
	if huCard < 1 || huCard >= int32(majiangcom.Zi[0]) {
		return
	}
	cards[huCard] -= 1
	count := 0
	for i := majiangcom.Wan[0]; i <= majiangcom.Wan[8]; i++ {
		tempOpt, _ := majiangcom.CanHu(cards, int32(i))
		if tempOpt&majiangcom.OptTypeHu != 0 {
			count++
		}
	}

	for i := majiangcom.Zi[0]; i <= majiangcom.Bai[0]; i++ {
		tempOpt, _ := majiangcom.CanHu(cards, int32(i))
		if tempOpt&majiangcom.OptTypeHu != 0 {
			count++
		}
	}

	//多个教不算坎也不算边
	if count >= 2 {
		return
	}

	temp := cards
	ret, _ := majiangcom.CanJiangHu(temp, huCard, huCard)
	if ret != 0 {
		return
	}

	if int(huCard)+1 > majiangcom.Wan[8] || int(huCard)-1 < majiangcom.Wan[0] {
		return
	}

	//这里看坎张
	if temp[int(huCard)+1] >= 1 && temp[int(huCard)-1] >= 1 {
		temp[int(huCard)+1] -= 1
		temp[int(huCard)-1] -= 1
		resC1, _ := majiangcom.CanHu(temp, 0)
		if resC1&majiangcom.OptTypeHu != 0 {
			*mask |= def.KanZhuang
			*noMask |= def.BianZhang
			return
		}
	}

	//如果不是坎张也不是单调将并且只有一个叫，又是3万和6万的情况必定是边张
	if int(huCard) == majiangcom.Wan[2] || int(huCard) == majiangcom.Wan[6] {
		*mask |= def.BianZhang
		*noMask |= def.KanZhuang
		return
	}
}

func IsAnKe(Cards [majiangcom.MaxCardValue]int, mask, noMask *int64) {
	var tempCard1 [4]int
	count := 0
	for key, val := range Cards {
		if val >= 3 {
			tempCard1[count] = key
			count++
		}
	}

	if count >= 3 {
		for i := 0; i < count; i++ {
			for j := i + 1; j < count; j++ {
				for m := j + 1; m < count; m++ {
					tempCard2 := Cards
					tempCard2[tempCard1[i]] -= 3
					tempCard2[tempCard1[j]] -= 3
					tempCard2[tempCard1[m]] -= 3
					tempOpt, _ := majiangcom.CanHu(tempCard2, 0)
					if tempOpt&majiangcom.OptTypeHu != 0 {
						*mask |= def.SanAnKe
						*noMask |= def.ShuangAnKe
						return
					}
				}
			}
		}
	}

	if count >= 2 {
		for i := 0; i < count; i++ {
			for j := i + 1; j < count; j++ {
				tempCard2 := Cards
				tempCard2[tempCard1[i]] -= 3
				tempCard2[tempCard1[j]] -= 3
				tempOpt, _ := majiangcom.CanHu(tempCard2, 0)
				if tempOpt&majiangcom.OptTypeHu != 0 {
					*mask |= def.ShuangAnKe
					return
				}
			}
		}
	}
}
