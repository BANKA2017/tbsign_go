package _plugin

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func init() {
	PluginList.Register(ForumSupportPluginInfo)

	npcList, _ := EmbeddedNPCList.ReadFile("s_forum_support.json")

	json.Unmarshal(npcList, &ForumSupportList)
}

type TypeForumSupportList struct {
	Fid      int    `json:"fid"`
	Nid      int    `json:"nid"`
	Name     string `json:"name"`
	Portrait string `json:"portrait"`
	Tieba    string `json:"tieba"`

	Icon        string `json:"icon"`
	IconName    string `json:"icon_name"`
	IconTitle   string `json:"icon_title"`
	IconPropsID int    `json:"icon_props_id"`
}

type TypeForumSupportResponse struct {
	No    int    `json:"no,omitempty"`
	Error string `json:"error,omitempty"`
	// Data is unused
}

//go:embed s_forum_support.json
var EmbeddedNPCList embed.FS

var ForumSupportList []TypeForumSupportList

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

			{Method: http.MethodGet, Path: "npc/:pid/:fid/items", Function: PluginForumSupportGetNpcItems},
			{Method: http.MethodPost, Path: "npc/:pid/:fid/items/:props_id/claim", Function: PluginForumSupportClaimNpcItem},
			{Method: http.MethodPost, Path: "npc/:pid/:fid/items/:props_id/drop", Function: PluginForumSupportDropNpcItem},
		},
	},
})

func PostForumSupport(cookie *_type.TypeCookie, fid int32, nid string) (*TypeForumSupportResponse, error) {
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

type ClaimForumSupportItemResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
	// Data  struct {
	// 	PropsID        string `json:"props_id"`
	// 	PropsCategory  string `json:"props_category"`
	// 	PropsType      string `json:"props_type"`
	// 	Title          string `json:"title"`
	// 	Name           string `json:"name"`
	// 	Description    string `json:"description"`
	// 	LockNpcLevel   string `json:"lock_npc_level"`
	// 	LockNeedVip    string `json:"lock_need_vip"`
	// 	LockNpcScore   string `json:"lock_npc_score"`
	// 	LimitSignLevel string `json:"limit_sign_level"`
	// 	Scores         string `json:"scores"`
	// 	LogoPicurl     string `json:"logo_picurl"`
	// 	PicURL         string `json:"pic_url"`
	// 	ExampleURL     string `json:"example_url"`
	// 	Pic            string `json:"pic"`
	// 	NpcID          string `json:"npc_id"`
	// 	ForumID        string `json:"forum_id"`
	// 	ForumName      string `json:"forum_name"`
	// 	EndTime        string `json:"end_time"`
	// 	TimeInterval   string `json:"time_interval"`
	// 	CountInterval  string `json:"count_interval"`
	// 	IsCommend      string `json:"is_commend"`
	// 	IsAndroid      string `json:"is_android"`
	// 	IsIos          string `json:"is_ios"`
	// 	IsWap          string `json:"is_wap"`
	// 	IsPc           string `json:"is_pc"`
	// 	PropsState     string `json:"props_state"`
	// 	Status         string `json:"status"`
	// 	MaterialID     string `json:"material_id"`
	// 	CommitUID      string `json:"commit_uid"`
	// 	UpdateTime     string `json:"update_time"`
	// 	CreateTime     string `json:"create_time"`
	// 	Ext1           string `json:"ext1"`
	// 	Ext2           string `json:"ext2"`
	// 	Ext            string `json:"ext"`
	// } `json:"data"`
}

func ClaimForumSupportItem(cookie *_type.TypeCookie, propsID int) (*ClaimForumSupportItemResponse, error) {
	_body := url.Values{
		"tbs":      {cookie.Tbs},
		"props_id": {strconv.Itoa(propsID)},
	}

	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/celebrity/submit/getNpcProps", http.MethodPost, []byte(_body.Encode()), map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var supportDecode ClaimForumSupportItemResponse
	err = _function.JsonDecode(supportResponse, &supportDecode)
	return &supportDecode, err
}

type DropForumSupportItemResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
}

func DropForumSupportItem(cookie *_type.TypeCookie, propsID int) (*DropForumSupportItemResponse, error) {
	_body := url.Values{
		"tbs":      {cookie.Tbs},
		"props_id": {strconv.Itoa(propsID)},
	}

	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/tbmall/post/closeProps", http.MethodPost, []byte(_body.Encode()), map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var supportDecode DropForumSupportItemResponse
	err = _function.JsonDecode(supportResponse, &supportDecode)
	return &supportDecode, err
}

type ForumSupportStatusResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
	Data  []any  `json:"data"`
}

func GetForumSupportStatus(cookie *_type.TypeCookie, fid int32) (*ForumSupportStatusResponse, error) {
	_body := url.Values{
		"tbs":      {cookie.Tbs},
		"forum_id": {strconv.Itoa(int(fid))},
	}

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/celebrity/submit/getForumSupport", http.MethodPost, []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	var supportDecode ForumSupportStatusResponse
	err = _function.JsonDecode(supportResponse, &supportDecode)
	return &supportDecode, err
}

type SetIconResponse struct {
	No    int    `json:"no"`
	Error string `json:"error"`
}

func SetIcon(cookie *_type.TypeCookie, slotNo int, iconName string) (*SetIconResponse, error) {
	_body := url.Values{
		"ie":        {"utf-8"},
		"tbs":       {cookie.Tbs},
		"icon_name": {iconName},
		"slot_no":   {strconv.Itoa(slotNo)},
	}

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/icon/seticon", http.MethodPost, []byte(_body.Encode()), headersMap)

	if err != nil {
		return nil, err
	}

	var supportDecode SetIconResponse
	err = _function.JsonDecode(supportResponse, &supportDecode)
	return &supportDecode, err
}

func SetIcon2(cookie *_type.TypeCookie, slotNo int, iconName string) (*SetIconResponse, error) {
	_body := url.Values{}
	_body.Set("icon_name", iconName)
	_body.Set("slot_no", strconv.Itoa(slotNo))
	_body.Set("tbs", cookie.Tbs)
	_body.Set("ie", "utf-8")

	// no login -> https://passport.baidu.com/v2/?login&u=http%3A%2F%2Ftieba.baidu.com%2Fmo%2Fq%2Ficon%2Fhome
	supportResponse, err := _function.TBFetch("https://tieba.baidu.com/mo/q/icon/setIcon", http.MethodPost, []byte(_body.Encode()), map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var supportDecode SetIconResponse
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

func (pluginInfo *ForumSupportPluginInfoType) ExportAccount(uid int32, tx *gorm.DB) (map[string]any, error) {
	if !pluginInfo.GetSwitch() {
		return nil, nil
	}

	tableName := (&model.TcVer4RankLog{}).TableName()
	var exportData []*model.TcVer4RankLog

	if tx == nil {
		tx = _function.GormDB.R
	}

	err := tx.Model(&model.TcVer4RankLog{}).Where("uid = ?", uid).Find(&exportData).Error

	return map[string]any{
		tableName: exportData,
		"tc_users_options": _function.GetUserOptionBatch(strconv.Itoa(int(uid)), _function.OptionExt{
			Tx:      tx,
			KeyName: "ver4_rank_check",
		}),
	}, err
}

func (pluginInfo *ForumSupportPluginInfoType) ImportAccount(uid int32, pid map[int32]int32, data map[string]json.RawMessage, tx *gorm.DB) error {
	if !pluginInfo.GetSwitch() {
		return errors.New("plugin is not enabled")
	}

	if tx == nil {
		tx = _function.GormDB.W
	}

	tableName := (&model.TcVer4RankLog{}).TableName()

	var data2 []*model.TcVer4RankLog
	if err := _function.JsonDecode(data[tableName], &data2); err != nil {
		return errors.New("invalid data format")
	}

	var data3 []*model.TcVer4RankLog

	var localTasks []*model.TcVer4RankLog
	_function.GormDB.R.Model(&model.TcVer4RankLog{}).Select("pid", "fid").Where("uid = ?", uid).Find(&localTasks)

	pidFidMap := make(map[int32][]int32)

	for _, task := range localTasks {
		if _, ok := pidFidMap[task.Pid]; !ok {
			pidFidMap[task.Pid] = []int32{}
		}
		pidFidMap[task.Pid] = append(pidFidMap[task.Pid], task.Fid)
	}

	for i := range data2 {
		if pid, ok := pid[data2[i].Pid]; ok {
			if fid, ok := pidFidMap[pid]; !ok || !slices.Contains(fid, data2[i].Fid) {
				data2[i].Pid = pid
				data2[i].ID = 0
				data2[i].UID = uid
				data3 = append(data3, data2[i])
			}
		}
	}

	if len(data3) == 0 {
		return nil
	}

	return tx.Model(&model.TcVer4RankLog{}).Create(data3).Error
}

// endpoints
func PluginForumSupportGetCharactersList(c *echo.Context) error {
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumSupportList, "tbsign"))
}
func PluginForumSupportGetSettings(c *echo.Context) error {
	uid := c.Get("uid").(string)

	var rankList []*model.TcVer4RankLog
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&rankList)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", rankList, "tbsign"))
}

func PluginForumSupportUpdateSettings(c *echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusForbidden, _function.ApiTemplate(403, "非法 pid", _function.EchoEmptyObject, "tbsign"))
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
				if _forum.Nid == int(numNid) {
					forum = _forum
					break
				}
			}

			if forum.Nid <= 0 {
				failedList = append(failedList, numNid)
				continue
			}

			addRankList = append(addRankList, &model.TcVer4RankLog{
				UID:   int32(numUID),
				Pid:   int32(numPid),
				Fid:   int32(forum.Fid),
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

func PluginForumSupportGetSwitch(c *echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_rank_check", uid)
	if status == "" {
		status = "0"
		_function.SetUserOption("ver4_rank_check", status, uid)
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status != "0", "tbsign"))
}

func PluginForumSupportSwitch(c *echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_rank_check", uid) != "0"

	err := _function.SetUserOption("ver4_rank_check", !status, uid)

	if err != nil {
		slog.Debug("plugin.forum-support.switch", "uid", uid, "current_status", status, "error", err)
		return c.JSON(http.StatusInternalServerError, _function.ApiTemplate(500, "无法修改名人堂助攻插件状态", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", !status, "tbsign"))
}

type PluginForumSupportNpcItemsBinding struct {
	Pid     uint   `param:"pid"`
	Fid     uint   `param:"fid"`
	PropsID uint   `param:"props_id"`
	Name    string `param:"name"`
	Slot    uint   `param:"slot"`
}

func PluginForumSupportGetNpcItems(c *echo.Context) error {
	uid := c.Get("uid").(string)

	var req PluginForumSupportNpcItemsBinding
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Request", _function.EchoEmptyArray, "tbsign"))
	}

	// check pid
	var user []*model.TcBaiduid
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ? AND id = ?", uid, req.Pid).Find(&user)

	if len(user) == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Pid", _function.EchoEmptyArray, "tbsign"))
	}

	// check fid
	var realFid int
	for _, npcList := range ForumSupportList {
		if npcList.Fid == int(req.Fid) {
			realFid = npcList.Fid
		}
	}
	if realFid == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Fid", _function.EchoEmptyArray, "tbsign"))
	}

	res, err := GetForumSupportStatus(_function.GetCookie(user[0].ID), int32(realFid))
	if err != nil {
		slog.Error("plugin.forum-support.api.status", "error", err)
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, "Response error", _function.EchoEmptyArray, "tbsign"))
	} else if res.No != 0 {
		slog.Error("plugin.forum-support.api.status", "error", errors.New(res.Error))
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, res.Error, _function.EchoEmptyArray, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", res.Data, "tbsign"))
}

func PluginForumSupportClaimNpcItem(c *echo.Context) error {
	uid := c.Get("uid").(string)

	var req PluginForumSupportNpcItemsBinding
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Request", false, "tbsign"))
	}

	// check pid
	var user []model.TcBaiduid
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ? AND id = ?", uid, req.Pid).Find(&user)

	if len(user) == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Pid", false, "tbsign"))
	}

	// check fid
	var realFid int
	for _, npcList := range ForumSupportList {
		if npcList.Fid == int(req.Fid) {
			realFid = npcList.Fid
		}
	}
	if realFid == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Fid", false, "tbsign"))
	}

	res, err := ClaimForumSupportItem(_function.GetCookie(user[0].ID), int(req.PropsID))
	if err != nil {
		slog.Error("plugin.forum-support.api.claim", "error", err)
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, "Response error", false, "tbsign"))
	} else if res.No != 0 {
		slog.Error("plugin.forum-support.api.claim", "error", errors.New(res.Error))
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, res.Error, false, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}

func PluginForumSupportDropNpcItem(c *echo.Context) error {
	uid := c.Get("uid").(string)

	var req PluginForumSupportNpcItemsBinding
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Request", false, "tbsign"))
	}

	// check pid
	var user []model.TcBaiduid
	_function.GormDB.R.Model(&model.TcBaiduid{}).Where("uid = ? AND id = ?", uid, req.Pid).Find(&user)

	if len(user) == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Pid", false, "tbsign"))
	}

	// check fid
	var realFid int
	for _, npcList := range ForumSupportList {
		if npcList.Fid == int(req.Fid) {
			realFid = npcList.Fid
		}
	}
	if realFid == 0 {
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(400, "Invalid Fid", false, "tbsign"))
	}

	res, err := DropForumSupportItem(_function.GetCookie(user[0].ID), int(req.PropsID))
	if err != nil {
		slog.Error("plugin.forum-support.api.drop", "error", err)
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, "Response error", false, "tbsign"))
	} else if res.No != 0 {
		slog.Error("plugin.forum-support.api.drop", "error", errors.New(res.Error))
		return c.JSON(http.StatusBadRequest, _function.ApiTemplate(500, res.Error, false, "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", true, "tbsign"))
}
