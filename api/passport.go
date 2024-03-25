package _api

import (
	"encoding/hex"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// func Register(c echo.Context) error {
//
// }
//
// func DeleteAccount(c echo.Context) error {
//
// }

func Login(c echo.Context) error {
	account := c.Request().PostFormValue("account") // username or email
	password := c.Request().PostFormValue("password")
	timestamp := strconv.Itoa(int(time.Now().UnixMilli()))

	if account == "" || password == "" {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid account or password", make(map[string]interface{}, 0), "tbsign"))
	}

	// check
	var accountInfo []model.TcUser
	_function.GormDB.Where("name = ? OR email = ?", account, account).Limit(1).Find(&accountInfo)
	//log.Println(accountInfo[0].Pw, password, "password")

	err := bcrypt.CompareHashAndPassword([]byte(accountInfo[0].Pw), []byte(password))
	if err != nil {
		// Compatible with older versions
		md5_ := _function.Md5(_function.Md5(_function.Md5(password)))
		if md5_ == accountInfo[0].Pw {
			log.Println("md5")
		} else {
			return c.JSON(http.StatusOK, apiTemplate(403, "Invalid account or password", make(map[string]interface{}, 0), "tbsign"))
		}
	} else {
		log.Print("right")
	}

	var resp struct {
		UID       int32  `json:"uid"`
		Pwd       string `json:"pwd"` // <- static session
		Timestamp string `json:"timestamp"`
	}

	resp.UID = accountInfo[0].ID
	resp.Pwd = hex.EncodeToString(_function.GenHMAC256([]byte(accountInfo[0].Pw+":"+timestamp), []byte(strconv.Itoa(int(accountInfo[0].ID))+accountInfo[0].Pw)))
	resp.Timestamp = timestamp

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func Logout(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTemplate(200, "In fact, you just need to clear your local cache", make(map[string]interface{}, 0), "tbsign"))
}

func UpdatePassword(c echo.Context) error {
	uid := c.Get("uid").(string)

	oldPwd := c.Request().PostFormValue("old_password")
	newPwd := c.Request().PostFormValue("new_password")

	var accountInfo []model.TcUser
	_function.GormDB.Where("id = ?", uid).Limit(1).Find(&accountInfo)

	// compare old password
	err := bcrypt.CompareHashAndPassword([]byte(accountInfo[0].Pw), []byte(oldPwd))
	if err != nil {
		// Compatible with older versions
		md5_ := _function.Md5(_function.Md5(_function.Md5(oldPwd)))
		if md5_ == accountInfo[0].Pw {
			log.Println("md5")
		} else {
			return c.JSON(http.StatusOK, apiTemplate(403, "Invalid password", make(map[string]interface{}, 0), "tbsign"))
		}
	} else {
		log.Print("right")
	}

	// create new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPwd), 12)
	if err != nil {
		return c.JSON(http.StatusOK, apiTemplate(403, "Encrypt password failed...", make(map[string]interface{}, 0), "tbsign"))
	}

	_function.GormDB.Model(model.TcUser{}).Where("id = ?", uid).Update("pw", string(hash))

	var resp struct {
		UID       int32  `json:"uid"`
		Pwd       string `json:"pwd"` // <- static session
		Timestamp string `json:"timestamp"`
	}

	timestamp := strconv.Itoa(int(time.Now().UnixMilli()))

	numberUID, _ := strconv.ParseInt(uid, 10, 64)
	resp.UID = int32(numberUID)
	resp.Pwd = hex.EncodeToString(_function.GenHMAC256([]byte(string(hash)+":"+timestamp), []byte(uid+string(hash))))
	resp.Timestamp = timestamp

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func GetAccountInfo(c echo.Context) error {
	uid := c.Get("uid").(string)

	// check filter

	var accountInfo []model.TcUser
	_function.GormDB.Where("id = ?", uid).Find(&accountInfo)

	var accountSettings []model.TcUsersOption
	_function.GormDB.Where("uid = ?", uid).Find(&accountSettings)

	var resp struct {
		UID      int32             `json:"uid"`
		Name     string            `json:"name"`
		Email    string            `json:"email"`
		Role     string            `json:"role"`
		Settings map[string]string `json:"settings"`
	}

	resp.UID = accountInfo[0].ID
	resp.Name = accountInfo[0].Name
	resp.Email = accountInfo[0].Email
	resp.Role = accountInfo[0].Role

	resp.Settings = make(map[string]string)

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
