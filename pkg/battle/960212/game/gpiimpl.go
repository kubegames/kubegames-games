package game

import (
	"fmt"
	"game_poker/doudizhu/msg"
	"reflect"

	"github.com/kubegames/kubegames-sdk/pkg/log"

	"github.com/antonmedv/expr"
	b3 "github.com/magicsea/behavior3go"
	b3cfg "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
	b3loader "github.com/magicsea/behavior3go/loader"
)

var cfg *b3cfg.BTProjectCfg

func ErrorCheck(cond bool, err int32, logargs ...interface{}) {
	if !cond {
		log.Traceln(logargs...)
		//panic(err)
	}
}

func B3cfgInit() {
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	log.Traceln(err)
	//}
	//log.Traceln(dir)

	if config, ok := b3cfg.LoadProjectCfg("config/ddz.json"); ok {
		cfg = config
	} else {
		ErrorCheck(false, 1, "load b tree is failed")
	}
}

//CreateBehaviorTree 创建行为树
func CreateBehaviorTree() *b3core.BehaviorTree {
	B3cfgInit()
	maps := b3.NewRegisterStructMaps()
	maps.Register("Wait", new(Wait))
	maps.Register("isInRange", new(IsInRange))
	maps.Register("isEqual", new(IsEqual))
	maps.Register("isMore", new(IsMore))
	maps.Register("isLess", new(IsLess))
	maps.Register("callbanker", new(callbanker))
	maps.Register("robbanker", new(robbanker))
	maps.Register("makedouble", new(makedouble))
	maps.Register("passcard", new(passcard))
	maps.Register("outcard", new(outcard))
	maps.Register("outAllCard", new(outAllcard))
	maps.Register("outAbsBig", new(outAbsBig))
	maps.Register("gencard", new(gencard))
	maps.Register("outSingle", new(outSingle))
	maps.Register("outMinSingle", new(outMinSingle))
	maps.Register("outNotSingle", new(outNotSingle))
	maps.Register("outDouble", new(outDouble))
	maps.Register("outMinDouble", new(outMinDouble))
	maps.Register("outNotDouble", new(outNotDouble))
	maps.Register("outBomb", new(outBomb))
	maps.Register("outBombMax", new(outBombMax))
	maps.Register("bankerGenPrevOut", new(bankerGenPrevOut))
	maps.Register("bankerGenNextOut", new(bankerGenNextOut))
	maps.Register("prevGenBankerOut", new(prevGenBankerOut))
	maps.Register("prevGenNextOut", new(prevGenNextOut))
	maps.Register("nextGenBankerOut", new(nextGenBankerOut))
	maps.Register("nextGenPrevOut", new(nextGenPrevOut))
	maps.Register("out4With2DoubleOr2Single", new(out4With2DoubleOr2Single))
	maps.Register("outCardToStopBanker", new(outCardToStopBanker))
	maps.Register("IsContain", new(IsContain))
	maps.Register("doOutCard", new(doOutCard))
	maps.Register("outThePairForSigle", new(outThePairForSigle))
	maps.Register("OutAbsSmall", new(OutAbsSmall))
	maps.Register("outNotMinCard", new(outNotMinCardType))
	maps.Register("outNotSigleAndDouble", new(outNotSigleAndDouble))
	maps.Register("outMiddleCountSigle", new(outMiddleCountSigle))
	maps.Register("outCardByType", new(outCardByType))
	var firstTree *b3core.BehaviorTree
	//载入
	for _, v := range cfg.Trees {
		tree := b3loader.CreateBevTreeFromConfig(&v, maps)
		if firstTree == nil {
			firstTree = tree
		}
	}
	//firstTree.Print()
	//log.Debugf("------ %v", firstTree.Print())
	return firstTree
}
func (cab *IsContain) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.name = setting.GetPropertyAsString("arrayName")
	cab.v = setting.GetPropertyAsInt("v")
}
func (cab *IsContain) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	m := reflect.ValueOf(*robot.exportData)
	v := m.FieldByName(cab.name)
	switch v.Type().Elem().Kind() {
	case reflect.Int:
		a := v.Interface().([]int)
		for _, v := range a {
			if v == cab.v {
				return b3.SUCCESS
			}
		}
	}

	log.Debugf("failed")
	return b3.FAILURE
}

func (cab *doOutCard) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.groupType = setting.GetPropertyAsInt("t")
}
func (cab *doOutCard) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	if robot.doOut(cab.groupType) {
		return b3.SUCCESS
	} else {
		log.Debugf("failed")
		return b3.FAILURE
	}
}
func (cab *callbanker) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.call = GetProperty("call", setting) //GetValueAsInt(setting, "calltype", 0)

	ErrorCheck(nil != cab.call, 1, "invalid b tree cfg")
	var ec error
	cab.Program, ec = expr.Compile(fmt.Sprintf("%v", cab.call), expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
}

func (cab *callbanker) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	exportData := robot.GetExportData()
	ret, ec := expr.Run(cab.Program, exportData)
	if nil != ec {
		log.Warnf("run expr is failed, ec = %v", ec)
		log.Debugf("failed")
		return b3.FAILURE
	}
	call := ret.(bool)
	log.Debugf("robot[%d] callbanker 【%v】 !!!", robot.mySeat, call)
	robot.doCallBanker(call)
	return b3.SUCCESS
}
func (cab *robbanker) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.rob = GetProperty("rob", setting) //GetValueAsInt(setting, "calltype", 0)

	ErrorCheck(nil != cab.rob, 1, "invalid b tree cfg")
	var ec error
	cab.Program, ec = expr.Compile(fmt.Sprintf("%v", cab.rob), expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
}

func (cab *robbanker) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	exportData := robot.GetExportData()
	ret, ec := expr.Run(cab.Program, exportData)
	if nil != ec {
		log.Warnf("run expr is failed, ec = ", ec)
		log.Debugf("failed")
		return b3.FAILURE
	}
	rob := ret.(bool)
	log.Debugf("robot[%d] robbanker 【%v】 !!!", robot.mySeat, rob)
	robot.doRobBanker(rob)
	return b3.SUCCESS
}

func (mkd *makedouble) Initialize(setting *b3cfg.BTNodeCfg) {
	mkd.Action.Initialize(setting)
	//mkd.t = GetValueAsInt(setting, "type", 0)

	mkd.t = GetProperty("type", setting) //GetValueAsInt(setting, "calltype", 0)

	ErrorCheck(nil != mkd.t, 1, "invalid b tree cfg")
	var ec error
	mkd.Program, ec = expr.Compile(fmt.Sprintf("%v", mkd.t), expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
}

func (mkd *makedouble) OnTick(tick *b3core.Tick) b3.Status {
	//robot := tick.GetTarget().(*robot)
	//robot.doDouble(mkd.t)

	robot := tick.GetTarget().(*robot)
	exportData := robot.GetExportData()
	ret, ec := expr.Run(mkd.Program, exportData)
	if nil != ec {
		log.Warnf("run expr is failed, ec = ", ec)
		log.Debugf("failed")
		return b3.FAILURE
	}
	rob := ret.(int)
	log.Debugf("robot[%d] makedouble 【%v】 !!!", robot.mySeat, rob)
	robot.doDouble(rob)

	return b3.SUCCESS
}

func (pass *passcard) Initialize(setting *b3cfg.BTNodeCfg) {
	pass.Action.Initialize(setting)
}

func (pass *passcard) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] passcard !!!", robot.mySeat)
	robot.doPassCard()
	return b3.SUCCESS
}

func (out *outcard) Initialize(setting *b3cfg.BTNodeCfg) {
	out.Action.Initialize(setting)
}

func (out *outcard) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outcard !!!", robot.mySeat)
	robot.doOutCard()
	return b3.SUCCESS
}

func (outa *outAllcard) Initialize(setting *b3cfg.BTNodeCfg) {
	outa.Action.Initialize(setting)
}

func (outa *outAllcard) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outAllcard !!!", robot.mySeat)
	robot.doOutAllCard()
	return b3.SUCCESS
}

func (outb *outAbsBig) Initialize(setting *b3cfg.BTNodeCfg) {
	outb.Action.Initialize(setting)
}

func (outb *outAbsBig) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outAbsBig !!!", robot.mySeat)
	robot.doOutAbsBig()
	return b3.SUCCESS
}

func (outs *outSingle) Initialize(setting *b3cfg.BTNodeCfg) {
	outs.Action.Initialize(setting)
	outs.minNumber = GetProperty("minNumber", setting) //GetValueAsInt(setting, "minNumber", 0)
	outs.maxNumber = GetProperty("maxNumber", setting) //GetValueAsInt(setting, "maxNumber", 0)
	outs.isFromBig = GetProperty("isFromBig", setting) //GetValueAsBool(setting, "isFromBig", false)
	outs.isForce = GetProperty("isForce", setting)     //GetValueAsBool(setting, "isForce", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, outs.minNumber, outs.maxNumber)
	var ec error
	outs.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(outs.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (outs *outSingle) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outSingle !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", outs.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outs.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outs.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outs.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	robot.doOutSingle(minNumber, maxNumber, isFromBig, isForce)
	return b3.SUCCESS
}

func (oms *outMinSingle) Initialize(setting *b3cfg.BTNodeCfg) {
	oms.Action.Initialize(setting)
}

func (oms *outMinSingle) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outMinSingle !!!", robot.mySeat)
	robot.doOutMinSingle()
	return b3.SUCCESS
}

func (nots *outNotSingle) Initialize(setting *b3cfg.BTNodeCfg) {
	nots.Action.Initialize(setting)
	nots.minNumber = GetProperty("minNumber", setting) //GetValueAsInt(setting, "minNumber", 0)
	nots.maxNumber = GetProperty("maxNumber", setting) //GetValueAsInt(setting, "maxNumber", 0)
	nots.isFromBig = GetProperty("isFromBig", setting) //GetValueAsBool(setting, "isFromBig", false)
	nots.isForce = GetProperty("isForce", setting)     //GetValueAsBool(setting, "isForce", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, nots.minNumber, nots.maxNumber)
	var ec error
	nots.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(nots.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (nots *outNotSingle) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outNotSingle !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", nots.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", nots.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", nots.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", nots.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	robot.doOutNotSingle(minNumber, maxNumber, isFromBig, isForce)
	return b3.SUCCESS
}

func (outd *outDouble) Initialize(setting *b3cfg.BTNodeCfg) {
	outd.Action.Initialize(setting)
	outd.minNumber = GetProperty("minNumber", setting) //GetValueAsInt(setting, "minNumber", 0)
	outd.maxNumber = GetProperty("maxNumber", setting) //GetValueAsInt(setting, "maxNumber", 0)
	outd.isFromBig = GetProperty("isFromBig", setting) //GetValueAsBool(setting, "isFromBig", false)
	outd.isForce = GetProperty("isForce", setting)     //GetValueAsBool(setting, "isForce", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, outd.minNumber, outd.maxNumber)
	var ec error
	outd.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(outd.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (outd *outDouble) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outDouble !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", outd.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outd.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outd.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", outd.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	robot.doOutDouble(minNumber, maxNumber, isFromBig, isForce)
	return b3.SUCCESS
}

func (omd *outMinDouble) Initialize(setting *b3cfg.BTNodeCfg) {
	omd.Action.Initialize(setting)
}

func (omd *outMinDouble) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outMinDouble !!!", robot.mySeat)
	robot.doOutMinDouble()
	return b3.SUCCESS
}

func (notd *outNotDouble) Initialize(setting *b3cfg.BTNodeCfg) {
	notd.Action.Initialize(setting)
	notd.minNumber = GetProperty("minNumber", setting) //GetValueAsInt(setting, "minNumber", 0)
	notd.maxNumber = GetProperty("maxNumber", setting) //GetValueAsInt(setting, "maxNumber", 0)
	notd.isFromBig = GetProperty("isFromBig", setting) //GetValueAsBool(setting, "isFromBig", false)
	notd.isForce = GetProperty("isForce", setting)     //GetValueAsBool(setting, "isForce", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, notd.minNumber, notd.maxNumber)
	var ec error
	notd.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(notd.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (notd *outNotDouble) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outNotDouble !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", notd.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", notd.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", notd.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", notd.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	robot.doOutNotDouble(minNumber, maxNumber, isFromBig, isForce)
	return b3.SUCCESS
}

func (gcd *gencard) Initialize(setting *b3cfg.BTNodeCfg) {
	gcd.Action.Initialize(setting)
	gcd.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	gcd.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	gcd.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	gcd.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	gcd.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	gcd.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, gcd.minNumber, gcd.maxNumber)
	var ec error
	gcd.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(gcd.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (gcd *gencard) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] gencard !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", gcd.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", gcd.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", gcd.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", gcd.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", gcd.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", gcd.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doGenCard(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (outb *outBomb) Initialize(setting *b3cfg.BTNodeCfg) {
	outb.Action.Initialize(setting)
}

func (outb *outBomb) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outBomb !!!", robot.mySeat)
	robot.doOutBomb()
	return b3.SUCCESS
}

func (outb *outBombMax) Initialize(setting *b3cfg.BTNodeCfg) {
	outb.Action.Initialize(setting)
}

func (outb *outBombMax) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outBomb !!!", robot.mySeat)
	robot.doOutBombMax()
	return b3.SUCCESS
}

func (outb *out4With2DoubleOr2Single) Initialize(setting *b3cfg.BTNodeCfg) {
	outb.Action.Initialize(setting)
}

func (outb *out4With2DoubleOr2Single) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] outBomb !!!", robot.mySeat)
	robot.doOut4With2DoubleOr2Single()
	return b3.SUCCESS
}

func (outb *outCardToStopBanker) Initialize(setting *b3cfg.BTNodeCfg) {
	outb.Action.Initialize(setting)
}

func (outb *outCardToStopBanker) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] stopBanker !!!", robot.mySeat)

	robot.doOutCardToStopBanker(robot.getOutCardPriorityGroups1())
	return b3.SUCCESS
}
func (bgp *bankerGenPrevOut) Initialize(setting *b3cfg.BTNodeCfg) {
	bgp.Action.Initialize(setting)
	bgp.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	bgp.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	bgp.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	bgp.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	bgp.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	bgp.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, bgp.minNumber, bgp.maxNumber)
	var ec error
	bgp.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(bgp.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (bgp *bankerGenPrevOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] bankerGenPrevOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", bgp.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgp.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgp.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgp.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgp.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgp.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doBankerGenPrevOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (bgn *bankerGenNextOut) Initialize(setting *b3cfg.BTNodeCfg) {
	bgn.Action.Initialize(setting)
	bgn.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	bgn.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	bgn.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	bgn.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	bgn.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	bgn.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, bgn.minNumber, bgn.maxNumber)
	var ec error
	bgn.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(bgn.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (bgn *bankerGenNextOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] bankerGenNextOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", bgn.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgn.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgn.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgn.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgn.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", bgn.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doBankerGenNextOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (pgb *prevGenBankerOut) Initialize(setting *b3cfg.BTNodeCfg) {
	pgb.Action.Initialize(setting)
	pgb.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	pgb.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	pgb.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	pgb.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	pgb.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	pgb.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, pgb.minNumber, pgb.maxNumber)
	var ec error
	pgb.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(pgb.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (pgb *prevGenBankerOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] prevGenBankerOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", pgb.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgb.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgb.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgb.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgb.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgb.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doPrevGenBankerOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (pgn *prevGenNextOut) Initialize(setting *b3cfg.BTNodeCfg) {
	pgn.Action.Initialize(setting)
	pgn.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	pgn.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	pgn.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	pgn.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	pgn.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	pgn.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, pgn.minNumber, pgn.maxNumber)
	var ec error
	pgn.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(pgn.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (pgn *prevGenNextOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] prevGenNextOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", pgn.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgn.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgn.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgn.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgn.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", pgn.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doPrevGenNextOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (ngb *nextGenBankerOut) Initialize(setting *b3cfg.BTNodeCfg) {
	ngb.Action.Initialize(setting)
	ngb.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	ngb.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	ngb.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	ngb.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	ngb.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	ngb.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, ngb.minNumber, ngb.maxNumber)
	var ec error
	ngb.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(ngb.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (ngb *nextGenBankerOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] nextGenBankerOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", ngb.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngb.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngb.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngb.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngb.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngb.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doNextGenBankerOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (ngp *nextGenPrevOut) Initialize(setting *b3cfg.BTNodeCfg) {
	ngp.Action.Initialize(setting)
	ngp.minNumber = GetProperty("minNumber", setting)     //GetValueAsInt(setting, "minNumber", 0)
	ngp.maxNumber = GetProperty("maxNumber", setting)     //GetValueAsInt(setting, "maxNumber", 0)
	ngp.isFromBig = GetProperty("isFromBig", setting)     //GetValueAsBool(setting, "isFromBig", false)
	ngp.isForce = GetProperty("isForce", setting)         //GetValueAsBool(setting, "isForce", false)
	ngp.isUseBomb = GetProperty("isUseBomb", setting)     //GetValueAsBool(setting, "isUseBomb", false)
	ngp.isUseRocket = GetProperty("isUseRocket", setting) //GetValueAsBool(setting, "isUseRocket", false)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, ngp.minNumber, ngp.maxNumber)
	var ec error
	ngp.Program, ec = expr.Compile(express, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	var ret interface{}
	ret, ec = expr.Run(ngp.Program, expr.Env(ExportData{}))
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	ErrorCheck(ret.(bool), 1, "invalid b tree cfg")
}

func (ngp *nextGenPrevOut) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	log.Debugf("robot[%d] nextGenPrevOut !!!", robot.mySeat)
	temp, ec := expr.Eval(fmt.Sprintf("%v", ngp.minNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	minNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngp.maxNumber), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	maxNumber := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngp.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngp.isForce), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isForce := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngp.isUseBomb), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseBomb := temp.(bool)
	temp, ec = expr.Eval(fmt.Sprintf("%v", ngp.isUseRocket), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isUseRocket := temp.(bool)
	robot.doNextGenPrevOut(minNumber, maxNumber, isFromBig, isForce, isUseBomb, isUseRocket)
	return b3.SUCCESS
}

func (cab *outThePairForSigle) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.isFromBig = GetProperty("isFromBig", setting)
}
func (cab *outThePairForSigle) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	temp, ec := expr.Eval(fmt.Sprintf("%v", cab.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)

	if robot.exportData.IsBanker {
		next := (robot.User.ChairID + 1) % 3
		prev := (next + 1) % 3
		find := int32(-1)
		if len(robot.GameLogic.Chairs[next].Cards) == 2 {
			find = next
		} else if len(robot.GameLogic.Chairs[prev].Cards) == 2 {
			find = prev
		}

		if find != -1 {
			if robot.GameLogic.Chairs[prev].Cards[0]>>4 == robot.GameLogic.Chairs[prev].Cards[1]>>4 {
				for _, group := range robot.cardGroups {
					if group.Type == msg.CardsType_Pair {
						if group.Cards[0] < robot.GameLogic.Chairs[prev].Cards[0] {
							if isFromBig {
								outcard := ChaiLargePairs(robot.cardGroups)
								if outcard != 0 {
									robot.sendOutCardmsg(1, []byte{outcard})
									return b3.SUCCESS
								}
							} else {
								outcard := ChaiLessPairs(robot.cardGroups)
								if outcard != 0 {
									robot.sendOutCardmsg(1, []byte{outcard})
									return b3.SUCCESS
								}
							}

							break
						}
					}
				}
			}
		}
	} else if len(robot.GameLogic.Dizhu.Cards) == 2 {
		if robot.GameLogic.Dizhu.Cards[0]>>4 == robot.GameLogic.Dizhu.Cards[1]>>4 {
			for _, group := range robot.cardGroups {
				if group.Type == msg.CardsType_Pair {
					if group.Cards[0] < robot.GameLogic.Dizhu.Cards[0] {
						if isFromBig {
							outcard := ChaiLargePairs(robot.cardGroups)
							if outcard != 0 {
								robot.sendOutCardmsg(1, []byte{outcard})
								return b3.SUCCESS
							}
						} else {
							outcard := ChaiLessPairs(robot.cardGroups)
							if outcard != 0 {
								robot.sendOutCardmsg(1, []byte{outcard})
								return b3.SUCCESS
							}
						}

						break
					}
				}
			}
		}
	}
	log.Debugf("failed")
	return b3.FAILURE
}

func (cab *OutAbsSmall) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
}
func (cab *OutAbsSmall) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	temp := robot.CheckAllOff()
	if len(temp) > 0 {
		robot.sendOutCardmsg(1, temp[0])
		return b3.SUCCESS
	}

	log.Debugf("outNotMinCardType FAILURE")
	return b3.FAILURE
}

func (cab *outNotMinCardType) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.isFromBig = GetProperty("isFromBig", setting)
}
func (cab *outNotMinCardType) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	temp, ec := expr.Eval(fmt.Sprintf("%v", cab.isFromBig), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	isFromBig := temp.(bool)

	if robot.exportData.IsBanker {
		next := (robot.User.ChairID + 1) % 3
		prev := (next + 1) % 3
		find := int32(-1)
		if len(robot.GameLogic.Chairs[next].Cards) <= 2 {
			find = next
		} else if len(robot.GameLogic.Chairs[prev].Cards) <= 2 {
			find = prev
		}

		if find != -1 {
			if IsCanAbsWin(robot.cardGroups, robot.GameLogic.Chairs[find].Cards, false) {
				_, cardgroup := GetMostValueGroup(robot.GameLogic.Chairs[find].Cards)
				if isFromBig {
					group := SearchFirstLargeGroup(robot.cardGroups, cardgroup[0].Cards, false)
					if group != nil {
						robot.sendOutCardmsg(1, group.Cards)
						return b3.SUCCESS
					}
				} else {
					group := SearchFirstLargeGroup(robot.cardGroups, cardgroup[0].Cards, false)
					if group != nil {
						robot.sendOutCardmsg(1, group.Cards)
						return b3.SUCCESS
					}
				}
			}
		}
	} else if len(robot.GameLogic.Dizhu.Cards) <= 2 {
		if IsCanAbsWin(robot.cardGroups, robot.GameLogic.Dizhu.Cards, false) {
			_, cardgroup := GetMostValueGroup(robot.GameLogic.Dizhu.Cards)
			if isFromBig {
				group := SearchFirstLargeGroup(robot.cardGroups, cardgroup[0].Cards, false)
				if group != nil {
					robot.sendOutCardmsg(1, group.Cards)
					return b3.SUCCESS
				}
			} else {
				group := SearchFirstLargeGroup(robot.cardGroups, cardgroup[0].Cards, false)
				if group != nil {
					robot.sendOutCardmsg(1, group.Cards)
					return b3.SUCCESS
				}
			}
		}
	}

	log.Debugf("outNotMinCardType FAILURE")
	return b3.FAILURE
}

func (cab *outNotSigleAndDouble) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
}
func (cab *outNotSigleAndDouble) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	if robot.exportData.IsBanker {
		next := (robot.User.ChairID + 1) % 3
		prev := (next + 1) % 3
		find := int32(-1)
		if len(robot.GameLogic.Chairs[next].Cards) == 2 {
			find = next
		} else if len(robot.GameLogic.Chairs[prev].Cards) == 2 {
			find = prev
		}

		if find != -1 {
			if robot.GameLogic.Chairs[prev].Cards[0]>>4 != robot.GameLogic.Chairs[prev].Cards[1]>>4 {
				for _, group := range robot.cardGroups {
					if group.Type >= msg.CardsType_Bomb {
						continue
					}

					if group.Type > msg.CardsType_Pair {
						if group.Cards[0] < robot.GameLogic.Chairs[prev].Cards[0] {
							robot.sendOutCardmsg(1, group.Cards)
						}
					}
				}
			}
		}
	} else if len(robot.GameLogic.Dizhu.Cards) == 2 {
		if robot.GameLogic.Dizhu.Cards[0]>>4 != robot.GameLogic.Dizhu.Cards[1]>>4 {
			for _, group := range robot.cardGroups {
				if group.Type >= msg.CardsType_Bomb {
					continue
				}

				if group.Type > msg.CardsType_Pair {
					if group.Cards[0] < robot.GameLogic.Dizhu.Cards[0] {
						robot.sendOutCardmsg(1, group.Cards)
						return b3.SUCCESS
					}
				}
			}
		}
	}

	log.Debugf("outNotSigleAndDouble FAILURE")
	return b3.FAILURE
}

func (cab *outMiddleCountSigle) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
}
func (cab *outMiddleCountSigle) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	if robot.exportData.IsBanker {
		next := (robot.User.ChairID + 1) % 3
		prev := (next + 1) % 3
		find := int32(-1)
		if len(robot.GameLogic.Chairs[next].Cards) == 2 {
			find = next
		} else if len(robot.GameLogic.Chairs[prev].Cards) == 2 {
			find = prev
		}

		if find != -1 {
			count := [2]int{}
			index := -1
			if robot.GameLogic.Chairs[prev].Cards[0]>>4 != robot.GameLogic.Chairs[prev].Cards[1]>>4 {
				for i, group := range robot.cardGroups {
					if group.Type == msg.CardsType_SingleCard {
						if group.Cards[0] < robot.GameLogic.Chairs[prev].Cards[0] {
							count[0]++
						} else if group.Cards[0] < robot.GameLogic.Chairs[prev].Cards[1] {
							count[1]++
							if index == -1 {
								index = i
							}
						}
					}
				}

				if count[0] != 0 && count[1] != 0 {
					robot.sendOutCardmsg(1, robot.cardGroups[index].Cards)
					return b3.SUCCESS
				}
			}
		}
	} else if len(robot.GameLogic.Dizhu.Cards) == 2 {
		if robot.GameLogic.Dizhu.Cards[0]>>4 != robot.GameLogic.Dizhu.Cards[1]>>4 {
			count := [2]int{}
			index := -1
			for i, group := range robot.cardGroups {
				if group.Type == msg.CardsType_SingleCard {
					if group.Cards[0] < robot.GameLogic.Dizhu.Cards[0] {
						count[0]++
					} else if group.Cards[0] < robot.GameLogic.Dizhu.Cards[1] {
						count[1]++
						if index == -1 {
							index = i
						}
					}
				}
			}

			if count[0] != 0 && count[1] != 0 {
				robot.sendOutCardmsg(1, robot.cardGroups[index].Cards)
				return b3.SUCCESS
			}
		}
	}
	log.Debugf("outMiddleCountSigle FAILURE")
	return b3.FAILURE
}

func (cab *outCardByType) Initialize(setting *b3cfg.BTNodeCfg) {
	cab.Action.Initialize(setting)
	cab.BigToSmall = GetProperty("BigToSmall", setting)
}

func (cab *outCardByType) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	temp, ec := expr.Eval(fmt.Sprintf("%v", cab.BigToSmall), nil)
	ErrorCheck(nil == ec, 1, "invalid b tree cfg")
	BigToSmall := temp.(bool)

	var group *CardGroup
	if BigToSmall {
		group = SearchFirstLargeGroup(robot.cardGroups, robot.GameLogic.CurrentCards.Cards, false)
	} else {
		group = SearchLastLargeGroup(robot.cardGroups, robot.GameLogic.CurrentCards.Cards, false)
	}

	if group != nil {
		robot.sendOutCardmsg(1, group.Cards)
		return b3.SUCCESS
	}

	robot.doPassCard()
	log.Debugf("outCardByType FAILURE")
	return b3.FAILURE
}
