package _function

import (
	"fmt"
	"log"
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
		_body + "\r\n")
	err := smtp.SendMail(smtp_host+":"+smtp_port, client, mail, to, msg)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
