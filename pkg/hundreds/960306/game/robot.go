package game

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"game_frame_v2/game/clock"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"

	"math/rand"

	"github.com/bitly/go-simplejson"
	"github.com/sipt/GoJsoner"
)

const (
	coutLineID = 0 //点数
	SDLineID   = 1 //单双
	BSLineID   = 2 //大小
	WEILineID  = 3 //豹子

)

type RobotProbability struct {
	Min         int
	Max         int
	Probability int
}

type RobotBetConfig struct {
	BetWeight   []int
	TotalWeight int
}

type RobotConfig struct {
	BetTime            []int              //下注时间
	BetWeight          []int              //下注权重
	ToTalBetWeoght     int                //
	BetPlace           []int              //下注区域权重
	TotalBetPlace      int                //下注区域总权重
	Line               [][]int            //平衡线，少于这个线和多余这个线使用不同的配置
	LessLine           []int              //低于这个线时下
	TotalLessLine      int                //
	OverLine           []int              //高于下注线时下注权重
	TotalOverLine      int                //
	Limit              int                //机器人下注次数限制
	SitDownTime        []int              //坐下随机时间
	StandUpTime        []int              //起立随机时间
	SitDownProbability []RobotProbability //机器人坐下的概率
	StandUpProbability []RobotProbability //机器人起立的概率
	BetPro             []int              //下注万分比区间
	BetConfig          []RobotBetConfig   //下注配置
}

type Robot struct {
	User           player.RobotInterface
	GameLogic      *Game
	BetCount       int        //下注限制
	TimerJob       *clock.Job //job
	LastBetPlace   int        //机器人上次大小下注的区域
	LastBetBSPlace int        //机器人上次大小下注的区域
	LastBetSDPlace int        //机器人上次单双下注的区域
}

var RConfig RobotConfig

//读取配置文件
func (rc *RobotConfig) LoadLabadCfg() {
	data, err := ioutil.ReadFile("conf/robot.json")
	if err != nil {
		log.Traceln("File reading error", err)
		return
	}
	//去除配置文件中的注释
	result, _ := GoJsoner.Discard(string(data))
	rc.analysiscfg(result)
}

//解析配置文件
func (rc *RobotConfig) analysiscfg(json_str string) {
	//使用简单json来解析。
	js, err := simplejson.NewJson([]byte(json_str))
	if err != nil {
		fmt.Printf("analysiscfg err%v\n", err)
		fmt.Printf("%v\n", json_str)
		return
	}

	rc.getBetTime(js)
	rc.getBetWeight(js)
	rc.getBetPlace(js)
	rc.getLineConfig(js)
	rc.getLessLine(js)
	rc.getOverLine(js)
	rc.getSitDownTime(js)
	rc.getStandUpTime(js)
	rc.getSitDownProbability(js)
	rc.getStandUpProbability(js)
	rc.getBetPro(js)
	rc.getBetConfig(js)
	rc.Limit, _ = js.Get("limit").Int()
}

//获取平衡线配置
func (rc *RobotConfig) getLineConfig(js *simplejson.Json) {
	i := 1
	for {
		str := strconv.Itoa(i)
		arr, err := js.Get("line").Get(str).Array()
		if err != nil {
			break
		}

		var tmp []int
		for n := 0; n < len(arr); n++ {
			line, _ := arr[n].(json.Number).Int64()
			tmp = append(tmp, int(line))
		}

		rc.Line = append(rc.Line, tmp)
		i++
	}
}
func (rc *RobotConfig) getBetConfig(js *simplejson.Json) {
	i := 1
	for {
		str := "bet" + strconv.Itoa(i)
		arr, err := js.Get("bet").Get(str).Array()
		if err != nil {
			break
		}

		var tmp RobotBetConfig
		for n := 0; n < len(arr); n++ {
			weight, _ := arr[n].(json.Number).Int64()
			tmp.BetWeight = append(tmp.BetWeight, int(weight))
			tmp.TotalWeight += int(weight)
		}

		rc.BetConfig = append(rc.BetConfig, tmp)
		i++
	}
}

func (rc *RobotConfig) getBetTime(js *simplejson.Json) {
	arr, _ := js.Get("bettime").Array()
	for n := 0; n < len(arr); n++ {
		bettime, _ := arr[n].(json.Number).Int64()
		rc.BetTime = append(rc.BetTime, int(bettime))
	}

	//	log.Tracef("下注时间%v", rc.BetTime)
}

func (rc *RobotConfig) getBetPro(js *simplejson.Json) {
	arr, _ := js.Get("betpro").Array()
	for n := 0; n < len(arr); n++ {
		betpro, _ := arr[n].(json.Number).Int64()
		rc.BetPro = append(rc.BetPro, int(betpro))
	}
}

func (rc *RobotConfig) getBetWeight(js *simplejson.Json) {
	arr, _ := js.Get("betweight").Array()
	for n := 0; n < len(arr); n++ {
		betweight, _ := arr[n].(json.Number).Int64()
		rc.BetWeight = append(rc.BetWeight, int(betweight))
		rc.ToTalBetWeoght += int(betweight)
	}
}

func (rc *RobotConfig) getBetPlace(js *simplejson.Json) {
	arr, _ := js.Get("betplace").Array()
	for n := 0; n < len(arr); n++ {
		betplace, _ := arr[n].(json.Number).Int64()
		rc.BetPlace = append(rc.BetPlace, int(betplace))
		rc.TotalBetPlace += int(betplace)
	}
}

func (rc *RobotConfig) getLessLine(js *simplejson.Json) {
	less, _ := js.Get("lessline").Get("less").Int()
	more, _ := js.Get("lessline").Get("more").Int()
	rc.LessLine = append(rc.LessLine, less)
	rc.LessLine = append(rc.LessLine, more)
	rc.TotalLessLine += less
	rc.TotalLessLine += more
}

func (rc *RobotConfig) getOverLine(js *simplejson.Json) {
	less, _ := js.Get("overline").Get("less").Int()
	more, _ := js.Get("overline").Get("more").Int()
	rc.OverLine = append(rc.OverLine, less)
	rc.OverLine = append(rc.OverLine, more)
	rc.TotalOverLine += less
	rc.TotalOverLine += more
}

func (rc *RobotConfig) getSitDownTime(js *simplejson.Json) {
	arr, _ := js.Get("sitdowntime").Array()
	for n := 0; n < len(arr); n++ {
		time, _ := arr[n].(json.Number).Int64()
		rc.SitDownTime = append(rc.SitDownTime, int(time))
	}
}

func (rc *RobotConfig) getStandUpTime(js *simplejson.Json) {
	arr, _ := js.Get("standuptime").Array()
	for n := 0; n < len(arr); n++ {
		time, _ := arr[n].(json.Number).Int64()
		rc.StandUpTime = append(rc.StandUpTime, int(time))
	}
}

func (rc *RobotConfig) getSitDownProbability(js *simplejson.Json) {
	i := 0
	for {
		str := strconv.Itoa(i)
		min, err := js.Get("sitdownprobability").Get(str).Get("min").Int()
		if err != nil {
			break
		}
		max, _ := js.Get("sitdownprobability").Get(str).Get("max").Int()
		probability, _ := js.Get("sitdownprobability").Get(str).Get("probability").Int()
		var rp RobotProbability
		rp.Min = min
		rp.Max = max
		rp.Probability = probability
		rc.SitDownProbability = append(rc.SitDownProbability, rp)
		i++
	}
}

func (rc *RobotConfig) getStandUpProbability(js *simplejson.Json) {
	var tmp RobotProbability
	rc.StandUpProbability = append(rc.StandUpProbability, tmp)

	i := 1
	for {
		str := strconv.Itoa(i)
		min, err := js.Get("standupprobability").Get(str).Get("min").Int()
		if err != nil {
			break
		}
		max, _ := js.Get("standupprobability").Get(str).Get("max").Int()
		probability, _ := js.Get("standupprobability").Get(str).Get("probability").Int()
		var rp RobotProbability
		rp.Min = min
		rp.Max = max
		rp.Probability = probability
		rc.StandUpProbability = append(rc.StandUpProbability, rp)
		i++
	}
}

func (r *Robot) OnGameMessage(subCmd int32, buffer []byte) {
	switch subCmd {
	case int32(BRTB.SendToClientMessageType_Status):
		{

			r.LastBetBSPlace = -1
			r.LastBetSDPlace = -1

			r.OnStatusMsg(buffer)
		}
		break
	}
}

func (r *Robot) Init(User player.RobotInterface, g table.TableHandler) {
	r.User = User
	r.GameLogic = g.(*Game)
}

func (r *Robot) OnStatusMsg(b []byte) {
	msg := &BRTB.StatusMessage{}
	proto.Unmarshal(b, msg)
	if msg.Status == int32(BRTB.GameStatus_BetStatus) {
		r.BetCount = 0
		r.AddBetTimer()
	} else if msg.Status == int32(BRTB.GameStatus_EndBetMovie) {
		if r.TimerJob != nil {
			r.TimerJob.Cancel()
			r.TimerJob = nil
		}
	}
}

func (r *Robot) RobotBet() {
	//	log.Tracef("RobotBet")
	//先决定下那边
	weight := rand.Intn(RConfig.TotalBetPlace)
	place := 0
	if weight < RConfig.BetPlace[0] {
		//0下大小  1下单双
		if rand.Intn(2) == 0 {
			place = r.ChoseBetRorB(Big, Single, r.LastBetBSPlace, BSLineID)
			r.LastBetBSPlace = place
		} else {
			place = r.ChoseBetRorB(Small, Double, r.LastBetBSPlace, SDLineID)
			r.LastBetSDPlace = place
		}

	} else if weight < RConfig.BetPlace[1]+RConfig.BetPlace[0] {
		//随机点数 4-17 14个区域
		place = r.ChoseBetMorePlace(0, coutLineID)
	} else {
		//随机下豹子18 -24 7区域位置
		place = r.ChoseBetMorePlace(1, WEILineID)
	}

	index := r.GetBetIndex()

	r.SendBetMsg(place, index)

	r.AddBetTimer()
}

//选择下红黑
/*
功能:选择下注区域
参数: BigOrSinglePlace大单或者SmallOrDoublePlace小双  最后一次下注区域lastbetplace lintindex 单双 1 大小2 点数0 豹子3
结果:上次有下注,则下返回上一次区域,else 根据平衡线下确定大小 单双下注区域
*/
func (r *Robot) ChoseBetRorB(BigOrSinglePlace int, SmallOrDoublePlace int, LastBetPlace int, lintindex int) int {
	if r.GameLogic.SenceSeat.IsSitDown(r.User) && LastBetPlace != -1 {
		return LastBetPlace
	}
	line := r.GameLogic.BetTotal[BigOrSinglePlace] - r.GameLogic.BetTotal[SmallOrDoublePlace]
	weight := RConfig.TotalLessLine
	var lo []int
	//当前差值大于平衡线
	if line < int64(-r.GameLogic.Rule.RobotLine[lintindex]) || line > int64(r.GameLogic.Rule.RobotLine[lintindex]) {
		lo = RConfig.OverLine
		weight = RConfig.TotalOverLine
	} else {
		lo = RConfig.LessLine
	}

	w := rand.Intn(weight)
	betplace := 0

	if w < lo[0] {
		//下少
		if line < 0 {
			//下红色
			betplace = BigOrSinglePlace
		} else {
			//下黑色
			betplace = SmallOrDoublePlace
		}
	} else {
		//下多
		if line < 0 {
			//下黑色
			betplace = SmallOrDoublePlace
		} else {
			//下红色
			betplace = BigOrSinglePlace
		}
	}

	//r.LastBetPlace = betplace
	return betplace
}

//获取下多少金币
func (r *Robot) GetBetIndex() int {
	betweight := 0
	if r.User.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*100) {
		betweight = 0
	} else if r.User.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*500) {
		betweight = 1
	} else if r.User.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*5000) {
		betweight = 2
	} else {
		betweight = 3
	}

	tmp := rand.Intn(RConfig.BetConfig[betweight].TotalWeight)

	for index, v := range RConfig.BetConfig[betweight].BetWeight {
		if tmp < v {
			return index
		}
		tmp -= v
	}

	return 0
}

func (r *Robot) SendBetMsg(place int, index int) {
	r.BetCount++
	msg := new(BRTB.Bet)
	msg.BetIndex = int32(index)
	msg.BetType = int32(place)
	//	log.Tracef("发送机器人下注")
	r.User.SendMsgToServer(int32(BRTB.ReceiveMessageType_BetID), msg)
}

func (r *Robot) AddBetTimer() {
	//达到限制次数了以后就不下注了
	if r.BetCount >= RConfig.Limit {
		if r.TimerJob != nil {
			r.TimerJob.Cancel()
			r.TimerJob = nil
		}
		return
	}

	t := rand.Intn((RConfig.BetTime[1] - RConfig.BetTime[0])) + RConfig.BetTime[0]
	r.TimerJob, _ = r.User.AddTimer(time.Duration(t), r.RobotBet)
}

//选择下豹子 点数区域
/*
功能:选择下注区域多区域 找出下注最多的区域和最小区域的差值 如果大于差值则下注区域除最大区域以外区域下注.下于随机下注
参数: types 0点数 1豹子
结果:
*/
func (r *Robot) ChoseBetMorePlace(types int, lintindex int) int {
	max := int64(0)
	min := int64(0)
	maxIndex := 0
	place := 0
	if types == 0 {
		for i := 4; i <= 17; i++ {
			if r.GameLogic.BetTotal[i] > max {
				max = r.GameLogic.BetTotal[i]
				maxIndex = i
			}
			if r.GameLogic.BetTotal[i] < min {
				min = r.GameLogic.BetTotal[i]
			}
		}
	} else {
		for i := 18; i <= 24; i++ {
			if r.GameLogic.BetTotal[i] > max {
				max = r.GameLogic.BetTotal[i]
				maxIndex = i
			}
			if r.GameLogic.BetTotal[i] < min {
				min = r.GameLogic.BetTotal[i]
			}
		}
	}

	line := max - min
	//当前差值大于平衡线
	if line < int64(-r.GameLogic.Rule.RobotLine[lintindex]) || line > int64(r.GameLogic.Rule.RobotLine[lintindex]) {
		if types == 0 {
			for i := 0; i <= 10; i++ {
				place = rand.Intn(14) + 4
				if place == maxIndex {
					continue
				}
				return place
			}
		} else {
			for i := 0; i <= 10; i++ {
				place = rand.Intn(7) + 18
				if place == maxIndex {
					continue
				}
				return place
			}
		}

	} else {
		if types == 0 {
			place = rand.Intn(14) + 4
			return place
		} else {
			place = rand.Intn(7) + 18
			return place
		}

	}
	return place

}
