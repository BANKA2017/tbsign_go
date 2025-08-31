//go:build exclude

// ^^^^^^^^^^^^^^^ REMOVE `//go:build exclude`!!!
// This is the template file for standard plugin development

package _plugin

import (
	"net/http"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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
		Name:              "kd_example", // 文件名跟 Name 不一定要相同，但都要求唯一
		PluginNameCN:      "示例插件",       // 用于前端插件总开关
		PluginNameCNShort: "示例插件",       // 用于前端侧边栏展示，不建议太长
		PluginNameFE:      "kd_example", // 在前端 `pages/` 的路径，如果没有就留空，允许跟 Name 不同，但要求唯一

		Version: "1.0",
		Options: map[string]string{
			"key1":              "value1",
			"key2":              "val2",
			"key3_action_limit": "50",
			// ...
		},
		// SettingOptions 用于允许前端修改的项目，变量以字符串形式传递，如需在前端显示特殊效果，请修改前端页面
		/// 含有特殊后缀 `_action_limit` 的变量将会在前端自动识别成 number 类型
		SettingOptions: map[string]PluinSettingOption{
			"key2": {
				OptionName:   "key2",
				OptionNameCN: "变量的解释",
				Validate:     PluginExampleOptionValidatorKey2,
			},
			"key3_action_limit": {
				OptionName:   "key3_action_limit",
				OptionNameCN: "这个变量在前端会被识别成数字",
				Validate: func(value string) bool {
					return value == "50" // ...
				},
			},
			// ...
		},
		Test: true,
		Endpoints: []PluginEndpintStruct{
			{Method: http.MethodGet, Path: "example", Function: pluginApiSample},
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

// _type: `uid`, `pid`
func (pluginInfo *ExamplePluginType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	// 清理账号
	// tx 用于事务，但可能为 nil，如果不需要一致性(不建议)的话可以无视直接用 _function.GormDB.W
	return nil
}

func (pluginInfo *ExamplePluginType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *ExamplePluginType) Reset(int32, int32, int32) error {
	return nil
}

// OptionValidator

func PluginExampleOptionValidatorKey2(value string) bool {
	return value == "0" || value == "1"
}

// endpoint
func pluginApiSample(c echo.Context) error {
	// uid := c.Get("uid").(string)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", _function.EchoEmptyObject, "tbsign"))
}
