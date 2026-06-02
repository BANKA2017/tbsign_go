package _api

import (
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/assets"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/jellydator/ttlcache/v3"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"golang.org/x/sync/singleflight"
)

var RoleList = []string{_function.RoleDeleted, _function.RoleBanned, _function.RoleUser, _function.RoleVIP, _function.RoleAdmin}

var IndependentFEPath = []string{"/robots.txt", "/favicon.ico", "/icp.jsonp", "/site.jsonp"}

func echoRobots(c echo.Context) error {
	if val := _function.GetOption("go_robots_txt"); val != "" {
		return c.String(http.StatusOK, val)
	}

	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

var FaviconBinCache []byte
var FaviconCacheTime time.Time
var NoFavicon = false

func echoFavicon(c echo.Context) (err error) {
	if len(FaviconBinCache) > 0 {
		c.Response().Header().Set("Last-Modified", FaviconCacheTime.Format(http.TimeFormat))
		return c.Blob(http.StatusOK, "image/x-icon", FaviconBinCache)
	}

	if NoFavicon {
		return _function.EchoNoContent(c)
	}

	// load from options
	optionVal := _function.GetOption("go_favicon")
	if val, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(optionVal, "data:image/x-icon;base64,")); err == nil && len(val) > 0 {
		FaviconBinCache = val
		FaviconCacheTime = time.Now()
		c.Response().Header().Set("Last-Modified", FaviconCacheTime.Format(http.TimeFormat))
		return c.Blob(http.StatusOK, "image/x-icon", FaviconBinCache)
	}

	// load from embedded
	fe, _ := fs.Sub(assets.EmbeddedFrontend, "dist")
	FaviconBinCache, err = fs.ReadFile(fe, "favicon.ico")
	if err != nil {
		NoFavicon = true
		return _function.EchoNoContent(c)
	}
	c.Response().Header().Set("Last-Modified", share.BuildAtTime.Format(http.TimeFormat))
	return c.Blob(http.StatusOK, "image/x-icon", FaviconBinCache)
}

const authorizationLengthMax = 100

func verifyAuthorization(authorization string) (string, string) {
	authorization = strings.TrimSpace(authorization)

	if authorization == "" || len(authorization) > authorizationLengthMax {
		return "0", _function.RoleGuest
	}

	token := strings.SplitN(authorization, ":", 3)
	// TODO static target
	if len(token) != 3 {
		return "0", _function.RoleGuest
	}

	uid, err := strconv.ParseInt(token[0], 10, 64)
	if err != nil || uid <= 0 {
		return "0", _function.RoleGuest
	}

	strUID := strconv.Itoa(int(uid))

	now := time.Now().Unix()

	expiredAt, err := strconv.ParseInt(token[2], 10, 64)
	if err != nil || expiredAt <= now {
		return "0", _function.RoleGuest
	}

	savedSessionExpiredAt := GetSessionExpiredAt(strUID)

	if savedSessionExpiredAt <= now || savedSessionExpiredAt != expiredAt {
		return "0", _function.RoleGuest
	}

	dbPwd := _function.GetPassword(int(uid))

	byteToken, err := base64.RawURLEncoding.DecodeString(token[1])
	// len(hmac256)==32
	if err != nil || len(byteToken) != 32 {
		return "0", _function.RoleGuest
	}

	if dbPwd != "" {
		if !hmac.Equal(HmacSessionToken(strUID, dbPwd, strconv.FormatInt(savedSessionExpiredAt, 10)), byteToken) {
			return "0", _function.RoleGuest
		}
		var accountInfo model.TcUser
		if err := _function.GormDB.R.Where("id = ?", uid).First(&accountInfo).Error; err == nil && accountInfo.ID > 0 {
			return strconv.FormatInt(int64(accountInfo.ID), 10), accountInfo.Role
		}

		return "0", _function.RoleGuest
	}

	return "0", _function.RoleGuest
}

// func legacyTokenBuilder(uid int32, password string) string {
// 	return _function.Base64URLEncode([]byte(strconv.Itoa(int(uid)) + ":" + hex.EncodeToString(_function.GenHMAC256([]byte(password), []byte(strconv.Itoa(int(uid))+password)))))
// }

// type HttpAuthRefreshTokenMapItemStruct struct {
// 	Content  string
// 	ExpireAt int64
// }

// var HttpAuthRefreshTokenMap _function.KV[int, string]

func HmacSessionToken(uid, password, expiredAt string) []byte {
	return _function.GenHMAC256([]byte(password), []byte(uid+":"+expiredAt))
}

func tokenBuilder(uid int, password string) (string, int64, int64) {
	// expire
	strCookieExpire := _function.GetOption("cktime")
	numberCookieExpire, err := strconv.ParseInt(strCookieExpire, 10, 64)
	if err != nil {
		numberCookieExpire = 10 * 60
	} else if numberCookieExpire > 10*24*60*60 {
		numberCookieExpire = 10 * 24 * 60 * 60
	} else if numberCookieExpire < 30 {
		numberCookieExpire = 30
	}
	expiredAt := time.Now().Add(time.Duration(numberCookieExpire) * time.Second).Unix()
	strExpiredAt := strconv.FormatInt(expiredAt, 10)

	token := base64.RawURLEncoding.EncodeToString(HmacSessionToken(strconv.Itoa(uid), password, strExpiredAt))

	// HttpAuthRefreshTokenMap.Store(int(uid), token, numberCookieExpire)

	return strconv.Itoa(uid) + ":" + token + ":" + strExpiredAt, expiredAt, numberCookieExpire
}

var ExpiredTimeCache = _function.NewKV(
	ttlcache.WithTTL[string, int64](time.Hour),
	ttlcache.WithCapacity[string, int64](100),
)

func GetSessionExpiredAt(uid string) int64 {
	if t, ok := ExpiredTimeCache.Load(uid); ok {
		return t
	}
	strT := _function.GetUserOption("session_expired_at", uid)

	t, _ := strconv.ParseInt(strT, 10, 64)
	ExpiredTimeCache.Store(uid, t, int64(ttlcache.DefaultTTL))
	return t
}

func UpdateSessionExpiredAt(uid string, t int64) (int64, error) {
	if err := _function.SetUserOption("session_expired_at", strconv.FormatInt(t, 10), uid); err != nil {
		return 0, err
	}
	ExpiredTimeCache.Store(uid, t, int64(ttlcache.DefaultTTL))
	return t, nil
}

func DeleteSessionExpiredAt(uid string) (bool, error) {
	if err := _function.DeleteUserOption("session_expired_at", uid); err != nil {
		return false, err
	}
	ExpiredTimeCache.Delete(uid)
	return true, nil
}

var resetPasswordVerifyCodeByteLength int64 = 6

var resetPasswordVerifyCodeLength = int(math.Ceil(float64(resetPasswordVerifyCodeByteLength*4) / float64(3)))

func ResetMessageBuilder(uid int32, forceMode bool) *_function.VerifyCodeStruct {
	_v, ok := _function.VerifyCodeList.LoadCode("reset_password", uid)
	var v *_function.VerifyCodeStruct

	if !ok || _v == nil {
		v = &_function.VerifyCodeStruct{
			Expire: time.Now().Add(time.Second * time.Duration(_function.ResetPwdExpire)).Unix(),
		}
	} else {
		v = _v
	}

	v.Type = "reset_password"

	if forceMode || v.ResetTime < _function.ResetPwdMaxTimes {
		// init a callback code
		_code, _ := _function.RandomTokenBuilder(resetPasswordVerifyCodeByteLength)
		code := _function.Base64URLEncode(_code)

		v.ResetTime += 1
		v.VerifyCode = _function.RandomEmoji()
		v.TryTime = 0

		if len(code) != resetPasswordVerifyCodeLength {
			return v
		}

		v.Value = code
	}

	_function.VerifyCodeList.StoreCode("reset_password", uid, v)

	return v
}

func SendResetMessage(uid int32, pushType string, forceMode bool) (string, error) {

	v := ResetMessageBuilder(uid, forceMode)
	if len(v.Value) != resetPasswordVerifyCodeLength {
		return "", errors.New("验证码生成失败，请重试")
	}

	if !forceMode && (v.ResetTime >= _function.ResetPwdMaxTimes || v.TryTime >= _function.ResetPwdMaxTimes) {
		if len(v.Value) == resetPasswordVerifyCodeLength {
			tmpValue, _ := _function.RandomTokenBuilder(48)
			v.Value = _function.Base64URLEncode(tmpValue)
			_function.VerifyCodeList.StoreCode("reset_password", uid, v)
		}
		return "", errors.New("已超过最大验证次数，请稍后再试")
	}

	mailObject := _function.PushMessageTemplateResetPassword(v.VerifyCode, v.Value)

	// user default message type
	userMessageType := "email"
	if pushType != "" && slices.Contains(_function.MessageTypeList, pushType) {
		userMessageType = pushType
	} else {
		localPushType := _function.GetUserOption("go_message_type", strconv.Itoa(int(uid)))
		if slices.Contains(_function.MessageTypeList, localPushType) {
			userMessageType = localPushType
		}
	}

	err := _function.SendMessage(userMessageType, uid, mailObject.Title, mailObject.Body)
	if err != nil {
		slog.Error("passport.reset-password", "uid", uid, "error", err)
		return "", errors.New("消息发送失败")
	}

	return v.VerifyCode, nil
}

func IsArrayMode(c echo.Context) bool {
	arrayModeValue := c.QueryParam("array_mode")
	return arrayModeValue != "" && arrayModeValue != "0" && arrayModeValue != "false"
}

var RequestSingleFlight singleflight.Group

func GetBodyMap(c echo.Context) (map[string][]string, error) {
	request := c.Request()

	if err := request.ParseForm(); err != nil {
		return nil, err
	}

	bodyMap := make(map[string][]string)
	for k, vs := range request.Form {
		if len(vs) > 0 {
			bodyMap[k] = vs
		} else {
			bodyMap[k] = []string{""}
		}
	}

	return bodyMap, nil
}
