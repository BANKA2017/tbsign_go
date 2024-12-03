package _api

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
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

	BaiduAccountCount int `json:"baidu_account_count"`
	ForumCount        int `json:"forum_count"`

	// checkin status
	CheckinSuccess int `json:"checkin_success"`
	CheckinFailed  int `json:"checkin_failed"`
	CheckinWaiting int `json:"checkin_waiting"`
	CheckinIgnore  int `json:"checkin_ignore"`
}

func GetAdminSettings(c echo.Context) error {
	var adminSettings []*model.TcOption
	_function.GormDB.R.Where("name in ?", _function.SettingsFilter).Find(&adminSettings)

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

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", settings, "tbsign"))
}

func UpdateAdminSettings(c echo.Context) error {
	c.Request().ParseForm()

	var adminSettings []*model.TcOption
	_function.GormDB.R.Where("name in ?", _function.SettingsFilter).Find(&adminSettings)

	var errStr []string
	settings := make(map[string]string)
	for _, vName := range _function.SettingsFilter {
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
					case "enable_reg", "mail_auth", "go_export_personal_data", "go_import_personal_data":
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
					case "cron_limit", "retry_max", "sign_sleep", "mail_port", "cktime":
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
					case "go_ntfy_addr", "go_bark_addr":
						if _function.VerifyURL(v1[0]) {
							settings[vName] = v1[0]
						} else {
							errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
						}
					default:
						if PluginOptionValidatorAny, ok := _plugin.PluginOptionValidatorMap.Load(vName); ok {
							if PluginOptionValidator, ok := PluginOptionValidatorAny.(func(value string) bool); ok && PluginOptionValidator(v1[0]) {
								settings[vName] = v1[0]
								_function.SetOption(vName, v1[0])
							} else {
								errStr = append(errStr, vName+": Invalid value `"+v1[0]+"`")
								log.Println("invalid settings kv", vName, v1[0])
							}
						} else {
							settings[vName] = v1[0]
							_function.SetOption(vName, v1[0])
						}
					}
				}
				break
			}
		}
	}

	if len(errStr) == 0 {
		errStr = append(errStr, "OK")
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, strings.Join(errStr, "\n"), settings, "tbsign"))
}

func AdminModifyAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)
	targetUID := c.Param("uid")

	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyObject, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法修改自己的账号", _function.EchoEmptyObject, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyObject, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	_function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Find(&accountInfo)
	if accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyObject, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变管理员状态", _function.EchoEmptyObject, "tbsign"))
	}

	var newAccountInfo = accountInfo

	newName := c.FormValue("name")
	newEmail := c.FormValue("email")
	newRole := c.FormValue("role")

	// email
	if newEmail != "" && accountInfo.Email != newEmail {
		if !_function.VerifyEmail(newEmail) {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效邮箱", _function.EchoEmptyObject, "tbsign"))
		} else {
			var emailExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("email = ?", newEmail).Count(&emailExistsCount)
			if emailExistsCount > 0 {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "邮箱已存在", _function.EchoEmptyObject, "tbsign"))
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
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "用户名已存在", _function.EchoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Name = newName
		}
	} else {
		newName = accountInfo.Name
	}

	// role
	if newRole != "" && accountInfo.Role != newRole {
		if !slices.Contains(RoleList, newRole) {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "新用户组 "+newRole+" 不存在", _function.EchoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Role = newRole
		}
	} else {
		newRole = accountInfo.Role
	}

	// soft delete?
	if newRole == "deleted" {
		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			var err error
			// plugins
			if err = _plugin.DeleteAccount("uid", accountInfo.ID, tx); err != nil {
				return err
			}
			// account
			if err = tx.Where("id = ?", accountInfo.ID).Delete(&model.TcUser{}).Error; err != nil {
				return err
			}
			if err = tx.Where("uid = ?", accountInfo.ID).Delete(&model.TcTieba{}).Error; err != nil {
				return err
			}
			if err = tx.Where("uid = ?", accountInfo.ID).Delete(&model.TcBaiduid{}).Error; err != nil {
				return err
			}
			if err = tx.Where("uid = ?", accountInfo.ID).Delete(&model.TcUsersOption{}).Error; err != nil {
				return err
			}
			return err
		})

		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, fmt.Sprintf("删除用户 %d:%s 失败 (%s)", accountInfo.ID, accountInfo.Name, err.Error()), _function.EchoEmptyObject, "tbsign"))
		}

		keyBucket.Delete(strconv.Itoa(int(accountInfo.ID)))
	} else {
		_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", accountInfo.ID).Updates(&newAccountInfo)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", &SiteAccountsResponse{
		ID:    accountInfo.ID,
		Name:  newName,
		Email: newEmail,
		Role:  newRole,
		T:     accountInfo.T,
	}, "tbsign"))
}

func AdminResetTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	targetUID := c.Param("uid")
	resetFailedOnly := strings.TrimSpace(c.FormValue("failed_only")) != "0"

	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).First(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" && uid != targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变其他管理员状态", false, "tbsign"))
	}

	var PIDCount int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", targetUID).Count(&PIDCount)
	if PIDCount > 0 {
		var err error
		if resetFailedOnly {
			err = _function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ? AND status != 0", targetUID).Update("latest", 0).Error
		} else {
			err = _function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ?", targetUID).Update("latest", 0).Error
		}
		if err != nil {
			log.Println(err)
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", err == nil, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	}
}

func AdminDeleteTiebaAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	targetUID := c.Param("uid")

	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法修改自己的账号", false, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).First(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变管理员状态", false, "tbsign"))
	}

	var PIDCount int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", targetUID).Count(&PIDCount)
	if PIDCount > 0 {
		err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			var err error
			// plugins
			if err = _plugin.DeleteAccount("uid", int32(numTargetUID), tx); err != nil {
				return err
			}

			if err = tx.Where("uid = ?", targetUID).Delete(&model.TcBaiduid{}).Error; err != nil {
				return err
			}
			if err = tx.Where("uid = ?", targetUID).Delete(&model.TcTieba{}).Error; err != nil {
				return err
			}
			return err
		})
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, fmt.Sprintf("清空用户 %d:%s 的贴吧列表失败 (%s)", accountInfo.ID, accountInfo.Name, err.Error()), false, "tbsign"))
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	}
}

func AdminDeleteAccountToken(c echo.Context) error {
	uid := c.Get("uid").(string)
	targetUID := c.Param("uid")

	if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法踢自己下线", false, "tbsign"))
	}

	if _, ok := keyBucket.LoadAndDelete(targetUID); ok {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不在线上", false, "tbsign"))
	}
}

func GetAccountsList(c echo.Context) error {
	page := c.QueryParams().Get("page")
	count := c.QueryParams().Get("count")
	query := strings.TrimSpace(c.QueryParams().Get("q"))

	if page == "" {
		page = "1"
	}
	if count == "" {
		count = "10"
	}

	var respAccountInfo struct {
		List  []SiteAccountsResponse `json:"list"`
		Page  int64                  `json:"page"`
		Total int64                  `json:"total"`
	}
	numPage, err := strconv.ParseInt(page, 10, 64)
	if err != nil || numPage <= 0 {
		log.Println(err, page)
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "Invalid page", respAccountInfo, "tbsign"))
	}
	numCount, err := strconv.ParseInt(count, 10, 64)
	if err != nil || numCount <= 0 {
		log.Println(err, count)
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "Invalid count", respAccountInfo, "tbsign"))
	}

	var accountInfoCount int64
	_function.GormDB.R.Model(&model.TcUser{}).Count(&accountInfoCount)

	// TODO better query
	/// TODO any injection attacks?
	accountInfo := new([]SiteAccountsResponse)

	pidCountQuery := _function.GormDB.R.Model(&model.TcBaiduid{}).
		Select("uid, COUNT(*) AS pid_count").
		Group("uid")

	today := strconv.Itoa(_function.Now.Local().Day())
	forumCountQuery := _function.GormDB.R.Model(&model.TcTieba{}).
		Select("uid, COUNT(*) AS forum_count, SUM(CASE WHEN NOT no AND status = 0 AND latest = ? THEN 1 ELSE 0 END) AS success, SUM(CASE WHEN NOT no AND status <> 0 AND latest = ? THEN 1 ELSE 0 END) AS failed, SUM(CASE WHEN NOT no AND latest <> ? THEN 1 ELSE 0 END) AS waiting, SUM(CASE WHEN no THEN 1 ELSE 0 END) AS `ignore`", today, today, today).
		Group("uid")

	if query != "" {
		accountsQuery := _function.GormDB.R.Model(&model.TcUser{}).
			Select("*").
			Where("name LIKE ? OR email LIKE ?", _function.AppendStrings("%", query, "%"), _function.AppendStrings("%", query, "%")).
			Order("id").
			Limit(int(numCount)).
			Offset(int((numPage - 1) * numCount))

		_function.GormDB.R.Table("(?) as accounts", accountsQuery).
			Select("accounts.*, COALESCE(pid_count, 0) AS baidu_account_count, COALESCE(forum_count, 0) AS forum_count, COALESCE(success, 0) AS checkin_success, COALESCE(failed, 0) AS checkin_failed, COALESCE(waiting, 0) AS checkin_waiting, COALESCE(`ignore`, 0) AS checkin_ignore").
			Joins("LEFT JOIN (?) pid_counts ON accounts.id = pid_counts.uid", pidCountQuery).
			Joins("LEFT JOIN (?) forum_counts ON accounts.id = forum_counts.uid", forumCountQuery).
			Order("accounts.id").
			Scan(&accountInfo)

	} else {
		accountsQuery := _function.GormDB.R.Model(&model.TcUser{}).
			Order("id").
			Limit(int(numCount)).
			Offset(int((numPage - 1) * numCount))

		_function.GormDB.R.Table("(?) as accounts", accountsQuery).
			Select("accounts.*, COALESCE(pid_count, 0) AS baidu_account_count, COALESCE(forum_count, 0) AS forum_count,COALESCE(success, 0) AS checkin_success,COALESCE(failed, 0) AS checkin_failed,COALESCE(waiting, 0) AS checkin_waiting,COALESCE(`ignore`, 0) AS checkin_ignore").
			Joins("LEFT JOIN (?) pid_counts ON accounts.id = pid_counts.uid", pidCountQuery).
			Joins("LEFT JOIN (?) forum_counts ON accounts.id = forum_counts.uid", forumCountQuery).
			Order("accounts.id").
			Scan(&accountInfo)
	}

	if accountInfo != nil {
		respAccountInfo.List = *accountInfo
	}

	respAccountInfo.Page = numPage
	respAccountInfo.Total = accountInfoCount

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", respAccountInfo, "tbsign"))
}

func PluginSwitch(c echo.Context) error {
	c.Request().ParseForm()
	pluginName := c.Param("plugin_name")

	_pluginInfo, ok := _plugin.PluginList[pluginName]
	if !ok {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "插件不存在", map[string]any{
			"name":    pluginName,
			"exists":  false,
			"status":  false,
			"version": -1,
		}, "tbsign"))
	}

	// auto install
	if _pluginInfo.(_plugin.PluginHooks).GetDBInfo().Ver == "-1" {
		// TODO more flexible?
		_pluginInfo.Delete()
		err := _pluginInfo.Install()
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, "插件安装失败", map[string]any{
				"name":    pluginName,
				"exists":  false,
				"status":  false,
				"version": -1,
			}, "tbsign"))
		}
	}

	newPluginStatus := _pluginInfo.(_plugin.PluginHooks).Switch()

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"name":    pluginName,
		"exists":  true,
		"status":  newPluginStatus,
		"version": _pluginInfo.(_plugin.PluginHooks).GetDBInfo().Ver,
	}, "tbsign"))
}

func PluginUninstall(c echo.Context) error {
	c.Request().ParseForm()
	pluginName := c.Param("plugin_name")

	_pluginInfo, ok := _plugin.PluginList[pluginName]
	if !ok {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "插件不存在", map[string]any{
			"name":    pluginName,
			"exists":  false,
			"status":  false,
			"version": -1,
		}, "tbsign"))
	}

	if _pluginInfo.(_plugin.PluginHooks).GetDBInfo().Ver == "-1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, "插件尚未安装", map[string]any{
			"name":    pluginName,
			"exists":  false,
			"status":  false,
			"version": -1,
		}, "tbsign"))
	}

	err := _pluginInfo.Delete()

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, err.Error(), _function.EchoEmptyObject, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
			"name":    pluginName,
			"exists":  false,
			"status":  false,
			"version": -1,
		}, "tbsign"))
	}
}

func SendTestMessage(c echo.Context) error {
	uid := c.Get("uid").(string)

	messageType := c.QueryParams().Get("type")
	if !slices.Contains(_function.MessageTypeList, messageType) {
		messageType = "email"
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	messageObject := _function.PushMessageTestTemplate()
	err := _function.SendMessage(messageType, int32(numUID), messageObject.Title, messageObject.Body)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, err.Error(), false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	}
}
