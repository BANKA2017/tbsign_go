package _plugin

import (
	"log"
	"slices"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_type "github.com/BANKA2017/tbsign_go/types"
)

type RefreshTiebaListPluginType struct {
	PluginInfo
}

var RefreshTiebaListPlugin = _function.VariablePtrWrapper(RefreshTiebaListPluginType{
	PluginInfo{
		Name: "ver4_ref",
	},
})

var RefreshTiebaListPluginName = ""

func ScanTiebaByPid(pid int32) {
	account := _function.GetCookie(pid)

	var localTiebaList = &[]model.TcTieba{}
	_function.GormDB.R.Model(&model.TcTieba{}).Where("pid = ?", account.ID).Find(&localTiebaList)

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
		var tiebaList = []*_type.TcTieba{}
		for _, tiebaInfo := range response.Data.LikeForum.List {
			if tiebaInfo.ForumID == 0 {
				continue
			}
			//log.Println(tiebaInfo)
			//合并或被封禁的贴吧会怎样?

			tmpTcTieba := &_type.TcTieba{
				TcTieba: model.TcTieba{
					Pid: pid,
					Fid: int32(tiebaInfo.ForumID),
					UID: account.UID,
				},
				Tieba:     _function.VariablePtrWrapper(tiebaInfo.ForumName),
				Status:    _function.VariablePtrWrapper(int32(0)),
				LastError: _function.VariablePtrWrapper(""),
			}

			if !slices.Contains(localTiebaFidList, tiebaInfo.ForumID) && !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
				tiebaList = append(tiebaList, tmpTcTieba)
				localTiebaFidList = append(localTiebaFidList, tiebaInfo.ForumID)
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

			if response.ErrorCode != "0" {
				break
			}
			// merge list
			var mergedList []_type.ForumInfo
			if response.ForumList.NonGconforum != nil && len(*response.ForumList.NonGconforum) > 0 {
				mergedList = append(mergedList, *response.ForumList.NonGconforum...)
			}

			if response.ForumList.Gconforum != nil && len(*response.ForumList.Gconforum) > 0 {
				mergedList = append(mergedList, *response.ForumList.Gconforum...)
			}

			if len(mergedList) == 0 {
				break
			}

			var tiebaList = []*_type.TcTieba{}
			for _, tiebaInfo := range mergedList {
				//log.Println(tiebaInfo)
				//合并或被封禁的贴吧会怎样?

				numFID, _ := strconv.ParseInt(tiebaInfo.ID, 10, 64)
				if numFID == 0 {
					continue
				}

				tmpTcTieba := &_type.TcTieba{
					TcTieba: model.TcTieba{
						Pid: pid,
						Fid: int32(numFID),
						UID: account.UID,
					},
					Tieba:     _function.VariablePtrWrapper(tiebaInfo.Name),
					Status:    _function.VariablePtrWrapper(int32(0)),
					LastError: _function.VariablePtrWrapper(""),
				}

				if !slices.Contains(localTiebaFidList, int(numFID)) && !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
					tiebaList = append(tiebaList, tmpTcTieba)
					localTiebaFidList = append(localTiebaFidList, int(numFID))
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

	if (_function.GetOption("go_forum_sync_policy") == "add_delete" || _function.GetOption("go_forum_sync_policy") == "delete_only") && len(wholeTiebaFidList) != len(localTiebaFidList) {
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

func (pluginInfo RefreshTiebaListPluginType) Action() {

	//activeAfter := 18 //GMT+8 18:00

	// day, _ := strconv.ParseInt(_function.GetOption("ver4_ref_day"), 10, 64)

	// if day != int64(_function.Now.Local().Day()) {
	lastdo, _ := strconv.ParseInt(_function.GetOption("ver4_ref_lastdo"), 10, 64)
	refID := _function.GetOption("ver4_ref_id")

	// 4 hours
	if refID != "0" || _function.Now.Unix() > lastdo+60*60*4 {
		var accounts = &[]model.TcBaiduid{}
		// TODO fix hard limit
		_function.GormDB.R.Model(&model.TcBaiduid{}).Where("id > ?", refID).Limit(50).Find(accounts)

		if len(*accounts) == 0 {
			_function.SetOption("ver4_ref_id", "0")
			_function.SetOption("ver4_ref_day", strconv.Itoa(_function.Now.Local().Day()))
		} else {
			for _, account := range *accounts {
				ScanTiebaByPid(account.ID)
				_function.SetOption("ver4_ref_id", strconv.Itoa(int(account.ID)))
				_function.SetOption("ver4_ref_lastdo", strconv.Itoa(int(_function.Now.Unix())))
			}
		}
	}
	// }
}
