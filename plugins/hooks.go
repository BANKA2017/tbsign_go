package _plugin

import (
	"log"
	"sync"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm/clause"
)

var PluginList = make(map[string]PluginActionHooks)
var PluginOptionValidatorMap = sync.Map{}

func RegisterPlugin(name string, plugin PluginActionHooks) {
	PluginList[name] = plugin
}

type PluginEndpintStruct struct {
	Method   string
	Path     string
	Function echo.HandlerFunc
}

type PluginInfo struct {
	Name            string
	Version         string
	Active          bool
	Options         map[string]string
	OptionValidator map[string]func(value string) bool
	Info            model.TcPlugin
	Test            bool
	Endpoints       []PluginEndpintStruct
	sync.RWMutex
}

type PluginHooks interface {
	GetInfo() *PluginInfo
	GetDBInfo() model.TcPlugin
	SetDBInfo(model.TcPlugin) error
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
	// for future
	Ext() ([]any, error)
}

func (pluginInfo *PluginInfo) GetInfo() *PluginInfo {
	return pluginInfo
}

func (pluginInfo *PluginInfo) GetDBInfo() model.TcPlugin {
	return pluginInfo.Info
}

func (pluginInfo *PluginInfo) SetDBInfo(info model.TcPlugin) error {
	pluginInfo.RWMutex.Lock()
	pluginInfo.Info = info
	pluginInfo.RWMutex.Unlock()
	return nil
}

func (pluginInfo *PluginInfo) Switch() bool {
	pluginInfo.RWMutex.Lock()
	pluginInfo.Info.Status = !pluginInfo.Info.Status
	_function.GormDB.W.Model(&model.TcPlugin{}).Where("name = ?", pluginInfo.Name).Update("status", pluginInfo.Info.Status)
	pluginInfo.RWMutex.Unlock()
	return pluginInfo.Info.Status
}

func (pluginInfo *PluginInfo) GetSwitch() bool {
	return pluginInfo.Info.Status
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
	pluginListDB := new([]model.TcPlugin)
	// get plugin list

	pluginNameList := []string{}
	pluginNameSet := sync.Map{}
	for name := range PluginList {
		pluginNameList = append(pluginNameList, name)
		pluginNameSet.Store(name, nil)
	}

	_function.GormDB.R.Where("name in ?", pluginNameList).Find(pluginListDB)

	for _, pluginStatus := range *pluginListDB {
		pluginNameSet.Delete(pluginStatus.Name)
		PluginList[pluginStatus.Name].(PluginHooks).SetDBInfo(pluginStatus)

		// sync option
		for option, optionValue := range PluginList[pluginStatus.Name].(PluginHooks).GetInfo().Options {
			if optionValue != "" && _function.GetOption(option) == "" {
				log.Println("plugin:option:sync:", option, optionValue)
				_function.SetOption(option, optionValue)
			}
		}

		// option validator
		for optionKey, optionValidator := range PluginList[pluginStatus.Name].(PluginHooks).GetInfo().OptionValidator {
			PluginOptionValidatorMap.Store(optionKey, optionValidator)
		}
	}

	pluginNameSet.Range(func(key any, value any) bool {
		pluginNameSet.Delete(key)
		PluginList[key.(string)].(PluginHooks).SetDBInfo(model.TcPlugin{
			Name:    key.(string),
			Status:  false,
			Ver:     "-1",
			Options: "",
		})
		return true
	})
	AddToSettingsFilter()
}

func UpdatePluginInfo(name string, version string, status bool, options string) error {
	// db
	err := _function.GormDB.W.Model(&model.TcPlugin{}).Clauses(clause.OnConflict{UpdateAll: true}).Create(&model.TcPlugin{
		Name:    name,
		Ver:     version,
		Status:  status,
		Options: options,
	}).Error

	if err != nil {
		return err
	}

	// memory cache
	info := PluginList[name].(PluginHooks).GetInfo()
	PluginList[name].(PluginHooks).SetDBInfo(model.TcPlugin{
		Name:    name,
		Status:  false,
		Ver:     info.Version,
		Options: "",
	})

	// option validator
	for optionKey, optionValidator := range PluginList[name].(PluginHooks).GetInfo().OptionValidator {
		PluginOptionValidatorMap.Store(optionKey, optionValidator)
	}
	AddToSettingsFilter()

	return err
}

func DeletePluginInfo(name string) error {
	// memory cache
	PluginList[name].(PluginHooks).SetDBInfo(model.TcPlugin{
		Name:    name,
		Status:  false,
		Ver:     "-1",
		Options: "",
	})

	// option validator
	for optionKey := range PluginList[name].(PluginHooks).GetInfo().OptionValidator {
		PluginOptionValidatorMap.Delete(optionKey)
	}
	AddToSettingsFilter()

	return _function.GormDB.W.Where("name = ?", name).Delete(&model.TcPlugin{}).Error
}

func AddToSettingsFilter() {
	tmpSettingsFilter := _function.SettingsKeys
	PluginOptionValidatorMap.Range(func(key, value any) bool {
		tmpSettingsFilter = append(tmpSettingsFilter, key.(string))
		return true
	})

	_function.SettingsFilter = tmpSettingsFilter
}
