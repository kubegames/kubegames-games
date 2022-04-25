package game

//ExportData 导出变量
type ExportData struct {
	//游戏步骤
	GameStep int
	//当前最大叫地主类型
	//Calltype int
	//是否是地主
	IsBanker bool
	//是否是地主上家
	IsBankerPrev bool
	//是否是地主下家
	IsBankerNext bool
	//地主剩余手牌数量
	BankerCardsCount int
	//地主上家剩余手牌数量
	BankerPrevCardsCount int
	//地主下家剩余手牌数量
	BankerNextCardsCount int
	//手牌价值
	HandCardsValue int
	//自己手上牌的组数
	HandCardGroupNum int
	//自己的最高价值炸弹是否比剩余中的大
	IsMyBombGreaterThanLeft bool
	//自己的最高价值单牌是否比剩余中的大
	IsMySingleGreaterThanLeft bool
	//自己的最高价值对子是否比剩余中的大
	IsMyDoubleGreaterThanLeft bool
	//手上是否只剩下一组牌
	IsLeftOneCards bool
	//是否能出完手上所有牌
	IsCanOutAll bool
	//是否是主动出牌
	IsFirstOut bool
	//当前是否是地主出的牌
	IsBankeOut bool
	//当前是否是地主上家出的牌
	IsBankePrevOut bool
	//当前是否是地主下家出的牌
	IsBankeNextOut bool
	//当前出牌回合，地主是否已经过牌
	IsBankerPass bool
	//当前出牌回合，地主上家是否已经过牌
	IsBankerPrevPass bool
	//当前出牌回合，地主下家是否已经过牌
	IsBankerNextPass bool
	//当前出牌是否是单张
	IsOutSingle bool
	//当前出牌是否是对子
	IsOutDouble bool
	//当前回合出牌多少张
	OutCardCount int
	//当前回合出牌的价值
	OutCardValue int
	//自己手牌是否满足一组小牌， 其他绝对大牌的赢牌路径
	IsAllBigWin bool
	//自己跟牌之后是否能进入一组小牌， 其他绝对大牌的赢牌路径
	IsCanAllBig bool
	//自己是否有大王
	IsHasBigKing bool
	//自己是否有小王
	IsHasSmallKing bool
	//自己手上2的数量
	MyPoker2Count int
	//自己手上大于等于2的数量
	MyPoker2More int
	//自己手上A的数量
	MyPokerACount int
	//自己手上大于等于A的数量
	MyPokerAMore int
	//自己手上炸弹的数量
	MyBombCount         int
	MyBombCountMaxValue int
	//自己手上炸弹的数量（不含双王）
	MyNoKingBombCount int
	//自己手上的单牌数量
	MySingleCardsNum int
	//自己手上的对子数量
	MyDoubleCardsNum int
	//大王是否在其他人手牌中
	IsOtherBigKing bool
	//小王是否在其他人手牌中
	IsOtherSmallKing bool
	//其他人手上2的数量
	OtherPoker2Count int
	//其他人手上大于等于2的数量
	OtherPoker2More int
	//其他人手上A的数量
	OtherPokerACount int
	//其他人手上大于等于A的数量
	OtherPokerAMore int
	//其他人手上炸弹的数量
	OtherBombCount int
	//其他人手上最大单牌点数
	OtherMaxSingle int
	//其他人手上最大对子点数
	OtherMaxDouble int
	//是否有最优出牌策略
	IsHaveOptimalStrategy bool
	//判定地主手里的牌是否为对子
	IsBankerApair bool
	//判定地主下家手里的牌是否为对子
	IsBankerNextApair bool
	//判定地主上家手里的牌是否为对子
	IsBankerPrevAPair bool
	//其他玩家是否有双王
	IsOtherDoubleKing bool
	//其他玩家手上炸弹的数量
	OthersBombCount int
	//其他玩家手上有没有火箭
	IsOtherRoket bool
	//小于报牌玩家最小点数牌的个数
	MyMinAlertCardCount int
	//大于报牌玩家最大点数牌的个数
	MyMaxAlertCardCount int
	//自己手上的单牌点数小于报牌玩家最小的点数
	IsMySigleGreaterThanRight bool
	//自己手上的对子点数小于报牌玩家最小的点数
	IsMyDoubleGreaterThanRight bool
	//地主下家单牌比地主最大牌点数大
	IsBankerNextSingleGreaterThanLeft bool
	//接的是非单非双的牌型
	IsOutNotSigleAndDouble bool
	//有与出牌玩家一样的牌型，且可以接牌
	IsHaveSametype bool
	//判定拆牌之后可以接的起
	IsCanDemCardBeConnected bool
	//拆牌之后单牌数量增加数
	DemCardMySingleCardsNum int
	// 绝对大牌的数量
	AllBigCount int
	//自己手上有王炸
	IsMyRoket bool

	MyHandCardsValue             int
	CardsValueDiffWithBanker     int
	CardsValueDiffWithAnotherMax int

	BankerLastCardType      int
	EarlyBankerLastCardType int
	LateBankerLastCardType  int
	BankerPassType          int
	EarlyBankerPassType     int
	LateBankerPassType      int

	HandCardGroupTypes []int
}

/**
 * @导出行为
 * @title:叫地主
 * @type: 3
 * @describe: 机器人叫地主
 */
type callbanker struct {
	NodeBase
	call interface{}
}

type robbanker struct {
	NodeBase
	rob interface{}
}

type doOutCard struct {
	NodeBase
	groupType int
}

type IsContain struct {
	NodeBase
	name string
	v    int
}

/**
 * @导出行为
 * @title:加倍
 * @type: 3
 * @describe: 机器人加倍（不加倍）
 */
type makedouble struct {
	NodeBase
	//加倍（bool）true, false
	t interface{} // 0, 1, 2
}

/**
 * @导出行为
 * @title:不出
 * @type: 3
 * @describe: 机器人不出，主动出牌时不允许不出
 */
type passcard struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出牌
 * @type: 3
 * @describe: 机器人默认主动出牌方式
 */
type outcard struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出完所有牌
 * @type: 3
 * @describe: 机器人手上只剩一组牌时，出完所有牌
 */
type outAllcard struct {
	NodeBase
}

/**
 * @导出行为
 * @title:先出大牌,再出小牌的赢牌路径
 * @type: 3
 * @describe: 判断条件：外面的炸弹个数小于等于自己的炸弹个数，自己手上除了炸弹以外的牌，只有一组小牌，其他绝对大牌的情况下
 */
type outAbsBig struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出单张
 * @type: 3
 * @describe: 主动出牌方式：出单张
 */
type outSingle struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
}

/**
 * @导出行为
 * @title:出最小单张
 * @type: 3
 * @describe: 主动出牌方式：出最小单张
 */
type outMinSingle struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出单张以外的
 * @type: 3
 * @describe: 主动出牌方式：出单张以外的
 */
type outNotSingle struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
}

/**
 * @导出行为
 * @title:出对子
 * @type: 3
 * @describe: 主动出牌方式：出对子
 */
type outDouble struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
}

/**
 * @导出行为
 * @title:出最小对子
 * @type: 3
 * @describe: 主动出牌方式：出最小对子
 */
type outMinDouble struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出对子以外的
 * @type: 3
 * @describe: 主动出牌方式：出对子以外的
 */
type outNotDouble struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
}

/**
 * @导出行为
 * @title:跟牌
 * @type: 3
 * @describe: 机器人默认被动出牌方式，出能压过当前出牌同类型的牌
 */
type gencard struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:出炸弹
 * @type: 3
 * @describe: 机器人被动出牌方式：出手上最小能打过的炸弹
 */
type outBomb struct {
	NodeBase
}

type outBombMax struct {
	NodeBase
}

/**
 * @导出行为
 * @title:出4带2双或者4带2单
 * @type: 3
 * @describe: 优先出4带2双
 */
type out4With2DoubleOr2Single struct {
	NodeBase
}

/**
 * @导出行为
 * @title:拦着地主出牌
 * @type: 3
 * @describe: 拦着地主出牌
 */
type outCardToStopBanker struct {
	NodeBase
}

/**
 * @导出行为
 * @title:地主跟上家的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主上家的出牌最大，地主跟牌
 */
type bankerGenPrevOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:地主跟下家的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主下家的出牌最大，由于逆时针出牌顺序，可以判断地主上家是pass，地主跟牌
 */
type bankerGenNextOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:地主上家跟地主的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主的出牌最大，由于逆时针出牌顺序，可以判断地主下家是pass，地主上家跟牌
 */
type prevGenBankerOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:地主上家跟地主下家的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主下家的出牌最大，地主上家跟牌
 */
type prevGenNextOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:地主下家跟地主的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主的出牌最大，地主下家跟牌
 */
type nextGenBankerOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

/**
 * @导出行为
 * @title:地主下家跟地主上家的出牌
 * @type: 3
 * @describe: 机器人被动出牌方式，当前回合地主上家的出牌最大，由于逆时针出牌顺序，可以判断地主是pass，地主下家跟牌
 */
type nextGenPrevOut struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
	//同类型找不到时是否可以用炸弹
	isUseBomb interface{}
	//同类型和炸弹都找不到时，是否可以用王炸
	isUseRocket interface{}
}

//拆对子为单张出牌
type outThePairForSigle struct {
	NodeBase
	//是否从大到小, 否则从小到大, int
	isFromBig interface{}
}

//出绝对大牌中的最小牌
type OutAbsSmall struct {
	NodeBase
}

//出非最小牌型
type outNotMinCardType struct {
	NodeBase
	//是否从大到小, 否则从小到大, int
	isFromBig interface{}
}

//出绝对大牌中的最小牌
type outNotSigleAndDouble struct {
	NodeBase
}

//出报双玩家中间区域的牌
type outMiddleCountSigle struct {
	NodeBase
	//决定牌的最小点数, int
	minNumber interface{}
	//决定牌的最大点数, int
	maxNumber interface{}
	//报双玩家最小的牌, int
	minAlert interface{}
	//报双玩家最大的牌, int
	maxAlert interface{}
	//是否从大到小, 否则从小到大
	isFromBig interface{}
	//是否强制出牌, 打破现有拆分牌型，有则必出
	isForce interface{}
}

//出报双玩家中间区域的牌
type outCardByType struct {
	NodeBase
	//决定牌的最小点数, int
	t interface{}
	//是否从大到小, 否则从小到大
	BigToSmall interface{}
}
