package _api

import (
	"net/http"
	"strconv"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
)

func AddTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.FormValue("pid")
	fname := c.FormValue("fname")

	if fname == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "贴吧名无效", echoEmptyObject, "tbsign"))
	}
	// get tieba info by fname
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "\""+fname+"吧\" 不存在", echoEmptyObject, "tbsign"))
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaItems []model.TcTieba
	_function.GormDB.R.Where("uid = ? AND pid = ? AND fid = ?", uid, pid, fid).Limit(1).Find(&tiebaItems)

	if len(tiebaItems) > 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "贴吧已存在", tiebaItems[0], "tbsign"))
	}

	// TOO STUPID!
	newTieba := _type.TcTieba{
		TcTieba: model.TcTieba{
			UID:    int32(numUID),
			Pid:    int32(numPid),
			Fid:    int32(fid),
			No:     false,
			Latest: 0,
		},
		Tieba:     _function.VariablePtrWrapper(fname),
		Status:    _function.VariablePtrWrapper(int32(0)),
		LastError: _function.VariablePtrWrapper(""),
	}

	_function.GormDB.W.Create(&newTieba)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", newTieba, "tbsign"))
}

func RemoveTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}
	numFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 fid", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, numFid).Delete(&model.TcTieba{})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", echoEmptyObject, "tbsign"))
}

func ResetTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}
	numFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 fid", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, numFid).Update("latest", 0)

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
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}
	numFid, err := strconv.ParseInt(fid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 fid", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, numFid).Update("no", method != "DELETE")

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"uid": numUID,
		"pid": numPid,
		"fid": numFid,
		"no":  method != "DELETE",
	}, "tbsign"))
}

func CleanTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcTieba{})
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]string{
		"uid": uid,
	}, "tbsign"))
}

func RefreshTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	// get account list
	for _, v := range tiebaAccounts {
		_plugin.ScanTiebaByPid(v.ID)
	}

	var tiebaList []model.TcTieba
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaList, "tbsign"))
}

func GetTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaList []model.TcTieba
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaList, "tbsign"))
}
