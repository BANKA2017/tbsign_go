package _api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func AddTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.FormValue("pid")
	fname := c.FormValue("fname")

	if fname == "" {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "贴吧名无效", _function.EchoEmptyObject, "tbsign"))
	}
	// get tieba info by fname
	fid := _function.GetFid(fname)
	if fid == 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(404, "\""+fname+"吧\" 不存在", _function.EchoEmptyObject, "tbsign"))
	}

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	// pre-check
	var tiebaItems []*model.TcTieba
	_function.GormDB.R.Where("uid = ? AND pid = ? AND fid = ?", uid, pid, fid).Limit(1).Find(&tiebaItems)

	if len(tiebaItems) > 0 {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "贴吧已存在", tiebaItems[0], "tbsign"))
	}

	// TOO STUPID!
	newTieba := model.TcTieba{
		UID:       int32(numUID),
		Pid:       int32(numPid),
		Fid:       int32(fid),
		No:        0,
		Latest:    0,
		Tieba:     fname,
		Status:    0,
		LastError: "",
	}

	_function.GormDB.W.Create(&newTieba)

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", newTieba, "tbsign"))
}

type ModifyForumIDList struct {
	InvalidFid []string `json:"invalid_fid"`
	ValidFid   []int64  `json:"valid_fid"`
}

const MaxFidSeqCount = 100

func VerifyFidList(fid string) ModifyForumIDList {
	list := ModifyForumIDList{
		[]string{},
		[]int64{},
	}

	fidArray := strings.SplitSeq(strings.TrimSpace(fid), ",")

	seq := 0
	for f := range fidArray {
		if seq > MaxFidSeqCount {
			break
		}
		numFid, err := strconv.ParseInt(f, 10, 64)
		if err != nil {
			list.InvalidFid = append(list.InvalidFid, f)
		} else {
			list.ValidFid = append(list.ValidFid, numFid)
		}
		seq++
	}

	return list
}

func RemoveTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}

	response := VerifyFidList(fid)

	if len(response.ValidFid) > 0 {
		if len(response.ValidFid) == 1 {
			_function.GormDB.W.Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, response.ValidFid[0]).Delete(&model.TcTieba{})
		} else {
			_function.GormDB.W.Where("uid = ? AND pid = ? AND fid IN (?)", numUID, numPid, response.ValidFid).Delete(&model.TcTieba{})
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", response, "tbsign"))
}

func ResetTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}
	response := VerifyFidList(fid)

	if len(response.ValidFid) > 0 {
		if len(response.ValidFid) == 1 {
			_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, response.ValidFid[0]).Update("latest", 0)
		} else {
			_function.GormDB.W.Model(&model.TcTieba{}).Where("uid = ? AND pid = ? AND fid IN (?)", numUID, numPid, response.ValidFid).Update("latest", 0)
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", response, "tbsign"))
}

func IgnoreTieba(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")
	fid := c.Param("fid")

	method := c.Request().Method

	numUID, _ := strconv.ParseInt(uid, 10, 64)
	numPid, err := strconv.ParseInt(pid, 10, 64)
	if err != nil {
		return c.JSON(http.StatusOK, _function.ApiTemplate(403, "无效 pid", _function.EchoEmptyObject, "tbsign"))
	}
	response := VerifyFidList(fid)

	if len(response.ValidFid) > 0 {
		if len(response.ValidFid) == 1 {
			_function.GormDB.W.Model(&model.TcTieba{}).Select("no", "latest").Where("uid = ? AND pid = ? AND fid = ?", numUID, numPid, response.ValidFid[0]).Updates(&model.TcTieba{
				No:     _function.When(method == http.MethodDelete, 0, 1),
				Latest: -1,
			})
		} else {
			_function.GormDB.W.Model(&model.TcTieba{}).Select("no", "latest").Where("uid = ? AND pid = ? AND fid IN (?)", numUID, numPid, response.ValidFid).Updates(&model.TcTieba{
				No:     _function.When(method == http.MethodDelete, 0, 1),
				Latest: -1,
			})
		}
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]any{
		"uid": numUID,
		"pid": numPid,
		"fid": response,
		"no":  method != http.MethodDelete,
	}, "tbsign"))
}

func CleanTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	var numPid int64
	if pid != "" {
		numPid, _ = strconv.ParseInt(pid, 10, 64)
	}

	tx := _function.GormDB.W.Where("uid = ?", uid)

	if numPid > 0 {
		tx = tx.Where("pid = ?", numPid)
	}

	tx.Delete(&model.TcTieba{})
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]string{
		"uid": uid,
	}, "tbsign"))
}

func RefreshTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	var numPid int64
	if pid != "" {
		numPid, _ = strconv.ParseInt(pid, 10, 64)
	}

	arrayMode := IsArrayMode(c)

	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	// get account list
	synced := false
	for _, v := range tiebaAccounts {
		if (numPid > 0 && v.ID == int32(numPid)) || numPid == 0 {
			_function.ScanTiebaByPid(v.ID)
			synced = true
		}
	}

	var tiebaList []*model.TcTieba

	if numPid == 0 {
		_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)
	} else if synced {
		_function.GormDB.R.Where("uid = ? AND pid = ?", uid, numPid).Order("id ASC").Find(&tiebaList)
	}

	if arrayMode {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumListObj2Arr(tiebaList), "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaList, "tbsign"))
}

func GetTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	var numPid int64
	if pid != "" {
		numPid, _ = strconv.ParseInt(pid, 10, 64)
	}

	arrayMode := IsArrayMode(c)

	var tiebaList []*model.TcTieba
	var tiebaListBatchQueryList []*model.TcTieba
	// _function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)

	tx := _function.GormDB.R.Where("uid = ?", uid).Order("id ASC")

	if numPid > 0 {
		tx = tx.Where("pid = ?", numPid)
	}

	tx.FindInBatches(&tiebaListBatchQueryList, 1000, func(tx *gorm.DB, batch int) error {
		tiebaList = append(tiebaList, tiebaListBatchQueryList...)
		return nil
	})

	if arrayMode {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumListObj2Arr(tiebaList), "tbsign"))
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaList, "tbsign"))
}

func ForumListObj2Arr(tiebaList []*model.TcTieba) [][9]any {
	listArray := make([][9]any, len(tiebaList))
	for i, forumInfo := range tiebaList {
		listArray[i] = [9]any{
			forumInfo.ID,
			forumInfo.UID,
			forumInfo.Pid,
			forumInfo.Fid,
			forumInfo.Tieba,
			forumInfo.No,
			forumInfo.Status,
			forumInfo.Latest,
			forumInfo.LastError,
		}
	}
	return listArray
}

func GetForumStatus(c echo.Context) error {
	uid := c.Get("uid").(string)

	pid := c.Param("pid")

	var numPid int64
	if pid != "" {
		numPid, _ = strconv.ParseInt(pid, 10, 64)
	}

	var status struct {
		Uid        int32 `json:"uid"`
		Pid        int32 `json:"pid,omitempty"`
		ForumCount int   `json:"forum_count"`
		Success    int   `json:"success"`
		Failed     int   `json:"failed"`
		Waiting    int   `json:"waiting"`
		IsIgnore   int   `json:"ignore"`
	}

	today := strconv.Itoa(time.Now().Day())
	tx := _function.GormDB.R.Model(&model.TcTieba{}).
		Select("uid, pid, COUNT(*) AS forum_count, SUM(CASE WHEN (no = 0) AND status = 0 AND latest = ? THEN 1 ELSE 0 END) AS success, SUM(CASE WHEN (no = 0) AND status <> 0 AND latest = ? THEN 1 ELSE 0 END) AS failed, SUM(CASE WHEN (no = 0) AND latest <> ? THEN 1 ELSE 0 END) AS waiting, SUM(CASE WHEN no <> 0 THEN 1 ELSE 0 END) AS is_ignore", today, today, today).
		Where("uid = ?", uid)

	if numPid > 0 {
		tx = tx.Where("pid = ?", numPid)
	}

	tx.Scan(&status)

	if numPid <= 0 {
		status.Pid = 0
	}

	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", status, "tbsign"))
}
