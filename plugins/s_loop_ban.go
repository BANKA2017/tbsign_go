package _plugin

import (
	"bytes"
	"encoding/binary"
	"log"
	"net/url"
	"strconv"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	tbpb "github.com/BANKA2017/tbsign_go/proto"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"
)

type BanAccountResponse struct {
	Un         string `json:"un,omitempty"`
	ServerTime string `json:"server_time,omitempty"`
	Time       int    `json:"time,omitempty"`
	Ctime      int    `json:"ctime,omitempty"`
	Logid      int    `json:"logid,omitempty"`
	ErrorCode  string `json:"error_code,omitempty"`
	ErrorMsg   string `json:"error_msg,omitempty"`
	Info       []any  `json:"info,omitempty"`
}

type IsManagerPreCheckResponse struct {
	IsManager bool   `json:"is_manager"`
	Role      string `json:"role"`
}

type LoopBanPluginType struct {
	PluginInfo
}

var LoopBanPlugin = _function.VariablePtrWrapper(LoopBanPluginType{
	PluginInfo{
		Name:    "ver4_ban",
		Version: "1.4",
		Options: map[string]string{
			"ver4_ban_break_check": "0",
			"ver4_ban_id":          "0",
			"ver4_ban_limit":       "5",
		},
	},
})

var banDays = []int32{1, 3, 10}

func PostClientBan(cookie _type.TypeCookie, fid int32, portrait string, day int32, reason string) (*BanAccountResponse, error) {
	isSvipBlock := "0"
	if day <= 90 && !slices.Contains(banDays, day) {
		isSvipBlock = "1"
	}

	var form = map[string]string{
		"BDUSS":       cookie.Bduss,
		"day":         strconv.Itoa(int(day)),
		"fid":         strconv.Itoa(int(fid)),
		"is_loop_ban": isSvipBlock, // <- Users have to check their svip status in advance
		"ntn":         "banid",
		"portrait":    portrait,
		"reason":      reason,
		"tbs":         cookie.Tbs,
		"word":        "-",
		"z":           "6",
	}
	_function.AddSign(&form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	banResponse, err := _function.TBFetch("http://c.tieba.baidu.com/c/c/bawu/commitprison", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var banDecode BanAccountResponse
	err = _function.JsonDecode(banResponse, &banDecode)
	return &banDecode, err
}

func GetManagerInfo(fid uint64) (*tbpb.GetBawuInfoResIdl_DataRes, error) {
	pbBytes, err := proto.Marshal(&tbpb.GetBawuInfoReqIdl_DataReq{
		Common: &tbpb.CommonReq{
			XClientVersion: _function.ClientVersion,
		},
		Fid: fid,
	})
	if err != nil {
		return nil, err
	}

	pbBytesLen := make([]byte, 8)
	binary.BigEndian.PutUint64(pbBytesLen, uint64(len(pbBytes)))

	body, contentType, err := _function.MultipartBodyBuilder(map[string]any{}, _function.MultipartBodyBinaryFileType{
		Fieldname: "data",
		Filename:  "file",
		Binary:    bytes.Join([][]byte{[]byte("\n"), _function.RemoveLeadingZeros(pbBytesLen), pbBytes}, []byte{}),
	})

	if err != nil {
		return nil, err
	}

	resp, err := _function.TBFetch("http://tiebac.baidu.com/c/f/forum/getBawuInfo?cmd=301007", "POST", body, map[string]string{
		"Content-Type":   contentType,
		"x_bd_data_type": "protobuf",
	})

	if err != nil {
		return nil, err
	}
	// log.Println(resp, string(resp))
	var res tbpb.GetBawuInfoResIdl
	err = proto.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.GetData(), nil
}

func GetManagerStatus(portrait string, fid int64) (*IsManagerPreCheckResponse, error) {
	managerList, _ := GetManagerInfo(uint64(fid))
	for _, v := range managerList.BawuTeamInfo.BawuTeamList {
		if v.RoleName == "吧主助手" {
			continue
		}
		for _, v2 := range v.RoleInfo {
			if v2.Portrait == portrait {
				return &IsManagerPreCheckResponse{
					IsManager: true,
					Role:      v.RoleName,
				}, nil
			}
		}
	}

	return &IsManagerPreCheckResponse{}, nil
}

func (pluginInfo *LoopBanPluginType) Action() {
	if !pluginInfo.PluginInfo.CheckActive() {
		return
	}
	defer pluginInfo.PluginInfo.SetActive(false)

	id, err := strconv.ParseInt(_function.GetOption("ver4_ban_id"), 10, 64)
	if err != nil {
		id = 0
	}
	otime := _function.Now.Add(time.Hour * -24).Unix()
	var localBanAccountList = new([]model.TcVer4BanList)
	subQuery := _function.GormDB.R.Model(&model.TcUsersOption{}).Select("uid").Where("name = 'ver4_ban_open' AND value = '1'")

	// TODO fix hard limit
	_function.GormDB.R.Model(&model.TcVer4BanList{}).Where("id > ? AND date < ? AND stime < ? AND etime > ? AND uid IN (?)", id, otime, _function.Now.Unix(), _function.Now.Unix(), subQuery).Order("id ASC").Limit(50).Find(&localBanAccountList)

	var reasonList = &[]model.TcVer4BanUserset{}
	_function.GormDB.R.Model(&model.TcVer4BanUserset{}).Find(&reasonList)

	for _, banAccountInfo := range *localBanAccountList {
		// find reason
		var reason = "您因为违反吧规，已被吧务封禁，如有疑问请联系吧务！"
		for _, reasonDB := range *reasonList {
			if reasonDB.UID == banAccountInfo.UID && reasonDB.C != "" {
				reason = reasonDB.C
				break
			}
		}

		//get fid
		fid := _function.GetFid(banAccountInfo.Tieba)
		if fid == 0 {
			log.Println("fname: ", banAccountInfo.Tieba, "is not exists!")
			continue
		}

		// !!! warning: unable to check permission !!!
		response, err := PostClientBan(_function.GetCookie(banAccountInfo.Pid), int32(fid), banAccountInfo.Portrait, 1, reason)
		if err != nil {
			log.Println("ban:", err)
			continue
		}
		msg := banAccountInfo.Log
		if response.ErrorMsg != "" {
			msg += _function.Now.Local().Format("2006-01-02 15:04:05") + " 执行结果：<font color=\"red\">操作失败</font>#" + response.ErrorCode + " " + response.ErrorMsg + "<br>"
		} else {
			msg += _function.Now.Local().Format("2006-01-02 15:04:05") + " 执行结果：<font color=\"green\">操作成功</font><br>"
		}

		_function.GormDB.W.Model(&model.TcVer4BanList{}).Where("id = ?", banAccountInfo.ID).Updates(model.TcVer4BanList{
			Log:  msg,
			Date: int32(_function.Now.Unix()),
		})
		_function.SetOption("ver4_ban_id", strconv.Itoa(int(banAccountInfo.ID)))
	}
	_function.SetOption("ver4_ban_id", "0")

	// clean

}

func (pluginInfo *LoopBanPluginType) Install() error {
	for k, v := range LoopBanPlugin.Options {
		_function.SetOption(k, v)
	}
	_function.UpdatePluginInfo(pluginInfo.Name, pluginInfo.Version, false, "")

	_function.GormDB.W.Migrator().DropTable(&model.TcVer4BanUserset{}, &model.TcVer4BanList{})

	// index ?
	if share.DBMode == "mysql" {
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcVer4BanUserset{})
		_function.GormDB.W.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci").Migrator().CreateTable(&model.TcVer4BanList{})
		_function.GormDB.W.Exec("ALTER TABLE `tc_ver4_ban_list` ADD KEY `uid` (`uid`), ADD KEY `id_uid` (`id`,`uid`), ADD KEY `pid` (`pid`), ADD KEY `id_date_stime_etime_uid` (`id`,`date`,`stime`,`etime`,`uid`) USING BTREE;")
		_function.GormDB.W.Exec("ALTER TABLE `tc_ver4_ban_userset` ADD UNIQUE KEY `uid` (`uid`);")
	} else {
		_function.GormDB.W.Set("gorm:table_options", "WITHOUT ROWID").Migrator().CreateTable(&model.TcVer4BanUserset{})
		_function.GormDB.W.Migrator().CreateTable(&model.TcVer4BanList{})

		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_uid" ON "tc_ver4_ban_list" ("uid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_uid" ON "tc_ver4_ban_list" ("id","uid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_pid" ON "tc_ver4_ban_list" ("pid");`)
		_function.GormDB.W.Exec(`CREATE INDEX IF NOT EXISTS "idx_tc_ver4_ban_list_id_date_stime_etime_uid" ON "tc_ver4_ban_list" ("id","date","stime","etime","uid");`)
	}
	return nil
}

func (pluginInfo *LoopBanPluginType) Delete() error {
	return nil
}
func (pluginInfo *LoopBanPluginType) Upgrade() error {
	return nil
}
func (pluginInfo *LoopBanPluginType) Ext() ([]any, error) {
	return []any{}, nil
}
