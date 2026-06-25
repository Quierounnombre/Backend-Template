package main

import (
	"log/slog"

	"gopkg.in/gomail.v2"
)

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
	m.SetBody("text/html", resetPasswordHTML(s.Frontend + "/reset_pass_new/" + id))
	err = Send_Mail(s, m)
	if err != nil {
		return err
	}
	return nil
}

func resetPasswordHTML(link string) string {
	return (`<html>
		<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
			<h2>Cambiar contraseña</h2>
			<p>Pulsa el boton para cambiar tu contraseña.</p>
			<a href="` + link + `" style="display:inline-block;padding:12px 24px;background:#000;color:#fff;text-decoration:none;border-radius:6px">Cambiar contraseña</a>
			<p style="color:#999;font-size:12px;margin-top:20px">En caso de no haberlo solicitado, ponte en contacto con nosotros.</p>
		</body>
	</html>`)
}

func TwoFA_Mail(s *Settings, db *Db_data, target string, id string) error {
	var err		error

	m := gomail.NewMessage()
	m.SetHeader("From", s.Mail.User)
	m.SetHeader("To", target)
	m.SetHeader("Subject", "Doble factor de autentificación")
	m.SetBody("text/html", TwoFAHTML(s.Frontend + "/2FA_validate/" + id))
	err = Send_Mail(s, m)
	if err != nil {
		return err
	}
	return nil
}

func TwoFAHTML(link string) string {
	return (`<html>
		<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
			<h2>Doble factor de autentificación</h2>
			<p>Haz click en el boton para terminar de registrarte.</p>
			<a href="` + link + `" style="display:inline-block;padding:12px 24px;background:#000;color:#fff;text-decoration:none;border-radius:6px">Registrarme</a>
			<p style="color:#999;font-size:12px;margin-top:20px">En caso de no haberlo solicitado, ponte en contacto con nosotros.</p>
		</body>
	</html>`)
}
