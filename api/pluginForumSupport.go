package _api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/BANKA2017/tbsign_go/dao/model"
	_function "github.com/BANKA2017/tbsign_go/functions"
	_plugin "github.com/BANKA2017/tbsign_go/plugins"
	"github.com/labstack/echo/v4"
)

// ver4_rank
func PluginForumSupportGetCharactersList(c echo.Context) error {
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", _plugin.ForumSupportList, "tbsign"))
}
func PluginForumSupportGetSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	var rankList []model.TcVer4RankLog
	_function.GormDB.Where("uid = ?", uid).Find(&rankList)

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", rankList, "tbsign"))
}

func PluginForumSupportUpdateSettings(c echo.Context) error {
	uid := c.Get("uid").(string)

	numUID, _ := strconv.ParseInt(uid, 10, 64)

	pid := c.FormValue("pid")
	numPid, err := strconv.ParseInt(pid, 10, 64)

	if err != nil || numPid <= 0 {
		return c.JSON(http.StatusOK, apiTemplate(403, "Invalid pid", echoEmptyObject, "tbsign"))
	}

	var rankList []model.TcVer4RankLog
	_function.GormDB.Where("uid = ? AND pid = ?", uid, pid).Find(&rankList)

	c.Request().ParseForm()

	nid := c.Request().Form["nid[]"]

	var addRankList []model.TcVer4RankLog
	var delRankList []model.TcVer4RankLog
	var delRankIDList []int32
	var failedList []int64

	// add
	for _, v := range nid {
		exist := false
		for _, v1 := range rankList {
			if v1.Nid == v {
				exist = true
			}
		}
		if !exist {
			numNid, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				failedList = append(failedList, numNid)
				continue
			}

			// fid nid
			var forum _plugin.TypeForumSupportList
			for _, _forum := range _plugin.ForumSupportList {
				if _forum.Nid == numNid {
					forum = _forum
					break
				}
			}

			if forum.Nid <= 0 {
				failedList = append(failedList, numNid)
				continue
			}

			numFid, _ := strconv.ParseInt(forum.Fid, 10, 64)

			addRankList = append(addRankList, model.TcVer4RankLog{
				UID:   int32(numUID),
				Pid:   int32(numPid),
				Fid:   int32(numFid),
				Nid:   v,
				Name:  forum.Name,
				Tieba: forum.Tieba,
				Log:   "",
				Date:  0,
			})
		}
	}

	_function.GormDB.Create(&addRankList)

	// del
	for _, v := range rankList {
		exist := false
		for _, v1 := range nid {
			if v.Nid == v1 {
				exist = true
			}
		}
		if !exist {
			delRankList = append(delRankList, v)
			delRankIDList = append(delRankIDList, v.ID)
		}
	}

	_function.GormDB.Delete(&model.TcVer4RankLog{}, delRankIDList)

	var resp = struct {
		Add    []model.TcVer4RankLog `json:"add"`
		Del    []model.TcVer4RankLog `json:"del"`
		Failed []int64               `json:"failed"`
	}{
		Add:    addRankList,
		Del:    delRankList,
		Failed: failedList,
	}

	return c.JSON(http.StatusOK, apiTemplate(200, "OK", resp, "tbsign"))
}

func PluginForumSupportSwitch(c echo.Context) error {
	uid := c.Get("uid").(string)
	status := _function.GetUserOption("ver4_rank_check", uid) != "0"

	newValue := "1"
	if status {
		newValue = "0"
	}
	err := _function.SetUserOption("ver4_rank_check", newValue, uid)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusOK, apiTemplate(500, "无法启用名人堂助攻功能", status, "tbsign"))
	}
	return c.JSON(http.StatusOK, apiTemplate(200, "OK", !status, "tbsign"))
}
