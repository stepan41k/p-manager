package app

import (
	"fmt"
	"net/smtp"

	"github.com/stepan41k/p-manager/internal/lib/logger/sl"
	"github.com/zalando/go-keyring"
)

func (m *Model) sendOTPEmail(code string) error {
	smtpPassword, err := keyring.Get("p-manager", "smtp_password")
	if err != nil {
		m.log.Warn("failed to get password", sl.Err(err))
		return fmt.Errorf("smtp password not found in keyring: %w", err)
	}

	cfg := m.config.SMTPConfig

	host := cfg.SMTPHost
	port := cfg.SMTPPort
	from := cfg.SMTPSender
	to := cfg.Email

	subject := "Subject: Vault Access Code\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf("Your security code is: <b>%s</b>", code)
	msg := []byte(subject + mime + body)

	auth := smtp.PlainAuth("", from, smtpPassword, host)
	addr := fmt.Sprintf("%s:%s", host, port)

	err = smtp.SendMail(addr, auth, from, []string{to}, msg)
	if err != nil {
		m.log.Warn("failed to send mail:", sl.Err(err))
		return fmt.Errorf("failed to send mail: %w", err)
	}

	return nil
}
