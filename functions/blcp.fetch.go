package _function

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	_type "github.com/BANKA2017/tbsign_go/types"
	"golang.org/x/exp/slices"
)

type GenerateLcmTokenResponse struct {
	// RequestID string `json:"request_id"`
	ErrorCode string `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	Token     string `json:"token"`
	// Protocols []struct {
	// 	Protocol string `json:"protocol"`
	// 	Domain   string `json:"domain"`
	// 	Port     string `json:"port"`
	// } `json:"protocols"`
	// ClientLogConfig struct {
	// 	ClientUploadLogSwitch int `json:"client_upload_log_switch"`
	// 	RealtimeLogSwitch     []struct {
	// 		ID     int `json:"id"`
	// 		Switch int `json:"switch"`
	// 	} `json:"realtime_log_switch"`
	// 	ClientPingSwitch int `json:"client_ping_switch"`
	// 	ClientPingConfig struct {
	// 		PingTimes      int    `json:"ping_times"`
	// 		PingBodySize   int    `json:"ping_body_size"`
	// 		PingDeadlineMs int    `json:"ping_deadline_ms"`
	// 		PingTimeoutMs  int    `json:"ping_timeout_ms"`
	// 		PingIntervalMs int    `json:"ping_interval_ms"`
	// 		PingTarget     string `json:"ping_target"`
	// 	} `json:"client_ping_config"`
	// } `json:"client_log_config"`
	// ClientTestConfig struct {
	// 	ClientTestConfigSwitch int   `json:"client_test_config_switch"`
	// 	ClientTestSwitch       []any `json:"client_test_switch"`
	// } `json:"client_test_config"`
	// BddnsEnable  bool `json:"bddns_enable"`
	// Ipv6Strategy int  `json:"ipv6_strategy"`
}

func (b *BLCPCore) GenerateLcmToken() (*GenerateLcmTokenResponse, error) {
	cuidGalaxy2, _ := b.Client.CuidGalaxy2()

	// request_id === string(ts)
	ts := time.Now().UnixMilli()
	requestId := strconv.Itoa(int(ts))

	bodyPayload := map[string]any{
		"app_id":      strconv.Itoa(ChatAppId),
		"app_version": ChatVersion,
		"cuid":        cuidGalaxy2,
		"device_type": BlcpPlatform,
		"manufacture": "",
		"model_type":  "",
		"request_id":  requestId,
		"sdk_version": LcmSdkVersion,
		"sign":        Md5(strconv.Itoa(ChatAppId) + cuidGalaxy2 + BlcpPlatform + requestId), // appid
		"ts":          ts,
		"user_key":    "",
	}

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/5.0/generate_lcm_token", http.MethodPost, body, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
	})

	if err != nil {
		return nil, err
	}

	var resp GenerateLcmTokenResponse
	err = JsonDecode(res, &resp)
	return &resp, err
}

type EnterChatroomClientResponse struct {
	RequestID int    `json:"request_id"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	// TraceLocation string `json:"trace_location"`
	RoomID  int    `json:"room_id"`
	BaiduUK string `json:"baidu_uk"`

	RoomInfo struct {
		Appid       int    `json:"appid"`
		RoomID      int    `json:"room_id"`
		ChatMcastID int    `json:"chat_mcast_id"`
		RoomType    int    `json:"room_type"`
		RoomName    string `json:"room_name"`
		RoomImage   string `json:"room_image"`
		RoomNotice  string `json:"room_notice"`
		ResourceID  string `json:"resource_id"`
		Ext         string `json:"ext"`
		State       int    `json:"state"`
		UserCount   int    `json:"user_count"`
	} `json:"room_info"`
	UserInfo struct {
		Character     int    `json:"character"`
		CharacterName string `json:"character_name"`
		Vip           int    `json:"vip"`
	} `json:"user_info"`
}

func (b *BLCPCore) EnterChatroomClient(roomID, accountType int) (*EnterChatroomClientResponse, error) {
	cuidGalaxy2, _ := b.Client.CuidGalaxy2()

	bodyPayload := map[string]any{
		"appid":        ChatAppId,
		"room_id":      roomID,
		"app_version":  ChatVersion,
		"device_id":    cuidGalaxy2,
		"sdk_version":  ChatSdkVersion,
		"timestamp":    time.Now().Unix(),
		"account_type": accountType,
	}

	BlcpHTTPClientSign(bodyPayload)

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/3.0/im/chatroom/enter_chatroom_client", http.MethodPost, body, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
		"Cookie":          "BDUSS=" + b.Client.Account.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var resp EnterChatroomClientResponse
	err = JsonDecode(res, &resp)
	return &resp, err
}

type ExitChatroomClientResponse struct {
	RequestID int    `json:"request_id"`
	ErrorCode int    `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`
	// TraceLocation string `json:"trace_location"`
}

func (b *BLCPCore) ExitChatroomClient(roomID int) (*ExitChatroomClientResponse, error) {
	cuidGalaxy2, _ := b.Client.CuidGalaxy2()

	bodyPayload := map[string]any{
		"appid":       ChatAppId,
		"room_id":     roomID,
		"app_version": ChatVersion,
		"device_id":   cuidGalaxy2,
		"sdk_version": ChatSdkVersion,
		"timestamp":   time.Now().Unix(),
	}

	BlcpHTTPClientSign(bodyPayload)

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/3.0/im/chatroom/exit_chatroom_client", http.MethodPost, body, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
		"Cookie":          "BDUSS=" + b.Client.Account.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var resp ExitChatroomClientResponse
	err = JsonDecode(res, &resp)
	return &resp, err
}

type FetchMcastMsgClientResponse struct {
	// RequestID     int                                   `json:"request_id"`
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	// TraceLocation string                                `json:"trace_location"`
	Messages  *[]FetchMcastMsgClientResponseMessage `json:"messages,omitempty"`
	HasMore   bool                                  `json:"has_more"`
	ErrorCode int                                   `json:"error_code"`
	ErrorMsg  string                                `json:"error_msg"`
}

type FetchMcastMsgClientResponseMessage struct {
	Msgid         int64  `json:"msgid"`
	UserType      int    `json:"user_type"`
	ContacterType int    `json:"contacter_type"`
	ContacterUk   int    `json:"contacter_uk"`
	PaUID         int    `json:"pa_uid"`
	Contacter     int64  `json:"contacter"`
	FromUser      int64  `json:"from_user"`
	Category      int    `json:"category"`
	Content       string `json:"content"`
	CreateTime    int    `json:"create_time"`
	ToUser        int    `json:"to_user"`
	GroupID       int    `json:"group_id"`
	McastID       int64  `json:"mcast_id"`
	GroupMsgid    int    `json:"group_msgid"`
	IsRead        int    `json:"is_read"`
	Type          int    `json:"type"`
	Cmd           int    `json:"cmd"`
	OriginID      int    `json:"origin_id"`
	Appid         int    `json:"appid"`
	MsgKey        string `json:"msg_key"`
	MsgExpires    int    `json:"msg_expires"`
	BusinessType  int    `json:"business_type"`
}

type McastMessageTextObject struct {
	Text string `json:"text"`
}

type McastMessageTextObjectText struct {
	RoomID        string `json:"room_id"`
	Type          string `json:"type"`
	BaiduUk       string `json:"baidu_uk"`
	Character     string `json:"character"`
	CharacterName string `json:"character_name"`
	Vip           string `json:"vip"`
	Name          string `json:"name"`
	Portrait      string `json:"portrait"`
	ContentType   string `json:"content_type"`
	ContentBody   string `json:"content_body"`
	AtData        []struct {
		AtType          string `json:"at_type"`
		AtBaiduUk       string `json:"at_baidu_uk"`
		AtName          string `json:"at_name"`
		AtVip           string `json:"at_vip"`
		AtCharacter     string `json:"at_character"`
		AtCharacterName string `json:"at_character_name"`
		AtPortrait      string `json:"at_portrait"`
		// Position        int    `json:"position"` // <- int or string
	} `json:"at_data"`
	Ext string `json:"ext"`
}

// default count=30
func (b *BLCPCore) FetchMcastMsgClient(roomID, accountType, count int) (*FetchMcastMsgClientResponse, error) {
	cuidGalaxy2, _ := b.Client.CuidGalaxy2()

	// extInfoJSON, _ := JsonEncode(map[string]int{"last_callback_msg_id": 0, "cast_id": 0, "local_ts": 0, "latest_msg_id": 0})
	// extInfo := url.QueryEscape(string(extInfoJSON))
	// extInfo = strings.ReplaceAll(extInfo, "+", "%20")
	extInfo := `%7B%22cast_id%22%3A0%2C%22last_callback_msg_id%22%3A0%2C%22latest_msg_id%22%3A0%2C%22local_ts%22%3A0%7D`

	bodyPayload := map[string]any{
		"appid":        ChatAppId,
		"mcast_id":     roomID,
		"msgid_begin":  0,
		"msgid_end":    math.MaxInt64,
		"count":        -count,
		"category":     4,
		"app_version":  ChatVersion,
		"sdk_version":  ChatSdkVersion,
		"device_id":    cuidGalaxy2,
		"device_type":  2,
		"from_action":  1,
		"ext_info":     extInfo,
		"timestamp":    time.Now().Unix(),
		"account_type": accountType,
	}

	BlcpHTTPClientSign(bodyPayload)

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/3.0/im/fetch_mcast_msg_client", http.MethodPost, body, map[string]string{
		"Content-Type":    "application/json",
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
		"Cookie":          "BDUSS=" + b.Client.Account.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var resp FetchMcastMsgClientResponse
	err = JsonDecode(res, &resp)
	return &resp, err
}

// ?idk
func (b *BLCPCore) GroupChat() (any, error) {
	ts := time.Now().Unix()

	bodyPayload := map[string]any{
		"method":    "get_joined_groups",
		"appid":     ChatAppId,
		"timestamp": ts,
		"sign":      Md5(strconv.FormatInt(ts, 10) + b.Client.Account.Bduss + strconv.FormatInt(ChatAppId, 10)),
	}

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/2.0/im/groupchat", http.MethodPost, body, map[string]string{
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
		"Cookie":          "BDUSS=" + b.Client.Account.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var resp any
	err = JsonDecode(res, &resp)
	return &resp, err
}

// ?idk
func (b *BLCPCore) GroupChatV1() (any, error) {
	cuidGalaxy2, _ := b.Client.CuidGalaxy2()

	ts := time.Now().Unix()

	bodyPayload := map[string]any{
		"method":      "get_joined_groups",
		"appid":       ChatAppId,
		"timestamp":   ts,
		"sign":        Md5(strconv.FormatInt(ts, 10) + b.Client.Account.Bduss + strconv.FormatInt(ChatAppId, 10)),
		"group_type":  3,
		"source":      0,
		"cuid":        cuidGalaxy2,
		"app_version": ChatVersion,
		"sdk_version": strconv.FormatInt(ChatSdkVersion, 10),
		"device_type": 2,
	}

	body, _ := JsonEncode(bodyPayload)

	res, err := TBFetch("https://pim.baidu.com/rest/2.0/im/groupchatv1", http.MethodPost, body, map[string]string{
		"Accept-Encoding": "gzip",
		"User-Agent":      BlcpUserAgent,
		"Cookie":          "BDUSS=" + b.Client.Account.Bduss,
	})

	if err != nil {
		return nil, err
	}

	var resp any
	err = JsonDecode(res, &resp)
	return &resp, err
}

func BlcpHTTPClientSign(form map[string]any) {
	var formKeys []string
	for k := range form {
		formKeys = append(formKeys, k)
	}

	slices.Sort(formKeys)

	var payload strings.Builder

	for _, v := range formKeys {
		payload.WriteString(v)
		payload.WriteString("=")
		switch any(form[v]).(type) {
		case string:
			payload.WriteString(form[v].(string))
		case int64, int32, int16, int8, int:
			fmt.Fprintf(&payload, "%d", form[v])
		case float64, float32:
			// TODO really need float?
			fmt.Fprintf(&payload, "%f", form[v])
		case bool:
			payload.WriteString(When(form[v].(bool), "true", "false"))
		}
	}

	form["sign"] = strings.ToLower(Md5(payload.String()))
}

type GetRoomListByFidResponse struct {
	Data struct {
		MidGroupID string `json:"mid_group_id"`
		List       []struct {
			GroupName   string `json:"group_name"`
			GroupID     string `json:"group_id"`
			GroupWeight string `json:"group_weight"`
			RoomList    []struct {
				RoomID   string `json:"room_id"`
				RoomInfo struct {
					ForumID   string `json:"forum_id"`
					ForumName string `json:"forum_name"`
					GroupID   string `json:"group_id"`
					Name      string `json:"name"`
					Avatar    string `json:"avatar"`
					UnreadNum string `json:"unread_num"`
				} `json:"room_info"`
				MsgInfo struct {
					FromUID        string `json:"from_uid"`
					FromName       string `json:"from_name"`
					Content        string `json:"content"`
					MsgID          string `json:"msg_id"`
					MsgTime        string `json:"msg_time"`
					SpecialMsg     string `json:"special_msg"`
					SpecialMsgType string `json:"special_msg_type"`
				} `json:"msg_info"`
				JumpURL    string `json:"jump_url"`
				RoomWeight string `json:"room_weight"`
				// AtInfo     struct {
				// 	AtCountAll       string `json:"at_count_all"`
				// 	AtAllMsgCount    string `json:"at_all_msg_count"`
				// 	AtSingleMsgCount string `json:"at_single_msg_count"`
				// 	AtAllMsgList     []any  `json:"at_all_msg_list"`
				// 	AtSingleMsgList  []any  `json:"at_single_msg_list"`
				// } `json:"at_info"`
			} `json:"room_list"`
		} `json:"list"`
		// Footer struct {
		// 	Src  string `json:"src"`
		// 	Link string `json:"link"`
		// } `json:"footer"`
		// Banner struct {
		// 	HeadImg       []string `json:"head_img"`
		// 	IsLike        string   `json:"is_like"`
		// 	FrsScheme     string   `json:"frs_scheme"`
		// 	HotNumsText   string   `json:"hot_nums_text"`
		// 	ManagerSwitch string   `json:"manager_switch"`
		// } `json:"banner"`
	} `json:"data"`

	ErrorMsg  string `json:"error_msg"`
	ErrorCode string `json:"error_code"`
}

func GetRoomListByFid(fid int) (*GetRoomListByFidResponse, error) {
	var form = make(map[string]string)
	form["subapp_type"] = "tieba"
	form["fid"] = strconv.FormatInt(int64(fid), 10)

	ClientTypeFallBack(form, "android")
	AddSign(form, "android")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	res, err := TBFetch("https://tiebac.baidu.com/c/f/chat/getRoomListByFid", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp GetRoomListByFidResponse
	err = JsonDecode(res, &resp)

	return &resp, err
}

type GetChatRoomDetailResponse struct {
	ErrorMsg  string `json:"error_msg"`
	ErrorCode string `json:"error_code"`

	Data struct {
		BasicInfo struct {
			ID         string `json:"id"`
			ForumID    string `json:"forum_id"`
			GroupID    string `json:"group_id"`
			Name       string `json:"name"`
			Avatar     string `json:"avatar"`
			IsDelete   string `json:"is_delete"`
			CreateTime string `json:"create_time"`
			DiscussNum string `json:"discuss_num"`
			ForumName  string `json:"forum_name"`
		} `json:"basic_info"`
		ExtraInfo struct {
			ChatroomID             string `json:"chatroom_id"`
			Weight                 string `json:"weight"`
			Announce               string `json:"announce"`
			TalkThresholdLevel     string `json:"talk_threshold_level"`
			ChatroomViewType       string `json:"chatroom_view_type"`
			ForumEntranceSwitch    string `json:"forum_entrance_switch"`
			ForumEntranceCloseType string `json:"forum_entrance_close_type"`
			IsFreezing             string `json:"is_freezing"`
			FreezingType           string `json:"freezing_type"`
			IsTalkBanned           string `json:"is_talk_banned"`
			AnnounceOpTime         string `json:"announce_op_time"`
			OpTime                 string `json:"op_time"`
			TalkType               string `json:"talk_type"`
		} `json:"extra_info"`
		UserInfo struct {
			IsCanViewChatroom string `json:"is_can_view_chatroom"`
			IsBeenTalkBanned  string `json:"is_been_talk_banned"`
			IsSubscription    string `json:"is_subscription"`
			IdentityRole      string `json:"identity_role"`
			ForumLevel        string `json:"forum_level"`
			UserGrowthLevel   string `json:"user_growth_level"`
			IsCanReaction     string `json:"is_can_reaction"`
			BazhuType         string `json:"bazhu_type"`
			BubbleInfo        struct {
				Type    string `json:"type"`
				ImgInfo struct {
					AndroidLeft      string `json:"android_left"`
					AndroidRight     string `json:"android_right"`
					IosLeft          string `json:"ios_left"`
					IosRight         string `json:"ios_right"`
					AndroidLeftDark  string `json:"android_left_dark"`
					AndroidRightDark string `json:"android_right_dark"`
					IosLeftDark      string `json:"ios_left_dark"`
					IosRightDark     string `json:"ios_right_dark"`
				} `json:"img_info"`
				EndTime   string `json:"end_time"`
				ColorInfo struct {
					TextColor       string `json:"text_color"`
					TextColorDark   string `json:"text_color_dark"`
					AtTextColor     string `json:"at_text_color"`
					AtTextColorDark string `json:"at_text_color_dark"`
				} `json:"color_info"`
				JumpURL string `json:"jump_url"`
				ID      string `json:"id"`
			} `json:"bubble_info"`
			MainData []struct {
				Type string `json:"type"`
				Text *struct {
					Str           string `json:"str"`
					Priority      string `json:"priority"`
					ShortEnable   string `json:"short_enable"`
					ShortLength   string `json:"short_length"`
					ShortPriority string `json:"short_priority"`
					Suffix        string `json:"suffix"`
					Type          string `json:"type"`
				} `json:"text,omitempty"`
				Icon *struct {
					URL      string `json:"url"`
					Width    string `json:"width"`
					Height   string `json:"height"`
					Priority string `json:"priority"`
					Type     string `json:"type"`
					Schema   string `json:"schema"`
				} `json:"icon,omitempty"`
			} `json:"main_data"`
		} `json:"user_info"`
		Bots struct {
			Guide   string `json:"guide"`
			BotList []struct {
				Scene   string `json:"scene"`
				Guide   string `json:"guide"`
				AtGuide string `json:"at_guide"`
				Skill   []struct {
					Type     string `json:"type"`
					Name     string `json:"name"`
					Alias    string `json:"alias"`
					Desc     string `json:"desc"`
					Exp      string `json:"exp"`
					Exec     string `json:"exec"`
					CallType string `json:"call_type"`
					Keyboard string `json:"keyboard"`
					Items    []struct {
						Name  string `json:"name"`
						Value string `json:"value"`
						Opts  []struct {
							Name    string `json:"name"`
							Default string `json:"default"`
						} `json:"opts"`
						Type string `json:"type"`
					} `json:"items"`
					Version struct {
						Ios     string `json:"ios"`
						Android string `json:"android"`
					} `json:"version"`
					IsShow string `json:"is_show"`
				} `json:"skill"`
				User struct {
					Name       string `json:"name"`
					Portrait   string `json:"portrait"`
					NameShow   string `json:"name_show"`
					Uk         string `json:"uk"`
					RobotRole  string `json:"robot_role"`
					ForumLevel string `json:"forum_level"`
				} `json:"user"`
			} `json:"bot_list"`
		} `json:"bots"`
		HasChated  string `json:"has_chated"`
		TextIsShow string `json:"text_is_show"`
		ShowText   string `json:"show_text"`
		MaskInfo   []struct {
			MaskID  string `json:"mask_id"`
			Content string `json:"content"`
		} `json:"mask_info"`
		Conf struct {
			ChatroomResourceMaxTimeDelay string `json:"chatroom_resource_max_time_delay"`
		} `json:"conf"`
	} `json:"data"`
}

func GetChatRoomDetail(cookie *_type.TypeCookie, chatroomID int) (*GetChatRoomDetailResponse, error) {
	var form = make(map[string]string)
	form["subapp_type"] = "tieba"
	form["BDUSS"] = cookie.Bduss
	form["chatroom_id"] = strconv.FormatInt(int64(chatroomID), 10)

	ClientTypeFallBack(form, "android")
	AddSign(form, "android")
	_body := url.Values{}
	for k, v := range form {
		if k != "sign" {
			_body.Set(k, v)
		}
	}

	res, err := TBFetch("https://tiebac.baidu.com/c/f/chat/getChatRoomDetail", http.MethodPost, []byte(_body.Encode()+"&sign="+form["sign"]), EmptyHeaders)

	if err != nil {
		return nil, err
	}

	var resp GetChatRoomDetailResponse
	err = JsonDecode(res, &resp)

	return &resp, err
}
