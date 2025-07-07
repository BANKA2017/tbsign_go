package _plugin

import (
	"log"
	"sync"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var PluginList = make(map[string]PluginActionHooks)
var PluginOptionValidatorMap _function.KV[string, func(value string) bool]

func RegisterPlugin(name string, plugin PluginActionHooks) {
	PluginList[name] = plugin
}

type PluginEndpintStruct struct {
	Method   string
	Path     string
	Function echo.HandlerFunc
}

type PluinSettingOption struct {
	OptionName   string
	OptionNameCN string
	Validate     func(value string) bool
}

type PluginInfo struct {
	Name              string
	PluginNameCN      string `json:"plugin_name_cn"`
	PluginNameCNShort string `json:"plugin_name_cn_short"`
	PluginNameFE      string `json:"plugin_name_fe"`

	Version        string
	Active         bool
	Options        map[string]string
	SettingOptions map[string]PluinSettingOption
	Info           model.TcPlugin
	Test           bool
	Endpoints      []PluginEndpintStruct
	sync.Mutex
}

type PluginHooks interface {
	GetInfo() *PluginInfo
	GetDBInfo() model.TcPlugin
	SetDBInfo(*model.TcPlugin) error
	Switch() bool
	GetSwitch() bool
	CheckActive() bool
	SetActive(bool) bool
	GetEndpoints() []PluginEndpintStruct
}

type PluginActionHooks interface {
	Install() error
	Action() //error
	Delete() error
	Upgrade() error
	RemoveAccount(string, int32, *gorm.DB) error
	// ExportAccount(int32) any
	// for future
	Report(int32, *gorm.DB) (string, error)
}

func (pluginInfo *PluginInfo) GetInfo() *PluginInfo {
	return pluginInfo
}

func (pluginInfo *PluginInfo) GetDBInfo() model.TcPlugin {
	return pluginInfo.Info
}

func (pluginInfo *PluginInfo) SetDBInfo(info *model.TcPlugin) error {
	pluginInfo.Mutex.Lock()
	defer pluginInfo.Mutex.Unlock()
	pluginInfo.Info = *info
	return nil
}

func (pluginInfo *PluginInfo) Switch() bool {
	pluginInfo.Mutex.Lock()
	defer pluginInfo.Mutex.Unlock()
	if pluginInfo.Info.Status == 0 {
		pluginInfo.Info.Status = 1
	} else {
		pluginInfo.Info.Status = 0
	}
	_function.GormDB.W.Model(&model.TcPlugin{}).Where("name = ?", pluginInfo.Name).Update("status", pluginInfo.Info.Status)
	return _function.TinyIntToBool(pluginInfo.Info.Status)
}

func (pluginInfo *PluginInfo) GetSwitch() bool {
	return _function.TinyIntToBool(pluginInfo.Info.Status)
}

func (pluginInfo *PluginInfo) CheckActive() bool {
	if pluginInfo.Active {
		return false
	}
	pluginInfo.Active = true
	return true
}

func (pluginInfo *PluginInfo) SetActive(v bool) bool {
	pluginInfo.Active = v
	return v
}

func (pluginInfo *PluginInfo) GetEndpoints() []PluginEndpintStruct {
	return pluginInfo.Endpoints
}

func InitPluginList() {
	var pluginListDB []*model.TcPlugin
	// get plugin list

	pluginNameList := []string{}
	pluginNameSet := make(map[string]bool, len(PluginList))
	for name := range PluginList {
		pluginNameList = append(pluginNameList, name)
		pluginNameSet[name] = true
	}

	_function.GormDB.R.Where("name in ?", pluginNameList).Find(&pluginListDB)

	for _, pluginStatus := range pluginListDB {
		delete(pluginNameSet, pluginStatus.Name)
		PluginList[pluginStatus.Name].(PluginHooks).SetDBInfo(pluginStatus)

		// sync option
		for option, optionValue := range PluginList[pluginStatus.Name].(PluginHooks).GetInfo().Options {
			if optionValue != "" && _function.GetOption(option) == "" {
				log.Println("plugin:option:sync:", option, optionValue)
				_function.SetOption(option, optionValue)
			}
		}

		// option validator
		for optionKey, optionValidator := range PluginList[pluginStatus.Name].(PluginHooks).GetInfo().SettingOptions {
			PluginOptionValidatorMap.Store(optionKey, optionValidator.Validate, -1)
		}
	}
	for key := range pluginNameSet {
		delete(pluginNameSet, key)
		PluginList[key].(PluginHooks).SetDBInfo(&model.TcPlugin{
			Name:    key,
			Status:  0,
			Ver:     "-1",
			Options: "",
		})
	}
	AddToSettingsFilter()
}

func UpdatePluginInfo(name string, version string, status bool, options string) error {
	// db
	err := _function.GormDB.W.Model(&model.TcPlugin{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcPlugin{
		Name:    name,
		Ver:     version,
		Status:  _function.BoolToTinyInt(status),
		Options: options,
	}).Error

	if err != nil {
		return err
	}

	// memory cache
	info := PluginList[name].(PluginHooks).GetInfo()
	PluginList[name].(PluginHooks).SetDBInfo(&model.TcPlugin{
		Name:    name,
		Status:  0,
		Ver:     info.Version,
		Options: "",
	})

	// option validator
	for optionKey, optionValidator := range PluginList[name].(PluginHooks).GetInfo().SettingOptions {
		PluginOptionValidatorMap.Store(optionKey, optionValidator.Validate, -1)
	}
	AddToSettingsFilter()

	return err
}

func DeletePluginInfo(name string) error {
	// memory cache
	PluginList[name].(PluginHooks).SetDBInfo(&model.TcPlugin{
		Name:    name,
		Status:  0,
		Ver:     "-1",
		Options: "",
	})

	// option validator
	for optionKey := range PluginList[name].(PluginHooks).GetInfo().SettingOptions {
		PluginOptionValidatorMap.Delete(optionKey)
	}
	AddToSettingsFilter()

	return _function.GormDB.W.Where("name = ?", name).Delete(&model.TcPlugin{}).Error
}

func AddToSettingsFilter() {
	tmpSettingsFilter := _function.SettingsKeys
	PluginOptionValidatorMap.Range(func(key string, value func(value string) bool) bool {
		tmpSettingsFilter = append(tmpSettingsFilter, key)
		return true
	})

	_function.SettingsFilter = tmpSettingsFilter
}

func DeleteAccount(_type string, id int32, tx *gorm.DB) error {
	for _, p := range PluginList {
		if p.(PluginHooks).GetDBInfo().Ver != "-1" {
			if err := p.RemoveAccount(_type, id, tx); err != nil {
				return err
			}
		}
	}
	return nil
}
