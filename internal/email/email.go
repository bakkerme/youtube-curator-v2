package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"

	"embed"
	"youtube-curator-v2/internal/rss"
)

//go:embed templates/*.tmpl
var templateFS embed.FS

// EmailSender handles sending emails
type EmailSender struct {
	SMTPServer   string
	SMTPPort     string
	SMTPUsername string
	SMTPPassword string
}

// NewEmailSender creates a new EmailSender instance
func NewEmailSender(server, port, username, password string) *EmailSender {
	return &EmailSender{
		SMTPServer:   server,
		SMTPPort:     port,
		SMTPUsername: username,
		SMTPPassword: password,
	}
}

// SendEmail sends an email with the given subject, body, and recipient
// Send sends an HTML email to the specified recipient
func (c *EmailSender) Send(recipient string, subject string, htmlContent string) error {
	if recipient == "" {
		return errors.New("recipient email cannot be empty")
	}

	// Validate recipient contains @
	if !strings.Contains(recipient, "@") {
		return errors.New("invalid recipient email format")
	}

	// Set up authentication
	auth := smtp.PlainAuth("", c.SMTPUsername, c.SMTPPassword, c.SMTPServer)

	// Construct MIME headers
	headers := make(map[string]string)
	headers["From"] = c.SMTPUsername
	headers["To"] = recipient
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build message from headers
	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + htmlContent)

	// Send email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", c.SMTPServer, c.SMTPPort),
		auth,
		c.SMTPUsername,
		[]string{recipient},
		[]byte(message.String()),
	)

	return err
}

// FormatNewVideosEmail formats an email for new video notifications
func FormatNewVideosEmail(videos []rss.Entry) (string, error) {
	tmplContent, err := templateFS.ReadFile("templates/videos_email_template.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to read template: %w", err)
	}

	// Create and parse the template
	funcMap := template.FuncMap{
		"cleanHTML": func(s string) string {
			// This is a very basic way to remove tags; consider a library for production.
			// For MVP, this should be okay.
			return rss.CleanContent(s, 300, false) // Using existing CleanContent
		},
	}

	t, err := template.New("newVideosEmail").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, videos); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return body.String(), nil
}
