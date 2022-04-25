package gamelogic

import (
	"go-game-sdk/define"
	"go-game-sdk/example/game_LaBa/970102/config"
	"go-game-sdk/example/game_LaBa/970102/msg"
	"go-game-sdk/inter"
	"go-game-sdk/lib/clock"
	"go-game-sdk/sdk/global"
	frameMsg "go-game-sdk/sdk/msg"
	"math/rand"
	"sort"
	"strconv"
	"sync"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"github.com/golang/protobuf/proto"
)

type Game struct {
	table table.TableInterface

	userMux     sync.RWMutex
	Status      msg.GameStatus  //游戏状态
	Users       map[int64]*User // 玩家列表
	TimerJob    *clock.Job      // 定时
	Cards       []byte
	SpareCards  []byte
	FirstCards  []byte
	SaveCards   []byte
	Bet         int32
	Poker       GamePoker
	PayKey      int
	Changed     bool
	TestModel   bool
	PokerType   msg.PokerType
	WeightCards []byte
}

var pokerTypeString = []string{"无", "对子", "两队", "三条", "顺子", "同花", "葫芦", "四梅", "同花顺", "同花大顺", "lucky5"}
var specialDuiZI = []byte{14, 13, 12, 9, 8, 7, 4, 3, 2}

func (lbr *Game) BindRobot(ai inter.AIUserInter) player.RobotHandler {
	return nil
}

// 实现接口
func (game *Game) ResetTable() {
	game.init()
}

func (game *Game) CloseTable() {

}

func (game *Game) OnActionUserSitDown(user player.PlayerInterface, chairId int, cfg string) int {
	log.Debugf("OnActionUserSitDown user", user.GetID())
	if user.GetScore() < game.table.GetEntranceRestrictions() {
		return define.SIT_DOWN_ERROR_OVER
	}
	if game.Bet == 0 {
		game.Bet = config.GetBet(cfg)
	}
	u := game.Users[user.GetID()]
	if u != nil {
		return define.SIT_DOWN_OK
	}
	if len(game.Users) > 1 {
		return define.SIT_DOWN_ERROR_NORMAL
	}
	return define.SIT_DOWN_OK
}

func (game *Game) OnGameMessage(subCmd int32, buffer []byte, user player.PlayerInterface) {
	switch subCmd {
	case int32(msg.MsgId_BET_Req):
		game.UserBet(buffer, user)
	case int32(msg.MsgId_CHANGECARD_Req):
		game.changeCard(buffer, user)
	case int32(msg.MsgId_COUNT_Req):
		game.count(user)
	case int32(msg.MsgId_TEST_Req):
		game.test(buffer)
	}
}

func (game *Game) SendScene(user player.PlayerInterface) bool {
	u := NewUser(user)
	game.Users[user.GetID()] = u
	game.SendSceneMsg(user, u.Bet)
	game.table.StartGame()
	return true
}

func (game *Game) UserReady(player.PlayerInterface) bool {
	return true
}

func (game *Game) GameStart(player.PlayerInterface) bool {
	return true
}

func (game *Game) LeaveGame(user player.PlayerInterface) bool {
	game.userLeave(user)
	return true
}

func (game *Game) UserExit(user player.PlayerInterface) bool {
	game.userLeave(user)
	return true
}

func (game *Game) userLeave(user player.PlayerInterface) {
	if game.Status != msg.GameStatus_game_End {
		game.count(user)
	}
	u := game.Users[user.GetID()]
	delete(game.Users, user.GetID())
	gameNum := game.table.GetGameNum()
	user.SendRecord(gameNum, u.Score-int64(u.AllBet), int64(u.AllBet), int64(u.Gold)-u.Score, u.Score, "")
	game.table.EndGame()
}

func NewGame(table table.TableInterface) *Game {
	return &Game{
		table: table,
	}
}

func (game *Game) init() {
	game.Users = make(map[int64]*User, 0)
	game.Bet = 0
	game.PayKey = -1
	game.TestModel = false
	game.PokerType = msg.PokerType_ZERO
	game.Reset()
}

func (game *Game) Reset() {
	game.Status = msg.GameStatus_game_End
	game.Changed = false
	game.Cards = make([]byte, 0)
	game.SpareCards = make([]byte, 0)
	game.FirstCards = make([]byte, 0)
	game.SaveCards = make([]byte, 0)
	game.WeightCards = make([]byte, 0)
	game.PayKey = game.getAKey("7")
	game.Poker.InitPoker()
	game.Poker.ShuffleCards()
}

func (game *Game) test(buf []byte) {
	if global.GConfig.Runmode != "dev" {
		return
	}
	req := &msg.TestReq{}
	proto.Unmarshal(buf, req)
	log.Tracef("test req =", req)
	pokerType := req.GetPokerType()
	if pokerType >= 0 && pokerType < 11 {
		game.TestModel = req.GetIsOpen()
		game.PokerType = pokerType
	}
}

// 发送场景消息
func (game *Game) SendSceneMsg(user player.PlayerInterface, bet int32) {
	res := &msg.EnterRoomRes{}
	tableInfo := &msg.TableInfo{
		TableId:    int32(game.table.GetID()),
		GameStatus: game.Status,
	}
	if len(game.Cards) > 0 {
		cards := make([]int32, 0)
		for _, v := range game.Cards {
			cards = append(cards, int32(v))
		}
		tableInfo.Poker = cards
	}
	bets := make([]int32, 5)
	for i := 0; i < 5; i++ {
		bets[i] = game.Bet * int32(i+1)
	}
	tableInfo.Bets = bets
	userInfo := &msg.UserInfo{
		UserID:   user.GetID(),
		NickName: user.GetNike(),
		Sex:      user.GetSex(),
		Head:     user.GetHead(),
		Gold:     user.GetScore(),
		Bet:      bet,
	}
	userInfos := make([]*msg.UserInfo, 0)
	userInfos = append(userInfos, userInfo)
	tableInfo.UserInfoArr = userInfos
	res.TableInfo = tableInfo
	user.SendMsg(int32(msg.MsgId_ENTER_ROOM_Res), res)
}

func (game *Game) UserBet(buf []byte, user player.PlayerInterface) {
	if game.Status != msg.GameStatus_game_End {
		return
	}
	req := &msg.BetReq{}
	proto.Unmarshal(buf, req)
	num := req.GetBet() / game.Bet
	if num < 0 || num > 5 || int64(req.GetBet()) > user.GetScore() {
		return
	}
	u := game.Users[user.GetID()]
	if u != nil {
		u.Bet = req.Bet
		u.AllBet += u.Bet
		game.Start(buf, user)
	}
	//user.SendMsg(int32(msg.MsgId_BET_Res), &msg.BetRes{})
}

func (game *Game) changeCard(buf []byte, user player.PlayerInterface) {
	if game.Changed || game.Status != msg.GameStatus_dealcard {
		return
	}
	req := &msg.ChangeCardReq{}
	proto.Unmarshal(buf, req)
	indexs := req.GetIndexs()
	length := len(indexs)
	maxIndex := int32(len(game.Cards)) - 1
	log.Tracef("change indexs", indexs)
	log.Tracef("change before", game.Cards)
	game.SaveCards = append(game.SaveCards, game.Cards...)
	for i := 0; i < length; i++ {
		index := indexs[i]
		if index > maxIndex {
			continue
		}
		weight := game.checkIsWeightCard(i, game.Cards[index])
		game.Cards[index] = game.SpareCards[i]
		if weight > 0 {
			game.Cards[index] = weight
		}
		game.SaveCards[index] = 0
	}
	log.Tracef("change after", game.Cards)
	res := &msg.ChangeCardRes{
		Poker: game.transformCardForMsg(game.Cards),
	}
	game.Changed = true
	user.SendMsg(int32(msg.MsgId_CHANGECARD_Res), res)
}

// 检查是否权重牌
func (game *Game) checkIsWeightCard(index int, card byte) byte {
	value, _ := GetCardValueAndColor(card)
	for _, v := range game.WeightCards {
		if value == v {
			for k, v := range game.SpareCards {
				if k < index {
					continue
				}
				cardValue, _ := GetCardValueAndColor(v)
				if value == cardValue && k == index {
					return 0
				}
				if value == cardValue && k != index {
					tem := game.SpareCards[index]
					game.SpareCards[index] = v
					game.SpareCards[k] = tem
					return 0
				}
			}
			weighrCard := game.dealADesignativeCard(v)
			cardValue, _ := GetCardValueAndColor(weighrCard)
			if value == cardValue {
				return weighrCard
			}
		}
	}
	return 0
}

// 开始游戏
func (game *Game) Start(buf []byte, user player.PlayerInterface) {
	u := game.Users[user.GetID()]
	if u == nil || u.Bet == 0 {
		return
	}
	game.Status = msg.GameStatus_game_Start
	//game.table.StartGame()
	//game.TimerJob, _ = game.table.AddTimer(time.Duration(2000), func() {
	game.dealcard(user)
	//})

}

func (game *Game) dealcard(user player.PlayerInterface) {
	game.Status = msg.GameStatus_dealcard
	chance := game.getWinChance(user)
	cards := make([]byte, 0)
	pokerType := msg.PokerType_ZERO
	log.Tracef("chance =", chance)
	if !game.TestModel && chance == 0 {
		cards = game.GetARandomCards()
	}
	if !game.TestModel && chance != 0 {
		pokerType = game.GetPokerType(chance)
		cards = game.GetACards(pokerType)
	}
	if game.TestModel {
		cards = game.GetACards(game.PokerType)
	}
	//if pokerType == msg.PokerType_DUIZI {
	//	game.Cards = append(game.Cards, cards[:5]...)
	//	game.SpareCards = append(game.SpareCards, cards[5:]...)
	//}
	//if pokerType != msg.PokerType_DUIZI {
	length := len(cards)
	for i := 0; i < length; i++ {
		r := rand.Intn(100)
		card := cards[i]
		if (len(game.SpareCards) == 5) || (r < 80 && len(game.Cards) < 5) {
			game.Cards = append(game.Cards, card)
			continue
		}
		game.SpareCards = append(game.SpareCards, card)
	}
	//}
	game.FirstCards = game.Cards

	rand.Shuffle(5, func(i, j int) {
		game.Cards[i], game.Cards[j] = game.Cards[j], game.Cards[i]
	})

	user.SendMsg(int32(msg.MsgId_BET_Res), &msg.BetRes{
		Poker:     game.transformCardForMsg(game.Cards),
		PokerType: game.CheckPokerType(game.Cards),
	})

	//game.TimerJob, _ = game.table.AddTimer(time.Duration(config.GetOperationTime()), func() {//
	//	game.count(user)
	//})

}

func (game *Game) transformCardForMsg(cards []byte) []int32 {
	tem := make([]int32, 0)
	values := make([]byte, 0)
	huaSe := make([]byte, 0)
	for _, v := range cards {
		value, se := GetCardValueAndColor(v)
		values = append(values, value)
		huaSe = append(huaSe, se)
		tem = append(tem, int32(v))
	}
	//fmt.Println("card value =", values)
	//fmt.Println("card huaSe =", huaSe)
	log.Tracef("cards =", cards)
	log.Tracef("card value =", values)
	log.Tracef("card huaSe =", huaSe)
	return tem
}

func (game *Game) GetPokerType(chance int32) msg.PokerType {
	if game.checkWinChance(chance) {
		return msg.PokerType(game.getAKey(strconv.Itoa(game.PayKey + 3)))
	}
	return msg.PokerType_ZERO
}

func (game *Game) getWinChance(user player.PlayerInterface) int32 {
	key := game.table.GetRoomProb()
	PointKey := user.GetProb()
	chance := int32(0)
	u := game.Users[user.GetID()]
	if key == 0 {
		key = 1000
	}
	if PointKey == 0 {
		u.ControlKey = key
		u.IsPoint = "否"
		chance = config.GetXueChiChance(key)
	}
	if PointKey != 0 {
		u.ControlKey = PointKey
		u.IsPoint = "是"
		chance = config.GetPlayerChance(PointKey)
	}
	return chance
}

func (game *Game) checkWinChance(chance int32) bool {
	r := rand.Intn(10000)
	if int32(r) > chance {
		return true
	}
	return false
}

func (game *Game) GetARandomCards() []byte {
	cards := make([]byte, 0)
	wildNum := game.getWildNum()
	cards = append(cards, game.getWildCards(wildNum)...)
	for i := 0; i < 10-wildNum; i++ {
		cards = append(cards, game.Poker.DealCards())
	}
	return cards
}

func (game *Game) GetACards(pokerType msg.PokerType) []byte {
	cards := make([]byte, 0)
B:
	if pokerType == msg.PokerType_ZERO {
		cards = game.getZeroCards()
	}
	if pokerType == msg.PokerType_DUIZI {
		cards = game.getDuiZiCards()
	}
	if pokerType == msg.PokerType_LIANGDUI {
		cards = game.getTwoDuiCards()
	}
	if pokerType == msg.PokerType_SANTIAO {
		cards = game.getSanTiaoCards()
	}
	if pokerType == msg.PokerType_SHUNZI {
		cards = game.getShunZiCards()
	}
	if pokerType == msg.PokerType_TONGHUA {
		cards = game.getTongHuaCards()
	}
	if pokerType == msg.PokerType_HULU {
		cards = game.getHuLuCards()
	}
	if pokerType == msg.PokerType_SIMEI {
		cards = game.getSiMeiCards()
	}
	if pokerType == msg.PokerType_TONGHUASHUN {
		cards = game.getTongHuaShunCards()
	}
	if pokerType == msg.PokerType_TONGHUADASHUN {
		cards = game.getTongHuaDaShunCards()
	}
	if pokerType == msg.PokerType_LUCKY5 {
		cards = game.getLuck5Cards()
	}
	//fmt.Println("pokerType = ", pokerType, cards, len(game.Poker.Cards))
	log.Tracef("pokerType = ", pokerType, cards)
A:
	if len(game.Poker.Cards) == 0 {
		log.Errorf("not find cards by pokerType", pokerType)
		game.Poker.InitPoker()
		game.Poker.ShuffleCards()
		pokerType = msg.PokerType_ZERO
		goto B
	}
	if len(cards) < 10 {
		card := game.Poker.DealCards()
		tem := make([]byte, 0)
		tem = append(tem, cards...)
		tem = append(tem, card)
		//game.transformCardForMsg(tem)
		if !game.checkAllCompose(tem, pokerType) {
			goto A
		}
		cards = tem
	}
	if len(cards) < 10 {
		goto A
	}
	game.transformCardForMsg(cards)
	log.Tracef("get card by type", pokerType, cards)
	return cards
}

func (game *Game) getZeroCards() []byte {
	cards := make([]byte, 0)
	tongHua := true
	huaSe := byte(0)
A:
	card := game.Poker.DealCards()
	tem := make([]byte, 0)
	tem = append(tem, cards...)
	tem = append(tem, card)
	if !game.checkIsZeroType(tem) {
		goto A
	}
	value, se := GetCardValueAndColor(card)
	if value == 15 {
		goto A
	}
	if huaSe == 0 {
		huaSe = se
	}
	if huaSe != se {
		tongHua = false
	}
	cards = tem
	if len(cards) < 5 {
		goto A
	}
	if len(cards) == 5 && tongHua {
		cards = append(cards[:0], cards[1:]...)
		goto A
	}
	return cards
}

func (game *Game) checkIsZeroType(cards []byte) bool {
	values, cardValue, wildNum := game.getCardValue(cards)
	length := len(values)
	if wildNum > 0 && len(cardValue) != length {
		return false
	}
	if len(cardValue) < length-1 {
		return false
	}
	isShunZi := true
	tem := byte(0)
	for i := 0; i < length; i++ {
		card := values[i]
		if wildNum > 0 && card > 10 && card != 15 {
			return false
		}
		if i+1 < length && values[i+1] == card && card > 10 {
			return false
		}
		if i+1 < length && values[i+1]-card > 1 {
			tem += values[i+1] - card - 1
		}
		if tem > wildNum {
			isShunZi = false
		}
	}
	if length == 5 && isShunZi {
		return false
	}
	return true
}

func (game *Game) checkIsShunZi(values []byte, wildNum byte) (bool, byte) {
	isShunZi := true
	tem := byte(0)
	length := len(values)
	for i := 0; i < length-1; i++ {
		card := values[i]
		sub := values[i+1] - card
		if sub > 1 {
			tem += sub - 1
		}
		if sub == 0 || tem > wildNum {
			isShunZi = false
			break
		}
	}
	return isShunZi, wildNum - tem
}

func (game *Game) getCardValue(cards []byte) ([]byte, map[byte]int, byte) {
	tem := make([]byte, 0)
	cardValue := make(map[byte]int, 0)
	wildNum := byte(0)
	for _, v := range cards {
		value, _ := GetCardValueAndColor(v)
		tem = append(tem, value)
		cardValue[value]++
		if value == 15 {
			wildNum++
		}
	}
	sort.Slice(tem, func(i, j int) bool {
		return tem[i] < tem[j]
	})
	return tem, cardValue, wildNum
}

func (game *Game) getDuiZiCards() []byte {
	cards := make([]byte, 0)
	r := rand.Intn(4) + 11
	game.WeightCards = append(game.WeightCards, byte(r))
	wildNum := game.getWildNum()
	if wildNum > 0 {
		cards = append(cards, game.getAWildCard())
		cards = append(cards, game.getSpecialDuiZi()...)
		rand.Shuffle(10, func(i, j int) {
			cards[i], cards[j] = cards[j], cards[i]
		})
		return cards
	}
A:
	cards = append(cards, game.dealADesignativeCard(byte(r)))
	if len(cards) < 2 {
		goto A
	}
	lastCardValue := byte(0)
B:
	card := game.Poker.DealCards()
	value, _ := GetCardValueAndColor(card)
	if value == byte(r) || value == lastCardValue || value == 15 {
		goto B
	}
	tem := make([]byte, 0)
	tem = append(tem, cards...)
	tem = append(tem, card)
	if len(tem) == 5 && game.CheckPokerType(tem) != msg.PokerType_DUIZI {
		goto B
	}
	lastCardValue = value
	cards = tem
	if len(cards) < 5 {
		goto B
	}
	return cards
}

func (game *Game) getSpecialDuiZi() []byte {
	cards := make([]byte, 0)
	huaSeNum := make(map[byte]int, 0)
	for _, v := range specialDuiZI {
	A:
		card := game.dealADesignativeCard(v)
		_, se := GetCardValueAndColor(card)
		if huaSeNum[se] > 3 {
			goto A
		}
		huaSeNum[se]++
		cards = append(cards, card)
	}
	return cards
}

func (game *Game) getTwoDuiCards() []byte {
	cards := make([]byte, 0)
	r1 := rand.Intn(13) + 2
A1:
	r2 := rand.Intn(13) + 2
	if r2 == r1 {
		goto A1
	}
	cards = append(cards, game.dealADesignativeCard(byte(r1)))
	cards = append(cards, game.dealADesignativeCard(byte(r2)))
	cards = append(cards, game.dealADesignativeCard(byte(r1)))
	cards = append(cards, game.dealADesignativeCard(byte(r2)))

	game.WeightCards = append(game.WeightCards, byte(r1))
	game.WeightCards = append(game.WeightCards, byte(r2))
B:
	card := game.Poker.DealCards()
	tem := make([]byte, 0)
	tem = append(tem, cards...)
	tem = append(tem, card)
	if game.CheckPokerType(tem) != msg.PokerType_LIANGDUI {
		goto B
	}
	cards = tem

	return cards
}

func (game *Game) getSanTiaoCards() []byte {
	cards := make([]byte, 0)
	r1 := rand.Intn(13) + 2
	wildNum := game.getWildNum()
	game.WeightCards = append(game.WeightCards, byte(r1))
	cards = append(cards, game.getWildCards(wildNum)...)
	for i := 0; i < 3-wildNum; i++ {
		cards = append(cards, game.dealADesignativeCard(byte(r1)))
	}
B:
	card := game.Poker.DealCards()
	value, _ := GetCardValueAndColor(card)
	if value == byte(r1) || value == 15 {
		goto B
	}
	//tem := make([]byte, 0)
	//tem = append(tem, cards...)
	//tem = append(tem, card)
	cards = append(cards, card)
	if len(cards) < 5 {
		goto B
	}
	if game.CheckPokerType(cards) != msg.PokerType_SANTIAO {
		cards = append(cards[:4], cards[5:]...)
		goto B
	}
	return cards
}

func (game *Game) getShunZiCards() []byte {
	cards := make([]byte, 5)
	r1 := rand.Intn(13) + 2
	wildNum := game.getWildNum()
	add := game.checkAddOrSub(r1, wildNum)
	tongHua := true
	huaSe := byte(0)
	for i := 0; i < 5-wildNum; i++ {
		cards[i] = game.dealADesignativeCard(byte(r1))
		_, se := GetCardValueAndColor(cards[i])
		if i == 0 {
			huaSe = se
		}
		if huaSe != se {
			tongHua = false
		}
		if add {
			r1++
		}
		if !add {
			r1--
		}
	}
	cards = append(cards[:5-wildNum], game.getWildCards(wildNum)...)
	if tongHua {
		r, _ := GetCardValueAndColor(cards[0])
		cards = append(cards[:0], cards[1:]...)
		cards = append(cards, game.dealADesignativeCard(r))
	}
	return cards
}

func (game *Game) getTongHuaCards() []byte {
	cards := make([]byte, 0)
	wildNum := game.getWildNum()
	if wildNum > 0 {
		cards = append(cards, game.getAWildCard())
	}
	huaSe := byte(0)
A:
	if huaSe == 0 {
		card := game.Poker.DealCards()
		cards = append(cards, card)
		_, se := GetCardValueAndColor(card)
		huaSe = se
		goto A
	}
	if huaSe != 0 {
		card := game.dealADesignativeHuaSeCard(huaSe)
		cards = append(cards, card)
	}
	if len(cards) < 5 {
		goto A
	}
	if len(cards) == 5 && game.CheckPokerType(cards) != msg.PokerType_TONGHUA {
		cards = append(cards[:4], cards[5:]...)
		goto A
	}
	return cards
}

func (game *Game) checkIsContainCard(cards []byte, card byte) bool {
	for _, v := range cards {
		value, _ := GetCardValueAndColor(v)
		if card == value {
			return true
		}
	}
	return false
}

func (game *Game) getHuLuCards() []byte {
	cards := make([]byte, 0)
	r1 := rand.Intn(13) + 2
A1:
	r2 := rand.Intn(13) + 2
	if r2 == r1 {
		goto A1
	}
	wildNum := game.getWildNum()
	if wildNum > 0 {
		cards = append(cards, game.getAWildCard())
	}
	num1 := 0
	num2 := 0
	if wildNum == 0 {
		num1 = 3
		num2 = 2
	}
	if wildNum > 0 {
		num1 = 2
		num2 = 2
	}
	for i := 0; i < num1; i++ {
		cards = append(cards, game.dealADesignativeCard(byte(r1)))
	}
	for i := 0; i < num2; i++ {
		cards = append(cards, game.dealADesignativeCard(byte(r2)))
	}
	return cards
}

func (game *Game) getSiMeiCards() []byte {
	cards := make([]byte, 0)
	r1 := rand.Intn(13) + 2
	wildNum := game.getWildNum()
	cards = append(cards, game.getWildCards(wildNum)...)
	num := 0
	if wildNum == 0 {
		num = 4
	}
	if wildNum == 1 {
		num = 3
	}
	if wildNum == 2 {
		num = 2
	}
	for i := 0; i < num; i++ {
		cards = append(cards, game.dealADesignativeCard(byte(r1)))
	}
B:
	card := game.Poker.DealCards()
	value, _ := GetCardValueAndColor(card)
	if value == byte(r1) || value == 15 {
		goto B
	}
	cards = append(cards, card)
	return cards
}

func (game *Game) getTongHuaShunCards() []byte {
	cards := make([]byte, 5)
	r1 := rand.Intn(13) + 2
	wildNum := game.getWildNum()
	cards = append(cards[:5-wildNum], game.getWildCards(wildNum)...)
	add := game.checkAddOrSub(r1, wildNum)
	if (add && r1 == 10) || (!add && r1 == 14) {
		r1--
	}
	huaSe := byte(0)
	for i := 0; i < 5-wildNum; i++ {
		if i == 0 {
			cards[i] = game.dealADesignativeCard(byte(r1))
			_, se := GetCardValueAndColor(cards[i])
			huaSe = se
		}
		if i != 0 {
			cards[i] = game.dealADesignativeTongHuaCard(byte(r1), huaSe)
		}
		if add {
			r1++
		}
		if !add {
			r1--
		}
	}
	return cards
}

func (game *Game) getTongHuaDaShunCards() []byte {
	cards := make([]byte, 5)
	wildNum := game.getWildNum()
	cards = append(cards[:5-wildNum], game.getWildCards(wildNum)...)
	r1 := 14 - wildNum
	huaSe := byte(0)
	for i := 0; i < 5-wildNum; i++ {
		if i == 0 {
			cards[i] = game.dealADesignativeCard(byte(r1))
			_, se := GetCardValueAndColor(cards[i])
			huaSe = se
		}
		if i != 0 {
			cards[i] = game.dealADesignativeTongHuaCard(byte(r1), huaSe)
		}
		r1--
	}
	return cards
}
func (game *Game) getLuck5Cards() []byte {
	cards := make([]byte, 0)
	r1 := rand.Intn(13) + 2
	wildNum := game.getWildNum()
	if wildNum == 0 {
		wildNum = 1
	}
	cards = append(cards, game.getWildCards(wildNum)...)
	for i := 0; i < 5-wildNum; i++ {
		cards = append(cards, game.dealADesignativeCard(byte(r1)))
	}
	return cards
}

func (game *Game) checkAddOrSub(v, wildNum int) bool {
	return !(v-wildNum > 10)
}

func (game *Game) dealADesignativeHuaSeCard(huaSe byte) byte {
	for k, v := range game.Poker.Cards {
		_, se := GetCardValueAndColor(v)
		if se == huaSe {
			game.Poker.Cards = append(game.Poker.Cards[:k], game.Poker.Cards[k+1:]...)
			return v
		}
	}
	log.Errorf("not find card value and color", huaSe)
	return game.Poker.DealCards()
}

func (game *Game) dealADesignativeTongHuaCard(cardValue, huaSe byte) byte {
	for k, v := range game.Poker.Cards {
		value, se := GetCardValueAndColor(v)
		if value == cardValue && se == huaSe {
			game.Poker.Cards = append(game.Poker.Cards[:k], game.Poker.Cards[k+1:]...)
			return v
		}
	}
	log.Errorf("not find card value and color", cardValue, huaSe)
	return game.Poker.DealCards()
}

func (game *Game) dealADesignativeCard(cardValue byte) byte {
	for k, v := range game.Poker.Cards {
		value, _ := GetCardValueAndColor(v)
		if value == cardValue {
			game.Poker.Cards = append(game.Poker.Cards[:k], game.Poker.Cards[k+1:]...)
			return v
		}
	}
	log.Errorf("not find card value", cardValue)
	return game.Poker.DealCards()
}

func (game *Game) getWildCards(wildNum int) []byte {
	wildCards := make([]byte, 0)

	if wildNum == 1 {
		wildCards = append(wildCards, game.getAWildCard())
	}
	if wildNum == 2 {
		wildCards = append(wildCards, King...)
	}
	return wildCards
}

func (game *Game) getAWildCard() byte {
	wildIndex := rand.Intn(2)
	return King[wildIndex]
}

func (game *Game) getWildNum() int {
	r := rand.Intn(10000)
	if r < 4 {
		return 2
	}
	if r < 185 {
		return 1
	}
	return 0
}

func (game *Game) checkAllCompose(cards []byte, pokerType msg.PokerType) bool {
	result := game.getAllCompose(len(cards), 5)
	for _, v := range result {
		tem := make([]byte, 0)
		for k, v2 := range v {
			if v2 == 1 {
				tem = append(tem, cards[k])
			}
		}
		p := game.CheckPokerType(tem)
		if p > pokerType {
			log.Tracef("error type", p, game.transformCardForMsg(tem))
			return false
		}
	}
	return true
}

func (game *Game) getAllCompose(n, m int) [][]byte {
	result := make([][]byte, 0)
	first := make([]byte, n)
	for i := 0; i < m; i++ {
		first[i] = 1
	}
	if n >= m {
		result = append(result, first)
	}
	if n > m {
	A:
		tem := game.getACompose(first)
		if len(tem) != 0 {
			first = tem
			result = append(result, first)
			goto A
		}
	}
	return result
}

func (game *Game) getACompose(fisrt []byte) []byte {
	tem := make([]byte, 0)
	tem = append(tem, fisrt...)
	length := len(fisrt)
	for i := 0; i < length; i++ {
		if i+1 < length && tem[i] == 1 && tem[i+1] == 0 {
			tem[i], tem[i+1] = tem[i+1], tem[i]
			tmp := i - 1
			for j := 0; j < i; j++ {
				//if j < num {
				//	tem[j] = 1
				//}
				//if j >= num {
				//	tem[j] = 0
				//}
				if j >= tmp {
					break
				}
				if tmp > 1 && tem[tmp] != 1 {
					tmp--
				}
				if tem[j] == 0 {
					tem[j], tem[tmp] = tem[tmp], tem[j]
				}
			}
			return tem
		}
	}
	return []byte{}
}

func (game *Game) getChangeNum(arr []byte) int {
	sum := 0
	for _, v := range arr {
		if v == 1 {
			sum++
		}
	}
	return sum
}

func (game *Game) CheckPokerType(cards []byte) msg.PokerType {
	tem := make([]byte, 0)
	tem = append(tem, cards...)
	//log.Tracef("tem card =", tem)
	sort.Slice(tem, func(i, j int) bool {
		value1, _ := GetCardValueAndColor(tem[i])
		value2, _ := GetCardValueAndColor(tem[j])
		return value1 < value2
	})
	//log.Tracef("sort tem card =", tem)
	isTongHua := false
	isShunZi := true
	isSanTiao := false
	isSiTiao := false
	hasJQKA := false
	huaSe := make(map[byte]int, 0)
	cardValue := make(map[byte]int, 0)
	values := make([]byte, 0)
	dui := make([]byte, 0)
	maxCardValue, _ := GetCardValueAndColor(tem[0])
	wildNum := 0
	length := len(tem)
	for i := 0; i < length; i++ {
		card1 := tem[i]
		value, se := GetCardValueAndColor(card1)
		if value == 15 {
			wildNum++
			continue
		}
		huaSe[se]++
		cardValue[value]++
		values = append(values, value)
		if value > 10 {
			hasJQKA = true
		}
		if value > maxCardValue {
			maxCardValue = value
		}
		if cardValue[value] == 2 {
			dui = append(dui, value)
		}
		if cardValue[value] == 3 {
			isSanTiao = true
		}
		if cardValue[value] == 4 {
			isSiTiao = true
		}
	}
	if len(huaSe) == 1 {
		isTongHua = true
	}
	isShunZi, spareNum := game.checkIsShunZi(values, byte(wildNum))
	cardValueLen := len(cardValue)
	if cardValueLen == 5 && !isShunZi && !isTongHua {
		return msg.PokerType_ZERO
	}
	if cardValueLen == 4 && wildNum == 0 && dui[0] <= 10 {
		return msg.PokerType_ZERO
	}
	if cardValueLen == 4 && !isShunZi && !hasJQKA && !isTongHua {
		return msg.PokerType_ZERO
	}
	if cardValueLen == 4 && !isShunZi && !isTongHua && (hasJQKA || dui[0] > 10) {
		return msg.PokerType_DUIZI
	}
	if cardValueLen == 3 && !isSanTiao && !isShunZi && !isTongHua && wildNum == 0 {
		return msg.PokerType_LIANGDUI
	}
	if cardValueLen == 3 && !isShunZi && (wildNum == 2 || isSanTiao || len(dui) > 0) && !isTongHua {
		return msg.PokerType_SANTIAO
	}
	if isShunZi && !isTongHua {
		return msg.PokerType_SHUNZI
	}
	if !isShunZi && isTongHua {
		return msg.PokerType_TONGHUA
	}
	if cardValueLen == 2 && !isSiTiao && (isSanTiao && wildNum == 0) {
		return msg.PokerType_HULU
	}
	if cardValueLen == 2 && len(dui) == 2 {
		return msg.PokerType_HULU
	}
	if cardValueLen == 2 && ((isSiTiao && wildNum == 0) || isSanTiao || wildNum == 2) {
		return msg.PokerType_SIMEI
	}
	if isTongHua && isShunZi && maxCardValue < 14-spareNum && maxCardValue != 14 {
		return msg.PokerType_TONGHUASHUN
	}
	if isTongHua && isShunZi && (maxCardValue >= 14-spareNum || maxCardValue == 14) {
		return msg.PokerType_TONGHUADASHUN
	}
	if (isSanTiao && wildNum == 2) || (isSiTiao && wildNum == 1) {
		return msg.PokerType_LUCKY5
	}
	return msg.PokerType_ZERO
}

// 发送结算消息
func (game *Game) count(user player.PlayerInterface) {
	if game.Status != msg.GameStatus_dealcard {
		return
	}
	if game.TimerJob != nil {
		game.TimerJob.Cancel()
	}
	game.Status = msg.GameStatus_game_Start
	u := game.Users[user.GetID()]
	if u != nil {
		pokerType := game.CheckPokerType(game.Cards)
		gold := int32(0)
		score := int64(0)
		if pokerType != msg.PokerType_ZERO {
			multiple := config.GetPokerPay(int(pokerType), game.PayKey)
			gold = multiple * u.Bet
			score, _ = user.SetScore(game.table.GetGameNum(), int64(gold), game.table.GetRoomRate())
		}
		u.Gold += gold
		u.Score += score
		user.SetScore(game.table.GetGameNum(), -int64(u.Bet), game.table.GetRoomRate())
		user.SendChip(int64(u.Bet))
		//gameNum := user.GetRoomNum()
		//user.SendRecord(gameNum, score - int64(u.Bet), int64(u.Bet), int64(gold) - score, score, GetCardString(game.Cards))
		res := &msg.GameEndRes{
			PokerType: pokerType,
			Gold:      score,
			Score:     user.GetScore(),
			PayKey:    int32(game.PayKey),
		}
		user.SendMsg(int32(msg.MsgId_COUNT_Res), res)
		game.createOperationLog(u, int64(gold), int64(u.Bet), pokerType)
		game.checkMarquee(u.user.GetNike(), pokerType, int64(gold))
	}
	game.Reset()
	//game.table.EndGame()
}

func (game *Game) createOperationLog(user *User, coinChange, bet int64, pokerType msg.PokerType) {
	userId := user.user.GetID()
	content := "用户ID:" + strconv.FormatInt(userId, 10) +
		" 下注金额:" + score.GetScoreStr(bet) +
		" 赔付金额:" + score.GetScoreStr(coinChange) +
		" 用户剩余金额:" + score.GetScoreStr(user.user.GetScore()) +
		" 作弊值: " + strconv.Itoa(int(user.ControlKey)) +
		" 是否点控: " + user.IsPoint +
		" 转动一牌 " + GetCardString(game.FirstCards) +
		" 保留牌 " + GetCardString(game.SaveCards) +
		" 转动二牌 " + GetCardString(game.Cards)
	//" 牌型 " + game.getPokerTypeString(pokerType)
	game.table.WriteLogs(userId, content)
}

func (game *Game) getPokerTypeString(pokerType msg.PokerType) string {
	return pokerTypeString[int(pokerType)]
}

func (game *Game) checkMarquee(nickName string, pokerType msg.PokerType, coin int64) {
	orderRules := game.orderMarqueeRules(game.table.GetMarqueeConfig())
	//for _, v := range orderRules {
	length := len(orderRules)
	for i := 0; i < length; i++ {
		v := orderRules[i]
		//SpecialCondition
		special, _ := strconv.ParseInt(v.GetSpecialCondition(), 10, 64)
		specialType := msg.PokerType_ZERO
		if special > 0 && special < 11 {
			specialType = msg.PokerType(special)
		}
		if v.GetAmountLimit() < 0 || coin < v.GetAmountLimit() || specialType > pokerType {
			continue
		}
		game.table.CreateMarquee(nickName, coin, pokerTypeString[int(pokerType)], v.GetRuleId())
		break
	}
}

func (game *Game) orderMarqueeRules(rules []*frameMsg.MarqueeConfig) []*frameMsg.MarqueeConfig {
	orderRules := make([]*frameMsg.MarqueeConfig, 0)
	orderRules = append(orderRules, rules...)
	length := len(orderRules)
	for i := 0; i < length; i++ {
		for j := i + 1; j < length; j++ {
			change := false
			special, _ := strconv.ParseInt(orderRules[i].GetSpecialCondition(), 10, 64)
			special1, _ := strconv.ParseInt(orderRules[j].GetSpecialCondition(), 10, 64)
			if special != 0 {
				if special1 != 0 && special1 > special {
					change = true
				}
			}
			if special == 0 && (orderRules[i].GetAmountLimit() < orderRules[j].GetAmountLimit() ||
				special1 != 0) {
				change = true
			}
			if change {
				tem := orderRules[i]
				orderRules[i] = orderRules[j]
				orderRules[j] = tem
			}
		}
	}
	return orderRules
}

func (game *Game) getAKey(key string) int {
	r := rand.Intn(10000)
	chance := 0
	for i := 1; i < 11; i++ {
		chance += config.GetValue(key, strconv.Itoa(i))
		if r < chance {
			return i
		}
	}
	return 1
}
