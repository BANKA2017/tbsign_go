package _plugin

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func init() {
	PluginList.Register(ForumSupportPluginInfo)
}

type TypeForumSupportList struct {
	Fid   string `json:"fid"`
	Nid   int64  `json:"nid"`
	Name  string `json:"name"`
	Tieba string `json:"tieba"`
}

type TypeForumSupportResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	// Data is unused
}

var ForumSupportList = []TypeForumSupportList{
	{
		Fid:   "280050",
		Nid:   2257282262,
		Name:  "最强王者",
		Tieba: "lol",
	},
	{
		Fid:   "11772",
		Nid:   1337628265,
		Name:  "路飞",
		Tieba: "海贼王",
	},
	{
		Fid:   "8230522",
		Nid:   1337623685,
		Name:  "贴吧用户_QRNVQaG",
		Tieba: "风暴英雄",
	},
	{
		Fid:   "10866131",
		Nid:   1866273558,
		Name:  "月老嘉哥",
		Tieba: "英三嘉哥",
	},
	{
		Fid:   "81570",
		Nid:   2257269894,
		Name:  "赛丽亚😇",
		Tieba: "地下城与勇士",
	},
	{
		Fid:   "401299",
		Nid:   1337623678,
		Name:  "戒色守望者",
		Tieba: "戒色",
	},
	{
		Fid:   "711567",
		Nid:   1337628263,
		Name:  "黄鱼哥",
		Tieba: "内涵",
	},
	{
		Fid:   "110019",
		Nid:   1337628266,
		Name:  "漩涡鸣人",
		Tieba: "火影忍者",
	},
	{
		Fid:   "43927",
		Nid:   2257275375,
		Name:  "剑侠客😇",
		Tieba: "梦幻西游",
	},
	{
		Fid:   "339",
		Nid:   2257275437,
		Name:  "江户川柯南😇",
		Tieba: "柯南",
	},
	{
		Fid:   "1701120",
		Nid:   2257275573,
		Name:  "陆逊😇",
		Tieba: "三国杀",
	},
	{
		Fid:   "30227",
		Nid:   1337623700,
		Name:  "灰原哀",
		Tieba: "灰原哀",
	},
	{
		Fid:   "2862817",
		Nid:   1337623686,
		Name:  "炉石旅店萌板娘",
		Tieba: "炉石传说",
	},
	{
		Fid:   "122873",
		Nid:   1337628268,
		Name:  "黑崎一护",
		Tieba: "死神",
	},
	{
		Fid:   "738100",
		Nid:   1337623692,
		Name:  "初音ミク",
		Tieba: "初音ミク",
	},
	{
		Fid:   "1627732",
		Nid:   2257275681,
		Name:  "圣堂刺客😇",
		Tieba: "dota2",
	},
	{
		Fid:   "13839385",
		Nid:   2257269364,
		Name:  "大喵😇",
		Tieba: "奇迹暖暖",
	},
	{
		Fid:   "1525417",
		Nid:   1337623679,
		Name:  "张起灵",
		Tieba: "盗墓笔记",
	},
	{
		Fid:   "667580",
		Nid:   1337623697,
		Name:  "坂田银时",
		Tieba: "银魂",
	},
	{
		Fid:   "1111175",
		Nid:   1337628284,
		Name:  "贴吧用户_QRNVGb3",
		Tieba: "黑丝",
	},
	{
		Fid:   "574961",
		Nid:   1337628277,
		Name:  "夏目贵志",
		Tieba: "夏目友人帐",
	},
	{
		Fid:   "2358322",
		Nid:   1337623698,
		Name:  "遮天叶凡",
		Tieba: "遮天",
	},
	{
		Fid:   "11760190",
		Nid:   2257282271,
		Name:  "宁海😇",
		Tieba: "战舰少女",
	},
	{
		Fid:   "5024455",
		Nid:   1337628281,
		Name:  "暖暖",
		Tieba: "暖暖环游世界",
	},
	{
		Fid:   "1089593",
		Nid:   1337623683,
		Name:  "荆天明",
		Tieba: "秦时明月",
	},
	{
		Fid:   "501133",
		Nid:   2257282249,
		Name:  "SerB😇",
		Tieba: "坦克世界",
	},
	{
		Fid:   "1786526",
		Nid:   1337628286,
		Name:  "高坂穗乃果",
		Tieba: "lovelive",
	},
	{
		Fid:   "47320",
		Nid:   1337623693,
		Name:  "沢田纲吉",
		Tieba: "家庭教师",
	},
	{
		Fid:   "15975428",
		Nid:   1337623688,
		Name:  "",
		Tieba: "大话西游手游",
	},
	{
		Fid:   "2432903",
		Nid:   2257269753,
		Name:  "天空之城TC😇",
		Tieba: "minecraft",
	},
	{
		Fid:   "1498934",
		Nid:   1337628261,
		Name:  "纳兹",
		Tieba: "妖精的尾巴",
	},
	{
		Fid:   "2941239",
		Nid:   2257269548,
		Name:  "洛天依😇",
		Tieba: "洛天依",
	},
	{
		Fid:   "27829",
		Nid:   2257269560,
		Name:  "桔梗😇",
		Tieba: "桔梗",
	},
	{
		Fid:   "16779",
		Nid:   1337623680,
		Name:  "素还真",
		Tieba: "霹雳",
	},
	{
		Fid:   "1564063",
		Nid:   2257275856,
		Name:  "御坂美琴",
		Tieba: "御坂美琴",
	},
	{
		Fid:   "551358",
		Nid:   2257275816,
		Name:  "saber",
		Tieba: "saber",
	},
	{
		Fid:   "185228",
		Nid:   2257269662,
		Name:  "龙马😇",
		Tieba: "网球王子",
	},
	{
		Fid:   "46797",
		Nid:   2257269533,
		Name:  "工藤新一",
		Tieba: "工藤新一",
	},
	{
		Fid:   "21473",
		Nid:   1337628276,
		Name:  "毛利兰",
		Tieba: "毛利兰",
	},
	{
		Fid:   "1299394",
		Nid:   1337623690,
		Name:  "吴邪",
		Tieba: "吴邪",
	},
	{
		Fid:   "1319679",
		Nid:   2257275577,
		Name:  "皮卡超人😇",
		Tieba: "部落战争",
	},
	{
		Fid:   "1321783",
		Nid:   2257275225,
		Name:  "夏尔😇",
		Tieba: "黑执事",
	},
	{
		Fid:   "5437659",
		Nid:   1337623684,
		Name:  "kiana",
		Tieba: "崩坏学园2",
	},
	{
		Fid:   "216681",
		Nid:   2257269425,
		Name:  "古河渚😇",
		Tieba: "clannad",
	},
	{
		Fid:   "9714340",
		Nid:   1337628279,
		Name:  "孟浩",
		Tieba: "我欲封天",
	},
	{
		Fid:   "78279",
		Nid:   1337623689,
		Name:  "伏见猿比古",
		Tieba: "k",
	},
	{
		Fid:   "14823520",
		Nid:   2257269780,
		Name:  "刀锋😇",
		Tieba: "cf手游",
	},
	{
		Fid:   "1866137",
		Nid:   2257269448,
		Name:  "鹿目圆香😇",
		Tieba: "魔法少女小圆",
	},
	{
		Fid:   "2099286",
		Nid:   2257269826,
		Name:  "夜刀神十香😇",
		Tieba: "datealive",
	},
	{
		Fid:   "1575589",
		Nid:   1337623701,
		Name:  "东方爱",
		Tieba: "浪漫传说",
	},
	{
		Fid:   "154782",
		Nid:   2257282283,
		Name:  "江户川哀😇",
		Tieba: "柯哀",
	},
	{
		Fid:   "4300",
		Nid:   2257282328,
		Name:  "奇犽·揍敌客😇",
		Tieba: "全职猎人",
	},
	{
		Fid:   "10254689",
		Nid:   2257269311,
		Name:  "绚濑绘里😇",
		Tieba: "lovelive国服",
	},
	{
		Fid:   "10631925",
		Nid:   2257275308,
		Name:  "温文儒雅阳光😇",
		Tieba: "炫舞时代",
	},
	{
		Fid:   "4295466",
		Nid:   1337623694,
		Name:  "柳鸣",
		Tieba: "魔天记",
	},
	{
		Fid:   "149985",
		Nid:   2257275387,
		Name:  "日向雏田😇",
		Tieba: "雏田",
	},
	{
		Fid:   "957111",
		Nid:   2257275941,
		Name:  "枫音",
		Tieba: "中萌",
	},
	{
		Fid:   "347203",
		Nid:   2257275921,
		Name:  "工藤兰😇",
		Tieba: "新兰",
	},
	{
		Fid:   "644596",
		Nid:   2257275770,
		Name:  "勒鲁什😇",
		Tieba: "叛逆的勒鲁什",
	},
	{
		Fid:   "701877",
		Nid:   2072780523,
		Name:  "如果爱粉红帖😇",
		Tieba: "如果爱",
	},
	{
		Fid:   "420456",
		Nid:   2257275411,
		Name:  "风铃😇",
		Tieba: "世萌",
	},
	{
		Fid:   "138887",
		Nid:   2257275510,
		Name:  "不二周助",
		Tieba: "不二周助",
	},
	{
		Fid:   "711389",
		Nid:   2257269774,
		Name:  "宇智波佐助😇",
		Tieba: "佐助",
	},
	{
		Fid:   "60129",
		Nid:   2257269964,
		Name:  "怪盗基德😇",
		Tieba: "怪盗基德",
	},
	{
		Fid:   "128494",
		Nid:   2257269820,
		Name:  "旗木卡卡西😇",
		Tieba: "卡卡西",
	},
	{
		Fid:   "1773771",
		Nid:   2257275519,
		Name:  "时崎狂三😇",
		Tieba: "时崎狂三",
	},
	{
		Fid:   "2812935",
		Nid:   1337623681,
		Name:  "金木研",
		Tieba: "东京食尸鬼",
	},
	{
		Fid:   "2297729",
		Nid:   2257275832,
		Name:  "小奏😇",
		Tieba: "立华奏",
	},
	{
		Fid:   "493375",
		Nid:   2257269969,
		Name:  "宇智波鼬",
		Tieba: "宇智波鼬",
	},
	{
		Fid:   "2757769",
		Nid:   2257275212,
		Name:  "闪光亚丝娜😇",
		Tieba: "亚丝娜",
	},
	{
		Fid:   "2515521",
		Nid:   2257269875,
		Name:  "利威尔阿克曼😇",
		Tieba: "利威尔",
	},
	{
		Fid:   "2966494",
		Nid:   2257275688,
		Name:  "神兽萌萌😇",
		Tieba: "炫舞2",
	},
	{
		Fid:   "1550649",
		Nid:   2257269996,
		Name:  "鲁鲁修😇",
		Tieba: "鲁鲁修",
	},
	{
		Fid:   "225307",
		Nid:   2257269677,
		Name:  "越前龙马😇",
		Tieba: "越前龙马",
	},
	{
		Fid:   "13913",
		Nid:   2257269464,
		Name:  "一露😇",
		Tieba: "一露",
	},
	{
		Fid:   "799600",
		Nid:   2257269684,
		Name:  "春野樱😇",
		Tieba: "春野樱",
	},
	{
		Fid:   "1308858",
		Nid:   1337628275,
		Name:  "盖聂",
		Tieba: "盖聂",
	},
	{
		Fid:   "4266803",
		Nid:   2257275502,
		Name:  "次元酱😇",
		Tieba: "二次元界",
	},
	{
		Fid:   "2083514",
		Nid:   2257269862,
		Name:  "秦时明月少司😇",
		Tieba: "少司命",
	},
	{
		Fid:   "2530050",
		Nid:   2257275916,
		Name:  "黄濑凉太😇",
		Tieba: "黄濑凉太",
	},
	{
		Fid:   "1278749",
		Nid:   2257275638,
		Name:  "C.C.😇",
		Tieba: "c.c.",
	},
	{
		Fid:   "11768634",
		Nid:   1337623691,
		Name:  "暴雨心奴",
		Tieba: "罪雨台",
	},
	{
		Fid:   "1446622",
		Nid:   2257275490,
		Name:  "卫庄😇",
		Tieba: "卫庄",
	},
	{
		Fid:   "683722",
		Nid:   1337623696,
		Name:  "雾岛董香",
		Tieba: "雾岛董香",
	},
	{
		Fid:   "1590884",
		Nid:   2257275344,
		Name:  "张良",
		Tieba: "秦时明月张良",
	},
	{
		Fid:   "2485436",
		Nid:   2257269581,
		Name:  "西木野真姬",
		Tieba: "西木野真姬",
	},
	{
		Fid:   "280289",
		Nid:   2257269626,
		Name:  "娜美😇",
		Tieba: "娜美",
	},
	{
		Fid:   "361521",
		Nid:   2257275803,
		Name:  "白凤😇",
		Tieba: "白凤",
	},
	{
		Fid:   "1310236",
		Nid:   2257269636,
		Name:  "端木蓉😇",
		Tieba: "端木蓉",
	},
	{
		Fid:   "2827493",
		Nid:   2257269567,
		Name:  "高渐离😇",
		Tieba: "秦时明月all高",
	},
	{
		Fid:   "2181095",
		Nid:   1337623682,
		Name:  "戒撸吧小管家",
		Tieba: "戒撸",
	},
}

type ForumSupportPluginInfoType struct {
	PluginInfo
}

var ForumSupportPluginInfo = _function.VPtr(ForumSupportPluginInfoType{
	PluginInfo{
		Name:              "ver4_rank",
		PluginNameCN:      "贴吧名人堂助攻",
		PluginNameCNShort: "名人堂",
		PluginNameFE:      "forum_support",
		Version:           "1.2",
		Options: map[string]string{
			"ver4_rank_daily":        "1",
			"ver4_rank_id":           "0",
			"ver4_rank_action_limit": "50",
		},
		SettingOptions: map[string]PluginSettingOption{
			"ver4_rank_action_limit": {
				OptionName:   "ver4_rank_action_limit",
				OptionNameCN: "每分钟最大执行数",
				Validate: &_function.OptionRule{
					Min: _function.VPtr(int64(0)),
				},
			},
		},
		Endpoints: []PluginEndpointStruct{
			{Method: http.MethodGet, Path: "switch", Function: PluginForumSupportGetSwitch},
			{Method: http.MethodPost, Path: "switch", Function: PluginForumSupportSwitch},
			{Method: http.MethodGet, Path: "list", Function: PluginForumSupportGetCharactersList},
			{Method: http.MethodGet, Path: "settings", Function: PluginForumSupportGetSettings},
			{Method: http.MethodPut, Path: "settings", Function: PluginForumSupportUpdateSettings},
		},
	},
})

func PostForumSupport(cookie _type.TypeCookie, fid int32, nid string) (*TypeForumSupportResponse, error) {
	_body := url.Values{
		"tbs":      {cookie.Tbs},
		"forum_id": {strconv.Itoa(int(fid))},
		"npc_id":   {nid},
	}

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	// when login is invalid, 302 to http://static.tieba.baidu.com/tb/error.html
	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/celebrity/submit/support", http.MethodPost, []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	var supportDecode TypeForumSupportResponse
	err = _function.JsonDecode(supportResponse, &supportDecode)
	return &supportDecode, err
}

func (pluginInfo *ForumSupportPluginInfoType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	batchMode := _function.GetOption("go_plugin_batch_tasks") == "1"
	id, err := strconv.ParseInt(_function.GetOption("ver4_rank_id"), 10, 64)
	if err != nil {
		id = 0
	}
	// status list
	var accountStatusList = make(map[int32]string)

	// get list
	todayBeginning := _function.LocaleTimeDiff(0) //GMT+8
	var ver4RankLog []*model.TcVer4RankLog

	limit := _function.GetOption("ver4_rank_action_limit")
	numLimit, _ := strconv.ParseInt(limit, 10, 64)

	if batchMode {
		BatchPluginQuery(_function.GormDB.R.Model(&model.TcVer4RankLog{}).Where("date < ?", todayBeginning), int(numLimit), 3, []string{"*"}, &ver4RankLog)
	} else {
		_function.GormDB.R.Model(&model.TcVer4RankLog{}).Where("date < ? AND id > ?", todayBeginning, id).Limit(int(numLimit)).Find(&ver4RankLog)
	}

	for _, forumSupportItem := range ver4RankLog {
		if _, ok := accountStatusList[forumSupportItem.UID]; !ok {
			accountStatusList[forumSupportItem.UID] = _function.GetUserOption("ver4_rank_check", strconv.Itoa(int(forumSupportItem.UID)))
		}
		if accountStatusList[forumSupportItem.UID] == "" {
			// clean
			_function.GormDB.W.Where("uid = ?", forumSupportItem.UID).Delete(&model.TcVer4RankLog{})
			accountStatusList[forumSupportItem.UID] = "NOT_EXISTS"
		} else if accountStatusList[forumSupportItem.UID] == "1" {
			cookie := _function.GetCookie(forumSupportItem.Pid)
			message := ""
			errno := 0 // response.No / -1(login) / -2(request error)

			if !cookie.IsLogin {
				errno = -1
				message = "账号未登录，跳过助攻~"
				slog.Error("账号未登录，跳过助攻 (plugin.forum-support.action.login)", "tieba", forumSupportItem.Tieba, "name", forumSupportItem.Name, "pid", forumSupportItem.Pid)
			} else {
				response, err := PostForumSupport(cookie, forumSupportItem.Fid, forumSupportItem.Nid)

				if err != nil {
					errno = -2
					message = "助攻失败，发生了一些未知错误~"
					slog.Error(message+" (plugin.forum-support.action)", "tieba", forumSupportItem.Tieba, "name", forumSupportItem.Name, "error", err, "pid", forumSupportItem.Pid)
				} else {
					errno = response.No
					switch response.No {
					case 0:
						message = "助攻成功啦~明天记得继续呦~"
					case 340027:
						message = "很抱歉，封禁用户无法助攻"
					case 3110004:
						message = "你还未关注当前吧哦, 快去关注吧~"
					case 2280006:
						message = "今日已助攻过了，或者度受抽风了~"
					default:
						message = "抽风了~"
					}

					if response.No != 0 {
						slog.Error(message+" (plugin.forum-support.action)", "tieba", forumSupportItem.Tieba, "name", forumSupportItem.Name, "code", response.No, "error", response.Error, "pid", forumSupportItem.Pid)
					}
				}
			}

			// previous logs
			previousLogs := []string{}
			for i, s := range strings.Split(forumSupportItem.Log, "<br/>") {
				if i <= 30 {
					previousLogs = append(previousLogs, s)
				} else {
					break
				}
			}

			_function.GormDB.W.Model(&model.TcVer4RankLog{}).Where("id = ?", forumSupportItem.ID).Updates(model.TcVer4RankLog{
				Log:  fmt.Sprintf("<br/>%s #%d,%s%s", time.Now().Format(time.DateOnly), errno, message, strings.Join(previousLogs, "<br/>")),
				Date: int32(time.Now().Unix()),
			})

			if !batchMode {
				_function.SetOption("ver4_rank_id", strconv.Itoa(int(forumSupportItem.ID)))
			}
		}
	}
	if !batchMode {
		_function.SetOption("ver4_rank_id", "0")
	}
}

func (pluginInfo *ForumSupportPluginInfoType) Install() error {
	var err error

	for k, v := range pluginInfo.Options {
		_function.SetOption(k, v)
	}
	err = UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")
	if err != nil {
		return err
	}

	return _function.GormDB.W.Migrator().CreateTable(&model.TcVer4RankLog{})
}

func (pluginInfo *ForumSupportPluginInfoType) Delete() error {
	for k := range pluginInfo.Options {
		_function.DeleteOption(k)
	}
	DeletePluginInfo(pluginInfo.Name)

	_function.GormDB.W.Migrator().DropTable(&model.TcVer4RankLog{})

	// user options
	_function.GormDB.W.Where("name = ?", "ver4_rank_check").Delete(&model.TcUsersOption{})

	return nil
}
func (pluginInfo *ForumSupportPluginInfoType) Upgrade() error {
	return nil
}

func (pluginInfo *ForumSupportPluginInfoType) RemoveAccount(_type string, id int32, tx *gorm.DB) error {
	_sql := _function.GormDB.W
	if tx != nil {
		_sql = tx
	}
	return _sql.Where(_type+" = ?", id).Delete(&model.TcVer4RankLog{}).Error
}

func (pluginInfo *ForumSupportPluginInfoType) Report(int32, *gorm.DB) (string, error) {
	return "", nil
}

func (pluginInfo *ForumSupportPluginInfoType) Reset(uid, pid, tid int32) error {
	if uid == 0 {
		return errors.New("invalid uid")
	}

	_sql := _function.GormDB.W.Model(&model.TcVer4RankLog{}).Where("uid = ?", uid)
	if pid != 0 {
		_sql = _sql.Where("pid = ?", pid)
	}

	if tid != 0 {
		_sql = _sql.Where("id = ?", tid)
	}

	return _sql.Update("date", 0).Error
}

// endpoints
func PluginForumSupportGetCharactersList(c echo.Context) error {
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumSupportList, "tbsign"))
}
func PluginForumSupportGetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	var rankList []*model.TcVer4RankLog
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&rankList)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", rankList, "tbsign"))
}

func PluginForumSupportUpdateSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "非法 pid", _function.EchoEmptyObject, "tbsign"))
	}

	var rankList []*model.TcVer4RankLog
	_function.GormDB.R.Where("uid = ? AND pid = ?", uid, pid).Order("id ASC").Find(&rankList)

	c.Request().ParseForm()

	nid := c.Request().Form["nid[]"]

	var addRankList []*model.TcVer4RankLog
	var delRankList []*model.TcVer4RankLog
	var delRankIDList []int32
	var failedList []int64

	// add
	for _, v := range nid {
		exist := false
		for _, v1 := range rankList {
			if v1.Nid == v {
				exist = true
			}
		}
		if !exist {
			numNid, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				failedList = append(failedList, numNid)
				continue
			}

			// fid nid
			var forum TypeForumSupportList
			for _, _forum := range ForumSupportList {
				if _forum.Nid == numNid {
					forum = _forum
					break
				}
			}

			if forum.Nid <= 0 {
				failedList = append(failedList, numNid)
				continue
			}

			numFid, _ := strconv.ParseInt(forum.Fid, 10, 64)

			addRankList = append(addRankList, &model.TcVer4RankLog{
				UID:   int32(numUID),
				Pid:   int32(numPid),
				Fid:   int32(numFid),
				Nid:   v,
				Name:  forum.Name,
				Tieba: forum.Tieba,
				Log:   "",
				Date:  0,
			})
		}
	}

	if len(addRankList) > 0 {
		_function.GormDB.W.Create(&addRankList)
	}

	// del
	for _, v := range rankList {
		exist := false
		for _, v1 := range nid {
			if v.Nid == v1 {
				exist = true
			}
		}
		if !exist {
			delRankList = append(delRankList, v)
			delRankIDList = append(delRankIDList, v.ID)
		}
	}

	_function.GormDB.W.Where("id IN ?", delRankIDList).Delete(&model.TcVer4RankLog{})

	var resp = struct {
		Add    []*model.TcVer4RankLog `json:"add"`
		Del    []*model.TcVer4RankLog `json:"del"`
		Failed []int64                `json:"failed"`
	}{
		Add:    addRankList,
		Del:    delRankList,
		Failed: failedList,
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", resp, "tbsign"))
}

func PluginForumSupportGetSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_rank_check", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("ver4_rank_check", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginForumSupportSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_rank_check", uid) != "0"

	err := _function.SetUserOption("ver4_rank_check", !status, uid)

	if err != nil {
		slog.Debug("plugin.forum-support.switch", "uid", uid, "current_status", status, "error", err)
		return c.JSON(http.StatusOK, _function.ApiTemplate(500, "无法修改名人堂助攻插件状态", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}
