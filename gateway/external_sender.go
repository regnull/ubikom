package gateway

import (
	"bytes"
	"os/exec"
)

// ExternalSender represents anything that can send an email to the outside world.
type ExternalSender interface {
	Send(from string, message string) error
}

// SendmailSender implements ExternalSender and sends mail using sendmail.
type SendmailSender struct {
}

// NewSendmailSender creates and returns a new SendmailSender.
func NewSendmailSender() *SendmailSender {
	return &SendmailSender{}
}

// Send sends email to the outside world.
func (s *SendmailSender) Send(from string, message string) error {
	cmd := exec.Command("sendmail", "-t", "-f", from)
	cmd.Stdin = bytes.NewReader([]byte(message))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
