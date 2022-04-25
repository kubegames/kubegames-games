package game

import (
	"github.com/kubegames/kubegames-games/pkg/battle/960201/data"
)

//将玩家添加上桌，如果原先已经分配的则返回true，否则返回false
func (game *Game) AddUserIntoTable(user *data.User) (isDised bool) {

	user.TableId = game.TableId
	if game.GetUserListMap(user.User.GetID()) == nil {
		user.InTableCount = 0
		game.SetUserListMap(user)
		//fmt.Println("user list : ",game.userListArr)
		//设置椅子号
		game.SetCurUserChairId(user)
	} else {
		//fmt.Println("被分配过")
		isDised = true
	}
	return
}
