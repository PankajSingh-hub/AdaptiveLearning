package service

import (
	"regexp"
	"testing"

	"adaptive-assessment/internal/mailer"
	"adaptive-assessment/internal/models"
)

func newTestSignup() (*SignupService, *AuthService) {
	auth := NewAuthService(nil, "secret", 1)
	return NewSignupService(nil, mailer.ConsoleMailer{}, "secret", auth), auth
}

func TestSignupTokenRoundTrip(t *testing.T) {
	s, _ := newTestSignup()
	tok, err := s.signupToken("a@b.com")
	if err != nil {
		t.Fatal(err)
	}
	if email, err := s.parseSignupToken(tok); err != nil || email != "a@b.com" {
		t.Fatalf("round trip failed: %v %v", email, err)
	}
}

func TestLoginAndSignupTokensAreNotInterchangeable(t *testing.T) {
	s, auth := newTestSignup()
	login, _ := auth.generateToken(models.User{ID: 1, Role: models.RoleAdmin})
	if _, err := s.parseSignupToken(login); err == nil {
		t.Fatal("a login token must not work as a signup token")
	}
	signup, _ := s.signupToken("a@b.com")
	if _, err := auth.ParseToken(signup); err == nil {
		t.Fatal("a signup token must not work as a login token")
	}
}

func TestCodeIsSixDigits(t *testing.T) {
	re := regexp.MustCompile(`^\d{6}$`)
	for i := 0; i < 100; i++ {
		c, err := newCode()
		if err != nil || !re.MatchString(c) {
			t.Fatalf("bad code %q (%v)", c, err)
		}
	}
}

func TestCodeHashIsBoundToTheEmail(t *testing.T) {
	s, _ := newTestSignup()
	if s.hashCode("a@b.com", "123456") == s.hashCode("c@d.com", "123456") {
		t.Fatal("the same code must hash differently for different emails")
	}
}

func TestPasswordRules(t *testing.T) {
	for pw, ok := range map[string]bool{
		"abc12345": true, "short1": false, "onlyletters": false, "12345678": false,
	} {
		if (validatePassword(pw) == nil) != ok {
			t.Errorf("validatePassword(%q) expected ok=%v", pw, ok)
		}
	}
}