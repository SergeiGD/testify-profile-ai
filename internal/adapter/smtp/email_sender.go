package smtp

import (
	"fmt"
	"net/smtp"
)

type EmailSender interface {
	SendConfirmationEmail(to, confirmationLink string) error
}

type smtpEmailSender struct {
	host     string
	port     int
	username string
	password string
	from     string
}

func NewEmailSender(host string, port int, username, password, from string) EmailSender {
	return &smtpEmailSender{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (s *smtpEmailSender) SendConfirmationEmail(to, confirmationLink string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	subject := "Подтверждение регистрации"
	body := fmt.Sprintf(
		"Для подтверждения регистрации перейдите по ссылке:\n%s\n\nСсылка действительна в течение 10 минут.",
		confirmationLink,
	)
	msg := []byte(
		"From: " + s.from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, []string{to}, msg)
}
