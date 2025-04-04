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
	_, err := smtp2go.Send(&email)
	if err != nil {
		return err
	}
	return nil
}
