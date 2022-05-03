package game

import (
	"go-game-sdk/example/game_LaBa/970601/config"
	"go-game-sdk/example/game_LaBa/970601/data"
	"go-game-sdk/example/game_LaBa/970601/msg"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
)

func TestNewGame1(t *testing.T) {
	t1 := time.Now()
	for i := 0; i < 1; i++ {
		TestNewGame(t)
	}
	log.Traceln(time.Now().Sub(t1))
}

func TestNewGame(t *testing.T) {
	config.LoadLhdbConfig()

	game := &Game{
		level: 1,
		user: &data.User{
			Status:          0,
			User:            &UserInterTest{},
			IsIntoSmallGame: false,
			IsNormalQuit:    false,
			Cheat:           1000,
			LastTime:        time.Time{},
			CurBox:          0,
			Level:           0,
			TotalInvest:     0,
		},
		Table: &Table{},
	}
	game.CheatConfig = config.GetLhdbConfig(game.user.Cheat)
	log.Traceln("game.CheatConfig : ", game.CheatConfig)
	game.InitIcons(true)
	//game.StartGame()

}

func TestGame(t *testing.T) {
	config.LoadLhdbConfig()

	game := &Game{
		level: 1,
		user: &data.User{
			Status:          0,
			User:            &UserInterTest{},
			IsIntoSmallGame: false,
			IsNormalQuit:    false,
			Cheat:           1000,
			LastTime:        time.Time{},
			CurBox:          0,
			Level:           0,
			TotalInvest:     0,
		},
		Table: &Table{},
	}
	game.CheatConfig = config.GetLhdbConfig(game.user.Cheat)
	log.Traceln("game.CheatConfig : ", game.CheatConfig)
	c2sMsg, _ := proto.Marshal(&msg.C2STestTool{
		Icon: 1, Count: 6,
	})
	game.ProcTestTool(c2sMsg, nil)
	game.PrintIcons()
	//game.InitIcons(true)
	//game.StartGame()

}

func TestFillSelf(t *testing.T) {
	config.LoadLhdbConfig()
	t1 := time.Now()
	game := &Game{
		level: 1,
		user: &data.User{
			Cheat: 1000, User: &UserInterTest{},
		},
		CheatConfig: config.GetLhdbConfig(1000),
	}
	game.InitIcons(true)
	game.Icons = [][]int32{
		{2, 1, 4, 2},
		{3, 3, 2, 0},
		{4, 2, 3, 0},
		{0, 0, 0, 0},
	}
	for y, _ := range game.Icons {
		for x, _ := range game.Icons[y] {
			if game.Icons[x][y] == 0 {
				continue
			}
			game.PushIntoSameList(x, y)
			if game.GetWinTmpArrCount() < 4 {
				game.ClearWinTmpArr()
				continue
			} else {
				log.Traceln("> 5 的值：", x, y, game.Icons[x][y])
				game.PrintWinTmpArr()
				game.ClearWinTmpArr()
			}
		}
	}
	for _, v := range game.WinTmpArr {
		log.Traceln(v)
		if v != nil {
			game.Icons[v.X][v.Y] = 0
		}
	}
	game.PrintIcons()

	log.Traceln(time.Now().Sub(t1))
}

func TestGetErgodicTimes(t *testing.T) {
	config.LoadLhdbConfig()
	t1 := time.Now()
	game := &Game{
		level: 1,
		user: &data.User{
			Cheat: 1000, User: &UserInterTest{},
		},
		CheatConfig: config.GetLhdbConfig(1000),
	}
	game.InitIcons(true)
	game.Icons = [][]int32{
		{2, 1, 4, 2},
		{3, 3, 1, 3},
		{2, 2, 2, 3},
		{1, 1, 2, 1},
	}
	totalAxisList := make([]*Axis, 0)
	var times int64 = 0
	icons := game.CopyGameIcons()
	log.Traceln(icons, game.Icons)
	res, times := game.GetErgodicTimes(totalAxisList, times)
	for _, v := range res {
		log.Traceln("坐标：", v.X, v.Y, v.Value)
	}
	//但是有问题，已经下落的坐标要还原，所以还是得把pushLeft，pushRight等用单独的icons提出来
	log.Traceln("结束：", times, icons)
	game.PrintIcons()
	game.AssignGameIcons(icons)
	log.Traceln("还原之后：")
	game.PrintIcons()
	log.Traceln(time.Now().Sub(t1))
}

func (game *Game) PrintWinTmpArr() {
	log.Traceln("打印PrintWinTmpArr")
	for _, v := range game.WinTmpArr {
		if v == nil {
			continue
		}
		log.Traceln(v.X, v.Y, game.Icons[v.X][v.Y])
		game.Icons[v.X][v.Y] = 0
		//log.Traceln(game.Icons[v.X][v.Y])
	}
}

func TestMaopao(t *testing.T) {
	m := make(map[int]int)
	for i := 0; i < 1000; i++ {
		m[i] = i
	}
	t1 := time.Now()
	log.Traceln(m[237])
	log.Traceln(time.Now().Sub(t1))
	return
	values := []int{2, 3, 0, 1}
	for i := len(values) - 2; i >= 0; i-- {
		values[i], values[i+1] = values[i+1], values[i]
	}
	log.Traceln(values)
}

func TestGame_GameStart(t *testing.T) {
	//log.Traceln(fmt.Sprintf(`%v`,false))
	//return
	config.LoadLhdbConfig()
	game := &Game{
		level: 1,
	}
	game.CheatConfig = config.GetLhdbConfig(game.user.Cheat)
	c2sMsg, _ := proto.Marshal(&msg.C2STestTool{
		Icon: 1, Count: 5,
	})
	game.ProcTestTool(c2sMsg, nil)
	game.PrintIcons()
}

func TestGetdiffIcon(t *testing.T) {
	config.LoadLhdbConfig()
	game := &Game{
		level:    1,
		TopIcons: []int32{4, 5, 3, 5},
	}
	game.Icons = [][]int32{
		// 4,3
		{5, 5, 2, 5},
		{2, 5, 2, 2},
		{3, 3, 3, 2},
		{2, 3, 5, 5},
	}
	//log.Traceln(game.GetDifferentIcon2(1,1,5))
	axis := make([]*Axis, 0)
	res := make([]*Axis, 0)
	var times int64
	res, times = game.GetErgodicTimes(axis, times)
	log.Traceln("times : ", times)
	for _, v := range res {
		log.Traceln(v)
	}
	game.PrintIcons()

}

func TestGetdiffTopIcon(t *testing.T) {
	config.LoadLhdbConfig()
	game := &Game{
		level:    1,
		TopIcons: []int32{4, 5, 3, 5},
	}
	game.Icons = [][]int32{
		// 4,3
		{5, 5, 2, 5},
		{2, 5, 2, 2},
		{3, 3, 3, 2},
		{2, 3, 5, 5},
	}
	for i := range game.TopIcons {
		game.TopIcons[i] = game.GetDifferentTopIcon(i, nil)
	}
	log.Traceln(game.TopIcons)
}
