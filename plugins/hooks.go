package _plugin

import (
	"sync"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
)

var PluginList = make(map[string]PluginActionHooks)

func RegisterPlugin(name string, plugin PluginActionHooks) {
	PluginList[name] = plugin
}

type PluginEndpintStruct struct {
	Method   string
	Path     string
	Function echo.HandlerFunc
}

type PluginInfo struct {
	Name      string
	Version   string
	Active    bool
	Options   map[string]string
	Info      model.TcPlugin
	Test      bool
	Endpoints []PluginEndpintStruct
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
}
