package main

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"gopkg.in/gomail.v2"
)

// DON'T REMOVE THE UNDERLAYING COMMENT

//go:embed templates/*.html
var templateFS	embed.FS
var tmpls		*template.Template

// May need hardening when sending emails, could be abuse to launch 10x

//This works fine when sending 20 mails/min...
func Send_Mail(s *Settings, m *gomail.Message) error {
	//587 is hardcoded for the SMPT protocol
	d := gomail.NewDialer(s.Mail.Provider, 587, s.Mail.User, s.Mail_key)
	err := d.DialAndSend(m)
	if err != nil {
		slog.Error("email send failed", "err", err)
		return err
	}
	return nil
}

func Mail_Reset_Pass(s *Settings, db *Db_data, target string) error {
	var err		error

	m := gomail.NewMessage()
	m.SetHeader("From", s.Mail.User)
	m.SetHeader("To", target)
	m.SetHeader("Subject", "Cambiar contraseña")
	id, err := create_a_password_reset(db, target)
	if err != nil {
		return err
	}
	str, err := resetPasswordHTML(s.Frontend + "/reset_pass_new/" + id)
	if err != nil {
		return err
	}
	m.SetBody("text/html", str)
	err = Send_Mail(s, m)
	if err != nil {
		return err
	}
	return nil
}

func resetPasswordHTML(link string) (string, error) {
	var buf		bytes.Buffer
	var err		error

	err = tmpls.ExecuteTemplate(&buf, "reset_pass.html", struct{ Link string }{ link })
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func TwoFA_Mail(s *Settings, db *Db_data, target string, id string) error {
	var err		error

	m := gomail.NewMessage()
	m.SetHeader("From", s.Mail.User)
	m.SetHeader("To", target)
	m.SetHeader("Subject", "Doble factor de autentificación")
	str, err := TwoFAHTML(s.Frontend + "/2FA_validate/" + id)
	if err != nil {
		return err
	}
	m.SetBody("text/html", str)
	err = Send_Mail(s, m)
	if err != nil {
		return err
	}
	return nil
}

func TwoFAHTML(link string) (string, error) {
	var buf		bytes.Buffer
	var err		error

	err = tmpls.ExecuteTemplate(&buf, "2FA_validate.html", struct{ Link string }{ link })
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
