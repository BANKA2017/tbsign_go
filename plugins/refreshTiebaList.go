package _plugin

import (
	"log"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
)

func scanTiebaByPid(pid int32) {
	account := _function.GetCookie(pid)

	var localTiebaList = &[]model.TcTieba{}
	_function.GormDB.Model(&model.TcTieba{ID: account.UID}).Find(&localTiebaList)
	var pn int64 = 1

	for {
		log.Println(pid, pn)
		response, err := _function.GetForumList(account, pn)
		//log.Println(rc, err)
		if err != nil {
			log.Println(err)
			break
		}
		if response.Errmsg != "成功" || len(response.Data.LikeForum.List) <= 0 {
			break
		}
		var tiebaList = []*model.TcTieba{}
		for _, tiebaInfo := range response.Data.LikeForum.List {
			//log.Println(tiebaInfo)
			exists := false
			for _, tiebaInfoDB := range *localTiebaList {
				//合并或被封禁的贴吧会怎样?
				if tiebaInfoDB.Fid == int32(tiebaInfo.ForumID) && tiebaInfoDB.Pid == pid {
					exists = true
					break
				}
			}
			if !exists {
				tiebaList = append(tiebaList, &model.TcTieba{
					Pid:   pid,
					Fid:   int32(tiebaInfo.ForumID),
					UID:   account.UID,
					Tieba: tiebaInfo.ForumName,
				})
			}
		}
		if len(tiebaList) > 0 {
			err := _function.GormDB.Create(tiebaList)
			log.Println(err)
		}
		pn++
		if pn > int64(response.Data.LikeForum.Page.TotalPage) {
			break
		}
	}
}

func RefreshTiebaListAction() {

	activeAfter := 18 //GMT+8 18:00

	day, _ := strconv.ParseInt(_function.GetOption("ver4_ref_day"), 10, 64)
	if day != int64(_function.Now.Local().Day()) {
		lastdo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
		if (_function.Now.Unix() > lastdo+90) && _function.Now.Local().Hour() > activeAfter {
			var accounts = &[]model.TcBaiduid{}
			_function.GormDB.Model(&model.TcBaiduid{}).Find(accounts)
			for _, account := range *accounts {
				scanTiebaByPid(account.ID)
				_function.SetOption("ver4_ref_id", strconv.Itoa(int(account.ID)))
				_function.SetOption("ver4_ref_lastdo", strconv.Itoa(int(_function.Now.Unix())))
			}
			_function.SetOption("ver4_ref_id", "0")
			_function.SetOption("ver4_ref_day", strconv.Itoa(_function.Now.Local().Day()))
		}
	}
}
