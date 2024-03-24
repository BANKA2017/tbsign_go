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

var PreCheckWhiteList = []string{
	"/*",
	"/favicon.ico",
	"/robots.txt",
	"/passport/login",
	"/passport/logout",
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
	var template _type.ApiTemplate
	template.Code = code
	template.Message = message
	template.Data = data
	template.Version = version
	return template
}

func echoReject(c echo.Context) error {
	var response = apiTemplate(403, "Invalid Request", make(map[string]interface{}, 0), "tbsign")
	return c.JSON(http.StatusForbidden, response)
}

func echoFavicon(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}

func echoRobots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nDisallow: /*")
}

func verifyAuthorization(authorization string) string {
	if authorization == "" {
		return "0"
	} else {
		parsedAuth, err := base64.RawURLEncoding.DecodeString(authorization)
		if err != nil {
			return "0"
		}
		authArray := strings.Split(string(parsedAuth), ":")
		if len(authArray) != 3 || authArray[0] == "" || authArray[1] == "" || authArray[2] == "" {
			return "0"
		}
		var accountInfo []model.TcUser
		_function.GormDB.Where("id = ?", authArray[0]).Limit(1).Find(&accountInfo)

		if hex.EncodeToString(_function.GenHMAC256([]byte(accountInfo[0].Pw+":"+authArray[1]), []byte(strconv.Itoa(int(accountInfo[0].ID))+accountInfo[0].Pw))) == authArray[2] {
			return authArray[0]
		} else {
			return "0"
		}
	}
}
