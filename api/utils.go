package _api

import (
	"encoding/hex"
	"errors"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
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
	} else {
		return slices.Contains(PreCheckWhiteListWithoutFE, path)
	}
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

func verifyAuthorization(authorization string) (string, string) {
	if authorization == "" {
		return "0", "guest"
	} else {
		token := strings.Split(strings.TrimSpace(authorization), ":")
		// TODO static target
		if len(token) != 2 {
			return "0", "guest"
		}

		uid, err := strconv.ParseInt(token[0], 10, 64)
		if err != nil || uid <= 0 {
			return "0", "guest"
		}

		if storeToken, ok := HttpAuthRefreshTokenMap.Load(int(uid)); ok {
			tokenContent, ok := storeToken.(*HttpAuthRefreshTokenMapItemStruct)
			if !ok || tokenContent.ExpireAt <= time.Now().Unix() || tokenContent.Content != token[1] {
				return "0", "guest"
			}
			var accountInfo []*model.TcUser
			_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)
			if len(accountInfo) == 1 {
				return strconv.Itoa(int(accountInfo[0].ID)), accountInfo[0].Role
			} else {
				return "0", "guest"
			}
		} else {
			return "0", "guest"
		}
	}
}

func legacyTokenBuilder(uid int32, password string) string {
	return _function.Base64URLEncode([]byte(strconv.Itoa(int(uid)) + ":" + hex.EncodeToString(_function.GenHMAC256([]byte(password), []byte(strconv.Itoa(int(uid))+password)))))
}

type HttpAuthRefreshTokenMapItemStruct struct {
	Content  string
	ExpireAt int64
}

var HttpAuthRefreshTokenMap sync.Map // int -> HttpAuthRefreshTokenMapItemStruct

func tokenBuilder(uid int) (string, int64) {
	_token, err := _function.RandomTokenBuilder(48)
	if err != nil {
		return "", 0
	}
	token := _function.Base64URLEncode(_token)

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

	expireAt := time.Now().Add(time.Duration(numberCookieExpire) * time.Second).Unix()

	HttpAuthRefreshTokenMap.Store(int(uid), &HttpAuthRefreshTokenMapItemStruct{
		Content:  token,
		ExpireAt: expireAt,
	})

	return _function.AppendStrings(strconv.Itoa(uid), ":", token), expireAt
}

var resetPasswordVerifyCodeByteLength int64 = 6

var resetPasswordVerifyCodeLength = int(math.Ceil(float64(resetPasswordVerifyCodeByteLength*4) / float64(3)))

func ResetMessageBuilder(uid int32, forceMode bool) *_function.VerifyCodeStruct {
	_v, ok := _function.VerifyCodeList.LoadCode("reset_password", uid)
	var v *_function.VerifyCodeStruct

	if !ok || _v == nil || _v.Expire <= _function.Now.Unix() {
		v = _function.VariablePtrWrapper(_function.VerifyCodeStruct{
			Expire: _function.Now.Unix() + _function.ResetPwdExpire,
		})
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
	} else {
		return v.VerifyCode, nil
	}
}
