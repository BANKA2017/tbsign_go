package _plugin

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func init() {
	RegisterPlugin(LoopBanPlugin.Name, LoopBanPlugin)
}

type BanAccountResponse struct {
	Un         string `json:"un,omitempty"`
	ServerTime string `json:"server_time,omitempty"`
	Time       int    `json:"time,omitempty"`
	Ctime      int    `json:"ctime,omitempty"`
	Logid      int    `json:"logid,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_msg,omitempty"`
	Info       []any  `json:"info,omitempty"`
}

type LoopBanPluginType struct {
	PluginInfo
}

var LoopBanPlugin = _function.VariablePtrWrapper(LoopBanPluginType{
	PluginInfo{
		Name:              "ver4_ban",
		PluginNameCN:      "循环封禁",
		PluginNameCNShort: "循环封禁",
		PluginNameFE:      "loop_ban",
		Version:           "1.4",
		Options: map[string]string{
			"ver4_ban_break_check":  "0",
			"ver4_ban_id":           "0",
			"ver4_ban_limit":        "5",
			"ver4_ban_action_limit": "50",
		},
		SettingOptions: map[string]PluinSettingOption{
			"ver4_ban_break_check": {
				OptionName:   "ver4_ban_break_check",
				OptionNameCN: "跳过吧务权限检查",
				Validate:     PluginLoopBanOptionValidatorVer4BanBreakCheck,
			},
			"ver4_ban_limit": {
				OptionName:   "ver4_ban_limit",
				OptionNameCN: "可添加循环封禁账号上限",
				Validate:     PluginLoopBanOptionValidatorVer4BanLimit,
			},
			"ver4_ban_action_limit": {
				OptionName:   "ver4_ban_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: func(value string) bool {
					numLimit, err := strconv.ParseInt(value, 10, 64)
					return err == nil && numLimit >= 0
				},
			},
		},
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "switch", Function: PluginLoopBanGetSwitch},
			{Method: "POST", Path: "switch", Function: PluginLoopBanSwitch},
			{Method: "GET", Path: "reason", Function: PluginLoopBanGetReason},
			{Method: "PUT", Path: "reason", Function: PluginLoopBanSetReason},
			{Method: "GET", Path: "list", Function: PluginLoopBanGetList},
			{Method: "PATCH", Path: "list", Function: PluginLoopBanAddAccounts},
			{Method: "DELETE", Path: "list/:id", Function: PluginLoopBanDelAccount},
			{Method: "POST", Path: "list/empty", Function: PluginLoopBanDelAllAccounts},
		},
	},
})

var banDays = []int32{1, 3, 10}

func PostClientBan(cookie _type.TypeCookie, fid int32, portrait string, day int32, reason string) (*BanAccountResponse, error) {
	isSvipBlock := "0"
	if day <= 90 && !slices.Contains(banDays, day) {
		isSvipBlock = "1"
	}

	var form = map[string]string{
		"BDUSS":       cookie.Bduss,
		"day":         strconv.Itoa(int(day)),
		"fid":         strconv.Itoa(int(fid)),
		"is_loop_ban": isSvipBlock, // <- Users have to check their svip status in advance
		"ntn":         "banid",
		"portrait":    portrait,
		"reason":      reason,
		"tbs":         cookie.Tbs,
		"word":        "-",
		"z":           "6",
	}
	_function.AddSign(&form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	banResponse, err := _function.TBFetch("http://c.tieba.baidu.com/c/c/bawu/commitprison", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var banDecode BanAccountResponse
	err = _function.JsonDecode(banResponse, &banDecode)
	return &banDecode, err
}

func (pluginInfo *LoopBanPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id, err := strconv.ParseInt(_function.GetOption("ver4_ban_id"), 10, 64)
	if err != nil {
		id = 0
	}
	otime := _function.Now.Add(time.Hour * -24).Unix()
	var localBanAccountList []*model.TcVer4BanList
	subQuery := _function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid").Where("name = 'ver4_ban_open' AND value = '1'")

	limit := _function.GetOption("ver4_ban_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	_function.GormDB.R.Model(&model.TcVer4BanList{}).Where("id > ? AND date < ? AND stime < ? AND etime > ? AND uid IN (?)", id, otime, _function.Now.Unix(), _function.Now.Unix(), subQuery).Order("id ASC").Limit(int(numLimit)).Find(&localBanAccountList)

	var reasonList = &[]model.TcVer4BanUserset{}
	_function.GormDB.R.Model(&model.TcVer4BanUserset{}).Find(&reasonList)

	for _, banAccountInfo := range localBanAccountList {
		// find reason
		var reason = "您因为违反吧规，已被吧务封禁，如有疑问请联系吧务！"
		for _, reasonDB := range *reasonList {
			if reasonDB.UID == banAccountInfo.UID && reasonDB.C != "" {
				reason = reasonDB.C
				break
			}
		}

		//get fid
		fid := _function.GetFid(banAccountInfo.Tieba)
		if fid == 0 {
			log.Println("fname: ", banAccountInfo.Tieba, "is not exists!")
			continue
		}

		// !!! warning: unable to check permission !!!
		response, err := PostClientBan(_function.GetCookie(banAccountInfo.Pid), int32(fid), banAccountInfo.Portrait, 1, reason)
		if err != nil {
			log.Println("ban:", err)
			continue
		}
		msg := banAccountInfo.Log
		if response.ErrorMsg != "" {
			msg += _function.Now.Local().Format(time.DateTime) + " 执行结果：<font color=\"red\">操作失败</font>#" + response.ErrorCode + " " + response.ErrorMsg + "<br>"
		} else {
			msg += _function.Now.Local().Format(time.DateTime) + " 执行结果：<font color=\"green\">操作成功</font><br>"
		}

		_function.GormDB.W.Model(&model.TcVer4BanList{}).Where("id = ?", banAccountInfo.ID).Updates(model.TcVer4BanList{
			Log:  msg,
			Date: int32(_function.Now.Unix()),
		})
		_function.SetOption("ver4_ban_id", strconv.Itoa(int(banAccountInfo.ID)))
	}
	_function.SetOption("ver4_ban_id", "0")

	// clean

}

func (pluginInfo *LoopBanPluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcVer4BanUserset{})
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcVer4BanList{})
		_function.GormDB.W.Exec("ALTER TABLE `tc_ver4_ban_list` ADD KEY `uid` (`uid`), ADD KEY `id_uid` (`id`,`uid`), ADD KEY `pid` (`pid`), ADD KEY `id_date_stime_etime_uid` (`id`,`date`,`stime`,`etime`,`uid`) USING BTREE;")
		_function.GormDB.W.Exec("ALTER TABLE `tc_ver4_ban_userset` ADD UNIQUE KEY `uid` (`uid`);")
	} else {
		_function.GormDB.W.Set("gorm:table_options", "WITHOUT ROWID").Migrator().CreateTable(&model.TcVer4BanUserset{})
		_function.GormDB.W.Migrator().CreateTable(&model.TcVer4BanList{})

		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_uid" ON "tc_ver4_ban_list" ("uid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_uid" ON "tc_ver4_ban_list" ("id","uid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_pid" ON "tc_ver4_ban_list" ("pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_date_stime_etime_uid" ON "tc_ver4_ban_list" ("id","date","stime","etime","uid");`)
	}
	return nil
}

func (pluginInfo *LoopBanPluginType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)
	_function.GormDB.W.Migrator().DropTable(&model.TcVer4BanUserset{}, &model.TcVer4BanList{})

	// user options
	_function.GormDB.W.Where("name = ?", "ver4_ban_open").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *LoopBanPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *LoopBanPluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}

	if err := _sql.Where(_function.AppendStrings(_type, " = ?"), id).Delete(&model.TcVer4BanList{}).Error; err != nil {
		return err
	}
	if _type == "uid" {
		return _sql.Where("uid = ?", id).Delete(&model.TcVer4BanUserset{}).Error
	}
	return nil
}

func (pluginInfo *LoopBanPluginType) Ext() ([]any, error) {
	return []any{}, nil
}

// OptionValidator

func PluginLoopBanOptionValidatorVer4BanBreakCheck(value string) bool {
	return value == "0" || value == "1"
}

func PluginLoopBanOptionValidatorVer4BanLimit(value string) bool {
	numValue, err := strconv.ParseInt(value, 10, 64)
	return err == nil && numValue >= 0
}

// endpoint

type addAccountsResponseList struct {
	ID       int32  `json:"id"`
	PID      int32  `json:"pid"`
	Name     string `json:"name,omitempty"`
	NameShow string `json:"name_show,omitempty"`
	Portrait string `json:"portrait"`
	Fname    string `json:"fname,omitempty"`
	Start    int64  `json:"start,omitempty"`
	End      int64  `json:"end,omitempty"`
	Success  bool   `json:"success"`
	Msg      string `json:"msg,omitempty"`
	Log      string `json:"log,omitempty"`
	Date     int32  `json:"date,omitempty"`
}

func PluginLoopBanGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_ban_open", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("ver4_ban_open", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginLoopBanSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_ban_open", uid) != "0"

	err := _function.SetUserOption("ver4_ban_open", !status, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法启用循环封禁功能", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}

func PluginLoopBanGetReason(c echo.Context) error {
	uid := c.Get("uid").(string)

	var loopBanSettings []model.TcVer4BanUserset
	_function.GormDB.R.Where("uid = ?", uid).Limit(1).Find(&loopBanSettings)

	reason := ""
	if len(loopBanSettings) == 0 {
		reason = "您因为违反吧规，已被吧务封禁，如有疑问请联系吧务！"
	} else {
		reason = loopBanSettings[0].C
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]string{
		"reason": reason,
	}, "tbsign"))
}

func PluginLoopBanSetReason(c echo.Context) error {
	uid := c.Get("uid").(string)

	reason := c.FormValue("reason")

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	err := _function.GormDB.W.Model(&model.TcVer4BanUserset{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcVer4BanUserset{UID: int32(numUID), C: reason}).Error
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法更新封禁理由", map[string]string{
			"reason": reason,
		}, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]string{
		"reason": reason,
	}, "tbsign"))

}

func PluginLoopBanGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var loopBanList []model.TcVer4BanList
	_function.GormDB.R.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Order("id ASC").Find(&loopBanList)

	limit := _function.GetOption("ver4_ban_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)

	var responseList []addAccountsResponseList

	for _, v := range loopBanList {
		responseList = append(responseList, addAccountsResponseList{
			ID:       v.ID,
			PID:      v.Pid,
			Name:     v.Name,
			NameShow: v.NameShow,
			Portrait: v.Portrait,
			Fname:    v.Tieba,
			Start:    int64(v.Stime),
			End:      int64(v.Etime),
			Success:  true,
			Log:      v.Log,
			Date:     v.Date,
		})
	}

	var list = struct {
		Count int64                     `json:"count"`
		Limit int64                     `json:"limit"`
		List  []addAccountsResponseList `json:"list"`
	}{
		Count: int64(len(responseList)),
		Limit: numLimit,
		List:  responseList,
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", list, "tbsign"))
}

func PluginLoopBanAddAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)
	pid := c.FormValue("pid")
	start := c.FormValue("start")
	end := c.FormValue("end")
	fname := c.FormValue("fname")
	portraits := strings.TrimSpace(c.FormValue("portrait"))

	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// time
	startTime := time.Now()
	if start != "" {
		startTime, err = time.Parse(time.DateOnly, start)
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "开始日期格式错误", _function.EchoEmptyObject, "tbsign"))
		}
	}

	endTime, err := time.Parse(time.DateOnly, end)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "结束日期格式错误", _function.EchoEmptyObject, "tbsign"))
	}

	if startTime.Unix() >= endTime.Unix() {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "开始时刻晚于结束时刻", _function.EchoEmptyObject, "tbsign"))
	}

	if endTime.Unix() < time.Now().Unix() {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "现在时刻晚于结束时刻", _function.EchoEmptyObject, "tbsign"))
	}

	// pre check
	var accountInfo model.TcBaiduid
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).First(&accountInfo)
	if accountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// portrait
	if portraits == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "待封禁 portrait 列表为空!", _function.EchoEmptyObject, "tbsign"))
	}
	portraitList := []string{}
	for _, portrait := range strings.Split(portraits, "\n") {
		if strings.HasPrefix(portrait, "tb.1.") {
			portraitList = append(portraitList, portrait)
		}
	}

	// limit
	limit := _function.GetOption("ver4_ban_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)

	var existsAccountList []model.TcVer4BanList
	_function.GormDB.R.Model(&model.TcVer4BanList{}).Where("uid = ?", uid).Order("id ASC").Find(&existsAccountList)

	count := len(existsAccountList)
	if count >= int(numLimit) || count+len(portraitList) > int(numLimit) {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, fmt.Sprintf("添加账号数超限（%d/%s）", count+len(portraitList), limit), _function.EchoEmptyObject, "tbsign"))
	}

	// fid
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "贴吧不存在", _function.EchoEmptyObject, "tbsign"))
	}

	// is manager?
	if _function.GetOption("ver4_ban_break_check") == "0" {
		managerStatus, err := _function.GetManagerStatus(_function.GetCookie(int32(numPid)).Portrait, fid)
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法获取吧务列表", _function.EchoEmptyObject, "tbsign"))
		}
		if !managerStatus.IsManager {
			return c.JSON(http.StatusOK, _function.ApiTemplate(403, "您不是 fname:"+fname+" 的吧务成员", _function.EchoEmptyObject, "tbsign"))
		}
	}

	// get account info
	var accountsResult []addAccountsResponseList
	var accountsToInsert []model.TcVer4BanList
	var successPortraitList []string
	for _, portrait := range portraitList {
		// check db
		dbExists := false
		for _, v := range existsAccountList {
			if v.Portrait == portrait {
				accountsResult = append(accountsResult, addAccountsResponseList{
					ID:       v.ID,
					PID:      v.Pid,
					Name:     v.Name,
					NameShow: v.NameShow,
					Portrait: portrait,
					Fname:    v.Tieba,
					Start:    int64(v.Stime),
					End:      int64(v.Etime),
					Success:  false,
					Msg:      "账号已存在",
				})
				dbExists = true
				break
			}
		}
		if dbExists {
			continue
		}

		// check exists
		banUserInfo, err := _function.GetUserInfoByUsernameOrPortrait("portrait", portrait)
		if err != nil && banUserInfo.No != 0 {
			accountsResult = append(accountsResult, addAccountsResponseList{
				PID:      int32(numPid),
				Portrait: portrait,
				Success:  false,
				Msg:      "账号不存在",
			})
		}
		successPortraitList = append(successPortraitList, portrait)
		accountsToInsert = append(accountsToInsert, model.TcVer4BanList{
			UID:      int32(numUID),
			Pid:      int32(numPid),
			Name:     banUserInfo.Data.Name,
			NameShow: banUserInfo.Data.NameShow,
			Portrait: portrait,
			Tieba:    fname,
			Stime:    int32(startTime.Unix()),
			Etime:    int32(endTime.Unix()),
			Date:     0,
		})
	}

	if len(accountsToInsert) > 0 {
		_function.GormDB.W.Create(&accountsToInsert)
	}

	var loopBanList []model.TcVer4BanList
	_function.GormDB.R.Model(&model.TcVer4BanList{}).Where("uid = ? AND portrait IN ?", uid, successPortraitList).Order("id ASC").Find(&loopBanList)

	for _, v := range loopBanList {
		accountsResult = append(accountsResult, addAccountsResponseList{
			ID:       v.ID,
			PID:      v.Pid,
			Name:     v.Name,
			NameShow: v.NameShow,
			Portrait: v.Portrait,
			Fname:    v.Tieba,
			Start:    int64(v.Stime),
			End:      int64(v.Etime),
			Success:  true,
			Log:      v.Log,
			Date:     v.Date,
		})
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", accountsResult, "tbsign"))
}

func PluginLoopBanDelAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	id := c.Param("id")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无效 id", map[string]any{
			"success": false,
			"id":      id,
		}, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numUID),
		ID:  int32(numID),
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginLoopBanDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.W.Model(&model.TcVer4BanList{}).Delete(&model.TcVer4BanList{
		UID: int32(numUID),
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}
