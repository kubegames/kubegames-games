package server

import (
	"sync"

	"github.com/kubegames/kubegames-games/pkg/fishing/980101/data"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

var waitGroup sync.WaitGroup

var userOnlineMap = make(map[int64]*data.User)

func GetUser(uid int64) *data.User {
	return userOnlineMap[uid]
}
func SetUser(user *data.User) {
	//userOnlineMap[user.Userinfo.UserId] = user
}

type Server struct {
	Conf int32
}

func (self *Server) InitTable(table table.TableInterface) {
	tableLogic := new(TableLogic)
	tableLogic.init(table)
	table.Start(tableLogic, nil, nil)
}
