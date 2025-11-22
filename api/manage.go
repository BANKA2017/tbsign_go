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

	settings := make(map[string]string, len(adminSettings))
	for _, v := range adminSettings {
		if v.Name == "sign_mode" {
			var tmpOption []string
			for _, match := range regexp.MustCompile(`(?m)"([123])"`).FindAllString(v.Value, -1) {
				tmpOption = append(tmpOption, strings.ReplaceAll(match, "\"", ""))
			}
			settings[v.Name] = strings.Join(tmpOption, ",")
		} else {
			settings[v.Name] = v.Value
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", settings, "tbsign"))
}

func encodeSignMode(val []string) string {
	// a:1:{i:0;s:1:"1";}
	var sb strings.Builder
	sb.Grow(6 + len(val)*12)
	sb.WriteString(`a:` + strconv.Itoa(len(val)) + `:{`)

	for i := range val {
		sb.WriteString("i:" + strconv.Itoa(i) + `;s:1:"` + val[i] + `";`)
	}

	sb.WriteByte('}')
	return sb.String()
}

var SettingsRules = map[string]*_function.OptionRule{
	"sign_mode": {
		Custom: func(val string) error {
			if val == "" {
				return nil
			}
			signMode := strings.SplitSeq(val, ",")
			for v := range signMode {
				if !slices.Contains([]string{"1", "2", "3"}, v) {
					return errors.New("invalid value " + v)
				}
			}
			return nil
		},
		Transform: func(val string) string {
			signMode := strings.Split(val, ",")

			if len(signMode) == 0 {
				signMode = append(signMode, "1")
			}

			return encodeSignMode(signMode)
		},
	},
	"mail_auth":                   {Enum: []string{"0", "1", "2"}},
	"enable_reg":                  {Enum: []string{"0", "1"}},
	"go_export_personal_data":     {Enum: []string{"0", "1"}},
	"go_import_personal_data":     {Enum: []string{"0", "1"}},
	"mail_secure":                 {Enum: []string{"none", "ssl", "tls"}},
	"cron_limit":                  {Min: _function.VPtr(int64(0))},
	"retry_max":                   {Min: _function.VPtr(int64(0))},
	"sign_sleep":                  {Min: _function.VPtr(int64(0))},
	"sign_multith":                {Min: _function.VPtr(int64(0))},
	"mail_port":                   {Min: _function.VPtr(int64(0)), Max: _function.VPtr(int64(65535))},
	"cktime":                      {Min: _function.VPtr(int64(0))},
	"go_re_check_in_max_interval": {Min: _function.VPtr(int64(1))},

	"sign_hour":            {Min: _function.VPtr(int64(-1)), Max: _function.VPtr(int64(23))},
	"go_daily_report_hour": {Min: _function.VPtr(int64(-1)), Max: _function.VPtr(int64(23))},

	"go_forum_sync_policy": {
		Enum: []string{"add_delete", "add_only", ""},
		Transform: func(val string) string {
			if val == "" {
				return "add_only"
			}
			return val
		},
	},

	"go_ntfy_addr":     {IsURL: true},
	"go_bark_addr":     {IsURL: true},
	"go_pushdeer_addr": {IsURL: true},

	"bduss_num": {Min: _function.VPtr(int64(-1)), Max: _function.VPtr(int64(999999999))},
	// "tb_max": {Min: _function.VPtr(int64(-1)), Max: _function.VPtr(int64(10000))},
}

func UpdateAdminSettings(c echo.Context) error {
	var errStr []string
	settings := make(map[string]string)

	for _, key := range _function.SettingsFilter {
		val := c.Request().FormValue(key)
		if val == "" {
			continue
		}

		oldVal := _function.GetOption(key)
		if oldVal == val {
			continue
		}

		var validator *_function.OptionRule

		if pluginValidator, ok := _plugin.PluginOptionValidatorMap.Load(key); ok {
			validator = pluginValidator

		} else if optionRule, ok := SettingsRules[key]; ok {
			validator = optionRule
		}

		if validator != nil {
			newVal, err := _function.ValidateOptionValue(val, validator)
			if err != nil {
				errStr = append(errStr, key+": "+err.Error())
				continue
			}
			if newVal == oldVal {
				continue
			}

			settings[key] = newVal
			_function.SetOption(key, newVal)
		} else {
			settings[key] = val
			_function.SetOption(key, val)
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
		if newRole == "deleted" {
			return c.JSON(http.StatusOK, _function.ApiTemplate(400, "请使用 DELETE:/admin/account/:uid 删除用户", _function.EchoEmptyObject, "tbsign"))
		} else if !slices.Contains(RoleList, newRole) {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "新用户组 "+newRole+" 不存在", _function.EchoEmptyObject, "tbsign"))
		} else {
			newAccountInfo.Role = newRole
		}
	} else {
		newRole = accountInfo.Role
	}

	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", accountInfo.ID).Updates(&newAccountInfo)

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
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Take(&accountInfo).Error
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
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Take(&accountInfo).Error
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

	numTargetUID, _ := strconv.ParseInt(targetUID, 10, 64)

	if _, err := DeleteSessionExpiredAt(strconv.Itoa(int(numTargetUID))); err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "令牌错误", false, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func AdminResetPassword(c echo.Context) error {
	uid := c.Get("uid").(string)
	targetUID := c.Param("uid")

	var resetCodeResponse _function.VerifyCodeStruct

	if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "您已登录，请到首页修改密码", resetCodeResponse, "tbsign"))
	}
	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", resetCodeResponse, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法修改自己的账号", resetCodeResponse, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", resetCodeResponse, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Take(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", resetCodeResponse, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变管理员状态", resetCodeResponse, "tbsign"))
	}

	resetCodeResponse = *ResetMessageBuilder(int32(numTargetUID), true)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resetCodeResponse, "tbsign"))
}

func GetAccountsList(c echo.Context) error {
	page := c.QueryParam("page")
	count := c.QueryParam("count")
	query := strings.TrimSpace(c.QueryParam("q"))

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

	today := strconv.Itoa(_function.Now.Day())
	forumCountQuery := _function.GormDB.R.Model(&model.TcTieba{}).
		Select("uid, COUNT(*) AS forum_count, SUM(CASE WHEN (no = 0) AND status = 0 AND latest = ? THEN 1 ELSE 0 END) AS success, SUM(CASE WHEN (no = 0) AND status <> 0 AND latest = ? THEN 1 ELSE 0 END) AS failed, SUM(CASE WHEN (no = 0) AND latest <> ? THEN 1 ELSE 0 END) AS waiting, SUM(CASE WHEN no <> 0 THEN 1 ELSE 0 END) AS is_ignore", today, today, today).
		Group("uid")

	if query != "" {
		accountsQuery := _function.GormDB.R.Model(&model.TcUser{}).
			Select("*").
			Where("name LIKE ? OR email LIKE ?", "%"+query+"%", "%"+query+"%").
			Order("id").
			Limit(int(numCount)).
			Offset(int((numPage - 1) * numCount))

		_function.GormDB.R.Table("(?) as accounts", accountsQuery).
			Select("accounts.*, COALESCE(pid_count, 0) AS baidu_account_count, COALESCE(forum_count, 0) AS forum_count, COALESCE(success, 0) AS checkin_success, COALESCE(failed, 0) AS checkin_failed, COALESCE(waiting, 0) AS checkin_waiting, COALESCE(is_ignore, 0) AS checkin_ignore").
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
			Select("accounts.*, COALESCE(pid_count, 0) AS baidu_account_count, COALESCE(forum_count, 0) AS forum_count,COALESCE(success, 0) AS checkin_success,COALESCE(failed, 0) AS checkin_failed,COALESCE(waiting, 0) AS checkin_waiting,COALESCE(is_ignore, 0) AS checkin_ignore").
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

func AdminDeleteAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	targetUID := c.Param("uid")

	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyObject, "tbsign"))
	} else if uid == targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无法删除自己的账号", _function.EchoEmptyObject, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Take(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" && uid != targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变其他管理员状态", false, "tbsign"))
	}

	err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
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
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, fmt.Sprintf("删除用户 %d:%s 失败 (%s)", accountInfo.ID, accountInfo.Name, err.Error()), false, "tbsign"))
	}

	//HttpAuthRefreshTokenMap.Delete(int(accountInfo.ID))
	_function.PasswordCache.Delete(int(accountInfo.ID))

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func AdminResetAccountPlugin(c echo.Context) error {
	uid := c.Get("uid").(string)

	c.Request().ParseForm()
	pluginName := c.Param("plugin_name")
	targetUID := c.Param("uid")

	// plugin
	_pluginInfo, ok := _plugin.PluginList[pluginName]
	if !ok {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "插件不存在", false, "tbsign"))
	}

	if _pluginInfo.(_plugin.PluginHooks).GetDBInfo().Ver == "-1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, "插件尚未安装", false, "tbsign"))
	}

	// target uid
	if targetUID == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	numTargetUID, err := strconv.ParseInt(targetUID, 10, 64)
	if err != nil || numTargetUID <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}

	// exists?
	var accountInfo model.TcUser
	err = _function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", targetUID).Take(&accountInfo).Error
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) || accountInfo.ID == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "用户不存在", false, "tbsign"))
	}
	if accountInfo.Role == "admin" && uid != "1" && uid != targetUID {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "只有根管理员允许改变其他管理员状态", false, "tbsign"))
	}

	err = _pluginInfo.Reset(int32(numTargetUID), 0, 0)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, err.Error(), false, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
	}
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

	messageType := c.QueryParam("type")
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
