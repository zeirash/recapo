package email

import (
	"fmt"
	"net/smtp"
	"strings"

	mailjet "github.com/mailjet/mailjet-apiv3-go/v4"
	"github.com/zeirash/recapo/arion/common/config"
	"github.com/zeirash/recapo/arion/common/i18n"
	"github.com/zeirash/recapo/arion/common/logger"
)

// SendOTP sends a 6-digit OTP to the given email address.
func SendOTP(to, code, lang string) error {
	subject := i18n.T(lang, "email_otp_subject")
	body := fmt.Sprintf(i18n.T(lang, "email_otp_body"), code)
	return sendEmail(to, subject, body)
}

// SendPasswordResetOTP sends a password reset OTP to the given email address.
func SendPasswordResetOTP(to, code, lang string) error {
	subject := i18n.T(lang, "email_reset_otp_subject")
	body := fmt.Sprintf(i18n.T(lang, "email_reset_otp_body"), code)
	return sendEmail(to, subject, body)
}

func sendEmail(to, subject, body string) error {
	cfg := config.GetConfig()

	if cfg.MailjetAPIKeyPublic != "" {
		return sendViaMailjet(cfg, to, subject, body)
	}

	if cfg.SMTPHost != "" {
		return sendViaSMTP(cfg, to, subject, body)
	}

	logger.Infof("[DEV] Email to %s | Subject: %s | Body: %s", to, subject, body)
	return nil
}

func sendViaMailjet(cfg config.Config, to, subject, body string) error {
	client := mailjet.NewMailjetClient(cfg.MailjetAPIKeyPublic, cfg.MailjetAPIKeyPrivate)
	messages := mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{
			{
				From: &mailjet.RecipientV31{
					Email: cfg.MailjetFromEmail,
					Name:  cfg.MailjetFromName,
				},
				To: &mailjet.RecipientsV31{
					{Email: to},
				},
				Subject:  subject,
				TextPart: body,
			},
		},
	}
	_, err := client.SendMailV31(&messages)
	return err
}

func sendViaSMTP(cfg config.Config, to, subject, body string) error {
	from := cfg.SMTPUser
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	msg := strings.Join([]string{
		"From: " + from,
		"To: " + to,
		"Subject: " + subject,
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if cfg.SMTPUser != "" && cfg.SMTPPass != "" {
		auth = smtp.PlainAuth("", cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPHost)
	}

	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}
