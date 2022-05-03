package game

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"

	"github.com/bitly/go-simplejson"
	"github.com/golang/protobuf/proto"
	rbwar "github.com/kubegames/kubegames-games/pkg/hundreds/960301/pb"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/sipt/GoJsoner"
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
	Line               int                //平衡线，少于这个线和多余这个线使用不同的配置
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
	Robot        player.RobotInterface
	GameLogic    *Game
	BetCount     int         //下注限制
	TimerJob     *player.Job //job
	LastBetPlace int         //机器人上次下注的区域
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
	rc.Line, _ = js.Get("line").Int()
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

func (rc *RobotConfig) getBetTime(js *simplejson.Json) {
	arr, _ := js.Get("bettime").Array()
	for n := 0; n < len(arr); n++ {
		bettime, _ := arr[n].(json.Number).Int64()
		rc.BetTime = append(rc.BetTime, int(bettime))
	}
}

func (rc *RobotConfig) getBetPro(js *simplejson.Json) {
	arr, _ := js.Get("betpro").Array()
	for n := 0; n < len(arr); n++ {
		betpro, _ := arr[n].(json.Number).Int64()
		rc.BetPro = append(rc.BetPro, int(betpro))
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
	case int32(rbwar.SendToClientMessageType_Status):
		{
			r.LastBetPlace = -1
			r.OnStatusMsg(buffer)
		}
		break
	}
}

func (r *Robot) Init(robot player.RobotInterface, g *Game) {
	r.Robot = robot
	r.GameLogic = g
}

func (r *Robot) OnStatusMsg(b []byte) {
	msg := &rbwar.StatusMessage{}
	proto.Unmarshal(b, msg)
	if msg.Status == int32(rbwar.GameStatus_BetStatus) {
		r.BetCount = 0
		r.AddBetTimer()
	} else if msg.Status == int32(rbwar.GameStatus_EndBetMovie) {
		if r.TimerJob != nil {
			r.Robot.DeleteJob(r.TimerJob)
			r.TimerJob = nil
		}
	}
}

func (r *Robot) RobotBet() {
	//先决定下那边
	weight := rand.Intn(RConfig.TotalBetPlace)
	place := 0
	if weight < RConfig.BetPlace[0] {
		//下红或者黑
		place = r.ChoseBetRorB()
	} else {
		//下幸运一击
		place = ONEHIT
	}

	index := r.GetBetIndex()

	r.SendBetMsg(place, index)

	r.AddBetTimer()
}

//选择下红黑
func (r *Robot) ChoseBetRorB() int {
	if r.GameLogic.SenceSeat.IsSitDown(r.Robot) && r.LastBetPlace != -1 {
		return r.LastBetPlace
	}

	line := r.GameLogic.BetTotal[RED] - r.GameLogic.BetTotal[BLACK]
	weight := RConfig.TotalLessLine
	var lo []int
	//当前差值大于平衡线
	if line < int64(-RConfig.Line) || line > int64(RConfig.Line) {
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
			betplace = RED
		} else {
			//下黑色
			betplace = BLACK
		}
	} else {
		//下多
		if line < 0 {
			//下黑色
			betplace = BLACK
		} else {
			//下红色
			betplace = RED
		}
	}

	r.LastBetPlace = betplace
	return betplace
}

//获取下多少金币
func (r *Robot) GetBetIndex() int {
	betweight := 0
	if r.Robot.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*100) {
		betweight = 0
	} else if r.Robot.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*500) {
		betweight = 1
	} else if r.Robot.GetScore() <= int64(r.GameLogic.Rule.BetList[0]*5000) {
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
	msg := new(rbwar.Bet)
	msg.BetIndex = int32(index)
	msg.BetType = int32(place)
	r.Robot.SendMsgToServer(int32(rbwar.ReceiveMessageType_BetID), msg)
}

func (r *Robot) AddBetTimer() {
	//达到限制次数了以后就不下注了
	if r.BetCount >= RConfig.Limit {
		if r.TimerJob != nil {
			r.Robot.DeleteJob(r.TimerJob)
			r.TimerJob = nil
		}
		return
	}

	//不重复添加
	t := rand.Intn((RConfig.BetTime[1] - RConfig.BetTime[0])) + RConfig.BetTime[0]
	r.TimerJob, _ = r.Robot.AddTimer(int64(t), r.RobotBet)
}
