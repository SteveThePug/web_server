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
	tag    atomic.Int64
}

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

func (c *imapClient) Close() error {
	return c.conn.Close()
}

func (c *imapClient) nextTag() string {
	return fmt.Sprintf("A%04d", c.tag.Add(1))
}

func (c *imapClient) readLine() (string, error) {
	line, err := c.reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n"), err
}

// sendCommand sends a tagged command and reads lines until the tagged response.
// Returns all untagged response lines and the final tagged status line.
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
	// Status line is like: A0001 OK ...  or  A0001 NO ...
	parts := strings.SplitN(status, " ", 3)
	if len(parts) < 2 || parts[1] != "OK" {
		return untagged, fmt.Errorf("command %q failed: %s", cmd, status)
	}
	return untagged, nil
}

// fetchLiteral reads an IMAP literal {N}\r\n followed by N bytes.
func (c *imapClient) readLiteral(size int) (string, error) {
	buf := make([]byte, size)
	_, err := io.ReadFull(c.reader, buf)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// sendFetch sends a FETCH command and collects the full response including literals.
// Returns raw response lines (with literals inlined after their header lines).
func (c *imapClient) sendFetch(cmd string) ([]string, error) {
	tag := c.nextTag()
	_, err := fmt.Fprintf(c.conn, "%s %s\r\n", tag, cmd)
	if err != nil {
		return nil, err
	}

	var lines []string
	for {
		line, err := c.readLine()
		if err != nil {
			return lines, err
		}

		// Check for literal marker {N} at end of line
		if idx := strings.LastIndex(line, "{"); idx >= 0 && strings.HasSuffix(line, "}") {
			sizeStr := line[idx+1 : len(line)-1]
			var size int
			if _, err := fmt.Sscanf(sizeStr, "%d", &size); err == nil && size > 0 {
				literal, err := c.readLiteral(size)
				if err != nil {
					return lines, fmt.Errorf("reading literal of %d bytes: %w", size, err)
				}
				lines = append(lines, line)
				lines = append(lines, literal)
				continue
			}
		}

		if strings.HasPrefix(line, tag+" ") {
			// Check for OK
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 2 && parts[1] != "OK" {
				return lines, fmt.Errorf("FETCH failed: %s", line)
			}
			return lines, nil
		}
		lines = append(lines, line)
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

	// Quote the password to handle special characters
	quotedPass := fmt.Sprintf("%q", s.Config.IMAP.Password)
	if _, err := c.sendCommandOK(fmt.Sprintf("LOGIN %s %s", s.Config.IMAP.Email, quotedPass)); err != nil {
		return nil, fmt.Errorf("IMAP login: %w", err)
	}

	if _, err := c.sendCommandOK("SELECT INBOX"); err != nil {
		return nil, fmt.Errorf("IMAP select INBOX: %w", err)
	}

	// SEARCH SINCE uses date only (no time), per IMAP spec
	dateStr := since.UTC().Format("02-Jan-2006")
	untagged, err := c.sendCommandOK(fmt.Sprintf("SEARCH SINCE %s", dateStr))
	if err != nil {
		return nil, fmt.Errorf("IMAP search: %w", err)
	}

	// Parse sequence numbers from "* SEARCH 1 2 3 ..."
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

	seqSet := strings.Join(seqNums, ",")
	fetchLines, err := c.sendFetch(fmt.Sprintf("FETCH %s (BODY[])", seqSet))
	if err != nil {
		return nil, fmt.Errorf("IMAP fetch: %w", err)
	}

	// Parse fetched messages - literals contain the raw RFC822 message
	var messages []graphMessage
	for i := 0; i < len(fetchLines); i++ {
		line := fetchLines[i]
		// Look for a literal following this line
		if strings.Contains(line, "BODY[]") && strings.HasSuffix(line, "}") {
			if i+1 < len(fetchLines) {
				raw := fetchLines[i+1]
				i++ // skip the literal
				gm, err := parseRawEmail(raw)
				if err != nil {
					log.Printf("[EmailSync/IMAP] Error parsing email: %v", err)
					continue
				}
				messages = append(messages, gm)
			}
		}
	}

	c.sendCommand("LOGOUT")
	return messages, nil
}

// parseRawEmail parses an RFC822 email into a graphMessage.
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

// extractTextBody pulls the text/plain or text/html content from a message body.
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

// decodeHeader decodes RFC 2047 encoded header values.
func decodeHeader(s string) string {
	dec := new(mime.WordDecoder)
	decoded, err := dec.DecodeHeader(s)
	if err != nil {
		return s
	}
	return decoded
}
