package _plugin

import (
	"log"
	"net/url"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/exp/slices"
)

var UserGrowthTasksPluginName = "kd_growth"
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
	ID                 int    `json:"id,omitempty"`
	Name               string `json:"name,omitempty"`
	ActType            string `json:"act_type,omitempty"`
	URL                string `json:"url,omitempty"`
	Detail             string `json:"detail,omitempty"`
	Exp                int    `json:"exp,omitempty"`
	Current            int    `json:"current,omitempty"`
	Total              int    `json:"total,omitempty"`
	Status             int    `json:"status,omitempty"`
	SortStatus         int    `json:"sort_status,omitempty"`
	CompleteTime       int    `json:"complete_time,omitempty"`
	StartTime          int    `json:"start_time,omitempty"`
	ExpireTime         int    `json:"expire_time,omitempty"`
	MinLevel           int    `json:"min_level,omitempty"`
	TaskDoneNum        int    `json:"task_done_num,omitempty"`
	TaskThreadID       []any  `json:"task_thread_id,omitempty"`
	TargetKw           string `json:"target_kw,omitempty"`
	TargetScheme       string `json:"target_scheme,omitempty"`
	TargetChatroomName string `json:"target_chatroom_name,omitempty"`
	TargetChatroomID   int    `json:"target_chatroom_id,omitempty"`
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

	response, err := _function.Fetch("https://tieba.baidu.com/mo/q/usergrowth/commitUGTaskInfo", "POST", []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	var resp UserGrowthTasksWebResponse
	err = _function.JsonDecode(response, &resp)
	return &resp, err
}

// share_thread page_sign
func PostGrowthTaskByClient(cookie _type.TypeCookie, task string) (*UserGrowthTasksClientResponse, error) {
	form := map[string]string{
		"BDUSS":    cookie.Bduss,
		"act_type": task,
		"cuid":     "-",
		"tbs":      cookie.Tbs,
	}
	_function.AddSign(&form)
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	response, err := _function.Fetch("https://tiebac.baidu.com/c/c/user/commitUGTaskInfo", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp UserGrowthTasksClientResponse
	err = _function.JsonDecode(response, &resp)
	return &resp, err
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
	response, err := _function.Fetch("https://tieba.baidu.com/mo/q/icon/collectStamp", "POST", []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	var resp UserGrowthTaskCollectStampResponse
	err = _function.JsonDecode(response, &resp)
	return &resp, err
}

func GetUserGrowthTasksList(cookie _type.TypeCookie) (*UserGrowthTasksListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	response, err := _function.Fetch("https://tieba.baidu.com/mo/q/usergrowth/showUserGrowth", "GET", nil, headersMap)

	if err != nil {
		return nil, err
	}

	var resp UserGrowthTasksListResponse
	err = _function.JsonDecode(response, &resp)
	return &resp, err
}

var activeTasks = []string{"daily_task", "live_task"}

// TODO redo growth tasks(?)
func DoGrowthTasksAction() {
	id, err := strconv.ParseInt(_function.GetOption("kd_growth_offset"), 10, 64)
	if err != nil {
		id = 0
	}
	// status list
	var accountStatusList = make(map[int64]string)
	// cookie list
	var accountCookiesList = make(map[int64]_type.TypeCookie)

	// get list
	todayBeginning := _function.TodayBeginning() //GMT+8
	kdGrowthTasksUserList := &[]model.TcKdGrowth{}
	_function.GormDB.Model(&model.TcKdGrowth{}).Where("date < ? AND id > ?", todayBeginning, id).Find(&kdGrowthTasksUserList)
	for _, taskUserItem := range *kdGrowthTasksUserList {
		if _, ok := accountStatusList[taskUserItem.UID]; !ok {
			accountStatusList[taskUserItem.UID] = _function.GetUserOption("kd_growth_sign_only", strconv.Itoa(int(taskUserItem.UID)))
		}
		if accountStatusList[taskUserItem.UID] == "" {
			// check uid is exists
			var accountInfo model.TcBaiduid
			_function.GormDB.Model(&model.TcBaiduid{}).Where("uid = ?", taskUserItem.UID).First(&accountInfo)
			if accountInfo.Portrait == "" {
				// clean
				_function.GormDB.Where("uid = ?", taskUserItem.UID).Delete(&model.TcKdGrowth{})
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
		if accountStatusList[taskUserItem.UID] == "1" {
			tasksResponse, err := GetUserGrowthTasksList(cookie)
			if err != nil {
				log.Println(err)
				log.Println("user_tasks: ", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, "Unable to fetch tasks list")
				//continue
			}
			for _, taskTypeListList := range tasksResponse.Data.TabList {
				if taskTypeListList.TabName == "basic" {
					for _, taskTypeList := range taskTypeListList.TaskTypeList {
						if slices.Contains(activeTasks, taskTypeList.TaskType) {
							tasksList = append(tasksList, taskTypeList.TaskList...)
						}
						if taskTypeList.TaskType == "icon_task" {
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
				Status:     1,
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

			log.Println("user_tasks:", taskUserItem.ID, taskUserItem.Pid, taskUserItem.UID, string(jsonResult))
			_function.GormDB.Model(&model.TcKdGrowth{}).Where("id = ?", taskUserItem.ID).Updates(model.TcKdGrowth{
				Status: string(jsonResult),
				Log:    "<br/>" + _function.Now.Local().Format("2006-01-02") + ": " + string(jsonResult) + taskUserItem.Log,
				Date:   int32(_function.Now.Unix()),
			})
		}

		_function.SetOption("kd_growth_offset", strconv.Itoa(int(taskUserItem.ID)))

	}
	_function.SetOption("kd_growth_offset", "0")
}
