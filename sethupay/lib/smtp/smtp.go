package smtp

import (
	"crypto/tls"
	"fmt"
	"sethupay/lib/config"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(to, htmlBody string) error {

	cfg, err := config.Configuration()
	if err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", "Mario Gallery", cfg.SMTP.User))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Sethu Thanks You for your Generous Donation")
	m.SetBody("text/html", htmlBody)

	// Settings for SMTP server
	d := gomail.NewDialer(
		cfg.SMTP.Server,
		cfg.SMTP.Port,
		cfg.SMTP.User,
		cfg.SMTP.Password,
	)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
