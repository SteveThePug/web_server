package services

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net"
	"net/mail"
	"strings"
	"sync/atomic"
	"time"
)

// This file is a hand-rolled IMAP client, used as an alternative mail source
// to Microsoft Graph. It exists because the Graph path needs an Azure app
// registration and admin consent, while IMAP needs only an app password.
//
// It implements the bare minimum of RFC 3501 — LOGIN, SELECT, SEARCH, FETCH,
// LOGOUT — by matching on string prefixes rather than parsing IMAP's grammar
// properly. It works against Outlook and similar servers with the response
// shapes seen in practice; it is not a general-purpose IMAP client.

// IMAPConfig holds the mailbox credentials. Password is an app password, not
// an account password.
type IMAPConfig struct {
	Host     string
	Port     string
	Email    string
	Password string
}

// imapClient is a minimal, lenient IMAP client that tolerates
// Outlook's non-standard response formatting.
type imapClient struct {
	conn   net.Conn
	reader *bufio.Reader
	// tag numbers the commands. IMAP allows several in flight at once
	// identified by tag; this client never does that, so the atomic is
	// belt-and-braces rather than load-bearing.
	tag atomic.Int64
}

// imapDial opens an implicit-TLS IMAP connection (port 993 style, TLS from the
// first byte — there is no STARTTLS support here) and consumes the server
// greeting, which arrives unsolicited before any command.
func imapDial(addr string) (*imapClient, error) {
	conn, err := tls.Dial("tcp", addr, nil)
	if err != nil {
		return nil, err
	}
	c := &imapClient{
		conn:   conn,
		reader: bufio.NewReader(conn),
	}
	// Read server greeting
	if _, err := c.readLine(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("reading greeting: %w", err)
	}
	return c, nil
}

// Close drops the TCP connection. Callers normally send LOGOUT first.
func (c *imapClient) Close() error {
	return c.conn.Close()
}

// nextTag returns the next command tag, e.g. "A0001". The width is cosmetic:
// past 9999 the tags simply get longer, and they stay unique.
func (c *imapClient) nextTag() string {
	return fmt.Sprintf("A%04d", c.tag.Add(1))
}

// readLine reads one CRLF-terminated protocol line with the terminator
// stripped. There is no read deadline anywhere in this client, so an
// unresponsive server can block a sync until the TCP connection itself fails.
func (c *imapClient) readLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

// sendCommand sends a tagged command and reads lines until the line beginning
// with that tag, which is the server's final status for the command.
// Everything before it is an untagged ("* ...") response and is returned.
//
// This must not be used for FETCH: it reads line by line and would mistake the
// bytes of a literal for protocol lines. Use sendFetch instead.
func (c *imapClient) sendCommand(cmd string) (untagged []string, status string, err error) {
	tag := c.nextTag()
	_, err = fmt.Fprintf(c.conn, "%s %s\r\n", tag, cmd)
	if err != nil {
		return nil, "", err
	}

	for {
		line, err := c.readLine()
		if err != nil {
			return untagged, "", err
		}
		if strings.HasPrefix(line, tag+" ") {
			return untagged, line, nil
		}
		untagged = append(untagged, line)
	}
}

// sendCommandOK sends a command and returns an error if the response is not OK.
func (c *imapClient) sendCommandOK(cmd string) ([]string, error) {
	untagged, status, err := c.sendCommand(cmd)
	if err != nil {
		return untagged, err
	}
	// Status line is "<tag> OK ..." on success, "<tag> NO ..." or
	// "<tag> BAD ..." otherwise; only the second field matters.
	parts := strings.SplitN(status, " ", 3)
	if len(parts) < 2 || parts[1] != "OK" {
		return untagged, fmt.Errorf("command %q failed: %s", cmd, status)
	}
	return untagged, nil
}

// readLiteral reads exactly size bytes of an IMAP literal, i.e. the payload
// announced by a trailing "{N}" on the preceding line. It must read by byte
// count rather than by line because the payload is arbitrary binary data that
// will itself contain CRLFs.
func (c *imapClient) readLiteral(size int) (string, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(c.reader, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// fetchPart is one FETCH response line plus the literal payload that followed
// it, if any. Pairing them in the type removes the need for callers to walk an
// interleaved slice and manually skip every second entry — the previous shape,
// where a literal was simply the next element and a caller that forgot to
// advance would parse message bytes as a protocol line.
type fetchPart struct {
	Line string
	// Literal is the payload that followed Line, valid only when HasLiteral.
	Literal    string
	HasLiteral bool
}

// sendFetch sends a FETCH and collects its response, handling literals.
//
// The literal is detected with LastIndex("{") plus a "}" suffix rather than a
// real parse, so a line whose text merely ends in braces could be misread —
// tolerable because only FETCH responses reach this code.
func (c *imapClient) sendFetch(cmd string) ([]fetchPart, error) {
	tag := c.nextTag()
	_, err := fmt.Fprintf(c.conn, "%s %s\r\n", tag, cmd)
	if err != nil {
		return nil, err
	}

	var parts []fetchPart
	for {
		line, err := c.readLine()
		if err != nil {
			return parts, err
		}

		// Check for literal marker {N} at end of line
		if idx := strings.LastIndex(line, "{"); idx >= 0 && strings.HasSuffix(line, "}") {
			sizeStr := line[idx+1 : len(line)-1]
			var size int
			if _, err := fmt.Sscanf(sizeStr, "%d", &size); err == nil && size > 0 {
				literal, err := c.readLiteral(size)
				if err != nil {
					return parts, fmt.Errorf("reading literal of %d bytes: %w", size, err)
				}
				parts = append(parts, fetchPart{Line: line, Literal: literal, HasLiteral: true})
				continue
			}
		}

		if strings.HasPrefix(line, tag+" ") {
			// Check for OK
			fields := strings.SplitN(line, " ", 3)
			if len(fields) >= 2 && fields[1] != "OK" {
				return parts, fmt.Errorf("FETCH failed: %s", line)
			}
			return parts, nil
		}
		parts = append(parts, fetchPart{Line: line})
	}
}

// fetchEmailsIMAP connects to an IMAP server and retrieves emails since the given time.
func (s *EmailSyncService) fetchEmailsIMAP(since time.Time) ([]graphMessage, error) {
	addr := fmt.Sprintf("%s:%s", s.Config.IMAP.Host, s.Config.IMAP.Port)

	c, err := imapDial(addr)
	if err != nil {
		return nil, fmt.Errorf("IMAP connect: %w", err)
	}
	defer c.Close()

	// %q wraps the password in double quotes and backslash-escapes any quote
	// or backslash inside it, which happens to be exactly IMAP's quoted-string
	// syntax. The username is NOT quoted: an address needs no escaping.
	// Note this sends the password in a LOGIN command, so it is only safe
	// because the connection is TLS from the first byte.
	quotedPass := fmt.Sprintf("%q", s.Config.IMAP.Password)
	if _, err := c.sendCommandOK(fmt.Sprintf("LOGIN %s %s", s.Config.IMAP.Email, quotedPass)); err != nil {
		return nil, fmt.Errorf("IMAP login: %w", err)
	}

	if _, err := c.sendCommandOK("SELECT INBOX"); err != nil {
		return nil, fmt.Errorf("IMAP select INBOX: %w", err)
	}

	// SEARCH SINCE has day granularity only — the time of day is not
	// expressible — so this always over-fetches back to midnight of `since`.
	// Harmless: SyncEmails deduplicates by message id. The date must be in
	// IMAP's dd-Mon-yyyy form, hence this exact layout string.
	dateStr := since.UTC().Format("02-Jan-2006")
	untagged, err := c.sendCommandOK(fmt.Sprintf("SEARCH SINCE %s", dateStr))
	if err != nil {
		return nil, fmt.Errorf("IMAP search: %w", err)
	}

	// Parse sequence numbers from "* SEARCH 1 2 3 ...". parts[2:] skips the
	// "*" and "SEARCH" tokens; a bare "* SEARCH" (no matches) has len 2 and
	// is skipped by the length test.
	var seqNums []string
	for _, line := range untagged {
		if strings.HasPrefix(line, "* SEARCH") {
			parts := strings.Fields(line)
			if len(parts) > 2 {
				seqNums = append(seqNums, parts[2:]...)
			}
		}
	}

	if len(seqNums) == 0 {
		log.Printf("[EmailSync/IMAP] No messages found since %s", dateStr)
		c.sendCommand("LOGOUT")
		return nil, nil
	}

	log.Printf("[EmailSync/IMAP] Found %d messages since %s", len(seqNums), dateStr)

	// Every matching message is fetched in one command with a comma-separated
	// sequence set, and each full body is held in memory. A mailbox with a
	// very busy day could make this large — there is no batching.
	seqSet := strings.Join(seqNums, ",")
	fetchParts, err := c.sendFetch(fmt.Sprintf("FETCH %s (BODY[])", seqSet))
	if err != nil {
		return nil, fmt.Errorf("IMAP fetch: %w", err)
	}

	// Each body arrives as a "... BODY[] {N}" line carrying its literal, so
	// the pairing is now part of the value rather than something this loop has
	// to reconstruct by skipping entries. A message that fails to parse is
	// logged and skipped rather than failing the whole sync.
	var messages []graphMessage
	for _, part := range fetchParts {
		if part.HasLiteral && strings.Contains(part.Line, "BODY[]") {
			gm, err := parseRawEmail(part.Literal)
			if err != nil {
				log.Printf("[EmailSync/IMAP] Error parsing email: %v", err)
				continue
			}
			messages = append(messages, gm)
		}
	}

	// LOGOUT errors are ignored: the emails are already in hand and the
	// deferred Close will drop the connection regardless.
	c.sendCommand("LOGOUT")
	return messages, nil
}

// parseRawEmail parses a raw RFC822 message into the same graphMessage struct
// the Graph API path produces, so everything downstream is source-agnostic.
// An unparseable Date falls back to now, and a missing From leaves both name
// and address empty rather than failing.
func parseRawEmail(raw string) (graphMessage, error) {
	msg, err := mail.ReadMessage(strings.NewReader(raw))
	if err != nil {
		return graphMessage{}, fmt.Errorf("parsing message: %w", err)
	}

	header := msg.Header

	// Parse From
	var fromName, fromAddr string
	if fromList, err := header.AddressList("From"); err == nil && len(fromList) > 0 {
		fromName = fromList[0].Name
		fromAddr = fromList[0].Address
	}

	// Parse Date
	dateStr := header.Get("Date")
	parsedDate, err := mail.ParseDate(dateStr)
	if err != nil {
		parsedDate = time.Now()
	}

	// Extract body text
	bodyContent := extractTextBody(msg.Header, msg.Body)

	return graphMessage{
		ID:               header.Get("Message-ID"),
		Subject:          decodeHeader(header.Get("Subject")),
		ReceivedDateTime: parsedDate.Format(time.RFC3339),
		From: graphFrom{
			EmailAddress: graphEmailAddress{
				Name:    fromName,
				Address: fromAddr,
			},
		},
		Body: graphBody{
			ContentType: "text",
			Content:     bodyContent,
		},
	}, nil
}

// extractTextBody pulls readable content out of a message body, preferring
// text/plain and falling back to text/html.
//
// It recurses into nested multipart parts by re-wrapping the part's bytes in a
// synthetic mail.Header carrying just its Content-Type — a shortcut that works
// because this function only ever reads Content-Type.
//
// Quoted-printable and base64 transfer encodings are NOT decoded, so a
// base64-encoded body reaches Claude as base64. Read errors on a part are
// swallowed and the part is skipped.
func extractTextBody(header mail.Header, body io.Reader) string {
	contentType := header.Get("Content-Type")
	if contentType == "" {
		contentType = "text/plain"
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		// Try reading as plain text
		b, _ := io.ReadAll(body)
		return string(b)
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			b, _ := io.ReadAll(body)
			return string(b)
		}

		mr := multipart.NewReader(body, boundary)
		var textContent, htmlContent string
		for {
			part, err := mr.NextPart()
			if err != nil {
				break
			}
			partType := part.Header.Get("Content-Type")
			partMedia, _, _ := mime.ParseMediaType(partType)
			b, err := io.ReadAll(part)
			if err != nil {
				continue
			}
			switch partMedia {
			case "text/plain":
				textContent = string(b)
			case "text/html":
				htmlContent = string(b)
			case "multipart/alternative", "multipart/related", "multipart/mixed":
				// Recursively handle nested multipart
				nested := extractTextBody(
					mail.Header{"Content-Type": {partType}},
					strings.NewReader(string(b)),
				)
				if nested != "" && textContent == "" {
					textContent = nested
				}
			}
		}
		if textContent != "" {
			return textContent
		}
		return htmlContent
	}

	b, _ := io.ReadAll(body)
	return string(b)
}

// decodeHeader decodes RFC 2047 encoded-words (the "=?UTF-8?Q?...?=" form
// used for non-ASCII subjects), returning the input unchanged if it is not
// encoded or cannot be decoded.
func decodeHeader(s string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return decoded
}
