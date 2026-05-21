package _function

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/BANKA2017/tbsign_go/model"
	"golang.org/x/sync/singleflight"
)

var SyncForumListSf singleflight.Group

func ScanTiebaByPid(pid int32) {
	SyncForumListSf.Do(strconv.Itoa(int(pid)), func() (any, error) {
		account := GetCookie(pid)

		// if !account.IsLogin {
		// 	fmt.Println("scanTiebaByPid:", errors.New("account "+strconv.Itoa(int(pid))+" login status failed"))
		// 	return
		// }

		var localTiebaList []*model.TcTieba
		GormDB.R.Model(&model.TcTieba{}).Where("pid = ?", account.ID).Find(&localTiebaList)

		localTiebaFidList := make(map[int32]string, len(localTiebaList))
		pendingDelete := make(map[int32]int32, len(localTiebaList))

		for _, v := range localTiebaList {
			localTiebaFidList[v.Fid] = v.Tieba
			pendingDelete[v.Fid] = v.ID
		}

		var pn int64 = 1

		for {
			//fmt.Println(pid, pn)
			response, err := GetForumList2(account, pn)
			//fmt.Println(rc, err)
			if err != nil {
				slog.Error("scan forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
				break
			}
			if response.ErrorCode != 0 || len(response.LikeForum) == 0 {
				break
			}
			var tiebaList []*model.TcTieba
			for _, tiebaInfo := range response.LikeForum {
				if tiebaInfo.ForumID == 0 || tiebaInfo.IsForbidden == 1 {
					continue
				}
				//fmt.Println(tiebaInfo)
				//合并或被封禁的贴吧会怎样?
				/// - 被封的现在有 is_forbidden
				/// - 被合并的贴吧暂时没有办法直接判断

				/// latest 的数字等于日期，正常从 1 开始
				/// 0 从未签到过/已重置，-1 已忽略，-2 一键签到中

				latest := 0
				if tiebaInfo.IsSign == 1 {
					latest = time.Now().Day()
				}

				tmpTcTieba := &model.TcTieba{
					Pid:       pid,
					Fid:       int32(tiebaInfo.ForumID),
					UID:       account.UID,
					Latest:    int32(latest),
					Tieba:     tiebaInfo.ForumName,
					Status:    0,
					LastError: "",
				}

				if _, ok := localTiebaFidList[int32(tiebaInfo.ForumID)]; !ok {
					tiebaList = append(tiebaList, tmpTcTieba)
					localTiebaFidList[int32(tiebaInfo.ForumID)] = tiebaInfo.ForumName
				}

				delete(pendingDelete, int32(tiebaInfo.ForumID))
			}
			if len(tiebaList) > 0 {
				err := GormDB.W.Create(tiebaList).Error
				if err != nil {
					slog.Error("update forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
				}
			}

			pn++
			// 20 * 200 -> 4000
			// avoid loop
			if !response.LikeForumHasMore || pn > 20 {
				break
			}
		}

		if GetOption("go_forum_sync_policy") == "add_delete" && len(pendingDelete) > 0 {
			delList := make([]int32, 0, len(pendingDelete))
			for _, id := range pendingDelete {
				delList = append(delList, id)
			}

			if len(delList) > 0 {
				GormDB.W.Delete(&model.TcTieba{}, delList)
			}
		}

		return nil, nil
	})
}
