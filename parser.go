package bounce_collector

import (
	gomime "github.com/ProtonMail/go-mime"
	"io"
	"mime"
	"net/mail"
	"regexp"
	"time"
)

type ReportType int

const (
	ReportTypeUnknown ReportType = iota
	ReportTypeDeliveryStatus
	ReportTypeFeedback
	ReportTypeXMailer
)

func (t ReportType) String() string {
	switch t {
	case ReportTypeXMailer:
		return "X-Mailer"
	case ReportTypeDeliveryStatus:
		return "Delivery Status"
	case ReportTypeFeedback:
		return "Feedback Loop"
	default:
		return "Unknown"
	}
}

func getMessageType(msg *mail.Message) (ReportType, error) {
	if msg.Header.Get("X-Mailer-Daemon-Error") != "" && msg.Header.Get("X-Mailer-Daemon-Recipients") != "" {
		return ReportTypeXMailer, nil
	}

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return ReportTypeUnknown, err
	}
	if mediaType == "multipart/report" {
		if reportType, ok := params["report-type"]; ok {
			switch reportType {
			case "feedback-report":
				return ReportTypeFeedback, nil
			case "delivery-status":
				return ReportTypeDeliveryStatus, nil
			}
		}
	}

	return ReportTypeUnknown, nil
}

type Report struct {
	ID                string
	Date              time.Time
	Reporter          mail.Address
	Sender            mail.Address
	Recipient         mail.Address
	ReportType        ReportType
	MessageID         string
	PostmasterMsgType string
	Message           string
}

func Parser(r io.Reader) (Report, error) {
	msg, err := mail.ReadMessage(r)
	if err != nil {
		return Report{}, err
	}

	msgType, err := getMessageType(msg)
	if err != nil {
		return Report{}, err
	}

	switch msgType {
	case ReportTypeXMailer:
		return ReportXMailer(msg)

	case ReportTypeFeedback:
		return ReportParser(msg, ReportTypeFeedback)

	case ReportTypeDeliveryStatus:
		return ReportParser(msg, ReportTypeDeliveryStatus)

	default:
		return Report{ReportType: ReportTypeUnknown}, nil

	}
}

func ReportParser(msg *mail.Message, reportType ReportType) (Report, error) {
	var report Report
	report.ReportType = reportType

	reporter, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		return Report{}, err
	}
	report.Reporter = *reporter

	report.Date, err = parseDate(msg.Header.Get("Date"))
	if err != nil {
		return Report{}, err
	}

	report.Message, err = gomime.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		return Report{}, err
	}

	partHdr, err := getHeadersFromRFC822Part(msg)
	if err != nil {
		return Report{}, err
	}

	recipient, err := mail.ParseAddress(partHdr.Get("To"))
	if err != nil {
		return Report{}, err
	}
	report.Recipient = *recipient

	sender, err := mail.ParseAddress(partHdr.Get("From"))
	if err != nil {
		return Report{}, err
	}
	report.Sender = *sender

	report.MessageID = partHdr.Get("Message-ID")

	report.PostmasterMsgType = partHdr.Get("X-Postmaster-Msgtype")

	return report, nil
}

func ReportXMailer(msg *mail.Message) (Report, error) {
	var report Report
	report.ReportType = ReportTypeXMailer

	reporter, err := mail.ParseAddress(msg.Header.Get("From"))
	if err != nil {
		return Report{}, err
	}
	report.Reporter = *reporter

	recipient, err := mail.ParseAddress(msg.Header.Get("X-Mailer-Daemon-Recipients"))
	if err != nil {
		return Report{}, err
	}
	report.Recipient = *recipient

	report.Date, err = parseDate(msg.Header.Get("Date"))
	if err != nil {
		return Report{}, err
	}

	report.Message = msg.Header.Get("X-Mailer-Daemon-Error")

	includeMsg, err := getHeadersFromTextPart(msg)
	if err != nil {
		return Report{}, err
	}

	sender, err := mail.ParseAddress(includeMsg.Get("From"))
	if err != nil {
		return Report{}, err
	}
	report.Sender = *sender

	report.MessageID = includeMsg.Get("Message-ID")

	report.PostmasterMsgType = includeMsg.Get("X-Postmaster-Msgtype")

	return report, nil
}

var reDate = regexp.MustCompile(`(\w{3},\s)?\d{1,2}\s+\w{3}\s+\d{4}\s+\d{1,2}:\d{1,2}:\d{1,2}\s+([+-]?\d{4}|\w{3})`)

func parseDate(s string) (time.Time, error) {
	reS := reDate.FindString(s)
	// time.RFC1123Z ???
	t, err := time.Parse("Mon, _2 Jan 2006 15:04:05 -0700", reS)
	if err != nil {
		t, err = time.Parse("_2 Jan 2006 15:04:05 -0700", reS)
		// ToDo "Wed, 04 Jul 2018 15:37:57 GMT"
		if err != nil {
			t, err = time.Parse("Mon, _2 Jan 2006 15:04:05 MST", reS)
			if err != nil {
				t, err = time.Parse("_2 Jan 2006 15:04:05 MST", reS)
			}
		}
	}
	return t, err
}
