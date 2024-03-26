package _api

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
)

var echoEmptyObject = make(map[string]interface{}, 0)

var PreCheckWhiteList = []string{
	"/*",
	"/favicon.ico",
	"/robots.txt",
	"/passport/login",
	"/passport/logout",
	"/passport/register",
}

func PreCheckWhiteListExists(path string) bool {
	for _, v := range PreCheckWhiteList {
		if path == v {
			return true
		}
	}
	return false
}

func apiTemplate[T any](code int, message string, data T, version string) _type.ApiTemplate {
	return _type.ApiTemplate{
		Code:    code,
		Message: message,
		Data:    data,
		Version: version,
	}
}

func echoReject(c echo.Context) error {
	var response = apiTemplate(403, "Invalid Request", echoEmptyObject, "tbsign")
	return c.JSON(http.StatusForbidden, response)
}

func echoNoContent(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

func verifyAuthorization(authorization string) (string, string) {
	if authorization == "" {
		return "0", "guest"
	} else {
		parsedAuth, err := base64.RawURLEncoding.DecodeString(authorization)
		if err != nil {
			return "0", "guest"
		}
		authArray := strings.Split(string(parsedAuth), ":")
		if len(authArray) != 3 || authArray[0] == "" || authArray[1] == "" || authArray[2] == "" {
			return "0", "guest"
		}
		var accountInfo []model.TcUser
		_function.GormDB.Where("id = ?", authArray[0]).Limit(1).Find(&accountInfo)

		if hex.EncodeToString(_function.GenHMAC256([]byte(accountInfo[0].Pw+":"+authArray[1]), []byte(strconv.Itoa(int(accountInfo[0].ID))+accountInfo[0].Pw))) == authArray[2] {
			return strconv.Itoa(int(accountInfo[0].ID)), accountInfo[0].Role
		} else {
			return "0", "guest"
		}
	}
}
