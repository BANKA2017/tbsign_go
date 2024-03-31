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

	pid := c.FormValue("pid")
	fname := c.FormValue("fname")

	if fname == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid fname", echoEmptyObject, "tbsign"))
	}
	// get tieba info by fname
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "Tieba \""+fname+"\" is not exists", echoEmptyObject, "tbsign"))
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaItems []model.TcTieba
	_function.GormDB.Where("uid = ? AND pid = ? AND fid = ?", uid, pid, fid).Limit(1).Find(&tiebaItems)

	if len(tiebaItems) > 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "Tieba is already exists", tiebaItems[0], "tbsign"))
	}

	newTieba := model.TcTieba{
		UID:       int32(numUID),
		Pid:       int32(numPid),
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

	pid := c.Param("pid")
	fid := c.Param("fid")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}
	numFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid fid", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.Delete(&model.TcTieba{
		UID: int32(numUID),
		Pid: int32(numPid),
		Fid: int32(numFid),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", echoEmptyObject, "tbsign"))
}

func IgnoreTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	method := c.Request().Method

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}
	numFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid fid", echoEmptyObject, "tbsign"))
	}

	if method == "DELETE" {
		_function.GormDB.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, numFid).Update("no", false)
	} else {
		_function.GormDB.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, numFid).Update("no", true)
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"uid": numUID,
		"pid": numPid,
		"fid": numFid,
		"no":  method != "DELETE",
	}, "tbsign"))
}

func CleanTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	_function.GormDB.Where("uid = ?", uid).Delete(&model.TcTieba{})
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]string{
		"uid": uid,
	}, "tbsign"))
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
