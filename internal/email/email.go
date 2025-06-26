package email

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/wneessen/go-mail"
)

type EmailDetails struct {
	From        string
	To          string
	Host        string
	Port        int
	Subject     string
	Body        string
	Attachments []string
	Username    string
	Password    string
}

// verify verfies the EmailDetails for necessary details
func (e *EmailDetails) verify() error {
	switch {
	case strings.TrimSpace(e.From) == "":
		return fmt.Errorf("From field should not be empty")
	case strings.TrimSpace(e.To) == "":
		return fmt.Errorf("To field should not be empty")
	case strings.TrimSpace(e.Host) == "":
		return fmt.Errorf("Host field should not be empty")
	case e.Port == 0:
		return fmt.Errorf("Port field should be non-zero")
	case strings.TrimSpace(e.Subject) == "":
		return fmt.Errorf("Subject field should not be empty")
	case strings.TrimSpace(e.Body) == "":
		return fmt.Errorf("Body field should not be empty")
	case strings.TrimSpace(e.Username) == "":
		return fmt.Errorf("Username field should not be empty")
	case strings.TrimSpace(e.Password) == "":
		return fmt.Errorf("Password field should not be empty")
	}

	return nil
}

func Send(details EmailDetails) error {
	// TODO: add task id context or direct parameter
	if err := details.verify(); err != nil {
		return fmt.Errorf("error while validating email details: %w", err)
	}

	msg := mail.NewMsg()
	err := msg.From(details.From)
	if err != nil {
		return nil
	}

	err = msg.To(details.To)
	if err != nil {
		return nil
	}

	msg.Subject(details.Subject)
	msg.SetBodyString(mail.TypeTextPlain, details.Body)
	if len(details.Attachments) > 0 {
		for _, item := range details.Attachments {
			msg.AttachFile(item)
		}
	}

	slog.Info("Creating email client object")
	client, err := mail.NewClient(details.Host,
		mail.WithPort(details.Port),
		mail.WithTLSPolicy(mail.DefaultTLSPolicy),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithUsername(details.Username),
		mail.WithPassword(details.Password),
		mail.WithTLSPolicy(mail.TLSMandatory),
	)
	if err != nil {
		return fmt.Errorf("error while constructing email client, %w", err)
	}

	slog.Info("Attempting to send email")

	// send the email
	err = client.DialAndSend(msg)
	if err != nil {
		return fmt.Errorf("error while sending email, %w", err)
	}
	slog.Info("Email sent successfully")

	return nil
}
