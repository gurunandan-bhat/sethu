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

	emailTemplate, ok := s.Template[template]
	if !ok {
		logErr := fmt.Errorf("cannot find template %s in template cache", template)
		fmt.Println("Checking Cache: " + logErr.Error())
		return logErr
	}
	var emailBuf bytes.Buffer
	if err := emailTemplate.ExecuteTemplate(&emailBuf, "email", data); err != nil {
		logErr := fmt.Errorf("error generating email from template %s: %w", template, err)
		fmt.Println("Generating Template: " + logErr.Error())
		return logErr
	}
	fmt.Println("Email Text " + emailBuf.String())

	if err := smtp.SendEmail(validEmail.Address, emailBuf.String()); err != nil {
		logErr := fmt.Errorf("error sending email %w", err)
		fmt.Println("Sending : " + logErr.Error())
		return err
	}

	return nil
}
