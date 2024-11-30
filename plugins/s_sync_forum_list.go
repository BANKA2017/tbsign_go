package _plugin

import (
	"net/http"
	"strconv"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func init() {
	RegisterPlugin(RefreshTiebaListPlugin.Name, RefreshTiebaListPlugin)
}

type RefreshTiebaListPluginType struct {
	PluginInfo
}

var RefreshTiebaListPlugin = _function.VariablePtrWrapper(RefreshTiebaListPluginType{
	PluginInfo{
		Name:              "ver4_ref",
		PluginNameCN:      "自动同步贴吧列表",
		PluginNameCNShort: "自动同步贴吧列表",
		PluginNameFE:      "",
		Version:           "1.0",
		Options: map[string]string{
			"ver4_ref_day":          "1",
			"ver4_ref_id":           "0",
			"ver4_ref_action_limit": "50",
		},
		SettingOptions: map[string]PluinSettingOption{
			"ver4_ref_action_limit": {
				OptionName:   "ver4_ref_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: func(value string) bool {
					numLimit, err := strconv.ParseInt(value, 10, 64)
					return err == nil && numLimit >= 0
				},
			},
		},
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "list", Function: PluginRefreshTiebaListGetAccountList},
			{Method: "POST", Path: "sync", Function: PluginRefreshTiebaListRefreshTiebaList},
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

	// if day != int64(_function.Now.Local().Day()) {
	lastdo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
	refID := _function.GetOption("ver4_ref_id")

	// 4 hours
	if refID != "0" || _function.Now.Unix() > lastdo+60*60*4 {
		var accounts []*model.TcBaiduid

		limit := _function.GetOption("ver4_ref_action_limit")
		numLimit, _ := strconv.ParseInt(limit, 10, 64)
		_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id > ?", refID).Limit(int(numLimit)).Find(&accounts)

		if len(accounts) == 0 {
			_function.SetOption("ver4_ref_id", "0")
			_function.SetOption("ver4_ref_day", strconv.Itoa(_function.Now.Local().Day()))
		} else {
			for _, account := range accounts {
				_function.ScanTiebaByPid(account.ID)
				_function.SetOption("ver4_ref_id", strconv.Itoa(int(account.ID)))
				_function.SetOption("ver4_ref_lastdo", strconv.Itoa(int(_function.Now.Unix())))
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

func (pluginInfo *RefreshTiebaListPluginType) Ext() ([]any, error) {
	return []any{}, nil
}

// endpoint
func PluginRefreshTiebaListGetAccountList(c echo.Context) error {
	uid := c.Get("uid").(string)

	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	var tiebaList []*model.TcTieba
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)

	type accountListResponse struct {
		PID      int32  `json:"pid"`
		Name     string `json:"name"`
		Portrait string `json:"portrait"`
		Count    int32  `json:"count"`
	}

	var response []accountListResponse
	for _, v := range tiebaAccounts {
		var count int32
		for _, v1 := range tiebaList {
			if v1.Pid == v.ID {
				count++
			}
		}
		response = append(response, accountListResponse{
			PID:      v.ID,
			Name:     v.Name,
			Portrait: v.Portrait,
			Count:    count,
		})
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", response, "tbsign"))

}

func PluginRefreshTiebaListRefreshTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.FormValue("pid")

	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	// get account list
	for _, v := range tiebaAccounts {
		if v.ID == int32(numPid) {
			_function.ScanTiebaByPid(v.ID)
			var tiebaList []*model.TcTieba
			_function.GormDB.R.Where("uid = ? AND pid = ?", uid, pid).Order("id ASC").Find(&tiebaList)
			return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaList, "tbsign"))
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(404, "找不到 pid:"+pid, _function.EchoEmptyObject, "tbsign"))
}
