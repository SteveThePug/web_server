package services

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
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

	c, err := imapclient.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("IMAP connect: %w", err)
	}
	defer c.Close()

	if err := c.Login(s.Config.IMAP.Email, s.Config.IMAP.Password).Wait(); err != nil {
		return nil, fmt.Errorf("IMAP login: %w", err)
	}

	if _, err := c.Select("INBOX", nil).Wait(); err != nil {
		return nil, fmt.Errorf("IMAP select INBOX: %w", err)
	}

	// Search for emails since the given time
	sinceDate := since.UTC()
	searchCriteria := &imap.SearchCriteria{
		Since: sinceDate,
	}
	searchData, err := c.Search(searchCriteria, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("IMAP search: %w", err)
	}

	if searchData.AllSeqNums() == nil || len(searchData.AllSeqNums()) == 0 {
		return nil, nil
	}

	seqNums := searchData.AllSeqNums()
	log.Printf("[EmailSync/IMAP] Found %d messages since %s", len(seqNums), sinceDate.Format(time.RFC3339))

	seqSet := imap.SeqSetNum(seqNums...)
	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	fetchCmd := c.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	var messages []graphMessage

	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		var envelope *imap.Envelope
		var bodyContent string

		for {
			item := msg.Next()
			if item == nil {
				break
			}

			switch data := item.(type) {
			case imapclient.FetchItemDataEnvelope:
				envelope = data.Envelope
			case imapclient.FetchItemDataBodySection:
				body, err := io.ReadAll(data.Literal)
				if err != nil {
					log.Printf("[EmailSync/IMAP] Error reading body: %v", err)
					continue
				}
				bodyContent = string(body)
			}
		}

		if envelope == nil {
			continue
		}

		// Convert to graphMessage format for unified processing
		var fromName, fromAddr string
		if len(envelope.From) > 0 {
			fromName = envelope.From[0].Name
			fromAddr = envelope.From[0].Addr()
		}

		gm := graphMessage{
			ID:               envelope.MessageID,
			Subject:          envelope.Subject,
			ReceivedDateTime: envelope.Date.Format(time.RFC3339),
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

	if err := fetchCmd.Close(); err != nil {
		return nil, fmt.Errorf("IMAP fetch: %w", err)
	}

	if err := c.Logout().Wait(); err != nil {
		log.Printf("[EmailSync/IMAP] Logout error: %v", err)
	}

	return messages, nil
}
