package email

// MockEmailSender is a mock implementation of EmailSenderInterface for testing
// It stores sent emails in a slice for inspection

type MockSender struct {
	SentEmails []SentEmail
}

type SentEmail struct {
	Recipient string
	Subject   string
	Body      string
}

func (m *MockSender) Send(recipient, subject, htmlContent string) error {
	m.SentEmails = append(m.SentEmails, SentEmail{
		Recipient: recipient,
		Subject:   subject,
		Body:      htmlContent,
	})
	return nil
}
