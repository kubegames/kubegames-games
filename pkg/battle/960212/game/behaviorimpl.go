package game

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/antonmedv/expr"
	"github.com/kubegames/kubegames-games/internal/pkg/rand"
	"github.com/kubegames/kubegames-sdk/pkg/log"
	b3 "github.com/magicsea/behavior3go"
	b3cfg "github.com/magicsea/behavior3go/config"
	b3core "github.com/magicsea/behavior3go/core"
)

//临时机器人，在初始化的时候传进来，这样IsInRange、IsEqual这些节点才能捕获不同游戏所导出的变量，该变量只有在编译表达式的时候才有用
var tempRobot robot

func fillEnv(env map[string]interface{}, robot robot) map[string]interface{} {
	if nil == env {
		env = make(map[string]interface{})
	}
	object := reflect.ValueOf(robot.GetExportData())
	ref := object.Elem()
	typeOfType := ref.Type()
	for i := 0; i < ref.NumField(); i++ {
		field := ref.Field(i)
		//log.Traceln(typeOfType.Field(i).Name)
		env[typeOfType.Field(i).Name] = field.Interface()
	}
	return env
}

//屎山函数，求优化
func GetProperty(key string, setting *b3cfg.BTNodeCfg) interface{} {
	//errors.Catch()函数会打印大量的堆栈，这里单独定义个函数捕获异常
	catchFn := func(err *error) {
		if ec := recover(); nil != ec {
			if nil != err {
				switch ec.(type) {
				case error:
					temp := ec.(error)
					*err = temp
				case string:
					temp := ec.(string)
					*err = fmt.Errorf(temp)
				default:
					log.Errorf("unkown type exception, ", ec)
				}
			}
		}
	}

	floatFn := func() (ret interface{}, err error) {
		defer catchFn(&err)
		value := setting.GetProperty(key)
		return value, nil
	}
	boolFn := func() (ret interface{}, err error) {
		defer catchFn(&err)
		value := setting.GetPropertyAsBool(key)
		return value, nil
	}
	intFn := func() (ret interface{}, err error) {
		defer catchFn(&err)
		value := setting.GetPropertyAsInt(key)
		return value, nil
	}
	int64Fn := func() (ret interface{}, err error) {
		defer catchFn(&err)
		value := setting.GetPropertyAsInt64(key)
		return value, nil
	}
	stringFn := func() (ret interface{}, err error) {
		defer catchFn(&err)
		value := setting.GetPropertyAsString(key)
		value = strings.Replace(value, "$", "", -1)
		return value, nil
	}

	if value, ec := floatFn(); nil == ec {
		return value
	}
	if value, ec := intFn(); nil == ec {
		return value
	}
	if value, ec := int64Fn(); nil == ec {
		return value
	}
	if value, ec := stringFn(); nil == ec {
		return value
	}
	if value, ec := boolFn(); nil == ec {
		return value
	}

	log.Errorf("can not parse b tree param")
	return nil
}

func GetValueAsFloat32(cfg *b3cfg.BTNodeCfg, name string, defValue float32) float32 {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	ret, fok := v.(float32)
	if !fok {
		return defValue
	}
	return ret
}

func GetValueAsFloat64(cfg *b3cfg.BTNodeCfg, name string, defValue float64) float64 {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	ret, fok := v.(float64)
	if !fok {
		return defValue
	}
	return ret
}

func GetValueAsInt(cfg *b3cfg.BTNodeCfg, name string, defValue int) int {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	f64, fok := v.(float64)
	if !fok {
		return defValue
	}
	return int(f64)
}
func GetValueAsInt64(cfg *b3cfg.BTNodeCfg, name string, defValue int64) int64 {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	f64, fok := v.(float64)
	if !fok {
		return defValue
	}
	return int64(f64)
}
func GetValueAsBool(cfg *b3cfg.BTNodeCfg, name string, defValue bool) bool {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	ret, fok := v.(bool)
	if !fok {
		if str, sok := v.(string); sok {
			return str == "true"
		}
		return defValue
	}
	return ret
}
func GetValueAsString(cfg *b3cfg.BTNodeCfg, name string, defValue string) string {
	v, ok := cfg.Properties[name]
	if !ok {
		return defValue
	}
	ret, fok := v.(string)
	if !fok {
		return defValue
	}
	return ret
}

func (inRan *IsInRange) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		LEFT  = "left"
		RIGHT = "right"
		VALUE = "value"
	)

	inRan.Action.Initialize(setting)
	inRan.env = fillEnv(inRan.env, tempRobot)
	inRan.env[LEFT] = GetProperty(LEFT, setting)
	inRan.env[RIGHT] = GetProperty(RIGHT, setting)
	inRan.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	inRan.express = fmt.Sprintf(`(%v)>=(%v) && (%v)<=(%v)`, inRan.env[VALUE], inRan.env[LEFT], inRan.env[VALUE], inRan.env[RIGHT])
	var ec error
	inRan.Program, ec = expr.Compile(inRan.express, expr.Env(inRan.env))
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(inRan.Program, inRan.env)
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}
}

func (inRan *IsInRange) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	inRan.env = fillEnv(inRan.env, *robot)
	//log.Debugf("isInRange tick, all var = %v", inRan.env)
	ret, ec := expr.Run(inRan.Program, inRan.env)
	if nil != ec {
		log.Warnf("run expr is failed, ec = ", ec)
		return b3.FAILURE
	}

	realRet := ret.(bool)
	if realRet {
		return b3.SUCCESS
	} else {
		return b3.FAILURE
	}
}

func (wa *Wait) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		LEFT  = "left"
		RIGHT = "right"
	)

	wa.Action.Initialize(setting)
	wa.env = fillEnv(wa.env, tempRobot)
	wa.env[LEFT] = GetProperty(LEFT, setting)
	wa.env[RIGHT] = GetProperty(RIGHT, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v)>=0 && (%v)>=(%v)`, wa.env[LEFT], wa.env[RIGHT], wa.env[LEFT])
	var ec error
	wa.Program, ec = expr.Compile(express, expr.Env(wa.env))
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}

	//运行下，如果有语法错误会马上报出来
	_, ec = expr.Run(wa.Program, wa.env)
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}

	wa.left = wa.env[LEFT]
	wa.right = wa.env[RIGHT]
}

func (wa *Wait) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	wa.env = fillEnv(wa.env, *robot)

	temp, ec := expr.Eval(fmt.Sprintf("%v", wa.left), wa.env)
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}
	left := temp.(int)
	temp, ec = expr.Eval(fmt.Sprintf("%v", wa.right), wa.env)
	if ec != nil {
		log.Errorf("invalid b tree cfg %v", ec)
	}
	right := temp.(int)
	rand := rand.RandInt(left, right)
	log.Debugf("wait tick, all var = %v, rand = %d", wa.env, rand)
	//	ErrorCheck(rand > 0, 4, fmt.Sprintf("invalid b tree cfg, left = %d, right = %d", left, right))

	time.Sleep(time.Millisecond * time.Duration(rand))
	return b3.SUCCESS
}

func (eq *IsEqual) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		BENCHMARK = "benchmark"
		VALUE     = "value"
	)
	eq.Action.Initialize(setting)

	eq.env = fillEnv(eq.env, tempRobot)
	eq.env[BENCHMARK] = GetProperty(BENCHMARK, setting)
	eq.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) == (%v)`, eq.env[VALUE], eq.env[BENCHMARK])
	var ec error
	eq.Program, ec = expr.Compile(express, expr.Env(eq.env))

	ErrorCheck(nil == ec, 4, "invalid b tree cfg123")

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(eq.Program, eq.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
}

func (eq *IsEqual) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	eq.env = fillEnv(eq.env, *robot)
	//log.Debugf("is equal tick, all var = %v", eq.env)

	ret, ec := expr.Run(eq.Program, eq.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
	if nil != ec {
		log.Debugf("%v %v", ec.Error(), eq.GetTitle())
	}
	realRet := ret.(bool)

	if realRet {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

func (pMore *IsMore) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		BENCHMARK = "benchmark"
		VALUE     = "value"
	)
	pMore.Action.Initialize(setting)

	pMore.env = fillEnv(pMore.env, tempRobot)
	pMore.env[BENCHMARK] = GetProperty(BENCHMARK, setting)
	pMore.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) > (%v)`, pMore.env[VALUE], pMore.env[BENCHMARK])
	var ec error
	pMore.Program, ec = expr.Compile(express, expr.Env(pMore.env))
	ErrorCheck(nil == ec, 4, "invalid b tree cfg1")

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(pMore.Program, pMore.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg2")
}

func (pMore *IsMore) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	pMore.env = fillEnv(pMore.env, *robot)
	//log.Debugf("is equal tick, all var = %v", pMore.env)

	ret, ec := expr.Run(pMore.Program, pMore.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg3")
	realRet := ret.(bool)
	if realRet {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

func (pme *IsMoreEqual) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		BENCHMARK = "benchmark"
		VALUE     = "value"
	)
	pme.Action.Initialize(setting)

	pme.env = fillEnv(pme.env, tempRobot)
	pme.env[BENCHMARK] = GetProperty(BENCHMARK, setting)
	pme.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) >= (%v)`, pme.env[VALUE], pme.env[BENCHMARK])
	var ec error
	pme.Program, ec = expr.Compile(express, expr.Env(pme.env))
	ErrorCheck(nil == ec, 4, "invalid b tree cfg4")

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(pme.Program, pme.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg5")
}

func (pme *IsMoreEqual) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	pme.env = fillEnv(pme.env, *robot)
	//log.Debugf("is equal tick, all var = %v", pMore.env)

	ret, ec := expr.Run(pme.Program, pme.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
	realRet := ret.(bool)
	if realRet {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

func (pLess *IsLess) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		BENCHMARK = "benchmark"
		VALUE     = "value"
	)
	pLess.Action.Initialize(setting)

	pLess.env = fillEnv(pLess.env, tempRobot)
	pLess.env[BENCHMARK] = GetProperty(BENCHMARK, setting)
	pLess.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) < (%v)`, pLess.env[VALUE], pLess.env[BENCHMARK])
	var ec error
	pLess.Program, ec = expr.Compile(express, expr.Env(pLess.env))
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(pLess.Program, pLess.env)
	//	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
}

func (pLess *IsLess) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	pLess.env = fillEnv(pLess.env, *robot)
	//log.Debugf("is equal tick, all var = %v", pLess.env)

	ret, ec := expr.Run(pLess.Program, pLess.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
	realRet := ret.(bool)
	if realRet {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

func (ple *IsLessEqual) Initialize(setting *b3cfg.BTNodeCfg) {
	const (
		BENCHMARK = "benchmark"
		VALUE     = "value"
	)
	ple.Action.Initialize(setting)

	ple.env = fillEnv(ple.env, tempRobot)
	ple.env[BENCHMARK] = GetProperty(BENCHMARK, setting)
	ple.env[VALUE] = GetProperty(VALUE, setting)

	//构建一个临时的动态表达式的运行环境，主要是检查决策树的配置规则有没有语法错误
	express := fmt.Sprintf(`(%v) <= (%v)`, ple.env[VALUE], ple.env[BENCHMARK])
	var ec error
	ple.Program, ec = expr.Compile(express, expr.Env(ple.env))
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")

	//运行一下，如果有语法错误会马上报出来
	_, ec = expr.Run(ple.Program, ple.env)
	//	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
}

func (ple *IsLessEqual) OnTick(tick *b3core.Tick) b3.Status {
	robot := tick.GetTarget().(*robot)
	ple.env = fillEnv(ple.env, *robot)
	//log.Debugf("is equal tick, all var = %v", pLess.env)

	ret, ec := expr.Run(ple.Program, ple.env)
	ErrorCheck(nil == ec, 4, "invalid b tree cfg")
	realRet := ret.(bool)
	if realRet {
		return b3.SUCCESS
	}
	return b3.FAILURE
}

func Init(rb robot) {
	//ErrorCheck(nil != robot, errors.ErrorParam, "robot is nil, then maybe nil pointer reference !!!")
	tempRobot = rb
}
