package smtp

import "github.com/sirupsen/logrus"

type mockEmailSender struct {
	log logrus.FieldLogger
}

func NewMockEmailSender(log logrus.FieldLogger) EmailSender {
	return &mockEmailSender{log: log}
}

func (m *mockEmailSender) SendConfirmationEmail(to, confirmationLink string) error {
	m.log.WithFields(logrus.Fields{
		"to":                to,
		"confirmation_link": confirmationLink,
	}).Info("mock: confirmation email")
	return nil
}
