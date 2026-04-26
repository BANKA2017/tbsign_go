package _plugin

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
	"gorm.io/gorm"
)

func init() {
	PluginList.Register(UserGrowthTasksPlugin)
}

type UserGrowthTasksPluginType struct {
	PluginInfo
}

var UserGrowthTasksPlugin = _function.VPtr(UserGrowthTasksPluginType{
	PluginInfo{
		Name:              "kd_growth",
		PluginNameCN:      "用户成长任务",
		PluginNameCNShort: "成长任务",
		PluginNameFE:      "user_growth_tasks",
		Version:           "0.2",
		Options: map[string]string{
			"kd_growth_offset":       "0",
			"kd_growth_action_limit": "50",
			// "kd_growth_client_version": "12.84.3.0",
		},
		SettingOptions: map[string]PluginSettingOption{
			"kd_growth_action_limit": {
				OptionName:   "kd_growth_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
			// "kd_growth_client_version": {
			// 	OptionName:   "kd_growth_client_version",
			// 	OptionNameCN: "客户端版本号",
			// 	Validate: func(value string) bool {
			// 		parts := strings.Split(value, ".")
			// 		if len(parts) == 0 {
			// 			return false
			// 		}
			// 		for _, p := range parts {
			// 			if p == "" {
			// 				return false
			// 			}
			// 			if _, err := strconv.Atoi(p); err != nil {
			// 				return false
			// 			}
			// 		}
			// 		return true
			// 	},
			// },
		},
		Endpoints: []PluginEndpointStruct{
			{Method: http.MethodGet, Path: "settings", Function: PluginGrowthTasksGetSettings},
			{Method: http.MethodPut, Path: "settings", Function: PluginGrowthTasksSetSettings},
			{Method: http.MethodGet, Path: "list", Function: PluginGrowthTasksGetList},
			{Method: http.MethodPatch, Path: "list", Function: PluginGrowthTasksAddAccount},
			{Method: http.MethodDelete, Path: "list/:id", Function: PluginGrowthTasksDelAccount},
			{Method: http.MethodPost, Path: "list/empty", Function: PluginGrowthTasksDelAllAccounts},
			{Method: http.MethodGet, Path: "status/:pid", Function: PluginGrowthTasksGetTasksStatus},
		},
	},
})

var UserGrowthTasksPluginClientVersion = "22.4.1.0"

var activeTasks = []string{"daily_task", "live_task", "exchange_flow_task"}
var UserGrowthTasksBreakList = []string{"open_push_switch"}
var UserGrowthTasksExchangeFlowTaskIDs = []int{591}

type UserGrowthTasksWebResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
}

type UserGrowthTasksClientResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	Data  struct {
		Toast          json.RawMessage `json:"toast"` // they use PHP, so empty array `[]` means do nothing, an object and `success_task_ids` not empty means success
		SuccessTaskIds []int           `json:"success_task_ids"`
	} `json:"data"`
}

type LevelInfo struct {
	Level          int    `json:"level"`
	Name           string `json:"name"`
	GrowthValue    int    `json:"growth_value"`
	NextLevelValue int    `json:"next_level_value"`
	Status         int    `json:"status"`
	IsCurrent      int    `json:"is_current"`
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
		LevelInfo []LevelInfo `json:"level_info,omitempty"`
		Tmoney    struct {
			Current int `json:"current,omitempty"`
		} `json:"tmoney,omitempty"`
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
	TargetScheme string `json:"target_scheme,omitempty"`
	// TargetChatroomName string `json:"target_chatroom_name,omitempty"`
	// TargetChatroomID   int    `json:"target_chatroom_id,omitempty"`
}

type UserGrowthTaskToSave struct {
	TaskID  int    `json:"task_id"`
	Name    string `json:"name"`
	ActType string `json:"act_type"`
	Status  int    `json:"status"`
	Msg     string `json:"msg"`
}

type UserGrowthTaskCollectStampResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
}

func PostUserGrowth(cookie *_type.TypeCookie, task string) (*UserGrowthTasksClientResponse, error) {
	_body := url.Values{}
	_body.Set("tbs", cookie.Tbs)
	_body.Set("act_type", task)
	_body.Set("cuid", "-")
	// _body.Set("subapp_type", "hybrid") // NEVER SET HYBRID IN WEB ENDPOINT!!!

	if task == "page_sign" {
		_body.Set("scene_name", "taskCenter")
	}

	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"User-Agent": "tieba/" + UserGrowthTasksPluginClientVersion,
	}

	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/usergrowth/commitUGTaskInfo", http.MethodPost, []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksClientResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

// share_thread page_sign
func PostUserTask(cookie *_type.TypeCookie, task string, taskID int) (*UserGrowthTasksClientResponse, error) {
	form := map[string]string{
		"act_type":        task,
		"cuid":            "-",
		"tbs":             cookie.Tbs,
		"_client_version": UserGrowthTasksPluginClientVersion,

		"subapp_type": "hybrid",
	}

	if task == "task_entry_page" && taskID > 0 {
		form["act_data[task_id]"] = strconv.Itoa(taskID)
	}

	_function.ClientTypeFallBack(form, "android")
	_function.AddSign(form, "android")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/usertask/commitUGTaskInfo", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), map[string]string{
		"Cookie":      "BDUSS=" + cookie.Bduss,
		"Subapp-Type": "hybrid",
	})

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksClientResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

func PostUserTaskInfoWidget(cookie *_type.TypeCookie) (any, error) {
	_body := url.Values{
		"BDUSS":           {cookie.Bduss},
		"push_switch":     {"1"},
		"cuid":            {"-"},
		"_client_version": {UserGrowthTasksPluginClientVersion},
		"_client_type":    {"2"},
	}

	taskInfoResponse, err := _function.TBFetch("https://tiebac.baidu.com/c/f/widget/getUserTaskInfo", http.MethodPost, []byte(_body.Encode()), map[string]string{
		"User-Agent": "TiebaWidgets/" + UserGrowthTasksPluginClientVersion + " CFNetwork/3826.500.131 Darwin/24.5.0",
	})

	if err != nil {
		return nil, err
	}

	return string(taskInfoResponse), err
}

func PostCollectStamp(cookie *_type.TypeCookie, task_id int) (*UserGrowthTaskCollectStampResponse, error) {
	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"User-Agent": "tieba/" + UserGrowthTasksPluginClientVersion,
	}
	_body := url.Values{
		"type":     {"3"}, // why 3?
		"task_id":  {strconv.Itoa(task_id)},
		"act_type": {"active"},
		"tbs":      {cookie.Tbs},
		"cuid":     {"-"},
	}
	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/icon/collectStamp", http.MethodPost, []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTaskCollectStampResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

// {"no":110003,"error":"fail to call service","data":{}}
func GetUserGrowthTasksList(cookie *_type.TypeCookie) (*UserGrowthTasksListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	response, err := _function.TBFetch("https://tieba.baidu.com/mo/q/usergrowth/showUserGrowth?client_type=2&client_version="+UserGrowthTasksPluginClientVersion, http.MethodGet, nil, headersMap)

	if err != nil {
		return nil, err
	}

	resp := new(UserGrowthTasksListResponse)
	err = _function.JsonDecode(response, &resp)
	return resp, err
}

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
	var accountStatusList = make(map[int32]string)
	// cookie list
	// var accountCookiesList = make(map[int32]_type.TypeCookie)
	var extTasksList = make(map[int32]map[string]string)

	// get list
	todayBeginning := _function.LocaleTimeDiff(0) //GMT+8
	var kdGrowthTasksUserList []*model.TcKdGrowth

	limit := _function.GetOption("kd_growth_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("date < ? AND id > ?", todayBeginning, id).Limit(int(numLimit)).Find(&kdGrowthTasksUserList)
	for _, taskUserItem := range kdGrowthTasksUserList {
		if _, ok := accountStatusList[taskUserItem.UID]; !ok {
			accountStatusList[taskUserItem.UID] = _function.GetUserOption("kd_growth_sign_only", strconv.Itoa(int(taskUserItem.UID)))
		}
		if accountStatusList[taskUserItem.UID] == "" {
			// check uid is exists
			var accountInfo model.TcBaiduid
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", taskUserItem.UID).Take(&accountInfo)
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

		// cookies
		cookie := _function.GetCookie(int32(taskUserItem.Pid))
		var result []UserGrowthTaskToSave

		if !cookie.IsLogin {
			result = append(result, UserGrowthTaskToSave{
				TaskID:  -1,
				Name:    "未登录",
				ActType: "login_failed",
				Status:  0,
				Msg:     "failed",
			})
		} else {
			var tasksList = make(map[string]UserGrowthTask)
			doCollectStampTasks := false

			/// levelInfo := LevelInfo{}
			if accountStatusList[taskUserItem.UID] == "2" {
				// ext tasks
				if extTasks, ok := extTasksList[taskUserItem.UID]; !ok {
					if err := _function.JsonDecode([]byte(_function.GetUserOption("kd_growth_ext_tasks", strconv.Itoa(int(taskUserItem.UID)))), &extTasks); err != nil {
						extTasksList[taskUserItem.UID] = make(map[string]string)
					} else {
						extTasksList[taskUserItem.UID] = extTasks
					}
				}
				extTasks := extTasksList[taskUserItem.UID]

				if len(extTasks) > 0 {
					for actType, taskName := range extTasks {
						if actType != "" && taskName != "" {
							var taskID int

							// acttype:123
							// acttype
							if strings.Contains(actType, ":") {
								actTypeSplit := strings.Split(actType, ":")
								actType = actTypeSplit[0]
								strTaskID := actTypeSplit[1]
								taskID, _ = strconv.Atoi(strTaskID)
							}

							// remove task id now
							tasksList[actType] = UserGrowthTask{
								ID:         taskID,
								Name:       taskName,
								ActType:    actType,
								SortStatus: 1,
								ExpireTime: 0,
							}
						}
					}
				}
			}

			if accountStatusList[taskUserItem.UID] != "0" {
				tasksResponse, err := GetUserGrowthTasksList(cookie)
				if err != nil {
					slog.Error("plugin.user-growth-tasks.get-list", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "error", err)
					continue
				} else if tasksResponse.No != 0 {
					slog.Error("plugin.user-growth-tasks.get-list", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "code", tasksResponse.No, "error", tasksResponse.Error)
					continue
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
								if taskTypeList.TaskType == "exchange_flow_task" {
									for _, taskItem := range taskTypeList.TaskList {
										if slices.Contains(UserGrowthTasksExchangeFlowTaskIDs, taskItem.ID) {
											tasksList[taskItem.ActType+strconv.Itoa(taskItem.ID)] = taskItem
										} else if _, exist := tasksList[taskItem.ActType]; exist {
											tasksList[taskItem.ActType] = taskItem // replace ext task
										}
									}
								} else {
									for _, taskItem := range taskTypeList.TaskList {
										tasksList[taskItem.ActType] = taskItem
									}
								}
							}
							if taskTypeList.TaskType == "icon_task" && slices.Contains([]string{"0", ""}, _function.GetUserOption("kd_growth_break_icon_tasks", strconv.Itoa(int(taskUserItem.UID)))) {
								for _, iconTaskItem := range taskTypeList.TaskList {
									switch iconTaskItem.SortStatus {
									case 0:
										postCollectStampRES, err := PostCollectStamp(cookie, iconTaskItem.ID)
										if err != nil {
											result = append(result, UserGrowthTaskToSave{
												TaskID:  iconTaskItem.ID,
												Name:    iconTaskItem.Name,
												ActType: iconTaskItem.ActType,
												Status:  0,
												Msg:     "failed",
											})
										} else {
											if postCollectStampRES.No == 0 {
												result = append(result, UserGrowthTaskToSave{
													TaskID:  iconTaskItem.ID,
													Name:    iconTaskItem.Name,
													ActType: iconTaskItem.ActType,
													Status:  1,
													Msg:     "success",
												})
											} else {
												result = append(result, UserGrowthTaskToSave{
													TaskID:  iconTaskItem.ID,
													Name:    iconTaskItem.Name,
													ActType: iconTaskItem.ActType,
													Status:  0,
													Msg:     postCollectStampRES.Error,
												})
											}
										}
										doCollectStampTasks = true
									case 1:
										doCollectStampTasks = true
									}
								}
							}
						}
					}
				}
			} else {
				tasksList["page_sign"] = UserGrowthTask{
					ID:         20,
					Name:       "每日签到",
					ActType:    "page_sign",
					SortStatus: 1,
					ExpireTime: 0,
				}
			}

			for _, task := range tasksList {
				if task.SortStatus == -1 || slices.Contains(UserGrowthTasksBreakList, task.ActType) {
					continue
				} else if task.SortStatus == 2 {
					result = append(result, UserGrowthTaskToSave{
						TaskID:  task.ID,
						Name:    task.Name,
						ActType: task.ActType,
						Status:  1,
						Msg:     "success",
					})
				} else if task.SortStatus == 1 && (task.ExpireTime == 0 || task.ExpireTime > int(time.Now().Unix())) {
					response := new(UserGrowthTasksClientResponse)

					if slices.Contains(UserGrowthTasksExchangeFlowTaskIDs, task.ID) {
						switch task.ID {
						case 591:
							var res *growthTasks591ExecuteResponse
							res, err = growthTasks591(task.TargetScheme)
							if err == nil {
								response.No = res.Errno
								response.Error = res.Errmsg
								if res.Errno == 0 {
									response.Data.SuccessTaskIds = []int{task.ID}
								}
							}
						// case 563:
						// 	response, err = PostGrowthTaskByClient(cookie, task.ActType, task.ID)
						default:
							err = fmt.Errorf("unknown exchange_flow_task name: %s, act_type: %s, task_id: %d", task.Name, task.ActType, task.ID)
						}
					} else if task.ActType == "page_sign" || task.ID == 0 {
						response, err = PostUserGrowth(cookie, task.ActType)
					} else {
						response, err = PostUserTask(cookie, task.ActType, task.ID)
					}

					// {"no":110003,"error":"fail to call service","data":"\u4efb\u52a1\u5931\u8d25"}
					// WTF string in struct
					if err != nil {
						if _function.JsonIsUnmarshalTypeError(err) && response.No > 0 {
							slog.Error("plugin.user-growth-tasks.action", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "task_id", task.ID, "task_name", task.Name, "act_type", task.ActType, "code", response.No, "error", response.Error)
							result = append(result, UserGrowthTaskToSave{
								TaskID:  task.ID,
								Name:    task.Name,
								ActType: task.ActType,
								Status:  0,
								Msg:     response.Error,
							})
						} else {
							slog.Error("plugin.user-growth-tasks.action", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "task_id", task.ID, "task_name", task.Name, "act_type", task.ActType, "error", err)
							result = append(result, UserGrowthTaskToSave{
								TaskID:  task.ID,
								Name:    task.Name,
								ActType: task.ActType,
								Status:  0,
								Msg:     "failed",
							})
						}
					} else {
						if response.No == 0 {
							if task.ID == 0 || len(response.Data.Toast) > 2 {
								result = append(result, UserGrowthTaskToSave{
									TaskID:  task.ID,
									Name:    task.Name,
									ActType: task.ActType,
									Status:  1,
									Msg:     "success",
								})
							} else {
								slog.Error("plugin.user-growth-tasks.action", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "task_id", task.ID, "task_name", task.Name, "act_type", task.ActType, "error", "received, but not successfully")
								result = append(result, UserGrowthTaskToSave{
									TaskID:  task.ID,
									Name:    task.Name,
									ActType: task.ActType,
									Status:  0,
									Msg:     "received, but not successfully",
								})
							}
						} else {
							slog.Error("plugin.user-growth-tasks.action", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "task_id", task.ID, "task_name", task.Name, "act_type", task.ActType, "code", response.No, "error", response.Error)
							result = append(result, UserGrowthTaskToSave{
								TaskID:  task.ID,
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
						Name:    "印记任务签到",
						ActType: "active",
						Status:  0,
						Msg:     "failed",
					})
				} else {
					result = append(result, UserGrowthTaskToSave{
						Name:    "印记任务签到",
						ActType: "active",
						Status:  1,
						Msg:     "success",
					})
				}
			}
		}

		if len(result) > 0 {
			jsonResult, _ := _function.JsonEncode(result)
			var tmpLog strings.Builder
			for i, r := range result {
				if i > 0 {
					tmpLog.WriteString(",")
				}
				tmpLog.WriteString(strconv.Itoa(r.TaskID) + ":" + r.Name + ":" + r.ActType + ":" + strconv.Itoa(r.Status))
			}

			slog.Debug("plugin.user-growth-tasks.action", "id", taskUserItem.ID, "pid", taskUserItem.Pid, "uid", taskUserItem.UID, "result", string(jsonResult))

			// previous logs
			previousLogs := []string{}
			for i, s := range strings.Split(taskUserItem.Log, "<br/>") {
				if i <= 30 {
					previousLogs = append(previousLogs, s)
				} else {
					break
				}
			}

			_function.GormDB.W.Model(&model.TcKdGrowth{}).Where("id = ?", taskUserItem.ID).Updates(model.TcKdGrowth{
				Status: string(jsonResult),
				Log:    fmt.Sprintf("%s: %s<br/>%s", time.Now().Format(time.DateOnly), tmpLog.String(), strings.Join(previousLogs, "<br/>")),
				Date:   int32(time.Now().Unix()),
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

	return _function.GormDB.W.Migrator().CreateTable(&model.TcKdGrowth{})
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
	_function.GormDB.W.Where("name = ?", "kd_growth_ext_tasks").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *UserGrowthTasksPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *UserGrowthTasksPluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}
	return _sql.Where(_type+" = ?", id).Delete(&model.TcKdGrowth{}).Error
}

func (pluginInfo *UserGrowthTasksPluginType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *UserGrowthTasksPluginType) Reset(uid, pid, tid int32) error {
	if uid == 0 {
		return errors.New("invalid uid")
	}

	_sql := _function.GormDB.W.Model(&model.TcKdGrowth{}).Where("uid = ?", uid)
	if pid != 0 {
		_sql = _sql.Where("pid = ?", pid)
	}

	if tid != 0 {
		_sql = _sql.Where("id = ?", tid)
	}

	return _sql.Update("date", 0).Error
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

	// ext tasks
	var extTasksMap = make(map[string]string)
	extTasks := _function.GetUserOption("kd_growth_ext_tasks", uid)
	if extTasks == "" {
		extTasks = "{}"
		_function.SetUserOption("kd_growth_ext_tasks", extTasks, uid)
	} else {
		err := _function.JsonDecode([]byte(extTasks), &extTasksMap)
		if err != nil {
			slog.Error("plugin.user-growth-tasks.settings.ext-tasks.read", "error", err)
			extTasks = "{}"
			_function.SetUserOption("kd_growth_ext_tasks", extTasks, uid)
			extTasksMap = map[string]string{}
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"sign_only":        signOnly,
		"break_icon_tasks": noIconTasks,
		"ext_tasks":        extTasksMap,
	}, "tbsign"))
}

func PluginGrowthTasksSetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	signOnly := strings.TrimSpace(c.FormValue("sign_only"))
	noIconTasks := c.FormValue("break_icon_tasks") != "0"
	extTasks := c.FormValue("ext_tasks")

	// invalid sign only value
	if !slices.Contains([]string{"0", "1", "2"}, signOnly) {
		signOnly = "0"
	}

	// ext tasks list
	var extTasksMap = make(map[string]string)
	if extTasks != "" {
		err := _function.JsonDecode([]byte(extTasks), &extTasksMap)
		if err != nil {
			slog.Error("plugin.user-growth-tasks.settings.ext-tasks.write", "error", err)
			extTasks = "{}"
		}
	} else {
		extTasks = "{}"
	}

	_function.SetUserOption("kd_growth_sign_only", signOnly, uid)
	_function.SetUserOption("kd_growth_break_icon_tasks", noIconTasks, uid)
	_function.SetUserOption("kd_growth_ext_tasks", extTasks, uid)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]bool{
		"success": true,
	}, "tbsign"))
}

func PluginGrowthTasksGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accounts []*model.TcKdGrowth
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
			UID:  int32(numUID),
			Pid:  int32(numPid),
			Date: 0,
		}
		_function.GormDB.W.Create(&dataToInsert)
		_function.GormDB.R.Model(&model.TcKdGrowth{}).Where("uid = ? AND pid = ?", uid, numPid).Take(&dataToInsert)
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
		UID: int32(numUID),
		ID:  int32(numID),
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
		UID: int32(numUID),
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

// special tasks

// 591
var growthTasks591Request1 = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 101, 111, 112, 97, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 97, 112, 105, 47, 116, 97, 115, 107, 47, 101, 120, 116, 101, 114, 110, 97, 108, 47, 98, 105, 122, 116, 97, 115, 107, 47, 99, 111, 109, 112, 108, 101, 116, 101})
var growthTasks591RequestReferrer = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 99, 116, 105, 118, 105, 116, 121, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47})

var growthTasks591SecretMap = map[string]string{
	"1309": string([]byte{109, 118, 82, 83, 55, 100, 106, 90, 81, 122, 101, 99, 117, 109, 84, 90, 49, 110, 88, 112, 81, 79, 65, 50, 113, 120, 89, 107, 118, 101, 49, 117}),
	"578":  string([]byte{110, 114, 115, 114, 107, 108, 114, 55, 68, 121, 69, 88, 86, 53, 65, 97, 115, 117, 51, 105, 88, 113, 104, 108, 80, 48, 71, 98, 84, 67, 111, 49}),
	"798":  string([]byte{82, 107, 49, 43, 50, 100, 88, 75, 43, 100, 52, 116, 108, 105, 98, 103, 103, 108, 100, 77, 52, 87, 82, 117, 119, 84, 112, 97, 65, 85, 107, 89, 79, 105, 49, 109, 109, 112, 76, 107, 105, 110, 107, 61}),
	"799":  string([]byte{57, 106, 73, 52, 101, 122, 83, 72, 68, 85, 69, 81, 85, 85, 70, 78, 86, 105, 107, 118, 99, 51, 102, 78, 87, 102, 109, 105, 118, 87, 86, 83, 113, 75, 77, 53, 79, 106, 89, 68, 104, 72, 52, 61}),
	"800":  string([]byte{100, 81, 69, 87, 70, 98, 81, 120, 66, 104, 68, 113, 106, 101, 69, 109, 99, 117, 90, 57, 57, 101, 54, 50, 113, 54, 85, 112, 75, 54, 98, 116, 90, 98, 75, 75, 71, 103, 108, 112, 101, 66, 65, 61}),
	"983":  string([]byte{56, 97, 49, 54, 48, 51, 49, 100, 45, 57, 49, 99, 99, 45, 52, 51, 50, 56, 45, 55, 57, 54, 56, 45, 102, 57, 51, 49, 57, 51, 97, 52, 98, 48, 101, 51}),
	"1150": string([]byte{49, 86, 122, 105, 105, 52, 106, 90, 99, 76, 78, 50, 80, 122, 100, 56, 118, 79, 102, 70, 75, 120, 89, 108, 74, 89, 71, 53, 73, 66, 110, 80}),
}

type growthTasks591TaskInfo struct {
	ID        string         `json:"id"`
	SDKParams map[string]any `json:"sdkParams"`
}

type growthTasks591ExecuteResponse struct {
	Errno  int    `json:"errno"`
	Errmsg string `json:"errmsg"`
}

func growthTasks591GenerateSign(body map[string]string, secret string) map[string]string {
	if secret == "" {
		return body
	}

	keys := make([]string, 0, len(body))
	for k := range body {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// build sign payload
	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, body[k]))
	}
	data := strings.Join(parts, "&")

	body["sign"] = hex.EncodeToString(_function.GenHMAC256([]byte(data), []byte(secret)))
	return body
}

func growthTasks591(uri string) (*growthTasks591ExecuteResponse, error) {
	schemeURL, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("解析 uri 失败: %v (%s)", err, uri)
	}

	innerURL := schemeURL.Query().Get("url")
	innerParsed, err := url.Parse(innerURL)
	if err != nil {
		return nil, fmt.Errorf("解析内部 URL 失败: %v (%s)", err, uri)
	}

	taskInfoStr := innerParsed.Query().Get("taskInfo")
	var taskInfo growthTasks591TaskInfo
	if err := json.Unmarshal([]byte(taskInfoStr), &taskInfo); err != nil {
		return nil, fmt.Errorf("解析 taskInfo 失败: %v (%s)", err, uri)
	}

	body := make(map[string]string)
	for k, v := range taskInfo.SDKParams {
		switch val := v.(type) {
		case string:
			body[k] = val
		default:
			jsonVal, _ := json.Marshal(val)
			body[k] = string(jsonVal)
		}
	}

	body["taskId"] = taskInfo.ID
	body["timestamp"] = strconv.FormatInt(time.Now().Unix()*1000, 10)

	growthTasks591GenerateSign(body, growthTasks591SecretMap[taskInfo.ID])

	formData := url.Values{}
	for k, v := range body {
		formData.Set(k, v)
	}

	res, err := _function.TBFetch(
		growthTasks591Request1, http.MethodPost,
		[]byte(formData.Encode()), map[string]string{
			"Content-Type": "application/x-www-form-urlencoded",
			"Referer":      growthTasks591RequestReferrer,
		},
	)
	if err != nil {
		return nil, err
	}

	var result growthTasks591ExecuteResponse
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
