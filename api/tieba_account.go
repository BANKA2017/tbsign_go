package _api

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetLoginQRCode(c echo.Context) error {
	qrcode, err := _function.GetLoginQRCode()
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "获取二维码失败", _function.EchoEmptyObject, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", qrcode, "tbsign"))
}

func GetBDUSS(c echo.Context) error {
	uid := c.Get("uid").(string)
	isAdmin := strings.EqualFold(c.Get("role").(string), "admin")

	sign := c.FormValue("sign")

	res, err := _function.GetUnicastResponse(sign)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "获取状态失败", _function.EchoEmptyObject, "tbsign"))
	}

	// tmpBDUSS
	if res.ChannelV.V == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, "未确认", _function.EchoEmptyObject, "tbsign"))
	}

	res2, err := _function.GetLoginResponse(res.ChannelV.V)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "登录失败", _function.EchoEmptyObject, "tbsign"))
	}

	stokenStr := strings.ReplaceAll(res2.Data.Session.StokenList, "&quot;", "\"")
	var stokenArray []string
	err = _function.JsonDecode([]byte(stokenStr), &stokenArray)
	if err != nil || res2.Data.Session.Bduss == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "登录失败", _function.EchoEmptyObject, "tbsign"))
	}
	bduss := res2.Data.Session.Bduss

	stokenKV := make(map[string]string, len(stokenArray))
	for _, v := range stokenArray {
		tmpSplit := strings.Split(v, "#")
		stokenKV[tmpSplit[0]] = tmpSplit[1]
	}

	if stoken, ok := stokenKV["tb"]; !ok || stoken == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "登录失败", _function.EchoEmptyObject, "tbsign"))
	}
	stoken := stokenKV["tb"]

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无法验证登录状态 BDUSS", _function.EchoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ? AND portrait = ?", uid, baiduAccountInfo.User.Portrait).Limit(1).Find(&tiebaAccounts)

	if len(share.DataEncryptKeyByte) > 0 {
		encryptedBDUSS, _ := _function.AES256GCMEncrypt(bduss, share.DataEncryptKeyByte)
		bduss = _function.Base64URLEncode(encryptedBDUSS)

		encryptedStoken, _ := _function.AES256GCMEncrypt(stoken, share.DataEncryptKeyByte)
		stoken = _function.Base64URLEncode(encryptedStoken)
	}

	if len(tiebaAccounts) > 0 {
		if tiebaAccounts[0].Bduss != bduss || tiebaAccounts[0].Stoken != stoken {
			newData := model.TcBaiduid{
				Bduss:    bduss,
				Stoken:   stoken,
				Name:     baiduAccountInfo.User.Name,
				Portrait: baiduAccountInfo.User.Portrait,
			}
			_function.GormDB.W.Model(&model.TcBaiduid{}).Where("id = ?", tiebaAccounts[0].ID).Updates(&newData)
			newData.ID = tiebaAccounts[0].ID
			newData.UID = tiebaAccounts[0].UID
			newData.Bduss = ""
			newData.Stoken = ""

			_function.CookieList.Delete(tiebaAccounts[0].ID)
			_function.GormDB.W.Model(&model.TcTieba{}).Select("status", "latest", "last_error").Where("pid = ? AND status = ?", tiebaAccounts[0].ID, 110000).Updates(&model.TcTieba{
				Status:    0,
				Latest:    0,
				LastError: "等待重签",
			})

			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "已更新 BDUSS", newData, "tbsign"))
		} else if tiebaAccounts[0].Bduss == bduss && tiebaAccounts[0].Stoken == stoken {
			tiebaAccounts[0].Bduss = ""
			tiebaAccounts[0].Stoken = ""
			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "贴吧账号已存在", tiebaAccounts[0], "tbsign"))
		}
	}

	// bduss-num
	if !isAdmin {
		bdussNUM := _function.GetOption("bduss_num")
		numBDUSSLimit, err := strconv.ParseInt(bdussNUM, 10, 64)
		if err != nil || numBDUSSLimit <= -1 {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法添加更多账号", _function.EchoEmptyObject, "tbsign"))
		}

		if numBDUSSLimit > 0 {
			var tiebaAccountsCount int64
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Count(&tiebaAccountsCount)

			if tiebaAccountsCount >= numBDUSSLimit {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法添加更多账号", _function.EchoEmptyObject, "tbsign"))
			}
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

	return c.JSON(http.StatusOK, _function.ApiTemplate(201, "OK", newAccount, "tbsign"))
}

func AddTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)
	isAdmin := strings.EqualFold(c.Get("role").(string), "admin")

	bduss := strings.TrimSpace(c.FormValue("bduss"))
	stoken := strings.TrimSpace(c.FormValue("stoken"))

	if bduss == "" || stoken == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(401, "BDUSS 或 Stoken 无效", _function.EchoEmptyObject, "tbsign"))
	}

	// get tieba account info
	baiduAccountInfo, err := _function.GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	if err != nil || baiduAccountInfo.User.Portrait == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无法验证登录状态 BDUSS", _function.EchoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ? AND portrait = ?", uid, baiduAccountInfo.User.Portrait).Limit(1).Find(&tiebaAccounts)

	if len(share.DataEncryptKeyByte) > 0 {
		encryptedBDUSS, _ := _function.AES256GCMEncrypt(bduss, share.DataEncryptKeyByte)
		bduss = _function.Base64URLEncode(encryptedBDUSS)

		encryptedStoken, _ := _function.AES256GCMEncrypt(stoken, share.DataEncryptKeyByte)
		stoken = _function.Base64URLEncode(encryptedStoken)
	}

	if len(tiebaAccounts) > 0 {
		if tiebaAccounts[0].Bduss != bduss || tiebaAccounts[0].Stoken != stoken {
			newData := model.TcBaiduid{
				Bduss:    bduss,
				Stoken:   stoken,
				Name:     baiduAccountInfo.User.Name,
				Portrait: baiduAccountInfo.User.Portrait,
			}
			_function.GormDB.W.Model(&model.TcBaiduid{}).Where("id = ?", tiebaAccounts[0].ID).Updates(&newData)
			newData.ID = tiebaAccounts[0].ID
			newData.UID = tiebaAccounts[0].UID

			newData.Bduss = ""
			newData.Stoken = ""

			_function.CookieList.Delete(tiebaAccounts[0].ID)
			_function.GormDB.W.Model(&model.TcTieba{}).Select("status", "latest", "last_error").Where("pid = ? AND status = ?", tiebaAccounts[0].ID, 110000).Updates(&model.TcTieba{
				Status:    0,
				Latest:    0,
				LastError: "等待重签",
			})

			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "已更新 BDUSS", newData, "tbsign"))
		} else if tiebaAccounts[0].Bduss == bduss && tiebaAccounts[0].Stoken == stoken {
			tiebaAccounts[0].Bduss = ""
			tiebaAccounts[0].Stoken = ""

			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "贴吧账号已存在", tiebaAccounts[0], "tbsign"))
		}
	}

	// bduss-num
	if !isAdmin {
		bdussNUM := _function.GetOption("bduss_num")
		numBDUSSLimit, err := strconv.ParseInt(bdussNUM, 10, 64)
		if err != nil || numBDUSSLimit <= -1 {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法添加更多账号", _function.EchoEmptyObject, "tbsign"))
		}

		if numBDUSSLimit > 0 {
			var tiebaAccountsCount int64
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Count(&tiebaAccountsCount)

			if tiebaAccountsCount >= numBDUSSLimit {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法添加更多账号", _function.EchoEmptyObject, "tbsign"))
			}
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

	return c.JSON(http.StatusOK, _function.ApiTemplate(201, "OK", newAccount, "tbsign"))
}

func RemoveTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// fid pid
	var tiebaAccount model.TcBaiduid
	err = _function.GormDB.R.Where("uid = ? AND id = ?", uid, numPid).Find(&tiebaAccount).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "Pid 不存在", _function.EchoEmptyObject, "tbsign"))
	} else if err != nil {
		log.Println("remove-tieba-account", uid, pid, err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyObject, "tbsign"))
	}

	err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		var err error

		// plugins
		if err = _plugin.DeleteAccount("pid", int32(numPid), tx); err != nil {
			return err
		}
		if err = tx.Where("id = ?", numPid).Delete(&model.TcBaiduid{}).Error; err != nil {
			return err
		}
		if err = tx.Where("pid = ?", numPid).Delete(&model.TcTieba{}).Error; err != nil {
			return err
		}
		return err
	})

	if err != nil {
		log.Println("remove-tieba-account", uid, pid, err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyObject, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]int32{
		"pid": int32(numPid),
	}, "tbsign"))

}

func GetTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	arrayMode := IsArrayMode(c)

	var tiebaAccounts []*model.TcBaiduid
	var tiebaAccountsBatchQueryList []*model.TcBaiduid
	// _function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	_function.GormDB.R.Select("id, uid, name, portrait").Where("uid = ?", uid).Order("id ASC").FindInBatches(&tiebaAccountsBatchQueryList, 1000, func(tx *gorm.DB, batch int) error {
		tiebaAccounts = append(tiebaAccounts, tiebaAccountsBatchQueryList...)
		return nil
	})

	// for k := range tiebaAccounts {
	// 	tiebaAccounts[k].Bduss = ""
	// 	tiebaAccounts[k].Stoken = ""
	// }

	if arrayMode {
		listArray := make([][4]any, len(tiebaAccounts))
		for i, accountInfo := range tiebaAccounts {
			listArray[i] = [4]any{
				accountInfo.ID,
				accountInfo.UID,
				// accountInfo.Bduss,
				// accountInfo.Stoken,
				accountInfo.Name,
				accountInfo.Portrait,
			}
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", listArray, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaAccounts, "tbsign"))
	}
}

func GetTiebaAccountItem(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	var tiebaAccount model.TcBaiduid
	_function.GormDB.R.Where("id = ? AND uid = ?", numPid, uid).Take(&tiebaAccount)

	tiebaAccount.Bduss = ""
	tiebaAccount.Stoken = ""

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaAccount, "tbsign"))
}

func CheckTiebaAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.Param("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	cookie := _function.GetCookie(int32(numPid), false, true)

	if cookie.ID == 0 || cookie.UID != int32(numUID) || (cookie.ID != 0 && cookie.ID != int32(numPid)) {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", cookie.IsLogin, "tbsign"))
}

func CheckIsManager(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fname := c.Param("fname")

	// pre-check pid
	var pidCheck []*model.TcBaiduid
	_function.GormDB.R.Where("id = ? AND uid = ?", pid, uid).Limit(1).Find(&pidCheck)

	if len(pidCheck) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _type.IsManagerPreCheckResponse{}, "tbsign"))
	}

	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _type.IsManagerPreCheckResponse{}, "tbsign"))
	}
	resp, err := _function.GetManagerStatus(pidCheck[0].Portrait, fid)
	if err != nil {
		log.Println(err)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}
