package _api

import (
	"net/http"
	"strconv"
	"strings"

	_function "github.com/BANKA2017/tbsign_go/functions"
	"github.com/BANKA2017/tbsign_go/model"
	_type "github.com/BANKA2017/tbsign_go/types"
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
	newTieba := _type.TcTieba{
		TcTieba: model.TcTieba{
			UID:    int32(numUID),
			Pid:    int32(numPid),
			Fid:    int32(fid),
			No:     0,
			Latest: 0,
		},
		Tieba:     _function.VPtr(fname),
		Status:    _function.VPtr(int32(0)),
		LastError: _function.VPtr(""),
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

	_function.GormDB.W.Where("uid = ?", uid).Delete(&model.TcTieba{})
	return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", map[string]string{
		"uid": uid,
	}, "tbsign"))
}

func RefreshTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	arrayMode := IsArrayMode(c)

	var tiebaAccounts []*model.TcBaiduid
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaAccounts)

	// get account list
	for _, v := range tiebaAccounts {
		_function.ScanTiebaByPid(v.ID)
	}

	var tiebaList []*model.TcTieba
	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)

	if arrayMode {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumListObj2Arr(tiebaList), "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaList, "tbsign"))
	}
}

func GetTiebaList(c echo.Context) error {
	uid := c.Get("uid").(string)

	arrayMode := IsArrayMode(c)

	var tiebaList []*model.TcTieba
	var tiebaListBatchQueryList []*model.TcTieba
	// _function.GormDB.R.Where("uid = ?", uid).Order("id ASC").Find(&tiebaList)

	_function.GormDB.R.Where("uid = ?", uid).Order("id ASC").FindInBatches(&tiebaListBatchQueryList, 1000, func(tx *gorm.DB, batch int) error {
		tiebaList = append(tiebaList, tiebaListBatchQueryList...)
		return nil
	})

	if arrayMode {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", ForumListObj2Arr(tiebaList), "tbsign"))
	} else {
		return c.JSON(http.StatusOK, _function.ApiTemplate(200, "OK", tiebaList, "tbsign"))
	}
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
