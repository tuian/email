package sender

import (
	"crypto/tls"
	"net/smtp"

	"../builder"
)

// SendMail 用来发送邮件
func SendMail(mailer *builder.MailBuilder, raw []byte,
	smtpserver string, xtls bool, auth smtp.Auth) error {
	c, err := smtp.Dial(smtpserver)
	if err != nil {
		return err
	}

	if xtls {
		config := tls.Config{InsecureSkipVerify: true}
		if err = c.StartTLS(&config); err != nil {
			return err
		}
	}

	// 用户信息认证
	if err = c.Auth(auth); err != nil {
		return err
	}

	// From
	if err = c.Mail(mailer.From.Address); err != nil {
		return err
	}

	// TO
	for _, item := range mailer.To {
		if err = c.Rcpt(item.Address); err != nil {
			return err
		}
	}

	// Cc
	for _, item := range mailer.Cc {
		if err = c.Rcpt(item.Address); err != nil {
			return err
		}
	}

	// TODO(user) Bcc

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(raw)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()

	return nil
}
