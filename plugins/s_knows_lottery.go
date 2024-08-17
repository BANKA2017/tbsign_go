package _plugin

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
)

type LotteryPluginPluginType struct {
	PluginInfo
}

var LotteryPluginPlugin = _function.VariablePtrWrapper(LotteryPluginPluginType{
	PluginInfo{
		Name:    "ver4_lottery",
		Version: "1.0",
		Options: map[string]string{
			"ver4_lottery_pid": "0",
			"ver4_lottery_day": "0",
		},
	},
})

type GetLotteryResponse struct {
	Errno int `json:"errno,omitempty"`
	Data  *struct {
		PrizeList []struct {
			GoodsName string `json:"goodsName,omitempty"`
		} `json:"prizeList,omitempty"`
	} `json:"data,omitempty"`
	Errmsg string `json:"errmsg,omitempty"`
}

func GetLotteryToken(cookie _type.TypeCookie) (string, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}
	resp, err := _function.TBFetch("https://zhidao.baidu.com/shop/lottery", "GET", []byte{}, headersMap)
	if err != nil {
		return "", err
	}

	for _, match := range regexp.MustCompile(`(?m)'luckyToken',(?:\s+|)'([^']+)'`).FindAllSubmatch(resp, -1) {
		return string(match[1]), nil
	}
	return "", fmt.Errorf("get_lottery_token: No token")
}

func GetLottery(cookie _type.TypeCookie, token string) (*GetLotteryResponse, error) {
	headersMap := map[string]string{
		"Cookie":   "BDUSS=" + cookie.Bduss,
		"x-ik-ssl": "1",
		"Referer":  "https://zhidao.baidu.com/shop/lottery",
	}
	response, err := _function.TBFetch(fmt.Sprintf("https://zhidao.baidu.com/shop/submit/lottery?type=0&token=%s&_=%d", token, _function.Now.UnixMilli()), "GET", []byte{}, headersMap)
	if err != nil {
		return nil, err
	}

	// log.Println(string(response))

	resp := new(GetLotteryResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

var notCompleteActionPid sync.Map

func (pluginInfo *LotteryPluginPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id := _function.GetOption("ver4_lottery_pid")

	// 10 am gmt+8
	stime := _function.LocaleTimeDiff(10)

	queryLotteryLogs := _function.GormDB.R.Model(&model.TcVer4LotteryLog{}).Select("pid").Where("date > ?", stime).Where("pid > ?", id).Group("pid").Having("max(date) >= ? OR COUNT(*) >= ?", _function.Now.Add(time.Minute*-10).Unix(), 2)

	queryUserOptions := _function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid").Where("name='ver4_lottery_check' AND value = '1'")

	accounts := new([]model.TcBaiduid)

	// TODO fix hard limit
	_function.GormDB.R.Model(&model.TcBaiduid{}).Select("id").Where("id > ? AND id NOT IN (?) AND uid IN (?)", id, queryLotteryLogs, queryUserOptions).Order("id").Limit(50).Find(accounts)

	w := time.NewTicker(time.Second)
	defer w.Stop()

	if len(*accounts) > 0 {
		for i, account := range *accounts {
			if i > 0 {
				// wait 2s
				w.Reset(time.Second)
				<-w.C
			}
			cookie := _function.GetCookie(account.ID, true)
			dataToInsert := model.TcVer4LotteryLog{
				UID:   cookie.UID,
				Pid:   account.ID,
				Date:  int32(_function.Now.Unix()),
				Prize: "-",
			}

			token, err := GetLotteryToken(cookie)
			if err != nil || token == "" {
				dataToInsert.Result = "无法获取 token"
				log.Println(err)
			}

			_, hasNotCompleted := notCompleteActionPid.Load(account.ID)

			resp, err := GetLottery(cookie, token)
			if err != nil && (resp == nil || resp.Data == nil) {
				dataToInsert.Result = "无法解析物品信息"
				if hasNotCompleted {
					notCompleteActionPid.Delete(account.ID)
				}
				log.Println(err, resp)
			} else if err != nil && resp.Errno == 0 && len(resp.Data.PrizeList) == 0 {
				if hasNotCompleted {
					dataToInsert.Result = "未完成抽奖"
					notCompleteActionPid.Delete(account.ID)
				} else {
					log.Printf("knows_lottery: %d:%s[ %s ] 第一次未完成抽奖\n", account.ID, account.Name, account.Portrait)
					notCompleteActionPid.Store(account.ID, nil)
				}
			} else if resp.Errno != 0 {
				dataToInsert.Result = resp.Errmsg
				if hasNotCompleted {
					notCompleteActionPid.Delete(account.ID)
				}
			} else {
				dataToInsert.Prize = resp.Data.PrizeList[0].GoodsName
				if hasNotCompleted {
					notCompleteActionPid.Delete(account.ID)
				}
			}

			_function.GormDB.W.Create(&dataToInsert)
			_function.SetOption("ver4_lottery_pid", strconv.Itoa(int(account.ID)))
		}
	} else {
		_function.SetOption("ver4_lottery_pid", "0")
	}

	latestDay := _function.GetOption("ver4_lottery_day")
	nowDate := _function.Now.Local().Day()
	if latestDay != strconv.Itoa(nowDate) {
		err := _function.GormDB.W.Where("date <= ?", _function.Now.Add(time.Hour*-24*30).Unix()).Delete(&model.TcVer4LotteryLog{}).Error
		if err != nil {
			log.Println(err)
		}
		_function.SetOption("ver4_lottery_day", nowDate)
	}
}

func (pluginInfo *LotteryPluginPluginType) Install() error {
	for k, v := range LotteryPluginPlugin.Options {
		_function.SetOption(k, v)
	}
	_function.UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	_function.GormDB.W.Migrator().DropTable(&model.TcVer4LotteryLog{})

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcVer4LotteryLog{})

		_function.GormDB.W.Exec("ALTER TABLE `tc_ver4_lottery_log` ADD KEY `pid` (`pid`), ADD KEY `date` (`date`), ADD KEY `pid_date` (`pid`,`date`);")
	} else {
		_function.GormDB.W.Migrator().CreateTable(&model.TcVer4LotteryLog{})

		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_lottery_log_pid" ON "tc_ver4_lottery_log" ("pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_lottery_log_date" ON "tc_ver4_lottery_log" ("date");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_lottery_log_pid_date" ON "tc_ver4_lottery_log" ("pid","date");`)
	}
	return nil
}

func (pluginInfo *LotteryPluginPluginType) Delete() error {
	return nil
}
func (pluginInfo *LotteryPluginPluginType) Upgrade() error {
	return nil
}
func (pluginInfo *LotteryPluginPluginType) Ext() ([]any, error) {
	return []any{}, nil
}
