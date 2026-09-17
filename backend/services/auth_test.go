package services

import (
	"errors"
	"testing"
	"time"

	"adam-french.co.uk/backend/models"
	"github.com/golang-jwt/jwt/v5"
)

// The tests below deliberately stop short of AuthenticateToken's happy path:
// that branch needs a live *gorm.DB and the only driver in go.mod is postgres,
// so exercising it would mean a database in CI. Everything that runs before the
// row lookup (parsing, signature, claim extraction) is covered here, and the
// revocation rule itself is covered through tokenVersionOf, which is the whole
// of the comparison AuthenticateTokenWithClaims performs.

func testAuth() *Auth {
	return InitAuth(&AuthConfig{
		Secret:               []byte("test-secret"),
		Domain:               "example.test",
		AccessTokenLifetime:  time.Hour,
		RefreshTokenLifetime: 24 * time.Hour,
	})
}

func TestTokenVersionOf(t *testing.T) {
	cases := []struct {
		name string
		in   uint
		want uint
	}{
		{"zero normalises to one so pre-versioning rows stay logged in", 0, 1},
		{"one is the column default", 1, 1},
		{"first bump", 2, 2},
		{"large value passes through", 99, 99},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := tokenVersionOf(c.in); got != c.want {
				t.Fatalf("tokenVersionOf(%d) = %d, want %d", c.in, got, c.want)
			}
		})
	}
}

// The comparison AuthenticateTokenWithClaims makes is tokenVersionOf(claim) vs
// tokenVersionOf(row). The table therefore encodes the revocation contract: an
// absent claim (0) must still match an un-bumped user, and any bump must not.
func TestTokenVersionMatch(t *testing.T) {
	cases := []struct {
		name      string
		claimed   uint
		stored    uint
		wantMatch bool
	}{
		{"legacy token with no tv claim against a pre-migration row", 0, 0, true},
		{"legacy token with no tv claim against a default row", 0, 1, true},
		{"legacy token after a bump is revoked", 0, 2, false},
		{"current token", 2, 2, true},
		{"token minted before the bump is revoked", 1, 2, false},
		{"stale row cannot revoke a newer token either", 3, 2, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tokenVersionOf(c.claimed) == tokenVersionOf(c.stored)
			if got != c.wantMatch {
				t.Fatalf("claimed %d vs stored %d: match = %v, want %v", c.claimed, c.stored, got, c.wantMatch)
			}
		})
	}
}

func TestGenerateJWTTokenVersionClaim(t *testing.T) {
	auth := testAuth()
	cases := []struct {
		name    string
		version uint
		wantTV  float64
	}{
		{"unset version is minted as 1", 0, 1},
		{"explicit default", 1, 1},
		{"after a bump", 5, 5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tokens, err := auth.GenerateJWT(&models.User{ID: 7, Username: "adam", Admin: true, TokenVersion: c.version})
			if err != nil {
				t.Fatalf("GenerateJWT: %v", err)
			}
			for label, tok := range map[string]string{"access": tokens.AccessToken, "refresh": tokens.RefreshToken} {
				claims, err := auth.VerifyJWT(tok)
				if err != nil {
					t.Fatalf("VerifyJWT(%s): %v", label, err)
				}
				if tv, _ := (*claims)["tv"].(float64); tv != c.wantTV {
					t.Errorf("%s token tv = %v, want %v", label, tv, c.wantTV)
				}
			}
		})
	}
}

// The split of claims is load-bearing: admin lives in the access token so most
// requests need no row, and is absent from the refresh token so that revoking
// admin takes effect on the next refresh.
func TestGenerateJWTClaimSplit(t *testing.T) {
	auth := testAuth()
	tokens, err := auth.GenerateJWT(&models.User{ID: 7, Username: "adam", Admin: true, TokenVersion: 1})
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	access, err := auth.VerifyJWT(tokens.AccessToken)
	if err != nil {
		t.Fatalf("VerifyJWT(access): %v", err)
	}
	if id, _ := (*access)["id"].(float64); id != 7 {
		t.Errorf("access id = %v, want 7", id)
	}
	if admin, _ := (*access)["admin"].(bool); !admin {
		t.Error("access token should carry admin")
	}
	if (*access)["username"] != "adam" {
		t.Errorf("access username = %v, want adam", (*access)["username"])
	}

	refresh, err := auth.VerifyJWT(tokens.RefreshToken)
	if err != nil {
		t.Fatalf("VerifyJWT(refresh): %v", err)
	}
	if _, ok := (*refresh)["admin"]; ok {
		t.Error("refresh token must not carry admin")
	}
	if _, ok := (*refresh)["username"]; ok {
		t.Error("refresh token must not carry username")
	}
}

func TestVerifyJWTRejections(t *testing.T) {
	auth := testAuth()
	secret := auth.Config.Secret

	signed := func(method jwt.SigningMethod, key any, claims jwt.MapClaims) string {
		s, err := jwt.NewWithClaims(method, claims).SignedString(key)
		if err != nil {
			t.Fatalf("signing: %v", err)
		}
		return s
	}

	expired := signed(jwt.SigningMethodHS256, secret, jwt.MapClaims{
		"id":  float64(1),
		"exp": time.Now().Add(-time.Minute).Unix(),
	})
	wrongSecret := signed(jwt.SigningMethodHS256, []byte("other-secret"), jwt.MapClaims{
		"id":  float64(1),
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	// An unsigned token is the classic algorithm-confusion attack; VerifyJWT
	// pins HS256, so it must not verify even though the "signature" matches.
	none := signed(jwt.SigningMethodNone, jwt.UnsafeAllowNoneSignatureType, jwt.MapClaims{
		"id":  float64(1),
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	cases := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"garbage", "not.a.token"},
		{"expired", expired},
		{"signed with another secret", wrongSecret},
		{"alg none", none},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := auth.VerifyJWT(c.token); err == nil {
				t.Fatal("expected VerifyJWT to reject the token")
			}
			// Every VerifyJWT failure must surface as ErrInvalidToken, since
			// call sites map that sentinel onto a 401.
			if _, err := auth.AuthenticateToken(nil, c.token); !errors.Is(err, ErrInvalidToken) {
				t.Fatalf("AuthenticateToken err = %v, want ErrInvalidToken", err)
			}
		})
	}
}

// A verified token with no usable id is a server-side inconsistency, not a
// client error, so it gets its own sentinel. Both cases below return before the
// row lookup, which is why a nil *gorm.DB is safe here.
func TestAuthenticateTokenInvalidClaims(t *testing.T) {
	auth := testAuth()
	cases := []struct {
		name   string
		claims jwt.MapClaims
	}{
		{"no id claim", jwt.MapClaims{"tv": 1, "exp": time.Now().Add(time.Hour).Unix()}},
		{"id is not a number", jwt.MapClaims{"id": "7", "exp": time.Now().Add(time.Hour).Unix()}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c.claims).SignedString(auth.Config.Secret)
			if err != nil {
				t.Fatalf("signing: %v", err)
			}
			if _, err := auth.AuthenticateToken(nil, tok); !errors.Is(err, ErrInvalidClaims) {
				t.Fatalf("err = %v, want ErrInvalidClaims", err)
			}
		})
	}
}
