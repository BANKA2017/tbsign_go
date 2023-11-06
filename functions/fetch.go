package _function

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"crypto/md5"
	"encoding/hex"
	"net/url"
	"strconv"

	"golang.org/x/exp/slices"

	_type "github.com/BANKA2017/tbsign_go/types"
)

var Client = &http.Client{
	Timeout: time.Second * 10,
}

func Fetch[T any](_url string, _method string, _body string, _headers map[string]string, responseTemplate T) (*T, error) {
	var body io.Reader
	if strings.ToUpper(_method) == "POST" {
		body = strings.NewReader(_body)
	} else {
		body = nil
	}
	req, err := http.NewRequest(_method, _url, body)
	if err != nil {
		log.Println("fetch:", err)
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Linux; Android 9; ONEPLUS A3010 Build/PKQ1.181203.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/117.0.0.0 Mobile Safari/537.36 tieba/12.22.1.0")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	for k, v := range _headers {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		log.Println("fetch:", err)
		return nil, err
	}
	defer resp.Body.Close()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("fetch:", err)
		return nil, err
	}
	//log.Println(string(response[:]))

	if err = JsonDecode(response, &responseTemplate); err != nil {
		log.Println("fetch:", err)
		return nil, err
	}
	return &responseTemplate, err
}

func AddSign(form *map[string]string) {
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
	//log.Println("payload", payload)
	_md5 := md5.Sum([]byte(payload + "tiebaclient!!!"))
	(*form)["sign"] = strings.ToUpper(hex.EncodeToString(_md5[:]))
}

func GetTbs(bduss string) string {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + bduss,
	}
	tbsResponse, err := Fetch("http://tieba.baidu.com/dc/common/tbs", "GET", "", headersMap, _type.TbsResponse{})
	if err != nil {
		return ""
	}
	return tbsResponse.Tbs
}

func PostSignClient(cookie _type.TypeCookie, kw string, fid int32) (_type.ClientSignResponse, error) {
	//log.Println(cookie, kw, fid)
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["fid"] = strconv.Itoa(int(fid))
	form["kw"] = kw
	form["tbs"] = cookie.Tbs
	AddSign(&form)
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss,
	}

	//log.Println(_body.Encode() + "&sign=" + form["sign"])
	signResponse, err := Fetch("http://c.tieba.baidu.com/c/c/forum/sign", "POST", _body.Encode()+"&sign="+form["sign"], headersMap, _type.ClientSignResponse{})

	return *signResponse, err
}

func GetForumList(cookie _type.TypeCookie, page int64) (_type.ForumListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := Fetch("https://tieba.baidu.com/mg/o/getForumHome?st=0&pn="+strconv.Itoa(int(page))+"&rn=200", "GET", "", headersMap, _type.ForumListResponse{})

	return *forumListResponse, err
}

func GetForumNameShare(name string) (_type.ForumNameShareResponse, error) {
	queryStr := url.Values{}
	queryStr.Set("ie", "utf-8")
	queryStr.Set("fname", name)

	ForumNameShare, err := Fetch("http://tieba.baidu.com/f/commit/share/fnameShareApi?"+queryStr.Encode(), "GET", "", map[string]string{}, _type.ForumNameShareResponse{})

	return *ForumNameShare, err
}
