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

func GetLoginQRCode(c echo.Context) error {
	qrcode, err := _function.GetLoginQRCode()
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "获取二维码失败", echoEmptyObject, "tbsign"))
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", qrcode, "tbsign"))
}

func GetBDUSS(c echo.Context) error {
	uid := c.Get("uid").(string)
	sign := c.FormValue("sign")

	res, err := _function.GetUnicastResponse(sign)

	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "获取状态失败", echoEmptyObject, "tbsign"))
	}

	// tmpBDUSS
	if res.ChannelV.V == "" {
		return c.JSON(http.StatusOK, apiTemplate(400, "未确认", echoEmptyObject, "tbsign"))
	}

	res2, err := _function.GetLoginResponse(res.ChannelV.V)

	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "登录失败", echoEmptyObject, "tbsign"))
	}

	stokenStr := strings.ReplaceAll(res2.Data.Session.StokenList, "&quot;", "\"")
	var stokenArray []string
	err = _function.JsonDecode([]byte(stokenStr), &stokenArray)
	if err != nil || res2.Data.Session.Bduss == "" {
		return c.JSON(http.StatusOK, apiTemplate(500, "登录失败", echoEmptyObject, "tbsign"))
	}
	bduss := res2.Data.Session.Bduss

	stokenKV := make(map[string]string)
	for _, v := range stokenArray {
		tmpSplit := strings.Split(v, "#")
		stokenKV[tmpSplit[0]] = tmpSplit[1]
	}

	if stoken, ok := stokenKV["tb"]; !ok || stoken == "" {
		return c.JSON(http.StatusOK, apiTemplate(500, "登录失败", echoEmptyObject, "tbsign"))
	}
	stoken := stokenKV["tb"]

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "无法验证登录状态 BDUSS", echoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.R.Where("uid = ? AND portrait = ?", uid, baiduAccountInfo.User.Portrait).Limit(1).Find(&tiebaAccounts)

	if len(tiebaAccounts) > 0 {
		if tiebaAccounts[0].Bduss != bduss || tiebaAccounts[0].Stoken != stoken {
			newData := model.TcBaiduid{
				Bduss:    bduss,
				Stoken:   stoken,
				Name:     baiduAccountInfo.User.Name,
				Portrait: baiduAccountInfo.User.Portrait,
			}
			_function.GormDB.W.Model(model.TcBaiduid{}).Where("id = ?", tiebaAccounts[0].ID).Updates(&newData)
			newData.ID = tiebaAccounts[0].ID
			newData.UID = tiebaAccounts[0].UID
			newData.Bduss = ""
			newData.Stoken = ""
			return c.JSON(http.StatusOK, apiTemplate(200, "已更新 BDUSS", newData, "tbsign"))
		} else if tiebaAccounts[0].Bduss == bduss && tiebaAccounts[0].Stoken == stoken {
			tiebaAccounts[0].Bduss = ""
			tiebaAccounts[0].Stoken = ""
			return c.JSON(http.StatusOK, apiTemplate(200, "贴吧账号已存在", tiebaAccounts[0], "tbsign"))
		}
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	newAccount := model.TcBaiduid{
		UID:      int32(numUID),
		Bduss:    bduss,
		Stoken:   stoken,
		Name:     baiduAccountInfo.User.Name,
		Portrait: baiduAccountInfo.User.Portrait,
	}
	_function.GormDB.W.Create(&newAccount)
	newAccount.Bduss = ""
	newAccount.Stoken = ""
	return c.JSON(http.StatusOK, apiTemplate(201, "OK", newAccount, "tbsign"))
}

func AddTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	bduss := strings.TrimSpace(c.FormValue("bduss"))
	stoken := strings.TrimSpace(c.FormValue("stoken"))

	includeBDUSSAndStoken := c.QueryParams().Get("all") == "1"

	if bduss == "" || stoken == "" {
		return c.JSON(http.StatusOK, apiTemplate(401, "BDUSS 或 Stoken 无效", echoEmptyObject, "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "无法验证登录状态 BDUSS", echoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.R.Where("uid = ? AND portrait = ?", uid, baiduAccountInfo.User.Portrait).Limit(1).Find(&tiebaAccounts)

	if len(tiebaAccounts) > 0 {
		if tiebaAccounts[0].Bduss != bduss || tiebaAccounts[0].Stoken != stoken {
			newData := model.TcBaiduid{
				Bduss:    bduss,
				Stoken:   stoken,
				Name:     baiduAccountInfo.User.Name,
				Portrait: baiduAccountInfo.User.Portrait,
			}
			_function.GormDB.W.Model(model.TcBaiduid{}).Where("id = ?", tiebaAccounts[0].ID).Updates(&newData)
			newData.ID = tiebaAccounts[0].ID
			newData.UID = tiebaAccounts[0].UID
			if !includeBDUSSAndStoken {
				newData.Bduss = ""
				newData.Stoken = ""
			}
			return c.JSON(http.StatusOK, apiTemplate(200, "已更新 BDUSS", newData, "tbsign"))
		} else if tiebaAccounts[0].Bduss == bduss && tiebaAccounts[0].Stoken == stoken {
			if !includeBDUSSAndStoken {
				tiebaAccounts[0].Bduss = ""
				tiebaAccounts[0].Stoken = ""
			}
			return c.JSON(http.StatusOK, apiTemplate(200, "贴吧账号已存在", tiebaAccounts[0], "tbsign"))
		}
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	newAccount := model.TcBaiduid{
		UID:      int32(numUID),
		Bduss:    bduss,
		Stoken:   stoken,
		Name:     baiduAccountInfo.User.Name,
		Portrait: baiduAccountInfo.User.Portrait,
	}
	_function.GormDB.W.Create(&newAccount)
	if !includeBDUSSAndStoken {
		newAccount.Bduss = ""
		newAccount.Stoken = ""
	}
	return c.JSON(http.StatusOK, apiTemplate(201, "OK", newAccount, "tbsign"))
}

func RemoveTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	// fid pid
	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Find(&tiebaAccounts)

	for _, v := range tiebaAccounts {
		if v.ID == int32(numPid) {
			_function.GormDB.W.Where("id = ?", v.ID).Delete(&model.TcBaiduid{})
			_function.GormDB.W.Where("pid = ?", v.ID).Delete(&model.TcTieba{})

			// plugins
			_function.GormDB.W.Where("pid = ?", v.ID).Delete(&model.TcVer4BanList{})
			_function.GormDB.W.Where("pid = ?", v.ID).Delete(&model.TcVer4RankLog{})
			_function.GormDB.W.Where("pid = ?", v.ID).Delete(&model.TcKdGrowth{})

			return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]int32{
				"pid": v.ID,
			}, "tbsign"))
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(404, "Pid 不存在", echoEmptyObject, "tbsign"))
}

func GetTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	includeBDUSSAndStoken := c.QueryParams().Get("all") == "1"

	var tiebaAccounts []model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	if !includeBDUSSAndStoken {
		for k := range tiebaAccounts {
			tiebaAccounts[k].Bduss = ""
			tiebaAccounts[k].Stoken = ""
		}
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaAccounts, "tbsign"))
}

func GetTiebaAccountItem(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	includeBDUSSAndStoken := c.QueryParams().Get("all") == "1"

	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	var tiebaAccount model.TcBaiduid
	_function.GormDB.R.Where("id = ? AND uid = ?", numPid, uid).First(&tiebaAccount)

	if !includeBDUSSAndStoken {
		tiebaAccount.Bduss = ""
		tiebaAccount.Stoken = ""
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", tiebaAccount, "tbsign"))
}

func CheckTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
	}

	var tiebaAccount model.TcBaiduid
	_function.GormDB.R.Where("id = ? AND uid = ?", pid, uid).Order("id ASC").Find(&tiebaAccount)

	if tiebaAccount.ID != 0 && tiebaAccount.ID != int32(numPid) {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 pid", echoEmptyObject, "tbsign"))
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
