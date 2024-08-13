package _function

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"net/url"
	"strconv"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"

	tbpb "github.com/BANKA2017/tbsign_go/proto"
	_type "github.com/BANKA2017/tbsign_go/types"
)

var IgnoreProxy bool

var Client *http.Client

func InitClient(timeout int) *http.Client {
	transport := http.DefaultTransport

	if IgnoreProxy {
		transport.(*http.Transport).Proxy = nil
	}

	return &http.Client{
		Timeout:   time.Second * time.Duration(timeout),
		Transport: transport,
	}
}

var EmptyHeaders = map[string]string{}

const BrowserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

const ClientVersion = "12.58.1.0"
const ClientUserAgent = "tieba/" + ClientVersion

func TBFetch(_url string, _method string, _body []byte, _headers map[string]string) ([]byte, error) {
	if Client == nil {
		Client = InitClient(10)
	}
	return Fetch(_url, _method, _body, _headers, Client)
}

func Fetch(_url string, _method string, _body []byte, _headers map[string]string, client *http.Client) ([]byte, error) {
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
	req.Header.Set("User-Agent", ClientUserAgent)
	if slices.Contains([]string{"POST", "PUT", "PATCH"}, strings.ToUpper(_method)) {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	for k, v := range _headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
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

	return response[:], err
}

type MultipartBodyBinaryFileType struct {
	Fieldname string
	Filename  string
	Binary    []byte
}

func MultipartBodyBuilder(_body map[string]any, files ...MultipartBodyBinaryFileType) ([]byte, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range _body {
		part, _ := writer.CreateFormField(k)
		part.Write([]byte(v.(string)))
	}

	for _, _file := range files {
		part, err := writer.CreateFormFile(_file.Fieldname, _file.Filename)
		if err != nil {
			return nil, "", err
		}
		part.Write(_file.Binary)
	}

	err := writer.Close()
	if err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}

func AddSign(form *map[string]string, client_type string) {
	if client_type == "" {
		client_type = "4"
	}
	(*form)["_client_version"] = ClientVersion
	(*form)["_client_type"] = client_type

	var formKeys []string
	for k := range *form {
		formKeys = append(formKeys, k)
	}

	slices.Sort(formKeys)

	var payload strings.Builder

	for _, v := range formKeys {
		payload.WriteString(v)
		payload.WriteString("=")
		payload.WriteString((*form)[v])
	}
	//log.Println("payload", payload)
	_md5 := md5.Sum([]byte(AppendStrings(payload.String() + "tiebaclient!!!")))
	(*form)["sign"] = strings.ToUpper(hex.EncodeToString(_md5[:]))
}

func GetTbs(bduss string) string {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + bduss,
	}
	tbsResponse, err := TBFetch("http://tieba.baidu.com/dc/common/tbs", "GET", nil, headersMap)

	if err != nil {
		return ""
	}

	var tbsDecode _type.TbsResponse
	if err = JsonDecode(tbsResponse, &tbsDecode); err != nil {
		return ""
	}
	return tbsDecode.Tbs

	/// userInfo, err := GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	/// if err != nil {
	/// 	return ""
	/// } else {
	/// 	return userInfo.Anti.Tbs
	/// }
}

func PostSignClient(cookie _type.TypeCookie, kw string, fid int32) (*_type.ClientSignResponse, error) {
	//log.Println(cookie, kw, fid)
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["fid"] = strconv.Itoa(int(fid))
	form["kw"] = kw
	form["tbs"] = cookie.Tbs
	AddSign(&form, "2")
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
	signResponse, err := TBFetch("https://tiebac.baidu.com/c/c/forum/sign", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var signDecode _type.ClientSignResponse
	err = JsonDecode(signResponse, &signDecode)
	return &signDecode, err
}

func GetWebForumList(cookie _type.TypeCookie, page int64) (*_type.WebForumListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := TBFetch(fmt.Sprintf("https://tieba.baidu.com/mg/o/getForumHome?st=0&pn=%d&rn=200", page), "GET", nil, headersMap)

	if err != nil {
		return nil, err
	}

	var forumListDecode _type.WebForumListResponse
	err = JsonDecode(forumListResponse, &forumListDecode)
	return &forumListDecode, err
}

func GetForumList(cookie _type.TypeCookie, uid string, page int64) (*_type.ForumListResponse, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["friend_uid"] = uid
	form["page_no"] = strconv.Itoa(int(page))
	form["page_size"] = "200"
	form["tbs"] = cookie.Tbs

	AddSign(&form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := TBFetch("https://tiebac.baidu.com/c/f/forum/like", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var forumListDecode _type.ForumListResponse
	err = JsonDecode(forumListResponse, &forumListDecode)
	return &forumListDecode, err
}

func GetOneKeySignList(cookie _type.TypeCookie) (any, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["tbs"] = cookie.Tbs

	AddSign(&form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := TBFetch("https://tieba.baidu.com/c/f/forum/getforumlist", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	//var forumListDecode _type.ForumListResponse
	//err = JsonDecode(forumListResponse, &forumListDecode)
	return string(forumListResponse), err

}

func GetForumNameShare(name string) (*_type.ForumNameShareResponse, error) {
	queryStr := url.Values{}
	queryStr.Set("ie", "utf-8")
	queryStr.Set("fname", name)

	forumNameShare, err := TBFetch("http://tieba.baidu.com/f/commit/share/fnameShareApi?"+queryStr.Encode(), "GET", nil, EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var forumNameShareDecode _type.ForumNameShareResponse
	err = JsonDecode(forumNameShare, &forumNameShareDecode)
	return &forumNameShareDecode, err
}

func GetBaiduUserInfo(cookie _type.TypeCookie) (*_type.BaiduUserInfoResponse, error) {
	var form = make(map[string]string)
	form["bdusstoken"] = cookie.Bduss + "|null" //why '|null' ?
	AddSign(&form, "4")
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
	accountInfo, err := TBFetch("http://c.tieba.baidu.com/c/s/login", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var accountInfoDecode _type.BaiduUserInfoResponse
	err = JsonDecode(accountInfo, &accountInfoDecode)
	return &accountInfoDecode, err
}

func GetUserInfoByTiebaUID(tbuid string) (*tbpb.GetUserByTiebaUidResIdl_DataRes, error) {
	pbBytes, err := proto.Marshal(&tbpb.GetUserByTiebaUidReqIdl_DataReq{
		Common: &tbpb.CommonReq{
			XClientVersion: ClientVersion,
		},
		TiebaUid: tbuid,
	})
	if err != nil {
		return nil, err
	}

	pbBytesLen := make([]byte, 8)
	binary.BigEndian.PutUint64(pbBytesLen, uint64(len(pbBytes)))

	body, contentType, err := MultipartBodyBuilder(map[string]any{}, MultipartBodyBinaryFileType{
		Fieldname: "data",
		Filename:  "file",
		Binary:    bytes.Join([][]byte{[]byte("\n"), RemoveLeadingZeros(pbBytesLen), pbBytes}, []byte{}),
	})

	if err != nil {
		return nil, err
	}

	resp, err := TBFetch("http://tiebac.baidu.com/c/u/user/getUserByTiebaUid?cmd=309702", "POST", body, map[string]string{
		"Content-Type":   contentType,
		"x_bd_data_type": "protobuf",
	})

	if err != nil {
		return nil, err
	}
	//log.Println(resp, string(resp))
	var res tbpb.GetUserByTiebaUidResIdl
	err = proto.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.GetData(), nil
}

func GetUserInfoByUsernameOrPortrait(requestType string, value string) (*_type.TiebaPanelUserInfoResponse, error) {
	query := "ie=utf-8"
	if requestType == "portrait" && strings.HasPrefix(value, "tb.1.") {
		query += "&id=" + value
	} else if requestType == "username" && value != "" {
		query += "&un=" + value
	} else {
		return nil, errors.New("invalid type or portrait/username")
	}
	resp, err := TBFetch("https://tieba.baidu.com/home/get/panel?"+query, "GET", nil, map[string]string{
		"User-Agent": BrowserUserAgent,
	})
	if err != nil {
		return nil, err
	}

	var res _type.TiebaPanelUserInfoResponse
	err = JsonDecode(resp, &res)
	return &res, err
}

// !!! Calling this api will change the IP location !!!
// / DO NOT ASK ME WHY THE RESPONSE IS `ANY`!!!
func PostSync(cookie _type.TypeCookie) (any, error) {
	form := map[string]string{
		"BDUSS": cookie.Bduss,
		"cuid":  "-", //TODO cuid
	}
	AddSign(&form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	response, err := TBFetch("https://tiebac.baidu.com/c/s/sync", "POST", []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp any
	err = JsonDecode(response, &resp)
	return &resp, err
}

func GetLoginQRCode() (*_type.LoginQRCode, error) {
	response, err := TBFetch("https://passport.baidu.com/v2/api/getqrcode?lp=pc", "GET", nil, EmptyHeaders)

	if err != nil {
		return nil, err
	}

	resp := new(_type.LoginQRCode)
	err = JsonDecode(response, &resp)
	return resp, err
}

func GetUnicastResponse(sign string) (*_type.WrapUnicastResponse, error) {
	callbackName := fmt.Sprintf("tangram_guid_%d", Now.UnixMilli())

	res, err := TBFetch(fmt.Sprintf("https://passport.baidu.com/channel/unicast?channel_id=%s&tpl=mn&_sdkFrom=1&callback=%s&apiver=v3", sign, callbackName), "GET", nil, EmptyHeaders)
	if err != nil {
		return nil, err
	}

	resStr := res[len(callbackName)+1 : len(res)-2]

	var parsed _type.UnicastResponse
	err = JsonDecode(resStr, &parsed)
	if err != nil {
		return nil, err
	}
	if parsed.ChannelV != "" {
		var ChannelV _type.UnicastResponseChannelV
		err = JsonDecode([]byte(parsed.ChannelV), &ChannelV)
		if err != nil {
			return nil, err
		}
		return &_type.WrapUnicastResponse{
			ChannelV: &ChannelV,
			UnicastResponse: _type.UnicastResponse{
				ChannelID: parsed.ChannelID,
				Errno:     parsed.Errno,
			},
		}, nil
	} else {
		return &_type.WrapUnicastResponse{
			UnicastResponse: _type.UnicastResponse{
				Errno: parsed.Errno,
			},
		}, nil
	}
}

func GetLoginResponse(tmpBDUSS string) (*_type.LoginResponse, error) {
	res, err := TBFetch(fmt.Sprintf("https://passport.baidu.com/v3/login/main/qrbdusslogin?bduss=%s", tmpBDUSS), "GET", nil, EmptyHeaders)
	if len(res) <= 2 || err != nil {
		return nil, err
	}

	resStr := strings.ReplaceAll(strings.ReplaceAll(string(res), "'", "\""), "\\&", "&")

	var parsed _type.LoginResponse
	err = JsonDecode([]byte(resStr), &parsed)
	if err != nil {
		return nil, err
	}
	return &parsed, err
}

func GetManagerTasks(cookie _type.TypeCookie, fid int64) (*_type.ManagerTasksResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}

	res, err := TBFetch(fmt.Sprintf("https://tieba.baidu.com/mo/q/bawu/taskInfo?fid=%d&tbs=%s", fid, cookie.Tbs), "GET", nil, headersMap)
	if err != nil {
		return nil, err
	}

	var parsed _type.ManagerTasksResponse
	err = JsonDecode(res, &parsed)
	if err != nil {
		return nil, err
	}
	return &parsed, err
}
