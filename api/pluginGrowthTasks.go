package _api

import (
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

func PluginGrowthTasksGetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	// sign only
	signOnly := _function.GetUserOption("kd_growth_sign_only", uid)
	if signOnly == "" {
		signOnly = "0"
		_function.SetUserOption("kd_growth_sign_only", signOnly, uid)
	}

	// no icon tasks
	noIconTasks := _function.GetUserOption("kd_growth_break_icon_tasks", uid)
	if noIconTasks == "" {
		noIconTasks = "0"
		_function.SetUserOption("kd_growth_break_icon_tasks", noIconTasks, uid)
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"sign_only":        signOnly,
		"break_icon_tasks": noIconTasks,
	}, "tbsign"))
}

func PluginGrowthTasksSetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	signOnly := c.FormValue("sign_only") == "0"
	noIconTasks := c.FormValue("break_icon_tasks") != "0"

	_function.SetUserOption("kd_growth_sign_only", signOnly, uid)
	_function.SetUserOption("kd_growth_break_icon_tasks", noIconTasks, uid)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"success": true,
	}, "tbsign"))
}

func PluginGrowthTasksGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accounts []model.TcKdGrowth
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&accounts)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", accounts, "tbsign"))
}

func PluginGrowthTasksAddAccount(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("uid = ? AND pid = ?", uid, numPid).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "帐号已存在", echoEmptyObject, "tbsign"))
	} else {
		dataToInsert := model.TcKdGrowth{
			UID:  numUID,
			Pid:  numPid,
			Date: 0,
		}
		_function.GormDB.W.Create(&dataToInsert)
		_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("uid = ? AND pid = ?", uid, numPid).First(&dataToInsert)
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", dataToInsert, "tbsign"))
	}
}

func PluginGrowthTasksDelAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	id := c.Param("id")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "无效任务 id", map[string]any{
			"success": false,
			"id":      id,
		}, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcKdGrowth{}).Delete(&model.TcKdGrowth{
		UID: numUID,
		ID:  numID,
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginGrowthTasksDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.W.Model(&model.TcKdGrowth{}).Delete(&model.TcKdGrowth{
		UID: numUID,
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
}

func PluginGrowthTasksGetTasksStatus(c echo.Context) error {
	uid := c.Get("uid").(string)
	pid := c.Param("pid")

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).Count(&count)

	if count > 0 {
		numPid, _ := strconv.ParseInt(pid, 10, 64)
		status, err := _plugin.GetUserGrowthTasksList(_function.GetCookie(int32(numPid)))
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(500, "获取任务列表失败", echoEmptyObject, "tbsign"))
		} else if status.No != 0 {
			return c.JSON(http.StatusOK, apiTemplate(500, status.Error, echoEmptyObject, "tbsign"))
		}
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", status.Data, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(404, "帐号不存在", echoEmptyObject, "tbsign"))
	}
}
