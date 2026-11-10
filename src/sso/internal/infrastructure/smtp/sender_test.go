package smtp

import (
	"context"
	"strings"
	"testing"
)

func TestNewSender_Validation(t *testing.T) {
	t.Parallel()

	if _, err := NewSender(Config{Port: 25, From: "a@b.c"}); err == nil {
		t.Error("expected error for empty host")
	}
	if _, err := NewSender(Config{Host: "smtp.example.com", Port: 25}); err == nil {
		t.Error("expected error for empty from")
	}
	if _, err := NewSender(Config{Host: "smtp.example.com", Port: 587, From: "a@b.c"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildMessage(t *testing.T) {
	t.Parallel()

	msg := string(buildMessage("noreply@example.com", "user@example.com", "Сброс пароля", "тело\r\n"))
	for _, want := range []string{
		"From: noreply@example.com\r\n",
		"To: user@example.com\r\n",
		"Subject: =?utf-8?q?",
		"Content-Type: text/plain; charset=utf-8\r\n",
		"\r\n\r\nтело\r\n",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message does not contain %q:\n%s", want, msg)
		}
	}
}

func TestSend_RejectsHeaderInjection(t *testing.T) {
	t.Parallel()

	s, err := NewSender(Config{Host: "smtp.example.com", Port: 587, From: "a@b.c"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SendVerificationEmail(context.Background(), "x@y.z\r\nBcc: evil@example.com", "123456"); err == nil {
		t.Error("expected error for recipient with CRLF")
	}
}
