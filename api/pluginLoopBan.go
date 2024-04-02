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
		return c.JSON(http.StatusOK, apiTemplate(500, "无法启用循环封禁功能", status, "tbsign"))
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
		return c.JSON(http.StatusOK, apiTemplate(500, "无法更新封禁理由", map[string]string{
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
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	// time
	startTime := time.Now()
	if start != "" {
		startTime, err = time.Parse("2006-01-02", start)
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(403, "开始日期格式错误", echoEmptyObject, "tbsign"))
		}
	}

	endTime, err := time.Parse("2006-01-02", end)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "结束日期格式错误", echoEmptyObject, "tbsign"))
	}

	if startTime.Unix() >= endTime.Unix() {
		return c.JSON(http.StatusOK, apiTemplate(403, "开始时刻晚于结束时刻", echoEmptyObject, "tbsign"))
	}

	if endTime.Unix() < time.Now().Unix() {
		return c.JSON(http.StatusOK, apiTemplate(403, "现在时刻晚于结束时刻", echoEmptyObject, "tbsign"))
	}

	// pre check
	var accountInfo model.TcBaiduid
	_function.GormDB.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).First(&accountInfo)
	if accountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "无效 pid", echoEmptyObject, "tbsign"))
	}

	// portrait
	if portraits == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "待封禁 portrait 列表为空!", echoEmptyObject, "tbsign"))
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
		return c.JSON(http.StatusOK, apiTemplate(403, "添加帐号数超限（"+strconv.Itoa(int(count+int64(len(portraitList))))+"/"+limit+"）", echoEmptyObject, "tbsign"))
	}

	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
	}

	// is manager?
	if _function.GetOption("ver4_ban_break_check") == "0" {
		managerStatus, err := _plugin.GetManagerStatus(_function.GetCookie(int32(numPid)).Portrait, fid)
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(500, "无法获取吧务列表", echoEmptyObject, "tbsign"))
		}
		if !managerStatus.IsManager {
			return c.JSON(http.StatusOK, apiTemplate(403, "您不是 fid:"+fname+" 的吧务成员", echoEmptyObject, "tbsign"))
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
					Msg:      "账号已存在",
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
				Msg:      "帐号不存在",
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
		return c.JSON(http.StatusOK, apiTemplate(500, "无效 id", map[string]any{
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
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", _plugin.IsManagerPreCheckResponse{}, "tbsign"))
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
