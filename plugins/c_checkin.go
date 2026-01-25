package _plugin

import (
	"context"
	"errors"
	"log"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/kdnetwork/code-snippet/go/worker"
)

const AgainErrorId = "160002"

// 1#用户未登录或登录失败，请更换账号或重试
// 340006#贴吧目录出问题啦，请到贴吧签到吧反馈
// 2280007#签到服务忙，请重新签到
// 110001#未知错误
// 110000#user not login/请先登录
// 340011#你签得太快了，先看看贴子再来签吧:)
// 320022#tbs check fail
// 2280015#您的账号现已被封禁，不能进行签到
// 2280001#您尚在黑名单中，不能操作。
// 1990055#帐号未实名，功能禁用。请先完成帐号的手机实名验证
var recheckinErrorID = []int64{340011, 2280007, 110001, 1989004, 255, 1, 340006}

var tcPrivateErrorID = map[int]string{
	110000: "请先登录",
}

// 重复签到错误代码
// again_error_id_2 := "1101"
// 特殊的重复签到错误代码！！！签到过快=已签到
// again_error_id_3 := "1102"

var tableList = []string{"tieba"}
var checkinToday string

var LastBreakHour = 0

func Dosign(_ string, retry bool) (bool, error) {
	//signMode := _function.GetOption("sign_mode")// client mode only
	hasFailed := false
	signHour, _ := strconv.ParseInt(_function.GetOption("sign_hour"), 10, 64)
	h := _function.Now.Hour()
	if int64(h) <= signHour {
		if h != LastBreakHour {
			log.Println("checkin:", strconv.FormatInt(signHour, 10)+"点时忽略签到")
			LastBreakHour = h
		}

		return hasFailed, nil
	}
	limit, _ := strconv.ParseInt(_function.GetOption("cron_limit"), 10, 64)
	var tiebaList []*model.TcTieba

	// retry has no limit
	if retry || limit == 0 {
		limit = -1
	}

	today := _function.Now.Day()
	if retry {
		// 重签
		_function.GormDB.R.Where("no = ? AND latest = ? AND status IN ?", 0, today, recheckinErrorID).Limit(int(limit)).Find(&tiebaList)
	} else {
		BatchPluginQuery(_function.GormDB.R.Model(&model.TcTieba{}).Where("no = ? AND latest != ?", 0, today), int(limit), 3, []string{"id", "pid", "tieba", "fid"}, &tiebaList)
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

	threadCount, _ := strconv.ParseInt(_function.GetOption("sign_multith"), 10, 64)
	if threadCount <= 0 {
		threadCount = 10
	}

	badBdussPidChan := make(chan int32, len(tiebaList))
	defer close(badBdussPidChan)

	ctx := context.Background()

	errs := worker.RunWorkerPool(ctx, tiebaList, int(threadCount), func(ctx context.Context, task *model.TcTieba, store map[int32]struct{}) error {
		if _, ok := store[task.Pid]; ok {
			return nil
		}

		now := _function.Now
		ck := _function.GetCookie(task.Pid)
		if ck.Bduss == "" {
			log.Println("checkin: Failed, no such account", task.Pid, task.Tieba, task.Fid, task.ID, today)
			return nil
		} else if !ck.IsLogin {
			store[task.Pid] = struct{}{}
			log.Println("checkin: Failed, account login status failed", task.Pid, task.Tieba, task.Fid, task.ID, today)

			// tc err && today
			if !(task.Status == 110000 && task.Latest == int32(today)) {
				badBdussPidChan <- task.Pid
			}

			return errors.New("account " + strconv.Itoa(int(task.Pid)) + " login status failed")
		}
		response, err := _function.PostCheckinClient(ck, task.Tieba, task.Fid)

		if err != nil {
			log.Println(err)
		} else if response.ErrorCode != "" {
			var errorCode int64 = 0
			errorMsg := "NULL"
			if !(response.ErrorCode == "0" || response.ErrorCode == AgainErrorId) {
				errorCode, _ = strconv.ParseInt(response.ErrorCode, 10, 64)
				errorMsg = response.ErrorMsg
			} else if response.ErrorCode == AgainErrorId {
				errorMsg = ""
			}

			// TODO better sql update
			if err := _function.GormDB.W.Model(&model.TcTieba{}).Select("status", "last_error", "latest").Where("id = ?", task.ID).Updates(&model.TcTieba{
				Status:    int32(errorCode),
				LastError: errorMsg,
				Latest:    int32(today),
			}).Error; err != nil {
				log.Println(err)
			}
		}

		log.Println("checkin:", task.Pid, task.Tieba, task.Fid, task.ID, today, time.Now().UnixMilli()-now.UnixMilli())
		time.Sleep(time.Millisecond * time.Duration(sleep))
		return err
	})

	for _, err := range errs {
		if err != nil {
			hasFailed = true
			break
		}
	}

	badBdussPidChanLen := len(badBdussPidChan)
	if badBdussPidChanLen > 0 {
		var badBdussPidArr []int32

		for range badBdussPidChanLen {
			badPid := <-badBdussPidChan
			if !slices.Contains(badBdussPidArr, badPid) {
				badBdussPidArr = append(badBdussPidArr, badPid)
			}
		}

		_function.GormDB.W.Model(&model.TcTieba{}).Select("status", "last_error", "latest").Where("pid IN (?)", badBdussPidArr).Updates(&model.TcTieba{
			Latest:    int32(today),
			Status:    110000,
			LastError: tcPrivateErrorID[110000],
		})
	}

	log.Println("checkin: done!")
	return hasFailed, nil
}

var cronSignAgainInterface CronSignAgainType

const StandardCronSignAgainLength = len(`a:2:{s:6:"lastdo";s:10:"2000-01-01";s:3:"num";i:0;}`)

type CronSignAgainType struct {
	Num    int
	LastDo string

	mu sync.Mutex
}

func (st *CronSignAgainType) Decode(str string) *CronSignAgainType {
	st.mu.Lock()
	defer st.mu.Unlock()

	if len(str) < StandardCronSignAgainLength {
		st.LastDo = "2000-01-01"
		return st
	}

	var valStart, valEnd = 0, 0

	// overflow?
	numStart := strings.Index(str, `s:3:"num";i:`)
	valStart = numStart + len(`s:3:"num";i:`)
	valEnd = strings.Index(str[valStart:], ";")
	num, _ := strconv.Atoi(str[valStart : valStart+valEnd])

	lastdoStart := strings.Index(str, `s:6:"lastdo";s:10:"`)
	valStart = lastdoStart + len(`s:6:"lastdo";s:10:"`)
	lastdo := str[valStart : valStart+10]

	st.Num = num
	st.LastDo = lastdo

	return st
}

func (st *CronSignAgainType) Encode() string {
	return `a:2:{s:6:"lastdo";s:10:"` + st.LastDo + `";s:3:"num";i:` + strconv.Itoa(st.Num) + `;}`
}

// 关于一键签到
// 会员 100 个吧，非会员 50 个 7 级以上 （文案里面还出现过超会 200 个吧和最多签 400 个吧……它开心就好）
// 每天一次机会，forum_ids 传什么都行，反正也不看，是直接按照等级倒序排下来的

// show_dialog == "1" && sign_notice != "" 时，按了会跳消息，不请求签到接口，强行签到过快// show_dialog: "0"|"1"
// 签到前 valid == "1" && today_exp == "0"，签到后 valid == "0"，today_exp 的值也会大于 "0"

func DoBatchCheckinAction() {}

func DoCheckinAction() {
	checkinToday = _function.Now.Format(time.DateOnly)
	// a:2:{s:3:"num";i:0;s:6:"lastdo";s:10:"2000-01-01";}
	cronSignAgain := _function.GetOption("cron_sign_again")
	cronSignAgainInterface.Decode(cronSignAgain)

	if checkinToday != cronSignAgainInterface.LastDo {
		// update lastdo
		cronSignAgainInterface.Num = 0
		cronSignAgainInterface.LastDo = checkinToday
		RecheckInStatus.UnixTimestamp = 0
		cronSignAgainEncoded := cronSignAgainInterface.Encode()

		_function.SetOption("cron_sign_again", cronSignAgainEncoded)

		//log.Println(string(cronSignAgainEncoded))
	}

	for _, table := range tableList {
		Dosign(table, false)
	}
}

var RecheckInStatus struct {
	UnixTimestamp int64
	NowInterval   int
}

func DoReCheckinAction() {
	retryMax, _ := strconv.ParseInt(_function.GetOption("retry_max"), 10, 64)

	if RecheckInStatus.UnixTimestamp > 0 && _function.Now.Add(time.Minute*-1*time.Duration(RecheckInStatus.NowInterval)).Unix() < RecheckInStatus.UnixTimestamp {
		return
	}

	retryNum := cronSignAgainInterface.Num

	if retryMax >= 0 && retryNum >= int(retryMax) {
		return
	}

	// all accounts are done?
	var checkinStatus = struct {
		UnDoneCount int
		FailedCount int
	}{}

	_function.GormDB.R.Model(&model.TcTieba{}).Select("COUNT(CASE WHEN latest != ? THEN 1 END) AS un_done_count, COUNT(CASE WHEN status IN (?) THEN 1 END) as failed_count", _function.Now.Day(), recheckinErrorID).Where("no = 0").Scan(&checkinStatus)

	if checkinStatus.UnDoneCount == 0 && checkinStatus.FailedCount > 0 {
		RecheckInStatus.UnixTimestamp = _function.Now.Unix()
		ReCheckInCaps, _ := strconv.ParseInt(_function.GetOption("go_re_check_in_max_interval"), 10, 64)
		if ReCheckInCaps < 1 {
			ReCheckInCaps = 1
			_function.SetOption("go_re_check_in_max_interval", "1")
		}
		if RecheckInStatus.NowInterval < int(ReCheckInCaps) {
			if RecheckInStatus.NowInterval > int(ReCheckInCaps)>>2 {
				RecheckInStatus.NowInterval = int(ReCheckInCaps)
			} else {
				RecheckInStatus.NowInterval <<= 2
			}
		}
		//for retryMax == 0 || int64(retryNum) <= retryMax {
		for _, table := range tableList {
			_, err := Dosign(table, true)
			if err != nil {
				log.Println(err)
			}
		}
		retryNum++
		cronSignAgainInterface.Num = retryNum
		cronSignAgainEncoded := cronSignAgainInterface.Encode()

		_function.SetOption("cron_sign_again", cronSignAgainEncoded)
		//}
	}
}
