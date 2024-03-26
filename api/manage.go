package _api

import (
	"net/http"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

func GetAdminSettings(c echo.Context) error {
	var adminSettings []model.TcOption
	_function.GormDB.Find(&adminSettings)

	settings := make(map[string]string)
	for _, v := range adminSettings {
		settings[v.Name] = v.Value
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func UpdateAdminSettings(c echo.Context) error {
	c.Request().ParseForm()

	var adminSettings []model.TcOption
	_function.GormDB.Find(&adminSettings)

	var newSettings []model.TcOption

	for _, v := range adminSettings {
		for k1, v1 := range c.Request().Form {
			if v.Name == k1 {
				if v.Value != v1[0] {
					v.Value = v1[0]
					newSettings = append(newSettings, v)
				}
				break
			}
		}
	}

	settings := make(map[string]string)
	if len(newSettings) > 0 {
		for _, v := range newSettings {
			settings[v.Name] = v.Value
			_function.GormDB.Model(model.TcOption{}).Where("name = ?", v.Name).Updates(&model.TcOption{
				Value: v.Value,
			})
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

// TODO
// func AddAccount()  {
// }
//
// func RemoveAccount()  {
// }
//
// func UpdateAccount()  {
// }

func GetAccountList(c echo.Context) error {
	var accountInfo []model.TcUser
	_function.GormDB.Find(&accountInfo)

	type respStruct struct {
		ID    int32  `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
		T     string `json:"t"`
	}

	var respAccountInfo []respStruct

	for _, v := range accountInfo {
		respAccountInfo = append(respAccountInfo, respStruct{
			ID:    v.ID,
			Name:  v.Name,
			Email: v.Email,
			Role:  v.Role,
			T:     v.T,
		})
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", respAccountInfo, "tbsign"))
}
