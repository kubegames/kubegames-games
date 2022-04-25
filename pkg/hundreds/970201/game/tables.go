package game

import (
	"game_buyu/rob_red/msg"
	"time"
)

//红包发送记录
type TableRed struct {
	Id         int64 //红包id
	SenderId   int64 `gorm:"index:sender_id"` //发送红包者id
	SenderName string
	Amount     int64          //红包金额
	Count      int64          //红包数量
	MineNum    int64          //雷号
	Route      []*msg.S2CAxis `gorm:"-"` //
	StartPoint int32          //红包飞进界面的点，一共5个点
	Time       int64          //红包发送时间
	Level      string
	Status     int32
	//TotalMineAmount int64 //中雷赔付总金额
}

func NewTableRed(senderId, amount, count, mineNum int64, startPoint, status int32, senderName, level string) *TableRed {
	return &TableRed{
		SenderName: senderName, SenderId: senderId, Amount: amount, Count: count, MineNum: mineNum, StartPoint: startPoint,
		Time: time.Now().Unix(), Level: level, Status: status,
	}
}

//抢红包记录
type TableRedRob struct {
	Id           int64 //抢包id
	RedId        int64 `gorm:"index:red_id"`
	Time         int64
	Level        string
	SenderName   string
	RobbedAmount int64
	RedAmount    int64
	IsMine       bool
	MineNum      int64
	Uid          int64
	MineAmount   int64 //中雷金额
}

func NewTableRedRob(uid, redId, robbedAmount, redAmount, mineNum int64, level, senderName string, isMine bool, mineAmount int64) *TableRedRob {
	return &TableRedRob{
		RedAmount: redAmount, RedId: redId, RobbedAmount: robbedAmount, MineNum: mineNum, Level: level, SenderName: senderName,
		IsMine: isMine, Time: time.Now().Unix(), Uid: uid, MineAmount: mineAmount,
	}
}
