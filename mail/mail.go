package mail

import (
	"bytes"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

const (
	ubikomShortSuffix = "@x"
	ubikomLongSuffix  = "@ubikom.cc"
)

// InternalToExternalAddress parses the address and returns the short and full address suitable
// for sending messages to the external destinations.
func InternalToExternalAddress(address string) (shortAddr, fullAddr string, err error) {
	mailAddr, err := mail.ParseAddress(address)
	if err != nil {
		return
	}

	shortAddr = mailAddr.Address
	shortAddr = strings.Replace(shortAddr, ubikomShortSuffix, ubikomLongSuffix, 1)
	fullAddr = fmt.Sprintf("%s <%s>", mailAddr.Name, shortAddr)
	return
}

// ExtractSenderAddress returns full sender's address extracted from the content.
func ExtractSenderAddress(content string) (address string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}
	address = mailMsg.Header.Get("From")
	return
}

func ExtractReceiverInternalName(content string) (receiver string, err error) {
	contentReader := strings.NewReader(content)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}
	addressStr := mailMsg.Header.Get("To")
	address, err := mail.ParseAddress(addressStr)
	if err != nil {
		return
	}
	receiver = strings.Replace(address.Address, "@ubikom.cc", "", 1)
	return
}

// RewriteFromHeader rewrites the message to change sender from internal to external format.
func RewriteFromHeader(message string) (rewrittenMessage string, fromAddr, toAddr string, err error) {
	contentReader := strings.NewReader(message)
	mailMsg, err := mail.ReadMessage(contentReader)
	if err != nil {
		return
	}

	from := mailMsg.Header.Get("From")
	fromAddr, fullAddr, err := InternalToExternalAddress(from)
	if err != nil {
		return
	}

	to := mailMsg.Header.Get("To")
	a, err := mail.ParseAddress(to)
	if err != nil {
		return
	}
	toAddr = a.Address

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("To: %s\n", mailMsg.Header.Get("To"))))
	buf.Write([]byte(fmt.Sprintf("From: %s\n", fullAddr)))
	for name, values := range mailMsg.Header {
		if name == "From" {
			continue
		}
		if name == "To" {
			continue
		}
		for _, value := range values {
			buf.Write([]byte(fmt.Sprintf("%s: %s\n", name, value)))
		}
	}
	buf.Write([]byte("\n"))
	io.Copy(&buf, mailMsg.Body)
	rewrittenMessage = string(buf.Bytes())
	return
}

// StripDomain removes the domain name from name.
func StripDomain(name string) string {
	i := strings.Index(name, "@")
	if i == -1 {
		return name
	}
	return name[:i]
}

// IsInternal returns true if the name refers to an internal recipient.
func IsInternal(name string) bool {
	i := strings.Index(name, "@")
	if i == -1 {
		return true
	}
	if strings.HasSuffix(name, ubikomShortSuffix) {
		return true
	}
	if strings.HasSuffix(name, ubikomLongSuffix) {
		return true
	}
	return false
}
