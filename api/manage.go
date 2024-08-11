package _api

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
	"github.com/leeqvip/gophp"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type SiteAccountsResponse struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	T     string `json:"t"`
}

var settingsFilter = []string{"ann", "system_url", "stop_reg", "enable_reg", "yr_reg", "cktime", "sign_mode", "sign_hour", "cron_limit", "sign_sleep", "retry_max", "mail_name", "mail_yourname", "mail_host", "mail_port", "mail_secure", "mail_auth", "mail_smtpname", "mail_smtppw", "ver4_ban_limit", "ver4_ban_break_check", "go_forum_sync_policy"} // "system_name", "system_keywords", "system_description"

func GetAdminSettings(c echo.Context) error {
	var adminSettings []model.TcOption
	_function.GormDB.R.Where("name in ?", settingsFilter).Find(&adminSettings)

	settings := make(map[string]string)
	for _, v := range adminSettings {
		if v.Name == "sign_mode" {
			tmpOption := []string{}
			for _, match := range regexp.MustCompile(`(?m)\"([123])\"`).FindAllString(v.Value, -1) {
				tmpOption = append(tmpOption, strings.ReplaceAll(match, "\"", ""))
			}
			settings[v.Name] = strings.Join(tmpOption, ",")
		} else {
			settings[v.Name] = v.Value
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func UpdateAdminSettings(c echo.Context) error {
	c.Request().ParseForm()

	var adminSettings []model.TcOption
	_function.GormDB.R.Where("name in ?", settingsFilter).Find(&adminSettings)

	var errStr []string
	settings := make(map[string]string)
	for _, vName := range settingsFilter {
		for k1, v1 := range c.Request().Form {
			if vName == k1 && len(v1) > 0 {
				vValue := _function.GetOption(vName)
				if vValue != v1[0] {
					// verify
					switch vName {
					case "sign_mode":
						signMode := strings.Split(v1[0], ",")
						check := true
						for _, signModeItem := range signMode {
							if !slices.Contains([]string{"1", "2", "3"}, signModeItem) {
								check = false
								break
							}
						}
						if !check {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
							continue
						}

						signModeEncoded, err := gophp.Serialize(signMode)
						if err != nil {
							errStr = append(errStr, "sign_mode: encode php serialize failed")
							log.Println("sign_mode: encode php serialize failed", err)
							continue
						}
						if string(signModeEncoded) != vValue {
							settings[vName] = v1[0]
							_function.SetOption(vName, string(signModeEncoded))
						}

					case "enable_reg", "ver4_ban_break_check", "mail_auth":
						if v1[0] == "0" || v1[0] == "1" {
							settings[vName] = v1[0]
							_function.SetOption(vName, v1[0])
						} else {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
						}
					case "mail_secure":
						if v1[0] == "none" || v1[0] == "ssl" || v1[0] == "tls" {
							settings[vName] = v1[0]
							_function.SetOption(vName, v1[0])
						} else {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
						}
					case "cron_limit", "retry_max", "sign_sleep", "ver4_ban_limit", "mail_port", "cktime":
						numValue, err := strconv.ParseInt(v1[0], 10, 64)
						if err == nil && numValue >= 0 {
							settings[vName] = v1[0]
							_function.SetOption(vName, v1[0])
						} else {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
							log.Println(vName, numValue, err)
						}
					case "sign_hour":
						numValue, err := strconv.ParseInt(v1[0], 10, 64)
						if err != nil {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
							log.Println(vName, err)
							continue
						} else if numValue < -1 {
							numValue = -1
						} else if numValue > 23 {
							numValue = 23
						}
						settings[vName] = strconv.Itoa(int(numValue))
						_function.SetOption(vName, settings[vName])
					case "go_forum_sync_policy":
						if v1[0] == "add_delete" || v1[0] == "add_only" || v1[0] == "" {
							if v1[0] == "" {
								settings[vName] = "add_only"
							} else {
								settings[vName] = v1[0]
							}
							err := _function.SetOption(vName, settings[vName])
							log.Println(err)
						} else {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
						}
					default:
						settings[vName] = v1[0]
						_function.SetOption(vName, v1[0])
					}
				}
				break
			}
		}
	}

	if len(errStr) == 0 {
		errStr = append(errStr, "OK")
	}

	return c.JSON(http.StatusOK, apiTemplate(200, strings.Join(errStr, "\n"), settings, "tbsign"))
}

func AdminModifyAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)
	targetUID := c.Param("uid")

	if targetUID == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", echoEmptyObject, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, apiTemplate(403, "无法修改自己的帐号", echoEmptyObject, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", echoEmptyObject, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	_function.GormDB.R.Model([]model.TcUser{}).Where("id = ?", targetUID).Find(&accountInfo)
	if accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", echoEmptyObject, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" {
		return c.JSON(http.StatusOK, apiTemplate(403, "只有根管理员允许改变管理员状态", echoEmptyObject, "tbsign"))
	}

	var newAccountInfo = accountInfo

	newName := c.FormValue("name")
	newEmail := c.FormValue("email")
	newRole := c.FormValue("role")

	// email
	if newEmail != "" && accountInfo.Email != newEmail {
		if !_function.VerifyEmail(newEmail) {
			return c.JSON(http.StatusOK, apiTemplate(403, "无效邮箱", echoEmptyObject, "tbsign"))
		} else {
			var emailExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("email = ?", newEmail).Count(&emailExistsCount)
			if emailExistsCount > 0 {
				return c.JSON(http.StatusOK, apiTemplate(403, "邮箱已存在", echoEmptyObject, "tbsign"))
			} else {
				newAccountInfo.Email = newEmail
			}
		}
	} else {
		newEmail = accountInfo.Email
	}

	// name
	if newName != "" && newName != accountInfo.Name {
		var nameExistsCount int64
		_function.GormDB.R.Model(&model.TcUser{}).Where("name = ?", newName).Count(&nameExistsCount)
		if nameExistsCount > 0 {
			return c.JSON(http.StatusOK, apiTemplate(403, "用户名已存在", echoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Name = newName
		}
	} else {
		newName = accountInfo.Name
	}

	// role
	if newRole != "" && accountInfo.Role != newRole {
		if !slices.Contains(RoleList, newRole) {
			return c.JSON(http.StatusOK, apiTemplate(403, "新用户组 "+newRole+" 不存在", echoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Role = newRole
		}
	} else {
		newRole = accountInfo.Role
	}

	// soft delete?
	if newRole == "delete" {
		// account
		_function.GormDB.W.Where("id = ?", accountInfo.ID).Delete(&model.TcUser{})
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcTieba{})
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcBaiduid{})
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcUsersOption{})

		// plugins
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcVer4BanList{})
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcVer4RankLog{})
		_function.GormDB.W.Where("uid = ?", accountInfo.ID).Delete(&model.TcKdGrowth{})
		keyBucket.Delete(strconv.Itoa(int(accountInfo.ID)))
	} else {
		_function.GormDB.W.Model(model.TcUser{}).Where("id = ?", accountInfo.ID).Updates(&newAccountInfo)
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", &SiteAccountsResponse{
		ID:    accountInfo.ID,
		Name:  newName,
		Email: newEmail,
		Role:  newRole,
		T:     accountInfo.T,
	}, "tbsign"))
}

func AdminDeleteTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	targetUID := c.Param("uid")

	if targetUID == "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", false, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, apiTemplate(403, "无法修改自己的帐号", false, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", false, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model([]model.TcUser{}).Where("id = ?", targetUID).First(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", false, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" {
		return c.JSON(http.StatusOK, apiTemplate(403, "只有根管理员允许改变管理员状态", false, "tbsign"))
	}

	var PIDCount int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", targetUID).Count(&PIDCount)
	if PIDCount > 0 {
		_function.GormDB.W.Where("uid = ?", targetUID).Delete(&model.TcBaiduid{})
		_function.GormDB.W.Where("uid = ?", targetUID).Delete(&model.TcTieba{})

		// plugins
		_function.GormDB.W.Where("uid = ?", targetUID).Delete(&model.TcVer4BanList{})
		_function.GormDB.W.Where("uid = ?", targetUID).Delete(&model.TcVer4RankLog{})
		_function.GormDB.W.Where("uid = ?", targetUID).Delete(&model.TcKdGrowth{})
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}
}

func AdminDeleteAccountToken(c echo.Context) error {
	uid := c.Get("uid").(string)
	targetUID := c.Param("uid")

	if uid == targetUID {
		return c.JSON(http.StatusOK, apiTemplate(403, "无法踢自己下线", false, "tbsign"))
	}

	if _, ok := keyBucket.LoadAndDelete(targetUID); ok {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不在线上", false, "tbsign"))
	}
}

func GetAccountsList(c echo.Context) error {
	page := c.QueryParams().Get("page")
	count := c.QueryParams().Get("count")

	if page == "" {
		page = "1"
	}
	if count == "" {
		count = "10"
	}

	var respAccountInfo []SiteAccountsResponse
	numPage, err := strconv.ParseInt(page, 10, 64)
	if err != nil || numPage <= 0 {
		log.Println(err, page)
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid page", respAccountInfo, "tbsign"))
	}
	numCount, err := strconv.ParseInt(count, 10, 64)
	if err != nil || numCount <= 0 {
		log.Println(err, count)
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid count", respAccountInfo, "tbsign"))
	}

	var accountInfoCount int64
	_function.GormDB.R.Model(&model.TcUser{}).Count(&accountInfoCount)

	var accountInfo []model.TcUser
	_function.GormDB.R.Offset(int((numPage - 1) * numCount)).Limit(int(numCount)).Find(&accountInfo)

	for _, v := range accountInfo {
		respAccountInfo = append(respAccountInfo, SiteAccountsResponse{
			ID:    v.ID,
			Name:  v.Name,
			Email: v.Email,
			Role:  v.Role,
			T:     v.T,
		})
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"list":  respAccountInfo,
		"page":  numPage,
		"total": accountInfoCount,
	}, "tbsign"))
}

func PluginSwitch(c echo.Context) error {
	c.Request().ParseForm()
	pluginName := c.Param("plugin_name")

	_function.GetOptionsAndPluginList()
	_pluginValue, ok := _function.PluginList.Load(pluginName)
	if !ok || _pluginValue == nil {
		return c.JSON(http.StatusOK, apiTemplate(404, "插件不存在", map[string]any{
			"name":   pluginName,
			"exists": false,
			"status": false,
		}, "tbsign"))
	}

	pluginValue := _pluginValue.(model.TcPlugin)
	newPluginStatus := !pluginValue.Status

	pluginValue.Status = newPluginStatus
	_function.GormDB.W.Model(&model.TcPlugin{}).Where("name = ?", pluginName).Update("status", newPluginStatus)
	_function.PluginList.Store(pluginName, pluginValue)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"name":   pluginName,
		"exists": true,
		"status": newPluginStatus,
	}, "tbsign"))
}

func SendTestMail(c echo.Context) error {
	uid := c.Get("uid").(string)
	var accountInfo model.TcUser

	_function.GormDB.R.Where("id = ?", uid).Find(&accountInfo)

	mailObject := _function.EmailTestTemplate()
	err := _function.SendEmail(accountInfo.Email, mailObject.Object, mailObject.Body)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, err.Error(), false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}
}
