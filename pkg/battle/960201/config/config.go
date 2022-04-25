package config

import (
	"encoding/json"
	"os"
)

type config struct {
	ListenerPort int    //web socket 侦听端口
	MaxConn      int32  //最大连接数
	DBUri        string //数据库链接字符串
	LogPath      string //需要制定输出的日志路径
}

type AiConfig struct {
	//未看牌决定是否要看牌
	SeeCardsBase        int //看牌基本概率
	SeeCardsCount1      int //看牌用户人数为1
	SeeCardsCount2      int //看牌用户人数为2
	SeeCardsCount3      int //看牌用户人数为3
	SeeCardsCount4      int //看牌用户人数为4
	SeeCardsRaise1      int //看牌后加注人数为1
	SeeCardsRaise2      int //看牌后加注人数为2
	SeeCardsRaise3      int //看牌后加注人数为3
	SeeCardsRaise4      int //看牌后加注人数为4
	SeeCardsRound3      int //游戏轮数 < 3轮
	SeeCardsRound5      int //游戏轮数 < 5轮
	SeeCardsRound10     int //游戏轮数 < 10轮
	SeeCardsRound15     int //游戏轮数 < 15轮
	SeeCardsAllIn       int //全押之后看牌概率
	SeeCardsCardsIndex1 int //新增-自己的牌在牌局中第1大（最大）
	SeeCardsCardsIndex2 int //新增-自己的牌在牌局中第2大
	SeeCardsCardsIndex3 int //新增-自己的牌在牌局中第3大
	SeeCardsCardsIndex4 int //新增-自己的牌在牌局中第4大
	SeeCardsCardsIndex5 int //新增-自己的牌在牌局中第5大

	////未看牌的操作
	NotSeeFollowCardsIndex1  int //新增-未看牌自己的牌第1大，跟注
	NotSeeFollowCardsIndex2  int //新增-未看牌自己的牌第2大，跟注
	NotSeeFollowCardsIndex3  int //新增-未看牌自己的牌第3大，跟注
	NotSeeFollowCardsIndex4  int //新增-未看牌自己的牌第4大，跟注
	NotSeeFollowCardsIndex5  int //新增-未看牌自己的牌第5大，跟注
	NotSeeRaiseCardsIndex1   int //新增-未看牌自己的牌第1大，加注
	NotSeeRaiseCardsIndex2   int //新增-未看牌自己的牌第2大，加注
	NotSeeRaiseCardsIndex3   int //新增-未看牌自己的牌第3大，加注
	NotSeeRaiseCardsIndex4   int //新增-未看牌自己的牌第4大，加注
	NotSeeRaiseCardsIndex5   int //新增-未看牌自己的牌第5大，加注
	NotSeeCompareCardsIndex1 int //新增-未看牌自己的牌第1大，比牌
	NotSeeCompareCardsIndex2 int //新增-未看牌自己的牌第2大，比牌
	NotSeeCompareCardsIndex3 int //新增-未看牌自己的牌第3大，比牌
	NotSeeCompareCardsIndex4 int //新增-未看牌自己的牌第4大，比牌
	NotSeeCompareCardsIndex5 int //新增-未看牌自己的牌第5大，比牌
	//看牌用户数
	NotSeeBaseFollow  int //未看牌跟注的基础概率
	NotSeeBaseRaise   int //未看牌跟注的基础概率
	NotSeeBaseCompare int //未看牌跟注的基础概率
	NotSeeFollowSaw1  int //未看牌情况下，已经看了牌的人数为1，跟注
	NotSeeFollowSaw2  int //未看牌情况下，已经看了牌的人数为2，跟注
	NotSeeFollowSaw3  int //未看牌情况下，已经看了牌的人数为3，跟注
	NotSeeFollowSaw4  int //未看牌情况下，已经看了牌的人数为4，跟注
	NotSeeRaiseSaw1   int //未看牌情况下，已经看了牌的人数为1，加注
	NotSeeRaiseSaw2   int //未看牌情况下，已经看了牌的人数为2，加注
	NotSeeRaiseSaw3   int //未看牌情况下，已经看了牌的人数为3，加注
	NotSeeRaiseSaw4   int //未看牌情况下，已经看了牌的人数为4，加注
	NotSeeCompareSaw1 int //未看牌情况下，已经看了牌的人数为1，比牌
	NotSeeCompareSaw2 int //未看牌情况下，已经看了牌的人数为2，比牌
	NotSeeCompareSaw3 int //未看牌情况下，已经看了牌的人数为3，比牌
	NotSeeCompareSaw4 int //未看牌情况下，已经看了牌的人数为4，比牌
	//剩余人数
	NotSeeFollowRemain2  int //未看牌情况下，剩余人数2，选择跟注
	NotSeeFollowRemain3  int //未看牌情况下，剩余人数3，选择跟注
	NotSeeFollowRemain4  int //未看牌情况下，剩余人数4，选择跟注
	NotSeeFollowRemain5  int //未看牌情况下，剩余人数5，选择跟注
	NotSeeRaiseRemain2   int //未看牌情况下，剩余人数2，选择加注
	NotSeeRaiseRemain3   int //未看牌情况下，剩余人数3，选择加注
	NotSeeRaiseRemain4   int //未看牌情况下，剩余人数4，选择加注
	NotSeeRaiseRemain5   int //未看牌情况下，剩余人数5，选择加注
	NotSeeCompareRemain2 int //未看牌情况下，剩余人数2，选择比牌
	NotSeeCompareRemain3 int //未看牌情况下，剩余人数3，选择比牌
	NotSeeCompareRemain4 int //未看牌情况下，剩余人数4，选择比牌
	NotSeeCompareRemain5 int //未看牌情况下，剩余人数5，选择比牌
	//其他用户比牌胜利次数
	NotSeeOtherCompare1Follow  int //未看牌情况下，其他用户比牌胜利次数1，跟注
	NotSeeOtherCompare2Follow  int //未看牌情况下，其他用户比牌胜利次数2，跟注
	NotSeeOtherCompare3Follow  int //未看牌情况下，其他用户比牌胜利次数3，跟注
	NotSeeOtherCompare1Raise   int //未看牌情况下，其他用户比牌胜利次数1，加注
	NotSeeOtherCompare2Raise   int //未看牌情况下，其他用户比牌胜利次数2，加注
	NotSeeOtherCompare3Raise   int //未看牌情况下，其他用户比牌胜利次数3，加注
	NotSeeOtherCompare1Compare int //未看牌情况下，其他用户比牌胜利次数1，比牌
	NotSeeOtherCompare2Compare int //未看牌情况下，其他用户比牌胜利次数2，比牌
	NotSeeOtherCompare3Compare int //未看牌情况下，其他用户比牌胜利次数3，比牌
	//自己比牌胜利次数
	NotSeeSelfCompare1Follow  int //未看牌情况下，自己比牌胜利次数1，跟注
	NotSeeSelfCompare2Follow  int //未看牌情况下，自己比牌胜利次数2，跟注
	NotSeeSelfCompare3Follow  int //未看牌情况下，自己比牌胜利次数3，跟注
	NotSeeSelfCompare1Raise   int //未看牌情况下，自己比牌胜利次数1，加注
	NotSeeSelfCompare2Raise   int //未看牌情况下，自己比牌胜利次数2，加注
	NotSeeSelfCompare3Raise   int //未看牌情况下，自己比牌胜利次数3，加注
	NotSeeSelfCompare1Compare int //未看牌情况下，自己比牌胜利次数1，比牌
	NotSeeSelfCompare2Compare int //未看牌情况下，自己比牌胜利次数2，比牌
	NotSeeSelfCompare3Compare int //未看牌情况下，自己比牌胜利次数3，比牌
	//当前加注档位
	NotSeeRaiseAmount1Follow  int //未看牌情况下，加注档位1，跟注
	NotSeeRaiseAmount2Follow  int //未看牌情况下，加注档位2，跟注
	NotSeeRaiseAmount3Follow  int //未看牌情况下，加注档位3，跟注
	NotSeeRaiseAmount1Raise   int //未看牌情况下，加注档位1，加注
	NotSeeRaiseAmount2Raise   int //未看牌情况下，加注档位2，加注
	NotSeeRaiseAmount3Raise   int //未看牌情况下，加注档位3，加注
	NotSeeRaiseAmount1Compare int //未看牌情况下，加注档位1，比牌
	NotSeeRaiseAmount2Compare int //未看牌情况下，加注档位2，比牌
	NotSeeRaiseAmount3Compare int //未看牌情况下，加注档位3，比牌
	//当前轮数
	NotSeeRound2Follow   int //当前轮数 <=2 ，跟注
	NotSeeRound5Follow   int //当前轮数 <=5 ，跟注
	NotSeeRound10Follow  int //当前轮数 <=10 ，跟注
	NotSeeRound20Follow  int //当前轮数 <=20 ，跟注
	NotSeeRound2Raise    int //当前轮数 <=2 ，加注
	NotSeeRound5Raise    int //当前轮数 <=5 ，加注
	NotSeeRound10Raise   int //当前轮数 <=10 ，加注
	NotSeeRound20Raise   int //当前轮数 <=20 ，加注
	NotSeeRound2Compare  int //当前轮数 <=2 ，比牌
	NotSeeRound5Compare  int //当前轮数 <=5 ，比牌
	NotSeeRound10Compare int //当前轮数 <=10 ，比牌
	NotSeeRound20Compare int //当前轮数 <=20 ，比牌

	////看了牌的操作
	//新增-自己的牌在牌局中第几大，对应不看牌的操作
	SawFollowCardsIndex1  int //新增-未看牌自己的牌第1大，跟注
	SawFollowCardsIndex2  int //新增-未看牌自己的牌第2大，跟注
	SawFollowCardsIndex3  int //新增-未看牌自己的牌第3大，跟注
	SawFollowCardsIndex4  int //新增-未看牌自己的牌第4大，跟注
	SawFollowCardsIndex5  int //新增-未看牌自己的牌第5大，跟注
	SawRaiseCardsIndex1   int //新增-未看牌自己的牌第1大，加注
	SawRaiseCardsIndex2   int //新增-未看牌自己的牌第2大，加注
	SawRaiseCardsIndex3   int //新增-未看牌自己的牌第3大，加注
	SawRaiseCardsIndex4   int //新增-未看牌自己的牌第4大，加注
	SawRaiseCardsIndex5   int //新增-未看牌自己的牌第5大，加注
	SawCompareCardsIndex1 int //新增-未看牌自己的牌第1大，比牌
	SawCompareCardsIndex2 int //新增-未看牌自己的牌第2大，比牌
	SawCompareCardsIndex3 int //新增-未看牌自己的牌第3大，比牌
	SawCompareCardsIndex4 int //新增-未看牌自己的牌第4大，比牌
	SawCompareCardsIndex5 int //新增-未看牌自己的牌第5大，比牌
	SawGiveUpCardsIndex1  int //新增-未看牌自己的牌第1大，弃牌
	SawGiveUpCardsIndex2  int //新增-未看牌自己的牌第2大，弃牌
	SawGiveUpCardsIndex3  int //新增-未看牌自己的牌第3大，弃牌
	SawGiveUpCardsIndex4  int //新增-未看牌自己的牌第4大，弃牌
	SawGiveUpCardsIndex5  int //新增-未看牌自己的牌第5大，弃牌
	//看牌用户数
	SawBaseFollow  int //看了牌的跟注的基础概率
	SawBaseRaise   int //看了牌的加注的基础概率
	SawBaseCompare int //看了牌的比牌的基础概率
	SawBaseGiveUp  int //看了牌的弃牌的基础概率
	SawFollowSaw1  int //看了牌的情况下，已经看了牌的人数为1，跟注
	SawFollowSaw2  int //看了牌的情况下，已经看了牌的人数为2，跟注
	SawFollowSaw3  int //看了牌的情况下，已经看了牌的人数为3，跟注
	SawFollowSaw4  int //看了牌的情况下，已经看了牌的人数为4，跟注
	SawRaiseSaw1   int //看了牌的情况下，已经看了牌的人数为1，加注
	SawRaiseSaw2   int //看了牌的情况下，已经看了牌的人数为2，加注
	SawRaiseSaw3   int //看了牌的情况下，已经看了牌的人数为3，加注
	SawRaiseSaw4   int //看了牌的情况下，已经看了牌的人数为4，加注
	SawCompareSaw1 int //看了牌的情况下，已经看了牌的人数为1，比牌
	SawCompareSaw2 int //看了牌的情况下，已经看了牌的人数为2，比牌
	SawCompareSaw3 int //看了牌的情况下，已经看了牌的人数为3，比牌
	SawCompareSaw4 int //看了牌的情况下，已经看了牌的人数为4，比牌
	SawGiveUpSaw1  int //看了牌的情况下，已经看了牌的人数为1，弃牌
	SawGiveUpSaw2  int //看了牌的情况下，已经看了牌的人数为2，弃牌
	SawGiveUpSaw3  int //看了牌的情况下，已经看了牌的人数为3，弃牌
	SawGiveUpSaw4  int //看了牌的情况下，已经看了牌的人数为4，弃牌
	//剩余人数
	SawFollowRemain2  int //看了牌的情况下，剩余人数2，选择跟注
	SawFollowRemain3  int //看了牌的情况下，剩余人数3，选择跟注
	SawFollowRemain4  int //看了牌的情况下，剩余人数4，选择跟注
	SawFollowRemain5  int //看了牌的情况下，剩余人数5，选择跟注
	SawRaiseRemain2   int //看了牌的情况下，剩余人数2，选择加注
	SawRaiseRemain3   int //看了牌的情况下，剩余人数3，选择加注
	SawRaiseRemain4   int //看了牌的情况下，剩余人数4，选择加注
	SawRaiseRemain5   int //看了牌的情况下，剩余人数5，选择加注
	SawCompareRemain2 int //看了牌的情况下，剩余人数2，选择比牌
	SawCompareRemain3 int //看了牌的情况下，剩余人数3，选择比牌
	SawCompareRemain4 int //看了牌的情况下，剩余人数4，选择比牌
	SawCompareRemain5 int //看了牌的情况下，剩余人数5，选择比牌
	SawGiveUpRemain2  int //看了牌的情况下，剩余人数2，弃牌
	SawGiveUpRemain3  int //看了牌的情况下，剩余人数3，弃牌
	SawGiveUpRemain4  int //看了牌的情况下，剩余人数4，弃牌
	SawGiveUpRemain5  int //看了牌的情况下，剩余人数5，弃牌
	//其他用户比牌胜利次数
	SawOtherCompare1Follow  int //看了牌的情况下，其他用户比牌胜利次数1，跟注
	SawOtherCompare2Follow  int //看了牌的情况下，其他用户比牌胜利次数2，跟注
	SawOtherCompare3Follow  int //看了牌的情况下，其他用户比牌胜利次数3，跟注
	SawOtherCompare1Raise   int //看了牌的情况下，其他用户比牌胜利次数1，加注
	SawOtherCompare2Raise   int //看了牌的情况下，其他用户比牌胜利次数2，加注
	SawOtherCompare3Raise   int //看了牌的情况下，其他用户比牌胜利次数3，加注
	SawOtherCompare1Compare int //看了牌的情况下，其他用户比牌胜利次数1，比牌
	SawOtherCompare2Compare int //看了牌的情况下，其他用户比牌胜利次数2，比牌
	SawOtherCompare3Compare int //看了牌的情况下，其他用户比牌胜利次数3，比牌
	SawOtherCompare1GiveUp  int //看了牌的情况下，其他用户比牌胜利次数1，弃牌
	SawOtherCompare2GiveUp  int //看了牌的情况下，其他用户比牌胜利次数2，弃牌
	SawOtherCompare3GiveUp  int //看了牌的情况下，其他用户比牌胜利次数3，弃牌
	//自己比牌胜利次数
	SawSelfCompare1Follow  int //看了牌的情况下，自己比牌胜利次数1，跟注
	SawSelfCompare2Follow  int //看了牌的情况下，自己比牌胜利次数2，跟注
	SawSelfCompare3Follow  int //看了牌的情况下，自己比牌胜利次数3，跟注
	SawSelfCompare1Raise   int //看了牌的情况下，自己比牌胜利次数1，加注
	SawSelfCompare2Raise   int //看了牌的情况下，自己比牌胜利次数2，加注
	SawSelfCompare3Raise   int //看了牌的情况下，自己比牌胜利次数3，加注
	SawSelfCompare1Compare int //看了牌的情况下，自己比牌胜利次数1，比牌
	SawSelfCompare2Compare int //看了牌的情况下，自己比牌胜利次数2，比牌
	SawSelfCompare3Compare int //看了牌的情况下，自己比牌胜利次数3，比牌
	SawSelfCompare1GiveUp  int //看了牌的情况下，自己比牌胜利次数1，弃牌
	SawSelfCompare2GiveUp  int //看了牌的情况下，自己比牌胜利次数2，弃牌
	SawSelfCompare3GiveUp  int //看了牌的情况下，自己比牌胜利次数3，弃牌
	//当前加注档位
	SawRaiseAmount1Follow  int //看了牌情况下，加注档位1，跟注
	SawRaiseAmount2Follow  int //看了牌情况下，加注档位2，跟注
	SawRaiseAmount3Follow  int //看了牌情况下，加注档位3，跟注
	SawRaiseAmount1Raise   int //看了牌情况下，加注档位1，加注
	SawRaiseAmount2Raise   int //看了牌情况下，加注档位2，加注
	SawRaiseAmount3Raise   int //看了牌情况下，加注档位3，加注
	SawRaiseAmount1Compare int //看了牌情况下，加注档位1，比牌
	SawRaiseAmount2Compare int //看了牌情况下，加注档位2，比牌
	SawRaiseAmount3Compare int //看了牌情况下，加注档位3，比牌
	SawRaiseAmount1GiveUp  int //看了牌情况下，加注档位1，弃牌
	SawRaiseAmount2GiveUp  int //看了牌情况下，加注档位2，弃牌
	SawRaiseAmount3GiveUp  int //看了牌情况下，加注档位3，弃牌
	//当前轮数
	SawRound2Follow   int //当前轮数 <=2 ，跟注
	SawRound5Follow   int //当前轮数 <=5 ，跟注
	SawRound10Follow  int //当前轮数 <=10 ，跟注
	SawRound20Follow  int //当前轮数 <=20 ，跟注
	SawRound2Raise    int //当前轮数 <=2 ，加注
	SawRound5Raise    int //当前轮数 <=5 ，加注
	SawRound10Raise   int //当前轮数 <=10 ，加注
	SawRound20Raise   int //当前轮数 <=20 ，加注
	SawRound2Compare  int //当前轮数 <=2 ，比牌
	SawRound5Compare  int //当前轮数 <=5 ，比牌
	SawRound10Compare int //当前轮数 <=10 ，比牌
	SawRound20Compare int //当前轮数 <=20 ，比牌
	SawRound2GiveUp   int //当前轮数 <=2 ，弃牌
	SawRound5GiveUp   int //当前轮数 <=5 ，弃牌
	SawRound10GiveUp  int //当前轮数 <=10 ，弃牌
	SawRound20GiveUp  int //当前轮数 <=20 ，弃牌
	//看了牌之后牌型
	//SawCardTypeSingleFollow  int //高牌（就是单张），跟注
	SawCardTypeSingleFollowOverA  int
	SawCardTypeSingleFollowBelowA int
	SawCardTypeDzFollow           int //对子，跟注
	SawCardTypeSzFollow           int //顺子，跟注
	SawCardTypeJhFollow           int //金花，跟注
	SawCardTypeSjFollow           int //顺金，跟注
	SawCardTypeBzFollow           int //豹子，跟注
	//SawCardTypeSingleRaise   int //高牌，加注
	SawCardTypeSingleRaiseOverA  int
	SawCardTypeSingleRaiseBelowA int
	SawCardTypeDzRaise           int //对子，加注
	SawCardTypeSzRaise           int //顺子，加注
	SawCardTypeJhRaise           int //金花，加注
	SawCardTypeSjRaise           int //顺金，加注
	SawCardTypeBzRaise           int //豹子，加注
	//SawCardTypeSingleCompare int //高牌，比牌
	SawCardTypeSingleCompareOverA  int
	SawCardTypeSingleCompareBelowA int
	SawCardTypeDzCompare           int //对子，比牌
	SawCardTypeSzCompare           int //顺子，比牌
	SawCardTypeJhCompare           int //金花，比牌
	SawCardTypeSjCompare           int //顺金，比牌
	SawCardTypeBzCompare           int //豹子，比牌
	//SawCardTypeSingleGiveUp  int //高牌，弃牌
	SawCardTypeSingleGiveUpOverA  int
	SawCardTypeSingleGiveUpBelowA int
	SawCardTypeDzGiveUp           int //对子，弃牌
	SawCardTypeSzGiveUp           int //顺子，弃牌
	SawCardTypeJhGiveUp           int //金花，弃牌
	SawCardTypeSjGiveUp           int //顺金，弃牌
	SawCardTypeBzGiveUp           int //豹子，弃牌

	IntoRoom1Person  int //房间只有1个玩家分配机器人的概率
	IntoRoom2Person  int
	IntoRoom3Person  int
	LeaveRoom5Person int //房间有5个机器人退出的概率
	LeaveRoom3Person int //房间小于3个机器人退出的概率
	FirstRoundFollow int //第一轮跟注概率
	FirstRoundRaise  int //第一轮加注概率

	//SecondRoundNotSee1       int //第二轮未看牌有效玩家数为1对应的看牌概率
	//SecondRoundNotSee2       int //第二轮未看牌有效玩家数为2对应的看牌概率
	//SecondRoundNotSeeOver2   int //第二轮未看牌有效玩家数>2对应的看牌概率
	//SecondRoundSee           int //第二轮看牌概率
	//SecondRoundGiveUp        int //第二轮弃牌概率
	//SecondRoundNotSeeFollow  int //第二轮没看牌跟注概率
	//SecondRoundNotSeeRaise   int //第二轮没看牌加注概率
	//SecondRoundNotSeeCompare int //第二轮没看牌比牌概率
	//SecondRoundNotSeeAllin   int //第二轮没看牌全押概率
	//SecondRoundSawFollow     int //第二轮看了牌跟注概率
	//SecondRoundSawRaise      int //第二轮看了牌加注概率
	//SecondRoundSawCompare    int //第二轮看了牌比牌概率
	//SecondRoundSawAllin      int //第二轮看了牌全押概率

	SecondRoundNotSee1         int
	SecondRoundNotSee2         int
	SecondRoundNotSeeOver2     int
	SecondRoundNotSeeGiveUp    int
	SecondRoundNotSeeFollow    int
	SecondRoundNotSeeRaise     int
	SecondRoundNotSeeCompare   int
	SecondRoundNotSeeAllin     int
	SecondRoundSawOverKGiveUp  int
	SecondRoundSawBelowKGiveUp int
	SecondRoundSawFollow       int
	SecondRoundSawRaise        int
	SecondRoundSawCompare      int
	SecondRoundSawAllin        int

	ThirdRoundSawBelowKGiveUp int
	ThirdRoundDzGiveUp        int
	ThirdRoundNotSee1         int //第三轮未看牌有效玩家数为1对应的看牌概率
	ThirdRoundNotSee2         int //第三轮未看牌有效玩家数为2对应的看牌概率
	ThirdRoundNotSeeOver2     int //第三轮未看牌有效玩家数>2对应的看牌概率
	ThirdRoundSee             int //第三轮看牌概率
	ThirdRoundNotSeeGiveUp    int //第三轮弃牌概率
	ThirdRoundNotSeeFollow    int //第三轮没看牌跟牌概率
	ThirdRoundNotSeeRaise     int //第三轮没看牌加注概率
	ThirdRoundNotSeeCompare   int //第三轮没看牌比牌概率
	ThirdRoundNotSeeAllin     int //第三轮没看牌全押概率

	ThirdRoundDzFollow  int //第三轮对子及以下跟注的概率
	ThirdRoundDzRaise   int //第三轮对子及以下加注的概率
	ThirdRoundDzCompare int //第三轮对子及以下比牌的概率
	ThirdRoundDzAllin   int //第三轮对子及以下全押的概率

	ThirdRoundSzGiveUp  int //第三轮顺子全押的概率
	ThirdRoundSzAllin   int //第三轮顺子全押的概率
	ThirdRoundSzCompare int //第三轮顺子比牌的概率
	ThirdRoundSzRaise   int //第三轮顺子加注的概率
	ThirdRoundSzFollow  int //第三轮顺子跟注的概率

	ThirdRoundJhGiveUp  int //第三轮金花跟注的概率
	ThirdRoundJhFollow  int //第三轮金花跟注的概率
	ThirdRoundJhRaise   int //第三轮金花加注的概率
	ThirdRoundJhCompare int //第三轮金花比牌的概率
	ThirdRoundJhAllin   int //第三轮金花全押的概率

	ThirdRoundOverJhGiveUp  int //第三轮金花以上跟注的概率
	ThirdRoundOverJhFollow  int //第三轮金花以上跟注的概率
	ThirdRoundOverJhRaise   int //第三轮金花以上加注的概率
	ThirdRoundOverJhCompare int //第三轮金花以上比牌的概率
	ThirdRoundOverJhAllin   int //第三轮金花以上全押的概率

	FourOverNotSee1       int //第4轮未看牌有效玩家数为1对应的看牌概率
	FourOverNotSee2       int //第4轮未看牌有效玩家数为2对应的看牌概率
	FourOverNotSeeOver2   int //第4轮未看牌有效玩家数>2对应的看牌概率
	FourOverNotSeeGiveUp  int //
	FourOverNotSeeFollow  int //第4轮未看牌跟注概率
	FourOverNotSeeRaise   int //第4轮未看牌加注概率
	FourOverNotSeeCompare int //第4轮未看牌比牌概率
	FourOverNotSeeAllin   int //第4轮未看牌全押概率

	FourOverDzGiveUp  int //第4轮对子及以下跟注的概率
	FourOverDzFollow  int //第4轮对子及以下跟注的概率
	FourOverDzRaise   int //第4轮对子及以下加注的概率
	FourOverDzCompare int //第4轮对子及以下比牌的概率
	FourOverDzAllin   int //第4轮对子及以下全押的概率

	FourOverSzGiveUp  int //第4轮顺子全押的概率
	FourOverSzAllin   int //第4轮顺子全押的概率
	FourOverSzCompare int //第4轮顺子比牌的概率
	FourOverSzRaise   int //第4轮顺子加注的概率
	FourOverSzFollow  int //第4轮顺子跟注的概率

	FourOverJhGiveUp  int //第4轮金花跟注的概率
	FourOverJhFollow  int //第4轮金花跟注的概率
	FourOverJhRaise   int //第4轮金花加注的概率
	FourOverJhCompare int //第4轮金花比牌的概率
	FourOverJhAllin   int //第4轮金花全押的概率

	FourOverOverJhGiveUp  int //第4轮金花以上跟注的概率
	FourOverOverJhFollow  int //第4轮金花以上跟注的概率
	FourOverOverJhRaise   int //第4轮金花以上加注的概率
	FourOverOverJhCompare int //第4轮金花以上比牌的概率
	FourOverOverJhAllin   int //第4轮金花以上全押的概率

	FirstRoundRaise1        int //第1轮等级1加注概率
	FirstRoundRaise2        int //第1轮等级2加注概率
	FirstRoundRaise3        int //第1轮等级3加注概率
	SecondRoundNotSeeRaise1 int //第2轮没看牌等级1加注概率
	SecondRoundNotSeeRaise2 int //第2轮没看牌等级2加注概率
	SecondRoundNotSeeRaise3 int //第2轮没看牌等级3加注概率
	SecondRoundSawRaise1    int //第2轮看了牌等级1加注概率
	SecondRoundSawRaise2    int //第2轮看了牌等级1加注概率
	SecondRoundSawRaise3    int //第2轮看了牌等级1加注概率

	ThirdRoundNotSeeRaise1 int //第2轮没看牌等级1加注概率
	ThirdRoundNotSeeRaise2 int //第2轮没看牌等级2加注概率
	ThirdRoundDzRaise1     int //第2轮没看牌等级2加注概率
	ThirdRoundDzRaise2     int //第2轮没看牌等级2加注概率
	ThirdRoundDzRaise3     int //第2轮没看牌等级2加注概率
	ThirdRoundSzRaise1     int //第2轮没看牌等级2加注概率
	ThirdRoundSzRaise2     int //第2轮没看牌等级2加注概率
	ThirdRoundSzRaise3     int //第2轮没看牌等级2加注概率
	ThirdRoundJhRaise1     int //第2轮没看牌等级2加注概率
	ThirdRoundJhRaise2     int //第2轮没看牌等级2加注概率
	ThirdRoundJhRaise3     int //第2轮没看牌等级2加注概率
	ThirdRoundOverJhRaise1 int //第2轮没看牌等级2加注概率
	ThirdRoundOverJhRaise2 int //第2轮没看牌等级2加注概率
	ThirdRoundOverJhRaise3 int //第2轮没看牌等级2加注概率

	FourRoundNotSeeRaise1 int //第2轮没看牌等级1加注概率
	FourRoundNotSeeRaise2 int //第2轮没看牌等级2加注概率
	FourRoundDzRaise1     int //第2轮没看牌等级2加注概率
	FourRoundDzRaise2     int //第2轮没看牌等级2加注概率
	FourRoundDzRaise3     int //第2轮没看牌等级2加注概率
	FourRoundSzRaise1     int //第2轮没看牌等级2加注概率
	FourRoundSzRaise2     int //第2轮没看牌等级2加注概率
	FourRoundSzRaise3     int //第2轮没看牌等级2加注概率
	FourRoundJhRaise1     int //第2轮没看牌等级2加注概率
	FourRoundJhRaise2     int //第2轮没看牌等级2加注概率
	FourRoundJhRaise3     int //第2轮没看牌等级2加注概率
	FourRoundOverJhRaise1 int //第2轮没看牌等级2加注概率
	FourRoundOverJhRaise2 int //第2轮没看牌等级2加注概率
	FourRoundOverJhRaise3 int //第2轮没看牌等级2加注概率
}

type GameConfig struct {
	Id           int64
	MinAction    int64   //底分额度
	RaiseAmount  []int64 //加注额度
	MaxRound     int32   //最高轮数
	MaxAllIn     int64   //全押最高额度
	AiZhengChang int
	AiJiJin      int
	AiWenZhong   int
	AiTouJi      int
}

type CheatConfig struct {
	MustLoseCheat  int //必输作弊率
	MustLoseRate   int //必输概率
	BigLoseCheat   int //大输作弊率
	BigLoseRate    int //大输概率
	SmallLoseCheat int //小输作弊率
	SmallLoseRate  int //小输概率
	SmallWinCheat  int //小赢作弊率
	SmallWinRate   int //小赢概率
	BigWinCheat    int //大赢作弊率
	BigWinRate     int //大赢概率
	MustWinCheat   int //必赢作弊率
	MustWinRate    int //必赢概率
}

type changeCardsConfig struct {
	Times      int64
	ChangeRate int
	Cheat      int
}

var ChangeCardsArr = make([]*changeCardsConfig, 0)
var Config = new(config)
var AiConfigArr = make([]*AiConfig, 0)
var GameConfigArr = make([]*GameConfig, 0)
var CheatConfigArr = make([]*CheatConfig, 0)

func LoadJsonConfig(_filename string, _config interface{}) (err error) {
	f, err := os.Open(_filename)
	if err == nil {
		defer f.Close()
		var fileInfo os.FileInfo
		fileInfo, err = f.Stat()
		if err == nil {
			bytes := make([]byte, fileInfo.Size())
			_, err = f.Read(bytes)
			if err == nil {
				BOM := []byte{0xEF, 0xBB, 0xBF}

				if bytes[0] == BOM[0] && bytes[1] == BOM[1] && bytes[2] == BOM[2] {
					bytes = bytes[3:]
				}
				err = json.Unmarshal(bytes, _config)
			}
		}
	}
	return
}
