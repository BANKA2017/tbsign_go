// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameTcUsersOption = "tc_users_options"

// TcUsersOption mapped from table <tc_users_options>
type TcUsersOption struct {
	UID   int32  `gorm:"column:uid;primaryKey" json:"uid"`
	Name  string `gorm:"column:name;primaryKey" json:"name"`
	Value string `gorm:"column:value;not null" json:"value"`
}

// TableName TcUsersOption's table name
func (*TcUsersOption) TableName() string {
	return TableNameTcUsersOption
}