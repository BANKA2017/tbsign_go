package _api

import (
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
)

// TODO QR login
// func LoginTiebaAccount(c echo.Context) error {
// 	uid := c.Get("uid").(string)
//
// 	return c.JSON(http.StatusOK, apiTemplate(200, "OK", "", "tbsign"))
//
// }

func AddTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	bduss := c.Request().PostFormValue("bduss")
	stoken := c.Request().PostFormValue("stoken")

	if bduss == "" || stoken == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid bduss or stoken", make(map[string]interface{}, 0), "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Unable to verify BDUSS", make(map[string]interface{}, 0), "tbsign"))
	}

	numberUID, _ := strconv.ParseInt(uid, 10, 64)

	newAccount := model.TcBaiduid{
		UID:      int32(numberUID),
		Bduss:    bduss,
		Stoken:   stoken,
		Name:     baiduAccountInfo.Name,
		Portrait: baiduAccountInfo.Portrait,
	}
	_function.GormDB.Create(&newAccount)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", newAccount, "tbsign"))

}

func RemoveTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Request().PostFormValue("pid")
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", make(map[string]interface{}, 0), "tbsign"))
	}

	// fid pid
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)

	for _, v := range tiebaAccounts {
		if v.ID == int32(numberPid) {
			_function.GormDB.Delete(&v)
			return c.JSON(http.StatusOK, apiTemplate(200, "OK", v, "tbsign"))
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(404, "pid not found", make(map[string]interface{}, 0), "tbsign"))
}

func GetTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaAccounts, "tbsign"))
}

func CheckTiebaAccount(c echo.Context) error {
	pid := c.QueryParams().Get("pid")
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", make(map[string]interface{}, 0), "tbsign"))
	}

	var tiebaAccount model.TcBaiduid
	_function.GormDB.Where("id = ?", pid).Find(&tiebaAccount)

	if tiebaAccount.ID != 0 && tiebaAccount.ID != int32(numberPid) {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", make(map[string]interface{}, 0), "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: tiebaAccount.Bduss})

	if err != nil || baiduAccountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}

}
