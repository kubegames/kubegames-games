package game

import (
	"github.com/antonmedv/expr/vm"
	b3core "github.com/magicsea/behavior3go/core"
)

//节点基类，决策树任何节点都请组合该节点，使用范例可参阅炸金花的hehavior.go
type NodeBase struct {
	b3core.Action
	Program *vm.Program
}

//判断给定的值是否落在某个区间
type IsInRange struct {
	NodeBase
	env map[string]interface{}
	//调试打日志用的，没啥卵用
	express string
}

//sleep功能
type Wait struct {
	NodeBase
	env   map[string]interface{}
	left  interface{}
	right interface{}
}

//判等
type IsEqual struct {
	NodeBase
	env map[string]interface{}
}

//大于
type IsMore struct {
	NodeBase
	env map[string]interface{}
}

//大于等于
type IsMoreEqual struct {
	NodeBase
	env map[string]interface{}
}

//小于
type IsLess struct {
	NodeBase
	env map[string]interface{}
}

//小于等于
type IsLessEqual struct {
	NodeBase
	env map[string]interface{}
}
