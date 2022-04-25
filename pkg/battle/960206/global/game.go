package global

//一些时间、金额常量
const (
	WAN_RATE_TOTAL  = 10000
	RED_TOTAL_ROUTE = 15
)

//牌桌当前状态
const (
	TABLE_CUR_STATUS_WAIT = 0 //准备开赛

	//  mod by wd in 2020.3.12
	//TABLE_CUR_STATUS_WAIT_SEND_CARDS = 1 //等待客户端发牌完毕
	//TABLE_CUR_STATUS_ING             = 2 //正在游戏
	//TABLE_CUR_STATUS_END             = 3 //游戏结束
	TABLE_CUR_STATUS_COMPOSE = 1 // 摆牌状态
	TABLE_CUR_STATUS_COMPARE = 2 // 比牌状态
	TABLE_CUR_STATUS_SETTLE  = 3 // 结算状态
	TABLE_CUR_STATUS_END     = 4 // 游戏结束
)

const (
	SCORE_KIND_PEACE = 0 // 积分变动类型-和
	SCORE_KIND_BET   = 1 // 积分变动类型-下注
	SCORE_KIND_WIN   = 2 // 积分变动类型-赢
	SCORE_KIND_BACK  = 2 // 积分变动类型-退还
)

const (
	USER_STATUS_WAIT = 0
	USER_STATUS_ING  = 1
	USER_STATUS_END  = 2
)

const (
	COMPARE_LOSE = -1
	COMPARE_EQ   = 0
	COMPARE_WIN  = 1
)

const (
	//1:投注 2:结算
	SET_SCORE_BET    = 1 //下注
	SET_SCORE_SETTLE = 2 //结算
)

const (
	//头墩冲三的分
	HEAD_BZ_SCORE = 3
)

//时间
const (
	TIMER_START_GAME = 100
)
