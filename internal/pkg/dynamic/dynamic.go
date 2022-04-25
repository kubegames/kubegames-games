package dynamic

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
)

type Function struct {
	Foo      reflect.Value
	IsMethod bool
	Target   reflect.Value
}

func NewDynamic() (d *Dynamic) {
	d = new(Dynamic)
	d.RWMutex = new(sync.RWMutex)
	d.funcList = make(map[string]*Function)
	return
}

type Dynamic struct {
	*sync.RWMutex
	funcList map[string]*Function
}

func (self *Dynamic) AddFunc(funcName string, target interface{}, callback interface{}) {
	var f = new(Function)
	if target != nil {
		if reflect.TypeOf(callback) == reflect.TypeOf("") {
			f.Foo = reflect.ValueOf(target).MethodByName(callback.(string))
			f.IsMethod = true
		} else {
			f.Target = reflect.ValueOf(target)
			f.Foo = reflect.ValueOf(callback)
		}
	} else {
		f.Foo = reflect.ValueOf(callback)
	}
	if f.Foo.IsValid() {
		self.Lock()
		self.funcList[funcName] = f
		self.Unlock()
	} else {
		fmt.Println(fmt.Sprintf("Dynamic.AddFunc::Function[%s] Fail!", funcName))
	}
}
func (self *Dynamic) RemoveFunc(funcName string) {
	self.Lock()
	delete(self.funcList, funcName)
	self.Unlock()
}
func (self *Dynamic) Run(funcName string, args ...interface{}) (res []interface{}, err error) {
	res = make([]interface{}, 0)
	self.RLock()
	f, isFind := self.funcList[funcName]
	self.RUnlock()
	if isFind {
		var tmp []reflect.Value = make([]reflect.Value, 0)
		if f.Target.IsValid() && !f.IsMethod {
			tmp = append(tmp, f.Target)
		}
		for _, volume := range args {
			tmp = append(tmp, reflect.ValueOf(volume))
		}
		t := time.Now().UnixNano()
		resTmp := f.Foo.Call(tmp)
		n := (time.Now().UnixNano() - t) / int64(time.Millisecond)
		if n > 200 {
			fmt.Println("Dynamic.Run::>100ms:", funcName, args, n)
		}
		for _, v := range resTmp {
			res = append(res, v.Interface())
		}
	} else {
		fmt.Println(fmt.Sprintf("Dynamic.Run(not Find)::[%s] args=%v", funcName, args))
		err = errors.New("Dynamic.Run(not Find)")
	}
	return
}
func (d *Dynamic) RemoveAllFunc() {
	d.Lock()
	for k := range d.funcList {
		delete(d.funcList, k)
	}
	d.Unlock()
}
