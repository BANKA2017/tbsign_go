package _function

import (
	"log"
	"strconv"

	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/exp/slices"
)

var AccountLoginFailedChannel = make(chan _type.TypeCookie, 100)

func ScanTiebaByPid(pid int32) {
	account := GetCookie(pid)

	var localTiebaList = &[]model.TcTieba{}
	GormDB.R.Model(&model.TcTieba{}).Where("pid = ?", account.ID).Find(&localTiebaList)

	localTiebaFidList := []int{}

	for _, v := range *localTiebaList {
		localTiebaFidList = append(localTiebaFidList, int(v.Fid))
	}

	var pn int64 = 1

	wholeTiebaFidList := []int32{}

	for {
		//log.Println(pid, pn)
		response, err := GetWebForumList(account, pn)
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
				Tieba:     VariablePtrWrapper(tiebaInfo.ForumName),
				Status:    VariablePtrWrapper(int32(0)),
				LastError: VariablePtrWrapper(""),
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
			err := GormDB.W.Create(tiebaList).Error
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

	accountInfo, err := GetBaiduUserInfo(account)
	if accountInfo.ErrorCode == "1" {
		AccountLoginFailedChannel <- account
	}
	if err == nil {
		pn = 1
		for {
			response, err := GetForumList(account, accountInfo.User.ID, pn)
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
			if response.ForumList != nil {
				if len(response.ForumList.NonGconforum) > 0 {
					mergedList = append(mergedList, response.ForumList.NonGconforum...)
				}
				if len(response.ForumList.Gconforum) > 0 {
					mergedList = append(mergedList, response.ForumList.Gconforum...)
				}
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
					Tieba:     VariablePtrWrapper(tiebaInfo.Name),
					Status:    VariablePtrWrapper(int32(0)),
					LastError: VariablePtrWrapper(""),
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
				err := GormDB.W.Create(tiebaList)
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

	if (GetOption("go_forum_sync_policy") == "add_delete" || GetOption("go_forum_sync_policy") == "delete_only") && len(wholeTiebaFidList) != len(localTiebaFidList) {
		delList := []int32{}
		for _, v := range *localTiebaList {
			if !slices.Contains(wholeTiebaFidList, v.Fid) && v.Fid != 0 {
				delList = append(delList, v.ID)
			}
		}
		if len(delList) > 0 {
			GormDB.W.Delete(&model.TcTieba{}, delList)
		}
	}
}
