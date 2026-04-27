package _function

import (
	"github.com/kdnetwork/code-snippet/go/db"
	"gorm.io/gorm/logger"
)

var GormDB = new(db.GormDBCtx)

func UpdateGormLogger(newLogger logger.Interface) {
	GormDB.R.Logger = newLogger
	GormDB.W.Logger = newLogger
}
