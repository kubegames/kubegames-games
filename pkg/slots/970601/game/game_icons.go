package game

import (
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/global"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/msg"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

//图标中有钥匙（砖头）的操作
func (game *Game) HasKeyOperate(serial int32) (hasKey bool) {
	hasKey, x, y := game.IsIconsHasKey()
	if !hasKey {
		return
	}
	game.Icons[x][y] = 0
	if game.EndGameInfo == nil {
		game.EndGameInfo = &msg.S2CEndGame{
			AllWinInfoArr: make([]*msg.AllWinInfo, 0),
		}
	}
	if game.AllWinInfoCache == nil {
		game.AllWinInfoCache = &msg.AllWinInfo{
			Serial:           serial,
			DisappearInfoArr: make([]*msg.DisappearInfo, 1),
		}
	}
	game.AllWinInfoCache.DisappearInfoArr[0] = &msg.DisappearInfo{
		Count:    1,
		WinScore: 0,
		WinAxis:  make([]*msg.AxisValue, 1),
	}
	game.AllWinInfoCache.DisappearInfoArr[0].WinAxis[0] = &msg.AxisValue{
		X:     int32(x),
		Y:     int32(y),
		Value: global.ICON_KEY,
	}

	game.fillIconsSelfErgodic()
	game.FillIconsByTop4()

	game.FillRemainIcons(false)
	game.FillTopIcons()
	//allInfo := &msg.AllWinInfo{
	//	DisappearInfoArr:     make([]*msg.DisappearInfo,1),
	//	//FillArr:              make([]*msg.AxisValue,0),
	//	TopIcons:             game.TopIcons,
	//}
	game.EndGameInfo.AllWinInfoArr = append(game.EndGameInfo.AllWinInfoArr, game.AllWinInfoCache)
	game.AllWinInfoCache = nil
	game.CurBoxNum--
	if game.CurBoxNum < 0 {
		game.CurBoxNum = 0
	}
	return
}

//根据坐标获取对应的作弊的位置   1  2  3  4  =>y
//						    8 7  6  5
//							9 10 11 12
//							16 15 14 13
//							|
//							x
func (game *Game) GetCheatIconIndex(x, y int) int {
	if game.level == 1 {
		switch y {
		case 0: //第一行
			return x + 1 // 1 2 3 4
		case 1: //第二行
			return 8 - x
		case 2: //第三行
			return 9 + x
		case 3: //第四行
			return 16 - x
		}
	} else if game.level == 2 {
		//根据坐标获取对应的作弊的位置   1   2  3  4   5 =>y
		//						    10  9  8  7  6
		//							11 12 13 14  15
		//							20 19 18 17 16
		//							21 22 23 24 25
		//							|
		//							x
		switch y {
		case 0: //第一行
			return x + 1 // 1 2 3 4
		case 1: //第二行
			return 10 - x
		case 2: //第三行
			return 11 + x
		case 3: //第四行
			return 20 - x
		case 4: //第五行
			return 21 + x
		}
	} else {
		//1		2	3	4	5	6
		//12	11	10	9	8	7
		//13	14	15	16	17	18
		//24	23	22	21	20	19
		//25	26	27	28	29	30
		//36	35	34	33	32	31

		switch y {
		case 0: //第一行
			return x + 1 // 1 2 3 4
		case 1: //第二行
			return 12 - x
		case 2: //第三行
			return 13 + x
		case 3: //第四行
			return 24 - x
		case 4: //第五行
			return 25 + x
		case 5:
			return 36 - x
		}
	}
	return 0
}

//
////生成一个与上下左右都不同的icon
func (game *Game) GetDifferentIcon(x, y int) int32 {
	var start int32 = global.ICON_BAIYU
	var end int32 = global.ICON_HUPO
	lenth := 4
	if game.level == 2 {
		start = global.ICON_ZUMULV
		end = global.ICON_BAIZHENZHU
		lenth = 5
	}
	if game.level == 3 {
		start = global.ICON_HONGBAOSHI
		end = global.ICON_BAIZUANSHI
		lenth = 6
	}
	//alternative := make([]int32,0)
	for i := start; i <= end; i++ {
		//上下左右中只要有一个满足相同就继续
		//上
		if x-1 >= 0 {
			if game.Icons[x-1][y] == i {
				continue
			}
		}
		//下
		if x+1 < lenth {
			if game.Icons[x+1][y] == i {
				continue
			}
		}
		//左
		if y-1 >= 0 {
			if game.Icons[x][y-1] == i {
				continue
			}
		}
		//右
		if y+1 < lenth {
			if game.Icons[x][y+1] == i {
				continue
			}
		}
		return i
		//alternative = append(alternative,i)
	}
	//index := rand.RandInt(0,len(alternative)-1)
	//return alternative[index]
	return start
}

////生成一个与上下左右都不同的icon
//并且不能与给定的相同
func (game *Game) GetDifferentIcon2(x, y int, value int32) int32 {
	var start int32 = global.ICON_BAIYU
	var end int32 = global.ICON_HUPO
	lenth := 4
	if game.level == 2 {
		start = global.ICON_ZUMULV
		end = global.ICON_BAIZHENZHU
		lenth = 5
	}
	if game.level == 3 {
		start = global.ICON_HONGBAOSHI
		end = global.ICON_BAIZUANSHI
		lenth = 6
	}
	//alternative := make([]int32,0)
	for i := start; i <= end; i++ {
		//上下左右中只要有一个满足相同就继续
		//上
		if x-1 >= 0 {
			if game.Icons[x-1][y] == i {
				continue
			}
		}
		//下
		if x+1 < lenth {
			if game.Icons[x+1][y] == i {
				continue
			}
		}
		//左
		if y-1 >= 0 {
			if game.Icons[x][y-1] == i {
				continue
			}
		}
		//右
		if y+1 < lenth {
			if game.Icons[x][y+1] == i {
				continue
			}
		}
		if i == value {
			continue
		}
		return i
		//alternative = append(alternative,i)
	}
	//index := rand.RandInt(0,len(alternative)-1)
	//return alternative[index]
	log.Traceln("GetDifferentIcon2 return start ::: ", start)
	return start
}

//当前图标的倍数
//该函数返回中奖的倍数和中奖的坐标信息等
func (game *Game) GetIconTimesAndTmp() (axisList []*Axis, times int64) {
	//if hasKey,_,_ := game.IsIconsHasKey();!hasKey {
	//	return
	//}
	axisList = make([]*Axis, 0)
	for y := range game.Icons {
		for x := range game.Icons {
			if game.Icons[x][y] == 0 {
				continue
			}
			game.PushIntoSameList(x, y)
			//判断是否中奖
			winCountVar := 4
			switch game.level {
			case 2:
				winCountVar = 5
			case 3:
				winCountVar = 6
			}
			//log.Traceln("game.PushIntoSameList(x, y) ::: ",game.GetWinTmpArrCount(), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
			if game.GetWinTmpArrCount() >= winCountVar {
				times += game.getIcocTimes(int32(game.GetWinTmpArrCount()), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])

				log.Traceln("GetIconTimesAndTmp ====== ", times, "int32(game.GetWinTmpArrCount()):", int32(game.GetWinTmpArrCount()), "game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y] ", game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
				for _, tmp := range game.WinTmpArr {
					if tmp != nil {
						tmp.Value = game.Icons[tmp.X][tmp.Y]
						axisList = append(axisList, tmp)
						game.Icons[tmp.X][tmp.Y] = 0
					}
				}
			}
			game.ClearWinTmpArr()
		}
	}
	for _, v := range axisList {
		game.Icons[v.X][v.Y] = v.Value
	}
	//log.Traceln("GetIconTimes ::: ",times)
	return

}

//递归遍历所有icons中所有消除后然后自由下落的倍数
//注意：该函数会将icons变为自动下落之后的值，所以需要调用之前记录game.icons，然后调完函数还原
func (game *Game) GetErgodicTimes(totalAxisList []*Axis, times int64) (totalAxisListRes []*Axis, totalTimes int64) {
	tmpAxisList := make([]*Axis, 0)
	for y := range game.Icons {
		for x := range game.Icons {
			if game.Icons[x][y] == 0 {
				continue
			}
			game.PushIntoSameList(x, y)
			//判断是否中奖
			winCountVar := 4
			switch game.level {
			case 2:
				winCountVar = 5
			case 3:
				winCountVar = 6
			}
			//log.Traceln("game.PushIntoSameList(x, y) ::: ",game.GetWinTmpArrCount(), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
			if game.GetWinTmpArrCount() >= winCountVar {
				times += game.getIcocTimes(int32(game.GetWinTmpArrCount()), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
				//log.Traceln("times ====== ",times)
				for _, tmp := range game.WinTmpArr {
					if tmp != nil {
						tmp.Value = game.Icons[tmp.X][tmp.Y]
						tmpAxisList = append(tmpAxisList, tmp)
						totalAxisList = append(totalAxisList, tmp)
						game.Icons[tmp.X][tmp.Y] = 0
					}
				}
			}
			game.ClearWinTmpArr()
		}
	}
	if len(tmpAxisList) == 0 {
		totalAxisListRes = totalAxisList
		totalTimes = times
		//log.Traceln("return出去：",totalAxisListRes,"total : ",totalAxisList)
		return
	} else {
		game.fillIconsSelfErgodic()
		if len(game.TopIcons) != 0 && len(game.TopIcons) == len(game.Icons) {
			game.FillIconsByTop4Ergodic()
		}
		return game.GetErgodicTimes(totalAxisList, times)
	}
	//for _,v := range axisList{
	//	game.Icons[v.X][v.Y] = v.Value
	//}
	//log.Traceln("GetIconTimes ::: ",times)

}

////生成一个与上下左右都不同的top icon , x始终为0
//disListRes 是本次game.icons 还要消除的图标
func (game *Game) GetDifferentTopIcon(y int, disListRes []*Axis) int32 {
	//log.Traceln("需要补充的顶部y：",y)
	//for _,v := range disListRes {
	//	log.Traceln("消的坐标：",v)
	//}
	//game.PrintIcons()
	var start int32 = global.ICON_BAIYU
	var end int32 = global.ICON_HUPO
	//lenth := 4
	if game.level == 2 {
		start = global.ICON_ZUMULV
		end = global.ICON_BAIZHENZHU
		//lenth = 5
	}
	if game.level == 3 {
		start = global.ICON_HONGBAOSHI
		end = global.ICON_BAIZUANSHI
		//lenth = 6
	}
	//log.Traceln("超过了，补充新图标")
	//既不能等于相邻的，又不能等于下面相同的
	if len(disListRes) != 0 {
		var maxX *Axis
		for _, v := range disListRes {
			if v.Y == y {
				maxX = v
				break
			}
		}
		if maxX == nil {
			log.Traceln("---------maxX为空，不管")
		} else {
			var value int32 = 0
			if y < len(game.TopIcons)-1 {
				log.Traceln("GetDifferentTopIcon >>>> top icons : ", game.TopIcons)
				value = game.TopIcons[y+1]
			}
			return game.GetDifferentIcon2(maxX.X, maxX.Y, value)
		}
	}

	//if len(disListRes) != 0 {
	//	maxX := disListRes[0]
	//	for _,v := range disListRes {
	//		if v.X > maxX.X && v.Y == maxX.Y {
	//			maxX = v
	//		}
	//	}
	//	//log.Traceln("max ::: ",maxX)
	//	dif := game.GetDifferentIcon(maxX.X,maxX.Y)
	//	//log.Traceln("get dif ::: ",dif)
	//	return dif
	//	//log.Traceln("-------补充顶部图标时，下面的还能继续消除，并且没有key")
	//	//var maxX *Axis
	//	//for _,v := range disListRes {
	//	//	if v.Y == y {
	//	//		maxX = v
	//	//		break
	//	//	}
	//	//}
	//	//if maxX == nil {
	//	//	log.Traceln("---------maxX为空，不管")
	//	//}else {
	//	//	for _,v := range disListRes {
	//	//		if v.X > maxX.X && v.Y == maxX.Y {
	//	//			maxX = v
	//	//		}
	//	//	}
	//	//	//log.Traceln("max ::: ",maxX)
	//		dif := game.GetDifferentIcon(maxX.X,maxX.Y)
	//	//	//log.Traceln("get dif ::: ",dif)
	//	//	return dif
	//	//}
	//}else {
	//	//log.Traceln("===补充顶部图标时,按原来的逻辑处理")
	//}
	if y == 0 {
		//alter = append(alter,start)
		for i := start; i <= end; i++ {
			if i != game.Icons[0][0] {
				return i
			}
		}
		return start + 1
	}
	for i := start; i <= end; i++ {
		if i != game.TopIcons[y-1] && i != game.Icons[y][0] {
			return i
		}
	}

	log.Warnf("没有相同的 top icon ")
	return start
}

//随机获取一个图标
//needKey 是否需要随机填充key，top的时候不需要
func (game *Game) GetRandIcon(needKey bool, x, y int) int32 {
	//hasKey, _, _ := game.IsIconsHasKey()
	//是否需要生成钻头
	//log.Traceln("game.CheatConfig.KeyRate ",game.CheatConfig.KeyRate,needKey,hasKey,game.CurBoxNum)
	//if rand.RateToExec(game.CheatConfig.KeyRate) && !hasKey && needKey && game.CurBoxNum > 0 {
	//	return global.ICON_KEY
	//}
	return game.GetRandIconWithCheat(x, y)
	//if !needKey {
	//min, max := global.ICON_BAIYU, global.ICON_HUPO+1
	//switch game.level {
	//case 2:
	//	min, max = global.ICON_ZUMULV, global.ICON_BAIZHENZHU+1
	//case 3:
	//	min, max = global.ICON_HONGBAOSHI, global.ICON_BAIZUANSHI+1
	//}
	//return game.GetRandIconWithCheat(x, y)
	//return int32(rand.RandInt(min, max))
	//} else {
	//	return game.GetRandIconWithCheat(x, y)
	//}

}

//2月27号添加 根据作弊率获取icon
func (game *Game) GetRandIconWithCheat(x, y int) int32 {

	index := game.GetCheatIconIndex(x, y)
	cheatConfig := config.GetLhdbConfig(game.user.Cheat)
	//log.Traceln("x, y,index : ",x, y,index)
	if index%2 == 1 {
		//奇数选第1个方案
		switch game.level {
		case 1:
			total := cheatConfig.BaiyuRateA + cheatConfig.BiyuRateA + cheatConfig.MoyuRateA + cheatConfig.ManaoRateA + cheatConfig.HupoRateA
			index := rand.RandInt(0, total)
			//log.Traceln("奇数选第1个方案随机值：",index,cheatConfig.BaiyuRateA,cheatConfig.BiyuRateA)
			if index < cheatConfig.BaiyuRateA {
				return global.ICON_BAIYU
			}
			if index < cheatConfig.BaiyuRateA+cheatConfig.BiyuRateA {
				return global.ICON_BIYU
			}
			if index < cheatConfig.BaiyuRateA+cheatConfig.BiyuRateA+cheatConfig.MoyuRateA {
				return global.ICON_MOYU
			}
			if index < cheatConfig.BaiyuRateA+cheatConfig.BiyuRateA+cheatConfig.MoyuRateA+cheatConfig.ManaoRateA {
				return global.ICON_MANAO
			}
			return global.ICON_HUPO
		case 2:
			//第二关
			total := cheatConfig.ZumulvRateA + cheatConfig.MaoyanshiRateA + cheatConfig.ZishuijingRateA + cheatConfig.FeicuishiRateA + cheatConfig.BaizhenzhuRateA
			index := rand.RandInt(0, total)
			if index < cheatConfig.ZumulvRateA {
				return global.ICON_ZUMULV
			}
			if index < cheatConfig.ZumulvRateA+cheatConfig.MaoyanshiRateA {
				return global.ICON_MAOYANSHI
			}
			if index < cheatConfig.ZumulvRateA+cheatConfig.MaoyanshiRateA+cheatConfig.ZishuijingRateA {
				return global.ICON_ZISHUIJING
			}
			if index < cheatConfig.ZumulvRateA+cheatConfig.MaoyanshiRateA+cheatConfig.ZishuijingRateA+cheatConfig.FeicuishiRateA {
				return global.ICON_FEICUISHI
			}
			return global.ICON_BAIZHENZHU
		}
		//第三关
		total := cheatConfig.HongbaoshiRateA + cheatConfig.HuangbaoshiRateA + cheatConfig.LvbaoshiRateA + cheatConfig.LanbaoshiRateA + cheatConfig.BaizuanshiRateA
		index := rand.RandInt(0, total)
		if index < cheatConfig.HongbaoshiRateA {
			return global.ICON_HONGBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateA+cheatConfig.HuangbaoshiRateA {
			return global.ICON_HUANGBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateA+cheatConfig.HuangbaoshiRateA+cheatConfig.LvbaoshiRateA {
			return global.ICON_LVBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateA+cheatConfig.HuangbaoshiRateA+cheatConfig.LvbaoshiRateA+cheatConfig.LanbaoshiRateA {
			return global.ICON_LANBAOSHI
		}
		return global.ICON_BAIZUANSHI
	} else {
		//偶数选第2个方案
		switch game.level {
		case 1:
			total := cheatConfig.BaiyuRateB + cheatConfig.BiyuRateB + cheatConfig.MoyuRateB + cheatConfig.ManaoRateB + cheatConfig.HupoRateB
			index := rand.RandInt(0, total)
			//log.Traceln("偶数选第1个方案随机值：",index,cheatConfig.BaiyuRateB,cheatConfig.BiyuRateB)
			if index < cheatConfig.BaiyuRateB {
				return global.ICON_BAIYU
			}
			if index < cheatConfig.BaiyuRateB+cheatConfig.BiyuRateB {
				return global.ICON_BIYU
			}
			if index < cheatConfig.BaiyuRateB+cheatConfig.BiyuRateB+cheatConfig.MoyuRateB {
				return global.ICON_MOYU
			}
			if index < cheatConfig.BaiyuRateB+cheatConfig.BiyuRateB+cheatConfig.MoyuRateB+cheatConfig.ManaoRateB {
				return global.ICON_MANAO
			}
			return global.ICON_HUPO
		case 2:
			//第二关
			total := cheatConfig.ZumulvRateB + cheatConfig.MaoyanshiRateB + cheatConfig.ZishuijingRateB + cheatConfig.FeicuishiRateB + cheatConfig.BaizhenzhuRateB
			index := rand.RandInt(0, total)
			if index < cheatConfig.ZumulvRateB {
				return global.ICON_ZUMULV
			}
			if index < cheatConfig.ZumulvRateB+cheatConfig.MaoyanshiRateB {
				return global.ICON_MAOYANSHI
			}
			if index < cheatConfig.ZumulvRateB+cheatConfig.MaoyanshiRateB+cheatConfig.ZishuijingRateB {
				return global.ICON_ZISHUIJING
			}
			if index < cheatConfig.ZumulvRateB+cheatConfig.MaoyanshiRateB+cheatConfig.ZishuijingRateB+cheatConfig.FeicuishiRateB {
				return global.ICON_FEICUISHI
			}
			return global.ICON_BAIZHENZHU
		}
		//第三关
		total := cheatConfig.HongbaoshiRateB + cheatConfig.HuangbaoshiRateB + cheatConfig.LvbaoshiRateB + cheatConfig.LanbaoshiRateB + cheatConfig.BaizuanshiRateB
		index := rand.RandInt(0, total)
		if index < cheatConfig.HongbaoshiRateB {
			return global.ICON_HONGBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateB+cheatConfig.HuangbaoshiRateB {
			return global.ICON_HUANGBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateB+cheatConfig.HuangbaoshiRateB+cheatConfig.LvbaoshiRateB {
			return global.ICON_LVBAOSHI
		}
		if index < cheatConfig.HongbaoshiRateB+cheatConfig.HuangbaoshiRateB+cheatConfig.LvbaoshiRateB+cheatConfig.LanbaoshiRateB {
			return global.ICON_LANBAOSHI
		}
		return global.ICON_BAIZUANSHI
	}

}

//复制一个icons出来
func (game *Game) CopyGameIcons() (icons [][]int32) {
	icons = make([][]int32, len(game.Icons))
	for y := range game.Icons {
		icons[y] = make([]int32, len(game.Icons))
	}
	for y := range game.Icons {
		for x := range game.Icons {
			icons[x][y] = game.Icons[x][y]
		}
	}
	return
}

func (game *Game) CopyTopIcons() (topIcons []int32) {
	topIcons = make([]int32, len(game.TopIcons))
	for i := range game.TopIcons {
		topIcons[i] = game.TopIcons[i]
	}
	return
}

func (game *Game) AssignTopIcons(topIcons []int32) {
	for i := range topIcons {
		game.TopIcons[i] = topIcons[i]
	}
}

//将icons赋值给game.icons
func (game *Game) AssignGameIcons(icons [][]int32) {
	for y := range game.Icons {
		for x := range game.Icons {
			game.Icons[x][y] = icons[x][y]
		}
	}
}

func (game *Game) getIconName(icon int32) string {
	switch icon {
	case global.ICON_BAIYU:
		return "白玉"
	case global.ICON_BIYU:
		return "碧玉"
	case global.ICON_MOYU:
		return "墨玉"
	case global.ICON_MANAO:
		return "玛瑙"
	case global.ICON_HUPO:
		return "琥珀"
	case global.ICON_ZUMULV:
		return "祖母绿"
	case global.ICON_MAOYANSHI:
		return "猫眼石"
	case global.ICON_ZISHUIJING:
		return "紫水晶"
	case global.ICON_FEICUISHI:
		return "翡翠石"
	case global.ICON_BAIZHENZHU:
		return "白珍珠"
	case global.ICON_HONGBAOSHI:
		return "红宝石"
	case global.ICON_LVBAOSHI:
		return "绿宝石"
	case global.ICON_HUANGBAOSHI:
		return "黄宝石"
	case global.ICON_LANBAOSHI:
		return "蓝宝石"
	case global.ICON_BAIZUANSHI:
		return "白钻石"
	case global.ICON_KEY:
		return "钻头"
	}
	return "错误图标"
}
