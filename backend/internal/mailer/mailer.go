package mailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type Mailer interface {
	Send(to, subject, body string) error
	Mode() string // "smtp" or "console"
}

// ConsoleMailer is the dev fallback: it prints the message instead of sending it.
type ConsoleMailer struct{}

func (ConsoleMailer) Mode() string { return "console" }
func (ConsoleMailer) Send(to, subject, body string) error {
	log.Printf("EMAIL NOT SENT (SMTP not configured, console mode)\n  to: %s\n  subject: %s\n  ----\n%s\n  ----", to, subject, body)
	return nil
}

type SMTPMailer struct {
	host, port, user, pass string
	from                   *mail.Address
}

func (*SMTPMailer) Mode() string { return "smtp" }

// FromEnv reads SMTP_HOST, SMTP_PORT (default 587), SMTP_USER, SMTP_PASSWORD, SMTP_FROM.
func FromEnv() (Mailer, error) {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		log.Println("mail: SMTP_HOST not set; verification codes will be printed to this console")
		return ConsoleMailer{}, nil
	}
	user := os.Getenv("SMTP_USER")
	fromRaw := os.Getenv("SMTP_FROM")
	if fromRaw == "" {
		fromRaw = user
	}
	from, err := mail.ParseAddress(fromRaw)
	if err != nil {
		return nil, fmt.Errorf("SMTP_FROM is invalid: %w", err)
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	log.Printf("mail: SMTP configured (%s:%s)", host, port)
	return &SMTPMailer{host: host, port: port, user: user, pass: os.Getenv("SMTP_PASSWORD"), from: from}, nil
}

func (m *SMTPMailer) Send(to, subject, body string) error {
	if strings.ContainsAny(to+subject, "\r\n") {
		return errors.New("invalid characters in email header")
	}
	addr := net.JoinHostPort(m.host, m.port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	var conn net.Conn
	var err error
	if m.port == "465" { // implicit TLS
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.host})
	} else {
		conn, err = dialer.Dial("tcp", addr)
	}
	if err != nil {
		return err
	}
	_ = conn.SetDeadline(time.Now().Add(25 * time.Second))
	c, err := smtp.NewClient(conn, m.host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()
	if m.port != "465" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: m.host}); err != nil {
				return err
			}
		}
	}
	if m.user != "" { // PlainAuth refuses to send credentials over an unencrypted connection
		if err := c.Auth(smtp.PlainAuth("", m.user, m.pass, m.host)); err != nil {
			return err
		}
	}
	if err := c.Mail(m.from.Address); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	msg := "From: " + m.from.String() + "\r\nTo: " + to + "\r\nSubject: " + subject +
		"\r\nDate: " + time.Now().Format(time.RFC1123Z) +
		"\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" +
		strings.ReplaceAll(body, "\n", "\r\n")
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}