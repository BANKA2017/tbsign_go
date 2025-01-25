package _api

import (
	"encoding/base64"
	"log"
	"net/http"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

type tokenResponse struct {
	Type     string `json:"type"`
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
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
	_function.GormDB.R.Model(&model.TcUser{}).Where("id = ?", uid).Take(&accountInfo)

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

	HttpAuthRefreshTokenMap.Delete(int(numUID))

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

	token, expireAt := tokenBuilder(int(accountInfo[0].ID))

	var resp = tokenResponse{
		Type:     "session",
		Token:    token,
		ExpireAt: expireAt,
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func Logout(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	HttpAuthRefreshTokenMap.Delete(int(numUID))

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

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	token, expireAt := tokenBuilder(int(numUID))

	var resp = tokenResponse{
		Type:     "session",
		Token:    token,
		ExpireAt: expireAt,
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
	resp.SystemSettings["allow_export_personal_data"] = _function.GetOption("go_export_personal_data")
	resp.SystemSettings["allow_import_personal_data"] = _function.GetOption("go_import_personal_data")

	if !share.EnableBackup {
		resp.SystemSettings["allow_export_personal_data"] = "0"
		resp.SystemSettings["allow_import_personal_data"] = "0"
	}

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

	settings := make(map[string]string, len(accountSettings))

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
	_account := strings.TrimSpace(c.FormValue("account"))
	verifyCode := strings.TrimSpace(c.FormValue("code"))
	newPwd := strings.TrimSpace(c.FormValue("password"))

	// pushType := strings.TrimSpace(c.QueryParams().Get("push_type"))

	resMessage := map[string]string{
		"verify_emoji": "",
	}

	// find account
	var accountInfo model.TcUser
	_function.GormDB.R.Where("name = ? OR email = ?", _account, _account).Find(&accountInfo)
	if accountInfo.ID == 0 && verifyCode != "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
	} else if accountInfo.ID == 0 {
		// defense scan
		// TODO Implement a delay of several seconds to prevent a side-channel attack.
		resMessage["verify_emoji"] = _function.RandomEmoji()
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
	}

	if verifyCode != "" {
		if len(verifyCode) != resetPasswordVerifyCodeLength {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
		}
		_v, ok := _function.VerifyCodeList.LoadCode("reset_password", accountInfo.ID)

		if !ok || _v == nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
		}

		if _v.TryTime <= _function.ResetPwdMaxTimes {
			_v.TryTime += 1
			_function.VerifyCodeList.StoreCode("reset_password", accountInfo.ID, _v)

			if _v.Value == verifyCode {
				if newPwd == "" {
					return c.JSON(http.StatusOK, _function.ApiTemplate(404, "密码不能为空", resMessage, "tbsign"))
				} else {
					// create new password
					hash, err := _function.CreatePasswordHash(newPwd)
					if err != nil {
						return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法更新密码...", resMessage, "tbsign"))
					}

					_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", accountInfo.ID).Update("pw", string(hash))

					_function.VerifyCodeList.DeleteCode("reset_password", accountInfo.ID)
					HttpAuthRefreshTokenMap.Delete(int(accountInfo.ID))
					return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
				}
			} else {
				return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
			}
		} else {
			return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
		}
	} else if verifyCode == "" && newPwd != "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效验证码", resMessage, "tbsign"))
	} else {
		VerifyCode, err := SendResetMessage(accountInfo.ID, "", false)
		if err != nil {
			switch err.Error() {
			case "已超过最大验证次数，请稍后再试":
				return c.JSON(http.StatusOK, _function.ApiTemplate(429, "已超过最大验证次数，请稍后再试", resMessage, "tbsign"))
			case "消息发送失败":
				return c.JSON(http.StatusOK, _function.ApiTemplate(500, "消息发送失败", resMessage, "tbsign"))
			case "验证码生成失败，请重试":
				return c.JSON(http.StatusOK, _function.ApiTemplate(500, "验证码生成失败，请重试", resMessage, "tbsign"))
			}
		}
		resMessage["verify_emoji"] = VerifyCode
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resMessage, "tbsign"))
	}
}

type TcBackupExportStructTcTieba struct {
	*model.TcTieba
	No bool `json:"no"`
}

func ExportAccountData(c echo.Context) error {
	uid := c.Get("uid").(string)

	// isPureGoMode
	if _function.GetOption("go_ver") != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "运行在兼容模式下的云签数据不允许导出", _function.EchoEmptyObject, "tbsign"))
	}

	// allowed?
	if _function.GetOption("go_export_personal_data") != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "站点管理员已关闭数据导出功能", _function.EchoEmptyObject, "tbsign"))
	}

	password := c.FormValue("password")

	var tcUser []*model.TcUser
	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Find(&tcUser)
	if len(tcUser) > 0 {
		err := _function.VerifyPasswordHash(tcUser[0].Pw, password)
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}

	var tcTieba []*model.TcTieba
	var tcBaiduid []*model.TcBaiduid

	// TODO plugin data export
	// var tcUsersOption []*model.TcUsersOption
	// var tcVer4BanList []*model.TcVer4BanList
	// var tcVer4RankLog []*model.TcVer4RankLog
	// var tcKdGrowth []*model.TcKdGrowth

	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ?", uid).Find(&tcTieba)
	_function.GormDB.W.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Find(&tcBaiduid)

	// _function.GormDB.W.Model(&model.TcUsersOption{}).Where("uid = ?", uid).Find(&tcUsersOption)
	// _function.GormDB.W.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Find(&tcVer4BanList)
	// _function.GormDB.W.Model(&model.TcVer4RankLog{}).Where("uid = ?", uid).Find(&tcVer4RankLog)
	// _function.GormDB.W.Model(&model.TcKdGrowth{}).Where("uid = ?", uid).Find(&tcKdGrowth)

	if len(share.DataEncryptKeyByte) > 0 {
		for _, tcBaiduidItem := range tcBaiduid {
			decryptedBDUSS, _ := _function.AES256GCMDecrypt(tcBaiduidItem.Bduss, share.DataEncryptKeyByte)
			tcBaiduidItem.Bduss = string(decryptedBDUSS)

			decryptedStoken, _ := _function.AES256GCMDecrypt(tcBaiduidItem.Stoken, share.DataEncryptKeyByte)
			tcBaiduidItem.Stoken = string(decryptedStoken)
		}
	}

	var tcTiebaForExport []*TcBackupExportStructTcTieba
	for _, tcTiebaItem := range tcTieba {
		tcTiebaForExport = append(tcTiebaForExport, &TcBackupExportStructTcTieba{
			tcTiebaItem,
			_function.TinyIntToBool(tcTiebaItem.No),
		})
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"tc_tieba":   tcTiebaForExport,
		"tc_baiduid": tcBaiduid,
		// "tc_users_option":  tcUsersOption,
		// "tc_ver4_ban_list": tcVer4BanList,
		// "tc_ver4_bank_log": tcVer4RankLog,
		// "tc_kd_growth":     tcKdGrowth,
	}, "tbsign"))
}

type TcBackupUploadStructTcBaiduid struct {
	Label    int    `json:"label"`
	Bduss    string `json:"bduss"`
	Stoken   string `json:"stoken"`
	Name     string `json:"name"`
	Portrait string `json:"portrait"`
}

type TcBackupUploadStructTcTieba struct {
	Label     int    `json:"label"`
	Fid       int    `json:"fid"`
	Tieba     string `json:"tieba"`
	No        bool   `json:"no"`
	Status    int    `json:"status"`
	Latest    int    `json:"latest"`
	LastError string `json:"last_error"`
}

type TcBackupUploadStruct struct {
	TcBaiduid []TcBackupUploadStructTcBaiduid `json:"tc_baiduid,omitempty"`
	TcTieba   []TcBackupUploadStructTcTieba   `json:"tc_tieba,omitempty"`
}

func ImportAccountData(c echo.Context) error {
	uid := c.Get("uid").(string)

	// isPureGoMode
	if _function.GetOption("go_ver") != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "运行在兼容模式下的云签数据不允许导入", _function.EchoEmptyObject, "tbsign"))
	}

	// allowed?
	if _function.GetOption("go_import_personal_data") != "1" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "站点管理员已关闭数据导入功能", _function.EchoEmptyObject, "tbsign"))
	}

	password := strings.TrimSpace(c.FormValue("password"))

	var tcUser []*model.TcUser
	_function.GormDB.W.Model(&model.TcUser{}).Where("id = ?", uid).Find(&tcUser)
	if len(tcUser) > 0 {
		err := _function.VerifyPasswordHash(tcUser[0].Pw, password)
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "密码错误", _function.EchoEmptyObject, "tbsign"))
		}
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}

	backupData := strings.TrimSpace(c.FormValue("data"))

	decodedData := new(TcBackupUploadStruct)
	err := _function.JsonDecode([]byte(backupData), decodedData)
	if err != nil {
		log.Println("decode-backup-data-err", err, uid)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "备份数据读取失败", _function.EchoEmptyObject, "tbsign"))
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	var tcTieba []*model.TcTieba
	var tcBaiduid []*model.TcBaiduid
	_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ?", uid).Find(&tcTieba)
	_function.GormDB.W.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Find(&tcBaiduid)

	var labelPidKV = make(map[int]int32)
	var newTcBaiduID []model.TcBaiduid
	var newTcBaiduPortrait []string
	var newTcTieba []model.TcTieba
	var newTcTiebaWithoutAccount []TcBackupUploadStructTcTieba

	// newAccount
	for _, importBaiduidItem := range decodedData.TcBaiduid {
		exists := false
		for _, localTcBaiduidItem := range tcBaiduid {
			if importBaiduidItem.Portrait == localTcBaiduidItem.Portrait {
				if _, ok := labelPidKV[importBaiduidItem.Label]; !ok {
					labelPidKV[importBaiduidItem.Label] = localTcBaiduidItem.ID
				}
				exists = true
				break
			}
		}
		if !exists {
			if len(share.DataEncryptKeyByte) > 0 {
				encryptedBDUSS, _ := _function.AES256GCMEncrypt(importBaiduidItem.Bduss, share.DataEncryptKeyByte)
				importBaiduidItem.Bduss = strings.ReplaceAll(base64.URLEncoding.EncodeToString(encryptedBDUSS), "=", "")

				encryptedStoken, _ := _function.AES256GCMEncrypt(importBaiduidItem.Stoken, share.DataEncryptKeyByte)
				importBaiduidItem.Stoken = strings.ReplaceAll(base64.URLEncoding.EncodeToString(encryptedStoken), "=", "")
			}

			newTcBaiduID = append(newTcBaiduID, model.TcBaiduid{
				UID:      int32(numUID),
				Bduss:    importBaiduidItem.Bduss,
				Stoken:   importBaiduidItem.Stoken,
				Name:     importBaiduidItem.Name,
				Portrait: importBaiduidItem.Portrait,
			})
			newTcBaiduPortrait = append(newTcBaiduPortrait, importBaiduidItem.Portrait)
		}
	}

	for _, importTiebaItem := range decodedData.TcTieba {
		if pid, ok := labelPidKV[importTiebaItem.Label]; ok {
			exists := false
			for _, localTcTiebaItem := range tcTieba {
				if localTcTiebaItem.Pid == pid && localTcTiebaItem.Fid == int32(importTiebaItem.Fid) {
					exists = true
					break
				}
			}
			if !exists {
				newTcTieba = append(newTcTieba, model.TcTieba{
					UID:       int32(numUID),
					Pid:       pid,
					Fid:       int32(importTiebaItem.Fid),
					Tieba:     importTiebaItem.Tieba,
					No:        _function.BoolToTinyInt(importTiebaItem.No),
					Status:    int32(importTiebaItem.Status),
					Latest:    int32(importTiebaItem.Latest),
					LastError: importTiebaItem.LastError,
				})
			}
		} else {
			newTcTiebaWithoutAccount = append(newTcTiebaWithoutAccount, importTiebaItem)
		}
	}

	err = _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		var err error

		if len(newTcBaiduID) > 0 {
			if err = tx.Create(&newTcBaiduID).Error; err != nil {
				return err
			}
		}

		if len(newTcTieba) > 0 {
			if err = tx.Create(&newTcTieba).Error; err != nil {
				return err
			}
		}

		return err
	})

	if err != nil {
		log.Println("create-backup-data-err", err, uid)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "备份数据导入失败", _function.EchoEmptyObject, "tbsign"))
	}

	if len(newTcTiebaWithoutAccount) > 0 {
		_function.GormDB.W.Model(&model.TcBaiduid{}).Where("uid = ? AND portrait IN (?)", uid, newTcBaiduPortrait).Find(&newTcBaiduID)

		for _, importBaiduidItem := range decodedData.TcBaiduid {
			for _, localTcBaiduidItem := range newTcBaiduID {
				if importBaiduidItem.Portrait == localTcBaiduidItem.Portrait {
					if _, ok := labelPidKV[importBaiduidItem.Label]; !ok {
						labelPidKV[importBaiduidItem.Label] = localTcBaiduidItem.ID
					}
					break
				}
			}
		}

		var newTcTiebaWithoutAccountToInsert []model.TcTieba
		for _, importTiebaItem := range newTcTiebaWithoutAccount {
			if pid, ok := labelPidKV[importTiebaItem.Label]; ok {
				newTcTiebaWithoutAccountToInsert = append(newTcTiebaWithoutAccountToInsert, model.TcTieba{
					UID:       int32(numUID),
					Pid:       pid,
					Fid:       int32(importTiebaItem.Fid),
					Tieba:     importTiebaItem.Tieba,
					No:        _function.BoolToTinyInt(importTiebaItem.No),
					Status:    int32(importTiebaItem.Status),
					Latest:    int32(importTiebaItem.Latest),
					LastError: importTiebaItem.LastError,
				})
			}
		}

		if len(newTcTiebaWithoutAccountToInsert) > 0 {
			if err := _function.GormDB.W.Create(&newTcTiebaWithoutAccountToInsert).Error; err != nil {
				log.Println("create-backup-data-err2", err, uid)
				return c.JSON(http.StatusOK, _function.ApiTemplate(500, "部分贴吧列表导入失败", _function.EchoEmptyObject, "tbsign"))
			}
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]int{
		"tc_tieba":   len(decodedData.TcTieba) - len(tcTieba),
		"tc_baiduid": len(decodedData.TcBaiduid) - len(tcBaiduid),
	}, "tbsign"))
}
