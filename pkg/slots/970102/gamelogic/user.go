package gamelogic

import "github.com/kubegames/kubegames-sdk/pkg/player"

type User struct {
	user       player.PlayerInterface
	Bet        int32 // 下注
	AllBet     int32 // 总下注
	Gold       int32 // 总赢钱
	Score      int64 // 税后
	ControlKey int32
	IsPoint    string
}

func NewUser(user player.PlayerInterface) *User {
	return &User{
		user: user,
	}
}
