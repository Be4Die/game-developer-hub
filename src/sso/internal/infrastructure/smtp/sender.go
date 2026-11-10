// Package smtp реализует отправку служебных писем SSO через SMTP-сервер.
package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// Config — параметры подключения к SMTP-серверу.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	// TLS включает шифрование: неявный TLS на порту 465, STARTTLS на остальных.
	TLS bool
	// Timeout ограничивает время одной отправки.
	Timeout time.Duration
}

// Sender отправляет письма с кодом подтверждения и токеном сброса пароля.
type Sender struct {
	cfg Config
}

// NewSender создаёт отправщик. Возвращает ошибку при неполной конфигурации.
func NewSender(cfg Config) (*Sender, error) {
	if cfg.Host == "" || cfg.Port <= 0 {
		return nil, fmt.Errorf("smtp.NewSender: host and port are required")
	}
	if cfg.From == "" {
		return nil, fmt.Errorf("smtp.NewSender: from address is required")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 15 * time.Second
	}
	return &Sender{cfg: cfg}, nil
}

// SendVerificationEmail отправляет код подтверждения регистрации.
func (s *Sender) SendVerificationEmail(ctx context.Context, email, code string) error {
	body := "Здравствуйте!\r\n\r\n" +
		"Код подтверждения регистрации в Game Developer Hub: " + code + "\r\n\r\n" +
		"Если вы не регистрировались, просто проигнорируйте это письмо.\r\n"
	return s.send(ctx, email, "Подтверждение регистрации", body)
}

// SendPasswordResetEmail отправляет токен сброса пароля.
func (s *Sender) SendPasswordResetEmail(ctx context.Context, email, token string) error {
	body := "Здравствуйте!\r\n\r\n" +
		"Токен для сброса пароля в Game Developer Hub:\r\n\r\n" + token + "\r\n\r\n" +
		"Если вы не запрашивали сброс пароля, просто проигнорируйте это письмо.\r\n"
	return s.send(ctx, email, "Сброс пароля", body)
}

func (s *Sender) send(ctx context.Context, to, subject, body string) error {
	const op = "smtp.Sender.send"

	// Защита от внедрения заголовков через адрес получателя.
	if strings.ContainsAny(to, "\r\n") {
		return fmt.Errorf("%s: invalid recipient address", op)
	}

	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	deadline := time.Now().Add(s.cfg.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}

	dialer := &net.Dialer{Deadline: deadline}
	tlsCfg := &tls.Config{ServerName: s.cfg.Host, MinVersion: tls.VersionTLS12}

	var conn net.Conn
	var err error
	implicitTLS := s.cfg.TLS && s.cfg.Port == 465
	if implicitTLS {
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, tlsCfg)
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("%s: dial: %w", op, err)
	}
	_ = conn.SetDeadline(deadline)

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("%s: handshake: %w", op, err)
	}
	defer func() { _ = client.Close() }()

	if s.cfg.TLS && !implicitTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("%s: server does not support STARTTLS", op)
		}
		if err := client.StartTLS(tlsCfg); err != nil {
			return fmt.Errorf("%s: starttls: %w", op, err)
		}
	}

	if s.cfg.Username != "" {
		if err := client.Auth(smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)); err != nil {
			return fmt.Errorf("%s: auth: %w", op, err)
		}
	}

	if err := client.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("%s: mail from: %w", op, err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("%s: rcpt to: %w", op, err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("%s: data: %w", op, err)
	}
	if _, err := w.Write(buildMessage(s.cfg.From, to, subject, body)); err != nil {
		return fmt.Errorf("%s: write: %w", op, err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("%s: close data: %w", op, err)
	}

	return client.Quit()
}

// buildMessage собирает письмо в формате RFC 5322 с UTF-8 темой и текстом.
func buildMessage(from, to, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n")
	b.WriteString("Date: " + time.Now().Format(time.RFC1123Z) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
