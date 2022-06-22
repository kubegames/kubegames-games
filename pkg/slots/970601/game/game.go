package game

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/config"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/data"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/global"
	"github.com/kubegames/kubegames-games/pkg/slots/970601/msg"
	msg2 "github.com/kubegames/kubegames-sdk/app/message"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type Game struct {
	Table table.TableInterface // table interface
	user  *data.User           //一个游戏只有一个玩家
	Room  *LhdbRoom
	//Status      int32
	lock        sync.Mutex
	Bottom      int64     //底注
	BottomCount int64     //注数
	Icons       [][]int32 //游戏的图标
	//NextIcons   [][]int32      //游戏的图标
	level           int32           //关卡等级
	TopIcons        []int32         //顶端4个图标
	WinTmpArr       [36]*Axis       // 暂存当前中奖的坐标
	CacheScore      int64           //暂存玩家中奖
	CurBoxNum       int32           //当前关卡宝箱数量
	EndGameInfo     *msg.S2CEndGame //结束比赛的返回信息
	AllWinInfoCache *msg.AllWinInfo //缓存当前的中奖信息
	IsWin           bool            //当前局是否中奖，如果中奖则递归下一次遍历
	Bottom2C        []int64         //返回给客户端的地主选项
	BottomCount2C   []int64
	TotalInvest     int64 //玩家总下注，要累计，退出才清零
	MaxTimes        int64 //最大限制倍数
	MaxDisCount     int   //限制最大中奖次数
	IsIntoCaijin    bool
	CheatConfig     *config.IconRate
	IsTest          bool
	IsShuzhi        bool //数值相关
	HoseLampArr     []*msg2.MarqueeConfig
}
type Axis struct {
	X     int
	Y     int
	Value int32
}

func NewGame(level int32, room *LhdbRoom) (game *Game) {
	game = &Game{
		Room: room, level: level, HoseLampArr: make([]*msg2.MarqueeConfig, 0),
	}
	//go game.checkTime()
	return
}

//检查用户超时没玩儿
func (game *Game) checkTime() {
	ticker := time.NewTimer(3 * time.Second)
	for {
		select {
		case <-ticker.C:
			log.Traceln("3分钟检查用户超时没玩儿")
			ticker = time.NewTimer(3 * time.Second)
		}
	}

}

//初始化游戏的二维数组宝石图标
func (game *Game) InitIcons(isChangeLevel bool) {
	if game.level <= 0 {
		game.level = 1
	}
	if game.level > 3 {
		log.Traceln("传过来的game.level怎么会大于3  ")
		return
	}
	//如果没有改变关卡，并且不是初始化，就只将所有的图标数组置为0
	if game.Icons != nil && !isChangeLevel {
		//log.Traceln("没有变换关卡，全部置0")
		for y := range game.Icons {
			for x := range game.Icons[y] {
				game.Icons[x][y] = game.GetRandIconWithCheat(x, y)
			}
		}
	} else {
		//log.Traceln("重新初始化关卡")
		switch game.level {
		case 1:
			game.Icons = make([][]int32, 4)
		case 2:
			game.Icons = make([][]int32, 5)
		case 3:
			game.Icons = make([][]int32, 6)
		}
		for i := range game.Icons {
			game.Icons[i] = make([]int32, len(game.Icons))
		}
		for y := range game.Icons {
			for x := range game.Icons[y] {
				game.Icons[x][y] = game.GetRandIconWithCheat(x, y)
			}
		}
	}

	//game.PrintIcons()
}

//初始化顶端4个图标
func (game *Game) InitTopIcons(isChangeLevel bool) {
	//min, max := global.ICON_BAIYU, global.ICON_HUPO
	//switch game.level {
	//case 2:
	//	min, max = global.ICON_ZUMULV, global.ICON_BAIZHENZHU
	//case 3:
	//	min, max = global.ICON_HONGBAOSHI, global.ICON_BAIZUANSHI
	//}
	if isChangeLevel {
		game.TopIcons = make([]int32, len(game.Icons))
	}
	for i := range game.TopIcons {
		game.TopIcons[i] = game.GetRandIconWithCheat(0, i)
		//game.TopIcons[i] = int32(rand.RandInt(min, max))
	}

	totalAxisList := make([]*Axis, 0)
	var times int64 = 0
	icons := game.CopyGameIcons()
	topIcons := game.CopyTopIcons()
	_, times = game.GetErgodicTimes(totalAxisList, times)
	game.AssignGameIcons(icons)
	game.AssignTopIcons(topIcons)
	//wins, times := game.GetIconTimesAndTmp()
	if times > game.CheatConfig.MaxTimes {
		log.Traceln("InitIcons 大于最大倍数：", times, game.CheatConfig.MaxTimes)
		for y := range game.Icons {
			for x := range game.Icons[y] {
				game.Icons[x][y] = game.GetDifferentIcon(x, y)
			}
		}
		//for _, tmp := range wins {
		//	game.Icons[tmp.X][tmp.Y] = game.GetDifferentIcon(tmp.X, tmp.Y)
		//}
	}

}

//初始化返回信息
func (game *Game) InitEndGame2C() {
	game.EndGameInfo = &msg.S2CEndGame{
		Icons:         game.iconsToOneArr(),
		TopIcons:      game.TopIcons,
		AllWinInfoArr: make([]*msg.AllWinInfo, 0),
		CurLevel:      game.level,
	}
}

//判断icons里面是否有钥匙（砖头），如果有并且返回坐标
func (game *Game) IsIconsHasKey() (bool, int, int) {
	for y := range game.Icons {
		for x := range game.Icons[y] {
			if game.Icons[x][y] == global.ICON_KEY {
				return true, x, y
			}
		}
	}
	return false, 0, 0
}

func (game *Game) SetUserList(user *data.User) {
	game.user = user
}
func (game *Game) GetUserList() *data.User {

	return game.user
}
func (game *Game) DelUserList() {
	game.user = nil
}

//获取房间基本信息
func (game *Game) GetRoomInfo2C(userSelf *data.User) *msg.S2CRoomInfo {
	info := &msg.S2CRoomInfo{
		Bottom: game.Bottom2C, Count: game.BottomCount2C,
		Level: game.level, TotalInvest: game.TotalInvest,
		IsIntoCaijin: game.IsIntoCaijin,
	}
	if game.level == 1 {
		info.FirstBoxNum = game.CurBoxNum
		info.SecondBoxNum = global.TOTAL_BOX_COUNT
		info.ThirdBoxNum = global.TOTAL_BOX_COUNT
	}
	if game.level == 2 {
		info.FirstBoxNum = global.TOTAL_BOX_COUNT
		info.SecondBoxNum = game.CurBoxNum
		info.ThirdBoxNum = global.TOTAL_BOX_COUNT
	}
	if game.level == 3 {
		info.FirstBoxNum = global.TOTAL_BOX_COUNT
		info.SecondBoxNum = global.TOTAL_BOX_COUNT
		info.ThirdBoxNum = game.CurBoxNum
	}
	return info
}

//添加 下 元素
func (game *Game) pushBelow(curX, curY int) {
	//log.Traceln("curX , curY ：",curX,curY)
	if curX < 0 || curX >= len(game.Icons) || curY < 0 || curY >= len(game.Icons) {
		return
	}
	//if game.Icons[curX][curY] == 3 {
	//	log.Traceln("pushBelow x y : ",curX,curY)
	//}

	nextX := curX + 1
	if nextX < 0 || nextX >= len(game.Icons) {
		return
	}
	if game.Icons[curX][curY] != game.Icons[nextX][curY] {
		return
	}
	//log.Traceln("x,y对应的值：",game.Icons[curX][curY])
	//将下一个元素入栈
	game.AddIntoWinTmpArr(&Axis{X: nextX, Y: curY})
	game.pushBelow(nextX, curY)
	game.pushRight(nextX, curY)
	game.pushLeft(nextX, curY)
}

//添加 上 元素
func (game *Game) pushAbove(curX, curY int) {
	//log.Traceln("下：",curX)
	if curX < 0 || curX >= len(game.Icons) || curY < 0 || curY >= len(game.Icons) {
		return
	}
	//if game.Icons[curX][curY] == 3 {
	//	log.Traceln("pushAbove x y : ",curX,curY)
	//}

	nextX := curX - 1
	if nextX < 0 || nextX >= len(game.Icons) {
		return
	}
	if game.Icons[curX][curY] != game.Icons[nextX][curY] {
		return
	}
	//将下一个元素入栈
	game.AddIntoWinTmpArr(&Axis{X: nextX, Y: curY})
	game.pushAbove(nextX, curY)
	game.pushLeft(nextX, curY)
	game.pushRight(nextX, curY)
}

//添加 左 元素
func (game *Game) pushLeft(curX, curY int) {
	if curX < 0 || curX >= len(game.Icons) || curY < 0 || curY >= len(game.Icons) {
		return
	}
	nextY := curY - 1
	if nextY < 0 || nextY >= len(game.Icons) {
		return
	}
	if game.Icons[curX][curY] != game.Icons[curX][nextY] {
		return
	}
	//将下一个元素入栈
	game.AddIntoWinTmpArr(&Axis{X: curX, Y: nextY})
	game.pushLeft(curX, nextY)
	//game.pushBelow(curX, curY)
	//game.pushAbove(curX, curY)
}

//添加 右 元素
func (game *Game) pushRight(curX, curY int) {
	if curX < 0 || curX >= len(game.Icons) || curY < 0 || curY >= len(game.Icons) {
		//game.pushLeft(curX, curY)
		return
	}
	//if game.Icons[curX][curY] == 3 {
	//	log.Traceln("pushRight x y : ",curX,curY)
	//}
	nextY := curY + 1
	if nextY < 0 || nextY >= len(game.Icons) {
		//game.pushLeft(curX, curY)
		return
	}
	if game.Icons[curX][curY] != game.Icons[curX][nextY] {
		//game.pushLeft(curX, curY)
		return
	}
	//将下一个元素入栈
	game.AddIntoWinTmpArr(&Axis{X: curX, Y: nextY})
	game.pushRight(curX, nextY)
	game.pushBelow(curX, nextY)
	game.pushAbove(curX, nextY)
}

//放入钥匙 并返回钥匙位置
func (game *Game) PutKey() (x, y int) {
	game.Icons[2][2] = global.ICON_KEY
	return 2, 2
}

//找出坐标中相邻相同的点，并存储于game.SameList当中
//x,y是坐标，函数执行完会将结果存入game.SameList中
func (game *Game) PushIntoSameList(x, y int) {
	game.AddIntoWinTmpArr(&Axis{X: x, Y: y})
	game.pushBelow(x, y)
	game.pushRight(x, y)
	game.pushLeft(x, y)
}

//遍历-找出整个二维数组中存在的相邻相同图标    serial 表示当前中奖的次数 1:第1次中奖
//效率大概 1000次 170ms
func (game *Game) Ergodic(serial int32) {
	//log.Traceln("-----------第", serial, "次图标-----------")
	//log.Traceln(game.TopIcons)
	//game.PrintIcons()
	if game.HasKeyOperate(serial) {
		game.user.KeyCount++
		log.Traceln("第", serial, "次有key：", game.user.KeyCount)
		game.MaxDisCount++
		serial++
		game.Ergodic(serial)
		return
	}
	for y := range game.Icons {
		for x := range game.Icons {
			if game.Icons[x][y] == 0 {
				continue
			}
			//将相同的元素放入sameList中
			game.PushIntoSameList(x, y)
			//判断是否中奖
			winCountVar := 4
			switch game.level {
			case 2:
				winCountVar = 5
			case 3:
				winCountVar = 6
			}
			winCount := game.GetWinTmpArrCount()
			if winCount < winCountVar {
				game.ClearWinTmpArr()
				continue
			}
			times := game.getIcocTimes(int32(winCount), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
			game.MaxTimes += times
			game.Win(serial)
			game.IsWin = true
			game.MaxDisCount++
		}
	}
	if game.IsWin {
		game.user.WinCountNew++
		//装填新的图标 在该函数里面会存 game.AllWinInfoCache 的数据
		game.fillIconsSelfErgodic()
		game.FillIconsByTop4()
		game.FillRemainIcons(true)
		game.FillTopIcons()
	}

	if game.AllWinInfoCache != nil {
		game.EndGameInfo.AllWinInfoArr = append(game.EndGameInfo.AllWinInfoArr, game.AllWinInfoCache)
	}
	//game.EndGameInfo.Icons = game.iconsToOneArr()
	//game.EndGameInfo.TopIcons = game.TopIcons
	game.EndGameInfo.IsIntoSmallGame = false
	game.AllWinInfoCache = nil
	if game.IsWin {
		game.IsWin = false
		serial++
		game.Ergodic(serial)
	}
}

func (game *Game) Win(serial int32) {
	game.updateAllWinInfoDisCache(serial)
	for _, axis := range game.WinTmpArr {
		if axis != nil {
			//将中奖的坐标位置置为0
			//log.Traceln("zhongjiang坐标：",axis)
			game.Icons[axis.X][axis.Y] = 0
		}
	}
	game.ClearWinTmpArr()
}

//更新缓存的中奖信息-消除了的图标那部分数据结构
func (game *Game) updateAllWinInfoDisCache(serial int32) {
	if game.AllWinInfoCache == nil {
		game.AllWinInfoCache = &msg.AllWinInfo{
			Serial: serial,
		}
	}
	// ----- 消除的图标和积分信息 -----
	if game.AllWinInfoCache.DisappearInfoArr == nil {
		game.AllWinInfoCache.DisappearInfoArr = make([]*msg.DisappearInfo, 0)
	}
	//消除的图标信息x
	disInfo := &msg.DisappearInfo{
		WinAxis: make([]*msg.AxisValue, 0),
	}
	for _, tmpAxis := range game.WinTmpArr {
		if tmpAxis == nil {
			continue
		}
		disInfo.Count++
		disInfo.WinAxis = append(disInfo.WinAxis, &msg.AxisValue{
			X:     int32(tmpAxis.X),
			Y:     int32(tmpAxis.Y),
			Value: game.Icons[tmpAxis.X][tmpAxis.Y],
		})
	}

	times := game.getIcocTimes(int32(game.GetWinTmpArrCount()), game.Icons[game.WinTmpArr[0].X][game.WinTmpArr[0].Y])
	game.user.Times += times
	//game.MaxTimes += times
	disInfo.WinScore += game.Bottom * game.BottomCount * times / 10
	game.EndGameInfo.WinScore += disInfo.WinScore
	//log.Traceln("disInfo.WinScore::: ",disInfo.WinScore,game.Bottom,game.BottomCount,times,game.EndGameInfo.WinScore)
	game.AllWinInfoCache.DisappearInfoArr = append(game.AllWinInfoCache.DisappearInfoArr, disInfo)
	// ----- 消除的图标和积分信息 -----

}

//获取消除的图标数量产生的倍数
func (game *Game) getIcocTimes(count, icon int32) int64 {
	//log.Traceln("count : ",count,"icon : ",icon)
	score := config.IconScoreMap[icon][count]
	//log.Traceln("score : ",score)
	return score
}

func (game *Game) AddIntoWinTmpArr(axis *Axis) {
	//先去重
	for _, v := range game.WinTmpArr {
		if v != nil && v.X == axis.X && v.Y == axis.Y {
			return
		}
	}
	for i := 0; i < len(game.WinTmpArr); i++ {
		if game.WinTmpArr[i] == nil {
			game.WinTmpArr[i] = axis
			return
		}
	}
}

func (game *Game) GetWinTmpArrCount() (count int) {
	for i := 0; i < len(game.WinTmpArr); i++ {
		if game.WinTmpArr[i] != nil {
			count++
		}
	}
	return
}

//清除list中所有元素
func (game *Game) ClearWinTmpArr() {
	for i := 0; i < len(game.WinTmpArr); i++ {
		game.WinTmpArr[i] = nil
	}
}

//将顶部的4个补到图标数组中，再补齐顶部4个 level不一样会有5个
func (game *Game) FillIconsByTop4() {
	for x := len(game.Icons) - 1; x >= 0; x-- {
		for y := len(game.Icons) - 1; y >= 0; y-- {
			if game.Icons[x][y] == 0 && game.TopIcons[y] != 0 {
				game.Icons[x][y] = game.TopIcons[y]
				if game.AllWinInfoCache.FillArr == nil {
					game.AllWinInfoCache.FillArr = make([]*msg.AxisValue, 0)
				}
				game.AllWinInfoCache.FillArr = append(game.AllWinInfoCache.FillArr, &msg.AxisValue{
					X:     int32(x),
					Y:     int32(y),
					Value: game.Icons[x][y],
				})
				game.TopIcons[y] = 0
			}
		}
	}
	//再补齐顶部的
	//for i:= range game.TopIcons {
	//	if game.TopIcons[i] == 0 {
	//		game.TopIcons[i] = game.GetRandIcon(false)
	//	}
	//}
	//log.Traceln("顶部填充之后的：")
	//game.PrintIcons()
}

//将顶部的4个补到图标数组中，再补齐顶部4个 level不一样会有5个
func (game *Game) FillIconsByTop4Ergodic() {
	for x := len(game.Icons) - 1; x >= 0; x-- {
		for y := len(game.Icons) - 1; y >= 0; y-- {
			if game.Icons[x][y] == 0 && game.TopIcons[y] != 0 {
				game.Icons[x][y] = game.TopIcons[y]
				game.TopIcons[y] = 0
			}
		}
	}
}

//补齐图标数组中剩余的，并返回坐标对应的值
type AxisValue struct {
	X     int
	Y     int
	Value int32
}

//装填顶部icons
func (game *Game) FillTopIcons() {
	//装填新的topIcons
	if game.AllWinInfoCache.TopIcons == nil {
		game.AllWinInfoCache.TopIcons = make([]int32, 0)
	}

	tmp := make([]int, 0) //暂存添加了的坐标
	for i, v := range game.TopIcons {
		if v == 0 {
			tmp = append(tmp, i)
			game.TopIcons[i] = game.GetRandIcon(false, 0, 0)
			//if game.MaxTimes+times >= game.CheatConfig.MaxTimes-1 || game.MaxDisCount >= game.CheatConfig.MaxDisCount-2 {
			//	log.Traceln("顶部图标需要获取不一样的：",game.TopIcons)
			//	game.TopIcons[i] = game.GetDifferentTopIcon(i, disListRes)
			//} else {
			//
			//}
		}
	}

	//检测是否超过最大倍数
	totalAxisList := make([]*Axis, 0)
	var times int64 = 0
	icons := game.CopyGameIcons()
	topIcons := game.CopyTopIcons()
	var disListRes = make([]*Axis, 0)
	disListRes, times = game.GetErgodicTimes(totalAxisList, times)
	game.AssignGameIcons(icons)
	game.AssignTopIcons(topIcons)
	if game.MaxTimes+times >= game.CheatConfig.MaxTimes || game.MaxDisCount >= game.CheatConfig.MaxDisCount {
		log.Traceln("FillTopIcons 超过最大倍数，原来准备fill的：", game.TopIcons)
		for _, v := range tmp {
			game.TopIcons[v] = game.GetDifferentTopIcon(v, disListRes)
		}
		log.Traceln("FillTopIcons 超过最大倍数，现在fill的：", game.TopIcons)
	}

	game.AllWinInfoCache.TopIcons = game.TopIcons
}

//装填了key返回true，没有则返回false
//todo 现在生成钻头会有概率导致超过倍数
func (game *Game) fillKey(needKey bool) (int, int, bool) {
	//return 0, 0, false

	hasKey, _, _ := game.IsIconsHasKey()
	//是否需要生成钻头
	//log.Traceln("game.CheatConfig.KeyRate ",game.CheatConfig.KeyRate,needKey,hasKey,game.CurBoxNum)
	randIndex := rand.RandInt(0, 100)
	//log.Traceln("randIndex : ",randIndex,randIndex <= game.CheatConfig.KeyRate)
	if randIndex <= game.CheatConfig.KeyRate && !hasKey && needKey && game.CurBoxNum > 0 {
		for y := range game.Icons {
			for x := range game.Icons[y] {
				if game.Icons[x][y] == 0 {
					game.Icons[x][y] = global.ICON_KEY
					//log.Traceln("生成钻头------")
					//game.PrintIcons()
					return x, y, true
					//log.Traceln(x,y,"装填：",value)
				}
			}
		}
	}
	//log.Traceln("不生成钻头------")

	return 0, 0, false
}

//isDisKey：true 表示该次填充是在填充key
//todo 现在的问题是砖头消失后，掉落下来的会有概率连上
//todo 以及传到remain的时候已经存在可以消除的
func (game *Game) FillRemainIcons(needKey bool) (resArr []*AxisValue) {
	//log.Traceln("FillRemainIcons >>> ")
	//game.PrintIcons()
	//装填消除了的积分信息
	if game.AllWinInfoCache == nil {
		game.AllWinInfoCache = new(msg.AllWinInfo)
	}
	if game.AllWinInfoCache.FillArr == nil {
		game.AllWinInfoCache.FillArr = make([]*msg.AxisValue, 0)
	}
	resArr = make([]*AxisValue, 0)

	///////--------
	tmpAxisList := make([]*Axis, 0)
	if x, y, fKey := game.fillKey(needKey); fKey {
		tmpAxisList = append(tmpAxisList, &Axis{X: x, Y: y, Value: global.ICON_KEY})
	}
	for y := range game.Icons {
		for x := range game.Icons[y] {
			if game.Icons[x][y] == 0 {
				value := game.GetRandIcon(needKey, x, y)
				game.Icons[x][y] = value
				//log.Traceln(x,y,"装填：",value)
				tmpAxisList = append(tmpAxisList, &Axis{X: x, Y: y, Value: value})
			}
		}
	}
	totalAxisList := make([]*Axis, 0)
	var times int64 = 0
	icons := game.CopyGameIcons()
	topIcons := game.CopyTopIcons()
	_, times = game.GetErgodicTimes(totalAxisList, times)
	//log.Traceln("times : ",times)
	game.AssignTopIcons(topIcons)
	game.AssignGameIcons(icons)
	//_, times := game.GetIconTimesAndTmp()
	if game.MaxTimes+times >= game.CheatConfig.MaxTimes || game.MaxDisCount >= game.CheatConfig.MaxDisCount {
		log.Traceln("game.MaxTimes > game.CheatConfig.MaxTimes ------- 进行还原", game.MaxTimes+times)
		//log.Traceln(" is IsShuzhi : ", game.IsShuzhi)
		for _, tmp := range tmpAxisList {
			var max int32 = global.ICON_HUPO
			if game.level == 2 {
				max = global.ICON_BAIZHENZHU
			}
			if game.level == 3 {
				max = global.ICON_BAIZUANSHI
			}
			game.Icons[tmp.X][tmp.Y] = game.GetDifferentIcon2(tmp.X, tmp.Y, max)
			tmp.Value = game.Icons[tmp.X][tmp.Y]
		}

		_, timesNow := game.GetIconTimesAndTmp()
		if timesNow > 0 {
			log.Traceln("timesNow > 0")
		}
		if (game.MaxTimes + times) > 15*game.CheatConfig.MaxTimes {
			log.Traceln(" > 15*game.CheatConfig.MaxTimes ..... ")
			//panic("123")
			for y := range game.Icons {
				for x := range game.Icons[y] {
					game.Icons[x][y] = game.GetDifferentIcon(x, y)
				}
			}
		}
	}
	for _, v := range tmpAxisList {
		resArr = append(resArr, &AxisValue{X: v.X, Y: v.Y, Value: v.Value})
	}
	//返回给客户端的数据
	for _, res := range resArr {
		game.AllWinInfoCache.FillArr = append(game.AllWinInfoCache.FillArr, &msg.AxisValue{
			X:     int32(res.X),
			Y:     int32(res.Y),
			Value: game.Icons[res.X][res.Y],
		})
	}
	//log.Traceln("fill remain 之后：")
	//game.PrintIcons()
	return
}

//图标出现相连的消失之后，自动依次下移 ，遍历
func (game *Game) fillIconsSelfErgodic() {
	for y := range game.Icons {
		for x := range game.Icons[y] {
			game.fillIconsSelf(x, y)
		}
	}
	//log.Traceln("自由落体之后的：")
	//game.PrintIcons()
}

//数组内图标消失，内部自己往下移进行填充 x,y 是当前为0的位置，剩下的依次往下落
//从下网上遍历，所以横坐标 --
func (game *Game) fillIconsSelf(curX, curY int) {
	if game.Icons[curX][curY] != 0 {
		return
	}
	if curX < 0 || curX >= len(game.Icons) || curY < 0 || curY >= len(game.Icons) {
		return
	}
	nextX := curX - 1
	if nextX < 0 || nextX >= len(game.Icons) {
		return
	}
	//从该位置开始，依次往上交换
	for i := curX - 1; i >= 0; i-- {
		game.Icons[i][curY], game.Icons[i+1][curY] = game.Icons[i+1][curY], game.Icons[i][curY]
	}
	if game.Icons[curX][curY] == 0 && game.Icons[nextX][curY] != 0 {
		log.Traceln("curX : ", curX)
		game.fillIconsSelf(curX, curY)
	} else {
		game.fillIconsSelf(nextX, curY)
	}

}

func (game *Game) PrintIcons() {
	for y := range game.Icons {
		log.Traceln(game.Icons[y])
	}
	log.Traceln("----------")
	//for y:= range game.Icons {
	//	for x:= range game.Icons {
	//		fmt.Print(game.GetCheatIconIndex(x, y))
	//	}
	//	log.Traceln()
	//}
}

//二维数组转一维数组
func (game *Game) iconsToOneArr() []int32 {
	icons := make([]int32, 0)
	for i := 0; i < len(game.Icons); i++ {
		icons = append(icons, game.Icons[i]...)
	}
	return icons
}

//设置图标为某个特定的值
func (game *Game) SetSpecialIcons() {
	switch game.level {
	case 1:
		game.Icons = [][]int32{
			{3, 2, 4, 2},
			{2, 3, 2, 1},
			{2, 1, 1, 2},
			{4, 4, 2, 4},
		}
	case 2:
		game.Icons = [][]int32{
			{13, 12, 14, 12, 11},
			{12, 13, 12, 11, 12},
			{12, 11, 11, 12, 13},
			{14, 14, 12, 14, 12},
			{15, 15, 12, 14, 13},
		}
	case 3:
		game.Icons = [][]int32{
			{23, 22, 24, 22, 21, 25},
			{22, 23, 22, 21, 22, 25},
			{22, 21, 21, 22, 23, 25},
			{24, 24, 22, 24, 22, 24},
			{25, 25, 22, 24, 23, 22},
			{22, 21, 21, 25, 23, 25},
		}
	}

}

//是否触发跑马灯,有特殊条件就是and，没有特殊条件满足触发金额即可
func (game *Game) TriggerHorseLamp(winAmount int64) {
	//log.Traceln("TriggerHorseLamp >>> ",winAmount,fmt.Sprintf(`%+v`,game.HoseLampArr))
	for _, v := range game.HoseLampArr {
		if strings.TrimSpace(v.SpecialCondition) == "" {
			if winAmount >= v.AmountLimit && fmt.Sprintf(`%d`, game.Table.GetRoomID()) == v.RoomId {
				log.Traceln("创建没有特殊条件的跑马灯")
				if err := game.Table.CreateMarquee(game.user.User.GetNike(), winAmount, "", v.RuleId); err != nil {
				}
				break
			} else {
				log.Traceln("不创建跑马灯：", winAmount, v.AmountLimit, game.Table.GetRoomID(), v.RoomId)
			}
		} else {
			log.Traceln("跑马灯有特殊条件 : ", strings.TrimSpace(v.SpecialCondition))
		}

	}
}
