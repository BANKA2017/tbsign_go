package _type

import "github.com/BANKA2017/tbsign_go/model"

type TcTieba struct {
	Tieba     *string `gorm:"column:tieba;not null" json:"tieba"`
	Status    *int32  `gorm:"column:status;not null" json:"status"`
	LastError *string `gorm:"column:last_error" json:"last_error"`
	model.TcTieba
}

type StatusStruct struct {
	Success  int64 `json:"success"`
	Failed   int64 `json:"failed"`
	Waiting  int64 `json:"waiting"`
	IsIgnore int64 `json:"ignore"` // `ignore` is the keyword of SQLite
}
