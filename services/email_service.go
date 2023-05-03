package services

import (
	"github.com/go-mail/mail"
	"khoomi-api-io/khoomi_api/configs"
)

type KhoomiEmailService struct {
	mailer  *mail.Message
	content KhoomiEmailComposer
}

func NewKhoomiEmailService(content KhoomiEmailComposer) KhoomiEmailService {
	return KhoomiEmailService{
		mailer:  mail.NewMessage(),
		content: content,
	}
}

func (s *KhoomiEmailService) SendMail() error {
	m := s.mailer
	for _, content := range s.content.Header {
		m.SetHeader(content.Field, content.Value...)
	}

	for _, content := range s.content.AddressHeader {
		m.SetAddressHeader(content.Field, content.Address, content.Name)
	}

	body := s.content.Body
	m.SetBody(body.ContentType, body.Body)

	SmtpHost := configs.LoadEnvFor("SMTP_HOST")
	SmtpUsername := configs.LoadEnvFor("SMTP_USERNAME")
	SmtpPassword := configs.LoadEnvFor("SMTP_PASSWORD")
	dialer := mail.NewDialer(SmtpHost, 2525, SmtpUsername, SmtpPassword)
	if err := dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

type KhoomiEmailComposer struct {
	Header        []SetHeader
	AddressHeader []SetAddressHeader
	Body          SetBody
	Attach        string
}

type SetHeader struct {
	Field string
	Value []string
}

type SetAddressHeader struct {
	Field   string
	Address string
	Name    string
}

type SetBody struct {
	ContentType string
	Body        string
}
