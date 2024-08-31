package _plugin

import (
	"log"
	"strconv"
	"sync"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/leeqvip/gophp"
)

var wg sync.WaitGroup

const AgainErrorId = "160002"

var recheckinErrorID = []int64{340011, 2280007, 110001, 1989004, 255}

var spCheckinQuota = 2 //?

// 1#用户未登录或登录失败，请更换账号或重试
// 340006#贴吧目录出问题啦，请到贴吧签到吧反馈
var spRecheckinErrorID = []int64{1, 340006}

type SPCheckinDBItem struct {
	ID        int32
	Quota     int
	Latest    int32
	Timestamp int64
}

var SPCheckinDB sync.Map

// 重复签到错误代码
// again_error_id_2 := "1101"
// 特殊的重复签到错误代码！！！签到过快=已签到
// again_error_id_3 := "1102"

var cornSignAgainInterface map[string]any
var tableList = []string{"tieba"}
var checkinToday string

func Dosign(table string, retry bool) (bool, error) {
	//signMode := _function.GetOption("sign_mode")// client mode only
	hasFailed := false
	signHour, _ := strconv.ParseInt(_function.GetOption("sign_hour"), 10, 64)
	if int64(_function.Now.Hour()) <= signHour {
		log.Println("sign:", strconv.FormatInt(signHour, 10)+"点时忽略签到")
		return hasFailed, nil
	}
	limit, _ := strconv.ParseInt(_function.GetOption("cron_limit"), 10, 64)
	var tiebaList []model.TcTieba

	// retry has no limit
	if retry || limit == 0 {
		limit = -1
	}
	if retry {
		// 重签

		today := _function.Now.Local().Day()
		_function.GormDB.R.Where("no = ? AND latest = ? AND status IN ?", 0, today, recheckinErrorID).Limit(int(limit)).Find(&tiebaList)

		// special re-checkin
		spReCheckinList := new([]model.TcTieba)
		err := _function.GormDB.R.Where("no = ? AND latest = ? AND status IN ?", 0, today, spRecheckinErrorID).Limit(int(limit)).Find(spReCheckinList).Error
		if err == nil {
			for _, spReCheckinListItem := range *spReCheckinList {
				_data, ok := SPCheckinDB.Load(spReCheckinListItem.ID)
				var data SPCheckinDBItem
				if !ok {
					data = SPCheckinDBItem{
						ID:        spReCheckinListItem.ID,
						Quota:     spCheckinQuota,
						Latest:    spReCheckinListItem.Latest,
						Timestamp: time.Now().Unix(),
					}
				} else {
					data, _ = _data.(SPCheckinDBItem)
				}

				if data.Latest == int32(today) && data.Quota > 0 {
					data.Quota--
					data.Timestamp = time.Now().Unix() // what?
					tiebaList = append(tiebaList, spReCheckinListItem)
					SPCheckinDB.Store(spReCheckinListItem.ID, data)
				}
			}
		}

	} else {
		_function.GormDB.R.Table(
			"(?) as forums",
			_function.GormDB.R.Table(
				"(?) as filtered_forums",
				_function.GormDB.R.Model(&model.TcTieba{}).Where("no = ? AND latest != ?", 0, _function.Now.Local().Day()).Select("*"),
			).Select("*", "ROW_NUMBER() OVER (PARTITION BY pid ORDER BY id) AS rn"),
		).Select("id", "pid", "tieba", "fid").Where("no = ? AND latest != ? AND rn <= ?", 0, _function.Now.Local().Day(), limit).Limit(int(limit * 3)).Find(&tiebaList)
	}

	if len(tiebaList) <= 0 {
		//log.Println("sign: Empty list")
		return hasFailed, nil
	}

	//log.Println(tiebaList)
	sleep, _ := strconv.ParseInt(_function.GetOption("sign_sleep"), 10, 64)

	//force sleep
	if sleep <= 0 {
		sleep = 100
	}

	var forceWaitCount = 50
	for _, v := range tiebaList {
		// we will not auto update fid
		if v.Fid == 0 {
			continue
		}
		wg.Add(1)
		go func(pid int32, kw string, fid int32, id int32, now time.Time) {
			defer wg.Done()
			// success := false
			ck := _function.GetCookie(pid)
			if ck.Bduss == "" {
				log.Println("sign: Failed, no such account", pid, kw, fid, id, now.Local().Day())
				return
			}
			response, err := _function.PostSignClient(ck, kw, fid)

			if err != nil {
				log.Println(err)
				hasFailed = true
			} else if response.ErrorCode != "" {
				var errorCode int64 = 0
				errorMsg := "NULL"
				if !(response.ErrorCode == "0" || response.ErrorCode == AgainErrorId) {
					errorCode, _ = strconv.ParseInt(response.ErrorCode, 10, 64)
					errorMsg = response.ErrorMsg
				} else if response.ErrorCode == AgainErrorId {
					errorMsg = ""
				}

				//TODO better sql update
				_function.GormDB.W.Model(model.TcTieba{}).Where("id = ?", id).Updates(&_type.TcTieba{
					Status:    _function.VariablePtrWrapper(int32(errorCode)),
					LastError: _function.VariablePtrWrapper(errorMsg),
					TcTieba: model.TcTieba{
						Latest: int32(now.Local().Day()),
					},
				})
			}

			log.Println("sign:", pid, kw, fid, id, now.Local().Day(), time.Now().UnixMilli()-now.UnixMilli())
		}(v.Pid, v.Tieba, v.Fid, v.ID, _function.Now)

		time.Sleep(time.Millisecond * time.Duration(sleep))

		forceWaitCount--
		if forceWaitCount <= 0 {
			forceWaitCount = 50
			wg.Wait()
		}
	}
	wg.Wait()
	log.Println("sign: done!")
	return hasFailed, nil
}

func DoCheckinAction() {
	checkinToday = _function.Now.Local().Format("2006-01-02")
	// a:2:{s:3:"num";i:0;s:6:"lastdo";s:10:"2000-01-01";}
	cornSignAgain := _function.GetOption("cron_sign_again")
	cornSignAgainParsed, err := gophp.Unserialize([]byte(cornSignAgain))
	if err != nil {
		log.Println("sign:", err)
		return
	}

	var ok bool
	if cornSignAgainInterface, ok = cornSignAgainParsed.(map[string]any); !ok {
		log.Println("sign: parse config failed (lastdo)")
		return
	}

	if checkinToday != cornSignAgainInterface["lastdo"].(string) {
		// update lastdo
		cornSignAgainInterface["num"] = 0
		cornSignAgainInterface["lastdo"] = checkinToday
		cornSignAgainEncoded, err := gophp.Serialize(cornSignAgainInterface)
		if err != nil {
			log.Println("sign: encode php serialize failed", err)
			return
		}

		_function.SetOption("cron_sign_again", string(cornSignAgainEncoded))

		//log.Println(string(cornSignAgainEncoded))
	}

	for _, table := range tableList {
		Dosign(table, false)
	}

}

func DoReCheckinAction() {
	retryMax, _ := strconv.ParseInt(_function.GetOption("retry_max"), 10, 64)

	retryNum := cornSignAgainInterface["num"].(int)

	// all accounts are done?
	var unDoneCount int64
	_function.GormDB.R.Model(&model.TcTieba{}).Where("no = 0 AND latest != ?", _function.Now.Local().Day()).Count(&unDoneCount)

	var failedCount int64
	_function.GormDB.R.Model(&model.TcTieba{}).Where("no = 0 AND status IN ?", recheckinErrorID).Count(&failedCount)

	if unDoneCount == 0 && failedCount > 0 && (retryMax == 0 || int64(retryNum) < retryMax && retryMax > 0) {
		for retryMax == 0 || int64(retryNum) <= retryMax {
			retryAgain := false
			for _, table := range tableList {
				hasFailed, _ := Dosign(table, true)
				if hasFailed {
					retryAgain = true
				}
			}
			retryNum++
			cornSignAgainInterface["num"] = retryNum
			cornSignAgainEncoded, err := gophp.Serialize(cornSignAgainInterface)
			if err != nil {
				log.Println("sign_retry: encode failed")
				return
			}
			_function.SetOption("cron_sign_again", string(cornSignAgainEncoded))
			if !retryAgain {
				break
			}
		}
	}
}
