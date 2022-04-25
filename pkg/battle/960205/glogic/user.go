package glogic

import "github.com/kubegames/kubegames-sdk/pkg/player"

type User struct {
	InterUser  player.PlayerInterface // 用户信息
	IsDeposit  bool
	ControlKey int32
	IsPoint    string
}
