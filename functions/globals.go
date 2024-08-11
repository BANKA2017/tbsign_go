package _function

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/gorm/clause"
)

var Options sync.Map    //  make(map[string]string)
var CookieList sync.Map //= make(map[int32]_type.TypeCookie)
var FidList sync.Map    //= make(map[string]int64)
var PluginListDB []model.TcPlugin

var PluginNameList = []string{"kd_growth", "ver4_ban", "ver4_rank", "ver4_ref"}
var PluginList sync.Map //= make(map[string]model.TcPlugin)

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

	err := GormDB.W.Model(&model.TcOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcOption{Name: keyName, Value: newValue}).Error

	if err == nil {
		Options.Store(keyName, newValue)
	}
	return err
}

func DeleteOption(keyName string) error {
	Options.Delete(keyName)
	return GormDB.W.Where("name = ?", keyName).Delete(&model.TcOption{}).Error
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

func GetCookie(pid int32) _type.TypeCookie {
	cookie, ok := CookieList.Load(pid)
	if !ok {
		var _cookie _type.TypeCookie
		var cookieDB model.TcBaiduid
		GormDB.R.Model(&model.TcBaiduid{}).Where("id = ?", pid).First(&cookieDB)
		_cookie.Tbs = GetTbs(cookieDB.Bduss)
		if _cookie.Tbs == "" {
			return _cookie
		}
		_cookie.Bduss = cookieDB.Bduss
		_cookie.Stoken = cookieDB.Stoken
		_cookie.ID = cookieDB.ID
		_cookie.Name = cookieDB.Name
		_cookie.Portrait = cookieDB.Portrait
		_cookie.UID = cookieDB.UID
		CookieList.Store(pid, cookie)
		return _cookie
	}

	return cookie.(_type.TypeCookie)
}

func GetFid(name string) int64 {
	fid, ok := FidList.Load(name)
	if !ok {
		// find in db
		var tiebaInfo model.TcTieba
		GormDB.R.Model(&model.TcTieba{}).Where("tieba = ? AND fid IS NOT NULL AND fid != ''", name).First(&tiebaInfo)
		fid = int64(tiebaInfo.Fid)
		if fid == 0 {
			forumNameInfo, err := GetForumNameShare(name)
			if err != nil {
				log.Println("fid:", err)
			}
			fid = int64(forumNameInfo.Data.Fid)
		}
		FidList.Store(name, fid)
	}
	return fid.(int64)
}

func GetOptionsAndPluginList() {
	// get options
	var tmpOptions []model.TcOption

	GormDB.R.Find(&tmpOptions)
	for _, v := range tmpOptions {
		Options.Store(v.Name, v.Value)
	}

	// get plugin list
	GormDB.R.Where("name in ?", PluginNameList).Find(&PluginListDB)

	for _, pluginStatus := range PluginListDB {
		PluginList.Store(pluginStatus.Name, pluginStatus)
	}
}

// for GMT+8
func TodayBeginning() int64 {
	if Now.Local().Hour() >= 8 {
		return Now.Unix() - Now.Unix()%86400 - 8*3600
	}
	return Now.Unix() - Now.Unix()%86400 + 86400 - 8*3600
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
