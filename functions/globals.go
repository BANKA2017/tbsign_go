package _function

import (
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"gorm.io/gorm"
)

var Options []model.TcOption
var CookieList = make(map[int32]_type.TypeCookie)
var GormDB *gorm.DB

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

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

// for GMT+8
func TodayBeginning() int64 {
	if Now.Local().Hour() >= 8 {
		return Now.Unix() - Now.Unix()%86400 - 8*3600
	}
	return Now.Unix() - Now.Unix()%86400 + 86400 - 8*3600
}
