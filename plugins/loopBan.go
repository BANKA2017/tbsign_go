package _plugin

import (
	"log"
	"net/url"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_type "github.com/BANKA2017/tbsign_go/types"
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

var LoopBanPluginName = "ver4_ban"

func PostClientBan(cookie _type.TypeCookie, fid int32, portrait string, day int32, reason string) (BanAccountResponse, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["day"] = strconv.Itoa(int(day))
	form["fid"] = strconv.Itoa(int(fid))
	form["ntn"] = "banid"
	form["portrait"] = portrait
	form["reason"] = reason
	form["tbs"] = cookie.Tbs
	form["word"] = "-"
	form["z"] = "6"
	_function.AddSign(&form)
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	banResponse, err := _function.Fetch("http://c.tieba.baidu.com/c/c/bawu/commitprison", "POST", _body.Encode()+"&sign="+form["sign"], map[string]string{}, BanAccountResponse{})

	return *banResponse, err
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
