package _api

import (
	"log"
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
)

// ver4_lottery
func PluginKnowsLotteryGetLogs(c echo.Context) error {
	uid := c.Get("uid").(string)

	log := new([]model.TcVer4LotteryLog)
	_function.GormDB.R.Model(&model.TcVer4LotteryLog{}).Where("uid = ?", uid).Order("id DESC").Find(log)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", log, "tbsign"))
}

func PluginKnowsLotteryGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_lottery_check", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("ver4_lottery_check", status, uid)
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginKnowsLotterySwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_lottery_check", uid) != "0"

	err := _function.SetUserOption("ver4_lottery_check", !status, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "无法修改知道商城抽奖插件状态", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", !status, "tbsign"))
}
