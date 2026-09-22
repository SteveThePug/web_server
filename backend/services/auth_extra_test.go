package services

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/gin-gonic/gin"
)

// Companion to auth_test.go, covering the two pure things it leaves out: the
// expiry the minting code bakes into each token, and the cookie attributes
// SetAuthCookies/ClearAuthCookies emit.
//
// DELIBERATELY NOT COVERED HERE:
//   - AuthenticateToken past the id claim, and BumpTokenVersion: both need a
//     live *gorm.DB and the only driver in go.mod is postgres.
//   - Everything auth_test.go already covers (tokenVersionOf, the claim split,
//     the VerifyJWT rejection table) — not duplicated.
//   - Password hashing: there is none in this package; handlers own it.
//
// Cookie assertions go through a real gin context because gin, not this
// package, is what turns the SetCookie arguments into a header — asserting on
// the arguments alone would test nothing.

// cookiesFrom runs fn against a throwaway gin context and returns what it set,
// keyed by cookie name.
func cookiesFrom(fn func(*gin.Context)) map[string]*http.Cookie {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(rec)
	fn(gc)

	out := map[string]*http.Cookie{}
	for _, c := range rec.Result().Cookies() {
		out[c.Name] = c
	}
	return out
}

// exp must track the configured lifetime, because the cookie max-age is derived
// from the same number: if they disagreed the browser would hold a dead cookie
// or drop a live one.
func TestGenerateJWTExpiryMatchesConfiguredLifetime(t *testing.T) {
	auth := testAuth()
	issued := time.Now()

	tokens, err := auth.GenerateJWT(&models.User{ID: 7, Username: "adam", TokenVersion: 1})
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	cases := []struct {
		name  string
		token string
		want  time.Duration
	}{
		{"access token expires after AccessTokenLifetime", tokens.AccessToken, auth.Config.AccessTokenLifetime},
		{"refresh token expires after RefreshTokenLifetime", tokens.RefreshToken, auth.Config.RefreshTokenLifetime},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			claims, err := auth.VerifyJWT(c.token)
			if err != nil {
				t.Fatalf("VerifyJWT: %v", err)
			}
			exp, ok := (*claims)["exp"].(float64)
			if !ok {
				t.Fatal("token carries no numeric exp claim")
			}
			// A couple of seconds of slack: exp is whole seconds and the two
			// tokens are minted a hair apart.
			got := time.Unix(int64(exp), 0).Sub(issued)
			if got < c.want-2*time.Second || got > c.want+2*time.Second {
				t.Fatalf("exp is %v after minting, want ~%v", got, c.want)
			}
		})
	}
}

func TestSetAuthCookies(t *testing.T) {
	config := testAuth().Config
	tokens := &Tokens{AccessToken: "access-value", RefreshToken: "refresh-value"}
	got := cookiesFrom(func(gc *gin.Context) { SetAuthCookies(gc, config, tokens) })

	cases := []struct {
		name       string
		cookie     string
		wantValue  string
		wantMaxAge int
	}{
		{"access cookie carries the access token and its lifetime", "access_token", "access-value", int(config.AccessTokenLifetime.Seconds())},
		{"refresh cookie carries the refresh token and its lifetime", "refresh_token", "refresh-value", int(config.RefreshTokenLifetime.Seconds())},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ck := got[c.cookie]
			if ck == nil {
				t.Fatalf("no %s cookie was set", c.cookie)
			}
			if ck.Value != c.wantValue {
				t.Errorf("value = %q, want %q", ck.Value, c.wantValue)
			}
			if ck.MaxAge != c.wantMaxAge {
				t.Errorf("max-age = %d, want %d (the token's own lifetime)", ck.MaxAge, c.wantMaxAge)
			}
			// These four are the security contract: unreadable from JS, never
			// sent in clear, and not attached to cross-site POSTs.
			if !ck.HttpOnly {
				t.Error("cookie must be HttpOnly")
			}
			if !ck.Secure {
				t.Error("cookie must be Secure")
			}
			if ck.SameSite != http.SameSiteLaxMode {
				t.Errorf("SameSite = %v, want Lax", ck.SameSite)
			}
			if ck.Path != "/" || ck.Domain != config.Domain {
				t.Errorf("path/domain = %q/%q, want \"/\"/%q", ck.Path, ck.Domain, config.Domain)
			}
		})
	}
}

// A browser only replaces a cookie when path, domain and flags match the
// original, so a deletion that differs in any of them silently leaves the user
// logged in. Hence asserting the attributes, not just the negative max-age.
func TestClearAuthCookiesMatchesTheCookiesItDeletes(t *testing.T) {
	config := testAuth().Config
	set := cookiesFrom(func(gc *gin.Context) {
		SetAuthCookies(gc, config, &Tokens{AccessToken: "a", RefreshToken: "r"})
	})
	cleared := cookiesFrom(func(gc *gin.Context) { ClearAuthCookies(gc, config) })

	for _, name := range []string{"access_token", "refresh_token"} {
		t.Run(name+" is expired with attributes matching the original", func(t *testing.T) {
			ck, orig := cleared[name], set[name]
			if ck == nil || orig == nil {
				t.Fatalf("missing %s cookie", name)
			}
			if ck.Value != "" {
				t.Errorf("value = %q, want empty", ck.Value)
			}
			if ck.MaxAge >= 0 {
				t.Errorf("max-age = %d, want negative so the browser deletes it", ck.MaxAge)
			}
			if ck.Path != orig.Path || ck.Domain != orig.Domain ||
				ck.Secure != orig.Secure || ck.HttpOnly != orig.HttpOnly || ck.SameSite != orig.SameSite {
				t.Errorf("attributes differ from the cookie being deleted:\n cleared %+v\n set     %+v", ck, orig)
			}
		})
	}
}
