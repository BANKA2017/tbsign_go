// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameTcKdGrowth = "tc_kd_growth"

// TcKdGrowth mapped from table <tc_kd_growth>
type TcKdGrowth struct {
	ID     int64  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID    int64  `gorm:"column:uid;not null" json:"uid"`
	Pid    int64  `gorm:"column:pid;not null" json:"pid"`
	Status string `gorm:"column:status" json:"status"`
	Log    string `gorm:"column:log" json:"log"`
	Date   int32  `gorm:"column:date;not null;default:0" json:"date"`
}

// TableName TcKdGrowth's table name
func (*TcKdGrowth) TableName() string {
	return TableNameTcKdGrowth
}
