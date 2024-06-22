package _api

import (
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
)

type SiteAccountsResponse struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	T     string `json:"t"`
}

// ?
// var InviteCodeList = make(map[string]string)
var settingsFilter = []string{"ann", "system_url", "system_keywords", "system_name", "system_description", "enable_reg", "yr_reg", "ver4_ban_limit", "ver4_ban_break_check", "mail_name", "mail_yourname", "mail_host", "mail_port", "mail_secure", "mail_auth", "mail_smtpname", "mail_smtppw", "ver4_ref_lastdo", "sign_mode", "sign_hour", "cron_limit", "sign_sleep", "cron_sign_again", "retry_max"}

func GetAdminSettings(c echo.Context) error {
	var adminSettings []model.TcOption
	_function.GormDB.Where("name in ?", settingsFilter).Find(&adminSettings)

	settings := make(map[string]string)
	for _, v := range adminSettings {
		settings[v.Name] = v.Value
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func UpdateAdminSettings(c echo.Context) error {
	c.Request().ParseForm()

	var adminSettings []model.TcOption
	_function.GormDB.Find(&adminSettings)

	var newSettings []model.TcOption

	for _, v := range adminSettings {
		for k1, v1 := range c.Request().Form {
			if v.Name == k1 {
				if v.Value != v1[0] {
					v.Value = v1[0]
					newSettings = append(newSettings, v)
				}
				break
			}
		}
	}

	settings := make(map[string]string)
	if len(newSettings) > 0 {
		for _, v := range newSettings {
			settings[v.Name] = v.Value
			_function.GormDB.Model(model.TcOption{}).Where("name = ?", v.Name).Updates(&model.TcOption{
				Value: v.Value,
			})
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func AdminUpdateAccount(c echo.Context) error {
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
	_function.GormDB.Model([]model.TcUser{}).Where("id = ?", targetUID).Find(&accountInfo)
	if accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, apiTemplate(404, "用户不存在", echoEmptyObject, "tbsign"))
	}
	if accountInfo.Role == "admin" {
		return c.JSON(http.StatusOK, apiTemplate(403, "无法修改管理员帐号", echoEmptyObject, "tbsign"))
	}

	var newAccountInfo model.TcUser

	newName := c.FormValue("name")
	newEmail := c.FormValue("email")
	newRole := c.FormValue("role")

	// email
	if newEmail != "" {
		if !_function.VerifyEmail(newEmail) {
			return c.JSON(http.StatusOK, apiTemplate(403, "无效邮箱", echoEmptyObject, "tbsign"))
		} else {
			var emailExistsCount int64
			_function.GormDB.Model(&model.TcUser{}).Where("email = ?", newEmail).Count(&emailExistsCount)
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
		_function.GormDB.Model(&model.TcUser{}).Where("name = ?", newName).Count(&nameExistsCount)
		if nameExistsCount > 0 {
			return c.JSON(http.StatusOK, apiTemplate(403, "用户名已存在", echoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Name = newName
		}
	} else {
		newName = accountInfo.Name
	}

	// role
	if newRole != "" {
		if !slices.Contains(RoleList, newRole) {
			return c.JSON(http.StatusOK, apiTemplate(403, "新用户组 "+newRole+" 不存在", echoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Role = newRole
		}
	} else {
		newRole = accountInfo.Role
	}

	_function.GormDB.Model(model.TcUser{}).Where("id = ?", accountInfo.ID).Updates(&newAccountInfo)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", &SiteAccountsResponse{
		ID:    accountInfo.ID,
		Name:  newName,
		Email: newEmail,
		Role:  newRole,
		T:     accountInfo.T,
	}, "tbsign"))
}

func GetAccountsList(c echo.Context) error {
	var accountInfo []model.TcUser
	_function.GormDB.Find(&accountInfo)

	var respAccountInfo []SiteAccountsResponse

	for _, v := range accountInfo {
		respAccountInfo = append(respAccountInfo, SiteAccountsResponse{
			ID:    v.ID,
			Name:  v.Name,
			Email: v.Email,
			Role:  v.Role,
			T:     v.T,
		})
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", respAccountInfo, "tbsign"))
}

func GetPluginsList(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", _function.PluginList, "tbsign"))
}

func PluginSwitch(c echo.Context) error {
	c.Request().ParseForm()
	pluginName := c.Param("plugin_name")

	_function.GetOptionsAndPluginList()

	if _, ok := _function.PluginList[pluginName]; !ok {
		return c.JSON(http.StatusOK, apiTemplate(404, "插件不存在", map[string]any{
			"name":   pluginName,
			"exists": false,
			"status": false,
		}, "tbsign"))
	}
	newStatus := !_function.PluginList[pluginName]
	_function.GormDB.Model(&model.TcPlugin{}).Where("name = ?", pluginName).Update("status", newStatus)
	_function.PluginList[pluginName] = newStatus

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"name":   pluginName,
		"exists": true,
		"status": newStatus,
	}, "tbsign"))
}

func SendTestMail(c echo.Context) error {
	uid := c.Get("uid").(string)
	var accountInfo model.TcUser

	_function.GormDB.Where("id = ?", uid).Find(&accountInfo)

	mailObject := _function.EmailTestTemplate()
	err := _function.SendEmail(accountInfo.Email, mailObject.Object, mailObject.Body)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, err.Error(), false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}
}
