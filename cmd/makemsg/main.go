package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emersion/go-message/mail"
)

// Create a mail message with attachment.
// Example:
// makemsg --content-type=application/pdf --file-name=report.pdf --attach-file-name=report.pdf \
//     --subject="Updated SMA200/10 SPY report" --from="sot@ubikom.cc" --to=lgx@ubikom.cc \
//    < message_body.txt

func main() {
	var (
		contentType        string // application/pdf
		fileName           string
		attachmentFileName string
		subject            string
		from               string
		to                 string
	)
	flag.StringVar(&contentType, "content-type", "", "content type")
	flag.StringVar(&fileName, "file-name", "", "file name")
	flag.StringVar(&attachmentFileName, "attach-file-name", "", "attachment file name")
	flag.StringVar(&subject, "subject", "", "subject")
	flag.StringVar(&from, "from", "", "from")
	flag.StringVar(&to, "to", "", "to")
	flag.Parse()

	if subject == "" {
		log.Fatal("--subject must be specified")
	}

	if from == "" {
		log.Fatal("--from must be specified")
	}

	if to == "" {
		log.Fatal("--to must be specified")
	}

	var lines []string
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your message, dot on an empty line to finish: \n")
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if text == ".\n" {
			break
		}
		lines = append(lines, strings.ReplaceAll(text, "\n", ""))
	}
	body := strings.Join(lines, "\n")

	var b bytes.Buffer

	var h mail.Header
	h.SetContentType("multipart/alternative", nil)
	h.SetText("To", to)
	h.SetText("From", from)
	h.SetText("Subject", subject)
	h.SetText("Date", time.Now().Format("02 Jan 06 15:04:05 -0700"))
	w, err := mail.CreateWriter(&b, h)
	if err != nil {
		log.Fatal(err)
	}

	// Create an attachment
	if fileName != "" {
		if contentType == "" {
			log.Fatal("--content-type must be specified")
		}

		if attachmentFileName == "" {
			attachmentFileName = filepath.Base(fileName)
		}

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
	}

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
	io.WriteString(w2, body)
	w2.Close()
	tw.Close()

	w.Close()

	// Write the output.
	fmt.Println(b.String())
}
