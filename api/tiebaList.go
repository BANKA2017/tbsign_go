package _api

import (
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

func AddTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Request().PostFormValue("pid")
	fname := c.Request().PostFormValue("fname")

	if fname == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid fname", make(map[string]interface{}, 0), "tbsign"))
	}
	// get tieba info by fname
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "Tieba \""+fname+"\" is not exists", make(map[string]interface{}, 0), "tbsign"))
	}

	numberUID, _ := strconv.ParseInt(uid, 10, 64)
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", make(map[string]interface{}, 0), "tbsign"))
	}
	newTieba := model.TcTieba{
		UID:       int32(numberUID),
		Pid:       int32(numberPid),
		Fid:       int32(fid),
		Tieba:     fname,
		No:        false,
		Status:    0,
		Latest:    0,
		LastError: "NULL",
	}

	_function.GormDB.Create(&newTieba)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", newTieba, "tbsign"))
}

func RemoveTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Request().PostFormValue("pid")
	fid := c.Request().PostFormValue("fid")

	numberUID, _ := strconv.ParseInt(uid, 10, 64)
	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", make(map[string]interface{}, 0), "tbsign"))
	}
	numberFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid fid", make(map[string]interface{}, 0), "tbsign"))
	}

	_function.GormDB.Delete(&model.TcTieba{
		UID: int32(numberUID),
		Pid: int32(numberPid),
		Fid: int32(numberFid),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", "tiebaAccounts", "tbsign"))
}

func RefreshTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)

	// get account list
	for _, v := range tiebaAccounts {
		_plugin.ScanTiebaByPid(v.ID)
	}

	var tiebaList []model.TcTieba
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaList)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaList, "tbsign"))
}

func GetTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaList []model.TcTieba
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaList)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaList, "tbsign"))
}
