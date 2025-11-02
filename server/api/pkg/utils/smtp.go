package utils

import (
	"github.com/smtp2go-oss/smtp2go-go"
)

func SendSMTPEmail(from, to, subject, textBody string) error {
	email := smtp2go.Email{
		From: from,
		To: []string{
			to,
		},
		Subject:  subject,
		TextBody: textBody,
	}
	if _, err := smtp2go.Send(&email); err != nil {
		return err
	}
	return nil
}
