package _function

import (
	"log"

	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/exp/slices"
)

func ScanTiebaByPid(pid int32) {
	account := GetCookie(pid)

	var localTiebaList []*model.TcTieba
	GormDB.R.Model(&model.TcTieba{}).Where("pid = ?", account.ID).Find(&localTiebaList)

	localTiebaFidList := make([]int, len(localTiebaList))

	for _, v := range localTiebaList {
		localTiebaFidList = append(localTiebaFidList, int(v.Fid))
	}

	var pn int64 = 1

	wholeTiebaFidList := []int32{}

	for {
		//log.Println(pid, pn)
		response, err := GetForumList2(account, pn)
		//log.Println(rc, err)
		if err != nil {
			log.Println("scanTiebaByPid:", err)
			break
		}
		if response.ErrorCode != 0 || len(response.LikeForum) == 0 {
			break
		}
		var tiebaList = []*_type.TcTieba{}
		for _, tiebaInfo := range response.LikeForum {
			if tiebaInfo.ForumID == 0 || tiebaInfo.IsForbidden == 1 {
				continue
			}
			//log.Println(tiebaInfo)
			//合并或被封禁的贴吧会怎样?
			/// - 被封的现在有 is_forbidden
			/// - 被合并的暂时没有办法直接判断

			latest := 0
			if tiebaInfo.IsSign == 1 {
				latest = Now.Day()
			}

			tmpTcTieba := &_type.TcTieba{
				TcTieba: model.TcTieba{
					Pid:    pid,
					Fid:    int32(tiebaInfo.ForumID),
					UID:    account.UID,
					Latest: int32(latest),
				},
				Tieba:     VariablePtrWrapper(tiebaInfo.ForumName),
				Status:    VariablePtrWrapper(int32(0)),
				LastError: VariablePtrWrapper(""),
			}

			if !slices.Contains(localTiebaFidList, tiebaInfo.ForumID) {
				tiebaList = append(tiebaList, tmpTcTieba)
				localTiebaFidList = append(localTiebaFidList, tiebaInfo.ForumID)
			}
			if !slices.Contains(wholeTiebaFidList, tmpTcTieba.Fid) {
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
		if !response.LikeForumHasMore || pn > 20 {
			break
		}
	}

	if (GetOption("go_forum_sync_policy") == "add_delete" || GetOption("go_forum_sync_policy") == "delete_only") && len(wholeTiebaFidList) != len(localTiebaFidList) {
		delList := []int32{}
		for _, v := range localTiebaList {
			if !slices.Contains(wholeTiebaFidList, v.Fid) && v.Fid != 0 {
				delList = append(delList, v.ID)
			}
		}
		if len(delList) > 0 {
			GormDB.W.Delete(&model.TcTieba{}, delList)
		}
	}
}
