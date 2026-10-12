package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"adam-french.co.uk/backend/services"
	"github.com/gin-gonic/gin"
)

func envRouter(t *testing.T) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	env, err := services.NewEnvManager(dir)
	if err != nil {
		t.Fatalf("NewEnvManager: %v", err)
	}
	store := &Store{Env: env}
	r := gin.New()
	r.GET("/admin/env", store.GetEnv)
	r.PUT("/admin/env", store.UpdateEnv)
	r.GET("/admin/env/status", store.GetEnvStatus)
	r.DELETE("/admin/env/pending", store.CancelEnv)
	return r, dir
}

func envDo(r *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// The property the whole tab rests on: a secret's value never leaves the
// server, while a non-secret's does so it can be edited in place.
func TestGetEnvNeverReturnsSecrets(t *testing.T) {
	r, _ := envRouter(t)
	t.Setenv("CLAUDE_API_KEY", "sk-ant-very-secret-value-9f3a")
	t.Setenv("POSTGRES_PASSWORD", "hunter2")
	t.Setenv("STEAM_ID", "76561198000000000")
	t.Setenv("TFL_APP_KEY", "")

	rec := envDo(r, http.MethodGet, "/admin/env", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	for _, secret := range []string{"sk-ant-very-secret-value-9f3a", "hunter2"} {
		if strings.Contains(body, secret) {
			t.Fatalf("response contains the secret %q", secret)
		}
	}

	var parsed struct {
		Groups []struct {
			Vars []envVarView `json:"vars"`
		} `json:"groups"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("parsing response: %v", err)
	}
	byName := map[string]envVarView{}
	for _, g := range parsed.Groups {
		for _, v := range g.Vars {
			byName[v.Name] = v
		}
	}

	if v := byName["CLAUDE_API_KEY"]; !v.Set || !v.Secret || v.Value != "" || v.Hint != "9f3a" {
		t.Errorf("CLAUDE_API_KEY = %+v, want set, secret, no value, hint 9f3a", v)
	}
	// Too short for four characters to be a safe hint.
	if v := byName["POSTGRES_PASSWORD"]; !v.Set || v.Hint != "" || v.Editable {
		t.Errorf("POSTGRES_PASSWORD = %+v, want set, no hint, locked", v)
	}
	if v := byName["STEAM_ID"]; v.Value != "76561198000000000" || v.Secret {
		t.Errorf("STEAM_ID = %+v, want its value shown", v)
	}
	if v := byName["TFL_APP_KEY"]; v.Set {
		t.Errorf("TFL_APP_KEY = %+v, want unset", v)
	}
}

func TestUpdateEnv(t *testing.T) {
	r, dir := envRouter(t)
	requestPath := filepath.Join(dir, "request.json")

	t.Run("a locked variable is refused and nothing is queued", func(t *testing.T) {
		rec := envDo(r, http.MethodPut, "/admin/env", `{"set":{"POSTGRES_PASSWORD":"x"}}`)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
		if _, err := os.Stat(requestPath); err == nil {
			t.Fatal("a refused change left a request behind")
		}
	})

	t.Run("a malformed body is refused", func(t *testing.T) {
		if rec := envDo(r, http.MethodPut, "/admin/env", `{"set":`); rec.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", rec.Code)
		}
	})

	t.Run("a valid change is queued and a generated value returned once", func(t *testing.T) {
		rec := envDo(r, http.MethodPut, "/admin/env", `{"set":{"STEAM_ID":"42"},"generate":["BACKEND_SECRET"]}`)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("status = %d, want 202: %s", rec.Code, rec.Body)
		}
		var parsed struct {
			Generated map[string]string       `json:"generated"`
			Status    services.EnvApplyStatus `json:"status"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
			t.Fatalf("parsing response: %v", err)
		}
		secret := parsed.Generated["BACKEND_SECRET"]
		if len(secret) != 64 {
			t.Fatalf("generated BACKEND_SECRET = %q, want 64 hex characters", secret)
		}
		if parsed.Status.State != "pending" {
			t.Fatalf("status = %+v, want pending", parsed.Status)
		}

		// Reading the state back must not repeat it.
		for _, path := range []string{"/admin/env", "/admin/env/status"} {
			if strings.Contains(envDo(r, http.MethodGet, path, "").Body.String(), secret) {
				t.Fatalf("GET %s repeats the generated secret", path)
			}
		}
	})

	t.Run("a pending change can be withdrawn", func(t *testing.T) {
		rec := envDo(r, http.MethodDelete, "/admin/env/pending", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if _, err := os.Stat(requestPath); err == nil {
			t.Fatal("the request file survived the cancel")
		}
	})
}
