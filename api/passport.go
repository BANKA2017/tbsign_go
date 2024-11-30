package _api

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type tokenResponse struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type userInfoStruct struct {
	UID    int32  `json:"uid"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar string `json:"avatar"`
	Role   string `json:"role"`

	// push
	NtfyTopic string `json:"ntfy_topic"`
	BarkKey   string `json:"bark_key"`
	PushType  string `json:"push_type"`
}
type userInfoWithSettingsStruct struct {
	userInfoStruct

	SystemSettings map[string]string `json:"system_settings"`
}

func Signup(c echo.Context) error {
	// site status
	isRegistrationEnable := _function.GetOption("enable_reg") == "1"
	if !isRegistrationEnable {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "注册已关闭", _function.EchoEmptyObject, "tbsign"))
	}

	// form
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	inviteCode := c.FormValue("invite_code")

	if name == "" || strings.Contains(name, "@") || !_function.VerifyEmail(email) || password == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 用户名/邮箱/密码", _function.EchoEmptyObject, "tbsign"))
	}

	// invite code
	localInviteCode := _function.GetOption("yr_reg")
	if localInviteCode != "" {
		if localInviteCode != inviteCode {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效邀请码", _function.EchoEmptyObject, "tbsign"))
		}
	}

	role := "user"

	// pre check
	var emailOrNameExistsCount int64
	_function.GormDB.R.Model(&model.TcUser{}).Where("email = ? OR name = ?", email, name).Count(&emailOrNameExistsCount)
	if emailOrNameExistsCount > 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "用户名或邮箱已注册", _function.EchoEmptyObject, "tbsign"))
	}

	passwordHash, err := _function.CreatePasswordHash(password)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法建立账号", _function.EchoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Create(&model.TcUser{
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  role,
		T:     "tieba",
	})

	msg := "注册成功🎉"

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]string{
		"name": name,
		"role": role,
		"msg":  msg,
	}, "tbsign"))
}

func DeleteAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	password := c.FormValue("password")
	if password == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效密码", _function.EchoEmptyObject, "tbsign"))
	}

	var accountInfo model.TcUser
	_function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", uid).First(&accountInfo)

	// verify password
	err := _function.VerifyPasswordHash(accountInfo.Pw, password)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效密码", _function.EchoEmptyObject, "tbsign"))
	}

	// find root admin
	if uid == "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "您不能删除账号，因为您是根管理员", _function.EchoEmptyObject, "tbsign"))
	}

	// set role -> deleted

	// plugins
	numUID, _ := strconv.ParseInt(uid, 10, 64)

	err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		var err error
		if err = _plugin.DeleteAccount("uid", int32(numUID), tx); err != nil {
			return err
		}

		if err = tx.Where("id = ?", uid).Delete(&model.TcUser{}).Error; err != nil {
			return err
		}
		if err = tx.Where("uid = ?", uid).Delete(&model.TcTieba{}).Error; err != nil {
			return err
		}
		if err = tx.Where("uid = ?", uid).Delete(&model.TcBaiduid{}).Error; err != nil {
			return err
		}
		if err = tx.Where("uid = ?", uid).Delete(&model.TcUsersOption{}).Error; err != nil {
			return err
		}

		return err

	})
	if err != nil {
		log.Println("del-account", uid, err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "账号删除失败", _function.EchoEmptyObject, "tbsign"))
	}

	keyBucket.Delete(uid)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "账号已删除，感谢您的使用", map[string]any{
		"uid":  int64(accountInfo.ID),
		"name": accountInfo.Name,
		"role": accountInfo.Role,
	}, "tbsign"))
}

func Login(c echo.Context) error {
	account := strings.TrimSpace(c.FormValue("account")) // username or email
	password := strings.TrimSpace(c.FormValue("password"))

	if account == "" || password == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(401, "账号或密码错误", _function.EchoEmptyObject, "tbsign"))
	}

	// check
	var accountInfo []*model.TcUser
	_function.GormDB.R.Where("name = ? OR email = ?", account, account).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(401, "账号或密码错误", _function.EchoEmptyObject, "tbsign"))
	}

	err := _function.VerifyPasswordHash(accountInfo[0].Pw, password)
	if err != nil && _function.GetOption("go_ver") != "1" {
		// Compatible with older versions -> md5(md5(md5($pwd)))
		if _function.Md5(_function.Md5(_function.Md5(password))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, _function.ApiTemplate(401, "账号或密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(401, "账号或密码错误", _function.EchoEmptyObject, "tbsign"))
	}

	if accountInfo[0].Role == "banned" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号已封禁", _function.EchoEmptyObject, "tbsign"))
	} else if accountInfo[0].Role == "deleted" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号已删除", _function.EchoEmptyObject, "tbsign"))
	}

	var resp = tokenResponse{
		Type:  "bearer",
		Token: bearerTokenBuilder(strconv.Itoa(int(accountInfo[0].ID)), true),
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func Logout(c echo.Context) error {
	uid := c.Get("uid").(string)
	keyBucket.Delete(uid)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func UpdateAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accountInfo []*model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}

	username := strings.TrimSpace(c.FormValue("username"))
	email := strings.TrimSpace(c.FormValue("email"))
	barkKey := strings.TrimSpace(c.FormValue("bark_key"))
	ntfyTopic := strings.TrimSpace(c.FormValue("ntfy_topic"))
	pushType := strings.TrimSpace(c.FormValue("push_type"))

	password := strings.TrimSpace(c.FormValue("password"))

	// compare old password
	err := _function.VerifyPasswordHash(accountInfo[0].Pw, password)
	if err != nil && _function.GetOption("go_ver") != "1" {
		// Compatible with older versions
		if _function.Md5(_function.Md5(_function.Md5(password))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "密码错误", _function.EchoEmptyObject, "tbsign"))
	}

	// TODO use transaction
	// email
	if email != "" {
		if !_function.VerifyEmail(email) {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "邮箱不合法", false, "tbsign"))
		}

		// compare email
		if email != accountInfo[0].Email {
			var emailExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("email = ?", email).Count(&emailExistsCount)

			if emailExistsCount > 0 {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "邮箱已存在", _function.EchoEmptyObject, "tbsign"))
			} else {
				_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Update("email", email)
			}
		}
	} else {
		email = accountInfo[0].Email
	}

	// username
	if username != "" {
		// compare username
		if username != accountInfo[0].Name {
			var usernameExistsCount int64
			_function.GormDB.R.Model(&model.TcUser{}).Where("name = ?", username).Count(&usernameExistsCount)

			if usernameExistsCount > 0 {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "用户名已存在", _function.EchoEmptyObject, "tbsign"))
			} else {
				_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Update("name", username)
			}
		}
	} else {
		username = accountInfo[0].Name
	}

	// push
	localPushNtfyTopic := _function.GetUserOption("go_ntfy_topic", uid)
	if localPushNtfyTopic != ntfyTopic {
		_function.SetUserOption("go_ntfy_topic", ntfyTopic, uid)
		localPushNtfyTopic = ntfyTopic
	}
	localPushBarkKey := _function.GetUserOption("go_bark_key", uid)
	if localPushBarkKey != barkKey {
		_function.SetUserOption("go_bark_key", barkKey, uid)
		localPushBarkKey = barkKey
	}
	localPushType := _function.GetUserOption("go_message_type", uid)
	if localPushType != pushType && slices.Contains(_function.MessageTypeList, pushType) {
		_function.SetUserOption("go_message_type", pushType, uid)
		localPushType = pushType
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	var resp = userInfoStruct{
		UID:       int32(numUID),
		Name:      username,
		Email:     email,
		Avatar:    _function.GetGravatarLink(email),
		NtfyTopic: localPushNtfyTopic,
		BarkKey:   localPushBarkKey,
		PushType:  localPushType,
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func UpdatePassword(c echo.Context) error {
	uid := c.Get("uid").(string)

	oldPwd := c.FormValue("old_password")
	newPwd := c.FormValue("new_password")

	if oldPwd == "" || newPwd == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "新/旧密码都不可为空", _function.EchoEmptyObject, "tbsign"))
	}

	var accountInfo []*model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}

	// compare old password
	err := _function.VerifyPasswordHash(accountInfo[0].Pw, oldPwd)
	if err != nil && _function.GetOption("go_ver") != "1" {
		// Compatible with older versions
		if _function.Md5(_function.Md5(_function.Md5(oldPwd))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "旧密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(401, "账号或密码错误", _function.EchoEmptyObject, "tbsign"))
	}

	// create new password
	hash, err := _function.CreatePasswordHash(newPwd)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法更新密码...", _function.EchoEmptyObject, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Update("pw", string(hash))

	var resp = tokenResponse{
		Type:  "bearer",
		Token: bearerTokenBuilder(uid, true),
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func GetAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	// check filter

	var accountInfo []*model.TcUser
	_function.GormDB.R.Where("id = ?", uid).Find(&accountInfo)

	// var accountSettings []*model.TcUsersOption
	// _function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	if len(accountInfo) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}

	var resp = userInfoWithSettingsStruct{
		userInfoStruct{
			UID:    accountInfo[0].ID,
			Name:   accountInfo[0].Name,
			Email:  accountInfo[0].Email,
			Avatar: _function.GetGravatarLink(accountInfo[0].Email),
			Role:   accountInfo[0].Role,

			NtfyTopic: _function.GetUserOption("go_ntfy_topic", uid),
			BarkKey:   _function.GetUserOption("go_bark_key", uid),
			PushType:  _function.GetUserOption("go_message_type", uid),
		},
		make(map[string]string),
	}
	resp.SystemSettings["forum_sync_policy"] = _function.GetOption("go_forum_sync_policy")
	resp.SystemSettings["bark_addr"] = _function.GetOption("go_bark_addr")
	resp.SystemSettings["ntfy_addr"] = _function.GetOption("go_ntfy_addr")

	if resp.PushType == "" {
		_function.SetUserOption("go_message_type", "email", uid)
		resp.PushType = "email"
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func GetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accountSettings []*model.TcUsersOption
	_function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	settings := make(map[string]string)

	for _, v := range accountSettings {
		settings[v.Name] = v.Value
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", settings, "tbsign"))
}

// TODO verify password
func UpdateSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	c.Request().ParseForm()

	var accountSettings []*model.TcUsersOption
	_function.GormDB.R.Where("uid = ?", uid).Find(&accountSettings)

	var newSettings []*model.TcUsersOption

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

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", settings, "tbsign"))
}

func ResetPassword(c echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	verifyCode := strings.TrimSpace(c.FormValue("code"))
	newPwd := strings.TrimSpace(c.FormValue("password"))

	pushType := strings.TrimSpace(c.QueryParams().Get("push_type"))

	resMessage := map[string]string{
		"verify_emoji": "",
	}

	if !_function.VerifyEmail(email) {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "邮箱不合法", resMessage, "tbsign"))
	}

	// find account
	var accountInfo model.TcUser
	_function.GormDB.R.Where("email = ?", email).Find(&accountInfo)
	if accountInfo.ID == 0 {
		// defense scan
		// TODO Implement a delay of several seconds to prevent a side-channel attack.
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
	}

	if verifyCode != "" {
		_v, ok := _function.ResetPwdList.Load(accountInfo.ID)

		if !ok || _v == nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
		}
		if __v, ok := _v.(*_function.ResetPwdStruct); ok && __v.Value == verifyCode {
			if newPwd == "" {
				return c.JSON(http.StatusOK, _function.ApiTemplate(404, "密码不能为空", resMessage, "tbsign"))
			} else {
				// create new password
				hash, err := _function.CreatePasswordHash(newPwd)
				if err != nil {
					return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法更新密码...", resMessage, "tbsign"))
				}

				_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", accountInfo.ID).Update("pw", string(hash))

				_function.ResetPwdList.Delete(accountInfo.ID)
				keyBucket.Delete(strconv.Itoa(int(accountInfo.ID)))
				return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
			}
		} else {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
		}
	} else if verifyCode == "" && newPwd != "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
	} else {
		_v, ok := _function.ResetPwdList.Load(accountInfo.ID)
		v := new(_function.ResetPwdStruct)

		if !ok || _v == nil {
			v = _function.VariablePtrWrapper(_function.ResetPwdStruct{
				Expire: _function.Now.Unix() + _function.ResetPwdExpire,
			})
		} else {
			v = _v.(*_function.ResetPwdStruct)
			if v.Time >= _function.ResetPwdMaxTimes {
				return c.JSON(http.StatusOK, _function.ApiTemplate(429, "已超过最大验证次数，请稍后再试", resMessage, "tbsign"))
			}
		}
		// init a callback code
		code := strconv.Itoa(int(rand.Uint32()))
		for len(code) < 6 {
			code = "0" + code
		}

		code = code[0:6]

		v.Value = code
		v.Time += 1
		v.VerifyCode = _function.RandomEmoji()

		_function.ResetPwdList.Store(accountInfo.ID, v)

		mailObject := _function.PushMessageTemplateResetPassword(v.VerifyCode, code)

		// user default message type
		userMessageType := "email"
		if pushType != "" && slices.Contains(_function.MessageTypeList, pushType) {
			userMessageType = pushType
		} else {
			localPushType := _function.GetUserOption("go_message_type", strconv.Itoa(int(accountInfo.ID)))
			if slices.Contains(_function.MessageTypeList, localPushType) {
				userMessageType = localPushType
			}
		}

		err := _function.SendMessage(userMessageType, accountInfo.ID, mailObject.Title, mailObject.Body)
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, "消息发送失败", resMessage, "tbsign"))
		} else {
			resMessage["verify_emoji"] = v.VerifyCode
			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
		}
	}
}

func ExportAccountData(c echo.Context) error {
	uid := c.Get("uid").(string)

	password := c.FormValue("password")

	var tcUser []*model.TcUser
	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Find(&tcUser)
	if len(tcUser) > 0 {
		err := _function.VerifyPasswordHash(tcUser[0].Pw, password)
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在 (请问是如何登录的)", _function.EchoEmptyObject, "tbsign"))
	}
	oneTcUser := tcUser[0]
	oneTcUser.Pw = ""
	oneTcUser.T = ""

	var tcTieba []*model.TcTieba
	var tcBaiduid []*model.TcBaiduid
	var tcUsersOption []*model.TcUsersOption
	var tcVer4BanList []*model.TcVer4BanList
	var tcVer4RankLog []*model.TcVer4RankLog
	var tcKdGrowth []*model.TcKdGrowth

	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ?", uid).Find(&tcTieba)
	_function.GormDB.W.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Find(&tcBaiduid)
	_function.GormDB.W.Model(&model.TcUsersOption{}).Where("uid = ?", uid).Find(&tcUsersOption)

	_function.GormDB.W.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Find(&tcVer4BanList)
	_function.GormDB.W.Model(&model.TcVer4RankLog{}).Where("uid = ?", uid).Find(&tcVer4RankLog)
	_function.GormDB.W.Model(&model.TcKdGrowth{}).Where("uid = ?", uid).Find(&tcKdGrowth)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"tc_user":          oneTcUser,
		"tc_tieba":         tcTieba,
		"tc_baiduid":       tcBaiduid,
		"tc_users_option":  tcUsersOption,
		"tc_ver4_ban_list": tcVer4BanList,
		"tc_ver4_bank_log": tcVer4RankLog,
		"tc_kd_growth":     tcKdGrowth,
	}, "tbsign"))
}
