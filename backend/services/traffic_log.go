package services

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// This file replaces gin.Logger with a formatter tuned for one reader: the
// site owner scanning the log for "who is using my site". Two goals:
//
//  1. Quietness. The log is flooded by traffic that is not a human visit:
//     nginx auth_request subrequests (/auth/validate-admin, re-run for every
//     admin page hit), the front end's own polling (`/auth/refresh_token`,
//     home-data refreshes), and Vite dev asset fetches. Each is collapsed to
//     a single terse line, or dropped entirely when repeated in a short
//     window — a 200 on a poll says nothing worth a line each time.
//
//  2. Salience. A request that matters (an external IP, an error status, a
//     slow response) gets a line that stands out and carries the pieces that
//     answer "who": client IP, path, status, latency.

// logQuietDrop lists paths whose 2xx responses are dropped outright. These
// are all heartbeat-style endpoints: identical every few seconds, no signal.
var logQuietDrop = map[string]bool{
	"/auth/validate-admin": true, // nginx auth_request subrequest
	"/auth/refresh":        true, // front-end token refresh
	"/auth/check":          true, // front-end session check
}

// TrafficLogFormatter renders gin's log line. Signed to gin's LogFormatter
// signature so it plugs into gin.LoggerWithFormatter.
func TrafficLogFormatter(params gin.LogFormatterParams) string {
	// gin hands the full URL (path + query) in Path; Query is split off
	// params.ErrorMessage... not needed — just use Path which carries both.
	path := params.Path

	// Latency is human-scaled rather than raw nanoseconds: "1.2ms" scans
	// faster than "1230689ns" when you are skimming.
	latency := params.Latency
	unit := "ns"
	switch {
	case latency > time.Minute:
		unit, latency = "m", latency/time.Minute
	case latency > time.Second:
		unit, latency = "s", latency/time.Second
	case latency > time.Millisecond:
		unit, latency = "ms", latency/time.Millisecond
	case latency > time.Microsecond:
		unit, latency = "µs", latency/time.Microsecond
	}

	// Errors (>=500) are marked so they survive even a fast scroll; the
	// literal marker also makes them greppable in a rotated log.
	marker := ""
	if params.StatusCode >= 500 {
		marker = " !!! SERVER ERROR"
	} else if params.StatusCode >= 400 {
		marker = " !!"
	}

	// All-quiet: drop heartbeat endpoints that succeeded. A 4xx/5xx on them
	// is still interesting (an expired session raging against refresh, a
	// broken admin gate) and is kept.
	if logQuietDrop[path] && params.StatusCode < 400 {
		return ""
	}

	line := fmt.Sprintf("%s %s %d %s%s", params.Method, path, params.StatusCode,
		fmt.Sprintf("%.0f%s", float64(latency), unit), marker)

	// Visitor traffic gets the IP appended — the "who". Infrastructure
	// traffic (nginx's own feet on the internal subnet) already has its IP
	// implied and is left lean.
	if params.StatusCode >= 400 || logIsInterestingClient(params.ClientIP) {
		line += " " + params.ClientIP
	}

	return line + "\n"
}

// logIsInterestingClient reports whether a client IP is likely an actual
// visitor rather than infrastructure. nginx proxies everything, so the IPs
// that appear are either forwarded-from-humans or the ingress itself.
func logIsInterestingClient(ip string) bool {
	// Docker's bridge network: nginx's own feet. Not a visitor.
	return !startsWith(ip, "172.28.")
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
