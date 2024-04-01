package _plugin

import (
	"log"
	"net/url"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	tbpb "github.com/BANKA2017/tbsign_go/proto"
	_type "github.com/BANKA2017/tbsign_go/types"
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

var LoopBanPluginName = "ver4_ban"

func PostClientBan(cookie _type.TypeCookie, fid int32, portrait string, day int32, reason string) (*BanAccountResponse, error) {
	isSvipBlock := "0"
	if day == 90 {
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
	_function.AddSign(&form)
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	banResponse, err := _function.Fetch("http://c.tieba.baidu.com/c/c/bawu/commitprison", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), _function.EmptyHeaders)

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

	body, contentType, err := _function.MultipartBodyBuilder(pbBytes)

	if err != nil {
		return nil, err
	}

	resp, err := _function.Fetch("http://tiebac.baidu.com/c/f/forum/getBawuInfo?cmd=301007", "POST", body, map[string]string{
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

func LoopBanAction() {
	id, err := strconv.ParseInt(_function.GetOption("ver4_ban_id"), 10, 64)
	if err != nil {
		id = 0
	}
	otime := _function.Now.Unix() - 86400
	var localBanAccountList = &[]model.TcVer4BanList{}
	subQuery := _function.GormDB.Select("uid").Where("name = 'ver4_ban_open' AND value = '1'").Table("tc_users_options")
	_function.GormDB.Model(&model.TcVer4BanList{}).Where("id > ? AND date < ? AND stime < ? AND etime > ? AND uid IN (?)", id, otime, _function.Now.Unix(), _function.Now.Unix(), subQuery).Order("id ASC").Find(&localBanAccountList)

	var reasonList = &[]model.TcVer4BanUserset{}
	_function.GormDB.Model(&model.TcVer4BanUserset{}).Find(&reasonList)

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

		_function.GormDB.Model(&model.TcVer4BanList{}).Where("id = ?", banAccountInfo.ID).Updates(model.TcVer4BanList{
			Log:  msg,
			Date: int32(_function.Now.Unix()),
		})
		_function.SetOption("ver4_ban_id", strconv.Itoa(int(banAccountInfo.ID)))
	}
	_function.SetOption("ver4_ban_id", "0")

	// clean

}
