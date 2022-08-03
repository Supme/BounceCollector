package bounce_collector

import (
	"net/mail"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParser(t *testing.T) {
	eml := map[string]Report{
		"testdata/mail_ru_invalid_mailbox.eml": {
			// "Fri, 15 Jul 2022 16:40:55 +0300"
			Date:              time.Date(2022, time.July, 15, 16, 40, 55, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "mailer-daemon@corp.mail.ru", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeXMailer,
			MessageID:         "<16578924547016.111222333@gonder>",
			PostmasterMsgType: "campaign7016",
			Message:           "user_not_found",
		},
		"testdata/mail_ru_report_FBL.eml": {
			// "2022-07-19 21:09:46 +0300 MSK"
			Date:              time.Date(2022, time.July, 19, 21, 9, 46, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "noreply@corp.mail.ru", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeFeedback,
			MessageID:         "<16093366945646.54672397@gonder>",
			PostmasterMsgType: "campaign5646",
			Message:           "Mail.ru abuse report (Feedback Loop)",
		},
		"testdata/aoaes_ru_report_delivery-status.eml": {
			// "Mon, 8 Apr 2019 16:55:07 +0300"
			Date:              time.Date(2019, time.April, 8, 16, 55, 07, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "postmaster@atomsbt.ru", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeDeliveryStatus,
			MessageID:         "<15547288993611.20249452@gonder>",
			PostmasterMsgType: "campaign3611",
			Message:           "Undeliverable: [!!Mass Mail]Выиграйте поездку на завод Nissan!",
		},
		"testdata/gazprom-neft_report_delivery-status.eml": {
			// "2019-04-08 17:59:17 +0300 MSK"
			Date:              time.Date(2019, time.April, 8, 17, 59, 17, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "GPNpostmaster@gazprom-neft.ru", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: "Recipient"},
			ReportType:        ReportTypeDeliveryStatus,
			MessageID:         "<15547355523611.20333125@gonder>",
			PostmasterMsgType: "campaign3611",
			Message:           "Undeliverable: [!!Mass Mail]Выиграйте поездку на завод Nissan!",
		},
		"testdata/indexgroup_ru_report_delivery-status.eml": {
			// "2019-05-16 19:04:22 +0300 MSK"
			Date:              time.Date(2019, time.May, 16, 19, 4, 22, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "postmaster@indexgroup.ru", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeDeliveryStatus,
			MessageID:         "<15580225773754.20866238@gonder>",
			PostmasterMsgType: "campaign3754",
			Message:           "Undeliverable: [MASSMAIL] Выгодные предложения, специальные условия и новые технологии в дайджесте Nissan!",
		},
		"testdata/tnt_com_report_delivery-status.eml": {
			// "2019-04-08 17:16:56 +0300 MSK"
			Date:              time.Date(2019, time.April, 8, 17, 16, 56, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "Postmaster@tnt.com", Name: ""},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeDeliveryStatus,
			MessageID:         "<OFC740BF7C.982022CF-ON802583D6.004E7815-802583D6.004E7835@tnt.com>",
			PostmasterMsgType: "campaign3611",
			Message:           "DELIVERY FAILURE: User recipient (recipient@domain.tld) not listed in Domino Directory",
		},
		"testdata/mvd_ru_report_delivery-status.eml": {
			// "2021-12-22 16:43:14 +0300 MSK"
			Date:              time.Date(2021, time.December, 22, 16, 43, 14, 00, time.FixedZone("MSK", 10800)),
			Reporter:          mail.Address{Address: "helpdesk@mvd.ru", Name: "ЕЦЭ ИСОД МВД России"},
			Sender:            mail.Address{Address: "sender@domain.tld", Name: "Sender"},
			Recipient:         mail.Address{Address: "recipient@domain.tld", Name: ""},
			ReportType:        ReportTypeDeliveryStatus,
			MessageID:         "<16401805916732.72318109@gonder>",
			PostmasterMsgType: "campaign6732",
			Message:           "Почта не может быть доставлена получателю",
		},
	}

	for filename, report := range eml {
		m, err := os.Open(filename)
		if err != nil {
			t.Fatal(err)
		}

		parseMsg, err := Parser(m)
		if err != nil {
			t.Fatal(err)
		}

		if parseMsg.Date.Unix() != report.Date.Unix() {
			t.Error("bad message date", parseMsg.Date.String(), "!=", report.Date.String(), "  in", filename)
		}

		if strings.Compare(parseMsg.Reporter.Name, report.Reporter.Name) != 0 {
			t.Errorf("bad reporter name '%s' != '%s' in %s", parseMsg.Reporter.Name, report.Reporter.Name, filename)
		}
		if strings.Compare(parseMsg.Reporter.Address, report.Reporter.Address) != 0 {
			t.Errorf("bad reporter email '%s' != '%s' in %s", parseMsg.Reporter.Address, report.Reporter.Address, filename)
		}

		if strings.Compare(parseMsg.Sender.Name, report.Sender.Name) != 0 {
			t.Errorf("bad sender name '%s' != '%s' in %s", parseMsg.Sender.Name, report.Sender.Name, filename)
		}
		if strings.Compare(parseMsg.Sender.Address, report.Sender.Address) != 0 {
			t.Errorf("bad sender email '%s' != '%s' in %s", parseMsg.Sender.Address, report.Sender.Address, filename)
		}

		if strings.Compare(parseMsg.Recipient.Name, report.Recipient.Name) != 0 {
			t.Errorf("bad recipient name '%s' != '%s' in %s", parseMsg.Recipient.Name, report.Recipient.Name, filename)
		}
		if strings.Compare(parseMsg.Recipient.Address, report.Recipient.Address) != 0 {
			t.Errorf("bad recipient email '%s' != '%s' in %s", parseMsg.Recipient.Address, report.Recipient.Address, filename)
		}

		if parseMsg.ReportType != report.ReportType {
			t.Error("bad report type", parseMsg.ReportType, "!=", report.ReportType, " in", filename)
		}

		if strings.Compare(parseMsg.MessageID, report.MessageID) != 0 {
			t.Error("bad message id address", parseMsg.MessageID, "!=", report.MessageID, " in", filename)
		}

		if strings.Compare(parseMsg.PostmasterMsgType, report.PostmasterMsgType) != 0 {
			t.Error("bad postmaster message type", parseMsg.PostmasterMsgType, "!=", report.PostmasterMsgType, " in", filename)
		}

		if strings.Compare(parseMsg.Message, report.Message) != 0 {
			t.Error("bad message", parseMsg.Message, "!=", report.Message, " in", filename)
		}
	}
}
