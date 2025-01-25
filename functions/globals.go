package _function

import (
	crypto_rand "crypto/rand"
	"log"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/mod/semver"

	_ "time/tzdata"
)

var Options sync.Map    //  make(map[string]string)
var CookieList sync.Map //= make(map[int32]_type.TypeCookie)
var FidList sync.Map    //= make(map[string]int64)

const ResetPwdMaxTimes = 5
const ResetPwdExpire = 60 * 5 // 5 mins

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

func UpdateNow() {
	Now = time.Now().In(LocalTime)
}

func GetCookie(pid int32, bduss_only ...bool) _type.TypeCookie {
	cookie, ok := CookieList.Load(pid)
	if !ok || cookie == nil {
		var _cookie _type.TypeCookie
		var cookieDB model.TcBaiduid
		GormDB.R.Model(&model.TcBaiduid{}).Where("id = ?", pid).Take(&cookieDB)

		if len(share.DataEncryptKeyByte) > 0 {
			decryptedBDUSS, _ := AES256GCMDecrypt(cookieDB.Bduss, share.DataEncryptKeyByte)
			cookieDB.Bduss = string(decryptedBDUSS)

			decryptedStoken, _ := AES256GCMDecrypt(cookieDB.Stoken, share.DataEncryptKeyByte)
			cookieDB.Stoken = string(decryptedStoken)
		}

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
		GormDB.R.Model(&model.TcTieba{}).Where("tieba = ? AND fid IS NOT NULL AND fid != ''", name).Take(&tiebaInfo)
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
	var tmpOptions []*model.TcOption

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

func VerifyURL(_url string) bool {
	_, err := url.ParseRequestURI(_url)
	return err == nil
}

func RandomEmoji() string {
	emojiMap := []string{"💻", "✅", "➡️", "🎉", "🤖", "🐱", "⚙️", "😊", "📌", "✒️", "⌛", "🔔"}
	randNum := rand.Perm(len(emojiMap))

	resStr := []string{}
	for _, v := range randNum {
		resStr = append(resStr, emojiMap[v])
	}
	return strings.Join(resStr, "")
}

func RandomTokenBuilder(n int64) ([]byte, error) {
	token := make([]byte, n)

	_, err := crypto_rand.Read(token)
	return token, err
}

func TinyIntToBool(t int) bool {
	return t != 0
}
func BoolToTinyInt(b bool) int {
	return When(b, 1, 0)
}

func When[T any](c bool, d1, d2 T) T {
	if c {
		return d1
	} else {
		return d2
	}
}
