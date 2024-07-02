package _function

import (
	"fmt"
	"mime"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-smtp"

	"github.com/emersion/go-sasl"
	"github.com/google/uuid"
)

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

// TODO GPG?
func SendEmail(_to, _subject, _body string) error {
	mail := GetOption("mail_name")
	mail_name := GetOption("mail_yourname")
	smtp_host := GetOption("mail_host")
	smtp_port := GetOption("mail_port")
	smtp_secure := GetOption("mail_secure")
	smtp_auth := GetOption("mail_auth") != "0"
	smtp_username := GetOption("mail_smtpname")
	smtp_password := GetOption("mail_smtppw")

	if mail == "" || mail_name == "" || smtp_host == "" || smtp_port == "" || smtp_secure == "" {
		return fmt.Errorf("mail: Email settings not completed")
	}

	if smtp_auth && (smtp_username == "" || smtp_password == "") {
		return fmt.Errorf("mail: Login failed")
	}

	var client sasl.Client
	if smtp_auth {
		client = sasl.NewLoginClient(smtp_username, smtp_password)
	} else {
		client = sasl.NewAnonymousClient(uuid.New().String())
	}
	to := []string{_to}
	msg := strings.NewReader("To: " + _to + "\r\n" +
		"From: \"TbSign Push Service\" <" + mail + ">\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n" +
		"Subject: " + mime.QEncoding.Encode("UTF-8", _subject) + " \r\n" +
		"Date: " + Now.Format(time.RFC3339) + "\r\n" +
		"Content-Transfer-Encoding: 8bit\r\n" +
		"Message-ID: <" + Now.Format("20060102150405") + "." + strconv.Itoa(Now.Nanosecond()) + "." + mail + ">\r\n" +
		"\r\n" +
		_body + "<br /><br />" + Now.Format("2006-01-02") + "\r\n")
	return smtp.SendMail(smtp_host+":"+smtp_port, client, mail, to, msg)
}

type emailTemplateStruct struct {
	Object string
	Body   string
}

func EmailTemplateResetPassword(email, code string) emailTemplateStruct {
	return emailTemplateStruct{
		Object: code + " 是你的验证码",
		Body: "亲爱的 " + email + "<br /><br />" +
			"你正在 TbSign 进行找回密码，需要进行身份验证。 本次行为的验证码是:<br /><br />" +
			code + "<br /><br />" +
			"请在页面输入验证码，进行重置。<br />" +
			"该邮件" + strconv.Itoa(ResetPwdExpire/60) + "分钟内有效，为了你的帐号安全，请勿将验证码提供给他人。",
	}
}
func EmailTestTemplate() emailTemplateStruct {
	return emailTemplateStruct{
		Object: "TbSign 测试邮件",
		Body:   "TbSign 推送服务测试<br />这是一封测试消息，如果您能阅读到这里，说明邮件已经发送成功",
	}
}
