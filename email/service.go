package email

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
		m.SetHeader(content.field, content.value...)
	}

	for _, content := range s.content.AddressHeader {
		m.SetAddressHeader(content.field, content.address, content.name)
	}

	body := s.content.Body
	m.SetBody(body.contentType, body.body)

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
	field string
	value []string
}

type SetAddressHeader struct {
	field   string
	address string
	name    string
}

type SetBody struct {
	contentType string
	body        string
}
