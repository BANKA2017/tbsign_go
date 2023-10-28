package _function

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	_type "github.com/BANKA2017/tbsign_go/types"
)

func addSign(form *map[string]string) {
	(*form)["_client_version"] = "12.22.1.0"
	(*form)["_client_type"] = "4"

	var formKeys []string
	for k := range *form {
		formKeys = append(formKeys, k)
	}

	slices.Sort(formKeys)
	payload := ""

	for _, v := range formKeys {
		payload += v + "=" + (*form)[v]
	}
	log.Println("payload", payload)
	_md5 := md5.Sum([]byte(payload + "tiebaclient!!!"))
	(*form)["sign"] = strings.ToUpper(hex.EncodeToString(_md5[:]))
}

func DoSignClient(client *http.Client, cookie _type.Cookie, kw string, fid int32) (_type.ClientSignResponse, error) {
	log.Println(cookie, kw, fid)
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["fid"] = strconv.Itoa(int(fid))
	form["kw"] = kw
	form["tbs"] = cookie.Tbs
	addSign(&form)
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	log.Println(_body.Encode() + "&sign=" + form["sign"])
	signResponse, err := Fetch(client, "http://c.tieba.baidu.com/c/c/forum/sign", "POST", _body.Encode()+"&sign="+form["sign"], headersMap, _type.ClientSignResponse{})

	return *signResponse, err
}

func GetTbs(client *http.Client, bduss string) string {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + bduss,
	}
	tbsResponse, err := Fetch(client, "http://tieba.baidu.com/dc/common/tbs", "GET", "", headersMap, _type.TbsResponse{})
	if err != nil {
		return ""
	}
	return tbsResponse.Tbs
}
