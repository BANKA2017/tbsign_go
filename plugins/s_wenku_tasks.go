package _plugin

import (
	"errors"
	"fmt"
	"log"
	"net/http"
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
			"kd_wenku_tasks_offset":     "0",
			"kd_wenku_tasks_vip_matrix": "0",
			// "kd_wenku_tasks_vip_matrix_id_set": "|",
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
			{Method: http.MethodGet, Path: "settings", Function: PluginWenkuTasksGetSettings},
			{Method: http.MethodPut, Path: "settings", Function: PluginWenkuTasksSetSettings},
			{Method: http.MethodGet, Path: "list", Function: PluginWenkuTasksGetList},
			{Method: http.MethodPatch, Path: "list", Function: PluginWenkuTasksAddAccount},
			{Method: http.MethodDelete, Path: "list/:id", Function: PluginWenkuTasksDelAccount},
			{Method: http.MethodPost, Path: "list/empty", Function: PluginWenkuTasksDelAllAccounts},
			{Method: http.MethodGet, Path: "status/:pid", Function: PluginWenkuTasksGetTasksStatus},
			{Method: http.MethodPost, Path: "claim/:pid", Function: PluginWenkuTasksClaim7DaySignVIP},
		},
	},
})

const IOSVersion = "18.1.1"
const WenkuSemver = "9.1.40"

var WenkuUserAgent = "%E7%99%BE%E5%BA%A6%E6%96%87%E5%BA%93/" + WenkuSemver + ".5 CFNetwork/1568.200.51 Darwin/24.1.0"

var wenkuPassTasks = []int{4}

var wenkuTasksLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 112, 112, 119, 107, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 110, 97, 97, 112, 105, 47, 116, 97, 115, 107, 47, 116, 97, 115, 107, 108, 105, 115, 116, 63, 115, 114, 99, 61, 37, 115, 38, 110, 97, 95, 117, 110, 99, 104, 101, 99, 107, 61, 49})

var updateWenkuTaskLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 97, 112, 112, 119, 107, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 110, 97, 97, 112, 105, 47, 116, 97, 115, 107, 47, 117, 112, 100, 97, 116, 101, 116, 97, 115, 107, 63, 116, 97, 115, 107, 73, 100, 61, 37, 100, 38, 115, 121, 115, 95, 118, 101, 114, 61, 37, 115, 38, 117, 105, 100, 61, 98, 100, 95, 48, 38, 97, 112, 112, 95, 118, 101, 114, 61, 37, 115, 38, 98, 105, 100, 61, 49, 38, 102, 114, 111, 109, 61, 105, 111, 115, 95, 38, 66, 100, 105, 95, 98, 101, 97, 114, 61, 38, 97, 112, 112, 95, 117, 97, 61, 105, 80, 97, 100, 49, 49, 44, 49, 38, 102, 114, 61, 50, 38, 112, 105, 100, 61, 49, 37, 115})

var claimWenku7DaySignVIPLink = string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 116, 97, 110, 98, 105, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 104, 53, 97, 112, 112, 116, 111, 112, 105, 99, 47, 112, 114, 111, 120, 121, 47, 110, 97, 97, 112, 105, 47, 97, 99, 116, 105, 118, 105, 116, 121, 47, 108, 111, 116, 116, 101, 114, 121, 63, 97, 99, 116, 105, 111, 110, 61, 100, 114, 97, 119, 38, 110, 97, 95, 117, 110, 99, 104, 101, 99, 107, 61, 49, 38, 99, 111, 109, 98, 111, 61, 55, 100, 97, 121, 115, 105, 103, 110, 38, 95, 116, 61, 37, 100})

type WenkuTaskToSave struct {
	TaskName    string `json:"task_name"`
	TaskID      int    `json:"task_id"`
	TaskStatus  int    `json:"task_status"`
	Msg         string `json:"msg"`
	SignDay     int64  `json:"sign_day,omitempty"`
	ClaimStatus string `json:"claim_status,omitempty"`
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

type ClaimWenku7DaySignVIPResponse struct {
	Status struct {
		Code int    `json:"code,omitempty"`
		Msg  string `json:"msg,omitempty"`
	} `json:"status,omitempty"`
	Data struct {
		IsForbiddenUser int `json:"isForbiddenUser,omitempty"`
		IsWin           int `json:"isWin,omitempty"`
		MyBean          int `json:"myBean,omitempty"`
		RemaiNum        int `json:"remaiNum,omitempty"`
		Prize           struct {
			Prizeid int    `json:"prizeid,omitempty"`
			Icon    string `json:"icon,omitempty"`
			Name    string `json:"name,omitempty"`
			Desc    string `json:"desc,omitempty"`
		} `json:"prize,omitempty"`
		Errstr string `json:"errstr,omitempty"`
	} `json:"data,omitempty"`
}

// type: tasklist, signin
func GetWenkuTaskList(cookie _type.TypeCookie, _type string) (*GetWenkuTaskListResponse, error) {
	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"User-Agent": WenkuUserAgent,
	}

	response, err := _function.TBFetch(fmt.Sprintf(wenkuTasksLink, _type), http.MethodGet, []byte{}, headersMap)
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
	naUncheckStr := "&extra=%7B%0A%20%20%22app_ver%22%20%3A%20%22" + minVersion + "%22%0A%7D"
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
		"User-Agent": strings.Replace(WenkuUserAgent, WenkuSemver, minVersion, 1),
	}

	response, err := _function.TBFetch(fmt.Sprintf(updateWenkuTaskLink, taskID, IOSVersion, minVersion, naUncheckStr), http.MethodGet, []byte{}, headersMap)
	if err != nil {
		return nil, err
	}

	// log.Println(string(response))

	resp := new(UpdateWenkuTaskResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

func ClaimWenku7DaySignVIP(cookie _type.TypeCookie) (*ClaimWenku7DaySignVIPResponse, error) {
	headersMap := map[string]string{
		"Cookie":     "BDUSS=" + cookie.Bduss,
		"User-Agent": WenkuUserAgent,
		"Referrer":   string([]byte{104, 116, 116, 112, 115, 58, 47, 47, 116, 97, 110, 98, 105, 46, 98, 97, 105, 100, 117, 46, 99, 111, 109, 47, 104, 53, 97, 112, 112, 116, 111, 112, 105, 99, 47, 98, 114, 111, 119, 115, 101, 47, 108, 111, 116, 116, 101, 114, 121, 118, 105, 112, 50, 48, 50, 50, 49, 49}),
	}

	response, err := _function.TBFetch(fmt.Sprintf(claimWenku7DaySignVIPLink, _function.Now.UnixMilli()), http.MethodGet, []byte{}, headersMap)
	if err != nil {
		return nil, err
	}

	// log.Println(string(response))

	resp := new(ClaimWenku7DaySignVIPResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

type WenkuTasksPluginVipMatrixIDSet struct {
	MatrixIDMap map[string][4]string
	WeekDayList map[string]struct{}
	LastDay     string
}

func (m *WenkuTasksPluginVipMatrixIDSet) Init() {
	m.MatrixIDMap = make(map[string][4]string, 0)
	m.WeekDayList = make(map[string]struct{}, 0)
}

func (m *WenkuTasksPluginVipMatrixIDSet) Import(str string, uid string) error {
	// filter
	var idList []*model.TcKdWenkuTask
	_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Select("id").Where("uid = ?", uid).Find(&idList)

	if len(idList) == 0 {
		return nil
	}

	idArray := make([]string, len(idList))
	for i := range idList {
		idArray[i] = strconv.Itoa(int(idList[i].ID))
	}

	if len(str) > 1 && strings.HasPrefix(str, "|") && strings.HasSuffix(str, "|") {
		for _, v := range strings.Split(str[1:len(str)-1], "|") {
			idSet := strings.Split(v, ",")
			if slices.Contains(idArray, idSet[0]) {
				m.MatrixIDMap[idSet[0]] = [4]string{idSet[0], idSet[1], idSet[2], uid}
				if _, ok := m.WeekDayList[idSet[1]]; !ok {
					m.WeekDayList[idSet[1]] = struct{}{}
				}
				m.LastDay = idSet[1]
			}
		}
	}

	return nil
}

func (m *WenkuTasksPluginVipMatrixIDSet) Export(uid string) string {
	tmpStr := []string{}

	for _, arrayValue := range m.MatrixIDMap {
		if uid == "*" || len(arrayValue) == 4 && arrayValue[3] == uid || len(arrayValue) == 3 {
			// log.Println(arrayValue)
			tmpStr = append(tmpStr, strings.Join(arrayValue[0:3], ","))
		}
	}

	if len(tmpStr) == 0 {
		return "|"
	}

	return "|" + strings.Join(tmpStr, "|") + "|"
}

// !!! use this func for ONLY ONE UID !!!
func (m *WenkuTasksPluginVipMatrixIDSet) AddID(id int32, uid string, day int64, autoBreak bool) error {
	strID := strconv.Itoa(int(id))
	if _, ok := m.MatrixIDMap[strID]; ok {
		return nil
	}
	// time.Weekday -> Sunday is 0
	weekDayList := []string{"0", "1", "2", "3", "4", "5", "6"}
	var currentDay int64 = day
	strCurrentDay := strconv.Itoa(int(currentDay))

	if currentDay == -1 {
		autoBreak = true
		for i, d := range weekDayList {
			if _, ok := m.WeekDayList[d]; !ok {
				currentDay = int64(i)
				m.WeekDayList[d] = struct{}{}
				m.LastDay = d
				strCurrentDay = d
				break
			}
		}

		if currentDay == -1 {
			tmpLastDay, _ := strconv.ParseInt(m.LastDay, 10, 64)
			if tmpLastDay != 6 {
				currentDay = tmpLastDay + 1
			} else {
				currentDay = 0
			}
			strCurrentDay = strconv.Itoa(int(currentDay))
			m.LastDay = strCurrentDay
		}
	}

	m.MatrixIDMap[strID] = [4]string{strID, strCurrentDay, _function.When(autoBreak, "0", "1"), uid}

	return nil
}

// !!! use this func for ONLY ONE UID !!!
func (m *WenkuTasksPluginVipMatrixIDSet) DelID(id int32) error {
	strID := strconv.Itoa(int(id))

	if data, ok := m.MatrixIDMap[strID]; ok {
		delete(m.MatrixIDMap, strID)
		delete(m.WeekDayList, data[1])
	}

	return nil
}

func (m *WenkuTasksPluginVipMatrixIDSet) Clean() {
	m.MatrixIDMap = make(map[string][4]string, 0)
	m.WeekDayList = make(map[string]struct{}, 0)
	m.LastDay = ""
}

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
	var kdWenkuTasksUserList []*model.TcKdWenkuTask

	limit := _function.GetOption("kd_wenku_tasks_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)
	_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("date < ? AND id > ?", todayBeginning, id).Limit(int(numLimit)).Find(&kdWenkuTasksUserList)

	var wenkuTasksPluginVipMatrixIDSetMap WenkuTasksPluginVipMatrixIDSet
	wenkuTasksPluginVipMatrixIDSetMap.Init()

	for _, taskUserItem := range kdWenkuTasksUserList {
		strUID := strconv.Itoa(int(taskUserItem.UID))

		if _, ok := accountStatusList[taskUserItem.UID]; !ok {
			accountStatusList[taskUserItem.UID] = _function.GetUserOption("kd_wenku_tasks_checkin_only", strUID)
		}
		if accountStatusList[taskUserItem.UID] == "" {
			// check uid is exists
			var accountInfo model.TcBaiduid
			_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ?", taskUserItem.UID).Take(&accountInfo)
			if accountInfo.Portrait == "" {
				// clean
				_function.GormDB.W.Where("uid = ?", taskUserItem.UID).Delete(&model.TcKdWenkuTask{})
				accountStatusList[taskUserItem.UID] = "NOT_EXISTS"
				continue
			}
			// auto set -> 1
			accountStatusList[taskUserItem.UID] = "1"
			_function.SetUserOption("kd_wenku_tasks_checkin_only", "1", strUID)
		}

		var tasksList []WenkuTaskList
		var result []WenkuTaskToSave

		tasksIDList := make(map[int]bool)

		// vip matrix
		var vipMatrixIDSet [4]string
		isVipMatrix := _function.GetUserOption("kd_wenku_tasks_vip_matrix", strUID) == "1"
		if isVipMatrix {
			if vipMatrixIDSetUnknow, ok := wenkuTasksPluginVipMatrixIDSetMap.MatrixIDMap[strconv.Itoa(int(taskUserItem.ID))]; ok {
				vipMatrixIDSet = vipMatrixIDSetUnknow
			} else {
				wenkuTasksPluginVipMatrixIDSetMap.Import(_function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", strUID), strUID)
				_vipMatrixIDSet, ok := wenkuTasksPluginVipMatrixIDSetMap.MatrixIDMap[strconv.Itoa(int(taskUserItem.ID))]
				if !ok {
					wenkuTasksPluginVipMatrixIDSetMap.AddID(int32(taskUserItem.ID), strUID, -1, true)
					_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", wenkuTasksPluginVipMatrixIDSetMap.Export(strUID), strUID)
					_vipMatrixIDSet = wenkuTasksPluginVipMatrixIDSetMap.MatrixIDMap[strconv.Itoa(int(taskUserItem.ID))]
				}
				vipMatrixIDSet = _vipMatrixIDSet
			}
		}
		if isVipMatrix && vipMatrixIDSet[1] == strconv.Itoa(int(_function.Now.Weekday())) && vipMatrixIDSet[2] == "0" {
			result = append(result, WenkuTaskToSave{
				TaskName:   "VIP 账号组自动跳过",
				TaskID:     -100,
				TaskStatus: 3,
				Msg:        "跳过",
			})
			vipMatrixIDSet[2] = "1"
			wenkuTasksPluginVipMatrixIDSetMap.MatrixIDMap[vipMatrixIDSet[0]] = vipMatrixIDSet

			_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", wenkuTasksPluginVipMatrixIDSetMap.Export(strUID), strUID)
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
					if !tasksIDList[v.TaskID] && !slices.Contains(wenkuPassTasks, v.TaskID) && v.TaskStatus >= 1 && v.TaskStatus <= 3 {
						tasksIDList[v.TaskID] = true
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
						if !tasksIDList[v.TaskID] && !slices.Contains(wenkuPassTasks, v.TaskID) && v.TaskStatus >= 1 && v.TaskStatus <= 3 {
							tasksIDList[v.TaskID] = true
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
				hasError := false

				if task.TaskStatus == 1 {
					doTask, err := UpdateWenkuTask(cookie, task.TaskID, task.MinAppVer, false)
					if err != nil {
						hasError = true
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: -999,
							Msg:        "未知错误",
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
						log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status1", err.Error())
					} else if doTask.Status.Code == 9 {
						hasError = true
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: 9,
							Msg:        "您的账号因涉嫌刷分作弊而被封禁，不能进行此项操作",
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
						log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status9", task)
					} else if doTask.Status.Code != 0 {
						hasError = true
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
				if !hasError && task.TaskStatus == 2 {
					claimResponse, err := UpdateWenkuTask(cookie, task.TaskID, task.MinAppVer, true)
					if err != nil {
						hasError = true
						result = append(result, WenkuTaskToSave{
							TaskName:   task.TaskName,
							TaskID:     task.TaskID,
							TaskStatus: -999,
							Msg:        "未知错误",
							// RewardType: task.RewardType,
							// RewardNum:  task.RewardNum,
						})
						log.Println("wenku_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "task_status2", err.Error())
					} else if claimResponse.Status.Code != 0 {
						hasError = true
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

				if !hasError && task.TaskStatus == 3 {
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
				} else if !hasError && task.TaskStatus != 3 {
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
					verifyDay := (int64(_function.Now.Weekday()) + 7 - since) % 7
					if verifyDay == 0 {
						verifyDay = 7
					}

					if task.TaskStatus != 3 || task.TaskExtra.SignDay != int(verifyDay) {
						vipMatrixIDSet[2] = "0"
						wenkuTasksPluginVipMatrixIDSetMap.MatrixIDMap[vipMatrixIDSet[0]] = vipMatrixIDSet

						_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", wenkuTasksPluginVipMatrixIDSetMap.Export(strUID), strUID)
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
				if i <= 30 {
					previousLogs = append(previousLogs, s)
				} else {
					break
				}
			}

			_function.GormDB.W.Model(&model.TcKdWenkuTask{}).Where("id = ?", taskUserItem.ID).Updates(model.TcKdWenkuTask{
				Status: string(jsonResult),
				Log:    fmt.Sprintf("%s: %s<br/>%s", _function.Now.Format(time.DateOnly), tmpLog, strings.Join(previousLogs, "<br/>")),
				Date:   int32(_function.Now.Unix()),
			})
		}

		_function.SetOption("kd_wenku_tasks_offset", strconv.Itoa(int(taskUserItem.ID)))
	}
	_function.SetOption("kd_wenku_tasks_offset", "0")
	wenkuTasksPluginVipMatrixIDSetMap.Clean()
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

	// user options
	_function.GormDB.W.Where("name = ?", "kd_wenku_tasks_checkin_only").Delete(&model.TcUsersOption{})
	_function.GormDB.W.Where("name = ?", "kd_wenku_tasks_vip_matrix").Delete(&model.TcUsersOption{})
	_function.GormDB.W.Where("name = ?", "kd_wenku_tasks_vip_matrix_id_set").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *WenkuTasksPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *WenkuTasksPluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}

	var err error

	switch _type {
	case "pid":
		// get uid
		account := new(model.TcBaiduid)
		_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ?", id).Take(account)
		if account.ID == 0 {
			// pid is not exists
			return nil
		}
		uid := strconv.Itoa(int(account.UID))
		// get task id
		task := new(model.TcKdWenkuTask)
		_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("pid = ?", id).Take(task)

		// rebuild vip matrix set
		if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
			var vipMatrixSet WenkuTasksPluginVipMatrixIDSet
			vipMatrixSet.Init()
			vipMatrixSet.Import(_function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid), uid)
			vipMatrixSet.DelID(int32(task.ID))
			err = _function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", vipMatrixSet.Export(uid), uid, tx)
			if err != nil {
				return err
			}
		}
	case "uid":
		uid := strconv.Itoa(int(id))
		if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
			err = _function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", "|", uid, tx)
			if err != nil {
				return err
			}
		}
	}

	return _sql.Where(_type+" = ?", id).Delete(&model.TcKdWenkuTask{}).Error
}

func (pluginInfo *WenkuTasksPluginType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *WenkuTasksPluginType) Reset(uid, pid, tid int32) error {
	if uid == 0 {
		return errors.New("invalid uid")
	}

	_sql := _function.GormDB.W.Model(&model.TcKdWenkuTask{}).Where("uid = ?", uid)
	if pid != 0 {
		_sql = _sql.Where("pid = ?", pid)
	}

	if tid != 0 {
		_sql = _sql.Where("id = ?", tid)
	}

	return _sql.Update("date", 0).Error
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

	// vipMatrixSet := _function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid)
	// if vipMatrixSet == "" {
	// 	vipMatrixSet = "|"
	// 	_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", vipMatrixSet, uid)
	// }

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"checkin_only": checkinOnly,
		"vip_matrix":   vipMatrix,
		// "vip_matrix_set": vipMatrixSet,
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
		var uidTasksList []*model.TcKdWenkuTask
		_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("uid = ?", uid).Find(&uidTasksList)

		var vipMatrixSet WenkuTasksPluginVipMatrixIDSet
		vipMatrixSet.Init()

		for _, task := range uidTasksList {
			vipMatrixSet.AddID(int32(task.ID), uid, -1, true)
		}

		_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", vipMatrixSet.Export(uid), uid)
	} else if !vipMatrix {
		_function.DeleteUserOption("kd_wenku_tasks_vip_matrix_id_set", uid)
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"success": true,
	}, "tbsign"))
}

func PluginWenkuTasksGetList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var accounts []*model.TcKdWenkuTask
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

	day := c.FormValue("day")
	autoBreakValue := c.FormValue("auto_break")
	autoBreak := autoBreakValue != "" && autoBreakValue != "0" && autoBreakValue != "false"

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
		_function.GormDB.R.Model(&model.TcKdWenkuTask{}).Where("uid = ? AND pid = ?", uid, numPid).Take(&dataToInsert)

		// vip matrix
		if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
			numDay, err := strconv.ParseInt(day, 10, 64)
			if err != nil || numDay < -1 || numDay > 6 {
				return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效兑换日", _function.EchoEmptyObject, "tbsign"))
			}

			var vipMatrixSet WenkuTasksPluginVipMatrixIDSet
			vipMatrixSet.Init()
			vipMatrixSet.Import(_function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid), uid)
			vipMatrixSet.AddID(int32(dataToInsert.ID), uid, numDay, autoBreak)
			_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", vipMatrixSet.Export(uid), uid)
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
		var vipMatrixSet WenkuTasksPluginVipMatrixIDSet
		vipMatrixSet.Init()
		vipMatrixSet.Import(_function.GetUserOption("kd_wenku_tasks_vip_matrix_id_set", uid), uid)
		vipMatrixSet.DelID(int32(numID))
		_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", vipMatrixSet.Export(uid), uid)
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

	// vip matrix
	if !slices.Contains([]string{"", "0"}, _function.GetUserOption("kd_wenku_tasks_vip_matrix", uid)) {
		_function.SetUserOption("kd_wenku_tasks_vip_matrix_id_set", "|", uid)
	}

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

		tasksIDList := make(map[int]bool)

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
				if !tasksIDList[v.TaskID] {
					tasksIDList[v.TaskID] = true
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
				if !tasksIDList[v.TaskID] {
					tasksIDList[v.TaskID] = true
					tasksList = append(tasksList, v)
				}
			}
		}

		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tasksList, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "账号不存在", _function.EchoEmptyObject, "tbsign"))
	}
}

func PluginWenkuTasksClaim7DaySignVIP(c echo.Context) error {
	uid := c.Get("uid").(string)
	pid := c.Param("pid")

	// pre check
	var count int64
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id = ? AND uid = ?", pid, uid).Count(&count)

	if count > 0 {
		numPid, _ := strconv.ParseInt(pid, 10, 64)
		cookie := _function.GetCookie(int32(numPid))

		res, err := ClaimWenku7DaySignVIP(cookie)

		if err != nil {
			log.Println("wenku_tasks_api: claim 7days vip", pid, err, res)
			c.JSON(http.StatusOK, _function.ApiTemplate(500, "领取失败", ClaimWenku7DaySignVIPResponse{}, "tbsign"))
		}
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", res, "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "账号不存在", ClaimWenku7DaySignVIPResponse{}, "tbsign"))
	}
}
