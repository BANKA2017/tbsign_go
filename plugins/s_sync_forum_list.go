package _plugin

import (
	"net/http"
	"strconv"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/kdnetwork/code-snippet/go/utils"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func init() {
	PluginList.Register(RefreshTiebaListPlugin)
}

type RefreshTiebaListPluginType struct {
	PluginInfo
}

var RefreshTiebaListPlugin = _function.VPtr(RefreshTiebaListPluginType{
	PluginInfo{
		Name:              "ver4_ref",
		PluginNameCN:      "自动同步贴吧列表",
		PluginNameCNShort: "自动同步贴吧列表",
		PluginNameFE:      "",
		Version:           "1.1",
		Options: map[string]string{
			"ver4_ref_day":          "1",
			"ver4_ref_id":           "0",
			"ver4_ref_action_limit": "50",
			"ver4_ref_interval":     "240", // mins
		},
		SettingOptions: map[string]PluginSettingOption{
			"ver4_ref_action_limit": {
				OptionName:   "ver4_ref_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
			"ver4_ref_interval": {
				OptionName:   "ver4_ref_interval",
				OptionNameCN: "每轮同步任务间隔（分钟）",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(1)),
				},
			},
		},
		Endpoints: []PluginEndpointStruct{
			// duplicate endpoints /list/sync ...
			// {Method: http.MethodGet, Path: "list", Function: PluginRefreshTiebaListGetAccountList},
			// {Method: http.MethodPost, Path: "sync", Function: PluginRefreshTiebaListRefreshTiebaList},
			{Method: http.MethodGet, Path: "status", Function: PluginRefreshTiebaListStatus},
		},
	},
})

func (pluginInfo *RefreshTiebaListPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	//activeAfter := 18 //GMT+8 18:00

	// day, _ := strconv.ParseInt(_function.GetOption("ver4_ref_day"), 10, 64)

	// if day != int64(time.Now().Day()) {
	lastdo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
	numRefID, _ := strconv.ParseInt(_function.GetOption("ver4_ref_id"), 10, 64)

	numInterval, _ := strconv.ParseInt(_function.GetOption("ver4_ref_interval"), 10, 64)

	if numInterval < 1 {
		numInterval = 1
	}

	// 4 hours
	if numRefID > 0 || time.Now().Unix() > lastdo+60*numInterval {
		var accounts []*model.TcBaiduid

		numLimit, _ := strconv.ParseInt(_function.GetOption("ver4_ref_action_limit"), 10, 64)

		// hard code limit 1000
		numLimit = utils.Clamp(numLimit, 0, 1000)

		_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id > ?", numRefID).Limit(int(numLimit)).Find(&accounts)

		if len(accounts) == 0 {
			_function.SetOption("ver4_ref_id", "0")
			_function.SetOption("ver4_ref_day", time.Now().Day())
		} else {
			for _, account := range accounts {
				_function.ScanTiebaByPid(account.ID)
				_function.SetOption("ver4_ref_id", int(account.ID))
				_function.SetOption("ver4_ref_lastdo", int(time.Now().Unix()))
			}
		}
	}
	// }
}

func (pluginInfo *RefreshTiebaListPluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")
	return nil
}

func (pluginInfo *RefreshTiebaListPluginType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	return nil
}
func (pluginInfo *RefreshTiebaListPluginType) Upgrade() error {
	return nil
}

func (pluginInfo *RefreshTiebaListPluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	return nil
}

func (pluginInfo *RefreshTiebaListPluginType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *RefreshTiebaListPluginType) Reset(int32, int32, int32) error { return nil }

// endpoint

func PluginRefreshTiebaListStatus(c echo.Context) error {
	numLastDo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
	numInterval, _ := strconv.ParseInt(_function.GetOption("ver4_ref_interval"), 10, 64)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]int64{
		"last_do":  numLastDo,
		"interval": numInterval,
	}, "tbsign"))
}
