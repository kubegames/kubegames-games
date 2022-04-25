package iconlogic

import (
	"go-game-sdk/example/game_LaBa/labacom/config"
	"math/rand"
)

const max_times = 10

type LineInfo struct {
	Index     int32
	Count     int32
	Gold      int64
	Start     int32 //线的起始方向，0左起，1右起
	WildCount int32 //wild的个数
}

type Iconinfo struct {
	Iconarr []int32
	Line    []LineInfo
	Odds    int
	Count   int
}

//计算图形
func (icon *Iconinfo) GetIcon(cheatvalue int64, lbcfg *config.LabaConfig, isfree bool, isSpecialStyle bool, oddFunc func(cheatvalue int64), iconFunc func(cheatvalue int64)) {
	cheatcfg := lbcfg.GetCheatCfg(cheatvalue)
	var iconweightcfg map[int64]config.Iconweight
	//倍数限制
	limit := cheatcfg.Limit
	//特殊玩法和免费游戏不能同时出现
	if isSpecialStyle {
		iconweightcfg = cheatcfg.SpecialStyleicon
		//出特殊玩法时倍数限制
		limit = cheatcfg.SpecialStyleLimit
	} else {
		if isfree {
			iconweightcfg = cheatcfg.Freeicon
		} else {
			iconweightcfg = cheatcfg.Normalicon
		}
	}
	hasvalue := false
	//最多计算10次
	for i := 0; i < max_times; i++ {
		icon.geticonbyweight(lbcfg, iconweightcfg)
		if iconFunc != nil {
			iconFunc(cheatvalue)
		}

		icon.Gettotalodds(lbcfg)
		if oddFunc != nil {
			oddFunc(cheatvalue)
		}

		if icon.Odds < limit {
			hasvalue = true
			break
		}
	}

	//10次再计算不出来就使用特殊的计算
	if !hasvalue {
		icon.geticonbyweight(lbcfg, lbcfg.Special)
		if iconFunc != nil {
			iconFunc(cheatvalue)
		}
		icon.Gettotalodds(lbcfg)
		if oddFunc != nil {
			oddFunc(cheatvalue)
		}
		icon.Count++
	}
}

//通过权重计算出对应的图标
func (icon *Iconinfo) geticonbyweight(lbcfg *config.LabaConfig, iconweightcfg map[int64]config.Iconweight) {
	icon.Iconarr = make([]int32, 0)
	freegameicon := int64(-1)
	//限制免费游戏图标个数
	if lbcfg.FreeGame.IconId != -1 && lbcfg.FreeGame.LimitCount == 1 {
		freegameicon = int64(lbcfg.FreeGame.IconId)
	}

	jackpoticon := int64(-1)
	//限制jackpot的个数
	if lbcfg.Jackpot.IconId != -1 && lbcfg.Jackpot.Limitcount == 1 {
		jackpoticon = int64(lbcfg.Jackpot.IconId)
	}

	wildicon := int64(-1)
	//限制wild的个数
	if lbcfg.Wild.IconId != -1 && lbcfg.Wild.LimitCount == 1 {
		wildicon = int64(lbcfg.Wild.IconId)
	}

	for i := 1; i <= lbcfg.Row; i++ {
		iconweight1 := 0
		iconweight2 := 0
		iconweight3 := 0
		for j := 0; j < lbcfg.Line; j++ {
			//先处理异形的问题，如果是-1就略过
			if lbcfg.Matrix[i-1][j] == -1 {
				icon.Iconarr = append(icon.Iconarr, -1)
				continue
			}

			w := iconweightcfg[int64(i)].TotalWeight
			r := rand.Intn(w - iconweight1 - iconweight2 - iconweight3)
			for k, v := range iconweightcfg[int64(i)].Weight {
				//遇到特殊图标，但是特殊图标又有了就跳过
				if iconweight1 != 0 && freegameicon == k {
					continue
				}

				if iconweight2 != 0 && jackpoticon == k {
					continue
				}

				if iconweight3 != 0 && wildicon == k {
					continue
				}

				if r < v {
					//将生成的图标加入到列表中去
					icon.Iconarr = append(icon.Iconarr, int32(k))
					//设置特殊图标权重
					if freegameicon == k {
						iconweight1 = v
					}

					if jackpoticon == k {
						iconweight2 = v
					}

					if wildicon == k {
						iconweight3 = v
					}

					break
				}

				r -= v
			}
		}
	}
}

//获取总的赔率
func (icon *Iconinfo) Gettotalodds(lbcfg *config.LabaConfig) int {
	icon.Line = make([]LineInfo, 0)
	icon.Odds = 0
	if lbcfg.AwardType == 4 {
		//243线特殊处理
		for i := 0; i < lbcfg.Line; i++ {
			extraodds := 1
			count := 0
			wildcount := 0
			iconid := icon.Iconarr[i]
			for j := 1; j < lbcfg.Row; j++ {
				tempcount := 0
				tmpwild := 0
				for m := 0; m < lbcfg.Line; m++ {
					index := j*lbcfg.Line + m
					if icon.Iconarr[index] == int32(lbcfg.Wild.IconId) &&
						iconid != int32(lbcfg.FreeGame.IconId) &&
						iconid != int32(lbcfg.Jackpot.IconId) {
						tempcount++
						tmpwild++
					} else if iconid == int32(lbcfg.Wild.IconId) &&
						icon.Iconarr[index] != int32(lbcfg.FreeGame.IconId) &&
						icon.Iconarr[index] != int32(lbcfg.Jackpot.IconId) {
						iconid = icon.Iconarr[index]
						tempcount++
						tmpwild++
					} else if iconid == icon.Iconarr[index] {
						tempcount++
					}
				}

				wildcount += tmpwild
				if tempcount > 0 {
					count++
					extraodds = extraodds * tempcount
				} else {
					break
				}
			}

			iconaward, _ := lbcfg.IconAward[int64(iconid)]
			if iconaward[int64(count)] != 0 {
				var li LineInfo
				li.Count = int32(count) + 1
				li.Gold = iconaward[int64(count)]
				li.Index = iconid
				li.Start = 1
				li.WildCount = int32(wildcount)
				icon.Line = append(icon.Line, li)
			}

			icon.Odds += int(iconaward[int64(count)]) * extraodds
		}
	} else {
		for index, v := range lbcfg.LineBall {
			count := 0
			wildcount := 0
			//左起连续压线记奖
			if (lbcfg.AwardType & 1) != 0 {
				iconid := icon.Iconarr[v[0]]
				for i := 1; i < len(v); i++ {
					if icon.Iconarr[v[i]] == int32(lbcfg.Wild.IconId) &&
						iconid != int32(lbcfg.FreeGame.IconId) &&
						iconid != int32(lbcfg.Jackpot.IconId) {
						count++
						wildcount++
					} else if iconid == int32(lbcfg.Wild.IconId) &&
						icon.Iconarr[v[i]] != int32(lbcfg.FreeGame.IconId) &&
						icon.Iconarr[v[i]] != int32(lbcfg.Jackpot.IconId) {
						iconid = icon.Iconarr[v[i]]
						count++
						wildcount++
					} else if iconid == icon.Iconarr[v[i]] {
						count++
					} else {
						break
					}
				}

				iconaward, _ := lbcfg.IconAward[int64(iconid)]

				icon.Odds += int(iconaward[int64(count)])
				if iconaward[int64(count)] != 0 {
					var li LineInfo
					li.Count = int32(count) + 1
					li.Gold = iconaward[int64(count)]
					li.Index = int32(index)
					li.Start = 1
					li.WildCount = int32(wildcount)
					icon.Line = append(icon.Line, li)
				}
			}

			//右起连续压线记奖
			if (lbcfg.AwardType&0x2) != 0 && count < lbcfg.Line {
				count = 0
				iconid := icon.Iconarr[v[len(v)-1]]

				for i := len(v) - 2; i >= 0; i-- {
					if icon.Iconarr[v[i]] == int32(lbcfg.Wild.IconId) &&
						icon.Iconarr[v[i]] != int32(lbcfg.FreeGame.IconId) &&
						icon.Iconarr[v[i]] != int32(lbcfg.Jackpot.IconId) {
						wildcount++
						count++
					} else if iconid == int32(lbcfg.Wild.IconId) &&
						iconid != int32(lbcfg.FreeGame.IconId) &&
						iconid != int32(lbcfg.Jackpot.IconId) {
						iconid = icon.Iconarr[v[i]]
						wildcount++
						count++
					} else if iconid == icon.Iconarr[v[i]] {
						count++
					} else {
						break
					}
				}

				iconaward, _ := lbcfg.IconAward[int64(iconid)]
				icon.Odds += int(iconaward[int64(count)])
				if iconaward[int64(count)] != 0 {
					var li LineInfo
					li.Count = int32(count) + 1
					li.Gold = iconaward[int64(count)]
					li.Index = int32(index)
					li.Start = 2
					li.WildCount = int32(wildcount)
					icon.Line = append(icon.Line, li)
				}
			}
		}
	}

	return icon.Odds
}

func (icon *Iconinfo) Getfreegametimes(lbcfg *config.LabaConfig) []int {
	var count []int
	//左起压线上记奖，不需要连续
	if lbcfg.FreeGame.Type == 1 {
		for _, arr := range lbcfg.LineBall {
			tmp := 0
			for i := 0; i < len(arr); i++ {
				if icon.Iconarr[arr[i]] == int32(lbcfg.FreeGame.IconId) {
					tmp++
				}
			}

			if tmp > 0 {
				tmp--
				if tmp >= len(lbcfg.FreeGame.Times) {
					tmp = len(lbcfg.FreeGame.Times) - 1
				}
				count = append(count, tmp)
			}
		}
	} else if lbcfg.FreeGame.Type == 2 {
		//连续列
		//hasfree := false
		tmp := 0
		for _, arr := range lbcfg.Matrix {
			for i := 0; i < len(arr); i++ {
				if icon.Iconarr[arr[i]] == int32(lbcfg.FreeGame.IconId) {
					//hasfree = true
					tmp++
					break
				}
			}
		}

		if tmp > 0 {
			tmp--
			if tmp >= len(lbcfg.FreeGame.Times) {
				tmp = len(lbcfg.FreeGame.Times) - 1
			}
			count = append(count, tmp)
		}
	} else if lbcfg.FreeGame.Type == 3 {
		//分散计算个数
		tmp := 0
		for _, v := range icon.Iconarr {
			if v == int32(lbcfg.FreeGame.IconId) {
				tmp++
			}
		}

		if tmp > 0 {
			tmp--
			if tmp >= len(lbcfg.FreeGame.Times) {
				tmp = len(lbcfg.FreeGame.Times) - 1
			}
			count = append(count, tmp)
		}
	} else if lbcfg.FreeGame.Type == 5 {
		for _, arr := range lbcfg.LineBall {
			tmp := 0
			if icon.Iconarr[arr[0]] == int32(lbcfg.FreeGame.IconId) {
				for i := 0; i < len(arr); i++ {
					if icon.Iconarr[arr[i]] == int32(lbcfg.FreeGame.IconId) {
						tmp++
					} else {
						break
					}
				}
			}

			if tmp > 0 {
				tmp--
				if tmp >= len(lbcfg.FreeGame.Times) {
					tmp = len(lbcfg.FreeGame.Times) - 1
				}
				count = append(count, tmp)
			}
		}
	}

	return count
}

func (icon *Iconinfo) Getjackpot(lbcfg *config.LabaConfig) int {
	count := 0
	//左起连线上记奖
	if lbcfg.Jackpot.Type == 1 {
		for _, v := range lbcfg.Matrix[0] {
			if v == -1 {
				continue
			}
			if icon.Iconarr[v] != int32(lbcfg.Jackpot.IconId) {
				continue
			}
			for _, arr := range lbcfg.LineBall {
				if arr[0] != v {
					continue
				}

				tmp := 0
				for i := 1; i < len(arr); i++ {
					if icon.Iconarr[arr[i]] == int32(lbcfg.Jackpot.IconId) {
						tmp++
					} else {
						break
					}
				}

				if count < tmp {
					count = tmp
				}
			}
		}
	} else if lbcfg.Jackpot.Type == 2 {
		//连续列
		hasfree := false
		for _, arr := range lbcfg.Matrix {
			for i := 0; i < len(arr); i++ {
				if arr[i] == int64(lbcfg.Jackpot.IconId) {
					hasfree = true
					count++
					break
				}
			}

			if !hasfree {
				count--
				break
			}
		}
	} else if lbcfg.Jackpot.Type == 3 {
		//分散计算个数
		for _, v := range icon.Iconarr {
			if v == int32(lbcfg.Jackpot.IconId) {
				count++
			}
		}

		count--
	}

	if count < 0 {
		count = 0
	} else if count >= len(lbcfg.Jackpot.Award) {
		count = len(lbcfg.Jackpot.Award) - 1
	}

	return int(lbcfg.Jackpot.Award[count])
}

//获取连线上连续图片个数，普通图片不包括wild
func (icon *Iconinfo) GetOnLineIconCount(iconid int, lbcfg *config.LabaConfig) []int {
	var ret []int
	for _, arr := range lbcfg.LineBall {
		tmp := 0
		for _, v := range arr {
			if icon.Iconarr[v] == int32(iconid) {
				tmp++
			} else if tmp > 0 {
				ret = append(ret, tmp)
				tmp = 0
			}
		}

		if tmp > 0 {
			ret = append(ret, tmp)
		}
	}

	return ret
}

//获取连线上分散图片个数，普通图片不包括wild
func (icon *Iconinfo) GetOnLineScatterIconCount(iconid int, lbcfg *config.LabaConfig) []int {
	var ret []int
	for _, arr := range lbcfg.LineBall {
		tmp := 0
		for _, v := range arr {
			if icon.Iconarr[v] == int32(iconid) {
				tmp++
			}
		}

		if tmp > 0 {
			ret = append(ret, tmp)
		}
	}

	return ret
}

//243线算连续的列
func (icon *Iconinfo) GetLineIconCount(iconid int, lbcfg *config.LabaConfig) int {
	var count int
	for _, arr := range lbcfg.Matrix {
		tmp := count
		for i := 0; i < len(arr); i++ {
			if icon.Iconarr[arr[i]] == int32(iconid) {
				count++
				break
			}
		}

		if tmp == count {
			break
		}
	}

	return count
}
