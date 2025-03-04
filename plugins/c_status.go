package _plugin

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
)

// go_daily_report_status
// go_daily_report

type DailyReportStruct struct {
	UID  int32  `json:"uid"`
	Name string `json:"name"`

	GoDailyReport       string `json:"go_daily_report"`
	GoDailyReportStatus string `json:"go_daily_report_status"`
	GoMessageType       string `json:"go_message_type"`
}

var DailyReportLock int32

const dailyReportLimit = 100
const isLogoutPercent = 0.95

func DailyReportAction() {
	if !atomic.CompareAndSwapInt32(&DailyReportLock, 0, 1) {
		return
	}

	defer atomic.StoreInt32(&DailyReportLock, 0)

	dbDailyReportHour := _function.GetOption("go_daily_report_hour")
	if dbDailyReportHour == "-1" || dbDailyReportHour == "" {
		return
	}
	dailyReportHour, _ := strconv.ParseInt(dbDailyReportHour, 10, 64)

	signHour, _ := strconv.ParseInt(_function.GetOption("sign_hour"), 10, 64)
	if dailyReportHour < signHour {
		// log.Println("daily-report: 不可在签到开始前生成报告")
		return
	}

	if _function.Now.Hour() < int(dailyReportHour) {
		// not work now
		return
	}

	offsetUID := _function.GetOption("go_daily_report_uid")
	numOffsetUID, _ := strconv.ParseInt(offsetUID, 10, 64)

	todayBeginning := _function.LocaleTimeDiff(0)

	var dailyData []*DailyReportStruct

	_function.GormDB.R.Table("(?) as user_options_a", _function.GormDB.R.Model(&model.TcUsersOption{})).
		Joins("LEFT JOIN (?) user ON user_options_a.uid = user.id", _function.GormDB.R.Model(&model.TcUser{})).
		Joins("LEFT JOIN (?) user_options_c ON user_options_a.uid = user_options_c.uid AND user_options_c.name = ?", _function.GormDB.R.Model(&model.TcUsersOption{}), "go_message_type").
		Joins("LEFT JOIN (?) user_options_b ON user_options_a.uid = user_options_b.uid AND user_options_b.name = ?", _function.GormDB.R.Model(&model.TcUsersOption{}), "go_daily_report_status").
		Select("user_options_a.uid, user.name, COALESCE(user_options_a.value, '0') AS go_daily_report, COALESCE(user_options_b.value, '-1') AS go_daily_report_status, COALESCE(user_options_c.value, 'email') AS go_message_type").
		Where("user_options_a.uid > ? AND user_options_a.name = ? AND user_options_a.value = ?", numOffsetUID, "go_daily_report", "1").Limit(dailyReportLimit).Scan(&dailyData)

	if len(dailyData) > 0 {
		for _, data := range dailyData {
			if int64(data.UID) > numOffsetUID {
				numOffsetUID = int64(data.UID)
			}
			if data.GoDailyReportStatus == "-1" || strings.HasPrefix(data.GoDailyReportStatus, "-") { // and a.value === 1
				continue
			}

			// parse last status
			numDailyReportTime, _ := strconv.ParseInt(data.GoDailyReportStatus, 10, 64)
			sendTime := time.Unix(numDailyReportTime, 0)
			if todayBeginning > sendTime.Unix() {
				// send
				dailyStatus, err := GetForumCheckInStatus([]int32{data.UID})
				if err != nil {
					log.Println(err, data.UID)
					continue
				}
				// check-in
				messageObject := PushMessageTemplateDailyReport(int32(data.UID), data.Name, dailyStatus) //_function.PushMessageTestTemplate()

				// plugins
				for _, _pluginInfo := range PluginList {
					if _pluginInfo.(PluginHooks).GetSwitch() {
						// TDOO disable endpoint before install?
						message, err := _pluginInfo.Report(data.UID, nil)
						if err != nil || message == "" {

						} else {
							messageObject.Body += "<br />" + message
						}
					}
				}

				err = _function.SendMessage(data.GoMessageType, data.UID, messageObject.Title, messageObject.Body)
				if err != nil {
					log.Println(err, data.UID)
				}
				err = _function.SetUserOption("go_daily_report_status", int(_function.Now.Unix()), strconv.Itoa(int(data.UID)))
				if err != nil {
					log.Println(err, data.UID)
				}
			}
		}
	} else {
		numOffsetUID = 0
	}

	_function.SetOption("go_daily_report_uid", int(numOffsetUID))
}

type AccountStatusStruct struct {
	PID      int32  `json:"pid"`
	Name     string `json:"name"`
	Portrait string `json:"portrait"`
	IsLogout bool   `json:"is_logout"`
	_type.StatusStruct
}

func GetForumCheckInStatus(uid []int32) ([]*AccountStatusStruct, error) {
	var Status []*AccountStatusStruct

	today := strconv.Itoa(_function.Now.Day())

	err := _function.GormDB.R.Select("tc_baiduid.id, tc_baiduid.name, tc_baiduid.portrait, tc_tieba.is_logout, tc_tieba.success, tc_tieba.failed, tc_tieba.waiting, tc_tieba.is_ignore").Table("(?) AS tc_baiduid", _function.GormDB.R.Model(&model.TcBaiduid{}).Select("id, uid, name, portrait")).Joins("INNER JOIN (?) AS tc_tieba ON tc_tieba.pid=tc_baiduid.id", _function.GormDB.R.Model(&model.TcTieba{}).Select(`pid, CASE WHEN COALESCE(CAST(SUM(CASE WHEN no = 0 AND latest = ? and status = 1 THEN 1 ELSE 0 END) AS FLOAT)/COALESCE(SUM(CASE WHEN no = 0 AND latest = ? THEN 1 ELSE 0 END), NULL), 0) > ? then 1 ELSE 0 end AS is_logout, SUM(CASE WHEN (no = 0) AND status = 0 AND latest = ? THEN 1 ELSE 0 END) AS success, SUM(CASE WHEN (no = 0) AND status <> 0 AND latest = ? THEN 1 ELSE 0 END) AS failed, SUM(CASE WHEN (no = 0) AND latest <> ? THEN 1 ELSE 0 END) AS waiting, SUM(CASE WHEN no <> 0 THEN 1 ELSE 0 END) AS is_ignore`, today, today, isLogoutPercent, today, today, today).Group("pid")).Where("uid IN (?)", uid).Order("tc_baiduid.id ASC").Scan(&Status).Error

	return Status, err
}

func PushMessageTemplateDailyReport(uid int32, username string, accountStatus []*AccountStatusStruct) _function.PushMessageTemplateStruct {
	now := _function.Now.Format(time.DateOnly)

	msg := []string{}

	for _, status := range accountStatus {
		msg = append(msg, fmt.Sprintf("- [ %s ]：%d / %d / %d / %d %s", _function.When(status.Name != "", status.Name, status.Portrait), status.Success, status.Failed, status.Waiting, status.IsIgnore, _function.When(status.IsLogout, "，登录失效", "")))
	}

	return _function.PushMessageTemplateStruct{
		Title: now + " 签到报告",
		Body: fmt.Sprintf(
			"贴吧云签到账号 [ %s ] %s 签到情况:<br /><br />"+strings.Join(msg, "<br />")+"<br /><br />格式说明：成功 / 失败 / 等待 / 忽略<br /><br />UID: %d",
			username, now, uid),
	}
}
