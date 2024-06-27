package _api

import (
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
)

type SiteAccountsResponse struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
	T     string `json:"t"`
}

var settingsFilter = []string{"ann", "system_url", "system_name", "system_keywords", "system_description", "stop_reg", "enable_reg", "yr_reg", "sign_mode", "sign_hour", "cron_limit", "sign_sleep", "retry_max", "mail_name", "mail_yourname", "mail_host", "mail_port", "mail_secure", "mail_auth", "mail_smtpname", "mail_smtppw", "ver4_ban_limit", "ver4_ban_break_check"}

func GetAdminSettings(c echo.Context) error {
	var adminSettings []model.TcOption
	_function.GormDB.Where("name in ?", settingsFilter).Find(&adminSettings)

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
	_function.GormDB.Where("name in ?", settingsFilter).Find(&adminSettings)

	var errStr []string
	settings := make(map[string]string)
	for _, v := range adminSettings {
		for k1, v1 := range c.Request().Form {
			if v.Name == k1 && len(v1) > 0 {
				if v.Value != v1[0] {
					// verify
					switch v.Name {
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
							errStr = append(errStr, v.Name+": Invalid value `"+v1[0]+"`")
							continue
						}

						signModeEncoded, err := gophp.Serialize(signMode)
						if err != nil {
							errStr = append(errStr, "sign_mode: encode php serialize failed")
							log.Println("sign_mode: encode php serialize failed", err)
							continue
						}
						if string(signModeEncoded) != v.Value {
							settings[v.Name] = v1[0]
							_function.SetOption(v.Name, string(signModeEncoded))
						}

					case "enable_reg", "ver4_ban_break_check", "mail_auth":
						if v1[0] == "0" || v1[0] == "1" {
							settings[v.Name] = v1[0]
							_function.SetOption(v.Name, v1[0])
						} else {
							errStr = append(errStr, v.Name+": Invalid value `"+v1[0]+"`")
						}
					case "mail_secure":
						if v1[0] == "none" || v1[0] == "ssl" || v1[0] == "tls" {
							settings[v.Name] = v1[0]
							_function.SetOption(v.Name, v1[0])
						} else {
							errStr = append(errStr, v.Name+": Invalid value `"+v1[0]+"`")
						}
					case "cron_limit", "retry_max", "sign_sleep", "ver4_ban_limit", "mail_port":
						numValue, err := strconv.ParseInt(v1[0], 10, 64)
						if err == nil && numValue >= 0 {
							settings[v.Name] = v1[0]
							_function.SetOption(v.Name, v1[0])
						} else {
							errStr = append(errStr, v.Name+": Invalid value `"+v1[0]+"`")
							log.Println(v.Name, numValue, err)
						}
					case "sign_hour":
						numValue, err := strconv.ParseInt(v1[0], 10, 64)
						if err != nil {
							errStr = append(errStr, v.Name+": Invalid value `"+v1[0]+"`")
							log.Println(v.Name, err)
							continue
						} else if numValue < -1 {
							numValue = -1
						} else if numValue > 23 {
							numValue = 23
						}
						settings[v.Name] = strconv.Itoa(int(numValue))
						_function.SetOption(v.Name, settings[v.Name])
					default:
						settings[v.Name] = v1[0]
						_function.SetOption(v.Name, v1[0])
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

	pluginValue := _function.PluginList[pluginName]
	pluginValue.Status = !pluginValue.Status
	_function.GormDB.Model(&model.TcPlugin{}).Where("name = ?", pluginName).Update("status", pluginValue.Status)
	_function.PluginList[pluginName] = pluginValue

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"name":   pluginName,
		"exists": true,
		"status": pluginValue.Status,
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
