package _function

import (
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var SettingsKeys = []string{"ann", "system_url", "stop_reg", "enable_reg", "yr_reg", "cktime", "sign_mode", "sign_hour", "cron_limit", "sign_sleep", "retry_max", "mail_name", "mail_yourname", "mail_host", "mail_port", "mail_secure", "mail_auth", "mail_smtpname", "mail_smtppw", "go_forum_sync_policy", "go_ntfy_addr", "go_bark_addr"} // "system_name", "system_keywords", "system_description"

var SettingsFilter = SettingsKeys

func GetOption(keyName string) string {
	v, ok := Options.Load(keyName)
	if ok {
		return v.(string)
	} else {
		return ""
	}
}

func SetOption[T ~string | ~bool | ~int](keyName string, value T, ext ...any) error {
	newValue := ""
	switch any(value).(type) {
	case string:
		newValue = strings.TrimSpace(any(value).(string))
	case bool:
		if any(value).(bool) {
			newValue = "1"
		} else {
			newValue = "0"
		}
	case int:
		newValue = strconv.Itoa(any(value).(int))
	}

	v, ok := Options.Load(keyName)
	if ok && v == newValue {
		return nil
	}

	_sql := GormDB.W
	if len(ext) > 0 {
		if tx, ok := ext[0].(*gorm.DB); ok {
			_sql = tx
		}
	}
	err := _sql.Model(&model.TcOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcOption{Name: keyName, Value: newValue}).Error

	if err == nil {
		Options.Store(keyName, newValue)
	}
	return err
}

func DeleteOption(keyName string, ext ...any) error {
	_sql := GormDB.W
	if len(ext) > 0 {
		if tx, ok := ext[0].(*gorm.DB); ok {
			_sql = tx
		}
	}
	err := _sql.Where("name = ?", keyName).Delete(&model.TcOption{}).Error
	if err == nil {
		Options.Delete(keyName)
	}
	return err
}

func GetUserOption(keyName string, uid string) string {
	var tmpUserOption model.TcUsersOption
	GormDB.R.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", uid, keyName).First(&tmpUserOption)
	return tmpUserOption.Value
}

func SetUserOption[T ~string | ~bool | ~int](keyName string, value T, uid string, ext ...any) error {
	numUID, _ := strconv.ParseInt(uid, 10, 64)
	newValue := ""
	switch any(value).(type) {
	case string:
		newValue = any(value).(string)
	case bool:
		if any(value).(bool) {
			newValue = "1"
		} else {
			newValue = "0"
		}
	case int:
		newValue = strconv.Itoa(any(value).(int))
	}

	_sql := GormDB.W
	if len(ext) > 0 {
		if tx, ok := ext[0].(*gorm.DB); ok {
			_sql = tx
		}
	}
	return _sql.Model(&model.TcUsersOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcUsersOption{UID: int32(numUID), Name: keyName, Value: newValue}).Error
}

func DeleteUserOption(keyName string, uid string, ext ...any) error {
	_sql := GormDB.W
	if len(ext) > 0 {
		if tx, ok := ext[0].(*gorm.DB); ok {
			_sql = tx
		}
	}
	return _sql.Where("uid = ? AND name = ?", uid, keyName).Delete(&model.TcUsersOption{}).Error
}
