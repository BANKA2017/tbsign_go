// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameTcTieba = "tc_tieba"

// TcTieba mapped from table <tc_tieba>
type TcTieba struct {
	ID        int32  `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	UID       int32  `gorm:"column:uid;not null" json:"uid"`
	Pid       int32  `gorm:"column:pid;not null" json:"pid"`
	Fid       int32  `gorm:"column:fid;not null" json:"fid"`
	Tieba     string `gorm:"column:tieba;not null" json:"tieba"`
	No        bool   `gorm:"column:no;not null" json:"no"`
	Status    int32  `gorm:"column:status;not null" json:"status"`
	Latest    int32  `gorm:"column:latest;not null" json:"latest"`
	LastError string `gorm:"column:last_error" json:"last_error"`
}

// TableName TcTieba's table name
func (*TcTieba) TableName() string {
	return TableNameTcTieba
}
