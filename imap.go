package bounce_collector

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"io/ioutil"
	"time"
)

func ReportCollectorIMAP(server, username, password, mailbox string, fromDate, toDate time.Time, result chan<- Report) error {
	defer close(result)

	c, err := client.DialTLS(server, nil)
	if err != nil {
		return err
	}
	defer c.Logout()

	if err := c.Login(username, password); err != nil {
		return err
	}

	done := make(chan error, 1)

	mbox, err := c.Select(mailbox, true)
	if err != nil {
		return err
	}

	if mbox.Messages == 0 {
		return fmt.Errorf("no message in mailbox")
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = fromDate
	criteria.Before = toDate
	ids, err := c.Search(criteria)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		return fmt.Errorf("no message in search")
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	// Get the whole message body
	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, 50)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	go func() {
		for msg := range messages {
			r := msg.GetBody(section)
			if r == nil {
				done <- fmt.Errorf("server didn't returned message body")
			}

			report, e := Parser(r)
			if e != nil {
				//log.Printf("parse: %v", e)
				body, _ := ioutil.ReadAll(r)
				l := 8192
				if len(body) < l {
					l = len(body)
				}
			}
			if e == nil && report.ReportType != ReportTypeUnknown {
				report.ID = msg.Envelope.MessageId
				result <- report
			}
		}
	}()

	err = <-done

	return err
}
