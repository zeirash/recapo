package email

import (
	"fmt"
	"net/smtp"
	"strings"

	resend "github.com/resend/resend-go/v2"
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

// SendInvitation sends an admin invitation email with the accept link.
func SendInvitation(to, inviterName, shopName, inviteURL, lang string) error {
	subject := fmt.Sprintf(i18n.T(lang, "email_invitation_subject"), shopName)
	body := fmt.Sprintf(i18n.T(lang, "email_invitation_body"), inviterName, shopName, inviteURL)
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

	if cfg.ResendAPIKey != "" {
		return sendViaResend(cfg, to, subject, body)
	}

	if cfg.SMTPHost != "" {
		return sendViaSMTP(cfg, to, subject, body)
	}

	logger.Infof("[DEV] Email to %s | Subject: %s | Body: %s", to, subject, body)
	return nil
}

func sendViaResend(cfg config.Config, to, subject, body string) error {
	client := resend.NewClient(cfg.ResendAPIKey)
	params := &resend.SendEmailRequest{
		From:    fmt.Sprintf("%s <%s>", cfg.ResendFromName, cfg.ResendFromEmail),
		To:      []string{to},
		Subject: subject,
		Text:    body,
	}
	_, err := client.Emails.Send(params)
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
