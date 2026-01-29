package _function

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var SettingsKeys = []string{"ann", "icp", "system_url", "stop_reg", "enable_reg", "yr_reg", "cktime", "sign_mode", "sign_hour", "cron_limit", "sign_sleep", "retry_max", "mail_name", "mail_yourname", "mail_host", "mail_port", "mail_secure", "mail_auth", "mail_smtpname", "mail_smtppw", "go_forum_sync_policy", "go_ntfy_addr", "go_bark_addr", "go_pushdeer_addr", "go_export_personal_data", "go_import_personal_data", "go_re_check_in_max_interval", "sign_multith", "go_daily_report_hour", "bduss_num"} // "system_name", "system_keywords", "system_description", "tb_max"

var SettingsFilter = SettingsKeys

func GetOption(keyName string) string {
	if v, ok := Options.Load(keyName); ok {
		return v
	}

	return ""
}

type OptionExt struct {
	Tx         *gorm.DB
	EncryptKey *[]byte
}

func SetOption[T ~string | ~bool | ~int](keyName string, value T, ext ...OptionExt) error {
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
		if ext[0].Tx != nil {
			_sql = ext[0].Tx
		}
	}
	err := _sql.Model(&model.TcOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcOption{Name: keyName, Value: newValue}).Error

	if err == nil {
		Options.Store(keyName, newValue, -1)
	}
	return err
}

func DeleteOption(keyName string, ext ...OptionExt) error {
	_sql := GormDB.W
	if len(ext) > 0 {
		if ext[0].Tx != nil {
			_sql = ext[0].Tx
		}
	}
	err := _sql.Where("name = ?", keyName).Delete(&model.TcOption{}).Error
	if err == nil {
		Options.Delete(keyName)
	}
	return err
}

var CanEncryptUserOption = []string{"go_bark_key", "go_ntfy_topic", "go_pushdeer_key"}

func GetUserOption(keyName string, uid string, ext ...OptionExt) string {
	var tmpUserOption model.TcUsersOption
	err := GormDB.R.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", uid, keyName).Take(&tmpUserOption).Error
	isNotFound := err != nil && errors.Is(err, gorm.ErrRecordNotFound)
	if isNotFound {
		if GetOption("go_create_user_option_if_not_exist") == "1" {
			SetUserOption(keyName, "", uid)
		}

		return ""
	}
	if len(ext) > 0 {
		if ext[0].EncryptKey != nil {
			if len(*ext[0].EncryptKey) == 32 && tmpUserOption.Value != "" {
				newDecryptedValue, err := AES256GCMDecrypt(tmpUserOption.Value, *ext[0].EncryptKey)
				if err == nil && newDecryptedValue != nil {
					tmpUserOption.Value = string(newDecryptedValue)
				}
			}
		}
	}
	return tmpUserOption.Value
}

func SetUserOption[T ~string | ~bool | ~int](keyName string, value T, uid string, ext ...OptionExt) error {
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
		if ext[0].Tx != nil {
			_sql = ext[0].Tx
		}

		if ext[0].EncryptKey != nil {
			newEncryptedValue, err := AES256GCMEncrypt(newValue, *ext[0].EncryptKey)
			if err == nil && newEncryptedValue != nil {
				newValue = Base64URLEncode(newEncryptedValue)
			}
		}
	}
	return _sql.Model(&model.TcUsersOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcUsersOption{UID: int32(numUID), Name: keyName, Value: newValue}).Error
}

func DeleteUserOption(keyName string, uid string, ext ...OptionExt) error {
	_sql := GormDB.W
	if len(ext) > 0 {
		if ext[0].Tx != nil {
			_sql = ext[0].Tx
		}
	}
	return _sql.Where("uid = ? AND name = ?", uid, keyName).Delete(&model.TcUsersOption{}).Error
}

// option validation rule
// by ChatGPT

type OptionRule struct {
	Enum []string
	Min  *int64
	Max  *int64
	// Length    *int64
	IsURL     bool
	Custom    func(string) error
	Transform func(string) string
}

func ValidateOptionValue(val string, rule *OptionRule) (string, error) {
	if val == "" || rule == nil {
		return val, nil
	}
	// enum
	if len(rule.Enum) > 0 {
		if !slices.Contains(rule.Enum, val) {
			return "", fmt.Errorf("invalid value `%s`", val)
		}
	}

	// url
	if rule.IsURL {
		if !VerifyURL(val) {
			return "", fmt.Errorf("invalid URL `%s`", val)
		}
	}

	// num value
	if rule.Min != nil || rule.Max != nil {
		num, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number `%s`", val)
		}
		if rule.Min != nil && num < *rule.Min {
			return "", fmt.Errorf("value < min (%d)", *rule.Min)
		}
		if rule.Max != nil && num > *rule.Max {
			return "", fmt.Errorf("value > max (%d)", *rule.Max)
		}
	}

	// if rule.Length != nil {
	// 	if int64(len(val)) != *rule.Length {
	// 		return "", fmt.Errorf("length != (%d)", *rule.Length)
	// 	}
	// }

	// custom
	if rule.Custom != nil {
		if err := rule.Custom(val); err != nil {
			return "", err
		}
	}

	// transform
	if rule.Transform != nil {
		val = rule.Transform(val)
	}

	return val, nil
}
