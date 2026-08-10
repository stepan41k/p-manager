package email

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(code string) error {
		from := os.Getenv("SMTP_FROM")
		pass := os.Getenv("SMTP_PASS")
		to := os.Getenv("TARGET_EMAIL")
		host := os.Getenv("SMTP_HOST")
		port := os.Getenv("SMTP_PORT")

		subject := "Subject: Vault Access Code\n"
		body := fmt.Sprintf("Your one-time access code is: %s\nThis code will expire shortly.", code)
		msg := []byte(subject + "\n" + body)

		auth := smtp.PlainAuth("", from, pass, host)
		return smtp.SendMail(host+":"+port, auth, from, []string{to}, msg)
}