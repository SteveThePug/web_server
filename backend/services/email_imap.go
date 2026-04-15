package services

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

type IMAPConfig struct {
	Host     string
	Port     string
	Email    string
	Password string
}

// fetchEmailsIMAP connects to an IMAP server and retrieves emails since the given time.
func (s *EmailSyncService) fetchEmailsIMAP(since time.Time) ([]graphMessage, error) {
	addr := fmt.Sprintf("%s:%s", s.Config.IMAP.Host, s.Config.IMAP.Port)

	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP connect: %w", err)
	}
	defer c.Logout()

	if err := c.Login(s.Config.IMAP.Email, s.Config.IMAP.Password); err != nil {
		return nil, fmt.Errorf("IMAP login: %w", err)
	}

	_, err = c.Select("INBOX", false)
	if err != nil {
		return nil, fmt.Errorf("IMAP select INBOX: %w", err)
	}

	// Search for emails since the given time
	criteria := imap.NewSearchCriteria()
	criteria.Since = since.UTC()

	seqNums, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("IMAP search: %w", err)
	}

	if len(seqNums) == 0 {
		return nil, nil
	}

	log.Printf("[EmailSync/IMAP] Found %d messages since %s", len(seqNums), since.UTC().Format(time.RFC3339))

	seqSet := new(imap.SeqSet)
	seqSet.AddNum(seqNums...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{imap.FetchEnvelope, section.FetchItem()}

	msgChan := make(chan *imap.Message, len(seqNums))
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet, items, msgChan)
	}()

	var messages []graphMessage

	for msg := range msgChan {
		if msg.Envelope == nil {
			continue
		}

		var fromName, fromAddr string
		if len(msg.Envelope.From) > 0 {
			fromName = msg.Envelope.From[0].PersonalName
			fromAddr = msg.Envelope.From[0].Address()
		}

		var bodyContent string
		bodyReader := msg.GetBody(section)
		if bodyReader != nil {
			mr, err := mail.CreateReader(bodyReader)
			if err == nil {
				for {
					p, err := mr.NextPart()
					if err != nil {
						break
					}
					switch p.Header.(type) {
					case *mail.InlineHeader:
						b, err := io.ReadAll(p.Body)
						if err == nil {
							bodyContent = string(b)
						}
					}
				}
			}
		}

		gm := graphMessage{
			ID:               msg.Envelope.MessageId,
			Subject:          msg.Envelope.Subject,
			ReceivedDateTime: msg.Envelope.Date.Format(time.RFC3339),
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
		}

		messages = append(messages, gm)
	}

	if err := <-done; err != nil {
		return nil, fmt.Errorf("IMAP fetch: %w", err)
	}

	return messages, nil
}
