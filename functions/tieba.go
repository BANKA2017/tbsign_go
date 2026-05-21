package _function

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/BANKA2017/tbsign_go/model"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
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
		localForumCounter := make(map[string][]int32, len(localTiebaList))
		pendingDelete := make(map[int32]int32, len(localTiebaList))

		for _, v := range localTiebaList {
			localTiebaFidList[v.Fid] = v.Tieba
			pendingDelete[v.Fid] = v.ID
			localForumCounter[v.Tieba] = append(localForumCounter[v.Tieba], v.Fid)
		}

		// duplicate forum name
		var shouldUpdateFid = make(map[int32]struct{})
		for _, v := range localForumCounter {
			if len(v) > 1 {
				for _, id := range v {
					shouldUpdateFid[id] = struct{}{}
				}
			}
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
			var updateList = make(map[int]string)
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

				localFname, ok := localTiebaFidList[int32(tiebaInfo.ForumID)]
				if !ok {
					tiebaList = append(tiebaList, &model.TcTieba{
						Pid:       pid,
						Fid:       int32(tiebaInfo.ForumID),
						UID:       account.UID,
						Latest:    int32(latest),
						Tieba:     tiebaInfo.ForumName,
						Status:    0,
						LastError: "",
					})
					localTiebaFidList[int32(tiebaInfo.ForumID)] = tiebaInfo.ForumName
				} else if localFname != tiebaInfo.ForumName {
					updateList[tiebaInfo.ForumID] = tiebaInfo.ForumName
					slog.Warn("forum name changed1", "fid", tiebaInfo.ForumID, "old", localFname, "new", tiebaInfo.ForumName)
				} else if _, ok := shouldUpdateFid[int32(tiebaInfo.ForumID)]; ok {
					updateList[tiebaInfo.ForumID] = tiebaInfo.ForumName
					delete(shouldUpdateFid, int32(tiebaInfo.ForumID))
					slog.Warn("forum name changed2", "fid", tiebaInfo.ForumID, "old", localFname, "new", tiebaInfo.ForumName)
				}

				delete(pendingDelete, int32(tiebaInfo.ForumID))
			}

			GormDB.W.Transaction(func(tx *gorm.DB) error {
				if len(tiebaList) > 0 {
					err := tx.Create(tiebaList).Error
					if err != nil {
						slog.Error("add forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
						return err
					}
				}

				if len(updateList) > 0 {
					for k, v := range updateList {
						err := tx.Model(&model.TcTieba{}).Where("fid = ?", k).Update("tieba", v).Error
						if err != nil {
							slog.Error("update forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
						}
					}
				}

				return nil
			})

			pn++
			// 20 * 200 -> 4000
			// avoid loop
			if !response.LikeForumHasMore || pn > 20 {
				break
			}
		}

		// delete
		if GetOption("go_forum_sync_policy") == "add_delete" && len(pendingDelete) > 0 {
			delList := make([]int32, 0, len(pendingDelete))
			for _, id := range pendingDelete {
				delList = append(delList, id)
			}

			if len(delList) > 0 {
				if err := GormDB.W.Delete(&model.TcTieba{}, delList).Error; err != nil {
					slog.Error("delete forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
				}
			}
		}

		// no record in duplicated forum name
		// if len(shouldUpdateFid) > 0 {
		// 	for k := range shouldUpdateFid {
		// 		forum := GetFname(int64(k), true)
		// 		if err := GormDB.W.Model(&model.TcTieba{}).Where("fid = ?", k).Update("tieba", forum).Error; err != nil {
		// 			slog.Error("update forum list(scanTiebaByPid) failed", "pid", pid, "error", err)
		// 		}
		// 	}
		// }

		return nil, nil
	})
}
