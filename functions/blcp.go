package _function

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"time"

	tbpb "github.com/BANKA2017/tbsign_go/proto"
	"github.com/BANKA2017/tbsign_go/share"
	"github.com/jellydator/ttlcache/v3"
	"google.golang.org/protobuf/proto"
)

const (
	ChatAppId      = 10773430
	ChatSdkVersion = 11250036
	ChatVersion    = "12.68.1.0" // ClientVersion->12.58.1.0
	LcmSdkVersion  = "3460016"
	BlcpUserAgent  = "okhttp/3.11.0"
	BlcpPlatform   = "android"
)

var BlcpMagic = []byte{'l', 'c', 'p', 1}

type BlcpState int

const (
	BLCPStateDisconnected BlcpState = iota - 1
	BLCPStateConnected
	BLCPStateLoggedIn
)

type BLCPCore struct {
	Client    *Client
	Conn      net.Conn
	BLCPToken string
	TTL       time.Duration

	MessageChan     chan *BLCPWaitListContent
	PendingRequests *KV[int64, chan *BLCPWaitListContent]
	NotifyCallback  func(*BLCPWaitListContent) error // <- NotifyCallback should use goroutine

	Ctx         context.Context
	CloseSignal chan struct{}

	LoginID int64

	State BlcpState
}

type BLCPNotify struct {
	Appid     int `json:"appid"`
	EventList []struct {
		Event       string `json:"event"`
		TimestampMs int    `json:"timestamp_ms"`
	} `json:"event_list"`
	McastID   int `json:"mcast_id"`
	McastType int `json:"mcast_type"`
	Messages  []struct {
		Category   int    `json:"category"`
		Contacter  int    `json:"contacter"`
		Content    string `json:"content"`
		CreateTime int    `json:"create_time"`
		FromUser   int    `json:"from_user"`
		IsRead     int    `json:"is_read"`
		McastID    int    `json:"mcast_id"`
		MsgKey     string `json:"msg_key"`
		Msgid      int    `json:"msgid"`
		OriginID   int    `json:"origin_id,omitempty"`
		Priority   int    `json:"priority"`
		Type       int    `json:"type"`
	} `json:"messages"`
	Method   int `json:"method"`
	NotifyID int `json:"notify_id"`
}

type BLCPWaitListContent struct {
	IsNotify      bool
	IsHeartBeat   bool
	BLCPData      *BLCPData
	RPC           *tbpb.RpcMeta
	LCMType       string // protobuf/json/notify(gzipped json)
	GzipLCM       bool   // <- for debug
	ProtoLCM      *tbpb.RpcData
	JsonLCM       map[string]any
	NotifyLCM     *BLCPNotify
	CorrelationId int
}

func NewBLCP(ctx context.Context, client *Client) *BLCPCore {
	ctx, cancel := context.WithCancel(ctx)
	ttl := time.Second * 5
	blcp := &BLCPCore{
		Ctx:         ctx,
		CloseSignal: make(chan struct{}, 1),

		Client: client,
		State:  BLCPStateDisconnected,

		TTL:         ttl,
		MessageChan: make(chan *BLCPWaitListContent, 100),
		PendingRequests: NewKV(
			ttlcache.WithCapacity[int64, chan *BLCPWaitListContent](500),
			ttlcache.WithTTL[int64, chan *BLCPWaitListContent](ttl+time.Second*5),
		),
	}

	// do after ctx done
	go func() {
		defer cancel()
		<-blcp.CloseSignal
		blcp.State = BLCPStateDisconnected
		time.Sleep(blcp.TTL + time.Second)
		if blcp.Conn != nil {
			blcp.Conn.Close()
		}
	}()
	return blcp
}

func (b *BLCPCore) Connect() error {
	conn, err := (&tls.Dialer{
		NetDialer: &net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		},
		Config: &tls.Config{
			ServerName:         "common.lcs.baidu.com",
			RootCAs:            CACertPool,
			ClientSessionCache: tls.NewLRUClientSessionCache(5),
		},
	}).Dial("tcp", "common.lcs.baidu.com:443")

	if err != nil {
		return err
	}

	b.Conn = conn

	go b.messageDispatcher()
	go b.readLoop()
	go func() {
		ticker := time.NewTicker(time.Second * 15)
		for {
			select {
			case <-b.Ctx.Done():
				return
			case <-ticker.C:
				if b.State == BLCPStateLoggedIn {
					b.heartbeat()
				}
			}
		}
	}()

	b.State = BLCPStateConnected

	return nil
}

func (b *BLCPCore) messageDispatcher() {
	for {
		select {
		case <-b.Ctx.Done():
			return
		case message, ok := <-b.MessageChan:
			if share.TestMode {
				fmt.Printf("<--ok: %t\n", ok)
			}
			if !ok {
				continue
			}

			if message.IsHeartBeat {
				if share.TestMode {
					fmt.Println("<--heartbeat")
				}
				continue
			}

			if message.IsNotify {
				if b.NotifyCallback != nil {
					if err := b.NotifyCallback(message); err != nil {
						slog.Error("blcp notify callback", "error", err)
					}
				}
			} else {
				// find place
				if val, ok := b.PendingRequests.Load(message.BLCPData.correlationId); ok && val != nil {
					val <- message
				}
			}
		}
	}
}

func (b *BLCPCore) SyncSendBLCPData(data *BLCPData) (*BLCPWaitListContent, error) {
	if share.TestMode {
		strBplcData, _ := JsonEncode(data)
		fmt.Println("send->", "data", string(strBplcData))
	}

	return b.SyncSend(data.correlationId, data.toBytes())
}

func (b *BLCPCore) SyncSend(key int64, payload []byte) (*BLCPWaitListContent, error) {
	if share.TestMode {
		fmt.Println("send->", key)
	}

	if b.Conn == nil {
		return nil, errors.New("no conn (snyc send)")
	}

	ctx, cancel := context.WithTimeout(b.Ctx, b.TTL)
	defer cancel()

	var msgSignal = make(chan *BLCPWaitListContent, 1)

	b.PendingRequests.Store(key, msgSignal, 0)

	_, err := b.Conn.Write(payload)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data := <-msgSignal:
		if data == nil || data.RPC == nil || (data.ProtoLCM == nil && data.JsonLCM == nil && data.NotifyLCM == nil) {
			return nil, errors.New("BLCP Handshake error.")
		} else if data.RPC.Response.ErrorText != "success" {
			return nil, errors.New(data.RPC.Response.ErrorText)
		}
		return data, nil
	}
}

func (b *BLCPCore) Login() error {
	lcmTokenResponse, err := b.GenerateLcmToken()
	if err != nil {
		return err
	}

	b.BLCPToken = lcmTokenResponse.Token

	cuidGalaxy2, err := b.Client.CuidGalaxy2()
	if err != nil {
		return err
	}

	// login part1
	loginBLCPRequest := NewBLCPData(1, 1)

	if loginBLCPRequest.RpcBody, err = b.buildRpcBody(1, 1, loginBLCPRequest.correlationId, 0, 1); err != nil {
		return err
	}

	if loginBLCPRequest.LcmBody, err = proto.Marshal(&tbpb.RpcData{
		LcmRequest: &tbpb.LcmRequest{
			LogId: loginBLCPRequest.correlationId,
			Token: b.BLCPToken,
			Common: &tbpb.Common{
				Cuid:       cuidGalaxy2,
				Device:     BlcpPlatform,
				AppId:      strconv.FormatInt(ChatAppId, 10),
				AppVersion: ChatVersion,
				SdkVersion: LcmSdkVersion,
				Network:    "wifi",
			},
			Timestamp: time.Now().UnixMilli(),
			StartType: -1,
			ConnType:  1,
		},
	}); err != nil {
		return err
	}

	resp, err := b.SyncSendBLCPData(loginBLCPRequest)
	if err != nil {
		return err
	} else if resp.ProtoLCM.LcmResponse.ErrorMsg != "success" {
		return errors.New(resp.ProtoLCM.LcmResponse.ErrorMsg)
	}

	// login part2
	// com.baidu.searchbox.cloudcontrolblcp.CloudControlBlCPManager
	// com.baidu.common.param.CommonUrlParamManager
	sid := b.Client.SampleID
	ua := "900_1600_android_" + ChatVersion + "_240"
	uid, _ := EnuidEncode(cuidGalaxy2) // Enuid -> cuid_galaxy2->libbase64encoder_v2_0.so
	login_from := "1008550l"           // baidu_imsdk_common_data.xml
	cfrom := login_from
	c3_aid, _ := b.Client.C3Aid()
	reqdict := map[string]any{
		"params": map[string]any{
			"appname": "tieba",
			"sid":     sid,
			"ua":      ua,
			"uid":     uid,
			"cfrom":   cfrom,
			"from":    login_from,
			"network": "1_-1",
			"p_sv":    "32",
			"mps":     "",
			"mpv":     "1",
			"c3_aid":  c3_aid,
			"type_id": "0",
		},
		"filter": map[string]any{"aps": map[string]string{"cpu_abi": "armeabi-v7a"}, "command": map[string]string{"step": "0"}},
	}

	loginLcmBody, err := JsonEncode(reqdict)
	if err != nil {
		return err
	}

	loginBLCPRequest = NewBLCPData(4, 1)
	loginBLCPRequest.RpcBody, err = b.buildRpcBody(4, 1, loginBLCPRequest.correlationId, 0, 1)
	if err != nil {
		return err
	}
	loginBLCPRequest.LcmBody = loginLcmBody

	resp, err = b.SyncSendBLCPData(loginBLCPRequest)
	if err != nil {
		return err
	} else if resp.LCMType != "json" {
		return errors.New("wrong lcm response type")
	} else if resp.JsonLCM["errno"] != "0" {
		return errors.New("BLCP Handshake error.1")
	}

	// login part3
	loginBLCPRequest = NewBLCPData(2, 50)
	loginBLCPRequest.correlationId = 2000003149381050
	loginBLCPRequest.RpcBody, err = b.buildRpcBody(2, 50, loginBLCPRequest.correlationId, 0, 1)
	if err != nil {
		return err
	}

	clientIdentifier := map[string]string{"zid": "", "version_code": ""} // no zid is allowed now
	cookie := ""                                                         // no cookie is allowed now
	token := b.Client.Account.Bduss

	strClientIdentifier, _ := JsonEncode(clientIdentifier)
	strRPCConfig, _ := JsonEncode(map[string]int{"rpc_retry_time": 0})

	reqdict = map[string]any{
		"method":            50,
		"appid":             ChatAppId,
		"device_id":         "android_" + cuidGalaxy2,
		"account_type":      1,
		"token":             token,
		"version":           4,
		"sdk_version":       ChatSdkVersion,
		"app_version":       ChatVersion,
		"app_open_type":     0,
		"client_identifier": string(strClientIdentifier),
		"tail":              0,
		"timeout":           10,
		"cookie":            cookie,
		"device_info": map[string]any{
			"app_version": ChatVersion,
			"os_version":  "32",
			"platform":    BlcpPlatform,
			"appid":       strconv.FormatInt(ChatAppId, 10),
			"from":        login_from,
			"cfrom":       cfrom,
		},
		"rpc":          string(strRPCConfig),
		"user_type":    0,
		"client_logid": time.Now().Nanosecond(),
	}

	loginLcmBody, err = JsonEncode(reqdict)
	if err != nil {
		return err
	}

	loginBLCPRequest.LcmBody = loginLcmBody
	resp, err = b.SyncSendBLCPData(loginBLCPRequest)
	if err != nil {
		return err
	} else if resp.LCMType != "json" {
		return errors.New("wrong lcm response type")
	} else if int(resp.JsonLCM["err_code"].(float64)) != 0 {
		return errors.New("BLCP Handshake error.3")
	}

	// {"err_code":0,"logid":<int>,"msg":"success","uk":<int>,"bd_uid":<int>,"role":3,"authority":<int>,"bddns_enable":1,"pa_uid":0,"b_user_type":0,"login_id":<int64>,"trigger_id":[<int>],"trace_location":"","cookie":"","disable_restapi":0,"conn_type":0,"client_upload_log_switch":1,"log_switch":[{"id":<int>,"switch":0,"ext_conf":"{}"}]}

	// TODO fix nil ptr
	b.Client.Account.TriggerID = int(resp.JsonLCM["trigger_id"].([]any)[0].(float64))
	b.Client.Account.UK = int(resp.JsonLCM["uk"].(float64))

	b.Client.Account.BaiduUID = int(resp.JsonLCM["bd_uid"].(float64))
	b.Client.Account.BDUK, _ = GetBDUKfromUserId(strconv.Itoa(b.Client.Account.BaiduUID))

	b.LoginID = int64(resp.JsonLCM["login_id"].(float64))

	b.State = BLCPStateLoggedIn

	return nil
}

func (b *BLCPCore) heartbeat() {
	request := NewBLCPData(1, 3)
	request.RpcBody, _ = b.buildRpcBody(1, 3, request.correlationId, 0, 1)
	request.LcmBody, _ = proto.Marshal(&tbpb.RpcData{
		LcmRequest: &tbpb.LcmRequest{
			LogId:     request.correlationId,
			Timestamp: request.timestamp,
		},
	})

	// go b.SyncSendBLCPData(request)

	if b.Conn != nil {
		n, err := b.Conn.Write(request.toBytes())
		if err != nil {
			slog.Error("heartbeat err", "bytes", n, "error", err)
		}
	}
}

func GetBDUKfromUserId(userID string) (string, error) {
	padded := pkcs7Pad([]byte(userID), aes.BlockSize)
	block, err := aes.NewCipher([]byte("AFD311832EDEEAEF"))
	if err != nil {
		return "", err
	}
	encrypted := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, []byte("2011121211143000"))
	mode.CryptBlocks(encrypted, padded)
	return base64.RawURLEncoding.EncodeToString(encrypted), nil
}

func (b *BLCPCore) MsgKey(bduk string) string {
	// TODO fix random.randint(-LongMax, LongMax)
	return bduk + strconv.FormatInt(time.Now().UnixNano(), 10) + strconv.FormatInt(rand.Int63(), 10)
}

// default compress_type=0, need_common=1
func (b *BLCPCore) buildRpcBody(serviceId, methodId, correlationId int64, compressType int32, needCommon int32) ([]byte, error) {
	eventTimestamp := &tbpb.EventTimestamp{
		Event:       "CLCPReqBegin",
		TimestampMs: time.Now().UnixMilli(),
	}

	rpcRequestMeta := &tbpb.RpcRequestMeta{
		LogId:      correlationId,
		ServiceId:  serviceId,
		MethodId:   methodId,
		NeedCommon: needCommon,
		EventList:  []*tbpb.EventTimestamp{eventTimestamp},
	}

	rpcMeta := &tbpb.RpcMeta{
		Request:            rpcRequestMeta,
		CorrelationId:      correlationId,
		CompressType:       VPtr(compressType),
		AcceptCompressType: 1,
	}

	return proto.Marshal(rpcMeta)
}

const MaxBLCPPacketSize = 16 * 1024 * 1024

func (b *BLCPCore) readLoop() {
	buffer := make([]byte, 0, 8192)
	tmp := make([]byte, 4096)

	for {
		if b.Ctx.Err() != nil {
			// ctx done
			slog.Info("blcp read loop exit (ctx done)", "error", b.Ctx.Err)
			return
		}
		n, err := b.Conn.Read(tmp)
		if err != nil {
			slog.Info("blcp read loop exit (read error)", "error", err)
			b.CloseSignal <- struct{}{}
			return
		} else if len(buffer) > MaxBLCPPacketSize {
			slog.Info("blcp read loop exit (buffer overflow)", "buffer_length", len(buffer), "MaxBLCPPacketSize", MaxBLCPPacketSize, "error", errors.New("blcp buffer overflow"))
			b.CloseSignal <- struct{}{}
			return
		}

		buffer = append(buffer, tmp[:n]...)

		for {
			packet, remain, ok := b.parsePacket(buffer)
			if !ok {
				break
			}
			buffer = remain
			msg, err := b.ParseBLCPResponse(packet)
			// fmt.Println("ParseBLCPResponse", "msg", msg, "error", err)
			if err != nil {
				// drop package
				slog.Error("ParseBLCPResponse failed", "error", err)
				break
			}

			// if responseBLCP.isNotify {
			if share.TestMode {
				fmt.Println("<--")
			}
			b.MessageChan <- msg
			// }
		}
	}
}

// packet, remain, ok
func (b *BLCPCore) parsePacket(buf []byte) ([]byte, []byte, bool) {
	if len(buf) < 12 {
		return nil, buf, false
	}

	index := bytes.Index(buf, BlcpMagic)
	if index < 0 {
		if len(buf) > 3 {
			buf = buf[len(buf)-3:]
		}
		return nil, buf, false
	}

	if index > 0 {
		buf = buf[index:]
	}

	total := int(binary.BigEndian.Uint32(buf[4:8]))

	packetLen := 12 + total

	if len(buf) < packetLen {
		return nil, buf, false
	}

	return buf[:packetLen], buf[packetLen:], true
}

// parsePacket will provide full data
// TODO bytes -> struct -> bytes
func (b *BLCPCore) ParseBLCPResponse(receivedBytes []byte) (*BLCPWaitListContent, error) {
	origBytes := receivedBytes

	receivedBytes = receivedBytes[4:]

	totalLength := binary.BigEndian.Uint32(receivedBytes[0:4])
	rpcLength := binary.BigEndian.Uint32(receivedBytes[4:8])

	receivedBytes = receivedBytes[8:]

	rpcBytes := receivedBytes[:rpcLength]
	lcmBytes := receivedBytes[rpcLength:totalLength]

	if share.TestMode {
		fmt.Println("-->")
		fmt.Printf("original bytes(no magic): %v\n", origBytes)
		fmt.Printf("rpc meta: %v\n", rpcBytes)
		fmt.Printf("response: %v\n", lcmBytes)
	}

	rpcMeta := &tbpb.RpcMeta{}

	if err := proto.Unmarshal(rpcBytes, rpcMeta); err != nil {
		return nil, err
	}

	serviceID := rpcMeta.GetResponse().GetServiceId()
	methodID := rpcMeta.GetResponse().GetMethodId()
	correlationID := rpcMeta.GetCorrelationId()

	// heartbeat: methodID == 3 && serviceID == 1
	isNotify := rpcMeta.GetNotify() != nil && !(methodID == 3 && serviceID == 1)

	lcmType, gzipped, lcm, err := b.parseLcm(lcmBytes, isNotify)
	if err != nil {
		return nil, err
	}

	if share.TestMode {
		fmt.Println("/->")
		fmt.Println("msgid:", correlationID)
		lcmJson, _ := JsonEncode(lcm)
		rpcJson, _ := JsonEncode(rpcMeta)
		fmt.Printf("meta: %s\n", rpcJson)
		fmt.Printf("res: %s\n", lcmJson)
		fmt.Printf("lcm: %s, gzip: %t\n", lcmType, gzipped)
	}

	responseBLCP := NewBLCPData(serviceID, methodID)
	responseBLCP.correlationId = correlationID
	responseBLCP.ifRequest = false
	responseBLCP.RpcBody = rpcBytes
	responseBLCP.isNotify = isNotify

	if lcm != nil {
		responseBLCP.LcmBody = lcmBytes
	}

	res := &BLCPWaitListContent{
		IsNotify:      responseBLCP.isNotify,
		IsHeartBeat:   methodID == 3 && serviceID == 1,
		BLCPData:      responseBLCP, // <- seems useless now...
		RPC:           rpcMeta,
		LCMType:       lcmType,
		GzipLCM:       gzipped,
		CorrelationId: int(correlationID),
	}

	switch lcmType {
	case "json":
		res.JsonLCM = lcm.(map[string]any)
	case "notify":
		res.NotifyLCM = lcm.(*BLCPNotify)
	default:
		res.ProtoLCM = lcm.(*tbpb.RpcData)
	}

	return res, nil
}

func (b *BLCPCore) parseLcm(data []byte, isNotify bool) (string, bool, any, error) {
	lcmType := "protobuf"

	if isNotify {
		raw, err := GzipDecode(data)
		if err != nil {
			return "", false, nil, errors.New("parse LCM data failed")
		}
		lcmType = "notify"
		var jsonLcm BLCPNotify
		if err := JsonDecode(raw, &jsonLcm); err == nil {
			return lcmType, false, &jsonLcm, nil
		}

		return "", false, nil, errors.New("parse LCM data failed")
	}

	rpcData := &tbpb.RpcData{}

	if err := proto.Unmarshal(data, rpcData); err == nil {
		return lcmType, false, rpcData, nil
	}

	// gzip protobuf
	raw, err := GzipDecode(data)
	if err == nil {
		rpcData = &tbpb.RpcData{}
		if err := proto.Unmarshal(raw, rpcData); err == nil {
			return lcmType, true, rpcData, nil
		}
	}

	// json
	lcmType = "json"
	var jsonLcm map[string]any
	if err := JsonDecode(data, &jsonLcm); err == nil {
		return lcmType, false, jsonLcm, nil
	}

	// gzip json
	if err := JsonDecode(raw, &jsonLcm); err == nil {
		return lcmType, true, jsonLcm, nil
	}

	return "", false, nil, errors.New("parse LCM data failed")
}

type BLCPData struct {
	serviceId     int64
	methodId      int64
	RpcBody       []byte
	LcmBody       []byte
	timestamp     int64
	ifRequest     bool
	correlationId int64
	isNotify      bool
}

func NewBLCPData(serviceId, methodId int64) *BLCPData {
	return &BLCPData{
		serviceId:     serviceId,
		methodId:      methodId,
		RpcBody:       nil,
		LcmBody:       nil,
		timestamp:     time.Now().UnixMilli(),
		ifRequest:     true,
		isNotify:      false,
		correlationId: rand.Int63(),
	}
}

func NewBLCPDataWithOption(serviceId, methodId int64, rpcBody, lcmBody []byte, timestamp int64, ifRequest, isNotify bool) *BLCPData {
	if timestamp == 0 {
		timestamp = time.Now().UnixMilli()
	}

	return &BLCPData{
		serviceId:     serviceId,
		methodId:      methodId,
		RpcBody:       rpcBody,
		LcmBody:       lcmBody,
		timestamp:     timestamp,
		ifRequest:     ifRequest,
		isNotify:      isNotify,
		correlationId: rand.Int63(),
	}
}

func (b *BLCPData) toBytes() []byte {
	buf := bytes.NewBuffer(nil)

	// magic
	buf.Write(BlcpMagic)
	totalLen := len(b.RpcBody) + len(b.LcmBody)
	tmp := make([]byte, 4)

	// total length
	binary.BigEndian.PutUint32(tmp, uint32(totalLen))
	buf.Write(tmp)

	// RPC length
	binary.BigEndian.PutUint32(tmp, uint32(len(b.RpcBody)))
	buf.Write(tmp)

	// RPC body
	buf.Write(b.RpcBody)

	// LCM body
	buf.Write(b.LcmBody)

	return buf.Bytes()
}
