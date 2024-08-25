package _type

import "github.com/BANKA2017/tbsign_go/model"

type TcTieba struct {
	Tieba     *string `gorm:"column:tieba;not null" json:"tieba"`
	Status    *int32  `gorm:"column:status;not null" json:"status"`
	LastError *string `gorm:"column:last_error" json:"last_error"`
	model.TcTieba
}
