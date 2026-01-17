package _function

import (
	crypto_rand "crypto/rand"
	"encoding/base64"
	"log"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/jellydator/ttlcache/v3"
	"golang.org/x/mod/semver"
	"golang.org/x/sync/singleflight"

	"maps"
	_ "time/tzdata"
)

var Options = NewKV[string, string]() //  make(map[string]string)
var CookieList = NewKV(
	ttlcache.WithCapacity[int32, _type.TypeCookie](100),
) //= make(map[int32]_type.TypeCookie)
var FidList = NewKV(
	ttlcache.WithCapacity[string, int64](500),
) //= make(map[string]int64)

const ResetPwdMaxTimes = 5
const ResetPwdExpire = 60 * 5 // 5 mins

// Tieba works in GMT+8
var LocalTime, _ = time.LoadLocation("Asia/Shanghai")

var Now = time.Now().In(LocalTime)

func UpdateNow() {
	Now = time.Now().In(LocalTime)
}

var syncCookieTasks singleflight.Group

// ext [bduss_only, force_sync]
func GetCookie(pid int32, ext ...bool) _type.TypeCookie {
	cookie, ok := CookieList.Load(pid)
	bdussOnly := len(ext) >= 1 && ext[0]
	forceSync := len(ext) >= 2 && ext[1]

	if !ok || forceSync {
		_cookie, _, _ := syncCookieTasks.Do(strconv.Itoa(int(pid))+When(forceSync, "1", "0"), func() (any, error) {
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

			if bdussOnly {
				_cookie.Bduss = cookieDB.Bduss
				_cookie.Stoken = cookieDB.Stoken
				_cookie.IsLogin = true
				return _cookie, nil
			}

			tbsResponse, err := GetTbs(cookieDB.Bduss)
			if err != nil {
				return _cookie, nil
			}
			_cookie.IsLogin = tbsResponse.IsLogin != 0
			_cookie.Tbs = tbsResponse.Tbs
			_cookie.Stoken = cookieDB.Stoken
			_cookie.Bduss = cookieDB.Bduss
			CookieList.Store(pid, _cookie, 60*60*4)
			return _cookie, nil
		})
		return _cookie.(_type.TypeCookie)
	}

	return cookie
}

var syncFidTasks singleflight.Group

func GetFid(name string) int64 {
	fid, ok := FidList.Load(name)
	if !ok || fid == 0 {
		_fid, _, _ := syncFidTasks.Do(name, func() (any, error) {
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
			FidList.Store(name, _fid, 60*60*4)
			return _fid, nil
		})
		return _fid.(int64)
	}
	return fid
}

func InitOptions() {
	// get db options
	var tmpOptions []*model.TcOption

	GormDB.R.Find(&tmpOptions)

	// sync options
	defaultOptionsCopy := make(map[string]string)
	if len(tmpOptions) != len(assets.DefaultOptions) {
		maps.Copy(defaultOptionsCopy, assets.DefaultOptions)
	}

	for _, v := range tmpOptions {
		Options.Store(v.Name, v.Value, -1)
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

func VPtr[T any](anyValue T) *T {
	return &anyValue
}

func GetGravatarLink(email string) string {
	return "https://www.gravatar.com/avatar/" + Sha256([]byte(email))
}

func NewerSemver(cur, ver2 string) string {
	cver := When(strings.HasPrefix(strings.ToLower(cur), "v"), "", "v") + strings.ToLower(cur)
	nver := When(strings.HasPrefix(strings.ToLower(ver2), "v"), "", "v") + strings.ToLower(ver2)

	if semver.Compare(cver, nver) == 1 {
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

	var resStr []string
	for _, v := range randNum[:3] {
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

func Base64URLEncode(originalBuffer []byte) string {
	return base64.RawURLEncoding.EncodeToString(originalBuffer)
}

func Base64URLDecode(originalBuffer string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.ReplaceAll(originalBuffer, "=", ""))
}
