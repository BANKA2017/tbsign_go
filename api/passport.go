package _api

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

type tokenResponse struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

func Signup(c echo.Context) error {
	// site status
	isRegistrationEnable := _function.GetOption("enable_reg") == "1"
	if !isRegistrationEnable {
		return c.JSON(http.StatusOK, apiTemplate(403, "注册已关闭", echoEmptyObject, "tbsign"))
	}

	// form
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	inviteCode := c.FormValue("invite_code")

	if name == "" || strings.Contains(name, "@") || !_function.VerifyEmail(email) || password == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 用户名/邮箱/密码", echoEmptyObject, "tbsign"))
	}

	// invite code
	localInviteCode := _function.GetOption("yr_reg")
	if localInviteCode != "" {
		if localInviteCode != inviteCode {
			return c.JSON(http.StatusOK, apiTemplate(403, "无效邀请码", echoEmptyObject, "tbsign"))
		}
	}

	role := "user"

	// pre check
	var emailOrNameExistsCount int64
	_function.GormDB.R.Model(&model.TcUser{}).Where("email = ? OR name = ?", email, name).Count(&emailOrNameExistsCount)
	if emailOrNameExistsCount > 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "用户名或邮箱已注册", echoEmptyObject, "tbsign"))
	}

	passwordHash, err := _function.CreatePasswordHash(password)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "无法建立帐号", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Create(&model.TcUser{
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  role,
		T:     "tieba",
	})

	msg := "注册成功🎉"

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]string{
		"name": name,
		"role": role,
		"msg":  msg,
	}, "tbsign"))
}

func DeleteAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	password := c.FormValue("password")
	if password == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效密码", echoEmptyObject, "tbsign"))
	}

	var accountInfo model.TcUser
	_function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", uid).First(&accountInfo)

	// verify password
	err := _function.VerifyPasswordHash(accountInfo.Pw, password)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效密码", echoEmptyObject, "tbsign"))
	}

	// find root admin
	if uid == "1" {
		return c.JSON(http.StatusOK, apiTemplate(403, "您不能删除账号，因为您是根管理员", echoEmptyObject, "tbsign"))
	}

	// set role -> delete
	_function.GormDB.W.Where("id = ?", uid).Delete(&model.TcUser{})
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcTieba{})
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcBaiduid{})
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcUsersOption{})

	// plugins
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcVer4BanList{})
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcVer4RankLog{})
	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcKdGrowth{})

	delete(keyBucket, uid)

	return c.JSON(http.StatusOK, apiTemplate(200, "帐号已删除，感谢您的使用", map[string]any{
		"uid":  int64(accountInfo.ID),
		"name": accountInfo.Name,
		"role": accountInfo.Role,
	}, "tbsign"))
}

func Login(c echo.Context) error {
	account := strings.TrimSpace(c.FormValue("account")) // username or email
	password := strings.TrimSpace(c.FormValue("password"))

	if account == "" || password == "" {
		return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
	}

	// check
	var accountInfo []model.TcUser
	_function.GormDB.R.Where("name = ? OR email = ?", account, account).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
	}

	err := _function.VerifyPasswordHash(accountInfo[0].Pw, password)
	if err != nil && _function.GetOption("go_ver") != "1" {
		// Compatible with older versions -> md5(md5(md5($pwd)))
		if _function.Md5(_function.Md5(_function.Md5(password))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
		}
	} else if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
	}

	if accountInfo[0].Role == "banned" {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号已封禁", echoEmptyObject, "tbsign"))
	} else if accountInfo[0].Role == "deleted" {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号已删除", echoEmptyObject, "tbsign"))
	}

	var resp = tokenResponse{
		Type:  "bearer",
		Token: bearerTokenBuilder(strconv.Itoa(int(accountInfo[0].ID)), true),
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func Logout(c echo.Context) error {
	uid := c.Get("uid").(string)
	delete(keyBucket, uid)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
}

func UpdateAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	username := c.FormValue("username")
	email := c.FormValue("email")

	var accountInfo []model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号不存在", echoEmptyObject, "tbsign"))
	}

	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	if email != "" {
		if !_function.VerifyEmail(email) {
			return c.JSON(http.StatusOK, apiTemplate(404, "邮箱不合法", false, "tbsign"))
		}

		// compare email
		if email != accountInfo[0].Email {
			var emailExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("email = ?", email).Count(&emailExistsCount)

			if emailExistsCount > 0 {
				return c.JSON(http.StatusOK, apiTemplate(403, "邮箱已存在", echoEmptyObject, "tbsign"))
			} else {
				_function.GormDB.W.Model(model.TcUser{}).Where("id = ?", uid).Update("email", email)
			}
		}
	} else {
		email = accountInfo[0].Email
	}

	if username != "" {
		// compare username
		if username != accountInfo[0].Name {
			var usernameExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("name = ?", username).Count(&usernameExistsCount)

			if usernameExistsCount > 0 {
				return c.JSON(http.StatusOK, apiTemplate(403, "用户名已存在", echoEmptyObject, "tbsign"))
			} else {
				_function.GormDB.W.Model(model.TcUser{}).Where("id = ?", uid).Update("name", username)
			}
		}
	} else {
		username = accountInfo[0].Name
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	var resp = struct {
		UID   int32  `json:"uid"`
		Name  string `json:"username"`
		Email string `json:"email"`
	}{
		UID:   int32(numUID),
		Name:  username,
		Email: email,
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func UpdatePassword(c echo.Context) error {
	uid := c.Get("uid").(string)

	oldPwd := c.FormValue("old_password")
	newPwd := c.FormValue("new_password")

	if oldPwd == "" || newPwd == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "新/旧密码都不可为空", echoEmptyObject, "tbsign"))
	}

	var accountInfo []model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号不存在", echoEmptyObject, "tbsign"))
	}

	// compare old password
	err := _function.VerifyPasswordHash(accountInfo[0].Pw, oldPwd)
	if err != nil && _function.GetOption("go_ver") != "1" {
		// Compatible with older versions
		if _function.Md5(_function.Md5(_function.Md5(oldPwd))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, apiTemplate(403, "旧密码错误", echoEmptyObject, "tbsign"))
		}
	} else if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
	}

	// create new password
	hash, err := _function.CreatePasswordHash(newPwd)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "无法更新密码...", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Model(model.TcUser{}).Where("id = ?", uid).Update("pw", string(hash))

	var resp = tokenResponse{
		Type:  "bearer",
		Token: bearerTokenBuilder(uid, true),
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func GetAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	// check filter

	var accountInfo []model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Find(&accountInfo)

	var accountSettings []model.TcUsersOption
	_function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号不存在", echoEmptyObject, "tbsign"))
	}

	var resp = struct {
		UID            int32             `json:"uid"`
		Name           string            `json:"name"`
		Email          string            `json:"email"`
		Role           string            `json:"role"`
		Settings       map[string]string `json:"settings"`
		SystemSettings map[string]string `json:"system_settings"`
	}{
		UID:            accountInfo[0].ID,
		Name:           accountInfo[0].Name,
		Email:          accountInfo[0].Email,
		Role:           accountInfo[0].Role,
		Settings:       make(map[string]string),
		SystemSettings: make(map[string]string),
	}

	for _, v := range accountSettings {
		resp.Settings[v.Name] = v.Value
	}
	resp.SystemSettings["forum_sync_policy"] = _function.GetOption("go_forum_sync_policy")

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func GetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accountSettings []model.TcUsersOption
	_function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	settings := make(map[string]string)

	for _, v := range accountSettings {
		settings[v.Name] = v.Value
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func UpdateSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	c.Request().ParseForm()

	var accountSettings []model.TcUsersOption
	_function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	var newSettings []model.TcUsersOption

	for _, v := range accountSettings {
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
			_function.SetUserOption(v.Name, v.Value, uid)
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}

func ResetPassword(c echo.Context) error {
	email := c.FormValue("email")
	verifyCode := c.FormValue("code")
	newPwd := c.FormValue("password")

	if !_function.VerifyEmail(email) {
		return c.JSON(http.StatusOK, apiTemplate(404, "邮箱不合法", false, "tbsign"))
	}

	// find account
	var accountInfo model.TcUser
	_function.GormDB.R.Where("email = ?", email).Find(&accountInfo)
	if accountInfo.ID == 0 {
		// defense scan
		// TODO Implement a delay of several seconds to prevent a side-channel attack.
		return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
	}

	if verifyCode != "" {
		if _, ok := _function.ResetPwdList[accountInfo.ID]; !ok {
			return c.JSON(http.StatusOK, apiTemplate(404, "无效验证码", false, "tbsign"))
		}
		if _function.ResetPwdList[accountInfo.ID].Value != verifyCode {
			return c.JSON(http.StatusOK, apiTemplate(404, "无效验证码", false, "tbsign"))
		} else {
			if newPwd == "" {
				return c.JSON(http.StatusOK, apiTemplate(404, "密码不能为空", false, "tbsign"))
			} else {
				// create new password
				hash, err := _function.CreatePasswordHash(newPwd)
				if err != nil {
					return c.JSON(http.StatusOK, apiTemplate(500, "无法更新密码...", false, "tbsign"))
				}

				_function.GormDB.W.Model(model.TcUser{}).Where("id = ?", accountInfo.ID).Update("pw", string(hash))

				delete(_function.ResetPwdList, accountInfo.ID)
				return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
			}
		}
	} else if verifyCode == "" && newPwd != "" {
		return c.JSON(http.StatusOK, apiTemplate(404, "无效验证码", false, "tbsign"))
	} else {
		if _, ok := _function.ResetPwdList[accountInfo.ID]; !ok {
			_function.ResetPwdList[accountInfo.ID] = &_function.ResetPwdStruct{
				Expire: _function.Now.Unix() + _function.ResetPwdExpire,
			}
		} else {
			if _function.ResetPwdList[accountInfo.ID].Time >= _function.ResetPwdMaxTimes {
				return c.JSON(http.StatusOK, apiTemplate(403, "已超过最大验证次数，请稍后再试", false, "tbsign"))
			}
		}
		// init a callback code
		code := strconv.Itoa(int(rand.Uint32()))
		for len(code) < 6 {
			code = "0" + code
		}

		code = code[0:6]

		_function.ResetPwdList[accountInfo.ID].Value = code
		_function.ResetPwdList[accountInfo.ID].Time += 1

		mailObject := _function.EmailTemplateResetPassword(accountInfo.Email, code)
		err := _function.SendEmail(accountInfo.Email, mailObject.Object, mailObject.Body)
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusOK, apiTemplate(500, "邮件发送失败", false, "tbsign"))
		} else {
			return c.JSON(http.StatusOK, apiTemplate(200, "OK", true, "tbsign"))
		}
	}
}

func ExportAccountData(c echo.Context) error {
	uid := c.Get("uid").(string)

	password := c.FormValue("password")

	var tcUser []model.TcUser
	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Find(&tcUser)
	if len(tcUser) > 0 {
		err := _function.VerifyPasswordHash(tcUser[0].Pw, password)
		if err != nil {
			return c.JSON(http.StatusOK, apiTemplate(403, "密码错误", echoEmptyObject, "tbsign"))
		}
	} else {
		return c.JSON(http.StatusOK, apiTemplate(403, "账号不存在 (请问是如何登录的)", echoEmptyObject, "tbsign"))
	}
	oneTcUser := tcUser[0]
	oneTcUser.Pw = ""
	oneTcUser.T = ""

	var tcTieba []model.TcTieba
	var tcBaiduid []model.TcBaiduid
	var tcUsersOption []model.TcUsersOption
	var tcVer4BanList []model.TcVer4BanList
	var tcVer4RankLog []model.TcVer4RankLog
	var tcKdGrowth []model.TcKdGrowth

	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ?", uid).Find(&tcTieba)
	_function.GormDB.W.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Find(&tcBaiduid)
	_function.GormDB.W.Model(&model.TcUsersOption{}).Where("uid = ?", uid).Find(&tcUsersOption)
	_function.GormDB.W.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Find(&tcVer4BanList)
	_function.GormDB.W.Model(&model.TcVer4RankLog{}).Where("uid = ?", uid).Find(&tcVer4RankLog)
	_function.GormDB.W.Model(&model.TcKdGrowth{}).Where("uid = ?", uid).Find(&tcKdGrowth)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", map[string]any{
		"tc_user":          oneTcUser,
		"tc_tieba":         tcTieba,
		"tc_baiduid":       tcBaiduid,
		"tc_users_option":  tcUsersOption,
		"tc_ver4_ban_list": tcVer4BanList,
		"tc_ver4_bank_log": tcVer4RankLog,
		"tc_kd_growth":     tcKdGrowth,
	}, "tbsign"))
}
