package util

import (
	"github.com/smtp2go-oss/smtp2go-go"
)

func SendSMTPEmail(from, to, subject, textBody string) error {
	email := smtp2go.Email{
		// From: "Matt <matt@example.com>",
		// To: []string{
		// 	"Dave <dave@example.com>",
		// },
		// Subject:  "Trying out SMTP2GO",
		// TextBody: "Test Message",
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
