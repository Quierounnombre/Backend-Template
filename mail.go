package main

import (
	"bytes"
	"embed"
	"html/template"
	"log/slog"
	"time"

	"gopkg.in/gomail.v2"
)

// DON'T REMOVE THE UNDERLAYING COMMENT

//go:embed templates/*.html
var templateFS	embed.FS
var tmpls		*template.Template

// May need hardening when sending emails, could be abuse to launch 10x
func init_mail(s *Settings) {
	//587 is hardcoded for the SMPT protocol
	s.Mail.dialer = gomail.NewDialer(s.Mail.Provider, 587, s.Mail.User, s.Mail_key)
	s.Mail.queue = make(chan *gomail.Message, s.Mail.queue_size)
	s.Mail.retry_queue = make(chan *gomail.Message, s.Mail.queue_size)
	go s.Mail.Manager()
}

func (mr *Mail_settings)Enqueue(m *gomail.Message) {
	mr.queue <- m
}

func (mr *Mail_settings)Retry_Enqueue(m *gomail.Message) {
	mr.retry_queue <- m
}

func (mr *Mail_settings)run(stop chan struct{}) {
	var s		gomail.SendCloser
	var err		error

	for {
		select {
		case m := <-mr.queue:
			if s == nil {
				s, err = mr.dialer.Dial()
				if err != nil {
					slog.Error("smtp failed to dial", "err", err)
					time.AfterFunc(5*time.Second, func() { mr.Enqueue(m) })
					continue
				}
			}
			err := gomail.Send(s, m)
			if err != nil {
				time.AfterFunc(5*time.Second, func() { mr.Enqueue(m) })
				slog.Error("send failed", "err", err)
				s.Close()
				s = nil
			}
		case m := <-mr.retry_queue:
			if s == nil {
				s, err = mr.dialer.Dial()
				if err != nil {
					slog.Error("smtp failed to dial", "err", err)
					slog.Error("dropped email", "to", m.GetHeader("To"))
					continue
				}
			}
			err := gomail.Send(s, m)
			if err != nil {
				slog.Error("send failed", "err", err)
				slog.Error("dropped email", "to", m.GetHeader("To"))
				s.Close()
				s = nil
			}
		case <-stop:
			if s != nil {
				s.Close()
			}
			return
		}
	}
}

func (mr *Mail_settings)Manager() {
	var stop		chan struct{}
	var n_workers	int
	var ex_wk		int

	stop = make(chan struct{})
	n_workers = 0
	ex_wk = 0
	for {
		queue_size := len(mr.queue)
		ex_wk = (queue_size + mr.worker_per_qeueu - 1) / mr.worker_per_qeueu
		if n_workers <= ex_wk {
			if n_workers <= mr.max_workers {
				slog.Info("New email worker created", "Email backlog", queue_size, "ex_wk", ex_wk)
				go mr.run(stop)
				n_workers += 1
			}
		} else if n_workers > mr.min_workers {
			slog.Info("Email worker deleted", "Email backlog", queue_size, "ex_wk", ex_wk)
			stop <- struct{}{}
			n_workers -= 1
		}
		time.Sleep(mr.sleep_time)
	}
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
	str, err := resetPasswordHTML(s.Frontend + "/Reset_pass_new/" + id)
	if err != nil {
		return err
	}
	m.SetBody("text/html", str)
	s.Mail.Enqueue(m)
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
	s.Mail.Enqueue(m)
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
