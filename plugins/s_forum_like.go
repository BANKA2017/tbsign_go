package _plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func init() {
	PluginList.Register(ForumLikePluginInfo)
}

type ForumLikePluginInfoType struct {
	PluginInfo
}

var ForumLikePluginInfo = _function.VPtr(ForumLikePluginInfoType{
	PluginInfo{
		Name:              "kd_forum_like",
		PluginNameCN:      "批量关注贴吧",
		PluginNameCNShort: "批量关注",
		PluginNameFE:      "forum_like",
		Version:           "0.1",
		RandomDuration:    true,
		Test:              true,
		Options: map[string]string{
			"kd_forum_like_action_limit":         "50",
			"kd_forum_like_forum_limit_each_pid": "40",
			"kd_forum_like_cooldown_time_pid":    "900",
			"kd_forum_like_cooldown_time_fname":  "900",
		},
		SettingOptions: map[string]PluginSettingOption{
			"kd_forum_like_action_limit": {
				OptionName:   "kd_forum_like_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
			// "kd_forum_like_forum_limit_each_uid": {
			// 	OptionName:   "kd_forum_like_forum_limit_each_uid",
			// 	OptionNameCN: "用户关注上限",
			// 	Validate: &_function.OptionRule{
			// 		Min: _function.VPtr(int64(0)),
			// 	},
			// },
			"kd_forum_like_forum_limit_each_pid": {
				OptionName:   "kd_forum_like_forum_limit_each_pid",
				OptionNameCN: "单个贴吧账号关注上限",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
			"kd_forum_like_cooldown_time_pid": {
				OptionName:   "kd_forum_like_cooldown_time_pid",
				OptionNameCN: "用户关注间隔 (s) +[0, 300)s",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
			"kd_forum_like_cooldown_time_fname": {
				OptionName:   "kd_forum_like_cooldown_time_fname",
				OptionNameCN: "贴吧关注间隔 (s)",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
		},
		Endpoints: []PluginEndpointStruct{
			{Method: http.MethodGet, Path: "switch", Function: PluginForumLikeGetSwitch},
			{Method: http.MethodPost, Path: "switch", Function: PluginForumLikeSwitch},
			{Method: http.MethodGet, Path: "settings", Function: PluginForumLikeForumConfig},
			// {Method: http.MethodPost, Path: "settings", Function: _function.EchoNoContent},
			{Method: http.MethodGet, Path: "list", Function: PluginForumLikeForumList},
			{Method: http.MethodGet, Path: "list/:pid", Function: PluginForumLikeForumList},
			{Method: http.MethodPut, Path: "list/:pid", Function: PluginForumLikeForumListAdd},
			{Method: http.MethodPost, Path: "list/:pid/clone/:source_pid/:source", Function: PluginForumLikeForumListClone},
			{Method: http.MethodDelete, Path: "list", Function: PluginForumLikeForumListDelete},
			{Method: http.MethodDelete, Path: "list/succeed", Function: PluginForumLikeForumListDeleteSucceed},
			{Method: http.MethodDelete, Path: "list/:pid", Function: PluginForumLikeForumListDelete},
			{Method: http.MethodDelete, Path: "list/:pid/succeed", Function: PluginForumLikeForumListDeleteSucceed},
			{Method: http.MethodDelete, Path: "list/:pid/:tid", Function: PluginForumLikeForumListDelete},
		},
	},
})

func (pluginInfo *ForumLikePluginInfoType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	var forumTasksList []*model.TcKdForumLike

	now := time.Now().Unix()

	// options

	actionTime, _ := strconv.Atoi(_function.GetOption("kd_forum_like_action_limit"))
	pidCooldownTime, _ := strconv.Atoi(_function.GetOption("kd_forum_like_cooldown_time_pid"))
	fnameCooldownTime, _ := strconv.Atoi(_function.GetOption("kd_forum_like_cooldown_time_fname"))

	// safe check
	if actionTime <= 0 {
		return
	}

	if pidCooldownTime < 1 || fnameCooldownTime < 1 {
		slog.Warn("plugin.forum-like.action.too-short-cooldown-time", "action_limit", actionTime, "pid_cooldown_time", pidCooldownTime, "fname_cooldown_time", fnameCooldownTime)
	}

	// TODO remove hard coded
	pidCooldownTime += rand.Intn(300)

	// we have to use raw SQL here
	// 1e6 random forums, select 50 forums one time takes ~2s
	// 1e5 random forums, select 50 forums one time takes ~1s
	// 此方案有一个漏洞，如果某个 pid/fname 在表中唯一，可以通过反复删除已完成记录来无视冷却时间，但实际上无视冷却时间更可能会导致封号，所以不需要修复
	// TODO maybe fix raw SQL in the future, raw table: tc_kd_forum_like, tc_users_options
	if err := _function.GormDB.R.Raw(`WITH
	success_count AS ( SELECT pid, COUNT(*) AS cnt FROM tc_kd_forum_like WHERE status = 0 AND date >= ? GROUP BY pid),
	candidate AS (
	    SELECT t.id, t.pid, t.fname FROM tc_kd_forum_like t
		LEFT JOIN success_count sc ON sc.pid = t.pid
	    LEFT JOIN (SELECT pid, MAX(date) AS last_pid_time FROM tc_kd_forum_like WHERE date > ? GROUP BY pid) p ON p.pid = t.pid
	    LEFT JOIN (SELECT fname, MAX(date) AS last_fname_time FROM tc_kd_forum_like WHERE date > ? GROUP BY fname) f ON f.fname = t.fname
	    WHERE t.status = 0 AND t.date = 0 AND EXISTS (SELECT 1 FROM tc_users_options uo WHERE uo.uid = t.uid AND uo.name = 'kd_forum_like_check' AND uo.value = '1') AND (p.last_pid_time IS NULL OR p.last_pid_time <= ?) AND (f.last_fname_time IS NULL OR f.last_fname_time <= ?) AND COALESCE(sc.cnt, 0) < 40
	),
	pid_ranked AS ( SELECT id, pid, fname, ROW_NUMBER() OVER (PARTITION BY pid ORDER BY id ASC) AS rn_pid FROM candidate),
	pid_first AS ( SELECT id, pid, fname FROM pid_ranked WHERE rn_pid = 1),
	fname_ranked AS ( SELECT id, pid, fname, ROW_NUMBER() OVER (PARTITION BY fname ORDER BY id ASC ) AS rn_fname FROM pid_first )
	SELECT id, pid, fname FROM fname_ranked WHERE rn_fname = 1 ORDER BY id ASC LIMIT ?;`, now-24*60*60, int(now)-pidCooldownTime, int(now)-fnameCooldownTime, int(now)-pidCooldownTime, int(now)-fnameCooldownTime, actionTime).Find(&forumTasksList).Error; err != nil {
		slog.Error("plugin.forum-like.action.get-tasks-error", "error", err)
		return
	}

	for _, task := range forumTasksList {
		// time
		task.Date = now

		// cookie
		cookie := _function.GetCookie(task.Pid)

		if cookie == nil || !cookie.IsLogin {
			task.Status = 1
			task.LastError = "用户未登录或登录失败，请更换账号或重试"
			slog.Error("plugin.forum-like.action.cookie-not-found", "pid", task.Pid, "fname", task.Fname, "error", task.LastError)
			if err := _function.GormDB.W.Model(&model.TcKdForumLike{}).Select("status", "last_error", "date").Where("id = ?", task.ID).Updates(task).Error; err != nil {
				slog.Error("plugin.forum-like.action.update-task-error", "pid", task.Pid, "fname", task.Fname, "error", err)
			}
			continue
		}

		// get fid
		fid := _function.GetFid(task.Fname)

		if fid == 0 {
			task.Status = 300003
			task.LastError = "贴吧不存在"
			slog.Error("plugin.forum-like.action.fid-not-found", "pid", task.Pid, "fname", task.Fname, "error", task.LastError)
			if err := _function.GormDB.W.Model(&model.TcKdForumLike{}).Select("status", "last_error", "date").Where("id = ?", task.ID).Updates(task).Error; err != nil {
				slog.Error("plugin.forum-like.action.update-task-error", "pid", task.Pid, "fname", task.Fname, "error", err)
			}
			continue
		}

		res, err := PostForumLikeClient(cookie, fid)

		if err != nil {
			task.LastError = "请求失败"
			slog.Error("plugin.forum-like.action.post-error", "pid", task.Pid, "fname", task.Fname, "error", err)
			if err := _function.GormDB.W.Model(&model.TcKdForumLike{}).Select("status", "last_error", "date").Where("id = ?", task.ID).Updates(task).Error; err != nil {
				slog.Error("plugin.forum-like.action.update-task-error", "pid", task.Pid, "fname", task.Fname, "error", err)
			}
			continue
		}

		if res.ErrorCode != "" && res.ErrorCode != "0" {
			// 300000 不存在（fid=0）
			// 300003 null（fid 不存在，自增未到达/彻底清除）
			// 2410003 不合理（fid 非正整数）
			// 1 用户未登录或登录失败，请更换账号或重试
			// 1990029 操作频繁，请稍后再试
			// 3250001 您的帐号涉及违规操作，现已被贴吧官方系统封禁，可进行申诉。

			errCode, _ := strconv.Atoi(res.ErrorCode)
			task.Status = int32(errCode)
			task.LastError = res.ErrorMsg
			if task.Status == 300003 && task.LastError == "" {
				task.LastError = "贴吧不存在"
			}

			slog.Error("plugin.forum-like.action.post-error", "pid", task.Pid, "fname", task.Fname, "error", task.LastError)
		} else if res.Error.Errno != "" && res.Error.Errno != "0" {
			// errno 永远等于 0，暂时找不到例外

			errCode, _ := strconv.Atoi(res.Error.Errno)
			task.Status = int32(errCode)
			task.LastError = res.Error.Usermsg
			slog.Error("plugin.forum-like.action.post-error", "pid", task.Pid, "fname", task.Fname, "error", task.LastError)
		} else {
			// task.Status = 0

			if res.Info.IsBlack == "1" {
				task.LastError = "关注成功 [ 黑名单 ]"
			} else if res.Info.IsLike == "1" {
				task.LastError = "关注成功 [ 重复关注 ]"
			} else {
				task.LastError = "关注成功"
			}
		}

		if err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.TcKdForumLike{}).Select("status", "last_error", "date").Where("id = ?", task.ID).Updates(task).Error; err != nil {
				return err
			}

			if task.Status == 1990029 || task.Status == 3250001 {
				if err := tx.Model(&model.TcKdForumLike{}).Where("pid = ?", task.Pid).Select("status", "last_error", "date").Updates(&model.TcKdForumLike{
					Status:    task.Status,
					LastError: task.LastError,
					Date:      task.Date,
				}).Error; err != nil {
					return err
				}
			}

			return nil
		}); err != nil {
			slog.Error("plugin.forum-like.action.update-task-error", "pid", task.Pid, "fname", task.Fname, "error", err)
		}

		// force sleep now, maybe can be custom in the future
		time.Sleep(time.Second * time.Duration(10+rand.Intn(20)))
	}
}

func (pluginInfo *ForumLikePluginInfoType) Install() error {
	var err error
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	err = UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")
	if err != nil {
		return err
	}

	return _function.GormDB.W.Migrator().CreateTable(&model.TcKdForumLike{})
}

func (pluginInfo *ForumLikePluginInfoType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	_function.GormDB.W.Migrator().DropTable(&model.TcKdForumLike{})

	return nil
}
func (pluginInfo *ForumLikePluginInfoType) Upgrade() error {
	return nil
}

// _type: `uid`, `pid`
func (pluginInfo *ForumLikePluginInfoType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}
	return _sql.Where(_type+" = ?", id).Delete(&model.TcKdForumLike{}).Error
}

func (pluginInfo *ForumLikePluginInfoType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *ForumLikePluginInfoType) Reset(uid, pid, tid int32) error {
	if uid == 0 {
		return errors.New("invalid uid")
	}

	_sql := _function.GormDB.W.Model(&model.TcKdForumLike{}).Where("uid = ?", uid)
	if pid != 0 {
		_sql = _sql.Where("pid = ?", pid)
	}

	if tid != 0 {
		_sql = _sql.Where("id = ?", tid)
	}

	return _sql.Updates(map[string]any{
		"date":   0,
		"status": 0,
	}).Error
}

func (pluginInfo *ForumLikePluginInfoType) ExportAccount(uid int32, tx *gorm.DB) (map[string]any, error) {
	if !pluginInfo.GetSwitch() {
		return nil, nil
	}

	tableName := (&model.TcKdForumLike{}).TableName()
	var exportData []*model.TcKdForumLike

	if tx == nil {
		tx = _function.GormDB.R
	}

	err := tx.Model(&model.TcKdForumLike{}).Where("uid = ?", uid).Find(&exportData).Error

	return map[string]any{
		tableName: exportData,
		"tc_users_options": _function.GetUserOptionBatch(strconv.Itoa(int(uid)), _function.OptionExt{
			Tx:      tx,
			KeyName: "kd_forum_like_check",
		}),
	}, err
}

func (pluginInfo *ForumLikePluginInfoType) ImportAccount(uid int32, pid map[int32]int32, data map[string]json.RawMessage, tx *gorm.DB) error {
	if !pluginInfo.GetSwitch() {
		return errors.New("plugin is not enabled")
	}

	if tx == nil {
		tx = _function.GormDB.W
	}

	tableName := (&model.TcKdForumLike{}).TableName()

	var data2 []*model.TcKdForumLike
	if err := _function.JsonDecode(data[tableName], &data2); err != nil {
		return errors.New("invalid data format")
	}

	var data3 []*model.TcKdForumLike

	var localTasks []*model.TcKdForumLike
	_function.GormDB.R.Model(&model.TcKdForumLike{}).Select("pid", "fname").Where("uid = ?", uid).Find(&localTasks)

	pidFidMap := make(map[int32][]string)

	for _, task := range localTasks {
		if _, ok := pidFidMap[task.Pid]; !ok {
			pidFidMap[task.Pid] = []string{}
		}
		pidFidMap[task.Pid] = append(pidFidMap[task.Pid], task.Fname)
	}

	for i := range data2 {
		if pid, ok := pid[data2[i].Pid]; ok {
			if fname, ok := pidFidMap[pid]; !ok || !slices.Contains(fname, data2[i].Fname) {
				data2[i].Pid = pid
				data2[i].ID = 0
				data2[i].UID = uid
				data3 = append(data3, data2[i])
			}
		}
	}

	if len(data3) == 0 {
		return nil
	}

	return tx.Model(&model.TcKdForumLike{}).Create(data3).Error
}

func GetMethodForumLike(cookie *_type.TypeCookie, uid, fid int64, fname string) (any, error) {
	headersMap := map[string]string{
		"Cookie":  "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
		"Referer": "https://tieba.baidu.com/mo/q/hybrid-main-frs/basicProfile/hybrid?customfullscreen=1&forum_id=" + strconv.Itoa(int(fid)) + "&nonavigationbar=1&loadingSignal=1&skin=dark&_client_version=" + _function.ClientVersion,
	}

	query := url.Values{}
	query.Set("fid", strconv.Itoa(int(fid)))
	query.Set("kw", fname)
	query.Set("itb_tbs", cookie.Tbs)
	query.Set("uid", strconv.Itoa(int(uid)))

	forumLikeRes, err := _function.TBFetch("https://tieba.baidu.com/mo/q/favolike?"+query.Encode(), http.MethodGet, nil, headersMap)

	if err != nil {
		return nil, err
	}

	var forumLikeResStruct any
	err = _function.JsonDecode(forumLikeRes, &forumLikeResStruct)
	return &forumLikeResStruct, err
}

type PostForumLikeClientResponseStruct struct {
	Info struct {
		CurScore     string `json:"cur_score"`
		LevelupScore string `json:"levelup_score"`
		IsLike       string `json:"is_like"`
		IsBlack      string `json:"is_black"`
		LikeNum      string `json:"like_num"`
		LevelId      string `json:"level_id"`
		LevelName    string `json:"level_name"`
		MemberSum    string `json:"member_sum"`
	} `json:"info"`
	UserPerm struct {
		LevelId   string `json:"level_id"`
		LevelName string `json:"level_name"`
	} `json:"user_perm"`
	Error struct {
		Errno   string `json:"errno"`
		Errmsg  string `json:"errmsg"`
		Usermsg string `json:"usermsg"`
	} `json:"error"`
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
}

var phpArrayDataInObjectResponse = []byte(",\"info\":[],")

func PostForumLikeClient(cookie *_type.TypeCookie, fid int64) (*PostForumLikeClientResponseStruct, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	// form["stoken"] = cookie.Stoken
	form["tbs"] = cookie.Tbs
	form["fid"] = strconv.Itoa(int(fid))
	// form["subapp_type"] = "hybrid"

	_function.ClientTypeFallBack(form, "android")
	_function.AddSign(form, "android")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	likeResponse, err := _function.TBFetch("https://tiebac.baidu.com/c/c/forum/like", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

	if err != nil {
		return nil, err
	}

	if bytes.Contains(likeResponse, phpArrayDataInObjectResponse) {
		likeResponse = bytes.ReplaceAll(likeResponse, phpArrayDataInObjectResponse, []byte(",\"info\":{},"))
	}

	var resp PostForumLikeClientResponseStruct
	err = _function.JsonDecode(likeResponse, &resp)

	return &resp, err
}

func PluginForumLikeGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_forum_like_check", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("kd_forum_like_check", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginForumLikeSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("kd_forum_like_check", uid) != "0"

	err := _function.SetUserOption("kd_forum_like_check", !status, uid)

	if err != nil {
		slog.Debug("plugin.forum-like.switch", "uid", uid, "current_status", status, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "无法修改批量关注贴吧插件状态", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}

func PluginForumLikeForumList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	var numPid int64
	if pid != "" {
		numPid, _ = strconv.ParseInt(pid, 10, 64)
	}

	tx := _function.GormDB.W.Where("uid = ?", uid)

	if numPid > 0 {
		tx = tx.Where("pid = ?", numPid)
	}

	var forumList []*model.TcKdForumLike
	tx.Find(&forumList)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", forumList, "tbsign"))
}

func pluginForumLikeListMerge(uid, pid int32) []string {
	var fname []string

	_function.GormDB.R.Model(&model.TcTieba{}).Where("uid = ? AND pid = ?", uid, pid).Pluck("tieba", &fname)

	var fnameFromPlugin []string
	_function.GormDB.R.Model(&model.TcKdForumLike{}).Where("uid = ? AND pid = ?", uid, pid).Pluck("fname", &fnameFromPlugin)

	for _, singleFname := range fnameFromPlugin {
		if !slices.Contains(fname, singleFname) {
			fname = append(fname, singleFname)
		}
	}

	return fname
}

type PluginForumLikeListBinding struct {
	Pid   uint64   `param:"pid"`
	Fname []string `form:"fname" json:"fname"`
}

func PluginForumLikeForumListAdd(c echo.Context) error {
	uid := c.Get("uid").(string)

	bindings := new(PluginForumLikeListBinding)
	if err := c.Bind(bindings); err != nil {
		slog.Debug("plugin.forum-like.list.add", "error", err)
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "error", _function.EchoEmptyArray, "tbsign"))
	}

	// pid
	var tiebaAccount []int32

	if err := _function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Pluck("id", &tiebaAccount).Error; err != nil {
		slog.Error("plugin.forum-like.list.add.find", "uid", uid, "pid", bindings.Pid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyArray, "tbsign"))
	}

	if !slices.Contains(tiebaAccount, int32(bindings.Pid)) {
		return c.JSON(http.StatusNotFound, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyArray, "tbsign"))
	}

	numUID, _ := strconv.Atoi(uid)

	// forums
	numLimit, _ := strconv.Atoi(_function.GetOption("kd_forum_like_forum_limit_each_pid"))
	var pidCount int64
	_function.GormDB.R.Model(&model.TcKdForumLike{}).Where("uid = ? AND pid = ?", numUID, bindings.Pid).Count(&pidCount)

	if pidCount >= int64(numLimit) {
		return c.JSON(http.StatusForbidden, _function.ApiTemplate(403, "用户关注数量已达上限", _function.EchoEmptyArray, "tbsign"))
	}
	remainForumNum := numLimit - int(pidCount)

	fnameList := make(map[string]struct{}, len(bindings.Fname))
	var dataToInsert []*model.TcKdForumLike

	localExistsFnameList := pluginForumLikeListMerge(int32(numUID), int32(bindings.Pid))

	for _, fname := range bindings.Fname {
		fname = strings.TrimSpace(fname)

		if fname == "" {
			continue
		}

		if _, ok := fnameList[fname]; !ok {
			if slices.Contains(localExistsFnameList, fname) {
				continue
			}

			fnameList[fname] = struct{}{}
			dataToInsert = append(dataToInsert, &model.TcKdForumLike{
				UID:   int32(numUID),
				Pid:   int32(bindings.Pid),
				Fname: fname,
			})
			remainForumNum--
			if remainForumNum <= 0 {
				break
			}
		}
	}

	if len(dataToInsert) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _function.EchoEmptyArray, "tbsign"))
	}

	if err := _function.GormDB.W.Create(&dataToInsert).Error; err != nil {
		slog.Error("plugin.forum-like.list.add.create", "uid", uid, "pid", bindings.Pid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyArray, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", dataToInsert, "tbsign"))
}

type PluginForumLikeCloneBinding struct {
	Pid       uint64 `param:"pid"`
	SourcePid uint64 `param:"source_pid"`
	Source    string `param:"source"`
}

func PluginForumLikeForumListClone(c echo.Context) error {
	uid := c.Get("uid").(string)

	bindings := new(PluginForumLikeCloneBinding)
	if err := c.Bind(bindings); err != nil {
		slog.Debug("plugin.forum-like.list.add", "error", err)
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "error", _function.EchoEmptyArray, "tbsign"))
	}

	// source
	if !slices.Contains([]string{"plugin_tasks", "forum_list"}, bindings.Source) {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "error", _function.EchoEmptyArray, "tbsign"))
	}

	// pid
	var tiebaAccount []int32

	if err := _function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", uid).Pluck("id", &tiebaAccount).Error; err != nil {
		slog.Error("plugin.forum-like.list.clone.find", "uid", uid, "pid", bindings.Pid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyArray, "tbsign"))
	}

	if !slices.Contains(tiebaAccount, int32(bindings.Pid)) || !slices.Contains(tiebaAccount, int32(bindings.SourcePid)) {
		return c.JSON(http.StatusNotFound, _function.ApiTemplate(404, "用户不存在", _function.EchoEmptyArray, "tbsign"))
	}

	// forums
	numLimit, _ := strconv.Atoi(_function.GetOption("kd_forum_like_forum_limit_each_pid"))
	var pidCount int64
	_function.GormDB.R.Model(&model.TcKdForumLike{}).Where("uid = ? AND pid = ?", uid, bindings.Pid).Count(&pidCount)

	if pidCount >= int64(numLimit) {
		return c.JSON(http.StatusForbidden, _function.ApiTemplate(403, "用户关注数量已达上限", _function.EchoEmptyArray, "tbsign"))
	}
	remainForumNum := numLimit - int(pidCount)

	var dataToInsert []*model.TcKdForumLike

	numUID, _ := strconv.Atoi(uid)
	localExistsFnameList := pluginForumLikeListMerge(int32(numUID), int32(bindings.Pid))

	if bindings.Source == "plugin_tasks" {
		var sourceData []*model.TcKdForumLike
		_function.GormDB.R.Model(&model.TcKdForumLike{}).Where("uid = ? AND pid = ?", uid, bindings.SourcePid).Find(&sourceData)

		for _, data := range sourceData {
			if slices.Contains(localExistsFnameList, data.Fname) {
				continue
			}
			dataToInsert = append(dataToInsert, &model.TcKdForumLike{
				UID:   int32(numUID),
				Pid:   int32(bindings.Pid),
				Fid:   data.Fid,
				Fname: data.Fname,
			})
			remainForumNum--
			if remainForumNum <= 0 {
				break
			}
		}
	} else {
		var sourceData []*model.TcTieba
		_function.GormDB.R.Model(&model.TcTieba{}).Where("uid = ? AND pid = ?", uid, bindings.SourcePid).Find(&sourceData)

		var sourceForumMap = make(map[string]struct{}, len(sourceData))
		for _, data := range sourceData {
			if slices.Contains(localExistsFnameList, data.Tieba) {
				continue
			} else if _, ok := sourceForumMap[data.Tieba]; ok {
				continue
			}

			dataToInsert = append(dataToInsert, &model.TcKdForumLike{
				UID:   int32(numUID),
				Pid:   int32(bindings.Pid),
				Fid:   data.Fid,
				Fname: data.Tieba,
			})
			sourceForumMap[data.Tieba] = struct{}{}
			remainForumNum--
			if remainForumNum <= 0 {
				break
			}
		}
	}

	if len(dataToInsert) == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _function.EchoEmptyArray, "tbsign"))
	}

	if err := _function.GormDB.W.Create(&dataToInsert).Error; err != nil {
		slog.Error("plugin.forum-like.list.clone.create", "uid", uid, "pid", bindings.Pid, "source", bindings.Source, "source-pid", bindings.SourcePid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "未知错误", _function.EchoEmptyArray, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", dataToInsert, "tbsign"))
}

func PluginForumLikeForumConfig(c echo.Context) error {
	numLimit, _ := strconv.Atoi(_function.GetOption("kd_forum_like_forum_limit_each_pid"))
	pidCooldownTime, _ := strconv.Atoi(_function.GetOption("kd_forum_like_cooldown_time_pid"))
	fnameCooldownTime, _ := strconv.Atoi(_function.GetOption("kd_forum_like_cooldown_time_fname"))

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]int{
		"limit":               numLimit,
		"pid_cooldown_time":   pidCooldownTime,
		"fname_cooldown_time": fnameCooldownTime,
	}, "tbsign"))
}

type PluginForumLikeForumListDeleteParams struct {
	Pid uint64 `param:"pid"`
	Tid uint64 `param:"tid"`
}

func PluginForumLikeForumListDelete(c echo.Context) error {
	uid := c.Get("uid").(string)

	params := new(PluginForumLikeForumListDeleteParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "error", false, "tbsign"))
	}

	affectedRows := []*model.TcKdForumLike{}

	err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		publicTx := tx.Model(&model.TcKdForumLike{}).Where("uid = ?", uid)

		if params.Pid > 0 {
			publicTx = publicTx.Where("pid = ?", params.Pid)
		}

		if params.Tid > 0 {
			publicTx = publicTx.Where("id = ?", params.Tid)
		}

		if err := publicTx.Find(&affectedRows).Error; err != nil {
			return err
		}

		if err := publicTx.Delete(&model.TcKdForumLike{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		slog.Error("plugin.forum-like.list.delete", "uid", uid, "pid", params.Pid, "tid", params.Tid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "error", _function.EchoEmptyArray, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", affectedRows, "tbsign"))
}

type PluginForumLikeForumListDeleteSucceedParams struct {
	Pid uint64 `param:"pid"`
}

func PluginForumLikeForumListDeleteSucceed(c echo.Context) error {
	uid := c.Get("uid").(string)

	params := new(PluginForumLikeForumListDeleteSucceedParams)
	if err := c.Bind(params); err != nil {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "error", false, "tbsign"))
	}

	affectedRows := []*model.TcKdForumLike{}

	err := _function.GormDB.W.Transaction(func(tx *gorm.DB) error {
		publicTx := tx.Model(&model.TcKdForumLike{}).Where("uid = ?", uid)

		if params.Pid > 0 {
			publicTx = publicTx.Where("pid = ?", params.Pid)
		}

		publicTx = publicTx.Where("date > ? AND status = ? AND last_error LIKE ?", 0, 0, "关注成功%")

		if err := publicTx.Find(&affectedRows).Error; err != nil {
			return err
		}

		if err := publicTx.Delete(&model.TcKdForumLike{}).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		slog.Error("plugin.forum-like.list.delete-succeed", "uid", uid, "pid", params.Pid, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "error", _function.EchoEmptyArray, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", affectedRows, "tbsign"))
}
