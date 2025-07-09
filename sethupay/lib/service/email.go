package service

import (
	"bytes"
	"fmt"
	"net/mail"
	"sethupay/lib/smtp"
)

func (s *Service) sendEmail(recipient, template string, data any) error {

	validEmail, err := mail.ParseAddress(recipient)
	if err != nil {
		return err
	}

	s.Logger.Info("Template", "name", template)
	emailTemplate, ok := s.Template[template]
	if !ok {
		logErr := fmt.Errorf("cannot find template %s in template cache", template)
		return logErr
	}
	var emailBuf bytes.Buffer
	if err := emailTemplate.ExecuteTemplate(&emailBuf, "base", data); err != nil {
		logErr := fmt.Errorf("error generating email from template %s: %w", template, err)
		return logErr
	}

	s.Logger.Info("Email string", "email", emailBuf.String())
	if err := smtp.SendEmail(validEmail.Address, emailBuf.String()); err != nil {
		logErr := fmt.Errorf("error sending email %w", err)
		return logErr
	}

	return nil
}
