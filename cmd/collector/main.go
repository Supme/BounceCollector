package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	collector "github.com/Supme/BounceCollector"
	"log"
	"net/mail"
	"os"
	"time"
)

func main() {
	var (
		filename string
		server   string
		username string
		password string
		mailbox  string
		since    string
		before   string
	)

	flag.StringVar(&filename, "f", "", "Output csv file, default to stdout")
	flag.StringVar(&server, "i", "imap.yandex.ru:993", "IMAP server:port")
	flag.StringVar(&username, "u", "user@yandex.ru", "IMAP username")
	flag.StringVar(&password, "p", "pa55w0rd", "IMAP password")
	flag.StringVar(&mailbox, "m", "INBOX", "IMAP mailbox")
	flag.StringVar(&since, "s", time.Now().Add(-time.Hour*24*90).Format(time.RFC3339), "Date since. Format: "+time.RFC3339+" (RFC3339).")
	flag.StringVar(&before, "b", time.Now().Format(time.RFC3339), "Date before. Format: "+time.RFC3339+" (RFC3339).")
	flag.Parse()

	fromDate, e := time.Parse(time.RFC3339, since)
	if e != nil {
		log.Fatal(e)
	}
	toDate, e := time.Parse(time.RFC3339, before)
	if e != nil {
		log.Fatal(e)
	}

	//log.Println("Start")

	result := make(chan collector.Report, 10)
	done := make(chan error, 1)

	go func() {
		err := collector.ReportCollectorIMAP(server, username, password, mailbox, fromDate, toDate, result)
		done <- err
	}()

	var f *os.File
	if filename == "" {
		f = os.Stdout
	} else {
		f, e = os.Create(filename)
		if e != nil {
			log.Fatal(e)
		}
	}
	defer f.Close()

	w := csv.NewWriter(f)
	e = w.Write([]string{"ID", "Date", "Reporter", "Sender", "Recipient", "ReportType", "MessageID", "PostmasterMsgType", "Message"})
	if e != nil {
		log.Fatal(e)
	}
	defer w.Flush()

	for report := range result {
		e = w.Write([]string{
			report.ID,
			report.Date.Format(time.RFC3339),
			printMail(report.Reporter),
			printMail(report.Sender),
			printMail(report.Recipient),
			report.ReportType.String(),
			report.MessageID,
			report.PostmasterMsgType,
			report.Message,
		})
		if e != nil {
			log.Fatal(e)
		}
	}

	err := <-done
	if err != nil {
		log.Fatal(e)
	}
	//log.Println("Done")
}

func printMail(m mail.Address) string {
	if m.Name == "" {
		return "<" + m.Address + ">"
	}
	return fmt.Sprintf(`"%s" <%s>`, m.Name, m.Address)
}
