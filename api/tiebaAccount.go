package _api

import (
	"net/http"
	"strconv"
	"strings"

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

	bduss := strings.TrimSpace(c.FormValue("bduss"))
	stoken := strings.TrimSpace(c.FormValue("stoken"))

	if bduss == "" || stoken == "" {
		return c.JSON(http.StatusOK, apiTemplate(401, "Invalid BDUSS or stoken", echoEmptyObject, "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "Unable to verify BDUSS", echoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ? AND portrait = ?", uid, baiduAccountInfo.User.Portrait).Limit(1).Find(&tiebaAccounts)

	if len(tiebaAccounts) > 0 {
		if tiebaAccounts[0].Bduss != bduss || tiebaAccounts[0].Stoken != stoken {
			newData := model.TcBaiduid{
				Bduss:    bduss,
				Stoken:   stoken,
				Name:     baiduAccountInfo.User.Name,
				Portrait: baiduAccountInfo.User.Portrait,
			}
			_function.GormDB.Model(model.TcBaiduid{}).Where("id = ?", tiebaAccounts[0].ID).Updates(&newData)
			newData.ID = tiebaAccounts[0].ID
			newData.UID = tiebaAccounts[0].UID
			return c.JSON(http.StatusOK, apiTemplate(200, "OK", newData, "tbsign"))
		} else if tiebaAccounts[0].Bduss == bduss && tiebaAccounts[0].Stoken == stoken {
			return c.JSON(http.StatusOK, apiTemplate(200, "Account already exists", tiebaAccounts[0], "tbsign"))
		}
	}

	numberUID, _ := strconv.ParseInt(uid, 10, 64)

	newAccount := model.TcBaiduid{
		UID:      int32(numberUID),
		Bduss:    bduss,
		Stoken:   stoken,
		Name:     baiduAccountInfo.User.Name,
		Portrait: baiduAccountInfo.User.Portrait,
	}
	_function.GormDB.Create(&newAccount)
	return c.JSON(http.StatusOK, apiTemplate(201, "OK", newAccount, "tbsign"))

}

func RemoveTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	// fid pid
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)

	for _, v := range tiebaAccounts {
		if v.ID == int32(numberPid) {
			_function.GormDB.Model(&model.TcBaiduid{}).Delete("id = ?", v.ID)
			_function.GormDB.Model(&model.TcTieba{}).Delete("pid = ?", v.ID)
			return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]int32{
				"pid": v.ID,
			}, "tbsign"))
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(404, "pid not found", echoEmptyObject, "tbsign"))
}

func GetTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaAccounts, "tbsign"))
}

func CheckTiebaAccount(c echo.Context) error {
	pid := c.Param("pid")
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	var tiebaAccount model.TcBaiduid
	_function.GormDB.Where("id = ?", pid).Find(&tiebaAccount)

	if tiebaAccount.ID != 0 && tiebaAccount.ID != int32(numberPid) {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: tiebaAccount.Bduss})

	//log.Println(tiebaAccount, baiduAccountInfo)

	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}

}
