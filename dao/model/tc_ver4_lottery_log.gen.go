// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameTcVer4LotteryLog = "tc_ver4_lottery_log"

// TcVer4LotteryLog mapped from table <tc_ver4_lottery_log>
type TcVer4LotteryLog struct {
	ID     int32  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID    int32  `gorm:"column:uid;not null" json:"uid"`
	Pid    int32  `gorm:"column:pid;not null" json:"pid"`
	Result string `gorm:"column:result;not null" json:"result"`
	Prize  string `gorm:"column:prize;not null" json:"prize"`
	Date   int32  `gorm:"column:date;not null;default:0" json:"date"`
}

// TableName TcVer4LotteryLog's table name
func (*TcVer4LotteryLog) TableName() string {
	return TableNameTcVer4LotteryLog
}