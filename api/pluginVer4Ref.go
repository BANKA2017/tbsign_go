package _api

import (
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

func GetVer4RefList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)

	var tiebaList []model.TcTieba
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaList)

	type ver4RefListResponse struct {
		PID      int32  `json:"pid"`
		Name     string `json:"name"`
		Portrait string `json:"portrait"`
		Count    int32  `json:"count"`
	}

	var response []ver4RefListResponse
	for _, v := range tiebaAccounts {
		var count int32
		for _, v1 := range tiebaList {
			if v1.Pid == v.ID {
				count++
			}
		}
		response = append(response, ver4RefListResponse{
			PID:      v.ID,
			Name:     v.Name,
			Portrait: v.Portrait,
			Count:    count,
		})
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", response, "tbsign"))

}

func RefreshVer4RefTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.FormValue("pid")

	numberPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || numberPid <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.Where("uid = ?", uid).Find(&tiebaAccounts)

	// get account list
	for _, v := range tiebaAccounts {
		if v.ID == int32(numberPid) {
			_plugin.ScanTiebaByPid(v.ID)
			var tiebaList []model.TcTieba
			_function.GormDB.Where("uid = ? AND pid = ?", uid, pid).Find(&tiebaList)
			return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaList, "tbsign"))
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(404, "Pid not found", echoEmptyObject, "tbsign"))
}
