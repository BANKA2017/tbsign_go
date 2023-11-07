package _function

import (
	"log"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/gorm"
)

var Options []model.TcOption
var CookieList = make(map[int32]_type.TypeCookie)
var FidList = make(map[string]int64)
var PluginListDB []model.TcPlugin
var PluginList = make(map[string]bool)
var GormDB *gorm.DB

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

func UpdateNow() {
	Now = time.Now().In(LocalTime)
}

func GetOption(keyName string) string {
	for _, v := range Options {
		if v.Name == keyName {
			return v.Value
		}
	}
	return ""
}

func GetUserOption(keyName string, uid string) string {
	var tmpUserOption model.TcUsersOption
	GormDB.Model(&model.TcUsersOption{}).Where("uid = ? AND name = ?", uid, keyName).First(&tmpUserOption)
	return tmpUserOption.Value
}

func SetOption(keyName string, value string) {
	GormDB.Model(&model.TcOption{}).Where("name = ?", keyName).Update("value", value)
	for i := range Options {
		if Options[i].Name == keyName {
			Options[i].Value = value
			break
		}
	}
}

func GetCookie(pid int32) _type.TypeCookie {
	cookie, ok := CookieList[pid]
	if !ok {
		var cookieDB model.TcBaiduid
		GormDB.Model(&model.TcBaiduid{}).Where("id = ?", pid).First(&cookieDB)
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
		GormDB.Model(&model.TcTieba{}).Where("tieba = ? AND fid IS NOT NULL AND fid != ''", name).First(&tiebaInfo)
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
	GormDB.Find(&Options)

	// get plugin list
	GormDB.Find(&PluginListDB)

	for _, pluginStatus := range PluginListDB {
		PluginList[pluginStatus.Name] = pluginStatus.Status
	}
}

// for GMT+8
func TodayBeginning() int64 {
	if Now.Local().Hour() >= 8 {
		return Now.Unix() - Now.Unix()%86400 - 8*3600
	}
	return Now.Unix() - Now.Unix()%86400 + 86400 - 8*3600
}
