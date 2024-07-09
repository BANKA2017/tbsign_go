package _function

import (
	"log"
	"strconv"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/gorm/clause"
)

var Options = make(map[string]string)
var CookieList = make(map[int32]_type.TypeCookie)
var FidList = make(map[string]int64)
var PluginListDB []model.TcPlugin

var PluginNameList = []string{"kd_growth", "ver4_ban", "ver4_rank", "ver4_ref"}
var PluginList = make(map[string]model.TcPlugin)

type ResetPwdStruct struct {
	Expire int64
	Value  string
	Time   int64
}

const ResetPwdMaxTimes = 5
const ResetPwdExpire = 60 * 30

var ResetPwdList = make(map[int32]*ResetPwdStruct)

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

func UpdateNow() {
	Now = time.Now().In(LocalTime)
}

func GetOption(keyName string) string {
	return Options[keyName]
}

func SetOption(keyName string, value string) error {
	err := GormDB.W.Model(&model.TcOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcOption{Name: keyName, Value: value}).Error

	if err == nil {
		Options[keyName] = value
	}
	return err
}

func DeleteOption(keyName string) error {
	delete(Options, keyName)
	return GormDB.W.Where("name = ?", keyName).Delete(&model.TcOption{}).Error
}

func GetUserOption(keyName string, uid string) string {
	var tmpUserOption model.TcUsersOption
	GormDB.R.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", uid, keyName).First(&tmpUserOption)
	return tmpUserOption.Value
}

func SetUserOption(keyName string, value string, uid string) error {
	numUID, _ := strconv.ParseInt(uid, 10, 64)
	return GormDB.W.Model(&model.TcUsersOption{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcUsersOption{UID: int32(numUID), Name: keyName, Value: value}).Error
}

func DeleteUserOption(keyName string, uid string) error {
	return GormDB.W.Where("uid = ? AND name = ?", uid, keyName).Delete(&model.TcUsersOption{}).Error
}

func GetCookie(pid int32) _type.TypeCookie {
	cookie, ok := CookieList[pid]
	if !ok {
		var cookieDB model.TcBaiduid
		GormDB.R.Model(&model.TcBaiduid{}).Where("id = ?", pid).First(&cookieDB)
		cookie.Tbs = GetTbs(cookieDB.Bduss)
		if cookie.Tbs == "" {
			return cookie
		}
		cookie.Bduss = cookieDB.Bduss
		cookie.Stoken = cookieDB.Stoken
		cookie.ID = cookieDB.ID
		cookie.Name = cookieDB.Name
		cookie.Portrait = cookieDB.Portrait
		cookie.UID = cookieDB.UID
		CookieList[pid] = cookie
	}

	return cookie
}

func GetFid(name string) int64 {
	fid, ok := FidList[name]
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
		FidList[name] = fid
	}
	return fid
}

func GetOptionsAndPluginList() {
	// get options
	var tmpOptions []model.TcOption

	GormDB.R.Find(&tmpOptions)
	for _, v := range tmpOptions {
		Options[v.Name] = v.Value
	}

	// get plugin list
	GormDB.R.Where("name in ?", PluginNameList).Find(&PluginListDB)

	for _, pluginStatus := range PluginListDB {
		PluginList[pluginStatus.Name] = pluginStatus
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
