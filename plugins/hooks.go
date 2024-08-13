package _plugin

import (
	"sync"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
)

type PluginInfo struct {
	Name  string
	Hooks PluginHooks
	Info  model.TcPlugin
	mu    *sync.RWMutex
	//APIGroups *echo.Group
}

type PluginHooks interface {
	UpdateInfo() error
	GetInfo() model.TcPlugin
	Switch() bool
	GetSwitch() bool

	Install() error
	Action() //error
	Delete() error
	Update() error
}

func (pluginInfo *PluginInfo) GetInfo() model.TcPlugin {
	return pluginInfo.Info
}

func (pluginInfo *PluginInfo) UpdateInfo(info model.TcPlugin) error {
	//pluginInfo.mu.Lock()
	pluginInfo.Info = info
	//pluginInfo.mu.Unlock()
	return nil
}

func (pluginInfo *PluginInfo) Switch() bool {
	//pluginInfo.mu.Lock()
	pluginInfo.Info.Status = !pluginInfo.Info.Status
	_function.GormDB.W.Model(&model.TcPlugin{}).Where("name = ?", pluginInfo.Name).Update("status", pluginInfo.Info.Status)
	//pluginInfo.mu.Unlock()
	return pluginInfo.Info.Status
}

func (pluginInfo *PluginInfo) GetSwitch() bool {
	return pluginInfo.Info.Status
}

func (pluginInfo *PluginInfo) Install() error {
	return nil
}
func (pluginInfo *PluginInfo) Action() {}
func (pluginInfo *PluginInfo) Delete() error {
	return nil
}
func (pluginInfo *PluginInfo) Update() error {
	return nil
}

var PluginList = map[string]*PluginInfo{
	"ver4_rank": _function.VariablePtrWrapper(ForumSupportPluginInfo.PluginInfo),
	"ver4_ban":  _function.VariablePtrWrapper(LoopBanPlugin.PluginInfo),
	"ver4_ref":  _function.VariablePtrWrapper(RefreshTiebaListPlugin.PluginInfo),
	"kd_growth": _function.VariablePtrWrapper(RefreshTiebaListPlugin.PluginInfo),
}

func InitPluginList() {
	pluginListDB := new([]model.TcPlugin)
	// get plugin list

	pluginNameList := []string{}
	for name := range PluginList {
		pluginNameList = append(pluginNameList, name)
	}

	_function.GormDB.R.Where("name in ?", pluginNameList).Find(pluginListDB)

	for _, pluginStatus := range *pluginListDB {
		if PluginList[pluginStatus.Name].mu == nil {
			PluginList[pluginStatus.Name].mu = new(sync.RWMutex)
		}
		PluginList[pluginStatus.Name].UpdateInfo(pluginStatus)
	}
}
