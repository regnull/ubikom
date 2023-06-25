package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"

	"github.com/emersion/go-message/mail"
)

// Create a mail message with attachment.

func main() {
	var (
		contentType        string // application/pdf
		fileName           string
		attachmentFileName string
		subject            string
		from               string
		to                 string
		text               string
	)
	flag.StringVar(&contentType, "content-type", "", "content type")
	flag.StringVar(&fileName, "file-name", "", "file name")
	flag.StringVar(&attachmentFileName, "attach-file-name", "", "attachment file name")
	flag.StringVar(&subject, "subject", "", "subject")
	flag.StringVar(&from, "from", "", "from")
	flag.StringVar(&to, "to", "", "to")
	flag.StringVar(&text, "text", "Please see attached", "email body text")
	flag.Parse()

	if contentType == "" {
		log.Fatal("--content-type must be specified")
	}

	if fileName == "" {
		log.Fatal("--file-name must be specified")
	}

	if attachmentFileName == "" {
		log.Fatal("--attach-file-name must be specified")
	}

	if subject == "" {
		log.Fatal("--subject must be specified")
	}

	if from == "" {
		log.Fatal("--from must be specified")
	}

	if to == "" {
		log.Fatal("--to must be specified")
	}

	var b bytes.Buffer

	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetText("To", to)
	h.SetText("From", from)
	h.SetText("Subject", subject)
	w, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	// Create an attachment
	var ah mail.AttachmentHeader
	ah.Set("Content-Type", contentType)
	ah.SetFilename(attachmentFileName)
	w1, err := w.CreateAttachment(ah)
	if err != nil {
		log.Fatal(err)
	}
	bb, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Fatal(err)
	}
	w1.Write(bb)
	w1.Close()

	// Write the email body.
	tw, err := w.CreateInline()
	if err != nil {
		log.Fatal(err)
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/plain")
	w2, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal(err)
	}
	io.WriteString(w2, text)
	w2.Close()
	tw.Close()

	w.Close()

	// Write the output.
	fmt.Println(b.String())
}
