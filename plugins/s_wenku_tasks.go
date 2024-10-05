package _plugin

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slices"
)

func init() {
	RegisterPlugin(WenkuTasksPlugin.Name, WenkuTasksPlugin)
}

type WenkuTasksPluginType struct {
	PluginInfo
}

var WenkuTasksPlugin = _function.VariablePtrWrapper(WenkuTasksPluginType{
	PluginInfo{
		Name:              "kd_wenku_tasks",
		PluginNameCN:      "文库任务",
		PluginNameCNShort: "文库任务",
		PluginNameFE:      "wenku_tasks",
		Version:           "0.1",
		Options: map[string]string{
			"kd_wenku_tasks_offset":       "0",
			"kd_wenku_tasks_vip_matrix":   "0",
			"kd_wenku_tasks_action_limit": "50",
		},
		SettingOptions: map[string]PluinSettingOption{
			"kd_wenku_tasks_action_limit": {
				OptionName:   "kd_wenku_tasks_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: func(value string) bool {
					numLimit, err := strconv.ParseInt(value, 10, 64)
					return err == nil && numLimit >= 0
				},
			},
		},
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "settings", Function: PluginWenkuTasksGetSettings},
			{Method: "PUT", Path: "settings", Function: PluginWenkuTasksSetSettings},
			{Method: "GET", Path: "list", Function: PluginWenkuTasksGetList},
			{Method: "PATCH", Path: "list", Function: PluginWenkuTasksAddAccount},
			{Method: "DELETE", Path: "list/:id", Function: PluginWenkuTasksDelAccount},
			{Method: "POST", Path: "list/empty", Function: PluginWenkuTasksDelAllAccounts},
			{Method: "GET", Path: "status/:pid", Function: PluginWenkuTasksGetTasksStatus},
		},
	},
})

const WenkuSemver = "9.0.70"

var WenkuUserAgent = fmt.Sprintf("%%E7%%99%%BE%%E5%%BA%%A6%%E6%%96%%87%%E5%%BA%%93/%s.7 CFNetwork/1498.700.2 Darwin/23.6.0", WenkuSemver)

var wenkuPassTasks = []int{4}

var wenkuTasksLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 112, 112, 119, 107, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 110, 97, 97, 112, 105, 47, 116, 97, 115, 107, 47, 116, 97, 115, 107, 108, 105, 115, 116, 63, 115, 114, 99, 61, 37, 115, 38, 110, 97, 95, 117, 110, 99, 104, 101, 99, 107, 61, 49})

var updateWenkuTaskLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 112, 112, 119, 107, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 110, 97, 97, 112, 105, 47, 116, 97, 115, 107, 47, 117, 112, 100, 97, 116, 101, 116, 97, 115, 107, 63, 116, 97, 115, 107, 73, 100, 61, 37, 100, 38, 115, 121, 115, 95, 118, 101, 114, 61, 49, 55, 46, 54, 46, 49, 38, 117, 105, 100, 61, 98, 100, 95, 48, 38, 97, 112, 112, 95, 118, 101, 114, 61, 37, 115, 38, 98, 105, 100, 61, 49, 38, 102, 114, 111, 109, 61, 105, 111, 115, 95, 38, 66, 100, 105, 95, 98, 101, 97, 114, 61, 38, 97, 112, 112, 95, 117, 97, 61, 105, 80, 97, 100, 49, 49, 44, 49, 38, 102, 114, 61, 50, 38, 112, 105, 100, 61, 49, 37, 115})

type WenkuTaskToSave struct {
	TaskName   string `json:"task_name"`
	TaskID     int    `json:"task_id"`
	TaskStatus int    `json:"task_status"`
	Msg        string `json:"msg"`
	SignDay    int64  `json:"sign_day,omitempty"`
	// RewardNum  int    `json:"reward_num"`
	// RewardType int    `json:"reward_type"`
}

type WenkuTaskList struct {
	TaskID   int    `json:"taskId,omitempty"`
	TaskName string `json:"taskName,omitempty"`
	// TaskDesc   string `json:"taskDesc,omitempty"`
	TaskStatus int `json:"taskStatus,omitempty"`
	// TaskIcon   string `json:"taskIcon,omitempty"`
	TaskEnd int `json:"taskEnd,omitempty"`
	// RewardNum  int    `json:"rewardNum,omitempty"`
	// RewardType int    `json:"rewardType,omitempty"`
	TaskExtra struct {
		// 	Rewards       []int `json:"rewards,omitempty"`
		SignDay       int `json:"signDay,omitempty"`
		IsFinishToday int `json:"isFinishToday,omitempty"`
		// 	NextRewardNum int   `json:"nextRewardNum,omitempty"`
	} `json:"taskExtra,omitempty"`
	// RewardIcon string `json:"rewardIcon,omitempty"`
	MinAppVer string `json:"minAppVer,omitempty"`
}

type GetWenkuTaskListResponse struct {
	Status struct {
		Code int    `json:"code,omitempty"`
		Msg  string `json:"msg,omitempty"`
	} `json:"status,omitempty"`
	Data struct {
		TaskList        []WenkuTaskList `json:"taskList,omitempty"`
		IsForbiddenUser int             `json:"isForbiddenUser,omitempty"`
		Errstr          string          `json:"errstr,omitempty"`
	} `json:"data,omitempty"`
}

type UpdateWenkuTaskResponse struct {
	Status struct {
		Code int    `json:"code,omitempty"`
		Msg  string `json:"msg,omitempty"`
	} `json:"status,omitempty"`
	Data struct {
		Task   WenkuTaskList `json:"task,omitempty"`
		Errstr string        `json:"errstr,omitempty"`
	} `json:"data,omitempty"`
}

// type: tasklist, signin
func GetWenkuTaskList(cookie _type.TypeCookie, _type string) (*GetWenkuTaskListResponse, error) {
	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"user-agent": WenkuUserAgent,
	}

	response, err := _function.TBFetch(fmt.Sprintf(wenkuTasksLink, _type), "GET", []byte{}, headersMap)
	if err != nil {
		return nil, err
	}

	// log.Println(string(response))

	resp := new(GetWenkuTaskListResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

// isClaim = false -> do task
func UpdateWenkuTask(cookie _type.TypeCookie, taskID int, minVersion string, isClaim bool) (*UpdateWenkuTaskResponse, error) {
	naUncheckStr := fmt.Sprintf("&extra=%%7B%%0A++%%22app_ver%%22+:+%%22%s%%22%%0A%%7D", minVersion)
	if isClaim {
		naUncheckStr = "&na_uncheck=1"
	}

	if minVersion == "" {
		minVersion = WenkuSemver
	} else {
		minVersion = _function.GetSemver(WenkuSemver, minVersion)
	}

	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"user-agent": strings.Replace(WenkuUserAgent, WenkuSemver, minVersion, 1),
	}

	response, err := _function.TBFetch(fmt.Sprintf(updateWenkuTaskLink, taskID, minVersion, naUncheckStr), "GET", []byte{}, headersMap)
	if err != nil {
		return nil, err
	}

	// log.Println(string(response))

	resp := new(UpdateWenkuTaskResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

var WenkuTasksPluginVipMatrixIDSetMap sync.Map

func (pluginInfo *WenkuTasksPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id, err := strconv.ParseInt(_function.GetOption("kd_wenku_tasks_offset"), 10, 64)
	if err != nil {
		id = 0
	}
	// status list
	var accountStatusList = make(map[int64]string)
	// cookie list
	var accountCookiesList = make(map[int64]_type.TypeCookie)

	// get list
	todayBeginning := _function.LocaleTimeDiff(0) //GMT+8
	kdWenkuTasksUserList := &[]model.TcKdWenkuTask{}

	limit := _function.GetOption("kd_wenku_tasks_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("date < ? AND id > ?", todayBeginning, id).Limit(int(numLimit)).Find(&kdWenkuTasksUserList)
	for _, taskUserItem := range *kdWenkuTasksUserList {
		if _, ok := accountStatusList[taskUserItem.UID]; !ok {
			accountStatusList[taskUserItem.UID] = _function.GetUserOption("kd_wenku_tasks_checkin_only", strconv.Itoa(int(taskUserItem.UID)))
		}
		if accountStatusList[taskUserItem.UID] == "" {
			// check uid is exists
			var accountInfo model.TcBaiduid
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", taskUserItem.UID).First(&accountInfo)
			if accountInfo.Portrait == "" {
				// clean
				_function.GormDB.W.Where("uid = ?", taskUserItem.UID).Delete(&model.TcKdWenkuTask{})
				accountStatusList[taskUserItem.UID] = "NOT_EXISTS"
				continue
			}
			// auto set -> 1
			accountStatusList[taskUserItem.UID] = "1"
			_function.SetUserOption("kd_wenku_tasks_checkin_only", "1", strconv.Itoa(int(taskUserItem.UID)))
		}

		var tasksList []WenkuTaskList
		var result []WenkuTaskToSave

		tasksIDList := sync.Map{}

		// vip matrix
		dbIDSetStr := _function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", strconv.Itoa(int(taskUserItem.UID)))

		var vipMatrixIDSet []string
		isVipMatrix := _function.GetUserOption("kd_wenku_tasks_vip_matrix", strconv.Itoa(int(taskUserItem.UID))) == "1"
		if isVipMatrix {
			if vipMatrixIDSetUnknow, ok := WenkuTasksPluginVipMatrixIDSetMap.Load(strconv.Itoa(int(taskUserItem.ID))); ok {
				vipMatrixIDSet = vipMatrixIDSetUnknow.([]string)
			} else {
				if dbIDSetStr != "|" {
					for _, v := range strings.Split(dbIDSetStr[1:len(dbIDSetStr)-1], "|") {
						pidSet := strings.Split(v, ",")
						WenkuTasksPluginVipMatrixIDSetMap.Store(pidSet[0], pidSet)
						if pidSet[0] == strconv.Itoa(int(taskUserItem.ID)) {
							vipMatrixIDSet = pidSet
						}
					}
				}
			}
		}
		if isVipMatrix && vipMatrixIDSet[1] == strconv.Itoa(int(_function.Now.Local().Weekday())) && vipMatrixIDSet[2] == "0" {
			result = append(result, WenkuTaskToSave{
				TaskName:   "VIP 矩阵自动跳过",
				TaskID:     -100,
				TaskStatus: 3,
				Msg:        "跳过",
			})
			dbIDSetStr = strings.Replace(dbIDSetStr, fmt.Sprintf("|%d,%s,%d|", taskUserItem.ID, vipMatrixIDSet[1], 0), fmt.Sprintf("|%d,%s,%d|", taskUserItem.ID, vipMatrixIDSet[1], 1), 1)
			_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", strconv.Itoa(int(taskUserItem.UID)), dbIDSetStr)
		} else {
			if _, ok := accountCookiesList[taskUserItem.Pid]; !ok {
				accountCookiesList[taskUserItem.Pid] = _function.GetCookie(int32(taskUserItem.Pid))
			}
			cookie := accountCookiesList[taskUserItem.Pid]

			signinTasksResponse, err := GetWenkuTaskList(cookie, "signin")
			if err != nil {
				log.Println(err)
				log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "Unable to fetch signin list")
				//continue
			} else if signinTasksResponse.Status.Code != 0 {
				log.Println(&signinTasksResponse)
				log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, signinTasksResponse.Status.Msg)
			} else {
				for _, v := range signinTasksResponse.Data.TaskList {
					if _, ok := tasksIDList.Load(v.TaskID); !ok && !slices.Contains(wenkuPassTasks, v.TaskID) && v.TaskStatus >= 1 && v.TaskStatus <= 3 {
						tasksIDList.Store(v.TaskID, nil)
						tasksList = append(tasksList, v)
					}
				}
			}
			if accountStatusList[taskUserItem.UID] != "1" {
				tasksListResponse, err := GetWenkuTaskList(cookie, "tasklist")
				if err != nil {
					log.Println(err)
					log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "Unable to fetch tasklist list")
					//continue
				} else if tasksListResponse.Status.Code != 0 {
					log.Println(&tasksListResponse)
					log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, tasksListResponse.Status.Msg)
				} else {
					for _, v := range tasksListResponse.Data.TaskList {
						if _, ok := tasksIDList.Load(v.TaskID); !ok && !slices.Contains(wenkuPassTasks, v.TaskID) && v.TaskStatus >= 1 && v.TaskStatus <= 3 {
							tasksIDList.Store(v.TaskID, nil)
							tasksList = append(tasksList, v)
						}
					}
				}
			}

			for _, _task := range tasksList {
				if accountStatusList[taskUserItem.UID] == "1" && _task.TaskID != 1 {
					continue
				}

				task := _task

				if task.TaskStatus == 1 {
					doTask, err := UpdateWenkuTask(cookie, task.TaskID, task.MinAppVer, false)
					if err != nil {
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: -1,
							Msg:        "未知错误",
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
						log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status1", err.Error())
					} else if doTask.Status.Code != 0 {
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: doTask.Status.Code,
							Msg:        doTask.Status.Msg,
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
					} else {
						task = doTask.Data.Task
					}
				}
				if task.TaskStatus == 2 {
					claimResponse, err := UpdateWenkuTask(cookie, task.TaskID, task.MinAppVer, true)
					if err != nil {
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: -1,
							Msg:        "未知错误",
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
						log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status2", err.Error())
					} else if claimResponse.Status.Code != 0 {
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: claimResponse.Status.Code,
							Msg:        claimResponse.Status.Msg,
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
					} else {
						task = claimResponse.Data.Task
					}
				}

				if task.TaskStatus == 3 {
					r := WenkuTaskToSave{
						TaskName:   task.TaskName,
						TaskID:     task.TaskID,
						TaskStatus: task.TaskStatus,
						Msg:        "success",
						// RewardType: task.RewardType,
						// RewardNum:  task.RewardNum,
					}
					if task.TaskID == 1 {
						r.SignDay = int64(task.TaskExtra.SignDay)
					}
					result = append(result, r)
				} else if task.TaskStatus != 3 {
					result = append(result, WenkuTaskToSave{
						TaskName:   task.TaskName,
						TaskID:     task.TaskID,
						TaskStatus: task.TaskStatus,
						Msg:        "未知错误",
						// RewardType: task.RewardType,
						// RewardNum:  task.RewardNum,
					})
					log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status3", task)
				}

				if isVipMatrix && task.TaskID == 1 && vipMatrixIDSet[2] == "1" {
					since, _ := strconv.ParseInt(vipMatrixIDSet[1], 10, 64)
					verifyDay := (int64(_function.Now.Local().Weekday()) + 7 - since) % 7
					if verifyDay == 0 {
						verifyDay = 7
					}

					if task.TaskStatus != 3 || task.TaskExtra.SignDay != int(verifyDay) {
						dbIDSetStr = strings.Replace(dbIDSetStr, fmt.Sprintf("|%d,%s,%d|", taskUserItem.ID, vipMatrixIDSet[1], 1), fmt.Sprintf("|%d,%s,%d|", taskUserItem.ID, vipMatrixIDSet[1], 0), 1)
						_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", strconv.Itoa(int(taskUserItem.UID)), dbIDSetStr)
					}

				}
			}
		}

		if len(result) > 0 {
			jsonResult, _ := _function.JsonEncode(result)
			tmpLog := ""
			for i, r := range result {
				if i > 0 {
					tmpLog += ","
				}
				tmpLog += fmt.Sprintf("%s#%d:%d", r.TaskName, r.TaskID, r.TaskStatus)
			}

			log.Println("wenku_tasks:", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, string(jsonResult))

			// previous logs
			previousLogs := []string{}
			for i, s := range strings.Split(taskUserItem.Log, "<br/>") {
				if i <= 28 {
					previousLogs = append(previousLogs, s)
				} else {
					break
				}
			}

			_function.GormDB.W.Model(&model.TcKdWenkuTask{}).Where("id = ?", taskUserItem.ID).Updates(model.TcKdWenkuTask{
				Status: string(jsonResult),
				Log:    fmt.Sprintf("%s: %s<br/>%s", _function.Now.Local().Format(time.DateOnly), tmpLog, strings.Join(previousLogs, "<br/>")),
				Date:   int32(_function.Now.Unix()),
			})
		}

		_function.SetOption("kd_wenku_tasks_offset", strconv.Itoa(int(taskUserItem.ID)))
	}
	_function.SetOption("kd_wenku_tasks_offset", "0")
	WenkuTasksPluginVipMatrixIDSetMap.Range(func(key, value any) bool {
		WenkuTasksPluginVipMatrixIDSetMap.Delete(key)
		return true
	})
}

func (pluginInfo *WenkuTasksPluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	_function.GormDB.W.Migrator().DropTable(&model.TcKdWenkuTask{})

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcKdWenkuTask{})
		_function.GormDB.W.Exec("ALTER TABLE `tc_kd_wenku_tasks` ADD UNIQUE KEY `id_uid_pid` (`id`,`uid`,`pid`), ADD KEY `uid` (`uid`), ADD KEY `pid` (`pid`), ADD KEY `date_id` (`date`,`id`) USING BTREE;")
	} else {
		_function.GormDB.W.Migrator().CreateTable(&model.TcKdWenkuTask{})

		_function.GormDB.W.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS "idx_tc_kd_wenku_tasks_id_uid_pid" ON "tc_kd_wenku_tasks" ("id","uid","pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_wenku_tasks_date_id" ON "tc_kd_wenku_tasks" ("date","id");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_wenku_tasks_pid" ON "tc_kd_wenku_tasks" ("pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_kd_wenku_tasks_uid" ON "tc_kd_wenku_tasks" ("uid");`)
	}
	return nil
}

func (pluginInfo *WenkuTasksPluginType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	_function.GormDB.W.Migrator().DropTable(&model.TcKdWenkuTask{})

	return nil
}
func (pluginInfo *WenkuTasksPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *WenkuTasksPluginType) RemoveAccount(_type string, id int32) error {
	_function.GormDB.W.Where("? = ?", _type, id).Delete(&model.TcKdWenkuTask{})
	return nil
}

func (pluginInfo *WenkuTasksPluginType) Ext() ([]any, error) {
	return []any{}, nil
}

// endpoints
func PluginWenkuTasksGetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	// checkin only
	checkinOnly := _function.GetUserOption("kd_wenku_tasks_checkin_only", uid)
	if checkinOnly == "" {
		checkinOnly = "0"
		_function.SetUserOption("kd_wenku_tasks_checkin_only", checkinOnly, uid)
	}

	// build a vip matrix (at least 7 accounts)
	vipMatrix := _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)
	if vipMatrix == "" {
		vipMatrix = "0"
		_function.SetUserOption("kd_wenku_tasks_vip_matrix", vipMatrix, uid)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"checkin_only": checkinOnly,
		"vip_matrix":   vipMatrix,
	}, "tbsign"))
}

func PluginWenkuTasksSetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	checkinOnly := c.FormValue("checkin_only") != "0"
	vipMatrix := c.FormValue("vip_matrix") != "0"

	dbVipMatrix := !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid))

	_function.SetUserOption("kd_wenku_tasks_checkin_only", checkinOnly, uid)
	_function.SetUserOption("kd_wenku_tasks_vip_matrix", vipMatrix, uid)

	// vip matrix list
	if vipMatrix && !dbVipMatrix {
		uidTasksList := new([]model.TcKdWenkuTask)
		_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("uid = ?", uid).Find(uidTasksList)
		idList := []string{}
		for i, task := range *uidTasksList {
			idList = append(idList, fmt.Sprintf("%d,%d,%d", task.ID, i%7, 0))
		}
		idListStr := "|"
		if len(idList) > 0 {
			idListStr += strings.Join(idList, "|") + "|"
		}
		_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", idListStr, uid)
	} else if !vipMatrix {
		_function.DeleteUserOption("kd_wenku_tasks_vip_matrix_id_set", uid)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
	}, "tbsign"))
}

func PluginWenkuTasksGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accounts []model.TcKdWenkuTask
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&accounts)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", accounts, "tbsign"))
}

func PluginWenkuTasksAddAccount(c echo.Context) error {
	uid := c.Get("uid").(string)
	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("uid = ? AND pid = ?", uid, numPid).Count(&count)
	if count > 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "账号已存在", _function.EchoEmptyObject, "tbsign"))
	} else {
		dataToInsert := model.TcKdWenkuTask{
			UID:  numUID,
			Pid:  numPid,
			Date: 0,
		}
		_function.GormDB.W.Create(&dataToInsert)
		_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("uid = ? AND pid = ?", uid, numPid).First(&dataToInsert)

		// vip matrix
		if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
			dbIDSetStr := _function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid)

			day := (len(strings.Split(dbIDSetStr, "|")) - 2) % 7

			_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", dbIDSetStr+fmt.Sprintf("%d,%d,%d", dataToInsert.ID, day, 0)+"|", uid)
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", dataToInsert, "tbsign"))
	}
}

func PluginWenkuTasksDelAccount(c echo.Context) error {
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

	_function.GormDB.W.Model(&model.TcKdWenkuTask{}).Delete(&model.TcKdWenkuTask{
		UID: numUID,
		ID:  numID,
	})

	// vip matrix
	if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
		dbIDSetStr := _function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid)

		newIDSet := []string{}
		for _, s := range strings.Split(dbIDSetStr, "|") {
			if !strings.HasPrefix(s, fmt.Sprintf("%d,", numID)) {
				newIDSet = append(newIDSet, s)
			}
		}

		_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", strings.Join(newIDSet, "|"), uid)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
		"id":      id,
	}, "tbsign"))
}

func PluginWenkuTasksDelAllAccounts(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	_function.GormDB.W.Model(&model.TcKdWenkuTask{}).Delete(&model.TcKdWenkuTask{
		UID: numUID,
	})

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func PluginWenkuTasksGetTasksStatus(c echo.Context) error {
	uid := c.Get("uid").(string)
	pid := c.Param("pid")

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).Count(&count)

	if count > 0 {
		numPid, _ := strconv.ParseInt(pid, 10, 64)
		cookie := _function.GetCookie(int32(numPid))

		var tasksList []WenkuTaskList

		tasksIDList := sync.Map{}

		signinTasksResponse, err := GetWenkuTaskList(cookie, "signin")
		if err != nil {
			log.Println(err)
			log.Println("wenku_tasks_api: ", cookie.ID, cookie.UID, "Unable to fetch signin list")
			//continue
		} else if signinTasksResponse.Status.Code != 0 {
			log.Println(&signinTasksResponse)
			log.Println("wenku_tasks_api: ", cookie.ID, cookie.UID, "Unable to fetch signin list", signinTasksResponse.Status.Msg)
		} else {
			for _, v := range signinTasksResponse.Data.TaskList {
				if _, ok := tasksIDList.Load(v.TaskID); !ok {
					tasksIDList.Store(v.TaskID, nil)
					tasksList = append(tasksList, v)
				}
			}
		}
		tasksListResponse, err := GetWenkuTaskList(cookie, "tasklist")
		if err != nil {
			log.Println(err)
			log.Println("wenku_tasks_api: ", cookie.ID, cookie.UID, "Unable to fetch tasklist list")
			//continue
		} else if tasksListResponse.Status.Code != 0 {
			log.Println(&tasksListResponse)
			log.Println("wenku_tasks_api: ", cookie.ID, cookie.UID, "Unable to fetch tasklist list", tasksListResponse.Status.Msg)
		} else {
			for _, v := range tasksListResponse.Data.TaskList {
				if _, ok := tasksIDList.Load(v.TaskID); !ok {
					tasksIDList.Store(v.TaskID, nil)
					tasksList = append(tasksList, v)
				}
			}
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tasksList, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}
}
