package _plugin

import (
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/leeqvip/gophp"
)

var wg sync.WaitGroup

var AgainErrorId = "160002"

//重复签到错误代码
//again_error_id_2 := "1101"
//特殊的重复签到错误代码！！！签到过快=已签到
//again_error_id_3 := "1102"

func Dosign(table string, retry bool) (bool, error) {
	//signMode := GetOption(options, "sign_mode")
	signHour, _ := strconv.ParseInt(_function.GetOption("sign_hour"), 10, 64)
	if int64(_function.Now.Hour()) <= signHour {
		log.Println(strconv.FormatInt(signHour, 10) + "点时忽略签到")
		return true, nil
	}
	//limit := GetOption(options, "cron_limit")
	var tiebaList []model.TcTieba
	//处理所有未签到的贴吧
	//TODO support limit
	//if limit == "0" {
	if retry {
		_function.GormDB.Where("no = ? AND latest != ? AND status IN ?", 0, _function.Now.Local().Day(), []int64{340011, 2280007, 110001, 1989004, 255}).Find(&tiebaList)
	} else {
		_function.GormDB.Where("no = ? AND latest != ?", 0, _function.Now.Local().Day()).Find(&tiebaList)
	}

	if len(tiebaList) <= 0 {
		log.Println("Empty list")
		return true, nil
	}
	//} else {
	//}

	//log.Println(tiebaList)
	sleep, _ := strconv.ParseInt(_function.GetOption("sign_sleep"), 10, 64)

	//force sleep
	if sleep <= 0 {
		sleep = 100
	}

	// get all cookies
	for _, v := range tiebaList {
		_function.GetCookie(v.Pid)
	}
	for _, v := range tiebaList {
		// we will not auto update fid
		if v.Fid == 0 {
			continue
		}
		wg.Add(1)
		go func(pid int32, kw string, fid int32, id int32, now time.Time) {
			// success := false
			ck := _function.GetCookie(pid)
			if ck.Bduss == "" {
				log.Println("sign: Failed, no such account", pid, kw, fid, id, now.Local().Day())
				return
			}
			response, err := _function.PostSignClient(ck, kw, fid)

			if err == nil && response.ErrorCode != "" {
				var errorCode int64 = 0
				errorMsg := "NULL"
				if !(response.ErrorCode == "0" || response.ErrorCode == AgainErrorId) {
					errorCode, _ = strconv.ParseInt(response.ErrorCode, 10, 64)
					errorMsg = response.ErrorMsg
				}
				//TODO better sql update
				_function.GormDB.Model(model.TcTieba{}).Where("id = ?", id).Updates(&model.TcTieba{
					Latest:    int32(now.Local().Day()),
					Status:    int32(errorCode),
					LastError: errorMsg,
				})
			}

			log.Println(pid, kw, fid, id, now.Local().Day(), response, time.Now().UnixMilli()-now.UnixMilli())
			defer wg.Done()
		}(v.Pid, v.Tieba, v.Fid, v.ID, _function.Now)

		time.Sleep(time.Millisecond * time.Duration(sleep))
	}
	wg.Wait()
	log.Println(time.Now().UnixMilli() - _function.Now.UnixMilli())
	return false, nil
}

func DoSignAction() {
	var today = _function.Now.Local().Format("2006-01-02")
	//limit, _ := strconv.ParseInt(GetOption(options, "cron_limit"), 10, 64)
	cornSignAgain := _function.GetOption("cron_sign_again")
	cornSignAgainParsed, err := gophp.Unserialize([]byte(cornSignAgain))
	if err != nil {
		panic(err)
	}

	cornSignAgainInterface, ok1 := cornSignAgainParsed.(map[string]interface{})
	retryNum, ok3 := cornSignAgainInterface["num"].(int)
	lastdo, ok2 := cornSignAgainInterface["lastdo"].(string)

	if !(ok1 && ok2 && ok3) {
		log.Fatal("Parsed failed")
	}
	log.Println(today, lastdo)

	if today != lastdo {
		// update lastdo
		cornSignAgainInterface["num"] = 0
		retryNum = 0
		cornSignAgainInterface["lastdo"] = today
		cornSignAgainEncoded, err := gophp.Serialize(cornSignAgainInterface)
		if err != nil {
			log.Fatal("encode failed")
		}

		_function.SetOption("cron_sign_again", string(cornSignAgainEncoded))

		log.Println(string(cornSignAgainEncoded))
	}

	var tables = []string{"tieba"}

	for _, table := range tables {
		Dosign(table, false)
	}

	retryMax, _ := strconv.ParseInt(_function.GetOption("retry_max"), 10, 64)
	if retryMax == 0 || cornSignAgainInterface["lastdo"] == today && int64(retryNum) <= retryMax && retryMax > -1 {
		for retryMax == 0 || int64(retryNum) <= retryMax {
			retryAgain := false
			for _, table := range tables {
				checkRequest, _ := Dosign(table, true)
				if checkRequest {
					retryAgain = true
				}
			}
			retryNum++
			cornSignAgainInterface["num"] = retryNum
			cornSignAgainEncoded, err := gophp.Serialize(cornSignAgainInterface)
			if err != nil {
				log.Fatal("encode failed")
			}
			_function.SetOption("cron_sign_again", string(cornSignAgainEncoded))
			if !retryAgain {
				break
			}
		}
	}
}

//TODO resign
//func DoReSignAction() {
//
//}
