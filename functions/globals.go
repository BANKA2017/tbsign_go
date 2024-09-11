package _function

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/mod/semver"
	"gorm.io/gorm/clause"
)

var Options sync.Map    //  make(map[string]string)
var CookieList sync.Map //= make(map[int32]_type.TypeCookie)
var FidList sync.Map    //= make(map[string]int64)

type ResetPwdStruct struct {
	Expire int64
	Value  string
	Time   int64
}

const ResetPwdMaxTimes = 5
const ResetPwdExpire = 60 * 30

var ResetPwdList sync.Map //= make(map[int32]*ResetPwdStruct)

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

func UpdateNow() {
	Now = time.Now().In(LocalTime)
}

func GetOption(keyName string) string {
	v, ok := Options.Load(keyName)
	if ok {
		return v.(string)
	} else {
		return ""
	}
}

func SetOption[T ~string | ~bool | ~int](keyName string, value T) error {
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

	v, ok := Options.Load(keyName)
	if ok && v == newValue {
		return nil
	}

	err := GormDB.W.Model(&model.TcOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcOption{Name: keyName, Value: newValue}).Error

	if err == nil {
		Options.Store(keyName, newValue)
	}
	return err
}

func DeleteOption(keyName string) error {
	err := GormDB.W.Where("name = ?", keyName).Delete(&model.TcOption{}).Error
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

func SetUserOption[T ~string | ~bool | ~int](keyName string, value T, uid string) error {
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

	return GormDB.W.Model(&model.TcUsersOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcUsersOption{UID: int32(numUID), Name: keyName, Value: newValue}).Error
}

func DeleteUserOption(keyName string, uid string) error {
	return GormDB.W.Where("uid = ? AND name = ?", uid, keyName).Delete(&model.TcUsersOption{}).Error
}

func UpdatePluginInfo(name string, version string, status bool, options string) error {
	err := GormDB.W.Model(&model.TcPlugin{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcPlugin{
		Name:    name,
		Ver:     version,
		Status:  status,
		Options: options,
	}).Error

	return err
}

func DeletePluginInfo(name string) error {
	return GormDB.W.Where("name = ?", name).Delete(&model.TcPlugin{}).Error
}

func GetCookie(pid int32, bduss_only ...bool) _type.TypeCookie {
	cookie, ok := CookieList.Load(pid)
	if !ok || cookie == nil {
		var _cookie _type.TypeCookie
		var cookieDB model.TcBaiduid
		GormDB.R.Model(&model.TcBaiduid{}).Where("id = ?", pid).First(&cookieDB)

		_cookie.ID = cookieDB.ID
		_cookie.Name = cookieDB.Name
		_cookie.Portrait = cookieDB.Portrait
		_cookie.UID = cookieDB.UID

		if len(bduss_only) > 0 && bduss_only[0] {
			_cookie.Bduss = cookieDB.Bduss
			_cookie.Stoken = cookieDB.Stoken
			return _cookie
		}

		tbsResponse, err := GetTbs(cookieDB.Bduss)
		if err != nil || tbsResponse.IsLogin == 0 {
			return _cookie
		}
		_cookie.Tbs = tbsResponse.Tbs
		_cookie.Stoken = cookieDB.Stoken
		_cookie.Bduss = cookieDB.Bduss
		CookieList.Store(pid, _cookie)
		return _cookie
	}

	return cookie.(_type.TypeCookie)
}

func GetFid(name string) int64 {
	fid, ok := FidList.Load(name)
	if !ok || fid == nil {
		// find in db
		var tiebaInfo model.TcTieba
		GormDB.R.Model(&model.TcTieba{}).Where("tieba = ? AND fid IS NOT NULL AND fid != ''", name).First(&tiebaInfo)
		_fid := int64(tiebaInfo.Fid)
		if _fid == 0 {
			forumNameInfo, err := GetForumNameShare(name)
			if err != nil {
				log.Println("fid:", err)
			}
			_fid = int64(forumNameInfo.Data.Fid)
		}
		FidList.Store(name, _fid)
		return _fid
	}
	return fid.(int64)
}

func InitOptions() {
	// get db options
	var tmpOptions []model.TcOption

	GormDB.R.Find(&tmpOptions)

	// sync options
	defaultOptionsCopy := make(map[string]string)
	if len(tmpOptions) != len(assets.DefaultOptions) {
		for k, v := range assets.DefaultOptions {
			defaultOptionsCopy[k] = v
		}
	}

	for _, v := range tmpOptions {
		Options.Store(v.Name, v.Value)
		delete(defaultOptionsCopy, v.Name)
	}

	// sync options
	for k, v := range defaultOptionsCopy {
		SetOption(k, v)
	}
}

// for GMT+8
func LocaleTimeDiff(hour int64) int64 {
	targetTime := time.Date(Now.Year(), Now.Month(), Now.Day(), int(hour), 0, 0, 0, LocalTime)

	if targetTime.After(Now) {
		targetTime = targetTime.Add(-24 * time.Hour)
	}

	return targetTime.Unix()
}

func VariableWrapper[T any](anyValue T) T {
	return anyValue
}

func VariablePtrWrapper[T any](anyValue T) *T {
	return &anyValue
}

func GetGravatarLink(email string) string {
	return "https://www.gravatar.com/avatar/" + Sha256([]byte(email))
}

func GetSemver(cur, ver2 string) string {
	if semver.Compare(cur, ver2) == 1 {
		return cur
	} else {
		return ver2
	}
}
