package _function

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"strings"
	"time"

	"encoding/binary"
	"encoding/json"
	"net/url"
	"strconv"

	"golang.org/x/exp/slices"
	"google.golang.org/protobuf/proto"

	"github.com/BANKA2017/tbsign_go/assets"
	tbpb "github.com/BANKA2017/tbsign_go/proto"
	"github.com/BANKA2017/tbsign_go/share"
	_type "github.com/BANKA2017/tbsign_go/types"
)

func init() {
	var err error
	CACertPool, err = x509.SystemCertPool()
	if err != nil {
		log.Println("Unable to load system CA Cert Pool:", err)

	}

	if CACertPool == nil {
		CACertPool = x509.NewCertPool()
	}

	caFile, err := assets.EmbeddedCACert.ReadFile("ca/cacert.pem")
	if err != nil {
		log.Println("Unable to load embedded CA Cert file")
	} else {
		CACertPool.AppendCertsFromPEM(caFile)
	}
}

var IgnoreProxy bool

var DefaultCient *http.Client
var TBClient *http.Client

var CACertPool *x509.CertPool

func InitClient(timeout time.Duration) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		RootCAs: CACertPool,
	}

	if share.DNSAddress != "" {
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{
						Timeout: timeout,
					}
					// https://pkg.go.dev/net#Dial
					return d.DialContext(ctx, "udp", share.DNSAddress)
				},
			},
		}

		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		}
	}

	if IgnoreProxy {
		transport.Proxy = nil
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}

var EmptyHeaders = make(map[string]string)

const BrowserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36"

const ClientVersion = "12.58.1.0"
const ClientUserAgent = "tieba/" + ClientVersion
const ClientWidgetUserAgent = "TiebaWidgets/" + ClientVersion + " CFNetwork/3826.500.131 Darwin/24.5.0"

func TBFetch(_url string, _method string, _body []byte, _headers map[string]string) ([]byte, error) {
	return Fetch(_url, _method, _body, _headers, TBClient)
}

func Fetch(_url string, _method string, _body []byte, _headers map[string]string, client *http.Client) ([]byte, error) {
	var body io.Reader

	if strings.ToUpper(_method) == http.MethodPost {
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
	if slices.Contains([]string{http.MethodPost, http.MethodPut, http.MethodPatch}, strings.ToUpper(_method)) {
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

	if share.TestMode {
		strResponse := "[binary file]"
		if contentType, ok := resp.Header["Content-Type"]; ok && len(contentType) > 0 {
			mediatype, _, _ := mime.ParseMediaType(contentType[0])
			if slices.Contains([]string{"html", "txt", "json", "xml", "javascript", "x-javascript"}, strings.ReplaceAll(strings.ReplaceAll(mediatype, "application/", ""), "text/", "")) {
				strResponse = string(response)
			}
		}

		log.Printf("\n---TEST MODE FETCH-vvv\nurl: %s\nmethod: %s\nbody: %v\nheaders: %v\n------\nres code: %d\nres headers: %v\nres str: %s\n---TEST MODE FETCH-^^^\n\n", _url, _method, _body, _headers, resp.StatusCode, resp.Header, strResponse)
	}

	return response, err
}

type MultipartBodyBinaryFileType struct {
	Fieldname string
	Filename  string
	Binary    []byte
}

func MultipartBodyBuilder(_body map[string][]byte, files ...MultipartBodyBinaryFileType) ([]byte, string, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for k, v := range _body {
		part, _ := writer.CreateFormField(k)
		part.Write(v)
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

var clientSignSalt = string([]byte{116, 105, 101, 98, 97, 99, 108, 105, 101, 110, 116, 33, 33, 33})
var pcSignSalt = string([]byte{51, 54, 55, 55, 48, 98, 49, 102, 51, 52, 99, 57, 98, 98, 102, 50, 101, 55, 100, 49, 97, 57, 57, 100, 50, 98, 56, 50, 102, 97, 57, 101})
var otherSignSalt = string([]byte{48, 48, 51, 57, 100, 55, 57, 100, 99, 51, 99, 99, 50, 48, 55, 53, 49, 50, 57, 55, 52, 53, 97, 51, 48, 50, 51, 55, 97, 51, 99, 52})

func AddSign(form map[string]string, client_type string) {
	if form == nil {
		form = make(map[string]string)
	}

	if ct := form["_client_type"]; ct == "" {
		if client_type == "" {
			client_type = "4"
		}
		form["_client_type"] = client_type
	}

	// if form["subapp_type"] == "" {
	// 	form["subapp_type"] = "client"
	// }

	var signSalt = clientSignSalt
	switch form["subapp_type"] {
	case "pc":
		signSalt = pcSignSalt
	default:
		if v := form["_client_version"]; v == "" {
			form["_client_version"] = ClientVersion
		}
		// signSalt = clientSignSalt
	}

	var formKeys []string
	for k := range form {
		formKeys = append(formKeys, k)
	}

	slices.Sort(formKeys)

	var payload strings.Builder

	for _, v := range formKeys {
		payload.WriteString(v + "=" + form[v])
	}
	//log.Println("payload", payload)
	form["sign"] = strings.ToUpper(Md5(payload.String() + signSalt))
}

func GetTbs(bduss string) (*_type.TbsResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + bduss,
	}
	tbsResponse, err := TBFetch("http://tieba.baidu.com/dc/common/tbs", http.MethodGet, nil, headersMap)

	if err != nil {
		return nil, err
	}

	tbsDecode := new(_type.TbsResponse)
	if err = JsonDecode(tbsResponse, &tbsDecode); err != nil {
		return nil, err
	}
	return tbsDecode, err

	/// userInfo, err := GetBaiduUserInfo(_type.TypeCookie{Bduss: bduss})
	/// if err != nil {
	/// 	return ""
	/// } else {
	/// 	return userInfo.Anti.Tbs
	/// }
}

func PostCheckinClient(cookie _type.TypeCookie, kw string, fid int32) (*_type.ClientSignResponse, error) {
	//log.Println(cookie, kw, fid)
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["fid"] = strconv.Itoa(int(fid))
	form["kw"] = kw
	form["tbs"] = cookie.Tbs
	form["from_widget"] = "1"
	AddSign(form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	signResponse, err := TBFetch("https://tiebac.baidu.com/c/c/forum/sign", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var signDecode _type.ClientSignResponse
	err = JsonDecode(signResponse, &signDecode)
	return &signDecode, err
}

func PostForumInfoWidget(cookie _type.TypeCookie, fid int32) (any, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["forum_id"] = strconv.Itoa(int(fid))

	AddSign(form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	forumListResponse, err := TBFetch("https://tiebac.baidu.com/c/f/widget/getForumInfo", http.MethodPost, []byte(_body.Encode()), map[string]string{
		"User-Agent": ClientWidgetUserAgent,
	})

	if err != nil {
		return nil, err
	}

	return string(forumListResponse), err
}

func GetWebForumList(cookie _type.TypeCookie, page int64) (*_type.WebForumListResponse, error) {
	headersMap := map[string]string{
		"Cookie": "BDUSS=" + cookie.Bduss + ";STOKEN=" + cookie.Stoken,
	}
	forumListResponse, err := TBFetch("https://tieba.baidu.com/mg/o/getForumHome?st=0&pn="+strconv.Itoa(int(page))+"&rn=200", http.MethodGet, nil, headersMap)

	if err != nil {
		return nil, err
	}

	var forumListDecode _type.WebForumListResponse
	err = JsonDecode(forumListResponse, &forumListDecode)
	return &forumListDecode, err
}

func GetForumList(cookie _type.TypeCookie, uid string, page int64) (*_type.ForumListResponse[*_type.ForumList], error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["friend_uid"] = uid
	form["page_no"] = strconv.Itoa(int(page))
	form["page_size"] = "200"
	form["tbs"] = cookie.Tbs

	AddSign(form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	forumListResponse, err := TBFetch("https://tiebac.baidu.com/c/f/forum/like", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	tmpForumListDecode := new(_type.ForumListResponse[json.RawMessage])
	err = JsonDecode(forumListResponse, tmpForumListDecode)
	if err != nil {
		return nil, err
	}

	forumListDecode := VPtr(_type.ForumListResponse[*_type.ForumList]{
		ErrorCode: tmpForumListDecode.ErrorCode,
		HasMore:   tmpForumListDecode.HasMore,
		ForumList: new(_type.ForumList),
	})

	if !bytes.Equal([]byte{91, 93}, tmpForumListDecode.ForumList) {
		tmpForumList := new(_type.ForumList)
		err = JsonDecode(tmpForumListDecode.ForumList, tmpForumList)
		if err != nil {
			return forumListDecode, err
		}
		forumListDecode.ForumList = tmpForumList
	}

	return forumListDecode, err
}

func GetForumList2(cookie _type.TypeCookie, page int64) (*_type.ForumGuideResponse, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["sort_type"] = "3"
	form["call_from"] = "3"
	form["page_no"] = strconv.Itoa(int(page))
	form["res_num"] = "200"
	form["tbs"] = cookie.Tbs
	//form["top_forum_num"] = "0"

	AddSign(form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	headersMap := map[string]string{
		"Subapp-Type": "hybrid",
	}

	forumListResponse, err := TBFetch("https://tieba.baidu.com/c/f/forum/forumGuide", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), headersMap)

	if err != nil {
		return nil, err
	}

	var forumListDecode _type.ForumGuideResponse
	err = JsonDecode(forumListResponse, &forumListDecode)
	return &forumListDecode, err
}

func PostClientBatchCheckinForumList(cookie _type.TypeCookie) (*_type.BatchCheckinForumListResponse, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["tbs"] = cookie.Tbs

	AddSign(form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	forumListResponse, err := TBFetch("https://tiebac.baidu.com/c/f/forum/getforumlist", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var forumListDecode = new(_type.BatchCheckinForumListResponse)
	err = JsonDecode(forumListResponse, forumListDecode)
	return forumListDecode, err
}

func PostClientBatchCheckin(cookie _type.TypeCookie, fid []string) (*_type.BatchCheckinActionResponse, error) {
	var form = make(map[string]string)
	form["BDUSS"] = cookie.Bduss
	form["stoken"] = cookie.Stoken
	form["tbs"] = cookie.Tbs

	if len(fid) > 50 {
		fid = fid[:50]
	}

	form["forum_ids"] = strings.Join(fid, ",")

	AddSign(form, "2")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	batchCheckinResponse, err := TBFetch("https://tiebac.baidu.com/c/c/forum/msign", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	batchCheckinDecode := new(_type.BatchCheckinActionResponse)
	err = JsonDecode(batchCheckinResponse, batchCheckinDecode)
	return batchCheckinDecode, err
}

func GetForumNameShare(name string) (*_type.ForumNameShareResponse, error) {
	queryStr := url.Values{}
	queryStr.Set("ie", "utf-8")
	queryStr.Set("fname", name)

	forumNameShare, err := TBFetch("http://tieba.baidu.com/f/commit/share/fnameShareApi?"+queryStr.Encode(), http.MethodGet, nil, EmptyHeaders)

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
	AddSign(form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	accountInfo, err := TBFetch("https://tiebac.baidu.com/c/s/login", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

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

	body, contentType, err := MultipartBodyBuilder(map[string][]byte{}, MultipartBodyBinaryFileType{
		Fieldname: "data",
		Filename:  "file",
		Binary:    bytes.Join([][]byte{[]byte("\n"), RemoveLeadingZeros(pbBytesLen), pbBytes}, []byte{}),
	})

	if err != nil {
		return nil, err
	}

	resp, err := TBFetch("http://tiebac.baidu.com/c/u/user/getUserByTiebaUid?cmd=309702", http.MethodPost, body, map[string]string{
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
	resp, err := TBFetch("https://tieba.baidu.com/home/get/panel?"+query, http.MethodGet, nil, map[string]string{
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
		"cuid":  "-", // TODO cuid
	}
	AddSign(form, "4")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	response, err := TBFetch("https://tiebac.baidu.com/c/s/sync", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp any
	err = JsonDecode(response, &resp)
	return &resp, err
}

func GetLoginQRCode() (*_type.LoginQRCode, error) {
	// &qrloginfrom=native <- force to use the client to scan qrcode
	response, err := TBFetch("https://passport.baidu.com/v2/api/getqrcode?lp=pc", http.MethodGet, nil, EmptyHeaders)

	if err != nil {
		return nil, err
	}

	resp := new(_type.LoginQRCode)
	err = JsonDecode(response, &resp)
	return resp, err
}

func GetUnicastResponse(sign string) (*_type.WrapUnicastResponse, error) {
	callbackName := "tangram_guid_" + strconv.Itoa(int(Now.UnixMilli()))

	res, err := TBFetch("https://passport.baidu.com/channel/unicast?channel_id="+sign+"&tpl=mn&_sdkFrom=1&callback="+callbackName+"&apiver=v3", http.MethodGet, nil, EmptyHeaders)
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
	res, err := TBFetch("https://passport.baidu.com/v3/login/main/qrbdusslogin?bduss="+tmpBDUSS, http.MethodGet, nil, EmptyHeaders)
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

	res, err := TBFetch("https://tieba.baidu.com/mo/q/bawu/taskInfo?fid="+strconv.Itoa(int(fid))+"&tbs="+cookie.Tbs, http.MethodGet, nil, headersMap)
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

func GetManagerInfo(fid uint64) (*tbpb.GetBawuInfoResIdl_DataRes, error) {
	pbBytes, err := proto.Marshal(&tbpb.GetBawuInfoReqIdl_DataReq{
		Common: &tbpb.CommonReq{
			XClientVersion: ClientVersion,
		},
		Fid: fid,
	})
	if err != nil {
		return nil, err
	}

	pbBytesLen := make([]byte, 8)
	binary.BigEndian.PutUint64(pbBytesLen, uint64(len(pbBytes)))

	body, contentType, err := MultipartBodyBuilder(map[string][]byte{}, MultipartBodyBinaryFileType{
		Fieldname: "data",
		Filename:  "file",
		Binary:    bytes.Join([][]byte{[]byte("\n"), RemoveLeadingZeros(pbBytesLen), pbBytes}, []byte{}),
	})

	if err != nil {
		return nil, err
	}

	resp, err := TBFetch("http://tiebac.baidu.com/c/f/forum/getBawuInfo?cmd=301007", http.MethodPost, body, map[string]string{
		"Content-Type":   contentType,
		"x_bd_data_type": "protobuf",
	})

	if err != nil {
		return nil, err
	}
	// log.Println(resp, string(resp))
	var res tbpb.GetBawuInfoResIdl
	err = proto.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.GetData(), nil
}

func GetManagerStatus(portrait string, fid int64) (*_type.IsManagerPreCheckResponse, error) {
	managerList, _ := GetManagerInfo(uint64(fid))
	for _, v := range managerList.BawuTeamInfo.BawuTeamList {
		if v.RoleName == "吧主助手" {
			continue
		}
		for _, v2 := range v.RoleInfo {
			if v2.Portrait == portrait {
				return &_type.IsManagerPreCheckResponse{
					IsManager: true,
					Role:      v.RoleName,
				}, nil
			}
		}
	}

	return &_type.IsManagerPreCheckResponse{}, nil
}

func GetNewPCUserCard(portrait string) (*_type.GetNewPCUserCardResponse, error) {
	query := map[string]string{
		"portrait":     portrait,
		"subapp_type":  "pc",
		"_client_type": "20",
	}
	AddSign(query, "20")
	_query := url.Values{}
	for k, v := range query {
		if k != "sign" {
			_query.Set(k, v)
		}
	}

	response, err := TBFetch("https://tieba.baidu.com/c/u/pc/userCard?"+_query.Encode()+"&sign="+query["sign"], http.MethodGet, nil, EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp = new(_type.GetNewPCUserCardResponse)
	err = JsonDecode(response, resp)
	return resp, err
}

func GetNewPCForumCard(fid int) (*_type.GetNewPCForumCardResponse, error) {
	query := map[string]string{
		"forum_id":     strconv.Itoa(fid),
		"subapp_type":  "pc",
		"_client_type": "20",
	}
	AddSign(query, "20")
	_query := url.Values{}
	for k, v := range query {
		if k != "sign" {
			_query.Set(k, v)
		}
	}

	response, err := TBFetch("https://tieba.baidu.com/c/f/pc/forumCard?"+_query.Encode()+"&sign="+query["sign"], http.MethodGet, nil, EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp = new(_type.GetNewPCForumCardResponse)
	err = JsonDecode(response, resp)
	return resp, err
}
