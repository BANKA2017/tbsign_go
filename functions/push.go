package _function

import (
	"errors"
	"fmt"
	"mime"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BANKA2017/tbsign_go/model"
	"github.com/emersion/go-smtp"
	"golang.org/x/exp/slices"

	"github.com/emersion/go-sasl"
	"github.com/google/uuid"
)

var MessageTypeList = []string{"email", "ntfy", "bark"}

func VerifyEmail(email string) bool {
	if email == "" {
		return false
	}
	subEmailStr := strings.Split(email, "@")
	if len(subEmailStr) != 2 {
		return false
	}

	if len(subEmailStr[0]) > 64 {
		return false
	}

	return len(regexp.MustCompile(`(?m)^[\w\.\-]+@\w+(?:[\.\-]\w+)*$`).FindAllString(email, -1)) == 1
}

func SendMessage(_type string, uid int32, _subject, _body string) error {
	if _type == "default" {
		_type = GetUserOption("go_message_type", strconv.Itoa(int(uid)))
		if !slices.Contains(MessageTypeList, _type) {
			_type = "email"
		}
	}

	switch _type {
	case "ntfy":
		ntfyTopic := GetUserOption("go_ntfy_topic", strconv.Itoa(int(uid)))
		return SendNtfy(ntfyTopic, _subject, _body)
	case "bark":
		barkKey := GetUserOption("go_bark_key", strconv.Itoa(int(uid)))
		return SendBark(barkKey, _subject, _body)
	default:
		accountInfo := new(model.TcUser)
		GormDB.R.Where("id = ?", uid).Find(accountInfo)
		return SendEmail(accountInfo.Email, _subject, _body)
	}
}

type NtfyResponseStruct struct {
	ID      string `json:"id,omitempty"`
	Time    int64  `json:"time,omitempty"`
	Expires int64  `json:"expires,omitempty"`
	Event   string `json:"event,omitempty"`
	Topic   string `json:"topic,omitempty"`
	Message string `json:"message,omitempty"`
}

func SendNtfy(_to, title, body string) error {
	if _to == "" {
		return errors.New("ntfy: topic is empty")
	}
	// get custom address
	ntfyAddr := GetOption("go_ntfy_addr")
	if ntfyAddr == "" {
		ntfyAddr = "https://ntfy.sh"
	}

	res, err := Fetch(AppendStrings(ntfyAddr, "/", _to), "POST", []byte(strings.ReplaceAll(body, "<br />", "\n")), map[string]string{
		"Title":        title,
		"Content-Type": "text/plain",
		"Tags":         "tbsign",
	}, DefaultCient)

	if err != nil {
		return err
	}

	resp := new(NtfyResponseStruct)
	err = JsonDecode(res, resp)
	if err != nil {
		return err
	}

	if resp.ID != "" {
		return nil
	} else {
		return fmt.Errorf("ntfy: %s", string(res))
	}
}

type BarkResponseStruct struct {
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

func SendBark(_to, title, body string) error {
	if _to == "" {
		return errors.New("bark: key is empty")
	}
	// get custom address
	barkAddr := GetOption("go_bark_addr")
	if barkAddr == "" {
		barkAddr = "https://api.day.app"
	}

	_body := url.Values{}

	_body.Set("title", title)
	_body.Set("body", strings.ReplaceAll(body, "<br />", "\n"))
	_body.Set("device_key", _to)
	_body.Set("group", "tbsign")

	res, err := Fetch(AppendStrings(barkAddr, "/push"), "POST", []byte(_body.Encode()), map[string]string{}, DefaultCient)
	if err != nil {
		return err
	}

	resp := new(BarkResponseStruct)
	err = JsonDecode(res, resp)
	if err != nil {
		return err
	}

	if resp.Code == 200 {
		return nil
	} else {
		return fmt.Errorf("bark: %s", resp.Message)
	}
}

// TODO GPG?
func SendEmail(_to, title, body string) error {
	mail := GetOption("mail_name")
	mail_name := GetOption("mail_yourname")
	smtp_host := GetOption("mail_host")
	smtp_port := GetOption("mail_port")
	smtp_secure := GetOption("mail_secure")
	smtp_auth := GetOption("mail_auth")
	smtp_username := GetOption("mail_smtpname")
	smtp_password := GetOption("mail_smtppw")

	if mail == "" || mail_name == "" || smtp_host == "" || smtp_port == "" || smtp_secure == "" {
		return errors.New("mail: Email settings not completed")
	}

	if smtp_auth != "0" && (smtp_username == "" || smtp_password == "") {
		return errors.New("mail: Login failed")
	}

	var client sasl.Client
	if smtp_auth == "1" {
		client = sasl.NewLoginClient(smtp_username, smtp_password)
	} else if smtp_auth == "2" {
		// TODO well... might works?
		// https://learn.microsoft.com/en-us/exchange/client-developer/legacy-protocols/how-to-authenticate-an-imap-pop-smtp-application-by-using-oauth
		numSMTPPort, _ := strconv.ParseInt(smtp_port, 10, 64)
		client = sasl.NewOAuthBearerClient(&sasl.OAuthBearerOptions{
			Username: smtp_username,
			Token:    smtp_password,
			Host:     smtp_host,
			Port:     int(numSMTPPort),
		})
	} else {
		client = sasl.NewAnonymousClient(uuid.New().String())
	}
	to := []string{_to}
	msg := strings.NewReader("To: " + _to + "\r\n" +
		"From: \"TbSign Push Service\" <" + mail + ">\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Subject: " + mime.QEncoding.Encode("UTF-8", title) + " \r\n" +
		"Date: " + Now.Format(time.RFC3339) + "\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n" +
		"Message-ID: <" + Now.Format("20060102150405") + "." + strconv.Itoa(Now.Nanosecond()) + "." + mail + ">\r\n" +
		"\r\n" +
		body + "<br /><br />" + Now.Format(time.DateOnly) + "\r\n")
	return smtp.SendMail(smtp_host+":"+smtp_port, client, mail, to, msg)
}

type PushMessageTemplateStruct struct {
	Title string
	Body  string
}

func PushMessageTemplateResetPassword(verifyMessage, code string) PushMessageTemplateStruct {
	return PushMessageTemplateStruct{
		Title: code + " 是你的验证码",
		Body: "你正在 TbSign 进行找回密码，需要进行身份验证。 本次行为的验证码是:<br /><br />" +
			code + "<br /><br />" +
			"请在页面输入验证码，进行重置。<br />" +
			"该消息" + strconv.Itoa(ResetPwdExpire/60) + "分钟内有效，为了你的账号安全，请勿将验证码提供给他人。<br /><br />" +
			"请确认以下 emoji 的排列顺序与网页展示的一致:<br />" + verifyMessage,
	}
}
func PushMessageTestTemplate() PushMessageTemplateStruct {
	return PushMessageTemplateStruct{
		Title: "TbSign 测试消息",
		Body:  "TbSign 推送服务测试<br />这是一条测试消息，如果您能阅读到这里，说明消息已经发送成功<br />emoji 测试:<br />" + RandomEmoji(),
	}
}
