//go:build exclude

// ^^^^^^^^^^^^^^^ REMOVE `//go:build exclude`!!!
// This is the template file for standard plugin development

package _plugin

import (
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
)

// 注册插件
func init() {
	RegisterPlugin(ExamplePlugin.Name, ExamplePlugin)
}

type ExamplePluginType struct {
	PluginInfo
}

var ExamplePlugin = _function.VariablePtrWrapper(ExamplePluginType{
	PluginInfo{
		Name:    "kd_example", // 文件名跟 Name 不一定要相同，但都要求唯一
		Version: "1.0",
		Options: map[string]string{
			"key1": "value1",
			"key2": "val2",
			// ...
		},
		Test: true,
		Endpoints: []PluginEndpintStruct{
			{Method: "GET", Path: "example", Function: pluginApiSample},
		},
	},
})

func (pluginInfo *ExamplePluginType) Action() {
	// 计划任务，每分钟都会执行
	// ...
	/// if !pluginInfo.PluginInfo.CheckActive() {
	/// 	return
	/// }
	/// defer pluginInfo.PluginInfo.SetActive(false)

	// 建议恢复上面注释（///）掉的四行，避免重复执行
}

func (pluginInfo *ExamplePluginType) Install() error {
	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	// 上面四行必须要有，如果要处理数据表请在下面继续添加内容

	return nil
}

func (pluginInfo *ExamplePluginType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	// 上面四行必须要有，如果要处理数据表请在下面继续添加内容

	return nil
}
func (pluginInfo *ExamplePluginType) Upgrade() error {
	return nil
}
func (pluginInfo *ExamplePluginType) Ext() ([]any, error) {
	return []any{}, nil
}

// endpoint
func pluginApiSample(c echo.Context) error {
	// uid := c.Get("uid").(string)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _function.EchoEmptyObject, "tbsign"))
}
