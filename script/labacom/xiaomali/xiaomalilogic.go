package xiaomali

import (
	"math/rand"
)

//获取外圈的图形
func (xml *XiaoMaLiCfg) GetIconInfo(iconweight Iconweight, cfg CheatCfg) (int, int, []int) {
	iconid := geticonid(iconweight)

	extricon := iconid
	bTwo := false
	if iconid == -1 {
		//无奖励的情况
		r := rand.Intn(len(iconweight.Icon))
		iconid = iconweight.Icon[r]
	} else {
		r := rand.Intn(10000)
		tmp, _ := iconweight.Weight[int64(iconid)]

		if r < int(tmp[1]) {
			bTwo = true
		}
	}

	iconarr := xml.GetInIconIndex(extricon, iconid, bTwo, iconweight, cfg)
	iconranklen := len(xml.Iconrank)
	r := rand.Intn(iconranklen)
	var ir []int
	//将原来的数据重新连接一次获取到第一个相同的index
	ir = append(xml.Iconrank[r:], xml.Iconrank[:r]...)
	index := 0
	for i := 0; i < iconranklen; i++ {
		if ir[i] == iconid {
			index = (i + r) % iconranklen
			break
		}
	}

	return iconid, index, iconarr
}

//获取内圈
func (xml *XiaoMaLiCfg) GetInIconIndex(extricon int, icond int, bTwo bool, iconweight Iconweight, cfg CheatCfg) []int {
	var iconarr []int
	for {
		for i := 0; i < 4; i++ {
			r := rand.Intn(len(iconweight.Icon) - 1)
			tmp := iconweight.Icon[r]
			for {
				if tmp != icond {
					break
				}
				r = rand.Intn(len(iconweight.Icon) - 1)
				tmp = iconweight.Icon[r]
			}

			iconarr = append(iconarr, tmp)
		}

		if extricon != -1 && extricon != iconweight.Icon[len(iconweight.Icon)-1] {
			r := rand.Intn(4)
			iconarr[r] = extricon
			if bTwo {
				for {
					tmp := rand.Intn(4)
					if tmp != r {
						iconarr[tmp] = extricon
						break
					}
				}
			}
		}

		//强制不出500倍
		if iconarr[0] == iconarr[1] && iconarr[1] == iconarr[2] && iconarr[2] == iconarr[3] {
			iconarr = make([]int, 0)
			continue
		}

		//三连只在-3000的情况下出现，其他情况不出
		if cfg.Three == 1 {
			if (iconarr[0] == iconarr[1] && iconarr[1] == iconarr[2]) ||
				(iconarr[1] == iconarr[2] && iconarr[2] == iconarr[3]) {
				iconarr = make([]int, 0)
				continue
			}
		}

		break
	}

	return iconarr
}

//随机一个结果出去
func geticonid(iconweight Iconweight) int {
	r := rand.Int63n(iconweight.TotalWeight)
	for k, v := range iconweight.Weight {
		if r < v[0] {
			return int(k)
		}

		r -= v[0]
	}

	return 0
}

func GetOdds(outicon int64, inicon []int, iconAward map[int64]int) int {
	odds := 0
	if (inicon[0] == inicon[1] && inicon[0] == inicon[2]) ||
		(inicon[1] == inicon[2] && inicon[3] == inicon[2]) {
		odds += 20
	}

	award, ok := iconAward[outicon]
	if ok {
		for _, v := range inicon {
			if v == int(outicon) {
				odds += award
			}
		}
	}

	return odds
}
