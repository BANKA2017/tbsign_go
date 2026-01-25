package _api

import (
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/jellydator/ttlcache/v3"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
)

var PreCheckWhiteListWithoutFE = []string{
	"/",
	"/*",
	"/favicon.ico",
	"/robots.txt",
	"/passport/login",
	"/passport/signup",
	"/passport/reset/password",
	// "/tools/userinfo/tieba_uid/:tiebauid",
	// "/tools/userinfo/panel/:query_type/:user_value",
	// "/tools/tieba/fname_to_fid/:fname",
	"/config/page/login",
}
var PreCheckWhiteListWithFE = []string{
	"/api/passport/login",
	"/api/passport/signup",
	"/api/passport/reset/password",
	// "/api/tools/userinfo/tieba_uid/:tiebauid",
	// "/api/tools/userinfo/panel/:query_type/:user_value",
	// "/api/tools/tieba/fname_to_fid/:fname",
	"/api/config/page/login",
}

var RoleList = []string{"deleted", "banned", "user", "vip", "admin"}

func PreCheckWhiteListExists(path string) bool {
	if share.EnableFrontend {
		return slices.Contains(PreCheckWhiteListWithFE, path)
	}

	return slices.Contains(PreCheckWhiteListWithoutFE, path)
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

func verifyAuthorization(authorization string) (string, string) {
	if authorization == "" {
		return "0", "guest"
	}

	token := strings.Split(strings.TrimSpace(authorization), ":")
	// TODO static target
	if len(token) != 3 {
		return "0", "guest"
	}

	uid, err := strconv.ParseInt(token[0], 10, 64)
	if err != nil || uid <= 0 {
		return "0", "guest"
	}

	strUID := strconv.Itoa(int(uid))

	expiredAt, _ := strconv.ParseInt(token[2], 10, 64)
	expiredAtTime := time.Unix(expiredAt, 0)

	if time.Now().After(expiredAtTime) {
		return "0", "guest"
	}

	savedSessionExpiredAt := GetSessionExpiredAt(strUID)

	if savedSessionExpiredAt <= 0 {
		return "0", "guest"
	}

	savedSessionExpiredAtTime := time.Unix(savedSessionExpiredAt, 0)
	if savedSessionExpiredAtTime.After(expiredAtTime) {
		return "0", "guest"
	}

	dbPwd := _function.GetPassword(int(uid))

	byteToken, _ := base64.RawURLEncoding.DecodeString(token[1])

	if dbPwd != "" {
		if !hmac.Equal(HmacSessionToken(strUID, dbPwd, strconv.Itoa(int(savedSessionExpiredAt))), byteToken) {
			return "0", "guest"
		}
		var accountInfo []*model.TcUser
		_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)
		if len(accountInfo) == 1 {
			return strconv.Itoa(int(accountInfo[0].ID)), accountInfo[0].Role
		}

		return "0", "guest"
	}

	return "0", "guest"
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
	return _function.GenHMAC256([]byte(password), []byte(uid+":"+password+":"+expiredAt))
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
	strExpiredAt := strconv.Itoa(int(expiredAt))

	token := base64.RawURLEncoding.EncodeToString(HmacSessionToken(strconv.Itoa(uid), password, strExpiredAt))

	// HttpAuthRefreshTokenMap.Store(int(uid), token, numberCookieExpire)

	return strconv.Itoa(uid) + ":" + token + ":" + strconv.Itoa(int(expiredAt)), expiredAt, numberCookieExpire
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
	if err := _function.SetUserOption("session_expired_at", strconv.Itoa(int(t)), uid); err != nil {
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
			Expire: _function.Now.Add(time.Second * time.Duration(_function.ResetPwdExpire)).Unix(),
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
		log.Println("send-reset-message", err)
		return "", errors.New("消息发送失败")
	}

	return v.VerifyCode, nil
}

func IsArrayMode(c echo.Context) bool {
	arrayModeValue := c.QueryParam("array_mode")
	return arrayModeValue != "" && arrayModeValue != "0" && arrayModeValue != "false"
}
