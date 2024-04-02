package _api

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

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
	//inviteCode := c.FormValue("invite_code")

	if name == "" || strings.Contains(name, "@") || !_function.VerifyEmail(email) || password == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效 用户名/邮箱/密码", echoEmptyObject, "tbsign"))
	}

	role := "user"

	// pre check
	var userCount int64
	_function.GormDB.Model(&model.TcUser{}).Count(&userCount)
	if userCount > 0 {
		var emailOrNameExistsCount int64
		_function.GormDB.Model(&model.TcUser{}).Where("email = ? OR name = ?", email, name).Count(&emailOrNameExistsCount)
		if emailOrNameExistsCount > 0 {
			return c.JSON(http.StatusOK, apiTemplate(403, "用户名或邮箱已注册", echoEmptyObject, "tbsign"))
		}
	} else {
		role = "admin"
	}

	passwordHash, err := _function.CreatePasswordHash(password)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "无法建立帐号", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.Create(&model.TcUser{
		Name:  name,
		Email: email,
		Pw:    string(passwordHash),
		Role:  role,
		T:     "tieba",
	})

	msg := "注册成功🎉"
	if userCount <= 0 {
		msg = "注册成功🎉，您是第一个用户，将被设置为管理员用户组"
	}

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
	_function.GormDB.Model(&model.TcUser{}).Where("id = ?", uid).First(&accountInfo)

	// verify password
	err := _function.VerifyPasswordHash(accountInfo.Pw, password)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "无效密码", echoEmptyObject, "tbsign"))
	}

	// find other admin
	var adminCount int64
	_function.GormDB.Model(&model.TcUser{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount <= 1 {
		return c.JSON(http.StatusOK, apiTemplate(403, "您不能删除此账号，因为您是本站唯一的管理员", echoEmptyObject, "tbsign"))
	}

	// set role -> delete
	_function.GormDB.Model(&model.TcUser{}).Delete("id = ?", uid)
	_function.GormDB.Model(&model.TcTieba{}).Delete("uid = ?", uid)
	_function.GormDB.Model(&model.TcBaiduid{}).Delete("uid = ?", uid)
	_function.GormDB.Model(&model.TcUsersOption{}).Delete("uid = ?", uid)

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
	_function.GormDB.Where("name = ? OR email = ?", account, account).Limit(1).Find(&accountInfo)

	err := _function.VerifyPasswordHash(accountInfo[0].Pw, password)
	if err != nil {
		// Compatible with older versions -> md5(md5(md5($pwd)))
		if _function.Md5(_function.Md5(_function.Md5(password))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, apiTemplate(401, "帐号或密码错误", echoEmptyObject, "tbsign"))
		}
	}

	if accountInfo[0].Role == "banned" {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号已封禁", echoEmptyObject, "tbsign"))
	} else if accountInfo[0].Role == "deleted" {
		return c.JSON(http.StatusOK, apiTemplate(403, "帐号已删除", echoEmptyObject, "tbsign"))
	}

	var resp = struct {
		Type  string `json:"type"`
		Token string `json:"token"` // <- static session
	}{
		Type:  "basic",
		Token: base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(int(accountInfo[0].ID)) + ":" + hex.EncodeToString(_function.GenHMAC256([]byte(accountInfo[0].Pw), []byte(strconv.Itoa(int(accountInfo[0].ID))+accountInfo[0].Pw))))),
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTemplate(200, "无效接口，清理本地缓存即可", echoEmptyObject, "tbsign"))
}

func UpdatePassword(c echo.Context) error {
	uid := c.Get("uid").(string)

	oldPwd := c.FormValue("old_password")
	newPwd := c.FormValue("new_password")

	var accountInfo []model.TcUser
	_function.GormDB.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	// compare old password
	err := _function.VerifyPasswordHash(accountInfo[0].Pw, oldPwd)
	if err != nil {
		// Compatible with older versions
		if _function.Md5(_function.Md5(_function.Md5(oldPwd))) != accountInfo[0].Pw {
			return c.JSON(http.StatusOK, apiTemplate(403, "旧密码错误", echoEmptyObject, "tbsign"))
		}
	}

	// create new password
	hash, err := _function.CreatePasswordHash(newPwd)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(500, "无法更新密码...", echoEmptyObject, "tbsign"))
	}

	_function.GormDB.Model(model.TcUser{}).Where("id = ?", uid).Update("pw", string(hash))

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	var resp = struct {
		UID int32  `json:"uid"`
		Pwd string `json:"pwd"` // <- static session
	}{
		UID: int32(numUID),
		Pwd: hex.EncodeToString(_function.GenHMAC256([]byte(string(hash)), []byte(uid+string(hash)))),
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func GetAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	// check filter

	var accountInfo []model.TcUser
	_function.GormDB.Where("id = ?", uid).Find(&accountInfo)

	var accountSettings []model.TcUsersOption
	_function.GormDB.Where("uid = ?", uid).Find(&accountSettings)

	var resp = struct {
		UID      int32             `json:"uid"`
		Name     string            `json:"name"`
		Email    string            `json:"email"`
		Role     string            `json:"role"`
		Settings map[string]string `json:"settings"`
	}{
		UID:      accountInfo[0].ID,
		Name:     accountInfo[0].Name,
		Email:    accountInfo[0].Email,
		Role:     accountInfo[0].Role,
		Settings: make(map[string]string),
	}

	for _, v := range accountSettings {
		resp.Settings[v.Name] = v.Value
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func GetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accountSettings []model.TcUsersOption
	_function.GormDB.Where("uid = ?", uid).Find(&accountSettings)

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
	_function.GormDB.Where("uid = ?", uid).Find(&accountSettings)

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
			_function.GormDB.Model(model.TcUsersOption{}).Where("uid = ? AND name = ?", uid, v.Name).Updates(&model.TcUsersOption{
				Value: v.Value,
			})
		}
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", settings, "tbsign"))
}
