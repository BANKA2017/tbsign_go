package _function

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// JoinChatRoom why we need this?
func (b *BLCPCore) JoinChatRoom(chatroomID int) error {
	if b.State != BLCPStateLoggedIn {
		return errors.New("im not login")
	}

	request := NewBLCPData(3, 201)
	request.RpcBody, _ = b.buildRpcBody(3, 201, request.correlationId, 0, 1)

	rpcConfig, _ := JsonEncode(map[string]int{"rpc_retry_time": 0})

	reqdict := map[string]any{
		"method":       201,
		"mcast_id":     chatroomID,
		"appid":        ChatAppId,
		"uk":           b.Client.Account.UK,
		"origin_id":    b.Client.Account.TriggerID,
		"msg_key":      "k" + strconv.FormatInt(time.Now().Unix()*100000, 10),
		"sdk_version":  ChatSdkVersion,
		"is_reliable":  false,
		"client_logid": time.Now().Nanosecond(),
		"rpc":          string(rpcConfig),
	}
	loginLcmBody, err := JsonEncode(reqdict)
	if err != nil {
		return err
	}
	request.LcmBody = loginLcmBody
	resp, err := b.SyncSendBLCPData(request)
	if err != nil {
		return err
	} else if resp.LCMType != "json" {
		return errors.New("wrong lcm response type")
	} else if int(resp.JsonLCM["err_code"].(float64)) != 0 {
		return errors.New("BLCP Handshake error.3")
	}

	// await self.fetch_mcast_msg_client_request(self.account.cuid_galaxy2, chatroom_id)

	return nil
}

// QuitChatRoom is not work at all!!!
// methodId=202
func (b *BLCPCore) QuitChatRoom(chatroomID int) error {
	if b.State != BLCPStateLoggedIn {
		return errors.New("im not login")
	}

	// request := NewBLCPData(3, 202)
	// request.RpcBody, _ = b.buildRpcBody(3, 202, request.correlationId, 0, 1)
	//
	// rpcConfig, _ := JsonEncode(map[string]int{"rpc_retry_time": 0})
	//
	// reqdict := map[string]any{
	// 	"method":       202,
	// 	"mcast_id":     chatroomID,
	// 	"appid":        ChatAppId,
	// 	"uk":           b.Client.Account.UK,
	// 	"origin_id":    b.Client.Account.TriggerID,
	// 	"msg_key":      "k" + strconv.FormatInt(time.Now().Unix()*100000, 10),
	// 	"sdk_version":  ChatSdkVersion,
	// 	"client_logid": time.Now().Nanosecond(),
	// 	"rpc":          string(rpcConfig),
	// }
	// loginLcmBody, err := JsonEncode(reqdict)
	// if err != nil {
	// 	return err
	// }
	// request.LcmBody = loginLcmBody
	// resp, err := b.SyncSendBLCPData(request)
	// if err != nil {
	// 	return err
	// } else if resp.LCMType != "json" {
	// 	return errors.New("wrong lcm response type")
	// } else if int(resp.JsonLCM["err_code"].(float64)) != 0 {
	// 	return errors.New("BLCP Handshake error.3")
	// }

	return nil
}

type ChatroomMessageAtData struct {
	AtType          string `json:"at_type"`
	AtBaiduUk       string `json:"at_baidu_uk"`
	AtName          string `json:"at_name"`
	AtVip           string `json:"at_vip,omitempty"`
	AtCharacter     string `json:"at_character,omitempty"`
	AtCharacterName string `json:"at_character_name,omitempty"`
	AtPortrait      string `json:"at_portrait"`
	Position        int    `json:"position"`
}

type SendChatroomMessageUserMainData struct {
	Type int                                  `json:"type"`
	Text *SendChatroomMessageUserMainDataText `json:"text,omitempty"`
	Icon *SendChatroomMessageUserMainDataIcon `json:"icon,omitempty"`
}

type SendChatroomMessageUserMainDataText struct {
	Str           string                                    `json:"str"`
	Priority      int                                       `json:"priority"`
	ShortEnable   int                                       `json:"short_enable"`
	ShortLength   int                                       `json:"short_length"`
	ShortPriority int                                       `json:"short_priority"`
	Suffix        string                                    `json:"suffix"`
	Type          string                                    `json:"type"`
	TextColor     *SendChatroomMessageUserMainDataTextColor `json:"text_color,omitempty"`
}

type SendChatroomMessageUserMainDataTextColor struct {
	Day   string `json:"day"`
	Night string `json:"night"`
	Type  int    `json:"type"`
}

type SendChatroomMessageUserMainDataIcon struct {
	URL      string `json:"url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Priority int    `json:"priority"`
	Type     string `json:"type"`
	Schema   string `json:"schema"`
}

type SendChatroomMessageExt struct {
	ForumID int `json:"forum_id"`
	Level   int `json:"level"`

	Content struct {
		RobotParams *SendChatroomMessageExtRobot `json:"robot_params,omitempty"`
	} `json:"content"`
	From     string                            `json:"from"`
	Portrait string                            `json:"portrait"`
	MainData []SendChatroomMessageUserMainData `json:"main_data"`
	Type     int                               `json:"type"`
	UserName string                            `json:"user_name"`

	IsSysMsg   int    `json:"is_sys_msg"`
	Version    string `json:"version"`
	RobotRole  int    `json:"robot_role"`
	Role       int    `json:"role"`
	SendStatus int    `json:"send_status"`
	SessionID  int    `json:"session_id"`
}

type SendChatroomMessageExtRobot struct {
	Scene string `json:"scene"`
	Type  string `json:"type"`
}

type SendChatroomMessageContentText struct {
	RoomID      string                   `json:"room_id"`
	Type        string                   `json:"type"`
	ToUID       string                   `json:"to_uid"`
	Vip         string                   `json:"vip"`
	Name        string                   `json:"name"`
	Portrait    string                   `json:"portrait"`
	ContentType string                   `json:"content_type"`
	ContentBody string                   `json:"content_body"`
	Src         string                   `json:"src"`
	BaiduUk     string                   `json:"baidu_uk"`
	Ext         string                   `json:"ext"`
	AtData      *[]ChatroomMessageAtData `json:"at_data,omitempty"`
}

// default robot=-1
func (b *BLCPCore) SendChatroomMessage(chatroomID int, text string, isVIP bool, forumLevel, growthLevel, fid, robot int, atData []ChatroomMessageAtData) (*BLCPWaitListContent, error) {
	name := b.Client.Account.Name
	portrait := b.Client.Account.Portrait // + "?t=" + strconv.FormatInt(time.Now().Unix(), 10)

	appSafeExt, _ := JsonEncode(map[string]map[string]string{"haotianjing": {"zid": ""}}) // zid must be "" b.Client.Zid
	strContent, _ := JsonEncode(McastMessageTextObject{Text: text})

	contentText := &SendChatroomMessageContentText{
		RoomID:      strconv.FormatInt(int64(chatroomID), 10),
		Type:        "0",
		ToUID:       "0",
		Vip:         When(isVIP, "1", "0"),
		Name:        name,
		Portrait:    portrait,
		ContentType: "0",
		ContentBody: string(strContent),
		Src:         "",
		BaiduUk:     b.Client.Account.BDUK,
		Ext:         "",
	}

	// 名字、头像、昵称颜色、大会员标志等UI展示信息
	var mainData []SendChatroomMessageUserMainData
	if isVIP {
		mainData = append(mainData, SendChatroomMessageUserMainData{
			Type: 2,
			Icon: &SendChatroomMessageUserMainDataIcon{
				Height:   75,
				Priority: 2,
				Schema:   "https://tieba.baidu.com/mo/q/hybrid-business-vip/tbvip?customfullscreen=1&nonavigationbar=1",
				Type:     "1",
				URL:      "https://tieba-ares.cdn.bcebos.com/mis/2023-7/1689061482682/13afea50121d.png",
				Width:    75,
			},
		})
	}

	namedata := SendChatroomMessageUserMainData{
		Type: 1,
		Text: &SendChatroomMessageUserMainDataText{
			ShortEnable:   1,
			ShortLength:   5,
			ShortPriority: 1,
			Priority:      1,
			Str:           name,
			Suffix:        "...",
			Type:          "1",
		},
	}

	if isVIP {
		namedata.Text.TextColor = &SendChatroomMessageUserMainDataTextColor{
			Day:   "CAM_X0301",
			Night: "CAM_X0301",
			Type:  2,
		}
	}

	mainData = append(
		mainData,
		namedata, // <- not necessary
		SendChatroomMessageUserMainData{
			Type: 2,
			Icon: &SendChatroomMessageUserMainDataIcon{
				Height:   15,
				Priority: 5,
				Schema:   "https://tieba.baidu.com/mo/q/hybrid-main-user/taskCenter?customfullscreen=1&nonavigationbar=1",
				Type:     "3",
				URL:      "local://icon/icon_mask_level_usergrouth_" + strconv.FormatInt(int64(growthLevel), 10) + "?type=webp",
				Width:    24,
			},
		},
		SendChatroomMessageUserMainData{
			Type: 2,
			Icon: &SendChatroomMessageUserMainDataIcon{
				Height:   12,
				Priority: 2,
				Schema:   "https://tieba.baidu.com/mo/q/wise-bawu-core/forum-level?customfullscreen=1&forum_id=" + strconv.FormatInt(int64(fid), 10) + "&nonavigationbar=1&obj_locate=5&portrait=" + portrait,
				Type:     "4",
				URL:      "local://icon/icon_level_" + fmt.Sprintf("%02d", forumLevel) + "?type=webp",
				Width:    16,
			},
		},
	)

	extData := &SendChatroomMessageExt{
		Type:       1,
		MainData:   mainData,
		IsSysMsg:   0,
		Version:    "",
		Portrait:   portrait,
		RobotRole:  0,
		Role:       0,
		SendStatus: 0,
		From:       BlcpPlatform,
		SessionID:  chatroomID,
		UserName:   name,
	}

	// 携带文字、气泡信息，但这部分内容比较多，暂且未开发，注释之
	// content['text']['ext']['bubble_info'] = {'color_info': {'at_text_color': '', 'at_text_color_dark': 'AF9EFF', 'text_color': '141414', 'text_color_dark': 'FFFFFF'}, 'end_time': 0, 'id': 1380005, 'img_info': {'android_left': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678971300129/4d9fd67c26e3.png', 'android_left_dark': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678971300129/4d9fd67c26e3.png', 'android_right': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678971357436/745549b2ae39.png', 'android_right_dark': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678971357436/745549b2ae39.png', 'ios_left': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678876082814/0a1741207e0d.png', 'ios_left_dark': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678876082814/0a1741207e0d.png', 'ios_right': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678876082435/c07fbaa4ad57.png', 'ios_right_dark': 'https://tieba-ares.cdn.bcebos.com/mis/2023-3/1678876082435/c07fbaa4ad57.png'}, 'jump_url': 'https://tieba.baidu.com/mo/q/hybrid/decorator?pageType=4&from_page=4&nonavigationbar=1&customfullscreen=1&decoratorId=1380005', 'type': 2}

	if robot != -1 {
		extData.Content.RobotParams = &SendChatroomMessageExtRobot{
			Scene: "tieba_group_chat",
			Type:  strconv.FormatInt(int64(robot), 10),
		}
	}

	extData.Level = forumLevel
	extData.ForumID = fid

	if len(atData) > 0 {
		contentText.AtData = &atData
	}

	byteExtData, _ := JsonEncode(extData)
	contentText.Ext = string(byteExtData)

	byteContentText, _ := JsonEncode(contentText)

	content := map[string]string{
		"text": string(byteContentText),
	}

	contentJson, _ := JsonEncode(content)

	requestData := map[string]any{
		"method":       185,
		"mcast_id":     chatroomID,
		"role":         3,
		"token":        b.Client.Account.Bduss,
		"appid":        ChatAppId,
		"uk":           b.Client.Account.UK,
		"origin_id":    b.Client.Account.TriggerID,
		"type":         81,
		"app_safe_ext": string(appSafeExt),
		"content":      string(contentJson),
		"msg_key":      b.MsgKey(b.Client.Account.BDUK),
		"account_type": 1,
		"sdk_version":  ChatSdkVersion,
		"event_list": []map[string]any{
			{"event": "CClickSendBegin", "timestamp_ms": 0},
			{"event": "CSendBegin", "timestamp_ms": 0},
			{"event": "CIMSendBegin", "timestamp_ms": time.Now().UnixMilli()},
		},
	}

	bplcReq := NewBLCPData(3, 185)
	bplcReq.correlationId = time.Now().UnixNano()
	bplcReq.RpcBody, _ = b.buildRpcBody(3, 185, bplcReq.correlationId, 0, 1)

	requestData["client_logid"] = bplcReq.correlationId
	requestData["rpc"] = `{"rpc_retry_time":0}`

	bplcReq.LcmBody, _ = JsonEncode(requestData)

	// fmt.Println(string(bplcReq.LcmBody))

	return b.SyncSendBLCPData(bplcReq)
}
