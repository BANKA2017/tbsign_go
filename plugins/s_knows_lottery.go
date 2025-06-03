package _plugin

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func init() {
	RegisterPlugin(LotteryPluginPlugin.Name, LotteryPluginPlugin)
}

type LotteryPluginPluginType struct {
	PluginInfo
}

var LotteryPluginPlugin = _function.VariablePtrWrapper(LotteryPluginPluginType{
	PluginInfo{
		Name:              "ver4_lottery",
		PluginNameCN:      "知道商城抽奖",
		PluginNameCNShort: "知道商城",
		PluginNameFE:      "knows_lottery",
		Version:           "1.0",
		Options: map[string]string{
			"ver4_lottery_pid": "0",
			"ver4_lottery_day": "0",
		},
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "switch", Function: PluginKnowsLotteryGetSwitch},
			{Method: "POST", Path: "switch", Function: PluginKnowsLotterySwitch},
			{Method: "GET", Path: "log", Function: PluginKnowsLotteryGetLogs},
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
	return "", errors.New("get_lottery_token: No token")
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

	log.Println(string(response))

	resp := new(GetLotteryResponse)
	err = _function.JsonDecode(response, resp)
	return resp, err
}

var notCompleteActionPid = make(map[int32]struct{}, 0)

func (pluginInfo *LotteryPluginPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id := _function.GetOption("ver4_lottery_pid")

	// 10 am gmt+8
	stime := _function.LocaleTimeDiff(10)

	queryLotteryLogs := _function.GormDB.R.Model(&model.TcVer4LotteryLog{}).Select("pid").Where("date >= ?", stime).Where("pid > ?", id).Group("pid").Having("max(date) >= ? OR COUNT(*) >= ?", _function.Now.Add(time.Minute*-10).Unix(), 2)

	queryUserOptions := _function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid").Where("name='ver4_lottery_check' AND value = '1'")

	var accounts []*model.TcBaiduid

	waitSeconds := 2 // wait 2s// should not <=1
	limit := 60
	if waitSeconds <= 0 {
		waitSeconds = 0
	} else {
		limit = limit / waitSeconds
	}

	_function.GormDB.R.Model(&model.TcBaiduid{}).Select("id", "name", "portrait").Where("id > ? AND id NOT IN (?) AND uid IN (?)", id, queryLotteryLogs, queryUserOptions).Order("id").Limit(limit).Find(&accounts)

	if len(accounts) > 0 {
		for i, account := range accounts {
			if i > 0 {
				time.Sleep(time.Second * time.Duration(waitSeconds))
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

			_, hasNotCompleted := notCompleteActionPid[account.ID]

			resp, err := GetLottery(cookie, token)
			if err != nil && (resp == nil || resp.Data == nil) {
				dataToInsert.Result = "无法解析物品信息"
				if hasNotCompleted {
					delete(notCompleteActionPid, account.ID)
				}
				log.Println(err, resp)
			} else if err != nil && resp.Errno == 0 && len(resp.Data.PrizeList) == 0 {
				if hasNotCompleted {
					dataToInsert.Result = "未完成抽奖"
					delete(notCompleteActionPid, account.ID)
				} else {
					log.Printf("knows_lottery: %d:%s[ %s ] 第一次未完成抽奖\n", account.ID, account.Name, account.Portrait)
					notCompleteActionPid[account.ID] = struct{}{}
				}
			} else if resp.Errno != 0 {
				dataToInsert.Result = resp.Errmsg
				if hasNotCompleted {
					delete(notCompleteActionPid, account.ID)
				}
			} else {
				dataToInsert.Prize = resp.Data.PrizeList[0].GoodsName
				if hasNotCompleted {
					delete(notCompleteActionPid, account.ID)
				}
			}

			_function.GormDB.W.Create(&dataToInsert)
			_function.SetOption("ver4_lottery_pid", strconv.Itoa(int(account.ID)))
		}
	} else {
		_function.SetOption("ver4_lottery_pid", "0")
	}

	latestDay := _function.GetOption("ver4_lottery_day")
	nowDate := _function.Now.Day()
	if latestDay != strconv.Itoa(nowDate) {
		err := _function.GormDB.W.Where("date <= ?", _function.Now.Add(time.Hour*-24*30).Unix()).Delete(&model.TcVer4LotteryLog{}).Error
		if err != nil {
			log.Println(err)
		}
		_function.SetOption("ver4_lottery_day", nowDate)
	}
}

func (pluginInfo *LotteryPluginPluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

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
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)
	_function.GormDB.W.Migrator().DropTable(&model.TcVer4LotteryLog{})

	// user options
	_function.GormDB.W.Where("name = ?", "ver4_lottery_check").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *LotteryPluginPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *LotteryPluginPluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}
	return _sql.Where(_function.AppendStrings(_type, " = ?"), id).Delete(&model.TcVer4LotteryLog{}).Error
}

func (pluginInfo *LotteryPluginPluginType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

// endpoint

func PluginKnowsLotteryGetLogs(c echo.Context) error {
	uid := c.Get("uid").(string)

	var log []*model.TcVer4LotteryLog
	_function.GormDB.R.Model(&model.TcVer4LotteryLog{}).Where("uid = ?", uid).Order("id DESC").Find(&log)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", log, "tbsign"))
}

func PluginKnowsLotteryGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_lottery_check", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("ver4_lottery_check", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginKnowsLotterySwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_lottery_check", uid) != "0"

	err := _function.SetUserOption("ver4_lottery_check", !status, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法修改知道商城抽奖插件状态", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}
