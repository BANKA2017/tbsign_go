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
)

func init() {
	RegisterPlugin(RenewManager.Name, RenewManager)
}

type RenewManagerType struct {
	PluginInfo
}

var RenewManager = _function.VariablePtrWrapper(RenewManagerType{
	PluginInfo{
		Name:              "kd_renew_manager",
		PluginNameCN:      "吧主考核",
		PluginNameCNShort: "吧主考核",
		PluginNameFE:      "renew_manager",
		Version:           "0.1",
		Options: map[string]string{
			"kd_renew_manager_id":           "0",
			"kd_renew_manager_action_limit": "50",
		},
		SettingOptions: map[string]PluinSettingOption{
			"kd_renew_manager_action_limit": {
				OptionName:   "kd_renew_manager_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: func(value string) bool {
					numLimit, err := strconv.ParseInt(value, 10, 64)
					return err == nil && numLimit >= 0
				},
			},
		},
		Test: false,
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "switch", Function: PluginRenewManagerGetSwitch},
			{Method: "POST", Path: "switch", Function: PluginRenewManagerSwitch},
			{Method: "GET", Path: "alert/switch", Function: PluginRenewManagerGetAlertSwitch},
			{Method: "POST", Path: "alert/switch", Function: PluginRenewManagerAlertSwitch},
			{Method: "GET", Path: "list", Function: PluginRenewManagerGetList},
			{Method: "PATCH", Path: "list", Function: PluginRenewManagerAddAccount},
			{Method: "DELETE", Path: "list/:id", Function: PluginRenewManagerDelAccount},
			{Method: "POST", Path: "list/empty", Function: PluginRenewManagerDelAllAccounts},
			{Method: "GET", Path: "check/:pid/status/:fname", Function: PluginRenewManagerPreCheckStatus},
		},
	},
})

var managerTasksPageLink = []byte{104, 116, 116, 112, 115, 58, 47, 47, 116, 105, 101, 98, 97, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 109, 111, 47, 113, 47, 98, 97, 119, 117, 47, 116, 97, 115, 107, 105, 110, 102, 111, 118, 105, 101, 119, 63, 116, 98, 105, 111, 115, 119, 107, 61, 49, 38, 102, 105, 100, 61}

func (pluginInfo *RenewManagerType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id, err := strconv.ParseInt(_function.GetOption("kd_renew_manager_id"), 10, 64)
	if err != nil {
		id = 0
	}

	otime := _function.Now.Add(time.Hour * -24).Unix()

	limit := _function.GetOption("kd_renew_manager_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	localRenewList := new([]model.TcKdRenewManager)
	subQuery := _function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid").Where("name = 'kd_renew_manager_open' AND value = '1'")
	_function.GormDB.R.Model(&model.TcKdRenewManager{}).Where("id > ? AND date < ? AND uid IN (?)", id, otime, subQuery).Order("id ASC").Limit(int(numLimit)).Find(&localRenewList)

	for _, renewItem := range *localRenewList {
		renewItem.Date = int32(_function.Now.Unix())
		tmpLog := []string{}

		// send cancel top package
		res, err := PluginRenewManagerCancelTop(_function.GetCookie(renewItem.Pid), renewItem.Fname, renewItem.Tid)

		if err != nil {
			log.Println("renew_manager (cancel_top):", res, err)
			renewItem.Status = "failed"
			tmpLog = append(tmpLog, "cacnel_top: failed")
		} else {
			if res.No != 0 {
				log.Println("renew_manager (cancel_top):", res.ErrCode, res.Error)
				renewItem.Status = "failed"
				tmpLog = append(tmpLog, fmt.Sprintf("cacnel_top: %d#%s", res.ErrCode, res.Error))
			} else {
				renewItem.Status = "success"
				tmpLog = append(tmpLog, "cacnel_top: done")
			}
		}

		// sync tasks
		res2, err := _function.GetManagerTasks(_function.GetCookie(renewItem.Pid), int64(renewItem.Fid))
		if err != nil {
			log.Println("renew_manager (sync_tasks):", err)
			tmpLog = append(tmpLog, "sync: failed")
		} else {
			renewItem.End = int32(res2.Data.BawuTask.EndTime)
			if res2.No != 0 {
				log.Println("renew_manager (sync):", res2.ErrCode, res2.Error)
				tmpLog = append(tmpLog, fmt.Sprintf("sync: %d#%s", res2.ErrCode, res2.Error))
			} else {
				tmpLog = append(tmpLog, "sync: done")
			}

			// done?
			// done := false
			// for _,remoteTask := range res2.Data.BawuTask.TaskList{
			// 	if remoteTask.TaskStatus == "1" {
			//
			// 		break
			// 	}
			// }
		}

		// previous logs
		previousLogs := []string{}
		for i, s := range strings.Split(renewItem.Log, "<br />") {
			if i <= 28 {
				previousLogs = append(previousLogs, s)
			} else {
				break
			}
		}
		renewItem.Log = fmt.Sprintf("%s: %s<br />%s", _function.Now.Local().Format(time.DateOnly), strings.Join(tmpLog, ", "), strings.Join(previousLogs, "<br />"))

		_function.GormDB.W.Model(&model.TcKdRenewManager{}).Where("id = ?", renewItem.ID).Updates(renewItem)
		_function.SetOption("kd_renew_manager_id", strconv.Itoa(int(renewItem.ID)))

		// push
		// TODO
		/// now > 10days, end_time < 25 days

		if renewItem.End < int32(_function.Now.Add(time.Duration(time.Hour*24*15)).Unix()) && _function.GetUserOption("kd_renew_manager_alert", strconv.Itoa(int(renewItem.UID))) != "0" {
			endTime := time.Unix(int64(renewItem.End), 0)

			msg := PluginRenewManagerAlertMessage(_function.GetCookie(renewItem.Pid).Name, renewItem.Fname, endTime.Local().Format("2006年01月02日 15:04:05"), renewItem.Fid)

			err = _function.SendMessage("default", renewItem.UID, msg.Title, msg.Body)
			if err != nil {
				log.Println("renew_manager:", err)
			}
		}
	}
	_function.SetOption("kd_renew_manager_id", "0")

	// clean

}

func (pluginInfo *RenewManagerType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcKdRenewManager{})
		_function.GormDB.W.Exec("ALTER TABLE `tc_kd_renew_manager` ADD PRIMARY KEY (`id`), ADD UNIQUE KEY `pid_fid` (`pid`,`fid`) USING BTREE, ADD KEY `id_date_uid` (`id`,`date`,`uid`), ADD KEY `uid_pid_fid` (`uid`,`pid`,`fid`);")
	} else {
		_function.GormDB.W.Migrator().CreateTable(&model.TcKdRenewManager{})

		_function.GormDB.W.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_kd_renew_manager_pid_fid" ON "tc_kd_renew_manager" ("pid","fid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_renew_manager_id_date_uid" ON "tc_kd_renew_manager" ("id","date","uid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_renew_manager_uid_pid_fid" ON "tc_kd_renew_manager" ("uid","pid","fid");`)
	}

	return nil
}

func (pluginInfo *RenewManagerType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	_function.GormDB.W.Migrator().DropTable(&model.TcKdRenewManager{})

	// user options
	_function.GormDB.W.Where("name = ?", "kd_renew_manager_open").Delete(&model.TcUsersOption{})
	_function.GormDB.W.Where("name = ?", "kd_renew_manager_alert").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *RenewManagerType) Upgrade() error {
	return nil
}

func (pluginInfo *RenewManagerType) RemoveAccount(_type string, id int32) error {
	_function.GormDB.W.Where(fmt.Sprintf("%s = ?", _type), id).Delete(&model.TcKdRenewManager{})
	return nil
}

func (pluginInfo *RenewManagerType) Ext() ([]any, error) {
	return []any{}, nil
}

func PluginRenewManagerAlertMessage(name, fname, end string, fid int32) _function.PushMessageTemplateStruct {
	return _function.PushMessageTemplateStruct{
		Title: fmt.Sprintf("吧主考核提醒 - %s吧", fname),
		Body:  fmt.Sprintf("@%s 您的吧主账号在 %s吧 (fid:%d) 的考核任务将于 %s 截止，目前剩余不到 15 天。<br /><br />由于 TbSign 已超过 15 天 未能完成考核，请您尽快前往 [吧主考核页面](%s%d) 完成相关任务。", name, fname, fid, end, managerTasksPageLink, fid),
	}
}

type PluginRenewManagerCancelTopResponse struct {
	No      int `json:"no,omitempty"`
	ErrCode int `json:"err_code,omitempty"`
	Error   any `json:"error,omitempty"`
	// Data    struct {
	// 	SecondClassID string `json:"second_class_id,omitempty"`
	// 	AutoMsg       string `json:"autoMsg,omitempty"`
	// 	Fid           int    `json:"fid,omitempty"`
	// 	Fname         string `json:"fname,omitempty"`
	// 	Tid           int64  `json:"tid,omitempty"`
	// 	IsLogin       int    `json:"is_login,omitempty"`
	// 	Content       string `json:"content,omitempty"`
	// 	AccessState   any    `json:"access_state,omitempty"`
	// 	Experience    int    `json:"experience,omitempty"`
	// 	IsPopAward    int    `json:"is_pop_award,omitempty"`
	// 	PopURL        string `json:"pop_url,omitempty"`
	// 	Vcode         struct {
	// 		NeedVcode       int    `json:"need_vcode,omitempty"`
	// 		StrReason       string `json:"str_reason,omitempty"`
	// 		CaptchaVcodeStr string `json:"captcha_vcode_str,omitempty"`
	// 		CaptchaCodeType int    `json:"captcha_code_type,omitempty"`
	// 		Userstatevcode  int    `json:"userstatevcode,omitempty"`
	// 	} `json:"vcode,omitempty"`
	// } `json:"data,omitempty"`
}

func PluginRenewManagerCancelTop(cookie _type.TypeCookie, fname string, tid string) (*PluginRenewManagerCancelTopResponse, error) {
	headersMap := map[string]string{
		"Cookie":       "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
		"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
		"Referrer":     "https://tieba.baidu.com/p/" + tid,
	}

	body := url.Values{}
	body.Set("ie", "utf-8")
	body.Set("tbs", cookie.Tbs)
	body.Set("kw", fname)
	body.Set("fid", strconv.Itoa(int(_function.GetFid(fname))))
	body.Set("tid", tid)

	res, err := _function.TBFetch("https://tieba.baidu.com/f/commit/thread/top/cancel", "POST", []byte(body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	// log.Println(string(res))

	resp := new(PluginRenewManagerCancelTopResponse)

	err = _function.JsonDecode(res, resp)

	return resp, err
}

type PluginRenewManagerGetThreadInfoResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		Forum struct {
			// ForumHelper struct {
			// 	Name      string `json:"name,omitempty"`
			// 	AvatarURL string `json:"avatar_url,omitempty"`
			// } `json:"forum_helper,omitempty"`
			// ForumAvatar string `json:"forum_avatar,omitempty"`
			ForumName string `json:"forum_name,omitempty"`
		} `json:"forum,omitempty"`
		ThreadInfo struct {
			ThreadID int64 `json:"thread_id,omitempty"`
			// PostID          int64  `json:"post_id,omitempty"`
			Title string `json:"title,omitempty"`
			// Content         string `json:"content,omitempty"`
			// PostCate        int    `json:"post_cate,omitempty"`
			// PostTag         int    `json:"post_tag,omitempty"`
			PostCreateTime string `json:"post_create_time,omitempty"`
			// PostLocation    string `json:"post_location,omitempty"`
			// ShowPostContent string `json:"show_post_content,omitempty"`
			// PassPostContent string `json:"pass_post_content,omitempty"`
			// ContentDetail   []struct {
			// 	Type  int    `json:"type,omitempty"`
			// 	Value string `json:"value,omitempty"`
			// } `json:"content_detail,omitempty"`
			// AllPics      []any `json:"all_pics,omitempty"`
			// IsFrsMask    int   `json:"is_frs_mask,omitempty"`
			// CommentNum   int   `json:"comment_num,omitempty"`
			// ReadNum      int   `json:"read_num,omitempty"`
			// AgreeNum     int   `json:"agree_num,omitempty"`
			// ShareNum     int   `json:"share_num,omitempty"`
			// DisagreeNum  int   `json:"disagree_num,omitempty"`
			// ShareUserNum int   `json:"share_user_num,omitempty"`
			// CollectNum   int   `json:"collect_num,omitempty"`
		} `json:"thread_info,omitempty"`
		// Tbs      string `json:"tbs,omitempty"`
		UserInfo struct {
			UserName     string `json:"user_name,omitempty"`
			UserNick     string `json:"user_nick,omitempty"`
			ShowNickname string `json:"show_nickname,omitempty"`
			Portrait     string `json:"portrait,omitempty"`
		} `json:"user_info,omitempty"`
	} `json:"data,omitempty"`
}

var managerGetThreadInfoLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 116, 105, 101, 98, 97, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 109, 111, 47, 113, 47, 98, 97, 119, 117, 47, 103, 101, 116, 82, 101, 99, 111, 118, 101, 114, 73, 110, 102, 111, 63, 37, 115})

func PluginRenewManagerGetThreadInfo(cookie _type.TypeCookie, tid int64, fid int64) (*PluginRenewManagerGetThreadInfoResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}

	query := url.Values{}
	query.Set("thread_id", strconv.Itoa(int(tid)))
	query.Set("forum_id", strconv.Itoa(int(fid)))
	query.Set("type", "1")
	query.Set("sub_type", "1")
	query.Set("tbs", cookie.Tbs)

	res, err := _function.TBFetch(fmt.Sprintf(managerGetThreadInfoLink, query.Encode()), "GET", nil, headersMap)

	if err != nil {
		return nil, err
	}

	// log.Println(string(res))

	resp := new(PluginRenewManagerGetThreadInfoResponse)

	err = _function.JsonDecode(res, resp)

	return resp, err
}

// endpoint
func PluginRenewManagerGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_renew_manager_open", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("kd_renew_manager_open", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginRenewManagerSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_renew_manager_open", uid) != "0"

	err := _function.SetUserOption("kd_renew_manager_open", !status, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法启用吧主考核功能", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}

func PluginRenewManagerGetAlertSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_renew_manager_alert", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("kd_renew_manager_alert", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginRenewManagerAlertSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_renew_manager_alert", uid) != "0"

	err := _function.SetUserOption("kd_renew_manager_alert", !status, uid)

	if err != nil {
		log.Println("renew_manager:", err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法启用吧主考核提醒功能", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}

func PluginRenewManagerGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	list := new([]model.TcKdRenewManager)
	_function.GormDB.R.Model(&model.TcKdRenewManager{}).Where("uid = ?", uid).Order("id ASC").Find(list)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", list, "tbsign"))
}

func PluginRenewManagerAddAccount(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)
	pid := strings.TrimSpace(c.FormValue("pid"))
	fname := strings.TrimSpace(c.FormValue("fname"))
	tid := strings.TrimSpace(c.FormValue("tid"))

	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	numTid, err := strconv.ParseInt(tid, 10, 64)

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 tid", _function.EchoEmptyObject, "tbsign"))
	}

	// pre check
	var accountInfo model.TcBaiduid
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).First(&accountInfo)
	if accountInfo.Portrait == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// fid
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "贴吧不存在", _function.EchoEmptyObject, "tbsign"))
	}

	existsList := new([]model.TcKdRenewManager)
	_function.GormDB.R.Model(&model.TcKdRenewManager{}).Where("uid = ? AND pid = ? AND fid = ?", uid, pid, fid).Limit(1).Find(&existsList)
	if len(*existsList) > 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(400, "重复任务", _function.EchoEmptyObject, "tbsign"))
	}

	// is manager? && end time
	managerTaskStatus, err := _function.GetManagerTasks(_function.GetCookie(int32(numPid)), fid)
	if err != nil || managerTaskStatus.No != 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无效用户", _function.EchoEmptyObject, "tbsign"))
	}
	end := managerTaskStatus.Data.BawuTask.EndTime

	// TODO tid exists?

	err = _function.GormDB.W.Create(&model.TcKdRenewManager{
		Pid:    int32(numPid),
		UID:    int32(numUID),
		Fname:  fname,
		Fid:    int32(fid),
		Tid:    strconv.Itoa(int(numTid)),
		Status: "idle",
		Date:   0,
		End:    int32(end),
		Log:    "",
	}).Error

	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "添加失败", _function.EchoEmptyObject, "tbsign"))
	}

	renewerTasks := new([]model.TcKdRenewManager)
	_function.GormDB.R.Model(&model.TcKdRenewManager{}).Where("uid = ? AND pid = ? AND fid = ?", uid, numPid, fid).Find(renewerTasks)

	if len(*renewerTasks) == 1 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", (*renewerTasks)[0], "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "查询出错", _function.EchoEmptyObject, "tbsign"))
	}
}

func PluginRenewManagerDelAccount(c echo.Context) error {
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

	_function.GormDB.W.Model(&model.TcKdRenewManager{}).Delete(&model.TcKdRenewManager{
		UID: int32(numUID),
		ID:  int32(numID),
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginRenewManagerDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.W.Model(&model.TcKdRenewManager{}).Delete(&model.TcKdRenewManager{
		UID: int32(numUID),
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func PluginRenewManagerPreCheckStatus(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fname := c.Param("fname")

	// pre-check pid
	var pidCheck []model.TcBaiduid
	_function.GormDB.R.Where("id = ? AND uid = ?", pid, uid).Limit(1).Find(&pidCheck)

	if len(pidCheck) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _type.BawuTask{}, "tbsign"))
	}

	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _type.BawuTask{}, "tbsign"))
	}
	resp, err := _function.GetManagerTasks(_function.GetCookie(pidCheck[0].ID), fid)
	if err != nil {
		log.Println("renew_manager:", err)
	}

	if err != nil || resp.No != 0 {

		// TODO not a good idea
		errStr := resp.Error
		if errStr == "" {
			errStr = "未知错误"
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(500, errStr, _type.BawuTask{}, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp.Data.BawuTask, "tbsign"))
	}
}
