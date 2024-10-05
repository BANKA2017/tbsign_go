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
)

func init() {
	RegisterPlugin(UserGrowthTasksPlugin.Name, UserGrowthTasksPlugin)
}

type UserGrowthTasksPluginType struct {
	PluginInfo
}

var UserGrowthTasksPlugin = _function.VariablePtrWrapper(UserGrowthTasksPluginType{
	PluginInfo{
		Name:              "kd_growth",
		PluginNameCN:      "用户成长任务",
		PluginNameCNShort: "成长任务",
		PluginNameFE:      "user_growth_tasks",
		Version:           "0.1",
		Options: map[string]string{
			"kd_growth_offset":       "0",
			"kd_growth_action_limit": "50",
		},
		SettingOptions: map[string]PluinSettingOption{
			"kd_growth_action_limit": {
				OptionName:   "kd_growth_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: func(value string) bool {
					numLimit, err := strconv.ParseInt(value, 10, 64)
					return err == nil && numLimit >= 0
				},
			},
		},
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "settings", Function: PluginGrowthTasksGetSettings},
			{Method: "PUT", Path: "settings", Function: PluginGrowthTasksSetSettings},
			{Method: "GET", Path: "list", Function: PluginGrowthTasksGetList},
			{Method: "PATCH", Path: "list", Function: PluginGrowthTasksAddAccount},
			{Method: "DELETE", Path: "list/:id", Function: PluginGrowthTasksDelAccount},
			{Method: "POST", Path: "list/empty", Function: PluginGrowthTasksDelAllAccounts},
			{Method: "GET", Path: "status/:pid", Function: PluginGrowthTasksGetTasksStatus},
		},
	},
})

var UserGrowthTasksBreakList = []string{"open_push_switch"}

type UserGrowthTasksWebResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
}

type UserGrowthTasksClientResponse struct {
	ServerTime string `json:"server_time,omitempty"`
	Time       int    `json:"time,omitempty"`
	Ctime      int    `json:"ctime,omitempty"`
	Logid      int64  `json:"logid,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_msg,omitempty"`
	Info       []any  `json:"info,omitempty"`
}

type LevelInfo struct {
	Level          int    `json:"level,omitempty"`
	Name           string `json:"name,omitempty"`
	GrowthValue    int    `json:"growth_value,omitempty"`
	NextLevelValue int    `json:"next_level_value,omitempty"`
	Status         int    `json:"status,omitempty"`
	IsCurrent      int    `json:"is_current,omitempty"`
}

type UserGrowthTasksListResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		User struct {
			UserID     int    `json:"user_id,omitempty"`
			Uname      string `json:"uname,omitempty"`
			Portrait   string `json:"portrait,omitempty"`
			IsTiebaVip bool   `json:"is_tieba_vip,omitempty"`
		} `json:"user,omitempty"`
		/// LevelInfo []LevelInfo `json:"level_info,omitempty"`
		TabList []struct {
			TabName      string `json:"tab_name,omitempty"`
			Name         string `json:"name,omitempty"`
			Text         string `json:"text,omitempty"`
			TaskTypeList []struct {
				TaskType string           `json:"task_type,omitempty"`
				TaskList []UserGrowthTask `json:"task_list,omitempty"`
			} `json:"task_type_list,omitempty"`
		} `json:"tab_list,omitempty"`
		Tbs string `json:"tbs,omitempty"`
	} `json:"data,omitempty"`
}

type UserGrowthTask struct {
	ID      int    `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	ActType string `json:"act_type,omitempty"`
	URL     string `json:"url,omitempty"`
	Detail  string `json:"detail,omitempty"`
	Exp     int    `json:"exp,omitempty"`
	Current int    `json:"current,omitempty"`
	Total   int    `json:"total,omitempty"`
	//Status             int    `json:"status,omitempty"`
	SortStatus   int `json:"sort_status,omitempty"`
	CompleteTime int `json:"complete_time,omitempty"`
	StartTime    int `json:"start_time,omitempty"`
	ExpireTime   int `json:"expire_time,omitempty"`
	// MinLevel     int `json:"min_level,omitempty"`
	// TaskDoneNum  int   `json:"task_done_num,omitempty"`
	// TaskThreadID []any `json:"task_thread_id,omitempty"`
	// TargetKw           string `json:"target_kw,omitempty"`
	// TargetScheme       string `json:"target_scheme,omitempty"`
	// TargetChatroomName string `json:"target_chatroom_name,omitempty"`
	// TargetChatroomID   int    `json:"target_chatroom_id,omitempty"`
}

type UserGrowthTaskToSave struct {
	Name    string `json:"name"`
	ActType string `json:"act_type"`
	Status  int    `json:"status"`
	Msg     string `json:"msg"`
}

type UserGrowthTaskCollectStampResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
}

func PostGrowthTaskByWeb(cookie _type.TypeCookie, task string) (*UserGrowthTasksWebResponse, error) {
	_body := url.Values{}
	_body.Set("tbs", cookie.Tbs)
	_body.Set("act_type", task)
	_body.Set("cuid", "-")

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/usergrowth/commitUGTaskInfo", "POST", []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksWebResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

// share_thread page_sign
func PostGrowthTaskByClient(cookie _type.TypeCookie, task string) (*UserGrowthTasksClientResponse, error) {
	form := map[string]string{
		"BDUSS":    cookie.Bduss,
		"act_type": task,
		"cuid":     "-",
		"tbs":      cookie.Tbs,
	}
	_function.AddSign(&form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	response, err := _function.TBFetch("https://tiebac.baidu.com/c/c/user/commitUGTaskInfo", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksClientResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

func PostCollectStamp(cookie _type.TypeCookie, task_id int) (*UserGrowthTaskCollectStampResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}
	_body := url.Values{
		"type":     {"3"}, // why 3?
		"task_id":  {strconv.Itoa(task_id)},
		"act_type": {"active"},
		"tbs":      {cookie.Tbs},
		"cuid":     {"-"},
	}
	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/icon/collectStamp", "POST", []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTaskCollectStampResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

func GetUserGrowthTasksList(cookie _type.TypeCookie) (*UserGrowthTasksListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/usergrowth/showUserGrowth", "GET", nil, headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksListResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

var activeTasks = []string{"daily_task", "live_task"}

// TODO redo growth tasks(?)
func (pluginInfo *UserGrowthTasksPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id, err := strconv.ParseInt(_function.GetOption("kd_growth_offset"), 10, 64)
	if err != nil {
		id = 0
	}
	// status list
	var accountStatusList = make(map[int64]string)
	// cookie list
	var accountCookiesList = make(map[int64]_type.TypeCookie)

	// get list
	todayBeginning := _function.LocaleTimeDiff(0) //GMT+8
	kdGrowthTasksUserList := &[]model.TcKdGrowth{}

	limit := _function.GetOption("kd_growth_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("date < ? AND id > ?", todayBeginning, id).Limit(int(numLimit)).Find(&kdGrowthTasksUserList)
	for _, taskUserItem := range *kdGrowthTasksUserList {
		if _, ok := accountStatusList[taskUserItem.UID]; !ok {
			accountStatusList[taskUserItem.UID] = _function.GetUserOption("kd_growth_sign_only", strconv.Itoa(int(taskUserItem.UID)))
		}
		if accountStatusList[taskUserItem.UID] == "" {
			// check uid is exists
			var accountInfo model.TcBaiduid
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", taskUserItem.UID).First(&accountInfo)
			if accountInfo.Portrait == "" {
				// clean
				_function.GormDB.W.Where("uid = ?", taskUserItem.UID).Delete(&model.TcKdGrowth{})
				accountStatusList[taskUserItem.UID] = "NOT_EXISTS"
				continue
			}
			// auto set -> 1
			accountStatusList[taskUserItem.UID] = "1"
			_function.SetUserOption("kd_growth_sign_only", "1", strconv.Itoa(int(taskUserItem.UID)))
		}

		if _, ok := accountCookiesList[taskUserItem.Pid]; !ok {
			accountCookiesList[taskUserItem.Pid] = _function.GetCookie(int32(taskUserItem.Pid))
		}
		cookie := accountCookiesList[taskUserItem.Pid]
		var tasksList []UserGrowthTask
		var result []UserGrowthTaskToSave
		doCollectStampTasks := false

		/// levelInfo := LevelInfo{}

		if accountStatusList[taskUserItem.UID] == "1" {
			tasksResponse, err := GetUserGrowthTasksList(cookie)
			if err != nil {
				log.Println(err)
				log.Println("user_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "Unable to fetch tasks list")
				//continue
			}

			/// // find level info
			/// for _, levelInfoItem := range tasksResponse.Data.LevelInfo {
			/// 	if levelInfoItem.IsCurrent == 1 {
			/// 		levelInfo = levelInfoItem
			/// 		break
			/// 	}
			/// }

			for _, taskTypeListList := range tasksResponse.Data.TabList {
				if taskTypeListList.TabName == "basic" {
					for _, taskTypeList := range taskTypeListList.TaskTypeList {
						if slices.Contains(activeTasks, taskTypeList.TaskType) {
							tasksList = append(tasksList, taskTypeList.TaskList...)
						}
						if taskTypeList.TaskType == "icon_task" && slices.Contains([]string{"0", ""}, _function.GetUserOption("kd_growth_break_icon_tasks", strconv.Itoa(int(taskUserItem.UID)))) {
							for _, iconTaskItem := range taskTypeList.TaskList {
								if iconTaskItem.SortStatus == 0 {
									postCollectStampRES, err := PostCollectStamp(cookie, iconTaskItem.ID)
									if err != nil {
										result = append(result, UserGrowthTaskToSave{
											Name:    iconTaskItem.Name,
											ActType: iconTaskItem.ActType,
											Status:  0,
											Msg:     "failed",
										})
									} else {
										if postCollectStampRES.No == 0 {
											result = append(result, UserGrowthTaskToSave{
												Name:    iconTaskItem.Name,
												ActType: iconTaskItem.ActType,
												Status:  1,
												Msg:     "success",
											})
										} else {
											result = append(result, UserGrowthTaskToSave{
												Name:    iconTaskItem.Name,
												ActType: iconTaskItem.ActType,
												Status:  0,
												Msg:     postCollectStampRES.Error,
											})
										}
									}
									doCollectStampTasks = true
								} else if iconTaskItem.SortStatus == 1 {
									doCollectStampTasks = true
								}
							}
						}
					}
				}
			}
		} else {
			tasksList = append(tasksList, UserGrowthTask{
				Name:       "每日签到",
				ActType:    "page_sign",
				SortStatus: 1,
				ExpireTime: 0,
			})
		}

		for _, task := range tasksList {
			if task.SortStatus == -1 || slices.Contains(UserGrowthTasksBreakList, task.ActType) {
				continue
			} else if task.SortStatus == 2 {
				result = append(result, UserGrowthTaskToSave{
					Name:    task.Name,
					ActType: task.ActType,
					Status:  1,
					Msg:     "success",
				})
			} else if task.SortStatus == 1 && (task.ExpireTime == 0 || task.ExpireTime > int(_function.Now.Unix())) {
				response, err := PostGrowthTaskByWeb(cookie, task.ActType)
				if err != nil {
					result = append(result, UserGrowthTaskToSave{
						Name:    task.Name,
						ActType: task.ActType,
						Status:  0,
						Msg:     "failed",
					})
				} else {
					if response.No == 0 {
						result = append(result, UserGrowthTaskToSave{
							Name:    task.Name,
							ActType: task.ActType,
							Status:  1,
							Msg:     "success",
						})
					} else {
						result = append(result, UserGrowthTaskToSave{
							Name:    task.Name,
							ActType: task.ActType,
							Status:  0,
							Msg:     response.Error,
						})
					}
				}
			}
		}

		// do sync
		if doCollectStampTasks {
			_, err := _function.PostSync(cookie)
			if err != nil {
				result = append(result, UserGrowthTaskToSave{
					Name:    "签到类集章任务",
					ActType: "active",
					Status:  0,
					Msg:     "failed",
				})
			} else {
				result = append(result, UserGrowthTaskToSave{
					Name:    "签到类集章任务",
					ActType: "active",
					Status:  1,
					Msg:     "success",
				})
			}
		}

		if len(result) > 0 {
			jsonResult, _ := _function.JsonEncode(result)
			tmpLog := ""
			for i, r := range result {
				if i > 0 {
					tmpLog += ","
				}
				tmpLog += r.ActType + ":" + strconv.Itoa(r.Status)
			}

			log.Println("user_tasks:", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, string(jsonResult))

			// previous logs
			previousLogs := []string{}
			for i, s := range strings.Split(taskUserItem.Log, "<br/>") {
				if i <= 28 {
					previousLogs = append(previousLogs, s)
				} else {
					break
				}
			}

			_function.GormDB.W.Model(&model.TcKdGrowth{}).Where("id = ?", taskUserItem.ID).Updates(model.TcKdGrowth{
				Status: string(jsonResult),
				Log:    fmt.Sprintf("%s: %s<br/>%s", _function.Now.Local().Format(time.DateOnly), tmpLog, strings.Join(previousLogs, "<br/>")),
				Date:   int32(_function.Now.Unix()),
			})
		}

		_function.SetOption("kd_growth_offset", strconv.Itoa(int(taskUserItem.ID)))
	}
	_function.SetOption("kd_growth_offset", "0")
}

func (pluginInfo *UserGrowthTasksPluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcKdGrowth{})
		_function.GormDB.W.Exec("ALTER TABLE `tc_kd_growth` ADD UNIQUE KEY `id_uid_pid` (`id`,`uid`,`pid`), ADD KEY `uid` (`uid`), ADD KEY `pid` (`pid`), ADD KEY `date_id` (`date`,`id`) USING BTREE;")
	} else {
		_function.GormDB.W.Migrator().CreateTable(&model.TcKdGrowth{})

		_function.GormDB.W.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_kd_growth_id_uid_pid" ON "tc_kd_growth" ("id","uid","pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_date_id" ON "tc_kd_growth" ("date","id");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_pid" ON "tc_kd_growth" ("pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_growth_uid" ON "tc_kd_growth" ("uid");`)
	}
	return nil
}

func (pluginInfo *UserGrowthTasksPluginType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)
	_function.GormDB.W.Migrator().DropTable(&model.TcKdGrowth{})

	// user options
	_function.GormDB.W.Where("name = ?", "kd_growth_sign_only").Delete(&model.TcUsersOption{})
	_function.GormDB.W.Where("name = ?", "kd_growth_break_icon_tasks").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *UserGrowthTasksPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *UserGrowthTasksPluginType) RemoveAccount(_type string, id int32) error {
	_function.GormDB.W.Where("? = ?", _type, id).Delete(&model.TcKdGrowth{})
	return nil
}

func (pluginInfo *UserGrowthTasksPluginType) Ext() ([]any, error) {
	return []any{}, nil
}

// endpoints

func PluginGrowthTasksGetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	// sign only
	signOnly := _function.GetUserOption("kd_growth_sign_only", uid)
	if signOnly == "" {
		signOnly = "0"
		_function.SetUserOption("kd_growth_sign_only", signOnly, uid)
	}

	// no icon tasks
	noIconTasks := _function.GetUserOption("kd_growth_break_icon_tasks", uid)
	if noIconTasks == "" {
		noIconTasks = "0"
		_function.SetUserOption("kd_growth_break_icon_tasks", noIconTasks, uid)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"sign_only":        signOnly,
		"break_icon_tasks": noIconTasks,
	}, "tbsign"))
}

func PluginGrowthTasksSetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	signOnly := c.FormValue("sign_only") != "0"
	noIconTasks := c.FormValue("break_icon_tasks") != "0"

	_function.SetUserOption("kd_growth_sign_only", signOnly, uid)
	_function.SetUserOption("kd_growth_break_icon_tasks", noIconTasks, uid)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
	}, "tbsign"))
}

func PluginGrowthTasksGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accounts []model.TcKdGrowth
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&accounts)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", accounts, "tbsign"))
}

func PluginGrowthTasksAddAccount(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("uid = ? AND pid = ?", uid, numPid).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "账号已存在", _function.EchoEmptyObject, "tbsign"))
	} else {
		dataToInsert := model.TcKdGrowth{
			UID:  numUID,
			Pid:  numPid,
			Date: 0,
		}
		_function.GormDB.W.Create(&dataToInsert)
		_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("uid = ? AND pid = ?", uid, numPid).First(&dataToInsert)
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", dataToInsert, "tbsign"))
	}
}

func PluginGrowthTasksDelAccount(c echo.Context) error {
	uid := c.Get("uid").(string)

	id := c.Param("id")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无效任务 id", map[string]any{
			"success": false,
			"id":      id,
		}, "tbsign"))
	}

	_function.GormDB.W.Model(&model.TcKdGrowth{}).Delete(&model.TcKdGrowth{
		UID: numUID,
		ID:  numID,
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginGrowthTasksDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.W.Model(&model.TcKdGrowth{}).Delete(&model.TcKdGrowth{
		UID: numUID,
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func PluginGrowthTasksGetTasksStatus(c echo.Context) error {
	uid := c.Get("uid").(string)
	pid := c.Param("pid")

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).Count(&count)

	if count > 0 {
		numPid, _ := strconv.ParseInt(pid, 10, 64)
		status, err := GetUserGrowthTasksList(_function.GetCookie(int32(numPid)))
		if err != nil {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, "获取任务列表失败", _function.EchoEmptyObject, "tbsign"))
		} else if status.No != 0 {
			return c.JSON(http.StatusOK, _function.ApiTemplate(500, status.Error, _function.EchoEmptyObject, "tbsign"))
		}
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status.Data, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}
}
