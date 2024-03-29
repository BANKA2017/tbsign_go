package _api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

func PluginLoopBanSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_ban_open", uid) != "0"

	newValue := "1"
	if status {
		newValue = "0"
	}
	err := _function.SetUserOption("ver4_ban_open", newValue, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "Unable to update status", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", !status, "tbsign"))
}

func PluginLoopBanGetReason(c echo.Context) error {
	uid := c.Get("uid").(string)

	var loopBanSettings []model.TcVer4BanUserset
	_function.GormDB.Where("uid = ?", uid).Limit(1).Find(&loopBanSettings)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]string{
		"reason": loopBanSettings[0].C,
	}, "tbsign"))
}

func PluginLoopBanSetReason(c echo.Context) error {
	uid := c.Get("uid").(string)

	reason := c.FormValue("reason")

	err := _function.GormDB.Model(&model.TcVer4BanUserset{}).Where("uid = ?", uid).Update("c", reason).Error
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "Unable to update the ban reason", map[string]string{
			"reason": reason,
		}, "tbsign"))
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]string{
		"reason": reason,
	}, "tbsign"))

}

func PluginLoopBanGetList() {}

func PluginLoopBanAddAccounts() {}

func PluginLoopBanDelAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	id := c.Param("id")

	numberUID, _ := strconv.ParseInt(uid, 10, 64)
	numberID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, apiTemplate(500, "Invalid id", map[string]any{
			"success": false,
			"id":      id,
		}, "tbsign"))
	}

	_function.GormDB.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numberUID),
		ID:  int32(numberID),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginLoopBanDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numberUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numberUID),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
}

func PluginLoopBanPreCheckIsManager(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	tiebaName := c.Param("tieba_name")

	// pre-check pid
	var pidCheck []model.TcBaiduid
	_function.GormDB.Where("uid = ? AND id = ?", uid, pid).Limit(1).Find(&pidCheck)

	if len(pidCheck) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
	}

	resp, err := _plugin.GetManagerStatus(pidCheck[0].Portrait, tiebaName)
	if err != nil {
		log.Println(err)
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}
