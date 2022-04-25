package data

import (
	"sync"
	"time"
)

//该文件做一些比如登陆次数限制，抢包次数限制等

var userRobLimitMap = make(map[int64]*UserCountTime) //uid =>
var userRobLimitLock = new(sync.Mutex)

type UserCountTime struct {
	FailCount int
	Time      time.Time
}

func SetUserRobLimitMap(uid int64) {
	userRobLimitLock.Lock()
	defer userRobLimitLock.Unlock()
	if userRobLimitMap[uid] == nil {
		userRobLimitMap[uid] = &UserCountTime{FailCount: 0, Time: time.Now()}
	}
	userRobLimitMap[uid].FailCount++
	//userRobLimitMap[uid].Time = time.Now()
}

func GetUserRobLimitMap(uid int64) *UserCountTime {
	userRobLimitLock.Lock()
	defer userRobLimitLock.Unlock()
	return userRobLimitMap[uid]
}

func DelUserRobLimitMap(uid int64) {
	userRobLimitLock.Lock()
	defer userRobLimitLock.Unlock()
	delete(userRobLimitMap, uid)
}
