package _plugin

import (
	"log"
	"slices"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
)

var RefreshTiebaListPluginName = "ver4_ref"

func ScanTiebaByPid(pid int32) {
	account := _function.GetCookie(pid)

	var localTiebaList = &[]model.TcTieba{}
	_function.GormDB.R.Model(&model.TcTieba{Pid: account.ID}).Find(&localTiebaList)

	localTiebaFidList := []int{}

	for _, v := range *localTiebaList {
		localTiebaFidList = append(localTiebaFidList, int(v.Fid))
	}

	var pn int64 = 1

	wholeTiebaFidList := []int32{}

	for {
		//log.Println(pid, pn)
		response, err := _function.GetWebForumList(account, pn)
		//log.Println(rc, err)
		if err != nil {
			log.Println("scanTiebaByPid:", err)
			break
		}
		if response.Errmsg != "成功" || len(response.Data.LikeForum.List) <= 0 {
			break
		}
		var tiebaList = []*model.TcTieba{}
		for _, tiebaInfo := range response.Data.LikeForum.List {
			if tiebaInfo.ForumID == 0 {
				continue
			}
			//log.Println(tiebaInfo)
			//合并或被封禁的贴吧会怎样?

			tmpTcTieba := &model.TcTieba{
				Pid:   pid,
				Fid:   int32(tiebaInfo.ForumID),
				UID:   account.UID,
				Tieba: tiebaInfo.ForumName,
			}

			if !slices.Contains(localTiebaFidList, tiebaInfo.ForumID) && !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
				tiebaList = append(tiebaList, tmpTcTieba)
				wholeTiebaFidList = append(wholeTiebaFidList, tmpTcTieba.Fid)
			} else if !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
				wholeTiebaFidList = append(wholeTiebaFidList, tmpTcTieba.Fid)
			}
		}
		if len(tiebaList) > 0 {
			err := _function.GormDB.W.Create(tiebaList).Error
			if err != nil {
				log.Println("scanTiebaByPid:", err)
			}
		}

		pn++
		// 30 * 200 -> 6000
		// avoid loop
		if pn > int64(response.Data.LikeForum.Page.TotalPage) || pn > 20 {
			break
		}
	}

	accountInfo, err := _function.GetBaiduUserInfo(account)
	if err == nil {
		pn = 1
		for {
			response, err := _function.GetForumList(account, accountInfo.User.ID, pn)
			//log.Println(rc, err)
			if err != nil {
				log.Println("scanTiebaByPid:", err)
				break
			}
			if response.ErrorCode != "0" || len(response.ForumList.NonGconforum) <= 0 {
				break
			}
			var tiebaList = []*model.TcTieba{}
			for _, tiebaInfo := range response.ForumList.NonGconforum {
				//log.Println(tiebaInfo)
				//合并或被封禁的贴吧会怎样?

				numFID, _ := strconv.ParseInt(tiebaInfo.ID, 10, 64)
				if numFID == 0 {
					continue
				}

				tmpTcTieba := &model.TcTieba{
					Pid:   pid,
					Fid:   int32(numFID),
					UID:   account.UID,
					Tieba: tiebaInfo.Name,
				}

				if !slices.Contains(localTiebaFidList, int(numFID)) && !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
					tiebaList = append(tiebaList, tmpTcTieba)
					wholeTiebaFidList = append(wholeTiebaFidList, tmpTcTieba.Fid)
				} else if !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
					wholeTiebaFidList = append(wholeTiebaFidList, tmpTcTieba.Fid)
				}
			}
			if len(tiebaList) > 0 {
				err := _function.GormDB.W.Create(tiebaList)
				log.Println("scanTiebaByPid:", err)
			}

			pn++
			// 30 * 200 -> 6000
			// avoid loop
			if response.HasMore == "0" || pn > 20 {
				break
			}
		}
	}

	if _function.GetOption("go_forum_sync_policy") != "add_only" && len(wholeTiebaFidList) != len(localTiebaFidList) {
		delList := []int32{}
		for _, v := range *localTiebaList {
			if !slices.Contains(wholeTiebaFidList, v.Fid) && v.Fid != 0 {
				delList = append(delList, v.ID)
			}
		}
		if len(delList) > 0 {
			_function.GormDB.W.Delete(&model.TcTieba{}, delList)
		}
	}
}

func RefreshTiebaListAction() {

	//activeAfter := 18 //GMT+8 18:00

	day, _ := strconv.ParseInt(_function.GetOption("ver4_ref_day"), 10, 64)
	if day != int64(_function.Now.Local().Day()) {
		lastdo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
		if _function.Now.Unix() > lastdo+90 {
			var accounts = &[]model.TcBaiduid{}
			//TODO account limit per query
			_function.GormDB.R.Model(&model.TcBaiduid{}).Find(accounts)
			for _, account := range *accounts {
				ScanTiebaByPid(account.ID)
				_function.SetOption("ver4_ref_id", strconv.Itoa(int(account.ID)))
				_function.SetOption("ver4_ref_lastdo", strconv.Itoa(int(_function.Now.Unix())))
			}
			_function.SetOption("ver4_ref_id", "0")
			_function.SetOption("ver4_ref_day", strconv.Itoa(_function.Now.Local().Day()))
		}
	}
}
