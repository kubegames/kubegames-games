package model

import (
	"go-game-sdk/example/game_poker/saima/msg"
	"strconv"
	"strings"

	"github.com/kubegames/kubegames-games/internal/pkg/score"

	"github.com/kubegames/kubegames-sdk/pkg/log"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type User struct {
	Table         table.TableInterface
	User          player.PlayerInterface
	UserInfo      *msg.UserInfo
	AllBet        int64                   //从开始到现在的总下注
	BetInfo       map[msg.BetArea]int64   //区域的下注信息
	BetIndexsInfo map[msg.BetArea][]int32 //区域的下注信息
	CountInfo     map[msg.BetArea]int64   //区域的结算信息
	NoBetCount    int                     //未下注统计
	IsRebet       bool                    //是否重复下注
	Win           int64
}

//重置数据
func (user *User) ResetData() {
	user.AllBet = 0
	user.Win = 0
	user.BetInfo = make(map[msg.BetArea]int64, 38)
	user.CountInfo = make(map[msg.BetArea]int64, 38)
	user.BetIndexsInfo = make(map[msg.BetArea][]int32, 0)
	user.IsRebet = false
}

//获取下注信息
func (user *User) GetBetInfo() []*msg.BetInfo {
	betInfo := make([]*msg.BetInfo, 0)
	for k, v := range user.BetInfo {
		data := &msg.BetInfo{}
		data.BetArea = k
		data.Score = v
		betInfo = append(betInfo, data)
	}
	return betInfo
}

//获取赔付信息
func (user *User) GetCountInfo() []*msg.CountInfo {
	betInfo := make([]*msg.CountInfo, 0)
	for k, v := range user.CountInfo {
		data := &msg.CountInfo{}
		data.BetArea = k
		data.Score = v
		betInfo = append(betInfo, data)
		if v > 0 {
			user.Win += v
		}
	}
	return betInfo
}

//获取赔付
func (user *User) GetWin() int64 {
	win := int64(0)
	for _, v := range user.CountInfo {
		win += v
	}
	return win
}

//获取操作日志
func (user *User) GetOperationLog() string {
	content := "玩家ID: " + strconv.FormatInt(user.User.GetID(), 10) +
		" 用户剩余金额:" + score.GetScoreStr(user.User.GetScore())
	first := " 冠军投注: "
	fAndS := " 前两名投注: "
	fAndM := " 男女投注: "
	for k, v := range user.BetInfo {
		temString := " 投注金额: " + strconv.Itoa(int(v)) +
			" 赔付金额: " + strconv.Itoa(int(user.CountInfo[k]))
		if k < msg.BetArea_first_second_12 {
			first += user.transAreaToLogString(k) + temString
		} else if k < msg.BetArea_champion_Man {
			fAndS += user.transAreaToLogString(k) + temString
		} else {
			fAndM += user.transAreaToLogString(k) + temString
		}
	}
	return content + first + fAndS + fAndM
}

//转化下注区域字符串
func (user *User) transAreaToLogString(area msg.BetArea) string {
	areaString := " 编号"
	num := strings.Split(area.String(), "_")
	areaString += num[len(num)-1]
	if area == msg.BetArea_champion_Woman {
		areaString = "女: "
	}
	if area == msg.BetArea_champion_Man {
		areaString = "男: "
	}
	log.Tracef("areaString =", areaString)
	return areaString
}

//获取打码量
func (user *User) GetChip() int64 {
	woman := user.BetInfo[msg.BetArea_champion_Man]
	man := user.BetInfo[msg.BetArea_champion_Man]
	diff := woman - man
	if diff < 0 {
		diff = -diff
	}
	return user.AllBet - woman - man + diff
}

//取消下注
func (user *User) BetClear() {
	user.UserInfo.Amount += user.AllBet
	user.AllBet = 0
	user.BetInfo = make(map[msg.BetArea]int64, 38)
}
