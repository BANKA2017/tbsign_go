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

// 1#用户未登录或登录失败，请更换账号或重试
// 340006#贴吧目录出问题啦，请到贴吧签到吧反馈
var recheckinErrorID = []int64{340011, 2280007, 110001, 1989004, 255, 1, 340006}

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
		log.Println("checkin:", strconv.FormatInt(signHour, 10)+"点时忽略签到")
		return hasFailed, nil
	}
	limit, _ := strconv.ParseInt(_function.GetOption("cron_limit"), 10, 64)
	var tiebaList []model.TcTieba

	// retry has no limit
	if retry || limit == 0 {
		limit = -1
	}

	today := _function.Now.Local().Day()
	if retry {
		// 重签
		_function.GormDB.R.Where("no = ? AND latest = ? AND status IN ?", 0, today, recheckinErrorID).Limit(int(limit)).Find(&tiebaList)
	} else {
		_function.GormDB.R.Table(
			"(?) as forums",
			_function.GormDB.R.Table(
				"(?) as filtered_forums",
				_function.GormDB.R.Model(&model.TcTieba{}).Where("no = ? AND latest != ?", 0, today).Select("*"),
			).Select("*", "ROW_NUMBER() OVER (PARTITION BY pid ORDER BY id) AS rn"),
		).Select("id", "pid", "tieba", "fid").Where("no = ? AND latest != ? AND rn <= ?", 0, today, limit).Limit(int(limit * 3)).Find(&tiebaList)
	}

	if len(tiebaList) <= 0 {
		//log.Println("checkin: Empty list")
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
				log.Println("checkin: Failed, no such account", pid, kw, fid, id, now.Local().Day())
				return
			}
			response, err := _function.PostCheckinClient(ck, kw, fid)

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

			log.Println("checkin:", pid, kw, fid, id, now.Local().Day(), time.Now().UnixMilli()-now.UnixMilli())
		}(v.Pid, v.Tieba, v.Fid, v.ID, _function.Now)

		time.Sleep(time.Millisecond * time.Duration(sleep))

		forceWaitCount--
		if forceWaitCount <= 0 {
			forceWaitCount = 50
			wg.Wait()
		}
	}
	wg.Wait()
	log.Println("checkin: done!")
	return hasFailed, nil
}

func DoCheckinAction() {
	checkinToday = _function.Now.Local().Format(time.DateOnly)
	// a:2:{s:3:"num";i:0;s:6:"lastdo";s:10:"2000-01-01";}
	cornSignAgain := _function.GetOption("cron_sign_again")
	cornSignAgainParsed, err := gophp.Unserialize([]byte(cornSignAgain))
	if err != nil {
		log.Println("checkin:", err)
		return
	}

	var ok bool
	if cornSignAgainInterface, ok = cornSignAgainParsed.(map[string]any); !ok {
		log.Println("checkin: parse config failed (lastdo)")
		return
	}

	if checkinToday != cornSignAgainInterface["lastdo"].(string) {
		// update lastdo
		cornSignAgainInterface["num"] = 0
		cornSignAgainInterface["lastdo"] = checkinToday
		cornSignAgainEncoded, err := gophp.Serialize(cornSignAgainInterface)
		if err != nil {
			log.Println("checkin: encode php serialize failed", err)
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
