package main

import (
	"log"
	"gopkg.in/gomail.v2"
)

//This works fine when sending 20 mails/min...
func Send_Mail(s *Settings, m *gomail.Message) error {
	//587 is hardcoded for the SMPT protocol
	d := gomail.NewDialer(s.Mail.Provider, 587, s.Mail.User, s.Mail_key)
	err := d.DialAndSend(m)
	if err != nil {
		log.Printf("failed to send email to %v: %v", m.GetHeader("To"), err)
		return err
	}
	return nil
}

func Mail_Reset_Pass(s *Settings, target string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.Mail.User)
	m.SetHeader("To", target)
	m.SetHeader("Subject", "Reset your password")
	m.SetBody("text/html", resetPasswordHTML()) //HERE YOU NEED TO DO SOME WORK FOR THE PROPER LINK CREATION/INVALIDATION
	err := Send_Mail(s, m)
	if err != nil {
		return err
	}
	return nil
}

func resetPasswordHTML(link string) string {
	return (`<html>
		<body style="font-family:sans-serif;max-width:600px;margin:0 auto;padding:20px">
			<h2>Reset your password</h2>
			<p>Click the button below to reset your password.</p>
			<a href="` + link + `" style="display:inline-block;padding:12px 24px;background:#000;color:#fff;text-decoration:none;border-radius:6px">Reset Password</a>
			<p style="color:#999;font-size:12px;margin-top:20px">If you didn't request this, ignore this email.</p>
		</body>
	</html>`)
}
