package _function

import (
	"bytes"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"net/url"
	"strconv"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"

	pb "github.com/BANKA2017/tbsign_go/proto"
	_type "github.com/BANKA2017/tbsign_go/types"
)

var Client = &http.Client{
	Timeout: time.Second * 10,
}

func Fetch(_url string, _method string, _body []byte, _headers map[string]string) ([]byte, error) {
	var body io.Reader

	if strings.ToUpper(_method) == "POST" {
		body = bytes.NewReader(_body)
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
	log.Println(req.Header.Clone())
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
	//log.Println(string())

	return response[:], err
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
	tbsResponse, err := Fetch("http://tieba.baidu.com/dc/common/tbs", "GET", nil, headersMap)

	if err != nil {
		return ""
	}

	var tbsDecode _type.TbsResponse
	if err = JsonDecode(tbsResponse, &tbsDecode); err != nil {
		return ""
	}
	return tbsDecode.Tbs
}

func PostSignClient(cookie _type.TypeCookie, kw string, fid int32) (*_type.ClientSignResponse, error) {
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
	signResponse, err := Fetch("http://c.tieba.baidu.com/c/c/forum/sign", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var signDecode _type.ClientSignResponse
	err = JsonDecode(signResponse, &signDecode)
	return &signDecode, err
}

func GetForumList(cookie _type.TypeCookie, page int64) (*_type.ForumListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := Fetch("https://tieba.baidu.com/mg/o/getForumHome?st=0&pn="+strconv.Itoa(int(page))+"&rn=200", "GET", nil, headersMap)

	if err != nil {
		return nil, err
	}

	var forumListDecode _type.ForumListResponse
	err = JsonDecode(forumListResponse, &forumListDecode)
	return &forumListDecode, err
}

func GetForumNameShare(name string) (*_type.ForumNameShareResponse, error) {
	queryStr := url.Values{}
	queryStr.Set("ie", "utf-8")
	queryStr.Set("fname", name)

	forumNameShare, err := Fetch("http://tieba.baidu.com/f/commit/share/fnameShareApi?"+queryStr.Encode(), "GET", nil, map[string]string{})

	if err != nil {
		return nil, err
	}

	var forumNameShareDecode _type.ForumNameShareResponse
	err = JsonDecode(forumNameShare, &forumNameShareDecode)
	return &forumNameShareDecode, err
}

func GetBaiduUserInfo(cookie _type.TypeCookie) (*_type.BaiduUserInfoResponse, error) {
	var form = make(map[string]string)
	form["bdusstoken"] = cookie.Bduss
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
	accountInfo, err := Fetch("http://c.tieba.baidu.com/c/s/login", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var accountInfoDecode _type.BaiduUserInfoResponse
	err = JsonDecode(accountInfo, &accountInfoDecode)
	return &accountInfoDecode, err
}

func GetUserInfoByTiebaUID(tbuid string) (*pb.GetUserByTiebaUidResIdl_DataRes, error) {
	pbBytes, err := proto.Marshal(&pb.GetUserByTiebaUidReqIdl_DataReq{
		Common: &pb.CommonReq{
			XClientVersion: "12.57.4.2",
		},
		TiebaUid: tbuid,
	})
	if err != nil {
		return nil, err
	}

	pbBytesLen := make([]byte, 8)
	binary.BigEndian.PutUint64(pbBytesLen, uint64(len(pbBytes)))

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	multipartHeader := textproto.MIMEHeader{}
	multipartHeader.Set("Content-Disposition", "form-data; name=\"data\"; filename=\"file\"")
	part, err := writer.CreatePart(multipartHeader)
	if err != nil {
		return nil, err
	}
	part.Write([]byte("\n"))
	part.Write(RemoveLeadingZeros(pbBytesLen))
	part.Write(pbBytes)
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	resp, err := Fetch("http://tiebac.baidu.com/c/u/user/getUserByTiebaUid?cmd=309702", "POST", body.Bytes(), map[string]string{
		"Content-Type":   writer.FormDataContentType(),
		"x_bd_data_type": "protobuf",
	})

	if err != nil {
		return nil, err
	}
	//log.Println(resp, string(resp))
	var res pb.GetUserByTiebaUidResIdl
	err = proto.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.GetData(), nil
}
