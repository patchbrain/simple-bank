package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

const (
	smtpAuthAddress   = "smtp.163.com"
	smtpServerAddress = "smtp.163.com:25"
)

type EmailSender interface {
	SendEmail(
		subject string,
		to []string,
		content string,
		cc []string,
		bcc []string,
		attachFiles []string,
	) error
}

type WangYiEmailSender struct {
	FromEmailAddress  string
	FromEmailPassword string
	Name              string
}

func NewWangYiEmailSender(address, password, name string) EmailSender {
	return &WangYiEmailSender{
		FromEmailAddress:  address,
		FromEmailPassword: password,
		Name:              name,
	}
}

func (w *WangYiEmailSender) SendEmail(
	subject string, to []string, content string, cc []string, bcc []string, attachFiles []string) error {
	email := email.NewEmail()
	email.From = fmt.Sprintf("%s <%s>", w.Name, w.FromEmailAddress)
	email.To = to
	email.HTML = []byte(content)
	email.Cc = cc
	email.Bcc = bcc
	email.Subject = subject
	for _, file := range attachFiles {
		_, err := email.AttachFile(file)
		if err != nil {
			return fmt.Errorf("fail to attach file: %w", err)
		}
	}

	auth := smtp.PlainAuth("", w.FromEmailAddress, w.FromEmailPassword, smtpAuthAddress)
	err := email.Send(smtpServerAddress, auth)
	if err != nil {
		return fmt.Errorf("fail to send email: %w", err)
	}

	return nil
}
