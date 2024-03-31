package _api

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

type addAccountsResponseList struct {
	Name     string `json:"name,omitempty"`
	NameShow string `json:"name_show,omitempty"`
	Portrait string `json:"portrait"`
	Fname    string `json:"fname,omitempty"`
	Start    int64  `json:"start,omitempty"`
	End      int64  `json:"end,omitempty"`
	Success  bool   `json:"success,omitempty"`
	Msg      string `json:"msg,omitempty"`
}

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

func PluginLoopBanGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var loopBanList []model.TcVer4BanList
	_function.GormDB.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Find(&loopBanList)

	limit := _function.GetOption("ver4_ban_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)

	var responseList []addAccountsResponseList

	for _, v := range loopBanList {
		responseList = append(responseList, addAccountsResponseList{
			Name:     v.Name,
			NameShow: v.NameShow,
			Portrait: v.Portrait,
			Fname:    v.Tieba,
			Start:    int64(v.Stime),
			End:      int64(v.Etime),
			Success:  true,
		})
	}

	var list = struct {
		Count int64                     `json:"count"`
		Limit int64                     `json:"limit"`
		List  []addAccountsResponseList `json:"list"`
	}{
		Count: int64(len(responseList)),
		Limit: numLimit,
		List:  responseList,
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", list, "tbsign"))
}

func PluginLoopBanAddAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)
	pid := c.FormValue("pid")
	start := c.FormValue("start")
	end := c.FormValue("end")
	fname := c.FormValue("fname")
	portraits := c.FormValue("portrait")

	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	// time
	startTime := time.Now()
	if start != "" {
		startTime, err = time.Parse("2006-01-02", start)
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(403, "Invalid start date format", echoEmptyObject, "tbsign"))
		}
	}

	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid end date format", echoEmptyObject, "tbsign"))
	}

	if startTime.Unix() >= endTime.Unix() {
		return c.JSON(http.StatusOK, apiTemplate(403, "Start time is greater than end time", echoEmptyObject, "tbsign"))
	}

	if endTime.Unix() < time.Now().Unix() {
		return c.JSON(http.StatusOK, apiTemplate(403, "Now time is greater than end time", echoEmptyObject, "tbsign"))
	}

	// pre check
	var accountInfo model.TcBaiduid
	_function.GormDB.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).First(&accountInfo)
	if accountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	// portrait
	if portraits == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Value of portrait is empty!", echoEmptyObject, "tbsign"))
	}
	portraitList := []string{}
	for _, portrait := range strings.Split(portraits, "\n") {
		if strings.HasPrefix(portrait, "tb.1.") {
			portraitList = append(portraitList, portrait)
		}
	}

	// limit
	limit := _function.GetOption("ver4_ban_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)

	var existsAccountList []model.TcVer4BanList
	_function.GormDB.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Find(&existsAccountList)

	count := int64(len(existsAccountList))
	if count >= numLimit || count+int64(len(portraitList)) > numLimit {
		return c.JSON(http.StatusOK, apiTemplate(403, "Account limit exceeded", echoEmptyObject, "tbsign"))
	}

	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
	}

	// is manager?
	if _function.GetOption("ver4_ban_break_check") == "0" {
		managerStatus, err := _plugin.GetManagerStatus(_function.GetCookie(int32(numPid)).Portrait, fid)
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(500, "Unable to check manager status", echoEmptyObject, "tbsign"))
		}
		if !managerStatus.IsManager {
			return c.JSON(http.StatusOK, apiTemplate(403, "You are NOT the manager of fid:"+fname, echoEmptyObject, "tbsign"))
		}
	}

	// get account info
	var accountsResult []addAccountsResponseList
	var accountsToInsert []model.TcVer4BanList
	for _, portrait := range portraitList {
		// check db
		dbExists := false
		for _, v := range existsAccountList {
			if v.Portrait == portrait {
				accountsResult = append(accountsResult, addAccountsResponseList{
					Name:     v.Name,
					NameShow: v.NameShow,
					Portrait: portrait,
					Fname:    v.Tieba,
					Start:    int64(v.Stime),
					End:      int64(v.Etime),
					Success:  false,
					Msg:      "Account is already exists",
				})
				dbExists = true
				break
			}
		}
		if dbExists {
			continue
		}

		// check exists
		banUserInfo, err := _function.GetUserInfoByUsernameOrPortrait("portrait", portrait)
		if err != nil || banUserInfo.No != 0 {
			accountsResult = append(accountsResult, addAccountsResponseList{
				Portrait: portrait,
				Success:  false,
				Msg:      "Portrait not found",
			})
		}
		accountsResult = append(accountsResult, addAccountsResponseList{
			Name:     banUserInfo.Data.Name,
			NameShow: banUserInfo.Data.NameShow,
			Portrait: portrait,
			Fname:    fname,
			Start:    startTime.Unix(),
			End:      endTime.Unix(),
			Success:  true,
			Msg:      "OK",
		})
		accountsToInsert = append(accountsToInsert, model.TcVer4BanList{
			UID:      int32(numUID),
			Pid:      int32(numPid),
			Name:     banUserInfo.Data.Name,
			NameShow: banUserInfo.Data.NameShow,
			Portrait: portrait,
			Tieba:    fname,
			Stime:    int32(startTime.Unix()),
			Etime:    int32(endTime.Unix()),
			Date:     0,
		})
	}

	_function.GormDB.Create(&accountsToInsert)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", accountsResult, "tbsign"))

}

func PluginLoopBanDelAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	id := c.Param("id")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, apiTemplate(500, "Invalid id", map[string]any{
			"success": false,
			"id":      id,
		}, "tbsign"))
	}

	_function.GormDB.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numUID),
		ID:  int32(numID),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginLoopBanDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numUID),
	})

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
}

func PluginLoopBanPreCheckIsManager(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fname := c.Param("fname")

	// pre-check pid
	var pidCheck []model.TcBaiduid
	_function.GormDB.Where("uid = ? AND id = ?", uid, pid).Limit(1).Find(&pidCheck)

	if len(pidCheck) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
	}

	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
	}
	resp, err := _plugin.GetManagerStatus(pidCheck[0].Portrait, fid)
	if err != nil {
		log.Println(err)
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}
