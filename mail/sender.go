package mail

import (
	"fmt"
	"net/smtp"
	"time"

	"github.com/jordan-wright/email"
)

const (
	smtpAddressHost = "smtp.gmail.com"
	smtpServerAddr  = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(subject string, content string, to, cc, bcc []string, attachFiles []string) error
}

type Sender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

type EmailBody struct {
	Username string
	URL      string
	Today    time.Time
}

func NewSender(name, fromEmailAddress, fromEmailPassword string) EmailSender {
	return &Sender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (s *Sender) SendEmail(subject string, content string, to, cc, bcc []string, attachFiles []string) error {
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", s.name, s.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.To = to
	e.Cc = cc
	e.Bcc = bcc
	for _, file := range attachFiles {
		_, err := e.AttachFile(file)
		if err != nil {
			return fmt.Errorf("unable to attach file: %w", err)
		}
	}

	smptAuth := smtp.PlainAuth("", s.fromEmailAddress, s.fromEmailPassword, smtpAddressHost)
	if err := e.Send(smtpServerAddr, smptAuth); err != nil {
		return fmt.Errorf("unable to send email: %w", err)
	}
	return nil
}
