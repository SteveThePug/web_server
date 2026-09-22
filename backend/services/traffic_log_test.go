package services

import (
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func fmtParams(method, path string, status int, ip string, d time.Duration) gin.LogFormatterParams {
	return gin.LogFormatterParams{
		Method:     method,
		Path:       path,
		StatusCode: status,
		ClientIP:   ip,
		Latency:    d,
	}
}

func TestTrafficFormatterDropsQuietHeartbeats(t *testing.T) {
	out := TrafficLogFormatter(fmtParams("GET", "/auth/validate-admin", 200, "172.28.0.6", 2*time.Millisecond))
	if out != "" {
		t.Fatalf("plain validate-admin 200 should be dropped, got %q", out)
	}
}

func TestTrafficFormatterKeepsHeartbeatFailures(t *testing.T) {
	out := TrafficLogFormatter(fmtParams("GET", "/auth/validate-admin", 401, "172.28.0.6", time.Millisecond))
	if out == "" {
		t.Fatal("a failing heartbeat must stay visible")
	}
	if !strings.Contains(out, "401") || !strings.Contains(out, "172.28.0.6") {
		t.Fatalf("failure line missing fields: %q", out)
	}
}

func TestTrafficFormatterVisitorGetsIP(t *testing.T) {
	out := TrafficLogFormatter(fmtParams("GET", "/", 200, "82.69.124.224", 15*time.Millisecond))
	if !strings.Contains(out, "82.69.124.224") {
		t.Fatalf("visitor line should carry the IP: %q", out)
	}
}

func TestTrafficFormatterInfra200Lean(t *testing.T) {
	out := TrafficLogFormatter(fmtParams("POST", "/graphql", 200, "172.28.0.6", 5*time.Millisecond))
	if strings.Contains(out, "172.28.0.6") {
		t.Fatalf("infrastructure success should not repeat its IP: %q", out)
	}
}

func TestTrafficFormatterServerErrorMarked(t *testing.T) {
	out := TrafficLogFormatter(fmtParams("GET", "/spotify/listening", 500, "82.69.124.224", time.Second))
	if !strings.Contains(out, "!!! SERVER ERROR") {
		t.Fatalf("server error should be marked: %q", out)
	}
	// Human latency units, not nanoseconds.
	if strings.Contains(out, "ns") {
		t.Fatalf("latency should be human units: %q", out)
	}
}

func TestTrafficFormatterInfraIPClassification(t *testing.T) {
	if !logIsInterestingClient("82.69.124.224") {
		t.Fatal("public IP must read as a visitor")
	}
	if logIsInterestingClient("172.28.0.6") {
		t.Fatal("docker-network IP must not read as a visitor")
	}
}
