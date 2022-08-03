package bounce_collector

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
)

//func getMailFromRFC822Part(msg *mail.Message) (*mail.Message, error) {
//	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
//	if err != nil {
//		return nil, err
//	}
//	if strings.HasPrefix(mediaType, "multipart/") {
//		mr := multipart.NewReader(msg.Body, params["boundary"])
//		for {
//			p, err := mr.NextPart()
//			if err == io.EOF {
//				return nil, fmt.Errorf("this message doesn't have rfc822 part")
//			}
//			if err != nil {
//				return nil, err
//			}
//
//			if mt, _, e := mime.ParseMediaType(p.Header.Get("Content-Type")); strings.Contains(mt, "rfc822") && e == nil {
//				m, err := mail.ReadMessage(p)
//				if err != nil {
//					return nil, err
//				}
//				return m, nil
//			}
//		}
//	}
//	return nil, fmt.Errorf("this not multipart message")
//}

func getHeadersFromRFC822Part(msg *mail.Message) (*textproto.MIMEHeader, error) {
	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				return nil, fmt.Errorf("this message doesn't have rfc822 part")
			}
			if err != nil {
				return nil, err
			}

			if mt, _, e := mime.ParseMediaType(p.Header.Get("Content-Type")); strings.Contains(mt, "rfc822") && e == nil {
				tp := textproto.NewReader(bufio.NewReader(p))
				hdr, err := tp.ReadMIMEHeader()
				if err != nil && err != io.EOF {
					return nil, err
				}
				return &hdr, nil
			}
		}
	}
	return nil, fmt.Errorf("this not multipart message")
}

var reMailFromTextPart = regexp.MustCompile(`(?m:^[\r,\n]+)(?ms:^[a-zA-Z0-9\-]{2,25}:.*)`)

//func getMailFromTextPart(msg *mail.Message) (*mail.Message, error) {
//	b, err := io.ReadAll(msg.Body)
//	if err != nil {
//		return nil, err
//	}
//	reMail := reMailFromTextPart.FindIndex(b)
//	if len(reMail) != 0 {
//		mr := bytes.NewReader(b[reMail[0]+2:])
//		m, err := mail.ReadMessage(mr)
//		if err != nil {
//			return nil, err
//		}
//		return m, nil
//	}
//	return nil, fmt.Errorf("mail not found in text")
//}

func getHeadersFromTextPart(msg *mail.Message) (*textproto.MIMEHeader, error) {
	b, err := io.ReadAll(msg.Body)
	if err != nil {
		return nil, err
	}
	reMail := reMailFromTextPart.FindIndex(b)
	if len(reMail) != 0 {
		mr := bytes.NewReader(b[reMail[0]+2:])
		tp := textproto.NewReader(bufio.NewReader(mr))
		hdr, err := tp.ReadMIMEHeader()
		if err != nil && err != io.EOF {
			return nil, err
		}
		return &hdr, nil
	}
	return nil, fmt.Errorf("mail not found in text")
}
