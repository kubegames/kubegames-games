package game

import (
	"go-game-sdk/inter"
	"go-game-sdk/lib/clock"
	"go-game-sdk/sdk/msg"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kubegames/kubegames-sdk/pkg/player"
	"github.com/kubegames/kubegames-sdk/pkg/table"
)

type UserInterTest struct {
}

func (user *UserInterTest) SendLogs(gameNum string, records []*msg.GameLog) {

}

func (user *UserInterTest) GetRoomNum() string {
	return ""
}

//设置总下注
func (user *UserInterTest) SetBetsAmount(betsAmount int64) {}

/*
   发送战绩 gameNum 局号
   outputAmount:产出  用户赢的钱减去税
*/
func (user *UserInterTest) SendRecord(gameNum string, profitAmount, betsAmount, drawAmount, outputAmount int64, endCards string) {
	return
}

//设置码量
func (user *UserInterTest) SendChip(chip int64) { return }

//获取ip
func (user *UserInterTest) GetIp() string { return "" }

//获取城市
func (user *UserInterTest) GetCity() string { return "" }

//获取用户性别
func (user *UserInterTest) GetSex() int32 { return 1 }

//创建跑马灯，content 跑马灯内容
func (user *UserInterTest) CreateMarquee(content string) error { return nil }

//上庄
func (user *UserInterTest) UpBanker() {}

//下庄
func (user *UserInterTest) DownBanker() {}

//发送打码量
//设置结算牌
func (user *UserInterTest) SetEndCards(cards string) {}

//获取点控级别
func (user *UserInterTest) GetProb() int32 { return 1 }

//绑定机器人逻辑
func (user *UserInterTest) BindRobot(interRobot player.RobotHandler) inter.AIUserInter { return nil }

/*
*用户游戏开始
*返回值说明:
*@return:string   		昵称
 */
//GameRun()
/*
*用户游戏结束
*返回值说明:
*@return:string   		昵称
 */
//GameEnd()
/*
*设置状态
*返回值说明:
*@return:string   		昵称
 */
func (user *UserInterTest) SetStatus(status int) {}

/*
*获取用户昵称
*返回值说明:
*@return:string   		昵称
 */
func (user *UserInterTest) GetNike() string { return "" }

/*
*获取用户头像
*返回值说明:
*@return:string   		用户头像
 */
func (user *UserInterTest) GetHead() string { return "" }

/*
*获取用户ID
*返回值说明:
*@return:int   		用户ID
 */
func (user *UserInterTest) GetId() int64 { return 1 }

/*
*获取用户积分
*返回值说明:
*@return:int   		积分
 */
func (user *UserInterTest) GetScore() int64 { return 1 }

/*
	*获取用户积分
	*参数说明:
	*@param:gameNum		局号
	*@param:score		积分
	*@param:kind		操作类型 1:投注 2:结算
	*@param:remark		备注
	*@param:tax			税收（万分比）
	*@param:principal	本金
	*@param:bussType	业务类型 具体老胡文档
	*@param:betsAmount 总下注
	*返回值
	@return int64扣完税的钱
*/
//SetScore(gameNum string, score int64, kind int, remark string, tax int64, principal int64, intakePool bool) (int64, error)
func (user *UserInterTest) SetScore(gameNum string, score int64, tax int64) (int64, error) {
	return 1, nil
}

/*
*获取椅子ID
*返回值说明:
*@return:id   		椅子ID
 */
func (user *UserInterTest) GetChairId() int {
	return 1
}

/*
*设置用户数据
*参数说明:
*@param:data		用户数据
 */
func (user *UserInterTest) SetTableData(data string) {}

/*
*获取用户数据
*返回值说明:
*@return:data   		用户数据
 */
func (user *UserInterTest) GetTableData() string { return "" }

/*
*删除用户数据
*返回值说明:
 */
func (user *UserInterTest) DelTableData() {}

/*
*用户发送消息
*参数说明:
*@param:subCmd		消息类型
*@param:pb		消息
*返回值说明:
*@return:error   	错误
 */
func (user *UserInterTest) SendMsg(subCmd int32, pb proto.Message) error { return nil }

/*
*判断用户是否为机器人--逻辑
*参数说明:
*@param:bool	true是机器人 false不是机器人
 */
func (user *UserInterTest) IsRobot() bool { return false }

type Table struct {
}

func (table *Table) WriteLogs(userId int64, content string) {}

//获取跑马灯配置
func (table *Table) GetMarqueeConfig() (configs []*msg.MarqueeConfig) { return }

//创建跑马灯， nickName 昵称, gold 金额, special 特殊条件
func (table *Table) CreateMarquee(nickName string, gold int64, special string, ruleId int64) error {
	return nil
}

/*
*获取奖金池
*参数说明:
*@param:ration	万分比
*@param:userId	用户id
*返回值
*第一个 int64:总共的值
*第二个 int64:实际赢的数量
*error 错误
 */
func (table *Table) GetCollect(ration int32, userId int64) (int64, int64, error) { return 1, 1, nil }

//获取等级
func (table *Table) GetLevel() int32 { return 1 }

//获取房间税收
func (table *Table) GetRoomRate() int64 { return 1 }

//获取局号
func (table *Table) GetGameNum() string { return "" }

//获取房间id -1为没有初始化房间id
func (table *Table) GetRoomID() int64 { return 1 }

//获取入场限制 -1为没有初始化入场限制
func (table *Table) GetEntranceRestrictions() int64 { return 1 }

//获取平台id -1为没有平台id
func (table *Table) GetPlatformID() int64 { return 1 }

//获取自定义配置
func (table *Table) GetAdviceConfig() string { return "" }

//获取血池
func (table *Table) GetRoomProb() (int32, error) { return 1, nil }

/*
*大厅广播消息
*参数说明:
 */
func (table *Table) BroadcastAll(subCmd int32, pb proto.Message) error { return nil }

/*
*获取机器人
 */
func (table *Table) GetRobot(int32) error { return nil }

/*
*获取桌子编号
*返回值说明:
*@return:int   		编号
 */
func (table *Table) GetId() int { return 1 }

/*
*绑定游戏逻辑
*参数说明:
*@param:GameInter	游戏逻辑接口对象
 */
func (table *Table) BindGame(gameInter table.TableHandler) { return }

/*
*开始游戏
 */
func (table *Table) StartGame() {}

/*
*结束游戏
 */
func (table *Table) EndGame() {}

/*
*提出用户
*参数说明:
*@param:UserInetr	用户接口对象
 */
func (table *Table) KickOut(user player.PlayerInterface) {}

/*
*通知桌子下的所有用户
*参数说明:
*@param:subCmd		消息类型
*@param:pb		消息内容
 */
func (table *Table) Broadcast(subCmd int32, pb proto.Message) {}

/*
*是否开赛
*返回值说明:
*@return:bool   	是否可以开赛
 */
//IsStart() bool

/*
*添加一次性定时器
*参数说明:
*@param:interval	time.Duration时间
*@param:jobFunc		回调业务函数
*返回值说明:
*@return:job   		定时器任务对象
*@return:ok   		是否成功
 */
func (table *Table) AddTimer(interval time.Duration, jobFunc func()) (job *clock.Job, ok bool) {
	return
}

/*
*添加多次定时器
*参数说明:
*@param:interval	time.Duration时间
*@param:num			执行次数
*@param:jobFunc		回调业务函数
*返回值说明:
*@return:job   		定时器任务对象
*@return:ok   		是否成功
 */
func (table *Table) AddTimerRepeat(interval time.Duration, num uint64, jobFunc func()) (job *clock.Job, inserted bool) {
	return
}
